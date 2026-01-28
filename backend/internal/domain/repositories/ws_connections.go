package repositories

import (
	"context"
	"log"
	"sync"
	"videocall/internal/infrastructure/messaging"
)

type Connections struct {
	WsClients map[string]*messaging.Client
	mu        sync.RWMutex
}

func NewConnections() *Connections {
	return &Connections{
		WsClients: make(map[string]*messaging.Client),
	}
}

func (r *Connections) Publisher(ctx context.Context, sender *messaging.Client, read <-chan []byte, done chan<- struct{}) {
	defer func() {
		done <- struct{}{}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-read:
			if !ok {
				return
			}
			r.mu.RLock()
			for userID, c := range r.WsClients {
				if userID != sender.UserID {
					c.Send(msg)
				}
			}
			r.mu.RUnlock()
		}
	}
}

func (r *Connections) AddClient(c *messaging.Client, roomID string) {
	r.mu.Lock()
	r.WsClients[c.UserID] = c
	r.mu.Unlock()
	log.Printf("👋 %s established connection with room %s", c.Username, roomID)
}

func (r *Connections) RemoveClient(c *messaging.Client, roomID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.WsClients[c.UserID]; ok {
		delete(r.WsClients, c.UserID)
		log.Printf("👤 %s disconnected from room %s", c.Username, roomID)
	}

	c.Close()
}
