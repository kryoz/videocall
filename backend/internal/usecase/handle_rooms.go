package usecase

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// MaxRoomUsers Only 2 users in a room tested so far
const MaxRoomUsers = 2

type InviteRequest struct {
	InvitedUsername string `json:"invited_username"`
}

func (s *ApiUseCases) HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	oldJwt, claims, err := s.validateAuthHeader(r)
	if err != nil || !oldJwt.Valid {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	roomID := strings.Replace(uuid.NewString(), "-", "", -1)
	s.roomRepository.AddRoom(roomID, claims.UserID)

	//refresh token to add roomID
	jwtStr, _, err := s.jwt.Issue(claims.UserID, claims.Username, roomID)
	if err != nil {
		log.Printf("failed to generate token: %v", err)
		http.Error(w, "cannot issue jwt", http.StatusInternalServerError)
		return
	}

	log.Printf("User %s (%s) created room %s", claims.Username, claims.UserID, roomID)

	writeJSON(w, map[string]string{
		"room_id":  roomID,
		"jwt":      jwtStr,
		"join_url": "/join/" + roomID,
	})
}

func (s *ApiUseCases) HandleJoinRoom(w http.ResponseWriter, r *http.Request) {
	jwtTok, claims, err := s.validateAuthHeader(r)
	if err != nil || !jwtTok.Valid {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

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

	if len(s.connections.WsClients) > MaxRoomUsers-1 {
		http.Error(w, "room already full", http.StatusNotAcceptable)
		return
	}

	// Обновляем jwt, чтобы он стал содержать RoomID
	jwtStr, _, err := s.jwt.Issue(claims.UserID, claims.Username, roomID)
	if err != nil {
		log.Printf("failed to generate token: %v", err)
		http.Error(w, "cannot issue jwt", http.StatusInternalServerError)
		return
	}

	// Обновляем метку времени комнаты, чтобы продлить время её жизни
	s.roomRepository.RefreshRoom(roomID)

	// Send notification to room creator if he is absent but has push enabled
	if s.pushService != nil && room.CreatorUserID != claims.UserID {
		var notifyUsers []string
		if len(s.connections.WsClients) > 0 {
			for userID := range s.connections.WsClients {
				if userID != claims.UserID {
					notifyUsers = append(notifyUsers, userID)
				}
			}
		} else {
			notifyUsers = append(notifyUsers, room.CreatorUserID)
		}

		if len(notifyUsers) > 0 {
			go func(users []string) {
				for _, user := range users {
					creator, err := s.userRepository.GetUser(user)
					if err == nil && creator.PushSubscription != nil {
						if err := s.pushService.NotifyUserJoined(user, claims.Username, roomID); err != nil {
							log.Printf("Failed to send join notification: %v", err)
						}
					}
				}
			}(notifyUsers)
		}
	}

	log.Printf("User %s (%s) joined room %s", claims.Username, claims.UserID, roomID)

	writeJSON(w, map[string]string{
		"jwt": jwtStr,
	})
}

func (s *ApiUseCases) HandleInviteToRoom(w http.ResponseWriter, r *http.Request) {
	token, claims, err := s.validateAuthHeader(r)
	if err != nil || !token.Valid {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "room not specified", http.StatusBadRequest)
		return
	}
	roomID := parts[3]

	var req InviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.InvitedUsername == "" {
		http.Error(w, "invited_username required", http.StatusBadRequest)
		return
	}

	_, ok := s.roomRepository.GetRoom(roomID)
	if !ok {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	invitedUser, err := s.userRepository.GetUserByUsername(strings.Trim(req.InvitedUsername, " "))
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if invitedUser.PushSubscription == nil {
		http.Error(w, "user has no push subscription", http.StatusBadRequest)
		return
	}

	// Send notification
	if s.pushService != nil {
		if err := s.pushService.NotifyRoomInvite(claims.UserID, claims.Username, invitedUser.ID, roomID); err != nil {
			log.Printf("Failed to send invite notification: %v", err)
			http.Error(w, "failed to send notification", http.StatusInternalServerError)
			return
		}
	}

	log.Printf("✅ Invite sent from %s to %s for room %s", claims.Username, req.InvitedUsername, roomID)

	writeJSON(w, map[string]string{
		"status": "invited",
	})
}

func (s *ApiUseCases) HandleFetchRoom(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "room not specified", http.StatusBadRequest)
		return
	}
	roomID := parts[3]

	_, ok := s.roomRepository.GetRoom(roomID)
	if !ok {
		http.Error(w, fmt.Sprintf("room not found %s", roomID), http.StatusNotFound)
		return
	}

	writeJSON(w, map[string]string{
		"exists": "true",
	})
}
