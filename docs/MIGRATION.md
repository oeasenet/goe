# Migration Guide

This guide helps developers migrate to Goe from other popular frameworks.

## Table of Contents

- [From Gin to Goe](#from-gin-to-goe)
- [From Echo to Goe](#from-echo-to-goe)
- [From Chi to Goe](#from-chi-to-goe)
- [From Goravel to Goe](#from-goravel-to-goe)
- [From Standard Library to Goe](#from-standard-library-to-goe)
- [Common Migration Patterns](#common-migration-patterns)

## From Gin to Goe

### Basic Application

**Gin:**
```go
package main

import (
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })
    
    r.Run(":8080")
}
```

**Goe:**
```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
)

func main() {
    _ = goe.New(goe.Options{
        Name:     "My App",
        WithHTTP: true,
        Invokers: []any{
            func(http contract.HTTPKernel) {
                app := http.App()
                
                app.Get("/ping", func(c fiber.Ctx) error {
                    return c.JSON(fiber.Map{
                        "message": "pong",
                    })
                })
            },
        },
    })
    
    goe.Run()
}
```

### Middleware

**Gin:**
```go
// Logger middleware
r.Use(gin.Logger())

// Custom middleware
r.Use(func(c *gin.Context) {
    c.Set("user_id", "123")
    c.Next()
})
```

**Goe:**
```go
// Logger middleware is built-in

// Custom middleware
app.Use(func(c fiber.Ctx) error {
    c.Locals("user_id", "123")
    return c.Next()
})
```

### Route Parameters

**Gin:**
```go
r.GET("/users/:id", func(c *gin.Context) {
    id := c.Param("id")
    c.JSON(200, gin.H{"id": id})
})
```

**Goe:**
```go
app.Get("/users/:id", func(c fiber.Ctx) error {
    id := c.Params("id")
    return c.JSON(fiber.Map{"id": id})
})
```

### Request Binding

**Gin:**
```go
type CreateUserRequest struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

r.POST("/users", func(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    // Process request
})
```

**Goe:**
```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

app.Post("/users", func(c fiber.Ctx) error {
    var req CreateUserRequest
    if err := c.Bind().Body(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }
    
    // Validate separately
    if err := validate.Struct(req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }
    // Process request
    return nil
})
```

## From Echo to Goe

### Basic Setup

**Echo:**
```go
package main

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

func main() {
    e := echo.New()
    
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    
    e.GET("/", func(c echo.Context) error {
        return c.JSON(http.StatusOK, map[string]string{
            "message": "Hello, World!",
        })
    })
    
    e.Logger.Fatal(e.Start(":8080"))
}
```

**Goe:**
```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
)

func main() {
    _ = goe.New(goe.Options{
        Name:     "My App",
        WithHTTP: true,
        Invokers: []any{
            func(http contract.HTTPKernel) {
                app := http.App()
                // Logger and Recover are built-in
                
                app.Get("/", func(c fiber.Ctx) error {
                    return c.JSON(fiber.Map{
                        "message": "Hello, World!",
                    })
                })
            },
        },
    })
    
    goe.Run()
}
```

### Context Usage

**Echo:**
```go
e.GET("/users/:id", func(c echo.Context) error {
    id := c.Param("id")
    
    // Custom context values
    c.Set("user_id", id)
    userID := c.Get("user_id").(string)
    
    // Logger
    c.Logger().Info("Getting user")
    
    return c.JSON(http.StatusOK, map[string]string{"id": id})
})
```

**Goe:**
```go
app.Get("/users/:id", func(c fiber.Ctx) error {
    id := c.Params("id")
    
    // Custom context values
    c.Locals("user_id", id)
    userID := c.Locals("user_id").(string)
    
    // Logger with context
    logger := http.GetLogger(c)
    logger.Info("Getting user")
    
    return c.JSON(fiber.Map{"id": id})
})
```

### Error Handling

**Echo:**
```go
e.HTTPErrorHandler = func(err error, c echo.Context) {
    code := http.StatusInternalServerError
    if he, ok := err.(*echo.HTTPError); ok {
        code = he.Code
    }
    c.Logger().Error(err)
    c.JSON(code, map[string]string{"error": err.Error()})
}
```

**Goe:**
```go
// Error handling is built-in, but can be customized:
app := fiber.New(fiber.Config{
    ErrorHandler: func(c fiber.Ctx, err error) error {
        code := fiber.StatusInternalServerError
        if e, ok := err.(*fiber.Error); ok {
            code = e.Code
        }
        
        logger := http.GetLogger(c)
        logger.Error("Request error", log.NewField("error", err))
        
        return c.Status(code).JSON(fiber.Map{"error": err.Error()})
    },
})
```

## From Chi to Goe

### Router Setup

**Chi:**
```go
package main

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func main() {
    r := chi.NewRouter()
    
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.RequestID)
    
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello World!"))
    })
    
    r.Route("/users", func(r chi.Router) {
        r.Get("/", listUsers)
        r.Post("/", createUser)
        
        r.Route("/{userID}", func(r chi.Router) {
            r.Use(UserCtx)
            r.Get("/", getUser)
            r.Put("/", updateUser)
            r.Delete("/", deleteUser)
        })
    })
    
    http.ListenAndServe(":8080", r)
}
```

**Goe:**
```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
)

func main() {
    _ = goe.New(goe.Options{
        Name:     "My App",
        WithHTTP: true,
        Invokers: []any{
            func(http contract.HTTPKernel) {
                app := http.App()
                // Logger, Recover, RequestID are built-in
                
                app.Get("/", func(c fiber.Ctx) error {
                    return c.SendString("Hello World!")
                })
                
                users := app.Group("/users")
                users.Get("/", listUsers)
                users.Post("/", createUser)
                
                user := users.Group("/:userID", UserMiddleware)
                user.Get("/", getUser)
                user.Put("/", updateUser)
                user.Delete("/", deleteUser)
            },
        },
    })
    
    goe.Run()
}
```

### Context Pattern

**Chi:**
```go
func UserCtx(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := chi.URLParam(r, "userID")
        user, err := dbGetUser(userID)
        if err != nil {
            http.Error(w, http.StatusText(404), 404)
            return
        }
        ctx := context.WithValue(r.Context(), "user", user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func getUser(w http.ResponseWriter, r *http.Request) {
    user := r.Context().Value("user").(*User)
    // Use user
}
```

**Goe:**
```go
func UserMiddleware(c fiber.Ctx) error {
    userID := c.Params("userID")
    user, err := dbGetUser(userID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "User not found"})
    }
    c.Locals("user", user)
    return c.Next()
}

func getUser(c fiber.Ctx) error {
    user := c.Locals("user").(*User)
    // Use user
    return c.JSON(user)
}
```

## From Goravel to Goe

### Application Structure

**Goravel:**
```go
// main.go
package main

import "github.com/goravel/framework/facades"

func main() {
    facades.Boot()
    facades.Route.Get("/", func(ctx facades.Context) {
        ctx.Response().Json(200, facades.Map{
            "message": "Hello Goravel",
        })
    })
    facades.Artisan.Call("serve")
}
```

**Goe:**
```go
// main.go
package main

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
)

func main() {
    _ = goe.New(goe.Options{
        Name:     "My App",
        WithHTTP: true,
        Invokers: []any{
            func(http contract.HTTPKernel) {
                app := http.App()
                app.Get("/", func(c fiber.Ctx) error {
                    return c.JSON(fiber.Map{
                        "message": "Hello Goe",
                    })
                })
            },
        },
    })
    
    goe.Run()
}
```

### Service Providers

**Goravel:**
```go
// app/providers/app_service_provider.go
type AppServiceProvider struct{}

func (receiver *AppServiceProvider) Register() {
    facades.App.Bind("user.service", func(app foundation.Application) (any, error) {
        return &UserService{
            db: facades.DB,
        }, nil
    })
}

func (receiver *AppServiceProvider) Boot() {
    // Bootstrap services
}
```

**Goe:**
```go
// Using Fx providers
func NewUserService(db *sql.DB) *UserService {
    return &UserService{db: db}
}

_ = goe.New(goe.Options{
    Providers: []any{
        NewDatabase,
        NewUserService,
    },
})
```

### Configuration

**Goravel:**
```go
// config/app.go
facades.Config.Get("app.name")
facades.Config.GetString("app.env")
facades.Config.GetBool("app.debug")
```

**Goe:**
```go
// Direct access
config := goe.Config()
config.GetString("APP_NAME")
config.GetString("APP_ENV")
config.GetBool("DEBUG")

// In handlers
func handler(c fiber.Ctx) error {
    config := http.GetConfig(c)
    appName := config.GetString("APP_NAME")
    return c.SendString(appName)
}
```

### Database/ORM

**Goravel:**
```go
// Using Goravel's ORM
var user models.User
facades.DB.Where("id", id).First(&user)

// Transactions
facades.DB.Transaction(func(tx facades.Transaction) error {
    // Operations
    return nil
})
```

**Goe:**
```go
// Bring your own ORM (e.g., GORM, sqlx)
type Repository struct {
    db *sql.DB // or *gorm.DB
}

func (r *Repository) GetUser(id string) (*User, error) {
    var user User
    err := r.db.QueryRow("SELECT * FROM users WHERE id = ?", id).Scan(&user)
    return &user, err
}

// Dependency injection
func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}
```

## From Standard Library to Goe

### Basic HTTP Server

**Standard Library:**
```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
)

type Response struct {
    Message string `json:"message"`
}

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(Response{Message: "Hello, World!"})
    })
    
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**Goe:**
```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
)

func main() {
    _ = goe.New(goe.Options{
        Name:     "My App",
        WithHTTP: true,
        Invokers: []any{
            func(http contract.HTTPKernel) {
                app := http.App()
                
                app.Get("/", func(c fiber.Ctx) error {
                    return c.JSON(fiber.Map{
                        "message": "Hello, World!",
                    })
                })
            },
        },
    })
    
    goe.Log().Info("Server starting on :8080")
    goe.Run()
}
```

### Middleware Pattern

**Standard Library:**
```go
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    }
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    }
}

func main() {
    http.HandleFunc("/", loggingMiddleware(authMiddleware(handler)))
}
```

**Goe:**
```go
// Logging is built-in

func authMiddleware(c fiber.Ctx) error {
    token := c.Get("Authorization")
    if token == "" {
        return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
    }
    return c.Next()
}

func setupRoutes(http contract.HTTPKernel) {
    app := http.App()
    app.Get("/", authMiddleware, handler)
}
```

## Common Migration Patterns

### 1. Dependency Injection Migration

**Before (Global Variables):**
```go
var (
    db     *sql.DB
    logger *log.Logger
    config *Config
)

func initServices() {
    db = initDB()
    logger = initLogger()
    config = initConfig()
}

func handler(w http.ResponseWriter, r *http.Request) {
    logger.Println("Handling request")
    // Use global db, config
}
```

**After (Dependency Injection):**
```go
type Controller struct {
    db     *sql.DB
    logger contract.Logger
    config contract.Config
}

func NewController(params struct {
    fx.In
    DB     *sql.DB
    Logger contract.Logger
    Config contract.Config
}) *Controller {
    return &Controller{
        db:     params.DB,
        logger: params.Logger,
        config: params.Config,
    }
}

func (c *Controller) Handler(ctx fiber.Ctx) error {
    c.logger.Info("Handling request")
    // Use injected dependencies
    return nil
}
```

### 2. Configuration Migration

**Before (Viper/env files):**
```go
viper.SetConfigFile(".env")
viper.ReadInConfig()

dbHost := viper.GetString("DB_HOST")
dbPort := viper.GetInt("DB_PORT")
```

**After (Goe Config):**
```go
config := goe.Config()
dbHost := config.GetString("DB_HOST")
dbPort := config.GetInt("DB_PORT")

// Or with DI
func NewService(config contract.Config) *Service {
    return &Service{
        dbHost: config.GetString("DB_HOST"),
        dbPort: config.GetInt("DB_PORT"),
    }
}
```

### 3. Logging Migration

**Before (Various Loggers):**
```go
// logrus
log.WithFields(logrus.Fields{
    "user_id": userID,
    "action":  "login",
}).Info("User logged in")

// zap
logger.Info("User logged in",
    zap.String("user_id", userID),
    zap.String("action", "login"),
)

// standard library
log.Printf("User %s logged in", userID)
```

**After (Goe Logger):**
```go
logger := goe.Log()
logger.Info("User logged in",
    log.NewField("user_id", userID),
    log.NewField("action", "login"),
)

// In HTTP handlers
func handler(c fiber.Ctx) error {
    logger := http.GetLogger(c) // Includes request_id
    logger.Info("Processing request")
    return nil
}
```

### 4. Testing Migration

**Before:**
```go
func TestHandler(t *testing.T) {
    // Manual setup
    db := setupTestDB()
    logger := setupTestLogger()
    
    req := httptest.NewRequest("GET", "/users", nil)
    w := httptest.NewRecorder()
    
    handler(w, req)
    
    assert.Equal(t, 200, w.Code)
}
```

**After:**
```go
func TestHandler(t *testing.T) {
    // Use Fx for test setup
    var controller *Controller
    
    app := fx.New(
        fx.Provide(
            NewMockDB,
            log.NewTestLogger,
            config.NewTestConfig,
            NewController,
        ),
        fx.Populate(&controller),
    )
    
    require.NoError(t, app.Start(context.Background()))
    defer app.Stop(context.Background())
    
    // Test with Fiber
    app := fiber.New()
    app.Get("/users", controller.ListUsers)
    
    req := httptest.NewRequest("GET", "/users", nil)
    resp, err := app.Test(req)
    
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}
```

### 5. Error Handling Migration

**Before:**
```go
func handler(w http.ResponseWriter, r *http.Request) {
    user, err := getUser(id)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "User not found", 404)
            return
        }
        log.Printf("Error: %v", err)
        http.Error(w, "Internal error", 500)
        return
    }
    // ...
}
```

**After:**
```go
func (c *Controller) GetUser(ctx fiber.Ctx) error {
    logger := http.GetLogger(ctx)
    
    user, err := c.service.GetUser(ctx.Params("id"))
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            return ctx.Status(404).JSON(fiber.Map{
                "error": "User not found",
            })
        }
        
        logger.Error("Failed to get user",
            log.NewField("error", err),
            log.NewField("id", ctx.Params("id")),
        )
        
        return ctx.Status(500).JSON(fiber.Map{
            "error": "Internal server error",
        })
    }
    
    return ctx.JSON(user)
}
```

## Migration Checklist

When migrating to Goe:

1. **Dependencies**
   - [ ] Remove old framework dependencies
   - [ ] Add `go.oease.dev/goe/v2`
   - [ ] Update import statements

2. **Application Structure**
   - [ ] Create main application with `goe.New()`
   - [ ] Convert middleware to Fiber handlers
   - [ ] Update route definitions

3. **Dependency Injection**
   - [ ] Convert global variables to services
   - [ ] Create provider functions
   - [ ] Use parameter objects for complex dependencies

4. **Configuration**
   - [ ] Move config to `.env` files
   - [ ] Update config access to use `goe.Config()`
   - [ ] Remove old config libraries

5. **Logging**
   - [ ] Replace logger with `goe.Log()`
   - [ ] Convert to structured logging
   - [ ] Use context logger in handlers

6. **Testing**
   - [ ] Update test setup to use Fx
   - [ ] Use Fiber's test helpers
   - [ ] Mock dependencies properly

7. **Error Handling**
   - [ ] Implement proper error types
   - [ ] Use Fiber's error responses
   - [ ] Add structured error logging

## Conclusion

Migrating to Goe provides:

- **Better Structure**: Clear separation of concerns
- **Type Safety**: Compile-time dependency checking  
- **Developer Experience**: Simple APIs with powerful features
- **Performance**: Built on fast foundations (Fiber, Zap)
- **Maintainability**: Easy to test and extend

The migration effort is typically straightforward, with most changes being mechanical replacements that can be automated with find-and-replace operations.