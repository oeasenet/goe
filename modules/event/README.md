# Event Module

The Event module provides a simple event bus for publishing and subscribing to events within the application. It allows for loose coupling between components by enabling event-driven communication. The module is implemented in the [`bus`](https://github.com/oeasenet/goe/blob/main/modules/event/bus/bus.go) package.

## Features

- Publish-subscribe pattern
- Type-safe event handling using Go generics
- Synchronous and asynchronous event processing
- Support for multiple subscribers per event
- Cancellable subscriptions
- Efficient event broadcasting
- Automatic event type inference
- Low memory footprint

## Usage

### Initialization

The event module can be used in two ways:

1. Using the default dispatcher (recommended for most cases):

```go
import "go.oease.dev/goe/modules/event"
```

2. Creating a custom dispatcher:

```go
import "go.oease.dev/goe/modules/event/bus"

// Create a new dispatcher
dispatcher := bus.NewDispatcher()
```

### Defining Events

Define your event types by creating structs that implement the `bus.Bus` interface:

```go
import "go.oease.dev/goe/modules/event/bus"

// Define event types
const (
    UserCreatedEventType uint32 = iota + 1
    OrderCompletedEventType
)

// UserCreatedEvent represents an event that is fired when a new user is created
type UserCreatedEvent struct {
    ID    string
    Email string
    Name  string
}

// Type implements the bus.Bus interface
func (e UserCreatedEvent) Type() uint32 {
    return UserCreatedEventType
}

// OrderCompletedEvent represents an event that is fired when an order is completed
type OrderCompletedEvent struct {
    OrderID     string
    UserID      string
    TotalAmount float64
}

// Type implements the bus.Bus interface
func (e OrderCompletedEvent) Type() uint32 {
    return OrderCompletedEventType
}
```

### Publishing Events

#### Using the Default Dispatcher

```go
import "go.oease.dev/goe/modules/event"

// Create an event instance
userEvent := UserCreatedEvent{
    ID:    "user123",
    Email: "user@example.com",
    Name:  "John Doe",
}

// Publish the event
event.Emit(userEvent)
```

#### Using a Custom Dispatcher

```go
import "go.oease.dev/goe/modules/event/bus"

// Create a dispatcher
dispatcher := bus.NewDispatcher()

// Create an event instance
orderEvent := OrderCompletedEvent{
    OrderID:     "order456",
    UserID:      "user123",
    TotalAmount: 99.99,
}

// Publish the event
bus.Publish(dispatcher, orderEvent)
```

### Subscribing to Events

#### Using the Default Dispatcher

```go
import (
    "fmt"
    "go.oease.dev/goe/modules/event"
)

// Subscribe to UserCreatedEvent
cancelFunc := event.On(func(e UserCreatedEvent) {
    fmt.Printf("User created: %s (%s)\n", e.Name, e.Email)
})

// Later, to unsubscribe
cancelFunc()
```

#### Using a Custom Dispatcher

```go
import (
    "fmt"
    "go.oease.dev/goe/modules/event/bus"
)

// Create a dispatcher
dispatcher := bus.NewDispatcher()

// Subscribe to OrderCompletedEvent
cancelFunc := bus.Subscribe(dispatcher, func(e OrderCompletedEvent) {
    fmt.Printf("Order completed: %s, Total: $%.2f\n", e.OrderID, e.TotalAmount)
})

// Later, to unsubscribe
cancelFunc()
```

### Real-world Examples

#### User Registration Flow

```go
// In your user service
func RegisterUser(userData UserRegistrationData) (*User, error) {
    // Create the user in the database
    user, err := createUserInDatabase(userData)
    if err != nil {
        return nil, err
    }
    
    // Emit an event that the user was created
    event.Emit(UserCreatedEvent{
        ID:    user.ID,
        Email: user.Email,
        Name:  user.Name,
    })
    
    return user, nil
}

// In your email service
func init() {
    // Subscribe to user creation events to send welcome emails
    event.On(func(e UserCreatedEvent) {
        sendWelcomeEmail(e.Email, e.Name)
    })
}

// In your analytics service
func init() {
    // Subscribe to user creation events to track user registrations
    event.On(func(e UserCreatedEvent) {
        trackUserRegistration(e.ID)
    })
}
```

#### Order Processing Flow

```go
// In your order service
func CompleteOrder(orderID string) error {
    // Process the order in the database
    order, err := markOrderAsCompleted(orderID)
    if err != nil {
        return err
    }
    
    // Emit an event that the order was completed
    event.Emit(OrderCompletedEvent{
        OrderID:     order.ID,
        UserID:      order.UserID,
        TotalAmount: order.TotalAmount,
    })
    
    return nil
}

// In your notification service
func init() {
    // Subscribe to order completion events to send notifications
    event.On(func(e OrderCompletedEvent) {
        sendOrderConfirmation(e.UserID, e.OrderID, e.TotalAmount)
    })
}

// In your inventory service
func init() {
    // Subscribe to order completion events to update inventory
    event.On(func(e OrderCompletedEvent) {
        updateInventoryForOrder(e.OrderID)
    })
}
```

## API Reference

### Event Interface

```go
// Bus represents an event contract
type Bus interface {
    Type() uint32
}
```

### Default Dispatcher API

```go
// On subscribes to an event, the type of the event will be automatically
// inferred from the provided type.
func On[T frameworkEvent](handler func(T)) context.CancelFunc

// OnType subscribes to an event with the specified event type.
func OnType[T frameworkEvent](eventType uint32, handler func(T)) context.CancelFunc

// Emit writes an event into the dispatcher.
func Emit[T frameworkEvent](ev T)

// Close closes the default dispatcher.
func Close()
```

### Custom Dispatcher API

```go
// NewDispatcher creates a new dispatcher of events.
func NewDispatcher() *Dispatcher

// Subscribe subscribes to an event, the type of the event will be automatically
// inferred from the provided type.
func Subscribe[T Bus](broker *Dispatcher, handler func(T)) context.CancelFunc

// SubscribeTo subscribes to an event with the specified event type.
func SubscribeTo[T Bus](broker *Dispatcher, eventType uint32, handler func(T)) context.CancelFunc

// Publish writes an event into the dispatcher.
func Publish[T Bus](broker *Dispatcher, ev T)

// Close closes the dispatcher.
func (d *Dispatcher) Close() error
```

## Implementation Details

The event module is implemented using Go's generics to provide type-safe event handling. The implementation consists of several key components:

### Dispatcher

The [`Dispatcher`](https://github.com/oeasenet/goe/blob/main/modules/event/bus/bus.go#L20) is the central component that manages subscriptions and event publishing. It maintains a map of event types to subscriber groups.

### Subscriber Groups

Each event type has a [`group`](https://github.com/oeasenet/goe/blob/main/modules/event/bus/bus.go#L147) of subscribers. When an event is published, it is broadcast to all subscribers in the group.

### Consumers

Each subscriber is represented by a [`consumer`](https://github.com/oeasenet/goe/blob/main/modules/event/bus/bus.go#L109) that has its own event queue. Consumers process events asynchronously to avoid blocking the publisher.

### Event Processing

Events are processed asynchronously in separate goroutines. The dispatcher periodically broadcasts events to all subscribers, which then process the events in their own goroutines.

### Performance Considerations

- Events are processed in batches to improve performance
- The dispatcher uses a sync.Map for thread-safe access to subscriber groups
- Subscribers use a condition variable for efficient waiting and notification
- Event queues are pre-allocated to reduce memory allocations