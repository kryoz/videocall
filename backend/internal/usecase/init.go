package usecase

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"videocall/internal/domain/repositories"
	"videocall/internal/infrastructure/auth"
	"videocall/internal/infrastructure/config"
)

type ApiUseCases struct {
	ctx            context.Context
	roomRepository *repositories.RoomRepository
	cfg            *config.Config
	jwt            *auth.JWT
}

type SignalingUseCases struct {
	ctx            context.Context
	roomRepository *repositories.RoomRepository
	jwt            *auth.JWT
}

func NewApiUseCases(ctx context.Context, hub *repositories.RoomRepository, cfg *config.Config, jwt *auth.JWT) *ApiUseCases {
	return &ApiUseCases{ctx: ctx, roomRepository: hub, cfg: cfg, jwt: jwt}
}

func NewSignalingUseCases(ctx context.Context, hub *repositories.RoomRepository, jwt *auth.JWT) *SignalingUseCases {
	return &SignalingUseCases{ctx: ctx, roomRepository: hub, jwt: jwt}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Printf("writeJSON error: %v", err)
		return
	}
}
