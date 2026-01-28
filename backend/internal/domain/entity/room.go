package entity

import (
	"time"
)

type Room struct {
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CreatorUserID string
}
