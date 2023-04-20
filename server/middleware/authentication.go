package middleware

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"github.com/Leantar/elonwallet-backend/models"
	"github.com/Leantar/elonwallet-backend/server/common"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"net/http"
)

const (
	Enclave = "elonwallet-enclave"
	Backend = "elonwallet-backend"
)

type BackendClaims struct {
	jwt.RegisteredClaims
}

func CheckAuthentication() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			bearer := c.Request().Header.Get("Authorization")
			if len(bearer) < 8 {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or invalid session")
			}

			tx := c.Get("tx").(common.Transaction)

			user, err := validateJWT(bearer[7:], tx, c.Request().Context())
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or invalid session").SetInternal(err)
			}

			c.Set("user", user)

			return next(c)
		}
	}
}

func validateJWT(tokenString string, tx common.Transaction, ctx context.Context) (user models.User, err error) {
	parser := jwt.NewParser(
		jwt.WithIssuedAt(),
		jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
		jwt.WithAudience(Backend),
		jwt.WithIssuer(Enclave),
	)

	var claims BackendClaims
	_, err = parser.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		subject, err := token.Claims.GetSubject()
		if err != nil {
			return nil, err
		}

		user, err = tx.Users().GetUserByEmail(subject, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

		pk, err := hex.DecodeString(user.VerificationKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode verification key: %w", err)
		}

		return ed25519.PublicKey(pk), nil
	})

	return
}
