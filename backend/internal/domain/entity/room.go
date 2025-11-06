package entity

import (
	"time"
)

type Room struct {
	CreatedAt     time.Time
	CreatorUserID string
}
