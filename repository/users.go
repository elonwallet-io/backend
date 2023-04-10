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

type UserRepository struct {
	tx *sqlx.Tx
}

func (u *UserRepository) CreateUser(user models.User, ctx context.Context) error {
	const query = `INSERT INTO users("id", "name", "email", "enclave_url", "verification_key") VALUES($1,$2,$3,$4,$5)`

	_, err := u.tx.ExecContext(ctx, query, user.ID, user.Name, user.Email, user.EnclaveURL, user.VerificationKey)
	if e, ok := err.(*pq.Error); ok && e.Code == postgresUniqueViolationCode {
		err = common.ErrConflict
	}

	return err
}

func (u *UserRepository) AddWalletToUser(userID string, wallet models.Wallet, ctx context.Context) error {
	const query = `INSERT INTO wallets("address", "name", "user_id") VALUES($1,$2,$3)`

	_, err := u.tx.ExecContext(ctx, query, wallet.Address, wallet.Name, userID)
	if e, ok := err.(*pq.Error); ok && e.Code == postgresUniqueViolationCode {
		err = common.ErrConflict
	}

	return err
}

func (u *UserRepository) AddContactToUser(userID, contactID string, ctx context.Context) error {
	const query = `INSERT INTO contacts("user_id", "contact_id") VALUES($1,$2)`

	_, err := u.tx.ExecContext(ctx, query, userID, contactID)
	if e, ok := err.(*pq.Error); ok && e.Code == postgresUniqueViolationCode {
		err = common.ErrConflict
	}

	return err
}

func (u *UserRepository) RemoveContactFromUser(userID, contactID string, ctx context.Context) error {
	const query = `DELETE FROM contacts WHERE "user_id" = $1 and "contact_id" = $2`

	result, err := u.tx.ExecContext(ctx, query, userID, contactID)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n != 1 {
		return common.ErrNotFound
	}

	return nil
}

func (u *UserRepository) SetEnclaveURLAndVerificationKeyForUser(userID, enclaveURL, verificationKey string, ctx context.Context) error {
	const query = `UPDATE users SET "enclave_url" = $1, "verification_key" = $2 WHERE "id" = $3`

	_, err := u.tx.ExecContext(ctx, query, enclaveURL, verificationKey, userID)
	return err
}

func (u *UserRepository) GetUserByID(userID string, ctx context.Context) (models.User, error) {
	const userQuery = `SELECT * FROM users where "id" = $1`
	const walletQuery = `SELECT * FROM wallets where "user_id" = $1`
	const contactQuery = `SELECT * FROM contacts where "user_id" = $1`

	var user dbUser
	err := u.tx.GetContext(ctx, &user, userQuery, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, common.ErrNotFound
		}
		return models.User{}, fmt.Errorf("failed to get dbUser: %w", err)
	}

	wallets := make([]dbWallet, 0)
	err = u.tx.SelectContext(ctx, &wallets, walletQuery, userID)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to get dbWallets: %w", err)
	}

	contacts := make([]dbContact, 0)
	err = u.tx.SelectContext(ctx, &contacts, contactQuery, userID)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to get dbContacts: %w", err)
	}

	return models.User{
		ID:              user.ID,
		Name:            user.Name,
		Email:           user.Email,
		Wallets:         mapWallets(wallets),
		EnclaveURL:      user.EnclaveURL,
		Contacts:        mapContacts(contacts),
		VerificationKey: user.VerificationKey,
	}, nil
}

func (u *UserRepository) GetUserByEmail(email string, ctx context.Context) (models.User, error) {
	const userQuery = `SELECT * FROM users where "email" = $1`
	const walletQuery = `SELECT * FROM wallets where "user_id" = $1`
	const contactQuery = `SELECT * FROM contacts where "user_id" = $1`

	var user dbUser
	err := u.tx.GetContext(ctx, &user, userQuery, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, common.ErrNotFound
		}
		return models.User{}, fmt.Errorf("failed to get dbUser: %w", err)
	}

	wallets := make([]dbWallet, 0)
	err = u.tx.SelectContext(ctx, &wallets, walletQuery, user.ID)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to get dbWallets: %w", err)
	}

	contacts := make([]dbContact, 0)
	err = u.tx.SelectContext(ctx, &contacts, contactQuery, user.ID)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to get dbContacts: %w", err)
	}

	return models.User{
		ID:              user.ID,
		Name:            user.Name,
		Email:           user.Email,
		Wallets:         mapWallets(wallets),
		EnclaveURL:      user.EnclaveURL,
		Contacts:        mapContacts(contacts),
		VerificationKey: user.VerificationKey,
	}, nil
}

func mapWallets(wallets []dbWallet) []models.Wallet {
	mapped := make([]models.Wallet, len(wallets))
	for i, wallet := range wallets {
		mapped[i] = models.Wallet{
			Name:    wallet.Name,
			Address: wallet.Address,
		}
	}
	return mapped
}

func mapContacts(contacts []dbContact) []string {
	mapped := make([]string, len(contacts))
	for i, contact := range contacts {
		mapped[i] = contact.ContactID
	}
	return mapped
}
