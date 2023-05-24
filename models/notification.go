package models

type Notification struct {
	ID           int64  `json:"id"`
	SeriesID     string `json:"series_id"`
	CreationTime int64  `json:"creation_time"`
	SendAfter    int64  `json:"send_after"`
	TimesTried   int64  `json:"times_tried"`
	UserID       string `json:"user_id"`
	Title        string `json:"title"`
	Body         string `json:"body"`
}
