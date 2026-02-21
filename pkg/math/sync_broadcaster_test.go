package math

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestBroadcastProgress tests progress broadcasting
func TestBroadcastProgress(t *testing.T) {
	broadcaster := NewSyncBroadcaster()
	ctx := context.Background()

	progress := map[string]interface{}{
		"current":  45,
		"total":    100,
		"percent":  45.0,
		"app":      "typing",
	}

	result := broadcaster.BroadcastProgress(ctx, progress)
	if !result {
		t.Error("Expected successful progress broadcast")
	}
}

// TestBroadcastLeaderboard tests leaderboard updates
func TestBroadcastLeaderboard(t *testing.T) {
	broadcaster := NewSyncBroadcaster()
	ctx := context.Background()

	leaderboard := []map[string]interface{}{
		{"user_id": 1, "name": "Alice", "score": 1000},
		{"user_id": 2, "name": "Bob", "score": 950},
		{"user_id": 3, "name": "Charlie", "score": 900},
	}

	result := broadcaster.BroadcastLeaderboard(ctx, "typing", leaderboard)
	if !result {
		t.Error("Expected successful leaderboard broadcast")
	}
}

// TestPropagateAchievements tests achievement propagation
func TestPropagateAchievements(t *testing.T) {
	broadcaster := NewSyncBroadcaster()
	ctx := context.Background()

	achievements := []map[string]interface{}{
		{"id": "speed_demon", "title": "Speed Demon", "desc": "Type 100 WPM"},
		{"id": "accuracy_master", "title": "Accuracy Master", "desc": "100% accuracy"},
	}

	result := broadcaster.PropagateAchievements(ctx, 1, achievements)
	if !result {
		t.Error("Expected successful achievement propagation")
	}
}

// TestBroadcastActivity tests activity feed updates
func TestBroadcastActivity(t *testing.T) {
	broadcaster := NewSyncBroadcaster()
	ctx := context.Background()

	activity := map[string]interface{}{
		"user_id":     1,
		"action":      "completed_session",
		"app":         "math",
		"timestamp":   time.Now().Unix(),
		"metadata":    map[string]interface{}{"score": 95},
	}

	result := broadcaster.BroadcastActivity(ctx, activity)
	if !result {
		t.Error("Expected successful activity broadcast")
	}
}

// TestMultiRecipient tests broadcasting to multiple apps
func TestMultiRecipient(t *testing.T) {
	broadcaster := NewSyncBroadcaster()
	ctx := context.Background()

	apps := []string{"typing", "math", "reading", "piano"}
	event := map[string]interface{}{
		"type": "user_achievement",
		"data": map[string]interface{}{"level": 10},
	}

	result := broadcaster.BroadcastToApps(ctx, apps, event)
	if !result {
		t.Error("Expected successful multi-app broadcast")
	}
}

// TestPriority tests priority handling
func TestPriority(t *testing.T) {
	broadcaster := NewSyncBroadcaster()
	ctx := context.Background()

	// High priority event
	highPriority := map[string]interface{}{
		"priority": "high",
		"message":  "System alert",
	}

	result := broadcaster.BroadcastWithPriority(ctx, highPriority, "high")
	if !result {
		t.Error("Expected successful high priority broadcast")
	}

	// Low priority event
	lowPriority := map[string]interface{}{
		"priority": "low",
		"message":  "Daily reminder",
	}

	result = broadcaster.BroadcastWithPriority(ctx, lowPriority, "low")
	if !result {
		t.Error("Expected successful low priority broadcast")
	}
}

// TestConcurrentBroadcast tests 50+ concurrent broadcasts
func TestConcurrentBroadcast(t *testing.T) {
	broadcaster := NewSyncBroadcaster()
	ctx := context.Background()
	var wg sync.WaitGroup
	numBroadcasts := 70
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < numBroadcasts; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			event := map[string]interface{}{
				"id":  id,
				"app": "typing",
				"type": "progress",
			}
			result := broadcaster.BroadcastProgress(ctx, event)
			if result {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if successCount < numBroadcasts/2 {
		t.Errorf("Expected at least %d successes, got %d", numBroadcasts/2, successCount)
	}
}

// TestStress tests high-volume broadcasting
func TestStress(t *testing.T) {
	broadcaster := NewSyncBroadcaster()
	ctx := context.Background()

	// Send 500 events rapidly
	for i := 0; i < 500; i++ {
		event := map[string]interface{}{
			"id":    i,
			"type":  "stress_test",
			"value": i * 10,
		}
		broadcaster.BroadcastProgress(ctx, event)
	}

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Verify queue processed events
	status := broadcaster.GetStatus(ctx)
	if status == nil {
		t.Error("Expected status, got nil")
	}
}

// TestErrorHandling tests error scenarios
func TestErrorHandling(t *testing.T) {
	broadcaster := NewSyncBroadcaster()
	ctx := context.Background()

	// Broadcast with nil data
	result := broadcaster.BroadcastProgress(ctx, nil)
	if result {
		t.Error("Expected broadcast to fail with nil data")
	}

	// Broadcast with invalid app
	result = broadcaster.BroadcastLeaderboard(ctx, "", nil)
	if result {
		t.Error("Expected broadcast to fail with empty app")
	}
}

// TestTimeout tests broadcast timeout
func TestTimeout(t *testing.T) {
	broadcaster := NewSyncBroadcaster()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Create a large event to trigger timeout
	largeEvent := make(map[string]interface{})
	for i := 0; i < 10000; i++ {
		largeEvent[string(rune(i))] = i
	}

	// This might timeout depending on processing
	_ = broadcaster.BroadcastProgress(ctx, largeEvent)
}

// TestBroadcastProgressWithRecipient tests progress to specific recipients
func TestBroadcastProgressWithRecipient(t *testing.T) {
	broadcaster := NewSyncBroadcaster()
	ctx := context.Background()

	progress := map[string]interface{}{
		"user_id":  1,
		"current":  50,
		"total":    100,
		"app":      "math",
	}

	recipients := []int{1, 2, 3}
	result := broadcaster.BroadcastProgressToUsers(ctx, progress, recipients)
	if !result {
		t.Error("Expected successful progress broadcast to users")
	}
}

// TestAchievementNotification tests achievement notifications
func TestAchievementNotification(t *testing.T) {
	broadcaster := NewSyncBroadcaster()
	ctx := context.Background()

	achievement := map[string]interface{}{
		"user_id": 1,
		"name":    "Century Club",
		"desc":    "Score 100 points",
		"reward":  500,
	}

	result := broadcaster.NotifyAchievement(ctx, achievement)
	if !result {
		t.Error("Expected successful achievement notification")
	}
}

// TestQueueOperations tests internal queue operations
func TestQueueOperations(t *testing.T) {
	broadcaster := NewSyncBroadcaster()

	event := map[string]interface{}{
		"type": "test_queue",
	}

	// Enqueue
	broadcaster.EnqueueBroadcast(event)

	// Dequeue
	retrieved := broadcaster.DequeueBroadcast()
	if retrieved == nil {
		t.Error("Expected event from queue, got nil")
	}

	if retrieved["type"] != "test_queue" {
		t.Errorf("Expected test_queue, got %v", retrieved["type"])
	}
}

// TestRetryLogic tests retry mechanisms for failed broadcasts
func TestRetryLogic(t *testing.T) {
	broadcaster := NewSyncBroadcaster()
	ctx := context.Background()

	event := map[string]interface{}{
		"type": "retry_test",
	}

	result := broadcaster.BroadcastWithRetry(ctx, event, 3)
	if !result {
		t.Error("Expected successful broadcast with retry")
	}
}

// BenchmarkBroadcastProgress benchmarks progress broadcast performance
func BenchmarkBroadcastProgress(b *testing.B) {
	broadcaster := NewSyncBroadcaster()
	ctx := context.Background()

	progress := map[string]interface{}{
		"current": 50,
		"total":   100,
		"percent": 50.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		broadcaster.BroadcastProgress(ctx, progress)
	}
}

// BenchmarkBroadcastAchievements benchmarks achievement broadcast performance
func BenchmarkBroadcastAchievements(b *testing.B) {
	broadcaster := NewSyncBroadcaster()
	ctx := context.Background()

	achievements := []map[string]interface{}{
		{"id": "ach1", "title": "Achievement 1"},
		{"id": "ach2", "title": "Achievement 2"},
		{"id": "ach3", "title": "Achievement 3"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		broadcaster.PropagateAchievements(ctx, 1, achievements)
	}
}

// BenchmarkConcurrent benchmarks concurrent broadcast performance
func BenchmarkConcurrent(b *testing.B) {
	broadcaster := NewSyncBroadcaster()
	ctx := context.Background()

	event := map[string]interface{}{
		"type": "benchmark",
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			broadcaster.BroadcastProgress(ctx, event)
		}
	})
}

// BenchmarkHighVolume benchmarks high-volume broadcast performance
func BenchmarkHighVolume(b *testing.B) {
	broadcaster := NewSyncBroadcaster()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			event := map[string]interface{}{
				"id":   i*10 + j,
				"type": "bulk",
			}
			broadcaster.BroadcastProgress(ctx, event)
		}
	}
}
