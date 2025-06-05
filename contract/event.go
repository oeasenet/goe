package contract

import "context"

// Event represents an event in the system
type Event interface {
	// Name returns the event name
	Name() string

	// Payload returns the event payload
	Payload() any

	// Context returns the event context
	Context() context.Context
}

// EventDispatcher defines the event dispatcher interface
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

// EventListener handles events
type EventListener interface {
	// Handle handles an event
	Handle(event Event) error
}

// EventSubscriber subscribes to events
type EventSubscriber interface {
	// Subscribe returns event patterns to subscribe to
	Subscribe() []string

	// Handle handles an event
	Handle(event Event) error
}
