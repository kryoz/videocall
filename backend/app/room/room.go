package room

import (
	"context"
	"log"
	"sync"
	"time"
	"videocall/app/signaling"
)

type Room struct {
	ID        string
	Name      string
	Password  string
	CreatedAt time.Time
	Connections
}

type Connections struct {
	WsClients map[*signaling.Client]bool
	mu        sync.RWMutex
}

func (r *Room) Publisher(ctx context.Context, sender *signaling.Client, read <-chan []byte, done chan<- struct{}) {
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
			r.Connections.mu.RLock()
			for c := range r.Connections.WsClients {
				if c != sender {
					c.Send(msg)
				}
			}
			r.Connections.mu.RUnlock()
		}
	}
}

func (r *Room) AddClient(c *signaling.Client) {
	r.Connections.mu.Lock()
	r.Connections.WsClients[c] = true
	r.Connections.mu.Unlock()
	log.Printf("👤 %s joined room %s", c.Username, r.ID)
}

func (r *Room) RemoveClient(c *signaling.Client) {
	r.Connections.mu.Lock()
	defer r.Connections.mu.Unlock()

	if _, ok := r.Connections.WsClients[c]; ok {
		delete(r.Connections.WsClients, c)
		log.Printf("👋 %s left room %s", c.Username, r.ID)
	}

	c.Close()

	// Let handleEmptyRooms clean up later
}
