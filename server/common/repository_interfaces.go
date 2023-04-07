package common

import (
	"context"
	"github.com/Leantar/elonwallet-backend/models"
)

type UserRepository interface {
	CreateUser(user models.User, ctx context.Context) error
	AddWalletToUser(userID string, wallet models.Wallet, ctx context.Context) error
	AddContactToUser(userID, contactID string, ctx context.Context) error
	RemoveContactFromUser(userID, contactID string, ctx context.Context) error
	GetUserByID(userID string, ctx context.Context) (models.User, error)
	GetUserByEmail(email string, ctx context.Context) (models.User, error)
}

type Transaction interface {
	Commit() error
	Rollback() error
	Users() UserRepository
}

type TransactionFactory interface {
	Begin() (Transaction, error)
	IsErrNotFound(err error) bool
	IsErrConflict(err error) bool
}
