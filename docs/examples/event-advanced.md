# Advanced Event System Example

This example demonstrates advanced features of the GOE event system including delayed messages, dead letter queue handling, batch processing, and monitoring.

## Advanced Features Demonstrated

- Delayed message processing
- Dead letter queue management
- Batch event publishing
- Event replay and recovery
- Advanced monitoring and metrics
- Circuit breaker pattern
- Event versioning and migration

## Complete Example

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "sync"
    "time"

    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
    "go.oease.dev/goe/v2/core/event"
    "github.com/gofiber/fiber/v3"
)

// OrderEvent represents an order event
type OrderEvent struct {
    OrderID     string    `json:"orderId"`
    CustomerID  string    `json:"customerId"`
    Amount      float64   `json:"amount"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"createdAt"`
    Version     int       `json:"version"`
}

// PaymentEvent represents a payment event
type PaymentEvent struct {
    PaymentID   string    `json:"paymentId"`
    OrderID     string    `json:"orderId"`
    Amount      float64   `json:"amount"`
    Status      string    `json:"status"`
    ProcessedAt time.Time `json:"processedAt"`
}

// NotificationEvent represents a notification event
type NotificationEvent struct {
    UserID      string    `json:"userId"`
    Type        string    `json:"type"`
    Message     string    `json:"message"`
    ScheduledAt time.Time `json:"scheduledAt"`
}

// OrderService handles order processing
type OrderService struct {
    publisher contract.EventPublisher
    mu        sync.RWMutex
    orders    map[string]*OrderEvent
}

func NewOrderService(publisher contract.EventPublisher) *OrderService {
    return &OrderService{
        publisher: publisher,
        orders:    make(map[string]*OrderEvent),
    }
}

// CreateOrder creates a new order and publishes events
func (os *OrderService) CreateOrder(c fiber.Ctx) error {
    var order OrderEvent
    if err := c.Bind().Body(&order); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
    }

    order.OrderID = fmt.Sprintf("order-%d", time.Now().UnixNano())
    order.Status = "pending"
    order.CreatedAt = time.Now()
    order.Version = 1

    // Store order
    os.mu.Lock()
    os.orders[order.OrderID] = &order
    os.mu.Unlock()

    // Create batch publisher for multiple events
    batch := event.NewBatchPublisher(os.publisher, "order-events")

    // Add order created event
    orderCreatedEvent := event.NewEvent("order.created", order)
    orderCreatedEvent.WithHeaders(map[string]string{
        "source":      "order-service",
        "order_id":    order.OrderID,
        "customer_id": order.CustomerID,
        "version":     "1",
    })
    batch.Add(orderCreatedEvent)

    // Add inventory check event
    inventoryEvent := event.NewEvent("inventory.check", map[string]interface{}{
        "orderId":    order.OrderID,
        "customerId": order.CustomerID,
        "amount":     order.Amount,
    })
    batch.Add(inventoryEvent)

    // Publish batch
    if err := batch.Publish(c.Context()); err != nil {
        log.Printf("Failed to publish order events: %v", err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to create order"})
    }

    // Schedule reminder notification for 1 hour later
    reminderEvent := event.NewEvent("notification.reminder", NotificationEvent{
        UserID:      order.CustomerID,
        Type:        "order_reminder",
        Message:     fmt.Sprintf("Don't forget about your order %s", order.OrderID),
        ScheduledAt: time.Now().Add(time.Hour),
    })
    reminderEvent.WithHeaders(map[string]string{
        "order_id": order.OrderID,
        "type":     "reminder",
    })

    if err := os.publisher.PublishWithDelay(c.Context(), "notification-events", reminderEvent, time.Hour); err != nil {
        log.Printf("Failed to schedule reminder: %v", err)
        // Don't fail the request for reminder scheduling failure
    }

    return c.JSON(fiber.Map{
        "message": "Order created successfully",
        "order":   order,
    })
}

// PaymentService handles payment processing
type PaymentService struct {
    publisher   contract.EventPublisher
    failureRate float64 // Simulate failure rate
}

func NewPaymentService(publisher contract.EventPublisher) *PaymentService {
    return &PaymentService{
        publisher:   publisher,
        failureRate: 0.3, // 30% failure rate for demonstration
    }
}

// HandleOrderCreated processes order created events
func (ps *PaymentService) HandleOrderCreated(ctx context.Context, event contract.Event) error {
    var order OrderEvent
    if err := event.UnmarshalEventPayload(event, &order); err != nil {
        return fmt.Errorf("failed to unmarshal order event: %w", err)
    }

    log.Printf("Processing payment for order %s, amount: $%.2f", order.OrderID, order.Amount)

    // Simulate payment processing delay
    time.Sleep(200 * time.Millisecond)

    // Simulate random failures for demonstration
    if time.Now().UnixNano()%100 < int64(ps.failureRate*100) {
        return fmt.Errorf("payment processing failed for order %s", order.OrderID)
    }

    // Create payment event
    paymentEvent := event.NewEvent("payment.processed", PaymentEvent{
        PaymentID:   fmt.Sprintf("pay-%d", time.Now().UnixNano()),
        OrderID:     order.OrderID,
        Amount:      order.Amount,
        Status:      "completed",
        ProcessedAt: time.Now(),
    })
    paymentEvent.WithHeaders(map[string]string{
        "order_id":    order.OrderID,
        "customer_id": order.CustomerID,
        "source":      "payment-service",
    })

    if err := ps.publisher.Publish(ctx, "payment-events", paymentEvent); err != nil {
        return fmt.Errorf("failed to publish payment event: %w", err)
    }

    log.Printf("Payment processed successfully for order %s", order.OrderID)
    return nil
}

// NotificationService handles notifications
type NotificationService struct {
    publisher contract.EventPublisher
}

func NewNotificationService(publisher contract.EventPublisher) *NotificationService {
    return &NotificationService{publisher: publisher}
}

// HandlePaymentProcessed sends confirmation notifications
func (ns *NotificationService) HandlePaymentProcessed(ctx context.Context, event contract.Event) error {
    var payment PaymentEvent
    if err := event.UnmarshalEventPayload(event, &payment); err != nil {
        return fmt.Errorf("failed to unmarshal payment event: %w", err)
    }

    log.Printf("Sending confirmation notification for order %s", payment.OrderID)

    // Simulate notification sending
    time.Sleep(100 * time.Millisecond)
    
    log.Printf("Confirmation notification sent for order %s", payment.OrderID)
    return nil
}

// HandleReminderNotification processes reminder notifications
func (ns *NotificationService) HandleReminderNotification(ctx context.Context, event contract.Event) error {
    var notification NotificationEvent
    if err := event.UnmarshalEventPayload(event, &notification); err != nil {
        return fmt.Errorf("failed to unmarshal notification event: %w", err)
    }

    log.Printf("Sending reminder to user %s: %s", notification.UserID, notification.Message)

    // Simulate notification sending
    time.Sleep(50 * time.Millisecond)
    
    log.Printf("Reminder notification sent to user %s", notification.UserID)
    return nil
}

// DeadLetterQueueService handles failed messages
type DeadLetterQueueService struct {
    dlq       contract.DeadLetterQueueManager
    publisher contract.EventPublisher
}

func NewDeadLetterQueueService(dlq contract.DeadLetterQueueManager, publisher contract.EventPublisher) *DeadLetterQueueService {
    return &DeadLetterQueueService{
        dlq:       dlq,
        publisher: publisher,
    }
}

// ProcessDeadLetterQueue processes messages in the dead letter queue
func (dlqs *DeadLetterQueueService) ProcessDeadLetterQueue(c fiber.Ctx) error {
    topic := c.Query("topic", "order-events")
    limit := c.QueryInt("limit", 10)

    messages, err := dlqs.dlq.GetMessages(c.Context(), topic, limit)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to get DLQ messages"})
    }

    var results []map[string]interface{}
    for _, msg := range messages {
        results = append(results, map[string]interface{}{
            "id":        msg.ID(),
            "name":      msg.Name(),
            "topic":     msg.Topic(),
            "timestamp": msg.Timestamp(),
            "headers":   msg.Headers(),
            "payload":   msg.Payload(),
        })
    }

    return c.JSON(fiber.Map{
        "topic":    topic,
        "messages": results,
        "count":    len(results),
    })
}

// ReplayDeadLetterMessage replays a specific message from DLQ
func (dlqs *DeadLetterQueueService) ReplayDeadLetterMessage(c fiber.Ctx) error {
    topic := c.Params("topic")
    messageID := c.Params("messageId")

    if err := dlqs.dlq.ReplayMessage(c.Context(), topic, messageID); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to replay message"})
    }

    return c.JSON(fiber.Map{
        "message": "Message replayed successfully",
        "topic":   topic,
        "id":      messageID,
    })
}

// RemoveDeadLetterMessage removes a message from DLQ
func (dlqs *DeadLetterQueueService) RemoveDeadLetterMessage(c fiber.Ctx) error {
    topic := c.Params("topic")
    messageID := c.Params("messageId")

    if err := dlqs.dlq.RemoveMessage(c.Context(), topic, messageID); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to remove message"})
    }

    return c.JSON(fiber.Map{
        "message": "Message removed successfully",
        "topic":   topic,
        "id":      messageID,
    })
}

// MonitoringService provides metrics and health checks
type MonitoringService struct {
    eventManager contract.EventManager
}

func NewMonitoringService(eventManager contract.EventManager) *MonitoringService {
    return &MonitoringService{eventManager: eventManager}
}

// GetEventStats returns event system statistics
func (ms *MonitoringService) GetEventStats(c fiber.Ctx) error {
    stats, err := ms.eventManager.Stats(c.Context())
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to get stats"})
    }

    return c.JSON(stats)
}

// GetHealthCheck returns health status
func (ms *MonitoringService) GetHealthCheck(c fiber.Ctx) error {
    if err := ms.eventManager.Health(c.Context()); err != nil {
        return c.Status(503).JSON(fiber.Map{
            "status": "unhealthy",
            "error":  err.Error(),
        })
    }

    return c.JSON(fiber.Map{"status": "healthy"})
}

// CircuitBreakerHandler demonstrates circuit breaker pattern
type CircuitBreakerHandler struct {
    failures    int
    lastFailure time.Time
    threshold   int
    timeout     time.Duration
}

func NewCircuitBreakerHandler() *CircuitBreakerHandler {
    return &CircuitBreakerHandler{
        threshold: 5,
        timeout:   time.Minute,
    }
}

func (cbh *CircuitBreakerHandler) HandleEvent(ctx context.Context, event contract.Event) error {
    // Check if circuit breaker is open
    if cbh.failures >= cbh.threshold && time.Since(cbh.lastFailure) < cbh.timeout {
        log.Printf("Circuit breaker open, skipping event %s", event.ID())
        return nil // Skip processing
    }

    // Simulate some processing that might fail
    if time.Now().UnixNano()%10 < 2 { // 20% failure rate
        cbh.failures++
        cbh.lastFailure = time.Now()
        return fmt.Errorf("simulated failure for event %s", event.ID())
    }

    // Success - reset failure count
    cbh.failures = 0
    log.Printf("Successfully processed event %s", event.ID())
    return nil
}

// SetupEventHandlers configures all event subscriptions
func SetupEventHandlers(consumer contract.EventConsumer, 
    paymentService *PaymentService,
    notificationService *NotificationService,
    circuitBreakerHandler *CircuitBreakerHandler) {

    // Payment service subscribes to order events
    paymentHandler := event.CreateEventHandler(paymentService.HandleOrderCreated)
    if err := consumer.Subscribe(context.Background(), "order-events", "payment-service", paymentHandler); err != nil {
        log.Fatalf("Failed to subscribe payment service: %v", err)
    }

    // Notification service subscribes to payment events
    confirmationHandler := event.CreateEventHandler(notificationService.HandlePaymentProcessed)
    if err := consumer.Subscribe(context.Background(), "payment-events", "notification-service", confirmationHandler); err != nil {
        log.Fatalf("Failed to subscribe notification service to payment events: %v", err)
    }

    // Notification service subscribes to reminder events
    reminderHandler := event.CreateEventHandler(notificationService.HandleReminderNotification)
    if err := consumer.Subscribe(context.Background(), "notification-events", "reminder-service", reminderHandler); err != nil {
        log.Fatalf("Failed to subscribe notification service to reminder events: %v", err)
    }

    // Circuit breaker handler for demonstration
    circuitBreakerHandlerFunc := event.CreateEventHandler(circuitBreakerHandler.HandleEvent)
    if err := consumer.Subscribe(context.Background(), "order-events", "circuit-breaker-demo", circuitBreakerHandlerFunc); err != nil {
        log.Fatalf("Failed to subscribe circuit breaker handler: %v", err)
    }

    log.Println("Advanced event handlers configured successfully")
}

// Routes sets up all HTTP routes
func Routes(app *fiber.App, 
    orderService *OrderService,
    dlqService *DeadLetterQueueService,
    monitoringService *MonitoringService) {

    api := app.Group("/api/v1")
    
    // Order endpoints
    api.Post("/orders", orderService.CreateOrder)
    
    // Dead letter queue endpoints
    api.Get("/dlq/:topic", dlqService.ProcessDeadLetterQueue)
    api.Post("/dlq/:topic/:messageId/replay", dlqService.ReplayDeadLetterMessage)
    api.Delete("/dlq/:topic/:messageId", dlqService.RemoveDeadLetterMessage)
    
    // Monitoring endpoints
    api.Get("/stats", monitoringService.GetEventStats)
    api.Get("/health", monitoringService.GetHealthCheck)
}

// BackgroundTasks starts background monitoring tasks
func BackgroundTasks(dlq contract.DeadLetterQueueManager) {
    go func() {
        ticker := time.NewTicker(5 * time.Minute)
        defer ticker.Stop()

        for range ticker.C {
            ctx := context.Background()
            stats, err := dlq.Stats(ctx)
            if err != nil {
                log.Printf("Failed to get DLQ stats: %v", err)
                continue
            }

            for topic, topicStats := range stats.TopicQueues {
                if topicStats.MessagesCount > 0 {
                    log.Printf("DLQ Alert: Topic %s has %d messages", topic, topicStats.MessagesCount)
                }
            }
        }
    }()
}

func main() {
    app := goe.New(goe.Options{
        WithHTTP:  true,
        WithEvent: true,
    })

    // Register services
    app.Provide(NewOrderService)
    app.Provide(NewPaymentService)
    app.Provide(NewNotificationService)
    app.Provide(NewDeadLetterQueueService)
    app.Provide(NewMonitoringService)
    app.Provide(NewCircuitBreakerHandler)

    // Register route and event setup
    app.Invoke(Routes)
    app.Invoke(SetupEventHandlers)
    app.Invoke(BackgroundTasks)

    log.Println("Starting advanced event system application...")
    app.Run()
}
```

## Configuration

Create a `.env` file with advanced configuration:

```bash
# HTTP Configuration
HTTP_PORT=8080

# Event System Configuration
EVENT_REDIS_ADDR=localhost:6379
EVENT_REDIS_PASSWORD=
EVENT_REDIS_DB=0

# Consumer Configuration
EVENT_CONSUMER_TIMEOUT=30s
EVENT_MAX_RETRIES=3
EVENT_RETRY_BACKOFF=2s
EVENT_STALE_CONSUMER_TIMEOUT=5m

# Performance Configuration
EVENT_BATCH_SIZE=100
EVENT_MAX_PENDING_MESSAGES=1000
EVENT_CLAIM_MIN_IDLE_TIME=1m
EVENT_CLAIM_INTERVAL=30s

# Delayed Queue Configuration
EVENT_DELAYED_QUEUE_ENABLED=true
EVENT_DELAYED_QUEUE_CHECK_INTERVAL=1s

# Dead Letter Queue Configuration
EVENT_DLQ_TTL=24h
```

## Testing the Advanced Features

### 1. Create Orders (with batch events)
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customerId": "customer-123",
    "amount": 99.99
  }'
```

### 2. Monitor Event Statistics
```bash
curl http://localhost:8080/api/v1/stats
```

### 3. Check Dead Letter Queue
```bash
# Get DLQ messages
curl http://localhost:8080/api/v1/dlq/order-events

# Replay a message
curl -X POST http://localhost:8080/api/v1/dlq/order-events/MESSAGE_ID/replay

# Remove a message
curl -X DELETE http://localhost:8080/api/v1/dlq/order-events/MESSAGE_ID
```

### 4. Health Check
```bash
curl http://localhost:8080/api/v1/health
```

## Load Testing

Create a load test script:

```bash
#!/bin/bash

# Create 100 orders concurrently
for i in {1..100}; do
  curl -X POST http://localhost:8080/api/v1/orders \
    -H "Content-Type: application/json" \
    -d "{
      \"customerId\": \"customer-$i\",
      \"amount\": $((RANDOM % 1000 + 1))
    }" &
done

wait
echo "Load test completed"
```

## Monitoring Commands

```bash
# Redis monitoring
redis-cli --latency-history -i 1
redis-cli info memory
redis-cli info clients

# Stream monitoring
redis-cli XLEN event:stream:order-events
redis-cli XINFO GROUPS event:stream:order-events
redis-cli XPENDING event:stream:order-events payment-service

# DLQ monitoring
redis-cli XLEN event:dlq:order-events
redis-cli XRANGE event:dlq:order-events - + COUNT 10
```

## Key Features Demonstrated

1. **Batch Publishing**: Multiple events published atomically
2. **Delayed Messages**: Scheduled notifications using `PublishWithDelay`
3. **Dead Letter Queue**: Failed message handling with replay/remove capabilities
4. **Circuit Breaker**: Failure protection pattern
5. **Monitoring**: Comprehensive stats and health checks
6. **Background Tasks**: Automated DLQ monitoring
7. **Error Simulation**: Realistic failure scenarios for testing

This advanced example shows how to build a robust, production-ready event-driven system with the GOE event framework.