package dashboard

import (
	"context"
	"testing"

	"github.com/jgirmay/unified-go/pkg/realtime"
)

func TestNewAchievementNotifier(t *testing.T) {
	hub := realtime.NewHub()
	notifier := NewAchievementNotifier(hub)

	if notifier == nil {
		t.Fatal("NewAchievementNotifier returned nil")
	}
	if notifier.hub != hub {
		t.Error("hub not set correctly")
	}
}

func TestAchievementCheckStreakMilestone(t *testing.T) {
	hub := realtime.NewHub()
	notifier := NewAchievementNotifier(hub)

	// Test 7-day streak
	unlocks := notifier.CheckStreakMilestone(123, "testuser", 7)
	if len(unlocks) != 1 {
		t.Errorf("expected 1 unlock for 7-day streak, got %d", len(unlocks))
	}
	if unlocks[0].Achievement.Type != AchievementStreak7Days {
		t.Errorf("expected AchievementStreak7Days, got %s", unlocks[0].Achievement.Type)
	}

	// Test 30-day streak
	unlocks = notifier.CheckStreakMilestone(123, "testuser", 30)
	if len(unlocks) != 1 {
		t.Errorf("expected 1 unlock for 30-day streak, got %d", len(unlocks))
	}

	// Test 100-day streak
	unlocks = notifier.CheckStreakMilestone(123, "testuser", 100)
	if len(unlocks) != 1 {
		t.Errorf("expected 1 unlock for 100-day streak, got %d", len(unlocks))
	}
}

func TestCheckScoreMilestone(t *testing.T) {
	hub := realtime.NewHub()
	notifier := NewAchievementNotifier(hub)

	// Test score milestones
	tests := []struct {
		score          float64
		expectedType   AchievementType
	}{
		{100, AchievementScore100},
		{500, AchievementScore500},
		{1000, AchievementScore1000},
		{5000, AchievementScore5000},
		{10000, AchievementScore10000},
	}

	for _, tt := range tests {
		unlocks := notifier.CheckScoreMilestone(123, "testuser", tt.score, "typing")
		if len(unlocks) == 0 {
			t.Errorf("expected unlock for score %f", tt.score)
			continue
		}
		if unlocks[len(unlocks)-1].Achievement.Type != tt.expectedType {
			t.Errorf("expected %s, got %s", tt.expectedType, unlocks[len(unlocks)-1].Achievement.Type)
		}
	}
}

func TestCheckRankMilestone(t *testing.T) {
	hub := realtime.NewHub()
	notifier := NewAchievementNotifier(hub)

	// Test rank milestones
	tests := []struct {
		rank          int
		expectedType  AchievementType
	}{
		{10, AchievementRankTop10},
		{5, AchievementRankTop5},
		{1, AchievementRankFirst},
	}

	for _, tt := range tests {
		unlocks := notifier.CheckRankMilestone(123, "testuser", tt.rank, "typing_wpm")
		if len(unlocks) == 0 {
			t.Errorf("expected unlock for rank %d", tt.rank)
			continue
		}
		if unlocks[len(unlocks)-1].Achievement.Type != tt.expectedType {
			t.Errorf("expected %s, got %s", tt.expectedType, unlocks[len(unlocks)-1].Achievement.Type)
		}
	}
}

func TestCheckAccuracyMilestone(t *testing.T) {
	hub := realtime.NewHub()
	notifier := NewAchievementNotifier(hub)

	// Test perfect accuracy
	unlocks := notifier.CheckAccuracyMilestone(123, "testuser", 100.0)
	if len(unlocks) == 0 {
		t.Error("expected unlock for perfect accuracy")
	}

	// Test high accuracy
	unlocks = notifier.CheckAccuracyMilestone(456, "testuser", 95.0)
	if len(unlocks) == 0 {
		t.Error("expected unlock for 95% accuracy")
	}
}

func TestCheckConsistencyMilestone(t *testing.T) {
	hub := realtime.NewHub()
	notifier := NewAchievementNotifier(hub)

	unlocks := notifier.CheckConsistencyMilestone(123, "testuser", 10)
	if len(unlocks) == 0 {
		t.Error("expected unlock for 10 sessions")
	}
	if unlocks[0].Achievement.Type != AchievementConsistency {
		t.Errorf("expected AchievementConsistency, got %s", unlocks[0].Achievement.Type)
	}
}

func TestBroadcastAchievement(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	notifier := NewAchievementNotifier(hub)

	ctx := context.Background()
	unlock := &AchievementUnlock{
		UserID:   123,
		Username: "testuser",
		Achievement: &Achievement{
			Type:        AchievementScore100,
			Title:       "Test",
			Description: "Test achievement",
			Icon:        "🎯",
			Points:      25,
			Category:    "score",
		},
	}

	// Should not panic
	notifier.BroadcastAchievement(ctx, unlock)
	if !unlock.NotificationSent {
		t.Error("notification flag not set")
	}
}

func TestGetUnlockedAchievements(t *testing.T) {
	hub := realtime.NewHub()
	notifier := NewAchievementNotifier(hub)

	// Unlock some achievements
	notifier.CheckScoreMilestone(123, "testuser", 100.0, "typing")
	notifier.CheckStreakMilestone(123, "testuser", 7)

	unlocked := notifier.GetUnlockedAchievements(123)
	if len(unlocked) < 2 {
		t.Errorf("expected at least 2 unlocked, got %d", len(unlocked))
	}
}

func TestIsAchievementUnlocked(t *testing.T) {
	hub := realtime.NewHub()
	notifier := NewAchievementNotifier(hub)

	// Initially not unlocked
	if notifier.IsAchievementUnlocked(123, AchievementScore100) {
		t.Error("achievement should not be unlocked yet")
	}

	// Unlock it
	notifier.CheckScoreMilestone(123, "testuser", 100.0, "typing")

	// Now should be unlocked
	if !notifier.IsAchievementUnlocked(123, AchievementScore100) {
		t.Error("achievement should be unlocked")
	}
}

func TestAchievementGetStats(t *testing.T) {
	hub := realtime.NewHub()
	notifier := NewAchievementNotifier(hub)

	// Unlock various achievements
	notifier.CheckStreakMilestone(123, "testuser", 7)
	notifier.CheckScoreMilestone(123, "testuser", 500.0, "typing")
	notifier.CheckRankMilestone(123, "testuser", 10, "typing_wpm")

	stats := notifier.GetStats(123)

	if stats.TotalUnlocked < 3 {
		t.Errorf("expected at least 3 unlocked, got %d", stats.TotalUnlocked)
	}
	if stats.TotalPoints <= 0 {
		t.Error("expected positive total points")
	}
}

func TestNoDuplicateUnlocks(t *testing.T) {
	hub := realtime.NewHub()
	notifier := NewAchievementNotifier(hub)

	// Try to unlock same achievement twice
	_ = notifier.CheckScoreMilestone(123, "testuser", 100.0, "typing")
	unlocks := notifier.CheckScoreMilestone(123, "testuser", 150.0, "typing")

	if len(unlocks) > 0 {
		t.Error("should not unlock same achievement twice")
	}
}

// BenchmarkCheckScoreMilestone benchmarks milestone checking
func BenchmarkCheckScoreMilestone(b *testing.B) {
	hub := realtime.NewHub()
	notifier := NewAchievementNotifier(hub)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = notifier.CheckScoreMilestone(uint(i%1000), "user", float64(100+i%10000), "typing")
	}
}

// TestAchievementNotifierConcurrency verifies thread-safe access
func TestAchievementNotifierConcurrency(t *testing.T) {
	hub := realtime.NewHub()
	notifier := NewAchievementNotifier(hub)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 20; j++ {
				_ = notifier.CheckScoreMilestone(uint(id), "user", float64(100+j*50), "typing")
				_ = notifier.GetUnlockedAchievements(uint(id))
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
