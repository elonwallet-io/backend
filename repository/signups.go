package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Leantar/elonwallet-backend/models"
	"github.com/Leantar/elonwallet-backend/server/common"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type SignupRepository struct {
	tx *sqlx.Tx
}

func (s *SignupRepository) CreateSignup(signup models.Signup, ctx context.Context) error {
	const query = `INSERT INTO signups("user_id", "activated", "activation_string", "created", "valid_until") VALUES($1,$2,$3,$4,$5)`

	_, err := s.tx.ExecContext(ctx, query, signup.UserID, signup.Activated, signup.ActivationString, signup.Created, signup.ValidUntil)
	if e, ok := err.(*pq.Error); ok && e.Code == postgresUniqueViolationCode {
		err = common.ErrConflict
	}

	return err
}

func (s *SignupRepository) UpdateSignup(signup models.Signup, ctx context.Context) error {
	const query = `UPDATE signups SET "activated" = $1, "activation_string" = $2, "created" = $3,"valid_until" = $4 WHERE "user_id" = $5`

	_, err := s.tx.ExecContext(ctx, query, signup.Activated, signup.ActivationString, signup.Created, signup.ValidUntil, signup.UserID)
	return err
}

func (s *SignupRepository) GetSignup(userID string, ctx context.Context) (models.Signup, error) {
	const query = `SELECT * FROM signups WHERE "user_id" = $1`

	var signup dbSignup
	err := s.tx.GetContext(ctx, &signup, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Signup{}, common.ErrNotFound
		}
		return models.Signup{}, fmt.Errorf("failed to get dbSignup: %w", err)
	}

	return models.Signup(signup), err
}
