package repository

type dbUser struct {
	ID         string `db:"id"`
	Name       string `db:"name"`
	Email      string `db:"email"`
	EnclaveURL string `db:"enclave_url"`
}

type dbWallet struct {
	Address string `db:"address"`
	Name    string `db:"name"`
	UserID  string `db:"user_id"`
}

type dbContact struct {
	UserID    string `db:"user_id"`
	ContactID string `db:"contact_id"`
}
