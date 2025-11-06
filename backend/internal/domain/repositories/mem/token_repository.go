package mem

import (
	"context"
	"log"
	"sync"
	"time"
	"videocall/internal/domain/entity"
	"videocall/internal/domain/repositories"
)

const tokenCleanInterval = 60 * time.Second

type RefreshTokenRepository struct {
	mu     sync.RWMutex
	Tokens map[string]*entity.RefreshToken
}

func NewRefreshTokenRepository(ctx context.Context) *RefreshTokenRepository {
	tr := &RefreshTokenRepository{
		Tokens: make(map[string]*entity.RefreshToken),
	}

	tr.handleExpiredTokens(ctx)

	return tr
}

func (t *RefreshTokenRepository) Create(token *entity.RefreshToken) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	tok, exists := t.Tokens[token.Token]
	if exists && tok.Expiry.Before(time.Now()) {
		return repositories.ErrTokenAlreadyExists
	}

	t.Tokens[token.Token] = token

	return nil
}

func (t *RefreshTokenRepository) GetToken(token string) (*entity.RefreshToken, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	tok, ok := t.Tokens[token]
	if !ok {
		return nil, repositories.ErrTokenNotFound
	}

	return tok, nil
}

func (t *RefreshTokenRepository) Remove(token string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.Tokens, token)
}

func (t *RefreshTokenRepository) handleExpiredTokens(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(tokenCleanInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				t.mu.Lock()
				for id, tok := range t.Tokens {
					if tok.Expiry.Before(time.Now()) {
						log.Printf("autoclean: delete expired refresh token for user %s", tok.UserID)
						delete(t.Tokens, id)
					}
				}
				t.mu.Unlock()
			}
		}
	}()
}
