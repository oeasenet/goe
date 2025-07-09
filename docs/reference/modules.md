# Core Modules

GOE Framework provides a set of core modules that handle essential application concerns. Each module is designed to be independent, configurable, and easily testable.

## Available Modules

### Observability Module
- **Logging**: Structured logging with Zap
- **Metrics**: Application metrics collection
- **Tracing**: Distributed tracing support
- **Health Checks**: Application health monitoring

### Database Module
- **Connection Management**: Database connection pooling
- **Migration Support**: Database schema migrations
- **Transaction Management**: Automatic transaction handling
- **Query Builders**: Fluent query building

### HTTP Module
- **Fiber Integration**: High-performance HTTP server
- **Middleware Support**: Extensible middleware system
- **Request/Response Handling**: Structured request/response patterns
- **Error Handling**: Centralized error management

### Configuration Module
- **Environment Variables**: Environment-based configuration
- **File Support**: YAML, JSON, TOML configuration files
- **Validation**: Configuration validation
- **Hot Reload**: Runtime configuration updates

## Module Registration

Modules are registered using the dependency injection container:

```go
func main() {
    container := dig.New()
    
    // Register core modules
    observability.Register(container)
    database.Register(container)
    http.Register(container)
    
    // Start application
    err := container.Invoke(func(app *fiber.App) {
        app.Listen(":8080")
    })
    
    if err != nil {
        log.Fatal(err)
    }
}
```

## Custom Modules

Create custom modules by implementing the Module interface:

```go
type Module interface {
    Register(container *dig.Container) error
}

type UserModule struct{}

func (m *UserModule) Register(container *dig.Container) error {
    container.Provide(NewUserRepository)
    container.Provide(NewUserService)
    container.Provide(NewUserHandler)
    return nil
}
```