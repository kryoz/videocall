package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID        string
	Name      string
	Password  string
	CreatedAt time.Time
	RoomSockets
}

type RoomSockets struct {
	wsClients map[*Client]bool
	mu        sync.RWMutex
}

type RoomStorage struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewHub() *RoomStorage {
	return &RoomStorage{
		rooms: make(map[string]*Room),
	}
}

func (rs *RoomStorage) HandleCreateRoom(w http.ResponseWriter, r *http.Request, jwt *JWT) {
	type Req struct {
		User     string `json:"username"`
		Password string `json:"password"`
	}
	var req Req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	rs.mu.Lock()
	roomID := strings.Replace(uuid.NewString(), "-", "", -1)

	rs.rooms[roomID] = &Room{
		ID:        roomID,
		Password:  req.Password,
		CreatedAt: time.Now(),
		RoomSockets: RoomSockets{
			wsClients: make(map[*Client]bool),
		},
	}
	rs.mu.Unlock()

	token, err := jwt.Issue(req.User, roomID)
	if err != nil {
		http.Error(w, "cannot issue token", http.StatusInternalServerError)
		return
	}

	log.Printf("issued token for %s in room %s", req.User, roomID)

	writeJSON(w, map[string]string{
		"room_id":  roomID,
		"token":    token,
		"join_url": "/join/" + roomID,
	})
}

func (rs *RoomStorage) HandleJoinRoom(w http.ResponseWriter, r *http.Request, jwt *JWT) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "room not specified", http.StatusBadRequest)
		return
	}
	roomID := parts[3]

	type Req struct {
		Password string `json:"password"`
		User     string `json:"username"`
	}
	var req Req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	rs.mu.RLock()
	room, ok := rs.rooms[roomID]
	rs.mu.RUnlock()
	if !ok {
		http.Error(w, fmt.Sprintf("room not found %s", roomID), http.StatusNotFound)
		return
	}
	if room.Password != "" && room.Password != req.Password {
		http.Error(w, "invalid password", http.StatusUnauthorized)
		return
	}

	if len(room.RoomSockets.wsClients) > 1 {
		http.Error(w, "room already full", http.StatusNotAcceptable)
		return
	}

	token, err := jwt.Issue(req.User, roomID)
	if err != nil {
		http.Error(w, "cannot issue token", http.StatusInternalServerError)
		return
	}

	log.Printf("issued token for %s in room %s", req.User, roomID)

	writeJSON(w, map[string]string{
		"token": token,
	})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
