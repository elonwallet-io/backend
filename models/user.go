package models

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

type User struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Email           string   `json:"email"`
	Wallets         []Wallet `json:"wallets"`
	EnclaveURL      string   `json:"enclave_url"`
	Contacts        []string `json:"contacts"`
	VerificationKey string   `json:"verification_key"`
}

func NewUser(name, email string) (User, error) {
	id, err := generateUniqueUserID()
	if err != nil {
		return User{}, fmt.Errorf("failed to generate uuid: %w", err)
	}

	return User{
		ID:    id,
		Name:  name,
		Email: email,
	}, nil
}

func generateUniqueUserID() (string, error) {
	var charset = []rune("abcdefghijklmnopqrstuvwxyz")
	var charsetLength = new(big.Int).SetInt64(int64(len(charset)))

	var sb strings.Builder
	for i := 0; i < 28; i++ {
		index, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", fmt.Errorf("failed to generate random char: %w", err)
		}
		sb.WriteRune(charset[index.Int64()])
	}

	return sb.String(), nil
}
