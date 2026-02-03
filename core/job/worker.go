package job

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"go.oease.dev/goe/v2/contract"
)

// worker processes jobs from a specific queue
type worker struct {
	manager *Manager
	queue   string
	id      int
	running bool
}

// newWorker creates a new worker
func newWorker(manager *Manager, queue string, id int) *worker {
	return &worker{
		manager: manager,
		queue:   queue,
		id:      id,
	}
}

// run starts the worker loop
func (w *worker) run(ctx context.Context) {
	w.running = true
	w.manager.logger.Debug("Worker started", "queue", w.queue, "worker_id", w.id)

	for {
		select {
		case <-w.manager.stopCh:
			w.running = false
			w.manager.logger.Debug("Worker stopped", "queue", w.queue, "worker_id", w.id)
			return
		case <-ctx.Done():
			w.running = false
			return
		default:
			w.processNextJob(ctx)
		}
	}
}

// processNextJob attempts to process the next job from the queue
func (w *worker) processNextJob(ctx context.Context) {
	// Use BRPOPLPUSH for atomic dequeue
	queueKey := w.manager.queueKey(w.queue)
	processingKey := w.manager.processingKey(w.queue)

	// Block with timeout waiting for a job
	result, err := w.manager.redis.BRPopLPush(ctx, queueKey, processingKey, w.manager.config.PollInterval).Result()
	if err != nil {
		// Timeout or error - just continue
		return
	}

	jobID := result

	// Get job data
	job, err := w.manager.getJobInternal(ctx, jobID)
	if err != nil {
		w.manager.logger.Error("Failed to get job", "job_id", jobID, "error", err)
		// Remove from processing queue
		w.manager.redis.LRem(ctx, processingKey, 0, jobID)
		return
	}

	// Process the job
	w.processJob(ctx, job)

	// Remove from processing queue
	w.manager.redis.LRem(ctx, processingKey, 0, jobID)
}

// processJob executes a single job
func (w *worker) processJob(ctx context.Context, job *jobImpl) {
	// Get handler
	handler, ok := w.manager.getHandler(job.name)
	if !ok {
		w.manager.logger.Error("No handler registered for job", "name", job.name, "job_id", job.id)
		w.handleFailure(ctx, job, ErrHandlerNotFound)
		return
	}

	// Update job status
	job.status = contract.JobStatusRunning
	job.startedAt = time.Now()
	job.attempts++
	w.manager.saveJob(ctx, job)

	// Track active jobs
	w.manager.activeJobs.Add(1)
	defer w.manager.activeJobs.Add(-1)

	// Update queue stats
	w.manager.stats.mu.Lock()
	if qs, ok := w.manager.stats.queueStats[job.queue]; ok {
		qs.running.Add(1)
	}
	w.manager.stats.mu.Unlock()
	defer func() {
		w.manager.stats.mu.Lock()
		if qs, ok := w.manager.stats.queueStats[job.queue]; ok {
			qs.running.Add(-1)
		}
		w.manager.stats.mu.Unlock()
	}()

	// Create job context with timeout
	jobCtx, cancel := context.WithTimeout(ctx, job.timeout)
	defer cancel()

	// Update job context
	job.ctx = jobCtx

	// Execute with panic recovery
	err := w.executeWithRecovery(jobCtx, handler, job)

	if err != nil {
		w.handleFailure(ctx, job, err)
	} else {
		w.handleSuccess(ctx, job)
	}
}

// executeWithRecovery executes the handler with panic recovery
func (w *worker) executeWithRecovery(ctx context.Context, handler contract.JobHandler, job *jobImpl) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			err = fmt.Errorf("panic in job handler: %v\n%s", r, string(stack))
			w.manager.logger.Error("Panic in job handler",
				"job_id", job.id,
				"name", job.name,
				"panic", r,
				"stack", string(stack),
			)
		}
	}()

	return handler.Handle(ctx, job)
}

// handleSuccess handles successful job completion
func (w *worker) handleSuccess(ctx context.Context, job *jobImpl) {
	job.status = contract.JobStatusCompleted
	job.completedAt = time.Now()
	job.lastError = ""

	if err := w.manager.saveJob(ctx, job); err != nil {
		w.manager.logger.Error("Failed to save completed job", "job_id", job.id, "error", err)
	}

	// Record statistics
	w.manager.recordCompletion(job.queue, true)

	// Clear uniqueness key if set
	if job.uniqueKey != "" {
		uniqueKeyRedis := w.manager.uniqueKey(job.queue, job.uniqueKey)
		w.manager.redis.Del(ctx, uniqueKeyRedis)
	}

	w.manager.logger.Debug("Job completed successfully",
		"job_id", job.id,
		"name", job.name,
		"attempts", job.attempts,
		"duration", job.completedAt.Sub(job.startedAt),
	)
}

// handleFailure handles job failure with retry logic
func (w *worker) handleFailure(ctx context.Context, job *jobImpl, err error) {
	job.lastError = err.Error()

	w.manager.logger.Warn("Job failed",
		"job_id", job.id,
		"name", job.name,
		"attempt", job.attempts,
		"max_attempts", job.maxAttempts,
		"error", err,
	)

	// Check if we should retry
	if job.attempts < job.maxAttempts {
		// Calculate backoff
		backoff := w.calculateBackoff(job.attempts)

		// Schedule for retry
		job.status = contract.JobStatusScheduled
		job.scheduledAt = time.Now().Add(backoff)

		if err := w.manager.saveJob(ctx, job); err != nil {
			w.manager.logger.Error("Failed to save job for retry", "job_id", job.id, "error", err)
			return
		}

		// Add to scheduled queue
		scheduledKey := w.manager.scheduledKey(job.queue)
		w.manager.redis.ZAdd(ctx, scheduledKey, struct {
			Score  float64
			Member interface{}
		}{
			Score:  float64(job.scheduledAt.UnixMilli()),
			Member: job.id,
		})

		w.manager.stats.retried.Add(1)

		w.manager.logger.Info("Job scheduled for retry",
			"job_id", job.id,
			"name", job.name,
			"attempt", job.attempts,
			"next_attempt_at", job.scheduledAt,
			"backoff", backoff,
		)
	} else {
		// Max retries exceeded - mark as failed
		job.status = contract.JobStatusFailed
		job.completedAt = time.Now()

		if err := w.manager.saveJob(ctx, job); err != nil {
			w.manager.logger.Error("Failed to save failed job", "job_id", job.id, "error", err)
		}

		// Move to DLQ
		if err := w.manager.moveToDLQ(ctx, job); err != nil {
			w.manager.logger.Error("Failed to move job to DLQ", "job_id", job.id, "error", err)
		}

		// Record statistics
		w.manager.recordCompletion(job.queue, false)

		// Clear uniqueness key if set
		if job.uniqueKey != "" {
			uniqueKeyRedis := w.manager.uniqueKey(job.queue, job.uniqueKey)
			w.manager.redis.Del(ctx, uniqueKeyRedis)
		}

		w.manager.logger.Error("Job failed after max retries",
			"job_id", job.id,
			"name", job.name,
			"attempts", job.attempts,
			"error", job.lastError,
		)
	}
}

// calculateBackoff calculates the backoff duration for a retry attempt
func (w *worker) calculateBackoff(attempt int) time.Duration {
	backoff := w.manager.config.RetryBackoff

	// Apply exponential backoff
	for i := 1; i < attempt; i++ {
		backoff = time.Duration(float64(backoff) * w.manager.config.RetryBackoffFactor)
	}

	// Cap at max backoff
	if backoff > w.manager.config.MaxRetryBackoff {
		backoff = w.manager.config.MaxRetryBackoff
	}

	return backoff
}
