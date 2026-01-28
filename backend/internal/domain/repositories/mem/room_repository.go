package mem

import (
	"log"
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
		UpdatedAt:     time.Now(),
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

func (rs *RoomRepository) RefreshRoom(roomID string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	_, ok := rs.Rooms[roomID]
	if !ok {
		return
	}

	rs.Rooms[roomID].UpdatedAt = time.Now()
}

func (rs *RoomRepository) CleanRooms(ttl time.Duration) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	for roomID, room := range rs.Rooms {
		if room.UpdatedAt.Add(ttl).Before(time.Now()) {
			delete(rs.Rooms, roomID)
			log.Printf("autoclean: delete empty room %s", roomID)
		}
	}
}
