# Goe - Modern Go Application Framework 🚀

<div align="center">
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-Apache_2.0-green?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Status-Beta-yellow?style=for-the-badge" alt="Status">
  <!-- Consider adding other relevant badges like build status, code coverage, etc. -->
</div>

Goe is a modern Go application framework that combines best practices from leading frameworks and libraries. Built entirely on [Uber's Fx](https://uber-go.github.io/fx/) dependency injection framework and leveraging [GoFiber](https://gofiber.io/) for its HTTP layer, Goe prioritizes developer experience, modularity, extensibility, and concurrent safety.

It aims to provide a solid foundation for building robust and scalable Go applications with ease. 😊

## ✨ Core Features

- **🔌 Powerful Dependency Injection**: Built on Uber's Fx for type-safe management of application components and lifecycle.
- **🌐 High-Performance HTTP Server**: Integrated with GoFiber v3, featuring automatic request logging, middleware support, and fast routing.
- **📝 Structured & Flexible Logging**: Utilizes Uber's Zap logger with developer-friendly console output and production-ready JSON formatting.
- **⚙️ Environment-Aware Configuration**: Load configuration from environment variables and `.env` files, with type-safe accessors.
- **💾 Versatile Cache Support**: Unified caching interface via Fiber's storage, supporting drivers like Memory, Redis, SQLite, etc.
- **🧩 Extensible Module System**: Organize your application into logical modules with managed lifecycles (`OnStart`, `OnStop`).
- **🛡️ Contract-Driven Design**: Core components are defined by interfaces, promoting loose coupling and testability.
- **🎯 Intuitive Developer Experience**: Offers both simple global accessors for convenience and full support for explicit dependency injection.
- **🔄 Concurrency Safety**: Core framework components are designed to be safe for concurrent use.
- **🗄️ GORM Database Integration**: Seamless integration with GORM for database operations, supporting multiple SQL drivers.

## 🚀 Getting Started

Install the dependency:

```bash
go get go.oease.dev/goe/v2
```

Create a minimal application (`main.go`):

```go
package main

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
)

func main() {
	// Initialize Goe with the HTTP module enabled
	_ = goe.New(goe.Options{
		WithHTTP: true,
		Invokers: []any{ // Use Fx invoker to register routes
			func(httpKernel contract.HTTPKernel, logger contract.Logger) {
				app := httpKernel.App()
				app.Get("/", func(c fiber.Ctx) error {
					logger.Info("Hello endpoint was hit!")
					return c.SendString("Hello, World from Goe! 👋")
				})
				logger.Info("Main route registered.")
			},
		},
	})

	// Start the application (blocks until shutdown)
	goe.Run()
}
```

For a detailed step-by-step guide, please see our full [Getting Started documentation](docs/02-getting-started.md).

## 📖 Comprehensive Documentation

Our **new and comprehensive developer documentation** is available in the [`docs/`](docs/README.md) directory.

It covers everything from getting started, architecture, configuration, core modules (HTTP, DB, Cache, Logging), advanced topics like custom modules and dependency injection, to best practices and deployment.

➡️ **[Start exploring the documentation here!](docs/README.md)** ⬅️

## 🤝 Contributing

Contributions are welcome and greatly appreciated! Please see the [Contributing Guide](docs/17-contributing.md) for details on how to get started, including setting up your environment, running tests (`go test ./...`), and submitting pull requests.

We adhere to a [Code of Conduct](../CODE_OF_CONDUCT.md) to ensure a welcoming community.

## 📝 License

Goe is released under the [Apache 2.0 License](LICENSE).

```
