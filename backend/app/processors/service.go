package processors

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"videocall/app/auth"
	"videocall/app/config"
	"videocall/app/repositories"
)

type Service struct {
	ctx context.Context
	hub *repositories.RoomStorage
	cfg *config.Config
	jwt *auth.JWT
}

func New(ctx context.Context, hub *repositories.RoomStorage, cfg *config.Config, jwt *auth.JWT) *Service {
	return &Service{ctx: ctx, hub: hub, cfg: cfg, jwt: jwt}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Printf("writeJSON error: %v", err)
		return
	}
}
