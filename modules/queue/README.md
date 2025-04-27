# Queue Module

The Queue module provides a Redis-based message queue system for the GOE framework. It implements the `contracts.Queue` interface and is built on top of the [Delayqueue](https://github.com/HDT3213/delayqueue) library.

## Features

- Asynchronous task processing
- Delayed task execution
- Reliable message delivery
- Automatic retries for failed tasks
- Concurrent worker support
- Structured logging of queue operations
- Graceful shutdown support

## Usage

### Initialization

The queue module is automatically initialized by the GOE framework:

```
# Queue Configuration
QUEUE_CONCURRENCY=4           # Number of concurrent workers
QUEUE_FETCH_INTERVAL=1        # Interval in seconds to fetch new tasks
QUEUE_FETCH_LIMIT=10          # Maximum number of tasks to fetch at once
QUEUE_MAX_CONSUME_DURATION=5  # Maximum time in seconds to process a task
QUEUE_DEFAULT_RETRIES=3       # Default number of retries for failed tasks
```

### Publishing Tasks

```go
// Get the queue instance
queue := goe.UseMQ()

// Publish a task to a queue
err := queue.Publish("email", map[string]interface{}{
    "to":      "user@example.com",
    "subject": "Welcome to GOE",
    "body":    "Thank you for using GOE framework!",
}, 0) // 0 means execute immediately

// Publish a delayed task (execute after 60 seconds)
err := queue.Publish("notification", map[string]interface{}{
    "user_id":  "123",
    "message":  "Your report is ready",
    "priority": "high",
}, 60)

// Publish a task with custom options
err := queue.PublishWithOptions("report", map[string]interface{}{
    "report_id": "abc123",
    "format":    "pdf",
}, &contracts.QueueOptions{
    Delay:   120,    // 2 minutes delay
    Retries: 5,      // 5 retries
    TTR:     300,    // 5 minutes to process
})
```

### Subscribing to Queues

```go
// Subscribe to a queue
queue.Subscribe("email", func(payload []byte) error {
    // Parse the payload
    var data map[string]interface{}
    err := json.Unmarshal(payload, &data)
    if err != nil {
        return err
    }

    // Process the task
    to := data["to"].(string)
    subject := data["subject"].(string)
    body := data["body"].(string)

    // Send the email
    fmt.Printf("Sending email to %s with subject: %s\n", to, subject)

    // Return nil for success, or an error to trigger a retry
    return nil
})

// Subscribe to multiple queues
queue.Subscribe("notification", handleNotification)
queue.Subscribe("report", handleReport)
```

### Error Handling and Retries

When a task handler returns an error, the task is automatically retried according to the retry configuration:

```go
queue.Subscribe("import", func(payload []byte) error {
    // Simulate a temporary failure
    if rand.Intn(10) < 3 {
        return errors.New("temporary failure, will retry")
    }

    // Process the task
    fmt.Println("Processing import task")
    return nil
})
```

### Monitoring Queue Status

```go
// Get queue statistics
stats, err := queue.Stats()
fmt.Printf("Queue stats: %+v\n", stats)

// Get information about a specific queue
queueInfo, err := queue.QueueInfo("email")
fmt.Printf("Email queue: %+v\n", queueInfo)
```

## Implementation Details

The queue module uses Redis as the backend storage for the message queue. It provides:

- Reliable task delivery with at-least-once semantics
- Automatic retries for failed tasks
- Delayed task execution
- Concurrent task processing
- Graceful shutdown, ensuring that in-progress tasks are completed

The module is designed to be easy to use while providing the reliability needed for production applications.

## Acknowledgments

This module is built on top of the [Delayqueue](https://github.com/HDT3213/delayqueue) library, which provides the core functionality for reliable message queuing with Redis.
