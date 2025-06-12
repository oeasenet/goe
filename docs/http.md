# HTTP Module

Goe integrates [Fiber](https://github.com/gofiber/fiber) for fast HTTP handling. The module is enabled by setting `WithHTTP: true` when creating the application.

## Basic Usage

```go
_ = goe.New(goe.Options{
    WithHTTP: true,
    Invokers: []any{
        func(h contract.HTTPKernel) {
            h.App().Get("/ping", func(c fiber.Ctx) error { return c.SendString("pong") })
        },
    },
})
```

Run `goe.Run()` to start the server. The port and other options are read from configuration variables like `HTTP_PORT`.

## Access Patterns

- **Global method**: `goe.HTTP()` returns the `HTTPKernel` for manual registration.
- **Dependency injection**: accept `contract.HTTPKernel` in a constructor or function.
- **Request context**: inside handlers, `http.GetLogger(c)` and `http.GetConfig(c)` provide request-scoped values.

## Middleware

Fiber middleware can be registered globally or per route/group:

```go
func register(h contract.HTTPKernel) {
    app := h.App()
    app.Use(cors.New())
    api := app.Group("/api", myAuth)
    api.Get("/users", listUsers)
}
```

See the [examples](examples.md) for a complete application.
