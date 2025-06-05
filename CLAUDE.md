# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Goe is a modern Go application framework built on Uber's Fx dependency injection framework, combining GoFiber v3 for HTTP handling with structured configuration management and logging. The framework follows a configuration-first approach where application behavior is controlled through environment variables rather than code.

## Core Architecture

### Dependency Injection Layer
The framework is built entirely around Uber's Fx for dependency injection:
- **Global Instance Pattern**: The main `goe.go` maintains a singleton instance with thread-safe access to core services
- **Module System**: All functionality is organized into modules (`contract.Module`) with lifecycle hooks (`OnStart`/`OnStop`)
- **Service Injection**: Services are injected via Fx providers and accessible through global accessors (`goe.Config()`, `goe.Log()`, `goe.HTTP()`)

### Configuration System
Configuration follows a strict priority hierarchy:
1. `.env` (base configuration, committed to repo)
2. `.local.env` (local overrides, gitignored) 
3. `.{GOE_ENV}.env` (environment-specific, e.g., `.production.env`)
4. System environment variables (highest priority)

The `GOE_ENV` variable determines which environment-specific file to load (defaults to "dev").

### Core Modules Architecture
- **Config Module** (`core/config/`): Handles environment file loading with hot reload capability
- **Log Module** (`core/log/`): Zap-based structured logging with request correlation
- **HTTP Module** (`core/http/`): Fiber v3 integration with validation, context helpers, and comprehensive configuration
- **App Module** (`core/app/`): Application lifecycle management with Fx container

### HTTP Request Flow
1. Fiber handles request with configured settings from `FIBER_*` environment variables
2. Request ID middleware assigns unique ID for tracing
3. Service injection middleware provides access to config, logger, validator via context helpers
4. Request/response logging with structured fields
5. Custom error handling with content negotiation (JSON/HTML/plain text)

### Contracts System
All interfaces are defined in `contract/` package to enable clean dependency injection:
- `contract.Application`: Main app interface with lifecycle methods
- `contract.Config`: Configuration access with type-safe getters
- `contract.Logger`: Structured logging with field support  
- `contract.HTTPKernel`: HTTP server abstraction over Fiber
- `contract.Module`: Module lifecycle interface

## Development Commands

### Testing
```bash
# Run all tests
go test ./...

# Run specific test package
go test ./tests

# Run specific test function
go test ./tests -run TestHTTPValidator -v

# Run tests with coverage
go test ./... -cover
```

### Building and Running
```bash
# Build the module
go build ./...

# Run example applications
go run example/main.go
go run example/validation_demo/main.go

# Update dependencies
go mod tidy
```

### Configuration Testing
```bash
# Test different environments
GOE_ENV=production go run example/main.go
GOE_ENV=staging go run example/main.go

# Test configuration priority
echo "TEST_VAR=from_env" > .env
echo "TEST_VAR=from_local" > .local.env
TEST_VAR=from_system go run example/main.go
```

## Key Implementation Patterns

### Module Creation
Modules must implement the `contract.Module` interface:
```go
type MyModule struct {
    config contract.Config
    logger contract.Logger
}

func (m *MyModule) Name() string { return "mymodule" }
func (m *MyModule) OnStart(ctx context.Context) error { /* setup */ }
func (m *MyModule) OnStop(ctx context.Context) error { /* cleanup */ }
```

### HTTP Handler Patterns
Use context helpers instead of global accessors in handlers:
```go
func handler(c fiber.Ctx) error {
    logger := http.GetLogger(c)    // Includes request ID
    config := http.GetConfig(c)
    validator := http.GetValidator(c)
    
    // handler logic
}
```

### Configuration-First Principle
- Application name/version via `APP_NAME`/`APP_VERSION` environment variables
- Fiber configuration via `FIBER_*` environment variables
- Never hardcode configuration in `goe.Options` struct
- Use environment-specific files for different deployment targets

### Validation Integration
- go-playground/validator is integrated as Fiber's StructValidator
- Custom validations registered via `CustomValidator.RegisterValidation()`
- Accessible in handlers via `http.GetValidator(c)`

## File Structure Notes

- `goe.go`: Main framework entry point with global instance management
- `contract/`: All interface definitions for dependency injection
- `core/`: Core module implementations (config, log, http, app)
- `tests/`: Comprehensive test suite including integration tests
- `example/`: Example applications demonstrating features
- `utils/`: Utility functions used across the framework
- `types.go`: Common types and version constants

## Testing Strategy

The framework uses table-driven tests with environment variable setup/teardown. HTTP tests use Fiber's `Test()` method for in-memory testing without actual server startup. Configuration tests verify priority order by setting up multiple env files and checking which values take precedence.