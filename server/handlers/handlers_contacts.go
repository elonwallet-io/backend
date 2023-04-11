package handlers

import (
	"errors"
	"fmt"
	"github.com/Leantar/elonwallet-backend/models"
	"github.com/Leantar/elonwallet-backend/server/common"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (a *Api) HandleGetContacts() echo.HandlerFunc {
	type contact struct {
		ID              string          `json:"-"`
		Name            string          `json:"name"`
		Email           string          `json:"email"`
		Wallets         []models.Wallet `json:"wallets"`
		EnclaveURL      string          `json:"-"`
		Contacts        []string        `json:"-"`
		VerificationKey string          `json:"-"`
	}

	type output struct {
		Contacts []contact `json:"contacts"`
	}
	return func(c echo.Context) error {
		user := c.Get("user").(models.User)
		tx := c.Get("tx").(common.Transaction)

		out := output{
			Contacts: make([]contact, len(user.Contacts)),
		}
		for i, contactId := range user.Contacts {
			con, err := tx.Users().GetUserByID(contactId, c.Request().Context())
			if err != nil {
				return fmt.Errorf("failed to get user: %w", err)
			}
			out.Contacts[i] = contact(con)
		}

		return c.JSON(http.StatusOK, out)
	}
}

func (a *Api) HandleCreateContact() echo.HandlerFunc {
	type input struct {
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

		user := c.Get("user").(models.User)
		tx := c.Get("tx").(common.Transaction)

		con, err := tx.Users().GetUserByEmail(in.Email, c.Request().Context())
		if errors.Is(err, common.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "contact does not exist")
		}
		if err != nil {
			return fmt.Errorf("failed to get contact: %w", err)
		}

		err = tx.Users().AddContactToUser(user.ID, con.ID, c.Request().Context())
		if errors.Is(err, common.ErrNotFound) {
			return echo.NewHTTPError(http.StatusConflict, "contact does already exist")
		}
		if err != nil {
			return fmt.Errorf("failed to get contact: %w", err)
		}

		return c.NoContent(http.StatusCreated)
	}
}

func (a *Api) HandleRemoveContact() echo.HandlerFunc {
	type input struct {
		Email string `validate:"required,email"`
	}
	return func(c echo.Context) error {
		in := input{
			Email: c.Param("email"),
		}
		if err := c.Validate(&in); err != nil {
			return err
		}

		user := c.Get("user").(models.User)
		tx := c.Get("tx").(common.Transaction)

		con, err := tx.Users().GetUserByEmail(in.Email, c.Request().Context())
		if errors.Is(err, common.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "contact does not exist")
		}
		if err != nil {
			return fmt.Errorf("failed to get contact: %w", err)
		}

		err = tx.Users().RemoveContactFromUser(user.ID, con.ID, c.Request().Context())
		if errors.Is(err, common.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "contact does not exist")
		}
		if err != nil {
			return fmt.Errorf("failed to remove contact: %w", err)
		}

		return c.NoContent(http.StatusOK)
	}
}
