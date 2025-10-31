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
	Username string
	conn     *websocket.Conn
	send     chan []byte
	once     sync.Once // Ensure send channel is closed only once
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
		Username: claims.Username,
		conn:     conn,
		send:     make(chan []byte, BufferSize),
	}, nil
}

func (c *Client) ReadPump(ctx context.Context, read chan<- []byte) {
	defer func() {
		close(read) // закроет горутину с Publisher
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			//log.Printf("error: %v", err)
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
		//log.Println("closing client socket")
	})
}

func (c *Client) Send(msg []byte) {
	// Non-blocking send with proper error handling
	select {
	case c.send <- msg:
		//log.Printf("sent: %s", msg)
	default:
		// Channel is full, drop message instead of closing channel
		log.Printf("send channel full, dropping message for %s", c.Username)
	}
}
