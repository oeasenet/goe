# Goe Framework Architecture

This document provides a detailed overview of the Goe framework's architecture, design decisions, and internal workings.

## Table of Contents

- [Overview](#overview)
- [Core Architecture](#core-architecture)
- [Component Design](#component-design)
- [Dependency Flow](#dependency-flow)
- [Module System](#module-system)
- [Configuration Architecture](#configuration-architecture)
- [Logging Architecture](#logging-architecture)
- [HTTP Architecture](#http-architecture)
- [Thread Safety](#thread-safety)
- [Performance Considerations](#performance-considerations)

## Overview

Goe is built on a layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────┐
│                    Application Layer                     │
│                  (User's Application)                    │
├─────────────────────────────────────────────────────────┤
│                      Goe Framework                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │   Global    │  │   Module    │  │   Context   │    │
│  │  Accessors  │  │   System    │  │   Helpers   │    │
│  └─────────────┘  └─────────────┘  └─────────────┘    │
├─────────────────────────────────────────────────────────┤
│                    Core Modules                          │
│  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐    │
│  │ App  │  │Config│  │Logger│  │ HTTP │  │Cache │    │
│  └──────┘  └──────┘  └──────┘  └──────┘  └──────┘    │
├─────────────────────────────────────────────────────────┤
│                 Dependency Injection                     │
│                    (Uber's Fx)                          │
└─────────────────────────────────────────────────────────┘
```

## Core Architecture

### 1. Application Core

The application core (`core/app/app.go`) manages the application lifecycle:

```go
type app struct {
    name        string
    version     string
    environment string
    ctx         context.Context
    cancel      context.CancelFunc
    container   *fx.App
    mu          sync.RWMutex
}
```

**Key Responsibilities:**
- Application metadata management
- Lifecycle coordination
- Fx container management
- Context propagation

### 2. Global Instance Management

The global instance pattern (`goe.go`) provides convenient access:

```go
var instance struct {
    app    contract.Application
    config contract.Config
    logger contract.Logger
    http   contract.HTTPKernel
    mu     sync.RWMutex
}
```

**Thread Safety:**
- All global accessors use read locks
- Initialization uses write locks
- Prevents race conditions

### 3. Contract-Based Design

All major components implement contracts (interfaces):

```go
// Contract ensures loose coupling
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    // ...
}

// Implementation can be swapped
type zapLogger struct {
    logger *zap.Logger
}
```

## Component Design

### 1. Configuration Component

**Design Pattern:** Repository Pattern with Caching

```go
type config struct {
    data   map[string]any  // In-memory cache
    mu     sync.RWMutex    // Thread safety
    logger contract.Logger
}
```

**Features:**
- Environment file loading (.env, .env.{environment})
- Type-safe getters
- Hot reload capability
- Thread-safe operations

**Loading Order:**
1. Default values
2. `.env` file
3. `.env.{environment}` file
4. System environment variables

### 2. Logging Component

**Design Pattern:** Adapter Pattern over Zap

```go
type logger struct {
    logger *zap.Logger
    config contract.Config
}
```

**Features:**
- Structured logging
- Context propagation
- Request ID tracking
- Environment-based formatting

**Log Flow:**
```
User Call → Goe Logger → Zap Logger → Output
                ↓
          Field Conversion
```

### 3. HTTP Component

**Design Pattern:** Facade Pattern over Fiber

```go
type kernel struct {
    app    *fiber.App
    config contract.Config
    logger contract.Logger
}
```

**Middleware Stack:**
1. Recovery (panic protection)
2. Request ID generation
3. Service injection
4. Request logging
5. User middleware
6. Route handlers

## Dependency Flow

### 1. Registration Phase

```
goe.New() → Create Core Modules → Register with Fx → User Providers → User Invokers
```

### 2. Initialization Order

```
1. Config Module (provides configuration)
2. Log Module (depends on config)
3. HTTP Module (depends on config and logger)
4. User Modules (depend on core modules)
```

### 3. Service Injection

```go
// Registration
fx.Provide(NewUserService)

// Injection
func NewController(userService *UserService) *Controller

// Context injection for HTTP
ServiceProvider → ServiceMiddleware → fiber.Context → GetLogger/GetConfig
```

## Module System

### 1. Module Interface

```go
type Module interface {
    Name() string
    OnStart(ctx context.Context) error
    OnStop(ctx context.Context) error
}
```

### 2. Lifecycle Management

**Start Sequence:**
1. Fx calls module OnStart in dependency order
2. Modules initialize resources
3. Background tasks start
4. HTTP server starts

**Stop Sequence:**
1. Signal received (SIGINT/SIGTERM)
2. HTTP server stops accepting new requests
3. Active requests complete
4. Modules OnStop called in reverse order
5. Resources cleaned up
6. Application exits

### 3. Module Communication

Modules communicate through:
- Dependency injection
- Shared contracts
- Event system (future)

## Configuration Architecture

### 1. Storage Layer

```go
// In-memory storage with mutex protection
type config struct {
    data map[string]any
    mu   sync.RWMutex
}
```

### 2. Type Conversion

```go
// Safe type conversion with defaults
func (c *config) GetInt(key string) int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    if val, ok := c.data[key]; ok {
        // Type assertion and conversion
        return convertToInt(val)
    }
    return 0 // default
}
```

### 3. Environment Parsing

```go
// Parse .env files
func loadEnvFile(filename string) map[string]string {
    // Read file
    // Parse KEY=VALUE format
    // Handle quotes and escapes
    // Return parsed map
}
```

## Logging Architecture

### 1. Logger Hierarchy

```
Global Logger
    ├── Module Loggers (with module field)
    ├── Request Loggers (with request_id)
    └── Service Loggers (with service field)
```

### 2. Field Management

```go
// Structured fields
type Field interface {
    zap.Field
}

// Field creation
func NewField(key string, value any) Field {
    return zap.Any(key, value)
}
```

### 3. Output Formatting

**Development:**
- Colored console output
- Human-readable timestamps
- Pretty-printed JSON

**Production:**
- JSON output
- Machine-readable
- Structured for log aggregation

## HTTP Architecture

### 1. Request Flow

```
Request → Fiber → Recovery → RequestID → ServiceInjection → Logging → Router → Handler → Response
```

### 2. Context Enhancement

```go
// Services injected into context
fiber.Context → Locals("services") → Services{App, Config, Logger}
```

### 3. Error Handling

```go
// Centralized error handler
ErrorHandler → Log 5xx errors → Format response → Send JSON error
```

## Thread Safety

### 1. Global State Protection

```go
// All global access protected
func Config() contract.Config {
    instance.mu.RLock()
    defer instance.mu.RUnlock()
    return instance.config
}
```

### 2. Configuration Safety

```go
// Read operations use read lock
func (c *config) Get(key string) any {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.data[key]
}

// Write operations use write lock
func (c *config) Set(key string, value any) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = value
}
```

### 3. Concurrent Module Operations

- Module OnStart can run concurrently
- Module OnStop runs sequentially
- Shared resources protected by mutexes

## Performance Considerations

### 1. Zero Allocation Logging

```go
// Reuse field allocations
logger.With(fields...) // Creates new logger with fields
```

### 2. Configuration Caching

```go
// Values cached in memory
// No file I/O on Get operations
// Type conversions cached
```

### 3. HTTP Optimizations

```go
// Fiber's zero-allocation design
// Request pooling
// Minimal middleware overhead
```

### 4. Dependency Injection

```go
// One-time cost at startup
// No runtime reflection
// Direct function calls
```

## Design Decisions

### 1. Why Fx?

- **Type Safety**: Compile-time dependency checking
- **Lifecycle**: Built-in lifecycle management
- **Testing**: Easy to test with mocks
- **Performance**: No runtime overhead

### 2. Why Global Accessors?

- **Developer Experience**: Simple API
- **Compatibility**: Works with existing code
- **Optional**: Can use DI exclusively

### 3. Why Contracts?

- **Flexibility**: Easy to swap implementations
- **Testing**: Simple mocking
- **Documentation**: Clear API boundaries

### 4. Why Modules?

- **Organization**: Logical grouping
- **Lifecycle**: Managed startup/shutdown
- **Extensibility**: Easy to add features

## Extension Points

### 1. Custom Modules

```go
type CustomModule struct {
    // Your fields
}

func (m *CustomModule) OnStart(ctx context.Context) error {
    // Initialize
}
```

### 2. Custom Middleware

```go
func CustomMiddleware() fiber.Handler {
    return func(c fiber.Ctx) error {
        // Your logic
        return c.Next()
    }
}
```

### 3. Custom Providers

```go
func NewCustomService(logger contract.Logger) *CustomService {
    // Your implementation
}
```

## Future Architecture Plans

### 1. Event System

```go
type EventBus interface {
    Publish(event Event)
    Subscribe(topic string, handler Handler)
}
```

### 2. Plugin System

```go
type Plugin interface {
    Module
    Register(app contract.Application)
}
```

### 3. Clustering Support

```go
type Cluster interface {
    Join(nodes ...string)
    Leave()
    Broadcast(message Message)
}
```

## Conclusion

The Goe framework's architecture prioritizes:

1. **Simplicity**: Easy to understand and use
2. **Safety**: Thread-safe and type-safe
3. **Performance**: Minimal overhead
4. **Extensibility**: Easy to extend and customize

The layered architecture with clear contracts ensures that each component can evolve independently while maintaining compatibility.