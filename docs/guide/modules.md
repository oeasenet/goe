# Modules

GOE's module system is the foundation of its extensible architecture. Modules allow you to organize functionality into cohesive units with managed lifecycles. This guide covers both using built-in modules and creating custom modules.

## Overview

GOE modules provide:
- **Lifecycle Management**: Automatic startup and shutdown handling
- **Dependency Injection**: Seamless integration with Fx
- **Modular Architecture**: Enable only what you need
- **Extensibility**: Easy to create custom modules

### Module Initialization Order

When GOE starts up, modules are initialized in this specific order:

1. **Config Module** (always first) - Reads application configuration
2. **Logger Module** (always second) - Sets up logging based on config
3. **Core Modules** (if enabled via options):
   - Cache Module (`WithCache: true`)
   - Database Module (`WithDB: true`)
   - Event Module (`WithEvent: true`)
   - MongoDB Module (`WithMongoDB: true`)
   - HTTP Module (`WithHTTP: true`)
4. **Custom Modules** - Your modules added via `Modules` array
5. **Providers** - Dependency injection providers
6. **Invokers** - Functions that run after all dependencies are ready

This order ensures:
- Core infrastructure is ready before custom modules
- Custom modules can depend on core services
- All modules are started before invokers run
- Services are available through dependency injection

## Built-in Modules

GOE comes with several built-in modules that you can enable as needed:

### Core Modules

```go
goe.New(goe.Options{
    WithHTTP:          true,  // HTTP server with GoFiber
    WithDB:            true,  // Database integration with GORM
    WithCache:         true,  // Caching with multiple backends
    WithObservability: true,  // Metrics and tracing
})
```

### Module Dependencies

Some modules have dependencies:
- **HTTP Module**: Depends on Config and Logger
- **Database Module**: Depends on Config and Logger  
- **Cache Module**: Depends on Config and Logger
- **Observability Module**: Depends on Config and Logger

GOE automatically resolves these dependencies.

## Module Interface

All GOE modules implement the `contract.Module` interface:

```go
package contract

import "context"

type Module interface {
    Name() string                      // Unique module identifier
    OnStart(ctx context.Context) error // Initialization logic
    OnStop(ctx context.Context) error  // Cleanup logic
}
```

## Creating Custom Modules

### Basic Module Structure

```go
package mymodule

import (
    "context"
    "go.oease.dev/goe/v2/contract"
)

type Module struct {
    logger contract.Logger
    config contract.Config
    // Add your module's dependencies
}

func NewModule(logger contract.Logger, config contract.Config) *Module {
    return &Module{
        logger: logger,
        config: config,
    }
}

func (m *Module) Name() string {
    return "my-module"
}

func (m *Module) OnStart(ctx context.Context) error {
    m.logger.Info("Starting my module")
    
    // Initialize your module here
    // Connect to external services, start background tasks, etc.
    
    return nil
}

func (m *Module) OnStop(ctx context.Context) error {
    m.logger.Info("Stopping my module")
    
    // Clean up resources here
    // Close connections, stop background tasks, etc.
    
    return nil
}
```

### Service Provider Module

A module that provides a service to other parts of the application:

```go
package emailmodule

import (
    "context"
    "go.oease.dev/goe/v2/contract"
)

type EmailService interface {
    SendEmail(to, subject, body string) error
}

type emailService struct {
    logger contract.Logger
    config contract.Config
}

func (s *emailService) SendEmail(to, subject, body string) error {
    s.logger.Info("Sending email", "to", to, "subject", subject)
    
    // Email sending logic here
    
    return nil
}

type Module struct {
    service EmailService
    logger  contract.Logger
}

func NewModule(logger contract.Logger, config contract.Config) *Module {
    return &Module{
        service: &emailService{logger: logger, config: config},
        logger:  logger,
    }
}

func (m *Module) Name() string {
    return "email-module"
}

func (m *Module) OnStart(ctx context.Context) error {
    m.logger.Info("Email module started")
    return nil
}

func (m *Module) OnStop(ctx context.Context) error {
    m.logger.Info("Email module stopped")
    return nil
}

// Provide the service to other components
func (m *Module) EmailService() EmailService {
    return m.service
}
```

### Background Task Module

A module that runs background tasks:

```go
package taskmodule

import (
    "context"
    "time"
    "go.oease.dev/goe/v2/contract"
)

type Module struct {
    logger contract.Logger
    config contract.Config
    done   chan struct{}
}

func NewModule(logger contract.Logger, config contract.Config) *Module {
    return &Module{
        logger: logger,
        config: config,
        done:   make(chan struct{}),
    }
}

func (m *Module) Name() string {
    return "task-module"
}

func (m *Module) OnStart(ctx context.Context) error {
    m.logger.Info("Starting background task module")
    
    // Start background goroutine
    go m.backgroundTask()
    
    return nil
}

func (m *Module) OnStop(ctx context.Context) error {
    m.logger.Info("Stopping background task module")
    
    // Signal background task to stop
    close(m.done)
    
    return nil
}

func (m *Module) backgroundTask() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            m.logger.Info("Running background task")
            // Perform background work
        case <-m.done:
            m.logger.Info("Background task stopped")
            return
        }
    }
}
```

## Registering Custom Modules

### Method 1: Module Registration with Dependency Injection

```go
package main

import (
    "go.oease.dev/goe/v2"
    "yourapp/internal/modules/email"
    "yourapp/internal/modules/task"
)

func main() {
    goe.New(goe.Options{
        WithHTTP: true,
        Modules: []any{
            email.NewModule,  // Pass constructor functions
            task.NewModule,   // DI will inject dependencies automatically
        },
    })
    
    goe.Run()
}
```

The framework will automatically inject the required dependencies (logger, config, etc.) into your module constructors, just like it does for providers and invokers.

**Key Features:**
- ✅ **Multiple modules supported** - Register as many modules as needed
- ✅ **Automatic dependency injection** - Logger and config automatically provided
- ✅ **Flexible constructor signatures** - Supports both `(logger, config)` and `(config, logger)` parameter orders
- ✅ **Automatic service registration** - Modules can provide services to the DI container automatically
- ✅ **Access to core services** - Modules can use all GOE core services (cache, database, HTTP, etc.)
- ✅ **Proper lifecycle management** - OnStart/OnStop hooks managed by GOE

### Automatic Service Registration

GOE automatically detects and registers services that your modules provide. Simply add methods with these naming patterns to your module:

- `Provide()` - Generic service provider
- `ProvideService()` - Generic service provider  
- `ProvideClient()` - For client modules (gRPC, HTTP clients)
- `ProvideManager()` - For manager services
- `ProvideHandler()` - For handler services
- `ProvideRepository()` - For data access modules
- `ProvideCache()` - For cache services
- `ProvideDB()` - For database services

**Example with automatic service registration:**
```go
type EmailService interface {
    SendEmail(to, subject, body string) error
}

type emailModule struct {
    logger contract.Logger
    config contract.Config
    service EmailService
}

func NewEmailModule(logger contract.Logger, config contract.Config) contract.Module {
    return &emailModule{
        logger:  logger,
        config:  config,
        service: &emailService{logger: logger},
    }
}

// This method will be automatically detected and registered with DI
func (m *emailModule) Provide() EmailService {
    return m.service
}

// Then use it in your application:
goe.New(goe.Options{
    Modules: []any{
        NewEmailModule,
    },
    Invokers: []any{
        func(emailService EmailService) {
            // EmailService is automatically available!
            emailService.SendEmail("user@example.com", "Hello", "World")
        },
    },
})
```

### Method 2: Module with Service Provider

If your module provides services that other components need:

```go
package emailmodule

type EmailService interface {
    SendEmail(to, subject, body string) error
}

type Module struct {
    service EmailService
    logger  contract.Logger
}

// Constructor with DI - dependencies are injected automatically
func NewModule(logger contract.Logger, config contract.Config) contract.Module {
    return &Module{
        service: &emailService{logger: logger, config: config},
        logger:  logger,
    }
}

// Provide service for other components
func ProvideEmailService(module contract.Module) EmailService {
    if m, ok := module.(*Module); ok {
        return m.service
    }
    return nil
}
```

Then register both the module and its service provider:

```go
func main() {
    goe.New(goe.Options{
        WithHTTP: true,
        Modules: []any{
            email.NewModule,  // Module constructor
        },
        Providers: []any{
            email.ProvideEmailService,  // Service provider
        },
        Invokers: []any{
            func(emailService email.EmailService, httpKernel contract.HTTPKernel) {
                // Use the email service in your HTTP routes
                app := httpKernel.App()
                app.Post("/send-email", func(c fiber.Ctx) error {
                    return emailService.SendEmail("user@example.com", "Hello", "Welcome!")
                })
            },
        },
    })
    
    goe.Run()
}
```

## Module Configuration

### Environment-Based Configuration

```go
package mymodule

type Config struct {
    Enabled  bool
    APIKey   string
    Timeout  time.Duration
    Workers  int
}

func NewConfig(config contract.Config) *Config {
    return &Config{
        Enabled:  config.GetBoolWithDefault("MYMODULE_ENABLED", true),
        APIKey:   config.GetString("MYMODULE_API_KEY"),
        Timeout:  config.GetDurationWithDefault("MYMODULE_TIMEOUT", 30*time.Second),
        Workers:  config.GetIntWithDefault("MYMODULE_WORKERS", 5),
    }
}

func NewModule(logger contract.Logger, config contract.Config) *Module {
    moduleConfig := NewConfig(config)
    
    return &Module{
        logger: logger,
        config: moduleConfig,
    }
}
```

### Conditional Module Loading

```go
func main() {
    config := goe.Config()
    
    options := goe.Options{
        WithHTTP: true,
    }
    
    // Conditionally enable modules
    if config.GetBool("EMAIL_ENABLED") {
        options.Providers = append(options.Providers, email.NewModule)
    }
    
    if config.GetBool("TASK_ENABLED") {
        options.Providers = append(options.Providers, task.NewModule)
    }
    
    goe.New(options)
    goe.Run()
}
```

## Module Communication

### Service Injection

```go
// Email module provides EmailService
func (m *EmailModule) EmailService() EmailService {
    return m.service
}

// User module uses EmailService
type UserModule struct {
    emailService EmailService
    logger       contract.Logger
}

func NewUserModule(emailService EmailService, logger contract.Logger) *UserModule {
    return &UserModule{
        emailService: emailService,
        logger:       logger,
    }
}

func (m *UserModule) CreateUser(name, email string) error {
    // Create user logic...
    
    // Send welcome email
    return m.emailService.SendEmail(email, "Welcome!", "Welcome to our app!")
}
```

### Event-Driven Communication

```go
type EventBus interface {
    Publish(event string, data interface{})
    Subscribe(event string, handler func(data interface{}))
}

type UserCreatedEvent struct {
    UserID int
    Email  string
}

// User module publishes events
func (m *UserModule) CreateUser(name, email string) error {
    user := &User{Name: name, Email: email}
    
    // Save user...
    
    // Publish event
    m.eventBus.Publish("user.created", UserCreatedEvent{
        UserID: user.ID,
        Email:  user.Email,
    })
    
    return nil
}

// Email module subscribes to events
func (m *EmailModule) OnStart(ctx context.Context) error {
    m.eventBus.Subscribe("user.created", func(data interface{}) {
        event := data.(UserCreatedEvent)
        m.service.SendEmail(event.Email, "Welcome!", "Welcome to our app!")
    })
    
    return nil
}
```

## Testing Custom Modules

### Unit Testing

```go
package mymodule

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockLogger struct {
    mock.Mock
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
    m.Called(msg, keysAndValues)
}

type MockConfig struct {
    mock.Mock
}

func (m *MockConfig) GetString(key string) string {
    args := m.Called(key)
    return args.String(0)
}

func TestModule_OnStart(t *testing.T) {
    mockLogger := &MockLogger{}
    mockConfig := &MockConfig{}
    
    mockLogger.On("Info", "Starting my module", mock.Anything)
    
    module := NewModule(mockLogger, mockConfig)
    
    err := module.OnStart(context.Background())
    
    assert.NoError(t, err)
    mockLogger.AssertExpectations(t)
}
```

### Integration Testing

```go
func TestModule_Integration(t *testing.T) {
    app := goe.New(goe.Options{
        Modules: []contract.Module{
            NewModule,
        },
    })
    
    // Test module integration
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    err := app.Start(ctx)
    assert.NoError(t, err)
    
    // Test module functionality
    
    err = app.Stop(ctx)
    assert.NoError(t, err)
}
```

## Best Practices

### 1. Single Responsibility

Each module should have a single, well-defined responsibility:

```go
// Good - focused on email functionality
type EmailModule struct { ... }

// Bad - too many responsibilities
type EmailNotificationTaskModule struct { ... }
```

### 2. Graceful Shutdown

Always implement proper cleanup in `OnStop`:

```go
func (m *Module) OnStop(ctx context.Context) error {
    // Close connections
    if m.connection != nil {
        m.connection.Close()
    }
    
    // Stop background tasks
    if m.done != nil {
        close(m.done)
    }
    
    // Cancel contexts
    if m.cancel != nil {
        m.cancel()
    }
    
    return nil
}
```

### 3. Error Handling

Handle errors appropriately in lifecycle methods:

```go
func (m *Module) OnStart(ctx context.Context) error {
    conn, err := m.connectToService()
    if err != nil {
        return fmt.Errorf("failed to connect to service: %w", err)
    }
    
    m.connection = conn
    return nil
}
```

### 4. Configuration Validation

Validate configuration during module initialization:

```go
func NewModule(logger contract.Logger, config contract.Config) *Module {
    apiKey := config.GetString("MYMODULE_API_KEY")
    if apiKey == "" {
        logger.Fatal("MYMODULE_API_KEY is required")
    }
    
    return &Module{
        logger: logger,
        apiKey: apiKey,
    }
}
```

## Module Examples

### Database Connection Pool Module

```go
type DatabasePoolModule struct {
    logger contract.Logger
    config contract.Config
    pool   *sql.DB
}

func (m *DatabasePoolModule) OnStart(ctx context.Context) error {
    dsn := m.config.GetString("DATABASE_URL")
    
    pool, err := sql.Open("postgres", dsn)
    if err != nil {
        return fmt.Errorf("failed to open database: %w", err)
    }
    
    if err := pool.Ping(); err != nil {
        return fmt.Errorf("failed to ping database: %w", err)
    }
    
    m.pool = pool
    m.logger.Info("Database pool initialized")
    return nil
}

func (m *DatabasePoolModule) OnStop(ctx context.Context) error {
    if m.pool != nil {
        m.pool.Close()
    }
    return nil
}

func (m *DatabasePoolModule) DB() *sql.DB {
    return m.pool
}
```

### Redis Client Module

```go
type RedisModule struct {
    logger contract.Logger
    config contract.Config
    client *redis.Client
}

func (m *RedisModule) OnStart(ctx context.Context) error {
    client := redis.NewClient(&redis.Options{
        Addr:     m.config.GetString("REDIS_ADDR"),
        Password: m.config.GetString("REDIS_PASSWORD"),
        DB:       m.config.GetInt("REDIS_DB"),
    })
    
    if err := client.Ping(ctx).Err(); err != nil {
        return fmt.Errorf("failed to connect to Redis: %w", err)
    }
    
    m.client = client
    m.logger.Info("Redis client initialized")
    return nil
}

func (m *RedisModule) Client() *redis.Client {
    return m.client
}
```

## Next Steps

- [**Dependency Injection**](./dependency-injection.md) - Learn more about Fx patterns
- [**Testing**](./testing.md) - Test your custom modules
- [**Best Practices**](./best-practices.md) - Module development best practices