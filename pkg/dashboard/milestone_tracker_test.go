package dashboard

import (
	"context"
	"testing"

	"github.com/jgirmay/unified-go/pkg/realtime"
)

func TestNewMilestoneTracker(t *testing.T) {
	hub := realtime.NewHub()
	tracker := NewMilestoneTracker(hub)

	if tracker == nil {
		t.Fatal("NewMilestoneTracker returned nil")
	}
	if tracker.hub != hub {
		t.Error("hub not set correctly")
	}
}

func TestCheckSessionMilestone(t *testing.T) {
	hub := realtime.NewHub()
	tracker := NewMilestoneTracker(hub)

	tests := []struct {
		sessionCount int
		expectedType MilestoneType
	}{
		{1, MilestoneFirstSession},
		{10, Milestone10Sessions},
		{50, Milestone50Sessions},
		{100, Milestone100Sessions},
		{500, Milestone500Sessions},
	}

	for _, tt := range tests {
		milestones := tracker.CheckSessionMilestone(123, "testuser", tt.sessionCount)
		if len(milestones) == 0 {
			t.Errorf("expected milestone for %d sessions", tt.sessionCount)
			continue
		}
		if milestones[len(milestones)-1].Type != tt.expectedType {
			t.Errorf("expected %s, got %s", tt.expectedType, milestones[len(milestones)-1].Type)
		}
	}
}

func TestCheckTimeMilestone(t *testing.T) {
	hub := realtime.NewHub()
	tracker := NewMilestoneTracker(hub)

	tests := []struct {
		minutes      int
		expectedType MilestoneType
	}{
		{600, MilestoneTotalHours10},   // 10 hours
		{3000, MilestoneTotalHours50},  // 50 hours
		{6000, MilestoneTotalHours100}, // 100 hours
	}

	for _, tt := range tests {
		milestones := tracker.CheckTimeMilestone(123, "testuser", tt.minutes, false)
		if len(milestones) == 0 {
			t.Errorf("expected milestone for %d minutes", tt.minutes)
			continue
		}
		if milestones[len(milestones)-1].Type != tt.expectedType {
			t.Errorf("expected %s, got %s", tt.expectedType, milestones[len(milestones)-1].Type)
		}
	}
}

func TestCheckProgressMilestone(t *testing.T) {
	hub := realtime.NewHub()
	tracker := NewMilestoneTracker(hub)

	// Test first improvement
	milestones := tracker.CheckProgressMilestone(123, "testuser", 100.0, 150.0)
	if len(milestones) == 0 {
		t.Error("expected milestone for first improvement")
	}
	if milestones[0].Type != MilestoneFirstImprovement {
		t.Errorf("expected MilestoneFirstImprovement, got %s", milestones[0].Type)
	}

	// Test double score
	milestones = tracker.CheckProgressMilestone(456, "testuser", 100.0, 200.0)
	if len(milestones) == 0 {
		t.Error("expected milestone for double score")
	}
	if milestones[len(milestones)-1].Type != MilestoneDoubleScore {
		t.Errorf("expected MilestoneDoubleScore, got %s", milestones[len(milestones)-1].Type)
	}
}

func TestCheckStreakMilestone(t *testing.T) {
	hub := realtime.NewHub()
	tracker := NewMilestoneTracker(hub)

	milestones := tracker.CheckStreakMilestone(123, "testuser", 10)
	if len(milestones) == 0 {
		t.Error("expected milestone for 10-day streak")
	}
	if milestones[0].Type != Milestone10Streak {
		t.Errorf("expected Milestone10Streak, got %s", milestones[0].Type)
	}
}

func TestBroadcastMilestone(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	tracker := NewMilestoneTracker(hub)

	ctx := context.Background()
	milestone := &Milestone{
		Type:        MilestoneFirstSession,
		Title:       "First Steps",
		Description: "Complete your first session",
		Icon:        "🎬",
		Reward:      10,
		Category:    "practice",
		UserID:      123,
		Username:    "testuser",
	}

	// Should not panic
	tracker.BroadcastMilestone(ctx, milestone)
}

func TestGetUnlockedMilestones(t *testing.T) {
	hub := realtime.NewHub()
	tracker := NewMilestoneTracker(hub)

	// Unlock some milestones
	_ = tracker.CheckSessionMilestone(123, "testuser", 1)
	_ = tracker.CheckSessionMilestone(123, "testuser", 10)

	unlocked := tracker.GetUnlockedMilestones(123)
	if len(unlocked) < 2 {
		t.Errorf("expected at least 2 unlocked, got %d", len(unlocked))
	}
}

func TestIsMilestoneUnlocked(t *testing.T) {
	hub := realtime.NewHub()
	tracker := NewMilestoneTracker(hub)

	// Initially not unlocked
	if tracker.IsMilestoneUnlocked(123, MilestoneFirstSession) {
		t.Error("milestone should not be unlocked yet")
	}

	// Unlock it
	_ = tracker.CheckSessionMilestone(123, "testuser", 1)

	// Now should be unlocked
	if !tracker.IsMilestoneUnlocked(123, MilestoneFirstSession) {
		t.Error("milestone should be unlocked")
	}
}

func TestGetMilestoneHistory(t *testing.T) {
	hub := realtime.NewHub()
	tracker := NewMilestoneTracker(hub)

	// Create some milestones
	_ = tracker.CheckSessionMilestone(123, "testuser", 1)
	_ = tracker.CheckSessionMilestone(123, "testuser", 10)
	_ = tracker.CheckSessionMilestone(123, "testuser", 50)

	history := tracker.GetMilestoneHistory(123, 0)
	if len(history) < 3 {
		t.Errorf("expected at least 3 milestones in history, got %d", len(history))
	}

	// Test limit
	limited := tracker.GetMilestoneHistory(123, 2)
	if len(limited) != 2 {
		t.Errorf("expected 2 milestones with limit, got %d", len(limited))
	}
}

func TestGetStats(t *testing.T) {
	hub := realtime.NewHub()
	tracker := NewMilestoneTracker(hub)

	// Unlock various milestones
	_ = tracker.CheckSessionMilestone(123, "testuser", 10)
	_ = tracker.CheckTimeMilestone(123, "testuser", 600, false)
	_ = tracker.CheckProgressMilestone(123, "testuser", 100.0, 150.0)

	stats := tracker.GetStats(123)

	if stats.TotalUnlocked < 3 {
		t.Errorf("expected at least 3 unlocked, got %d", stats.TotalUnlocked)
	}
	if stats.TotalRewardPoints <= 0 {
		t.Error("expected positive reward points")
	}
}

func TestNoDuplicateMilestones(t *testing.T) {
	hub := realtime.NewHub()
	tracker := NewMilestoneTracker(hub)

	// Try to unlock same milestone twice
	_ = tracker.CheckSessionMilestone(123, "testuser", 10)
	milestones := tracker.CheckSessionMilestone(123, "testuser", 10)

	if len(milestones) > 0 {
		t.Error("should not unlock same milestone twice")
	}
}

// BenchmarkCheckSessionMilestone benchmarks milestone checking
func BenchmarkCheckSessionMilestone(b *testing.B) {
	hub := realtime.NewHub()
	tracker := NewMilestoneTracker(hub)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tracker.CheckSessionMilestone(uint(i%1000), "user", i%500+1)
	}
}

// BenchmarkGetMilestoneHistory benchmarks history retrieval
func BenchmarkGetMilestoneHistory(b *testing.B) {
	hub := realtime.NewHub()
	tracker := NewMilestoneTracker(hub)

	// Pre-populate some milestones
	for i := 0; i < 100; i++ {
		_ = tracker.CheckSessionMilestone(123, "testuser", i+1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tracker.GetMilestoneHistory(123, 10)
	}
}

// TestMilestoneTrackerConcurrency verifies thread-safe access
func TestMilestoneTrackerConcurrency(t *testing.T) {
	hub := realtime.NewHub()
	tracker := NewMilestoneTracker(hub)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 20; j++ {
				_ = tracker.CheckSessionMilestone(uint(id), "user", j+1)
				_ = tracker.GetUnlockedMilestones(uint(id))
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
