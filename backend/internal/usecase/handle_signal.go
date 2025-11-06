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

	client, err := messaging.NewClient(w, r, claims)
	if err != nil {
		log.Println("ws upgrade error:", err)
		return
	}

	s.connections.AddClient(client, claims.RoomID)

	read := make(chan []byte, messaging.BufferSize)
	done := make(chan struct{})

	go s.connections.Publisher(s.ctx, client, read, done)
	go client.WritePump(s.ctx, done)
	client.ReadPump(s.ctx, read)

	<-done // WritePump

	s.connections.RemoveClient(client, claims.RoomID)
}

func (s *SignalingUseCases) validateReq(w http.ResponseWriter, r *http.Request) (*auth.Claims, bool) {
	jwtStr := r.URL.Query().Get("jwt")
	if jwtStr == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return nil, false
	}

	token, claims, err := s.jwt.GetToken(jwtStr)
	if err != nil {
		log.Printf("error getting jwt %s", err.Error())
		http.Error(w, "error getting jwt", http.StatusUnauthorized)
		return nil, false
	}

	if !token.Valid {
		log.Printf("invalid jwt %s", jwtStr)
		http.Error(w, "invalid jwt", http.StatusUnauthorized)
		return nil, false
	}

	return claims, true
}
