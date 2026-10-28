package contract

import (
	"context"
	"time"
)

// ShutdownHook represents a function called during graceful shutdown
type ShutdownHook func(ctx context.Context) error

// ShutdownManager manages graceful shutdown of the application
type ShutdownManager interface {
	// RegisterHook adds a shutdown hook with the given priority
	// Higher priority hooks are executed first during shutdown
	RegisterHook(name string, priority int, hook ShutdownHook)

	// UnregisterHook removes a shutdown hook by name
	UnregisterHook(name string)

	// Shutdown executes all registered hooks in priority order
	Shutdown(ctx context.Context) error

	// DrainTimeout returns the drain timeout for HTTP connections
	DrainTimeout() time.Duration

	// Timeout returns the total shutdown timeout
	Timeout() time.Duration

	// IsShutdown returns true if shutdown has been initiated
	IsShutdown() bool

	// Stopping returns a channel that is closed once shutdown starts, before
	// any hook runs: long-lived work (streams, loops) watches it to end early.
	Stopping() <-chan struct{}

	// Go runs fn in a new goroutine that shutdown waits for. fn's context is
	// cancelled only when the shutdown budget runs out before fn returns.
	Go(name string, fn func(ctx context.Context))
}

// ShutdownConfig defines configuration for the shutdown module
type ShutdownConfig struct {
	// Timeout is the maximum time for all shutdown hooks (default: 30s)
	Timeout time.Duration
	// DrainTimeout is the time to wait for HTTP connections to drain (default: 5s)
	DrainTimeout time.Duration
}
