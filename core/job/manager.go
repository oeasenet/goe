package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"go.oease.dev/goe/v2/contract"
)

var (
	ErrJobNotFound       = errors.New("job not found")
	ErrHandlerNotFound   = errors.New("handler not found for job")
	ErrJobAlreadyExists  = errors.New("job with this unique key already exists")
	ErrManagerNotRunning = errors.New("job manager is not running")
	ErrScheduleExists    = errors.New("scheduled job with this name already exists")
	ErrScheduleNotFound  = errors.New("scheduled job not found")
)

// promoteScript atomically moves due jobs from the scheduled sorted set to the
// ready list. Because ZRangeByScore + ZRem + LPush must be atomic to prevent
// duplicate processing across multiple app instances, this is done in a Lua
// script executed inside Redis.
//
// KEYS[1] = scheduled sorted-set key
// KEYS[2] = ready list key
// ARGV[1] = current time in milliseconds (max score)
//
// Returns the list of promoted job IDs.
var promoteScript = redis.NewScript(`
local jobs = redis.call('ZRANGEBYSCORE', KEYS[1], '-inf', ARGV[1], 'LIMIT', 0, 100)
local promoted = {}
for _, job_id in ipairs(jobs) do
    if redis.call('ZREM', KEYS[1], job_id) == 1 then
        redis.call('LPUSH', KEYS[2], job_id)
        table.insert(promoted, job_id)
    end
end
return promoted
`)

// Manager implements the contract.JobManager interface
type Manager struct {
	config  *Config
	logger  contract.Logger
	redis   redis.UniversalClient
	running atomic.Bool
	stopCh  chan struct{}
	wg      sync.WaitGroup

	// Handler registry
	handlers   map[string]contract.JobHandler
	handlersMu sync.RWMutex

	// Scheduled jobs
	schedules   map[string]*contract.ScheduledJob
	schedulesMu sync.RWMutex

	// Worker management
	workers     []*worker
	workersMu   sync.Mutex
	activeJobs  atomic.Int64
	totalQueues []string

	// Statistics
	stats *managerStats
}

// managerStats tracks job statistics
type managerStats struct {
	processed  atomic.Int64
	failed     atomic.Int64
	completed  atomic.Int64
	retried    atomic.Int64
	mu         sync.RWMutex
	queueStats map[string]*queueStatsInternal
}

type queueStatsInternal struct {
	running   atomic.Int64
	completed atomic.Int64
	failed    atomic.Int64
}

// NewManager creates a new job manager
func NewManager(config *Config, logger contract.Logger) (*Manager, error) {
	// Create Redis client
	var redisClient redis.UniversalClient

	if config.RedisURL != "" {
		opt, err := redis.ParseURL(config.RedisURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
		}
		opt.PoolSize = config.RedisPoolSize
		redisClient = redis.NewClient(opt)
	} else {
		redisClient = redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:    config.RedisHosts,
			Username: config.RedisUsername,
			Password: config.RedisPassword,
			DB:       config.RedisDB,
			PoolSize: config.RedisPoolSize,
		})
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	m := &Manager{
		config:      config,
		logger:      logger,
		redis:       redisClient,
		stopCh:      make(chan struct{}),
		handlers:    make(map[string]contract.JobHandler),
		schedules:   make(map[string]*contract.ScheduledJob),
		totalQueues: []string{config.DefaultQueue},
		stats: &managerStats{
			queueStats: make(map[string]*queueStatsInternal),
		},
	}

	return m, nil
}

// RegisterHandler registers a handler for a job name
func (m *Manager) RegisterHandler(name string, handler contract.JobHandler) {
	m.handlersMu.Lock()
	defer m.handlersMu.Unlock()
	m.handlers[name] = handler
	m.logger.Debug("Registered job handler", "name", name)
}

// RegisterSchedule registers a recurring scheduled job
func (m *Manager) RegisterSchedule(job *contract.ScheduledJob) error {
	m.schedulesMu.Lock()
	defer m.schedulesMu.Unlock()

	if _, exists := m.schedules[job.Name]; exists {
		return ErrScheduleExists
	}

	// Set defaults
	if job.Queue == "" {
		job.Queue = m.config.DefaultQueue
	}
	if job.Timeout <= 0 {
		job.Timeout = m.config.DefaultTimeout
	}
	if job.MaxAttempts <= 0 {
		job.MaxAttempts = m.config.DefaultMaxAttempts
	}

	m.schedules[job.Name] = job
	m.logger.Debug("Registered scheduled job",
		"name", job.Name,
		"schedule", job.Schedule.String(),
		"queue", job.Queue,
	)

	return nil
}

// UnregisterSchedule removes a scheduled job
func (m *Manager) UnregisterSchedule(name string) error {
	m.schedulesMu.Lock()
	defer m.schedulesMu.Unlock()

	if _, exists := m.schedules[name]; !exists {
		return ErrScheduleNotFound
	}

	delete(m.schedules, name)
	m.logger.Debug("Unregistered scheduled job", "name", name)
	return nil
}

// Dispatch dispatches a job for immediate or delayed execution
func (m *Manager) Dispatch(ctx context.Context, def *contract.JobDefinition) (string, error) {
	// Apply defaults
	if def.Queue == "" {
		def.Queue = m.config.DefaultQueue
	}
	if def.MaxAttempts <= 0 {
		def.MaxAttempts = m.config.DefaultMaxAttempts
	}
	if def.Timeout <= 0 {
		def.Timeout = m.config.DefaultTimeout
	}
	if def.UniqueFor <= 0 {
		def.UniqueFor = m.config.DefaultUniqueTTL
	}

	// Create job first so we have the ID for the uniqueness key
	job, err := newJob(def)
	if err != nil {
		return "", fmt.Errorf("failed to create job: %w", err)
	}

	// Atomically reserve uniqueness key using SetNX (atomic check-and-set).
	// This prevents the TOCTOU race where two instances both pass an Exists
	// check before either sets the key.
	if def.UniqueKey != "" {
		uniqueKeyRedis := m.uniqueKey(def.Queue, def.UniqueKey)
		set, err := m.redis.SetNX(ctx, uniqueKeyRedis, job.id, def.UniqueFor).Result()
		if err != nil {
			return "", fmt.Errorf("failed to check job uniqueness: %w", err)
		}
		if !set {
			return "", ErrJobAlreadyExists
		}
	}

	// Serialize job
	jobData, err := job.toJobData()
	if err != nil {
		return "", fmt.Errorf("failed to serialize job: %w", err)
	}
	jobBytes, err := jobData.marshal()
	if err != nil {
		return "", fmt.Errorf("failed to marshal job: %w", err)
	}

	// Store job data
	jobKey := m.jobKey(job.id)
	if err := m.redis.Set(ctx, jobKey, jobBytes, 0).Err(); err != nil {
		return "", fmt.Errorf("failed to store job: %w", err)
	}

	// Add to appropriate queue
	if job.status == contract.JobStatusScheduled {
		// Add to scheduled queue (sorted set with score = scheduled time)
		scheduledKey := m.scheduledKey(job.queue)
		if err := m.redis.ZAdd(ctx, scheduledKey, redis.Z{
			Score:  float64(job.scheduledAt.UnixMilli()),
			Member: job.id,
		}).Err(); err != nil {
			return "", fmt.Errorf("failed to schedule job: %w", err)
		}
	} else {
		// Add to ready queue
		queueKey := m.queueKey(job.queue)
		if err := m.redis.LPush(ctx, queueKey, job.id).Err(); err != nil {
			return "", fmt.Errorf("failed to queue job: %w", err)
		}
	}

	// Track queue
	m.trackQueue(def.Queue) //nolint:contextcheck // trackQueue starts long-lived worker goroutines that must outlive the request context

	m.logger.Debug("Dispatched job",
		"id", job.id,
		"name", job.name,
		"queue", job.queue,
		"status", job.status,
		"scheduled_at", job.scheduledAt,
	)

	return job.id, nil
}

// DispatchMany dispatches multiple jobs in a batch
func (m *Manager) DispatchMany(ctx context.Context, jobs []*contract.JobDefinition) ([]string, error) {
	ids := make([]string, 0, len(jobs))
	for _, def := range jobs {
		id, err := m.Dispatch(ctx, def)
		if err != nil {
			// Continue on duplicate key errors
			if errors.Is(err, ErrJobAlreadyExists) {
				ids = append(ids, "")
				continue
			}
			return ids, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// Cancel cancels a pending or scheduled job
func (m *Manager) Cancel(ctx context.Context, jobID string) error {
	job, err := m.getJobInternal(ctx, jobID)
	if err != nil {
		return err
	}

	// Can only cancel pending or scheduled jobs
	if job.status != contract.JobStatusPending && job.status != contract.JobStatusScheduled {
		return fmt.Errorf("cannot cancel job with status %s", job.status)
	}

	// Save original status before overwriting so we remove from the correct queue
	originalStatus := job.status

	// Update status
	job.status = contract.JobStatusCancelled
	job.completedAt = time.Now()

	// Save updated job
	if err := m.saveJob(ctx, job); err != nil {
		return err
	}

	// Remove from queues based on original status
	if originalStatus == contract.JobStatusScheduled {
		scheduledKey := m.scheduledKey(job.queue)
		m.redis.ZRem(ctx, scheduledKey, jobID)
	} else {
		queueKey := m.queueKey(job.queue)
		m.redis.LRem(ctx, queueKey, 0, jobID)
	}

	m.logger.Debug("Cancelled job", "id", jobID)
	return nil
}

// Get retrieves a job by ID
func (m *Manager) Get(ctx context.Context, jobID string) (contract.Job, error) {
	return m.getJobInternal(ctx, jobID)
}

func (m *Manager) getJobInternal(ctx context.Context, jobID string) (*jobImpl, error) {
	jobKey := m.jobKey(jobID)
	data, err := m.redis.Get(ctx, jobKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrJobNotFound
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	jobData, err := unmarshalJobData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return fromJobData(jobData), nil
}

func (m *Manager) saveJob(ctx context.Context, job *jobImpl) error {
	jobData, err := job.toJobData()
	if err != nil {
		return err
	}
	jobBytes, err := jobData.marshal()
	if err != nil {
		return err
	}
	return m.redis.Set(ctx, m.jobKey(job.id), jobBytes, 0).Err()
}

// Start starts the job workers and scheduler
func (m *Manager) Start(_ context.Context) error {
	if m.running.Load() {
		return nil
	}

	m.running.Store(true)
	m.stopCh = make(chan struct{})

	// Use a background context for all long-running goroutines.
	// The Fx startup context passed here expires after StartTimeout (2 min),
	// which would kill all background goroutines. Graceful shutdown is handled
	// by closing stopCh in Stop().
	bgCtx := context.Background()

	// Start workers for each queue
	m.workersMu.Lock()
	for _, queue := range m.totalQueues {
		for i := 0; i < m.config.Concurrency; i++ {
			w := newWorker(m, queue, i)
			m.workers = append(m.workers, w)
			m.wg.Add(1)
			go func(w *worker) {
				defer m.wg.Done()
				w.run(bgCtx)
			}(w)
		}
	}
	m.workersMu.Unlock()

	// Start scheduler if enabled
	if m.config.SchedulerEnabled {
		m.wg.Go(func() {
			m.runScheduler(bgCtx)
		})
	}

	// Start scheduled job promoter
	m.wg.Go(func() {
		m.runPromoter(bgCtx)
	})

	m.logger.Info("Job manager started",
		"queues", m.totalQueues,
		"workers_per_queue", m.config.Concurrency,
		"scheduler_enabled", m.config.SchedulerEnabled,
	)

	return nil
}

// Stop gracefully stops all workers
func (m *Manager) Stop(ctx context.Context) error {
	if !m.running.Load() {
		return nil
	}

	m.logger.Debug("Stopping job manager...")

	// Signal all goroutines to stop
	close(m.stopCh)
	m.running.Store(false)

	// Wait for graceful shutdown with timeout
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.logger.Info("Job manager stopped gracefully")
	case <-time.After(m.config.ShutdownTimeout):
		m.logger.Warn("Job manager shutdown timed out")
	case <-ctx.Done():
		m.logger.Warn("Job manager shutdown cancelled")
	}

	// Close Redis connection
	if err := m.redis.Close(); err != nil {
		m.logger.Warn("Error closing Redis connection", "error", err)
	}

	return nil
}

// Health checks the health of the job system
func (m *Manager) Health(ctx context.Context) error {
	return m.redis.Ping(ctx).Err()
}

// Stats returns job system statistics
func (m *Manager) Stats(ctx context.Context) (*contract.JobStats, error) {
	stats := &contract.JobStats{
		Queues: make(map[string]*contract.QueueStats),
		Workers: &contract.WorkerStats{
			Total:  len(m.workers),
			Active: int(m.activeJobs.Load()),
		},
		Scheduler: &contract.SchedulerStats{
			RegisteredJobs: len(m.schedules),
			Running:        m.running.Load() && m.config.SchedulerEnabled,
		},
	}

	stats.Workers.Idle = stats.Workers.Total - stats.Workers.Active

	// Get queue stats
	for _, queue := range m.totalQueues {
		queueStats := &contract.QueueStats{Name: queue}

		// Count pending jobs
		queueKey := m.queueKey(queue)
		pending, _ := m.redis.LLen(ctx, queueKey).Result()
		queueStats.Pending = pending

		// Count scheduled jobs
		scheduledKey := m.scheduledKey(queue)
		scheduled, _ := m.redis.ZCard(ctx, scheduledKey).Result()
		queueStats.Scheduled = scheduled

		// Get internal stats if available
		m.stats.mu.RLock()
		if internal, ok := m.stats.queueStats[queue]; ok {
			queueStats.Running = internal.running.Load()
			queueStats.Completed = internal.completed.Load()
			queueStats.Failed = internal.failed.Load()
		}
		m.stats.mu.RUnlock()

		stats.Queues[queue] = queueStats
	}

	// Get next scheduled run time
	m.schedulesMu.RLock()
	var nextRun time.Time
	now := time.Now()
	for _, sched := range m.schedules {
		next := sched.Schedule.Next(now)
		if nextRun.IsZero() || next.Before(nextRun) {
			nextRun = next
		}
	}
	stats.Scheduler.NextRun = nextRun
	m.schedulesMu.RUnlock()

	return stats, nil
}

// runScheduler runs the scheduler for recurring jobs.
// In multi-instance deployments, a distributed lock ensures only one instance
// dispatches scheduled jobs per tick. The atomic uniqueness check in Dispatch
// serves as a second safety net against duplicate dispatches.
func (m *Manager) runScheduler(ctx context.Context) {
	ticker := time.NewTicker(m.config.SchedulerInterval)
	defer ticker.Stop()

	// Track last run times
	lastRuns := make(map[string]time.Time)

	for {
		select {
		case <-m.stopCh:
			return
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			// Try to acquire a distributed lock for this scheduler tick.
			// Only one instance across all replicas will win the lock.
			// The lock auto-expires so another instance takes over if this one dies.
			lockKey := m.config.KeyPrefix + "lock:scheduler"
			acquired, err := m.redis.SetNX(ctx, lockKey, "1",
				m.config.SchedulerInterval+500*time.Millisecond,
			).Result()
			if err != nil || !acquired {
				continue // another instance is handling this tick
			}

			m.schedulesMu.RLock()
			for name, sched := range m.schedules {
				lastRun := lastRuns[name]
				nextRun := sched.Schedule.Next(lastRun)

				if nextRun.After(lastRun) && !nextRun.After(now) {
					// Time to dispatch this job
					def := &contract.JobDefinition{
						Name:        name,
						Payload:     sched.Payload,
						Queue:       sched.Queue,
						MaxAttempts: sched.MaxAttempts,
						Timeout:     sched.Timeout,
						Tags:        sched.Tags,
					}

					// If no overlap allowed, use job name as unique key
					if !sched.Overlap {
						def.UniqueKey = fmt.Sprintf("scheduled:%s", name)
						def.UniqueFor = time.Hour // Prevent overlap for up to 1 hour
					}

					_, err := m.Dispatch(ctx, def)
					if err != nil && !errors.Is(err, ErrJobAlreadyExists) {
						m.logger.Error("Failed to dispatch scheduled job",
							"name", name,
							"error", err,
						)
					} else if err == nil {
						m.logger.Debug("Dispatched scheduled job", "name", name)
					}

					lastRuns[name] = now
				}
			}
			m.schedulesMu.RUnlock()
		}
	}
}

// runPromoter moves scheduled jobs to the ready queue when their time comes.
// Uses a Lua script for atomic promote to prevent duplicate processing when
// multiple app instances run concurrently.
func (m *Manager) runPromoter(ctx context.Context) {
	ticker := time.NewTicker(m.config.SchedulerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := fmt.Sprintf("%d", time.Now().UnixMilli())

			for _, queue := range m.totalQueues {
				scheduledKey := m.scheduledKey(queue)
				queueKey := m.queueKey(queue)

				// Atomically: find due jobs, remove from sorted set, push to ready queue.
				// The Lua script ensures only one instance promotes each job, even if
				// multiple instances call ZRangeByScore at the same time.
				result, err := promoteScript.Run(ctx, m.redis,
					[]string{scheduledKey, queueKey}, now,
				).StringSlice()

				if err != nil && !errors.Is(err, redis.Nil) {
					m.logger.Warn("Failed to promote scheduled jobs", "queue", queue, "error", err)
					continue
				}

				// Update status for each promoted job
				for _, jobID := range result {
					job, err := m.getJobInternal(ctx, jobID)
					if err == nil {
						job.status = contract.JobStatusPending
						_ = m.saveJob(ctx, job)
					}
				}
			}
		}
	}
}

// trackQueue ensures a queue is being watched
func (m *Manager) trackQueue(queue string) {
	m.workersMu.Lock()
	defer m.workersMu.Unlock()

	if slices.Contains(m.totalQueues, queue) {
		return
	}

	m.totalQueues = append(m.totalQueues, queue)

	// Start workers for new queue if manager is running
	if m.running.Load() {
		for i := 0; i < m.config.Concurrency; i++ {
			w := newWorker(m, queue, i)
			m.workers = append(m.workers, w)
			m.wg.Add(1)
			go func(w *worker) {
				defer m.wg.Done()
				w.run(context.Background())
			}(w)
		}
	}

	// Initialize queue stats
	m.stats.mu.Lock()
	if _, ok := m.stats.queueStats[queue]; !ok {
		m.stats.queueStats[queue] = &queueStatsInternal{}
	}
	m.stats.mu.Unlock()
}

// Redis key helpers
func (m *Manager) jobKey(id string) string {
	return m.config.KeyPrefix + "data:" + id
}

func (m *Manager) queueKey(queue string) string {
	return m.config.KeyPrefix + "queue:" + queue
}

func (m *Manager) scheduledKey(queue string) string {
	return m.config.KeyPrefix + "scheduled:" + queue
}

func (m *Manager) uniqueKey(queue, key string) string {
	return m.config.KeyPrefix + "unique:" + queue + ":" + key
}

func (m *Manager) processingKey(queue string) string {
	return m.config.KeyPrefix + "processing:" + queue
}

func (m *Manager) dlqKey(queue string) string {
	return m.config.KeyPrefix + "dlq:" + queue
}

// getHandler returns the handler for a job name
func (m *Manager) getHandler(name string) (contract.JobHandler, bool) {
	m.handlersMu.RLock()
	defer m.handlersMu.RUnlock()

	// First check direct registration
	if handler, ok := m.handlers[name]; ok {
		return handler, true
	}

	// Then check scheduled jobs
	m.schedulesMu.RLock()
	defer m.schedulesMu.RUnlock()
	if sched, ok := m.schedules[name]; ok && sched.Handler != nil {
		return sched.Handler, true
	}

	return nil, false
}

// recordCompletion records job completion statistics
func (m *Manager) recordCompletion(queue string, success bool) {
	m.stats.mu.Lock()
	qs, ok := m.stats.queueStats[queue]
	if !ok {
		qs = &queueStatsInternal{}
		m.stats.queueStats[queue] = qs
	}
	m.stats.mu.Unlock()

	if success {
		qs.completed.Add(1)
		m.stats.completed.Add(1)
	} else {
		qs.failed.Add(1)
		m.stats.failed.Add(1)
	}
	m.stats.processed.Add(1)
}

// moveToDLQ moves a failed job to the dead letter queue
func (m *Manager) moveToDLQ(ctx context.Context, job *jobImpl) error {
	if !m.config.DLQEnabled {
		return nil
	}

	dlqKey := m.dlqKey(job.queue)

	// Store job data with DLQ metadata
	dlqData := map[string]any{
		"job_id":    job.id,
		"name":      job.name,
		"queue":     job.queue,
		"error":     job.lastError,
		"attempts":  job.attempts,
		"failed_at": time.Now().UnixMilli(),
	}

	dlqBytes, err := json.Marshal(dlqData)
	if err != nil {
		return err
	}

	// Add to DLQ with TTL
	pipe := m.redis.Pipeline()
	pipe.LPush(ctx, dlqKey, dlqBytes)
	pipe.Expire(ctx, dlqKey, m.config.DLQTTL)
	_, err = pipe.Exec(ctx)

	return err
}
