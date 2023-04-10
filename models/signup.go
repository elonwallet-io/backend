package models

type Signup struct {
	UserID           string `json:"user_id"`
	Activated        bool   `json:"activated"`
	ActivationString string `json:"activation_string"`
	Created          int64  `json:"created"`
	ValidUntil       int64  `json:"valid_until"`
}
