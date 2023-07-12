package server

import (
	"context"
	"crypto/tls"
	"github.com/Leantar/elonwallet-backend/config"
	"github.com/Leantar/elonwallet-backend/server/common"
	customMiddleware "github.com/Leantar/elonwallet-backend/server/middleware"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
	"net/http"
	"time"
)

type Server struct {
	echo   *echo.Echo
	cfg    config.Config
	tf     common.TransactionFactory
	tlsMgr *autocert.Manager
}

func New(cfg config.Config, tf common.TransactionFactory) (*Server, error) {
	e := echo.New()
	s := &Server{
		echo: e,
		cfg:  cfg,
		tf:   tf,
	}

	if cfg.UseInsecureHTTP {
		e.Server.ReadTimeout = 5 * time.Second
		e.Server.WriteTimeout = 120 * time.Second
		e.Server.IdleTimeout = 120 * time.Second
		e.Server.ErrorLog = e.StdLogger
		e.Server.Addr = "0.0.0.0:8080"
	} else {
		log.Error().Caller().Msgf("value is: %s", cfg.BackendHost)

		tlsMgr := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache("/data/.cache"),
			HostPolicy: autocert.HostWhitelist(cfg.BackendHost),
		}
		s.tlsMgr = &tlsMgr

		e.TLSServer.ReadTimeout = 5 * time.Second
		e.TLSServer.WriteTimeout = 30 * time.Second
		e.TLSServer.IdleTimeout = 120 * time.Second
		e.Server.ErrorLog = e.StdLogger
		e.TLSServer.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			},
			GetCertificate: s.tlsMgr.GetCertificate,
			NextProtos:     []string{acme.ALPNProto},
		}
		e.TLSServer.Addr = "0.0.0.0:8443"
	}

	cv := newValidator()
	e.Binder = &BinderWithURLDecoding{&echo.DefaultBinder{}}
	e.Validator = &cv

	e.Use(middleware.RequestID())
	e.Use(customMiddleware.RequestLogger())
	e.Use(customMiddleware.Cors(cfg.FrontendURL))
	e.Use(customMiddleware.ManageTransaction(tf))

	return s, nil
}

func (s *Server) Run() (err error) {
	err = s.registerRoutes()
	if err != nil {
		return
	}

	go s.workOnNotifications(s.cfg.Email)

	if s.cfg.UseInsecureHTTP {
		log.Info().Caller().Msgf("http server started on %s", s.echo.Server.Addr)
		err = s.echo.Server.ListenAndServe()
	} else {
		log.Info().Caller().Msgf("https server started on %s", s.echo.TLSServer.Addr)
		err = s.echo.TLSServer.ListenAndServeTLS("", "")
	}

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
