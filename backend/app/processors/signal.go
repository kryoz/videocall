package processors

import (
	"log"
	"net/http"
	"videocall/app/signaling"
)

func (s *Service) SignalHandler(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	token, claims, err := s.jwt.GetToken(tokenStr)
	if err != nil {
		log.Printf("error validating jwt %s", err.Error())
		return
	}

	if !token.Valid {
		log.Printf("invalid token %s", tokenStr)
		return
	}

	room, ok := s.hub.GetRoom(claims.Room)
	if !ok {
		log.Printf("room not found %s", claims.Room)
		return
	}

	client, err := signaling.NewClient(w, r, claims)
	if err != nil {
		log.Println("ws upgrade error:", err)
		return
	}

	room.AddClient(client)

	read := make(chan []byte, signaling.BufferSize)
	done := make(chan struct{})

	go room.Publisher(s.ctx, client, read, done)
	go client.WritePump(s.ctx, done)
	client.ReadPump(s.ctx, read)

	<-done // WritePump

	room.RemoveClient(client)
}
