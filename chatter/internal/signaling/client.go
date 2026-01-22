package signaling

import (
	"context"

	"github.com/coder/websocket"
)

type Client struct {
	userID   uint64
	username string
	conn     *websocket.Conn
	room     *Room
	send     chan []byte
}

func NewClient(userID uint64, username string, conn *websocket.Conn, room *Room) *Client {
	return &Client{
		userID:   userID,
		username: username,
		conn:     conn,
		room:     room,
		send:     make(chan []byte, 32),
	}
}

func (c *Client) Run(ctx context.Context) {
	go c.writeLoop(ctx)
	c.readLoop(ctx)
}

func (c *Client) readLoop(ctx context.Context) {
	defer func() {
		c.room.Unregister(c)
		_ = c.conn.Close(websocket.StatusNormalClosure, "client closed")
	}()

	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			return
		}

		c.room.HandleIncoming(c, data)
	}
}

func (c *Client) ID() uint64 {
	return c.userID
}

func (c *Client) writeLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.conn.Write(ctx, websocket.MessageText, msg); err != nil {
				return
			}
		}
	}
}
