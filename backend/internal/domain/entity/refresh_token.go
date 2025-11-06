package entity

import "time"

type RefreshToken struct {
	Token  string
	UserID string
	Expiry time.Time
}
