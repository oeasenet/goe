# Goe API Reference

This document provides a comprehensive reference for all public APIs in the Goe framework.

## Table of Contents

- [Core APIs](#core-apis)
  - [Application Management](#application-management)
  - [Configuration](#configuration)
  - [Logging](#logging)
  - [HTTP Server](#http-server)
- [Contracts](#contracts)
- [Types](#types)
- [Middleware](#middleware)
- [Utilities](#utilities)

## Core APIs

### Application Management

#### `goe.New(opts ...Options) contract.Application`

Creates a new Goe application instance.

**Parameters:**
- `opts`: Optional configuration options

**Returns:**
- Application instance

**Example:**
```go
app := goe.New(goe.Options{
    Name:        "My App",
    Version:     "1.0.0",
    Environment: "production",
    WithHTTP:    true,
})
```

#### `goe.Run()`

Starts the application and blocks until shutdown.

**Example:**
```go
goe.Run() // Blocks until SIGINT/SIGTERM
```

#### `goe.App() contract.Application`

Returns the global application instance.

**Panics:** If called before `goe.New()`

**Example:**
```go
app := goe.App()
fmt.Println(app.Name())     // "My App"
fmt.Println(app.Version())  // "1.0.0"
```

#### `goe.IsRunning() bool`

Checks if the application is currently running.

**Returns:**
- `true` if running, `false` otherwise

**Example:**
```go
if goe.IsRunning() {
    // Application is running
}
```

#### `goe.GetEnvironment() string`

Returns the current environment name.

**Returns:**
- Environment string (e.g., "dev", "staging", "production")

**Example:**
```go
env := goe.GetEnvironment() // "production"
```

#### `goe.Context() context.Context`

Returns the application's root context.

**Returns:**
- Application context

**Example:**
```go
ctx := goe.Context()
// Use for background tasks that should respect app lifecycle
```

### Configuration

#### `goe.Config() contract.Config`

Returns the global configuration instance.

**Panics:** If called before `goe.New()`

**Example:**
```go
config := goe.Config()
port := config.GetInt("HTTP_PORT")
```

#### Config Methods

##### `Get(key string) any`

Retrieves a configuration value as `any`.

**Parameters:**
- `key`: Configuration key

**Returns:**
- Configuration value or `nil`

**Example:**
```go
value := config.Get("MY_KEY")
```

##### `GetString(key string) string`

Retrieves a string configuration value.

**Parameters:**
- `key`: Configuration key

**Returns:**
- String value or empty string

**Example:**
```go
appName := config.GetString("APP_NAME")
```

##### `GetInt(key string) int`

Retrieves an integer configuration value.

**Parameters:**
- `key`: Configuration key

**Returns:**
- Integer value or 0

**Example:**
```go
port := config.GetInt("HTTP_PORT") // 8080
```

##### `GetBool(key string) bool`

Retrieves a boolean configuration value.

**Parameters:**
- `key`: Configuration key

**Returns:**
- Boolean value or `false`

**Example:**
```go
debug := config.GetBool("DEBUG") // true
```

##### `GetFloat64(key string) float64`

Retrieves a float64 configuration value.

**Parameters:**
- `key`: Configuration key

**Returns:**
- Float64 value or 0.0

**Example:**
```go
rate := config.GetFloat64("RATE_LIMIT") // 10.5
```

##### `GetDuration(key string) time.Duration`

Retrieves a duration configuration value.

**Parameters:**
- `key`: Configuration key

**Returns:**
- Duration value or 0

**Example:**
```go
timeout := config.GetDuration("REQUEST_TIMEOUT") // 30s
```

##### `GetStringSlice(key string) []string`

Retrieves a string slice configuration value.

**Parameters:**
- `key`: Configuration key

**Returns:**
- String slice or empty slice

**Example:**
```go
// ALLOWED_ORIGINS=http://localhost:3000,https://example.com
origins := config.GetStringSlice("ALLOWED_ORIGINS")
// ["http://localhost:3000", "https://example.com"]
```

##### `GetStringMap(key string) map[string]any`

Retrieves a string map configuration value.

**Parameters:**
- `key`: Configuration key

**Returns:**
- String map or empty map

**Example:**
```go
// DATABASE={"host":"localhost","port":5432}
dbConfig := config.GetStringMap("DATABASE")
// map[string]any{"host": "localhost", "port": 5432}
```

##### `Set(key string, value any)`

Sets a configuration value.

**Parameters:**
- `key`: Configuration key
- `value`: Configuration value

**Example:**
```go
config.Set("FEATURE_FLAG", true)
```

##### `Has(key string) bool`

Checks if a configuration key exists.

**Parameters:**
- `key`: Configuration key

**Returns:**
- `true` if exists, `false` otherwise

**Example:**
```go
if config.Has("API_KEY") {
    // API key is configured
}
```

##### `All() map[string]any`

Returns all configuration values.

**Returns:**
- Map of all configuration values

**Example:**
```go
allConfig := config.All()
for key, value := range allConfig {
    fmt.Printf("%s: %v\n", key, value)
}
```

### Logging

#### `goe.Log() contract.Logger`

Returns the global logger instance.

**Panics:** If called before `goe.New()`

**Example:**
```go
logger := goe.Log()
logger.Info("Application started")
```

#### Logger Methods

##### `Debug(msg string, fields ...Field)`

Logs a debug message.

**Parameters:**
- `msg`: Log message
- `fields`: Optional structured fields

**Example:**
```go
logger.Debug("Processing request",
    log.NewField("request_id", "123"),
    log.NewField("user_id", 456),
)
```

##### `Info(msg string, fields ...Field)`

Logs an info message.

**Parameters:**
- `msg`: Log message
- `fields`: Optional structured fields

**Example:**
```go
logger.Info("User logged in",
    log.NewField("user_id", userID),
    log.NewField("ip", request.IP),
)
```

##### `Warn(msg string, fields ...Field)`

Logs a warning message.

**Parameters:**
- `msg`: Log message
- `fields`: Optional structured fields

**Example:**
```go
logger.Warn("Rate limit approaching",
    log.NewField("current", 95),
    log.NewField("limit", 100),
)
```

##### `Error(msg string, fields ...Field)`

Logs an error message.

**Parameters:**
- `msg`: Log message
- `fields`: Optional structured fields

**Example:**
```go
logger.Error("Database connection failed",
    log.NewField("error", err.Error()),
    log.NewField("retry_count", retries),
)
```

##### `Fatal(msg string, fields ...Field)`

Logs a fatal message and exits the application.

**Parameters:**
- `msg`: Log message
- `fields`: Optional structured fields

**Example:**
```go
logger.Fatal("Critical configuration missing",
    log.NewField("config_key", "DATABASE_URL"),
)
```

##### `With(fields ...Field) Logger`

Creates a new logger with additional fields.

**Parameters:**
- `fields`: Fields to add to all logs

**Returns:**
- New logger instance with fields

**Example:**
```go
userLogger := logger.With(
    log.NewField("user_id", userID),
    log.NewField("session_id", sessionID),
)
userLogger.Info("Action performed") // Includes user_id and session_id
```

#### `log.NewField(key string, value any) Field`

Creates a new log field.

**Parameters:**
- `key`: Field key
- `value`: Field value

**Returns:**
- Field instance

**Example:**
```go
field := log.NewField("status", "active")
```

### HTTP Server

#### `goe.HTTP() contract.HTTPKernel`

Returns the global HTTP kernel instance.

**Panics:** If called before `goe.New()` with `WithHTTP: true`

**Example:**
```go
http := goe.HTTP()
app := http.App()
```

#### HTTP Context Helpers

##### `http.GetLogger(c fiber.Ctx) contract.Logger`

Gets a logger from the Fiber context with request ID.

**Parameters:**
- `c`: Fiber context

**Returns:**
- Logger instance with request context

**Example:**
```go
func handler(c fiber.Ctx) error {
    logger := http.GetLogger(c)
    logger.Info("Processing request") // Includes request_id
    return nil
}
```

##### `http.GetConfig(c fiber.Ctx) contract.Config`

Gets configuration from the Fiber context.

**Parameters:**
- `c`: Fiber context

**Returns:**
- Config instance

**Example:**
```go
func handler(c fiber.Ctx) error {
    config := http.GetConfig(c)
    appName := config.GetString("APP_NAME")
    return c.SendString(appName)
}
```

##### `http.GetApp(c fiber.Ctx) contract.Application`

Gets application instance from the Fiber context.

**Parameters:**
- `c`: Fiber context

**Returns:**
- Application instance

**Example:**
```go
func handler(c fiber.Ctx) error {
    app := http.GetApp(c)
    version := app.Version()
    return c.JSON(fiber.Map{"version": version})
}
```

##### `http.GetServices(c fiber.Ctx) Services`

Gets all services from the Fiber context.

**Parameters:**
- `c`: Fiber context

**Returns:**
- Services struct

**Example:**
```go
func handler(c fiber.Ctx) error {
    services := http.GetServices(c)
    services.Logger.Info("Got services")
    return nil
}
```

## Contracts

### Application

```go
type Application interface {
    // Name returns the application name
    Name() string
    
    // Version returns the application version
    Version() string
    
    // Environment returns the current environment
    Environment() string
    
    // Context returns the application context
    Context() context.Context
    
    // IsRunning checks if the application is running
    IsRunning() bool
    
    // Container returns the Fx container
    Container() *fx.App
    
    // Register registers Fx options
    Register(opts ...fx.Option) error
}
```

### Config

```go
type Config interface {
    // Get retrieves a value by key
    Get(key string) any
    
    // GetString retrieves a string value
    GetString(key string) string
    
    // GetInt retrieves an integer value
    GetInt(key string) int
    
    // GetBool retrieves a boolean value
    GetBool(key string) bool
    
    // GetFloat64 retrieves a float64 value
    GetFloat64(key string) float64
    
    // GetDuration retrieves a duration value
    GetDuration(key string) time.Duration
    
    // GetStringSlice retrieves a string slice
    GetStringSlice(key string) []string
    
    // GetStringMap retrieves a string map
    GetStringMap(key string) map[string]any
    
    // Set sets a configuration value
    Set(key string, value any)
    
    // Has checks if a key exists
    Has(key string) bool
    
    // All returns all configuration
    All() map[string]any
}
```

### Logger

```go
type Logger interface {
    // Debug logs a debug message
    Debug(msg string, fields ...Field)
    
    // Info logs an info message
    Info(msg string, fields ...Field)
    
    // Warn logs a warning message
    Warn(msg string, fields ...Field)
    
    // Error logs an error message
    Error(msg string, fields ...Field)
    
    // Fatal logs a fatal message and exits
    Fatal(msg string, fields ...Field)
    
    // With creates a logger with fields
    With(fields ...Field) Logger
}
```

### HTTPKernel

```go
type HTTPKernel interface {
    // App returns the Fiber application
    App() *fiber.App
    
    // Listen starts the HTTP server
    Listen(addr string) error
    
    // Shutdown gracefully shuts down the server
    Shutdown() error
}
```

### Module

```go
type Module interface {
    // Name returns the module name
    Name() string
    
    // OnStart is called when starting
    OnStart(ctx context.Context) error
    
    // OnStop is called when stopping
    OnStop(ctx context.Context) error
}
```

### Field

```go
type Field interface {
    zap.Field
}
```

## Types

### Options

```go
type Options struct {
    // Name is the application name
    Name string
    
    // Version is the application version
    Version string
    
    // Environment is the environment name
    Environment string
    
    // Modules are custom modules to register
    Modules []contract.Module
    
    // Providers are Fx providers
    Providers []any
    
    // Invokers are Fx invokers
    Invokers []any
    
    // WithHTTP enables the HTTP module
    WithHTTP bool
}
```

### Services

```go
type Services struct {
    // App is the application instance
    App contract.Application
    
    // Config is the configuration instance
    Config contract.Config
    
    // Logger is the logger instance
    Logger contract.Logger
}
```

### ServiceProvider

```go
type ServiceProvider struct {
    fx.In
    
    App    contract.Application
    Config contract.Config
    Logger contract.Logger
}
```

### ContextKey

```go
type ContextKey string

const (
    // RequestIDKey is the key for request ID
    RequestIDKey ContextKey = "requestID"
    
    // ServicesKey is the key for services
    ServicesKey ContextKey = "services"
)
```

## Middleware

### Service Injection Middleware

The service injection middleware is automatically added when `WithHTTP: true` is set.

```go
// Automatically injected, no manual setup required
// Services are available via http.GetLogger(c), etc.
```

### Request Logging Middleware

Request logging is automatically enabled for all HTTP requests.

```go
// Automatically added, logs:
// - Method
// - Path
// - Status code
// - Duration
// - Request ID
```

### Recovery Middleware

Panic recovery is automatically enabled.

```go
// Automatically added from Fiber's recover middleware
```

### Request ID Middleware

Request ID generation is automatically enabled.

```go
// Automatically added from Fiber's requestid middleware
// Accessible via c.Locals("requestID")
```

## Utilities

### Environment Variables

Goe recognizes these environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `GOE_ENV` | Environment name | `"dev"` |
| `APP_NAME` | Application name | From Options |
| `APP_VERSION` | Application version | From Options |
| `DEBUG` | Enable debug mode | `false` |
| `HTTP_HOST` | HTTP server host | `"0.0.0.0"` |
| `HTTP_PORT` | HTTP server port | `8080` |
| `HTTP_READ_TIMEOUT` | Read timeout | `"10s"` |
| `HTTP_WRITE_TIMEOUT` | Write timeout | `"10s"` |
| `HTTP_IDLE_TIMEOUT` | Idle timeout | `"120s"` |
| `HTTP_BODY_LIMIT` | Max body size | `4194304` (4MB) |
| `LOG_LEVEL` | Log level | `"info"` |
| `LOG_FORMAT` | Log format | `"console"` or `"json"` |

### Error Handling

```go
// In HTTP handlers
func handler(c fiber.Ctx) error {
    user, err := service.GetUser(id)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            return c.Status(404).JSON(fiber.Map{
                "error": "User not found",
            })
        }
        
        logger := http.GetLogger(c)
        logger.Error("Failed to get user",
            log.NewField("error", err),
            log.NewField("id", id),
        )
        
        return c.Status(500).JSON(fiber.Map{
            "error": "Internal server error",
        })
    }
    
    return c.JSON(user)
}
```

### Graceful Shutdown

Goe automatically handles graceful shutdown on SIGINT and SIGTERM:

1. Stops accepting new HTTP requests
2. Waits for active requests to complete
3. Calls OnStop for all modules in reverse order
4. Exits cleanly

```go
// Modules can implement custom shutdown logic
func (m *MyModule) OnStop(ctx context.Context) error {
    // Cleanup resources
    // Context has deadline for shutdown
    return nil
}
```