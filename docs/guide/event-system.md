# Event System

The GOE event system provides a robust, Redis Streams-based publish-subscribe (pub/sub) eventing system for building scalable, distributed applications. It supports advanced features like consumer groups, dead letter queues, delayed messages, and guaranteed message ordering.

## Features

- **Redis Streams Backend**: Uses Redis Streams for high-throughput, persistent message storage
- **Consumer Groups**: Horizontal scaling with load balancing across multiple consumers
- **Dead Letter Queue**: Automatic isolation of failed messages with replay capabilities
- **Delayed Messages**: Schedule messages for future delivery
- **Message Ordering**: Guaranteed order within a single stream/topic
- **At-Least-Once Delivery**: Ensures messages are processed at least once
- **Stale Consumer Recovery**: Automatic recovery from failed consumer instances
- **Observability**: Built-in metrics and tracing support

## Quick Start

### 1. Enable the Event Module

```go
package main

import (
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
)

func main() {
    app := goe.New(goe.Options{
        WithEvent: true, // Enable event system
    })
    
    app.Run()
}
```

### 2. Global Accessors

The event system is available via global accessors anywhere in your application:

```go
// Access event services globally
eventManager := goe.EventManager()     // Full event manager
publisher := goe.EventPublisher()      // Just publishing
consumer := goe.EventConsumer()        // Just consuming
dlq := goe.DeadLetterQueue()          // Dead letter queue

// Use in HTTP handlers, services, etc.
func MyHandler(c fiber.Ctx) error {
    event := event.NewEvent("user.action", userData)
    return goe.EventPublisher().Publish(c.Context(), "user-events", event)
}
```

### 3. Basic Event Publishing

```go
// Inject the event publisher
func HandleUserSignup(publisher contract.EventPublisher) func(c fiber.Ctx) error {
    return func(c fiber.Ctx) error {
        // Create and publish an event
        event := event.NewEvent("user.signup", map[string]interface{}{
            "userId": 123,
            "email":  "user@example.com",
            "plan":   "premium",
        })
        
        err := publisher.Publish(c.Context(), "user-events", event)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to publish event"})
        }
        
        return c.JSON(fiber.Map{"message": "User created successfully"})
    }
}
```

### 4. Basic Event Consumption

```go
// Event handler
func HandleUserSignup(ctx context.Context, event contract.Event) error {
    // Extract payload
    var payload map[string]interface{}
    if err := event.UnmarshalEventPayload(event, &payload); err != nil {
        return err
    }
    
    userID := payload["userId"]
    email := payload["email"]
    
    // Process the event
    fmt.Printf("Processing user signup: %v, %v\n", userID, email)
    
    // Send welcome email, update analytics, etc.
    return nil
}

// Subscribe to events
func SetupEventSubscriptions(consumer contract.EventConsumer) {
    handler := event.CreateEventHandler(HandleUserSignup)
    
    // Subscribe to user events with consumer group
    consumer.Subscribe(
        context.Background(),
        "user-events",           // topic
        "user-signup-processor", // consumer group
        handler,
    )
}
```

## Configuration

The event system can be configured using environment variables:

### Redis Connection
```bash
# Redis connection settings
EVENT_REDIS_ADDR=localhost:6379
EVENT_REDIS_PASSWORD=
EVENT_REDIS_DB=0
```

### Consumer Settings
```bash
# Consumer timeout for processing messages
EVENT_CONSUMER_TIMEOUT=30s

# Maximum number of retries before moving to DLQ
EVENT_MAX_RETRIES=3

# Backoff time between retries
EVENT_RETRY_BACKOFF=1s

# Dead letter queue TTL
EVENT_DLQ_TTL=24h

# Timeout for detecting stale consumers
EVENT_STALE_CONSUMER_TIMEOUT=5m
```

### Performance Settings
```bash
# Number of messages to process in a batch
EVENT_BATCH_SIZE=10

# Maximum pending messages per consumer
EVENT_MAX_PENDING_MESSAGES=1000

# Minimum idle time before claiming stale messages
EVENT_CLAIM_MIN_IDLE_TIME=1m

# Interval for claiming stale messages
EVENT_CLAIM_INTERVAL=30s
```

### Delayed Queue Settings
```bash
# Enable delayed message processing
EVENT_DELAYED_QUEUE_ENABLED=true

# Interval for checking delayed messages
EVENT_DELAYED_QUEUE_CHECK_INTERVAL=1s
```

## Advanced Usage

### Delayed Messages

```go
// Publish a message with a 5-minute delay
err := publisher.PublishWithDelay(
    ctx,
    "notifications",
    event.NewEvent("reminder.send", reminderData),
    5*time.Minute,
)
```

### Batch Publishing

```go
// Create a batch publisher
batch := event.NewBatchPublisher(publisher, "analytics")

// Add events to batch
batch.AddJSON("user.action", actionData1)
batch.AddJSON("user.action", actionData2)
batch.AddJSON("user.action", actionData3)

// Publish all events in a single operation
err := batch.Publish(ctx)
```

### Typed Event Handlers

```go
type UserSignupEvent struct {
    UserID int    `json:"userId"`
    Email  string `json:"email"`
    Plan   string `json:"plan"`
}

// Create a typed handler
handler := event.CreateTypedEventHandler(
    func(ctx context.Context, event contract.Event, payload UserSignupEvent) error {
        // Payload is automatically unmarshaled
        fmt.Printf("User %d signed up with email %s\n", payload.UserID, payload.Email)
        return nil
    },
)
```

### Dead Letter Queue Management

```go
// Get the dead letter queue manager
dlq := eventManager.GetDeadLetterQueue()

// Retrieve failed messages
messages, err := dlq.GetMessages(ctx, "user-events", 10)
if err != nil {
    log.Printf("Failed to get DLQ messages: %v", err)
    return
}

// Replay a message
for _, msg := range messages {
    messageID := msg.Headers()["dlq_message_id"]
    err := dlq.ReplayMessage(ctx, "user-events", messageID)
    if err != nil {
        log.Printf("Failed to replay message %s: %v", messageID, err)
    }
}
```

### Pattern Matching

```go
// Create a pattern for filtering events
pattern := &event.EventPattern{
    Topic:     "user-events",
    EventName: "user.signup",
    Headers: map[string]string{
        "source": "web-app",
    },
}

// Wrap handler with pattern matching
patternHandler := event.NewPatternHandler(pattern, handler)
```

### Multiple Consumer Groups (Fan-out)

```go
// Multiple consumer groups can process the same messages
// Each group gets a copy of every message

// Analytics consumer group
consumer.Subscribe(ctx, "user-events", "analytics", analyticsHandler)

// Notification consumer group  
consumer.Subscribe(ctx, "user-events", "notifications", notificationHandler)

// Audit consumer group
consumer.Subscribe(ctx, "user-events", "audit", auditHandler)
```

## Error Handling

### Retry Logic

The event system automatically retries failed messages up to `EVENT_MAX_RETRIES` times with exponential backoff. After all retries are exhausted, messages are moved to the dead letter queue.

### Dead Letter Queue

Failed messages are automatically moved to a dead letter queue where they can be:
- Inspected for debugging
- Replayed after fixing the underlying issue
- Removed if they're no longer needed

### Stale Consumer Recovery

The system automatically detects and recovers from stale consumers:
- Messages from crashed consumers are reclaimed
- Automatic failover to healthy consumers
- Configurable timeout for stale detection

## Monitoring

### Health Checks

```go
// Check event system health
err := eventManager.Health(ctx)
if err != nil {
    log.Printf("Event system unhealthy: %v", err)
}
```

### Statistics

```go
// Get event system statistics
stats, err := eventManager.Stats(ctx)
if err != nil {
    log.Printf("Failed to get stats: %v", err)
    return
}

// Print topic statistics
for topic, topicStats := range stats.Topics {
    fmt.Printf("Topic %s: %d messages\n", topic, topicStats.MessagesCount)
}

// Print consumer group statistics
for group, groupStats := range stats.ConsumerGroups {
    fmt.Printf("Group %s: %d pending messages\n", group, groupStats.PendingMessagesCount)
}
```

### Observability Integration

When the observability module is enabled, the event system automatically provides:
- **Metrics**: Message throughput, processing latency, error rates
- **Tracing**: Distributed tracing across event publishing and consumption
- **Logging**: Structured logging with correlation IDs

## Best Practices

### 1. Event Design

- Use meaningful event names with namespacing (e.g., `user.signup`, `order.created`)
- Include all necessary data in the event payload
- Add metadata in headers for filtering and routing
- Keep events immutable and backward compatible

### 2. Consumer Groups

- Use descriptive consumer group names
- Group related consumers together
- Scale consumer groups horizontally for high throughput
- Monitor consumer group lag

### 3. Error Handling

- Implement proper error handling in event handlers
- Use dead letter queues for poison messages
- Set appropriate retry counts and backoff strategies
- Monitor and alert on DLQ growth

### 4. Performance

- Use batch publishing for high-volume scenarios
- Configure appropriate batch sizes for your workload
- Monitor Redis memory usage and performance
- Consider partitioning topics for very high volumes

### 5. Testing

- Use the provided test utilities for integration testing
- Test both happy path and error scenarios
- Mock event publishers in unit tests
- Test consumer group failover scenarios

## Common Patterns

### Event Sourcing
```go
// Store events as the source of truth
events := []contract.Event{
    event.NewEvent("user.created", userData),
    event.NewEvent("user.email.verified", emailData),
    event.NewEvent("user.plan.upgraded", planData),
}

for _, evt := range events {
    publisher.Publish(ctx, "user-events", evt)
}
```

### Saga Pattern
```go
// Coordinate distributed transactions
func HandleOrderPlaced(ctx context.Context, event contract.Event) error {
    // Step 1: Reserve inventory
    publisher.Publish(ctx, "inventory", 
        event.NewEvent("inventory.reserve", inventoryData))
    
    // Step 2: Process payment
    publisher.Publish(ctx, "payment", 
        event.NewEvent("payment.charge", paymentData))
    
    return nil
}
```

### CQRS (Command Query Responsibility Segregation)
```go
// Separate read and write models
func HandleUserCommand(ctx context.Context, event contract.Event) error {
    // Update write model
    writeModel.UpdateUser(userData)
    
    // Publish event for read model updates
    publisher.Publish(ctx, "user-projections", 
        event.NewEvent("user.updated", userData))
    
    return nil
}
```

## Troubleshooting

### Common Issues

1. **Messages not being consumed**
   - Check Redis connection
   - Verify consumer group is created
   - Check consumer is running and healthy

2. **High DLQ growth**
   - Check handler error logs
   - Verify message format compatibility
   - Review retry configuration

3. **Performance issues**
   - Monitor Redis memory and CPU
   - Check batch size configuration
   - Review consumer group scaling

4. **Message ordering issues**
   - Verify single topic usage
   - Check consumer group configuration
   - Review message publishing order

### Debug Commands

```bash
# Check Redis streams
redis-cli XINFO STREAM event:stream:user-events

# Check consumer groups
redis-cli XINFO GROUPS event:stream:user-events

# Check pending messages
redis-cli XPENDING event:stream:user-events user-group

# Check DLQ
redis-cli XLEN event:dlq:user-events
```