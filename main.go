package main

import (
	"fmt"
	"github.com/Leantar/elonwallet-backend/config"
	"github.com/Leantar/elonwallet-backend/repository"
	"github.com/Leantar/elonwallet-backend/server"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	if err := run(); err != nil {
		log.Fatal().Caller().Err(err).Msg("failed to start")
	}
}

func run() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM)

	var cfg config.Config
	err := config.FromEnv(&cfg)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	err = validator.New().Struct(cfg)
	if err != nil {
		return fmt.Errorf("validation of config failed: %w", err)
	}

	tf, err := repository.NewTransactionFactory(cfg.DBConnectionString)
	if err != nil {
		return fmt.Errorf("failed to create TransactionFactory: %w", err)
	}

	s := server.New(cfg, tf)
	go func() {
		err := s.Run()
		if err != nil {
			log.Fatal().Caller().Err(err).Msg("failed to start")
		}
	}()

	<-stop

	return s.Stop()
}
