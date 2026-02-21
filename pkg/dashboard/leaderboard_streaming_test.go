package dashboard

import (
	"context"
	"testing"
	"time"

	"github.com/jgirmay/unified-go/pkg/realtime"
)

func TestNewLeaderboardStreamingManager(t *testing.T) {
	hub := realtime.NewHub()
	service := NewLeaderboardService(nil)

	manager := NewLeaderboardStreamingManager(hub, service)

	if manager == nil {
		t.Fatal("NewLeaderboardStreamingManager returned nil")
	}
	if manager.hub != hub {
		t.Error("hub not set correctly")
	}
	if manager.leaderboardService != service {
		t.Error("leaderboardService not set correctly")
	}
	if manager.rankTracker == nil {
		t.Error("rankTracker not initialized")
	}
	if manager.velocityAnalyzer == nil {
		t.Error("velocityAnalyzer not initialized")
	}
}

func TestStartLeaderboardStream(t *testing.T) {
	hub := realtime.NewHub()
	service := NewLeaderboardService(nil)
	manager := NewLeaderboardStreamingManager(hub, service)

	err := manager.StartLeaderboardStream("typing_wpm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify stream started
	session := manager.GetStreamingSession("typing_wpm")
	if session == nil {
		t.Fatal("streaming session not created")
	}
	if session.Category != "typing_wpm" {
		t.Errorf("expected category typing_wpm, got %s", session.Category)
	}
}

func TestStartLeaderboardStreamDuplicate(t *testing.T) {
	hub := realtime.NewHub()
	service := NewLeaderboardService(nil)
	manager := NewLeaderboardStreamingManager(hub, service)

	// Start first stream
	err := manager.StartLeaderboardStream("typing_wpm")
	if err != nil {
		t.Fatalf("unexpected error on first start: %v", err)
	}

	// Try to start duplicate stream
	err = manager.StartLeaderboardStream("typing_wpm")
	if err == nil {
		t.Error("expected error when starting duplicate stream")
	}
}

func TestStopLeaderboardStream(t *testing.T) {
	hub := realtime.NewHub()
	service := NewLeaderboardService(nil)
	manager := NewLeaderboardStreamingManager(hub, service)

	// Start stream
	_ = manager.StartLeaderboardStream("typing_wpm")

	// Stop stream
	manager.StopLeaderboardStream("typing_wpm")

	// Verify stream stopped
	session := manager.GetStreamingSession("typing_wpm")
	if session != nil {
		t.Error("streaming session should be nil after stop")
	}
}

func TestGetStreamingSession(t *testing.T) {
	hub := realtime.NewHub()
	service := NewLeaderboardService(nil)
	manager := NewLeaderboardStreamingManager(hub, service)

	// Non-existent session
	session := manager.GetStreamingSession("typing_wpm")
	if session != nil {
		t.Error("expected nil for non-existent session")
	}

	// Start stream
	_ = manager.StartLeaderboardStream("typing_wpm")

	// Get session
	session = manager.GetStreamingSession("typing_wpm")
	if session == nil {
		t.Fatal("expected session to exist")
	}
	if session.Category != "typing_wpm" {
		t.Errorf("expected category typing_wpm, got %s", session.Category)
	}
}

func TestGetActiveStreams(t *testing.T) {
	hub := realtime.NewHub()
	service := NewLeaderboardService(nil)
	manager := NewLeaderboardStreamingManager(hub, service)

	// Initially no streams
	streams := manager.GetActiveStreams()
	if len(streams) != 0 {
		t.Errorf("expected 0 streams, got %d", len(streams))
	}

	// Start multiple streams
	_ = manager.StartLeaderboardStream("typing_wpm")
	_ = manager.StartLeaderboardStream("math_accuracy")
	_ = manager.StartLeaderboardStream("piano_score")

	// Get active streams
	streams = manager.GetActiveStreams()
	if len(streams) != 3 {
		t.Errorf("expected 3 streams, got %d", len(streams))
	}
}

func TestUpdateSubscriberCount(t *testing.T) {
	hub := realtime.NewHub()
	service := NewLeaderboardService(nil)
	manager := NewLeaderboardStreamingManager(hub, service)

	// Start stream
	_ = manager.StartLeaderboardStream("typing_wpm")

	// Update subscriber count
	manager.UpdateSubscriberCount("typing_wpm", 42)

	// Verify count updated
	session := manager.GetStreamingSession("typing_wpm")
	if session.ActiveSubscribers != 42 {
		t.Errorf("expected 42 subscribers, got %d", session.ActiveSubscribers)
	}
}

func TestBroadcastScoreUpdate(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	service := NewLeaderboardService(nil)
	manager := NewLeaderboardStreamingManager(hub, service)

	// Start stream
	_ = manager.StartLeaderboardStream("typing_wpm")

	// Broadcast score update (should not panic)
	manager.broadcastScoreUpdate("typing_wpm", 123, "testuser", 150.5, 5)

	// Test passes if no panic
}

func TestBroadcastRankChange(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	service := NewLeaderboardService(nil)
	manager := NewLeaderboardStreamingManager(hub, service)

	rankChange := &RankChange{
		UserID:       123,
		Category:     "typing_wpm",
		PreviousRank: 10,
		CurrentRank:  7,
		RankDelta:    3,
		Velocity:     2.5,
		Timestamp:    time.Now(),
		IsPromotion:  true,
	}

	// Broadcast rank change (should not panic)
	manager.broadcastRankChange(rankChange)

	// Test passes if no panic
}

func TestCheckAndBroadcastMilestones(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	service := NewLeaderboardService(nil)
	manager := NewLeaderboardStreamingManager(hub, service)

	tests := []struct {
		name       string
		change     *RankChange
		shouldTest bool
	}{
		{
			"top10_promotion",
			&RankChange{
				UserID:       123,
				Category:     "typing_wpm",
				PreviousRank: 15,
				CurrentRank:  8,
				IsPromotion:  true,
			},
			true,
		},
		{
			"top5_promotion",
			&RankChange{
				UserID:       456,
				Category:     "math_accuracy",
				PreviousRank: 10,
				CurrentRank:  3,
				IsPromotion:  true,
			},
			true,
		},
		{
			"first_place",
			&RankChange{
				UserID:       789,
				Category:     "piano_score",
				PreviousRank: 2,
				CurrentRank:  1,
				IsPromotion:  true,
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager.checkAndBroadcastMilestones(tt.change, "testuser")
			// Test passes if no panic
		})
	}
}

func TestGetStreamingStats(t *testing.T) {
	hub := realtime.NewHub()
	service := NewLeaderboardService(nil)
	manager := NewLeaderboardStreamingManager(hub, service)

	// Start multiple streams
	_ = manager.StartLeaderboardStream("typing_wpm")
	_ = manager.StartLeaderboardStream("math_accuracy")

	// Get stats
	stats := manager.GetStreamingStats()

	if stats.ActiveStreams != 2 {
		t.Errorf("expected 2 active streams, got %d", stats.ActiveStreams)
	}
}

func TestHandleLeaderboardEventNil(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	service := NewLeaderboardService(nil)
	manager := NewLeaderboardStreamingManager(hub, service)

	ctx := context.Background()

	// Test with nil event
	err := manager.HandleLeaderboardEvent(ctx, nil)
	if err == nil {
		t.Error("expected error for nil event")
	}
}

func TestFindUserRank(t *testing.T) {
	hub := realtime.NewHub()
	service := NewLeaderboardService(nil)
	manager := NewLeaderboardStreamingManager(hub, service)

	// Create test leaderboard
	lb := &realtime.LeaderboardMessage{
		Category: "typing_wpm",
		Entries: []realtime.LeaderboardEntry{
			{UserID: 1, Rank: 1, Username: "user1"},
			{UserID: 2, Rank: 2, Username: "user2"},
			{UserID: 3, Rank: 3, Username: "user3"},
		},
	}

	// Find existing user rank
	rank := manager.findUserRank(lb, 2)
	if rank != 2 {
		t.Errorf("expected rank 2, got %d", rank)
	}

	// Find non-existent user
	rank = manager.findUserRank(lb, 999)
	if rank != 0 {
		t.Errorf("expected rank 0 for non-existent user, got %d", rank)
	}
}

// BenchmarkProcessScoreUpdate benchmarks score update processing
func BenchmarkProcessScoreUpdate(b *testing.B) {
	hub := realtime.NewHub()
	go hub.Run()
	service := NewLeaderboardService(nil)
	manager := NewLeaderboardStreamingManager(hub, service)

	ctx := context.Background()

	// Start stream
	_ = manager.StartLeaderboardStream("typing_wpm")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Note: This will error since we don't have real leaderboard data
		// But it benchmarks the processing overhead
		_ = manager.ProcessScoreUpdate(ctx, uint(i%100), "user", "typing", "typing_wpm", 150.0)
	}
}

// BenchmarkBroadcastRankChange benchmarks broadcasting rank changes
func BenchmarkBroadcastRankChange(b *testing.B) {
	hub := realtime.NewHub()
	go hub.Run()
	service := NewLeaderboardService(nil)
	manager := NewLeaderboardStreamingManager(hub, service)

	rankChange := &RankChange{
		UserID:       123,
		Category:     "typing_wpm",
		PreviousRank: 10,
		CurrentRank:  7,
		RankDelta:    3,
		Velocity:     2.5,
		Timestamp:    time.Now(),
		IsPromotion:  true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.broadcastRankChange(rankChange)
	}
}

// BenchmarkGetStreamingStats benchmarks stats retrieval
func BenchmarkGetStreamingStats(b *testing.B) {
	hub := realtime.NewHub()
	service := NewLeaderboardService(nil)
	manager := NewLeaderboardStreamingManager(hub, service)

	// Start multiple streams
	for i := 0; i < 10; i++ {
		_ = manager.StartLeaderboardStream("category_" + string(rune(i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetStreamingStats()
	}
}

// TestStreamingSessionConcurrency verifies thread-safe access
func TestStreamingSessionConcurrency(t *testing.T) {
	hub := realtime.NewHub()
	service := NewLeaderboardService(nil)
	manager := NewLeaderboardStreamingManager(hub, service)

	// Start stream
	_ = manager.StartLeaderboardStream("typing_wpm")

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				manager.UpdateSubscriberCount("typing_wpm", id*10+j)
				_ = manager.GetStreamingSession("typing_wpm")
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// If no panic, test passes
}
