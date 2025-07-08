# Basic Application Example

This example shows how to create a simple GOE application with HTTP endpoints.

## Hello World Application

The simplest GOE application with HTTP server:

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2"
)

func main() {
    // Initialize GOE with HTTP module
    goe.New(goe.Options{
        WithHTTP: true,
    })

    // Register routes
    app := goe.HTTP().App()
    app.Get("/", func(c fiber.Ctx) error {
        return c.SendString("Hello, World from GOE!")
    })

    // Start the application
    goe.Run()
}
```

## Basic REST API

Here's a more complete example with multiple endpoints:

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2"
)

func main() {
    // Initialize GOE
    goe.New(goe.Options{
        WithHTTP: true,
    })

    // Get the Fiber app
    app := goe.HTTP().App()

    // Routes
    app.Get("/", handleHome)
    app.Get("/health", handleHealth)
    app.Get("/users", handleUsers)
    app.Post("/users", handleCreateUser)

    // Start the server
    goe.Log().Info("Starting server on :8080")
    goe.Run()
}

func handleHome(c fiber.Ctx) error {
    return c.JSON(fiber.Map{
        "message": "Welcome to GOE API",
        "version": "1.0.0",
    })
}

func handleHealth(c fiber.Ctx) error {
    return c.JSON(fiber.Map{
        "status": "healthy",
        "timestamp": time.Now().Unix(),
    })
}

func handleUsers(c fiber.Ctx) error {
    users := []fiber.Map{
        {"id": 1, "name": "John Doe", "email": "john@example.com"},
        {"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
    }
    return c.JSON(users)
}

func handleCreateUser(c fiber.Ctx) error {
    var user struct {
        Name  string `json:"name" validate:"required"`
        Email string `json:"email" validate:"required,email"`
    }

    if err := c.BodyParser(&user); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    // Log the creation
    goe.Log().Info("Creating user", "name", user.Name, "email", user.Email)

    // Return created user (with mock ID)
    return c.Status(201).JSON(fiber.Map{
        "id":    3,
        "name":  user.Name,
        "email": user.Email,
    })
}
```

## Using Dependency Injection

A more structured approach using dependency injection:

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
)

func main() {
    // Initialize GOE with dependency injection
    goe.New(goe.Options{
        WithHTTP: true,
        Providers: []any{
            NewUserService,
            NewUserHandler,
        },
        Invokers: []any{
            RegisterRoutes,
        },
    })

    // Start the application
    goe.Run()
}

// UserService handles business logic
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

func (s *UserService) GetUsers() []fiber.Map {
    s.logger.Info("Fetching users")
    return []fiber.Map{
        {"id": 1, "name": "John Doe", "email": "john@example.com"},
        {"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
    }
}

func (s *UserService) CreateUser(name, email string) fiber.Map {
    s.logger.Info("Creating user", "name", name, "email", email)
    return fiber.Map{
        "id":    3,
        "name":  name,
        "email": email,
    }
}

// UserHandler handles HTTP requests
type UserHandler struct {
    userService *UserService
    logger      contract.Logger
}

func NewUserHandler(userService *UserService, logger contract.Logger) *UserHandler {
    return &UserHandler{
        userService: userService,
        logger:      logger,
    }
}

func (h *UserHandler) GetUsers(c fiber.Ctx) error {
    users := h.userService.GetUsers()
    return c.JSON(users)
}

func (h *UserHandler) CreateUser(c fiber.Ctx) error {
    var req struct {
        Name  string `json:"name" validate:"required"`
        Email string `json:"email" validate:"required,email"`
    }

    if err := c.BodyParser(&req); err != nil {
        h.logger.Error("Invalid request body", "error", err)
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    user := h.userService.CreateUser(req.Name, req.Email)
    return c.Status(201).JSON(user)
}

// RegisterRoutes sets up HTTP routes
func RegisterRoutes(httpKernel contract.HTTPKernel, userHandler *UserHandler) {
    app := httpKernel.App()

    // API routes
    api := app.Group("/api")
    api.Get("/users", userHandler.GetUsers)
    api.Post("/users", userHandler.CreateUser)

    // Health check
    app.Get("/health", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status": "healthy",
            "service": "goe-api",
        })
    })
}
```

## With Environment Configuration

Example using environment variables:

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
)

func main() {
    goe.New(goe.Options{
        WithHTTP: true,
        Invokers: []any{setupRoutes},
    })

    goe.Run()
}

func setupRoutes(httpKernel contract.HTTPKernel, config contract.Config, logger contract.Logger) {
    app := httpKernel.App()

    // Get configuration values
    appName := config.GetString("APP_NAME")
    version := config.GetString("APP_VERSION")

    app.Get("/", func(c fiber.Ctx) error {
        logger.Info("Home endpoint accessed", "ip", c.IP())
        return c.JSON(fiber.Map{
            "app":     appName,
            "version": version,
            "message": "Welcome to " + appName,
        })
    })

    app.Get("/config", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "app_name":    appName,
            "version":     version,
            "environment": config.GetString("APP_ENV"),
            "port":        config.GetInt("HTTP_PORT"),
        })
    })
}
```

Create a `.env` file:

```bash
APP_NAME=My GOE App
APP_VERSION=1.0.0
APP_ENV=development
HTTP_PORT=8080
LOG_LEVEL=info
```

## Running the Examples

1. **Create a new Go module:**
   ```bash
   mkdir my-goe-app
   cd my-goe-app
   go mod init my-goe-app
   ```

2. **Install GOE:**
   ```bash
   go get go.oease.dev/goe/v2
   ```

3. **Create main.go with any of the examples above**

4. **Run the application:**
   ```bash
   go run main.go
   ```

5. **Test the endpoints:**
   ```bash
   # Test health endpoint
   curl http://localhost:8080/health

   # Test users endpoint
   curl http://localhost:8080/api/users

   # Create a user
   curl -X POST http://localhost:8080/api/users \
     -H "Content-Type: application/json" \
     -d '{"name":"Alice","email":"alice@example.com"}'
   ```

## Key Concepts Demonstrated

1. **Simple Setup**: GOE requires minimal configuration to get started
2. **Dependency Injection**: Use Fx for better code organization and testability
3. **Configuration**: Easy access to environment variables and configuration
4. **Logging**: Structured logging with Zap integration
5. **HTTP Routing**: Full GoFiber routing capabilities
6. **Modular Design**: Enable only the modules you need

## Next Steps

- [**Database CRUD**](./database-crud.md) - Add database operations
- [**Custom Module**](./custom-module.md) - Create your own GOE modules
- [**Authentication**](./authentication.md) - Add authentication middleware