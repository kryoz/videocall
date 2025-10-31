package repositories

import (
	"context"
	"log"
	"sync"
	"time"
	"videocall/internal/domain/entity"
	"videocall/internal/infrastructure/messaging"
)

const roomCheckInterval = 60 * time.Second
const emptyRoomTtl = 15 * time.Minute

type RoomRepository struct {
	mu    sync.RWMutex
	Rooms map[string]*entity.Room
}

func New(ctx context.Context) *RoomRepository {
	rs := &RoomRepository{
		Rooms: make(map[string]*entity.Room),
	}

	rs.handleEmptyRooms(ctx)

	return rs
}

func (rs *RoomRepository) AddRoom(roomID, password string) {
	rs.mu.Lock()
	rs.Rooms[roomID] = &entity.Room{
		Password:  password,
		CreatedAt: time.Now(),
		Connections: entity.Connections{
			WsClients: make(map[*messaging.Client]bool),
		},
	}
	rs.mu.Unlock()
}

func (rs *RoomRepository) GetRoom(roomID string) (*entity.Room, bool) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	r, ok := rs.Rooms[roomID]

	return r, ok
}

func (rs *RoomRepository) handleEmptyRooms(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(roomCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rs.mu.Lock()
				for roomID := range rs.Rooms {
					if len(rs.Rooms[roomID].Connections.WsClients) == 0 &&
						(rs.Rooms[roomID].CreatedAt.Add(emptyRoomTtl).Before(time.Now())) {
						log.Printf("autoclean: delete empty room %s", roomID)
						delete(rs.Rooms, roomID)
					}
				}
				rs.mu.Unlock()
			}
		}
	}()
}
