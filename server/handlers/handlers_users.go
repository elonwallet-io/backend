package handlers

import (
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
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slices"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (a *Api) HandleAddWalletInitialize() echo.HandlerFunc {
	type input struct {
		Address string `json:"address" validate:"required,ethereum_address"`
	}

	type output struct {
		Challenge string `json:"challenge"`
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
		if walletExists(in.Address, user) {
			return echo.NewHTTPError(http.StatusConflict, "Wallet is already registered")
		}

		challenge, err := getChallenge()
		if err != nil {
			return err
		}

		a.mu.Lock()
		a.challenges[in.Address] = challenge
		a.mu.Unlock()

		return c.JSON(http.StatusOK, output{Challenge: challenge})
	}
}

func (a *Api) HandleAddWalletFinalize() echo.HandlerFunc {
	type input struct {
		Name      string `json:"name" validate:"required,alphanum"`
		Address   string `json:"address" validate:"required,ethereum_address"`
		Signature string `json:"signature" validate:"required,hexadecimal"`
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

		a.mu.Lock()
		defer a.mu.Unlock()
		challenge, ok := a.challenges[in.Address]
		if !ok {
			return echo.NewHTTPError(http.StatusConflict, "No challenge requested for this address")
		}
		delete(a.challenges, in.Address)

		valid, err := verifyPersonalSignature(challenge, in.Signature, in.Address)
		if err != nil {
			return err
		}
		if !valid {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid signature")
		}

		err = tx.Users().AddWalletToUser(user.ID, models.Wallet{
			Name:    in.Name,
			Address: in.Address,
		}, c.Request().Context())
		if err != nil {
			return fmt.Errorf("failed to add wallet to user: %w", err)
		}

		if len(user.Wallets) == 0 { //Send some initial MATIC tokens to new users
			err = sendMumbaiTestMatic(in.Address, a.cfg.Wallet, c.Request().Context())
			if err != nil {
				return err
			}
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
			return echo.NewHTTPError(http.StatusNotFound)
		}
		if err != nil {
			return fmt.Errorf("failed to get user by email: %w", err)
		}

		oldSignup, err := tx.Signups().GetSignup(user.ID, c.Request().Context())
		if err != nil {
			return fmt.Errorf("failed to get old signup: %w", err)
		}

		if oldSignup.Activated {
			return echo.NewHTTPError(http.StatusBadRequest, "User is already activated")
		}

		if time.Unix(oldSignup.Created, 0).Add(time.Minute * 15).Before(time.Now()) {
			return echo.NewHTTPError(http.StatusBadRequest, "Please wait at least 15 minutes before requesting a new activation link")
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
			return echo.NewHTTPError(http.StatusNotFound)
		}
		if err != nil {
			return fmt.Errorf("failed to get user by email: %w", err)
		}

		signup, err := tx.Signups().GetSignup(user.ID, c.Request().Context())
		if err != nil {
			return fmt.Errorf("failed to get signup: %w", err)
		}

		if signup.Activated {
			return echo.NewHTTPError(http.StatusBadRequest, "User is already activated")
		}

		if time.Now().After(time.Unix(signup.ValidUntil, 0)) {
			return echo.NewHTTPError(http.StatusBadRequest, "The activation link has expired")
		}

		if signup.ActivationString != in.ActivationString {
			return echo.NewHTTPError(http.StatusBadRequest, "The activation link is invalid")
		}

		signup.Activated = true
		err = tx.Signups().UpdateSignup(signup, c.Request().Context())
		if err != nil {
			return fmt.Errorf("failed to save updated signup: %w", err)
		}

		deployerApiClient := common.NewDeployerApiClient(a.cfg.DeployerURL)
		enclaveURL, err := deployerApiClient.DeployEnclave(user.ID)
		if err != nil {
			return err
		}

		pk, err := getVerificationKey(enclaveURL)
		if err != nil {
			return err
		}

		err = tx.Users().SetEnclaveURLAndVerificationKeyForUser(user.ID, enclaveURL, hex.EncodeToString(pk), c.Request().Context())
		if err != nil {
			return fmt.Errorf("failed to save enclave url: %w", err)
		}

		if a.cfg.Environment == "docker" {
			enclaveURL = strings.ReplaceAll(enclaveURL, "host.docker.internal", "localhost")
		}

		return c.JSON(http.StatusOK, output{enclaveURL})
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
			return echo.NewHTTPError(http.StatusNotFound)
		}
		if err != nil {
			return fmt.Errorf("failed to get user by email: %w", err)
		}

		if user.EnclaveURL == "" {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		return c.JSON(http.StatusOK, output(user))
	}
}

func (a *Api) HandleGetEnclaveURL() echo.HandlerFunc {
	type input struct {
		Email      string `param:"email" validate:"required,email"`
		Questioner string `query:"questioner"`
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
			return echo.NewHTTPError(http.StatusNotFound)
		}
		if err != nil {
			return fmt.Errorf("failed to get user by email: %w", err)
		}

		if user.EnclaveURL == "" {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		if a.cfg.Environment == "docker" && in.Questioner != "enclave" {
			user.EnclaveURL = strings.ReplaceAll(user.EnclaveURL, "host.docker.internal", "localhost")
		}

		return c.JSON(http.StatusOK, output{user.EnclaveURL})
	}
}

func (a *Api) HandleRemoveUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(models.User)
		tx := c.Get("tx").(common.Transaction)

		err := tx.Users().RemoveUser(user.ID, c.Request().Context())
		if errors.Is(err, common.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		if err != nil {
			return fmt.Errorf("failed to remove user: %w", err)
		}

		deployerApiClient := common.NewDeployerApiClient(a.cfg.DeployerURL)
		err = deployerApiClient.RemoveEnclave(user.ID)
		if err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)
	}
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
	user, err := models.NewUser(name, email)
	if err != nil {
		return models.User{}, err
	}

	err = tx.Users().CreateUser(user, ctx)
	if errors.Is(err, common.ErrConflict) {
		return models.User{}, echo.NewHTTPError(http.StatusConflict, "User does already exist")
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
		return models.Signup{}, echo.NewHTTPError(http.StatusConflict, "User has already signed up")
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
	title := "Activate your Elonwallet.io Account"
	body := "Please follow the link below to activate your account:\r\n"
	body += fmt.Sprintf("%s/activate?user=%s&activation_string=%s\r\n", cfg.FrontendURL, url.QueryEscape(user.Email), signup.ActivationString)

	return common.SendEmail(cfg.Email, user.Email, title, body)
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

func walletExists(address string, user models.User) bool {
	addr := strings.ToLower(address)
	return slices.ContainsFunc(user.Wallets, func(wallet models.Wallet) bool {
		return strings.ToLower(wallet.Address) == addr
	})
}

func getChallenge() (string, error) {
	buf := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", fmt.Errorf("failed to generate challenge: %w", err)
	}

	return hex.EncodeToString(buf), nil
}
