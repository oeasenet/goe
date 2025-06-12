# Getting Started

This guide shows how to install Goe, initialise a new project, and run a simple HTTP server. It also explains how modules can be used directly via dependency injection or through the global helper functions that Goe exposes.

## Installation

```bash
go get go.oease.dev/goe/v2
```

Goe requires Go 1.21 or newer.

## Creating a Project

A typical Goe application keeps all domain code under `internal/` and entry points under `cmd/`.

```text
myapp/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── user/
│   │   ├── service.go
│   │   └── handler.go
└── go.mod
```

This layout keeps commands and implementation separated. Each package inside `internal/` should own its dependencies and be free from circular imports.

Initialise a Go module and install dependencies:

```bash
mkdir myapp && cd myapp
go mod init example.com/myapp
```

## Minimal Example

Create `cmd/api/main.go` with the following contents:

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
)

func main() {
    _ = goe.New(goe.Options{
        WithHTTP: true,
        Invokers: []any{
            func(h contract.HTTPKernel, log contract.Logger) {
                app := h.App()
                app.Get("/", func(c fiber.Ctx) error {
                    log.Info("home")
                    return c.SendString("hello")
                })
            },
        },
    })

    goe.Run()
}
```

Run the application:

```bash
go run ./cmd/api
```

Navigate to `http://localhost:8080` to see the response.

## Dependency Injection vs Global Helpers

Goe exposes global helper functions such as `goe.Log()` and `goe.Config()` for convenience. In larger applications it is often better to inject these dependencies explicitly.

```go
func withGlobals() {
    logger := goe.Log()
    logger.Info("using global logger")
}

func withDI(log contract.Logger) {
    log.Info("injected logger")
}
```

Both approaches are supported. When using HTTP handlers, you can also fetch components from the request context via the `http` package.

```go
func handler(c fiber.Ctx) error {
    log := http.GetLogger(c)
    log.Info("from context")
    return nil
}
```
