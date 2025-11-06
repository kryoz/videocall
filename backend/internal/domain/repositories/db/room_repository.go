package db

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"videocall/internal/domain/entity"
)

type MariaDBRoomRepository struct {
	db *sql.DB
}

func NewMariaDBRoomRepository(db *sql.DB) *MariaDBRoomRepository {
	repo := &MariaDBRoomRepository{db: db}
	return repo
}

func (r *MariaDBRoomRepository) AddRoom(roomID, creatorUserID string) {
	query := `
		INSERT INTO rooms (id, creator_user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE updated_at = VALUES(updated_at)
	`
	_, err := r.db.Exec(query, roomID, creatorUserID, time.Now(), time.Now())
	if err != nil {
		log.Printf("error adding room: %v", err)
	}
}

func (r *MariaDBRoomRepository) GetRoom(roomID string) (*entity.Room, bool) {
	query := `
		SELECT id, creator_user_id, created_at
		FROM rooms
		WHERE id = ?
	`
	var roomIDDB, creatorUserID string
	var createdAt time.Time

	err := r.db.QueryRow(query, roomID).Scan(&roomIDDB, &creatorUserID, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false
		}
		log.Printf("error getting room: %v", err)
		return nil, false
	}

	return &entity.Room{
		CreatedAt:     createdAt,
		CreatorUserID: creatorUserID,
	}, true
}

func (r *MariaDBRoomRepository) DeleteRoom(roomID string) {
	query := `
		DELETE FROM rooms
		WHERE id = ?
	`
	_, err := r.db.Exec(query, roomID)
	if err != nil {
		log.Printf("error deleting room: %v", err)
	}
}
