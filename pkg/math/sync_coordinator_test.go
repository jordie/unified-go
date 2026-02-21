package math

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestNewSyncCoordinator tests coordinator initialization
func TestNewSyncCoordinator(t *testing.T) {
	coord := NewSyncCoordinator()
	if coord == nil {
		t.Error("Expected coordinator, got nil")
	}
}

// TestSubscribeToEvents tests event subscription
func TestSubscribeToEvents(t *testing.T) {
	coord := NewSyncCoordinator()
	ctx := context.Background()

	subscriber := make(chan *SyncEvent, 10)
	id := coord.SubscribeToEvents(ctx, subscriber, "test_app")

	if id == "" {
		t.Error("Expected subscription ID, got empty string")
	}

	coord.UnsubscribeFromEvents(id)
}

// TestUnsubscribeFromEvents tests unsubscribe functionality
func TestUnsubscribeFromEvents(t *testing.T) {
	coord := NewSyncCoordinator()
	ctx := context.Background()

	subscriber := make(chan *SyncEvent, 10)
	id := coord.SubscribeToEvents(ctx, subscriber, "test_app")

	result := coord.UnsubscribeFromEvents(id)
	if !result {
		t.Error("Expected successful unsubscribe")
	}

	// Unsubscribe again should fail
	result = coord.UnsubscribeFromEvents(id)
	if result {
		t.Error("Expected unsubscribe to fail for already unsubscribed")
	}
}

// TestBroadcastToSubscribers tests broadcasting to multiple subscribers
func TestBroadcastToSubscribers(t *testing.T) {
	coord := NewSyncCoordinator()
	ctx := context.Background()

	// Create 3 subscribers
	sub1 := make(chan *SyncEvent, 10)
	sub2 := make(chan *SyncEvent, 10)
	sub3 := make(chan *SyncEvent, 10)

	id1 := coord.SubscribeToEvents(ctx, sub1, "app1")
	id2 := coord.SubscribeToEvents(ctx, sub2, "app2")
	id3 := coord.SubscribeToEvents(ctx, sub3, "app3")

	event := &SyncEvent{
		EventType: "test_broadcast",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"value": 42},
	}

	coord.BroadcastToSubscribers(event)

	// All should receive the event
	select {
	case e := <-sub1:
		if e.EventType != "test_broadcast" {
			t.Errorf("Expected test_broadcast, got %s", e.EventType)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Subscriber 1 did not receive broadcast")
	}

	select {
	case e := <-sub2:
		if e.EventType != "test_broadcast" {
			t.Errorf("Expected test_broadcast, got %s", e.EventType)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Subscriber 2 did not receive broadcast")
	}

	select {
	case e := <-sub3:
		if e.EventType != "test_broadcast" {
			t.Errorf("Expected test_broadcast, got %s", e.EventType)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Subscriber 3 did not receive broadcast")
	}

	coord.UnsubscribeFromEvents(id1)
	coord.UnsubscribeFromEvents(id2)
	coord.UnsubscribeFromEvents(id3)
}

// TestMetricNormalization tests normalizing metrics from different apps
func TestMetricNormalization(t *testing.T) {
	coord := NewSyncCoordinator()

	// Metrics from different apps with different scales
	metrics1 := map[string]interface{}{
		"accuracy": 95.5,
		"count":    100,
	}

	metrics2 := map[string]interface{}{
		"accuracy": 0.885,  // Different scale (0-1 instead of 0-100)
		"count":    50,
	}

	normalized1 := coord.NormalizeMetrics("app1", metrics1)
	normalized2 := coord.NormalizeMetrics("app2", metrics2)

	if normalized1 == nil || normalized2 == nil {
		t.Error("Expected normalized metrics, got nil")
	}
}

// TestScheduleSync tests scheduling synchronization tasks
func TestScheduleSync(t *testing.T) {
	coord := NewSyncCoordinator()
	ctx := context.Background()

	taskID := "sync_task_1"
	interval := 100 * time.Millisecond

	result := coord.ScheduleSync(ctx, taskID, interval)
	if !result {
		t.Error("Expected successful schedule")
	}

	// Give it time to execute
	time.Sleep(150 * time.Millisecond)

	coord.CancelSync(taskID)
}

// TestCancelSync tests cancelling scheduled syncs
func TestCancelSync(t *testing.T) {
	coord := NewSyncCoordinator()
	ctx := context.Background()

	taskID := "sync_task_cancel"
	interval := 100 * time.Millisecond

	coord.ScheduleSync(ctx, taskID, interval)
	result := coord.CancelSync(taskID)

	if !result {
		t.Error("Expected successful cancel")
	}

	// Cancel again should fail
	result = coord.CancelSync(taskID)
	if result {
		t.Error("Expected cancel to fail for already cancelled task")
	}
}

// TestSyncTimeout tests timeout handling
func TestSyncTimeout(t *testing.T) {
	coord := NewSyncCoordinator()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	taskID := "sync_timeout"
	interval := 1 * time.Second // Long interval will timeout

	result := coord.ScheduleSync(ctx, taskID, interval)

	// Should timeout before actually scheduling
	time.Sleep(100 * time.Millisecond)
	if result {
		t.Error("Expected timeout during schedule")
	}
}

// TestErrorRecovery tests error recovery mechanisms
func TestErrorRecovery(t *testing.T) {
	coord := NewSyncCoordinator()

	event := &SyncEvent{
		EventType: "error_test",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{},
	}

	// Attempt to recover from error
	recovery := coord.RecoverFromError(event)
	if recovery == nil {
		t.Error("Expected recovery strategy, got nil")
	}
}

// TestQueueManagement tests internal queue management
func TestQueueManagement(t *testing.T) {
	coord := NewSyncCoordinator()

	event := &SyncEvent{
		EventType: "queue_test",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"msg": "test"},
	}

	// Add to queue
	coord.EnqueueEvent(event)

	// Retrieve from queue
	retrieved := coord.DequeueEvent()
	if retrieved == nil {
		t.Error("Expected event from queue, got nil")
	}

	if retrieved.EventType != "queue_test" {
		t.Errorf("Expected queue_test, got %s", retrieved.EventType)
	}
}

// TestConcurrentSubscription tests 100+ concurrent subscriptions
func TestConcurrentSubscription(t *testing.T) {
	coord := NewSyncCoordinator()
	ctx := context.Background()
	var wg sync.WaitGroup
	numSubscribers := 120

	subscriptionIDs := make([]string, numSubscribers)
	var mu sync.Mutex

	for i := 0; i < numSubscribers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			subscriber := make(chan *SyncEvent, 10)
			subID := coord.SubscribeToEvents(ctx, subscriber, "concurrent_app")

			mu.Lock()
			subscriptionIDs[id] = subID
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify all subscriptions were created
	count := 0
	for _, id := range subscriptionIDs {
		if id != "" {
			count++
		}
	}

	if count != numSubscribers {
		t.Errorf("Expected %d subscriptions, got %d", numSubscribers, count)
	}

	// Unsubscribe all
	for _, id := range subscriptionIDs {
		if id != "" {
			coord.UnsubscribeFromEvents(id)
		}
	}
}

// TestConcurrentBroadcast tests broadcasting with 100+ subscribers
func TestConcurrentBroadcast(t *testing.T) {
	coord := NewSyncCoordinator()
	ctx := context.Background()
	var wg sync.WaitGroup
	numSubscribers := 120
	receivedCount := 0
	var mu sync.Mutex

	subscriptionIDs := make([]string, numSubscribers)

	// Create many subscribers
	for i := 0; i < numSubscribers; i++ {
		subscriber := make(chan *SyncEvent, 10)
		id := coord.SubscribeToEvents(ctx, subscriber, "broadcast_app")
		subscriptionIDs[i] = id

		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-subscriber:
				mu.Lock()
				receivedCount++
				mu.Unlock()
			case <-time.After(500 * time.Millisecond):
			}
		}()
	}

	// Broadcast event
	event := &SyncEvent{
		EventType: "broadcast_test",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{},
	}
	coord.BroadcastToSubscribers(event)

	wg.Wait()

	if receivedCount < numSubscribers/2 {
		t.Errorf("Expected at least %d receivers, got %d", numSubscribers/2, receivedCount)
	}

	// Cleanup
	for _, id := range subscriptionIDs {
		if id != "" {
			coord.UnsubscribeFromEvents(id)
		}
	}
}

// TestRaceConditions tests for data race detection
func TestRaceConditions(t *testing.T) {
	coord := NewSyncCoordinator()
	ctx := context.Background()
	var wg sync.WaitGroup

	// Mix of subscribes, broadcasts, unsubscribes
	for i := 0; i < 60; i++ {
		wg.Add(3)

		// Subscribe
		go func(id int) {
			defer wg.Done()
			subscriber := make(chan *SyncEvent, 10)
			coord.SubscribeToEvents(ctx, subscriber, "race_app")
		}(i)

		// Broadcast
		go func() {
			defer wg.Done()
			event := &SyncEvent{
				EventType: "race_test",
				Timestamp: time.Now(),
				Data:      map[string]interface{}{},
			}
			coord.BroadcastToSubscribers(event)
		}()

		// Get status
		go func() {
			defer wg.Done()
			coord.GetCoordinatorStatus(ctx)
		}()
	}

	wg.Wait()
}

// TestMemoryLeaks tests subscription memory management
func TestMemoryLeaks(t *testing.T) {
	coord := NewSyncCoordinator()
	ctx := context.Background()

	// Create and destroy many subscriptions
	for i := 0; i < 100; i++ {
		subscriber := make(chan *SyncEvent, 10)
		id := coord.SubscribeToEvents(ctx, subscriber, "memory_app")
		coord.UnsubscribeFromEvents(id)
	}

	// Check that subscriptions were properly cleaned up
	status := coord.GetCoordinatorStatus(ctx)
	if status.SubscriberCount > 10 {
		t.Errorf("Expected low subscriber count after cleanup, got %d", status.SubscriberCount)
	}
}

// TestEventOrdering tests ordered delivery of events
func TestEventOrdering(t *testing.T) {
	coord := NewSyncCoordinator()
	ctx := context.Background()

	subscriber := make(chan *SyncEvent, 50)
	id := coord.SubscribeToEvents(ctx, subscriber, "order_app")

	// Send events in order
	for i := 0; i < 10; i++ {
		event := &SyncEvent{
			EventType: "ordered",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"sequence": i},
		}
		coord.BroadcastToSubscribers(event)
	}

	// Receive and verify order
	for i := 0; i < 10; i++ {
		select {
		case event := <-subscriber:
			if event.Data["sequence"] != i {
				t.Errorf("Expected sequence %d, got %v", i, event.Data["sequence"])
			}
		case <-time.After(500 * time.Millisecond):
			t.Errorf("Timeout waiting for event %d", i)
		}
	}

	coord.UnsubscribeFromEvents(id)
}

// BenchmarkSubscribe benchmarks subscription performance
func BenchmarkSubscribe(b *testing.B) {
	coord := NewSyncCoordinator()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subscriber := make(chan *SyncEvent, 10)
		id := coord.SubscribeToEvents(ctx, subscriber, "bench_app")
		coord.UnsubscribeFromEvents(id)
	}
}

// BenchmarkBroadcast benchmarks broadcasting performance
func BenchmarkBroadcast(b *testing.B) {
	coord := NewSyncCoordinator()
	ctx := context.Background()

	// Pre-create subscribers
	for i := 0; i < 10; i++ {
		subscriber := make(chan *SyncEvent, 100)
		coord.SubscribeToEvents(ctx, subscriber, "bench_app")
	}

	event := &SyncEvent{
		EventType: "benchmark",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"value": 42},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		coord.BroadcastToSubscribers(event)
	}
}

// BenchmarkConcurrentBroadcast benchmarks parallel broadcast performance
func BenchmarkConcurrentBroadcast(b *testing.B) {
	coord := NewSyncCoordinator()
	ctx := context.Background()

	// Pre-create subscribers
	for i := 0; i < 10; i++ {
		subscriber := make(chan *SyncEvent, 100)
		coord.SubscribeToEvents(ctx, subscriber, "bench_app")
	}

	event := &SyncEvent{
		EventType: "benchmark",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"value": 42},
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			coord.BroadcastToSubscribers(event)
		}
	})
}

// BenchmarkMetricNormalization benchmarks normalization performance
func BenchmarkMetricNormalization(b *testing.B) {
	coord := NewSyncCoordinator()

	metrics := map[string]interface{}{
		"accuracy": 95.5,
		"count":    100,
		"xp":       1500,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		coord.NormalizeMetrics("app1", metrics)
	}
}
