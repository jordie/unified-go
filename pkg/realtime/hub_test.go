package realtime

import (
	"testing"
	"time"
)

// TestHubCreation tests hub initialization
func TestHubCreation(t *testing.T) {
	hub := NewHub()
	if hub == nil {
		t.Fatal("Hub creation failed")
	}

	if hub.GetClientCount() != 0 {
		t.Errorf("Expected 0 clients initially, got %d", hub.GetClientCount())
	}

	if hub.GetActiveChannels() != nil && len(hub.GetActiveChannels()) != 0 {
		t.Errorf("Expected 0 active channels initially")
	}
}

// TestClientRegistration tests registering a client
func TestClientRegistration(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	time.Sleep(10 * time.Millisecond)

	// Create mock client
	client := &Client{
		hub:      hub,
		userID:   1,
		channels: make(map[string]bool),
		send:     make(chan interface{}, 10),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	if hub.GetClientCount() != 1 {
		t.Errorf("Expected 1 client after registration, got %d", hub.GetClientCount())
	}
}

// TestClientUnregistration tests unregistering a client
func TestClientUnregistration(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	time.Sleep(10 * time.Millisecond)

	client := &Client{
		hub:      hub,
		userID:   1,
		channels: make(map[string]bool),
		send:     make(chan interface{}, 10),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	if hub.GetClientCount() != 1 {
		t.Errorf("Expected 1 client after registration")
	}

	hub.Unregister(client)
	time.Sleep(10 * time.Millisecond)

	if hub.GetClientCount() != 0 {
		t.Errorf("Expected 0 clients after unregistration, got %d", hub.GetClientCount())
	}
}

// TestChannelSubscription tests subscribing to a channel
func TestChannelSubscription(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	time.Sleep(10 * time.Millisecond)

	client := &Client{
		hub:      hub,
		userID:   1,
		channels: make(map[string]bool),
		send:     make(chan interface{}, 10),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	channel := "leaderboard:typing_wpm"
	hub.Subscribe(client, channel)
	time.Sleep(10 * time.Millisecond)

	if hub.GetChannelSubscribers(channel) != 1 {
		t.Errorf("Expected 1 subscriber for channel, got %d", hub.GetChannelSubscribers(channel))
	}

	if !client.IsSubscribed(channel) {
		t.Error("Client should be subscribed to channel")
	}
}

// TestChannelUnsubscription tests unsubscribing from a channel
func TestChannelUnsubscription(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	time.Sleep(10 * time.Millisecond)

	client := &Client{
		hub:      hub,
		userID:   1,
		channels: make(map[string]bool),
		send:     make(chan interface{}, 10),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	channel := "leaderboard:typing_wpm"
	hub.Subscribe(client, channel)
	time.Sleep(10 * time.Millisecond)

	if hub.GetChannelSubscribers(channel) != 1 {
		t.Errorf("Expected 1 subscriber")
	}

	hub.Unsubscribe(client, channel)
	time.Sleep(10 * time.Millisecond)

	if hub.GetChannelSubscribers(channel) != 0 {
		t.Errorf("Expected 0 subscribers after unsubscription")
	}
}

// TestBroadcast tests broadcasting a message
func TestBroadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	time.Sleep(10 * time.Millisecond)

	client1 := &Client{
		hub:      hub,
		userID:   1,
		channels: make(map[string]bool),
		send:     make(chan interface{}, 10),
	}

	client2 := &Client{
		hub:      hub,
		userID:   2,
		channels: make(map[string]bool),
		send:     make(chan interface{}, 10),
	}

	hub.Register(client1)
	hub.Register(client2)
	time.Sleep(10 * time.Millisecond)

	channel := "activity:feed"
	hub.Subscribe(client1, channel)
	hub.Subscribe(client2, channel)
	time.Sleep(10 * time.Millisecond)

	// Send test message
	testMsg := NewMessage(MessageTypeActivityFeed, channel, map[string]interface{}{
		"event": "test",
	})

	hub.Broadcast(channel, testMsg)
	time.Sleep(50 * time.Millisecond)

	// Check if both clients received the message
	received1 := false
	received2 := false

	select {
	case msg := <-client1.send:
		if msg != nil {
			received1 = true
		}
	default:
	}

	select {
	case msg := <-client2.send:
		if msg != nil {
			received2 = true
		}
	default:
	}

	if !received1 {
		t.Error("Client 1 did not receive broadcast message")
	}

	if !received2 {
		t.Error("Client 2 did not receive broadcast message")
	}
}

// TestBroadcastToUser tests broadcasting to a specific user
func TestBroadcastToUser(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	time.Sleep(10 * time.Millisecond)

	client1 := &Client{
		hub:      hub,
		userID:   1,
		channels: make(map[string]bool),
		send:     make(chan interface{}, 10),
	}

	client2 := &Client{
		hub:      hub,
		userID:   2,
		channels: make(map[string]bool),
		send:     make(chan interface{}, 10),
	}

	hub.Register(client1)
	hub.Register(client2)
	time.Sleep(10 * time.Millisecond)

	channel := "user:1:achievements"
	hub.Subscribe(client1, channel)
	hub.Subscribe(client2, channel)
	time.Sleep(10 * time.Millisecond)

	// Send message only to user 1
	testMsg := NewMessage(MessageTypeAchievementUnlocked, channel, map[string]interface{}{
		"achievement": "test",
	})

	hub.BroadcastToUser(channel, 1, testMsg)
	time.Sleep(50 * time.Millisecond)

	// Check if only client 1 received the message
	var msg1 interface{}
	select {
	case msg := <-client1.send:
		msg1 = msg
	default:
	}

	var msg2 interface{}
	select {
	case msg := <-client2.send:
		msg2 = msg
	default:
	}

	if msg1 == nil {
		t.Error("Client 1 (correct user) should receive the message")
	}

	if msg2 != nil {
		t.Error("Client 2 (wrong user) should not receive the message")
	}
}

// TestMultipleChannelSubscription tests subscribing to multiple channels
func TestMultipleChannelSubscription(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	time.Sleep(10 * time.Millisecond)

	client := &Client{
		hub:      hub,
		userID:   1,
		channels: make(map[string]bool),
		send:     make(chan interface{}, 10),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	channels := []string{
		"leaderboard:typing_wpm",
		"leaderboard:math_accuracy",
		"user:1:achievements",
		"activity:feed",
	}

	for _, ch := range channels {
		hub.Subscribe(client, ch)
	}
	time.Sleep(10 * time.Millisecond)

	if client.ChannelCount() != len(channels) {
		t.Errorf("Expected %d channels, got %d", len(channels), client.ChannelCount())
	}

	// Verify each channel has 1 subscriber
	for _, ch := range channels {
		if hub.GetChannelSubscribers(ch) != 1 {
			t.Errorf("Expected 1 subscriber for %s, got %d", ch, hub.GetChannelSubscribers(ch))
		}
	}
}

// TestGetUserClients tests getting all clients for a user
func TestGetUserClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	time.Sleep(10 * time.Millisecond)

	// Register 3 clients for user 1 and 2 for user 2
	for i := 0; i < 3; i++ {
		client := &Client{
			hub:      hub,
			userID:   1,
			channels: make(map[string]bool),
			send:     make(chan interface{}, 10),
		}
		hub.Register(client)
	}

	for i := 0; i < 2; i++ {
		client := &Client{
			hub:      hub,
			userID:   2,
			channels: make(map[string]bool),
			send:     make(chan interface{}, 10),
		}
		hub.Register(client)
	}

	time.Sleep(10 * time.Millisecond)

	clients1 := hub.GetUserClients(1)
	clients2 := hub.GetUserClients(2)

	if len(clients1) != 3 {
		t.Errorf("Expected 3 clients for user 1, got %d", len(clients1))
	}

	if len(clients2) != 2 {
		t.Errorf("Expected 2 clients for user 2, got %d", len(clients2))
	}
}

// TestGetStats tests retrieving hub statistics
func TestGetStats(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	time.Sleep(10 * time.Millisecond)

	for i := 0; i < 5; i++ {
		client := &Client{
			hub:      hub,
			userID:   uint(i),
			channels: make(map[string]bool),
			send:     make(chan interface{}, 10),
		}
		hub.Register(client)
	}

	time.Sleep(10 * time.Millisecond)

	stats := hub.GetStats()

	if stats.TotalClients != 5 {
		t.Errorf("Expected 5 total clients, got %d", stats.TotalClients)
	}
}

// TestGetActiveChannels tests getting list of active channels
func TestGetActiveChannels(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	time.Sleep(10 * time.Millisecond)

	client := &Client{
		hub:      hub,
		userID:   1,
		channels: make(map[string]bool),
		send:     make(chan interface{}, 10),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	channels := []string{
		"leaderboard:typing_wpm",
		"activity:feed",
		"user:1:achievements",
	}

	for _, ch := range channels {
		hub.Subscribe(client, ch)
	}

	time.Sleep(10 * time.Millisecond)

	activeChannels := hub.GetActiveChannels()

	if len(activeChannels) != len(channels) {
		t.Errorf("Expected %d active channels, got %d", len(channels), len(activeChannels))
	}
}

// TestEmptyChannelCleanup tests that empty channels are cleaned up
func TestEmptyChannelCleanup(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	time.Sleep(10 * time.Millisecond)

	client := &Client{
		hub:      hub,
		userID:   1,
		channels: make(map[string]bool),
		send:     make(chan interface{}, 10),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	channel := "test:channel"
	hub.Subscribe(client, channel)
	time.Sleep(10 * time.Millisecond)

	if hub.GetChannelSubscribers(channel) != 1 {
		t.Error("Expected channel to have 1 subscriber")
	}

	hub.Unsubscribe(client, channel)
	time.Sleep(10 * time.Millisecond)

	activeChannels := hub.GetActiveChannels()
	if len(activeChannels) > 0 {
		t.Error("Empty channel should be cleaned up")
	}
}

// BenchmarkBroadcast benchmarks broadcast performance
func BenchmarkBroadcast(b *testing.B) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	time.Sleep(10 * time.Millisecond)

	// Register 100 clients
	for i := 0; i < 100; i++ {
		client := &Client{
			hub:      hub,
			userID:   uint(i),
			channels: make(map[string]bool),
			send:     make(chan interface{}, 256),
		}
		hub.Register(client)
		hub.Subscribe(client, "benchmark:channel")
	}

	time.Sleep(50 * time.Millisecond)

	b.ResetTimer()

	msg := NewMessage(MessageTypeActivityFeed, "benchmark:channel", map[string]interface{}{
		"test": "data",
	})

	for i := 0; i < b.N; i++ {
		hub.Broadcast("benchmark:channel", msg)
	}

	b.StopTimer()
	time.Sleep(50 * time.Millisecond)
}
