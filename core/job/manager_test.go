package job

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/zap"
)

// nopLogger implements contract.Logger for unit tests that don't need output.
type nopLogger struct{}

func (l *nopLogger) Debug(msg string, args ...any)                   {}
func (l *nopLogger) Info(msg string, args ...any)                    {}
func (l *nopLogger) Warn(msg string, args ...any)                    {}
func (l *nopLogger) Error(msg string, args ...any)                   {}
func (l *nopLogger) Fatal(msg string, args ...any)                   {}
func (l *nopLogger) Debugf(template string, args ...any)             {}
func (l *nopLogger) Infof(template string, args ...any)              {}
func (l *nopLogger) Warnf(template string, args ...any)              {}
func (l *nopLogger) Errorf(template string, args ...any)             {}
func (l *nopLogger) Fatalf(template string, args ...any)             {}
func (l *nopLogger) Panicf(template string, args ...any)             {}
func (l *nopLogger) Debugw(msg string, keysAndValues ...any)         {}
func (l *nopLogger) Infow(msg string, keysAndValues ...any)          {}
func (l *nopLogger) Warnw(msg string, keysAndValues ...any)          {}
func (l *nopLogger) Errorw(msg string, keysAndValues ...any)         {}
func (l *nopLogger) Fatalw(msg string, keysAndValues ...any)         {}
func (l *nopLogger) With(keysAndValues ...any) contract.Logger       { return l }
func (l *nopLogger) WithContext(ctx context.Context) contract.Logger { return l }
func (l *nopLogger) WithError(err error) contract.Logger             { return l }
func (l *nopLogger) GetLogger() *zap.SugaredLogger                  { return nil }

func TestNewJob(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		j, err := newJob(&contract.JobDefinition{
			Name:    "test",
			Payload: "data",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, j.id)
		assert.Equal(t, "test", j.name)
		assert.Equal(t, "default", j.queue)
		assert.Equal(t, contract.JobStatusPending, j.status)
		assert.Equal(t, 3, j.maxAttempts)
		assert.Equal(t, 30*time.Minute, j.timeout)
	})

	t.Run("with delay sets scheduled status", func(t *testing.T) {
		before := time.Now()
		j, err := newJob(&contract.JobDefinition{
			Name:  "test",
			Delay: 5 * time.Minute,
		})
		require.NoError(t, err)
		assert.Equal(t, contract.JobStatusScheduled, j.status)
		assert.True(t, j.scheduledAt.After(before.Add(4*time.Minute)))
	})

	t.Run("with scheduled_at", func(t *testing.T) {
		target := time.Now().Add(time.Hour)
		j, err := newJob(&contract.JobDefinition{
			Name:        "test",
			ScheduledAt: target,
		})
		require.NoError(t, err)
		assert.Equal(t, contract.JobStatusScheduled, j.status)
		assert.Equal(t, target, j.scheduledAt)
	})

	t.Run("custom queue and attempts", func(t *testing.T) {
		j, err := newJob(&contract.JobDefinition{
			Name:        "test",
			Queue:       "high",
			MaxAttempts: 10,
			Timeout:     time.Hour,
		})
		require.NoError(t, err)
		assert.Equal(t, "high", j.queue)
		assert.Equal(t, 10, j.maxAttempts)
		assert.Equal(t, time.Hour, j.timeout)
	})
}

func TestJobSerialization(t *testing.T) {
	original, err := newJob(&contract.JobDefinition{
		Name:      "serialize-test",
		Payload:   map[string]string{"key": "value"},
		Queue:     "test-queue",
		UniqueKey: "unique-123",
		Tags:      map[string]string{"env": "test"},
	})
	require.NoError(t, err)

	// Serialize
	data, err := original.toJobData()
	require.NoError(t, err)
	bytes, err := data.marshal()
	require.NoError(t, err)

	// Deserialize
	parsed, err := unmarshalJobData(bytes)
	require.NoError(t, err)
	restored := fromJobData(parsed)

	assert.Equal(t, original.id, restored.id)
	assert.Equal(t, original.name, restored.name)
	assert.Equal(t, original.queue, restored.queue)
	assert.Equal(t, original.status, restored.status)
	assert.Equal(t, original.maxAttempts, restored.maxAttempts)
	assert.Equal(t, original.uniqueKey, restored.uniqueKey)
	assert.Equal(t, original.tags, restored.tags)
	assert.Equal(t, original.timeout, restored.timeout)
}

func TestWorkerBackoff(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RetryBackoff = 1 * time.Second
	cfg.RetryBackoffFactor = 2.0
	cfg.MaxRetryBackoff = 30 * time.Second

	w := &worker{manager: &Manager{config: cfg}}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 1 * time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{4, 8 * time.Second},
		{5, 16 * time.Second},
		{6, 30 * time.Second}, // capped at max
		{7, 30 * time.Second},
	}

	for _, tt := range tests {
		backoff := w.calculateBackoff(tt.attempt)
		assert.Equal(t, tt.expected, backoff, "attempt %d", tt.attempt)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 5, cfg.Concurrency)
	assert.Equal(t, "default", cfg.DefaultQueue)
	assert.Equal(t, 3, cfg.DefaultMaxAttempts)
	assert.Equal(t, 30*time.Minute, cfg.DefaultTimeout)
	assert.Equal(t, time.Second, cfg.RetryBackoff)
	assert.Equal(t, 2.0, cfg.RetryBackoffFactor)
	assert.Equal(t, 5*time.Minute, cfg.MaxRetryBackoff)
	assert.True(t, cfg.SchedulerEnabled)
	assert.True(t, cfg.DLQEnabled)
	assert.Equal(t, 7*24*time.Hour, cfg.DLQTTL)
}

func TestManagerStartIdempotent(t *testing.T) {
	// Verify calling Start twice returns nil (idempotent guard)
	m := &Manager{
		config:      DefaultConfig(),
		totalQueues: []string{"default"},
		stopCh:      make(chan struct{}),
		stats:       &managerStats{queueStats: make(map[string]*queueStatsInternal)},
	}
	m.running.Store(true)

	err := m.Start(context.Background())
	assert.NoError(t, err) // should be no-op
}

func TestRedisKeyHelpers(t *testing.T) {
	m := &Manager{config: &Config{KeyPrefix: "goe:job:"}}

	assert.Equal(t, "goe:job:data:abc-123", m.jobKey("abc-123"))
	assert.Equal(t, "goe:job:queue:default", m.queueKey("default"))
	assert.Equal(t, "goe:job:scheduled:high", m.scheduledKey("high"))
	assert.Equal(t, "goe:job:unique:default:my-key", m.uniqueKey("default", "my-key"))
	assert.Equal(t, "goe:job:processing:default", m.processingKey("default"))
	assert.Equal(t, "goe:job:dlq:default", m.dlqKey("default"))
}

func TestTrackQueueDeduplication(t *testing.T) {
	m := &Manager{
		config:      DefaultConfig(),
		totalQueues: []string{"default"},
		stats:       &managerStats{queueStats: make(map[string]*queueStatsInternal)},
	}

	// Track same queue twice — should not duplicate
	m.trackQueue("default")
	m.trackQueue("default")
	assert.Len(t, m.totalQueues, 1)

	// Track new queue — should add
	m.trackQueue("high")
	assert.Len(t, m.totalQueues, 2)
	assert.Contains(t, m.totalQueues, "high")
}

func TestGetHandlerFallbackToSchedule(t *testing.T) {
	handler := contract.JobHandlerFunc(func(ctx context.Context, j contract.Job) error {
		return nil
	})

	m := &Manager{
		logger:    &nopLogger{},
		handlers:  make(map[string]contract.JobHandler),
		schedules: make(map[string]*contract.ScheduledJob),
	}

	// No handler registered — should return false
	_, ok := m.getHandler("unknown")
	assert.False(t, ok)

	// Register via schedule
	m.schedules["scheduled-job"] = &contract.ScheduledJob{
		Name:    "scheduled-job",
		Handler: handler,
	}
	h, ok := m.getHandler("scheduled-job")
	assert.True(t, ok)
	assert.NotNil(t, h)

	// Direct handler takes precedence over schedule handler
	m.handlers["scheduled-job"] = handler
	h2, ok := m.getHandler("scheduled-job")
	assert.True(t, ok)
	assert.NotNil(t, h2)
}

func TestRecordCompletion(t *testing.T) {
	m := &Manager{
		stats: &managerStats{queueStats: make(map[string]*queueStatsInternal)},
	}

	m.recordCompletion("default", true)
	m.recordCompletion("default", true)
	m.recordCompletion("default", false)

	assert.Equal(t, int64(3), m.stats.processed.Load())
	assert.Equal(t, int64(2), m.stats.completed.Load())
	assert.Equal(t, int64(1), m.stats.failed.Load())

	// Verify per-queue stats were auto-created
	qs, ok := m.stats.queueStats["default"]
	assert.True(t, ok)
	assert.Equal(t, int64(2), qs.completed.Load())
	assert.Equal(t, int64(1), qs.failed.Load())
}

func TestRegisterScheduleValidation(t *testing.T) {
	cfg := DefaultConfig()
	m := &Manager{
		config:    cfg,
		logger:    &nopLogger{},
		schedules: make(map[string]*contract.ScheduledJob),
	}

	handler := contract.JobHandlerFunc(func(ctx context.Context, j contract.Job) error {
		return nil
	})

	// First registration succeeds
	err := m.RegisterSchedule(&contract.ScheduledJob{
		Name:     "job1",
		Schedule: Every(time.Minute),
		Handler:  handler,
	})
	assert.NoError(t, err)

	// Duplicate registration fails
	err = m.RegisterSchedule(&contract.ScheduledJob{
		Name:     "job1",
		Schedule: Every(time.Minute),
		Handler:  handler,
	})
	assert.ErrorIs(t, err, ErrScheduleExists)

	// Defaults applied
	sched := m.schedules["job1"]
	assert.Equal(t, cfg.DefaultQueue, sched.Queue)
	assert.Equal(t, cfg.DefaultTimeout, sched.Timeout)
	assert.Equal(t, cfg.DefaultMaxAttempts, sched.MaxAttempts)
}

func TestUnregisterSchedule(t *testing.T) {
	m := &Manager{
		config:    DefaultConfig(),
		logger:    &nopLogger{},
		schedules: make(map[string]*contract.ScheduledJob),
	}

	handler := contract.JobHandlerFunc(func(ctx context.Context, j contract.Job) error {
		return nil
	})

	_ = m.RegisterSchedule(&contract.ScheduledJob{
		Name:     "job1",
		Schedule: Every(time.Minute),
		Handler:  handler,
	})

	err := m.UnregisterSchedule("job1")
	assert.NoError(t, err)

	err = m.UnregisterSchedule("nonexistent")
	assert.ErrorIs(t, err, ErrScheduleNotFound)
}

func TestJobImplAccessors(t *testing.T) {
	now := time.Now()
	j := &jobImpl{
		id:          "test-id",
		name:        "test-name",
		queue:       "test-queue",
		payload:     "test-payload",
		status:      contract.JobStatusRunning,
		attempts:    2,
		maxAttempts: 5,
		createdAt:   now,
		scheduledAt: now,
		startedAt:   now,
		completedAt: now,
		lastError:   "some error",
		ctx:         context.Background(),
		tags:        map[string]string{"env": "test"},
		timeout:     time.Minute,
		uniqueKey:   "unique-key",
	}

	assert.Equal(t, "test-id", j.ID())
	assert.Equal(t, "test-name", j.Name())
	assert.Equal(t, "test-queue", j.Queue())
	assert.Equal(t, "test-payload", j.Payload())
	assert.Equal(t, contract.JobStatusRunning, j.Status())
	assert.Equal(t, 2, j.Attempts())
	assert.Equal(t, 5, j.MaxAttempts())
	assert.Equal(t, now, j.CreatedAt())
	assert.Equal(t, now, j.ScheduledAt())
	assert.Equal(t, now, j.StartedAt())
	assert.Equal(t, now, j.CompletedAt())
	assert.Equal(t, "some error", j.LastError())
	assert.NotNil(t, j.Context())
	assert.Equal(t, map[string]string{"env": "test"}, j.Tags())
	assert.Equal(t, time.Minute, j.Timeout())
	assert.Equal(t, "unique-key", j.UniqueKey())
}

func TestJobBuilderFull(t *testing.T) {
	def := NewJob("send-email", "payload").
		OnQueue("emails").
		WithDelay(5 * time.Minute).
		WithRetries(5).
		WithTimeout(10 * time.Minute).
		Unique("email-123").
		UniqueFor("email-123", 2*time.Hour).
		WithTag("env", "prod").
		WithTags(map[string]string{"env": "prod", "priority": "high"}).
		Build()

	assert.Equal(t, "send-email", def.Name)
	assert.Equal(t, "emails", def.Queue)
	assert.Equal(t, 5*time.Minute, def.Delay)
	assert.Equal(t, 5, def.MaxAttempts)
	assert.Equal(t, 10*time.Minute, def.Timeout)
	assert.Equal(t, "email-123", def.UniqueKey)
	assert.Equal(t, 2*time.Hour, def.UniqueFor)
	assert.Equal(t, "prod", def.Tags["env"])
	assert.Equal(t, "high", def.Tags["priority"])
}

func TestJobBuilderTimeHelpers(t *testing.T) {
	t.Run("In alias", func(t *testing.T) {
		def := NewJob("test", nil).In(3 * time.Minute).Build()
		assert.Equal(t, 3*time.Minute, def.Delay)
	})

	t.Run("InMinutes", func(t *testing.T) {
		def := NewJob("test", nil).InMinutes(10).Build()
		assert.Equal(t, 10*time.Minute, def.Delay)
	})

	t.Run("InHours", func(t *testing.T) {
		def := NewJob("test", nil).InHours(2).Build()
		assert.Equal(t, 2*time.Hour, def.Delay)
	})

	t.Run("Tomorrow", func(t *testing.T) {
		def := NewJob("test", nil).Tomorrow().Build()
		assert.Equal(t, 24*time.Hour, def.Delay)
	})

	t.Run("At", func(t *testing.T) {
		target := time.Now().Add(time.Hour)
		def := NewJob("test", nil).At(target).Build()
		assert.Equal(t, target, def.ScheduledAt)
	})
}

func TestDispatchHelpers(t *testing.T) {
	t.Run("Dispatch", func(t *testing.T) {
		def := Dispatch("test", "data").Build()
		assert.Equal(t, "test", def.Name)
		assert.Equal(t, "data", def.Payload)
	})

	t.Run("Later", func(t *testing.T) {
		def := Later("test", "data", 5*time.Minute).Build()
		assert.Equal(t, "test", def.Name)
		assert.Equal(t, 5*time.Minute, def.Delay)
	})

	t.Run("AtTime", func(t *testing.T) {
		target := time.Now().Add(time.Hour)
		def := AtTime("test", "data", target).Build()
		assert.Equal(t, target, def.ScheduledAt)
	})
}

func TestScheduledJobBuilder(t *testing.T) {
	handler := contract.JobHandlerFunc(func(ctx context.Context, j contract.Job) error {
		return nil
	})

	t.Run("full builder", func(t *testing.T) {
		sched := NewScheduledJob("cleanup").
			Runs(DailyAt(3, 0)).
			Handle(handler).
			OnQueue("maintenance").
			WithPayload("test-payload").
			WithTimeout(1 * time.Hour).
			WithRetries(2).
			AllowOverlap().
			WithTag("env", "prod").
			Build()

		assert.Equal(t, "cleanup", sched.Name)
		assert.Equal(t, "maintenance", sched.Queue)
		assert.Equal(t, "test-payload", sched.Payload)
		assert.Equal(t, time.Hour, sched.Timeout)
		assert.Equal(t, 2, sched.MaxAttempts)
		assert.True(t, sched.Overlap)
		assert.Equal(t, "prod", sched.Tags["env"])
	})

	t.Run("Using alias", func(t *testing.T) {
		sched := NewScheduledJob("test").
			Using(Every(time.Minute)).
			HandleFunc(handler).
			Build()
		assert.NotNil(t, sched.Schedule)
		assert.NotNil(t, sched.Handler)
	})

	t.Run("PreventOverlap", func(t *testing.T) {
		sched := NewScheduledJob("test").
			AllowOverlap().
			PreventOverlap().
			Build()
		assert.False(t, sched.Overlap)
	})
}

func TestStopIdempotent(t *testing.T) {
	m := &Manager{
		config: DefaultConfig(),
		stopCh: make(chan struct{}),
	}
	// Not running — Stop should be no-op
	err := m.Stop(context.Background())
	assert.NoError(t, err)
}
