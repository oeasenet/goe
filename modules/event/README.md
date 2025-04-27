# Event Module

The Event module provides a simple event bus for publishing and subscribing to events within the application. It allows for loose coupling between components by enabling event-driven communication.

## Features

- Publish-subscribe pattern
- Type-safe event handling
- Synchronous event processing
- Support for multiple subscribers per event

## Usage

### Initialization

The event module can be used directly without explicit initialization:

```go
import "go.oease.dev/goe/modules/event/bus"
```

### Defining Events

Define your event types by creating structs:

```go
// UserCreatedEvent represents an event that is fired when a new user is created
type UserCreatedEvent struct {
    ID    string
    Email string
    Name  string
}

// OrderCompletedEvent represents an event that is fired when an order is completed
type OrderCompletedEvent struct {
    OrderID     string
    UserID      string
    TotalAmount float64
}
```

### Publishing Events

```go
// Create an event instance
event := UserCreatedEvent{
    ID:    "user123",
    Email: "user@example.com",
    Name:  "John Doe",
}

// Publish the event
bus.Publish(event)
```

### Subscribing to Events

```go
// Subscribe to UserCreatedEvent
bus.Subscribe(func(e UserCreatedEvent) {
    fmt.Printf("User created: %s (%s)\n", e.Name, e.Email)
})

// Subscribe to OrderCompletedEvent
bus.Subscribe(func(e OrderCompletedEvent) {
    fmt.Printf("Order completed: %s, Total: $%.2f\n", e.OrderID, e.TotalAmount)
})
```

### Unsubscribing

```go
// Subscribe and store the subscription ID
subID := bus.Subscribe(func(e UserCreatedEvent) {
    fmt.Printf("User created: %s\n", e.Name)
})

// Later, unsubscribe using the subscription ID
bus.Unsubscribe(subID)
```

## Implementation Details

The event module uses a simple in-memory event bus implementation. Events are processed synchronously in the same goroutine that publishes them.

The module uses Go's type system to ensure that events are delivered to the correct subscribers based on the event type.