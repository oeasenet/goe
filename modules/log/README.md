# Log Module

The Log module provides a structured logging system for the GOE framework. It implements the `contracts.Logger` interface and is built on top of the [Zap](https://github.com/uber-go/zap) logging library.

## Features

- Structured logging
- Multiple log levels (Debug, Info, Warn, Error, Fatal, Panic)
- Development and production modes
- Caller information (file and line number)
- Formatted logging with support for variables

## Usage

### Initialization

The log module is automatically initialized by the GOE framework:

```go
// Get the logger instance
logger := goe.UseLog()
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

### Log Levels

The log module supports the following log levels, in order of increasing severity:

1. **Debug**: Detailed information, typically useful only when diagnosing problems
2. **Info**: Confirmation that things are working as expected
3. **Warn**: Indication that something unexpected happened, but the application can continue
4. **Error**: Due to a more serious problem, the application has not been able to perform a function
5. **Fatal**: Very severe error events that will lead the application to abort
6. **Panic**: Severe error events that cause the application to panic

## Implementation Details

The log module uses Zap for high-performance, structured logging. It configures Zap differently based on the environment:

- In development mode, logs are formatted for human readability and include caller information
- In production mode, logs are JSON-formatted for machine parsing and optimized for performance

The module provides a simple, consistent interface for logging across the entire application.