package models

type Notification struct {
	ID           string `json:"id"`
	CreationTime int64  `json:"creation_time"`
	SendAfter    int64  `json:"send_after"`
	UserID       string `json:"user_id"`
	Title        string `json:"title"`
	Body         string `json:"body"`
}
