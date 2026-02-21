package events

import "errors"

var (
	// ErrBusClosed is returned when trying to publish to a closed bus
	ErrBusClosed = errors.New("event bus is closed")

	// ErrEventTimeout is returned when waiting for an event times out
	ErrEventTimeout = errors.New("event timeout: no event received within timeout duration")

	// ErrInvalidEvent is returned for invalid events
	ErrInvalidEvent = errors.New("invalid event")

	// ErrInvalidHandler is returned for invalid handlers
	ErrInvalidHandler = errors.New("invalid event handler")
)
