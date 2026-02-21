package math

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestNewCrossAppSyncManager tests initialization
func TestNewCrossAppSyncManager(t *testing.T) {
	manager := NewCrossAppSyncManager("math", nil)
	if manager == nil {
		t.Error("Expected manager, got nil")
	}
	if manager.AppID != "math" {
		t.Errorf("Expected AppID 'math', got %s", manager.AppID)
	}
}

// TestRecordEvent tests event recording
func TestRecordEvent(t *testing.T) {
	manager := NewCrossAppSyncManager("math", nil)
	event := &SyncEvent{
		EventType: "question_answered",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"user_id": 1,
			"score":   95,
		},
	}

	err := manager.RecordEvent(context.Background(), event)
	if err != nil {
		t.Errorf("Failed to record event: %v", err)
	}
}

// TestRecordEventWithNilEvent tests error handling
func TestRecordEventWithNilEvent(t *testing.T) {
	manager := NewCrossAppSyncManager("math", nil)
	err := manager.RecordEvent(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil event")
	}
}

// TestGetSyncStatus tests status retrieval
func TestGetSyncStatus(t *testing.T) {
	manager := NewCrossAppSyncManager("math", nil)
	status := manager.GetSyncStatus(context.Background())
	if status == nil {
		t.Error("Expected status, got nil")
	}
}

// TestEventTransformation tests event data transformation
func TestEventTransformation(t *testing.T) {
	manager := NewCrossAppSyncManager("math", nil)
	event := &SyncEvent{
		EventType: "user_achievement",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"achievement": "100_perfect_streak",
			"xp":          500,
		},
	}

	transformed := manager.TransformEventData(event)
	if transformed == nil {
		t.Error("Expected transformed data")
	}
}

// TestMetricFiltering tests metric filtering
func TestMetricFiltering(t *testing.T) {
	manager := NewCrossAppSyncManager("math", nil)
	events := []*SyncEvent{
		{EventType: "question_answered", Data: map[string]interface{}{"score": 90}},
		{EventType: "level_up", Data: map[string]interface{}{"level": 5}},
		{EventType: "question_answered", Data: map[string]interface{}{"score": 85}},
	}

	filtered := manager.FilterMetrics(events, "question_answered")
	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered events, got %d", len(filtered))
	}
}

// TestConcurrentEventRecording tests concurrent recording
func TestConcurrentEventRecording(t *testing.T) {
	manager := NewCrossAppSyncManager("math", nil)
	var wg sync.WaitGroup
	ctx := context.Background()

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			event := &SyncEvent{
				EventType: "question_answered",
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"user_id": id,
					"score":   id * 10 % 100,
				},
			}
			_ = manager.RecordEvent(ctx, event)
		}(i)
	}

	wg.Wait()

	status := manager.GetSyncStatus(ctx)
	if status == nil {
		t.Error("Expected status after concurrent recording")
	}
}

// TestConcurrentWithHighGoroutines tests with 100+ goroutines
func TestConcurrentWithHighGoroutines(t *testing.T) {
	manager := NewCrossAppSyncManager("math", nil)
	var wg sync.WaitGroup
	ctx := context.Background()
	numGoroutines := 150

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				event := &SyncEvent{
					EventType: "practice_session",
					Timestamp: time.Now(),
					Data: map[string]interface{}{
						"user_id":    id,
						"session_id": j,
						"duration":   30 + j,
					},
				}
				_ = manager.RecordEvent(ctx, event)
			}
		}(i)
	}

	wg.Wait()
}

// TestEventBatching tests batching of events
func TestEventBatching(t *testing.T) {
	manager := NewCrossAppSyncManager("math", nil)
	events := make([]*SyncEvent, 0)
	for i := 0; i < 100; i++ {
		events = append(events, &SyncEvent{
			EventType: "question_answered",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"user_id": i,
			},
		})
	}

	batches := manager.BatchEvents(events, 25)
	if len(batches) != 4 {
		t.Errorf("Expected 4 batches, got %d", len(batches))
	}
}

// TestEventDeduplication tests duplicate removal
func TestEventDeduplication(t *testing.T) {
	manager := NewCrossAppSyncManager("math", nil)
	event := &SyncEvent{
		EventType: "user_login",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"user_id": 1},
	}

	events := []*SyncEvent{event, event, event}
	deduped := manager.DeduplicateEvents(events)
	if len(deduped) > 1 {
		t.Errorf("Expected 1 event after deduplication, got %d", len(deduped))
	}
}

// TestEventOrdering tests event ordering by timestamp
func TestEventOrdering(t *testing.T) {
	manager := NewCrossAppSyncManager("math", nil)
	now := time.Now()
	events := []*SyncEvent{
		{EventType: "event1", Timestamp: now.Add(2 * time.Second)},
		{EventType: "event2", Timestamp: now.Add(1 * time.Second)},
		{EventType: "event3", Timestamp: now},
	}

	ordered := manager.OrderEventsByTimestamp(events)
	if ordered[0].EventType != "event3" {
		t.Errorf("Expected first event to be event3, got %s", ordered[0].EventType)
	}
}

// TestSyncErrorHandling tests error handling in sync
func TestSyncErrorHandling(t *testing.T) {
	manager := NewCrossAppSyncManager("math", nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel context immediately

	event := &SyncEvent{
		EventType: "test",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{},
	}

	err := manager.RecordEvent(ctx, event)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

// BenchmarkRecordEvent benchmarks event recording
func BenchmarkRecordEvent(b *testing.B) {
	manager := NewCrossAppSyncManager("math", nil)
	event := &SyncEvent{
		EventType: "question_answered",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"score": 95},
	}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.RecordEvent(ctx, event)
	}
}

// BenchmarkGetStatus benchmarks status retrieval
func BenchmarkGetStatus(b *testing.B) {
	manager := NewCrossAppSyncManager("math", nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetSyncStatus(ctx)
	}
}

// BenchmarkConcurrentRecording benchmarks concurrent recording
func BenchmarkConcurrentRecording(b *testing.B) {
	manager := NewCrossAppSyncManager("math", nil)
	event := &SyncEvent{
		EventType: "question_answered",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"score": 95},
	}
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = manager.RecordEvent(ctx, event)
		}
	})
}

// BenchmarkEventTransformation benchmarks transformation
func BenchmarkEventTransformation(b *testing.B) {
	manager := NewCrossAppSyncManager("math", nil)
	event := &SyncEvent{
		EventType: "question_answered",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"score": 95, "user_id": 1},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.TransformEventData(event)
	}
}

// TestRaceConditions tests for race conditions
func TestRaceConditions(t *testing.T) {
	manager := NewCrossAppSyncManager("math", nil)
	var wg sync.WaitGroup
	ctx := context.Background()

	// Mix of reads and writes
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(id int) {
			defer wg.Done()
			event := &SyncEvent{
				EventType: "test",
				Timestamp: time.Now(),
				Data:      map[string]interface{}{"id": id},
			}
			_ = manager.RecordEvent(ctx, event)
		}(i)

		go func() {
			defer wg.Done()
			_ = manager.GetSyncStatus(ctx)
		}()
	}

	wg.Wait()
}

// TestEventDataIntegrity tests data integrity across sync
func TestEventDataIntegrity(t *testing.T) {
	manager := NewCrossAppSyncManager("math", nil)
	ctx := context.Background()

	originalData := map[string]interface{}{
		"user_id":  123,
		"score":    95,
		"duration": 45.5,
	}

	event := &SyncEvent{
		EventType: "question_answered",
		Timestamp: time.Now(),
		Data:      originalData,
	}

	_ = manager.RecordEvent(ctx, event)
	status := manager.GetSyncStatus(ctx)

	if status.EventCount != 1 {
		t.Errorf("Expected 1 event, got %d", status.EventCount)
	}
}

// TestStressTest performs stress testing
func TestStressTest(t *testing.T) {
	manager := NewCrossAppSyncManager("math", nil)
	ctx := context.Background()
	var wg sync.WaitGroup

	// 1000 events from 100 goroutines
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				event := &SyncEvent{
					EventType: "stress_test",
					Timestamp: time.Now(),
					Data: map[string]interface{}{
						"goroutine": id,
						"iteration": j,
					},
				}
				_ = manager.RecordEvent(ctx, event)
			}
		}(i)
	}

	wg.Wait()

	status := manager.GetSyncStatus(ctx)
	if status.EventCount < 100 {
		t.Errorf("Expected at least 100 events, got %d", status.EventCount)
	}
}
