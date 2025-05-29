package event

import (
	"context"
	"sync"

	"go.oease.dev/goe/v2/contract"
)

// Event implements the contract.Event interface
type Event struct {
	mu           sync.RWMutex
	subscribers  map[string][]subscriber
	eventCounter map[string]int
}

// subscriber represents an event subscriber
type subscriber struct {
	handler contract.EventHandler
	once    bool
}

// subscription represents an event subscription
type subscription struct {
	event  string
	index  int
	parent *Event
}

// New creates a new Event instance
func New() *Event {
	return &Event{
		subscribers:  make(map[string][]subscriber),
		eventCounter: make(map[string]int),
	}
}

// Name returns the name of the module
func (e *Event) Name() string {
	return "event"
}

// Initialize initializes the event module
func (e *Event) Initialize(ctx context.Context) error {
	return nil
}

// Start starts the event module
func (e *Event) Start(ctx context.Context) error {
	return nil
}

// Stop stops the event module
func (e *Event) Stop(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Clear all subscribers
	e.subscribers = make(map[string][]subscriber)
	e.eventCounter = make(map[string]int)

	return nil
}

// Publish publishes an event with the given name and payload
func (e *Event) Publish(ctx context.Context, name string, payload interface{}) error {
	e.mu.RLock()
	subs := e.subscribers[name]
	e.mu.RUnlock()

	// Increment event counter
	e.mu.Lock()
	e.eventCounter[name]++
	e.mu.Unlock()

	// Call all subscribers
	for i, sub := range subs {
		if err := sub.handler(ctx, payload); err != nil {
			return err
		}

		// If this is a once subscriber, remove it
		if sub.once {
			e.mu.Lock()
			e.removeSubscriberAt(name, i)
			e.mu.Unlock()
		}
	}

	return nil
}

// Subscribe subscribes to events with the given name
func (e *Event) Subscribe(name string, handler contract.EventHandler) (contract.Subscription, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	index := len(e.subscribers[name])
	e.subscribers[name] = append(e.subscribers[name], subscriber{
		handler: handler,
		once:    false,
	})

	return &subscription{
		event:  name,
		index:  index,
		parent: e,
	}, nil
}

// SubscribeAsync subscribes to events with the given name asynchronously
func (e *Event) SubscribeAsync(name string, handler contract.EventHandler) (contract.Subscription, error) {
	asyncHandler := func(ctx context.Context, payload interface{}) error {
		go func() {
			_ = handler(ctx, payload)
		}()
		return nil
	}

	return e.Subscribe(name, asyncHandler)
}

// SubscribeOnce subscribes to events with the given name and unsubscribes after the first event
func (e *Event) SubscribeOnce(name string, handler contract.EventHandler) (contract.Subscription, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	index := len(e.subscribers[name])
	e.subscribers[name] = append(e.subscribers[name], subscriber{
		handler: handler,
		once:    true,
	})

	return &subscription{
		event:  name,
		index:  index,
		parent: e,
	}, nil
}

// SubscribeOnceAsync subscribes to events with the given name asynchronously and unsubscribes after the first event
func (e *Event) SubscribeOnceAsync(name string, handler contract.EventHandler) (contract.Subscription, error) {
	asyncHandler := func(ctx context.Context, payload interface{}) error {
		go func() {
			_ = handler(ctx, payload)
		}()
		return nil
	}

	return e.SubscribeOnce(name, asyncHandler)
}

// HasSubscribers returns true if there are subscribers for the given event name
func (e *Event) HasSubscribers(name string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return len(e.subscribers[name]) > 0
}

// SubscribersCount returns the number of subscribers for the given event name
func (e *Event) SubscribersCount(name string) int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return len(e.subscribers[name])
}

// removeSubscriberAt removes a subscriber at the given index
func (e *Event) removeSubscriberAt(name string, index int) {
	if index < 0 || index >= len(e.subscribers[name]) {
		return
	}

	// Remove the subscriber
	e.subscribers[name] = append(e.subscribers[name][:index], e.subscribers[name][index+1:]...)

	// Update indices for all subscriptions
	for i := index; i < len(e.subscribers[name]); i++ {
		// This would require tracking all subscriptions, which we're not doing for simplicity
	}
}

// Unsubscribe unsubscribes from the event
func (s *subscription) Unsubscribe() error {
	s.parent.mu.Lock()
	defer s.parent.mu.Unlock()

	s.parent.removeSubscriberAt(s.event, s.index)
	return nil
}

// Name returns the name of the event
func (s *subscription) Name() string {
	return s.event
}

// Provider provides an Event instance
func Provider() contract.Event {
	return New()
}
