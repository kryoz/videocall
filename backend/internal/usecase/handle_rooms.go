package usecase

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type PasswordReq struct {
	Password string `json:"password"`
	User     string `json:"username"`
}

func (s *ApiUseCases) HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	var req PasswordReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	roomID := strings.Replace(uuid.NewString(), "-", "", -1)
	s.roomRepository.AddRoom(roomID, req.Password)

	token, err := s.jwt.Issue(req.User, roomID)
	if err != nil {
		log.Printf("failed to generate token: %v", err)
		http.Error(w, "cannot issue jwt", http.StatusInternalServerError)
		return
	}

	log.Printf("%s got jwt for room %s", req.User, roomID)

	writeJSON(w, map[string]string{
		"room_id":  roomID,
		"token":    token,
		"join_url": "/join/" + roomID,
	})
}

func (s *ApiUseCases) HandleJoinRoom(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "room not specified", http.StatusBadRequest)
		return
	}
	roomID := parts[3]

	var req PasswordReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.User == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	room, ok := s.roomRepository.GetRoom(roomID)
	if !ok {
		http.Error(w, fmt.Sprintf("room not found %s", roomID), http.StatusNotFound)
		return
	}

	if room.Password != "" && room.Password != req.Password {
		http.Error(w, "invalid password", http.StatusUnauthorized)
		return
	}

	if len(room.Connections.WsClients) > 1 {
		http.Error(w, "room already full", http.StatusNotAcceptable)
		return
	}

	token, err := s.jwt.Issue(req.User, roomID)
	if err != nil {
		http.Error(w, "cannot issue token", http.StatusInternalServerError)
		return
	}

	log.Printf("%s created room %s", req.User, roomID)

	writeJSON(w, map[string]string{
		"token": token,
	})
}

func (s *ApiUseCases) HandleFetchRoom(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "room not specified", http.StatusBadRequest)
		return
	}
	roomID := parts[3]

	room, ok := s.roomRepository.GetRoom(roomID)
	if !ok {
		http.Error(w, fmt.Sprintf("room not found %s", roomID), http.StatusNotFound)
		return
	}

	isProtected := "0"
	if room.Password != "" {
		isProtected = "1"
	}

	writeJSON(w, map[string]string{
		"is_protected": isProtected,
	})
}
