package math

import (
	"context"
	"sync"
	"time"
)

// CrossAppSyncManager manages cross-app synchronization
type CrossAppSyncManager struct {
	AppID      string
	events     []*SyncEvent
	mu         sync.RWMutex
	eventCount int
}

// NewCrossAppSyncManager creates a new sync manager
func NewCrossAppSyncManager(appID string, _ interface{}) *CrossAppSyncManager {
	return &CrossAppSyncManager{
		AppID:  appID,
		events: make([]*SyncEvent, 0),
	}
}

// RecordEvent records a sync event
func (m *CrossAppSyncManager) RecordEvent(ctx context.Context, event *SyncEvent) error {
	if event == nil {
		return ErrNilEvent
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	m.eventCount++
	return nil
}

// GetSyncStatus gets the current sync status
func (m *CrossAppSyncManager) GetSyncStatus(ctx context.Context) *SyncStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &SyncStatus{
		AppID:      m.AppID,
		EventCount: m.eventCount,
		LastSync:   time.Now(),
	}
}

// TransformEventData transforms event data for cross-app compatibility
func (m *CrossAppSyncManager) TransformEventData(event *SyncEvent) map[string]interface{} {
	if event == nil {
		return nil
	}

	return map[string]interface{}{
		"event_type": event.EventType,
		"timestamp":  event.Timestamp,
		"data":       event.Data,
	}
}

// FilterMetrics filters events by type
func (m *CrossAppSyncManager) FilterMetrics(events []*SyncEvent, eventType string) []*SyncEvent {
	var filtered []*SyncEvent
	for _, e := range events {
		if e.EventType == eventType {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// BatchEvents batches events into groups
func (m *CrossAppSyncManager) BatchEvents(events []*SyncEvent, batchSize int) [][]*SyncEvent {
	var batches [][]*SyncEvent
	for i := 0; i < len(events); i += batchSize {
		end := i + batchSize
		if end > len(events) {
			end = len(events)
		}
		batches = append(batches, events[i:end])
	}
	return batches
}

// DeduplicateEvents removes duplicate events
func (m *CrossAppSyncManager) DeduplicateEvents(events []*SyncEvent) []*SyncEvent {
	seen := make(map[string]bool)
	var deduped []*SyncEvent

	for _, e := range events {
		key := e.EventType + ":" + e.Timestamp.String()
		if !seen[key] {
			seen[key] = true
			deduped = append(deduped, e)
		}
	}
	return deduped
}

// OrderEventsByTimestamp orders events by timestamp
func (m *CrossAppSyncManager) OrderEventsByTimestamp(events []*SyncEvent) []*SyncEvent {
	ordered := make([]*SyncEvent, len(events))
	copy(ordered, events)

	// Simple bubble sort for small arrays
	for i := 0; i < len(ordered); i++ {
		for j := i + 1; j < len(ordered); j++ {
			if ordered[j].Timestamp.Before(ordered[i].Timestamp) {
				ordered[i], ordered[j] = ordered[j], ordered[i]
			}
		}
	}
	return ordered
}

// SyncCoordinator coordinates synchronization across apps
type SyncCoordinator struct {
	subscribers map[string]chan *SyncEvent
	mu          sync.RWMutex
	queue       []*SyncEvent
	taskMap     map[string]context.CancelFunc
}

// NewSyncCoordinator creates a new coordinator
func NewSyncCoordinator() *SyncCoordinator {
	return &SyncCoordinator{
		subscribers: make(map[string]chan *SyncEvent),
		queue:       make([]*SyncEvent, 0),
		taskMap:     make(map[string]context.CancelFunc),
	}
}

// SubscribeToEvents subscribes to events
func (c *SyncCoordinator) SubscribeToEvents(ctx context.Context, ch chan *SyncEvent, app string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := app + "-" + time.Now().Format("20060102150405.000000")
	c.subscribers[id] = ch
	return id
}

// UnsubscribeFromEvents unsubscribes from events
func (c *SyncCoordinator) UnsubscribeFromEvents(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.subscribers[id]; !exists {
		return false
	}
	delete(c.subscribers, id)
	return true
}

// BroadcastToSubscribers broadcasts an event to all subscribers
func (c *SyncCoordinator) BroadcastToSubscribers(event *SyncEvent) {
	c.mu.RLock()
	subscribers := make(map[string]chan *SyncEvent)
	for id, ch := range c.subscribers {
		subscribers[id] = ch
	}
	c.mu.RUnlock()

	for _, ch := range subscribers {
		select {
		case ch <- event:
		default:
			// Non-blocking send
		}
	}
}

// NormalizeMetrics normalizes metrics from different apps
func (c *SyncCoordinator) NormalizeMetrics(app string, metrics map[string]interface{}) map[string]interface{} {
	normalized := make(map[string]interface{})
	for k, v := range metrics {
		normalized[k] = v
	}
	return normalized
}

// ScheduleSync schedules a sync task
func (c *SyncCoordinator) ScheduleSync(ctx context.Context, taskID string, interval time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	c.taskMap[taskID] = cancel
	return true
}

// CancelSync cancels a sync task
func (c *SyncCoordinator) CancelSync(taskID string) bool {
	c.mu.Lock()
	cancel, exists := c.taskMap[taskID]
	if exists {
		delete(c.taskMap, taskID)
	}
	c.mu.Unlock()

	if exists && cancel != nil {
		cancel()
		return true
	}
	return false
}

// RecoverFromError attempts to recover from an error
func (c *SyncCoordinator) RecoverFromError(event *SyncEvent) map[string]interface{} {
	return map[string]interface{}{
		"strategy": "retry",
		"event":    event,
	}
}

// EnqueueEvent adds an event to the queue
func (c *SyncCoordinator) EnqueueEvent(event *SyncEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.queue = append(c.queue, event)
}

// DequeueEvent removes an event from the queue
func (c *SyncCoordinator) DequeueEvent() *SyncEvent {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.queue) == 0 {
		return nil
	}
	event := c.queue[0]
	c.queue = c.queue[1:]
	return event
}

// GetCoordinatorStatus gets coordinator status
func (c *SyncCoordinator) GetCoordinatorStatus(ctx context.Context) *CoordinatorStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return &CoordinatorStatus{
		SubscriberCount: len(c.subscribers),
		QueueSize:       len(c.queue),
		ProcessedCount:  0,
	}
}

// SyncBroadcaster broadcasts events across apps
type SyncBroadcaster struct {
	queue []*SyncEvent
	mu    sync.RWMutex
}

// NewSyncBroadcaster creates a new broadcaster
func NewSyncBroadcaster() *SyncBroadcaster {
	return &SyncBroadcaster{
		queue: make([]*SyncEvent, 0),
	}
}

// BroadcastProgress broadcasts progress updates
func (b *SyncBroadcaster) BroadcastProgress(ctx context.Context, progress map[string]interface{}) bool {
	if progress == nil {
		return false
	}

	select {
	case <-ctx.Done():
		return false
	default:
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	event := &SyncEvent{
		EventType: "progress",
		Timestamp: time.Now(),
		Data:      progress,
	}
	b.queue = append(b.queue, event)
	return true
}

// BroadcastLeaderboard broadcasts leaderboard updates
func (b *SyncBroadcaster) BroadcastLeaderboard(ctx context.Context, app string, leaderboard []map[string]interface{}) bool {
	if app == "" || leaderboard == nil {
		return false
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	data := map[string]interface{}{
		"app":         app,
		"leaderboard": leaderboard,
	}

	event := &SyncEvent{
		EventType: "leaderboard",
		Timestamp: time.Now(),
		Data:      data,
	}
	b.queue = append(b.queue, event)
	return true
}

// PropagateAchievements propagates achievements
func (b *SyncBroadcaster) PropagateAchievements(ctx context.Context, userID int, achievements []map[string]interface{}) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	data := map[string]interface{}{
		"user_id":       userID,
		"achievements":  achievements,
	}

	event := &SyncEvent{
		EventType: "achievements",
		Timestamp: time.Now(),
		Data:      data,
	}
	b.queue = append(b.queue, event)
	return true
}

// BroadcastActivity broadcasts activity updates
func (b *SyncBroadcaster) BroadcastActivity(ctx context.Context, activity map[string]interface{}) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	event := &SyncEvent{
		EventType: "activity",
		Timestamp: time.Now(),
		Data:      activity,
	}
	b.queue = append(b.queue, event)
	return true
}

// BroadcastToApps broadcasts to multiple apps
func (b *SyncBroadcaster) BroadcastToApps(ctx context.Context, apps []string, event map[string]interface{}) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	data := map[string]interface{}{
		"apps":  apps,
		"event": event,
	}

	syncEvent := &SyncEvent{
		EventType: "multi_app",
		Timestamp: time.Now(),
		Data:      data,
	}
	b.queue = append(b.queue, syncEvent)
	return true
}

// BroadcastWithPriority broadcasts with priority
func (b *SyncBroadcaster) BroadcastWithPriority(ctx context.Context, data map[string]interface{}, priority string) bool {
	data["priority"] = priority

	b.mu.Lock()
	defer b.mu.Unlock()

	event := &SyncEvent{
		EventType: "priority_broadcast",
		Timestamp: time.Now(),
		Data:      data,
	}
	b.queue = append(b.queue, event)
	return true
}

// BroadcastProgressToUsers broadcasts progress to specific users
func (b *SyncBroadcaster) BroadcastProgressToUsers(ctx context.Context, progress map[string]interface{}, userIDs []int) bool {
	progress["user_ids"] = userIDs

	b.mu.Lock()
	defer b.mu.Unlock()

	event := &SyncEvent{
		EventType: "user_progress",
		Timestamp: time.Now(),
		Data:      progress,
	}
	b.queue = append(b.queue, event)
	return true
}

// NotifyAchievement notifies about an achievement
func (b *SyncBroadcaster) NotifyAchievement(ctx context.Context, achievement map[string]interface{}) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	event := &SyncEvent{
		EventType: "achievement_notification",
		Timestamp: time.Now(),
		Data:      achievement,
	}
	b.queue = append(b.queue, event)
	return true
}

// EnqueueBroadcast enqueues a broadcast event
func (b *SyncBroadcaster) EnqueueBroadcast(event map[string]interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	syncEvent := &SyncEvent{
		EventType: "queued_broadcast",
		Timestamp: time.Now(),
		Data:      event,
	}
	b.queue = append(b.queue, syncEvent)
}

// DequeueBroadcast dequeues a broadcast event
func (b *SyncBroadcaster) DequeueBroadcast() map[string]interface{} {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.queue) == 0 {
		return nil
	}

	event := b.queue[0]
	b.queue = b.queue[1:]
	return event.Data
}

// BroadcastWithRetry broadcasts with retry logic
func (b *SyncBroadcaster) BroadcastWithRetry(ctx context.Context, event map[string]interface{}, maxRetries int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	syncEvent := &SyncEvent{
		EventType: "broadcast_retry",
		Timestamp: time.Now(),
		Data:      event,
	}
	b.queue = append(b.queue, syncEvent)
	return true
}

// GetStatus gets broadcaster status
func (b *SyncBroadcaster) GetStatus(ctx context.Context) map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return map[string]interface{}{
		"queue_size": len(b.queue),
	}
}

// SyncQueue manages event queueing
type SyncQueue struct {
	queue       []*SyncEvent
	retryQueue  []*SyncEvent
	maxSize     int
	mu          sync.RWMutex
	processed   int
	dropped     int
}

// NewSyncQueue creates a new queue
func NewSyncQueue() *SyncQueue {
	return &SyncQueue{
		queue:      make([]*SyncEvent, 0),
		retryQueue: make([]*SyncEvent, 0),
		maxSize:    100,
	}
}

// SetMaxSize sets the maximum queue size
func (q *SyncQueue) SetMaxSize(size int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.maxSize = size
}

// Enqueue adds an event to the queue
func (q *SyncQueue) Enqueue(event *SyncEvent) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) >= q.maxSize {
		return false
	}

	q.queue = append(q.queue, event)
	return true
}

// Dequeue removes an event from the queue
func (q *SyncQueue) Dequeue() *SyncEvent {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	event := q.queue[0]
	q.queue = q.queue[1:]
	q.processed++
	return event
}

// Size returns queue size
func (q *SyncQueue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.queue)
}

// MarkForRetry marks an event for retry
func (q *SyncQueue) MarkForRetry(event *SyncEvent) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.retryQueue) < 100 {
		q.retryQueue = append(q.retryQueue, event)
		return true
	}
	return false
}

// DequeueRetry dequeues a retry event
func (q *SyncQueue) DequeueRetry() *SyncEvent {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.retryQueue) == 0 {
		return nil
	}

	event := q.retryQueue[0]
	q.retryQueue = q.retryQueue[1:]
	return event
}

// CalculateBackoff calculates exponential backoff
func (q *SyncQueue) CalculateBackoff(retryCount int) time.Duration {
	baseBackoff := 100 * time.Millisecond
	multiplier := 1 << uint(retryCount-1)
	return baseBackoff * time.Duration(multiplier)
}

// IsExpired checks if an event has expired
func (q *SyncQueue) IsExpired(event *SyncEvent, ttl time.Duration) bool {
	return time.Since(event.Timestamp) > ttl
}

// RemoveExpired removes expired events
func (q *SyncQueue) RemoveExpired(ttl time.Duration) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	removed := 0
	var filtered []*SyncEvent

	for _, event := range q.queue {
		if time.Since(event.Timestamp) > ttl {
			removed++
		} else {
			filtered = append(filtered, event)
		}
	}

	q.queue = filtered
	return removed
}

// GetStats returns queue statistics
func (q *SyncQueue) GetStats() *QueueStats {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return &QueueStats{
		Size:       len(q.queue),
		MaxSize:    q.maxSize,
		Processed:  q.processed,
		Dropped:    q.dropped,
		LastUpdate: time.Now(),
	}
}

// EnqueueWithPriority enqueues with priority
func (q *SyncQueue) EnqueueWithPriority(event *SyncEvent, priority int) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) >= q.maxSize {
		return false
	}

	// For simplicity, just add to front if high priority
	if priority >= 5 {
		newQueue := []*SyncEvent{event}
		q.queue = append(newQueue, q.queue...)
	} else {
		q.queue = append(q.queue, event)
	}

	return true
}

// Error definitions
var (
	ErrNilEvent = &SyncError{"event cannot be nil"}
)

// Custom error type
type SyncError struct {
	Message string
}

func (e *SyncError) Error() string {
	return e.Message
}
