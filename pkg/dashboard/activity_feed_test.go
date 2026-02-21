package dashboard

import (
	"testing"
	"time"

	"github.com/jgirmay/unified-go/pkg/events"
	"github.com/jgirmay/unified-go/pkg/realtime"
)

func TestNewActivityFeed(t *testing.T) {
	hub := realtime.NewHub()
	eventBus := events.NewBus(100)
	feed := NewActivityFeed(hub, eventBus)

	if feed == nil {
		t.Fatal("NewActivityFeed returned nil")
	}
	if feed.hub != hub {
		t.Error("hub not set correctly")
	}
	if feed.eventBus != eventBus {
		t.Error("eventBus not set correctly")
	}
}

func TestRecordActivity(t *testing.T) {
	hub := realtime.NewHub()
	eventBus := events.NewBus(100)
	feed := NewActivityFeed(hub, eventBus)

	activity := &Activity{
		Type:     ActivityAchievementUnlocked,
		UserID:   123,
		Username: "testuser",
		App:      "typing",
		Title:    "First Achievement",
		Icon:     "🏆",
	}

	feed.RecordActivity(activity)

	userActivities := feed.GetUserActivity(123, nil)
	if len(userActivities) == 0 {
		t.Fatal("activity not recorded")
	}
	if userActivities[0].Type != ActivityAchievementUnlocked {
		t.Errorf("expected ActivityAchievementUnlocked, got %s", userActivities[0].Type)
	}
}

func TestGetUserActivity(t *testing.T) {
	hub := realtime.NewHub()
	eventBus := events.NewBus(100)
	feed := NewActivityFeed(hub, eventBus)

	// Record multiple activities
	for i := 0; i < 5; i++ {
		activity := &Activity{
			Type:     ActivityMilestoneUnlocked,
			UserID:   123,
			Username: "testuser",
			App:      "math",
			Title:    "Milestone",
		}
		feed.RecordActivity(activity)
	}

	activities := feed.GetUserActivity(123, &ActivityFilter{Limit: 10})
	if len(activities) != 5 {
		t.Errorf("expected 5 activities, got %d", len(activities))
	}
}

func TestGetGlobalActivity(t *testing.T) {
	hub := realtime.NewHub()
	eventBus := events.NewBus(100)
	feed := NewActivityFeed(hub, eventBus)

	// Record activities from different users
	for i := 0; i < 3; i++ {
		activity := &Activity{
			Type:     ActivitySessionEnded,
			UserID:   uint(100 + i),
			Username: "user" + string(rune('0'+i)),
			App:      "typing",
		}
		feed.RecordActivity(activity)
	}

	activities := feed.GetGlobalActivity(&ActivityFilter{Limit: 10})
	if len(activities) != 3 {
		t.Errorf("expected 3 activities, got %d", len(activities))
	}
}

func TestActivityFilter(t *testing.T) {
	hub := realtime.NewHub()
	eventBus := events.NewBus(100)
	feed := NewActivityFeed(hub, eventBus)

	// Record mixed activities
	for i := 0; i < 5; i++ {
		actType := ActivityAchievementUnlocked
		if i%2 == 0 {
			actType = ActivityMilestoneUnlocked
		}
		activity := &Activity{
			Type:     actType,
			UserID:   123,
			Username: "testuser",
			App:      "typing",
		}
		feed.RecordActivity(activity)
	}

	// Filter by type
	filter := &ActivityFilter{
		ActivityType: []ActivityType{ActivityAchievementUnlocked},
		Limit:        10,
	}
	activities := feed.GetUserActivity(123, filter)
	if len(activities) != 2 {
		t.Errorf("expected 2 achievement activities, got %d", len(activities))
	}

	// Verify all are correct type
	for _, activity := range activities {
		if activity.Type != ActivityAchievementUnlocked {
			t.Errorf("expected ActivityAchievementUnlocked, got %s", activity.Type)
		}
	}
}

func TestActivityPagination(t *testing.T) {
	hub := realtime.NewHub()
	eventBus := events.NewBus(100)
	feed := NewActivityFeed(hub, eventBus)

	// Record 10 activities
	for i := 0; i < 10; i++ {
		activity := &Activity{
			Type:     ActivitySessionEnded,
			UserID:   123,
			Username: "testuser",
			App:      "typing",
		}
		feed.RecordActivity(activity)
		time.Sleep(1 * time.Millisecond)
	}

	// Test pagination
	filter := &ActivityFilter{Limit: 3, Offset: 0}
	page1 := feed.GetUserActivity(123, filter)
	if len(page1) != 3 {
		t.Errorf("expected 3 items on page 1, got %d", len(page1))
	}

	filter.Offset = 3
	page2 := feed.GetUserActivity(123, filter)
	if len(page2) != 3 {
		t.Errorf("expected 3 items on page 2, got %d", len(page2))
	}
}

func TestActivityCount(t *testing.T) {
	hub := realtime.NewHub()
	eventBus := events.NewBus(100)
	feed := NewActivityFeed(hub, eventBus)

	// Record activities
	for i := 0; i < 5; i++ {
		activity := &Activity{
			Type:     ActivityAchievementUnlocked,
			UserID:   123,
			Username: "testuser",
			App:      "typing",
		}
		feed.RecordActivity(activity)
	}

	count := feed.GetActivityCount(nil, nil)
	if count != 5 {
		t.Errorf("expected 5 global activities, got %d", count)
	}

	userID := uint(123)
	count = feed.GetActivityCount(&userID, nil)
	if count != 5 {
		t.Errorf("expected 5 user activities, got %d", count)
	}
}

func TestActivityFeedStats(t *testing.T) {
	hub := realtime.NewHub()
	eventBus := events.NewBus(100)
	feed := NewActivityFeed(hub, eventBus)

	// Record different activity types
	types := []ActivityType{
		ActivityAchievementUnlocked,
		ActivityMilestoneUnlocked,
		ActivityRankChanged,
		ActivitySessionEnded,
	}

	for i, actType := range types {
		for j := 0; j < 2; j++ {
			activity := &Activity{
				Type:     actType,
				UserID:   uint(100 + i),
				Username: "user" + string(rune('0'+i)),
				App:      "typing",
			}
			feed.RecordActivity(activity)
		}
	}

	stats := feed.GetStats(nil)

	if stats.TotalActivities != 8 {
		t.Errorf("expected 8 total activities, got %d", stats.TotalActivities)
	}
	if stats.AchievementCount != 2 {
		t.Errorf("expected 2 achievements, got %d", stats.AchievementCount)
	}
	if stats.MilestoneCount != 2 {
		t.Errorf("expected 2 milestones, got %d", stats.MilestoneCount)
	}
	if stats.RankChangeCount != 2 {
		t.Errorf("expected 2 rank changes, got %d", stats.RankChangeCount)
	}
	if stats.SessionCount != 2 {
		t.Errorf("expected 2 sessions, got %d", stats.SessionCount)
	}
}

func TestClearUserActivity(t *testing.T) {
	hub := realtime.NewHub()
	eventBus := events.NewBus(100)
	feed := NewActivityFeed(hub, eventBus)

	activity := &Activity{
		Type:     ActivityAchievementUnlocked,
		UserID:   123,
		Username: "testuser",
		App:      "typing",
	}

	feed.RecordActivity(activity)

	// Verify recorded
	activities := feed.GetUserActivity(123, nil)
	if len(activities) != 1 {
		t.Fatal("activity not recorded")
	}

	// Clear
	feed.ClearUserActivity(123)

	// Verify cleared
	activities = feed.GetUserActivity(123, nil)
	if len(activities) != 0 {
		t.Error("activity not cleared")
	}
}

func TestClearGlobalActivity(t *testing.T) {
	hub := realtime.NewHub()
	eventBus := events.NewBus(100)
	feed := NewActivityFeed(hub, eventBus)

	// Record multiple activities
	for i := 0; i < 3; i++ {
		activity := &Activity{
			Type:     ActivitySessionEnded,
			UserID:   uint(100 + i),
			Username: "user" + string(rune('0'+i)),
			App:      "typing",
		}
		feed.RecordActivity(activity)
	}

	// Verify recorded
	activities := feed.GetGlobalActivity(nil)
	if len(activities) != 3 {
		t.Fatal("activities not recorded")
	}

	// Clear
	feed.ClearGlobalActivity()

	// Verify cleared
	activities = feed.GetGlobalActivity(nil)
	if len(activities) != 0 {
		t.Error("global activity not cleared")
	}
}

func TestActivityBroadcasting(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()
	eventBus := events.NewBus(100)
	feed := NewActivityFeed(hub, eventBus)

	activity := &Activity{
		Type:     ActivityAchievementUnlocked,
		UserID:   123,
		Username: "testuser",
		App:      "typing",
	}

	// Should not panic
	feed.RecordActivity(activity)
}

func TestActivityMetadata(t *testing.T) {
	hub := realtime.NewHub()
	eventBus := events.NewBus(100)
	feed := NewActivityFeed(hub, eventBus)

	activity := &Activity{
		Type:     ActivityRankChanged,
		UserID:   123,
		Username: "testuser",
		App:      "typing",
		Metadata: map[string]interface{}{
			"old_rank": 15,
			"new_rank": 10,
		},
	}

	feed.RecordActivity(activity)

	activities := feed.GetUserActivity(123, nil)
	if len(activities) == 0 {
		t.Fatal("activity not recorded")
	}

	if oldRank, ok := activities[0].Metadata["old_rank"]; ok {
		if oldRank != 15 {
			t.Errorf("expected old_rank 15, got %v", oldRank)
		}
	} else {
		t.Error("metadata not preserved")
	}
}

func TestActivityAppFilter(t *testing.T) {
	hub := realtime.NewHub()
	eventBus := events.NewBus(100)
	feed := NewActivityFeed(hub, eventBus)

	// Record activities from different apps
	apps := []string{"typing", "math", "reading"}
	for i, app := range apps {
		for j := 0; j < 2; j++ {
			activity := &Activity{
				Type:     ActivitySessionEnded,
				UserID:   uint(100 + i),
				Username: "user" + string(rune('0'+i)),
				App:      app,
			}
			feed.RecordActivity(activity)
		}
	}

	// Filter by app
	appStr := "math"
	filter := &ActivityFilter{
		App:   &appStr,
		Limit: 10,
	}

	activities := feed.GetGlobalActivity(filter)
	for _, activity := range activities {
		if activity.App != "math" {
			t.Errorf("expected math app, got %s", activity.App)
		}
	}
}

// BenchmarkRecordActivity benchmarks activity recording
func BenchmarkRecordActivity(b *testing.B) {
	hub := realtime.NewHub()
	eventBus := events.NewBus(100)
	feed := NewActivityFeed(hub, eventBus)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		activity := &Activity{
			Type:     ActivityAchievementUnlocked,
			UserID:   uint(i % 1000),
			Username: "user",
			App:      "typing",
		}
		feed.RecordActivity(activity)
	}
}

// BenchmarkActivityFeedGetUserActivity benchmarks user activity retrieval
func BenchmarkActivityFeedGetUserActivity(b *testing.B) {
	hub := realtime.NewHub()
	eventBus := events.NewBus(100)
	feed := NewActivityFeed(hub, eventBus)

	// Pre-populate activities
	for i := 0; i < 100; i++ {
		activity := &Activity{
			Type:     ActivitySessionEnded,
			UserID:   123,
			Username: "testuser",
			App:      "typing",
		}
		feed.RecordActivity(activity)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = feed.GetUserActivity(123, &ActivityFilter{Limit: 10})
	}
}

// BenchmarkGetStats benchmarks stats calculation
func BenchmarkActivityGetStats(b *testing.B) {
	hub := realtime.NewHub()
	eventBus := events.NewBus(100)
	feed := NewActivityFeed(hub, eventBus)

	// Pre-populate activities
	for i := 0; i < 100; i++ {
		activity := &Activity{
			Type:     ActivityAchievementUnlocked,
			UserID:   uint(i % 10),
			Username: "user" + string(rune('0'+(i%10))),
			App:      "typing",
		}
		feed.RecordActivity(activity)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = feed.GetStats(nil)
	}
}

// TestActivityFeedConcurrency verifies thread-safe access (without Hub.Run)
func TestActivityFeedConcurrency(t *testing.T) {
	hub := realtime.NewHub()
	// Don't call hub.Run() to avoid goroutine cleanup issues in tests
	eventBus := events.NewBus(100)
	feed := NewActivityFeed(hub, eventBus)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 20; j++ {
				activity := &Activity{
					Type:     ActivitySessionEnded,
					UserID:   uint(id),
					Username: "user",
					App:      "typing",
				}
				feed.RecordActivity(activity)
				_ = feed.GetUserActivity(uint(id), nil)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
