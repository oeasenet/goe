# Fx Usage Guide for Goe Framework

This guide explains how to use Uber's Fx dependency injection framework within the Goe framework.

## Table of Contents

1. [Introduction](#introduction)
2. [Basic Usage](#basic-usage)
3. [Registering Modules](#registering-modules)
4. [Registering Providers](#registering-providers)
5. [Dependency Injection](#dependency-injection)
6. [Lifecycle Management](#lifecycle-management)
7. [Advanced Features](#advanced-features)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)

## Introduction

The Goe framework uses [Uber's Fx](https://github.com/uber-go/fx) for dependency injection and module lifecycle management. Fx provides a clean way to manage dependencies between components and handle application lifecycle events.

## Basic Usage

Here's a simple example of how to use the Goe framework with Fx:

```go
package main

import (
    "time"
    "go.oease.dev/goe/v2"
)

func main() {
    // Create a new Goe application
    app := goe.New()

    // Register a provider function
    app.App().RegisterProvider(func() *MyService {
        return &MyService{Name: "example"}
    })

    // Invoke a function with injected dependencies
    app.App().Invoke(func(service *MyService) {
        service.DoSomething()
    })

    // Run the application with a timeout
    if err := app.RunWithTimeout(10 * time.Second); err != nil {
        app.Log().Fatal("Application failed", "error", err)
    }
}

type MyService struct {
    Name string
}

func (s *MyService) DoSomething() {
    // Do something
}
```

## Registering Modules

Modules in Goe implement the `contract.Module` interface:

```go
type Module interface {
    Name() string
    Initialize(ctx context.Context) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

You can register modules with the application using the `RegisterModule` or `RegisterModules` methods:

```go
// Register a single module
app.App().RegisterModule(myModule)

// Register multiple modules
app.App().RegisterModules(
    module1,
    module2,
    module3,
)
```

## Registering Providers

Providers are functions that create and return instances of services. You can register providers with the application using the `RegisterProvider` or `RegisterProviders` methods:

```go
// Register a single provider
app.App().RegisterProvider(func() *MyService {
    return &MyService{Name: "example"}
})

// Register multiple providers
app.App().RegisterProviders(
    NewUserService,
    NewAuthService,
    NewDatabaseService,
)
```

## Dependency Injection

Fx automatically injects dependencies based on function parameters. For example:

```go
// Provider function with dependencies
app.App().RegisterProvider(func(config contract.Config, log contract.Log) *MyService {
    return &MyService{
        Config: config,
        Log:    log,
        Name:   "example",
    }
})

// Invoke function with dependencies
app.App().Invoke(func(service *MyService, log contract.Log) {
    log.Info("Service name", "name", service.Name)
})
```

## Lifecycle Management

The Goe framework automatically manages the lifecycle of all registered modules. You can also register custom lifecycle hooks:

```go
// Register a function to be called when the application starts
app.App().OnStart(func(log contract.Log) {
    log.Info("Application started")
})

// Register a function to be called when the application stops
app.App().OnStop(func(log contract.Log) {
    log.Info("Application stopped")
})
```

## Advanced Features

### Named Instances

You can use Fx's annotation feature to provide named instances:

```go
app.App().RegisterProvider(
    fx.Annotate(
        func() *Database { return NewDatabase("primary") },
        fx.ResultTags(`name:"primary"`),
    ),
)

app.App().RegisterProvider(
    fx.Annotate(
        func() *Database { return NewDatabase("replica") },
        fx.ResultTags(`name:"replica"`),
    ),
)

// Inject named instances
app.App().Invoke(func(
    primary *Database `name:"primary"`,
    replica *Database `name:"replica"`,
) {
    // Use primary and replica databases
})
```

### Groups

Fx supports grouping similar components together:

```go
// Define a group
var RouteGroup = fx.Group("routes")

// Register routes in the group
app.App().RegisterProvider(
    fx.Annotate(
        func() Route { return Route{Path: "/users", Handler: handleUsers} },
        fx.ResultTags(`group:"routes"`),
    ),
)

// Inject all routes
app.App().Invoke(func(routes []Route `group:"routes"`) {
    for _, route := range routes {
        // Register each route with the HTTP server
    }
})
```

## Best Practices

1. **Use Constructor Functions**: Define clear constructor functions for your services.
2. **Keep Dependencies Explicit**: Always declare dependencies in function parameters.
3. **Use Interfaces**: Depend on interfaces rather than concrete implementations.
4. **Group Related Providers**: Use modules to group related providers.
5. **Handle Errors**: Always check for errors when running the application.
6. **Use RunWithTimeout for Testing**: Use `RunWithTimeout` instead of `Run` for testing to avoid blocking indefinitely.

## Troubleshooting

### Common Issues

1. **Circular Dependencies**: Fx will report an error if there are circular dependencies between components. Resolve this by breaking the cycle using interfaces or restructuring your code.

2. **Missing Dependencies**: If a dependency is not provided, Fx will report an error. Make sure all required dependencies are registered with the application.

3. **Type Mismatches**: Fx uses reflection to match dependencies by type. Make sure the types match exactly.

4. **Blocking in Lifecycle Hooks**: Avoid blocking operations in lifecycle hooks. Use goroutines for long-running operations.

### Debugging Tips

1. **Enable Fx Logging**: Fx provides detailed logs about the dependency graph and lifecycle events. These logs are automatically integrated with the Goe logger.

2. **Use RunWithTimeout**: Use `RunWithTimeout` instead of `Run` for testing to avoid blocking indefinitely.

3. **Check for Errors**: Always check for errors when running the application.