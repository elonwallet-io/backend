package server

import (
	"github.com/Leantar/elonwallet-backend/server/handlers"
	server "github.com/Leantar/elonwallet-backend/server/middleware"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

func (s *Server) registerRoutes() error {
	api := handlers.NewApi(s.tf, s.cfg)

	s.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{s.cfg.CorsAllowedUrl},
		AllowMethods:     []string{http.MethodHead, http.MethodGet, http.MethodPost, http.MethodPut},
		AllowCredentials: true,
	}))
	s.echo.Use(server.ManageTransaction(s.tf))

	//TODO remove this endpoint
	s.echo.POST("/postman", api.HandleCreateUserViaPostman())

	s.echo.POST("/users", api.HandleCreateUser())
	s.echo.GET("/users/:email/resend-activation-link", api.HandleResendActivationLink())
	s.echo.POST("/users/:email/activate", api.HandleActivateUser())
	s.echo.GET("/users/:email/enclave-url", api.HandleGetEnclaveURL())
	s.echo.POST("/users/:email/wallets", api.HandleAddWallet(), server.CheckAuthentication())
	s.echo.GET("/users/:email", api.HandleGetUser(), server.CheckAuthentication())

	s.echo.GET("/:address/balance", api.HandleGetBalance(), server.CheckAuthentication())
	s.echo.GET("/:address/transactions", api.HandleGetTransactions(), server.CheckAuthentication())

	s.echo.GET("/contacts", api.HandleGetContacts(), server.CheckAuthentication())
	s.echo.POST("/contacts", api.HandleCreateContact(), server.CheckAuthentication())
	s.echo.DELETE("/contacts/:email", api.HandleRemoveContact(), server.CheckAuthentication())
	return nil
}
