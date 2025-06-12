# Examples

Below are small snippets showing how different modules can be used. Each example demonstrates both global access and dependency injection patterns.

## Logging

```go
// Global
logger := goe.Log()
logger.Info("hello")

// Injected
func process(log contract.Logger) {
    log.Info("from DI")
}
```

## HTTP Handler

```go
func routes(h contract.HTTPKernel) {
    app := h.App()
    app.Get("/hello", func(c fiber.Ctx) error {
        // from request context
        logger := http.GetLogger(c)
        logger.Info("handling request")
        return c.SendString("hi")
    })
}
```

## Cache

```go
// Using globals
_ = goe.Cache().Forever("key", "value")

// Injected
func store(cache contract.Cache) {
    cache.Set("key", 123, time.Minute)
}
```
