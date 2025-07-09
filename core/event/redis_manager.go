package event

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.oease.dev/goe/v2/contract"
)

// RedisManager implements EventManager using Redis Streams
type RedisManager struct {
	client     *redis.Client
	config     *Config
	serializer *Serializer
	logger     contract.Logger
	dlq        *DeadLetterQueue
	consumers  map[string]*Consumer
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewRedisManager creates a new Redis-based event manager
func NewRedisManager(config *Config, logger contract.Logger) (*RedisManager, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	ctx, cancel = context.WithCancel(context.Background())

	dlq := NewDeadLetterQueue(client, config, logger)

	manager := &RedisManager{
		client:     client,
		config:     config,
		serializer: NewSerializer(),
		logger:     logger,
		dlq:        dlq,
		consumers:  make(map[string]*Consumer),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start background tasks
	manager.startBackgroundTasks()

	return manager, nil
}

// Publish publishes an event to a topic
func (r *RedisManager) Publish(ctx context.Context, topic string, event contract.Event) error {
	// Set topic in event
	if e, ok := event.(*Event); ok {
		e.WithTopic(topic)
	}

	// Serialize event
	fields, err := r.serializer.Serialize(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Publish to Redis Stream
	streamKey := r.getStreamKey(topic)
	_, err = r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		ID:     "*",
		Values: fields,
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to publish event to stream %s: %w", streamKey, err)
	}

	r.logger.Debug("Event published",
		"topic", topic,
		"event_id", event.ID(),
		"event_name", event.Name(),
	)

	return nil
}

// PublishWithDelay publishes an event with a delay
func (r *RedisManager) PublishWithDelay(ctx context.Context, topic string, event contract.Event, delay time.Duration) error {
	if !r.config.DelayedQueueEnabled {
		return fmt.Errorf("delayed queue is not enabled")
	}

	// Set topic in event
	if e, ok := event.(*Event); ok {
		e.WithTopic(topic)
	}

	// Serialize event
	fields, err := r.serializer.Serialize(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Add delay timestamp
	deliveryTime := time.Now().Add(delay)
	fields["delivery_time"] = deliveryTime.Format(time.RFC3339Nano)

	// Serialize fields to JSON for storage
	fieldsJSON, err := json.Marshal(fields)
	if err != nil {
		return fmt.Errorf("failed to marshal delayed event fields: %w", err)
	}

	// Publish to delayed queue
	delayedKey := r.getDelayedQueueKey(topic)
	_, err = r.client.ZAdd(ctx, delayedKey, redis.Z{
		Score:  float64(deliveryTime.Unix()),
		Member: string(fieldsJSON),
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to publish delayed event: %w", err)
	}

	r.logger.Debug("Delayed event published",
		"topic", topic,
		"event_id", event.ID(),
		"event_name", event.Name(),
		"delay", delay,
		"delivery_time", deliveryTime,
	)

	return nil
}

// PublishBatch publishes multiple events in a batch
func (r *RedisManager) PublishBatch(ctx context.Context, topic string, events []contract.Event) error {
	if len(events) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()
	streamKey := r.getStreamKey(topic)

	for _, event := range events {
		// Set topic in event
		if e, ok := event.(*Event); ok {
			e.WithTopic(topic)
		}

		// Serialize event
		fields, err := r.serializer.Serialize(event)
		if err != nil {
			return fmt.Errorf("failed to serialize event: %w", err)
		}

		// Add to pipeline
		pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: streamKey,
			ID:     "*",
			Values: fields,
		})
	}

	// Execute batch
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish batch events: %w", err)
	}

	r.logger.Debug("Batch events published",
		"topic", topic,
		"count", len(events),
	)

	return nil
}

// Subscribe subscribes to a topic with a consumer group
func (r *RedisManager) Subscribe(ctx context.Context, topic string, consumerGroup string, handler contract.EventHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	consumerKey := r.getConsumerKey(topic, consumerGroup)
	if _, exists := r.consumers[consumerKey]; exists {
		return fmt.Errorf("consumer group %s already subscribed to topic %s", consumerGroup, topic)
	}

	// Create consumer group if it doesn't exist
	streamKey := r.getStreamKey(topic)
	err := r.client.XGroupCreateMkStream(ctx, streamKey, consumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	// Create consumer
	consumer := NewConsumer(r.client, r.config, r.logger, r.dlq, topic, consumerGroup, handler)
	r.consumers[consumerKey] = consumer

	// Start consumer
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		consumer.Start(r.ctx)
	}()

	r.logger.Info("Subscribed to topic",
		"topic", topic,
		"consumer_group", consumerGroup,
	)

	return nil
}

// Unsubscribe stops consuming from a topic
func (r *RedisManager) Unsubscribe(ctx context.Context, topic string, consumerGroup string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	consumerKey := r.getConsumerKey(topic, consumerGroup)
	consumer, exists := r.consumers[consumerKey]
	if !exists {
		return fmt.Errorf("consumer group %s not subscribed to topic %s", consumerGroup, topic)
	}

	// Stop consumer
	consumer.Stop()
	delete(r.consumers, consumerKey)

	r.logger.Info("Unsubscribed from topic",
		"topic", topic,
		"consumer_group", consumerGroup,
	)

	return nil
}

// Acknowledge acknowledges message processing
func (r *RedisManager) Acknowledge(ctx context.Context, topic string, consumerGroup string, messageID string) error {
	streamKey := r.getStreamKey(topic)
	return r.client.XAck(ctx, streamKey, consumerGroup, messageID).Err()
}

// Reject rejects a message and optionally requeues it
func (r *RedisManager) Reject(ctx context.Context, topic string, consumerGroup string, messageID string, requeue bool) error {
	if requeue {
		// Move message back to pending
		streamKey := r.getStreamKey(topic)
		return r.client.XClaim(ctx, &redis.XClaimArgs{
			Stream:   streamKey,
			Group:    consumerGroup,
			Consumer: "requeue",
			MinIdle:  0,
			Messages: []string{messageID},
		}).Err()
	}

	// Just acknowledge to remove from pending
	return r.Acknowledge(ctx, topic, consumerGroup, messageID)
}

// Health checks the health of the event system
func (r *RedisManager) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Stats returns event system statistics
func (r *RedisManager) Stats(ctx context.Context) (*contract.EventStats, error) {
	stats := &contract.EventStats{
		Topics:         make(map[string]*contract.TopicStats),
		ConsumerGroups: make(map[string]*contract.ConsumerGroupStats),
	}

	// Get topic stats
	keys, err := r.client.Keys(ctx, r.getStreamKey("*")).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get stream keys: %w", err)
	}

	for _, key := range keys {
		topic := r.extractTopicFromStreamKey(key)
		info, err := r.client.XInfoStream(ctx, key).Result()
		if err != nil {
			continue
		}

		stats.Topics[topic] = &contract.TopicStats{
			Name:          topic,
			MessagesCount: info.Length,
			LastMessageID: info.LastGeneratedID,
		}

		// Get consumer group stats
		groups, err := r.client.XInfoGroups(ctx, key).Result()
		if err != nil {
			continue
		}

		for _, group := range groups {
			groupKey := fmt.Sprintf("%s:%s", topic, group.Name)
			stats.ConsumerGroups[groupKey] = &contract.ConsumerGroupStats{
				Name:                 group.Name,
				Topic:                topic,
				PendingMessagesCount: group.Pending,
				LastDeliveredID:      group.LastDeliveredID,
			}
		}
	}

	// Get DLQ stats
	dlqStats, err := r.dlq.Stats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DLQ stats: %w", err)
	}
	stats.DeadLetterQueue = dlqStats

	return stats, nil
}

// GetDeadLetterQueue returns the dead letter queue manager
func (r *RedisManager) GetDeadLetterQueue() contract.DeadLetterQueueManager {
	return r.dlq
}

// Close closes the event manager and all connections
func (r *RedisManager) Close(ctx context.Context) error {
	r.cancel()
	r.wg.Wait()

	r.mu.Lock()
	defer r.mu.Unlock()

	// Stop all consumers
	for _, consumer := range r.consumers {
		consumer.Stop()
	}

	// Close Redis connection
	return r.client.Close()
}

// Helper methods
func (r *RedisManager) getStreamKey(topic string) string {
	return fmt.Sprintf("event:stream:%s", topic)
}

func (r *RedisManager) getDelayedQueueKey(topic string) string {
	return fmt.Sprintf("event:delayed:%s", topic)
}

func (r *RedisManager) getConsumerKey(topic, consumerGroup string) string {
	return fmt.Sprintf("%s:%s", topic, consumerGroup)
}

func (r *RedisManager) extractTopicFromStreamKey(key string) string {
	prefix := "event:stream:"
	if len(key) > len(prefix) {
		return key[len(prefix):]
	}
	return key
}

// startBackgroundTasks starts background maintenance tasks
func (r *RedisManager) startBackgroundTasks() {
	if r.config.DelayedQueueEnabled {
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			r.processDelayedMessages()
		}()
	}

	// Start stale consumer cleanup
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		r.cleanupStaleConsumers()
	}()
}

// processDelayedMessages processes delayed messages
func (r *RedisManager) processDelayedMessages() {
	ticker := time.NewTicker(r.config.DelayedQueueCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			r.processReadyDelayedMessages()
		}
	}
}

// processReadyDelayedMessages processes messages that are ready to be delivered
func (r *RedisManager) processReadyDelayedMessages() {
	ctx := context.Background()
	now := time.Now().Unix()

	// Get all delayed queue keys
	keys, err := r.client.Keys(ctx, r.getDelayedQueueKey("*")).Result()
	if err != nil {
		r.logger.Error("Failed to get delayed queue keys", "error", err)
		return
	}

	for _, key := range keys {
		// Get ready messages
		messages, err := r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
			Min: "0",
			Max: fmt.Sprintf("%d", now),
		}).Result()
		if err != nil {
			r.logger.Error("Failed to get ready delayed messages", "key", key, "error", err)
			continue
		}

		for _, message := range messages {
			// Parse message (stored as string in Redis sorted set)
			var fields map[string]interface{}
			if err := json.Unmarshal([]byte(message), &fields); err != nil {
				r.logger.Error("Failed to parse delayed message", "message", message, "error", err)
				continue
			}

			// Remove delivery_time field
			delete(fields, "delivery_time")

			// Extract topic
			topic, ok := fields["topic"].(string)
			if !ok {
				r.logger.Error("Invalid topic in delayed message", "message", message)
				continue
			}

			// Publish to stream
			streamKey := r.getStreamKey(topic)
			_, err = r.client.XAdd(ctx, &redis.XAddArgs{
				Stream: streamKey,
				ID:     "*",
				Values: fields,
			}).Result()

			if err != nil {
				r.logger.Error("Failed to publish delayed message", "topic", topic, "error", err)
				continue
			}

			// Remove from delayed queue
			r.client.ZRem(ctx, key, message)

			r.logger.Debug("Delayed message published", "topic", topic)
		}
	}
}

// cleanupStaleConsumers cleans up stale consumers and reclaims their messages
func (r *RedisManager) cleanupStaleConsumers() {
	ticker := time.NewTicker(r.config.ClaimInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			r.reclaimStaleMessages()
		}
	}
}

// reclaimStaleMessages reclaims messages from stale consumers
func (r *RedisManager) reclaimStaleMessages() {
	ctx := context.Background()

	// Get all stream keys
	keys, err := r.client.Keys(ctx, r.getStreamKey("*")).Result()
	if err != nil {
		r.logger.Error("Failed to get stream keys for cleanup", "error", err)
		return
	}

	for _, streamKey := range keys {
		// Get consumer groups
		groups, err := r.client.XInfoGroups(ctx, streamKey).Result()
		if err != nil {
			continue
		}

		for _, group := range groups {
			// Get pending messages
			pending, err := r.client.XPending(ctx, streamKey, group.Name).Result()
			if err != nil {
				continue
			}

			if pending.Count > 0 {
				// Get detailed pending info
				pendingExt, err := r.client.XPendingExt(ctx, &redis.XPendingExtArgs{
					Stream: streamKey,
					Group:  group.Name,
					Start:  "-",
					End:    "+",
					Count:  int64(r.config.BatchSize),
				}).Result()
				if err != nil {
					continue
				}

				for _, msg := range pendingExt {
					// Check if message is stale
					if msg.Idle > r.config.StaleConsumerTimeout {
						// Claim the message
						_, err := r.client.XClaim(ctx, &redis.XClaimArgs{
							Stream:   streamKey,
							Group:    group.Name,
							Consumer: "reclaim",
							MinIdle:  r.config.ClaimMinIdleTime,
							Messages: []string{msg.ID},
						}).Result()
						if err != nil {
							r.logger.Error("Failed to reclaim stale message",
								"stream", streamKey,
								"group", group.Name,
								"message_id", msg.ID,
								"error", err,
							)
						} else {
							r.logger.Debug("Reclaimed stale message",
								"stream", streamKey,
								"group", group.Name,
								"message_id", msg.ID,
							)
						}
					}
				}
			}
		}
	}
}
