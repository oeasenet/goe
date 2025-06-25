# 1. Introduction to Goe Framework 🚀

Welcome to the Goe Framework! Goe is a modern, opinionated application framework for the Go programming language, designed to accelerate the development of robust, scalable, and maintainable applications.

## 🤔 What is Goe?

Goe (Go Ease) aims to provide a delightful developer experience by integrating best practices and powerful libraries from the Go ecosystem. It's built with the philosophy that a framework should handle the boilerplate and provide sensible defaults, allowing developers to focus on business logic.

**Key Design Principles:**

*   **Developer Experience**: Intuitive API with both global accessors and dependency injection patterns
*   **Modularity & Extensibility**: Optional modules (HTTP, Cache, Database) that can be enabled as needed
*   **Performance**: Built on high-performance libraries like Fiber (HTTP), Zap (logging), and GORM (database)
*   **Type Safety**: Contract-driven design with interfaces and Uber's Fx for type-safe dependency injection
*   **Concurrency Safety**: Thread-safe core components designed for concurrent environments
*   **Best Practices**: Incorporates proven patterns for configuration, logging, HTTP handling, and lifecycle management

## ✨ Core Features

Goe provides a comprehensive set of features designed for modern Go applications:

### 🔌 Dependency Injection & Lifecycle Management
Built on [Uber's Fx](https://uber-go.github.io/fx/) framework, providing:
- Type-safe dependency injection with automatic component wiring
- Structured application lifecycle with startup/shutdown hooks
- Easy testing through mockable interfaces and dependency substitution
- Graceful shutdown handling for all components

### 🌐 High-Performance HTTP Server (Optional)
Powered by [GoFiber v3](https://gofiber.io/) when enabled with `WithHTTP: true`:
- Lightning-fast HTTP routing built on Fasthttp
- Built-in middleware: request logging, recovery, request ID generation
- JSON serialization using Sonic for optimal performance
- Request validation and error handling
- Support for middleware, route groups, and parameter binding

### 📝 Structured Logging
Advanced logging built on [Zap](https://github.com/uber-go/zap):
- High-performance structured logging with zero allocations
- Development-friendly console output and production JSON format
- Configurable log levels, outputs, and custom encoders
- Context-aware logging with request tracing
- Integration with Fx for component lifecycle logging

### ⚙️ Configuration Management
Flexible configuration system:
- Environment variable loading with type-safe getters
- Support for `.env`, `.env.local`, `.env.development` files
- Hot-reloading capabilities for configuration changes
- Custom configuration sources through the `ConfigSource` interface
- Built-in validation and default value handling

### 💾 Caching System (Optional)
Unified caching interface when enabled with `WithCache: true`:
- Multiple cache backends: Memory, Redis, and custom drivers
- Rich API: `Get`, `Set`, `Remember`, `Forget`, `Increment`, `Decrement`
- TTL support and cache key prefixing
- JSON serialization for complex data types
- Cache manager for multiple named stores

### 🗄️ Database Integration (Optional)
GORM-based database support when enabled with `WithDB: true`:
- Multiple database drivers: PostgreSQL, MySQL, SQLite, SQL Server
- Connection pooling and lifecycle management
- Auto-migration support for development
- Multiple database connections with named instances
- Model registration for automatic migrations

### 🧩 Modular Architecture
Extensible module system:
- Optional core modules that can be enabled/disabled
- Custom module creation with lifecycle hooks (`OnStart`, `OnStop`)
- Clean separation of concerns
- Module-specific configuration and dependencies

### 🛡️ Contract-Driven Design
Interface-based architecture:
- All core components defined by contracts (interfaces)
- Loose coupling enabling easy testing and customization
- Consistent API patterns across all modules
- Support for custom implementations of any component

### 🎯 Dual Access Patterns
Flexible component access:
- **Global Accessors**: `goe.Log()`, `goe.Config()`, `goe.HTTP()` for convenience
- **Dependency Injection**: Explicit injection for better testability and structure
- Choose the pattern that fits your use case and team preferences

## 🏗️ Architecture Overview

Goe follows a modular, contract-driven architecture built on dependency injection:

```mermaid
graph TB
    subgraph "Application Layer"
        A[Your Business Logic]
        B[HTTP Handlers]
        C[Services & Repositories]
    end

    subgraph "Goe Framework"
        D[Global Accessors<br/>goe.Log(), goe.Config(), etc.]
        E[Contract Interfaces<br/>HTTPKernel, Logger, Cache, DB]
        F[Core Modules<br/>HTTP, Config, Log, Cache, DB]
    end

    subgraph "Foundation"
        G[Uber Fx<br/>Dependency Injection]
        H[Third-party Libraries<br/>Fiber, Zap, GORM]
    end

    A --> D
    A --> E
    B --> E
    C --> E
    D --> F
    E --> F
    F --> G
    F --> H

    style A fill:#e1f5fe
    style D fill:#f3e5f5
    style E fill:#e8f5e8
    style F fill:#fff3e0
    style G fill:#fce4ec
    style H fill:#f1f8e9
```

### Key Architectural Components

**Application Layer**
- Your business logic, HTTP handlers, and services
- Interacts with Goe through contracts or global accessors
- Completely decoupled from implementation details

**Goe Framework Layer**
- **Global Accessors**: Convenient functions for quick access (`goe.Log()`, `goe.HTTP()`)
- **Contracts**: Interface definitions that ensure loose coupling
- **Core Modules**: Implementation of HTTP, logging, caching, database functionality

**Foundation Layer**
- **Uber Fx**: Manages dependency injection and component lifecycle
- **Third-party Libraries**: High-performance Go libraries (Fiber, Zap, GORM)

### Module Lifecycle

Each Goe module follows a consistent lifecycle:

1. **Registration**: Modules are registered with the Fx container
2. **Dependency Resolution**: Fx resolves all dependencies automatically
3. **OnStart**: Module initialization and resource setup
4. **Runtime**: Module serves requests and handles operations
5. **OnStop**: Graceful shutdown and resource cleanup

### Benefits of This Architecture

- **Modularity**: Enable only the components you need
- **Testability**: Easy mocking through interface contracts
- **Extensibility**: Add custom modules with the same lifecycle
- **Performance**: Built on proven, high-performance libraries
- **Maintainability**: Clear separation of concerns and dependencies

## 🚀 Quick Example

Here's how these concepts work together in practice:

```go
// Using global accessors (simple)
goe.Log().Info("Application starting")
config := goe.Config().GetString("DATABASE_URL")

// Using dependency injection (recommended for larger apps)
func NewUserService(db contract.DB, logger contract.Logger) *UserService {
    return &UserService{db: db, logger: logger}
}

// Module registration
app := goe.New(goe.Options{
    WithHTTP:  true,  // Enable HTTP module
    WithDB:    true,  // Enable database module
    WithCache: true,  // Enable cache module
    Providers: []any{NewUserService}, // Register your services
})
```

For a deeper dive into the architecture and design patterns, see the [Architecture Deep Dive](04-architecture.md) section.

Ready to build your first application? Let's continue with [Getting Started](02-getting-started.md)!
