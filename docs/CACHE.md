# Cache Module Documentation

The Goe cache module provides a unified interface for caching with support for multiple storage backends using Fiber's storage drivers.

## Features

- **Fiber Storage Compatible**: Uses Fiber v3's storage interface, allowing you to use any Fiber storage driver
- **Configuration-based**: Choose and configure drivers through environment variables
- **Type-safe Operations**: Generic functions for type-safe cache operations
- **Multiple Stores**: Support for multiple cache stores with different configurations
- **Built-in Drivers**: Memory and Redis drivers included

## Configuration

### Basic Configuration

```bash
# Choose the cache driver (default: memory)
CACHE_DRIVER=memory

# Set cache prefix
CACHE_PREFIX=myapp

# Default TTL for cache entries
CACHE_TTL=2h
```

### Memory Driver Configuration

The memory driver is the default and requires no additional configuration:

```bash
CACHE_DRIVER=memory

# Optional: Configure garbage collection interval
CACHE_MEMORY_GC_INTERVAL=10s
```

### Redis Driver Configuration

The Redis driver uses the high-performance rueidis client with auto-pipelining:

```bash
CACHE_DRIVER=redis

# Option 1: Use Redis URL
CACHE_REDIS_URL=redis://username:password@localhost:6379/0

# Option 2: Use individual settings
CACHE_REDIS_HOSTS=localhost:6379,localhost:6380  # Comma-separated for cluster
CACHE_REDIS_USERNAME=
CACHE_REDIS_PASSWORD=
CACHE_REDIS_DATABASE=0
CACHE_REDIS_CLIENT_NAME=myapp

# Advanced settings
CACHE_REDIS_CACHE_SIZE=134217728          # Client-side cache size (default: 128MB)
CACHE_REDIS_BLOCKING_POOL_SIZE=1000       # Connection pool size for blocking commands
CACHE_REDIS_PIPELINE_MULTIPLEX=2          # TCP connections for pipelining
CACHE_REDIS_DISABLE_RETRY=false           # Disable retry on network errors
CACHE_REDIS_DISABLE_CACHE=false           # Disable client-side caching
CACHE_REDIS_ALWAYS_PIPELINING=true        # Always pipeline commands
CACHE_REDIS_CACHE_TTL=1m                  # Client-side cache TTL
```

### Multiple Stores Configuration

You can configure multiple cache stores with different drivers:

```bash
# Default store
CACHE_STORE=primary
CACHE_primary_DRIVER=redis
CACHE_primary_REDIS_URL=redis://localhost:6379/0
CACHE_primary_PREFIX=app

# Secondary store (e.g., for sessions)
CACHE_secondary_DRIVER=memory
CACHE_secondary_PREFIX=sessions
```

## Usage

### Basic Usage

```go
package main

import (
    "time"
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
)

func main() {
    app := goe.New(goe.Options{
        WithCache: true,
    })
    
    // Get cache from dependency injection
    var cache contract.Cache
    app.Container().Invoke(func(c contract.Cache) {
        cache = c
    })
    
    // Basic operations
    cache.Set("key", "value", 5*time.Minute)
    
    value, err := cache.Get("key")
    if err == nil && value != nil {
        println(value.(string))
    }
}
```

### Type-safe Operations

```go
import "go.oease.dev/goe/v2/core/cache"

// Store a struct
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

user := User{ID: 1, Name: "John"}
cache.Set("user:1", user, 10*time.Minute)

// Retrieve with type safety
retrievedUser, err := cache.GetT[User](cache, "user:1")
```

### Remember Pattern

```go
// Compute expensive operation only if not cached
product, err := cache.RememberT(cache, "product:1", 1*time.Hour, func() (Product, error) {
    // This runs only on cache miss
    return fetchProductFromDB(1)
})
```

### Using Multiple Stores

```go
func setupCache(manager contract.CacheManager) {
    // Default store
    defaultCache := manager.Store()
    
    // Session store
    sessionCache := manager.Store("sessions")
    
    // Use different stores for different purposes
    defaultCache.Set("app:config", config, 24*time.Hour)
    sessionCache.Set("session:123", sessionData, 30*time.Minute)
}
```

## Adding Custom Fiber Drivers

You can extend the cache module with any Fiber storage driver:

```go
import (
    "github.com/gofiber/storage/sqlite3"
    "go.oease.dev/goe/v2/contract"
)

func registerSQLiteDriver(manager contract.CacheManager) {
    manager.Extend("sqlite", func(config contract.Config) (contract.CacheStore, error) {
        return sqlite3.New(sqlite3.Config{
            Database: config.GetString("CACHE_SQLITE_DATABASE"),
            Table:    config.GetString("CACHE_SQLITE_TABLE"),
        }), nil
    })
}
```

## Available Fiber Storage Drivers

The following Fiber storage drivers can be used with the cache module:

- **memory**: In-memory storage (included by default)
- **redis/rueidis**: Redis with auto-pipelining (included by default)
- **sqlite3**: SQLite database storage
- **postgres**: PostgreSQL storage
- **mysql**: MySQL storage
- **mongodb**: MongoDB storage
- **aerospike**: Aerospike storage
- **arangodb**: ArangoDB storage
- **azureblob**: Azure Blob storage
- **badger**: BadgerDB storage
- **bbolt**: BoltDB storage
- **cassandra**: Cassandra storage
- **couchbase**: Couchbase storage
- **dynamodb**: AWS DynamoDB storage
- **etcd**: etcd storage
- **memcache**: Memcached storage
- **minio**: MinIO/S3 storage
- **neo4j**: Neo4j storage
- **pebble**: PebbleDB storage
- **scylladb**: ScyllaDB storage
- **clickhouse**: ClickHouse storage
- **valkey**: Valkey storage
- **surrealdb**: SurrealDB storage

To use any of these drivers, install the corresponding package and register it:

```bash
go get github.com/gofiber/storage/sqlite3
```

Then configure it:

```bash
CACHE_DRIVER=sqlite3
CACHE_SQLITE_DATABASE=./cache.db
CACHE_SQLITE_TABLE=cache
```

## Performance Tips

1. **Use Redis for distributed caching**: When running multiple instances of your application
2. **Enable client-side caching**: Redis driver includes client-side caching for better performance
3. **Configure appropriate TTLs**: Set reasonable expiration times to balance memory usage and cache hit rates
4. **Use typed operations**: Generic functions avoid reflection overhead
5. **Batch operations**: Use Remember pattern to avoid multiple cache lookups

## Migration from Custom Drivers

If you were using custom cache drivers, migrate to Fiber storage drivers:

```go
// Old custom driver
store := customdriver.New(...)

// New Fiber driver
store := memory.New(memory.Config{
    GCInterval: 10 * time.Second,
})
```

The interface remains the same, so no code changes are needed beyond driver initialization.