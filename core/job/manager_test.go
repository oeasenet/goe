package job

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/zap"
)

// unitTestLogger is a test logger for unit tests (not requiring integration build tag).
type unitTestLogger struct {
	t *testing.T
}

func (l *unitTestLogger) Debug(msg string, args ...any)                   {}
func (l *unitTestLogger) Info(msg string, args ...any)                    {}
func (l *unitTestLogger) Warn(msg string, args ...any)                    {}
func (l *unitTestLogger) Error(msg string, args ...any)                   {}
func (l *unitTestLogger) Fatal(msg string, args ...any)                   { l.t.Fatalf(msg) }
func (l *unitTestLogger) Debugf(template string, args ...any)             {}
func (l *unitTestLogger) Infof(template string, args ...any)              {}
func (l *unitTestLogger) Warnf(template string, args ...any)              {}
func (l *unitTestLogger) Errorf(template string, args ...any)             {}
func (l *unitTestLogger) Fatalf(template string, args ...any)             { l.t.Fatalf(template, args...) }
func (l *unitTestLogger) Panicf(template string, args ...any)             {}
func (l *unitTestLogger) Debugw(msg string, keysAndValues ...any)         {}
func (l *unitTestLogger) Infow(msg string, keysAndValues ...any)          {}
func (l *unitTestLogger) Warnw(msg string, keysAndValues ...any)          {}
func (l *unitTestLogger) Errorw(msg string, keysAndValues ...any)         {}
func (l *unitTestLogger) Fatalw(msg string, keysAndValues ...any)         {}
func (l *unitTestLogger) With(keysAndValues ...any) contract.Logger       { return l }
func (l *unitTestLogger) WithContext(ctx context.Context) contract.Logger { return l }
func (l *unitTestLogger) WithError(err error) contract.Logger             { return l }
func (l *unitTestLogger) GetLogger() *zap.SugaredLogger                   { return nil }

// newUnitTestManager creates a Manager directly without a real Redis connection.
// It uses a client configured to fail fast, suitable for testing Start/Stop lifecycle.
func newUnitTestManager(t *testing.T, cfg *Config) *Manager {
	t.Helper()
	// Port 1 is reserved and will result in connection refused immediately.
	redisClient := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:1",
		DialTimeout:  20 * time.Millisecond,
		ReadTimeout:  20 * time.Millisecond,
		WriteTimeout: 20 * time.Millisecond,
		MaxRetries:   0,
	})
	t.Cleanup(func() { _ = redisClient.Close() })

	return &Manager{
		config:      cfg,
		logger:      &unitTestLogger{t: t},
		redis:       redisClient,
		stopCh:      make(chan struct{}),
		handlers:    make(map[string]contract.JobHandler),
		schedules:   make(map[string]*contract.ScheduledJob),
		totalQueues: []string{cfg.DefaultQueue},
		stats: &managerStats{
			queueStats: make(map[string]*queueStatsInternal),
		},
	}
}

// TestManagerStart_UseBackgroundContext verifies that Start() uses context.Background()
// for long-running goroutines so they are not cancelled when the startup context expires.
func TestManagerStart_UsesBackgroundContext(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Concurrency = 1
	cfg.PollInterval = 20 * time.Millisecond
	cfg.SchedulerEnabled = true
	cfg.SchedulerInterval = 20 * time.Millisecond
	cfg.ShutdownTimeout = 500 * time.Millisecond

	m := newUnitTestManager(t, cfg)

	// Simulate an Fx startup context that expires quickly.
	startCtx, cancelStart := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancelStart()

	err := m.Start(startCtx)
	require.NoError(t, err)
	assert.True(t, m.running.Load(), "manager should be running after Start")

	// Let the startup context expire to confirm goroutines keep running via stopCh.
	time.Sleep(100 * time.Millisecond)
	assert.True(t, m.running.Load(), "manager should still be running after startup context expired")

	// Stop via the dedicated stop mechanism.
	err = m.Stop(context.Background())
	require.NoError(t, err)
	assert.False(t, m.running.Load(), "manager should not be running after Stop")
}

// TestManagerStart_SchedulerDisabled verifies Start() with the scheduler disabled.
func TestManagerStart_SchedulerDisabled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Concurrency = 1
	cfg.PollInterval = 20 * time.Millisecond
	cfg.SchedulerEnabled = false
	cfg.SchedulerInterval = 20 * time.Millisecond
	cfg.ShutdownTimeout = 500 * time.Millisecond

	m := newUnitTestManager(t, cfg)

	err := m.Start(context.Background())
	require.NoError(t, err)
	assert.True(t, m.running.Load())

	// Give goroutines time to start.
	time.Sleep(50 * time.Millisecond)

	err = m.Stop(context.Background())
	require.NoError(t, err)
	assert.False(t, m.running.Load())
}

// TestManagerStart_AlreadyRunning verifies that calling Start() twice is a no-op.
func TestManagerStart_AlreadyRunning(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Concurrency = 1
	cfg.PollInterval = 20 * time.Millisecond
	cfg.SchedulerEnabled = false
	cfg.ShutdownTimeout = 500 * time.Millisecond

	m := newUnitTestManager(t, cfg)

	err := m.Start(context.Background())
	require.NoError(t, err)

	// Give goroutines time to start.
	time.Sleep(50 * time.Millisecond)

	// Second call should be a no-op.
	err = m.Start(context.Background())
	require.NoError(t, err)

	err = m.Stop(context.Background())
	require.NoError(t, err)
}
