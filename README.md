# Goe - Modern Go Application Framework

<div align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-Apache_2.0-green?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Status-Beta-yellow?style=for-the-badge" alt="Status">
</div>

Goe is a modern Go application framework that combines the best practices from frameworks in other languages and
GoFiber. Built entirely on Uber's Fx dependency injection framework, Goe prioritizes developer experience,
extensibility, and concurrent safety.

## 🚀 Features

- **🔌 Dependency Injection**: Built on Uber's Fx for powerful, type-safe dependency injection
- **🌐 HTTP Server**: Integrated GoFiber v3 with automatic request logging and middleware
- **📝 Structured Logging**: Uber's Zap logger with pretty console output for development
- **⚙️ Configuration Management**: Environment-based configuration with hot reload support
- **💾 Cache Support**: Multiple cache drivers via Fiber's storage interface (Memory, Redis, SQLite, etc.)
- **🔧 Module System**: Extensible module system with lifecycle hooks
- **🛡️ Type Safety**: Leverages Go's type system for compile-time safety
- **🎯 Developer Experience**: Simple global accessors and intuitive APIs
- **🔄 Concurrent Safe**: Thread-safe operations throughout the framework

## 🚀 Getting Started

Install the dependency:

```bash
go get go.oease.dev/goe/v2
```

Create an application:

```go
package main

import "go.oease.dev/goe/v2"

func main() {
    _ = goe.New(goe.Options{WithHTTP: true})
    goe.Run()
}
```

For a step by step guide see [docs/Getting Started](docs/getting-started.md).

## 📖 Documentation

Detailed documentation is available in the [docs](docs/README.md) directory. Topics include:

- Configuration
- HTTP, Cache and Database modules
- Examples and Best Practices

## Contributing

Pull requests are welcome. Please run `go test ./...` before submitting.

## License

Released under the [Apache 2.0](LICENSE) license.
