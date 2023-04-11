package handlers

import (
	"github.com/Leantar/elonwallet-backend/config"
	"github.com/Leantar/elonwallet-backend/server/common"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type Api struct {
	tf  common.TransactionFactory
	cfg config.Config
}

func NewApi(tf common.TransactionFactory, config config.Config) *Api {
	return &Api{
		tf:  tf,
		cfg: config,
	}
}
