package event

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEvent(t *testing.T) {
	tests := []struct {
		name        string
		eventName   string
		payload     any
		wantNilID   bool
		wantHeaders bool
	}{
		{
			name:        "basic event with string payload",
			eventName:   "user.created",
			payload:     "test payload",
			wantNilID:   false,
			wantHeaders: true,
		},
		{
			name:        "event with struct payload",
			eventName:   "order.placed",
			payload:     struct{ ID string }{ID: "123"},
			wantNilID:   false,
			wantHeaders: true,
		},
		{
			name:        "event with nil payload",
			eventName:   "system.ping",
			payload:     nil,
			wantNilID:   false,
			wantHeaders: true,
		},
		{
			name:        "event with map payload",
			eventName:   "data.updated",
			payload:     map[string]interface{}{"key": "value"},
			wantNilID:   false,
			wantHeaders: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewEvent(tt.eventName, tt.payload)

			require.NotNil(t, event)
			assert.NotEmpty(t, event.ID(), "ID should be generated")
			assert.Equal(t, tt.eventName, event.Name())
			assert.Equal(t, tt.payload, event.Payload())
			assert.NotNil(t, event.Headers(), "Headers should be initialized")
			assert.Empty(t, event.Headers(), "Headers should be empty initially")
			assert.False(t, event.Timestamp().IsZero(), "Timestamp should be set")
			assert.NotNil(t, event.Context(), "Context should be set")
			assert.Empty(t, event.Topic(), "Topic should be empty initially")
		})
	}
}

func TestNewEventWithContext(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		eventName string
		payload   any
	}{
		{
			name:      "with background context",
			ctx:       context.Background(),
			eventName: "test.event",
			payload:   "data",
		},
		{
			name:      "with cancel context",
			ctx:       func() context.Context { ctx, _ := context.WithCancel(context.Background()); return ctx }(),
			eventName: "cancel.event",
			payload:   123,
		},
		{
			name:      "with timeout context",
			ctx:       func() context.Context { ctx, _ := context.WithTimeout(context.Background(), time.Hour); return ctx }(),
			eventName: "timeout.event",
			payload:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewEventWithContext(tt.ctx, tt.eventName, tt.payload)

			require.NotNil(t, event)
			assert.NotEmpty(t, event.ID())
			assert.Equal(t, tt.eventName, event.Name())
			assert.Equal(t, tt.payload, event.Payload())
			assert.Equal(t, tt.ctx, event.Context())
		})
	}
}

func TestEvent_Getters(t *testing.T) {
	event := NewEvent("test.event", "payload")

	t.Run("ID returns valid UUID", func(t *testing.T) {
		id := event.ID()
		assert.NotEmpty(t, id)
		assert.Len(t, id, 36, "UUID should be 36 characters")
	})

	t.Run("Name returns event name", func(t *testing.T) {
		assert.Equal(t, "test.event", event.Name())
	})

	t.Run("Payload returns payload", func(t *testing.T) {
		assert.Equal(t, "payload", event.Payload())
	})

	t.Run("Headers returns empty map initially", func(t *testing.T) {
		headers := event.Headers()
		assert.NotNil(t, headers)
		assert.Empty(t, headers)
	})

	t.Run("Timestamp returns valid time", func(t *testing.T) {
		ts := event.Timestamp()
		assert.False(t, ts.IsZero())
		assert.True(t, ts.Before(time.Now().Add(time.Second)))
	})

	t.Run("Context returns context", func(t *testing.T) {
		ctx := event.Context()
		assert.NotNil(t, ctx)
	})

	t.Run("Topic returns empty initially", func(t *testing.T) {
		assert.Empty(t, event.Topic())
	})
}

func TestEvent_WithTopic(t *testing.T) {
	tests := []struct {
		name  string
		topic string
	}{
		{name: "simple topic", topic: "orders"},
		{name: "namespaced topic", topic: "orders.created"},
		{name: "empty topic", topic: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewEvent("test", nil)
			result := event.WithTopic(tt.topic)

			assert.Same(t, event, result, "Should return same instance for chaining")
			assert.Equal(t, tt.topic, event.Topic())
		})
	}
}

func TestEvent_WithHeader(t *testing.T) {
	t.Run("single header", func(t *testing.T) {
		event := NewEvent("test", nil)
		result := event.WithHeader("key1", "value1")

		assert.Same(t, event, result)
		assert.Equal(t, "value1", event.Headers()["key1"])
	})

	t.Run("multiple headers chained", func(t *testing.T) {
		event := NewEvent("test", nil)
		event.WithHeader("key1", "value1").WithHeader("key2", "value2")

		assert.Equal(t, "value1", event.Headers()["key1"])
		assert.Equal(t, "value2", event.Headers()["key2"])
	})

	t.Run("overwrite existing header", func(t *testing.T) {
		event := NewEvent("test", nil)
		event.WithHeader("key", "old").WithHeader("key", "new")

		assert.Equal(t, "new", event.Headers()["key"])
	})
}

func TestEvent_WithHeaders(t *testing.T) {
	t.Run("multiple headers at once", func(t *testing.T) {
		event := NewEvent("test", nil)
		headers := map[string]string{
			"Content-Type": "application/json",
			"X-Request-ID": "123",
		}
		result := event.WithHeaders(headers)

		assert.Same(t, event, result)
		assert.Equal(t, "application/json", event.Headers()["Content-Type"])
		assert.Equal(t, "123", event.Headers()["X-Request-ID"])
	})

	t.Run("merge with existing headers", func(t *testing.T) {
		event := NewEvent("test", nil)
		event.WithHeader("existing", "value")
		event.WithHeaders(map[string]string{"new": "value"})

		assert.Equal(t, "value", event.Headers()["existing"])
		assert.Equal(t, "value", event.Headers()["new"])
	})

	t.Run("empty headers map", func(t *testing.T) {
		event := NewEvent("test", nil)
		event.WithHeader("existing", "value")
		event.WithHeaders(map[string]string{})

		assert.Equal(t, "value", event.Headers()["existing"])
	})

	t.Run("nil headers map", func(t *testing.T) {
		event := NewEvent("test", nil)
		event.WithHeader("existing", "value")
		event.WithHeaders(nil)

		assert.Equal(t, "value", event.Headers()["existing"])
	})
}

func TestEvent_WithContext(t *testing.T) {
	t.Run("replace context", func(t *testing.T) {
		event := NewEvent("test", nil)
		newCtx := context.WithValue(context.Background(), "key", "value")
		result := event.WithContext(newCtx)

		assert.Same(t, event, result)
		assert.Equal(t, newCtx, event.Context())
	})
}

func TestEvent_InternalMethods(t *testing.T) {
	t.Run("setID updates ID", func(t *testing.T) {
		event := NewEvent("test", nil)
		originalID := event.ID()
		event.setID("custom-id-123")

		assert.NotEqual(t, originalID, event.ID())
		assert.Equal(t, "custom-id-123", event.ID())
	})

	t.Run("setTimestamp updates timestamp", func(t *testing.T) {
		event := NewEvent("test", nil)
		customTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		event.setTimestamp(customTime)

		assert.Equal(t, customTime, event.Timestamp())
	})
}

func TestEvent_FluentChaining(t *testing.T) {
	t.Run("full chain", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "key", "value")

		event := NewEvent("order.created", map[string]string{"id": "123"}).
			WithTopic("orders").
			WithHeader("X-Request-ID", "req-123").
			WithHeaders(map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "tenant-1",
			}).
			WithContext(ctx)

		assert.Equal(t, "order.created", event.Name())
		assert.Equal(t, "orders", event.Topic())
		assert.Equal(t, "req-123", event.Headers()["X-Request-ID"])
		assert.Equal(t, "application/json", event.Headers()["Content-Type"])
		assert.Equal(t, "tenant-1", event.Headers()["X-Tenant-ID"])
		assert.Equal(t, ctx, event.Context())
	})
}

func TestEvent_UniqueIDs(t *testing.T) {
	t.Run("each event gets unique ID", func(t *testing.T) {
		ids := make(map[string]bool)
		for i := 0; i < 100; i++ {
			event := NewEvent("test", nil)
			assert.False(t, ids[event.ID()], "ID should be unique")
			ids[event.ID()] = true
		}
	})
}
