package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/Leantar/elonwallet-backend/config"
	"github.com/Leantar/elonwallet-backend/server/common"
	"github.com/rs/zerolog/log"
	"time"
)

func (s *Server) workOnNotifications(cfg config.EmailConfig) {
	ctx := context.Background()
	for {
		tx, err := s.tf.Begin()
		if err != nil {
			log.Error().Caller().Err(err).Msg("failed to start transaction")
			continue
		}

		err = handlePendingNotifications(tx, cfg, ctx)
		if errors.Is(err, common.ErrNotFound) {
			if err := tx.Commit(); err != nil {
				log.Error().Caller().Err(err).Msg("failed to commit tx")
			}
			sleepTime := 10 * time.Minute
			log.Info().Caller().Msgf("No pending notifications. Sleeping until %v", time.Now().Add(sleepTime))
			time.Sleep(sleepTime)
			continue
		}
		if err != nil {
			log.Error().Caller().Err(err).Msg("failed to handle pending notifications")
			if err := tx.Rollback(); err != nil {
				log.Fatal().Caller().Err(err).Msg("failed to rollback tx")
			}
			continue
		}

		if err := tx.Commit(); err != nil {
			log.Error().Caller().Err(err).Msg("failed to commit tx")
		}
	}
}

func handlePendingNotifications(tx common.Transaction, cfg config.EmailConfig, ctx context.Context) error {
	notifications, err := tx.Notifications().GetPendingNotificationsBatch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending notifications: %w", err)
	}

	for _, notification := range notifications {
		user, err := tx.Users().GetUserByID(notification.UserID, ctx)
		if err != nil {
			return fmt.Errorf("failed to get user by id: %w", err)
		}

		err = common.SendEmail(cfg, user.Email, notification.Title, notification.Body)
		if err != nil {
			log.Error().Caller().Err(err).Str("email", user.Email).Msg("failed to send email to user")
			notification.TimesTried++
			err = tx.Notifications().UpdateNotification(notification, ctx)
			if err != nil {
				return fmt.Errorf("failed to update times tried for notification: %w", err)
			}
			continue
		}

		err = tx.Notifications().DeleteNotification(notification.ID, notification.UserID, ctx)
		if err != nil {
			return fmt.Errorf("failed to delete notification: %w", err)
		}
	}

	return nil
}
