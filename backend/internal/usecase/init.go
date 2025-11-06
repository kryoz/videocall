package usecase

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"videocall/internal/domain/repositories"
	"videocall/internal/infrastructure/auth"
	"videocall/internal/infrastructure/config"
	"videocall/internal/infrastructure/push"
	"videocall/internal/infrastructure/token"
)

type ApiUseCases struct {
	ctx            context.Context
	roomRepository repositories.RoomRepositoryInterface
	userRepository repositories.UserRepositoryInterface
	cfg            *config.Config
	jwt            *auth.JWT
	tokenService   *token.RefreshTokenService
	pushService    *push.Service
	connections    *repositories.Connections
}

type SignalingUseCases struct {
	ctx         context.Context
	connections *repositories.Connections
	jwt         *auth.JWT
	pushService *push.Service
}

func NewApiUseCases(ctx context.Context, roomRepo repositories.RoomRepositoryInterface, userRepo repositories.UserRepositoryInterface, cfg *config.Config, jwt *auth.JWT, refreshTokenService *token.RefreshTokenService, pushService *push.Service, connections *repositories.Connections) *ApiUseCases {
	return &ApiUseCases{
		ctx:            ctx,
		roomRepository: roomRepo,
		userRepository: userRepo,
		cfg:            cfg,
		jwt:            jwt,
		tokenService:   refreshTokenService,
		pushService:    pushService,
		connections:    connections,
	}
}

func NewSignalingUseCases(ctx context.Context, connections *repositories.Connections, jwt *auth.JWT, pushService *push.Service) *SignalingUseCases {
	return &SignalingUseCases{
		ctx:         ctx,
		connections: connections,
		jwt:         jwt,
		pushService: pushService,
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Printf("writeJSON error: %v", err)
		return
	}
}
