//go:build integration
// +build integration

package job

import (
	"context"
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

func getTestConfig() *Config {
	cfg := DefaultConfig()
	cfg.RedisHosts = []string{"localhost:32768"} // Docker mapped port
	cfg.RedisDB = 15                             // Use a separate DB for tests
	cfg.KeyPrefix = "goe:job:test:"
	cfg.Concurrency = 2
	cfg.PollInterval = 100 * time.Millisecond
	cfg.SchedulerInterval = 100 * time.Millisecond
	return cfg
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
	cfg := getTestConfig()
	logger := &testLogger{t: t}

	manager, err := NewManager(cfg, logger)
	require.NoError(t, err)

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
	err = manager.Start(ctx)
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
	cfg := getTestConfig()
	logger := &testLogger{t: t}

	manager, err := NewManager(cfg, logger)
	require.NoError(t, err)

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

	err = manager.Start(ctx)
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
	cfg := getTestConfig()
	cfg.RetryBackoff = 100 * time.Millisecond // Short backoff for testing
	logger := &testLogger{t: t}

	manager, err := NewManager(cfg, logger)
	require.NoError(t, err)

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

	err = manager.Start(ctx)
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
	cfg := getTestConfig()
	logger := &testLogger{t: t}

	manager, err := NewManager(cfg, logger)
	require.NoError(t, err)

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

	err = manager.Start(ctx)
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
	cfg := getTestConfig()
	logger := &testLogger{t: t}

	manager, err := NewManager(cfg, logger)
	require.NoError(t, err)

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

	err = manager.Start(ctx)
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
	cfg := getTestConfig()
	logger := &testLogger{t: t}

	manager, err := NewManager(cfg, logger)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var executed atomic.Bool
	var executionTime time.Time

	// Register scheduled job that runs every second
	err = manager.RegisterSchedule(&contract.ScheduledJob{
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
	cfg := getTestConfig()
	logger := &testLogger{t: t}

	manager, err := NewManager(cfg, logger)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	manager.RegisterHandler("stats-job", contract.JobHandlerFunc(
		func(ctx context.Context, j contract.Job) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		},
	))

	err = manager.Start(ctx)
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
	cfg := getTestConfig()
	logger := &testLogger{t: t}

	manager, err := NewManager(cfg, logger)
	require.NoError(t, err)

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
	cfg := getTestConfig()
	logger := &testLogger{t: t}

	manager, err := NewManager(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Health check should pass
	err = manager.Health(ctx)
	assert.NoError(t, err)

	// Stop and verify cleanup
	err = manager.Stop(ctx)
	assert.NoError(t, err)
}

func TestJobBuilder(t *testing.T) {
	// Test fluent builder
	job := NewJob("test-job", map[string]string{"key": "value"}).
		OnQueue("high").
		WithDelay(5*time.Minute).
		WithRetries(5).
		WithTimeout(10*time.Minute).
		Unique("unique-key").
		WithTag("env", "test").
		Build()

	assert.Equal(t, "test-job", job.Name)
	assert.Equal(t, "high", job.Queue)
	assert.Equal(t, 5*time.Minute, job.Delay)
	assert.Equal(t, 5, job.MaxAttempts)
	assert.Equal(t, 10*time.Minute, job.Timeout)
	assert.Equal(t, "unique-key", job.UniqueKey)
	assert.Equal(t, "test", job.Tags["env"])
}

func TestScheduledJobBuilder(t *testing.T) {
	handler := contract.JobHandlerFunc(func(ctx context.Context, j contract.Job) error {
		return nil
	})

	sched := NewScheduledJob("cleanup").
		Runs(DailyAt(3, 0)).
		Handle(handler).
		OnQueue("maintenance").
		WithTimeout(1 * time.Hour).
		WithRetries(2).
		AllowOverlap().
		Build()

	assert.Equal(t, "cleanup", sched.Name)
	assert.Equal(t, "maintenance", sched.Queue)
	assert.Equal(t, 1*time.Hour, sched.Timeout)
	assert.Equal(t, 2, sched.MaxAttempts)
	assert.True(t, sched.Overlap)
}
