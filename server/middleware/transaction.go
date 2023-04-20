package middleware

import (
	"fmt"
	"github.com/Leantar/elonwallet-backend/server/common"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func ManageTransaction(tf common.TransactionFactory) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tx, err := tf.Begin()
			if err != nil {
				return fmt.Errorf("failed to begin tx: %w", err)
			}

			c.Set("tx", tx)

			err = next(c)
			if err != nil {
				if err := tx.Rollback(); err != nil {
					log.Fatal().Caller().Err(err).Msg("failed to rollback")
				}
				return err
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit: %w", err)
			}

			return nil
		}
	}
}
