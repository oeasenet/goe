# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GOE is a modern Go application framework built on Uber's Fx dependency injection system and GoFiber v3 for HTTP. It follows contract-driven design with all major components defined as interfaces in `/contract`.

## Essential Commands

### Testing
```bash
# Run all tests with race detection
make test

# Run tests with coverage report
make test_coverage

# Run specific test suites
make test_contract    # Interface/contract tests
make test_core        # Core component unit tests
make test_integration # Integration tests

# Run a single test
go test -v -run TestName ./tests/...
```

### Development
```bash
# Install dependencies
go mod download

# Run go mod tidy (required before commits)
go mod tidy

# Check for vulnerabilities
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Documentation development (uses bun, not npm)
bun run docs:dev      # Start VitePress docs server
bun run docs:build    # Build documentation site
bun run docs:preview  # Preview built documentation
```

### Documentation Deployment
```bash
# Deploy docs to GitHub Pages
git push origin v2    # Triggers automatic deployment via GitHub Actions
```

## Architecture

### Core Design Principles

1. **Contract-Driven Design**: All components are defined by interfaces in `/contract`
2. **Dependency Injection**: Built entirely on Uber's Fx framework
3. **Module Pattern**: Each component implements `contract.Module` with lifecycle hooks

### Project Structure
```
goe/
├── contract/          # Interface definitions
├── core/             # Core implementations
│   ├── app/          # Application lifecycle
│   ├── cache/        # Caching (memory, Redis only)
│   ├── config/       # Configuration management
│   ├── db/           # Database (GORM)
│   ├── event/        # Event system (empty - not implemented)
│   ├── http/         # HTTP server (GoFiber v3)
│   ├── log/          # Logging (Zap)
│   └── observability/# Metrics and tracing
├── middlewares/      # HTTP middleware
├── utils/            # Utilities
├── webresult/        # Web response helpers
├── tests/            # All tests
└── docs/             # VitePress documentation
    ├── guide/        # Comprehensive guides
    ├── examples/     # Practical examples
    └── reference/    # API reference (future)
```

### Key Patterns

1. **Module Implementation**:
   - Implement `contract.Module` interface
   - Provide `Name()`, `OnStart(ctx)`, `OnStop(ctx)` methods
   - Register with Fx container

2. **Service Access**:
   - Via dependency injection (preferred)
   - Via global accessors: `goe.App()`, `goe.Config()`, `goe.Log()`, etc.
   - Via Fiber context injection for HTTP handlers

3. **Configuration**:
   - Environment-based: `.env` → `.local.env` → `.{GOE_ENV}.env`
   - Access via `contract.Config` interface

## Framework Capabilities

### Supported Features
- **HTTP Server**: GoFiber v3 with middleware, routing, JSON handling
- **Database**: GORM with PostgreSQL, MySQL, SQLite, SQL Server drivers
- **Caching**: Memory and Redis drivers (not SQLite, PostgreSQL, etc. as docs once claimed)
- **Logging**: Zap-based structured logging with console/JSON formats
- **Configuration**: Environment variables and .env file support
- **Observability**: OpenTelemetry metrics and tracing
- **Dependency Injection**: Full Fx integration with lifecycle management

### Module System
- All core features are optional modules (`WithHTTP`, `WithDB`, `WithCache`, `WithObservability`)
- Custom modules can be created implementing `contract.Module`
- Automatic dependency resolution and lifecycle management
- Clean shutdown with proper resource cleanup

## Testing Guidelines

- Use table-driven tests for comprehensive coverage
- Mock dependencies using testify/mock
- Test against interfaces, not implementations
- Always run with `-race` flag
- Use `GOE_ENV=test` for test configurations
- Follow repository pattern for data access
- Test HTTP handlers using Fiber's test utilities

## Working with Documentation

The project uses VitePress for documentation:

### Local Development
```bash
bun run docs:dev      # Start local dev server at http://localhost:5173
```

### Building
```bash
bun run docs:build    # Build static site to docs/.vitepress/dist
```

### Deployment
- Automatic deployment to GitHub Pages via GitHub Actions
- Triggered on pushes to `v2` branch that modify `docs/` folder
- Accessible at `https://oeasenet.github.io/goe/`

### Documentation Structure
- **Guide**: Comprehensive tutorials and explanations
- **Examples**: Practical, runnable code examples
- **Reference**: API documentation (planned)

## Working with Modules

When creating new modules:
1. Define interface in `/contract`
2. Implement in `/core/{module}`
3. Follow existing module patterns
4. Add comprehensive tests in `/tests`
5. Register module in application options
6. Document in `/docs/guide/` if it's a major feature

## Common Patterns

### Basic Application Setup
```go
goe.New(goe.Options{
    WithHTTP:          true,  // Enable HTTP server
    WithDB:            true,  // Enable database
    WithCache:         true,  // Enable caching
    WithObservability: true,  // Enable metrics/tracing
    Providers: []any{
        // Your services
    },
    Invokers: []any{
        // Route registration, etc.
    },
})
```

### Service Layer Pattern
```go
type UserService struct {
    repository UserRepository
    logger     contract.Logger
}

func NewUserService(repository UserRepository, logger contract.Logger) *UserService {
    return &UserService{repository: repository, logger: logger}
}
```

### HTTP Handler Pattern
```go
func NewUserHandler(userService *UserService) *UserHandler {
    return &UserHandler{userService: userService}
}

func (h *UserHandler) GetUser(c fiber.Ctx) error {
    // Handler implementation
}
```

## Important Notes

- **Go version**: Project specifies 1.24 in go.mod but this version doesn't exist; works with 1.21+ (tested with 1.23)
- **Main branch**: `v2`
- **Framework uses GoFiber v3** (beta) - check docs for v3-specific APIs
- **All components** are designed for concurrent safety
- **Use structured logging** with Zap
- **Prefer dependency injection** over global state
- **Cache drivers**: Only memory and Redis are implemented (not SQLite, PostgreSQL, etc.)
- **Event system**: Directory exists but is not implemented
- **Package manager**: Project uses bun, not npm
- **Documentation**: Comprehensive VitePress site with examples and guides

## Recent Improvements

- Created comprehensive VitePress documentation site
- Fixed documentation inaccuracies (Go version, cache drivers)
- Added practical examples (basic app, CRUD, custom modules, authentication)
- Set up automatic GitHub Pages deployment
- Organized documentation into logical structure (guide, examples, reference)
- Added search functionality and responsive design