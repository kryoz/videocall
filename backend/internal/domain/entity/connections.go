package entity

import (
	"context"
	"log"
	"sync"
	"videocall/internal/infrastructure/messaging"
)

type Connections struct {
	WsClients map[*messaging.Client]bool
	mu        sync.RWMutex
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
				// Read channel closed - ReadPump finishes
				//log.Printf("room %s user %s: publisher closes", r.ID, sender.Username)
				return
			}
			r.mu.RLock()
			for c := range r.WsClients {
				if c != sender {
					c.Send(msg)
				}
			}
			r.mu.RUnlock()
		}
	}
}

func (r *Connections) AddClient(c *messaging.Client, roomID string) {
	r.mu.Lock()
	r.WsClients[c] = true
	r.mu.Unlock()
	log.Printf("👤 %s joined room %s", c.Username, roomID)
}

func (r *Connections) RemoveClient(c *messaging.Client, roomID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.WsClients[c]; ok {
		delete(r.WsClients, c)
		log.Printf("👋 %s left room %s", c.Username, roomID)
	}

	c.Close()

	// Let handleEmptyRooms clean up later
}
