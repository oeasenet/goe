package event

import (
	"context"
	"encoding/json"
	"fmt"

	"go.oease.dev/goe/v2/contract"
)

// Helper functions for common event operations

// PublishJSON publishes an event with JSON payload
func PublishJSON(ctx context.Context, publisher contract.EventPublisher, topic string, eventName string, payload interface{}) error {
	event := NewEvent(eventName, payload)
	return publisher.Publish(ctx, topic, event)
}

// PublishJSONWithHeaders publishes an event with JSON payload and headers
func PublishJSONWithHeaders(ctx context.Context, publisher contract.EventPublisher, topic string, eventName string, payload interface{}, headers map[string]string) error {
	event := NewEvent(eventName, payload).WithHeaders(headers)
	return publisher.Publish(ctx, topic, event)
}

// UnmarshalEventPayload unmarshals event payload to a struct
func UnmarshalEventPayload(event contract.Event, target interface{}) error {
	payload := event.Payload()

	// If payload is already a json.RawMessage, unmarshal directly
	if rawMessage, ok := payload.(json.RawMessage); ok {
		return json.Unmarshal(rawMessage, target)
	}

	// Otherwise, marshal to JSON first, then unmarshal
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	return json.Unmarshal(data, target)
}

// CreateEventHandler creates an event handler from a function
func CreateEventHandler(fn func(ctx context.Context, event contract.Event) error) contract.EventHandler {
	return contract.EventHandlerFunc(fn)
}

// CreateTypedEventHandler creates a typed event handler that automatically unmarshals the payload
func CreateTypedEventHandler[T any](fn func(ctx context.Context, event contract.Event, payload T) error) contract.EventHandler {
	return contract.EventHandlerFunc(func(ctx context.Context, event contract.Event) error {
		var payload T
		if err := UnmarshalEventPayload(event, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}
		return fn(ctx, event, payload)
	})
}

// BatchPublisher provides a convenient way to publish multiple events
type BatchPublisher struct {
	publisher contract.EventPublisher
	topic     string
	events    []contract.Event
}

// NewBatchPublisher creates a new batch publisher
func NewBatchPublisher(publisher contract.EventPublisher, topic string) *BatchPublisher {
	return &BatchPublisher{
		publisher: publisher,
		topic:     topic,
		events:    make([]contract.Event, 0),
	}
}

// Add adds an event to the batch
func (bp *BatchPublisher) Add(event contract.Event) *BatchPublisher {
	bp.events = append(bp.events, event)
	return bp
}

// AddJSON adds a JSON event to the batch
func (bp *BatchPublisher) AddJSON(eventName string, payload interface{}) *BatchPublisher {
	event := NewEvent(eventName, payload)
	bp.events = append(bp.events, event)
	return bp
}

// AddJSONWithHeaders adds a JSON event with headers to the batch
func (bp *BatchPublisher) AddJSONWithHeaders(eventName string, payload interface{}, headers map[string]string) *BatchPublisher {
	event := NewEvent(eventName, payload).WithHeaders(headers)
	bp.events = append(bp.events, event)
	return bp
}

// Publish publishes all events in the batch
func (bp *BatchPublisher) Publish(ctx context.Context) error {
	if len(bp.events) == 0 {
		return nil
	}

	err := bp.publisher.PublishBatch(ctx, bp.topic, bp.events)
	if err != nil {
		return err
	}

	// Clear the batch
	bp.events = bp.events[:0]
	return nil
}

// Count returns the number of events in the batch
func (bp *BatchPublisher) Count() int {
	return len(bp.events)
}

// Clear clears all events from the batch
func (bp *BatchPublisher) Clear() {
	bp.events = bp.events[:0]
}

// EventPattern provides pattern matching for events
type EventPattern struct {
	Topic     string
	EventName string
	Headers   map[string]string
}

// Matches checks if an event matches the pattern
func (ep *EventPattern) Matches(event contract.Event) bool {
	// Check topic
	if ep.Topic != "" && event.Topic() != ep.Topic {
		return false
	}

	// Check event name
	if ep.EventName != "" && event.Name() != ep.EventName {
		return false
	}

	// Check headers
	eventHeaders := event.Headers()
	for key, value := range ep.Headers {
		if eventHeaders[key] != value {
			return false
		}
	}

	return true
}

// PatternHandler wraps an event handler with pattern matching
type PatternHandler struct {
	pattern *EventPattern
	handler contract.EventHandler
}

// NewPatternHandler creates a new pattern handler
func NewPatternHandler(pattern *EventPattern, handler contract.EventHandler) *PatternHandler {
	return &PatternHandler{
		pattern: pattern,
		handler: handler,
	}
}

// Handle handles an event if it matches the pattern
func (ph *PatternHandler) Handle(ctx context.Context, event contract.Event) error {
	if !ph.pattern.Matches(event) {
		return nil // Skip if pattern doesn't match
	}

	return ph.handler.Handle(ctx, event)
}

// MultiHandler handles multiple event handlers
type MultiHandler struct {
	handlers []contract.EventHandler
}

// NewMultiHandler creates a new multi-handler
func NewMultiHandler(handlers ...contract.EventHandler) *MultiHandler {
	return &MultiHandler{
		handlers: handlers,
	}
}

// Handle handles an event with all handlers
func (mh *MultiHandler) Handle(ctx context.Context, event contract.Event) error {
	for _, handler := range mh.handlers {
		if err := handler.Handle(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Add adds a handler to the multi-handler
func (mh *MultiHandler) Add(handler contract.EventHandler) {
	mh.handlers = append(mh.handlers, handler)
}

// RetryHandler wraps an event handler with retry logic
type RetryHandler struct {
	handler     contract.EventHandler
	maxRetries  int
	shouldRetry func(error) bool
}

// NewRetryHandler creates a new retry handler
func NewRetryHandler(handler contract.EventHandler, maxRetries int, shouldRetry func(error) bool) *RetryHandler {
	return &RetryHandler{
		handler:     handler,
		maxRetries:  maxRetries,
		shouldRetry: shouldRetry,
	}
}

// Handle handles an event with retry logic
func (rh *RetryHandler) Handle(ctx context.Context, event contract.Event) error {
	var lastErr error

	for i := 0; i <= rh.maxRetries; i++ {
		err := rh.handler.Handle(ctx, event)
		if err == nil {
			return nil
		}

		lastErr = err

		if rh.shouldRetry != nil && !rh.shouldRetry(err) {
			return err
		}

		if i < rh.maxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Could add backoff logic here
			}
		}
	}

	return lastErr
}
