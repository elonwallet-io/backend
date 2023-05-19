package handlers

import (
	"fmt"
	"github.com/Leantar/elonwallet-backend/models"
	"github.com/Leantar/elonwallet-backend/server/common"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

func (a *Api) HandleScheduleNotification() echo.HandlerFunc {
	type input struct {
		SendAfter int64  `json:"send_after" validate:"gte=0"`
		Title     string `json:"title" validate:"max=1000"`
		Body      string `json:"body" validate:"max=10000"`
	}

	type output struct {
		ID string `json:"id"`
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

		id, err := uuid.NewRandom()
		if err != nil {
			return fmt.Errorf("failed to generate uuid: %w", err)
		}

		notification := models.Notification{
			ID:           id.String(),
			CreationTime: time.Now().Unix(),
			SendAfter:    in.SendAfter,
			UserID:       user.ID,
			Title:        in.Title,
			Body:         in.Body,
		}

		err = tx.Notifications().CreateNotification(notification, c.Request().Context())
		if err != nil {
			return fmt.Errorf("failed to create notification: %w", err)
		}

		return c.JSON(http.StatusOK, output{ID: id.String()})
	}
}
