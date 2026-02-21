package events

import (
	"fmt"
	"sync"
)

// HandlerRegistry manages event handlers with logging and error handling
type HandlerRegistry struct {
	bus           *Bus
	handlers      map[string][]EventHandler
	mu            sync.RWMutex
	errorHandler  func(error) // Optional error handler
	loggingEnabled bool
}

// NewHandlerRegistry creates a new handler registry
func NewHandlerRegistry(bus *Bus) *HandlerRegistry {
	return &HandlerRegistry{
		bus:      bus,
		handlers: make(map[string][]EventHandler),
	}
}

// RegisterHandler registers a handler for an event type
func (r *HandlerRegistry) RegisterHandler(eventType EventType, handler EventHandler) error {
	if handler == nil {
		return ErrInvalidHandler
	}

	key := string(eventType)

	r.mu.Lock()
	r.handlers[key] = append(r.handlers[key], handler)
	r.mu.Unlock()

	// Subscribe to bus
	r.bus.Subscribe(eventType, handler)

	if r.loggingEnabled {
		fmt.Printf("[Events] Registered handler for %s\n", eventType)
	}

	return nil
}

// RegisterHandlerWithName registers a named handler for an event type
func (r *HandlerRegistry) RegisterHandlerWithName(eventType EventType, name string, handler EventHandler) error {
	if handler == nil {
		return ErrInvalidHandler
	}

	key := string(eventType) + ":" + name

	r.mu.Lock()
	r.handlers[key] = append(r.handlers[key], handler)
	r.mu.Unlock()

	// Subscribe to bus
	r.bus.Subscribe(eventType, handler)

	if r.loggingEnabled {
		fmt.Printf("[Events] Registered handler %s for %s\n", name, eventType)
	}

	return nil
}

// RegisterHandlerForMultipleTypes registers a handler for multiple event types
func (r *HandlerRegistry) RegisterHandlerForMultipleTypes(eventTypes []EventType, handler EventHandler) error {
	for _, eventType := range eventTypes {
		if err := r.RegisterHandler(eventType, handler); err != nil {
			return err
		}
	}
	return nil
}

// GetHandlers returns all handlers for an event type
func (r *HandlerRegistry) GetHandlers(eventType EventType) []EventHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := string(eventType)
	return r.handlers[key]
}

// RemoveHandlers removes all handlers for an event type
func (r *HandlerRegistry) RemoveHandlers(eventType EventType) {
	r.mu.Lock()
	delete(r.handlers, string(eventType))
	r.mu.Unlock()

	r.bus.Unsubscribe(eventType)
}

// SetErrorHandler sets a function to handle errors
func (r *HandlerRegistry) SetErrorHandler(handler func(error)) {
	r.errorHandler = handler
}

// EnableLogging enables event handler logging
func (r *HandlerRegistry) EnableLogging(enabled bool) {
	r.loggingEnabled = enabled
}

// EventRouter routes events to appropriate handlers based on rules
type EventRouter struct {
	bus     *Bus
	routes  map[EventType][]EventHandler
	mu      sync.RWMutex
	filters map[string]*EventFilter
	filterMu sync.RWMutex
}

// NewEventRouter creates a new event router
func NewEventRouter(bus *Bus) *EventRouter {
	return &EventRouter{
		bus:     bus,
		routes:  make(map[EventType][]EventHandler),
		filters: make(map[string]*EventFilter),
	}
}

// AddRoute adds a route from an event type to a handler
func (r *EventRouter) AddRoute(eventType EventType, handler EventHandler) {
	if handler == nil {
		return
	}

	r.mu.Lock()
	r.routes[eventType] = append(r.routes[eventType], handler)
	r.mu.Unlock()

	// Subscribe to bus
	r.bus.Subscribe(eventType, handler)
}

// AddFilter adds a named filter for conditional routing
func (r *EventRouter) AddFilter(name string, filter *EventFilter) {
	r.filterMu.Lock()
	r.filters[name] = filter
	r.filterMu.Unlock()
}

// GetFilter returns a named filter
func (r *EventRouter) GetFilter(name string) *EventFilter {
	r.filterMu.RLock()
	defer r.filterMu.RUnlock()
	return r.filters[name]
}

// ConditionalHandler creates a handler that only executes if a filter matches
func (r *EventRouter) ConditionalHandler(filter *EventFilter, handler EventHandler) EventHandler {
	return func(event *Event) error {
		if event.Matches(filter) {
			return handler(event)
		}
		return nil
	}
}

// ChainHandlers creates a handler that calls multiple handlers in sequence
func ChainHandlers(handlers ...EventHandler) EventHandler {
	return func(event *Event) error {
		for _, handler := range handlers {
			if handler != nil {
				if err := handler(event); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

// ParallelHandlers creates a handler that calls multiple handlers concurrently
func ParallelHandlers(handlers ...EventHandler) EventHandler {
	return func(event *Event) error {
		var wg sync.WaitGroup
		errorsChan := make(chan error, len(handlers))

		for _, handler := range handlers {
			if handler != nil {
				wg.Add(1)
				go func(h EventHandler) {
					defer wg.Done()
					if err := h(event); err != nil {
						errorsChan <- err
					}
				}(handler)
			}
		}

		wg.Wait()
		close(errorsChan)

		// Return first error if any
		for err := range errorsChan {
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// ConditionalChainHandler creates a handler that chains handlers conditionally
func ConditionalChainHandler(conditions []EventFilter, handlers []EventHandler) EventHandler {
	return func(event *Event) error {
		for i, condition := range conditions {
			if event.Matches(&condition) && i < len(handlers) {
				return handlers[i](event)
			}
		}
		return nil
	}
}

// LoggingHandler wraps a handler to add logging
func LoggingHandler(name string, handler EventHandler) EventHandler {
	return func(event *Event) error {
		fmt.Printf("[%s] Processing event: %s (User: %d, App: %s)\n", name, event.Type, event.UserID, event.App)
		err := handler(event)
		if err != nil {
			fmt.Printf("[%s] Handler error: %v\n", name, err)
		}
		return err
	}
}

// RecoveryHandler wraps a handler with panic recovery
func RecoveryHandler(handler EventHandler) EventHandler {
	return func(event *Event) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("handler panic: %v", r)
			}
		}()
		return handler(event)
	}
}

// ThrottledHandler creates a handler that throttles event processing
func ThrottledHandler(handler EventHandler, maxConcurrent int) EventHandler {
	semaphore := make(chan struct{}, maxConcurrent)

	return func(event *Event) error {
		semaphore <- struct{}{}
		defer func() { <-semaphore }()
		return handler(event)
	}
}

// BufferedHandler creates a handler that buffers events for batch processing
func BufferedHandler(handler EventHandler, batchSize int, flushInterval int) EventHandler {
	buffer := make([]*Event, 0, batchSize)
	var mu sync.Mutex

	return func(event *Event) error {
		mu.Lock()
		buffer = append(buffer, event)
		shouldFlush := len(buffer) >= batchSize
		mu.Unlock()

		if shouldFlush {
			mu.Lock()
			if len(buffer) > 0 {
				// Process buffer (simplified - in real use would batch process)
				for _, e := range buffer {
					if err := handler(e); err != nil {
						mu.Unlock()
						return err
					}
				}
				buffer = buffer[:0]
			}
			mu.Unlock()
		}

		return nil
	}
}
