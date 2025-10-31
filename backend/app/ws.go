package app

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	jwt2 "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

type SignalMessage struct {
	Type      string          `json:"type"`
	Offer     json.RawMessage `json:"offer,omitempty"`
	Answer    json.RawMessage `json:"answer,omitempty"`
	Candidate json.RawMessage `json:"candidate,omitempty"`
}

type Client struct {
	conn     *websocket.Conn
	roomName string
	username string
	send     chan []byte
}

type Claims struct {
	Username string `json:"username"`
	Room     string `json:"room"`
	jwt2.RegisteredClaims
}

const bufferSize = 256

var upgrader = websocket.Upgrader{
	ReadBufferSize:  bufferSize,
	WriteBufferSize: bufferSize,
	WriteBufferPool: &sync.Pool{},
}

func SignalHandler(ctx context.Context, hub *RoomStorage, jwt *JWT) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.URL.Query().Get("token")
		if tokenStr == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		claims := &Claims{}
		token, err := jwt2.ParseWithClaims(tokenStr, claims, func(token *jwt2.Token) (interface{}, error) {
			return jwt.secret, nil
		})
		if err != nil {
			log.Printf("error validating jwt %s", err.Error())
			return
		}

		if !token.Valid {
			log.Printf("invalid token %s", tokenStr)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("ws upgrade error:", err)
			return
		}

		client := &Client{
			conn:     conn,
			roomName: claims.Room,
			username: claims.Username,
			send:     make(chan []byte, bufferSize),
		}

		hub.mu.RLock()
		room, ok := hub.rooms[claims.Room]
		if !ok {
			log.Printf("room not found %s", claims.Room)
			hub.mu.RUnlock()
			return
		}
		hub.mu.RUnlock()

		room.AddClient(client)

		go client.writePump(ctx)
		client.readPump(ctx, room)

		room.RemoveClient(client, hub)
	}
}

// Broadcast отправляет сообщение всем участникам комнаты (кроме отправителя)
func (r *Room) Broadcast(sender *Client, data []byte) {
	r.RoomSockets.mu.Lock()
	defer r.RoomSockets.mu.Unlock()
	for c := range r.RoomSockets.wsClients {
		if c != sender {
			select {
			case c.send <- data:
			default:
				close(c.send)
				delete(r.RoomSockets.wsClients, c)
			}
		}
	}
}

func (r *Room) AddClient(c *Client) {
	r.RoomSockets.mu.Lock()
	r.RoomSockets.wsClients[c] = true
	r.RoomSockets.mu.Unlock()
	log.Printf("👤 %s присоединился к комнате %s", c.username, r.ID)
}

func (r *Room) RemoveClient(c *Client, hub *RoomStorage) {
	r.RoomSockets.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.RoomSockets.wsClients[c]; ok {
		delete(r.RoomSockets.wsClients, c)
		close(c.send)
		log.Printf("👋 %s покинул комнату %s", c.username, r.ID)
	}

	if len(r.RoomSockets.wsClients) == 0 {
		log.Printf("🧹 комната %s пуста, удаляю", r.ID)

		hub.mu.Lock()
		delete(hub.rooms, r.ID)
		hub.mu.Unlock()
	}
}

// Чтение сообщений от клиента
func (c *Client) readPump(ctx context.Context, room *Room) {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}

			return
		}

		room.Broadcast(c, msg)
	}
}

// Отправка сообщений клиенту
func (c *Client) writePump(ctx context.Context) {
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Println("write:", err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
