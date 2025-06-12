# 6. Logging with Goe 📝

Effective logging is crucial for understanding application behavior, debugging issues, and monitoring performance. Goe integrates [Uber's Zap logger](https://github.com/uber-go/zap), a blazing fast, structured logging library for Go.

## Introduction to Goe's Logging

Goe's logging module provides:

*   **High-Performance Logging**: Leveraging Zap's speed and efficiency.
*   **Structured Logging**: Logs are emitted with key-value pairs, making them easy to parse, search, and analyze, especially in log management systems.
*   **Multiple Log Levels**: Supports standard levels like Debug, Info, Warn, Error, and Fatal.
*   **Configurable Formatting**:
    *   **Text (Pretty Console)**: For development, logs are colorized and human-readable.
    *   **JSON**: For production, logs are typically formatted as JSON for machine readability and compatibility with log aggregators.
*   **Configurable Output**: Logs can be directed to the console, files, or multiple destinations.
*   **Caller Information**: Optionally include file and line number where the log message originated.
*   **Stack Traces**: Optionally include stack traces for error-level logs.

## Configuration

The logger is configured through environment variables, typically prefixed with `LOG_`. Refer to the [Configuration](05-configuration.md#common-configuration-keys) guide for general setup.

Key logging configuration options:

*   **`LOG_LEVEL`**: Sets the minimum log level. Messages below this level will not be printed.
    *   Values: `debug`, `info`, `warn`, `error`.
    *   Default: `info`.
    *   Example: `LOG_LEVEL=debug`

*   **`LOG_FORMAT`**: Determines the log output format.
    *   Values:
        *   `text`: Human-readable, colorized output (default for `GOE_ENV=dev` or if `GOE_ENV` is not set).
        *   `json`: Structured JSON output (default for `GOE_ENV=prod` or `GOE_ENV=production`).
    *   Example: `LOG_FORMAT=json`

*   **`LOG_OUTPUT`**: Comma-separated list of output destinations.
    *   Values: `console`, `file`.
    *   Default: `console`.
    *   If `file` is included, logs are written to `app.log` in the application's root directory.
    *   Example: `LOG_OUTPUT=console,file`

*   **`LOG_CALLER`**: Whether to include caller information (file and line number) in logs.
    *   Values: `true`, `false`.
    *   Default: `true`.
    *   Example: `LOG_CALLER=false`

*   **`LOG_STACKTRACE`**: Whether to automatically include stack traces for logs at `Error` level and above.
    *   Values: `true`, `false`.
    *   Default: `false`.
    *   Example: `LOG_STACKTRACE=true`

Goe determines if the environment is "production" based on `GOE_ENV` being `prod` or `production`. If so, and `LOG_FORMAT` is not explicitly set, it defaults to `json`. Otherwise, it defaults to `text`.

## Using the Logger

Similar to other Goe components, you can access the logger via a global helper or through dependency injection.

### 1. Global Accessor (`goe.Log()`)

The simplest way to get the logger instance:

```go
package main

import (
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract" // For contract.Field
)

func main() {
	_ = goe.New(goe.Options{}) // Initialize Goe

	goe.Log().Info("Application started successfully!",
		contract.NewField("user_id", 123),
		contract.NewField("status", "running"),
	)

	goe.Log().Debug("This is a debug message. Might not show if LOG_LEVEL=info.")
	goe.Log().Warn("Something looks a bit off here.", contract.NewField("warning_code", "W001"))

	err := MyAppError{"Something failed"}
	goe.Log().Error("An error occurred",
		contract.NewField("error", err.Error()),
		contract.NewField("component", "payment_processor"),
	)

	// Fatal will log and then os.Exit(1)
	// goe.Log().Fatal("A critical error occurred, shutting down.", contract.NewField("reason", "unrecoverable_state"))
}

type MyAppError struct{ Message string }
func (e MyAppError) Error() string { return e.Message }
```

### 2. Dependency Injection (`contract.Logger`)

For services, handlers, and other components, injecting `contract.Logger` is recommended for better testability and explicit dependencies.

```go
package mymodule

import (
	"go.oease.dev/goe/v2/contract"
)

type MyService struct {
	logger contract.Logger
}

// NewMyService is an Fx provider
func NewMyService(logger contract.Logger) *MyService {
	return &MyService{logger: logger}
}

func (s *MyService) DoSomething(taskID string) {
	s.logger.Info("Starting task...",
		contract.NewField("task_id", taskID),
		contract.NewField("service_name", "MyService"),
	)

	// ... do something ...

	if err := s.processSubTask(); err != nil {
		s.logger.Error("Failed to process sub-task",
			contract.NewField("task_id", taskID),
			contract.NewField("error", err), // Zap handles error types well
		)
		return
	}

	s.logger.Info("Task completed successfully.", contract.NewField("task_id", taskID))
}

func (s *MyService) processSubTask() error {
	// ...
	return nil // or an error
}
```
You would then provide `NewMyService` to Fx: `fx.Provide(mymodule.NewMyService)`.

## Log Levels and Methods

The `contract.Logger` interface provides methods for different log severities:

*   **`Debug(msg string, fields ...contract.Field)`**: For detailed information, typically useful only during debugging.
*   **`Info(msg string, fields ...contract.Field)`**: General information about application operation (e.g., service started, request processed).
*   **`Warn(msg string, fields ...contract.Field)`**: Indicates a potential problem or an unusual situation that isn't necessarily an error yet.
*   **`Error(msg string, fields ...contract.Field)`**: An error occurred during an operation. The application might still be able to continue.
*   **`Fatal(msg string, fields ...contract.Field)`**: A critical error that prevents the application from continuing. Calling `Fatal` will log the message and then terminate the application with `os.Exit(1)`.

## Structured Logging with Fields

Goe (via Zap) encourages structured logging by adding key-value pairs (fields) to your log messages. This makes logs much more powerful for querying and analysis.

Create fields using `contract.NewField(key string, value any)`:

```go
userID := 101
productID := "prod_abc"
goe.Log().Info("User viewed product",
	contract.NewField("user_id", userID),
	contract.NewField("product_id", productID),
	contract.NewField("action", "view"),
)
```

**Output (JSON format example):**
```json
{"level":"info","timestamp":"2023-10-27T10:30:00.123Z","caller":"myapp/main.go:15","msg":"User viewed product","user_id":101,"product_id":"prod_abc","action":"view"}
```

**Output (Text format example):**
```
INFO  T=2023-10-27 10:30:00.123 C=myapp/main.go:15  M=User viewed product  user_id=101 product_id=prod_abc action=view
```

Zap handles various data types for field values automatically (strings, numbers, booleans, errors, time.Time, etc.).

## Contextual Logging with `With()`

Sometimes you want to add a set of fields to multiple log messages within a certain context (e.g., all logs related to a specific request or user session). The `With(fields ...contract.Field)` method creates a new logger instance with these fields pre-populated.

```go
requestID := "req_xyz789"
contextualLogger := goe.Log().With(
	contract.NewField("request_id", requestID),
	contract.NewField("component", "OrderProcessor"),
)

contextualLogger.Info("Processing new order.")
// ... later ...
contextualLogger.Debug("Order details fetched.", contract.NewField("item_count", 5))
// ... if error ...
contextualLogger.Error("Failed to update inventory.", contract.NewField("error_code", "INV_003"))
```
All messages logged via `contextualLogger` will automatically include `request_id` and `component` fields.

## Logging `error` Types

When logging errors, you can pass the `error` type directly as a field value. Zap will typically serialize it appropriately.

```go
err := errors.New("something went wrong")
goe.Log().Error("Operation failed", contract.NewField("error", err))
```
If `LOG_STACKTRACE=true` is set, error logs will also include a stack trace.

The `WithError(err error)` method is a convenient shortcut:
```go
err := errors.New("connection timed out")
loggerWithError := goe.Log().WithError(err)
loggerWithError.Error("Failed to connect to payment gateway")
// This is equivalent to: goe.Log().Error("...", contract.NewField("error", err))
```

## Logging in HTTP Handlers

When using Goe's HTTP module, a logger instance (often with a `request_id`) is made available in the `fiber.Ctx`. You can access it using `goe.HTTP().GetLogger(c)` or `http.GetLogger(c)` if you import the `go.oease.dev/goe/v2/core/http` package.

```go
package main

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	goehttp "go.oease.dev/goe/v2/core/http" // Alias for clarity
)

func main() {
	_ = goe.New(goe.Options{
		WithHTTP: true,
		Invokers: []any{SetupRoutes},
	})
	goe.Run()
}

func SetupRoutes(kernel contract.HTTPKernel) {
	app := kernel.App()
	app.Get("/test", func(c fiber.Ctx) error {
		// Access logger from Fiber context (it's pre-loaded with request_id)
		logger := goehttp.GetLogger(c)

		logger.Info("Handler for /test executed.", contract.NewField("custom_data", "some_value"))
		return c.SendString("Test handler logged.")
	})
}
```
This logger is already contextualized with fields like `request_id` from the HTTP middleware.

## Underlying Zap Logger

If you need to access the underlying `*zap.SugaredLogger` (e.g., for compatibility with libraries expecting Zap), you can use the `GetLogger()` method from the `contract.Logger` interface:

```go
sugaredLogger := goe.Log().GetLogger()
sugaredLogger.Infow("Message from sugared logger", "key1", "value1")
```
There's also `log.GetZapLogger(contract.Logger)` in `core/log/log.go` to get the `*zap.Logger`.

## Best Practices for Logging

*   **Log at the Right Level**: Don't overuse `Error` for non-critical issues. Use `Debug` for verbose logs only needed during development.
*   **Be Structured**: Always prefer adding structured fields over formatting data directly into the log message string. This makes your logs searchable and analyzable.
    *   Bad: `logger.Info(fmt.Sprintf("User %d ordered product %s", userID, productID))`
    *   Good: `logger.Info("User ordered product", contract.NewField("user_id", userID), contract.NewField("product_id", productID))`
*   **Provide Context**: Use `With()` to add relevant contextual information (e.g., request ID, user ID, session ID) to your logs.
*   **Don't Log Secrets**: Be careful not to log sensitive information like passwords, API keys, or personal data unless absolutely necessary and properly secured/anonymized.
*   **Consistent Field Names**: Establish conventions for common field names (e.g., `user_id`, `trace_id`, `error_code`) across your application.
*   **Log Errors, Don't Just Print Them**: When you handle an error, log it with relevant context.
*   **Consider Log Volume**: Be mindful of logging too much in production, as it can impact performance and cost (for log storage/management). Adjust log levels appropriately.

By following these practices, you can make your Goe application's logs a powerful tool for development, debugging, and operations.

Next up: Building web applications with the [HTTP Server](07-http-server.md).
```
