# Log Module

The Log module provides a structured logging system for the GOE framework. It implements the [`contracts.Logger`](https://github.com/oeasenet/goe/blob/main/contracts/log.go) interface and is built on top of the [Zap](https://github.com/uber-go/zap) logging library, which is known for its high performance and flexibility.

## Features

- Structured logging with key-value pairs
- Multiple log levels (Debug, Info, Warn, Error, Fatal, Panic)
- Development and production modes with appropriate formatting
- Caller information (file and line number) for easy debugging
- Colored output in development mode
- JSON output in production mode for machine parsing
- Formatted logging with support for variables
- Low memory allocation and high performance
- Integration with Zap for advanced use cases

## Usage

### Initialization

The log module is automatically initialized by the GOE framework:

```go
// Get the logger instance
logger := goe.UseLog()
```

If you need to create a logger instance directly:

```go
// Create a development logger with colored output
logger := log.New(log.LevelDev)

// Create a production logger with JSON output
logger := log.New(log.LevelProd)
```

### Basic Logging

```go
// Log at different levels
logger.Debug("This is a debug message")
logger.Info("This is an info message")
logger.Warn("This is a warning message")
logger.Error("This is an error message")
logger.Fatal("This is a fatal message") // Also exits the application
logger.Panic("This is a panic message") // Also panics

// Log with additional fields
logger.With("user_id", "123").Info("User logged in")
logger.With("request_id", "abc123").With("user_id", "456").Info("Request processed")
```

### Logging Errors

```go
// Log an error
err := someFunction()
if err != nil {
    logger.Error("Failed to execute function", err)
}

// Log an error with additional context
logger.With("function", "someFunction").Error("Operation failed", err)
```

### Formatted Logging

```go
// Log with formatting
logger.Debugf("User %s logged in from %s", username, ipAddress)
logger.Infof("Processing request #%d", requestID)
logger.Warnf("Rate limit reached: %d requests in %d seconds", count, seconds)
logger.Errorf("Failed to connect to %s: %v", serverName, err)
```

### Structured Logging

Structured logging is a powerful feature that allows you to add key-value pairs to your log messages, making them easier to filter and analyze:

```go
// Log with structured fields
logger.Infow("User logged in", 
    "user_id", "123", 
    "ip_address", "192.168.1.1", 
    "login_time", time.Now(),
)

// Log an error with structured fields
err := someFunction()
if err != nil {
    logger.Errorw("Operation failed", 
        "error", err, 
        "operation", "someFunction", 
        "retry_count", 3,
    )
}
```

### Real-world Examples

#### HTTP Request Logging

```go
func logRequest(ctx fiber.Ctx) error {
    startTime := time.Now()
    
    // Process the request
    err := ctx.Next()
    
    // Log the request details
    logger := goe.UseLog()
    logger.Infow("HTTP Request",
        "method", ctx.Method(),
        "path", ctx.Path(),
        "status", ctx.Response().StatusCode(),
        "duration_ms", time.Since(startTime).Milliseconds(),
        "ip", ctx.IP(),
        "user_agent", ctx.Get("User-Agent"),
    )
    
    return err
}
```

#### Database Operation Logging

```go
func getUserById(id string) (*User, error) {
    logger := goe.UseLog()
    
    logger.Debugf("Fetching user with ID: %s", id)
    
    user, err := db.FindById(&User{}, id, &User{})
    if err != nil {
        logger.Errorw("Failed to fetch user",
            "user_id", id,
            "error", err,
        )
        return nil, err
    }
    
    if !user {
        logger.Warnf("User not found: %s", id)
        return nil, nil
    }
    
    logger.Debugf("Successfully fetched user: %s", id)
    return user, nil
}
```

### Log Levels

The log module supports the following log levels, in order of increasing severity:

1. **Debug**: Detailed information, typically useful only when diagnosing problems
2. **Info**: Confirmation that things are working as expected
3. **Warn**: Indication that something unexpected happened, but the application can continue
4. **Error**: Due to a more serious problem, the application has not been able to perform a function
5. **Fatal**: Very severe error events that will lead the application to abort
6. **Panic**: Severe error events that cause the application to panic

### Accessing the Underlying Zap Logger

If you need access to the underlying Zap logger for advanced use cases:

```go
// Get the Zap logger
zapLogger := logger.GetZapLogger()

// Get the Zap sugared logger
zapSugar := logger.GetZapSugarLogger()
```

## API Reference

The Log module implements the [`contracts.Logger`](https://github.com/oeasenet/goe/blob/main/contracts/log.go) interface:

```go
type Logger interface {
    // Basic logging methods
    Debug(args ...any)
    Log(args ...any)
    Info(args ...any)
    Warn(args ...any)
    Error(args ...any)
    Fatal(args ...any)
    Panic(args ...any)

    // Formatted logging methods
    Debugf(format string, args ...any)
    Logf(format string, args ...any)
    Infof(format string, args ...any)
    Warnf(format string, args ...any)
    Errorf(format string, args ...any)
    Fatalf(format string, args ...any)
    Panicf(format string, args ...any)

    // Structured logging methods
    Debugw(msg string, keysAndValues ...any)
    Infow(msg string, keysAndValues ...any)
    Warnw(msg string, keysAndValues ...any)
    Errorw(msg string, keysAndValues ...any)
    Fatalw(msg string, keysAndValues ...any)
    Panicw(msg string, keysAndValues ...any)

    // Utility methods
    GetZapLogger() *zap.Logger
    GetZapSugarLogger() *zap.SugaredLogger
    Close()
}
```

## Implementation Details

The log module is implemented in the [`Log`](https://github.com/oeasenet/goe/blob/main/modules/log/handler.go) struct, which provides:

- A wrapper around the Zap logger and sugared logger
- Configuration for development and production environments
- Automatic caller information for easy debugging

### Development vs. Production Mode

The module configures Zap differently based on the environment:

- **Development Mode** (`log.LevelDev`):
  - Console-friendly, human-readable output
  - Colored level indicators for better visibility
  - Stack traces for errors
  - Caller information (file and line number)
  - More verbose output

- **Production Mode** (`log.LevelProd`):
  - JSON-formatted output for machine parsing
  - Optimized for performance
  - Lower verbosity
  - No colors
  - Minimal stack traces

### Performance Considerations

The Zap logging library is designed for high performance with minimal allocations:

- No allocations for common operations
- Efficient encoding of structured data
- Minimal CPU overhead
- Asynchronous writing for better throughput

### Closing the Logger

It's good practice to close the logger when your application shuts down to ensure all buffered logs are written:

```go
// Close the logger
logger.Close()
```

In the GOE framework, this is handled automatically during graceful shutdown.