# GOE Framework

GOE is a simple, lightweight, and easy-to-use web development framework for Go. Built on top of [GoFiber](https://gofiber.io/) v3, GOE provides a comprehensive set of tools and modules to help developers quickly build robust web applications while focusing on business logic rather than infrastructure concerns.

Inspired by frameworks like Spring Boot in the Java ecosystem, GOE aims to provide a similar developer experience but with Go's simplicity and performance.

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

> **Note**: This documentation was written by AI to provide comprehensive information about the GOE framework.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Architecture](#architecture)
- [Modules](#modules)
  - [MongoDB](#mongodb)
  - [Mailer](#mailer)
  - [Cache](#cache)
  - [Queue](#queue)
  - [Cron](#cron)
  - [Logging](#logging)
  - [Configuration](#configuration-1)
- [Middleware](#middleware)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)
- [Examples](#examples)

## Features

- **Modular Architecture**: Use only what you need
- **Dependency Injection**: Simple container-based DI system
- **MongoDB Integration**: Built-in MongoDB support
- **Meilisearch Integration**: Full-text search capabilities
- **Redis-based Caching**: High-performance caching
- **Message Queue**: Asynchronous task processing
- **Cron Jobs**: Scheduled task execution
- **Mailer**: Email sending with multiple provider support (SMTP, Resend, SES)
- **Logging**: Structured logging with Zap
- **Configuration Management**: Environment-based configuration
- **Middleware Support**: Various built-in middlewares
- **File Storage**: S3-compatible storage support
- **Graceful Shutdown**: Clean application termination

## Installation

```shell
go get -u go.oease.dev/goe
```

## Quick Start

Create a new Go project and add the following code to your main.go file:

```go
package main

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe"
	"go.oease.dev/goe/webresult"
)

func main() {
	// Initialize the GOE application
	err := goe.NewApp()
	if err != nil {
		panic(err)
	}

	// Define a simple route
	goe.UseFiber().App().Get("/hello", func(ctx fiber.Ctx) error {
		return webresult.SendSucceed(ctx, "Hello, World!")
	})

	// Start the server
	err = goe.Run()
	if err != nil {
		panic(err)
	}
}
```

## Configuration

GOE uses environment variables or configuration files for setup. Create a `configs` directory in your project root with the necessary configuration files.

### Environment Variables

Here are some of the key environment variables you can set:

```
# App Configuration
APP_NAME=MyApp
APP_VERSION=1.0.0
APP_ENV=dev  # dev or prod

# Feature Toggles
MONGODB_ENABLED=true
MEILISEARCH_ENABLED=false
MAILER_ENABLED=true

# MongoDB Configuration
MONGODB_URI=mongodb://localhost:27017
MONGODB_DB=myapp

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_USERNAME=
REDIS_PASSWORD=

# HTTP Server Configuration
HTTP_PORT=3000
HTTP_SERVER_HEADER=MyAppServer/1.0
HTTP_BODY_LIMIT=4194304  # 4MB
```

## Architecture

GOE follows a modular architecture with a central dependency injection container. The main components are:

1. **Core**: Contains the central container and core services
2. **Contracts**: Defines interfaces for all modules
3. **Modules**: Implements specific functionality (cache, mail, etc.)
4. **Middlewares**: HTTP middleware components
5. **Utils**: Utility functions

## Modules

### MongoDB

GOE provides a simple and powerful interface for MongoDB operations through the [`contracts.MongoDB`](https://github.com/oeasenet/goe/blob/main/contracts/mongodb.go) interface. The implementation is built on top of the official MongoDB Go driver with additional features.

#### Defining Models

To work with MongoDB, you need to define models that implement the `IDefaultModel` interface. The easiest way is to embed the `DefaultModel` struct:

```go
import (
    "go.oease.dev/goe/modules/mongodb"
)

// User represents a user in the system
type User struct {
    mongodb.DefaultModel `bson:",inline"`
    Name                 string   `bson:"name" json:"name"`
    Email                string   `bson:"email" json:"email"`
    Age                  int      `bson:"age" json:"age"`
    Roles                []string `bson:"roles" json:"roles"`
}

// ColName returns the MongoDB collection name for this model
func (u *User) ColName() string {
    return "users"
}
```

#### Basic Operations

```go
// Get the MongoDB client
db := goe.UseDB()

// Insert a document
user := &User{
    Name:  "John Doe",
    Email: "john@example.com",
    Age:   30,
    Roles: []string{"user"},
}
result, err := db.Insert(user)
if err != nil {
    // Handle error
}
userID := user.GetId() // Get the inserted document's ID

// Find a document by ID
var foundUser User
found, err := db.FindById(&User{}, userID, &foundUser)
if err != nil {
    // Handle error
}

// Find documents with a filter
var users []User
err = db.Find(&User{}, bson.M{"age": bson.M{"$gt": 18}}).All(&users)
if err != nil {
    // Handle error
}

// Update a document
foundUser.Name = "John Smith"
err = db.Update(&foundUser)
if err != nil {
    // Handle error
}

// Delete a document
err = db.Delete(&foundUser)
if err != nil {
    // Handle error
}
```

#### Pagination

```go
// Get paginated results
page := 1
pageSize := 10
var users []User
totalDocs, totalPages := db.FindPage(&User{}, bson.M{"age": bson.M{"$gt": 18}}, &users, pageSize, page)

// Access pagination information
fmt.Printf("Found %d users across %d pages\n", totalDocs, totalPages)
```

#### Meilisearch Integration

If Meilisearch is enabled, the MongoDB module can automatically sync documents to Meilisearch for full-text search:

```
# Enable Meilisearch integration in your configuration
MEILISEARCH_ENABLED=true
MEILISEARCH_DB_SYNC=true
```

For more details, see the [MongoDB module documentation](https://github.com/oeasenet/goe/tree/main/modules/mongodb).

### Mailer

Send emails easily with multiple provider support:

```go
// Get the mailer
mailer := goe.UseMailer()

// Send an email
err := mailer.DefaultSender().
    To(&[]*mail.Address{{Name: "John Doe", Address: "john@example.com"}}).
    Subject("Hello from GOE").
    HTML("<h1>Hello World</h1>").
    Send()
```

### Cache

Use Redis-based caching:

```go
// Get the cache
cache := goe.UseCache()

// Set a value
err := cache.Set("key", "value", 60) // 60 seconds TTL

// Get a value
val, err := cache.Get("key")
```

### Queue

Process tasks asynchronously:

```go
// Get the queue
queue := goe.UseMQ()

// Publish a task
err := queue.Publish("email", map[string]interface{}{
    "to": "user@example.com",
    "subject": "Welcome",
}, 0)

// Subscribe to a queue
queue.Subscribe("email", func(payload []byte) error {
    // Process the task
    return nil
})
```

### Cron

Schedule tasks:

```go
// Get the cron service
cron := goe.UseCron()

// Add a job
cron.AddJob("0 * * * *", func() {
    // Run every hour
})
```

### Logging

Structured logging:

```go
// Get the logger
logger := goe.UseLog()

// Log messages
logger.Info("This is an info message")
logger.Error("An error occurred", err)
```

### Configuration

Access configuration values:

```go
// Get the config
config := goe.UseCfg()

// Get values
dbName := config.GetOrDefaultString("MONGODB_DB", "default")
port := config.GetOrDefaultInt("HTTP_PORT", 3000)
```

## Middleware

GOE includes several built-in middlewares:

- **File Upload/Download**: Handle file operations
- **Rate Limiter**: Limit request rates
- **Login Check**: Authentication verification
- **OIDC**: OpenID Connect authentication
- **Request Logging**: Log HTTP requests
- **Session**: Session management
- **SPA**: Single Page Application support

Example:

```go
// Use the rate limiter middleware
limiter := middlewares.NewRateLimiter()
goe.UseFiber().App().Use(limiter.Limit())

// Use the session middleware
session := middlewares.NewSession()
goe.UseFiber().App().Use(session.Handle())
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please make sure your code follows the project's coding style and includes appropriate tests.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

GOE relies on or was inspired by the following projects:

- [GoFiber](https://gofiber.io/) - For handling HTTP related tasks
- [GoFr](https://gofr.dev/) - For the project structure and interface design
- [Qmgo](https://github.com/qiniu/qmgo) - For the MongoDB operations
- [Gookit Validate](https://github.com/gookit/validate) - For the data validation
- [PocketBase](https://pocketbase.io/) - For the mailer implementation and interface design
- [Delayqueue](https://github.com/HDT3213/delayqueue) - For the message queue implementation
- [Zap](https://github.com/uber-go/zap) - For the logger implementation
- [EMQX](https://www.emqx.com/) - For the MQTT broker implementation.

## Examples

For more examples, check out the [example directory](https://github.com/oeasenet/goe/tree/main/example).
