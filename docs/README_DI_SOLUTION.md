# Solution: Avoiding Import Cycles with Dependency Injection

## Problem

The issue was about how to inject (call) the config and log modules without creating import cycles in Go. The question was whether Fx should be helping to avoid these cycles.

## Solution

Yes, Uber's Fx framework is designed to help avoid import cycles through dependency injection. We've implemented a solution that leverages Fx's capabilities to allow modules to depend on each other without creating import cycles.

### Key Components

1. **Modified Framework Initialization**: Changed the framework to use Fx for module initialization and dependency injection.
2. **Created a DI Module**: Added a new `di` package with a `Module` struct that contains all core modules as fields.
3. **Provided Example Code**: Created an example that demonstrates how to use the DI module to access other modules without creating import cycles.
4. **Added Documentation**: Created a comprehensive guide on dependency injection in the Goe framework.

### How It Works

1. **Interfaces in a Central Package**: All module interfaces are defined in the `contract` package, which can be imported by any module without creating cycles.
2. **Module Implementations in Separate Packages**: Each module is implemented in its own package, which only imports the `contract` package.
3. **Dependency Injection with Fx**: Fx is used to wire everything together at runtime, allowing modules to depend on each other without creating import cycles.
4. **DI Module for Cross-Module Dependencies**: The `di` package provides a `Module` struct that can be injected into any component that needs access to multiple modules.

## How to Use

### 1. Import the DI Module

```go
import "go.oease.dev/goe/v2/di"
```

### 2. Define a Struct with DI Module as a Dependency

```go
type MyService struct {
    deps *di.Module // Inject all dependencies through the DI module
}
```

### 3. Create a Constructor Function

```go
func NewMyService(deps *di.Module) *MyService {
    return &MyService{
        deps: deps,
    }
}
```

### 4. Use the Dependencies

```go
func (s *MyService) DoSomething() {
    // Access the config module
    appEnv := s.deps.Config.GetDefault("APP_ENV", "development")
    
    // Access the log module
    s.deps.Log.Info("MyService is doing something", "env", appEnv)
    
    // Access other modules as needed
    s.deps.Http.Get("/my-service", func(c fiber.Ctx) error {
        return c.JSON(map[string]interface{}{
            "message": "MyService is running",
            "env":     appEnv,
        })
    })
}
```

### 5. Register Your Service with the Application

```go
app := goe.New()
app.App().RegisterProvider(NewMyService)
app.App().Invoke(func(service *MyService) {
    service.DoSomething()
})
```

## Benefits

1. **No Import Cycles**: Modules can depend on each other without creating import cycles.
2. **Clean Dependencies**: Dependencies are explicit and injected at runtime.
3. **Testability**: Services can be easily tested with mock dependencies.
4. **Modularity**: Modules can be developed and tested independently.
5. **Flexibility**: New modules can be added without modifying existing code.

## Further Reading

For more detailed information and examples, see the [INTERNAL_DI_GUIDE.md](INTERNAL_DI_GUIDE.md) file.