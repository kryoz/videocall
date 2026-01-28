package repositories

import (
	"errors"
	"time"

	"videocall/internal/domain/entity"
)

var (
	ErrTokenAlreadyExists = errors.New("token already exists")
	ErrTokenNotFound      = errors.New("token not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

type RoomRepositoryInterface interface {
	AddRoom(roomID, creatorUserID string)
	GetRoom(roomID string) (*entity.Room, bool)
	RefreshRoom(roomID string)
	CleanRooms(ts time.Duration)
}

type UserRepositoryInterface interface {
	CreateUser(user *entity.User) error
	GetUser(userID string) (*entity.User, error)
	GetUserByUsername(username string) (*entity.User, error)
	UpdatePushSubscription(userID string, sub *entity.PushSubscription) error
	RemovePushSubscription(userID string) error
}

type RefreshTokenRepositoryInterface interface {
	Create(token *entity.RefreshToken) error
	GetToken(token string) (*entity.RefreshToken, error)
	Remove(token string)
}
