package app

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"videocall/app/api"
	"videocall/app/auth"
	"videocall/app/config"
	"videocall/app/processors"
	"videocall/app/repositories"
)

func Run(ctx context.Context, cfg *config.Config) error {
	hub := repositories.NewHub(ctx)
	jwt := auth.NewJWT(cfg)
	processor := processors.New(ctx, hub, cfg, jwt)
	service := api.NewAPI(processor)
	service.RegisterHandlers()

	log.SetOutput(os.Stderr)
	slog.SetLogLoggerLevel(slog.LevelDebug)

	//log.Printf("CONFIG: %v", cfg)

	log.Println("Server listening on", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, nil); err != nil {
		log.Fatal(err)
	}

	return nil
}
