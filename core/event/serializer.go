package event

import (
	"encoding/json"
	"fmt"
	"time"

	"go.oease.dev/goe/v2/contract"
)

// EventData represents the serialized event data for Redis Streams
type EventData struct {
	ID        string            `json:"id"`
	Topic     string            `json:"topic"`
	Name      string            `json:"name"`
	Payload   json.RawMessage   `json:"payload"`
	Headers   map[string]string `json:"headers"`
	Timestamp time.Time         `json:"timestamp"`
}

// Serializer handles event serialization/deserialization
type Serializer struct{}

// NewSerializer creates a new serializer
func NewSerializer() *Serializer {
	return &Serializer{}
}

// Serialize converts an event to Redis Stream field-value pairs
func (s *Serializer) Serialize(event contract.Event) (map[string]interface{}, error) {
	payloadBytes, err := json.Marshal(event.Payload())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	headersBytes, err := json.Marshal(event.Headers())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal headers: %w", err)
	}

	return map[string]interface{}{
		"id":        event.ID(),
		"topic":     event.Topic(),
		"name":      event.Name(),
		"payload":   string(payloadBytes),
		"headers":   string(headersBytes),
		"timestamp": event.Timestamp().Format(time.RFC3339Nano),
	}, nil
}

// Deserialize converts Redis Stream field-value pairs to an event
func (s *Serializer) Deserialize(fields map[string]interface{}) (contract.Event, error) {
	eventData := &EventData{
		Headers: make(map[string]string),
	}

	// Extract basic fields
	if id, ok := fields["id"].(string); ok {
		eventData.ID = id
	} else {
		return nil, fmt.Errorf("missing or invalid event id")
	}

	if topic, ok := fields["topic"].(string); ok {
		eventData.Topic = topic
	} else {
		return nil, fmt.Errorf("missing or invalid event topic")
	}

	if name, ok := fields["name"].(string); ok {
		eventData.Name = name
	} else {
		return nil, fmt.Errorf("missing or invalid event name")
	}

	// Parse timestamp
	if timestampStr, ok := fields["timestamp"].(string); ok {
		timestamp, err := time.Parse(time.RFC3339Nano, timestampStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp: %w", err)
		}
		eventData.Timestamp = timestamp
	} else {
		return nil, fmt.Errorf("missing or invalid event timestamp")
	}

	// Parse payload
	if payloadStr, ok := fields["payload"].(string); ok {
		eventData.Payload = json.RawMessage(payloadStr)
	} else {
		return nil, fmt.Errorf("missing or invalid event payload")
	}

	// Parse headers
	if headersStr, ok := fields["headers"].(string); ok {
		if err := json.Unmarshal([]byte(headersStr), &eventData.Headers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
		}
	}

	// Create event
	event := NewEvent(eventData.Name, eventData.Payload)
	event.setID(eventData.ID)
	event.setTimestamp(eventData.Timestamp)
	event.WithTopic(eventData.Topic)
	event.WithHeaders(eventData.Headers)

	return event, nil
}

// SerializeBatch serializes multiple events
func (s *Serializer) SerializeBatch(events []contract.Event) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, len(events))
	for i, event := range events {
		serialized, err := s.Serialize(event)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize event %d: %w", i, err)
		}
		result[i] = serialized
	}
	return result, nil
}
