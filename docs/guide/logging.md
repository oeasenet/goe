# Logging

GOE provides structured logging capabilities built on top of Uber's Zap library. The logging system is designed to be fast, structured, and configurable for different environments.

## Overview

The logging module in GOE:
- Uses Uber's Zap for high-performance structured logging
- Supports multiple output formats (console, JSON)
- Provides different log levels (debug, info, warn, error, fatal)
- Integrates with the application lifecycle
- Includes request tracing capabilities

## Configuration

Logging is configured through environment variables:

```bash
# .env file
LOG_LEVEL=info          # debug, info, warn, error, fatal
LOG_FORMAT=console      # console, json
LOG_OUTPUT=stdout       # stdout, stderr, file path
```

### Environment-Specific Configuration

```bash
# Development
LOG_LEVEL=debug
LOG_FORMAT=console

# Production
LOG_LEVEL=warn
LOG_FORMAT=json
LOG_OUTPUT=/var/log/app.log
```

## Accessing the Logger

### Using Global Accessor

```go
import "go.oease.dev/goe/v2"

func someFunction() {
    logger := goe.Log()
    
    // Basic logging
    logger.Info("Application started")
    logger.Error("Something went wrong")
    
    // Structured logging with key-value pairs
    logger.Info("User logged in", "user_id", 123, "email", "user@example.com")
    logger.Error("Database connection failed", "host", "localhost", "port", 5432)
}
```

### Using Dependency Injection

```go
import "go.oease.dev/goe/v2/contract"

type UserService struct {
    logger contract.Logger
}

func NewUserService(logger contract.Logger) *UserService {
    return &UserService{logger: logger}
}

func (s *UserService) CreateUser(name, email string) error {
    s.logger.Info("Creating user", "name", name, "email", email)
    
    // Business logic here
    
    s.logger.Info("User created successfully", "name", name)
    return nil
}
```

## Log Levels

GOE supports standard log levels:

```go
logger := goe.Log()

// Debug - detailed information for diagnosing problems
logger.Debug("Debugging information", "variable", value)

// Info - general information about program execution
logger.Info("Application started successfully")

// Warn - potentially harmful situations
logger.Warn("Configuration value not set, using default", "key", "TIMEOUT")

// Error - error events that might still allow the application to continue
logger.Error("Failed to process request", "error", err)

// Fatal - severe error events that will presumably lead the application to abort
logger.Fatal("Cannot connect to database", "error", err) // This will exit the program
```

## Structured Logging

Use key-value pairs for structured logging:

```go
logger := goe.Log()

// Simple key-value pairs
logger.Info("User action", "user_id", 123, "action", "login")

// Multiple fields
logger.Info("HTTP request",
    "method", "GET",
    "path", "/api/users",
    "status", 200,
    "duration", "45ms",
    "ip", "192.168.1.1",
)

// Complex objects
user := User{ID: 123, Name: "John Doe"}
logger.Info("User created", "user", user)
```

## HTTP Request Logging

GOE automatically logs HTTP requests when the HTTP module is enabled:

```go
// Automatic request logging includes:
// - Request method and path
// - Response status code
// - Request duration
// - Request ID (for tracing)
// - IP address

// Example log output:
// INFO HTTP Request {"method": "GET", "path": "/api/users", "status": 200, "duration": "23ms", "request_id": "abc123", "ip": "127.0.0.1"}
```

### Custom Request Logging

You can add custom logging in your handlers:

```go
import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2/contract"
)

func NewUserHandler(logger contract.Logger) *UserHandler {
    return &UserHandler{logger: logger}
}

type UserHandler struct {
    logger contract.Logger
}

func (h *UserHandler) CreateUser(c fiber.Ctx) error {
    h.logger.Info("Creating user endpoint called", "ip", c.IP())
    
    var user User
    if err := c.BodyParser(&user); err != nil {
        h.logger.Error("Failed to parse request body", "error", err)
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    h.logger.Info("User creation successful", "user_id", user.ID)
    return c.JSON(user)
}
```

## Error Logging

Proper error logging with context:

```go
func (s *UserService) GetUser(id int) (*User, error) {
    s.logger.Debug("Fetching user", "user_id", id)
    
    user, err := s.repository.FindUser(id)
    if err != nil {
        s.logger.Error("Failed to fetch user from database",
            "user_id", id,
            "error", err,
            "operation", "repository.FindUser",
        )
        return nil, err
    }
    
    s.logger.Info("User fetched successfully", "user_id", id)
    return user, nil
}
```

## Logging with Context

Using context for request tracing:

```go
import "context"

func (s *UserService) ProcessUser(ctx context.Context, userID int) error {
    // Extract request ID from context (if available)
    requestID := ctx.Value("request_id")
    
    s.logger.Info("Processing user",
        "user_id", userID,
        "request_id", requestID,
    )
    
    // Process user...
    
    s.logger.Info("User processed successfully",
        "user_id", userID,
        "request_id", requestID,
    )
    
    return nil
}
```

## Testing with Logging

Mock logger for testing:

```go
type MockLogger struct {
    logs []LogEntry
}

type LogEntry struct {
    Level   string
    Message string
    Fields  map[string]interface{}
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
    m.logs = append(m.logs, LogEntry{
        Level:   "info",
        Message: msg,
        Fields:  parseFields(keysAndValues),
    })
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
    m.logs = append(m.logs, LogEntry{
        Level:   "error",
        Message: msg,
        Fields:  parseFields(keysAndValues),
    })
}

func TestUserService(t *testing.T) {
    mockLogger := &MockLogger{}
    service := NewUserService(mockLogger)
    
    service.CreateUser("John", "john@example.com")
    
    // Verify logging
    assert.Len(t, mockLogger.logs, 2)
    assert.Equal(t, "Creating user", mockLogger.logs[0].Message)
    assert.Equal(t, "John", mockLogger.logs[0].Fields["name"])
}
```

## Best Practices

### 1. Use Structured Logging

```go
// Good - structured with key-value pairs
logger.Info("User login", "user_id", 123, "email", "user@example.com")

// Bad - unstructured string interpolation
logger.Info(fmt.Sprintf("User %d logged in with email %s", 123, "user@example.com"))
```

### 2. Choose Appropriate Log Levels

```go
// Debug - for development debugging
logger.Debug("Processing item", "item_id", id)

// Info - for general information
logger.Info("Server started", "port", 8080)

// Warn - for potential issues
logger.Warn("Rate limit approaching", "current", 95, "limit", 100)

// Error - for error conditions
logger.Error("Database query failed", "error", err)
```

### 3. Include Context Information

```go
logger.Info("Processing order",
    "order_id", order.ID,
    "user_id", order.UserID,
    "amount", order.Amount,
    "currency", order.Currency,
)
```

### 4. Don't Log Sensitive Information

```go
// Good - don't log passwords, tokens, etc.
logger.Info("User authentication", "user_id", user.ID)

// Bad - logging sensitive data
logger.Info("User login", "password", password, "token", token)
```

### 5. Use Consistent Field Names

```go
// Consistent field naming
logger.Info("User created", "user_id", user.ID)
logger.Info("User updated", "user_id", user.ID)
logger.Info("User deleted", "user_id", user.ID)
```

## Performance Considerations

1. **Zap is Fast**: Uber's Zap is designed for high-performance logging
2. **Structured Logging**: Key-value pairs are more efficient than string formatting
3. **Log Levels**: Use appropriate log levels to avoid unnecessary logging in production
4. **Buffering**: Consider log buffering for high-throughput applications

## Production Logging

### JSON Format for Production

```bash
# Production configuration
LOG_FORMAT=json
LOG_LEVEL=warn
LOG_OUTPUT=/var/log/app.log
```

### Log Rotation

Consider using log rotation tools:

```bash
# Using logrotate
/var/log/app.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 app app
}
```

## Next Steps

- [**HTTP Server**](./http-server.md) - Learn about HTTP request handling
- [**Database**](./database.md) - Database operations and logging
