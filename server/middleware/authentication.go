package middleware

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Leantar/elonwallet-backend/models"
	"github.com/Leantar/elonwallet-backend/server/common"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slices"
	"net/http"
)

const (
	Enclave        = "elonwallet-enclave"
	Backend        = "elonwallet-backend"
	invalidSession = "Invalid or malformed jwt"
)

type BackendClaims struct {
	Scope string `json:"scope"`
	jwt.RegisteredClaims
}

func CheckAuthentication(allowedScopes ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			bearer := c.Request().Header.Get("Authorization")
			if len(bearer) < 8 {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or invalid session")
			}

			tx := c.Get("tx").(common.Transaction)

			user, claims, err := validateJWT(bearer[7:], tx, c.Request().Context())
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, invalidSession).SetInternal(err)
			}

			if !isAllowedScope(allowedScopes, claims) {
				return echo.NewHTTPError(http.StatusUnauthorized, invalidSession)
			}

			c.Set("user", user)

			return next(c)
		}
	}
}

func validateJWT(tokenString string, tx common.Transaction, ctx context.Context) (user models.User, claims BackendClaims, err error) {
	parser := jwt.NewParser(
		jwt.WithIssuedAt(),
		jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
		jwt.WithAudience(Backend),
		jwt.WithIssuer(Enclave),
	)

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

	if claims.Scope == "" {
		return models.User{}, BackendClaims{}, errors.New("scope is missing")
	}

	return
}

func isAllowedScope(scopes []string, claims BackendClaims) bool {
	return slices.Contains(scopes, claims.Scope)
}
