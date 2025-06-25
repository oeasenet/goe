# Goe Framework Developer Documentation

Welcome to the comprehensive developer documentation for the Goe Framework! 🚀

Goe is a modern, opinionated application framework for Go that provides a delightful developer experience by integrating best practices and powerful libraries from the Go ecosystem. Built on top of Uber's Fx dependency injection framework, Goe offers modular architecture, type safety, and high performance.

## Quick Start

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
)

func main() {
    app := goe.New(goe.Options{
        WithHTTP: true,
        Invokers: []any{RegisterRoutes},
    })
    goe.Run()
}

func RegisterRoutes(httpKernel contract.HTTPKernel, logger contract.Logger) {
    app := httpKernel.App()
    app.Get("/", func(c fiber.Ctx) error {
        return c.SendString("Hello, Goe! 👋")
    })
    logger.Info("Routes registered successfully")
}
```

## Core Features

- **🔌 Dependency Injection**: Built on Uber's Fx for type-safe component management
- **🌐 High-Performance HTTP**: Powered by GoFiber v3 for fast web applications
- **📝 Structured Logging**: Zap-based logging with development and production modes
- **⚙️ Configuration Management**: Environment-aware config with .env file support
- **💾 Caching**: Unified interface supporting multiple cache backends
- **🗄️ Database Integration**: GORM-based ORM with multiple database support
- **🧩 Modular Architecture**: Extensible module system with lifecycle management
- **🛡️ Contract-Driven Design**: Interface-based architecture for loose coupling

## Documentation Structure

### Getting Started
1. [Introduction](01-introduction.md) - Framework overview and core concepts
2. [Getting Started](02-getting-started.md) - Installation and first application
3. [Project Structure](03-project-structure.md) - Recommended project organization
4. [Architecture](04-architecture.md) - Understanding Goe's design principles

### Core Components
5. [Configuration](05-configuration.md) - Environment variables and config management
6. [Logging](06-logging.md) - Structured logging with Zap
7. [HTTP Server](07-http-server.md) - Web applications and API development
8. [Database](08-database.md) - GORM integration and database management
9. [Caching](09-caching.md) - Cache strategies and multiple backends

### Advanced Topics
10. [Modules](10-modules.md) - Creating custom modules and lifecycle management
11. [Dependency Injection](11-dependency-injection.md) - Leveraging Fx for DI
12. [Error Handling](12-error-handling.md) - Error management best practices
13. [Testing](13-testing.md) - Unit and integration testing strategies

### Practical Guides
14. [Examples](14-examples.md) - Real-world application examples
15. [Best Practices](15-best-practices.md) - Production-ready development tips
16. [Deployment](16-deployment.md) - Deploying Goe applications

### Reference
17. [Contributing](17-contributing.md) - Contributing to the Goe framework
18. [Glossary](18-appendix-glossary.md) - Terms and definitions

## Key Concepts

- **Contracts**: Interfaces that define component behavior for loose coupling
- **Modules**: Self-contained components with lifecycle hooks (OnStart/OnStop)
- **Global Accessors**: Convenient functions like `goe.Log()`, `goe.Config()` for quick access
- **Dependency Injection**: Type-safe component wiring using Uber's Fx framework
- **Lifecycle Management**: Automatic startup/shutdown handling for all components

## Support

- **Documentation**: Comprehensive guides and API reference
- **Examples**: Practical code samples and use cases
- **Testing**: Built-in support for unit and integration testing
- **Community**: Active development and community support

Ready to build amazing Go applications? Let's get started! 🚀
