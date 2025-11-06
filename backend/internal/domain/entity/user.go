package entity

import "time"

type User struct {
	ID               string
	Username         string
	Password         string // hashed (only for registered users)
	CreatedAt        time.Time
	IsGuest          bool
	PushSubscription *PushSubscription
}
