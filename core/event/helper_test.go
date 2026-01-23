package event

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
)

// mockPublisher implements contract.EventPublisher for testing
type mockPublisher struct {
	publishedEvents []struct {
		topic string
		event contract.Event
		delay time.Duration
	}
	publishErr error
	batchErr   error
}

func newMockPublisher() *mockPublisher {
	return &mockPublisher{}
}

func (m *mockPublisher) Publish(ctx context.Context, topic string, event contract.Event) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.publishedEvents = append(m.publishedEvents, struct {
		topic string
		event contract.Event
		delay time.Duration
	}{topic: topic, event: event, delay: 0})
	return nil
}

func (m *mockPublisher) PublishWithDelay(ctx context.Context, topic string, event contract.Event, delay time.Duration) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.publishedEvents = append(m.publishedEvents, struct {
		topic string
		event contract.Event
		delay time.Duration
	}{topic: topic, event: event, delay: delay})
	return nil
}

func (m *mockPublisher) PublishBatch(ctx context.Context, topic string, events []contract.Event) error {
	if m.batchErr != nil {
		return m.batchErr
	}
	for _, event := range events {
		m.publishedEvents = append(m.publishedEvents, struct {
			topic string
			event contract.Event
			delay time.Duration
		}{topic: topic, event: event, delay: 0})
	}
	return nil
}

func TestPublishJSON(t *testing.T) {
	t.Run("publishes event successfully", func(t *testing.T) {
		publisher := newMockPublisher()
		payload := map[string]string{"key": "value"}

		err := PublishJSON(context.Background(), publisher, "test-topic", "test.event", payload)

		require.NoError(t, err)
		require.Len(t, publisher.publishedEvents, 1)
		assert.Equal(t, "test-topic", publisher.publishedEvents[0].topic)
		assert.Equal(t, "test.event", publisher.publishedEvents[0].event.Name())
	})

	t.Run("returns error on publish failure", func(t *testing.T) {
		publisher := newMockPublisher()
		publisher.publishErr = errors.New("publish failed")

		err := PublishJSON(context.Background(), publisher, "topic", "event", nil)

		assert.Error(t, err)
		assert.Equal(t, "publish failed", err.Error())
	})
}

func TestPublishJSONWithHeaders(t *testing.T) {
	t.Run("publishes event with headers", func(t *testing.T) {
		publisher := newMockPublisher()
		headers := map[string]string{
			"X-Request-ID": "123",
			"X-Tenant-ID":  "tenant-1",
		}

		err := PublishJSONWithHeaders(context.Background(), publisher, "topic", "event", "payload", headers)

		require.NoError(t, err)
		require.Len(t, publisher.publishedEvents, 1)
		assert.Equal(t, "123", publisher.publishedEvents[0].event.Headers()["X-Request-ID"])
		assert.Equal(t, "tenant-1", publisher.publishedEvents[0].event.Headers()["X-Tenant-ID"])
	})
}

func TestUnmarshalEventPayload(t *testing.T) {
	type TestPayload struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	t.Run("unmarshal from json.RawMessage", func(t *testing.T) {
		rawPayload := json.RawMessage(`{"name":"test","value":42}`)
		event := NewEvent("test", rawPayload)

		var result TestPayload
		err := UnmarshalEventPayload(event, &result)

		require.NoError(t, err)
		assert.Equal(t, "test", result.Name)
		assert.Equal(t, 42, result.Value)
	})

	t.Run("unmarshal from struct", func(t *testing.T) {
		payload := TestPayload{Name: "original", Value: 100}
		event := NewEvent("test", payload)

		var result TestPayload
		err := UnmarshalEventPayload(event, &result)

		require.NoError(t, err)
		assert.Equal(t, "original", result.Name)
		assert.Equal(t, 100, result.Value)
	})

	t.Run("unmarshal from map", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":  "mapped",
			"value": float64(200), // JSON numbers are float64
		}
		event := NewEvent("test", payload)

		var result TestPayload
		err := UnmarshalEventPayload(event, &result)

		require.NoError(t, err)
		assert.Equal(t, "mapped", result.Name)
		assert.Equal(t, 200, result.Value)
	})

	t.Run("invalid json.RawMessage returns error", func(t *testing.T) {
		rawPayload := json.RawMessage(`invalid json`)
		event := NewEvent("test", rawPayload)

		var result TestPayload
		err := UnmarshalEventPayload(event, &result)

		assert.Error(t, err)
	})
}

func TestCreateEventHandler(t *testing.T) {
	t.Run("creates handler from function", func(t *testing.T) {
		called := false
		fn := func(ctx context.Context, event contract.Event) error {
			called = true
			assert.Equal(t, "test.event", event.Name())
			return nil
		}

		handler := CreateEventHandler(fn)
		event := NewEvent("test.event", nil)

		err := handler.Handle(context.Background(), event)

		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("propagates error from function", func(t *testing.T) {
		expectedErr := errors.New("handler error")
		fn := func(ctx context.Context, event contract.Event) error {
			return expectedErr
		}

		handler := CreateEventHandler(fn)

		err := handler.Handle(context.Background(), NewEvent("test", nil))

		assert.Equal(t, expectedErr, err)
	})
}

func TestCreateTypedEventHandler(t *testing.T) {
	type OrderPayload struct {
		OrderID string  `json:"order_id"`
		Amount  float64 `json:"amount"`
	}

	t.Run("unmarshals payload and calls handler", func(t *testing.T) {
		var receivedPayload OrderPayload
		handler := CreateTypedEventHandler(func(ctx context.Context, event contract.Event, payload OrderPayload) error {
			receivedPayload = payload
			return nil
		})

		eventPayload := map[string]interface{}{
			"order_id": "order-123",
			"amount":   99.99,
		}
		event := NewEvent("order.created", eventPayload)

		err := handler.Handle(context.Background(), event)

		require.NoError(t, err)
		assert.Equal(t, "order-123", receivedPayload.OrderID)
		assert.Equal(t, 99.99, receivedPayload.Amount)
	})

	t.Run("returns error on unmarshal failure", func(t *testing.T) {
		handler := CreateTypedEventHandler(func(ctx context.Context, event contract.Event, payload OrderPayload) error {
			return nil
		})

		// Create event with invalid payload that can't unmarshal to OrderPayload
		event := NewEvent("test", json.RawMessage(`invalid`))

		err := handler.Handle(context.Background(), event)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal payload")
	})
}

func TestBatchPublisher(t *testing.T) {
	t.Run("NewBatchPublisher creates empty publisher", func(t *testing.T) {
		publisher := newMockPublisher()
		batch := NewBatchPublisher(publisher, "orders")

		assert.Equal(t, 0, batch.Count())
	})

	t.Run("Add adds event to batch", func(t *testing.T) {
		publisher := newMockPublisher()
		batch := NewBatchPublisher(publisher, "orders")

		event := NewEvent("order.created", nil)
		result := batch.Add(event)

		assert.Same(t, batch, result, "Should return same instance for chaining")
		assert.Equal(t, 1, batch.Count())
	})

	t.Run("AddJSON adds JSON event", func(t *testing.T) {
		publisher := newMockPublisher()
		batch := NewBatchPublisher(publisher, "orders")

		batch.AddJSON("order.created", map[string]string{"id": "123"})

		assert.Equal(t, 1, batch.Count())
	})

	t.Run("AddJSONWithHeaders adds event with headers", func(t *testing.T) {
		publisher := newMockPublisher()
		batch := NewBatchPublisher(publisher, "orders")

		headers := map[string]string{"X-Request-ID": "456"}
		batch.AddJSONWithHeaders("order.created", "payload", headers)

		assert.Equal(t, 1, batch.Count())
	})

	t.Run("Publish publishes all events", func(t *testing.T) {
		publisher := newMockPublisher()
		batch := NewBatchPublisher(publisher, "orders")

		batch.AddJSON("event1", nil)
		batch.AddJSON("event2", nil)
		batch.AddJSON("event3", nil)

		err := batch.Publish(context.Background())

		require.NoError(t, err)
		assert.Len(t, publisher.publishedEvents, 3)
		assert.Equal(t, 0, batch.Count(), "Batch should be cleared after publish")
	})

	t.Run("Publish with empty batch does nothing", func(t *testing.T) {
		publisher := newMockPublisher()
		batch := NewBatchPublisher(publisher, "orders")

		err := batch.Publish(context.Background())

		require.NoError(t, err)
		assert.Empty(t, publisher.publishedEvents)
	})

	t.Run("Publish returns error on failure", func(t *testing.T) {
		publisher := newMockPublisher()
		publisher.batchErr = errors.New("batch publish failed")
		batch := NewBatchPublisher(publisher, "orders")
		batch.AddJSON("event", nil)

		err := batch.Publish(context.Background())

		assert.Error(t, err)
	})

	t.Run("Clear removes all events", func(t *testing.T) {
		publisher := newMockPublisher()
		batch := NewBatchPublisher(publisher, "orders")

		batch.AddJSON("event1", nil)
		batch.AddJSON("event2", nil)
		assert.Equal(t, 2, batch.Count())

		batch.Clear()

		assert.Equal(t, 0, batch.Count())
	})

	t.Run("chained operations", func(t *testing.T) {
		publisher := newMockPublisher()
		batch := NewBatchPublisher(publisher, "orders")

		batch.
			AddJSON("event1", nil).
			AddJSON("event2", nil).
			AddJSONWithHeaders("event3", nil, map[string]string{"key": "value"})

		assert.Equal(t, 3, batch.Count())
	})
}

func TestEventPattern(t *testing.T) {
	t.Run("Matches returns true for matching event", func(t *testing.T) {
		pattern := &EventPattern{
			Topic:     "orders",
			EventName: "order.created",
		}

		event := NewEvent("order.created", nil).WithTopic("orders")

		assert.True(t, pattern.Matches(event))
	})

	t.Run("Matches returns false for non-matching topic", func(t *testing.T) {
		pattern := &EventPattern{
			Topic:     "orders",
			EventName: "order.created",
		}

		event := NewEvent("order.created", nil).WithTopic("inventory")

		assert.False(t, pattern.Matches(event))
	})

	t.Run("Matches returns false for non-matching event name", func(t *testing.T) {
		pattern := &EventPattern{
			Topic:     "orders",
			EventName: "order.created",
		}

		event := NewEvent("order.updated", nil).WithTopic("orders")

		assert.False(t, pattern.Matches(event))
	})

	t.Run("Matches checks headers", func(t *testing.T) {
		pattern := &EventPattern{
			Headers: map[string]string{"X-Tenant-ID": "tenant-1"},
		}

		matchingEvent := NewEvent("test", nil).WithHeader("X-Tenant-ID", "tenant-1")
		nonMatchingEvent := NewEvent("test", nil).WithHeader("X-Tenant-ID", "tenant-2")
		missingHeaderEvent := NewEvent("test", nil)

		assert.True(t, pattern.Matches(matchingEvent))
		assert.False(t, pattern.Matches(nonMatchingEvent))
		assert.False(t, pattern.Matches(missingHeaderEvent))
	})

	t.Run("Empty pattern matches all events", func(t *testing.T) {
		pattern := &EventPattern{}

		event := NewEvent("any.event", nil).WithTopic("any-topic")

		assert.True(t, pattern.Matches(event))
	})
}

func TestPatternHandler(t *testing.T) {
	t.Run("calls handler when pattern matches", func(t *testing.T) {
		called := false
		innerHandler := CreateEventHandler(func(ctx context.Context, event contract.Event) error {
			called = true
			return nil
		})

		pattern := &EventPattern{EventName: "order.created"}
		handler := NewPatternHandler(pattern, innerHandler)

		event := NewEvent("order.created", nil)
		err := handler.Handle(context.Background(), event)

		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("skips handler when pattern does not match", func(t *testing.T) {
		called := false
		innerHandler := CreateEventHandler(func(ctx context.Context, event contract.Event) error {
			called = true
			return nil
		})

		pattern := &EventPattern{EventName: "order.created"}
		handler := NewPatternHandler(pattern, innerHandler)

		event := NewEvent("order.updated", nil)
		err := handler.Handle(context.Background(), event)

		require.NoError(t, err)
		assert.False(t, called)
	})
}

func TestMultiHandler(t *testing.T) {
	t.Run("calls all handlers in order", func(t *testing.T) {
		var callOrder []int

		handler1 := CreateEventHandler(func(ctx context.Context, event contract.Event) error {
			callOrder = append(callOrder, 1)
			return nil
		})
		handler2 := CreateEventHandler(func(ctx context.Context, event contract.Event) error {
			callOrder = append(callOrder, 2)
			return nil
		})
		handler3 := CreateEventHandler(func(ctx context.Context, event contract.Event) error {
			callOrder = append(callOrder, 3)
			return nil
		})

		multi := NewMultiHandler(handler1, handler2, handler3)
		err := multi.Handle(context.Background(), NewEvent("test", nil))

		require.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, callOrder)
	})

	t.Run("stops on first error", func(t *testing.T) {
		expectedErr := errors.New("handler 2 failed")
		var callOrder []int

		handler1 := CreateEventHandler(func(ctx context.Context, event contract.Event) error {
			callOrder = append(callOrder, 1)
			return nil
		})
		handler2 := CreateEventHandler(func(ctx context.Context, event contract.Event) error {
			callOrder = append(callOrder, 2)
			return expectedErr
		})
		handler3 := CreateEventHandler(func(ctx context.Context, event contract.Event) error {
			callOrder = append(callOrder, 3)
			return nil
		})

		multi := NewMultiHandler(handler1, handler2, handler3)
		err := multi.Handle(context.Background(), NewEvent("test", nil))

		assert.Equal(t, expectedErr, err)
		assert.Equal(t, []int{1, 2}, callOrder)
	})

	t.Run("Add appends handler", func(t *testing.T) {
		var callOrder []int

		multi := NewMultiHandler()
		multi.Add(CreateEventHandler(func(ctx context.Context, event contract.Event) error {
			callOrder = append(callOrder, 1)
			return nil
		}))
		multi.Add(CreateEventHandler(func(ctx context.Context, event contract.Event) error {
			callOrder = append(callOrder, 2)
			return nil
		}))

		err := multi.Handle(context.Background(), NewEvent("test", nil))

		require.NoError(t, err)
		assert.Equal(t, []int{1, 2}, callOrder)
	})
}

func TestRetryHandler(t *testing.T) {
	t.Run("succeeds on first try", func(t *testing.T) {
		attempts := 0
		innerHandler := CreateEventHandler(func(ctx context.Context, event contract.Event) error {
			attempts++
			return nil
		})

		handler := NewRetryHandler(innerHandler, 3, nil)
		err := handler.Handle(context.Background(), NewEvent("test", nil))

		require.NoError(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("retries on failure", func(t *testing.T) {
		attempts := 0
		innerHandler := CreateEventHandler(func(ctx context.Context, event contract.Event) error {
			attempts++
			if attempts < 3 {
				return errors.New("temporary error")
			}
			return nil
		})

		handler := NewRetryHandler(innerHandler, 5, nil)
		err := handler.Handle(context.Background(), NewEvent("test", nil))

		require.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})

	t.Run("returns error after max retries", func(t *testing.T) {
		attempts := 0
		innerHandler := CreateEventHandler(func(ctx context.Context, event contract.Event) error {
			attempts++
			return errors.New("persistent error")
		})

		handler := NewRetryHandler(innerHandler, 3, nil)
		err := handler.Handle(context.Background(), NewEvent("test", nil))

		assert.Error(t, err)
		assert.Equal(t, 4, attempts) // initial + 3 retries
	})

	t.Run("respects shouldRetry function", func(t *testing.T) {
		attempts := 0
		permanentErr := errors.New("permanent error")
		innerHandler := CreateEventHandler(func(ctx context.Context, event contract.Event) error {
			attempts++
			return permanentErr
		})

		shouldRetry := func(err error) bool {
			return err.Error() != "permanent error"
		}

		handler := NewRetryHandler(innerHandler, 5, shouldRetry)
		err := handler.Handle(context.Background(), NewEvent("test", nil))

		assert.Equal(t, permanentErr, err)
		assert.Equal(t, 1, attempts, "Should not retry for permanent errors")
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		attempts := 0

		innerHandler := CreateEventHandler(func(ctx context.Context, event contract.Event) error {
			attempts++
			if attempts == 2 {
				cancel()
			}
			return errors.New("retry")
		})

		handler := NewRetryHandler(innerHandler, 10, nil)
		err := handler.Handle(ctx, NewEvent("test", nil))

		assert.Equal(t, context.Canceled, err)
	})
}
