# 10. Extending Goe with Modules 🧩

Goe's module system is a cornerstone of its extensibility and organization. It allows you to group related functionalities, manage their lifecycles, and integrate them seamlessly into Goe's dependency injection system powered by Uber's Fx. All of Goe's core components (HTTP, Log, DB, Cache, Config) are themselves modules.

## What is a Module?

A Goe module is any Go struct that implements the `contract.Module` interface. This interface defines the basic lifecycle hooks that Fx will manage.

```go
package contract

import "context"

// Module defines the interface that all Goe modules must implement
type Module interface {
	// Name returns the unique name of the module.
	// This name is used by Fx for identification.
	Name() string

	// OnStart is called when the module starts, after its dependencies are ready.
	// The provided context is tied to the application's lifecycle.
	OnStart(ctx context.Context) error

	// OnStop is called when the module stops, during graceful shutdown.
	// The provided context can be used for cleanup deadlines.
	OnStop(ctx context.Context) error
}
```

*   **`Name() string`**: Must return a unique string identifier for the module. This is important for Fx, especially for debugging and visualization of the dependency graph.
*   **`OnStart(ctx context.Context) error`**: This method is executed when the application (and specifically, this module) is starting up. It's the place to:
    *   Initialize resources (e.g., database connections if not handled by a core module, message queue consumers).
    *   Start background goroutines.
    *   Perform any setup tasks required for the module to function.
    If `OnStart` returns an error, the application startup will be aborted. The `context.Context` passed in is tied to the application's lifecycle; it will be canceled if the application begins to shut down.
*   **`OnStop(ctx context.Context) error`**: This method is executed when the application is shutting down gracefully. It's the place to:
    *   Release resources (e.g., close connections, stop listeners).
    *   Wait for background goroutines to finish.
    *   Perform any cleanup tasks.
    The `context.Context` can be used to set a deadline for cleanup operations. Errors returned from `OnStop` are typically logged, but the shutdown process will usually continue.

## Creating a Custom Module

Let's create a simple custom module that prints a message on start and stop, and perhaps manages a dummy resource.

**1. Define the Module Struct:**

It's good practice to include any dependencies the module itself needs, like a logger.

```go
// internal/modules/greetingmodule/greeting_module.go
package greetingmodule

import (
	"context"
	"fmt"
	"go.oease.dev/goe/v2/contract"
)

type GreetingModule struct {
	logger contract.Logger
	prefix string
}

// NewGreetingModule is the constructor for our module.
// Fx will inject the contract.Logger.
// We might get 'prefix' from configuration or pass it during construction.
func NewGreetingModule(logger contract.Logger, cfg contract.Config) *GreetingModule {
	modulePrefix := cfg.GetString("GREETING_MODULE_PREFIX")
	if modulePrefix == "" {
		modulePrefix = "DefaultGreeting"
	}
	return &GreetingModule{
		logger: logger,
		prefix: modulePrefix,
	}
}

// Name implements contract.Module.
func (m *GreetingModule) Name() string {
	return "greeting" // Unique name for this module
}

// OnStart implements contract.Module.
func (m *GreetingModule) OnStart(ctx context.Context) error {
	m.logger.Info(fmt.Sprintf("%s: GreetingModule starting!", m.prefix), contract.NewField("module_name", m.Name()))
	// Imagine initializing a resource here
	return nil
}

// OnStop implements contract.Module.
func (m *GreetingModule) OnStop(ctx context.Context) error {
	m.logger.Info(fmt.Sprintf("%s: GreetingModule stopping!", m.prefix), contract.NewField("module_name", m.Name()))
	// Imagine cleaning up a resource here
	return nil
}

// Greet is a custom method for this module
func (m *GreetingModule) Greet(name string) string {
	greeting := fmt.Sprintf("%s: Hello, %s!", m.prefix, name)
	m.logger.Info("Greeting generated", contract.NewField("greeting", greeting))
	return greeting
}
```

**2. Providing Dependencies and Registering the Module**

Modules, and their dependencies, are registered with Goe's Fx application instance. Typically, you'd do this when initializing Goe using `goe.Options`.

Goe handles the basic Fx wiring for `contract.Module` implementations automatically when you provide them via `goe.Options.Modules` or `goe.Options.Providers` (if the provider returns a `contract.Module`).

If your module itself also provides other services to the application, you'll use `fx.Provide`.

```go
// main.go
package main

import (
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	"example.com/yourproject/internal/modules/greetingmodule" // Adjust path
	"go.uber.org/fx"
)

func main() {
	_ = goe.New(goe.Options{
		// Option 1: Registering as a contract.Module directly (Goe wraps it for Fx lifecycle)
		// This is suitable if the module doesn't provide other components to Fx.
		// If NewGreetingModule returns *GreetingModule which implements contract.Module:
		// Modules: []contract.Module{
		//  greetingmodule.NewGreetingModule(goe.Log(), goe.Config()), // Manual instantiation
		// },

		// Option 2: Using Fx Providers (Recommended for modules that might provide other services or need DI for construction)
		// Here, Fx will construct GreetingModule and also recognize it for lifecycle hooks.
		Providers: []any{
			greetingmodule.NewGreetingModule, // Fx will provide contract.Logger and contract.Config
		},

		// If GreetingModule itself should be available for injection elsewhere:
		// fx.Provide(greetingmodule.NewGreetingModule), // This is effectively what happens if NewGreetingModule returns contract.Module

		// We can also make the *GreetingModule instance itself available for injection
		// by ensuring NewGreetingModule is in fx.Provide and it's the concrete type.
		// Then we can invoke a function that uses it:
		Invokers: []any{
			func(gm *greetingmodule.GreetingModule, logger contract.Logger) {
				if gm == nil {
					logger.Warn("GreetingModule is nil, not registered correctly as a concrete type provider?")
					return
				}
				greeting := gm.Greet("Goe User")
				logger.Info("Received greeting from module", contract.NewField("greeting_text", greeting))
			},
		},
	})

	goe.Run()
}
```

**Explanation of Registration:**

*   **`goe.Options.Providers`**: This is the most common and flexible way. You provide the constructor (`greetingmodule.NewGreetingModule`). Fx resolves its dependencies (`contract.Logger`, `contract.Config`). If the result of the constructor implements `contract.Module`, Goe's core application setup (`core/app/app.go`) will automatically wrap it in an `fx.Module` structure, ensuring its `Name()`, `OnStart()`, and `OnStop()` methods are correctly wired into Fx's lifecycle management.
*   **`fx.Invoke`**: The invoker function demonstrates how other components can depend on your `*greetingmodule.GreetingModule` if it's provided to the Fx graph as a concrete type.

## Module Lifecycle and Fx Interaction

When you register a `contract.Module` with Goe (typically by providing its constructor to Fx):

1.  **Instantiation**: Fx creates an instance of your module, injecting any dependencies specified in its constructor (e.g., `contract.Logger`, `contract.Config`, other services).
2.  **Wrapping**: Goe's application bootstrap logic (specifically in `core/app/app.go`'s `wrapModule` function or similar Fx setup for core modules) takes this instance. It creates an `fx.Module` definition. This `fx.Module` uses `fx.Invoke` to append Fx lifecycle hooks that call your module's `OnStart` and `OnStop` methods.
    ```go
    // Simplified from core/app/app.go
    // For a user-provided module `m` of type contract.Module
    /*
    fx.Module(
        m.Name(), // Module's unique name
        fx.Invoke(func(lc fx.Lifecycle) { // Fx Lifecycle object
            lc.Append(fx.Hook{
                OnStart: func(ctx context.Context) error {
                    return m.OnStart(ctx) // Call the module's OnStart
                },
                OnStop: func(ctx context.Context) error {
                    return m.OnStop(ctx) // Call the module's OnStop
                },
            })
        }),
        // ... any fx.Provide options from the module itself would also be here ...
    )
    */
    ```
3.  **`OnStart` Execution**: During `goe.Run()`, Fx calls all registered `OnStart` hooks in dependency order. Your module's `OnStart` will be called when its turn comes.
4.  **`OnStop` Execution**: When the application shuts down, Fx calls `OnStop` hooks in the reverse order of `OnStart` execution.

## Modules Providing Services

Modules are not just for lifecycle management; they are a natural way to group and provide services to the rest of the application.

Imagine our `GreetingModule` also provides a `DailyMessageService`:

```go
// internal/modules/greetingmodule/greeting_module.go
// ... (GreetingModule struct and methods as before) ...

// internal/modules/greetingmodule/daily_message_service.go
package greetingmodule

import "go.oease.dev/goe/v2/contract"

type DailyMessageService struct {
	logger contract.Logger
}

func NewDailyMessageService(logger contract.Logger) *DailyMessageService {
	return &DailyMessageService{logger: logger}
}

func (s *DailyMessageService) GetMessage() string {
	// In a real app, this might come from config, DB, or an external API
	return "Have a productive day with Goe!"
}

// To make this service available via Fx, it needs to be provided.
// This can be done when registering the GreetingModule itself,
// or the GreetingModule can be an Fx module that provides services.

// Option A: Providing services alongside the module in main.go
// fx.Provide(greetingmodule.NewGreetingModule),
// fx.Provide(greetingmodule.NewDailyMessageService),

// Option B: The module itself is an Fx module that provides services
// This is how Goe's core modules are structured.
/*
// In greetingmodule/fx.go (hypothetical file)
package greetingmodule

import "go.uber.org/fx"

// FxModule is an Fx option that groups providers for this module.
var FxModule = fx.Module("greeting_fx", // Different from contract.Module.Name() if needed for Fx grouping
    fx.Provide(
        NewGreetingModule,        // Provides *GreetingModule
        NewDailyMessageService,   // Provides *DailyMessageService
    ),
    // If NewGreetingModule returns contract.Module, Goe's app setup
    // will handle its lifecycle hooks.
)

// Then in main.go:
// app := goe.New(goe.Options{
//     Providers: []any{
//         greetingmodule.FxModule, // Register the whole Fx module
//     },
//     Invokers: []any{
//         func(dms *greetingmodule.DailyMessageService, logger contract.Logger) {
//             logger.Info(dms.GetMessage())
//         },
//     },
// })
*/
```

Goe's own core modules (like `core/log.Module`, `core/http.Module`) are structured like Option B. They are Fx modules that provide the main service (e.g., `contract.Logger`) and also implement `contract.Module` for lifecycle management, which Goe's `App` setup integrates.

For custom application modules, providing the module's constructor (which returns a `contract.Module` implementation) and any services it offers directly in `goe.Options.Providers` is often the most straightforward approach.

## Benefits of Using Modules

*   **Organization**: Groups related code and functionality.
*   **Lifecycle Management**: Provides clear `OnStart` and `OnStop` hooks for managing resources.
*   **Encapsulation**: Can hide internal implementation details while exposing services via interfaces.
*   **Reusability**: Well-designed modules can potentially be reused across different projects (though `internal/` modules are project-specific).
*   **Testability**: Modules and their provided services can be tested independently, especially when dependencies are injected.

By understanding and utilizing Goe's module system, you can build complex applications in a more structured, maintainable, and extensible way.

Next, we'll dive deeper into [Dependency Injection with Fx](11-dependency-injection.md).
