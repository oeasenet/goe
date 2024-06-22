package event

import (
	"context"
	"go.oease.dev/goe/internal/event/bus"
)

type frameworkEvent struct {
}

const FrameworkEvent = 0

func (ev frameworkEvent) Type() uint32 {
	return FrameworkEvent
}

// FrameworkEventDispatcher initializes a default in-process dispatcher
var frameworkEventDispatcher = bus.NewDispatcher()

// On subscribes to an event, the type of the event will be automatically
// inferred from the provided type. Must be constant for this to work. This
// functions same way as Subscribe() but uses the default dispatcher instead.
func On[T frameworkEvent](handler func(T)) context.CancelFunc {
	return bus.Subscribe(frameworkEventDispatcher, handler)
}

// OnType subscribes to an event with the specified event type. This functions
// same way as SubscribeTo() but uses the default dispatcher instead.
func OnType[T frameworkEvent](eventType uint32, handler func(T)) context.CancelFunc {
	return bus.SubscribeTo(frameworkEventDispatcher, eventType, handler)
}

// Emit writes an event into the dispatcher. This functions same way as
// Publish() but uses the default dispatcher instead.
func Emit[T frameworkEvent](ev T) {
	bus.Publish(frameworkEventDispatcher, ev)
}

func Close() {
	frameworkEventDispatcher.Close()
}
