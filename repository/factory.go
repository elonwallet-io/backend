package repository

import (
	"fmt"
	"github.com/Leantar/elonwallet-backend/server/common"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	postgresUniqueViolationCode = "23505"
)

type TransactionFactory struct {
	db *sqlx.DB
}

func (tf *TransactionFactory) migrate() error {
	for _, schema := range schemas {
		_, err := tf.db.Exec(schema)
		if err != nil {
			return err
		}
	}

	return nil

}

func (tf *TransactionFactory) Begin() (common.Transaction, error) {
	tx, err := tf.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}

	return &Transaction{tx}, nil
}

func NewTransactionFactory(connectionString string) (*TransactionFactory, error) {
	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	tf := TransactionFactory{db: db}

	err = tf.migrate()
	if err != nil {
		return nil, fmt.Errorf("failed to migrate db: %w", err)
	}

	return &tf, nil
}
