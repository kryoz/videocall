package internal

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"videocall/internal/domain/repositories"
	"videocall/internal/infrastructure/auth"
	"videocall/internal/infrastructure/config"
	restApi "videocall/internal/transport/http"
	wsApi "videocall/internal/transport/ws"
	"videocall/internal/usecase"
)

func Run(ctx context.Context, cfg *config.Config) error {
	hub := repositories.New(ctx)
	jwt := auth.NewJWT(cfg)

	apiUseCases := usecase.NewApiUseCases(ctx, hub, cfg, jwt)
	signalingUseCases := usecase.NewSignalingUseCases(ctx, hub, jwt)

	httpService := restApi.NewAPI(apiUseCases)
	httpService.RegisterHandlers()

	wsService := wsApi.NewAPI(signalingUseCases)
	wsService.RegisterHandlers()

	log.SetOutput(os.Stderr)
	slog.SetLogLoggerLevel(slog.LevelDebug)

	//log.Printf("CONFIG: %v", cfg)

	log.Println("Server listening on", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, nil); err != nil {
		log.Fatal(err)
	}

	return nil
}
