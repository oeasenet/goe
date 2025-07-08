# Caching

GOE provides a flexible caching system that can help improve your application's performance by storing frequently accessed data in memory or external cache stores.

## Overview

The caching module in GOE:
- Provides a unified interface for different cache backends
- Supports TTL (Time-To-Live) for automatic cache expiration
- Includes JSON serialization for complex data types
- Integrates with the dependency injection system
- Supports multiple named cache stores

## Supported Cache Drivers

GOE currently supports the following cache drivers:

- **Memory**: In-memory cache using Fiber's memory storage (default)
- **Redis**: Redis-based caching using the Rueidis client

> **Note**: While GOE's architecture supports additional drivers, currently only memory and Redis are implemented. Support for other drivers like SQLite, PostgreSQL, MySQL, MongoDB, and filesystem may be added in future versions.

## Configuration

Cache behavior is configured via environment variables, typically prefixed with `CACHE_`.

### Basic Configuration

```bash
# .env file
CACHE_DRIVER=memory          # or 'redis'
CACHE_PREFIX=myapp          # Key prefix (defaults to APP_NAME)
CACHE_TTL=30m               # Default TTL (defaults to 2 hours)
```

### Redis Configuration

When using Redis as the cache driver:

```bash
CACHE_DRIVER=redis
CACHE_REDIS_HOST=localhost
CACHE_REDIS_PORT=6379
CACHE_REDIS_PASSWORD=
CACHE_REDIS_DB=0
```

### Multiple Cache Stores

You can configure multiple named cache stores:

```bash
# Default store
CACHE_DRIVER=memory

# Redis store for sessions
CACHE_SESSIONS_DRIVER=redis
CACHE_SESSIONS_PREFIX=sess_
CACHE_SESSIONS_TTL=1h

# Memory store for temporary data
CACHE_TEMP_DRIVER=memory
CACHE_TEMP_PREFIX=temp_
CACHE_TEMP_TTL=5m
```

## Enabling Cache Module

Enable the cache module in your GOE application:

```go
package main

import (
    "go.oease.dev/goe/v2"
)

func main() {
    goe.New(goe.Options{
        WithCache: true,  // Enable cache module
        WithHTTP:  true,
    })
    
    goe.Run()
}
```

## Accessing the Cache

### Using Global Accessor

```go
import (
    "go.oease.dev/goe/v2"
    "time"
)

func someFunction() {
    cache := goe.Cache()
    
    // Set a value
    cache.Set("user:123", "John Doe", 10*time.Minute)
    
    // Get a value
    value, err := cache.Get("user:123")
    if err != nil {
        // Handle error
    }
    
    // Check if value is a string
    if str, ok := value.(string); ok {
        fmt.Println("User name:", str)
    }
}
```

### Using Dependency Injection

```go
import (
    "go.oease.dev/goe/v2/contract"
    "time"
)

type UserService struct {
    cache contract.Cache
}

func NewUserService(cache contract.Cache) *UserService {
    return &UserService{cache: cache}
}

func (s *UserService) GetUserFromCache(id string) (string, error) {
    value, err := s.cache.Get("user:" + id)
    if err != nil {
        return "", err
    }
    
    if str, ok := value.(string); ok {
        return str, nil
    }
    
    return "", errors.New("user not found in cache")
}

func (s *UserService) CacheUser(id, name string) error {
    return s.cache.Set("user:"+id, name, 15*time.Minute)
}
```

## Cache Operations

### Basic Operations

```go
cache := goe.Cache()

// Set a value with TTL
cache.Set("key", "value", 5*time.Minute)

// Get a value
value, err := cache.Get("key")

// Delete a value
cache.Delete("key")

// Check if key exists
exists := cache.Has("key")

// Clear all cache (if supported by driver)
cache.Clear()
```

### Working with Complex Data

GOE cache automatically handles JSON serialization:

```go
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

user := User{ID: 123, Name: "John Doe"}

// Cache complex data
cache.Set("user:123", user, 10*time.Minute)

// Retrieve complex data
value, err := cache.Get("user:123")
if err != nil {
    // Handle error
}

// Type assertion for complex data
if userData, ok := value.(User); ok {
    fmt.Printf("User: %+v\n", userData)
}
```

### Remember Pattern

The "Remember" pattern retrieves from cache or computes and caches the value:

```go
func (s *UserService) GetUser(id int) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", id)
    
    // Try to get from cache first
    if value, err := s.cache.Get(cacheKey); err == nil {
        if user, ok := value.(*User); ok {
            return user, nil
        }
    }
    
    // Not in cache, fetch from database
    user, err := s.db.GetUser(id)
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    s.cache.Set(cacheKey, user, 10*time.Minute)
    
    return user, nil
}
```

## Multiple Cache Stores

Access different cache stores by name:

```go
// Get the cache manager
cacheManager := goe.CacheManager()

// Get different stores
defaultCache := cacheManager.Store("default")
sessionCache := cacheManager.Store("sessions")
tempCache := cacheManager.Store("temp")

// Use different stores
sessionCache.Set("session:abc123", sessionData, 1*time.Hour)
tempCache.Set("temp:data", tempData, 5*time.Minute)
```

## Cache in HTTP Handlers

Using cache in Fiber handlers:

```go
import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2/contract"
)

func NewUserHandler(cache contract.Cache) *UserHandler {
    return &UserHandler{cache: cache}
}

type UserHandler struct {
    cache contract.Cache
}

func (h *UserHandler) GetUser(c fiber.Ctx) error {
    userID := c.Params("id")
    cacheKey := "user:" + userID
    
    // Try cache first
    if value, err := h.cache.Get(cacheKey); err == nil {
        return c.JSON(value)
    }
    
    // Fetch from database (mock)
    user := User{ID: userID, Name: "John Doe"}
    
    // Cache the result
    h.cache.Set(cacheKey, user, 10*time.Minute)
    
    return c.JSON(user)
}
```

## Best Practices

### 1. Cache Key Naming

Use consistent, hierarchical key naming:

```go
// Good
"user:123"
"user:123:profile"
"session:abc123"
"product:456:details"

// Avoid
"user123"
"u123"
"userdata"
```

### 2. TTL Management

Set appropriate TTL values based on data volatility:

```go
// Frequently changing data
cache.Set("stock:123", stockData, 30*time.Second)

// User profile data
cache.Set("user:123:profile", profile, 15*time.Minute)

// Static configuration
cache.Set("config:settings", settings, 1*time.Hour)
```

### 3. Error Handling

Always handle cache errors gracefully:

```go
func GetUser(id string) (*User, error) {
    // Try cache first, but don't fail if cache is down
    if value, err := cache.Get("user:" + id); err == nil {
        if user, ok := value.(*User); ok {
            return user, nil
        }
    }
    
    // Fallback to database
    return database.GetUser(id)
}
```

### 4. Cache Invalidation

Implement proper cache invalidation:

```go
func UpdateUser(user *User) error {
    // Update in database
    if err := database.UpdateUser(user); err != nil {
        return err
    }
    
    // Invalidate cache
    cache.Delete("user:" + user.ID)
    cache.Delete("user:" + user.ID + ":profile")
    
    return nil
}
```

## Testing with Cache

Mock the cache for testing:

```go
type MockCache struct {
    data map[string]interface{}
}

func (m *MockCache) Get(key string) (interface{}, error) {
    if value, exists := m.data[key]; exists {
        return value, nil
    }
    return nil, errors.New("key not found")
}

func (m *MockCache) Set(key string, value interface{}, ttl time.Duration) error {
    m.data[key] = value
    return nil
}

func TestUserService(t *testing.T) {
    mockCache := &MockCache{data: make(map[string]interface{})}
    service := NewUserService(mockCache)
    
    // Test caching behavior
    service.CacheUser("123", "John Doe")
    
    name, err := service.GetUserFromCache("123")
    assert.NoError(t, err)
    assert.Equal(t, "John Doe", name)
}
```

## Performance Considerations

1. **Memory Usage**: Monitor memory usage when using in-memory cache
2. **Network Latency**: Redis adds network overhead but provides persistence
3. **Serialization**: JSON serialization has overhead for complex objects
4. **TTL Strategy**: Balance between cache hit rate and data freshness

## Troubleshooting

### Common Issues

1. **Cache Not Working**: Ensure `WithCache: true` is set in `goe.Options`
2. **Redis Connection**: Verify Redis connection parameters
3. **Memory Usage**: Monitor memory usage with in-memory cache
4. **Type Assertions**: Handle type assertions carefully for cached data

### Debug Cache Operations

Enable debug logging to see cache operations:

```bash
LOG_LEVEL=debug
```

This will show cache hit/miss operations in the logs.

## Next Steps

- [**Database Integration**](./database.md) - Learn about database operations
- [**HTTP Server**](./http-server.md) - Understand HTTP request handling
- [**Observability**](./observability.md) - Monitor cache performance