# GOE Framework Examples

This directory contains example applications demonstrating the GOE framework's features and best practices.

## Prerequisites

- Go 1.26+
- Docker and Docker Compose (for database/cache dependencies)

## Quick Start

```bash
# Start required services (Redis, MongoDB, PostgreSQL)
docker compose up -d

# Run any example
cd 01-hello-world && go run .
```

## Examples Overview

### [01-hello-world](./01-hello-world/)

**Minimal HTTP server** - The simplest GOE application demonstrating:
- Basic GOE initialization with `goe.New()`
- HTTP route registration with Fiber
- Logger usage via dependency injection
- Configuration access via `contract.Config`
- JSON responses with `webresult` helpers

```bash
cd 01-hello-world && go run .

# Test endpoints
curl http://localhost:3000/
curl http://localhost:3000/hello/World
curl http://localhost:3000/health
curl -X POST http://localhost:3000/echo -H "Content-Type: application/json" -d '{"message":"hello"}'
```

---

### [02-todo-api](./02-todo-api/)

**REST API with MongoDB and Redis caching** - A complete CRUD API demonstrating:
- MongoDB integration for data persistence
- Redis caching with the `Remember` pattern
- Service layer pattern with dependency injection
- Handler pattern for HTTP endpoints
- Input validation with `go-playground/validator`
- Error handling patterns

```bash
cd 02-todo-api && go run .

# CRUD operations
curl http://localhost:3000/todos
curl -X POST http://localhost:3000/todos -H "Content-Type: application/json" \
  -d '{"title":"Learn GOE","description":"Study the framework"}'
curl http://localhost:3000/todos/{id}
curl -X PUT http://localhost:3000/todos/{id} -H "Content-Type: application/json" \
  -d '{"completed":true}'
curl -X DELETE http://localhost:3000/todos/{id}

# Stats endpoint (demonstrates caching)
curl http://localhost:3000/todos/stats/summary
```

---

### [04-custom-module](./04-custom-module/)

**Custom modules with lifecycle hooks** - Advanced patterns demonstrating:
- Creating custom modules implementing `contract.Module`
- Module lifecycle hooks (OnStart, OnStop)
- Configuration validation with `ConfigValidator`
- Background workers with graceful shutdown
- Health check aggregation patterns

```bash
cd 04-custom-module && go run .

# Test endpoints
curl http://localhost:3000/health
curl http://localhost:3000/metrics
curl http://localhost:3000/stats
curl http://localhost:3000/demo  # Records a request metric
```

### [06-code-first-config](./06-code-first-config/)

**Configuring the HTTP server in Go instead of `.env`** - Demonstrating:
- Passing Fiber settings through `goe.Options.HTTP` (which also enables the module)
- How code, environment variables and GOE defaults layer together
- Reaching settings that have no environment equivalent, such as TLS
- The `WithFiberConfig` escape hatch for anything GOE does not wrap

```bash
cd 06-code-first-config && go run .

# Server header, body limit and routing flags all come from Go code
curl -i http://localhost:8080/

# Shows which values came from code and which fell through to the defaults
curl -s http://localhost:8080/config | jq .
```

### [07-cdn-trusted-proxy](./07-cdn-trusted-proxy/)

**Trusting a CDN's edge IPs so `Ctx.IP()` is the real client** - Demonstrating:
- Fetching published edge ranges with the optional `cdntrust` helper
- Failing startup on a fetch error rather than booting untrusted
- Setting a provider-specific client-IP header

```bash
cd 07-cdn-trusted-proxy && go run .   # requires internet access

curl -s http://localhost:8080/whoami | jq .

# A forged header from an untrusted hop is ignored — client_ip does not change
curl -s -H 'CF-Connecting-IP: 203.0.113.9' http://localhost:8080/whoami | jq .
```

---

## Architecture Patterns

### Dependency Injection

GOE uses [Uber Fx](https://github.com/uber-go/fx) for dependency injection:

```go
// Register providers (constructors)
goe.New(goe.Options{
    Providers: []any{
        NewTodoService,  // func(db MongoDB, cache Cache, logger Logger) *TodoService
        NewTodoHandler,  // func(service *TodoService, logger Logger) *TodoHandler
    },
})

// Dependencies are automatically injected
func NewTodoService(db contract.MongoDB, cache contract.CacheManager, logger contract.Logger) *TodoService {
    return &TodoService{db: db, cache: cache.Store(), logger: logger}
}
```

### Service Layer Pattern

```
Handler (HTTP) → Service (Business Logic) → Repository (Data Access)
```

- **Handlers**: Parse requests, validate input, call services, format responses
- **Services**: Implement business logic, coordinate data access
- **Contracts**: Define interfaces for loose coupling and testability

### Module Pattern

Custom modules implement `contract.Module`:

```go
type MyModule struct { /* dependencies */ }

func (m *MyModule) Name() string { return "my-module" }
func (m *MyModule) OnStart(ctx context.Context) error { /* init resources */ }
func (m *MyModule) OnStop(ctx context.Context) error { /* cleanup resources */ }
```

### Caching Patterns

```go
// Remember pattern - get from cache or compute
cache.Remember("key", &result, 5*time.Minute, func() (any, error) {
    return fetchFromDatabase()
})

// Manual cache operations
cache.Set("key", value, time.Hour)
cache.Get("key", &result)
cache.Forget("key")
cache.Forever("key", value)
```

---

## Configuration

GOE loads configuration from environment files in priority order:
1. `.env` - Base configuration
2. `.local.env` - Local overrides (gitignored)
3. `.{GOE_ENV}.env` - Environment-specific (e.g., `.prod.env`)
4. System environment variables (highest priority)

### Common Configuration Options

```bash
# Application
GOE_ENV=development
APP_NAME=my-app
APP_VERSION=1.0.0

# HTTP Server
HTTP_HOST=0.0.0.0
HTTP_PORT=3000

# Logging
LOG_LEVEL=debug          # debug, info, warn, error
LOG_FORMAT=text          # text, json
LOG_OUTPUT=console       # console, file

# MongoDB
MONGO_URI=mongodb://localhost:27017
MONGO_DB_NAME=myapp

# Cache (Redis)
CACHE_STORE=redis
CACHE_REDIS_URL=redis://localhost:6379/0
CACHE_PREFIX=myapp:
```

---

## Docker Services

The `docker-compose.yml` provides:

| Service    | Port  | Credentials                |
|------------|-------|----------------------------|
| Redis      | 6379  | No password                |
| MongoDB    | 27017 | goe:goepassword            |
| PostgreSQL | 5432  | goe:goepassword            |

```bash
# Start all services
docker compose up -d

# Stop all services
docker compose down

# Stop and remove volumes
docker compose down -v

# View logs
docker compose logs -f redis
```

---

## Testing the Examples

Each example can be tested with curl or any HTTP client:

```bash
# Health check (available in all examples)
curl http://localhost:3000/health

# JSON POST request
curl -X POST http://localhost:3000/endpoint \
  -H "Content-Type: application/json" \
  -d '{"key": "value"}'

# With jq for pretty output
curl -s http://localhost:3000/todos | jq .
```

---

## Learning Path

1. **Start with 01-hello-world** - Understand basic GOE structure
2. **Move to 02-todo-api** - Learn MongoDB, caching, and service patterns
3. **Study 04-custom-module** - Master module creation and lifecycle
4. **Explore 05-production-essentials** - Learn about health checks, metrics, and tracing
5. **Read 06-code-first-config** - Configure the server in Go rather than `.env`
6. **See 07-cdn-trusted-proxy** - Trust a CDN's edges so client IPs are correct

---

## Additional Resources

- [GOE Documentation](https://goe.oease.dev)
- [GoFiber Documentation](https://docs.gofiber.io)
- [Uber Fx Documentation](https://uber-go.github.io/fx/)
- [MongoDB Go Driver](https://pkg.go.dev/go.mongodb.org/mongo-driver/v2)
