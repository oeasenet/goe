# GOE Event System 🚀

A robust, scalable pub/sub eventing system for the GOE framework using Redis Streams as the backend message broker.

## Features

- **Redis Streams Backend**: High-throughput, persistent message storage
- **Consumer Groups**: Horizontal scaling with load balancing
- **Dead Letter Queue**: Automatic isolation of failed messages with replay capabilities
- **Delayed Messages**: Schedule messages for future delivery
- **Message Ordering**: Guaranteed order within a single stream/topic
- **At-Least-Once Delivery**: Ensures messages are processed at least once
- **Stale Consumer Recovery**: Automatic recovery from failed consumer instances
- **Observability**: Built-in metrics and tracing support

## Quick Start

### 1. Enable the Event Module

```go
app := goe.New(goe.Options{
    WithEvent: true, // Enable event system
})
```

### 2. Publish Events

```go
// Inject the event publisher
func CreateUser(publisher contract.EventPublisher) func(c fiber.Ctx) error {
    return func(c fiber.Ctx) error {
        event := event.NewEvent("user.created", userData)
        return publisher.Publish(c.Context(), "user-events", event)
    }
}
```

### 3. Subscribe to Events

```go
// Create event handler
handler := event.CreateEventHandler(func(ctx context.Context, event contract.Event) error {
    // Process the event
    return nil
})

// Subscribe to events
consumer.Subscribe(ctx, "user-events", "user-processor", handler)
```

## Configuration

Configure using environment variables:

```bash
# Redis connection
EVENT_REDIS_ADDR=localhost:6379
EVENT_REDIS_PASSWORD=
EVENT_REDIS_DB=0

# Consumer settings
EVENT_CONSUMER_TIMEOUT=30s
EVENT_MAX_RETRIES=3
EVENT_RETRY_BACKOFF=1s

# Performance settings
EVENT_BATCH_SIZE=10
EVENT_MAX_PENDING_MESSAGES=1000

# Delayed queue settings
EVENT_DELAYED_QUEUE_ENABLED=true
EVENT_DELAYED_QUEUE_CHECK_INTERVAL=1s
```

## Advanced Features

### Delayed Messages
```go
// Schedule a message for future delivery
publisher.PublishWithDelay(ctx, "notifications", event, 5*time.Minute)
```

### Batch Publishing
```go
batch := event.NewBatchPublisher(publisher, "analytics")
batch.AddJSON("user.action", data1)
batch.AddJSON("user.action", data2)
batch.Publish(ctx)
```

### Dead Letter Queue Management
```go
dlq := eventManager.GetDeadLetterQueue()
messages, _ := dlq.GetMessages(ctx, "topic", 10)
dlq.ReplayMessage(ctx, "topic", messageID)
```

## Documentation

- [Event System Guide](../../docs/guide/event-system.md)
- [Basic Example](../../docs/examples/event-basic.md)
- [Advanced Example](../../docs/examples/event-advanced.md)

## Testing

Run tests with Docker Redis:

```bash
# Start Redis
docker-compose -f tests/event/docker-compose.yml up -d

# Run tests
RUN_INTEGRATION_TESTS=true go test -v ./tests/event/
```

## Architecture

The event system is built with:

- **Redis Streams**: Message storage and distribution
- **Consumer Groups**: Parallel processing with load balancing
- **Dead Letter Queue**: Failed message recovery
- **Delayed Queue**: Scheduled message delivery
- **Stale Consumer Recovery**: Automatic failover

## Best Practices

1. **Event Design**: Use meaningful names with namespacing
2. **Consumer Groups**: Use descriptive names and scale horizontally
3. **Error Handling**: Implement proper retry logic and DLQ monitoring
4. **Performance**: Use batch publishing for high-volume scenarios
5. **Testing**: Test both success and failure scenarios

## Contributing

1. Fork the repository
2. Create your feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see LICENSE file for details