package handlers

import (
	"github.com/Leantar/elonwallet-backend/config"
	"github.com/Leantar/elonwallet-backend/server/common"
	"sync"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type Api struct {
	tf         common.TransactionFactory
	cfg        config.Config
	challenges map[string]string //Holds the address of the wallet as the key and the personal sign message challenge as the value. Used to verify ownership of a wallet
	mu         sync.Mutex
}

func NewApi(tf common.TransactionFactory, config config.Config) *Api {
	return &Api{
		tf:         tf,
		cfg:        config,
		challenges: make(map[string]string),
		mu:         sync.Mutex{},
	}
}
