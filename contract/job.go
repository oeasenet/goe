package contract

import (
	"context"
	"time"
)

// JobStatus represents the current state of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"   // Job is waiting to be executed
	JobStatusScheduled JobStatus = "scheduled" // Job is scheduled for future execution
	JobStatusRunning   JobStatus = "running"   // Job is currently being executed
	JobStatusCompleted JobStatus = "completed" // Job completed successfully
	JobStatusFailed    JobStatus = "failed"    // Job failed after all retries
	JobStatusCancelled JobStatus = "cancelled" // Job was cancelled
)

// Job represents a job in the system
type Job interface {
	// ID returns the unique job ID
	ID() string

	// Name returns the job name/type
	Name() string

	// Queue returns the queue name this job belongs to
	Queue() string

	// Payload returns the job payload data
	Payload() any

	// Status returns the current job status
	Status() JobStatus

	// Attempts returns the number of execution attempts
	Attempts() int

	// MaxAttempts returns the maximum number of retry attempts
	MaxAttempts() int

	// CreatedAt returns when the job was created
	CreatedAt() time.Time

	// ScheduledAt returns when the job is scheduled to run
	ScheduledAt() time.Time

	// StartedAt returns when the job started executing (zero if not started)
	StartedAt() time.Time

	// CompletedAt returns when the job completed (zero if not completed)
	CompletedAt() time.Time

	// LastError returns the last error message if any
	LastError() string

	// Context returns the job context
	Context() context.Context

	// Tags returns job metadata tags for filtering/grouping
	Tags() map[string]string
}

// JobHandler handles job execution
type JobHandler interface {
	// Handle executes the job and returns an error if processing fails
	Handle(ctx context.Context, job Job) error
}

// JobHandlerFunc is a function adapter for JobHandler
type JobHandlerFunc func(ctx context.Context, job Job) error

// Handle implements JobHandler interface
func (f JobHandlerFunc) Handle(ctx context.Context, job Job) error {
	return f(ctx, job)
}

// Schedule represents a job schedule configuration
// This provides a developer-friendly way to define schedules without knowing cron syntax
type Schedule interface {
	// Next returns the next execution time after the given time
	Next(after time.Time) time.Time

	// String returns a human-readable description of the schedule
	String() string

	// Cron returns the cron expression if applicable, empty string otherwise
	Cron() string
}

// ScheduledJob represents a recurring job definition
type ScheduledJob struct {
	// Name is the unique identifier for this scheduled job
	Name string

	// Handler is the job handler function
	Handler JobHandler

	// Schedule defines when the job should run
	Schedule Schedule

	// Queue is the queue to use (default: "default")
	Queue string

	// Payload is the static payload for the job (optional)
	Payload any

	// Timeout is the maximum execution time (default: 30 minutes)
	Timeout time.Duration

	// MaxAttempts is the maximum retry attempts (default: 3)
	MaxAttempts int

	// Overlap determines if the job can run while a previous instance is still running
	// If false (default), the next scheduled run will be skipped if the job is still running
	Overlap bool

	// Tags are metadata tags for filtering/grouping
	Tags map[string]string
}

// JobManager provides the main interface for job management
type JobManager interface {
	// Dispatch dispatches a job for immediate or delayed execution
	Dispatch(ctx context.Context, job *JobDefinition) (string, error)

	// DispatchMany dispatches multiple jobs in a batch
	DispatchMany(ctx context.Context, jobs []*JobDefinition) ([]string, error)

	// Cancel cancels a pending or scheduled job
	Cancel(ctx context.Context, jobID string) error

	// Get retrieves a job by ID
	Get(ctx context.Context, jobID string) (Job, error)

	// RegisterHandler registers a handler for a job name
	RegisterHandler(name string, handler JobHandler)

	// RegisterSchedule registers a recurring scheduled job
	RegisterSchedule(job *ScheduledJob) error

	// UnregisterSchedule removes a scheduled job
	UnregisterSchedule(name string) error

	// Start starts the job workers and scheduler
	Start(ctx context.Context) error

	// Stop gracefully stops all workers
	Stop(ctx context.Context) error

	// Health checks the health of the job system
	Health(ctx context.Context) error

	// Stats returns job system statistics
	Stats(ctx context.Context) (*JobStats, error)
}

// JobDefinition defines a job to be dispatched
type JobDefinition struct {
	// Name is the job handler name to use
	Name string

	// Payload is the job payload data
	Payload any

	// Queue is the queue name (default: "default")
	Queue string

	// Delay is the delay before the job should be executed
	Delay time.Duration

	// ScheduledAt is the specific time to execute the job (overrides Delay)
	ScheduledAt time.Time

	// MaxAttempts is the maximum number of retry attempts (default: 3)
	MaxAttempts int

	// Timeout is the maximum execution time (default: 30 minutes)
	Timeout time.Duration

	// UniqueKey is used for job deduplication (optional)
	// Jobs with the same UniqueKey won't be duplicated within UniqueFor duration
	UniqueKey string

	// UniqueFor is the duration to enforce uniqueness (default: 1 hour)
	UniqueFor time.Duration

	// Tags are metadata tags for filtering/grouping
	Tags map[string]string
}

// JobStats contains job system statistics
type JobStats struct {
	// Queue statistics
	Queues map[string]*QueueStats

	// Worker statistics
	Workers *WorkerStats

	// Scheduler statistics
	Scheduler *SchedulerStats
}

// QueueStats contains statistics for a specific queue
type QueueStats struct {
	// Name is the queue name
	Name string

	// Pending is the number of jobs waiting to be processed
	Pending int64

	// Scheduled is the number of jobs scheduled for future execution
	Scheduled int64

	// Running is the number of jobs currently being processed
	Running int64

	// Completed is the total number of completed jobs
	Completed int64

	// Failed is the total number of failed jobs
	Failed int64

	// ProcessedPerSecond is the average processing rate
	ProcessedPerSecond float64
}

// WorkerStats contains worker statistics
type WorkerStats struct {
	// Total is the total number of workers
	Total int

	// Active is the number of workers currently processing jobs
	Active int

	// Idle is the number of idle workers
	Idle int
}

// SchedulerStats contains scheduler statistics
type SchedulerStats struct {
	// RegisteredJobs is the number of registered scheduled jobs
	RegisteredJobs int

	// NextRun is the time of the next scheduled job
	NextRun time.Time

	// Running indicates if the scheduler is running
	Running bool
}

// JobOption configures a job dispatch
type JobOption func(*JobDefinition)

// WithQueue sets the queue for the job
func WithJobQueue(queue string) JobOption {
	return func(j *JobDefinition) {
		j.Queue = queue
	}
}

// WithDelay sets the delay for the job
func WithJobDelay(delay time.Duration) JobOption {
	return func(j *JobDefinition) {
		j.Delay = delay
	}
}

// WithScheduledAt sets the scheduled execution time
func WithJobScheduledAt(t time.Time) JobOption {
	return func(j *JobDefinition) {
		j.ScheduledAt = t
	}
}

// WithMaxAttempts sets the maximum retry attempts
func WithJobMaxAttempts(attempts int) JobOption {
	return func(j *JobDefinition) {
		j.MaxAttempts = attempts
	}
}

// WithTimeout sets the job timeout
func WithJobTimeout(timeout time.Duration) JobOption {
	return func(j *JobDefinition) {
		j.Timeout = timeout
	}
}

// WithUniqueKey sets the unique key for deduplication
func WithJobUniqueKey(key string, duration time.Duration) JobOption {
	return func(j *JobDefinition) {
		j.UniqueKey = key
		j.UniqueFor = duration
	}
}

// WithTags sets the job tags
func WithJobTags(tags map[string]string) JobOption {
	return func(j *JobDefinition) {
		j.Tags = tags
	}
}

// NewJobDefinition creates a new job definition with options
func NewJobDefinition(name string, payload any, opts ...JobOption) *JobDefinition {
	j := &JobDefinition{
		Name:        name,
		Payload:     payload,
		Queue:       "default",
		MaxAttempts: 3,
		Timeout:     30 * time.Minute,
		UniqueFor:   time.Hour,
	}
	for _, opt := range opts {
		opt(j)
	}
	return j
}
