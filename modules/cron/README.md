# Cron Module

The Cron module provides a simple interface for scheduling and running tasks at specified intervals. It implements the [`contracts.CronJob`](https://github.com/oeasenet/goe/blob/main/contracts/cron.go) interface and is built on top of the [gocron](https://github.com/go-co-op/gocron) library.

## Features

- Schedule tasks using cron expressions
- Run tasks at specified intervals
- Concurrent job execution control
- Automatic task management
- Graceful shutdown support
- Thread-safe operations

## Usage

### Initialization

The cron module is automatically initialized by the GOE framework:

```go
// Get the cron service
cron := goe.UseCron()
```

### Scheduling Tasks

To schedule a task, you need to define a job with a cron expression and a handler function:

```go
// Import the required packages
import (
    "fmt"
    "github.com/go-co-op/gocron/v2"
    "time"
)

// Define a job that runs every hour
err := cron.DefineJob(
    gocron.DurationJob(
        time.Hour,
    ),
    func() {
        fmt.Println("Running hourly task")
    },
)
if err != nil {
    // Handle error
}

// Define a job with a cron expression (every 5 minutes)
err = cron.DefineJob(
    gocron.CronJob(
        "*/5 * * * *",
        gocron.WithLocation(time.UTC),
    ),
    func() {
        fmt.Println("Running every 5 minutes")
    },
)
if err != nil {
    // Handle error
}

// Define a job that runs at midnight every day
err = cron.DefineJob(
    gocron.CronJob(
        "0 0 * * *",
        gocron.WithLocation(time.UTC),
    ),
    func() {
        fmt.Println("Running daily at midnight")
    },
)
if err != nil {
    // Handle error
}
```

### Job Definition Options

The `gocron` library provides several ways to define jobs:

#### Duration Jobs

```go
// Run every 30 minutes
err := cron.DefineJob(
    gocron.DurationJob(
        30 * time.Minute,
    ),
    func() {
        // Task logic
    },
)
```

#### Cron Expression Jobs

```go
// Run at 10:30 AM every weekday
err := cron.DefineJob(
    gocron.CronJob(
        "30 10 * * 1-5",
        gocron.WithLocation(time.UTC),
    ),
    func() {
        // Task logic
    },
)
```

#### One-time Jobs

```go
// Run once after 5 minutes
err := cron.DefineJob(
    gocron.OneTimeJob(
        time.Now().Add(5 * time.Minute),
    ),
    func() {
        // Task logic
    },
)
```

### Cron Expression Format

The cron module uses the standard cron expression format:

```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of the month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday)
│ │ │ │ │
│ │ │ │ │
* * * * *
```

Common cron expressions:
- `* * * * *`: Every minute
- `0 * * * *`: Every hour at minute 0
- `*/5 * * * *`: Every 5 minutes
- `0 0 * * *`: Every day at midnight
- `0 0 * * 0`: Every Sunday at midnight
- `0 0 1 * *`: First day of every month at midnight
- `30 10 * * 1-5`: At 10:30 AM every weekday

### Real-world Examples

#### Database Cleanup Task

```go
// Clean up expired sessions every day at 2 AM
err := cron.DefineJob(
    gocron.CronJob(
        "0 2 * * *",
        gocron.WithLocation(time.UTC),
    ),
    func() {
        // Connect to database
        db := goe.UseDB()
        
        // Delete expired sessions
        filter := bson.M{"expires_at": bson.M{"$lt": time.Now()}}
        result, err := db.DeleteMany(&Session{}, filter)
        if err != nil {
            goe.UseLog().Error("Failed to clean up sessions:", err)
            return
        }
        
        goe.UseLog().Infof("Cleaned up %d expired sessions", result.DeletedCount)
    },
)
```

#### Periodic Report Generation

```go
// Generate weekly report every Monday at 7 AM
err := cron.DefineJob(
    gocron.CronJob(
        "0 7 * * 1",
        gocron.WithLocation(time.UTC),
    ),
    func() {
        // Generate report
        report, err := generateWeeklyReport()
        if err != nil {
            goe.UseLog().Error("Failed to generate weekly report:", err)
            return
        }
        
        // Send report via email
        mailer := goe.UseMailer()
        err = mailer.DefaultSender().
            To(&[]*mail.Address{{Name: "Admin", Address: "admin@example.com"}}).
            Subject("Weekly Report").
            HTML(report).
            Send()
        if err != nil {
            goe.UseLog().Error("Failed to send weekly report:", err)
            return
        }
        
        goe.UseLog().Info("Weekly report sent successfully")
    },
)
```

## API Reference

The Cron module implements the [`contracts.CronJob`](https://github.com/oeasenet/goe/blob/main/contracts/cron.go) interface:

```go
type CronJob interface {
    // DefineJob schedules a new job with the given definition and handler
    DefineJob(definition gocron.JobDefinition, handler func()) error
}
```

The [`CronJobModule`](https://github.com/oeasenet/goe/blob/main/modules/cron/cron.go) implementation also provides these additional methods:

```go
// Start starts the scheduler
Start()

// Close stops the scheduler and releases resources
Close() error
```

## Implementation Details

The cron module is implemented in the [`CronJobModule`](https://github.com/oeasenet/goe/blob/main/modules/cron/cron.go) struct, which provides:

- A wrapper around the [gocron](https://github.com/go-co-op/gocron) library
- Automatic starting of the scheduler when the application starts
- Graceful shutdown when the application exits
- Thread-safe operations

### Scheduler Configuration

The scheduler is configured with the following options:

- **LimitConcurrentJobs**: Limits the number of jobs that can run concurrently to 1
- **LimitModeWait**: When the concurrent job limit is reached, new jobs will wait until a slot is available

### Lifecycle Management

The cron module is automatically managed by the GOE framework:

1. The module is initialized when the application starts
2. Jobs can be defined before the application starts running
3. The scheduler is automatically started when the application starts
4. The scheduler is gracefully shut down when the application exits, allowing running jobs to complete

### Error Handling

The module handles errors gracefully:

- If a job definition is invalid, an error is returned
- If a job is defined after the scheduler has started, an error is returned
- If the scheduler fails to start or stop, an error is returned