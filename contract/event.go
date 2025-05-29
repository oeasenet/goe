package contract

import (
	"context"
)

// Event represents the event module interface
type Event interface {
	Module

	// Publish publishes an event with the given name and payload
	Publish(ctx context.Context, name string, payload interface{}) error

	// Subscribe subscribes to events with the given name
	Subscribe(name string, handler EventHandler) (Subscription, error)

	// SubscribeAsync subscribes to events with the given name asynchronously
	SubscribeAsync(name string, handler EventHandler) (Subscription, error)

	// SubscribeOnce subscribes to events with the given name and unsubscribes after the first event
	SubscribeOnce(name string, handler EventHandler) (Subscription, error)

	// SubscribeOnceAsync subscribes to events with the given name asynchronously and unsubscribes after the first event
	SubscribeOnceAsync(name string, handler EventHandler) (Subscription, error)

	// HasSubscribers returns true if there are subscribers for the given event name
	HasSubscribers(name string) bool

	// SubscribersCount returns the number of subscribers for the given event name
	SubscribersCount(name string) int
}

// EventHandler is a function that handles an event
type EventHandler func(ctx context.Context, payload interface{}) error

// Subscription represents an event subscription
type Subscription interface {
	// Unsubscribe unsubscribes from the event
	Unsubscribe() error

	// Name returns the name of the event
	Name() string
}

// EventProvider is a function that provides an Event instance
type EventProvider func() Event
