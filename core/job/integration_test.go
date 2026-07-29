//go:build integration

package job

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/internal/testutil"
	"go.uber.org/zap"
)

// Run with: make test_integration, or:
//   docker compose -f docker-compose.test.yml up -d
//   go test -tags=integration -v ./core/job/...
//
// The address defaults to the port docker-compose.test.yml publishes. Override it
// with GOE_TEST_REDIS_ADDR, the same variable the lock tests use:
//   GOE_TEST_REDIS_ADDR=localhost:6380 go test -tags=integration -v ./core/job/...

// testRedisAddr is the Redis instance these tests use.
//
// This previously defaulted to an ephemeral Docker-assigned port (32768) captured
// from one developer's session, so the suite failed everywhere else with a
// connection-refused storm. 127.0.0.1 rather than localhost keeps it off IPv6,
// where the published port is not bound.
func testRedisAddr() string {
	if env := os.Getenv("GOE_TEST_REDIS_ADDR"); env != "" {
		return env
	}
	return "127.0.0.1:6379"
}

func getTestConfig() *Config {
	cfg := DefaultConfig()
	cfg.RedisHosts = []string{testRedisAddr()}
	cfg.RedisDB = 15 // Use a separate DB for tests
	cfg.KeyPrefix = "goe:job:test:"
	cfg.Concurrency = 2
	cfg.PollInterval = 100 * time.Millisecond
	cfg.SchedulerInterval = 100 * time.Millisecond
	return cfg
}

// newTestManager creates a Manager and flushes stale test keys so tests are isolated.
func newTestManager(t *testing.T) *Manager {
	t.Helper()
	cfg := getTestConfig()
	logger := &testLogger{t: t}
	manager, err := NewManager(cfg, logger)
	require.NoErrorf(t, err,
		"Redis unreachable at %s. Start the test services with "+
			"`docker compose -f docker-compose.test.yml up -d`, or set GOE_TEST_REDIS_ADDR.",
		testRedisAddr())
	// Flush stale test keys for this DB
	manager.redis.FlushDB(context.Background())
	return manager
}

type testLogger struct {
	t *testing.T
}

func (l *testLogger) Debug(msg string, args ...any)                   { l.t.Logf("[DEBUG] "+msg, args...) }
func (l *testLogger) Info(msg string, args ...any)                    { l.t.Logf("[INFO] "+msg, args...) }
func (l *testLogger) Warn(msg string, args ...any)                    { l.t.Logf("[WARN] "+msg, args...) }
func (l *testLogger) Error(msg string, args ...any)                   { l.t.Logf("[ERROR] "+msg, args...) }
func (l *testLogger) Fatal(msg string, args ...any)                   { l.t.Fatalf("[FATAL] "+msg, args...) }
func (l *testLogger) Debugf(template string, args ...any)             {}
func (l *testLogger) Infof(template string, args ...any)              {}
func (l *testLogger) Warnf(template string, args ...any)              {}
func (l *testLogger) Errorf(template string, args ...any)             {}
func (l *testLogger) Fatalf(template string, args ...any)             { l.t.Fatalf(template, args...) }
func (l *testLogger) Panicf(template string, args ...any)             {}
func (l *testLogger) Debugw(msg string, keysAndValues ...any)         {}
func (l *testLogger) Infow(msg string, keysAndValues ...any)          {}
func (l *testLogger) Warnw(msg string, keysAndValues ...any)          {}
func (l *testLogger) Errorw(msg string, keysAndValues ...any)         {}
func (l *testLogger) Fatalw(msg string, keysAndValues ...any)         {}
func (l *testLogger) With(keysAndValues ...any) contract.Logger       { return l }
func (l *testLogger) WithContext(ctx context.Context) contract.Logger { return l }
func (l *testLogger) WithError(err error) contract.Logger             { return l }
func (l *testLogger) GetLogger() *zap.SugaredLogger                   { return nil }

func TestManagerIntegration_DispatchAndProcess(t *testing.T) {
	manager := newTestManager(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Track processed jobs
	var processedCount atomic.Int32
	var processedPayloads sync.Map

	// Register handler
	manager.RegisterHandler("test-job", contract.JobHandlerFunc(
		func(ctx context.Context, j contract.Job) error {
			processedCount.Add(1)
			processedPayloads.Store(j.ID(), j.Payload())
			return nil
		},
	))

	// Start manager
	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// Dispatch jobs
	job1ID, err := manager.Dispatch(ctx, &contract.JobDefinition{
		Name:    "test-job",
		Payload: map[string]string{"key": "value1"},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, job1ID)

	job2ID, err := manager.Dispatch(ctx, &contract.JobDefinition{
		Name:    "test-job",
		Payload: map[string]string{"key": "value2"},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, job2ID)

	// Poll instead of sleeping a fixed interval: worker wake-ups are bounded by
	// Redis's 1s minimum blocking-pop, so a fixed wait flakes under -race.
	testutil.AssertEventually(t, func() bool { return processedCount.Load() == 2 },
		30*time.Second, "both jobs should be processed")
	assert.Equal(t, int32(2), processedCount.Load())

	// Verify job states
	job1, err := manager.Get(ctx, job1ID)
	require.NoError(t, err)
	assert.Equal(t, contract.JobStatusCompleted, job1.Status())

	job2, err := manager.Get(ctx, job2ID)
	require.NoError(t, err)
	assert.Equal(t, contract.JobStatusCompleted, job2.Status())
}

func TestManagerIntegration_DelayedJob(t *testing.T) {
	manager := newTestManager(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var processedAt time.Time
	var processed atomic.Bool

	manager.RegisterHandler("delayed-job", contract.JobHandlerFunc(
		func(ctx context.Context, j contract.Job) error {
			processedAt = time.Now()
			processed.Store(true)
			return nil
		},
	))

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// Dispatch with 2 second delay
	dispatchedAt := time.Now()
	delay := 2 * time.Second

	_, err = manager.Dispatch(ctx, &contract.JobDefinition{
		Name:    "delayed-job",
		Payload: "test",
		Delay:   delay,
	})
	require.NoError(t, err)

	// Wait for the delayed job rather than guessing how long promotion takes.
	// The lower-bound delay assertion below is what actually tests the delay.
	testutil.AssertEventually(t, processed.Load, 30*time.Second,
		"delayed job should have been processed")

	require.True(t, processed.Load(), "Job should have been processed")
	actualDelay := processedAt.Sub(dispatchedAt)
	assert.GreaterOrEqual(t, actualDelay, delay-500*time.Millisecond, "Job should have been delayed")
	t.Logf("Job was delayed by %v (expected ~%v)", actualDelay, delay)
}

func TestManagerIntegration_Retry(t *testing.T) {
	manager := newTestManager(t)
	manager.config.RetryBackoff = 100 * time.Millisecond // Short backoff for testing

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var attempts atomic.Int32

	manager.RegisterHandler("retry-job", contract.JobHandlerFunc(
		func(ctx context.Context, j contract.Job) error {
			count := attempts.Add(1)
			if count < 3 {
				return assert.AnError // Fail first 2 attempts
			}
			return nil // Succeed on 3rd attempt
		},
	))

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	jobID, err := manager.Dispatch(ctx, &contract.JobDefinition{
		Name:        "retry-job",
		Payload:     "test",
		MaxAttempts: 5,
	})
	require.NoError(t, err)

	// Wait for the retry sequence to settle on the successful third attempt.
	testutil.AssertEventually(t, func() bool { return attempts.Load() == 3 },
		30*time.Second, "job should be attempted three times")
	assert.Equal(t, int32(3), attempts.Load(), "Job should have been attempted 3 times")

	// Verify final state is completed
	job, err := manager.Get(ctx, jobID)
	require.NoError(t, err)
	assert.Equal(t, contract.JobStatusCompleted, job.Status())
}

func TestManagerIntegration_UniqueJob(t *testing.T) {
	manager := newTestManager(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var processedCount atomic.Int32

	manager.RegisterHandler("unique-job", contract.JobHandlerFunc(
		func(ctx context.Context, j contract.Job) error {
			processedCount.Add(1)
			time.Sleep(500 * time.Millisecond) // Simulate work
			return nil
		},
	))

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// Dispatch first job with unique key
	_, err = manager.Dispatch(ctx, &contract.JobDefinition{
		Name:      "unique-job",
		Payload:   "test",
		UniqueKey: "test-unique-key",
		UniqueFor: 10 * time.Second,
	})
	require.NoError(t, err)

	// Try to dispatch duplicate - should fail
	_, err = manager.Dispatch(ctx, &contract.JobDefinition{
		Name:      "unique-job",
		Payload:   "test",
		UniqueKey: "test-unique-key",
		UniqueFor: 10 * time.Second,
	})
	assert.ErrorIs(t, err, ErrJobAlreadyExists)

	// Wait for the single job to be processed. A fixed wait would also have to
	// be long enough to prove no second job appears, which the unique-key check
	// in the assertion below covers.
	testutil.AssertEventually(t, func() bool { return processedCount.Load() == 1 },
		30*time.Second, "exactly one job should be processed")
	assert.Equal(t, int32(1), processedCount.Load())
}

func TestManagerIntegration_MultipleQueues(t *testing.T) {
	manager := newTestManager(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var highPriorityCount atomic.Int32
	var lowPriorityCount atomic.Int32

	manager.RegisterHandler("queue-test", contract.JobHandlerFunc(
		func(ctx context.Context, j contract.Job) error {
			if j.Queue() == "high" {
				highPriorityCount.Add(1)
			} else if j.Queue() == "low" {
				lowPriorityCount.Add(1)
			}
			return nil
		},
	))

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// Dispatch to different queues
	for i := 0; i < 3; i++ {
		_, err = manager.Dispatch(ctx, &contract.JobDefinition{
			Name:    "queue-test",
			Payload: i,
			Queue:   "high",
		})
		require.NoError(t, err)

		_, err = manager.Dispatch(ctx, &contract.JobDefinition{
			Name:    "queue-test",
			Payload: i,
			Queue:   "low",
		})
		require.NoError(t, err)
	}

	// Poll rather than sleep a fixed interval: worker wake-ups depend on Redis
	// blocking-pop granularity (1s minimum), so a fixed wait races with the
	// scheduler under -race and parallel load.
	testutil.AssertEventually(t, func() bool {
		return highPriorityCount.Load() == 3 && lowPriorityCount.Load() == 3
	}, 30*time.Second, "both queues should drain")

	assert.Equal(t, int32(3), highPriorityCount.Load())
	assert.Equal(t, int32(3), lowPriorityCount.Load())
}

func TestManagerIntegration_ScheduledJob(t *testing.T) {
	manager := newTestManager(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var executed atomic.Bool
	var executionTime time.Time
	var recordOnce sync.Once

	// Register scheduled job that runs every second
	err := manager.RegisterSchedule(&contract.ScheduledJob{
		Name:     "scheduled-test",
		Schedule: Every(1 * time.Second),
		Handler: contract.JobHandlerFunc(func(ctx context.Context, j contract.Job) error {
			// sync.Once rather than a check-then-set on executed: with
			// Concurrency > 1 two workers could both pass the check and race on
			// executionTime, which is not atomic.
			//
			// executionTime is written BEFORE publishing the flag. The original
			// order was the reverse, so a reader that saw executed == true could
			// still read a zero timestamp. A fixed 3s sleep hid that; polling
			// returns as soon as the flag flips and exposed it.
			recordOnce.Do(func() {
				executionTime = time.Now()
				executed.Store(true)
			})
			return nil
		}),
	})
	require.NoError(t, err)

	startTime := time.Now()

	err = manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// Wait for the scheduler to fire rather than sleeping a fixed interval.
	// The previous 3s sleep plus a `delay < 3s` assertion failed roughly four
	// runs in five under -race: the scheduler's wake-up is bounded by Redis's
	// 1s minimum blocking-pop, so the first execution lands anywhere in a
	// multi-second window under load.
	const scheduleTimeout = 30 * time.Second
	testutil.AssertEventually(t, executed.Load, scheduleTimeout,
		"scheduled job should have been executed")

	require.True(t, executed.Load(), "scheduled job never executed")

	// executionTime is safe to read here: the handler writes it before
	// executed.Store(true), and Go's atomics are sequentially consistent, so the
	// successful Load above happens-after that write.
	delay := executionTime.Sub(startTime)
	assert.Positive(t, delay, "execution should be after start")
	assert.Less(t, delay, scheduleTimeout,
		"scheduled job should execute well within the timeout")
	t.Logf("Scheduled job executed after %v", delay)
}

func TestManagerIntegration_Stats(t *testing.T) {
	manager := newTestManager(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	manager.RegisterHandler("stats-job", contract.JobHandlerFunc(
		func(ctx context.Context, j contract.Job) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		},
	))

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// Dispatch some jobs
	for i := 0; i < 5; i++ {
		_, err = manager.Dispatch(ctx, &contract.JobDefinition{
			Name:    "stats-job",
			Payload: i,
		})
		require.NoError(t, err)
	}

	// Give jobs time to start processing
	time.Sleep(500 * time.Millisecond)

	// Check stats
	stats, err := manager.Stats(ctx)
	require.NoError(t, err)

	t.Logf("Stats: %+v", stats)
	assert.NotNil(t, stats.Queues["default"])
	assert.Greater(t, stats.Workers.Total, 0)
}

func TestManagerIntegration_CancelJob(t *testing.T) {
	manager := newTestManager(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var processed atomic.Bool

	manager.RegisterHandler("cancel-job", contract.JobHandlerFunc(
		func(ctx context.Context, j contract.Job) error {
			processed.Store(true)
			return nil
		},
	))

	// Don't start the manager yet - dispatch then cancel

	// Dispatch with delay so we can cancel it
	jobID, err := manager.Dispatch(ctx, &contract.JobDefinition{
		Name:    "cancel-job",
		Payload: "test",
		Delay:   10 * time.Second,
	})
	require.NoError(t, err)

	// Cancel the job
	err = manager.Cancel(ctx, jobID)
	require.NoError(t, err)

	// Verify job is cancelled
	job, err := manager.Get(ctx, jobID)
	require.NoError(t, err)
	assert.Equal(t, contract.JobStatusCancelled, job.Status())

	// Start manager and wait
	err = manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	time.Sleep(2 * time.Second)

	// Job should not have been processed
	assert.False(t, processed.Load(), "Cancelled job should not be processed")
}

func TestManagerIntegration_Health(t *testing.T) {
	manager := newTestManager(t)

	ctx := context.Background()

	// Health check should pass
	err := manager.Health(ctx)
	assert.NoError(t, err)

	// Stop and verify cleanup
	err = manager.Stop(ctx)
	assert.NoError(t, err)
}

// TestManagerIntegration_ContextLifecycle verifies Issue #61 fix:
// background goroutines must survive after the startup context expires.
// Before the fix, runPromoter/runScheduler/workers would exit after 2 minutes
// because they used the Fx startup context which has a StartTimeout.
func TestManagerIntegration_ContextLifecycle(t *testing.T) {
	manager := newTestManager(t)

	var processed atomic.Bool

	manager.RegisterHandler("lifecycle-test", contract.JobHandlerFunc(
		func(ctx context.Context, j contract.Job) error {
			processed.Store(true)
			return nil
		},
	))

	// Simulate the Fx startup context: expires very quickly (500ms).
	// Before the fix, all goroutines would die when this context expires.
	startupCtx, startupCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer startupCancel()

	err := manager.Start(startupCtx)
	require.NoError(t, err)
	defer manager.Stop(context.Background())

	// Wait for the startup context to expire
	<-startupCtx.Done()
	time.Sleep(200 * time.Millisecond) // give goroutines a moment

	// Now dispatch a job AFTER the startup context has expired.
	// If the goroutines died with the context, this job will never be processed.
	bgCtx := context.Background()
	_, err = manager.Dispatch(bgCtx, &contract.JobDefinition{
		Name:    "lifecycle-test",
		Payload: "after-ctx-expired",
	})
	require.NoError(t, err)

	testutil.AssertEventually(t, processed.Load, 30*time.Second,
		"job dispatched after startup context expiry should still be processed")

	assert.True(t, processed.Load(),
		"Job dispatched after startup context expired must still be processed (Issue #61)")
}

// TestManagerIntegration_PromoterSurvivesContextExpiry specifically tests that
// the promoter (which moves delayed jobs to the ready queue) keeps working
// after the startup context expires.
func TestManagerIntegration_PromoterSurvivesContextExpiry(t *testing.T) {
	manager := newTestManager(t)

	var processed atomic.Bool

	manager.RegisterHandler("promoter-test", contract.JobHandlerFunc(
		func(ctx context.Context, j contract.Job) error {
			processed.Store(true)
			return nil
		},
	))

	// Start with a very short-lived context
	startupCtx, startupCancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer startupCancel()

	err := manager.Start(startupCtx)
	require.NoError(t, err)
	defer manager.Stop(context.Background())

	// Wait for startup context to expire
	<-startupCtx.Done()
	time.Sleep(200 * time.Millisecond)

	// Dispatch a delayed job AFTER the context expired.
	// The promoter must still be alive to move it from scheduled → ready.
	bgCtx := context.Background()
	_, err = manager.Dispatch(bgCtx, &contract.JobDefinition{
		Name:    "promoter-test",
		Payload: "delayed-after-ctx",
		Delay:   500 * time.Millisecond,
	})
	require.NoError(t, err)

	testutil.AssertEventually(t, processed.Load, 30*time.Second,
		"delayed job should be promoted after startup context expiry")

	assert.True(t, processed.Load(),
		"Delayed job must be promoted and processed even after startup context expired")
}

// TestManagerIntegration_SchedulerSurvivesContextExpiry tests that the scheduler
// (which creates recurring jobs) keeps working after the startup context expires.
func TestManagerIntegration_SchedulerSurvivesContextExpiry(t *testing.T) {
	manager := newTestManager(t)

	var executedCount atomic.Int32

	err := manager.RegisterSchedule(&contract.ScheduledJob{
		Name:     "scheduler-lifecycle",
		Schedule: Every(500 * time.Millisecond),
		Handler: contract.JobHandlerFunc(func(ctx context.Context, j contract.Job) error {
			executedCount.Add(1)
			return nil
		}),
		Overlap: true,
	})
	require.NoError(t, err)

	// Start with context that expires in 300ms
	startupCtx, startupCancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer startupCancel()

	err = manager.Start(startupCtx)
	require.NoError(t, err)
	defer manager.Stop(context.Background())

	// Wait for context to expire, then wait for scheduled jobs to accumulate
	<-startupCtx.Done()

	// The scheduler must keep firing; wait for more than one execution rather
	// than assuming a fixed window is long enough for two 1s ticks under load.
	testutil.AssertEventually(t, func() bool { return executedCount.Load() > 1 },
		30*time.Second, "scheduler should keep creating recurring jobs")

	count := executedCount.Load()
	assert.Greater(t, count, int32(1),
		"Scheduler must continue creating recurring jobs after startup context expires (got %d executions)", count)
}

// TestManagerIntegration_CancelScheduledJob verifies that cancelling a scheduled
// (delayed) job properly removes it from the scheduled sorted set in Redis.
// Before the fix, Cancel() checked job.status after already overwriting it to
// Cancelled, so the ZRem on the scheduled queue never executed.
func TestManagerIntegration_CancelScheduledJob(t *testing.T) {
	manager := newTestManager(t)

	ctx := context.Background()

	var processed atomic.Bool

	manager.RegisterHandler("cancel-scheduled", contract.JobHandlerFunc(
		func(ctx context.Context, j contract.Job) error {
			processed.Store(true)
			return nil
		},
	))

	// Dispatch with delay (goes to scheduled sorted set)
	jobID, err := manager.Dispatch(ctx, &contract.JobDefinition{
		Name:    "cancel-scheduled",
		Payload: "should-not-run",
		Delay:   2 * time.Second,
	})
	require.NoError(t, err)

	// Verify it's in the scheduled set
	scheduledKey := manager.scheduledKey("default")
	count, err := manager.redis.ZCard(ctx, scheduledKey).Result()
	require.NoError(t, err)
	assert.Greater(t, count, int64(0), "Job should be in the scheduled set")

	// Cancel the job
	err = manager.Cancel(ctx, jobID)
	require.NoError(t, err)

	// Verify it's been removed from the scheduled set
	score, err := manager.redis.ZScore(ctx, scheduledKey, jobID).Result()
	assert.Error(t, err, "Cancelled job should be removed from scheduled set")
	assert.Equal(t, float64(0), score)

	// Verify job status is cancelled
	job, err := manager.Get(ctx, jobID)
	require.NoError(t, err)
	assert.Equal(t, contract.JobStatusCancelled, job.Status())

	// Start the manager and wait — cancelled job should not run
	err = manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	time.Sleep(4 * time.Second)
	assert.False(t, processed.Load(), "Cancelled scheduled job must not be processed")
}

// NOTE: TestJobBuilder and TestScheduledJobBuilder moved to manager_test.go
// (pure unit tests, no Redis needed)
