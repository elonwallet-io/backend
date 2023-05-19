package repository

import (
	"context"
	"github.com/Leantar/elonwallet-backend/models"
	"github.com/Leantar/elonwallet-backend/server/common"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"time"
)

type NotificationRepository struct {
	tx *sqlx.Tx
}

func (n *NotificationRepository) CreateNotification(notification models.Notification, ctx context.Context) error {
	const query = `INSERT INTO notifications("id", "user_id", "creation_time", "send_after", "title", "body") VALUES($1,$2,$3,$4,$5,$6)`

	_, err := n.tx.ExecContext(ctx, query, notification.ID, notification.UserID, notification.CreationTime, notification.SendAfter, notification.Title, notification.Body)
	if e, ok := err.(*pq.Error); ok && e.Code == postgresUniqueViolationCode {
		err = common.ErrConflict
	}

	return err
}

func (n *NotificationRepository) GetPendingNotificationsBatch(ctx context.Context) ([]models.Notification, error) {
	const query = `SELECT * FROM notifications WHERE $1 > "send_after" ORDER BY "creation_time" ASC LIMIT 100`

	now := time.Now().Unix()
	notifications := make([]dbNotification, 0)

	err := n.tx.SelectContext(ctx, &notifications, query, now)
	if err != nil {
		return nil, err
	}

	if len(notifications) == 0 {
		return nil, common.ErrNotFound
	}

	output := make([]models.Notification, len(notifications))
	for i, notification := range notifications {
		output[i] = models.Notification(notification)
	}

	return output, nil
}

func (n *NotificationRepository) DeleteNotification(ID string, ctx context.Context) error {
	const query = `DELETE FROM notifications WHERE "id" = $1`

	_, err := n.tx.ExecContext(ctx, query, ID)
	return err
}
