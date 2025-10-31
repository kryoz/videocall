package usecase

import (
	"log"
	"net/http"
	"videocall/internal/infrastructure/auth"
	"videocall/internal/infrastructure/messaging"
)

func (s *SignalingUseCases) SignalHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.validateReq(w, r)
	if !ok {
		return
	}

	room, ok := s.roomRepository.GetRoom(claims.RoomID)
	if !ok {
		log.Printf("room not found %s", claims.RoomID)
		return
	}

	client, err := messaging.NewClient(w, r, claims)
	if err != nil {
		log.Println("ws upgrade error:", err)
		return
	}

	room.AddClient(client, claims.RoomID)

	read := make(chan []byte, messaging.BufferSize)
	done := make(chan struct{})

	go room.Publisher(s.ctx, client, read, done)
	go client.WritePump(s.ctx, done)
	client.ReadPump(s.ctx, read)

	<-done // WritePump

	room.RemoveClient(client, claims.RoomID)
}

func (s *SignalingUseCases) validateReq(w http.ResponseWriter, r *http.Request) (*auth.Claims, bool) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return nil, false
	}

	token, claims, err := s.jwt.GetToken(tokenStr)
	if err != nil {
		log.Printf("error validating jwt %s", err.Error())
		return nil, false
	}

	if !token.Valid {
		log.Printf("invalid token %s", tokenStr)
		return nil, false
	}

	return claims, true
}
