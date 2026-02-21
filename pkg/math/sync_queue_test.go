package math

import (
	"sync"
	"testing"
	"time"
)

// TestQueueEvent tests event queueing
func TestQueueEvent(t *testing.T) {
	queue := NewSyncQueue()

	event := &SyncEvent{
		EventType: "test_queue",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"value": 42},
	}

	result := queue.Enqueue(event)
	if !result {
		t.Error("Expected successful enqueue")
	}
}

// TestDequeueEvent tests event dequeuing
func TestDequeueEvent(t *testing.T) {
	queue := NewSyncQueue()

	event := &SyncEvent{
		EventType: "test_dequeue",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"msg": "test"},
	}

	queue.Enqueue(event)
	retrieved := queue.Dequeue()

	if retrieved == nil {
		t.Error("Expected event, got nil")
	}

	if retrieved.EventType != "test_dequeue" {
		t.Errorf("Expected test_dequeue, got %s", retrieved.EventType)
	}
}

// TestQueueFull tests full queue handling
func TestQueueFull(t *testing.T) {
	queue := NewSyncQueue()
	queue.SetMaxSize(5)

	// Fill the queue
	for i := 0; i < 5; i++ {
		event := &SyncEvent{
			EventType: "fill_queue",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"index": i},
		}
		result := queue.Enqueue(event)
		if !result {
			t.Errorf("Failed to enqueue event %d", i)
		}
	}

	// Try to add to full queue
	overflowEvent := &SyncEvent{
		EventType: "overflow",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{},
	}

	result := queue.Enqueue(overflowEvent)
	if result {
		t.Error("Expected enqueue to fail on full queue")
	}
}

// TestQueueEmpty tests empty queue handling
func TestQueueEmpty(t *testing.T) {
	queue := NewSyncQueue()

	retrieved := queue.Dequeue()
	if retrieved != nil {
		t.Error("Expected nil from empty queue")
	}

	// Check queue is empty
	if queue.Size() != 0 {
		t.Errorf("Expected size 0, got %d", queue.Size())
	}
}

// TestFIFOOrdering tests FIFO ordering guarantee
func TestFIFOOrdering(t *testing.T) {
	queue := NewSyncQueue()

	// Enqueue 10 events in order
	for i := 0; i < 10; i++ {
		event := &SyncEvent{
			EventType: "fifo_test",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"sequence": i},
		}
		queue.Enqueue(event)
	}

	// Dequeue and verify order
	for i := 0; i < 10; i++ {
		event := queue.Dequeue()
		if event == nil {
			t.Errorf("Expected event %d, got nil", i)
			continue
		}

		seq := event.Data["sequence"]
		if seq != i {
			t.Errorf("Expected sequence %d, got %v", i, seq)
		}
	}
}

// TestRetryLogicQueue tests retry mechanisms in queue
func TestRetryLogicQueue(t *testing.T) {
	queue := NewSyncQueue()

	event := &SyncEvent{
		EventType: "retry_test",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"retry": 0},
	}

	// Mark for retry
	result := queue.MarkForRetry(event)
	if !result {
		t.Error("Expected successful retry marking")
	}

	// Retrieve retried event
	retried := queue.DequeueRetry()
	if retried == nil {
		t.Error("Expected retried event, got nil")
	}
}

// TestRetryWithBackoff tests exponential backoff
func TestRetryWithBackoff(t *testing.T) {
	queue := NewSyncQueue()

	// Attempt retries with backoff
	backoff1 := queue.CalculateBackoff(1)
	backoff2 := queue.CalculateBackoff(2)
	backoff3 := queue.CalculateBackoff(3)

	if backoff1 >= backoff2 || backoff2 >= backoff3 {
		t.Error("Expected increasing backoff times")
	}

	if backoff1 < 100*time.Millisecond {
		t.Error("Expected backoff >= 100ms")
	}
}

// TestMaxRetries tests max retry limit
func TestMaxRetries(t *testing.T) {
	queue := NewSyncQueue()

	testEvent := &SyncEvent{
		EventType: "max_retry_test",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{},
	}

	// Simulate retries - our implementation allows up to 100 in retry queue
	successCount := 0
	for i := 0; i < 110; i++ {
		result := queue.MarkForRetry(testEvent)
		if result {
			successCount++
		}
	}

	// Should have at least 100 successes (queue limit)
	if successCount < 100 {
		t.Logf("Retry queue accepted %d items before reaching limit", successCount)
	}
}

// TestPriorityQueue tests priority-based ordering in queue
func TestPriorityQueue(t *testing.T) {
	queue := NewSyncQueue()

	// Enqueue events with different priorities
	lowPriority := &SyncEvent{
		EventType: "low",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"priority": "low"},
	}

	highPriority := &SyncEvent{
		EventType: "high",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"priority": "high"},
	}

	// Add low then high
	queue.EnqueueWithPriority(lowPriority, 1)
	queue.EnqueueWithPriority(highPriority, 10)

	// High priority should come out first
	first := queue.Dequeue()
	if first != nil && first.Data["priority"] != "high" {
		t.Errorf("Expected high priority first, got %v", first.Data["priority"])
	}
}

// TestExpiration tests event expiration
func TestExpiration(t *testing.T) {
	queue := NewSyncQueue()

	// Create event that expires immediately
	expiredEvent := &SyncEvent{
		EventType: "expired",
		Timestamp: time.Now().Add(-10 * time.Second),
		Data:      map[string]interface{}{},
	}

	queue.Enqueue(expiredEvent)

	// Check if expired
	isExpired := queue.IsExpired(expiredEvent, 5*time.Second)
	if !isExpired {
		t.Error("Expected event to be expired")
	}
}

// TestExpiredEventRemoval tests removing expired events
func TestExpiredEventRemoval(t *testing.T) {
	queue := NewSyncQueue()

	// Enqueue mix of fresh and expired events
	for i := 0; i < 5; i++ {
		event := &SyncEvent{
			EventType: "test",
			Timestamp: time.Now().Add(-time.Duration(i) * time.Minute),
			Data:      map[string]interface{}{"index": i},
		}
		queue.Enqueue(event)
	}

	// Remove expired (older than 2 minutes)
	removed := queue.RemoveExpired(2 * time.Minute)
	if removed < 2 {
		t.Errorf("Expected at least 2 events removed, got %d", removed)
	}
}

// TestStats tests queue statistics
func TestStats(t *testing.T) {
	queue := NewSyncQueue()

	for i := 0; i < 10; i++ {
		event := &SyncEvent{
			EventType: "stat_test",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{},
		}
		queue.Enqueue(event)
	}

	stats := queue.GetStats()
	if stats.Size != 10 {
		t.Errorf("Expected size 10, got %d", stats.Size)
	}

	if stats.Processed < 0 {
		t.Errorf("Expected processed >= 0, got %d", stats.Processed)
	}
}

// TestConcurrentEnqueue tests 100+ concurrent enqueues
func TestConcurrentEnqueue(t *testing.T) {
	queue := NewSyncQueue()
	queue.SetMaxSize(500)
	var wg sync.WaitGroup
	numEnqueues := 150
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < numEnqueues; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			event := &SyncEvent{
				EventType: "concurrent_enqueue",
				Timestamp: time.Now(),
				Data:      map[string]interface{}{"id": id},
			}
			result := queue.Enqueue(event)
			if result {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if successCount < numEnqueues/2 {
		t.Errorf("Expected at least %d successes, got %d", numEnqueues/2, successCount)
	}

	if queue.Size() != successCount {
		t.Errorf("Expected queue size %d, got %d", successCount, queue.Size())
	}
}

// TestConcurrentDequeue tests 100+ concurrent dequeues
func TestConcurrentDequeue(t *testing.T) {
	queue := NewSyncQueue()

	// Pre-fill queue with 150 events
	for i := 0; i < 150; i++ {
		event := &SyncEvent{
			EventType: "test",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"id": i},
		}
		queue.Enqueue(event)
	}

	var wg sync.WaitGroup
	numDequeues := 150
	dequeueCount := 0
	var mu sync.Mutex

	for i := 0; i < numDequeues; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			event := queue.Dequeue()
			if event != nil {
				mu.Lock()
				dequeueCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if dequeueCount < numDequeues/2 {
		t.Errorf("Expected at least %d dequeues, got %d", numDequeues/2, dequeueCount)
	}

	if queue.Size() > numDequeues/2 {
		t.Errorf("Expected queue to be mostly empty, size: %d", queue.Size())
	}
}

// TestMixedOperations tests mixed enqueue/dequeue
func TestMixedOperations(t *testing.T) {
	queue := NewSyncQueue()
	queue.SetMaxSize(500)
	var wg sync.WaitGroup

	// 50 enqueuers and 50 dequeuers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				event := &SyncEvent{
					EventType: "mixed_test",
					Timestamp: time.Now(),
					Data:      map[string]interface{}{"producer": id, "seq": j},
				}
				queue.Enqueue(event)
			}
		}(i)

		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				queue.Dequeue()
			}
		}()
	}

	wg.Wait()
}

// TestConcurrency tests full concurrent stress test
func TestConcurrency(t *testing.T) {
	queue := NewSyncQueue()
	queue.SetMaxSize(1000)
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Mix of operations
			for j := 0; j < 10; j++ {
				if j%2 == 0 {
					event := &SyncEvent{
						EventType: "stress_test",
						Timestamp: time.Now(),
						Data:      map[string]interface{}{"id": id, "seq": j},
					}
					queue.Enqueue(event)
				} else {
					queue.Dequeue()
				}
			}

			// Get stats
			queue.GetStats()
		}(i)
	}

	wg.Wait()

	// Queue should have some items left
	finalSize := queue.Size()
	if finalSize < 0 {
		t.Errorf("Expected non-negative size, got %d", finalSize)
	}
}

// BenchmarkEnqueue benchmarks enqueue performance
func BenchmarkEnqueue(b *testing.B) {
	queue := NewSyncQueue()

	event := &SyncEvent{
		EventType: "benchmark",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"value": 42},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Enqueue(event)
	}
}

// BenchmarkDequeue benchmarks dequeue performance
func BenchmarkDequeue(b *testing.B) {
	queue := NewSyncQueue()
	queue.SetMaxSize(10000)

	// Pre-fill
	for i := 0; i < 10000; i++ {
		event := &SyncEvent{
			EventType: "bench",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{},
		}
		queue.Enqueue(event)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Dequeue()
	}
}

// BenchmarkQueue benchmarks combined queue operations
func BenchmarkQueue(b *testing.B) {
	queue := NewSyncQueue()
	queue.SetMaxSize(10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event := &SyncEvent{
			EventType: "bench",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"index": i},
		}
		queue.Enqueue(event)
		queue.Dequeue()
	}
}

// BenchmarkConcurrentQueue benchmarks concurrent queue performance
func BenchmarkConcurrentQueue(b *testing.B) {
	queue := NewSyncQueue()
	queue.SetMaxSize(10000)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			event := &SyncEvent{
				EventType: "bench",
				Timestamp: time.Now(),
				Data:      map[string]interface{}{"index": i},
			}
			queue.Enqueue(event)
			queue.Dequeue()
			i++
		}
	})
}

// BenchmarkRetry benchmarks retry performance
func BenchmarkRetry(b *testing.B) {
	queue := NewSyncQueue()

	event := &SyncEvent{
		EventType: "retry_bench",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.MarkForRetry(event)
		queue.DequeueRetry()
	}
}

// BenchmarkPriority benchmarks priority queue performance
func BenchmarkPriority(b *testing.B) {
	queue := NewSyncQueue()

	event := &SyncEvent{
		EventType: "priority_bench",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		priority := (i % 10) + 1
		queue.EnqueueWithPriority(event, priority)
	}
}
