package models

type User struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Email           string   `json:"email"`
	Wallets         []Wallet `json:"wallets"`
	EnclaveURL      string   `json:"enclave_url"`
	Contacts        []string `json:"contacts"`
	VerificationKey string   `json:"verification_key"`
}
