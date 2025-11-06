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
	"videocall/internal/infrastructure/database"
	"videocall/internal/infrastructure/push"
	"videocall/internal/infrastructure/token"
	restApi "videocall/internal/transport/http"
	wsApi "videocall/internal/transport/ws"
	"videocall/internal/usecase"
)

func Run(ctx context.Context, cfg *config.Config) error {
	storageFactory, err := database.NewStorageFactory(&cfg.Storage)
	if err != nil {
		return err
	}
	defer storageFactory.Close()

	roomRepo := storageFactory.CreateRoomRepository(ctx)
	userRepo := storageFactory.CreateUserRepository()
	tokenRepo := storageFactory.CreateRefreshTokenRepository(ctx)

	jwt := auth.NewJWT(cfg)
	refreshTokenService := token.NewRefreshTokenService(tokenRepo, cfg.RefreshToken.TTL)

	var pushService *push.Service
	if cfg.VAPID.PublicKey != "" && cfg.VAPID.PrivateKey != "" {
		pushService = push.NewService(cfg.VAPID.PublicKey, cfg.VAPID.PrivateKey, userRepo)
		log.Println("✅ Push notification service initialized")
	} else {
		log.Println("⚠️ VAPID keys not configured, push notifications disabled")
	}

	wsConns := repositories.NewConnections()

	apiUseCases := usecase.NewApiUseCases(ctx, roomRepo, userRepo, cfg, jwt, refreshTokenService, pushService, wsConns)
	signalingUseCases := usecase.NewSignalingUseCases(ctx, wsConns, jwt, pushService)

	httpService := restApi.NewAPI(apiUseCases)
	httpService.RegisterHandlers()

	wsService := wsApi.NewAPI(signalingUseCases)
	wsService.RegisterHandlers()

	log.SetOutput(os.Stderr)
	slog.SetLogLoggerLevel(slog.LevelDebug)

	log.Println("Server listening on", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, nil); err != nil {
		log.Fatal(err)
	}

	return nil
}
