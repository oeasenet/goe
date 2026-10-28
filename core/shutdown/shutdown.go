// Package shutdown provides graceful shutdown management with priority-based hooks
// for orderly shutdown of application components.
//
// Shutdown runs once, in three phases ordered by hook priority, highest first:
// hooks above PriorityHTTP run while the HTTP server still serves (pre-drain
// work such as deregistration), the hook at PriorityHTTP drains the server, and
// hooks below it run once the server stopped — among them the wait for tasks
// started with Go, at PriorityWorkers. Stopping is closed before the first hook
// runs. All phases share three quarters of the shutdown timeout; the remaining
// quarter is left for the modules that stop after the manager.
package shutdown

import (
	"cmp"
	"context"
	"errors"
	"runtime/debug"
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

// tasksHookName names the manager's own wait for tasks started with Go.
const tasksHookName = "background-tasks"

// lateHookGrace is how long a hook still gets to finish once the shared budget
// is gone, so a quick hook that runs after a hung one is not reported as hung.
const lateHookGrace = 25 * time.Millisecond

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
	// inline marks the manager's own hooks: they honour their context, so
	// they run on the shutdown goroutine and finish (and report) in order.
	inline bool
}

// Manager manages graceful shutdown of the application
type Manager struct {
	mu           sync.RWMutex
	hooks        []hookEntry
	timeout      time.Duration
	drainTimeout time.Duration
	logger       contract.Logger
	isShutdown   bool
	stopping     chan struct{}

	tasksMu     sync.Mutex
	tasks       sync.WaitGroup
	running     map[string]int
	tasksClosed bool
	taskCtx     context.Context
	cancelTasks context.CancelFunc
}

// NewManager creates a new shutdown manager
func NewManager(logger contract.Logger, timeout, drainTimeout time.Duration) *Manager {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if drainTimeout <= 0 {
		drainTimeout = 5 * time.Second
	}

	// The drain runs inside the hooks' shared budget, so it never gets more.
	if budget := timeout * 3 / 4; drainTimeout > budget {
		logger.Warn("Drain timeout exceeds the shutdown budget",
			"drain_timeout", drainTimeout.String(),
			"budget", budget.String(),
		)
	}

	taskCtx, cancelTasks := context.WithCancel(context.Background())
	return &Manager{
		hooks:        make([]hookEntry, 0),
		timeout:      timeout,
		drainTimeout: drainTimeout,
		logger:       logger,
		stopping:     make(chan struct{}),
		running:      make(map[string]int),
		taskCtx:      taskCtx,
		cancelTasks:  cancelTasks,
	}
}

// byPriority orders hooks highest priority first. It is used with a stable
// sort, so hooks of equal priority keep their registration order.
func byPriority(a, b hookEntry) int {
	return cmp.Compare(b.priority, a.priority)
}

// RegisterHook adds a shutdown hook with the given priority.
// Higher priority hooks are executed first during shutdown; hooks of equal
// priority run in the order they were registered.
func (m *Manager) RegisterHook(name string, priority int, hook contract.ShutdownHook) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.hooks = append(m.hooks, hookEntry{
		name:     name,
		priority: priority,
		hook:     hook,
	})
	slices.SortStableFunc(m.hooks, byPriority)

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

// Stopping returns a channel that is closed once shutdown starts, before any
// hook runs. It is never closed when the application does not shut down.
func (m *Manager) Stopping() <-chan struct{} {
	return m.stopping
}

// Go runs fn in a new goroutine that shutdown waits for. Its context is not
// cancelled when shutdown starts: tasks finish their work while the server
// drains. The wait happens at PriorityWorkers; when the shutdown budget runs
// out first, the context is cancelled and the tasks still running are reported.
// A panic in fn is recovered and logged. Tasks started once the wait began are
// not waited for, and those started after shutdown get a cancelled context.
func (m *Manager) Go(name string, fn func(ctx context.Context)) {
	m.tasksMu.Lock()
	if m.tasksClosed {
		m.tasksMu.Unlock()
		go m.runTask(name, fn)
		return
	}
	m.running[name]++
	m.tasks.Add(1)
	m.tasksMu.Unlock()

	go func() {
		defer m.tasks.Done()
		defer m.taskFinished(name)
		m.runTask(name, fn)
	}()
}

func (m *Manager) runTask(name string, fn func(ctx context.Context)) {
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("Background task panicked",
				"task", name,
				"panic", r,
				"stack", string(debug.Stack()),
			)
		}
	}()
	fn(m.taskCtx)
}

func (m *Manager) taskFinished(name string) {
	m.tasksMu.Lock()
	defer m.tasksMu.Unlock()
	if m.running[name]--; m.running[name] <= 0 {
		delete(m.running, name)
	}
}

// awaitTasks is the manager's own hook: it waits for the tasks started with Go
// until its context is done, then cancels them and reports the ones left.
func (m *Manager) awaitTasks(ctx context.Context) error {
	m.tasksMu.Lock()
	m.tasksClosed = true
	m.tasksMu.Unlock()

	done := make(chan struct{})
	go func() {
		m.tasks.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		m.tasksMu.Lock()
		names := make([]string, 0, len(m.running))
		for name := range m.running {
			names = append(names, name)
		}
		m.tasksMu.Unlock()
		slices.Sort(names)

		m.cancelTasks()
		// The wait may start with the budget already gone (a hung hook used
		// it up); with no task running there is nothing to report.
		if len(names) > 0 {
			m.logger.Warn("Background tasks still running", "tasks", names)
		}
	}
	return nil
}

// Shutdown runs the registered hooks once, highest priority first, within
// three quarters of the shutdown timeout. A hook that is still running when
// that budget is gone is reported and left behind; the hooks after it still
// run. Errors returned by hooks are joined and returned.
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	if m.isShutdown {
		m.mu.Unlock()
		return nil
	}
	m.isShutdown = true
	// The task wait goes first among the hooks at PriorityWorkers.
	hooks := append([]hookEntry{{name: tasksHookName, priority: PriorityWorkers, hook: m.awaitTasks, inline: true}}, m.hooks...)
	m.mu.Unlock()
	slices.SortStableFunc(hooks, byPriority)

	close(m.stopping)
	defer m.cancelTasks()

	budget := m.timeout * 3 / 4
	shutdownCtx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()

	m.logger.Info("Starting graceful shutdown",
		"hooks", len(hooks)-1,
		"timeout", m.timeout.String(),
		"budget", budget.String(),
	)

	var errs []error
	for _, entry := range hooks {
		m.logger.Debug("Executing shutdown hook", "name", entry.name, "priority", entry.priority)

		run := m.run
		if entry.inline {
			run = func(ctx context.Context, entry hookEntry) error { return entry.hook(ctx) }
		}
		if err := run(shutdownCtx, entry); err != nil {
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

// run runs one hook in its own goroutine and waits for it until ctx is done
// (or, when the budget is already gone, for lateHookGrace).
func (m *Manager) run(ctx context.Context, entry hookEntry) error {
	done := make(chan error, 1)
	go func() { done <- entry.hook(ctx) }()

	wait := ctx
	if ctx.Err() != nil {
		var cancel context.CancelFunc
		wait, cancel = context.WithTimeout(context.WithoutCancel(ctx), lateHookGrace)
		defer cancel()
	}

	select {
	case err := <-done:
		return m.result(ctx, entry, err)
	case <-wait.Done():
		select {
		case err := <-done:
			return m.result(ctx, entry, err)
		default:
		}
		m.outOfTime(entry, nil)
		return nil
	}
}

// result treats a hook that gave up because the budget ran out like one that
// is still running: it is reported, not returned as a failure.
func (m *Manager) result(ctx context.Context, entry hookEntry, err error) error {
	if err != nil && ctx.Err() != nil &&
		(errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)) {
		m.outOfTime(entry, err)
		return nil
	}
	return err
}

func (m *Manager) outOfTime(entry hookEntry, err error) {
	args := []any{"name", entry.name, "priority", entry.priority}
	if err != nil {
		args = append(args, "error", err)
	}
	m.logger.Warn("Shutdown hook ran out of time", args...)
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
