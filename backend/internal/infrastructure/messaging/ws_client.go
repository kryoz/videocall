package messaging

import (
	"context"
	"log"
	"net/http"
	"sync"
	"videocall/internal/infrastructure/auth"

	"github.com/gorilla/websocket"
)

const BufferSize = 256

type Client struct {
	UserID   string
	Username string
	RoomID   string
	conn     *websocket.Conn
	send     chan []byte
	once     sync.Once
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  BufferSize,
	WriteBufferSize: BufferSize,
	WriteBufferPool: &sync.Pool{},
}

func NewClient(w http.ResponseWriter, r *http.Request, claims *auth.Claims) (*Client, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		UserID:   claims.UserID,
		Username: claims.Username,
		RoomID:   claims.RoomID,
		conn:     conn,
		send:     make(chan []byte, BufferSize),
	}, nil
}

func (c *Client) ReadPump(ctx context.Context, read chan<- []byte) {
	defer func() {
		close(read)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		read <- msg
	}
}

func (c *Client) WritePump(ctx context.Context, finish chan<- struct{}) {
	defer func() {
		finish <- struct{}{}
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Println("write error:", err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) Close() {
	c.once.Do(func() {
		close(c.send)
		_ = c.conn.Close()
	})
}

func (c *Client) Send(msg []byte) {
	select {
	case c.send <- msg:
	default:
		log.Printf("send channel full, dropping message for user %s (%s)", c.Username, c.UserID)
	}
}
