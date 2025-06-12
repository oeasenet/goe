# 11. Mastering Dependency Injection with Fx 💉

Dependency Injection (DI) is a fundamental design pattern that promotes loose coupling and testability in software applications. Goe leverages [Uber's Fx](https://uber-go.github.io/fx/), a powerful and minimalistic DI framework, to manage its components and empower developers to build modular applications.

## Why Dependency Injection?

Before diving into Fx, let's briefly understand why DI is important:

*   **Decoupling**: Components declare their dependencies instead of creating them. This makes it easier to swap implementations or manage component lifecycles independently.
*   **Testability**: Dependencies can be easily mocked or replaced with test doubles during unit testing.
*   **Clarity**: Dependencies are explicit, making it easier to understand how components interact.
*   **Lifecycle Management**: A DI container can manage the creation, startup, and shutdown of components in the correct order.

## Fx Fundamentals in Goe

Goe is built entirely around Fx. When you use `goe.New(options ...)` and `goe.Run()`, you are interacting with an Fx application instance.

### Key Fx Concepts:

1.  **Providers (`fx.Provide`)**:
    *   A provider is a constructor function that tells Fx how to create an instance of a type (a "component" or "service").
    *   The constructor's parameters are its dependencies, which Fx automatically resolves and injects.
    *   Example:
        ```go
        // Logger and Config are dependencies Fx will provide
        func NewMyService(logger contract.Logger, config contract.Config) *MyService {
            return &MyService{log: logger, cfg: config}
        }

        // To register with Fx (typically in goe.Options.Providers):
        // fx.Provide(NewMyService)
        ```
    *   Fx builds a directed acyclic graph (DAG) of these providers to understand the dependency relationships.

2.  **Invokers (`fx.Invoke`)**:
    *   An invoker is a function that Fx executes during application startup, after all necessary dependencies have been provided.
    *   Invokers are used to perform actions that require dependencies, such as registering HTTP routes, starting background tasks, or running initial setup logic.
    *   Example:
        ```go
        func RegisterHttpRoutes(kernel contract.HTTPKernel, myService *MyService) {
            app := kernel.App()
            app.Get("/my-route", myService.HandleMyRoute) // Assuming HandleMyRoute exists
        }

        // To register with Fx (typically in goe.Options.Invokers):
        // fx.Invoke(RegisterHttpRoutes)
        ```
    *   If an invoker returns an error, the application startup is aborted.

3.  **Lifecycle Hooks (`fx.Lifecycle`)**:
    *   Fx manages the application lifecycle through `OnStart` and `OnStop` hooks.
    *   Components can request `fx.Lifecycle` as a dependency in their constructor and append hooks to it.
    *   **`OnStart(func(context.Context) error) error`**: Functions to run when the application starts. Executed in the order they are appended.
    *   **`OnStop(func(context.Context) error) error`**: Functions to run when the application shuts down. Executed in LIFO (Last-In, First-Out) order relative to `OnStart` hooks.
    *   Goe's `contract.Module` interface's `OnStart` and `OnStop` methods are integrated into this Fx lifecycle.
    *   Example (simplified):
        ```go
        func NewDatabaseConnection(lc fx.Lifecycle, config contract.Config) (*sql.DB, error) {
            dsn := config.GetString("DB_DSN")
            db, err := sql.Open("mysql", dsn)
            if err != nil {
                return nil, err
            }

            lc.Append(fx.Hook{
                OnStart: func(ctx context.Context) error {
                    // Ping DB on start to ensure connection is live
                    return db.PingContext(ctx)
                },
                OnStop: func(ctx context.Context) error {
                    // Close DB connection on stop
                    return db.Close()
                },
            })
            return db, nil
        }
        // fx.Provide(NewDatabaseConnection)
        ```

4.  **Fx Modules (`fx.Module`)**:
    *   Fx allows grouping related providers and invokers into an `fx.Module`. This helps organize larger applications.
    *   Goe's core components (like HTTP, Log, DB) are structured as Fx modules internally.
    *   You can also create your own Fx modules for parts of your application.
    *   Example:
        ```go
        var MyFeatureModule = fx.Module("my_feature",
            fx.Provide(NewMyFeatureService),
            fx.Provide(NewMyFeatureRepository),
            fx.Invoke(RegisterMyFeatureRoutes),
        )
        // Then, in goe.Options.Providers:
        // MyFeatureModule, // This single line registers all providers/invokers from the module
        ```

5.  **`fx.In` and `fx.Out` Structs (Parameter/Result Objects)**:
    *   For components with many dependencies or constructors that return multiple values, Fx supports parameter objects (`fx.In`) and result objects (`fx.Out`).
    *   **`fx.In`**: A struct where fields are dependencies. Tag fields with `fx:"optional"` or `fx:"name:..."` if needed.
        ```go
        type MyServiceParams struct {
            fx.In // Embed fx.In

            Logger contract.Logger
            Config contract.Config
            DB     contract.DB `name:"primary_db"` // Named dependency
        }

        func NewMyServiceWithParams(params MyServiceParams) *MyService {
            // Use params.Logger, params.Config, params.DB
            return &MyService{log: params.Logger, /* ... */}
        }
        // fx.Provide(NewMyServiceWithParams)
        ```
    *   **`fx.Out`**: A struct where fields are values provided by a constructor.
        ```go
        type MyModuleResults struct {
            fx.Out // Embed fx.Out

            Service   *MyModuleService
            Something contract.Something `group:"mygroup"` // Provide into a group
        }

        func NewMyModuleComponents(logger contract.Logger) (MyModuleResults, error) {
            svc := &MyModuleService{log: logger}
            return MyModuleResults{Service: svc, Something: svc}, nil
        }
        // fx.Provide(NewMyModuleComponents)
        ```

## How Goe Uses Fx

*   **`goe.New(options goe.Options)`**: This is the entry point for creating your Fx application.
    *   `options.Providers`: You pass your `fx.Provide` and `fx.Module` options here.
    *   `options.Invokers`: You pass your `fx.Invoke` options here.
    *   `options.Modules`: You can pass implementations of `contract.Module`. Goe wraps these into Fx lifecycle hooks.
    *   `options.WithHTTP`, `WithDB`, etc.: These flags enable Goe's core Fx modules. For example, `WithLog: true` adds `log.FxModule` (which provides `contract.Logger` and manages its lifecycle) to the Fx application.

*   **Core Modules as Fx Modules**: Each of Goe's core components (Config, Logger, HTTPKernel, DB, Cache) is provided by an internal Fx module. These modules:
    1.  Provide the main interface (e.g., `contract.Logger`).
    2.  Implement `contract.Module` for lifecycle management, which Goe integrates with `fx.Lifecycle`.
    3.  May provide other related components (e.g., the HTTP module also provides the `*CustomValidator`).

*   **`goe.Run()`**: This calls `fx.App.Run()`, which starts the application, executes `OnStart` hooks and invokers, and blocks until a shutdown signal is received. It then handles graceful shutdown by executing `OnStop` hooks.

## Best Practices for DI in Goe

1.  **Constructor Injection**: Prefer injecting dependencies into the constructor of your structs. This makes dependencies explicit and ensures components are valid upon creation.

2.  **Depend on Interfaces (Contracts)**: Where possible, depend on interfaces (like Goe's `contract.Logger`, `contract.DB`) rather than concrete types. This improves testability and flexibility. Goe provides these contracts for its core services.

3.  **Use `fx.In` for Many Dependencies**: If a constructor has more than 3-4 dependencies, consider using an `fx.In` parameter struct to improve readability.

4.  **Group Related Components with `fx.Module`**: For larger features or domains within your application, group their Fx providers and invokers into a dedicated `fx.Module`.

5.  **Clear Naming**: Use clear and consistent names for your constructors (e.g., `NewUserService`, `NewPostgresRepository`).

6.  **Single Responsibility Principle**: Design your components (services, repositories) to have a single, well-defined responsibility. This naturally leads to clearer dependency graphs.

7.  **Provide Concrete Types, Depend on Interfaces**:
    *   Your provider function: `func NewPostgresUserRepository(...) *PostgresUserRepository` (returns concrete type).
    *   Your Fx registration: `fx.Provide(NewPostgresUserRepository, fx.As(new(contract.UserRepository)))` (provides `*PostgresUserRepository` and also makes it available as `contract.UserRepository`).
    *   Your service depending on it: `func NewUserService(repo contract.UserRepository) *UserService`.

## Debugging Dependency Issues

Fx provides information when it can't build the dependency graph. Common issues include:

*   **Missing Provider**: A type is required by a component, but no provider exists for it. Fx will list the missing type.
    *   **Fix**: Add an `fx.Provide(NewMissingTypeConstructor)` for the missing type.

*   **Circular Dependencies**: Component A depends on B, and B depends on A (directly or indirectly). Fx will detect and report this.
    *   **Fix**: Refactor your components to break the cycle. This often involves extracting a new interface/service or rethinking responsibilities. Sometimes, using `fx.Decorate` or value groups can help, but architectural changes are usually better.

*   **Multiple Providers for the Same Type**: If multiple providers offer the same type without differentiation (e.g., using `fx.Supply` or `fx.Provide` for the same interface from different places), Fx might complain or pick one unpredictably.
    *   **Fix**: Use named instances (`fx.Provide(..., fx.Annotate(..., fx.ResultTags(`name:"primary_db"`)))` and `fx.In` with `fx.ParamTags(`name:"primary_db"`)`) or group components into modules to clarify intent.

*   **Invoker Errors**: An invoked function returns an error. Fx will halt startup and report the error.
    *   **Fix**: Debug the invoked function. The error message usually indicates the cause.

Fx's error messages are generally helpful in pinpointing these issues. Visualizing the graph (Fx has experimental support for this) can also be useful for complex applications.

By mastering these Fx concepts, you can build highly modular, testable, and maintainable applications with the Goe framework.

Next, we'll discuss [Error Handling Strategies](12-error-handling.md).
```
