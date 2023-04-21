package handlers

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Leantar/elonwallet-backend/config"
	"github.com/Leantar/elonwallet-backend/models"
	"github.com/Leantar/elonwallet-backend/server/common"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"
	"time"
)

func (a *Api) HandleCreateUserViaPostman() echo.HandlerFunc {
	type wallet struct {
		Name    string `json:"name" validate:"required,alphanum"`
		Address string `json:"address" validate:"required,eth_addr"`
	}
	type input struct {
		Name    string   `json:"name" validate:"required"`
		Email   string   `json:"email" validate:"required,email"`
		Wallets []wallet `json:"wallets"`
	}

	return func(c echo.Context) error {
		var in input
		if err := c.Bind(&in); err != nil {
			return err
		}
		if err := c.Validate(&in); err != nil {
			return err
		}

		user, err := newUser(in.Name, in.Email)
		if err != nil {
			return err
		}

		tx := c.Get("tx").(common.Transaction)

		err = tx.Users().CreateUser(user, c.Request().Context())
		if errors.Is(err, common.ErrConflict) {
			return echo.NewHTTPError(http.StatusConflict, "target resource does already exist")
		}
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		for _, w := range in.Wallets {
			err = tx.Users().AddWalletToUser(user.ID, models.Wallet(w), c.Request().Context())
			if errors.Is(err, common.ErrConflict) {
				return echo.NewHTTPError(http.StatusConflict, "target resource does already exist")
			}
			if err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		}

		return c.NoContent(http.StatusCreated)
	}
}

func (a *Api) HandleAddWallet() echo.HandlerFunc {
	type input struct {
		Name    string `json:"name" validate:"required,alphanum"`
		Address string `json:"address" validate:"required,eth_addr"`
	}

	return func(c echo.Context) error {
		var in input
		if err := c.Bind(&in); err != nil {
			return err
		}
		if err := c.Validate(&in); err != nil {
			return err
		}

		user := c.Get("user").(models.User)
		tx := c.Get("tx").(common.Transaction)

		err := tx.Users().AddWalletToUser(user.ID, models.Wallet(in), c.Request().Context())
		if errors.Is(err, common.ErrConflict) {
			return echo.NewHTTPError(http.StatusConflict, "target resource does already exist")
		}
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		return c.NoContent(http.StatusCreated)
	}
}

func (a *Api) HandleCreateUser() echo.HandlerFunc {
	type input struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	return func(c echo.Context) error {
		var in input
		if err := c.Bind(&in); err != nil {
			return err
		}
		if err := c.Validate(&in); err != nil {
			return err
		}

		tx := c.Get("tx").(common.Transaction)

		user, err := createUser(in.Name, in.Email, tx, c.Request().Context())
		if err != nil {
			return err
		}

		signup, err := createSignup(user.ID, tx, c.Request().Context())
		if err != nil {
			return err
		}

		err = sendActivationLink(user, signup, a.cfg)
		if err != nil {
			return err
		}

		return c.NoContent(http.StatusCreated)
	}
}

func (a *Api) HandleResendActivationLink() echo.HandlerFunc {
	type input struct {
		Email string `param:"email" validate:"required,email"`
	}

	return func(c echo.Context) error {
		var in input
		if err := c.Bind(&in); err != nil {
			return err
		}
		if err := c.Validate(&in); err != nil {
			return err
		}

		tx := c.Get("tx").(common.Transaction)

		user, err := tx.Users().GetUserByEmail(in.Email, c.Request().Context())
		if errors.Is(err, common.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "user does not exist")
		}
		if err != nil {
			return fmt.Errorf("failed to get user by email: %w", err)
		}

		oldSignup, err := tx.Signups().GetSignup(user.ID, c.Request().Context())
		if err != nil {
			return fmt.Errorf("failed to get old signup: %w", err)
		}

		if oldSignup.Activated {
			return echo.NewHTTPError(http.StatusBadRequest, "user is already activated")
		}

		if time.Unix(oldSignup.Created, 0).Add(time.Minute * 15).Before(time.Now()) {
			return echo.NewHTTPError(http.StatusBadRequest, "please wait at least 15 minutes before requesting a new activation link")
		}

		signup, err := recreateSignup(user.ID, tx, c.Request().Context())
		if err != nil {
			return err
		}

		err = sendActivationLink(user, signup, a.cfg)
		if err != nil {
			return err
		}

		return c.NoContent(http.StatusCreated)
	}
}

func (a *Api) HandleActivateUser() echo.HandlerFunc {
	type input struct {
		Email            string `param:"email" validate:"required,email"`
		ActivationString string `json:"activation_string" validate:"required,hexadecimal,len=64"`
	}
	return func(c echo.Context) error {
		var in input
		if err := c.Bind(&in); err != nil {
			return err
		}
		if err := c.Validate(&in); err != nil {
			return err
		}

		tx := c.Get("tx").(common.Transaction)

		user, err := tx.Users().GetUserByEmail(in.Email, c.Request().Context())
		if errors.Is(err, common.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "user does not exist")
		}
		if err != nil {
			return fmt.Errorf("failed to get user by email: %w", err)
		}

		signup, err := tx.Signups().GetSignup(user.ID, c.Request().Context())
		if err != nil {
			return fmt.Errorf("failed to get signup: %w", err)
		}

		if signup.ValidUntil < time.Now().Unix() {
			return echo.NewHTTPError(http.StatusBadRequest, "invitation link has expired")
		}

		if signup.ActivationString != in.ActivationString {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid activation link")
		}

		signup.Activated = true
		err = tx.Signups().UpdateSignup(signup, c.Request().Context())
		if err != nil {
			return fmt.Errorf("failed to save updated signup: %w", err)
		}

		enclaveURL, err := deployEnclave(a.cfg.DeployerURL, user.ID)
		if err != nil {
			return err
		}

		pk, err := getVerificationKey(enclaveURL)
		if err != nil {
			return err
		}

		enclaveURL = strings.ReplaceAll(enclaveURL, "host.docker.internal", "localhost")

		err = tx.Users().SetEnclaveURLAndVerificationKeyForUser(user.ID, enclaveURL, hex.EncodeToString(pk), c.Request().Context())
		if err != nil {
			return fmt.Errorf("failed to save enclave url: %w", err)
		}

		return c.NoContent(http.StatusOK)
	}
}

func (a *Api) HandleGetUser() echo.HandlerFunc {
	type input struct {
		Email string `param:"email" validate:"required,email"`
	}

	type output struct {
		ID              string          `json:"-"`
		Name            string          `json:"name"`
		Email           string          `json:"email"`
		Wallets         []models.Wallet `json:"wallets"`
		EnclaveURL      string          `json:"-"`
		Contacts        []string        `json:"-"`
		VerificationKey string          `json:"-"`
	}
	return func(c echo.Context) error {
		var in input
		if err := c.Bind(&in); err != nil {
			return err
		}
		if err := c.Validate(&in); err != nil {
			return err
		}

		tx := c.Get("tx").(common.Transaction)

		user, err := tx.Users().GetUserByEmail(in.Email, c.Request().Context())
		if errors.Is(err, common.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "user does not exist")
		}
		if err != nil {
			return fmt.Errorf("failed to get user by email: %w", err)
		}

		return c.JSON(http.StatusOK, output(user))
	}
}

func (a *Api) HandleGetEnclaveURL() echo.HandlerFunc {
	type input struct {
		Email string `param:"email" validate:"required,email"`
	}

	type output struct {
		EnclaveURL string `json:"enclave_url"`
	}
	return func(c echo.Context) error {
		var in input
		if err := c.Bind(&in); err != nil {
			return err
		}
		if err := c.Validate(&in); err != nil {
			return err
		}

		tx := c.Get("tx").(common.Transaction)

		user, err := tx.Users().GetUserByEmail(in.Email, c.Request().Context())
		if errors.Is(err, common.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "user does not exist")
		}
		if err != nil {
			return fmt.Errorf("failed to get user by email: %w", err)
		}

		if user.EnclaveURL == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "user has not been verified yet")
		}

		return c.JSON(http.StatusOK, output{user.EnclaveURL})
	}
}

func newUser(name, email string) (models.User, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return models.User{}, fmt.Errorf("failed to generate uuid: %w", err)
	}

	return models.User{
		ID:    id.String(),
		Name:  name,
		Email: email,
	}, nil
}

func newSignup(userID string) (models.Signup, error) {
	buf := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return models.Signup{}, fmt.Errorf("failed to create activation string: %w", err)
	}

	return models.Signup{
		UserID:           userID,
		Activated:        false,
		ActivationString: hex.EncodeToString(buf),
		ValidUntil:       time.Now().Add(time.Hour * 336).Unix(),
	}, nil
}

func createUser(name string, email string, tx common.Transaction, ctx context.Context) (models.User, error) {
	user, err := newUser(name, email)
	if err != nil {
		return models.User{}, nil
	}

	err = tx.Users().CreateUser(user, ctx)
	if errors.Is(err, common.ErrConflict) {
		return models.User{}, echo.NewHTTPError(http.StatusConflict, "user does already exist")
	}
	if err != nil {
		return models.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func createSignup(userID string, tx common.Transaction, ctx context.Context) (models.Signup, error) {
	signup, err := newSignup(userID)
	if err != nil {
		return models.Signup{}, err
	}

	err = tx.Signups().CreateSignup(signup, ctx)
	if errors.Is(err, common.ErrConflict) {
		return models.Signup{}, echo.NewHTTPError(http.StatusConflict, "user has already signed up")
	}
	if err != nil {
		return models.Signup{}, fmt.Errorf("failed to create signup: %w", err)
	}

	return signup, nil
}

func recreateSignup(userID string, tx common.Transaction, ctx context.Context) (models.Signup, error) {
	signup, err := newSignup(userID)
	if err != nil {
		return models.Signup{}, err
	}

	err = tx.Signups().UpdateSignup(signup, ctx)
	if err != nil {
		return models.Signup{}, fmt.Errorf("failed to create signup: %w", err)
	}

	return signup, nil
}

func sendActivationLink(user models.User, signup models.Signup, cfg config.Config) error {
	receiver := []string{user.Email}
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("From: %s\r\n", cfg.Email.User))
	builder.WriteString(fmt.Sprintf("To: %s\r\n", user.Email))
	builder.WriteString("Subject: Activate your Elonwallet.io Account\r\n\r\n")
	builder.WriteString("Please follow the link below to activate your account:\r\n")
	builder.WriteString(fmt.Sprintf("%s/activate?user=%s&activation_string=%s\r\n", cfg.FrontendURL, url.QueryEscape(user.Email), signup.ActivationString))

	auth := smtp.PlainAuth("", cfg.Email.User, cfg.Email.Password, cfg.Email.AuthHost)
	err := smtp.SendMail(cfg.Email.SmtpHost, auth, cfg.Email.User, receiver, []byte(builder.String()))
	if err != nil {
		return fmt.Errorf("failed to send mail: %w", err)
	}
	return nil
}

func getVerificationKey(enclaveURL string) (ed25519.PublicKey, error) {
	res, err := http.Get(fmt.Sprintf("%s/jwt-verification-key", enclaveURL))
	if err != nil {
		return nil, fmt.Errorf("failed to get verification key: %w", err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("received error status code: %d", res.StatusCode)
	}

	type input struct {
		VerificationKey []byte `json:"verification_key"`
	}

	var in input
	if err := json.NewDecoder(res.Body).Decode(&in); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return in.VerificationKey, nil
}

func deployEnclave(deployerURL, name string) (string, error) {
	type payload struct {
		Name string `json:"name"`
	}

	body, err := json.Marshal(payload{name})
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %w", err)
	}

	res, err := http.Post(fmt.Sprintf("%s/enclaves", deployerURL), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to deploy enclave: %w", err)
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("received error status code: %d", res.StatusCode)
	}

	type input struct {
		EnclaveURL string `json:"url"`
	}

	var in input
	if err := json.NewDecoder(res.Body).Decode(&in); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return in.EnclaveURL, nil
}
