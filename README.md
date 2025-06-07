# Goe - Modern Go Application Framework

<div align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-Apache_2.0-green?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Status-Beta-yellow?style=for-the-badge" alt="Status">
</div>

Goe is a modern Go application framework that combines the best practices from frameworks in other languages and
GoFiber. Built entirely on Uber's Fx dependency injection framework, Goe prioritizes developer experience,
extensibility, and concurrent safety.

## 🚀 Features

- **🔌 Dependency Injection**: Built on Uber's Fx for powerful, type-safe dependency injection
- **🌐 HTTP Server**: Integrated GoFiber v3 with automatic request logging and middleware
- **📝 Structured Logging**: Uber's Zap logger with pretty console output for development
- **⚙️ Configuration Management**: Environment-based configuration with hot reload support
- **💾 Cache Support**: Multiple cache drivers via Fiber's storage interface (Memory, Redis, SQLite, etc.)
- **🔧 Module System**: Extensible module system with lifecycle hooks
- **🛡️ Type Safety**: Leverages Go's type system for compile-time safety
- **🎯 Developer Experience**: Simple global accessors and intuitive APIs
- **🔄 Concurrent Safe**: Thread-safe operations throughout the framework

## 📚 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
    - [Application Lifecycle](#application-lifecycle)
    - [Dependency Injection](#dependency-injection)
    - [Modules](#modules)
    - [Configuration](#configuration)
    - [Logging](#logging)
    - [HTTP Server](#http-server)
    - [Cache](#cache)
- [Examples](#examples)
- [API Reference](#api-reference)
- [Best Practices](#best-practices)
- [Contributing](#contributing)

## 🔧 Installation

```bash
go get go.oease.dev/goe/v2
```

## 🚀 Quick Start

Create a simple Goe application:

```go
package main

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
)

func main() {
	// Create application with HTTP enabled
	// App name, version, and environment are configured via .env files
	_ = goe.New(goe.Options{
		WithHTTP: true,
		Invokers: []any{
			func(http contract.HTTPKernel, logger contract.Logger) {
				app := http.App()

				app.Get("/", func(c fiber.Ctx) error {
					logger.Info("Home route accessed")
					return c.JSON(fiber.Map{
						"message": "Welcome to Goe!",
					})
				})
			},
		},
	})

	// Run the application
	goe.Run()
}
```

## 🏗️ Core Concepts

### Application Lifecycle

Goe applications follow a structured lifecycle:

1. **Initialization**: Create application with `goe.New()`
2. **Registration**: Register modules, providers, and invokers
3. **Start**: Start the application with `goe.Run()`
4. **Running**: Application serves requests and runs background tasks
5. **Shutdown**: Graceful shutdown on interrupt signals

```go
// 1. Initialization (reads settings from config)
app := goe.New(goe.Options{
	WithHTTP: true,
	// Other options...
})

// 2. Registration happens internally

// 3. Start
goe.Run()

// 4. Running...

// 5. Graceful shutdown on SIGINT/SIGTERM
```

### Dependency Injection

Goe uses Uber's Fx for dependency injection. There are several patterns for accessing dependencies:

#### 1. Direct Injection

```go
type UserService struct {
logger contract.Logger
config contract.Config
}

func NewUserService(logger contract.Logger, config contract.Config) *UserService {
return &UserService{
logger: logger,
config: config,
}
}
```

#### 2. Parameter Objects (Recommended)

```go
type UserServiceParams struct {
fx.In
Logger contract.Logger
Config contract.Config
DB     *sql.DB `optional:"true"`
}

func NewUserService(params UserServiceParams) *UserService {
return &UserService{
logger: params.Logger,
config: params.Config,
db:     params.DB,
}
}
```

#### 3. Result Objects

```go
type UserServiceResult struct {
fx.Out
Service *UserService
Handler http.Handler `name:"user"`
}

func NewUserService(logger contract.Logger) UserServiceResult {
service := &UserService{logger: logger}
return UserServiceResult{
Service: service,
Handler: service,
}
}
```

### Modules

Modules are self-contained units of functionality with lifecycle hooks:

```go
type MyModule struct {
config contract.Config
logger contract.Logger
}

func NewMyModule(config contract.Config, logger contract.Logger) contract.Module {
return &MyModule{
config: config,
logger: logger,
}
}

func (m *MyModule) Name() string {
return "mymodule"
}

func (m *MyModule) OnStart(ctx context.Context) error {
m.logger.Info("MyModule starting")
// Initialize resources
return nil
}

func (m *MyModule) OnStop(ctx context.Context) error {
m.logger.Info("MyModule stopping")
// Cleanup resources
return nil
}

// Register the module
_ = goe.New(goe.Options{
Modules: []contract.Module{
NewMyModule(goe.Config(), goe.Log()),
},
})
```

### Configuration

Goe provides a powerful configuration system with automatic environment-based loading:

#### Environment Detection

The framework automatically detects the environment through the `GOE_ENV` environment variable. This determines which configuration files to load and affects logging behavior.

```bash
# Set environment before running
export GOE_ENV=production
./myapp

# Or inline
GOE_ENV=production ./myapp
```

If `GOE_ENV` is not set, it defaults to `dev`.

#### Configuration Files

Goe loads configuration files in a specific order, with later files overriding earlier ones:

```env
# .env - Base configuration (loaded first)
APP_NAME=My Application
APP_VERSION=1.0.0
DEBUG=true
HTTP_PORT=8080
HTTP_HOST=0.0.0.0
LOG_LEVEL=info
LOG_FORMAT=console
DATABASE_URL=postgres://localhost/myapp_dev
```

```env
# .local.env - Local overrides (loaded second)
# This file is typically gitignored for local development settings
DATABASE_URL=postgres://localhost/myapp_local
DEBUG=true
SECRET_KEY=local-secret-key
```

```env
# .{GOE_ENV}.env - Environment-specific configuration (loaded third)
# For example: .production.env, .staging.env, .dev.env
DEBUG=false
LOG_FORMAT=json
DATABASE_URL=postgres://prod-server/myapp_prod
```

#### Configuration Loading Order

1. **`.env`** - Base configuration file (lowest priority)
2. **`.local.env`** - Local overrides (useful for development)
3. **`.{GOE_ENV}.env`** - Environment-specific configuration (e.g., `.production.env`)
4. **System environment variables** - Highest priority, overrides all files

This loading order ensures that:
- Base settings are defined in `.env`
- Developers can override settings locally without affecting the repository
- Environment-specific settings are properly isolated
- System environment variables always win (useful for secrets in production)

#### Accessing Configuration

```go
// Using global accessor
config := goe.Config()
appName := config.GetString("APP_NAME")
httpPort := config.GetInt("HTTP_PORT")
debug := config.GetBool("DEBUG")

// In HTTP handlers using context
func handler(c fiber.Ctx) error {
config := http.GetConfig(c)
appName := config.GetString("APP_NAME")
return c.SendString(appName)
}
```

#### Configuration Best Practices

1. **Use `.env` for defaults**: Put all default configuration values in `.env` and commit it to version control
2. **Use `.local.env` for local development**: Add `.local.env` to `.gitignore` for developer-specific settings
3. **Use `.{GOE_ENV}.env` for environments**: Create separate files for each deployment environment
4. **Use system environment variables for secrets**: Never commit sensitive data; use environment variables in production
5. **Document all configuration options**: Add comments in your `.env` file to explain each setting

Example `.gitignore`:
```gitignore
.local.env
.env.local
*.local.env
```

### Logging

Goe uses Uber's Zap for high-performance structured logging:

```go
// Global logger
logger := goe.Log()

// Basic logging
logger.Info("User logged in")
logger.Error("Failed to connect to database")
logger.Debug("Processing request")

// Structured logging
logger.Info("User action",
log.NewField("user_id", 123),
log.NewField("action", "login"),
log.NewField("ip", "192.168.1.1"),
)

// In HTTP handlers with request context
func handler(c fiber.Ctx) error {
logger := http.GetLogger(c) // Includes request ID
logger.Info("Processing request")
return c.SendStatus(200)
}
```

#### Log Levels

- `DEBUG`: Detailed debugging information
- `INFO`: General informational messages
- `WARN`: Warning messages
- `ERROR`: Error messages
- `FATAL`: Fatal errors (will exit application)

### HTTP Server

Goe integrates GoFiber v3 for high-performance HTTP handling with comprehensive configuration options:

#### Basic Routes

```go
_ = goe.New(goe.Options{
WithHTTP: true,
Invokers: []any{
func (http contract.HTTPKernel) {
app := http.App()

// GET route
app.Get("/users", getUsers)

// POST route
app.Post("/users", createUser)

// Route with parameters
app.Get("/users/:id", getUser)

// Route groups
api := app.Group("/api/v1")
api.Get("/products", getProducts)
},
},
})
```

#### Middleware

```go
func setupRoutes(http contract.HTTPKernel) {
app := http.App()

// Global middleware
app.Use(cors.New())
app.Use(compress.New())

// Route-specific middleware
app.Get("/admin/*", authMiddleware, adminHandler)

// Group middleware
api := app.Group("/api", rateLimitMiddleware)
api.Get("/users", getUsers)
}
```

#### Configuration

The HTTP module supports extensive configuration through environment variables:

```env
# Basic HTTP configuration
HTTP_HOST=0.0.0.0
HTTP_PORT=8080
HTTP_READ_TIMEOUT=10s
HTTP_WRITE_TIMEOUT=10s
HTTP_IDLE_TIMEOUT=30s

# Fiber framework configuration
FIBER_SERVER_HEADER=MyApp              # Server header value
FIBER_STRICT_ROUTING=false             # Enable strict routing (exact match)
FIBER_CASE_SENSITIVE=false             # Enable case sensitive routing
FIBER_IMMUTABLE=false                  # Enable immutable mode
FIBER_UNESCAPE_PATH=false              # Unescape path values
FIBER_BODY_LIMIT=4194304               # Max body size in bytes (4MB)
FIBER_STREAM_REQUEST_BODY=true         # Enable request body streaming
FIBER_CONCURRENCY=262144               # Maximum concurrent connections
FIBER_REDUCE_MEMORY=false              # Reduce memory usage
FIBER_ENABLE_IP_VALIDATION=false       # Enable IP validation

# Proxy configuration
FIBER_TRUST_PROXY=false                # Trust proxy headers
FIBER_PROXY_HEADER=X-Forwarded-For     # Custom proxy header
FIBER_TRUST_PROXIES=192.168.1.0/24,10.0.0.0/8  # Trusted proxy IPs/CIDRs
FIBER_TRUST_LINK_LOCAL=true            # Trust link-local addresses
FIBER_TRUST_LOOPBACK=true              # Trust loopback addresses
FIBER_TRUST_PRIVATE=true               # Trust private addresses
```

#### Request Validation

Goe includes built-in request validation using go-playground/validator:

```go
// Define your request struct with validation tags
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"required,min=18,max=120"`
    Password string `json:"password" validate:"required,min=8"`
}

func createUserHandler(c fiber.Ctx) error {
    // Get validator from context
    validator := http.GetValidator(c)
    
    // Parse request body
    var req CreateUserRequest
    if err := c.Bind().JSON(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
    }
    
    // Validate request
    if err := validator.Validate(req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }
    
    // Process valid request...
    return c.JSON(fiber.Map{"status": "created"})
}

// Register custom validation rules
func registerCustomValidations(http contract.HTTPKernel) {
    validator := http.Validator().(*http.CustomValidator)
    
    // Register custom validation function
    validator.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
        phone := fl.Field().String()
        // Custom phone validation logic
        return len(phone) >= 10 && len(phone) <= 15
    })
    
    // Register struct-level validation
    validator.RegisterStructValidation(func(sl validator.StructLevel) {
        user := sl.Current().Interface().(CreateUserRequest)
        if user.Age < 21 && strings.Contains(user.Email, "@bar.com") {
            sl.ReportError(user.Age, "age", "Age", "bartender", "")
        }
    }, CreateUserRequest{})
}
```

#### Context Helpers

Access services in HTTP handlers without circular dependencies:

```go
func handler(c fiber.Ctx) error {
// Get logger with request ID
logger := http.GetLogger(c)

// Get configuration
config := http.GetConfig(c)

// Get application instance
app := http.GetApp(c)

// Get validator
validator := http.GetValidator(c)

logger.Info("Request received",
log.NewField("path", c.Path()),
log.NewField("method", c.Method()),
)

return c.JSON(fiber.Map{
"app": config.GetString("APP_NAME"),
"version": app.Version(),
})
}
```

### Cache

Goe provides a powerful caching system built on Fiber's storage interface, supporting multiple drivers out of the box:

#### Basic Usage

```go
// Enable cache in your application
app := goe.New(goe.Options{
    WithCache: true,
})

// Access cache via dependency injection
func useCacheExample(cache contract.Cache) {
    // Set a value with TTL
    cache.Set("user:123", user, 5*time.Minute)
    
    // Get a value
    value, err := cache.Get("user:123")
    
    // Check if key exists
    if cache.Has("user:123") {
        // Key exists
    }
    
    // Remove a key
    cache.Forget("user:123")
    
    // Store forever (no expiration)
    cache.Forever("config:app", appConfig)
}
```

#### Type-safe Operations

```go
import "go.oease.dev/goe/v2/core/cache"

// Store and retrieve with type safety
user := User{ID: 1, Name: "John"}
cache.Set("user:1", user, 10*time.Minute)

// Get with type
retrievedUser, err := cache.GetT[User](cache, "user:1")

// Remember pattern - compute only on cache miss
product, err := cache.RememberT(cache, "product:1", 1*time.Hour, func() (Product, error) {
    // This only runs if not in cache
    return fetchProductFromDB(1)
})
```

#### Cache Configuration

Configure cache drivers via environment variables:

```bash
# Memory driver (default)
CACHE_DRIVER=memory
CACHE_MEMORY_GC_INTERVAL=10s

# Redis driver for distributed caching
CACHE_DRIVER=redis
CACHE_REDIS_URL=redis://localhost:6379/0
# Or use individual settings:
CACHE_REDIS_HOSTS=localhost:6379
CACHE_REDIS_PASSWORD=secret
CACHE_REDIS_DATABASE=0

# Common settings
CACHE_PREFIX=myapp
CACHE_TTL=2h
```

#### Multiple Cache Stores

```go
func setupMultipleStores(manager contract.CacheManager) {
    // Default store
    defaultCache := manager.Store()
    
    // Named store with different configuration
    sessionCache := manager.Store("sessions")
    
    // Use different stores for different purposes
    defaultCache.Set("app:config", config, 24*time.Hour)
    sessionCache.Set("session:abc123", sessionData, 30*time.Minute)
}
```

Configuration for multiple stores:

```bash
# Primary cache (Redis)
CACHE_STORE=primary
CACHE_primary_DRIVER=redis
CACHE_primary_REDIS_DATABASE=0

# Session cache (Memory)
CACHE_sessions_DRIVER=memory
CACHE_sessions_PREFIX=sessions
```

#### Available Drivers

Goe supports all Fiber storage drivers:
- **memory**: Fast in-memory cache (default)
- **redis**: Redis with auto-pipelining and client-side caching
- **sqlite3**: SQLite-based persistent cache
- **postgres**: PostgreSQL storage
- **mysql**: MySQL storage
- **mongodb**: MongoDB storage
- Plus many more...

See the [Cache Documentation](docs/CACHE.md) for detailed information on all available drivers and advanced usage.

## 📖 Examples

### Complete Application Example

```go
package main

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/http"
	"go.oease.dev/goe/v2/core/log"
	"go.uber.org/fx"
)

// Domain models
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Service layer
type UserService struct {
	logger contract.Logger
	config contract.Config
}

func NewUserService(params struct {
	fx.In
	Logger contract.Logger
	Config contract.Config
}) *UserService {
	return &UserService{
		logger: params.Logger,
		config: params.Config,
	}
}

func (s *UserService) GetUser(id string) (*User, error) {
	s.logger.Info("Getting user", log.NewField("id", id))

	// Simulate database lookup
	return &User{
		ID:    id,
		Name:  "John Doe",
		Email: "john@example.com",
	}, nil
}

// Controller layer
type UserController struct {
	service *UserService
	logger  contract.Logger
}

func NewUserController(params struct {
	fx.In
	Service *UserService
	Logger  contract.Logger
}) *UserController {
	return &UserController{
		service: params.Service,
		logger:  params.Logger,
	}
}

func (c *UserController) RegisterRoutes(app *fiber.App) {
	users := app.Group("/users")
	users.Get("/:id", c.GetUser)
	users.Post("/", c.CreateUser)
}

func (c *UserController) GetUser(ctx fiber.Ctx) error {
	id := ctx.Params("id")

	user, err := c.service.GetUser(id)
	if err != nil {
		return ctx.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return ctx.JSON(user)
}

func (c *UserController) CreateUser(ctx fiber.Ctx) error {
	// Using context helpers
	logger := http.GetLogger(ctx)
	config := http.GetConfig(ctx)

	logger.Info("Creating user",
		log.NewField("app", config.GetString("APP_NAME")),
	)

	// Implementation...
	return ctx.JSON(fiber.Map{"status": "created"})
}

func main() {
	_ = goe.New(goe.Options{
		WithHTTP:    true,
		Providers: []any{
			NewUserService,
			NewUserController,
		},
		Invokers: []any{
			func(params struct {
				fx.In
				HTTP       contract.HTTPKernel
				Controller *UserController
				Logger     contract.Logger
			}) {
				// Register routes
				params.Controller.RegisterRoutes(params.HTTP.App())

				// Add health check
				params.HTTP.App().Get("/health", func(c fiber.Ctx) error {
					return c.JSON(fiber.Map{
						"status":  "healthy",
						"version": goe.App().Version(),
					})
				})

				params.Logger.Info("Routes registered")
			},
		},
	})

	goe.Log().Info("Starting User Service...")
	goe.Run()
}
```

### Custom Module Example

```go
package main

import (
	"context"
	"time"

	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
)

type CacheModule struct {
	logger contract.Logger
	config contract.Config
	ticker *time.Ticker
}

func NewCacheModule() contract.Module {
	return &CacheModule{}
}

func (m *CacheModule) Name() string {
	return "cache"
}

func (m *CacheModule) OnStart(ctx context.Context) error {
	m.logger = goe.Log()
	m.config = goe.Config()

	m.logger.Info("Cache module starting")

	// Start background cleanup
	m.ticker = time.NewTicker(5 * time.Minute)
	go m.cleanup()

	return nil
}

func (m *CacheModule) OnStop(ctx context.Context) error {
	m.logger.Info("Cache module stopping")

	if m.ticker != nil {
		m.ticker.Stop()
	}

	return nil
}

func (m *CacheModule) cleanup() {
	for range m.ticker.C {
		m.logger.Debug("Running cache cleanup")
		// Cleanup logic
	}
}

// Usage
func main() {
	_ = goe.New(goe.Options{
		Modules: []contract.Module{
			NewCacheModule(),
		},
	})

	goe.Run()
}
```

## 📚 API Reference

### Global Functions

```go
// Create new application
func New(opts ...Options) contract.Application

// Run the application
func Run()

// Global accessors
func App() contract.Application
func Config() contract.Config
func Log() contract.Logger
func HTTP() contract.HTTPKernel
func Context() context.Context
func IsRunning() bool
func GetEnvironment() string
```

### Options

```go
type Options struct {
Modules   []contract.Module // Custom modules
Providers []any              // Fx providers
Invokers  []any              // Fx invokers
WithHTTP  bool               // Enable HTTP module
}
```

### Contracts

```go
// Application interface
type Application interface {
Name() string
Version() string
Environment() string
Context() context.Context
IsRunning() bool
Container() *fx.App
Register(opts ...fx.Option) error
}

// Config interface
type Config interface {
Get(key string) any
GetString(key string) string
GetInt(key string) int
GetBool(key string) bool
GetFloat64(key string) float64
GetDuration(key string) time.Duration
GetStringSlice(key string) []string
GetStringMap(key string) map[string]any
Set(key string, value any)
Has(key string) bool
All() map[string]any
}

// Logger interface
type Logger interface {
Debug(msg string, fields ...Field)
Info(msg string, fields ...Field)
Warn(msg string, fields ...Field)
Error(msg string, fields ...Field)
Fatal(msg string, fields ...Field)
With(fields ...Field) Logger
}

// HTTPKernel interface
type HTTPKernel interface {
App() *fiber.App
Listen(addr string) error
Shutdown() error
}

// Module interface
type Module interface {
Name() string
OnStart(ctx context.Context) error
OnStop(ctx context.Context) error
}
```

## 🎯 Best Practices

### 1. Use Parameter Objects for Dependencies

```go
// Good
type ServiceParams struct {
fx.In
Logger contract.Logger
Config contract.Config
DB     *sql.DB `optional:"true"`
}

func NewService(params ServiceParams) *Service {
return &Service{
logger: params.Logger,
config: params.Config,
db:     params.DB,
}
}

// Avoid
func NewService(logger contract.Logger, config contract.Config, db *sql.DB) *Service {
// This becomes unwieldy with many dependencies
}
```

### 2. Organize Code by Domain

```
project/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── user/
│   │   ├── service.go
│   │   ├── controller.go
│   │   └── repository.go
│   └── product/
│       ├── service.go
│       ├── controller.go
│       └── repository.go
├── pkg/
│   └── middleware/
│       └── auth.go
├── .env
├── .env.production
└── go.mod
```

### 3. Use Context Helpers in HTTP Handlers

```go
// Good - No circular dependencies
func handler(c fiber.Ctx) error {
logger := http.GetLogger(c)
config := http.GetConfig(c)

logger.Info("Request processed")
return c.JSON(fiber.Map{
"app": config.GetString("APP_NAME"),
})
}

// Avoid - Can cause circular dependencies
var globalLogger contract.Logger // Don't do this
```

### 4. Implement Graceful Shutdown

```go
type Service struct {
logger contract.Logger
cancel context.CancelFunc
}

func (s *Service) OnStart(ctx context.Context) error {
ctx, cancel := context.WithCancel(ctx)
s.cancel = cancel

go s.backgroundTask(ctx)
return nil
}

func (s *Service) OnStop(ctx context.Context) error {
if s.cancel != nil {
s.cancel()
}

// Wait for cleanup with timeout
done := make(chan struct{})
go func () {
// Cleanup logic
close(done)
}()

select {
case <-done:
s.logger.Info("Service stopped gracefully")
case <-ctx.Done():
s.logger.Warn("Service stop timeout")
}

return nil
}
```

### 5. Use Structured Logging

```go
// Good
logger.Info("User action",
log.NewField("user_id", userID),
log.NewField("action", "login"),
log.NewField("duration_ms", duration.Milliseconds()),
)

// Avoid
logger.Info(fmt.Sprintf("User %d performed %s in %dms", userID, "login", duration.Milliseconds()))
```

### 6. Handle Errors Properly

```go
func (s *Service) GetUser(id string) (*User, error) {
user, err := s.repo.FindByID(id)
if err != nil {
if errors.Is(err, sql.ErrNoRows) {
s.logger.Debug("User not found", log.NewField("id", id))
return nil, ErrUserNotFound
}
s.logger.Error("Database error",
log.NewField("error", err),
log.NewField("id", id),
)
return nil, fmt.Errorf("get user: %w", err)
}
return user, nil
}
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

1. Clone the repository:

```bash
git clone https://github.com/oeasenet/goe.git
cd goe
```

2. Install dependencies:

```bash
go mod download
```

3. Run tests:

```bash
go test ./...
```

4. Run linting:

```bash
golangci-lint run
```

## 📄 License

Goe is open-source software licensed under the [Apache 2.0 license](LICENSE).

## 🙏 Acknowledgments

Goe is built on the shoulders of giants:

- [Uber's Fx](https://github.com/uber-go/fx) - Dependency injection framework
- [GoFiber](https://github.com/gofiber/fiber) - Fast HTTP framework
- [Uber's Zap](https://github.com/uber-go/zap) - Structured logging

## 📞 Support

- 📧 Email: godev@oease.net
- 🐛 Issues: [GitHub Issues](https://github.com/oeasenet/goe/issues)

---

Made with ❤️ by FantasticTony