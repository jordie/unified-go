package dashboard

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jgirmay/unified-go/pkg/realtime"
)

func TestNewWebSocketHandler(t *testing.T) {
	hub := realtime.NewHub()
	handler := NewWebSocketHandler(hub)

	if handler == nil {
		t.Fatal("NewWebSocketHandler returned nil")
	}
	if handler.hub != hub {
		t.Error("hub not set correctly")
	}
	if handler.maxMessageSize != 512*1024 {
		t.Errorf("maxMessageSize expected 512KB, got %d", handler.maxMessageSize)
	}
}

func TestHandleConnectionWithoutProperUpgrade(t *testing.T) {
	// This test verifies the handler checks for proper WebSocket upgrade headers
	hub := realtime.NewHub()
	go hub.Run()
	handler := NewWebSocketHandler(hub)

	// Create a test request without proper WebSocket headers
	req := httptest.NewRequest("GET", "ws://localhost/api/ws?user_id=123", nil)
	w := httptest.NewRecorder()

	// Upgrade to WebSocket (will fail without proper headers)
	handler.HandleConnection(w, req)

	// Should fail the upgrade without proper headers
	// Status should NOT be 101
	if w.Code == http.StatusSwitchingProtocols {
		t.Errorf("expected non-101 status without proper WebSocket headers, got %d", w.Code)
	}
}

func TestExtractUserIDFromQueryParam(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		expected  uint
		wantError bool
	}{
		{
			name:      "valid user_id query param",
			url:       "ws://localhost/api/ws?user_id=123",
			expected:  123,
			wantError: false,
		},
		{
			name:      "invalid user_id query param",
			url:       "ws://localhost/api/ws?user_id=invalid",
			expected:  0,
			wantError: true,
		},
		{
			name:      "missing user_id",
			url:       "ws://localhost/api/ws",
			expected:  0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			userID, err := extractUserID(req)

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if !tt.wantError && userID != tt.expected {
				t.Errorf("expected userID %d, got %d", tt.expected, userID)
			}
		})
	}
}

func TestExtractUserIDFromHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "ws://localhost/api/ws", nil)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-User-ID", "456")

	userID, err := extractUserID(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != 456 {
		t.Errorf("expected userID 456, got %d", userID)
	}
}

func TestGetStats(t *testing.T) {
	hub := realtime.NewHub()
	handler := NewWebSocketHandler(hub)

	// Simulate some activity
	handler.stats.TotalConnections = 10
	handler.stats.ActiveConnections = 5
	handler.stats.TotalMessagesReceived = 100
	handler.stats.TotalMessagesSent = 150
	handler.stats.ConnectionErrors = 2

	stats := handler.GetStats()

	if stats.TotalConnections != 10 {
		t.Errorf("TotalConnections expected 10, got %d", stats.TotalConnections)
	}
	if stats.ActiveConnections != 5 {
		t.Errorf("ActiveConnections expected 5, got %d", stats.ActiveConnections)
	}
	if stats.TotalMessagesReceived != 100 {
		t.Errorf("TotalMessagesReceived expected 100, got %d", stats.TotalMessagesReceived)
	}
	if stats.TotalMessagesSent != 150 {
		t.Errorf("TotalMessagesSent expected 150, got %d", stats.TotalMessagesSent)
	}
	if stats.ConnectionErrors != 2 {
		t.Errorf("ConnectionErrors expected 2, got %d", stats.ConnectionErrors)
	}
}

func TestGetActiveConnectionCount(t *testing.T) {
	hub := realtime.NewHub()
	handler := NewWebSocketHandler(hub)

	handler.stats.ActiveConnections = 7

	count := handler.GetActiveConnectionCount()
	if count != 7 {
		t.Errorf("expected count 7, got %d", count)
	}
}

func TestBroadcastToChannel(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	handler := NewWebSocketHandler(hub)

	initialCount := handler.stats.TotalMessagesSent
	handler.BroadcastToChannel("test:channel", map[string]string{"type": "test"})

	// Stats should increment
	if handler.stats.TotalMessagesSent != initialCount+1 {
		t.Errorf("TotalMessagesSent not incremented")
	}
}

func TestBroadcastToUser(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	handler := NewWebSocketHandler(hub)

	initialCount := handler.stats.TotalMessagesSent
	handler.BroadcastToUser("user:123:progress", 123, map[string]interface{}{"score": 100})

	// Stats should increment
	if handler.stats.TotalMessagesSent != initialCount+1 {
		t.Errorf("TotalMessagesSent not incremented")
	}
}

func TestConnectionMessageStruct(t *testing.T) {
	msg := ConnectionMessage{
		Type:      "connection",
		UserID:    123,
		Timestamp: time.Now(),
		Channels:  []string{"leaderboard:typing_wpm", "user:123:progress"},
	}

	if msg.Type != "connection" {
		t.Error("ConnectionMessage.Type not set correctly")
	}
	if msg.UserID != 123 {
		t.Error("ConnectionMessage.UserID not set correctly")
	}
	if len(msg.Channels) != 2 {
		t.Error("ConnectionMessage.Channels not set correctly")
	}
}

func TestSubscriptionRequestStruct(t *testing.T) {
	req := SubscriptionRequest{
		Type:     "subscribe",
		Channels: []string{"leaderboard:overall", "activity:feed"},
	}

	if req.Type != "subscribe" {
		t.Error("SubscriptionRequest.Type not set correctly")
	}
	if len(req.Channels) != 2 {
		t.Error("SubscriptionRequest.Channels not set correctly")
	}
}

func TestUnsubscriptionRequestStruct(t *testing.T) {
	req := UnsubscriptionRequest{
		Type:     "unsubscribe",
		Channels: []string{"leaderboard:overall"},
	}

	if req.Type != "unsubscribe" {
		t.Error("UnsubscriptionRequest.Type not set correctly")
	}
	if len(req.Channels) != 1 {
		t.Error("UnsubscriptionRequest.Channels not set correctly")
	}
}

func TestPingPongMessages(t *testing.T) {
	pingMsg := PingMessage{
		Type:      "ping",
		Timestamp: time.Now(),
	}

	pongMsg := PongMessage{
		Type:      "pong",
		Timestamp: time.Now(),
	}

	if pingMsg.Type != "ping" {
		t.Error("PingMessage.Type not set correctly")
	}
	if pongMsg.Type != "pong" {
		t.Error("PongMessage.Type not set correctly")
	}
}

func TestWebSocketHandlerCreation(t *testing.T) {
	// Test that handler is properly created with all fields
	hub := realtime.NewHub()
	go hub.Run()
	handler := NewWebSocketHandler(hub)

	// Verify handler is not nil
	if handler == nil {
		t.Fatal("handler is nil")
	}

	// Verify hub reference
	if handler.hub != hub {
		t.Error("hub reference not set correctly")
	}

	// Verify maxMessageSize is set
	if handler.maxMessageSize == 0 {
		t.Error("maxMessageSize not set")
	}

	// Verify upgrader is configured
	if handler.upgrader.ReadBufferSize == 0 {
		t.Error("upgrader not configured")
	}
}

// mockConn is a mock WebSocket connection for testing
type mockConn struct {
	readDeadline  time.Time
	writeDeadline time.Time
	isClosed      bool
	messages      [][]byte
}

func (m *mockConn) ReadMessage() (messageType int, data []byte, err error) {
	if m.isClosed {
		return 0, nil, websocket.ErrCloseSent
	}
	// Return empty message to simulate connection
	time.Sleep(100 * time.Millisecond)
	return websocket.TextMessage, []byte("{}"), nil
}

func (m *mockConn) WriteMessage(messageType int, data []byte) error {
	if m.isClosed {
		return websocket.ErrCloseSent
	}
	m.messages = append(m.messages, data)
	return nil
}

func (m *mockConn) Close() error {
	m.isClosed = true
	return nil
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	m.readDeadline = t
	return nil
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	m.writeDeadline = t
	return nil
}

func (m *mockConn) SetReadLimit(limit int64) {}

func (m *mockConn) SetPongHandler(h func(string) error) {}

// TestStatsConcurrency verifies thread-safe access to stats
func TestStatsConcurrency(t *testing.T) {
	hub := realtime.NewHub()
	handler := NewWebSocketHandler(hub)

	// Simulate concurrent stats updates
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			handler.stats.mu.Lock()
			handler.stats.TotalConnections++
			handler.stats.ActiveConnections++
			handler.stats.mu.Unlock()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	stats := handler.GetStats()
	if stats.TotalConnections != 10 {
		t.Errorf("expected 10 connections, got %d", stats.TotalConnections)
	}
	if stats.ActiveConnections != 10 {
		t.Errorf("expected 10 active connections, got %d", stats.ActiveConnections)
	}
}

// TestWebSocketHeaderParsing verifies proper header parsing
func TestWebSocketHeaderParsing(t *testing.T) {
	tests := []struct {
		name         string
		setupReq     func(*http.Request)
		expectedID   uint
		wantError    bool
	}{
		{
			name: "query param only",
			setupReq: func(r *http.Request) {
				r.URL.RawQuery = "user_id=100"
			},
			expectedID: 100,
			wantError:  false,
		},
		{
			name: "header with authorization",
			setupReq: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer token123")
				r.Header.Set("X-User-ID", "200")
			},
			expectedID: 200,
			wantError:  false,
		},
		{
			name: "query param takes precedence",
			setupReq: func(r *http.Request) {
				r.URL.RawQuery = "user_id=300"
				r.Header.Set("X-User-ID", "400")
			},
			expectedID: 300,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "ws://localhost/api/ws", nil)
			tt.setupReq(req)

			userID, err := extractUserID(req)
			if (err != nil) != tt.wantError {
				t.Errorf("wantError %v, got error: %v", tt.wantError, err)
			}
			if !tt.wantError && userID != tt.expectedID {
				t.Errorf("expected userID %d, got %d", tt.expectedID, userID)
			}
		})
	}
}

// BenchmarkHandleConnection benchmarks WebSocket connection handling
func BenchmarkHandleConnection(b *testing.B) {
	hub := realtime.NewHub()
	go hub.Run()
	handler := NewWebSocketHandler(hub)

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "ws://localhost/api/ws?user_id="+strconv.Itoa(i), nil)
		w := httptest.NewRecorder()

		// Note: This will fail the upgrade but measures the overhead
		handler.HandleConnection(w, req)
	}
}

// BenchmarkBroadcastToChannel benchmarks broadcasting to a channel
func BenchmarkBroadcastToChannel(b *testing.B) {
	hub := realtime.NewHub()
	go hub.Run()
	handler := NewWebSocketHandler(hub)

	message := map[string]interface{}{
		"type":  "leaderboard_update",
		"score": 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.BroadcastToChannel("leaderboard:typing_wpm", message)
	}
}

// BenchmarkGetStats benchmarks stats retrieval
func BenchmarkGetStats(b *testing.B) {
	hub := realtime.NewHub()
	handler := NewWebSocketHandler(hub)

	// Populate some stats
	handler.stats.TotalConnections = 1000
	handler.stats.ActiveConnections = 500
	handler.stats.TotalMessagesReceived = 10000
	handler.stats.TotalMessagesSent = 15000

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.GetStats()
	}
}
