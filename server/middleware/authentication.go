package server

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"github.com/Leantar/elonwallet-backend/server/common"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"net/http"
)

func CheckAuthentication(tf common.TransactionFactory) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sessionCookie, err := c.Request().Cookie("session")
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or invalid session")
			}

			idCookie, err := c.Request().Cookie("user_id")
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or invalid session")
			}

			tx := c.Get("tx").(common.Transaction)

			user, err := tx.Users().GetUserByID(idCookie.Value, c.Request().Context())
			if tf.IsErrNotFound(err) {
				return echo.NewHTTPError(http.StatusNotFound, "User unknown")
			}
			if err != nil {
				return fmt.Errorf("failed to get user by id: %w", err)
			}

			pk, err := getVerificationKey(user.EnclaveURL)
			if err != nil {
				return fmt.Errorf("failed to get verification key: %w", err)
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

func getVerificationKey(enclaveURL string) (ed25519.PublicKey, error) {
	res, err := http.Get(fmt.Sprintf("%s/jwt-verification-key", enclaveURL))
	if err != nil {
		return nil, fmt.Errorf("failed to get verification key: %w", err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("receiver error status code: %d", res.StatusCode)
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
