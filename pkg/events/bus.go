package events

import (
	"sync"
	"time"
)

// Bus is the central event bus for publishing and subscribing to events
type Bus struct {
	// Map of event types to subscriber lists
	subscribers map[EventType][]EventHandler
	mu          sync.RWMutex

	// Event queue for async processing
	eventQueue chan *Event
	quit       chan bool

	// Event history
	history     []*Event
	historySize int
	historyMu   sync.RWMutex

	// Bus statistics
	stats BusStats
}

// BusStats contains bus statistics
type BusStats struct {
	mu               sync.RWMutex
	TotalPublished   int64
	TotalDelivered   int64
	TotalErrors      int64
	ActiveSubscribers int64
}

// NewBus creates a new event bus
func NewBus(historySize int) *Bus {
	return &Bus{
		subscribers: make(map[EventType][]EventHandler),
		eventQueue:  make(chan *Event, 1000),
		quit:        make(chan bool),
		history:     make([]*Event, 0, historySize),
		historySize: historySize,
	}
}

// Subscribe adds a handler for a specific event type
func (b *Bus) Subscribe(eventType EventType, handler EventHandler) {
	if handler == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[eventType] = append(b.subscribers[eventType], handler)

	b.stats.mu.Lock()
	b.stats.ActiveSubscribers++
	b.stats.mu.Unlock()
}

// SubscribeMultiple adds a handler for multiple event types
func (b *Bus) SubscribeMultiple(eventTypes []EventType, handler EventHandler) {
	if handler == nil {
		return
	}

	for _, eventType := range eventTypes {
		b.Subscribe(eventType, handler)
	}
}

// SubscribeAll adds a handler for all events
func (b *Bus) SubscribeAll(handler EventHandler) {
	if handler == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Create a catch-all subscription
	b.subscribers[EventType("")] = append(b.subscribers[EventType("")], handler)

	b.stats.mu.Lock()
	b.stats.ActiveSubscribers++
	b.stats.mu.Unlock()
}

// Unsubscribe removes all handlers for a specific event type
func (b *Bus) Unsubscribe(eventType EventType) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if handlers, ok := b.subscribers[eventType]; ok {
		b.stats.mu.Lock()
		b.stats.ActiveSubscribers -= int64(len(handlers))
		b.stats.mu.Unlock()

		delete(b.subscribers, eventType)
	}
}

// Publish publishes an event to all subscribers
func (b *Bus) Publish(event *Event) error {
	if event == nil {
		return nil
	}

	select {
	case b.eventQueue <- event:
		b.stats.mu.Lock()
		b.stats.TotalPublished++
		b.stats.mu.Unlock()
		return nil
	case <-b.quit:
		return ErrBusClosed
	}
}

// PublishAsync publishes an event asynchronously (non-blocking)
func (b *Bus) PublishAsync(event *Event) {
	if event == nil {
		return
	}

	select {
	case b.eventQueue <- event:
		b.stats.mu.Lock()
		b.stats.TotalPublished++
		b.stats.mu.Unlock()
	default:
		// Queue is full, drop event
		b.stats.mu.Lock()
		b.stats.TotalErrors++
		b.stats.mu.Unlock()
	}
}

// Run starts the event bus processing loop
func (b *Bus) Run() {
	for {
		select {
		case event := <-b.eventQueue:
			b.handleEvent(event)

		case <-b.quit:
			return
		}
	}
}

// handleEvent delivers an event to all subscribers
func (b *Bus) handleEvent(event *Event) {
	if event == nil {
		return
	}

	// Add to history
	b.addToHistory(event)

	b.mu.RLock()
	handlers := b.subscribers[event.Type]
	catchAllHandlers := b.subscribers[EventType("")] // Catch-all handlers
	b.mu.RUnlock()

	// Call type-specific handlers
	for _, handler := range handlers {
		if err := handler(event); err != nil {
			b.stats.mu.Lock()
			b.stats.TotalErrors++
			b.stats.mu.Unlock()
		} else {
			b.stats.mu.Lock()
			b.stats.TotalDelivered++
			b.stats.mu.Unlock()
		}
	}

	// Call catch-all handlers
	for _, handler := range catchAllHandlers {
		if err := handler(event); err != nil {
			b.stats.mu.Lock()
			b.stats.TotalErrors++
			b.stats.mu.Unlock()
		} else {
			b.stats.mu.Lock()
			b.stats.TotalDelivered++
			b.stats.mu.Unlock()
		}
	}
}

// addToHistory adds an event to the history buffer
func (b *Bus) addToHistory(event *Event) {
	b.historyMu.Lock()
	defer b.historyMu.Unlock()

	b.history = append(b.history, event)

	// Keep history size limited
	if len(b.history) > b.historySize {
		b.history = b.history[1:]
	}
}

// GetHistory returns recent events matching the filter
func (b *Bus) GetHistory(filter *EventFilter, limit int) []*Event {
	b.historyMu.RLock()
	defer b.historyMu.RUnlock()

	var result []*Event
	// Iterate from end to get most recent first
	for i := len(b.history) - 1; i >= 0 && len(result) < limit; i-- {
		if b.history[i].Matches(filter) {
			result = append(result, b.history[i])
		}
	}

	return result
}

// GetRecentEvents returns the most recent events
func (b *Bus) GetRecentEvents(limit int) []*Event {
	b.historyMu.RLock()
	defer b.historyMu.RUnlock()

	start := len(b.history) - limit
	if start < 0 {
		start = 0
	}

	result := make([]*Event, 0, limit)
	for i := len(b.history) - 1; i >= start; i-- {
		result = append(result, b.history[i])
	}

	return result
}

// GetEventsByUser returns events for a specific user
func (b *Bus) GetEventsByUser(userID uint, limit int) []*Event {
	filter := &EventFilter{
		UserID: &userID,
	}
	return b.GetHistory(filter, limit)
}

// GetEventsByType returns events of a specific type
func (b *Bus) GetEventsByType(eventType EventType, limit int) []*Event {
	filter := &EventFilter{
		EventTypes: []EventType{eventType},
	}
	return b.GetHistory(filter, limit)
}

// GetEventsByApp returns events for a specific app
func (b *Bus) GetEventsByApp(app string, limit int) []*Event {
	filter := &EventFilter{
		App: &app,
	}
	return b.GetHistory(filter, limit)
}

// GetStats returns bus statistics
func (b *Bus) GetStats() BusStats {
	b.stats.mu.RLock()
	defer b.stats.mu.RUnlock()
	return b.stats
}

// Stop stops the event bus
func (b *Bus) Stop() {
	close(b.quit)
}

// Clear clears the event history
func (b *Bus) Clear() {
	b.historyMu.Lock()
	defer b.historyMu.Unlock()
	b.history = make([]*Event, 0, b.historySize)
}

// GetHistorySize returns the current size of event history
func (b *Bus) GetHistorySize() int {
	b.historyMu.RLock()
	defer b.historyMu.RUnlock()
	return len(b.history)
}

// WaitFor waits for an event of a specific type with timeout
func (b *Bus) WaitFor(eventType EventType, timeout time.Duration) (*Event, error) {
	received := make(chan *Event, 1)
	var unsubscribed bool

	handler := func(event *Event) error {
		select {
		case received <- event:
		default:
		}
		return nil
	}

	b.Subscribe(eventType, handler)

	select {
	case event := <-received:
		b.Unsubscribe(eventType)
		unsubscribed = true
		return event, nil
	case <-time.After(timeout):
		if !unsubscribed {
			b.Unsubscribe(eventType)
		}
		return nil, ErrEventTimeout
	}
}
