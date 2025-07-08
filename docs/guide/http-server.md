# HTTP Server

GOE provides a powerful HTTP server built on top of GoFiber v3. The HTTP module offers high-performance routing, middleware support, and seamless integration with the dependency injection system.

## Overview

The HTTP module provides:
- **High Performance**: Built on GoFiber v3 and Fasthttp
- **Middleware Support**: Built-in and custom middleware
- **Automatic Logging**: Request/response logging
- **Service Injection**: Access to GOE services in handlers
- **JSON Support**: High-performance JSON serialization with Sonic

## Enabling HTTP Module

Enable the HTTP module in your GOE application:

```go
package main

import "go.oease.dev/goe/v2"

func main() {
    goe.New(goe.Options{
        WithHTTP: true,
    })
    
    goe.Run()
}
```

## Configuration

Configure the HTTP server through environment variables:

```bash
# .env file
HTTP_HOST=0.0.0.0
HTTP_PORT=8080
HTTP_READ_TIMEOUT=30s
HTTP_WRITE_TIMEOUT=30s
HTTP_IDLE_TIMEOUT=60s
```

## Basic Routing

### Using Global Accessor

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2"
)

func main() {
    goe.New(goe.Options{
        WithHTTP: true,
    })
    
    app := goe.HTTP().App()
    
    // Basic routes
    app.Get("/", handleHome)
    app.Get("/users", handleUsers)
    app.Post("/users", handleCreateUser)
    app.Put("/users/:id", handleUpdateUser)
    app.Delete("/users/:id", handleDeleteUser)
    
    goe.Run()
}

func handleHome(c fiber.Ctx) error {
    return c.SendString("Welcome to GOE!")
}

func handleUsers(c fiber.Ctx) error {
    users := []fiber.Map{
        {"id": 1, "name": "John Doe"},
        {"id": 2, "name": "Jane Smith"},
    }
    return c.JSON(users)
}
```

### Using Dependency Injection

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
        Invokers: []any{RegisterRoutes},
    })
    
    goe.Run()
}

func RegisterRoutes(httpKernel contract.HTTPKernel, logger contract.Logger) {
    app := httpKernel.App()
    
    app.Get("/", func(c fiber.Ctx) error {
        logger.Info("Home endpoint accessed", "ip", c.IP())
        return c.SendString("Welcome to GOE!")
    })
    
    app.Get("/health", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status": "healthy",
            "timestamp": time.Now().Unix(),
        })
    })
}
```

## Route Parameters

```go
// Path parameters
app.Get("/users/:id", func(c fiber.Ctx) error {
    id := c.Params("id")
    return c.SendString("User ID: " + id)
})

// Optional parameters
app.Get("/users/:id?", func(c fiber.Ctx) error {
    id := c.Params("id")
    if id == "" {
        return c.SendString("All users")
    }
    return c.SendString("User ID: " + id)
})

// Wildcards
app.Get("/files/*", func(c fiber.Ctx) error {
    path := c.Params("*")
    return c.SendString("File path: " + path)
})
```

## Query Parameters

```go
app.Get("/search", func(c fiber.Ctx) error {
    // Single query parameter
    query := c.Query("q")
    
    // Query parameter with default
    page := c.Query("page", "1")
    
    // Query parameter as int
    limit := c.QueryInt("limit", 10)
    
    return c.JSON(fiber.Map{
        "query": query,
        "page":  page,
        "limit": limit,
    })
})
```

## Request Body Parsing

```go
type User struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

app.Post("/users", func(c fiber.Ctx) error {
    var user User
    
    // Parse JSON body
    if err := c.BodyParser(&user); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }
    
    // Validate struct (requires validator setup)
    if err := validator.Struct(&user); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Validation failed",
            "details": err.Error(),
        })
    }
    
    return c.Status(201).JSON(user)
})
```

## Route Groups

```go
func RegisterRoutes(httpKernel contract.HTTPKernel) {
    app := httpKernel.App()
    
    // API group
    api := app.Group("/api")
    
    // Version group
    v1 := api.Group("/v1")
    
    // Users routes
    users := v1.Group("/users")
    users.Get("/", handleGetUsers)
    users.Post("/", handleCreateUser)
    users.Get("/:id", handleGetUser)
    users.Put("/:id", handleUpdateUser)
    users.Delete("/:id", handleDeleteUser)
    
    // Admin routes with middleware
    admin := v1.Group("/admin", adminMiddleware)
    admin.Get("/stats", handleAdminStats)
}
```

## Middleware

### Built-in Middleware

GOE automatically includes several middleware:

- **Recovery**: Catches panics and returns 500 errors
- **Request ID**: Adds unique request ID for tracing
- **Logger**: Logs HTTP requests and responses
- **Service Injection**: Injects GOE services into context

### Custom Middleware

```go
func authMiddleware(c fiber.Ctx) error {
    token := c.Get("Authorization")
    if token == "" {
        return c.Status(401).JSON(fiber.Map{
            "error": "Authorization header required",
        })
    }
    
    // Validate token
    if !isValidToken(token) {
        return c.Status(401).JSON(fiber.Map{
            "error": "Invalid token",
        })
    }
    
    return c.Next()
}

// Apply middleware to specific routes
app.Get("/protected", authMiddleware, func(c fiber.Ctx) error {
    return c.SendString("Protected content")
})

// Apply middleware to route group
protected := app.Group("/api", authMiddleware)
```

### CORS Middleware

```go
import "github.com/gofiber/fiber/v3/middleware/cors"

func RegisterRoutes(httpKernel contract.HTTPKernel) {
    app := httpKernel.App()
    
    // CORS middleware
    app.Use(cors.New(cors.Config{
        AllowOrigins: "https://myapp.com",
        AllowMethods: "GET,POST,PUT,DELETE",
        AllowHeaders: "Content-Type,Authorization",
    }))
}
```

## Service Injection in Handlers

GOE provides helper functions to access services in handlers:

```go
import "go.oease.dev/goe/v2/core/http"

app.Get("/users", func(c fiber.Ctx) error {
    // Get services from context
    logger := http.GetLogger(c)
    config := http.GetConfig(c)
    
    logger.Info("Getting users", "ip", c.IP())
    
    appName := config.GetString("APP_NAME")
    
    return c.JSON(fiber.Map{
        "app": appName,
        "users": []string{"user1", "user2"},
    })
})
```

## Error Handling

### Custom Error Handler

```go
func RegisterRoutes(httpKernel contract.HTTPKernel) {
    app := httpKernel.App()
    
    // Custom error handler
    app.ErrorHandler = func(c fiber.Ctx, err error) error {
        code := fiber.StatusInternalServerError
        
        if e, ok := err.(*fiber.Error); ok {
            code = e.Code
        }
        
        return c.Status(code).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
}
```

### Structured Error Responses

```go
type ErrorResponse struct {
    Error   string `json:"error"`
    Code    int    `json:"code"`
    Message string `json:"message"`
}

func handleError(c fiber.Ctx, err error, code int) error {
    response := ErrorResponse{
        Error:   err.Error(),
        Code:    code,
        Message: "An error occurred",
    }
    
    return c.Status(code).JSON(response)
}

app.Get("/users/:id", func(c fiber.Ctx) error {
    id := c.Params("id")
    
    user, err := getUserByID(id)
    if err != nil {
        return handleError(c, err, 404)
    }
    
    return c.JSON(user)
})
```

## File Upload

```go
app.Post("/upload", func(c fiber.Ctx) error {
    file, err := c.FormFile("file")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "No file uploaded",
        })
    }
    
    // Save file
    if err := c.SaveFile(file, "./uploads/"+file.Filename); err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to save file",
        })
    }
    
    return c.JSON(fiber.Map{
        "message": "File uploaded successfully",
        "filename": file.Filename,
        "size": file.Size,
    })
})
```

## Static Files

```go
func RegisterRoutes(httpKernel contract.HTTPKernel) {
    app := httpKernel.App()
    
    // Serve static files
    app.Static("/", "./public")
    
    // Serve static files with prefix
    app.Static("/assets", "./assets")
    
    // Serve single file
    app.Get("/favicon.ico", func(c fiber.Ctx) error {
        return c.SendFile("./assets/favicon.ico")
    })
}
```

## WebSocket Support

```go
import "github.com/gofiber/fiber/v3/middleware/websocket"

func RegisterRoutes(httpKernel contract.HTTPKernel) {
    app := httpKernel.App()
    
    // WebSocket middleware
    app.Use("/ws", func(c fiber.Ctx) error {
        if websocket.IsWebSocketUpgrade(c) {
            return c.Next()
        }
        return fiber.ErrUpgradeRequired
    })
    
    // WebSocket handler
    app.Get("/ws", websocket.New(func(c *websocket.Conn) {
        defer c.Close()
        
        for {
            messageType, message, err := c.ReadMessage()
            if err != nil {
                break
            }
            
            if err := c.WriteMessage(messageType, message); err != nil {
                break
            }
        }
    }))
}
```

## Handler Patterns

### Controller Pattern

```go
type UserController struct {
    userService UserService
    logger      contract.Logger
}

func NewUserController(userService UserService, logger contract.Logger) *UserController {
    return &UserController{
        userService: userService,
        logger:      logger,
    }
}

func (uc *UserController) GetUsers(c fiber.Ctx) error {
    users, err := uc.userService.GetAllUsers()
    if err != nil {
        uc.logger.Error("Failed to get users", "error", err)
        return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
    }
    
    return c.JSON(users)
}

func (uc *UserController) CreateUser(c fiber.Ctx) error {
    var req CreateUserRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    user, err := uc.userService.CreateUser(req.Name, req.Email)
    if err != nil {
        uc.logger.Error("Failed to create user", "error", err)
        return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
    }
    
    return c.Status(201).JSON(user)
}

// Register controller routes
func RegisterRoutes(httpKernel contract.HTTPKernel, userController *UserController) {
    app := httpKernel.App()
    
    users := app.Group("/users")
    users.Get("/", userController.GetUsers)
    users.Post("/", userController.CreateUser)
}
```

### Handler with Dependency Injection

```go
func NewUserHandler(userService UserService, logger contract.Logger) func(c fiber.Ctx) error {
    return func(c fiber.Ctx) error {
        users, err := userService.GetAllUsers()
        if err != nil {
            logger.Error("Failed to get users", "error", err)
            return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
        }
        
        return c.JSON(users)
    }
}

func RegisterRoutes(httpKernel contract.HTTPKernel, userService UserService, logger contract.Logger) {
    app := httpKernel.App()
    
    app.Get("/users", NewUserHandler(userService, logger))
}
```

## Testing HTTP Handlers

```go
import (
    "net/http/httptest"
    "testing"
    "github.com/gofiber/fiber/v3"
    "github.com/stretchr/testify/assert"
)

func TestUserHandler(t *testing.T) {
    app := fiber.New()
    
    // Mock services
    mockUserService := &MockUserService{}
    mockLogger := &MockLogger{}
    
    // Register routes
    RegisterRoutes(app, mockUserService, mockLogger)
    
    // Test request
    req := httptest.NewRequest("GET", "/users", nil)
    resp, err := app.Test(req)
    
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}
```

## Performance Optimization

### JSON Serialization

GOE uses Sonic for high-performance JSON serialization:

```go
// Automatic JSON serialization with Sonic
app.Get("/users", func(c fiber.Ctx) error {
    users := []User{
        {ID: 1, Name: "John"},
        {ID: 2, Name: "Jane"},
    }
    return c.JSON(users) // Uses Sonic internally
})
```

### Response Compression

```go
import "github.com/gofiber/fiber/v3/middleware/compress"

func RegisterRoutes(httpKernel contract.HTTPKernel) {
    app := httpKernel.App()
    
    // Enable compression
    app.Use(compress.New(compress.Config{
        Level: compress.LevelBestSpeed,
    }))
}
```

## Best Practices

### 1. Use Dependency Injection

```go
// Good - explicit dependencies
func NewUserHandler(userService UserService, logger contract.Logger) *UserHandler {
    return &UserHandler{userService: userService, logger: logger}
}

// Avoid - global dependencies
func handleUsers(c fiber.Ctx) error {
    users := goe.UserService().GetAll() // Harder to test
    return c.JSON(users)
}
```

### 2. Validate Input

```go
func handleCreateUser(c fiber.Ctx) error {
    var req CreateUserRequest
    
    // Parse body
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
    }
    
    // Validate
    if err := validator.Struct(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Validation failed"})
    }
    
    // Process request
    return nil
}
```

### 3. Handle Errors Properly

```go
func handleGetUser(c fiber.Ctx) error {
    id := c.Params("id")
    
    user, err := userService.GetUser(id)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            return c.Status(404).JSON(fiber.Map{"error": "User not found"})
        }
        
        logger.Error("Failed to get user", "error", err, "id", id)
        return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
    }
    
    return c.JSON(user)
}
```

### 4. Use Middleware for Cross-Cutting Concerns

```go
// Authentication middleware
func authMiddleware(c fiber.Ctx) error {
    // Authentication logic
    return c.Next()
}

// Logging middleware
func loggingMiddleware(logger contract.Logger) fiber.Handler {
    return func(c fiber.Ctx) error {
        start := time.Now()
        
        err := c.Next()
        
        logger.Info("HTTP request",
            "method", c.Method(),
            "path", c.Path(),
            "duration", time.Since(start),
        )
        
        return err
    }
}
```

## Next Steps

- [**Database**](./database.md) - Learn about database integration
- [**Testing**](./testing.md) - Test your HTTP handlers
- [**Best Practices**](./best-practices.md) - HTTP development best practices