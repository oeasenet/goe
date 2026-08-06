// Package shutdown provides graceful shutdown management with priority-based hooks
// for orderly shutdown of application components.
package shutdown

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	"go.oease.dev/goe/v2/contract"
)

// Priority constants for shutdown hooks
const (
	// PriorityHTTP is the priority for HTTP server shutdown (first to stop)
	PriorityHTTP = 100
	// PriorityWorkers is the priority for background workers
	PriorityWorkers = 80
	// PriorityEvent is the priority for event consumers
	PriorityEvent = 60
	// PriorityDatabase is the priority for database connections
	PriorityDatabase = 40
	// PriorityCache is the priority for cache connections
	PriorityCache = 20
	// PriorityTelemetry is the priority for telemetry flush (last to stop)
	PriorityTelemetry = 10
)

// Hook represents a shutdown hook function
// Deprecated: Use contract.ShutdownHook instead
type Hook = contract.ShutdownHook

// Manager satisfies contract.ShutdownManager. Asserting it here means the
// interface is checked by the compiler rather than being a declaration nothing
// enforces, so the two cannot drift apart.
var _ contract.ShutdownManager = (*Manager)(nil)

// hookEntry stores a hook with its priority and name
type hookEntry struct {
	name     string
	priority int
	hook     contract.ShutdownHook
}

// Manager manages graceful shutdown of the application
type Manager struct {
	mu           sync.RWMutex
	hooks        []hookEntry
	timeout      time.Duration
	drainTimeout time.Duration
	logger       contract.Logger
	isShutdown   bool
}

// NewManager creates a new shutdown manager
func NewManager(logger contract.Logger, timeout, drainTimeout time.Duration) *Manager {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	if drainTimeout == 0 {
		drainTimeout = 5 * time.Second
	}

	return &Manager{
		hooks:        make([]hookEntry, 0),
		timeout:      timeout,
		drainTimeout: drainTimeout,
		logger:       logger,
	}
}

// RegisterHook adds a shutdown hook with the given priority
// Higher priority hooks are executed first during shutdown
func (m *Manager) RegisterHook(name string, priority int, hook contract.ShutdownHook) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.hooks = append(m.hooks, hookEntry{
		name:     name,
		priority: priority,
		hook:     hook,
	})

	// Sort hooks by priority (descending)
	slices.SortFunc(m.hooks, func(a, b hookEntry) int {
		return cmp.Compare(b.priority, a.priority) // higher priority first
	})

	m.logger.Debug("Shutdown hook registered",
		"name", name,
		"priority", priority,
	)
}

// UnregisterHook removes a shutdown hook by name
func (m *Manager) UnregisterHook(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, entry := range m.hooks {
		if entry.name == name {
			m.hooks = append(m.hooks[:i], m.hooks[i+1:]...)
			m.logger.Debug("Shutdown hook unregistered", "name", name)
			return
		}
	}
}

// Shutdown executes all registered hooks in priority order
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	if m.isShutdown {
		m.mu.Unlock()
		return nil
	}
	m.isShutdown = true
	hooks := make([]hookEntry, len(m.hooks))
	copy(hooks, m.hooks)
	m.mu.Unlock()

	// Create timeout context if not already set
	shutdownCtx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	m.logger.Info("Starting graceful shutdown",
		"hooks", len(hooks),
		"timeout", m.timeout.String(),
	)

	var errs []error
	for _, entry := range hooks {
		m.logger.Debug("Executing shutdown hook", "name", entry.name, "priority", entry.priority)

		if err := entry.hook(shutdownCtx); err != nil {
			m.logger.Error("Shutdown hook failed",
				"name", entry.name,
				"error", err,
			)
			errs = append(errs, err)
			// Continue with other hooks even if one fails
		}
	}

	if len(errs) > 0 {
		m.logger.Warn("Graceful shutdown completed with errors", "errors", len(errs))
		return errors.Join(errs...)
	}

	m.logger.Info("Graceful shutdown completed successfully")
	return nil
}

// DrainTimeout returns the drain timeout for HTTP connections
func (m *Manager) DrainTimeout() time.Duration {
	return m.drainTimeout
}

// Timeout returns the total shutdown timeout
func (m *Manager) Timeout() time.Duration {
	return m.timeout
}

// IsShutdown returns true if shutdown has been initiated
func (m *Manager) IsShutdown() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isShutdown
}
