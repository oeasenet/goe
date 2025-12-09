# Goe - Modern Go Application Framework 🐹

<div align="center">
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Status-Dev-yellow?style=for-the-badge" alt="Status">
  <img src="https://img.shields.io/codecov/c/gh/oeasenet/goe/v2?token=9SWCFFQ38U&style=for-the-badge" alt="Coverage">
</div>

Goe is a modern Go application framework that combines best practices from leading frameworks and libraries. Built
entirely on [Uber's Fx](https://uber-go.github.io/fx/) dependency injection framework and
leveraging [GoFiber](https://gofiber.io/) for its HTTP layer, Goe prioritizes developer experience, modularity,
extensibility, and concurrent safety.

It aims to provide a solid foundation for building robust and scalable Go applications with ease. 😊

## ✨ Core Features

- **🔌 Powerful Dependency Injection**: Built on Uber's Fx for type-safe management of application components and
  lifecycle.
- **🌐 High-Performance HTTP Server**: Integrated with GoFiber v3, featuring automatic request logging, middleware
  support, and fast routing.
- **📝 Structured & Flexible Logging**: Utilizes Uber's Zap logger with developer-friendly console output and
  production-ready JSON formatting.
- **⚙️ Environment-Aware Configuration**: Load configuration from environment variables and `.env` files, with type-safe
  accessors.
- **💾 Versatile Cache Support**: Unified caching interface via Fiber's storage, supporting Memory and Redis drivers.
- **🧩 Extensible Module System**: Organize your application into logical modules with managed lifecycles (`OnStart`,
  `OnStop`).
- **🛡️ Contract-Driven Design**: Core components are defined by interfaces, promoting loose coupling and testability.
- **🎯 Intuitive Developer Experience**: Offers both simple global accessors for convenience and full support for
  explicit dependency injection.
- **🔄 Concurrency Safety**: Core framework components are designed to be safe for concurrent use.
- **🗄️ GORM Database Integration**: Seamless integration with GORM for database operations, supporting multiple SQL
  drivers.
- **🔄 Event System**: Built-in event system with Redis backend for publish-subscribe patterns and asynchronous
  processing.

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
	// Initialize Goe with HTTP module enabled
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

Access your application at `http://localhost:8080`.

For a detailed step-by-step guide, please see the full
[Getting Started documentation](https://deepwiki.com/oeasenet/goe) on DeepWiki.

## 📚 Official Documentation

The **official documentation for GOE** is exclusively hosted on DeepWiki:

🌐 **[https://deepwiki.com/oeasenet/goe](https://deepwiki.com/oeasenet/goe)**

## 🤝 Contributing

Contributions are welcome and greatly appreciated! Here's how to get started:

1. **Fork & Clone**: Fork this repository and clone your fork locally.
2. **Create a Branch**: Use a descriptive branch name (e.g., `feature/add-cache-metrics`).
3. **Install Dependencies**: Run `go mod download` to ensure all modules are available.
4. **Format & Lint**: Ensure code is formatted (`gofmt -w .`) and linted as needed.
5. **Test**: Run the test suite before opening a PR:
   ```bash
   make test
   ```
6. **Commit & PR**: Commit with clear messages and open a pull request against `main`, describing the change and any relevant context.

If you encounter issues or have questions, please open a GitHub issue so we can help.

## 📝 License

Goe is released under the [MIT License](LICENSE).
