package dashboard

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jgirmay/unified-go/pkg/realtime"
)

// WebSocketHandler manages WebSocket connections
type WebSocketHandler struct {
	hub            *realtime.Hub
	upgrader       websocket.Upgrader
	maxMessageSize int64
	mu             sync.RWMutex
	stats          WSStats
}

// WSStats contains WebSocket statistics
type WSStats struct {
	mu                    sync.RWMutex
	TotalConnections      int64
	ActiveConnections     int64
	TotalMessagesReceived int64
	TotalMessagesSent     int64
	ConnectionErrors      int64
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *realtime.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub:            hub,
		maxMessageSize: 512 * 1024, // 512KB max message size
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
		},
	}
}

// HandleConnection handles a new WebSocket connection
func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from query params or auth header
	userID, err := extractUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.stats.mu.Lock()
		h.stats.ConnectionErrors++
		h.stats.mu.Unlock()
		fmt.Printf("[WebSocket] Upgrade error: %v\n", err)
		return
	}

	// Create client
	client := realtime.NewClient(h.hub, conn, userID)

	// Register with hub
	h.hub.Register(client)

	h.stats.mu.Lock()
	h.stats.TotalConnections++
	h.stats.ActiveConnections++
	h.stats.mu.Unlock()

	fmt.Printf("[WebSocket] User %d connected\n", userID)

	// Set connection limits
	conn.SetReadLimit(h.maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// Start read and write pumps
	go client.ReadPump()
	go client.WritePump()

	// Monitor connection
	go h.monitorConnection(client, userID)
}

// monitorConnection monitors a client connection
func (h *WebSocketHandler) monitorConnection(client *realtime.Client, userID uint) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if client.IsClosed() {
			h.hub.Unregister(client)

			h.stats.mu.Lock()
			h.stats.ActiveConnections--
			h.stats.mu.Unlock()

			fmt.Printf("[WebSocket] User %d disconnected\n", userID)
			return
		}
	}
}

// BroadcastToChannel broadcasts a message to all clients in a channel
func (h *WebSocketHandler) BroadcastToChannel(channel string, message interface{}) {
	h.hub.Broadcast(channel, message)

	h.stats.mu.Lock()
	h.stats.TotalMessagesSent++
	h.stats.mu.Unlock()
}

// BroadcastToUser broadcasts a message to a specific user
func (h *WebSocketHandler) BroadcastToUser(channel string, userID uint, message interface{}) {
	h.hub.BroadcastToUser(channel, userID, message)

	h.stats.mu.Lock()
	h.stats.TotalMessagesSent++
	h.stats.mu.Unlock()
}

// GetStats returns WebSocket handler statistics
func (h *WebSocketHandler) GetStats() WSStats {
	h.stats.mu.RLock()
	defer h.stats.mu.RUnlock()
	return h.stats
}

// GetActiveConnectionCount returns the number of active connections
func (h *WebSocketHandler) GetActiveConnectionCount() int {
	h.stats.mu.RLock()
	defer h.stats.mu.RUnlock()
	return int(h.stats.ActiveConnections)
}

// extractUserID extracts user ID from request
func extractUserID(r *http.Request) (uint, error) {
	// Try query param first
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err == nil {
			return uint(userID), nil
		}
	}

	// Try Authorization header (Bearer token format)
	// In production, decode and verify JWT
	if auth := r.Header.Get("Authorization"); auth != "" {
		// For now, extract user ID from X-User-ID header
		if userIDStr := r.Header.Get("X-User-ID"); userIDStr != "" {
			userID, err := strconv.ParseUint(userIDStr, 10, 32)
			if err == nil {
				return uint(userID), nil
			}
		}
	}

	return 0, fmt.Errorf("missing user ID")
}

// ConnectionMessage represents a connection message
type ConnectionMessage struct {
	Type      string   `json:"type"`
	UserID    uint     `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
	Channels  []string `json:"channels,omitempty"`
}

// SubscriptionRequest represents a subscription request
type SubscriptionRequest struct {
	Type     string   `json:"type"`
	Channels []string `json:"channels"`
}

// UnsubscriptionRequest represents an unsubscription request
type UnsubscriptionRequest struct {
	Type     string   `json:"type"`
	Channels []string `json:"channels"`
}

// PingMessage represents a ping message
type PingMessage struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

// PongMessage represents a pong message
type PongMessage struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

// HandleSubscribe handles subscription requests
func (h *WebSocketHandler) HandleSubscribe(client *realtime.Client, channels []string) error {
	for _, channel := range channels {
		h.hub.Subscribe(client, channel)
	}
	return nil
}

// HandleUnsubscribe handles unsubscription requests
func (h *WebSocketHandler) HandleUnsubscribe(client *realtime.Client, channels []string) error {
	for _, channel := range channels {
		h.hub.Unsubscribe(client, channel)
	}
	return nil
}
