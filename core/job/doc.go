// Package job provides a developer-friendly, production-ready background job
// processing system for the GOE framework.
//
// # Overview
//
// The job system is designed to be easy to use while providing all the features
// needed for production workloads:
//
//   - Simple, human-friendly scheduling (no cron syntax required)
//   - Delayed and scheduled job execution
//   - Automatic retries with exponential backoff
//   - Job uniqueness/deduplication
//   - Multiple named queues
//   - Dead letter queue for failed jobs
//   - Health checks and statistics
//
// # Differences from the Event System
//
// While events are fire-and-forget pub/sub messages, jobs provide:
//
//   - Persistent state tracking (pending → running → completed/failed)
//   - Guaranteed at-least-once delivery
//   - Scheduling and delayed execution
//   - Retry logic with backoff
//   - Unique job constraints
//
// # Quick Start
//
// Enable the job module in your GOE application:
//
//	app := goe.New(goe.Options{
//	    WithJob: true,
//	})
//
// Register a job handler:
//
//	goe.Job().RegisterHandler("send-email", contract.JobHandlerFunc(
//	    func(ctx context.Context, j contract.Job) error {
//	        payload := j.Payload().(map[string]any)
//	        email := payload["email"].(string)
//	        // Send the email...
//	        return nil
//	    },
//	))
//
// Dispatch a job:
//
//	jobID, err := goe.Job().Dispatch(ctx, &contract.JobDefinition{
//	    Name:    "send-email",
//	    Payload: map[string]any{"email": "user@example.com"},
//	})
//
// # Fluent Job Builder
//
// For a more expressive API, use the job builder:
//
//	// Dispatch immediately
//	jobID, err := goe.Job().Dispatch(ctx,
//	    job.NewJob("send-email", emailData).
//	        OnQueue("emails").
//	        WithRetries(5).
//	        Build(),
//	)
//
//	// Dispatch with delay
//	jobID, err := goe.Job().Dispatch(ctx,
//	    job.NewJob("send-reminder", data).
//	        InMinutes(30).
//	        Build(),
//	)
//
//	// Unique job (prevent duplicates)
//	jobID, err := goe.Job().Dispatch(ctx,
//	    job.NewJob("sync-user", userData).
//	        Unique("sync-user-" + userID).
//	        Build(),
//	)
//
// # Human-Friendly Scheduling
//
// The job package provides schedule helpers that don't require cron syntax:
//
//	// Simple intervals
//	job.Every(5 * time.Minute)     // Every 5 minutes
//	job.EveryFifteenMinutes()      // Every 15 minutes
//	job.Hourly()                   // Every hour
//
//	// Daily schedules
//	job.Daily()                    // Every day at midnight
//	job.DailyAt(9, 30)             // Every day at 9:30 AM
//	job.TwiceDaily(8, 20)          // At 8 AM and 8 PM
//
//	// Weekly schedules
//	job.Weekly()                   // Every Sunday at midnight
//	job.WeeklyOn(time.Monday, 9, 0) // Every Monday at 9 AM
//	job.Weekdays(9, 0)             // Mon-Fri at 9 AM
//	job.Weekends(10, 0)            // Sat-Sun at 10 AM
//
//	// Monthly schedules
//	job.Monthly()                  // 1st of each month at midnight
//	job.MonthlyOn(15, 9, 0)        // 15th of each month at 9 AM
//	job.LastDayOfMonth(17, 0)      // Last day of each month at 5 PM
//
//	// Special schedules
//	job.Quarterly(9, 0)            // First day of each quarter
//	job.Yearly()                   // January 1st at midnight
//
//	// Schedule modifiers
//	job.Between(job.Every(15*time.Minute), 9, 17)  // Only between 9 AM and 5 PM
//	job.SkipWeekends(job.Daily())                   // Daily except weekends
//
//	// Raw cron (when needed)
//	job.Cron("0 9 * * 1-5")        // Weekdays at 9 AM (cron syntax)
//
// # Scheduled Jobs
//
// Register recurring jobs that run on a schedule:
//
//	goe.Job().RegisterSchedule(&contract.ScheduledJob{
//	    Name:     "cleanup-expired",
//	    Schedule: job.DailyAt(3, 0),  // 3 AM daily
//	    Handler:  cleanupHandler,
//	    Queue:    "maintenance",
//	})
//
// Or use the fluent builder:
//
//	goe.Job().RegisterSchedule(
//	    job.NewScheduledJob("send-reports").
//	        Runs(job.Weekdays(9, 0)).
//	        Handle(reportHandler).
//	        OnQueue("reports").
//	        Build(),
//	)
//
// # Configuration
//
// Configure via environment variables:
//
//	JOB_REDIS_URL=redis://host:6379/0  # Redis URL (environment-only, wins over hosts)
//	JOB_REDIS_HOSTS=host1:6379         # Redis hosts (used when no URL is set)
//	JOB_REDIS_ADDR=localhost:6379      # Redis address (deprecated; use URL or HOSTS)
//	JOB_REDIS_USERNAME=svc             # Redis username (environment-only)
//	JOB_REDIS_PASSWORD=secret          # Redis password (environment-only)
//	JOB_CONCURRENCY=5                  # Workers per queue
//	JOB_DEFAULT_QUEUE=default          # Default queue name
//	JOB_DEFAULT_MAX_ATTEMPTS=3         # Default retry attempts
//	JOB_RETRY_BACKOFF=1s               # Initial retry delay
//	JOB_RETRY_BACKOFF_FACTOR=2.0       # Backoff multiplier
//	JOB_SCHEDULER_ENABLED=true         # Enable scheduler
//	JOB_DLQ_ENABLED=true               # Enable dead letter queue
//
// Or from Go code through goe.Options.Job — every Config field has a
// matching With<FieldName> Option, and code wins over the environment:
//
//	goe.New(goe.Options{
//	    Job: []job.Option{
//	        job.WithConcurrency(10),
//	        job.WithDefaultQueue("critical"),
//	        job.WithDLQTTL(48 * time.Hour),
//	    },
//	})
//
// Credentials are the exception: JOB_REDIS_URL, JOB_REDIS_USERNAME and
// JOB_REDIS_PASSWORD have no Option. Secrets stay in the environment and are
// applied to whichever endpoint wins — including one chosen in code with
// job.WithRedisHosts.
//
// # Job Status Flow
//
//	┌─────────┐    ┌───────────┐    ┌─────────┐    ┌───────────┐
//	│ pending │───▶│  running  │───▶│completed│    │  failed   │
//	└─────────┘    └───────────┘    └─────────┘    └───────────┘
//	     │               │                              ▲
//	     │               │         (retry)              │
//	     │               └──────────────────────────────┘
//	     │                         (max retries exceeded)
//	     ▼
//	┌───────────┐
//	│ scheduled │  (future execution)
//	└───────────┘
//
// # Multiple Queues
//
// Use named queues to prioritize different job types:
//
//	// High priority - more workers
//	goe.Job().Dispatch(ctx,
//	    job.NewJob("urgent-notification", data).OnQueue("high").Build(),
//	)
//
//	// Low priority - fewer workers
//	goe.Job().Dispatch(ctx,
//	    job.NewJob("generate-report", data).OnQueue("low").Build(),
//	)
//
// Configure workers per queue in your application startup if needed.
//
// # Error Handling and Retries
//
// Jobs automatically retry on failure with exponential backoff:
//
//	func myHandler(ctx context.Context, j contract.Job) error {
//	    if err := doWork(); err != nil {
//	        // Return error to trigger retry
//	        return fmt.Errorf("work failed: %w", err)
//	    }
//	    return nil  // Success - no retry
//	}
//
// After max retries are exhausted, jobs move to the dead letter queue.
//
// # Statistics and Monitoring
//
// Get job system statistics:
//
//	stats, err := goe.Job().Stats(ctx)
//	// stats.Queues["default"].Pending
//	// stats.Queues["default"].Running
//	// stats.Workers.Active
//	// stats.Scheduler.RegisteredJobs
//
// Health check integration is automatic when using WithHealth.
package job
