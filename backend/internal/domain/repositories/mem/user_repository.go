package mem

import (
	"sync"
	"videocall/internal/domain/entity"
	"videocall/internal/domain/repositories"
)

type UserRepository struct {
	mu            sync.RWMutex
	Users         map[string]*entity.User // key: user ID (UUID)
	UsernameIndex map[string]string       // username -> user ID for lookups
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		Users:         make(map[string]*entity.User),
		UsernameIndex: make(map[string]string),
	}
}

func (ur *UserRepository) CreateUser(user *entity.User) error {
	ur.mu.Lock()
	defer ur.mu.Unlock()

	// Check if username already exists (for registered users)
	if user.Password != "" {
		if _, exists := ur.UsernameIndex[user.Username]; exists {
			return repositories.ErrUserAlreadyExists
		}
		ur.UsernameIndex[user.Username] = user.ID
	}

	ur.Users[user.ID] = user
	return nil
}

func (ur *UserRepository) GetUser(userID string) (*entity.User, error) {
	ur.mu.RLock()
	defer ur.mu.RUnlock()

	user, ok := ur.Users[userID]
	if !ok {
		return nil, repositories.ErrUserNotFound
	}
	return user, nil
}

func (ur *UserRepository) GetUserByUsername(username string) (*entity.User, error) {
	ur.mu.RLock()
	defer ur.mu.RUnlock()

	userID, ok := ur.UsernameIndex[username]
	if !ok {
		return nil, repositories.ErrUserNotFound
	}

	user, ok := ur.Users[userID]
	if !ok {
		return nil, repositories.ErrUserNotFound
	}
	return user, nil
}

func (ur *UserRepository) UpdatePushSubscription(userID string, sub *entity.PushSubscription) error {
	ur.mu.Lock()
	defer ur.mu.Unlock()

	user, ok := ur.Users[userID]
	if !ok {
		return repositories.ErrUserNotFound
	}

	user.PushSubscription = sub
	return nil
}

func (ur *UserRepository) RemovePushSubscription(userID string) error {
	ur.mu.Lock()
	defer ur.mu.Unlock()

	user, ok := ur.Users[userID]
	if !ok {
		return repositories.ErrUserNotFound
	}

	user.PushSubscription = nil
	return nil
}
