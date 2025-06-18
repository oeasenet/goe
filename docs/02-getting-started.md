# 2. Getting Started with Goe 🛠️

This guide will walk you through installing Goe, setting up your first project, and running a basic "Hello, World" HTTP server. We'll also touch upon the recommended project structure and how you can interact with Goe's modules.

## ✅ Prerequisites

*   **Go**: Goe requires Go version 1.24 or newer. You can download it from [golang.org](https://golang.org/dl/).
    *   To check your Go version: `go version`

## 📦 Installation

To add Goe to your Go project, use `go get`:

```bash
go get go.oease.dev/goe/v2
```

This command will fetch the latest stable version of the Goe framework.

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

Let's create a simple HTTP server that responds with "Hello, World!".

Create a file named `main.go` (or `cmd/server/main.go` if you're following a more complex structure) with the following content:

```go
package main

import (
	"github.com/gofiber/fiber/v3" // Import Fiber
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract" // For contract.HTTPKernel if using DI for routes
)

// main is the entry point of our application.
func main() {
	// 1. Initialize a new Goe application instance.
	//    goe.Options allows you to enable specific modules.
	//    Here, we enable the HTTP server.
	appInstance := goe.New(goe.Options{
		WithHTTP: true, // Enable the HTTP module
		// We can also register routes directly here using an Fx invoker
		Invokers: []any{
			RegisterRoutes,
		},
	})

	// Check if initialization failed (e.g., critical Fx setup error)
	if appInstance == nil || goe.App().Container().Err() != nil {
		// A basic logger isn't available yet if goe.New fails badly,
		// so use standard log.Fatalf or similar.
		// However, if only goe.Run() fails, the goe.Log() might be available.
		if goe.App().Container() != nil && goe.App().Container().Err() != nil {
			goe.Log().Fatal("Failed to initialize Goe application", "error", goe.App().Container().Err())
		} else {
			// A more primitive logging for very early failures
			println("Critical error: Failed to create Goe application instance.")
		}
		return
	}

	// 2. Run the application.
	//    This starts all registered modules (like the HTTP server)
	//    and blocks until the application is shut down (e.g., by SIGINT).
	goe.Run()

	// After goe.Run() completes (application shutdown):
	goe.Log().Info("Application has shut down gracefully.")
}

// RegisterRoutes is an Fx invoker function that sets up our HTTP routes.
// Fx will automatically provide the dependencies (contract.HTTPKernel and contract.Logger).
func RegisterRoutes(httpKernel contract.HTTPKernel, logger contract.Logger) {
	// Get the underlying Fiber app
	fiberApp := httpKernel.App()

	// Define a simple route
	fiberApp.Get("/", func(c fiber.Ctx) error {
		logger.Info("Received request for /", "remote_ip", c.IP())
		return c.SendString("Hello, World from Goe! 👋")
	})

	logger.Info("Successfully registered HTTP routes.")
}
```

**Explanation:**

*   **`goe.New(goe.Options{...})`**: This initializes the Goe application.
    *   `WithHTTP: true`: This option tells Goe to initialize and start its built-in HTTP server module (which uses GoFiber).
    *   `Invokers: []any{RegisterRoutes}`: This is a powerful feature of Uber's Fx (which Goe uses internally). An "invoker" is a function that Fx will call during application startup. Fx automatically injects any dependencies this function needs. Here, `RegisterRoutes` needs `contract.HTTPKernel` (to access the Fiber app) and `contract.Logger`.
*   **`RegisterRoutes(...)`**: This function defines our HTTP routes.
    *   It takes `contract.HTTPKernel` and `contract.Logger` as parameters. Fx provides these.
    *   `httpKernel.App()` gives us the underlying `*fiber.App` instance from GoFiber.
    *   `fiberApp.Get("/", ...)` defines a GET route for the path `/`.
*   **`goe.Run()`**: This function starts the application. It initializes all modules, runs their `OnStart` hooks, and then blocks, typically keeping the HTTP server listening for requests. It also handles graceful shutdown.
*   **Logging**: We inject `contract.Logger` into `RegisterRoutes` to log when a request comes in. The HTTP module also has its own request logging.

## 🚀 Running the Application

1.  **Open your terminal** and navigate to your project directory (`goe-hello-world`).
2.  **Run the `main.go` file:**
    ```bash
    go run main.go
    ```
    (Or `go run cmd/server/main.go` if you used that structure)

You should see output similar to this (the exact log format might differ based on your environment):

```
INFO	HTTP server starting	{"address": "0.0.0.0:8080"}
INFO	Successfully registered HTTP routes.
INFO	Log module started
INFO	Application started
```

3.  **Open your web browser** or use a tool like `curl` to access `http://localhost:8080`.

    ```bash
    curl http://localhost:8080
    ```

You should see the response: `Hello, World from Goe! 👋`

In your terminal, you'll also see the log message from your handler:
```
INFO	Received request for /	{"remote_ip": "127.0.0.1"}
INFO	HTTP Request	{"method": "GET", "path": "/", "status": 200, "duration": "...", "request_id": "..."}
```
(The second log line is from Goe's built-in HTTP request logger).

4.  **To stop the application**, go back to your terminal and press `Ctrl+C`. You should see shutdown messages.

## 🧩 Two Ways to Use Modules

Goe provides flexibility in how you access its core components (like the logger, config, HTTP kernel, etc.):

1.  **Global Accessors (Convenience 🍬):**
    Goe offers global functions for easy access, e.g., `goe.Log()`, `goe.Config()`, `goe.HTTP()`. These are handy for quick scripting, smaller applications, or in places where dependency injection is cumbersome.

    ```go
    // Example using global logger
    goe.Log().Info("This is a log message using the global accessor.")
    ```

2.  **Dependency Injection (Robustness & Testability 🏗️):**
    For larger applications, testing, and better separation of concerns, Goe fully supports and encourages constructor/method injection via Uber's Fx. You define your components (services, handlers) as functions or structs that declare their dependencies, and Fx provides them. This was demonstrated in the `RegisterRoutes` function above.

    ```go
    // From our main example:
    // func RegisterRoutes(httpKernel contract.HTTPKernel, logger contract.Logger) { ... }
    // Fx automatically provides httpKernel and logger.
    ```

You'll see both patterns used in the documentation and examples. Choosing which to use depends on the specific context and your project's needs. Dependency injection is generally preferred for application logic that requires testing and maintainability.

## Next Steps

You've successfully set up and run your first Goe application! Now you're ready to explore more advanced topics:

*   [Project Structure](03-project-structure.md): Dive deeper into organizing your Goe projects.
*   [Architecture Deep Dive](04-architecture.md): Understand the core design of Goe.
*   [Configuration](05-configuration.md): Learn how to manage application settings.
```
