package events

import (
	"sync"
	"testing"
	"time"
)

// TestBusCreation tests bus initialization
func TestBusCreation(t *testing.T) {
	bus := NewBus(100)
	if bus == nil {
		t.Fatal("Bus creation failed")
	}

	if bus.GetHistorySize() != 0 {
		t.Errorf("Expected empty history initially")
	}

	stats := bus.GetStats()
	if stats.TotalPublished != 0 || stats.TotalDelivered != 0 {
		t.Error("Stats should be zero initially")
	}
}

// TestSubscribe tests subscribing to events
func TestSubscribe(t *testing.T) {
	bus := NewBus(100)
	go bus.Run()
	defer bus.Stop()

	time.Sleep(10 * time.Millisecond)

	handled := false
	handler := func(event *Event) error {
		handled = true
		return nil
	}

	bus.Subscribe(EventSessionStarted, handler)
	event := NewSessionStartedEvent(1, "session1", "typing")

	bus.Publish(event)
	time.Sleep(50 * time.Millisecond)

	if !handled {
		t.Error("Event was not handled")
	}

	stats := bus.GetStats()
	if stats.TotalPublished != 1 {
		t.Errorf("Expected 1 published event, got %d", stats.TotalPublished)
	}

	if stats.TotalDelivered != 1 {
		t.Errorf("Expected 1 delivered event, got %d", stats.TotalDelivered)
	}
}

// TestSubscribeMultiple tests subscribing to multiple event types
func TestSubscribeMultiple(t *testing.T) {
	bus := NewBus(100)
	go bus.Run()
	defer bus.Stop()

	time.Sleep(10 * time.Millisecond)

	handledCount := 0
	var mu sync.Mutex

	handler := func(event *Event) error {
		mu.Lock()
		handledCount++
		mu.Unlock()
		return nil
	}

	types := []EventType{EventSessionStarted, EventSessionEnded, EventScoreUpdated}
	bus.SubscribeMultiple(types, handler)

	// Publish events of each type
	bus.Publish(NewSessionStartedEvent(1, "s1", "typing"))
	bus.Publish(NewSessionEndedEvent(1, "s1", "typing", 10*time.Minute, 100, 5))
	bus.Publish(NewScoreUpdatedEvent(1, "typing", "s1", 50, 100))

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if handledCount != 3 {
		t.Errorf("Expected 3 handled events, got %d", handledCount)
	}
}

// TestSubscribeAll tests subscribing to all events
func TestSubscribeAll(t *testing.T) {
	bus := NewBus(100)
	go bus.Run()
	defer bus.Stop()

	time.Sleep(10 * time.Millisecond)

	handledCount := 0
	var mu sync.Mutex

	handler := func(event *Event) error {
		mu.Lock()
		handledCount++
		mu.Unlock()
		return nil
	}

	bus.SubscribeAll(handler)

	// Publish different event types
	bus.Publish(NewSessionStartedEvent(1, "s1", "typing"))
	bus.Publish(NewRankChangedEvent(1, "typing_wpm", 10, 5))
	bus.Publish(NewAchievementUnlockedEvent(1, "ach1", "Pro", "Amazing", "🏆", 50))

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if handledCount != 3 {
		t.Errorf("Expected 3 handled events by catch-all, got %d", handledCount)
	}
}

// TestEventHistory tests event history tracking
func TestEventHistory(t *testing.T) {
	bus := NewBus(100)
	go bus.Run()
	defer bus.Stop()

	time.Sleep(10 * time.Millisecond)

	handler := func(event *Event) error { return nil }
	bus.Subscribe(EventSessionStarted, handler)

	// Publish multiple events
	for i := 0; i < 5; i++ {
		bus.Publish(NewSessionStartedEvent(uint(i), "s1", "typing"))
	}

	time.Sleep(100 * time.Millisecond)

	if bus.GetHistorySize() != 5 {
		t.Errorf("Expected 5 events in history, got %d", bus.GetHistorySize())
	}

	recent := bus.GetRecentEvents(3)
	if len(recent) != 3 {
		t.Errorf("Expected 3 recent events, got %d", len(recent))
	}
}

// TestEventFiltering tests event history filtering
func TestEventFiltering(t *testing.T) {
	bus := NewBus(100)
	go bus.Run()
	defer bus.Stop()

	time.Sleep(10 * time.Millisecond)

	handler := func(event *Event) error { return nil }
	bus.Subscribe(EventSessionStarted, handler)
	bus.Subscribe(EventScoreUpdated, handler)

	// Publish mixed events
	bus.Publish(NewSessionStartedEvent(1, "s1", "typing"))
	bus.Publish(NewSessionStartedEvent(2, "s2", "math"))
	bus.Publish(NewScoreUpdatedEvent(1, "typing", "s1", 50, 100))

	time.Sleep(100 * time.Millisecond)

	// Filter by user
	userID := uint(1)
	filtered := bus.GetEventsByUser(userID, 10)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 events for user 1, got %d", len(filtered))
	}

	// Filter by type
	typeFiltered := bus.GetEventsByType(EventSessionStarted, 10)
	if len(typeFiltered) != 2 {
		t.Errorf("Expected 2 session start events, got %d", len(typeFiltered))
	}
}

// TestPublishAsync tests async publishing
func TestPublishAsync(t *testing.T) {
	bus := NewBus(100)
	go bus.Run()
	defer bus.Stop()

	time.Sleep(10 * time.Millisecond)

	handled := 0
	var mu sync.Mutex

	handler := func(event *Event) error {
		mu.Lock()
		handled++
		mu.Unlock()
		return nil
	}

	bus.Subscribe(EventSessionStarted, handler)

	// Publish async
	bus.PublishAsync(NewSessionStartedEvent(1, "s1", "typing"))

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if handled != 1 {
		t.Error("Async event not handled")
	}
}

// TestHandlerRegistry tests handler registration
func TestHandlerRegistry(t *testing.T) {
	bus := NewBus(100)
	go bus.Run()
	defer bus.Stop()

	time.Sleep(10 * time.Millisecond)

	registry := NewHandlerRegistry(bus)

	handledCount := 0
	handler := func(event *Event) error {
		handledCount++
		return nil
	}

	registry.RegisterHandler(EventSessionStarted, handler)

	bus.Publish(NewSessionStartedEvent(1, "s1", "typing"))
	time.Sleep(50 * time.Millisecond)

	if handledCount != 1 {
		t.Error("Registered handler not called")
	}
}

// TestEventRouter tests event routing
func TestEventRouter(t *testing.T) {
	bus := NewBus(100)
	go bus.Run()
	defer bus.Stop()

	time.Sleep(10 * time.Millisecond)

	router := NewEventRouter(bus)

	routed := false
	handler := func(event *Event) error {
		routed = true
		return nil
	}

	router.AddRoute(EventRankChanged, handler)
	bus.Publish(NewRankChangedEvent(1, "typing_wpm", 10, 5))

	time.Sleep(50 * time.Millisecond)

	if !routed {
		t.Error("Event not routed correctly")
	}
}

// TestConditionalHandler tests conditional handler
func TestConditionalHandler(t *testing.T) {
	bus := NewBus(100)
	go bus.Run()
	defer bus.Stop()

	time.Sleep(10 * time.Millisecond)

	router := NewEventRouter(bus)

	handled := false
	handler := func(event *Event) error {
		handled = true
		return nil
	}

	userID := uint(1)
	filter := &EventFilter{
		UserID: &userID,
	}

	conditionalHandler := router.ConditionalHandler(filter, handler)
	router.AddRoute(EventRankChanged, conditionalHandler)

	// Event for user 1 - should be handled
	bus.Publish(NewRankChangedEvent(1, "typing_wpm", 10, 5))
	time.Sleep(50 * time.Millisecond)

	if !handled {
		t.Error("Conditional handler did not handle event for correct user")
	}

	handled = false

	// Event for user 2 - should not be handled
	bus.Publish(NewRankChangedEvent(2, "typing_wpm", 10, 5))
	time.Sleep(50 * time.Millisecond)

	if handled {
		t.Error("Conditional handler handled event for wrong user")
	}
}

// TestChainHandlers tests handler chaining
func TestChainHandlers(t *testing.T) {
	handler1Called := false
	handler2Called := false

	h1 := func(event *Event) error {
		handler1Called = true
		return nil
	}

	h2 := func(event *Event) error {
		handler2Called = true
		return nil
	}

	chained := ChainHandlers(h1, h2)
	event := NewEvent(EventSessionStarted, 1, "typing", nil)

	err := chained(event)

	if err != nil {
		t.Errorf("Chained handler returned error: %v", err)
	}

	if !handler1Called || !handler2Called {
		t.Error("Not all handlers in chain were called")
	}
}

// TestWaitFor tests waiting for a specific event
func TestWaitFor(t *testing.T) {
	bus := NewBus(100)
	go bus.Run()
	defer bus.Stop()

	time.Sleep(10 * time.Millisecond)

	go func() {
		time.Sleep(50 * time.Millisecond)
		bus.Publish(NewRankChangedEvent(1, "typing_wpm", 10, 5))
	}()

	event, err := bus.WaitFor(EventRankChanged, 2*time.Second)

	if err != nil {
		t.Errorf("WaitFor returned error: %v", err)
	}

	if event == nil {
		t.Error("WaitFor returned nil event")
	}

	if event.Type != EventRankChanged {
		t.Errorf("Wrong event type: %s", event.Type)
	}
}

// TestWaitForTimeout tests waiting for event with timeout
func TestWaitForTimeout(t *testing.T) {
	bus := NewBus(100)
	go bus.Run()
	defer bus.Stop()

	time.Sleep(10 * time.Millisecond)

	event, err := bus.WaitFor(EventRankChanged, 100*time.Millisecond)

	if err == nil {
		t.Error("Expected timeout error")
	}

	if event != nil {
		t.Error("Expected nil event on timeout")
	}
}

// TestHistoryLimit tests that history size is limited
func TestHistoryLimit(t *testing.T) {
	bus := NewBus(5) // Small history size
	go bus.Run()
	defer bus.Stop()

	time.Sleep(10 * time.Millisecond)

	handler := func(event *Event) error { return nil }
	bus.Subscribe(EventSessionStarted, handler)

	// Publish more events than history size
	for i := 0; i < 10; i++ {
		bus.Publish(NewSessionStartedEvent(uint(i), "s1", "typing"))
	}

	time.Sleep(100 * time.Millisecond)

	if bus.GetHistorySize() > 5 {
		t.Errorf("History size exceeded limit: %d > 5", bus.GetHistorySize())
	}
}

// BenchmarkPublish benchmarks event publishing
func BenchmarkPublish(b *testing.B) {
	bus := NewBus(10000)
	go bus.Run()
	defer bus.Stop()

	time.Sleep(10 * time.Millisecond)

	handler := func(event *Event) error { return nil }
	bus.Subscribe(EventSessionStarted, handler)

	event := NewSessionStartedEvent(1, "s1", "typing")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish(event)
	}
}

// BenchmarkSubscribe benchmarks subscription
func BenchmarkSubscribe(b *testing.B) {
	bus := NewBus(100)

	handler := func(event *Event) error { return nil }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Subscribe(EventSessionStarted, handler)
	}
}
