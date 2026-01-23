package event

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Event represents a concrete event implementation
// The struct is thread-safe for concurrent access to headers
type Event struct {
	id        string
	topic     string
	name      string
	payload   any
	headers   map[string]string
	headersMu sync.RWMutex // Protects headers map for concurrent access
	timestamp time.Time
	ctx       context.Context
}

// NewEvent creates a new event
func NewEvent(name string, payload any) *Event {
	return &Event{
		id:        uuid.New().String(),
		name:      name,
		payload:   payload,
		headers:   make(map[string]string),
		timestamp: time.Now(),
		ctx:       context.Background(),
	}
}

// NewEventWithContext creates a new event with context
func NewEventWithContext(ctx context.Context, name string, payload any) *Event {
	return &Event{
		id:        uuid.New().String(),
		name:      name,
		payload:   payload,
		headers:   make(map[string]string),
		timestamp: time.Now(),
		ctx:       ctx,
	}
}

// ID returns the unique event ID
func (e *Event) ID() string {
	return e.id
}

// Topic returns the event topic/stream name
func (e *Event) Topic() string {
	return e.topic
}

// Name returns the event name
func (e *Event) Name() string {
	return e.name
}

// Payload returns the event payload
func (e *Event) Payload() any {
	return e.payload
}

// Headers returns a copy of the event headers for thread-safe access
func (e *Event) Headers() map[string]string {
	e.headersMu.RLock()
	defer e.headersMu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]string, len(e.headers))
	for k, v := range e.headers {
		result[k] = v
	}
	return result
}

// Timestamp returns when the event was created
func (e *Event) Timestamp() time.Time {
	return e.timestamp
}

// Context returns the event context
func (e *Event) Context() context.Context {
	return e.ctx
}

// WithTopic sets the event topic
func (e *Event) WithTopic(topic string) *Event {
	e.topic = topic
	return e
}

// WithHeader sets a header value in a thread-safe manner
func (e *Event) WithHeader(key, value string) *Event {
	e.headersMu.Lock()
	defer e.headersMu.Unlock()
	e.headers[key] = value
	return e
}

// WithHeaders sets multiple headers in a thread-safe manner
func (e *Event) WithHeaders(headers map[string]string) *Event {
	e.headersMu.Lock()
	defer e.headersMu.Unlock()
	for k, v := range headers {
		e.headers[k] = v
	}
	return e
}

// WithContext sets the event context
func (e *Event) WithContext(ctx context.Context) *Event {
	e.ctx = ctx
	return e
}

// internal methods for Redis Streams serialization
func (e *Event) setID(id string) {
	e.id = id
}

func (e *Event) setTimestamp(timestamp time.Time) {
	e.timestamp = timestamp
}
