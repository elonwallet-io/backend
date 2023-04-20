package server

import (
	"context"
	"github.com/Leantar/elonwallet-backend/config"
	"github.com/Leantar/elonwallet-backend/server/common"
	customMiddleware "github.com/Leantar/elonwallet-backend/server/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"time"
)

type Server struct {
	echo *echo.Echo
	cfg  config.Config
	tf   common.TransactionFactory
}

func New(cfg config.Config, tf common.TransactionFactory) *Server {
	e := echo.New()
	e.Server.ReadTimeout = 5 * time.Second
	e.Server.WriteTimeout = 10 * time.Second
	e.Server.IdleTimeout = 120 * time.Second

	cv := CustomValidator{
		validator: validator.New(),
	}

	e.Binder = &BinderWithURLDecoding{&echo.DefaultBinder{}}
	e.Validator = &cv
	e.Use(middleware.RequestID())
	e.Use(customMiddleware.RequestLogger())
	e.Use(customMiddleware.Cors(cfg.FrontendURL))
	e.Use(customMiddleware.ManageTransaction(tf))

	return &Server{
		echo: e,
		cfg:  cfg,
		tf:   tf,
	}
}

func (s *Server) Run() (err error) {
	err = s.registerRoutes()
	if err != nil {
		return
	}

	err = s.echo.Start("0.0.0.0:8080")
	if err == http.ErrServerClosed {
		err = nil
	}

	return
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.echo.Shutdown(ctx)
}
