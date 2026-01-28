package token

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"time"
	"videocall/internal/domain/entity"
	"videocall/internal/domain/repositories"
)

type RefreshTokenService struct {
	repo repositories.RefreshTokenRepositoryInterface
	ttl  time.Duration
}

func NewRefreshTokenService(tokenRepository repositories.RefreshTokenRepositoryInterface, ttl time.Duration) *RefreshTokenService {
	return &RefreshTokenService{
		repo: tokenRepository,
		ttl:  ttl,
	}
}

func (r *RefreshTokenService) GenerateRefreshToken(userID string) (*entity.RefreshToken, error) {
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)

	token := base64.URLEncoding.EncodeToString(randomBytes)

	refreshToken := &entity.RefreshToken{
		Token:  token,
		UserID: userID,
		Expiry: time.Now().Add(r.ttl),
	}

	if err := r.repo.Create(refreshToken); err != nil {
		return nil, err
	}

	return refreshToken, nil
}

func (r *RefreshTokenService) Validate(token string) (*entity.RefreshToken, bool) {
	tok, err := r.repo.GetToken(token)
	if err != nil {
		log.Printf("token repo error: %v", err)
		return nil, false
	}

	return tok, tok.Expiry.After(time.Now())
}

func (r *RefreshTokenService) Revoke(token string) {
	r.repo.Remove(token)
}
