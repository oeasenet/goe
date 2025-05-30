# Goe Framework (WIP)

Goe is a full-featured Golang application development framework inspired by many amazing frameworks. It provides a
modular, interface-based design with dependency injection at its core, making it ideal for building scalable and
maintainable applications.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Architecture Overview](#architecture-overview)
- [Quick Start](#quick-start)
- [Core Modules](#core-modules)
    - [App Module](#app-module)
    - [Config Module](#config-module)
    - [HTTP Module](#http-module)
    - [Log Module](#log-module)
    - [Event Module](#event-module)
    - [Cache Module](#cache-module)
- [Dependency Injection](#dependency-injection)
- [Environment Configuration](#environment-configuration)
- [Advanced Usage](#advanced-usage)
- [Project Structure](#project-structure)
- [Development Setup](#development-setup)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Modular Design**: Built with interfaces and modules for maximum flexibility and extensibility
- **Dependency Injection**: Based on Uber's Fx framework for clean, maintainable code
- **HTTP Routing**: Uses GoFiber v3 as the default HTTP router for high-performance web applications
- **Configuration Management**: Environment-based configuration with .env file support
- **Structured Logging**: Comprehensive logging system with context support and multiple levels
- **Event System**: Publish-subscribe event system for decoupled communication between components
- **Caching**: In-memory caching with tagging support for improved performance
- **Concurrent Safe**: All modules are designed to be thread-safe for reliable operation
- **Lifecycle Management**: Proper initialization, startup, and graceful shutdown of all components

## Installation

To install Goe, use the standard Go package installation command:

```bash
go get go.oease.dev/goe/v2
```

## Architecture Overview

Goe follows a modular architecture with clear separation of concerns:

1. **Contracts**: Define interfaces for all modules, allowing for easy substitution and testing
2. **Core Modules**: Provide default implementations of the contract interfaces
3. **Framework**: Ties everything together and provides a simple API for application developers

The framework is built around the concept of modules, each responsible for a specific aspect of the application. All
modules implement the `Module` interface, which defines lifecycle methods:

```go
type Module interface {
Name() string
Initialize(ctx context.Context) error
Start(ctx context.Context) error
Stop(ctx context.Context) error
}
```

The `App` module is the central component that manages the lifecycle of all other modules and provides dependency
injection using Uber's Fx framework.

## Quick Start

Here's a simple example to get you started with Goe:

```go
package main

import (
	"context"
	"log"

	"go.oease.dev/goe/v2"
)

func main() {
	// Create a new Goe application
	app := goe.New()

	// Configure the HTTP server
	app.Http().Get("/", func(c interface{}) error {
		return c.(interface{ JSON(int, interface{}) error }).JSON(200, map[string]interface{}{
			"message": "Hello, World!",
		})
	})

	// Run the application
	if err := app.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
```

## Core Modules

### App Module

The App module is the central component of the framework, managing the lifecycle of all other modules and providing
dependency injection.

```go
// Get the app module
app := goe.App()

// Register a module
app.RegisterModule(myModule)

// Register a provider
app.RegisterProvider(myProvider)

// Register multiple providers
app.RegisterProviders(
NewUserService,
NewAuthService,
NewDatabaseService,
)

// Execute a function with injected dependencies
app.Invoke(func (log contract.Log, config contract.Config) {
log.Info("Application started with config", "env", config.Get("APP_ENV"))
})

// Register lifecycle hooks
app.OnStart(func (log contract.Log) {
log.Info("Application started")
})

app.OnStop(func (log contract.Log) {
log.Info("Application stopped")
})

// Run the application
app.Run()
```

### Config Module

The Config module manages application configuration from environment variables and .env files.

```go
// Get the config module
config := goe.Config()

// Get a configuration value
value := config.Get("KEY")

// Get a configuration value with a default
value := config.GetDefault("KEY", "default")

// Get typed configuration values
intValue, err := config.GetInt("NUMBER")
boolValue, err := config.GetBool("FLAG")
floatValue, err := config.GetFloat("AMOUNT")
durationValue, err := config.GetDuration("TIMEOUT")
```

### HTTP Module

The HTTP module provides a web server based on GoFiber v3.

```go
// Get the HTTP module
http := goe.Http()

// Register a route
http.Get("/", func (c interface{}) error {
return c.(interface{ JSON(int, interface{}) error }).JSON(200, map[string]interface{}{
"message": "Hello, World!",
})
})

// Create a route group
api := http.Group("/api")
api.Get("/users", func (c interface{}) error {
// Handle request
return nil
})

// Add middleware
http.Use(func (c interface{}) error {
// Middleware logic
return c.(interface{ Next() error }).Next()
})

// Handle different HTTP methods
http.Post("/users", createUserHandler)
http.Put("/users/:id", updateUserHandler)
http.Delete("/users/:id", deleteUserHandler)

// Get route parameters
http.Get("/users/:id", func (c interface{}) error {
id := c.(interface{ Param(string) string }).Param("id")
// Use id
return nil
})
```

### Log Module

The Log module provides structured logging with different levels and context support.

```go
// Get the log module
log := goe.Log()

// Log messages at different levels
log.Debug("Debug message")
log.Info("Info message")
log.Warn("Warning message")
log.Error("Error message")
log.Fatal("Fatal message") // This will exit the application

// Log with fields
log.Info("User logged in", "user_id", 123, "username", "john")

// Create a named logger
userLogger := log.Named("user")
userLogger.Info("User action")

// Log with context
log.WithContext(ctx).Info("Request received")

// Log with fields
log.WithFields(map[string]interface{}{
"user_id": 123,
"action": "login",
}).Info("User action")

// Log with a single field
log.WithField("request_id", requestID).Info("Processing request")
```

### Event Module

The Event module provides a publish-subscribe event system for decoupled communication.

```go
// Get the event module
event := goe.Event()

// Publish an event
event.Publish(ctx, "user.created", map[string]interface{}{
"id": 123,
"username": "john",
})

// Subscribe to an event
event.Subscribe("user.created", func (ctx context.Context, payload interface{}) error {
// Handle event
return nil
})

// Subscribe to an event asynchronously
event.SubscribeAsync("user.created", func (ctx context.Context, payload interface{}) error {
// Handle event asynchronously
return nil
})

// Subscribe to multiple events
event.SubscribeMultiple([]string{"user.created", "user.updated"}, func (ctx context.Context, payload interface{}) error {
// Handle events
return nil
})

// Unsubscribe from an event
event.Unsubscribe("user.created", handlerFunc)
```

### Cache Module

The Cache module provides in-memory caching with tagging support for improved performance.

```go
// Get the cache module
cache := goe.Cache()

// Set a value in the cache
cache.Set(ctx, "key", "value")

// Set a value with TTL
cache.SetWithTTL(ctx, "key", "value", time.Hour)

// Get a value from the cache
value, err := cache.Get(ctx, "key")

// Get typed values from the cache
strValue, err := cache.GetString(ctx, "key")
intValue, err := cache.GetInt(ctx, "key")
boolValue, err := cache.GetBool(ctx, "key")
floatValue, err := cache.GetFloat(ctx, "key")

// Check if a key exists
exists, err := cache.Has(ctx, "key")

// Delete a key
cache.Delete(ctx, "key")

// Use tagged cache for easier management
userCache := cache.Tags("users")
userCache.Set(ctx, "user:123", userData)
userCache.Clear(ctx) // Clear all cache entries with the "users" tag

// Multiple tags
userProfileCache := cache.Tags("users", "profiles")
userProfileCache.Set(ctx, "profile:123", profileData)
```

## Dependency Injection

Goe uses Uber's Fx framework for dependency injection. This allows for clean, maintainable code with explicit
dependencies.

### Basic Usage

```go
// Define a service
type UserService struct {
Log    contract.Log
Config contract.Config
}

// Create a constructor function
func NewUserService(log contract.Log, config contract.Config) *UserService {
return &UserService{
Log:    log,
Config: config,
}
}

// Register the service with the application
app.RegisterProvider(NewUserService)

// Use the service
app.Invoke(func (userService *UserService) {
// Use userService
})
```

### Named Instances

```go
app.RegisterProvider(
fx.Annotate(
func () *Database { return NewDatabase("primary") },
fx.ResultTags(`name:"primary"`),
),
)

app.RegisterProvider(
fx.Annotate(
func () *Database { return NewDatabase("replica") },
fx.ResultTags(`name:"replica"`),
),
)

// Inject named instances
app.Invoke(func (
primary *Database `name:"primary"`,
replica *Database `name:"replica"`,
) {
// Use primary and replica databases
})
```

For more detailed information about dependency injection, see the [DI_GUIDE.md](DI_GUIDE.md) file.

## Environment Configuration

Goe uses environment variables for configuration. You can set these variables in your system environment or in .env
files.

The framework looks for the following files in order:

1. `.env` (default environment file)
2. `.<env>.env` (environment-specific file, where `<env>` is the value of `GOE_ENV`)

Environment variables set in the system override values from .env files.

Example .env file:

```
APP_NAME=MyApp
APP_ENV=development
HTTP_PORT=8080
LOG_LEVEL=debug
CACHE_TTL=3600
```

## Advanced Usage

### Custom Modules

You can create custom modules by implementing the `Module` interface:

```go
type MyModule struct {
// Module state
}

func (m *MyModule) Name() string {
return "my-module"
}

func (m *MyModule) Initialize(ctx context.Context) error {
// Initialize resources
return nil
}

func (m *MyModule) Start(ctx context.Context) error {
// Start background processes
return nil
}

func (m *MyModule) Stop(ctx context.Context) error {
// Clean up resources
return nil
}

// Register the module
app.RegisterModule(&MyModule{})
```

### Middleware

You can add middleware to the HTTP module:

```go
// Add global middleware
app.Http().Use(func (c interface{}) error {
// Middleware logic
return c.(interface{ Next() error }).Next()
})

// Add middleware to a group
api := app.Http().Group("/api")
api.Use(authMiddleware)
```

### Error Handling

```go
app.Http().Get("/users/:id", func(c interface{}) error {
id := c.(interface{ Param(string) string }).Param("id")

user, err := getUserByID(id)
if err != nil {
return c.(interface{ Status(int) interface{} }).Status(404).
JSON(map[string]interface{}{
"error": "User not found",
})
}

return c.(interface{ JSON(int, interface{}) error }).JSON(200, user)
})
```

## Project Structure

A typical Goe project might be structured as follows:

```
myapp/
├── cmd/
│   └── server/
│       └── main.go         # Application entry point
├── config/
│   └── config.go           # Application configuration
├── internal/
│   ├── handlers/           # HTTP handlers
│   ├── middleware/         # HTTP middleware
│   ├── models/             # Data models
│   └── services/           # Business logic
├── pkg/
│   └── utils/              # Shared utilities
├── .env                    # Environment variables
├── .env.example            # Example environment variables
├── go.mod                  # Go module file
└── go.sum                  # Go module checksum
```

## Development Setup

1. Install Go (1.18 or later)
2. Clone the repository
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Create a `.env` file based on `.env.example`
5. Run the application:
   ```bash
   go run cmd/server/main.go
   ```

## Testing

Goe provides utilities for testing with Fx. Here's an example of how to test a component:

```go
func TestMyService(t *testing.T) {
// Create a test Fx app
testApp := fxtest.New(t,
// Provide dependencies
fx.Provide(
NewMockLogger,
NewMockConfig,
NewMyService,
),

// Test the service
fx.Invoke(func (service *MyService) {
// Assert that the service works correctly
result := service.DoSomething()
if result != expected {
t.Errorf("Expected %v, got %v", expected, result)
}
}),
)

// Start and stop the test app
testApp.RequireStart()
testApp.RequireStop()
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Commit your changes: `git commit -am 'Add my feature'`
4. Push to the branch: `git push origin feature/my-feature`
5. Submit a pull request

Please make sure your code follows the project's coding style and includes appropriate tests.

## License

This project is licensed under the Apache-2.0 License - see the [LICENSE](LICENSE) file for details.
