# 4. Architecture Deep Dive 🏛️

Understanding the architecture of the Goe framework is key to leveraging its full potential and building well-structured
applications. This section provides a detailed look into Goe's design principles, core components, and how they
interact.

## Core Philosophy

Goe is built upon several core philosophies:

* **Convention over Configuration**: While highly configurable, Goe aims to provide sensible defaults that work for most
  applications out-of-the-box.
* **Modularity**: The framework is designed as a set of cohesive, yet loosely coupled modules. This allows developers to
  use only what they need and makes the system extensible.
* **Developer Experience**: Prioritizing ease of use, clear APIs, and helpful tools to make development productive and
  enjoyable.
* **Leveraging the Ecosystem**: Goe doesn't reinvent the wheel. It integrates best-in-class libraries from the Go
  ecosystem like Uber's Fx, GoFiber, Zap, and GORM.
* **Testability**: With its contract-driven design and emphasis on dependency injection, Goe applications are designed
  to be easily testable.

## Layered Architecture

Goe employs a layered architecture to ensure a clear separation of concerns. This makes the framework easier to
understand, maintain, and extend.

```mermaid
graph TD
    A["User's Application Code (Handlers, Services, Domain Logic)"] --> B{"Goe Framework Interfaces & Global Accessors
    (contract.*, goe.*)"};
    B --> C["Goe Core Modules (HTTP, Config, Log, DB, Cache, App, etc.)"];
    C --> D["Dependency Injection Layer (Uber's Fx)"];
    C --> E["External Go Libraries (GoFiber, Zap, GORM, Fiber Storage, etc.)"];

    subgraph Goe Framework Internals
        direction LR
        C
        D
        E
    end
```

1. **User's Application Code**: This is where your specific business logic resides. It includes your HTTP handlers,
   business services, domain models, repositories, and any custom modules you build. This layer interacts with the Goe
   framework through its defined contracts (interfaces) or convenient global accessors.

2. **Goe Framework Interfaces & Global Accessors**: These form the public API of the Goe framework.
    * **Contracts (`contract/`)**: A set of Go interfaces (e.g., `contract.Logger`, `contract.HTTPKernel`) that define
      the capabilities of each core module. Your application code should primarily depend on these interfaces, enabling
      loose coupling and testability.
    * **Global Accessors (`goe.*`)**: Static helper functions (e.g., `goe.Log()`, `goe.Config()`) that provide easy
      access to the default instances of core modules. Useful for convenience but dependency injection is preferred for
      core application logic.

3. **Goe Core Modules**: These are the pre-built components that provide essential functionalities. Each module is
   typically responsible for a specific concern:
    * `App`: Manages the overall application lifecycle, Fx container, and context.
    * `Config`: Handles configuration loading and access.
    * `Log`: Provides structured logging services.
    * `HTTP`: Manages the HTTP server, routing, and middleware (using GoFiber).
    * `DB`: Integrates with GORM for database operations.
    * `Cache`: Offers caching services (using Fiber's storage).
      These modules are registered with the Dependency Injection layer and their lifecycles are managed by it.

4. **Dependency Injection Layer (Uber's Fx)**: Goe is built entirely on [Uber's Fx](https://uber-go.github.io/fx/). Fx
   is a powerful dependency injection framework that manages:
    * **Object Graph Construction**: Automatically resolves and provides dependencies to your components.
    * **Lifecycle Management**: Handles the startup and shutdown sequences of your application and its modules in the
      correct order. Fx invokes `OnStart` and `OnStop` hooks for components that define them.
    * **Modularity**: Fx's module system is leveraged by Goe to organize its own core modules and to allow applications
      to register their own.

5. **External Go Libraries**: Goe integrates and builds upon several high-quality, battle-tested Go libraries:
    * [GoFiber](https://gofiber.io/): For the HTTP layer.
    * [Zap](https://github.com/uber-go/zap): For structured logging.
    * [GORM](https://gorm.io/): For database object-relational mapping.
    * [Fiber's Storage](https://docs.gofiber.io/storage): For cache backend support.
    * Viper (indirectly, common for .env parsing patterns, though Goe has its own lightweight .env loader).

## Dependency Injection with Uber's Fx

Fx is fundamental to Goe. Here's how it works:

* **Providers (`fx.Provide`)**: You "provide" constructors for your services or components. These constructors tell Fx
  how to create an instance of a type. The constructor's parameters are its dependencies, which Fx will resolve.
  ```go
  // Example: Providing a UserService
  func NewUserService(logger contract.Logger, db contract.DB) *UserService {
      return &UserService{log: logger, db: db}
  }
  // In your main or module:
  // fx.Provide(NewUserService)
  ```

* **Invokers (`fx.Invoke`)**: You "invoke" functions that need dependencies to perform some action during application
  startup (e.g., registering HTTP routes, starting a background process). Fx will call these functions, automatically
  supplying their dependencies.
  ```go
  // Example: Registering HTTP routes
  func RegisterRoutes(httpKernel contract.HTTPKernel, userService *UserService) {
      // ... use httpKernel and userService to set up routes ...
  }
  // In your main or module:
  // fx.Invoke(RegisterRoutes)
  ```

* **Lifecycle Hooks**: Fx manages `OnStart` and `OnStop` hooks.
    * `OnStart` functions are executed when the application starts. They can perform initialization tasks like
      connecting to a database or starting a listener. If an `OnStart` hook returns an error, the application startup is
      aborted.
    * `OnStop` functions are executed when the application shuts down (e.g., on SIGINT). They perform cleanup tasks like
      closing database connections or releasing resources.
      Goe's modules use these hooks extensively. Your custom modules can too.

* **`goe.New(options goe.Options)`**: When you call `goe.New()`, you're essentially configuring and creating an `fx.App`
  instance. The `goe.Options` struct allows you to enable Goe's core modules (which are Fx modules themselves) and
  provide your own Fx options (providers, invokers, etc.).

### Dependency Flow

```mermaid
flowchart TD
 subgraph subGraph0["App Initialization (goe.New)"]
        I2{"Goe Core Options"}
        I1["User calls goe.New(goe.Options{...})"]
        I3@{ label: "User's fx.Option(s) (Providers, Invokers)" }
        I4["fx.New() called"]
        I5["Fx Builds Dependency Graph"]
  end
 subgraph subGraph1["App Startup (goe.Run -> fx.App.Start)"]
        S2{"Goe Modules Initialize (Config, Log, DB, HTTP, etc.)"}
        S1["Fx Calls OnStart Hooks in Order"]
        S3["User Module OnStart Hooks"]
        S4["Fx Calls Invokers"]
        S5["Application is Running"]
  end
    I1 --> I2
    I2 --> I3
    I3 --> I4
    I4 --> I5
    S1 --> S2
    S2 --> S3
    S3 --> S4
    S4 --> S5
    subGraph0 --> subGraph1
    I3@{ shape: rect}
```

1. **Initialization (`goe.New`)**:
    * The user calls `goe.New()` with options, specifying which core Goe modules to enable and any custom Fx providers
      or invokers.
    * Goe translates these options into `fx.Option`s.
    * `fx.New()` is called, and Fx analyzes all providers to build a dependency graph. It ensures all dependencies can
      be satisfied.

2. **Startup (`goe.Run()` which internally calls `fx.App.Start`)**:
    * Fx executes `OnStart` hooks for all components in dependency order. Core modules like Config and Logger usually
      start first.
    * Then, user-defined module `OnStart` hooks are called.
    * Finally, Fx executes all `fx.Invoke` functions, again in dependency order. This is where route registration,
      background task initiation, etc., typically happen.
    * The application is now considered running.

## Module System

Goe's module system allows for organizing code into manageable, independent units. Each core Goe feature (Config, Log,
HTTP, DB, Cache) is implemented as a module. You can also create your own modules.

A Goe module must implement the `contract.Module` interface:

```go
package contract

import "context"

type Module interface {
	Name() string                      // Returns the unique name of the module
	OnStart(ctx context.Context) error // Called when the module starts
	OnStop(ctx context.Context) error  // Called when the module stops
}
```

Modules are typically registered with Fx using `fx.Module("module-name", fx.Provide(...), fx.Invoke(...))`, and Goe
wraps this by allowing you to provide `contract.Module` implementations which it then integrates into the Fx lifecycle.

### Module Lifecycle

```mermaid
graph TD
    ML1["Application Starts (goe.Run)"] --> ML2{"Fx Manages Lifecycle"};
    ML2 -- "Iterates through modules" --> ML3["Module.OnStart(ctx) Called (In dependency order or registration order if no deps)"];
    ML3 -- "All OnStart successful" --> ML4["Application Running"];
    ML4 -- "Shutdown Signal (e.g., Ctrl+C)" --> ML5{"Fx Initiates Shutdown"};
    ML5 -- "Iterates through modules" --> ML6["Module.OnStop(ctx) Called (In reverse OnStart order)"];
    ML6 -- "All OnStop successful" --> ML7["Application Exits Gracefully"];
    ML3 -- "OnStart returns error" --> ML_ERR["Application Startup Fails"];
    ML6 -- "OnStop returns error" --> ML_WARN["Error Logged, Shutdown Continues"];
```

* **`Name()`**: Provides a unique identifier for the module. Used by Fx.
* **`OnStart(ctx context.Context)`**: Executed during application startup. This is where a module should initialize its
  resources, start background goroutines, etc. The provided `context.Context` is tied to the application's lifecycle; if
  the application starts shutting down, this context will be canceled.
* **`OnStop(ctx context.Context)`**: Executed during application shutdown. This is where a module should release
  resources, stop background goroutines gracefully, etc. The context can be used to set deadlines for cleanup
  operations.

## HTTP Request Lifecycle (with GoFiber)

When Goe's HTTP module is enabled, it uses GoFiber to handle incoming requests.

```mermaid
sequenceDiagram
    participant Client;
    participant GoFiber;
    participant Goe Middleware;
    participant Your Handler;
    participant Goe Services;

    Client->>GoFiber: HTTP Request (e.g., GET /users);
    GoFiber->>Goe Middleware: Request ID, Recovery, Logging, Service Injection;
    Note over GoFiber,Goe Middleware: Standard Fiber middleware chain executes;
    Goe Middleware->>Your Handler: fiber.Ctx (with injected services);
    Your Handler->>Goe Services: Access Config, Logger, DB (via DI or context);
    Goe Services-->>Your Handler: Data / Results;
    Your Handler->>GoFiber: Response (e.g., c.JSON(...), c.SendString(...));
    GoFiber-->>Client: HTTP Response;
```

1. **Request Reception**: GoFiber receives an incoming HTTP request.
2. **Middleware Execution**: The request passes through a chain of middleware. Goe pre-configures several:
    * **Recovery**: Catches panics and converts them into errors.
    * **Request ID**: Assigns a unique ID to each request (useful for tracing).
    * **Service Injection**: Injects Goe services (Config, Logger, App, Validator) into the `fiber.Ctx` locals, making
      them accessible via helper functions like `http.GetLogger(c)`.
    * **Request Logging**: Logs details about the incoming request and its eventual response.
    * Any custom middleware you add.
3. **Routing**: GoFiber matches the request path and method to a registered route.
4. **Handler Execution**: The corresponding handler function is executed.
    * The handler receives a `fiber.Ctx` object, which contains request details and methods for sending a response.
    * Handlers can access Goe services either through DI (if the handler itself or its controller is an Fx component) or
      by retrieving them from the `fiber.Ctx` (e.g., `http.GetLogger(c)`).
5. **Response**: The handler generates a response (e.g., JSON, HTML, redirect) using `fiber.Ctx` methods.
6. **Middleware (Reverse)**: The response may pass back through middleware (some middleware operate on responses).
7. **Response Sent**: GoFiber sends the final HTTP response to the client.

## Contract-Driven Design

Goe's core components adhere to interfaces defined in the `contract/` directory. For example, `contract.Logger` defines
the logging interface, which is implemented by `core/log/zap_logger.go`. This design:

* **Promotes Loose Coupling**: Your application code depends on these stable interfaces, not concrete implementations.
* **Enhances Testability**: You can easily mock these interfaces in your tests.
* **Allows Extensibility**: You could, in theory, provide an alternative implementation for a core contract if needed (
  though this is an advanced use case).

## Global Accessors vs. Dependency Injection

Goe offers two ways to access core services:

1. **Global Accessors** (e.g., `goe.Log()`, `goe.Config()`):
    * **Pros**: Convenient for quick access, especially in `main.go`, scripts, or simple functions where full DI setup
      is overkill.
    * **Cons**: Can lead to less testable code if overused, as they create hidden dependencies. Can make it harder to
      manage different configurations of a service (though Goe primarily uses one global instance per service).

2. **Dependency Injection** (via Fx):
    * **Pros**: Promotes explicit dependencies, making code easier to understand, test, and maintain. Aligns well with
      Fx's lifecycle management. Preferred for services, handlers, and other significant components of your application.
    * **Cons**: Requires a bit more setup (defining constructors, Fx options).

Goe supports both, allowing you to choose the best approach for different parts of your application. The recommendation
is to favor DI for your application's core logic and use global accessors sparingly.

This architectural overview should provide a solid foundation for understanding how Goe works. Subsequent sections will
delve into the specifics of each core module.

Next, let's explore [Configuration](05-configuration.md).

```
