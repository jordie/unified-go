package realtime

import (
	"sync"
)

// Hub manages all active WebSocket connections
type Hub struct {
	// Registered clients
	clients map[*Client]bool
	mu      sync.RWMutex

	// Channel subscriptions (channel -> set of clients)
	subscriptions map[string]map[*Client]bool
	subMu         sync.RWMutex

	// User clients (userID -> set of clients)
	userClients map[uint][]*Client
	userMu      sync.RWMutex

	// Broadcast channel
	broadcast chan BroadcastMessage

	// Register a new client
	register chan *Client

	// Unregister a client
	unregister chan *Client

	// Stop the hub
	stop chan bool

	// Hub statistics
	stats HubStats
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	Channel string
	Message interface{}
	UserID  uint // Optional: only send to specific user
}

// HubStats contains hub statistics
type HubStats struct {
	mu              sync.RWMutex
	TotalClients    int64
	TotalMessages   int64
	TotalBroadcasts int64
	ActiveChannels  int64
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		subscriptions: make(map[string]map[*Client]bool),
		userClients:   make(map[uint][]*Client),
		broadcast:     make(chan BroadcastMessage, 256),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		stop:          make(chan bool),
	}
}

// Run starts the hub event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case msg := <-h.broadcast:
			h.broadcastMessage(msg)

		case <-h.stop:
			h.shutdown()
			return
		}
	}
}

// registerClient adds a client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	h.clients[client] = true
	h.mu.Unlock()

	h.userMu.Lock()
	h.userClients[client.userID] = append(h.userClients[client.userID], client)
	h.userMu.Unlock()

	h.stats.mu.Lock()
	h.stats.TotalClients++
	h.stats.mu.Unlock()
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)

		// Remove from all subscriptions
		h.subMu.Lock()
		for channel := range h.subscriptions {
			delete(h.subscriptions[channel], client)
			// Clean up empty channel subscriptions
			if len(h.subscriptions[channel]) == 0 {
				delete(h.subscriptions, channel)
			}
		}
		h.subMu.Unlock()
	}
	h.mu.Unlock()

	// Remove from user clients
	h.userMu.Lock()
	if clients, ok := h.userClients[client.userID]; ok {
		for i, c := range clients {
			if c == client {
				h.userClients[client.userID] = append(clients[:i], clients[i+1:]...)
				break
			}
		}
		if len(h.userClients[client.userID]) == 0 {
			delete(h.userClients, client.userID)
		}
	}
	h.userMu.Unlock()
}

// broadcastMessage sends a message to subscribed clients
func (h *Hub) broadcastMessage(msg BroadcastMessage) {
	h.subMu.RLock()
	subscribers, ok := h.subscriptions[msg.Channel]
	h.subMu.RUnlock()

	if !ok {
		return
	}

	h.stats.mu.Lock()
	h.stats.TotalBroadcasts++
	h.stats.mu.Unlock()

	// Send to all subscribers
	for client := range subscribers {
		// If UserID is specified, only send to that user
		if msg.UserID > 0 && client.userID != msg.UserID {
			continue
		}

		client.Send(msg.Message)

		h.stats.mu.Lock()
		h.stats.TotalMessages++
		h.stats.mu.Unlock()
	}
}

// Subscribe adds a client to a channel subscription
func (h *Hub) Subscribe(client *Client, channel string) {
	h.subMu.Lock()
	defer h.subMu.Unlock()

	if _, ok := h.subscriptions[channel]; !ok {
		h.subscriptions[channel] = make(map[*Client]bool)
	}

	h.subscriptions[channel][client] = true
	client.Subscribe(channel)

	// Update active channels stat
	h.stats.mu.Lock()
	h.stats.ActiveChannels = int64(len(h.subscriptions))
	h.stats.mu.Unlock()
}

// Unsubscribe removes a client from a channel subscription
func (h *Hub) Unsubscribe(client *Client, channel string) {
	h.subMu.Lock()
	defer h.subMu.Unlock()

	if subscribers, ok := h.subscriptions[channel]; ok {
		delete(subscribers, client)
		client.Unsubscribe(channel)

		// Clean up empty channel
		if len(subscribers) == 0 {
			delete(h.subscriptions, channel)
		}
	}

	// Update active channels stat
	h.stats.mu.Lock()
	h.stats.ActiveChannels = int64(len(h.subscriptions))
	h.stats.mu.Unlock()
}

// Broadcast sends a message to all clients subscribed to a channel
func (h *Hub) Broadcast(channel string, message interface{}) {
	h.broadcast <- BroadcastMessage{
		Channel: channel,
		Message: message,
	}
}

// BroadcastToUser sends a message to a specific user on a channel
func (h *Hub) BroadcastToUser(channel string, userID uint, message interface{}) {
	h.broadcast <- BroadcastMessage{
		Channel: channel,
		Message: message,
		UserID:  userID,
	}
}

// Register registers a new client
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Stop stops the hub
func (h *Hub) Stop() {
	h.stop <- true
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetChannelSubscribers returns the number of subscribers for a channel
func (h *Hub) GetChannelSubscribers(channel string) int {
	h.subMu.RLock()
	defer h.subMu.RUnlock()

	if subscribers, ok := h.subscriptions[channel]; ok {
		return len(subscribers)
	}
	return 0
}

// GetUserClients returns all connected clients for a user
func (h *Hub) GetUserClients(userID uint) []*Client {
	h.userMu.RLock()
	defer h.userMu.RUnlock()

	if clients, ok := h.userClients[userID]; ok {
		return clients
	}
	return nil
}

// GetStats returns hub statistics
func (h *Hub) GetStats() HubStats {
	h.stats.mu.RLock()
	defer h.stats.mu.RUnlock()
	return h.stats
}

// GetActiveChannels returns the list of active channels
func (h *Hub) GetActiveChannels() []string {
	h.subMu.RLock()
	defer h.subMu.RUnlock()

	channels := make([]string, 0, len(h.subscriptions))
	for ch := range h.subscriptions {
		channels = append(channels, ch)
	}
	return channels
}

// shutdown gracefully closes the hub
func (h *Hub) shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		client.Close()
	}
	h.clients = make(map[*Client]bool)
}
