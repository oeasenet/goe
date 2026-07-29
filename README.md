<div align="center">
  <img src="docs/public/goe_gopher_logo.png" alt="GOE Framework Logo" width="180">

# GOE Framework

### Build production Go applications in minutes, not days.

*Dependency injection, HTTP, databases, caching, jobs, and distributed locking — wired together and ready to go.*

<br/>

[![Go 1.26+](https://img.shields.io/badge/Go-1.26+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![Latest Release](https://img.shields.io/github/v/release/oeasenet/goe?style=for-the-badge&color=blue)](https://github.com/oeasenet/goe/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/oeasenet/goe/ci.yml?branch=v2&style=for-the-badge&label=tests)](https://github.com/oeasenet/goe/actions)
[![Coverage](https://img.shields.io/codecov/c/gh/oeasenet/goe/v2?token=9SWCFFQ38U&style=for-the-badge)](https://codecov.io/gh/oeasenet/goe)
[![License: MIT](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

<br/>

[Documentation](https://deepwiki.com/oeasenet/goe) · [Examples](examples/) · [Configuration](.example.env) · [Contributing](#contributing)

</div>

<br/>

## Why GOE?

Most Go projects start the same way: wiring up a logger, a config loader, an HTTP server, a database connection, graceful shutdown — before writing a single line of business logic.

GOE handles all of that. It combines [Uber Fx](https://uber-go.github.io/fx/) for dependency injection with [GoFiber v3](https://gofiber.io/) for HTTP, then layers on the infrastructure modules most applications need. Everything is interface-driven, concurrency-safe, and testable out of the box.

<br/>

## Quick Start

```bash
go get go.oease.dev/goe/v2
```

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
    "go.oease.dev/goe/v2/webresult"
)

func main() {
    _ = goe.New(goe.Options{
        WithHTTP: true,
        Invokers: []any{registerRoutes},
    })
    goe.Run() // blocks until SIGINT/SIGTERM
}

func registerRoutes(app contract.HTTPKernel, log contract.Logger) {
    app.App().Get("/", func(c fiber.Ctx) error {
        return c.SendString("Hello from GOE!")
    })

    app.App().Get("/health", func(c fiber.Ctx) error {
        return webresult.SendSucceed(c, fiber.Map{"status": "healthy"})
    })

    log.Info("Routes registered")
}
```

```bash
go run main.go
# → http://localhost:8080
```

> See the full [examples/](examples/) directory for real-world patterns including REST APIs with MongoDB, custom modules, and production configuration.

<br/>

## Features

### Core

| | Feature | What You Get |
|---|---|---|
| 🔌 | **Dependency Injection** | Uber Fx — type-safe, automatic resolution, lifecycle hooks |
| 🌐 | **HTTP Server** | GoFiber v3 — fast routing, middleware, request logging |
| 📝 | **Logging** | Uber Zap — structured JSON for production, colored console for dev |
| ⚙️ | **Configuration** | `.env` files with environment layering, type-safe accessors |

### Data

| | Feature | What You Get |
|---|---|---|
| 🗄️ | **SQL Databases** | GORM with MySQL, PostgreSQL, SQLite, SQL Server. Multiple named connections |
| 🍃 | **MongoDB** | Native driver with pooling, transactions, helper utilities, and a fluent migration system |
| 💾 | **Caching** | Unified interface across 10+ backends — Redis, Memory, S3, DynamoDB, Badger, and more |

### Infrastructure

| | Feature | What You Get |
|---|---|---|
| 📋 | **Job System** | Redis-backed background processing with scheduling, retries, and dead letter queues |
| 🔐 | **Distributed Lock** | Redlock algorithm — single instance, Sentinel, or Cluster modes |
| 🩺 | **Health & Metrics** | Built-in health checks and OpenTelemetry integration |

### Design

| | Feature | What You Get |
|---|---|---|
| 🧩 | **Module System** | Managed lifecycles (`OnStart`/`OnStop`), config validation, clean boundaries |
| 🛡️ | **Contracts** | Interface-driven design — swap implementations, mock in tests |
| ⚡ | **Concurrency Safe** | Singleflight cache protection, atomic operations, race-free by default |

<br/>

## Enabling Modules

Turn on what you need:

```go
goe.New(goe.Options{
    WithHTTP:    true,  // HTTP server (Fiber v3)
    WithCache:   true,  // Caching (10+ backends)
    WithDB:      true,  // SQL database (GORM)
    WithMongoDB: true,  // MongoDB with connection pooling
    WithMigrate: true,  // MongoDB schema migrations
    WithJob:     true,  // Background job processing
    WithLock:    true,  // Distributed mutex (Redlock)
})
```

<br/>

## Configuration

GOE loads config from multiple sources, in priority order:

| Priority | Source | Purpose |
|:---:|---|---|
| 1 | `.env` | Base defaults |
| 2 | `.local.env` | Local dev overrides (gitignored) |
| 3 | `.{GOE_ENV}.env` | Environment-specific (`prod`, `staging`) |
| 4 | **System ENV** | Deployment overrides — highest priority |

See [`.example.env`](.example.env) for every available option with documentation.

### Or configure in Go

The HTTP server can be configured entirely in code, with no `.env` required:

```go
import goehttp "go.oease.dev/goe/v2/core/http"

goe.New(goe.Options{
    // Passing options enables the module — WithHTTP: true is not needed
    HTTP: []goehttp.Option{
        goehttp.WithPort(8080),
        goehttp.WithBodyLimit(16 << 20),
        goehttp.WithCertFile("server.crt"),   // TLS, which has no env equivalent
        goehttp.WithCertKeyFile("server.key"),
    },
})
```

Options sit above the environment: a field set in code wins, and every field
left alone still reads from `.env`, so adding options to an existing app
changes nothing else. Every field of `fiber.Config` and `fiber.ListenConfig` is
available as `With<FieldName>`, plus `WithFiberConfig` for anything GOE does not
wrap. See [CONFIGURATION.md](CONFIGURATION.md#configuring-in-go-code) and
[examples/06-code-first-config](examples/06-code-first-config/).

<br/>

## Documentation

Full documentation is available on DeepWiki:

**[deepwiki.com/oeasenet/goe](https://deepwiki.com/oeasenet/goe)**

<br/>

## Contributing

Contributions are welcome.

1. Fork the repository and create a feature branch
2. Run `go mod download` to install dependencies
3. Make your changes and ensure tests pass:
   ```bash
   go test -race ./...
   ```
4. Open a pull request against `v2` with a clear description

Found a bug or have a question? [Open an issue](https://github.com/oeasenet/goe/issues).

<br/>

## License

MIT — see [LICENSE](LICENSE) for details.
