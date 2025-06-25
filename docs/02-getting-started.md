# 2. Getting Started with Goe 🛠️

This guide will walk you through installing Goe, setting up your first project, and running a basic "Hello, World" HTTP server. We'll also explore the modular architecture and show you different ways to structure your application.

## ✅ Prerequisites

*   **Go**: Goe requires Go version 1.21 or newer. You can download it from [golang.org](https://golang.org/dl/).
    *   To check your Go version: `go version`

## 📦 Installation

To add Goe to your Go project, use `go get`:

```bash
go get go.oease.dev/goe/v2
```

This command will fetch the latest stable version of the Goe framework.

## 🎯 Core Concepts

Before diving into code, let's understand Goe's key concepts:

- **Modular Design**: Enable only the components you need (`WithHTTP`, `WithCache`, `WithDB`)
- **Contracts**: Interface-based design for loose coupling and easy testing
- **Dual Access**: Use global accessors (`goe.Log()`) or dependency injection
- **Lifecycle Management**: Automatic startup/shutdown handling for all components

## 🏗️ Creating Your First Project

While Goe doesn't enforce a specific project structure, we recommend a layout that promotes clarity and scalability. A common and effective structure is:

```
your-project-name/
├── cmd/                    # Main application(s) entry points
│   └── myapp/              # Example: your web application
│       └── main.go
├── internal/               # Private application and library code
│   ├── config/             # Configuration loading, struct definitions
│   ├── core/               # Core business logic, domain types
│   ├── handler/            # HTTP handlers or other interface handlers
│   ├── module/             # Custom Goe modules for your application
│   ├── repository/         # Data access logic (database, external APIs)
│   └── service/            # Business logic services
├── pkg/                    # Public library code (if any, shareable)
├── configs/                # Configuration files (e.g., .env.example)
├── docs/                   # Project documentation
├── scripts/                # Helper scripts (build, deploy, etc.)
├── web/                    # Frontend assets (templates, static files)
│   ├── static/
│   └── templates/
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
└── .gitignore              # Git ignore file
```

For this guide, we'll create a simpler structure for our "Hello, World" example.

1.  **Create a project directory:**
    ```bash
    mkdir goe-hello-world
    cd goe-hello-world
    ```

2.  **Initialize Go modules:**
    ```bash
    go mod init example.com/goe-hello-world
    # Replace example.com/goe-hello-world with your actual module path
    ```

3.  **Install Goe (if not already done globally or for another project):**
    ```bash
    go get go.oease.dev/goe/v2
    ```

## 👋 Hello, World! - Your First Goe Application

Let's create a simple HTTP server that responds with "Hello, World!". We'll show you two approaches: using global accessors (simple) and dependency injection (recommended for larger applications).

### Approach 1: Simple Global Accessors

Create a file named `main.go` with the following content:

```go
package main

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
)

func main() {
	// Initialize Goe with HTTP module enabled
	goe.New(goe.Options{
		WithHTTP: true,
	})

	// Get the HTTP kernel and register routes
	app := goe.HTTP().App()
	app.Get("/", func(c fiber.Ctx) error {
		goe.Log().Info("Received request", "path", c.Path(), "ip", c.IP())
		return c.SendString("Hello, World from Goe! 👋")
	})

	// Start the application
	goe.Log().Info("Starting Goe application")
	goe.Run()
}
```

### Approach 2: Dependency Injection (Recommended)

For larger applications, use dependency injection for better testability:

```go
package main

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
)

func main() {
	// Initialize Goe with HTTP module and route registration
	goe.New(goe.Options{
		WithHTTP:  true,
		Invokers: []any{RegisterRoutes},
	})

	// Start the application
	goe.Run()
}

// RegisterRoutes is automatically called by Fx with injected dependencies
func RegisterRoutes(httpKernel contract.HTTPKernel, logger contract.Logger) {
	app := httpKernel.App()

	// Register routes
	app.Get("/", func(c fiber.Ctx) error {
		logger.Info("Received request", "path", c.Path(), "ip", c.IP())
		return c.SendString("Hello, World from Goe! 👋")
	})

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"service": "goe-app",
		})
	})

	logger.Info("HTTP routes registered successfully")
}
```

**Key Concepts Explained:**

### Approach 1 (Global Accessors)
- **`goe.New()`**: Initializes the application with specified modules
- **`WithHTTP: true`**: Enables the HTTP server module (built on GoFiber)
- **`goe.HTTP().App()`**: Gets the underlying Fiber app instance for route registration
- **`goe.Log()`**: Accesses the global logger instance
- **`goe.Run()`**: Starts all modules and blocks until shutdown

### Approach 2 (Dependency Injection)
- **`Invokers`**: Functions called during startup with automatic dependency injection
- **`contract.HTTPKernel`**: Interface for HTTP server operations
- **`contract.Logger`**: Interface for structured logging
- **Fx Integration**: Uber's Fx handles all dependency wiring automatically

### Benefits of Each Approach

**Global Accessors** (Approach 1):
- ✅ Simple and quick for small applications
- ✅ Less boilerplate code
- ✅ Easy to understand for beginners
- ❌ Harder to test (global state)
- ❌ Less explicit dependencies

**Dependency Injection** (Approach 2):
- ✅ Better for larger applications
- ✅ Easier to test (mockable dependencies)
- ✅ Explicit dependency management
- ✅ Better separation of concerns
- ❌ More initial setup
- ❌ Steeper learning curve

## 🚀 Running the Application

1. **Navigate to your project directory:**
   ```bash
   cd goe-hello-world
   ```

2. **Run the application:**
   ```bash
   go run main.go
   ```

3. **Expected output:**
   You should see structured log output similar to this:
   ```
   INFO    Starting Goe application
   INFO    HTTP routes registered successfully
   INFO    HTTP server starting    {"host": "0.0.0.0", "port": 8080}
   INFO    Application started successfully
   ```

4. **Test your application:**
   Open your browser or use curl to test the endpoints:
   ```bash
   # Test the main endpoint
   curl http://localhost:8080/
   # Response: Hello, World from Goe! 👋

   # Test the health endpoint (if using Approach 2)
   curl http://localhost:8080/health
   # Response: {"status":"healthy","service":"goe-app"}
   ```

5. **View request logs:**
   Each request will generate structured logs:
   ```
   INFO    Received request    {"path": "/", "ip": "127.0.0.1"}
   INFO    HTTP Request        {"method": "GET", "path": "/", "status": 200, "duration": "123μs", "request_id": "abc123"}
   ```

6. **Graceful shutdown:**
   Press `Ctrl+C` to stop the application. You'll see shutdown logs:
   ```
   INFO    Shutting down HTTP server
   INFO    Application stopped gracefully
   ```

## 🔧 Configuration and Environment

Goe applications can be configured using environment variables or `.env` files:

```bash
# .env file
HTTP_HOST=0.0.0.0
HTTP_PORT=8080
LOG_LEVEL=info
APP_NAME=My Goe App
APP_VERSION=1.0.0
```

Access configuration in your code:

```go
// Using global accessor
port := goe.Config().GetInt("HTTP_PORT")
appName := goe.Config().GetString("APP_NAME")

// Using dependency injection
func MyService(config contract.Config) *Service {
    port := config.GetInt("HTTP_PORT")
    return &Service{port: port}
}
```

## 🧩 Adding More Modules

Enable additional modules as needed:

```go
app := goe.New(goe.Options{
    WithHTTP:  true,  // Web server
    WithCache: true,  // Caching system
    WithDB:    true,  // Database integration
    Providers: []any{
        NewUserService,    // Your custom services
        NewOrderService,
    },
    Invokers: []any{
        RegisterRoutes,    // Route registration
        SetupMiddleware,   // Middleware setup
    },
})
```

## 🧪 Testing Your Application

Goe makes testing easy with dependency injection:

```go
func TestUserService(t *testing.T) {
    // Create mocks
    mockDB := &MockDB{}
    mockLogger := &MockLogger{}

    // Create service with mocked dependencies
    service := NewUserService(mockDB, mockLogger)

    // Test your service
    user, err := service.GetUser("123")
    assert.NoError(t, err)
    assert.Equal(t, "John", user.Name)
}
```

## 🚀 Next Steps

Congratulations! You've successfully created your first Goe application. Here's what to explore next:

### Core Concepts
- [Project Structure](03-project-structure.md) - Organize your Goe projects effectively
- [Architecture](04-architecture.md) - Understand Goe's design principles
- [Configuration](05-configuration.md) - Master environment and config management

### Essential Components
- [Logging](06-logging.md) - Structured logging with Zap
- [HTTP Server](07-http-server.md) - Build robust web applications
- [Database](08-database.md) - GORM integration and best practices
- [Caching](09-caching.md) - Improve performance with caching

### Advanced Topics
- [Modules](10-modules.md) - Create custom modules
- [Dependency Injection](11-dependency-injection.md) - Master Fx patterns
- [Testing](13-testing.md) - Comprehensive testing strategies

### Production Ready
- [Best Practices](15-best-practices.md) - Production-ready development
- [Deployment](16-deployment.md) - Deploy your Goe applications

Ready to dive deeper? Start with [Project Structure](03-project-structure.md) to learn how to organize larger applications!
