package event

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.oease.dev/goe/v2/contract"
)

// Consumer represents a Redis Streams consumer
type Consumer struct {
	client        *redis.Client
	config        *Config
	logger        contract.Logger
	dlq           *DeadLetterQueue
	topic         string
	consumerGroup string
	consumerID    string
	handler       contract.EventHandler
	serializer    *Serializer
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	stopped       bool
	mu            sync.RWMutex
}

// NewConsumer creates a new consumer
func NewConsumer(
	client *redis.Client,
	config *Config,
	logger contract.Logger,
	dlq *DeadLetterQueue,
	topic string,
	consumerGroup string,
	handler contract.EventHandler,
) *Consumer {
	consumerID := fmt.Sprintf("%s-%d", consumerGroup, time.Now().UnixNano())
	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		client:        client,
		config:        config,
		logger:        logger,
		dlq:           dlq,
		topic:         topic,
		consumerGroup: consumerGroup,
		consumerID:    consumerID,
		handler:       handler,
		serializer:    NewSerializer(),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start starts the consumer
func (c *Consumer) Start(ctx context.Context) {
	c.logger.Info("Starting consumer",
		"topic", c.topic,
		"consumer_group", c.consumerGroup,
		"consumer_id", c.consumerID,
	)

	// Start main consumption loop
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.consumeLoop(ctx)
	}()

	// Start pending message processing
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.processPendingMessages(ctx)
	}()
}

// Stop stops the consumer
func (c *Consumer) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stopped {
		return
	}

	c.stopped = true
	c.cancel()
	c.wg.Wait()

	c.logger.Info("Consumer stopped",
		"topic", c.topic,
		"consumer_group", c.consumerGroup,
		"consumer_id", c.consumerID,
	)
}

// consumeLoop is the main consumption loop
func (c *Consumer) consumeLoop(ctx context.Context) {
	streamKey := c.getStreamKey()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.ctx.Done():
			return
		default:
			// Read new messages from stream
			messages, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    c.consumerGroup,
				Consumer: c.consumerID,
				Streams:  []string{streamKey, ">"},
				Count:    int64(c.config.BatchSize),
				Block:    c.config.ConsumerTimeout,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					// No new messages, continue
					continue
				}
				c.logger.Error("Failed to read from stream",
					"topic", c.topic,
					"consumer_group", c.consumerGroup,
					"error", err,
				)
				time.Sleep(c.config.RetryBackoff)
				continue
			}

			// Process messages
			for _, stream := range messages {
				for _, message := range stream.Messages {
					c.processMessage(ctx, message)
				}
			}
		}
	}
}

// processPendingMessages processes pending messages
func (c *Consumer) processPendingMessages(ctx context.Context) {
	ticker := time.NewTicker(c.config.ClaimInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.claimPendingMessages(ctx)
		}
	}
}

// claimPendingMessages claims pending messages from other consumers
func (c *Consumer) claimPendingMessages(ctx context.Context) {
	streamKey := c.getStreamKey()

	// Get pending messages
	pending, err := c.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: streamKey,
		Group:  c.consumerGroup,
		Start:  "-",
		End:    "+",
		Count:  int64(c.config.BatchSize),
	}).Result()

	if err != nil {
		c.logger.Error("Failed to get pending messages",
			"topic", c.topic,
			"consumer_group", c.consumerGroup,
			"error", err,
		)
		return
	}

	for _, msg := range pending {
		// Skip if message is not idle long enough
		if msg.Idle < c.config.ClaimMinIdleTime {
			continue
		}

		// Skip if it's our own message
		if msg.Consumer == c.consumerID {
			continue
		}

		// Claim the message
		claimed, err := c.client.XClaim(ctx, &redis.XClaimArgs{
			Stream:   streamKey,
			Group:    c.consumerGroup,
			Consumer: c.consumerID,
			MinIdle:  c.config.ClaimMinIdleTime,
			Messages: []string{msg.ID},
		}).Result()

		if err != nil {
			c.logger.Error("Failed to claim message",
				"topic", c.topic,
				"consumer_group", c.consumerGroup,
				"message_id", msg.ID,
				"error", err,
			)
			continue
		}

		// Process claimed messages
		for _, claimedMsg := range claimed {
			c.logger.Debug("Claimed message",
				"topic", c.topic,
				"consumer_group", c.consumerGroup,
				"message_id", claimedMsg.ID,
				"original_consumer", msg.Consumer,
			)
			c.processMessage(ctx, claimedMsg)
		}
	}
}

// processMessage processes a single message
func (c *Consumer) processMessage(ctx context.Context, message redis.XMessage) {
	streamKey := c.getStreamKey()
	messageID := message.ID

	// Deserialize event
	event, err := c.serializer.Deserialize(message.Values)
	if err != nil {
		c.logger.Error("Failed to deserialize event",
			"topic", c.topic,
			"message_id", messageID,
			"error", err,
		)
		c.handleMessageError(ctx, messageID, err)
		return
	}

	// Process event with retries
	var processingErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return
		case <-c.ctx.Done():
			return
		default:
		}

		// Create context with timeout
		processCtx, cancel := context.WithTimeout(ctx, c.config.ConsumerTimeout)
		processingErr = c.handler.Handle(processCtx, event)
		cancel()

		if processingErr == nil {
			// Success - acknowledge message
			if err := c.client.XAck(ctx, streamKey, c.consumerGroup, messageID).Err(); err != nil {
				c.logger.Error("Failed to acknowledge message",
					"topic", c.topic,
					"message_id", messageID,
					"error", err,
				)
			} else {
				c.logger.Debug("Message processed successfully",
					"topic", c.topic,
					"message_id", messageID,
					"event_name", event.Name(),
				)
			}
			return
		}

		// Log attempt failure
		c.logger.Warn("Message processing failed",
			"topic", c.topic,
			"message_id", messageID,
			"attempt", attempt+1,
			"max_retries", c.config.MaxRetries,
			"error", processingErr,
		)

		// Wait before retry (except for last attempt)
		if attempt < c.config.MaxRetries {
			time.Sleep(c.config.RetryBackoff * time.Duration(attempt+1)) // Exponential backoff
		}
	}

	// All retries failed - move to dead letter queue
	c.handleMessageError(ctx, messageID, processingErr)
}

// handleMessageError handles message processing errors
func (c *Consumer) handleMessageError(ctx context.Context, messageID string, err error) {
	streamKey := c.getStreamKey()

	// Get message details
	messages, getErr := c.client.XRange(ctx, streamKey, messageID, messageID).Result()
	if getErr != nil {
		c.logger.Error("Failed to get message for DLQ",
			"topic", c.topic,
			"message_id", messageID,
			"error", getErr,
		)
		// Still acknowledge to prevent reprocessing
		c.client.XAck(ctx, streamKey, c.consumerGroup, messageID)
		return
	}

	if len(messages) == 0 {
		c.logger.Error("Message not found for DLQ",
			"topic", c.topic,
			"message_id", messageID,
		)
		// Still acknowledge to prevent reprocessing
		c.client.XAck(ctx, streamKey, c.consumerGroup, messageID)
		return
	}

	// Move to dead letter queue
	if dlqErr := c.dlq.AddMessage(ctx, c.topic, messageID, messages[0].Values, err); dlqErr != nil {
		c.logger.Error("Failed to add message to DLQ",
			"topic", c.topic,
			"message_id", messageID,
			"error", dlqErr,
		)
	} else {
		c.logger.Info("Message moved to dead letter queue",
			"topic", c.topic,
			"message_id", messageID,
			"error", err,
		)
	}

	// Acknowledge message to remove from pending
	if ackErr := c.client.XAck(ctx, streamKey, c.consumerGroup, messageID).Err(); ackErr != nil {
		c.logger.Error("Failed to acknowledge failed message",
			"topic", c.topic,
			"message_id", messageID,
			"error", ackErr,
		)
	}
}

// getStreamKey returns the Redis stream key for the topic
func (c *Consumer) getStreamKey() string {
	return fmt.Sprintf("event:stream:%s", c.topic)
}
