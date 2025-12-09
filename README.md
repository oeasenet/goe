<div align="center">
  <img src="docs/public/goe_gopher_logo.png" alt="GOE Framework Logo" width="200">

# GOE Framework

### Modern Go Application Framework

*Built on Uber's Fx & GoFiber for developer productivity and scalability*

  <br/>

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![Status](https://img.shields.io/badge/Status-Dev-yellow?style=for-the-badge)](https://github.com/oeasenet/goe)
[![Coverage](https://img.shields.io/codecov/c/gh/oeasenet/goe/v2?token=9SWCFFQ38U&style=for-the-badge)](https://codecov.io/gh/oeasenet/goe)

  <br/>

[Documentation](https://deepwiki.com/oeasenet/goe) •
[Getting Started](#-getting-started) •
[Features](#-features) •
[Contributing](#-contributing)

</div>

<br/>

## Overview

GOE is a modern Go application framework that combines best practices from leading frameworks and libraries. Built
entirely on [Uber's Fx](https://uber-go.github.io/fx/) dependency injection framework and
leveraging [GoFiber](https://gofiber.io/) for its HTTP layer, GOE prioritizes developer experience, modularity,
extensibility, and concurrent safety.

<br/>

## ✨ Features

### Core Infrastructure
- 🔌 **Dependency Injection** - Built on Uber's Fx for type-safe component management
- 🌐 **HTTP Server** - GoFiber v3 with middleware & fast routing
- 📝 **Logging** - Uber's Zap with console & JSON formatting
- ⚙️ **Configuration** - Environment-aware with `.env` file support

### Data & Storage
- 🗄️ **SQL Database** - GORM integration (MySQL, PostgreSQL, SQLite, SQL Server)
- 🍃 **MongoDB** - Native driver with connection pooling
- 💾 **Caching** - 10+ backends (Redis, Memory, S3, DynamoDB...)

### Distributed Systems
- 🔄 **Event System** - Redis Streams with consumer groups & DLQ
- 🔐 **Distributed Lock** - Redlock algorithm for multi-process coordination

### Developer Experience
- 🧩 **Module System** - Managed lifecycles (OnStart, OnStop)
- 🛡️ **Contract-Driven** - Interface-based design for testability
- ⚡ **Concurrency Safe** - Thread-safe core components

<br/>

## 🚀 Getting Started

### Installation

```bash
go get go.oease.dev/goe/v2
```

### Quick Start

Create a minimal application (`main.go`):

```go
package main

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
)

func main() {
	// Initialize GOE with HTTP module enabled
	_ = goe.New(goe.Options{
		WithHTTP: true,
		Invokers: []any{
			func(httpKernel contract.HTTPKernel, logger contract.Logger) {
				app := httpKernel.App()
				app.Get("/", func(c fiber.Ctx) error {
					logger.Info("Hello endpoint was hit!")
					return c.SendString("Hello, World from GOE!")
				})
				logger.Info("Main route registered.")
			},
		},
	})

	// Start the application (blocks until shutdown)
	goe.Run()
}
```

Access your application at `http://localhost:8080`

<br/>

## 📦 Available Modules

Enable modules via `goe.Options`:

```go
goe.New(goe.Options{
WithHTTP:    true, // HTTP server (Fiber)
WithCache:   true, // Caching system
WithDB:      true, // SQL database (GORM)
WithMongoDB: true, // MongoDB
WithEvent:   true,  // Event system (Redis Streams)
WithLock:    true, // Distributed locking
})
```

<br/>

## ⚙️ Configuration

GOE loads configuration from multiple sources in order of priority:

| Priority | Source           | Description                              |
|:--------:|------------------|------------------------------------------|
|    1     | `.env`           | Base configuration                       |
|    2     | `.local.env`     | Local overrides (gitignored)             |
|    3     | `.{GOE_ENV}.env` | Environment-specific (e.g., `.prod.env`) |
|    4     | **System ENV**   | Highest priority - overrides all         |

> **See:** [`.example.env`](.example.env) for all options or the [Configuration Reference](CONFIGURATION.md) for detailed documentation.

<br/>

## 📚 Documentation

The **official documentation** is hosted on DeepWiki:

<div align="center">

### **[deepwiki.com/oeasenet/goe](https://deepwiki.com/oeasenet/goe)**

</div>

<br/>

## 🤝 Contributing

Contributions are welcome and greatly appreciated!

1. **Fork & Clone** - Fork this repository and clone your fork locally
2. **Create a Branch** - Use a descriptive branch name (e.g., `feature/add-cache-metrics`)
3. **Install Dependencies** - Run `go mod download`
4. **Format & Lint** - Ensure code is formatted (`gofmt -w .`)
5. **Test** - Run the test suite:
   ```bash
   make test
   ```
6. **Commit & PR** - Open a pull request against `v2` with clear description

If you encounter issues or have questions, please [open a GitHub issue](https://github.com/oeasenet/goe/issues).

<br/>

## 📄 License

GOE is released under the [MIT License](LICENSE).

<div align="center">
  <br/>
  <sub>Built with ❤️ by the GOE community</sub>
</div>
