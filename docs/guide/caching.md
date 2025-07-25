# Caching

GOE provides a flexible caching system that can help improve your application's performance by storing frequently accessed data in memory or external cache stores.

## Overview

The caching module in GOE:
- Provides a unified interface for different cache backends
- Supports TTL (Time-To-Live) for automatic cache expiration
- Includes JSON serialization for complex data types
- Integrates with the dependency injection system
- Supports multiple named cache stores
- Uses type-safe pointer-based API for better type safety and predictability

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
    "fmt"
)

func someFunction() {
    cache := goe.Cache()
    
    // Set a value
    err := cache.Set("user:123", "John Doe", 10*time.Minute)
    if err != nil {
        // Handle error
    }
    
    // Get a value (pointer-based)
    var userName string
    err = cache.Get("user:123", &userName)
    if err != nil {
        // Handle error (validation or store errors)
        // Note: Cache miss is NOT an error
    } else if userName != "" {
        // Got a value from cache
        fmt.Println("User name:", userName)
    } else {
        // Cache miss - userName remains empty string (zero value)
        fmt.Println("User not found in cache")
    }
    
    // Get with default value
    var userAge int
    err = cache.GetWithDefault("user:age:123", &userAge, 25)
    if err != nil {
        // Handle error
    }
    // userAge will be 25 if key doesn't exist, otherwise the cached value
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
    var userName string
    err := s.cache.Get("user:" + id, &userName)
    if err != nil {
        return "", err // Handle validation or store errors
    }
    if userName == "" {
        return "", errors.New("user not found in cache")
    }
    return userName, nil
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
err := cache.Set("key", "value", 5*time.Minute)

// Get a value (pointer-based)
var value string
err = cache.Get("key", &value)

// Get with default
var count int
err = cache.GetWithDefault("counter", &count, 0)

// Delete a value
err = cache.Forget("key")

// Check if key exists
exists := cache.Has("key")

// Clear all cache
err = cache.Flush()

// Store value forever (no TTL)
err = cache.Forever("permanent-key", "permanent-value")
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

// Retrieve complex data (pointer-based)
var retrievedUser User
err = cache.Get("user:123", &retrievedUser)
if err != nil {
    // Handle validation or store errors
    // Note: Cache miss is NOT an error
} else if retrievedUser.ID != 0 {
    // Check if we got a valid user (non-zero value)
    fmt.Printf("User: %+v\n", retrievedUser)
} else {
    // Cache miss - retrievedUser remains zero value
    fmt.Println("User not found in cache")
}
```

### Remember Pattern

The "Remember" pattern retrieves from cache or computes and caches the value:

```go
func (s *UserService) GetUser(id int) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", id)
    
    var user User
    err := s.cache.Remember(cacheKey, &user, 10*time.Minute, func() (any, error) {
        // This callback is only called if the key is not in cache
        return s.db.GetUser(id)
    })
    
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}

// Remember forever (no TTL)
func (s *UserService) GetConfig(key string) (*Config, error) {
    var config Config
    err := s.cache.RememberForever(key, &config, func() (any, error) {
        return s.db.GetConfig(key)
    })
    
    if err != nil {
        return nil, err
    }
    
    return &config, nil
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
    var user User
    err := h.cache.Get(cacheKey, &user)
    if err == nil {
        return c.JSON(user)
    }
    
    // Not in cache, fetch from database
    user = User{ID: userID, Name: "John Doe"}
    
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
    var user User
    err := cache.Get("user:" + id, &user)
    if err != nil {
        log.Warn("Cache error", "error", err)
        // Fallback to database on error
        return database.GetUser(id)
    }
    
    if user.ID != 0 {
        // Got a valid user from cache
        return &user, nil
    }
    
    // Cache miss - fallback to database
    log.Debug("Cache miss for user", "id", id)
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
    cache.Forget("user:" + user.ID)
    cache.Forget("user:" + user.ID + ":profile")
    
    return nil
}

// Pull pattern - get and remove in one operation
func ConsumeToken(tokenID string) (*Token, error) {
    var token Token
    err := cache.Pull("token:" + tokenID, &token)
    if err != nil {
        return nil, err
    }
    return &token, nil
}
```

## Testing with Cache

Mock the cache for testing:

```go
type MockCache struct {
    data map[string][]byte
    mu   sync.RWMutex
}

func (m *MockCache) Get(key string, value any) error {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    if data, exists := m.data[key]; exists {
        return json.Unmarshal(data, value)
    }
    return contract.ErrCacheMiss
}

func (m *MockCache) GetWithDefault(key string, value any, defaultValue any) error {
    err := m.Get(key, value)
    if errors.Is(err, contract.ErrCacheMiss) {
        // Use reflection to set default value
        rv := reflect.ValueOf(value)
        rdv := reflect.ValueOf(defaultValue)
        rv.Elem().Set(rdv)
        return nil
    }
    return err
}

func (m *MockCache) Set(key string, value any, ttl time.Duration) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    m.data[key] = data
    return nil
}

func (m *MockCache) Has(key string) bool {
    m.mu.RLock()
    defer m.mu.RUnlock()
    _, exists := m.data[key]
    return exists
}

func (m *MockCache) Forget(key string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    delete(m.data, key)
    return nil
}

func TestUserService(t *testing.T) {
    mockCache := &MockCache{data: make(map[string][]byte)}
    service := NewUserService(mockCache)
    
    // Test caching behavior
    err := service.CacheUser("123", "John Doe")
    assert.NoError(t, err)
    
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

## Migration Guide

### Migrating from the Old API

The cache module has been updated to use a pointer-based API for better type safety. Here's how to migrate your code:

#### Old API (returning `any`):
```go
// Get
value, err := cache.Get("key")
if err != nil {
    // handle error
}
if str, ok := value.(string); ok {
    // use str
}

// Remember
value, err := cache.Remember("key", ttl, func() (any, error) {
    return computeValue()
})

// Pull
value, err := cache.Pull("key")
```

#### New API (pointer-based):
```go
// Get - cache miss is not an error
var str string
err := cache.Get("key", &str)
if err != nil {
    // Handle validation or store errors
    // Note: Cache miss is NOT an error
} else if str != "" {
    // Got a value from cache
} else {
    // Cache miss - str remains empty (zero value)
}

// Get with default
var count int
err := cache.GetWithDefault("counter", &count, 0)
// count will be 0 if key doesn't exist

// Remember
var result MyStruct
err := cache.Remember("key", &result, ttl, func() (any, error) {
    return computeValue()
})

// Pull
var value MyType
err := cache.Pull("key", &value)
// No error for cache miss, value remains zero if key doesn't exist
```

### Key Benefits of the New API

1. **Type Safety**: No more type assertions - the compiler ensures type correctness
2. **Pointer Validation**: Automatic validation that values are non-nil pointers
3. **Go-idiomatic**: Cache misses are not errors - values remain at zero state
4. **Default Values**: Built-in support for default values with `GetWithDefault`
5. **Better IDE Support**: Auto-completion and type hints work correctly
6. **Simplified Error Handling**: Only actual errors (validation, store failures) are returned

## Next Steps

- [**Database Integration**](./database.md) - Learn about database operations
- [**HTTP Server**](./http-server.md) - Understand HTTP request handling
