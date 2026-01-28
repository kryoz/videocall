package repositories

import (
	"context"
	"log"
	"time"
	"videocall/internal/infrastructure/config"
)

func HandleObsoleteRooms(ctx context.Context, rs RoomRepositoryInterface, conf config.RoomConfig) {
	go func() {
		ticker := time.NewTicker(conf.CleanInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Printf("dispatched room clean up task")
				rs.CleanRooms(conf.TTL)
			}
		}
	}()
}
