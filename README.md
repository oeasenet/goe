# Goe Framework

Goe is a full-featured Golang application development framework inspired by Java Spring Boot, Golang Goravel, and GoFiber. It provides a modular, interface-based design with dependency injection at its core.

## Features

- **Modular Design**: Built with interfaces and modules for maximum flexibility
- **Dependency Injection**: Based on Uber's Fx framework
- **HTTP Routing**: Uses GoFiber v3 as the default HTTP router
- **Configuration Management**: Environment-based configuration with .env file support
- **Logging**: Structured logging with context support
- **Event System**: Publish-subscribe event system
- **Caching**: In-memory caching with tagging support
- **Concurrent Safe**: All modules are designed to be thread-safe

## Installation

```bash
go get go.oease.dev/goe/v2
```

## Quick Start

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

## Modules

### App Module

The App module is the central component of the framework, managing the lifecycle of all other modules.

```go
// Get the app module
app := goe.App()

// Register a module
app.RegisterModule(myModule)

// Register a provider
app.RegisterProvider(myProvider)

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

// Get a typed configuration value
intValue, err := config.GetInt("NUMBER")
boolValue, err := config.GetBool("FLAG")
```

### HTTP Module

The HTTP module provides a web server based on GoFiber.

```go
// Get the HTTP module
http := goe.Http()

// Register a route
http.Get("/", func(c interface{}) error {
	return c.(interface{ JSON(int, interface{}) error }).JSON(200, map[string]interface{}{
		"message": "Hello, World!",
	})
})

// Create a route group
api := http.Group("/api")
api.Get("/users", func(c interface{}) error {
	// Handle request
	return nil
})
```

### Log Module

The Log module provides structured logging.

```go
// Get the log module
log := goe.Log()

// Log messages at different levels
log.Debug("Debug message")
log.Info("Info message")
log.Warn("Warning message")
log.Error("Error message")

// Log with fields
log.Info("User logged in", "user_id", 123, "username", "john")

// Create a named logger
userLogger := log.Named("user")
userLogger.Info("User action")
```

### Event Module

The Event module provides a publish-subscribe event system.

```go
// Get the event module
event := goe.Event()

// Publish an event
event.Publish(ctx, "user.created", map[string]interface{}{
	"id": 123,
	"username": "john",
})

// Subscribe to an event
event.Subscribe("user.created", func(ctx context.Context, payload interface{}) error {
	// Handle event
	return nil
})

// Subscribe to an event asynchronously
event.SubscribeAsync("user.created", func(ctx context.Context, payload interface{}) error {
	// Handle event asynchronously
	return nil
})
```

### Cache Module

The Cache module provides in-memory caching with tagging support.

```go
// Get the cache module
cache := goe.Cache()

// Set a value in the cache
cache.Set(ctx, "key", "value")

// Set a value with TTL
cache.SetWithTTL(ctx, "key", "value", time.Hour)

// Get a value from the cache
value, err := cache.Get(ctx, "key")

// Get a typed value from the cache
strValue, err := cache.GetString(ctx, "key")
intValue, err := cache.GetInt(ctx, "key")

// Check if a key exists
exists, err := cache.Has(ctx, "key")

// Delete a key
cache.Delete(ctx, "key")

// Use tagged cache
userCache := cache.Tags("users")
userCache.Set(ctx, "user:123", userData)
userCache.Clear(ctx) // Clear all cache entries with the "users" tag
```

## Environment Configuration

Goe uses environment variables for configuration. You can set these variables in your system environment or in .env files.

The framework looks for the following files in order:
1. `.env` (default environment file)
2. `.<env>.env` (environment-specific file, where `<env>` is the value of `GOE_ENV`)

Environment variables set in the system override values from .env files.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.