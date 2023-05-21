package repository

type dbUser struct {
	ID              string `db:"id"`
	Name            string `db:"name"`
	Email           string `db:"email"`
	EnclaveURL      string `db:"enclave_url"`
	VerificationKey string `db:"verification_key"`
}

type dbWallet struct {
	Address string `db:"address"`
	Name    string `db:"name"`
	UserID  string `db:"user_id"`
}

type dbContact struct {
	UserID    string `db:"user_id"`
	ContactID string `db:"contact_id"`
}

type dbSignup struct {
	UserID           string `db:"user_id"`
	Activated        bool   `db:"activated"`
	ActivationString string `db:"activation_string"`
	Created          int64  `db:"created"`
	ValidUntil       int64  `db:"valid_until"`
}

type dbNotification struct {
	ID           int64  `db:"id"`
	SeriesID     string `db:"series_id"`
	CreationTime int64  `db:"creation_time"`
	SendAfter    int64  `db:"send_after"`
	TimesTried   int64  `db:"times_tried"`
	UserID       string `db:"user_id"`
	Title        string `db:"title"`
	Body         string `db:"body"`
}
