# Middleware

GOE Framework provides a comprehensive middleware system built on top of Fiber. Middleware components handle cross-cutting concerns like logging, authentication, rate limiting, and more.

## Built-in Middleware

### Request Logging
```go
app.Use(middleware.Logger(middleware.LoggerConfig{
    Format: "${time} ${method} ${path} ${status} ${latency}\n",
}))
```

### CORS
```go
app.Use(middleware.CORS(middleware.CORSConfig{
    AllowOrigins: "*",
    AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
    AllowHeaders: "Origin, Content-Type, Accept, Authorization",
}))
```

### Rate Limiting
```go
app.Use(middleware.RateLimit(middleware.RateLimitConfig{
    Max:      100,
    Duration: time.Minute,
}))
```

### Authentication
```go
app.Use(middleware.JWT(middleware.JWTConfig{
    Secret: "your-secret-key",
    Skip: func(c *fiber.Ctx) bool {
        return c.Path() == "/login"
    },
}))
```

### Request ID
```go
app.Use(middleware.RequestID())
```

### Compression
```go
app.Use(middleware.Compress(middleware.CompressConfig{
    Level: compress.LevelBestSpeed,
}))
```

## Custom Middleware

### Simple Middleware
```go
func CustomMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Before request
        start := time.Now()
        
        // Process request
        err := c.Next()
        
        // After request
        duration := time.Since(start)
        log.Printf("Request took %v", duration)
        
        return err
    }
}
```

### Middleware with Configuration
```go
type TimingConfig struct {
    Enabled   bool
    Threshold time.Duration
}

func TimingMiddleware(config TimingConfig) fiber.Handler {
    if !config.Enabled {
        return func(c *fiber.Ctx) error {
            return c.Next()
        }
    }
    
    return func(c *fiber.Ctx) error {
        start := time.Now()
        err := c.Next()
        duration := time.Since(start)
        
        if duration > config.Threshold {
            log.Printf("Slow request: %s %s took %v", 
                c.Method(), c.Path(), duration)
        }
        
        return err
    }
}
```

## Middleware Order

The order of middleware matters. Here's the recommended order:

```go
// 1. Request ID (for tracing)
app.Use(middleware.RequestID())

// 2. Logging (to capture all requests)
app.Use(middleware.Logger())

// 3. Recovery (to handle panics)
app.Use(middleware.Recover())

// 4. CORS (for cross-origin requests)
app.Use(middleware.CORS())

// 5. Rate limiting (before authentication)
app.Use(middleware.RateLimit())

// 6. Authentication (before authorization)
app.Use(middleware.JWT())

// 7. Compression (before response)
app.Use(middleware.Compress())

// 8. Custom business logic middleware
app.Use(CustomMiddleware())
```

## Conditional Middleware

Skip middleware for certain routes:

```go
app.Use(middleware.JWT(middleware.JWTConfig{
    Secret: "your-secret-key",
    Skip: func(c *fiber.Ctx) bool {
        // Skip auth for health checks
        return c.Path() == "/health"
    },
}))
```

## Group Middleware

Apply middleware to specific route groups:

```go
api := app.Group("/api")
api.Use(middleware.JWT())
api.Use(middleware.RateLimit())

// Auth required for all /api routes
api.Get("/users", getUsersHandler)
api.Post("/users", createUserHandler)

// Public routes
app.Get("/health", healthHandler)
app.Post("/login", loginHandler)
```

## Error Handling Middleware

```go
func ErrorHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        err := c.Next()
        
        if err != nil {
            code := fiber.StatusInternalServerError
            
            if e, ok := err.(*fiber.Error); ok {
                code = e.Code
            }
            
            return c.Status(code).JSON(fiber.Map{
                "error": err.Error(),
                "code":  code,
            })
        }
        
        return nil
    }
}
```

## Middleware Best Practices

1. **Keep middleware focused**: Each middleware should have a single responsibility
2. **Handle errors gracefully**: Always handle errors and return appropriate responses
3. **Use configuration**: Make middleware configurable for different environments
4. **Order matters**: Place middleware in the correct order for optimal performance
5. **Skip when necessary**: Use skip functions for routes that don't need middleware
6. **Performance**: Consider the performance impact of middleware on every request
7. **Testing**: Write tests for custom middleware to ensure they work correctly