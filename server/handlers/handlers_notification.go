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

func (a *Api) HandleSendNotification() echo.HandlerFunc {
	type input struct {
		Title string `json:"title" validate:"required,max=1000"`
		Body  string `json:"body" validate:"required,max=10000"`
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

		err := common.SendEmail(a.cfg.Email, user.Email, in.Title, in.Body)
		if err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)
	}
}

func (a *Api) HandleScheduleNotificationSeries() echo.HandlerFunc {
	type notification struct {
		SendAfter int64  `json:"send_after" validate:"gte=0"`
		Title     string `json:"title" validate:"required,max=1000"`
		Body      string `json:"body" validate:"required,max=10000"`
	}

	type input struct {
		Notifications []notification `json:"notifications" validate:"required,gt=0,dive"`
	}

	type output struct {
		SeriesID string `json:"series_id"`
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

		seriesID, err := uuid.NewRandom()
		if err != nil {
			return fmt.Errorf("failed to generate uuid: %w", err)
		}

		creationTime := time.Now().Unix()
		nfs := make([]models.Notification, len(in.Notifications))

		for i, n := range in.Notifications {
			nf := models.Notification{
				ID:           0,
				SeriesID:     seriesID.String(),
				CreationTime: creationTime,
				SendAfter:    n.SendAfter,
				UserID:       user.ID,
				Title:        n.Title,
				Body:         n.Body,
			}

			nfs[i] = nf
		}
		err = tx.Notifications().CreateNotificationSeries(nfs, c.Request().Context())
		if err != nil {
			return fmt.Errorf("failed to create notification: %w", err)
		}

		return c.JSON(http.StatusAccepted, output{seriesID.String()})
	}
}

func (a *Api) HandleRemoveScheduledNotificationSeries() echo.HandlerFunc {
	type input struct {
		SeriesID string `param:"series_id" validate:"uuid4"`
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

		err := tx.Notifications().DeleteNotificationSeries(in.SeriesID, user.ID, c.Request().Context())
		if err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)
	}
}
