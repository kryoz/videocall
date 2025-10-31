package app

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"videocall/app/config"
)

func Run(ctx context.Context, cfg *config.Config) error {
	hub := NewHub()
	jwt := NewJWT(cfg)

	log.SetOutput(os.Stderr)
	slog.SetLogLoggerLevel(slog.LevelDebug)

	http.HandleFunc("/api/rooms", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		hub.HandleCreateRoom(w, r, jwt)
	})

	http.HandleFunc("/api/rooms/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if !strings.HasSuffix(r.URL.Path, "/join") {
			http.Error(w, "method is not supported yet", http.StatusMethodNotAllowed)
			return
		}
		hub.HandleJoinRoom(w, r, jwt)
	})

	http.HandleFunc("/api/turn", func(w http.ResponseWriter, r *http.Request) {
		handleTurn(w, r, cfg, jwt)
	})
	http.HandleFunc("/api/signal", SignalHandler(ctx, hub, jwt))

	log.Printf("CONFIG: %v", cfg)
	log.Println("Server listening on", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, nil); err != nil {
		log.Fatal(err)
	}

	return nil
}
