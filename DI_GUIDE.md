# Uber Fx Integration Guide for Goe Framework

This guide explains how the Goe framework integrates with Uber's Fx dependency injection framework and how you can leverage its features in your applications.

## Table of Contents

1. [Introduction to Fx](#introduction-to-fx)
2. [How Goe Uses Fx](#how-goe-uses-fx)
3. [Registering Providers](#registering-providers)
4. [Dependency Injection](#dependency-injection)
5. [Lifecycle Management](#lifecycle-management)
6. [Advanced Fx Features](#advanced-fx-features)
7. [Testing with Fx](#testing-with-fx)

## Introduction to Fx

[Uber's Fx](https://github.com/uber-go/fx) is a dependency injection framework for Go that:

- Provides a clean way to manage dependencies
- Handles application lifecycle (startup and shutdown)
- Supports concurrent initialization of components
- Offers a modular approach to application construction

Fx uses reflection to wire dependencies together based on function signatures, eliminating the need for manual wiring.

## How Goe Uses Fx

The Goe framework uses Fx internally to:

1. **Manage Module Lifecycle**: Fx's lifecycle hooks are used to initialize, start, and stop modules in the correct order.
2. **Provide Dependency Injection**: Modules and their services are made available for injection.
3. **Handle Graceful Shutdown**: Fx's shutdown mechanism ensures all components are properly stopped.

The `App` module in Goe is the central component that manages the Fx application.

## Registering Providers

Providers are functions that create and return instances of services. In Goe, you can register providers using the `RegisterProvider` method:

```go
// Register a simple provider
app.RegisterProvider(func() *MyService {
    return &MyService{Name: "example"}
})

// Register a provider that has dependencies
app.RegisterProvider(func(config contract.Config, log contract.Log) *MyService {
    return &MyService{
        Config: config,
        Log:    log,
        Name:   "example",
    }
})
```

You can also register multiple providers at once:

```go
app.RegisterProviders(
    NewUserService,
    NewAuthService,
    NewDatabaseService,
)
```

## Dependency Injection

Goe automatically makes all registered modules available for injection. For example, if you register a provider that requires the `Log` module, Fx will automatically inject it:

```go
app.RegisterProvider(func(log contract.Log) *MyService {
    return &MyService{Log: log}
})
```

You can also use the `Invoke` method to execute a function with injected dependencies:

```go
app.Invoke(func(log contract.Log, config contract.Config) {
    log.Info("Application started with config", "env", config.Get("APP_ENV"))
})
```

## Lifecycle Management

Goe modules implement the `Module` interface, which defines lifecycle methods:

```go
type Module interface {
    Name() string
    Initialize(ctx context.Context) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

When you register a module with the app, Goe automatically registers Fx lifecycle hooks for these methods:

```go
app.RegisterModule(myModule)
```

You can also register custom lifecycle hooks:

```go
// Register a function to be called when the application starts
app.OnStart(func(log contract.Log) {
    log.Info("Application started")
})

// Register a function to be called when the application stops
app.OnStop(func(log contract.Log) {
    log.Info("Application stopped")
})
```

## Advanced Fx Features

### Named Instances

Goe supports named instances using Fx's annotation feature:

```go
app.RegisterProvider(
    fx.Annotate(
        func() *Database { return NewDatabase("primary") },
        fx.ResultTags(`name:"primary"`),
    ),
)

app.RegisterProvider(
    fx.Annotate(
        func() *Database { return NewDatabase("replica") },
        fx.ResultTags(`name:"replica"`),
    ),
)

// Inject named instances
app.Invoke(func(
    primary *Database `name:"primary"`,
    replica *Database `name:"replica"`,
) {
    // Use primary and replica databases
})
```

### Groups

Fx supports grouping similar components together:

```go
type Route struct {
    Path    string
    Handler interface{}
}

// Define a group
var RouteGroup = fx.Group("routes")

// Register routes in the group
app.RegisterProvider(
    fx.Annotate(
        func() Route { return Route{Path: "/users", Handler: handleUsers} },
        fx.ResultTags(`group:"routes"`),
    ),
)

app.RegisterProvider(
    fx.Annotate(
        func() Route { return Route{Path: "/auth", Handler: handleAuth} },
        fx.ResultTags(`group:"routes"`),
    ),
)

// Inject all routes
app.Invoke(func(routes []Route `group:"routes"`) {
    for _, route := range routes {
        // Register each route with the HTTP server
    }
})
```

## Testing with Fx

Goe provides utilities for testing with Fx. Here's an example of how to test a component that uses Fx:

```go
func TestMyService(t *testing.T) {
    // Create a test Fx app
    testApp := fxtest.New(t,
        // Provide dependencies
        fx.Provide(
            NewMockLogger,
            NewMockConfig,
            NewMyService,
        ),
        
        // Test the service
        fx.Invoke(func(service *MyService) {
            // Assert that the service works correctly
            result := service.DoSomething()
            if result != expected {
                t.Errorf("Expected %v, got %v", expected, result)
            }
        }),
    )
    
    // Start and stop the test app
    testApp.RequireStart()
    testApp.RequireStop()
}
```

For more examples of testing with Fx, see the `app_test.go` file in the `core/app` package.

## Best Practices

1. **Use Constructor Functions**: Define clear constructor functions for your services.
2. **Keep Dependencies Explicit**: Always declare dependencies in function parameters.
3. **Use Interfaces**: Depend on interfaces rather than concrete implementations.
4. **Group Related Providers**: Use modules to group related providers.
5. **Test with fxtest**: Use `fxtest` to test components that use Fx.

## Further Reading

- [Uber Fx GitHub Repository](https://github.com/uber-go/fx)
- [Fx GoDoc](https://pkg.go.dev/go.uber.org/fx)
- [Dependency Injection in Go using Uber's Fx](https://blog.logrocket.com/dependency-injection-in-go-using-uber-fx/)