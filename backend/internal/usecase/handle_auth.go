package usecase

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"videocall/internal/domain/entity"
	"videocall/internal/infrastructure/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UsernameRequest struct {
	Username string `json:"username"`
}

type RegisterRequest struct {
	UsernameRequest
	Password string `json:"password"`
}

type LoginRequest struct {
	RegisterRequest
}

type GuestRequest struct {
	UsernameRequest
}

type RefreshTokenRequest struct {
	Token  string `json:"token"`
	RoomID string `json:"room_id,omitempty"`
}

type PushSubscriptionRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

func (s *ApiUseCases) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password required", http.StatusBadRequest)
		return
	}

	password := []byte(req.Password)
	if len(password) < 8 {
		http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	userID := uuid.NewString()
	user := &entity.User{
		ID:        userID,
		Username:  entity.UsernameNormalize(req.Username),
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		IsGuest:   false,
	}

	if err := s.userRepository.CreateUser(user); err != nil {
		log.Printf("failed to create user: %v", err)
		http.Error(w, "username already exists", http.StatusConflict)
		return
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		log.Printf("failed to create refresh token: %v", err)
		http.Error(w, "cannot create refresh token", http.StatusInternalServerError)
		return
	}

	jwtStr, _, err := s.jwt.Issue(user.ID, user.Username, "")
	if err != nil {
		log.Printf("failed to generate jwt: %v", err)
		http.Error(w, "cannot issue jwt", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ User registered: %s (ID: %s)", user.Username, user.ID)

	writeJSON(w, map[string]string{
		"token":    refreshToken.Token,
		"expires":  refreshToken.Expiry.Format(time.RFC3339),
		"jwt":      jwtStr,
		"user_id":  user.ID,
		"username": user.Username,
	})
}

func (s *ApiUseCases) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, err := s.userRepository.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	refreshToken, ok := s.tokenService.Validate(user.ID)
	if !ok {
		refreshToken, err = s.tokenService.GenerateRefreshToken(user.ID)
		if err != nil {
			log.Printf("failed to create refresh token: %v", err)
			http.Error(w, "cannot create refresh token", http.StatusInternalServerError)
			return
		}
	}

	jwtStr, _, err := s.jwt.Issue(user.ID, user.Username, "")
	if err != nil {
		log.Printf("failed to generate jwt: %v", err)
		http.Error(w, "cannot issue jwt", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ New jwt token issued by login: %s (ID: %s)", user.Username, user.ID)

	writeJSON(w, map[string]string{
		"token":    refreshToken.Token,
		"expires":  refreshToken.Expiry.Format(time.RFC3339),
		"jwt":      jwtStr,
		"user_id":  user.ID,
		"username": user.Username,
	})
}

func (s *ApiUseCases) HandleCreateGuest(w http.ResponseWriter, r *http.Request) {
	var req GuestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, "username required", http.StatusBadRequest)
		return
	}

	userID := uuid.NewString()
	user := &entity.User{
		ID:        userID,
		Username:  req.Username,
		Password:  uuid.NewString(),
		CreatedAt: time.Now(),
		IsGuest:   true,
	}

	if err := s.userRepository.CreateUser(user); err != nil {
		log.Printf("Failed to create guest: %v", err)
		http.Error(w, "failed to create guest", http.StatusInternalServerError)
		return
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		log.Printf("failed to create refresh token: %v", err)
		http.Error(w, "cannot create refresh token", http.StatusInternalServerError)
		return
	}

	jwtStr, _, err := s.jwt.Issue(user.ID, user.Username, "")
	if err != nil {
		log.Printf("failed to generate jwt: %v", err)
		http.Error(w, "cannot issue jwt", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Guest created: %s (ID: %s)", user.Username, user.ID)

	writeJSON(w, map[string]string{
		"user_id":  user.ID,
		"username": user.Username,
		"jwt":      jwtStr,
		"token":    refreshToken.Token,
		"expires":  refreshToken.Expiry.Format(time.RFC3339),
	})
}

func (s *ApiUseCases) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	tok, ok := s.tokenService.Validate(req.Token)
	if !ok {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	user, err := s.userRepository.GetUser(tok.UserID)
	if err != nil {
		http.Error(w, "unable to fetch user", http.StatusInternalServerError)
		return
	}

	jwtStr, _, err := s.jwt.Issue(tok.UserID, user.Username, req.RoomID)
	if err != nil {
		log.Printf("failed to generate jwt: %v", err)
		http.Error(w, "cannot issue jwt", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ New jwt token issued by refresh token: %s (room: %s)", user.Username, req.RoomID)

	writeJSON(w, map[string]string{
		"jwt":      jwtStr,
		"user_id":  user.ID,
		"username": user.Username,
	})
}

func (s *ApiUseCases) HandleRevokeToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	s.tokenService.Revoke(req.Token)

	log.Printf("✅ Revoke token: %s", req.Token)
}

func (s *ApiUseCases) HandleSubscribePush(w http.ResponseWriter, r *http.Request) {
	token, claims, err := s.validateAuthHeader(r)
	if err != nil || !token.Valid {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req PushSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Save subscription
	subscription := &entity.PushSubscription{
		Endpoint: req.Endpoint,
		Keys: entity.PushKeys{
			P256dh: req.Keys.P256dh,
			Auth:   req.Keys.Auth,
		},
	}

	if err := s.userRepository.UpdatePushSubscription(claims.UserID, subscription); err != nil {
		log.Printf("Failed to update push subscription: %v", err)
		http.Error(w, "failed to save subscription", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Push subscription saved for user: %s (ID: %s)", claims.Username, claims.UserID)

	writeJSON(w, map[string]string{
		"status": "subscribed",
	})
}

func (s *ApiUseCases) HandleUnsubscribePush(w http.ResponseWriter, r *http.Request) {
	token, claims, err := s.validateAuthHeader(r)
	if err != nil || !token.Valid {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := s.userRepository.RemovePushSubscription(claims.UserID); err != nil {
		log.Printf("Failed to remove push subscription: %v", err)
		http.Error(w, "failed to remove subscription", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Push subscription removed for user: %s (ID: %s)", claims.Username, claims.UserID)

	writeJSON(w, map[string]string{
		"status": "unsubscribed",
	})
}

func (s *ApiUseCases) HandleGetVapidPublicKey(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]string{
		"publicKey": s.pushService.GetPublicKey(),
	})
}

func (s *ApiUseCases) validateAuthHeader(r *http.Request) (*jwt.Token, *auth.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, nil, fmt.Errorf("missing bearer token")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	return s.jwt.GetToken(tokenStr)
}
