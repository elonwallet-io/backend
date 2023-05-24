package common

import (
	"context"
	"errors"
	"github.com/Leantar/elonwallet-backend/models"
)

var (
	ErrNotFound = errors.New("target resource not found")
	ErrConflict = errors.New("target resource does already exist")
)

type NotificationRepository interface {
	CreateNotificationSeries(notifications []models.Notification, ctx context.Context) error
	GetPendingNotificationsBatch(ctx context.Context) ([]models.Notification, error)
	UpdateNotification(notification models.Notification, ctx context.Context) error
	DeleteNotification(id int64, userID string, ctx context.Context) error
	DeleteNotificationSeries(seriesID string, userID string, ctx context.Context) error
}

type SignupRepository interface {
	CreateSignup(signup models.Signup, ctx context.Context) error
	UpdateSignup(signup models.Signup, ctx context.Context) error
	GetSignup(userID string, ctx context.Context) (models.Signup, error)
}

type UserRepository interface {
	CreateUser(user models.User, ctx context.Context) error
	RemoveUser(userID string, ctx context.Context) error
	AddWalletToUser(userID string, wallet models.Wallet, ctx context.Context) error
	AddContactToUser(userID, contactID string, ctx context.Context) error
	RemoveContactFromUser(userID, contactID string, ctx context.Context) error
	GetUserByID(userID string, ctx context.Context) (models.User, error)
	GetUserByEmail(email string, ctx context.Context) (models.User, error)
	SetEnclaveURLAndVerificationKeyForUser(userID, enclaveURL, verificationKey string, ctx context.Context) error
}

type Transaction interface {
	Commit() error
	Rollback() error
	Users() UserRepository
	Signups() SignupRepository
	Notifications() NotificationRepository
}

type TransactionFactory interface {
	Begin() (Transaction, error)
}
