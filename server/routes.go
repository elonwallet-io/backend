package server

import (
	"github.com/Leantar/elonwallet-backend/server/handlers"
	server "github.com/Leantar/elonwallet-backend/server/middleware"
)

func (s *Server) registerRoutes() error {
	api := handlers.NewApi(s.tf, s.cfg)

	s.echo.POST("/users", api.HandleCreateUser())
	s.echo.GET("/users/:email/resend-activation-link", api.HandleResendActivationLink())
	s.echo.POST("/users/:email/activate", api.HandleActivateUser())
	s.echo.GET("/users/:email/enclave-url", api.HandleGetEnclaveURL())
	s.echo.GET("/users/:email", api.HandleGetUser(), server.CheckAuthentication("user"))
	s.echo.POST("/users/my/wallets/initialize", api.HandleAddWalletInitialize(), server.CheckAuthentication("enclave"))
	s.echo.POST("/users/my/wallets/finalize", api.HandleAddWalletFinalize(), server.CheckAuthentication("enclave"))

	s.echo.GET("/:address/balance", api.HandleGetBalance(), server.CheckAuthentication("user"))
	s.echo.GET("/:address/transactions", api.HandleGetTransactions(), server.CheckAuthentication("user"))

	s.echo.GET("/contacts", api.HandleGetContacts(), server.CheckAuthentication("user"))
	s.echo.POST("/contacts", api.HandleCreateContact(), server.CheckAuthentication("user"))
	s.echo.DELETE("/contacts/:email", api.HandleRemoveContact(), server.CheckAuthentication("user"))

	s.echo.POST("/notifications", api.HandleSendNotification(), server.CheckAuthentication("enclave"))
	s.echo.POST("/notifications/series", api.HandleScheduleNotificationSeries(), server.CheckAuthentication("enclave"))
	s.echo.DELETE("/notifications/series/:series_id", api.HandleRemoveScheduledNotificationSeries(), server.CheckAuthentication("enclave"))
	return nil
}
