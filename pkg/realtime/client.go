package realtime

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a connected WebSocket client
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	userID    uint
	send      chan interface{}
	channels  map[string]bool
	mu        sync.RWMutex
	closed    bool
}

// NewClient creates a new client
func NewClient(hub *Hub, conn *websocket.Conn, userID uint) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		userID:   userID,
		send:     make(chan interface{}, 256), // Buffered channel
		channels: make(map[string]bool),
		closed:   false,
	}
}

// Subscribe adds the client to a channel
func (c *Client) Subscribe(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.channels[channel] = true
}

// Unsubscribe removes the client from a channel
func (c *Client) Unsubscribe(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.channels, channel)
}

// IsSubscribed checks if the client is subscribed to a channel
func (c *Client) IsSubscribed(channel string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.channels[channel]
}

// GetChannels returns all subscribed channels
func (c *Client) GetChannels() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	channels := make([]string, 0, len(c.channels))
	for ch := range c.channels {
		channels = append(channels, ch)
	}
	return channels
}

// Send sends a message to the client
func (c *Client) Send(msg interface{}) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return
	}
	c.mu.RUnlock()

	select {
	case c.send <- msg:
	default:
		// Channel full, close and remove client
		c.Close()
	}
}

// Close closes the client connection
func (c *Client) Close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	c.mu.Unlock()

	close(c.send)
	if c.conn != nil {
		c.conn.Close()
	}
}

// ReadPump reads messages from the WebSocket connection
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.Close()
	}()

	// Set reasonable timeouts
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		// Read incoming message
		var msg map[string]interface{}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log unexpected close
			}
			break
		}

		// Handle subscription messages
		if msgType, ok := msg["type"].(string); ok {
			switch msgType {
			case string(MessageTypeSubscribe):
				if channels, ok := msg["channels"].([]interface{}); ok {
					for _, ch := range channels {
						if channelStr, ok := ch.(string); ok {
							c.Subscribe(channelStr)
						}
					}
				}

			case string(MessageTypeUnsubscribe):
				if channels, ok := msg["channels"].([]interface{}); ok {
					for _, ch := range channels {
						if channelStr, ok := ch.(string); ok {
							c.Unsubscribe(channelStr)
						}
					}
				}

			case string(MessageTypePing):
				// Respond with pong
				pong := NewMessage(MessageTypePong, "", nil)
				c.Send(pong)
			}
		}
	}
}

// WritePump writes messages to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Convert message to JSON and send
			var jsonBytes []byte
			var err error

			switch v := msg.(type) {
			case *Message:
				jsonBytes, err = json.Marshal(v)
			case map[string]interface{}:
				jsonBytes, err = json.Marshal(v)
			default:
				jsonBytes, err = json.Marshal(msg)
			}

			if err != nil {
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ChannelCount returns the number of subscribed channels
func (c *Client) ChannelCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.channels)
}

// IsClosed returns whether the client is closed
func (c *Client) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}
