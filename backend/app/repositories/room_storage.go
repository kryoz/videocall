package repositories

import (
	"context"
	"log"
	"sync"
	"time"
	"videocall/app/room"
	"videocall/app/signaling"
)

const roomCheckInterval = 60 * time.Second
const emptyRoomTtl = 15 * time.Minute

type RoomStorage struct {
	mu    sync.RWMutex
	Rooms map[string]*room.Room
}

func NewHub(ctx context.Context) *RoomStorage {
	rs := &RoomStorage{
		Rooms: make(map[string]*room.Room),
	}

	rs.handleEmptyRooms(ctx)

	return rs
}

func (rs *RoomStorage) handleEmptyRooms(ctx context.Context) {
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

func (rs *RoomStorage) AddRoom(roomID, password string) {
	rs.mu.Lock()
	rs.Rooms[roomID] = &room.Room{
		ID:        roomID,
		Password:  password,
		CreatedAt: time.Now(),
		Connections: room.Connections{
			WsClients: make(map[*signaling.Client]bool),
		},
	}
	rs.mu.Unlock()
}

func (rs *RoomStorage) GetRoom(roomID string) (*room.Room, bool) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	r, ok := rs.Rooms[roomID]

	return r, ok
}
