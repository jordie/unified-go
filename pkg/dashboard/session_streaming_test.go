package dashboard

import (
	"context"
	"testing"
	"time"

	"github.com/jgirmay/unified-go/pkg/realtime"
)

func TestNewSessionStreamingManager(t *testing.T) {
	hub := realtime.NewHub()
	manager := NewSessionStreamingManager(hub)

	if manager == nil {
		t.Fatal("NewSessionStreamingManager returned nil")
	}
	if manager.hub != hub {
		t.Error("hub not set correctly")
	}
	if manager.progressTracker == nil {
		t.Error("progressTracker not initialized")
	}
	if manager.metricsAnalyzer == nil {
		t.Error("metricsAnalyzer not initialized")
	}
}

func TestStartSessionStream(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	manager := NewSessionStreamingManager(hub)

	ctx := context.Background()
	err := manager.StartSessionStream(ctx, "session1", 123, "typing")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify session was started
	progress := manager.GetSessionProgress("session1")
	if progress == nil {
		t.Fatal("session stream not started")
	}
	if progress.SessionID != "session1" {
		t.Errorf("expected session1, got %s", progress.SessionID)
	}
}

func TestStartSessionStreamDuplicate(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	manager := NewSessionStreamingManager(hub)

	ctx := context.Background()
	_ = manager.StartSessionStream(ctx, "session1", 123, "typing")

	// Try to start duplicate
	err := manager.StartSessionStream(ctx, "session1", 123, "typing")
	if err == nil {
		t.Error("expected error when starting duplicate session")
	}
}

func TestEndSessionStream(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	manager := NewSessionStreamingManager(hub)

	ctx := context.Background()
	_ = manager.StartSessionStream(ctx, "session1", 123, "typing")

	err := manager.EndSessionStream("session1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify session was removed from active sessions
	activeSessions := manager.GetActiveSessions()
	if len(activeSessions) != 0 {
		t.Error("expected no active sessions after ending")
	}
}

func TestEndSessionStreamNotFound(t *testing.T) {
	hub := realtime.NewHub()
	manager := NewSessionStreamingManager(hub)

	err := manager.EndSessionStream("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent session")
	}
}

func TestUpdateProgress(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	manager := NewSessionStreamingManager(hub)

	ctx := context.Background()
	_ = manager.StartSessionStream(ctx, "session1", 123, "typing")

	err := manager.UpdateProgress(ctx, "session1", "wpm", 150.5, "150.5 WPM")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify progress was updated
	progress := manager.GetSessionProgress("session1")
	if progress.CurrentMetric != 150.5 {
		t.Errorf("expected metric 150.5, got %f", progress.CurrentMetric)
	}
}

func TestUpdateProgressNotFound(t *testing.T) {
	hub := realtime.NewHub()
	manager := NewSessionStreamingManager(hub)

	ctx := context.Background()
	err := manager.UpdateProgress(ctx, "nonexistent", "wpm", 150.0, "")

	if err == nil {
		t.Error("expected error for non-existent session")
	}
}

func TestSessionGetSessionProgress(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	manager := NewSessionStreamingManager(hub)

	ctx := context.Background()

	// Non-existent session
	progress := manager.GetSessionProgress("nonexistent")
	if progress != nil {
		t.Error("expected nil for non-existent session")
	}

	// Existing session
	_ = manager.StartSessionStream(ctx, "session1", 123, "typing")
	progress = manager.GetSessionProgress("session1")
	if progress == nil {
		t.Fatal("expected session to exist")
	}
}

func TestGetMetricsAnalysis(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	manager := NewSessionStreamingManager(hub)

	ctx := context.Background()
	_ = manager.StartSessionStream(ctx, "session1", 123, "typing")

	// Record some metrics
	_ = manager.UpdateProgress(ctx, "session1", "wpm", 100.0, "100 WPM")
	time.Sleep(5 * time.Millisecond)
	_ = manager.UpdateProgress(ctx, "session1", "wpm", 120.0, "120 WPM")

	// Get analysis
	analysis := manager.GetMetricsAnalysis("session1")

	if len(analysis) == 0 {
		t.Error("expected metrics analysis")
	}

	if wpmStats, exists := analysis["wpm"]; exists {
		if wpmStats.Current != 120.0 {
			t.Errorf("expected current 120.0, got %f", wpmStats.Current)
		}
	}
}

func TestGetSessionSummary(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	manager := NewSessionStreamingManager(hub)

	ctx := context.Background()
	_ = manager.StartSessionStream(ctx, "session1", 123, "typing")

	// Record metrics
	_ = manager.UpdateProgress(ctx, "session1", "wpm", 150.0, "150 WPM")
	_ = manager.UpdateProgress(ctx, "session1", "accuracy", 95.0, "95%")

	// End session
	_ = manager.EndSessionStream("session1")

	// Get summary
	summary := manager.GetSessionSummary("session1")

	if summary == nil {
		t.Fatal("expected session summary")
	}
	if summary.SessionID != "session1" {
		t.Errorf("expected session1, got %s", summary.SessionID)
	}
	if summary.UserID != 123 {
		t.Errorf("expected userID 123, got %d", summary.UserID)
	}
}

func TestSessionUpdateSubscriberCount(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	manager := NewSessionStreamingManager(hub)

	ctx := context.Background()
	_ = manager.StartSessionStream(ctx, "session1", 123, "typing")

	// Update subscriber count
	manager.UpdateSubscriberCount("session1", 5)

	// Get stats to verify
	stats := manager.GetStreamingStats()
	if stats.TotalSubscribers != 5 {
		t.Errorf("expected 5 subscribers, got %d", stats.TotalSubscribers)
	}
}

func TestSessionGetActiveSessions(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	manager := NewSessionStreamingManager(hub)

	ctx := context.Background()

	// No active sessions
	sessions := manager.GetActiveSessions()
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}

	// Start multiple sessions
	_ = manager.StartSessionStream(ctx, "session1", 1, "typing")
	_ = manager.StartSessionStream(ctx, "session2", 2, "math")

	sessions = manager.GetActiveSessions()
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessions))
	}
}

func TestSessionGetStreamingStats(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	manager := NewSessionStreamingManager(hub)

	ctx := context.Background()
	_ = manager.StartSessionStream(ctx, "session1", 123, "typing")

	// Record some updates
	_ = manager.UpdateProgress(ctx, "session1", "wpm", 100.0, "100 WPM")
	_ = manager.UpdateProgress(ctx, "session1", "accuracy", 95.0, "95%")

	stats := manager.GetStreamingStats()

	if stats.ActiveSessions != 1 {
		t.Errorf("expected 1 active session, got %d", stats.ActiveSessions)
	}
	if stats.TotalUpdates != 2 {
		t.Errorf("expected 2 updates, got %d", stats.TotalUpdates)
	}
}

func TestBroadcastSessionStart(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	manager := NewSessionStreamingManager(hub)

	// Broadcast should not panic
	manager.broadcastSessionStart("session1", 123, "typing")

	// Test passes if no panic
}

func TestBroadcastSessionEnd(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	manager := NewSessionStreamingManager(hub)

	ctx := context.Background()
	_ = manager.StartSessionStream(ctx, "session1", 123, "typing")
	_ = manager.UpdateProgress(ctx, "session1", "wpm", 150.0, "150 WPM")

	summary := manager.GetSessionSummary("session1")

	// Broadcast should not panic
	manager.broadcastSessionEnd("session1", 123, summary)

	// Test passes if no panic
}

func TestBroadcastProgressUpdate(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	manager := NewSessionStreamingManager(hub)

	ctx := context.Background()
	_ = manager.StartSessionStream(ctx, "session1", 123, "typing")

	// Broadcast should not panic
	manager.broadcastProgressUpdate("session1", "wpm", 150.0, "150 WPM")

	// Test passes if no panic
}

func TestSessionCheckAndBroadcastMilestones(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	manager := NewSessionStreamingManager(hub)

	ctx := context.Background()

	// Test typing milestone
	_ = manager.StartSessionStream(ctx, "session1", 123, "typing")
	manager.checkAndBroadcastMilestones("session1", "wpm", 100.0)

	// Test math milestone
	_ = manager.StartSessionStream(ctx, "session2", 456, "math")
	manager.checkAndBroadcastMilestones("session2", "accuracy", 90.0)

	// Test piano milestone
	_ = manager.StartSessionStream(ctx, "session3", 789, "piano")
	manager.checkAndBroadcastMilestones("session3", "score", 500.0)

	// Test passes if no panic
}

// BenchmarkUpdateProgress benchmarks progress updates
func BenchmarkUpdateProgress(b *testing.B) {
	hub := realtime.NewHub()
	go hub.Run()
	manager := NewSessionStreamingManager(hub)

	ctx := context.Background()
	_ = manager.StartSessionStream(ctx, "session1", 123, "typing")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.UpdateProgress(ctx, "session1", "wpm", float64(100+i%50), "")
	}
}

// BenchmarkGetMetricsAnalysis benchmarks metrics analysis
func BenchmarkGetMetricsAnalysis(b *testing.B) {
	hub := realtime.NewHub()
	go hub.Run()
	manager := NewSessionStreamingManager(hub)

	ctx := context.Background()
	_ = manager.StartSessionStream(ctx, "session1", 123, "typing")

	for i := 0; i < 100; i++ {
		_ = manager.UpdateProgress(ctx, "session1", "wpm", float64(100+i), "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetMetricsAnalysis("session1")
	}
}

// BenchmarkSessionGetStreamingStats benchmarks stats retrieval
func BenchmarkSessionGetStreamingStats(b *testing.B) {
	hub := realtime.NewHub()
	go hub.Run()
	manager := NewSessionStreamingManager(hub)

	ctx := context.Background()
	for i := 0; i < 10; i++ {
		sessionID := "session_" + string(rune('0'+i))
		_ = manager.StartSessionStream(ctx, sessionID, uint(i), "typing")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetStreamingStats()
	}
}

// TestSessionStreamingConcurrency verifies thread-safe access
func TestSessionStreamingConcurrency(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	manager := NewSessionStreamingManager(hub)

	ctx := context.Background()

	// Start multiple sessions
	for i := 0; i < 5; i++ {
		sessionID := "session_" + string(rune('0'+i))
		_ = manager.StartSessionStream(ctx, sessionID, uint(i), "typing")
	}

	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func(id int) {
			sessionID := "session_" + string(rune('0'+id))
			for j := 0; j < 20; j++ {
				_ = manager.UpdateProgress(ctx, sessionID, "wpm", float64(100+j), "")
				_ = manager.GetSessionProgress(sessionID)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	// If no panic, test passes
}
