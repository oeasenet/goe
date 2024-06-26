package bus

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
)

// Bus represents an event contract
type Bus interface {
	Type() uint32
}

// ------------------------------------- Dispatcher -------------------------------------

// Dispatcher represents an event dispatcher.
type Dispatcher struct {
	subs sync.Map
	done chan struct{} // Cancellation
	df   time.Duration // Flush interval
}

// NewDispatcher creates a new dispatcher of events.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		df:   500 * time.Microsecond,
		done: make(chan struct{}),
	}
}

// Close closes the dispatcher
func (d *Dispatcher) Close() error {
	close(d.done)
	return nil
}

// isClosed returns whether the dispatcher is closed or not
func (d *Dispatcher) isClosed() bool {
	select {
	case <-d.done:
		return true
	default:
		return false
	}
}

// Subscribe subscribes to an event, the type of the event will be automatically
// inferred from the provided type. Must be constant for this to work.
func Subscribe[T Bus](broker *Dispatcher, handler func(T)) context.CancelFunc {
	var event T
	return SubscribeTo(broker, event.Type(), handler)
}

// SubscribeTo subscribes to an event with the specified event type.
func SubscribeTo[T Bus](broker *Dispatcher, eventType uint32, handler func(T)) context.CancelFunc {
	if broker.isClosed() {
		panic(errClosed)
	}

	// Add to consumer group, if it doesn't exist it will create one
	s, loaded := broker.subs.LoadOrStore(eventType, &group[T]{
		cond: sync.NewCond(new(sync.Mutex)),
	})
	group := groupOf[T](eventType, s)
	sub := group.Add(handler)

	// Start flushing asynchronously if we just created a new group
	if !loaded {
		go group.Process(broker.df, broker.done)
	}

	// Return unsubscribe function
	return func() {
		group.Del(sub)
	}
}

// Publish writes an event into the dispatcher
func Publish[T Bus](broker *Dispatcher, ev T) {
	if s, ok := broker.subs.Load(ev.Type()); ok {
		group := groupOf[T](ev.Type(), s)
		group.Broadcast(ev)
	}
}

// Count counts the number of subscribers, this is for testing only.
func (d *Dispatcher) count(eventType uint32) int {
	if group, ok := d.subs.Load(eventType); ok {
		return group.(interface{ Count() int }).Count()
	}
	return 0
}

// groupOf casts the subscriber group to the specified generic type
func groupOf[T Bus](eventType uint32, subs any) *group[T] {
	if group, ok := subs.(*group[T]); ok {
		return group
	}

	panic(errConflict[T](eventType, subs))
}

// ------------------------------------- Subscriber -------------------------------------

// consumer represents a consumer with a message queue
type consumer[T Bus] struct {
	queue []T  // Current work queue
	stop  bool // Stop signal
}

// Listen listens to the event queue and processes events
func (s *consumer[T]) Listen(c *sync.Cond, fn func(T)) {
	pending := make([]T, 0, 128)

	for {
		c.L.Lock()
		for len(s.queue) == 0 {
			switch {
			case s.stop:
				c.L.Unlock()
				return
			default:
				c.Wait()
			}
		}

		// Swap buffers and reset the current queue
		temp := s.queue
		s.queue = pending
		pending = temp
		s.queue = s.queue[:0]
		c.L.Unlock()

		// Outside of the critical section, process the work
		for i := 0; i < len(pending); i++ {
			fn(pending[i])
		}
	}
}

// ------------------------------------- Subscriber Group -------------------------------------

// group represents a consumer group
type group[T Bus] struct {
	cond *sync.Cond
	subs []*consumer[T]
}

// Process periodically broadcasts events
func (s *group[T]) Process(interval time.Duration, done chan struct{}) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			s.cond.Broadcast()
		}
	}
}

// Broadcast sends an event to all consumers
func (s *group[T]) Broadcast(ev T) {
	s.cond.L.Lock()
	for _, sub := range s.subs {
		sub.queue = append(sub.queue, ev)
	}
	s.cond.L.Unlock()
}

// Add adds a subscriber to the list
func (s *group[T]) Add(handler func(T)) *consumer[T] {
	sub := &consumer[T]{
		queue: make([]T, 0, 128),
	}

	// Add the consumer to the list of active consumers
	s.cond.L.Lock()
	s.subs = append(s.subs, sub)
	s.cond.L.Unlock()

	// Start listening
	go sub.Listen(s.cond, handler)
	return sub
}

// Del removes a subscriber from the list
func (s *group[T]) Del(sub *consumer[T]) {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()

	// Search and remove the subscriber
	sub.stop = true
	subs := make([]*consumer[T], 0, len(s.subs))
	for _, v := range s.subs {
		if v != sub {
			subs = append(subs, v)
		}
	}
	s.subs = subs
}

// ------------------------------------- Debugging -------------------------------------

var errClosed = fmt.Errorf("event dispatcher is closed")

// Count returns the number of subscribers in this group
func (s *group[T]) Count() int {
	return len(s.subs)
}

// String returns string representation of the type
func (s *group[T]) String() string {
	typ := reflect.TypeOf(s).String()
	idx := strings.LastIndex(typ, "/")
	typ = typ[idx+1 : len(typ)-1]
	return typ
}

// errConflict returns a conflict message
func errConflict[T any](eventType uint32, existing any) string {
	var want T
	return fmt.Sprintf(
		"conflicting event type, want=<%T>, registered=<%s>, event=0x%v",
		want, existing, eventType,
	)
}
