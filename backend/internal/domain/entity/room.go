package entity

import (
	"time"
)

type Room struct {
	Password  string
	CreatedAt time.Time
	Connections
}
