package shutdown

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type logEntry struct {
	level, msg string
	fields     map[string]any
}

// recorder is a contract.Logger that keeps the Info, Warn and Error lines it
// receives, so tests can assert on what the manager reported.
type recorder struct {
	*testLogger
	mu      sync.Mutex
	entries []logEntry
}

func newRecorder(t *testing.T) *recorder {
	return &recorder{testLogger: &testLogger{t: t}}
}

func (r *recorder) record(level, msg string, args []any) {
	fields := map[string]any{}
	for i := 0; i+1 < len(args); i += 2 {
		if key, ok := args[i].(string); ok {
			fields[key] = args[i+1]
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, logEntry{level: level, msg: msg, fields: fields})
}

func (r *recorder) Info(msg string, args ...any)  { r.record("info", msg, args) }
func (r *recorder) Warn(msg string, args ...any)  { r.record("warn", msg, args) }
func (r *recorder) Error(msg string, args ...any) { r.record("error", msg, args) }

func (r *recorder) find(level, msg string) []logEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	var found []logEntry
	for _, e := range r.entries {
		if e.level == level && e.msg == msg {
			found = append(found, e)
		}
	}
	return found
}

func TestManager_StoppingClosesBeforeTheFirstHook(t *testing.T) {
	m := NewManager(newRecorder(t), time.Second, 0)

	select {
	case <-m.Stopping():
		t.Fatal("Stopping is closed before shutdown started")
	default:
	}

	var closedInHook atomic.Bool
	m.RegisterHook("deregister", PriorityHTTP+10, func(context.Context) error {
		select {
		case <-m.Stopping():
			closedInHook.Store(true)
		default:
		}
		return nil
	})

	require.NoError(t, m.Shutdown(context.Background()))
	assert.True(t, closedInHook.Load(), "the first hook must already see Stopping closed")
}

func TestManager_EqualPrioritiesKeepRegistrationOrder(t *testing.T) {
	m := NewManager(newRecorder(t), time.Second, 0)

	var mu sync.Mutex
	var order, want []string
	priorities := []int{PriorityWorkers, PriorityHTTP + 10, PriorityDatabase}
	for i := range 48 {
		name := fmt.Sprintf("hook-%02d", i)
		priority := priorities[(i*7)%len(priorities)]
		m.RegisterHook(name, priority, func(context.Context) error {
			mu.Lock()
			defer mu.Unlock()
			order = append(order, name)
			return nil
		})
	}
	for _, p := range []int{PriorityHTTP + 10, PriorityWorkers, PriorityDatabase} {
		for i := range 48 {
			if priorities[(i*7)%len(priorities)] == p {
				want = append(want, fmt.Sprintf("hook-%02d", i))
			}
		}
	}

	require.NoError(t, m.Shutdown(context.Background()))
	assert.Equal(t, want, order, "hooks of equal priority must run in registration order")
}

func TestManager_HungHookDoesNotBlockLaterPhases(t *testing.T) {
	rec := newRecorder(t)
	m := NewManager(rec, 400*time.Millisecond, 0)

	release := make(chan struct{})
	defer close(release)
	m.RegisterHook("hung", PriorityHTTP+10, func(context.Context) error {
		<-release // ignores its context
		return nil
	})
	var ranDB atomic.Bool
	m.RegisterHook("db", PriorityDatabase, func(context.Context) error {
		ranDB.Store(true)
		return nil
	})

	start := time.Now()
	require.NoError(t, m.Shutdown(context.Background()))
	assert.Less(t, time.Since(start), 400*time.Millisecond, "shutdown must stop waiting for a hung hook")
	assert.True(t, ranDB.Load(), "hooks after a hung hook must still run")

	warns := rec.find("warn", "Shutdown hook ran out of time")
	require.Len(t, warns, 1)
	assert.Equal(t, "hung", warns[0].fields["name"])
}

func TestManager_BudgetIsThreeQuartersOfTheTimeout(t *testing.T) {
	m := NewManager(newRecorder(t), 4*time.Second, 0)

	var deadline time.Time
	var hasDeadline bool
	m.RegisterHook("probe", PriorityWorkers, func(ctx context.Context) error {
		deadline, hasDeadline = ctx.Deadline()
		return nil
	})

	start := time.Now()
	require.NoError(t, m.Shutdown(context.Background()))
	require.True(t, hasDeadline)
	assert.WithinDuration(t, start.Add(3*time.Second), deadline, 50*time.Millisecond,
		"hooks share three quarters of the shutdown timeout; the rest is for module teardown")
}

func TestManager_HookErrorIsReturned(t *testing.T) {
	errBoom := errors.New("boom")
	m := NewManager(newRecorder(t), time.Second, 0)
	m.RegisterHook("failing", PriorityCache, func(context.Context) error { return errBoom })

	assert.ErrorIs(t, m.Shutdown(context.Background()), errBoom)
}

func TestManager_TasksAreAwaitedBeforeLowerPriorityHooks(t *testing.T) {
	m := NewManager(newRecorder(t), 2*time.Second, 0)

	var taskDone atomic.Bool
	m.Go("index", func(context.Context) {
		time.Sleep(100 * time.Millisecond)
		taskDone.Store(true)
	})
	var sawDone atomic.Bool
	m.RegisterHook("db", PriorityDatabase, func(context.Context) error {
		sawDone.Store(taskDone.Load())
		return nil
	})

	require.NoError(t, m.Shutdown(context.Background()))
	assert.True(t, sawDone.Load(), "a hook below PriorityWorkers must run after tracked tasks finished")
}

func TestManager_TaskContextCancelledWhenSettleGivesUp(t *testing.T) {
	rec := newRecorder(t)
	m := NewManager(rec, 400*time.Millisecond, 0)

	cancelled := make(chan struct{})
	m.Go("t", func(ctx context.Context) {
		<-ctx.Done()
		close(cancelled)
	})

	require.NoError(t, m.Shutdown(context.Background()))
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("the task context was never cancelled")
	}

	warns := rec.find("warn", "Background tasks still running")
	require.Len(t, warns, 1)
	assert.Equal(t, []string{"t"}, warns[0].fields["tasks"])
}

func TestManager_TaskPanicIsRecoveredAndLogged(t *testing.T) {
	rec := newRecorder(t)
	m := NewManager(rec, time.Second, 0)

	m.Go("boom", func(context.Context) { panic("kaboom") })

	require.Eventually(t, func() bool {
		return len(rec.find("error", "Background task panicked")) == 1
	}, time.Second, 5*time.Millisecond)
	entry := rec.find("error", "Background task panicked")[0]
	assert.Equal(t, "boom", entry.fields["task"])
	assert.Equal(t, "kaboom", fmt.Sprint(entry.fields["panic"]))
	assert.Contains(t, entry.fields["stack"], "phases_test.go")

	require.NoError(t, m.Shutdown(context.Background()))
}

func TestManager_TaskStartedAfterSettleGetsCancelledContext(t *testing.T) {
	m := NewManager(newRecorder(t), time.Second, 0)
	require.NoError(t, m.Shutdown(context.Background()))

	got := make(chan error, 1)
	m.Go("late", func(ctx context.Context) { got <- ctx.Err() })

	select {
	case err := <-got:
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("a task started after shutdown never ran")
	}
}

func TestManager_ShutdownRunsOnce(t *testing.T) {
	m := NewManager(newRecorder(t), time.Second, 0)
	var runs atomic.Int32
	m.RegisterHook("once", PriorityCache, func(context.Context) error {
		runs.Add(1)
		return nil
	})

	require.NoError(t, m.Shutdown(context.Background()))
	require.NoError(t, m.Shutdown(context.Background()))
	assert.Equal(t, int32(1), runs.Load())
}

func TestNewManager_WarnsWhenTheDrainExceedsTheBudget(t *testing.T) {
	rec := newRecorder(t)
	_ = NewManager(rec, 4*time.Second, 5*time.Second)

	warns := rec.find("warn", "Drain timeout exceeds the shutdown budget")
	require.Len(t, warns, 1, "the drain can never get more than three quarters of the shutdown timeout")
	assert.Equal(t, "5s", warns[0].fields["drain_timeout"])
	assert.Equal(t, "3s", warns[0].fields["budget"])
}

func TestNewManager_NoWarningForTheDefaults(t *testing.T) {
	rec := newRecorder(t)
	_ = NewManager(rec, 0, 0)
	assert.Empty(t, rec.find("warn", "Drain timeout exceeds the shutdown budget"))
}

func TestManager_NoTaskWarningWhenNoTaskIsRunning(t *testing.T) {
	rec := newRecorder(t)
	m := NewManager(rec, 200*time.Millisecond, 0)
	release := make(chan struct{})
	defer close(release)
	m.RegisterHook("hung", PriorityHTTP+10, func(context.Context) error {
		<-release // eats the whole budget
		return nil
	})

	require.NoError(t, m.Shutdown(context.Background()))
	assert.Empty(t, rec.find("warn", "Background tasks still running"),
		"no task was started, so there is nothing to report")
}
