package event

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
)

func TestNewSerializer(t *testing.T) {
	serializer := NewSerializer()
	require.NotNil(t, serializer)
}

func TestSerializer_Serialize(t *testing.T) {
	serializer := NewSerializer()

	t.Run("serializes basic event", func(t *testing.T) {
		event := NewEvent("test.event", map[string]string{"key": "value"}).
			WithTopic("test-topic").
			WithHeader("X-Request-ID", "123")

		result, err := serializer.Serialize(event)

		require.NoError(t, err)
		assert.Equal(t, event.ID(), result["id"])
		assert.Equal(t, "test-topic", result["topic"])
		assert.Equal(t, "test.event", result["name"])
		assert.NotEmpty(t, result["payload"])
		assert.NotEmpty(t, result["headers"])
		assert.NotEmpty(t, result["timestamp"])
	})

	t.Run("serializes nil payload", func(t *testing.T) {
		event := NewEvent("test.event", nil)

		result, err := serializer.Serialize(event)

		require.NoError(t, err)
		assert.Equal(t, "null", result["payload"])
	})

	t.Run("serializes complex payload", func(t *testing.T) {
		payload := struct {
			Name   string   `json:"name"`
			Values []int    `json:"values"`
			Nested struct{} `json:"nested"`
		}{
			Name:   "test",
			Values: []int{1, 2, 3},
		}
		event := NewEvent("complex.event", payload)

		result, err := serializer.Serialize(event)

		require.NoError(t, err)

		// Verify payload can be parsed back
		var parsed map[string]interface{}
		err = json.Unmarshal([]byte(result["payload"].(string)), &parsed)
		require.NoError(t, err)
		assert.Equal(t, "test", parsed["name"])
	})

	t.Run("serializes empty headers", func(t *testing.T) {
		event := NewEvent("test.event", nil)

		result, err := serializer.Serialize(event)

		require.NoError(t, err)
		assert.Equal(t, "{}", result["headers"])
	})

	t.Run("serializes multiple headers", func(t *testing.T) {
		event := NewEvent("test.event", nil).
			WithHeader("Header1", "Value1").
			WithHeader("Header2", "Value2")

		result, err := serializer.Serialize(event)

		require.NoError(t, err)

		var headers map[string]string
		err = json.Unmarshal([]byte(result["headers"].(string)), &headers)
		require.NoError(t, err)
		assert.Equal(t, "Value1", headers["Header1"])
		assert.Equal(t, "Value2", headers["Header2"])
	})

	t.Run("timestamp format is RFC3339Nano", func(t *testing.T) {
		event := NewEvent("test.event", nil)

		result, err := serializer.Serialize(event)

		require.NoError(t, err)

		timestampStr := result["timestamp"].(string)
		_, err = time.Parse(time.RFC3339Nano, timestampStr)
		require.NoError(t, err, "timestamp should be RFC3339Nano format")
	})
}

func TestSerializer_Deserialize(t *testing.T) {
	serializer := NewSerializer()

	t.Run("deserializes valid event", func(t *testing.T) {
		timestamp := time.Now().UTC().Truncate(time.Nanosecond)
		fields := map[string]interface{}{
			"id":        "event-123",
			"topic":     "orders",
			"name":      "order.created",
			"payload":   `{"order_id":"123"}`,
			"headers":   `{"X-Request-ID":"req-456"}`,
			"timestamp": timestamp.Format(time.RFC3339Nano),
		}

		event, err := serializer.Deserialize(fields)

		require.NoError(t, err)
		assert.Equal(t, "event-123", event.ID())
		assert.Equal(t, "orders", event.Topic())
		assert.Equal(t, "order.created", event.Name())
		assert.Equal(t, "req-456", event.Headers()["X-Request-ID"])
		assert.Equal(t, timestamp.Format(time.RFC3339Nano), event.Timestamp().Format(time.RFC3339Nano))
	})

	t.Run("deserializes payload as json.RawMessage", func(t *testing.T) {
		fields := map[string]interface{}{
			"id":        "event-123",
			"topic":     "test",
			"name":      "test.event",
			"payload":   `{"key":"value"}`,
			"headers":   `{}`,
			"timestamp": time.Now().Format(time.RFC3339Nano),
		}

		event, err := serializer.Deserialize(fields)

		require.NoError(t, err)
		payload := event.Payload()
		assert.IsType(t, json.RawMessage{}, payload)
	})

	t.Run("error on missing id", func(t *testing.T) {
		fields := map[string]interface{}{
			"topic":     "test",
			"name":      "test.event",
			"payload":   `{}`,
			"headers":   `{}`,
			"timestamp": time.Now().Format(time.RFC3339Nano),
		}

		_, err := serializer.Deserialize(fields)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id")
	})

	t.Run("error on missing topic", func(t *testing.T) {
		fields := map[string]interface{}{
			"id":        "123",
			"name":      "test.event",
			"payload":   `{}`,
			"headers":   `{}`,
			"timestamp": time.Now().Format(time.RFC3339Nano),
		}

		_, err := serializer.Deserialize(fields)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "topic")
	})

	t.Run("error on missing name", func(t *testing.T) {
		fields := map[string]interface{}{
			"id":        "123",
			"topic":     "test",
			"payload":   `{}`,
			"headers":   `{}`,
			"timestamp": time.Now().Format(time.RFC3339Nano),
		}

		_, err := serializer.Deserialize(fields)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("error on missing timestamp", func(t *testing.T) {
		fields := map[string]interface{}{
			"id":      "123",
			"topic":   "test",
			"name":    "test.event",
			"payload": `{}`,
			"headers": `{}`,
		}

		_, err := serializer.Deserialize(fields)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timestamp")
	})

	t.Run("error on invalid timestamp format", func(t *testing.T) {
		fields := map[string]interface{}{
			"id":        "123",
			"topic":     "test",
			"name":      "test.event",
			"payload":   `{}`,
			"headers":   `{}`,
			"timestamp": "invalid-timestamp",
		}

		_, err := serializer.Deserialize(fields)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse timestamp")
	})

	t.Run("error on missing payload", func(t *testing.T) {
		fields := map[string]interface{}{
			"id":        "123",
			"topic":     "test",
			"name":      "test.event",
			"headers":   `{}`,
			"timestamp": time.Now().Format(time.RFC3339Nano),
		}

		_, err := serializer.Deserialize(fields)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "payload")
	})

	t.Run("error on invalid headers JSON", func(t *testing.T) {
		fields := map[string]interface{}{
			"id":        "123",
			"topic":     "test",
			"name":      "test.event",
			"payload":   `{}`,
			"headers":   `{invalid json`,
			"timestamp": time.Now().Format(time.RFC3339Nano),
		}

		_, err := serializer.Deserialize(fields)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "headers")
	})

	t.Run("handles missing headers gracefully", func(t *testing.T) {
		fields := map[string]interface{}{
			"id":        "123",
			"topic":     "test",
			"name":      "test.event",
			"payload":   `{}`,
			"timestamp": time.Now().Format(time.RFC3339Nano),
		}

		event, err := serializer.Deserialize(fields)

		require.NoError(t, err)
		assert.Empty(t, event.Headers())
	})
}

func TestSerializer_RoundTrip(t *testing.T) {
	serializer := NewSerializer()

	t.Run("serialize and deserialize preserves data", func(t *testing.T) {
		original := NewEvent("order.created", map[string]interface{}{
			"order_id": "ord-123",
			"amount":   99.99,
		}).
			WithTopic("orders").
			WithHeader("X-Request-ID", "req-456").
			WithHeader("X-Tenant-ID", "tenant-789")

		// Serialize
		fields, err := serializer.Serialize(original)
		require.NoError(t, err)

		// Deserialize
		restored, err := serializer.Deserialize(fields)
		require.NoError(t, err)

		// Verify
		assert.Equal(t, original.ID(), restored.ID())
		assert.Equal(t, original.Topic(), restored.Topic())
		assert.Equal(t, original.Name(), restored.Name())
		assert.Equal(t, original.Headers()["X-Request-ID"], restored.Headers()["X-Request-ID"])
		assert.Equal(t, original.Headers()["X-Tenant-ID"], restored.Headers()["X-Tenant-ID"])
		assert.Equal(t, original.Timestamp().Format(time.RFC3339Nano), restored.Timestamp().Format(time.RFC3339Nano))
	})
}

func TestSerializer_SerializeBatch(t *testing.T) {
	serializer := NewSerializer()

	t.Run("serializes multiple events", func(t *testing.T) {
		events := []contract.Event{
			NewEvent("event1", "payload1").WithTopic("topic1"),
			NewEvent("event2", "payload2").WithTopic("topic2"),
			NewEvent("event3", "payload3").WithTopic("topic3"),
		}

		result, err := serializer.SerializeBatch(events)

		require.NoError(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, "event1", result[0]["name"])
		assert.Equal(t, "event2", result[1]["name"])
		assert.Equal(t, "event3", result[2]["name"])
	})

	t.Run("returns empty slice for empty input", func(t *testing.T) {
		result, err := serializer.SerializeBatch(nil)

		require.NoError(t, err)
		assert.Empty(t, result)
	})
}
