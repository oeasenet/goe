package contract

import (
	"context"
	"time"
)

// Event represents an event in the system
type Event interface {
	// ID returns the unique event ID
	ID() string

	// Topic returns the event topic/stream name
	Topic() string

	// Name returns the event name
	Name() string

	// Payload returns the event payload
	Payload() any

	// Headers returns the event headers
	Headers() map[string]string

	// Timestamp returns when the event was created
	Timestamp() time.Time

	// Context returns the event context
	Context() context.Context
}

// EventPublisher defines the event publisher interface
type EventPublisher interface {
	// Publish publishes an event to a topic
	Publish(ctx context.Context, topic string, event Event) error

	// PublishWithDelay publishes an event with a delay
	PublishWithDelay(ctx context.Context, topic string, event Event, delay time.Duration) error

	// PublishBatch publishes multiple events in a batch
	PublishBatch(ctx context.Context, topic string, events []Event) error
}

// EventConsumer defines the event consumer interface
type EventConsumer interface {
	// Subscribe subscribes to a topic with a consumer group
	Subscribe(ctx context.Context, topic string, consumerGroup string, handler EventHandler) error

	// Unsubscribe stops consuming from a topic
	Unsubscribe(ctx context.Context, topic string, consumerGroup string) error

	// Acknowledge acknowledges message processing
	Acknowledge(ctx context.Context, topic string, consumerGroup string, messageID string) error

	// Reject rejects a message and optionally requeues it
	Reject(ctx context.Context, topic string, consumerGroup string, messageID string, requeue bool) error
}

// EventHandler handles consumed events
type EventHandler interface {
	// Handle handles an event and returns an error if processing fails
	Handle(ctx context.Context, event Event) error
}

// EventHandlerFunc is a function adapter for EventHandler
type EventHandlerFunc func(ctx context.Context, event Event) error

// Handle implements EventHandler interface
func (f EventHandlerFunc) Handle(ctx context.Context, event Event) error {
	return f(ctx, event)
}

// EventManager combines publisher and consumer functionality
type EventManager interface {
	EventPublisher
	EventConsumer

	// Health checks the health of the event system
	Health(ctx context.Context) error

	// Stats returns event system statistics
	Stats(ctx context.Context) (*EventStats, error)

	// GetDeadLetterQueue returns the dead letter queue manager
	GetDeadLetterQueue() DeadLetterQueueManager

	// Close closes the event manager and all connections
	Close(ctx context.Context) error
}

// EventStats contains event system statistics
type EventStats struct {
	// Topics contains per-topic statistics
	Topics map[string]*TopicStats

	// ConsumerGroups contains per-consumer-group statistics
	ConsumerGroups map[string]*ConsumerGroupStats

	// DeadLetterQueue contains dead letter queue statistics
	DeadLetterQueue *DeadLetterQueueStats
}

// TopicStats contains statistics for a specific topic
type TopicStats struct {
	// Name is the topic name
	Name string

	// MessagesCount is the total number of messages
	MessagesCount int64

	// ConsumersCount is the number of active consumers
	ConsumersCount int

	// LastMessageID is the ID of the last message
	LastMessageID string

	// LastMessageTime is the timestamp of the last message
	LastMessageTime time.Time
}

// ConsumerGroupStats contains statistics for a consumer group
type ConsumerGroupStats struct {
	// Name is the consumer group name
	Name string

	// Topic is the topic being consumed
	Topic string

	// ConsumersCount is the number of consumers in the group
	ConsumersCount int

	// PendingMessagesCount is the number of pending messages
	PendingMessagesCount int64

	// LastDeliveredID is the ID of the last delivered message
	LastDeliveredID string

	// LastDeliveredTime is the timestamp of the last delivery
	LastDeliveredTime time.Time
}

// DeadLetterQueueManager manages dead letter queues
type DeadLetterQueueManager interface {
	// GetMessages retrieves messages from the dead letter queue
	GetMessages(ctx context.Context, topic string, limit int) ([]Event, error)

	// ReplayMessage replays a message from the dead letter queue
	ReplayMessage(ctx context.Context, topic string, messageID string) error

	// RemoveMessage removes a message from the dead letter queue
	RemoveMessage(ctx context.Context, topic string, messageID string) error

	// Stats returns dead letter queue statistics
	Stats(ctx context.Context) (*DeadLetterQueueStats, error)
}

// DeadLetterQueueStats contains dead letter queue statistics
type DeadLetterQueueStats struct {
	// TopicQueues contains per-topic dead letter queue statistics
	TopicQueues map[string]*TopicDeadLetterQueueStats
}

// TopicDeadLetterQueueStats contains statistics for a topic's dead letter queue
type TopicDeadLetterQueueStats struct {
	// Topic is the topic name
	Topic string

	// MessagesCount is the number of messages in the dead letter queue
	MessagesCount int64

	// OldestMessageTime is the timestamp of the oldest message
	OldestMessageTime time.Time

	// NewestMessageTime is the timestamp of the newest message
	NewestMessageTime time.Time
}

// EventListener handles events (deprecated, use EventHandler instead)
type EventListener interface {
	// Handle handles an event
	Handle(event Event) error
}

// EventSubscriber subscribes to events (deprecated, use EventHandler instead)
type EventSubscriber interface {
	// Subscribe returns event patterns to subscribe to
	Subscribe() []string

	// Handle handles an event
	Handle(event Event) error
}

// EventDispatcher defines the event dispatcher interface (deprecated, use EventManager instead)
type EventDispatcher interface {
	// Dispatch dispatches an event
	Dispatch(event Event) error

	// DispatchAsync dispatches an event asynchronously
	DispatchAsync(event Event) error

	// Listen registers an event listener
	Listen(eventName string, listener EventListener)

	// Subscribe subscribes to events matching a pattern
	Subscribe(pattern string, subscriber EventSubscriber)
}
