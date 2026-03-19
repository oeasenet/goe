package shutdown

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/zap"
)

// testLogger implements contract.Logger for testing
type testLogger struct {
	t *testing.T
}

func (l *testLogger) Debug(msg string, args ...any) {
	if l.t != nil {
		l.t.Logf("[DEBUG] %s %v", msg, args)
	}
}
func (l *testLogger) Info(msg string, args ...any) {
	if l.t != nil {
		l.t.Logf("[INFO] %s %v", msg, args)
	}
}
func (l *testLogger) Warn(msg string, args ...any) {
	if l.t != nil {
		l.t.Logf("[WARN] %s %v", msg, args)
	}
}
func (l *testLogger) Error(msg string, args ...any) {
	if l.t != nil {
		l.t.Logf("[ERROR] %s %v", msg, args)
	}
}
func (l *testLogger) Fatal(msg string, args ...any) {
	if l.t != nil {
		l.t.Fatalf("[FATAL] %s %v", msg, args)
	}
}
func (l *testLogger) Debugf(template string, args ...any) {
	if l.t != nil {
		l.t.Logf("[DEBUG] "+template, args...)
	}
}
func (l *testLogger) Infof(template string, args ...any) {
	if l.t != nil {
		l.t.Logf("[INFO] "+template, args...)
	}
}
func (l *testLogger) Warnf(template string, args ...any) {
	if l.t != nil {
		l.t.Logf("[WARN] "+template, args...)
	}
}
func (l *testLogger) Errorf(template string, args ...any) {
	if l.t != nil {
		l.t.Logf("[ERROR] "+template, args...)
	}
}
func (l *testLogger) Fatalf(template string, args ...any) {
	if l.t != nil {
		l.t.Fatalf("[FATAL] "+template, args...)
	}
}
func (l *testLogger) Debugw(msg string, keysAndValues ...any) {
	if l.t != nil {
		l.t.Logf("[DEBUG] %s %v", msg, keysAndValues)
	}
}
func (l *testLogger) Infow(msg string, keysAndValues ...any) {
	if l.t != nil {
		l.t.Logf("[INFO] %s %v", msg, keysAndValues)
	}
}
func (l *testLogger) Warnw(msg string, keysAndValues ...any) {
	if l.t != nil {
		l.t.Logf("[WARN] %s %v", msg, keysAndValues)
	}
}
func (l *testLogger) Errorw(msg string, keysAndValues ...any) {
	if l.t != nil {
		l.t.Logf("[ERROR] %s %v", msg, keysAndValues)
	}
}
func (l *testLogger) Fatalw(msg string, keysAndValues ...any) {
	if l.t != nil {
		l.t.Fatalf("[FATAL] %s %v", msg, keysAndValues)
	}
}
func (l *testLogger) With(keysAndValues ...any) contract.Logger {
	return l
}
func (l *testLogger) WithContext(ctx context.Context) contract.Logger {
	return l
}
func (l *testLogger) WithError(err error) contract.Logger {
	return l
}
func (l *testLogger) GetLogger() *zap.SugaredLogger {
	return nil
}

func TestNewManager(t *testing.T) {
	logger := &testLogger{t: t}

	t.Run("default timeouts", func(t *testing.T) {
		manager := NewManager(logger, 0, 0)
		assert.NotNil(t, manager)
		assert.Equal(t, 30*time.Second, manager.Timeout())
		assert.Equal(t, 5*time.Second, manager.DrainTimeout())
	})

	t.Run("custom timeouts", func(t *testing.T) {
		manager := NewManager(logger, 60*time.Second, 10*time.Second)
		assert.NotNil(t, manager)
		assert.Equal(t, 60*time.Second, manager.Timeout())
		assert.Equal(t, 10*time.Second, manager.DrainTimeout())
	})
}

func TestManager_RegisterHook(t *testing.T) {
	logger := &testLogger{t: t}
	manager := NewManager(logger, 30*time.Second, 5*time.Second)

	executed := false
	manager.RegisterHook("test-hook", 50, func(ctx context.Context) error {
		executed = true
		return nil
	})

	// Hook should be registered
	manager.mu.RLock()
	assert.Len(t, manager.hooks, 1)
	assert.Equal(t, "test-hook", manager.hooks[0].name)
	assert.Equal(t, 50, manager.hooks[0].priority)
	manager.mu.RUnlock()

	// Execute shutdown
	ctx := context.Background()
	err := manager.Shutdown(ctx)
	require.NoError(t, err)
	assert.True(t, executed)
}

func TestManager_UnregisterHook(t *testing.T) {
	logger := &testLogger{t: t}
	manager := NewManager(logger, 30*time.Second, 5*time.Second)

	manager.RegisterHook("hook1", 50, func(ctx context.Context) error { return nil })
	manager.RegisterHook("hook2", 40, func(ctx context.Context) error { return nil })

	manager.mu.RLock()
	assert.Len(t, manager.hooks, 2)
	manager.mu.RUnlock()

	manager.UnregisterHook("hook1")

	manager.mu.RLock()
	assert.Len(t, manager.hooks, 1)
	assert.Equal(t, "hook2", manager.hooks[0].name)
	manager.mu.RUnlock()
}

func TestManager_HookPriorityOrder(t *testing.T) {
	logger := &testLogger{t: t}
	manager := NewManager(logger, 30*time.Second, 5*time.Second)

	var executionOrder []string

	// Register hooks in random priority order
	manager.RegisterHook("cache", PriorityCache, func(ctx context.Context) error {
		executionOrder = append(executionOrder, "cache")
		return nil
	})
	manager.RegisterHook("http", PriorityHTTP, func(ctx context.Context) error {
		executionOrder = append(executionOrder, "http")
		return nil
	})
	manager.RegisterHook("database", PriorityDatabase, func(ctx context.Context) error {
		executionOrder = append(executionOrder, "database")
		return nil
	})
	manager.RegisterHook("telemetry", PriorityTelemetry, func(ctx context.Context) error {
		executionOrder = append(executionOrder, "telemetry")
		return nil
	})

	ctx := context.Background()
	err := manager.Shutdown(ctx)
	require.NoError(t, err)

	// Hooks should be executed in priority order (highest first)
	expected := []string{"http", "database", "cache", "telemetry"}
	assert.Equal(t, expected, executionOrder)
}

func TestManager_ShutdownWithError(t *testing.T) {
	logger := &testLogger{t: t}
	manager := NewManager(logger, 30*time.Second, 5*time.Second)

	expectedErr := errors.New("hook failed")

	manager.RegisterHook("failing-hook", 50, func(ctx context.Context) error {
		return expectedErr
	})

	ctx := context.Background()
	err := manager.Shutdown(ctx)
	assert.ErrorIs(t, err, expectedErr)
}

func TestManager_ShutdownContinuesAfterError(t *testing.T) {
	logger := &testLogger{t: t}
	manager := NewManager(logger, 30*time.Second, 5*time.Second)

	var executed int32

	manager.RegisterHook("hook1", PriorityHTTP, func(ctx context.Context) error {
		atomic.AddInt32(&executed, 1)
		return errors.New("error 1")
	})
	manager.RegisterHook("hook2", PriorityDatabase, func(ctx context.Context) error {
		atomic.AddInt32(&executed, 1)
		return nil
	})

	ctx := context.Background()
	err := manager.Shutdown(ctx)

	// Should return the first error
	assert.Error(t, err)
	// But both hooks should have executed
	assert.Equal(t, int32(2), atomic.LoadInt32(&executed))
}

func TestManager_ShutdownOnlyOnce(t *testing.T) {
	logger := &testLogger{t: t}
	manager := NewManager(logger, 30*time.Second, 5*time.Second)

	var count int32

	manager.RegisterHook("counter", 50, func(ctx context.Context) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	ctx := context.Background()

	// Call shutdown multiple times
	err1 := manager.Shutdown(ctx)
	err2 := manager.Shutdown(ctx)
	err3 := manager.Shutdown(ctx)

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)

	// Hook should only execute once
	assert.Equal(t, int32(1), atomic.LoadInt32(&count))
}

func TestManager_IsShutdown(t *testing.T) {
	logger := &testLogger{t: t}
	manager := NewManager(logger, 30*time.Second, 5*time.Second)

	assert.False(t, manager.IsShutdown())

	ctx := context.Background()
	_ = manager.Shutdown(ctx)

	assert.True(t, manager.IsShutdown())
}

func TestManager_ShutdownTimeout(t *testing.T) {
	logger := &testLogger{t: t}
	manager := NewManager(logger, 100*time.Millisecond, 50*time.Millisecond)

	manager.RegisterHook("slow-hook", 50, func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			return nil
		}
	})

	ctx := context.Background()
	start := time.Now()
	err := manager.Shutdown(ctx)
	elapsed := time.Since(start)

	// Should timeout and return context deadline exceeded
	assert.Error(t, err)
	assert.True(t, elapsed < 500*time.Millisecond, "Should timeout quickly")
}

func TestPriorityConstants(t *testing.T) {
	// Verify priority ordering (higher = runs first)
	assert.Greater(t, PriorityHTTP, PriorityWorkers)
	assert.Greater(t, PriorityWorkers, PriorityEvent)
	assert.Greater(t, PriorityEvent, PriorityDatabase)
	assert.Greater(t, PriorityDatabase, PriorityCache)
	assert.Greater(t, PriorityCache, PriorityTelemetry)
}
