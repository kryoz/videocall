package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"videocall/internal/domain/entity"
	"videocall/internal/domain/repositories"
)

type MariaDBUserRepository struct {
	db *sql.DB
}

func NewMariaDBUserRepository(db *sql.DB) *MariaDBUserRepository {
	return &MariaDBUserRepository{db: db}
}

func (r *MariaDBUserRepository) createTable() {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(255) PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255),
			push_subscription TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)
	`
	_, err := r.db.Exec(query)
	if err != nil {
		log.Printf("Error creating users table: %v", err)
	}
}

func (r *MariaDBUserRepository) CreateUser(user *entity.User) error {
	var pushSubJSON []byte
	var err error

	if user.PushSubscription != nil {
		pushSubJSON, err = json.Marshal(user.PushSubscription)
		if err != nil {
			return fmt.Errorf("failed to marshal push subscription: %w", err)
		}
	}

	query := `
		INSERT INTO users (id, username, password, push_subscription)
		VALUES (?, ?, ?, ?)
	`
	_, err = r.db.Exec(query, user.ID, user.Username, user.Password, pushSubJSON)
	if err != nil {
		if isDuplicateKeyError(err) {
			return repositories.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *MariaDBUserRepository) GetUser(userID string) (*entity.User, error) {
	query := `
		SELECT id, username, password, push_subscription
		FROM users
		WHERE id = ?
	`
	var id, username, password sql.NullString
	var pushSubJSON []byte

	err := r.db.QueryRow(query, userID).Scan(&id, &username, &password, &pushSubJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repositories.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user := &entity.User{
		ID:       id.String,
		Username: username.String,
		Password: password.String,
	}

	if len(pushSubJSON) > 0 {
		var pushSub entity.PushSubscription
		err = json.Unmarshal(pushSubJSON, &pushSub)
		if err != nil {
			log.Printf("Error unmarshaling push subscription: %v", err)
		} else {
			user.PushSubscription = &pushSub
		}
	}

	return user, nil
}

func (r *MariaDBUserRepository) GetUserByUsername(username string) (*entity.User, error) {
	query := `
		SELECT id, username, password, push_subscription
		FROM users
		WHERE username = ?
	`
	var id, usernameDB, password sql.NullString
	var pushSubJSON []byte

	err := r.db.QueryRow(query, username).Scan(&id, &usernameDB, &password, &pushSubJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repositories.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	user := &entity.User{
		ID:       id.String,
		Username: usernameDB.String,
		Password: password.String,
	}

	if len(pushSubJSON) > 0 {
		var pushSub entity.PushSubscription
		err = json.Unmarshal(pushSubJSON, &pushSub)
		if err != nil {
			log.Printf("Error unmarshaling push subscription: %v", err)
		} else {
			user.PushSubscription = &pushSub
		}
	}

	return user, nil
}

func (r *MariaDBUserRepository) UpdatePushSubscription(userID string, sub *entity.PushSubscription) error {
	subJSON, err := json.Marshal(sub)
	if err != nil {
		return fmt.Errorf("failed to marshal push subscription: %w", err)
	}

	query := `
		UPDATE users
		SET push_subscription = ?, updated_at = NOW()
		WHERE id = ? AND DATE_ADD(updated_at, INTERVAL 1 DAY) > NOW()
	`
	result, err := r.db.Exec(query, subJSON, userID)
	if err != nil {
		return fmt.Errorf("failed to update push subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repositories.ErrUserNotFound
	}

	return nil
}

func (r *MariaDBUserRepository) RemovePushSubscription(userID string) error {
	query := `
		UPDATE users
		SET push_subscription = NULL, updated_at = NOW()
		WHERE id = ?
	`
	result, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to remove push subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repositories.ErrUserNotFound
	}

	return nil
}

func isDuplicateKeyError(err error) bool {
	// MariaDB/MySQL duplicate key error code
	if err != nil {
		errStr := err.Error()
		return errStr != "" && (errStr == "1062" || errStr == "ER_DUP_ENTRY")
	}
	return false
}
