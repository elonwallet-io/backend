package repository

import (
	"context"
	"github.com/Leantar/elonwallet-backend/models"
	"github.com/Leantar/elonwallet-backend/server/common"
	"github.com/jmoiron/sqlx"
	"math"
	"time"
)

type NotificationRepository struct {
	tx *sqlx.Tx
}

const (
	step = 100
)

func (n *NotificationRepository) CreateNotificationSeries(notifications []models.Notification, ctx context.Context) (err error) {
	const query = `INSERT INTO notifications("series_id", "creation_time", "send_after", "times_tried", "user_id", "title", "body") VALUES(:series_id, :creation_time, :send_after, :times_tried, :user_id, :title, :body)`

	dbNotifications := make([]dbNotification, len(notifications))
	for i, nf := range notifications {
		dbNotifications[i] = dbNotification(nf)
	}

	notificationsLen := len(dbNotifications)
	lowerBound := 0
	upperBound := int(math.Min(float64(step), float64(notificationsLen)))

	for lowerBound < notificationsLen {
		_, err = n.tx.NamedExecContext(ctx, query, dbNotifications[lowerBound:upperBound])
		if err != nil {
			return
		}

		lowerBound = upperBound
		upperBound = int(math.Min(float64(upperBound+step), float64(notificationsLen)))
	}

	return
}

func (n *NotificationRepository) GetPendingNotificationsBatch(ctx context.Context) ([]models.Notification, error) {
	const query = `SELECT * FROM notifications WHERE $1 > "send_after" AND "times_tried" < 3 ORDER BY "creation_time" ASC LIMIT 100`

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

func (n *NotificationRepository) UpdateNotification(notification models.Notification, ctx context.Context) (err error) {
	const query = `UPDATE notifications SET "series_id" = $1, "creation_time" = $2, "send_after" = $3, "times_tried" = $4, "user_id" = $5, "title" = $6, "body" = $7 WHERE "id" = $8`

	_, err = n.tx.ExecContext(ctx, query, notification.SeriesID, notification.CreationTime, notification.SendAfter, notification.TimesTried, notification.UserID, notification.Title, notification.Body, notification.ID)
	return
}

func (n *NotificationRepository) DeleteNotification(id int64, userID string, ctx context.Context) (err error) {
	const query = `DELETE FROM notifications WHERE "id" = $1 AND "user_id" = $2`

	_, err = n.tx.ExecContext(ctx, query, id, userID)
	return
}

func (n *NotificationRepository) DeleteNotificationSeries(seriesID string, userID string, ctx context.Context) (err error) {
	const query = `DELETE FROM notifications WHERE "series_id" = $1 AND "user_id" = $2`

	_, err = n.tx.ExecContext(ctx, query, seriesID, userID)
	return
}
