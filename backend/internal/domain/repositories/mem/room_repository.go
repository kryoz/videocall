package mem

import (
	"sync"
	"time"
	"videocall/internal/domain/entity"
)

type RoomRepository struct {
	mu    sync.RWMutex
	Rooms map[string]*entity.Room
}

func New() *RoomRepository {
	rs := &RoomRepository{
		Rooms: make(map[string]*entity.Room),
	}

	return rs
}

func (rs *RoomRepository) AddRoom(roomID, creatorUserID string) {
	rs.mu.Lock()
	rs.Rooms[roomID] = &entity.Room{
		CreatedAt:     time.Now(),
		CreatorUserID: creatorUserID,
	}
	rs.mu.Unlock()
}

func (rs *RoomRepository) GetRoom(roomID string) (*entity.Room, bool) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	r, ok := rs.Rooms[roomID]

	return r, ok
}

func (rs *RoomRepository) DeleteRoom(roomID string) {
	delete(rs.Rooms, roomID)
}
