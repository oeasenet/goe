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
	"go.uber.org/zap"
)

// Run with: go test -tags=integration -v ./core/job/...
//
// Override Redis address via environment variable:
//   TEST_REDIS_ADDR=localhost:32769 go test -tags=integration -v ./core/job/...

func getTestConfig() *Config {
	cfg := DefaultConfig()
	addr := "localhost:32768" // default: Docker-mapped Redis
	if env := os.Getenv("TEST_REDIS_ADDR"); env != "" {
		addr = env
	}
	cfg.RedisHosts = []string{addr}
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
	require.NoError(t, err)
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

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Verify jobs were processed
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

	// Wait for processing (delay + buffer)
	time.Sleep(4 * time.Second)

	// Verify job was processed after the delay
	assert.True(t, processed.Load(), "Job should have been processed")
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

	// Wait for retries
	time.Sleep(5 * time.Second)

	// Verify retry behavior
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

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Only one job should have been processed
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

	// Wait for processing
	time.Sleep(3 * time.Second)

	// Verify both queues were processed
	assert.Equal(t, int32(3), highPriorityCount.Load())
	assert.Equal(t, int32(3), lowPriorityCount.Load())
}

func TestManagerIntegration_ScheduledJob(t *testing.T) {
	manager := newTestManager(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var executed atomic.Bool
	var executionTime time.Time

	// Register scheduled job that runs every second
	err := manager.RegisterSchedule(&contract.ScheduledJob{
		Name:     "scheduled-test",
		Schedule: Every(1 * time.Second),
		Handler: contract.JobHandlerFunc(func(ctx context.Context, j contract.Job) error {
			if !executed.Load() {
				executed.Store(true)
				executionTime = time.Now()
			}
			return nil
		}),
	})
	require.NoError(t, err)

	startTime := time.Now()

	err = manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// Wait for scheduled execution
	time.Sleep(3 * time.Second)

	// Verify scheduled job was executed
	assert.True(t, executed.Load(), "Scheduled job should have been executed")
	delay := executionTime.Sub(startTime)
	assert.Less(t, delay, 3*time.Second, "Scheduled job should execute within 3 seconds")
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

	// Wait for processing
	time.Sleep(3 * time.Second)

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

	// Wait: delay + promoter interval + processing buffer
	time.Sleep(4 * time.Second)

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
	time.Sleep(4 * time.Second)

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
