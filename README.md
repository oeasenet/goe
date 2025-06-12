# Goe

**Goe** is a lightweight application framework for Go that brings dependency injection and sensible defaults. It is built on [Fx](https://github.com/uber-go/fx) and [Fiber](https://github.com/gofiber/fiber).

## Features

- Simple application bootstrap with modules
- Pluggable HTTP server using Fiber
- Structured logging via Zap
- Configuration loaded from `.env` files and environment variables
- Optional cache and database modules

## Getting Started

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

## Documentation

Detailed documentation is available in the [docs](docs/README.md) directory. Topics include:

- Configuration
- HTTP, Cache and Database modules
- Examples and Best Practices

## Contributing

Pull requests are welcome. Please run `go test ./...` before submitting.

## License

Released under the [Apache 2.0](LICENSE) license.
