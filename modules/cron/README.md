# Cron Module

The Cron module provides a simple interface for scheduling and running tasks at specified intervals. It implements the `contracts.CronJob` interface.

## Features

- Schedule tasks using cron expressions
- Run tasks at specified intervals
- Automatic task management
- Graceful shutdown support

## Usage

### Initialization

The cron module is automatically initialized by the GOE framework:

```go
// Get the cron service
cron := goe.UseCron()
```

### Scheduling Tasks

```go
// Add a job that runs every hour
cron.AddJob("0 * * * *", func() {
    // Task logic here
    fmt.Println("Running hourly task")
})

// Add a job that runs every 5 minutes
cron.AddJob("*/5 * * * *", func() {
    // Task logic here
    fmt.Println("Running every 5 minutes")
})

// Add a job that runs at midnight every day
cron.AddJob("0 0 * * *", func() {
    // Task logic here
    fmt.Println("Running daily at midnight")
})
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

## Implementation Details

The cron module uses a background goroutine to check for scheduled tasks and execute them at the appropriate times. It supports graceful shutdown, ensuring that running tasks are completed before the application exits.

The module is automatically started when the GOE application starts and is stopped when the application shuts down.