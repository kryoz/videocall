package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"videocall/internal/domain/entity"
	"videocall/internal/domain/repositories"
)

type MariaDBRefreshTokenRepository struct {
	db *sql.DB
}

func NewMariaDBRefreshTokenRepository(db *sql.DB) *MariaDBRefreshTokenRepository {
	repo := &MariaDBRefreshTokenRepository{db: db}
	repo.handleExpiredTokens(context.Background())
	return repo
}

func (r *MariaDBRefreshTokenRepository) Create(token *entity.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (token, user_id, expiry)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE expiry = VALUES(expiry)
	`
	_, err := r.db.Exec(query, token.Token, token.UserID, token.Expiry)
	if err != nil {
		if isDuplicateKeyError(err) {
			return repositories.ErrTokenAlreadyExists
		}
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *MariaDBRefreshTokenRepository) GetToken(token string) (*entity.RefreshToken, error) {
	query := `
		SELECT token, user_id, expiry
		FROM refresh_tokens
		WHERE token = ?
	`
	var tokenStr, userID string
	var expiry time.Time

	err := r.db.QueryRow(query, token).Scan(&tokenStr, &userID, &expiry)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &entity.RefreshToken{
		Token:  tokenStr,
		UserID: userID,
		Expiry: expiry,
	}, nil
}

func (r *MariaDBRefreshTokenRepository) Remove(token string) {
	query := `
		DELETE FROM refresh_tokens
		WHERE token = ?
	`
	_, err := r.db.Exec(query, token)
	if err != nil {
		log.Printf("Error removing refresh token: %v", err)
	}
}

func (r *MariaDBRefreshTokenRepository) handleExpiredTokens(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := r.cleanupExpiredTokens()
				if err != nil {
					log.Printf("Error cleaning up expired tokens: %v", err)
				}
			}
		}
	}()
}

func (r *MariaDBRefreshTokenRepository) cleanupExpiredTokens() error {
	query := `
		DELETE FROM refresh_tokens
		WHERE expiry < NOW()
	`
	_, err := r.db.Exec(query)
	return err
}
