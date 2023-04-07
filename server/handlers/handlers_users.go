package handlers

import (
	"fmt"
	"github.com/Leantar/elonwallet-backend/models"
	"github.com/Leantar/elonwallet-backend/server/common"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/url"
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

		user, err := newUser(in.Name, in.Email, a.cfg.EnclaveURL)
		if err != nil {
			return err
		}

		tx := c.Get("tx").(common.Transaction)

		err = tx.Users().CreateUser(user, c.Request().Context())
		if a.tf.IsErrConflict(err) {
			return echo.NewHTTPError(http.StatusConflict, "target resource does already exist")
		}
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		for _, w := range in.Wallets {
			err = tx.Users().AddWalletToUser(user.ID, models.Wallet(w), c.Request().Context())
			if a.tf.IsErrConflict(err) {
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
		if a.tf.IsErrConflict(err) {
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

		user, err := newUser(in.Name, in.Email, a.cfg.EnclaveURL)
		if err != nil {
			return err
		}

		tx := c.Get("tx").(common.Transaction)

		err = tx.Users().CreateUser(user, c.Request().Context())
		if a.tf.IsErrConflict(err) {
			return echo.NewHTTPError(http.StatusConflict, "target resource does already exist")
		}
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		//TODO Set this when fetching enclave url
		c.SetCookie(&http.Cookie{
			Name:     "user_id",
			Value:    user.ID,
			Expires:  time.Now().Add(time.Hour * 24),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
		})

		return c.NoContent(http.StatusCreated)
	}
}

func (a *Api) HandleGetUser() echo.HandlerFunc {
	type input struct {
		Email string `validate:"required,email"`
	}

	type output struct {
		ID         string          `json:"-"`
		Name       string          `json:"name"`
		Email      string          `json:"email"`
		Wallets    []models.Wallet `json:"wallets"`
		EnclaveURL string          `json:"-"`
		Contacts   []string        `json:"-"`
	}
	return func(c echo.Context) error {
		email, err := url.QueryUnescape(c.Param("email"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid escape sequence").SetInternal(err)
		}
		in := input{
			Email: email,
		}
		if err := c.Validate(&in); err != nil {
			return err
		}

		tx := c.Get("tx").(common.Transaction)

		user, err := tx.Users().GetUserByEmail(in.Email, c.Request().Context())
		if a.tf.IsErrNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "user does not exist")
		}
		if err != nil {
			return fmt.Errorf("failed to get user by email: %w", err)
		}

		return c.JSON(http.StatusOK, output(user))
	}
}

func newUser(name, email, enclaveURL string) (models.User, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return models.User{}, fmt.Errorf("failed to generate uuid: %w", err)
	}

	return models.User{
		ID:         id.String(),
		Name:       name,
		Email:      email,
		Wallets:    nil,
		EnclaveURL: enclaveURL,
	}, nil
}
