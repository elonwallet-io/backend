package server

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Leantar/elonwallet-backend/server/common"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"net/http"
)

func CheckAuthentication() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sessionCookie, err := c.Request().Cookie("session")
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or invalid session")
			}

			idCookie, err := c.Request().Cookie("user_email")
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or invalid session")
			}

			tx := c.Get("tx").(common.Transaction)

			user, err := tx.Users().GetUserByEmail(idCookie.Value, c.Request().Context())
			if errors.Is(err, common.ErrNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "User unknown")
			}
			if err != nil {
				return fmt.Errorf("failed to get user by id: %w", err)
			}

			pk, err := hex.DecodeString(user.VerificationKey)
			if err != nil {
				return fmt.Errorf("failed to decode verification key: %w", err)
			}

			if !validateJWT(sessionCookie.Value, pk) {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or invalid session")
			}

			c.Set("user", user)

			return next(c)
		}
	}
}

func validateJWT(tokenString string, pk ed25519.PublicKey) bool {
	parser := jwt.NewParser(
		jwt.WithIssuedAt(),
		jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
	)

	_, err := parser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return pk, nil
	})

	return err == nil
}
