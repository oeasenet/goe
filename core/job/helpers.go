package job

import (
	"time"

	"go.oease.dev/goe/v2/contract"
)

// =============================================================================
// Fluent Job Builder - More Developer-Friendly API
// =============================================================================

// JobBuilder provides a fluent API for creating job definitions
type JobBuilder struct {
	def *contract.JobDefinition
}

// NewJob creates a new job builder with the given name and payload
//
// Example:
//
//	job.NewJob("send-email", emailData).
//	    OnQueue("emails").
//	    WithDelay(5 * time.Minute).
//	    Build()
func NewJob(name string, payload any) *JobBuilder {
	return &JobBuilder{
		def: &contract.JobDefinition{
			Name:        name,
			Payload:     payload,
			Queue:       "default",
			MaxAttempts: 3,
			Timeout:     30 * time.Minute,
			UniqueFor:   time.Hour,
		},
	}
}

// OnQueue sets the queue for the job
func (b *JobBuilder) OnQueue(queue string) *JobBuilder {
	b.def.Queue = queue
	return b
}

// WithDelay sets a delay before the job should execute
func (b *JobBuilder) WithDelay(delay time.Duration) *JobBuilder {
	b.def.Delay = delay
	return b
}

// At schedules the job for a specific time
func (b *JobBuilder) At(t time.Time) *JobBuilder {
	b.def.ScheduledAt = t
	return b
}

// In schedules the job to run after a duration (alias for WithDelay)
func (b *JobBuilder) In(delay time.Duration) *JobBuilder {
	return b.WithDelay(delay)
}

// InMinutes schedules the job to run after N minutes
func (b *JobBuilder) InMinutes(minutes int) *JobBuilder {
	return b.WithDelay(time.Duration(minutes) * time.Minute)
}

// InHours schedules the job to run after N hours
func (b *JobBuilder) InHours(hours int) *JobBuilder {
	return b.WithDelay(time.Duration(hours) * time.Hour)
}

// Tomorrow schedules the job to run tomorrow at the same time
func (b *JobBuilder) Tomorrow() *JobBuilder {
	return b.WithDelay(24 * time.Hour)
}

// WithRetries sets the maximum number of retry attempts
func (b *JobBuilder) WithRetries(attempts int) *JobBuilder {
	b.def.MaxAttempts = attempts
	return b
}

// WithTimeout sets the maximum execution time
func (b *JobBuilder) WithTimeout(timeout time.Duration) *JobBuilder {
	b.def.Timeout = timeout
	return b
}

// Unique ensures only one job with this key exists
// If a job with the same key already exists, Dispatch will return an error
func (b *JobBuilder) Unique(key string) *JobBuilder {
	b.def.UniqueKey = key
	return b
}

// UniqueFor sets both unique key and duration
func (b *JobBuilder) UniqueFor(key string, duration time.Duration) *JobBuilder {
	b.def.UniqueKey = key
	b.def.UniqueFor = duration
	return b
}

// WithTag adds a tag to the job
func (b *JobBuilder) WithTag(key, value string) *JobBuilder {
	if b.def.Tags == nil {
		b.def.Tags = make(map[string]string)
	}
	b.def.Tags[key] = value
	return b
}

// WithTags sets all tags at once
func (b *JobBuilder) WithTags(tags map[string]string) *JobBuilder {
	b.def.Tags = tags
	return b
}

// Build returns the job definition
func (b *JobBuilder) Build() *contract.JobDefinition {
	return b.def
}

// =============================================================================
// Scheduled Job Builder
// =============================================================================

// ScheduledJobBuilder provides a fluent API for creating scheduled jobs
type ScheduledJobBuilder struct {
	job *contract.ScheduledJob
}

// NewScheduledJob creates a new scheduled job builder
//
// Example:
//
//	job.NewScheduledJob("cleanup-expired").
//	    Runs(job.Every(time.Hour)).
//	    Handle(myCleanupHandler).
//	    OnQueue("maintenance").
//	    Build()
func NewScheduledJob(name string) *ScheduledJobBuilder {
	return &ScheduledJobBuilder{
		job: &contract.ScheduledJob{
			Name:        name,
			Queue:       "default",
			Timeout:     30 * time.Minute,
			MaxAttempts: 3,
		},
	}
}

// Runs sets the schedule (shorthand for Using)
func (b *ScheduledJobBuilder) Runs(s contract.Schedule) *ScheduledJobBuilder {
	b.job.Schedule = s
	return b
}

// Using sets the schedule explicitly
func (b *ScheduledJobBuilder) Using(s contract.Schedule) *ScheduledJobBuilder {
	b.job.Schedule = s
	return b
}

// Handle sets the job handler
func (b *ScheduledJobBuilder) Handle(handler contract.JobHandler) *ScheduledJobBuilder {
	b.job.Handler = handler
	return b
}

// HandleFunc sets a function as the job handler
func (b *ScheduledJobBuilder) HandleFunc(fn contract.JobHandlerFunc) *ScheduledJobBuilder {
	b.job.Handler = fn
	return b
}

// OnQueue sets the queue for the scheduled job
func (b *ScheduledJobBuilder) OnQueue(queue string) *ScheduledJobBuilder {
	b.job.Queue = queue
	return b
}

// WithPayload sets a static payload for the job
func (b *ScheduledJobBuilder) WithPayload(payload any) *ScheduledJobBuilder {
	b.job.Payload = payload
	return b
}

// WithTimeout sets the maximum execution time
func (b *ScheduledJobBuilder) WithTimeout(timeout time.Duration) *ScheduledJobBuilder {
	b.job.Timeout = timeout
	return b
}

// WithRetries sets the maximum retry attempts
func (b *ScheduledJobBuilder) WithRetries(attempts int) *ScheduledJobBuilder {
	b.job.MaxAttempts = attempts
	return b
}

// AllowOverlap allows the job to run even if a previous instance is still running
func (b *ScheduledJobBuilder) AllowOverlap() *ScheduledJobBuilder {
	b.job.Overlap = true
	return b
}

// PreventOverlap prevents the job from running if a previous instance is still running (default)
func (b *ScheduledJobBuilder) PreventOverlap() *ScheduledJobBuilder {
	b.job.Overlap = false
	return b
}

// WithTag adds a tag to the scheduled job
func (b *ScheduledJobBuilder) WithTag(key, value string) *ScheduledJobBuilder {
	if b.job.Tags == nil {
		b.job.Tags = make(map[string]string)
	}
	b.job.Tags[key] = value
	return b
}

// Build returns the scheduled job
func (b *ScheduledJobBuilder) Build() *contract.ScheduledJob {
	return b.job
}

// =============================================================================
// Quick Dispatch Helpers
// =============================================================================

// Dispatch is a shorthand for creating and getting a job definition ready for dispatch
// Use with a JobManager:
//
//	goe.Job().Dispatch(ctx, job.Dispatch("send-email", data).Build())
func Dispatch(name string, payload any) *JobBuilder {
	return NewJob(name, payload)
}

// Later creates a delayed job
//
//	goe.Job().Dispatch(ctx, job.Later("send-email", data, 5*time.Minute).Build())
func Later(name string, payload any, delay time.Duration) *JobBuilder {
	return NewJob(name, payload).WithDelay(delay)
}

// AtTime creates a job scheduled for a specific time
//
//	goe.Job().Dispatch(ctx, job.AtTime("send-report", data, tomorrowNoon).Build())
func AtTime(name string, payload any, t time.Time) *JobBuilder {
	return NewJob(name, payload).At(t)
}
