package event

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.oease.dev/goe/v2/contract"
)

// DeadLetterQueue manages dead letter queues for failed messages
type DeadLetterQueue struct {
	client     *redis.Client
	config     *Config
	logger     contract.Logger
	serializer *Serializer
}

// NewDeadLetterQueue creates a new dead letter queue manager
func NewDeadLetterQueue(client *redis.Client, config *Config, logger contract.Logger) *DeadLetterQueue {
	return &DeadLetterQueue{
		client:     client,
		config:     config,
		logger:     logger,
		serializer: NewSerializer(),
	}
}

// AddMessage adds a failed message to the dead letter queue
func (dlq *DeadLetterQueue) AddMessage(ctx context.Context, topic string, messageID string, fields map[string]interface{}, err error) error {
	dlqKey := dlq.getDLQKey(topic)

	// Create DLQ entry
	dlqEntry := map[string]interface{}{
		"original_message_id": messageID,
		"topic":               topic,
		"error":               err.Error(),
		"timestamp":           time.Now().Format(time.RFC3339Nano),
		"retry_count":         0,
	}

	// Add original message fields
	for k, v := range fields {
		dlqEntry[k] = v
	}

	// Add to DLQ stream
	_, addErr := dlq.client.XAdd(ctx, &redis.XAddArgs{
		Stream: dlqKey,
		ID:     "*",
		Values: dlqEntry,
	}).Result()

	if addErr != nil {
		return fmt.Errorf("failed to add message to DLQ: %w", addErr)
	}

	// Set expiration for DLQ if configured
	if dlq.config.DeadLetterQueueTTL > 0 {
		dlq.client.Expire(ctx, dlqKey, dlq.config.DeadLetterQueueTTL)
	}

	dlq.logger.Debug("Message added to DLQ",
		"topic", topic,
		"message_id", messageID,
		"error", err.Error(),
	)

	return nil
}

// GetMessages retrieves messages from the dead letter queue
func (dlq *DeadLetterQueue) GetMessages(ctx context.Context, topic string, limit int) ([]contract.Event, error) {
	dlqKey := dlq.getDLQKey(topic)

	// Read messages from DLQ
	messages, err := dlq.client.XRange(ctx, dlqKey, "-", "+").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read DLQ messages: %w", err)
	}

	// Limit results
	if limit > 0 && len(messages) > limit {
		messages = messages[:limit]
	}

	events := make([]contract.Event, 0, len(messages))
	for _, message := range messages {
		event, err := dlq.deserializeDLQMessage(message)
		if err != nil {
			dlq.logger.Error("Failed to deserialize DLQ message",
				"topic", topic,
				"message_id", message.ID,
				"error", err,
			)
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

// ReplayMessage replays a message from the dead letter queue
func (dlq *DeadLetterQueue) ReplayMessage(ctx context.Context, topic string, messageID string) error {
	dlqKey := dlq.getDLQKey(topic)

	// Get the specific message
	messages, err := dlq.client.XRange(ctx, dlqKey, messageID, messageID).Result()
	if err != nil {
		return fmt.Errorf("failed to get DLQ message: %w", err)
	}

	if len(messages) == 0 {
		return fmt.Errorf("message not found in DLQ: %s", messageID)
	}

	message := messages[0]

	// Extract original message fields
	originalFields := make(map[string]interface{})
	for k, v := range message.Values {
		// Skip DLQ-specific fields
		if k == "original_message_id" || k == "error" || k == "timestamp" || k == "retry_count" {
			continue
		}
		originalFields[k] = v
	}

	// Publish back to original stream
	streamKey := dlq.getStreamKey(topic)
	_, err = dlq.client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		ID:     "*",
		Values: originalFields,
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to replay message: %w", err)
	}

	// Remove from DLQ
	err = dlq.client.XDel(ctx, dlqKey, messageID).Err()
	if err != nil {
		dlq.logger.Error("Failed to remove replayed message from DLQ",
			"topic", topic,
			"message_id", messageID,
			"error", err,
		)
		// Don't return error as the message was successfully replayed
	}

	dlq.logger.Info("Message replayed from DLQ",
		"topic", topic,
		"message_id", messageID,
	)

	return nil
}

// RemoveMessage removes a message from the dead letter queue
func (dlq *DeadLetterQueue) RemoveMessage(ctx context.Context, topic string, messageID string) error {
	dlqKey := dlq.getDLQKey(topic)

	err := dlq.client.XDel(ctx, dlqKey, messageID).Err()
	if err != nil {
		return fmt.Errorf("failed to remove message from DLQ: %w", err)
	}

	dlq.logger.Debug("Message removed from DLQ",
		"topic", topic,
		"message_id", messageID,
	)

	return nil
}

// Stats returns dead letter queue statistics
func (dlq *DeadLetterQueue) Stats(ctx context.Context) (*contract.DeadLetterQueueStats, error) {
	stats := &contract.DeadLetterQueueStats{
		TopicQueues: make(map[string]*contract.TopicDeadLetterQueueStats),
	}

	// Get all DLQ keys
	keys, err := dlq.client.Keys(ctx, dlq.getDLQKey("*")).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get DLQ keys: %w", err)
	}

	for _, key := range keys {
		topic := dlq.extractTopicFromDLQKey(key)

		// Get stream info
		info, err := dlq.client.XInfoStream(ctx, key).Result()
		if err != nil {
			dlq.logger.Error("Failed to get DLQ stream info",
				"topic", topic,
				"key", key,
				"error", err,
			)
			continue
		}

		var oldestTime, newestTime time.Time
		if info.Length > 0 {
			// Get oldest message
			oldest, err := dlq.client.XRange(ctx, key, "-", "+").Result()
			if err == nil && len(oldest) > 0 {
				if timestampStr, ok := oldest[0].Values["timestamp"].(string); ok {
					oldestTime, _ = time.Parse(time.RFC3339Nano, timestampStr)
				}
			}

			// Get newest message
			newest, err := dlq.client.XRevRange(ctx, key, "+", "-").Result()
			if err == nil && len(newest) > 0 {
				if timestampStr, ok := newest[0].Values["timestamp"].(string); ok {
					newestTime, _ = time.Parse(time.RFC3339Nano, timestampStr)
				}
			}
		}

		stats.TopicQueues[topic] = &contract.TopicDeadLetterQueueStats{
			Topic:             topic,
			MessagesCount:     info.Length,
			OldestMessageTime: oldestTime,
			NewestMessageTime: newestTime,
		}
	}

	return stats, nil
}

// Helper methods
func (dlq *DeadLetterQueue) getDLQKey(topic string) string {
	return fmt.Sprintf("event:dlq:%s", topic)
}

func (dlq *DeadLetterQueue) getStreamKey(topic string) string {
	return fmt.Sprintf("event:stream:%s", topic)
}

func (dlq *DeadLetterQueue) extractTopicFromDLQKey(key string) string {
	prefix := "event:dlq:"
	if len(key) > len(prefix) {
		return key[len(prefix):]
	}
	return key
}

// deserializeDLQMessage deserializes a DLQ message into an event
func (dlq *DeadLetterQueue) deserializeDLQMessage(message redis.XMessage) (contract.Event, error) {
	// Create a map with original event fields
	eventFields := make(map[string]interface{})
	for k, v := range message.Values {
		// Skip DLQ-specific fields
		if k == "original_message_id" || k == "error" || k == "timestamp" || k == "retry_count" {
			continue
		}
		eventFields[k] = v
	}

	// Deserialize the event
	event, err := dlq.serializer.Deserialize(eventFields)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize DLQ event: %w", err)
	}

	// Add DLQ-specific headers
	if e, ok := event.(*Event); ok {
		e.WithHeader("dlq_message_id", message.ID)
		if originalID, ok := message.Values["original_message_id"].(string); ok {
			e.WithHeader("original_message_id", originalID)
		}
		if errorStr, ok := message.Values["error"].(string); ok {
			e.WithHeader("error", errorStr)
		}
		if timestampStr, ok := message.Values["timestamp"].(string); ok {
			e.WithHeader("dlq_timestamp", timestampStr)
		}
	}

	return event, nil
}

// CleanupExpiredMessages removes expired messages from DLQ
func (dlq *DeadLetterQueue) CleanupExpiredMessages(ctx context.Context, topic string, maxAge time.Duration) error {
	dlqKey := dlq.getDLQKey(topic)

	// Calculate cutoff time
	cutoff := time.Now().Add(-maxAge)

	// Get all messages
	messages, err := dlq.client.XRange(ctx, dlqKey, "-", "+").Result()
	if err != nil {
		return fmt.Errorf("failed to get DLQ messages for cleanup: %w", err)
	}

	var toDelete []string
	for _, message := range messages {
		if timestampStr, ok := message.Values["timestamp"].(string); ok {
			if timestamp, err := time.Parse(time.RFC3339Nano, timestampStr); err == nil {
				if timestamp.Before(cutoff) {
					toDelete = append(toDelete, message.ID)
				}
			}
		}
	}

	// Delete expired messages
	if len(toDelete) > 0 {
		if err := dlq.client.XDel(ctx, dlqKey, toDelete...).Err(); err != nil {
			return fmt.Errorf("failed to delete expired DLQ messages: %w", err)
		}

		dlq.logger.Info("Cleaned up expired DLQ messages",
			"topic", topic,
			"count", len(toDelete),
		)
	}

	return nil
}
