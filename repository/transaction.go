package repository

import (
	"fmt"
	"github.com/Leantar/elonwallet-backend/server/common"
	"github.com/jmoiron/sqlx"
)

type Transaction struct {
	tx *sqlx.Tx
}

func (t *Transaction) Commit() error {
	err := t.tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}
	return nil
}

func (t *Transaction) Rollback() error {
	err := t.tx.Rollback()
	if err != nil {
		return fmt.Errorf("failed to rollback tx: %w", err)
	}
	return nil
}

func (t *Transaction) Users() common.UserRepository {
	return &UserRepository{tx: t.tx}
}

func (t *Transaction) Signups() common.SignupRepository {
	return &SignupRepository{tx: t.tx}
}
