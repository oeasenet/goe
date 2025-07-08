# Installation

This guide covers how to install and set up the GOE framework in your Go project.

## Prerequisites

- **Go**: Version 1.21 or newer (recommended: 1.23+)
- **Git**: For dependency management

> **Note**: The project currently specifies Go 1.24 in go.mod, but this version doesn't exist yet. The framework should work with Go 1.21+ and has been tested with Go 1.23.

## Installing GOE

### Using go get (Recommended)

```bash
go get go.oease.dev/goe/v2
```

### Using go mod

Add to your `go.mod` file:

```go
module your-project

go 1.21

require (
    go.oease.dev/goe/v2 latest
)
```

Then run:

```bash
go mod tidy
```

## Project Setup

### 1. Initialize Go Module

```bash
mkdir my-goe-app
cd my-goe-app
go mod init example.com/my-goe-app
```

### 2. Install GOE

```bash
go get go.oease.dev/goe/v2
```

### 3. Create Main Application

Create `main.go`:

```go
package main

import (
    "go.oease.dev/goe/v2"
)

func main() {
    // Initialize GOE with desired modules
    goe.New(goe.Options{
        WithHTTP: true,
    })

    // Start the application
    goe.Run()
}
```

### 4. Run Your Application

```bash
go run main.go
```

## Environment Configuration

Create a `.env` file in your project root:

```bash
# Application
APP_NAME=My GOE App
APP_VERSION=1.0.0
APP_ENV=development

# HTTP Server
HTTP_HOST=0.0.0.0
HTTP_PORT=8080

# Logging
LOG_LEVEL=info
LOG_FORMAT=console

# Database (if using WithDB: true)
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp
DB_USER=postgres
DB_PASSWORD=secret

# Cache (if using WithCache: true)
CACHE_DRIVER=memory
REDIS_HOST=localhost
REDIS_PORT=6379
```

## IDE Setup

### VS Code

Install the Go extension and add these settings to your `.vscode/settings.json`:

```json
{
    "go.useLanguageServer": true,
    "go.lintTool": "golangci-lint",
    "go.formatTool": "goimports",
    "go.testFlags": ["-v", "-race"],
    "go.buildFlags": ["-v"],
    "go.vetFlags": ["-all"]
}
```

### GoLand/IntelliJ

1. Enable Go modules support
2. Set GOROOT to your Go installation
3. Configure code style to use `gofmt`

## Development Tools

### Recommended Tools

```bash
# Code formatting
go install golang.org/x/tools/cmd/goimports@latest

# Linting
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Testing
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Air for hot reloading
go install github.com/cosmtrek/air@latest
```

### Hot Reloading with Air

Create `.air.toml`:

```toml
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main ."
  bin = "tmp/main"
  full_bin = "./tmp/main"
  include_ext = ["go", "tpl", "tmpl", "html"]
  exclude_dir = ["assets", "tmp", "vendor"]
  include_dir = []
  exclude_file = []
  log = "build-errors.log"
  delay = 1000
  stop_on_error = true

[log]
  time = false

[color]
  main = "magenta"
  watcher = "cyan"
  build = "yellow"
  runner = "green"

[misc]
  clean_on_exit = false
```

Run with:

```bash
air
```

## Docker Setup

Create `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/.env .

EXPOSE 8080
CMD ["./main"]
```

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - DB_HOST=db
    depends_on:
      - db
      - redis

  db:
    image: postgres:15
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: secret
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

## Troubleshooting

### Common Issues

#### Module Not Found
```bash
# Clear module cache
go clean -modcache

# Re-download dependencies
go mod download
```

#### Import Issues
```bash
# Update imports
go mod tidy

# Verify dependencies
go list -m all
```

#### Build Issues
```bash
# Check Go version
go version

# Verify module path
go list -m

# Check for syntax errors
go vet ./...
```

### Getting Help

- [GitHub Issues](https://github.com/oeasenet/goe/issues)
- [Documentation](https://goe.oease.dev)
- [Examples](../examples/basic-app.md)

## Next Steps

Now that you have GOE installed and configured:

1. [**Quick Start**](./getting-started.md) - Build your first application
2. [**Project Structure**](./project-structure.md) - Organize your code
3. [**Configuration**](./configuration.md) - Set up environment variables