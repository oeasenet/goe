# Cache Module

The Cache module provides a simple interface for caching data using Redis. It implements the `contracts.Cache` interface.

## Features

- Redis-based caching
- Simple key-value storage
- TTL (Time-To-Live) support
- Structured logging of cache operations

## Usage

### Initialization

The cache module is automatically initialized by the GOE framework if Redis configuration is provided:

```go
// Redis Configuration in environment variables or config file
// REDIS_HOST=localhost
// REDIS_PORT=6379
// REDIS_USERNAME=
// REDIS_PASSWORD=
```

### Basic Operations

```go
// Get the cache instance
cache := goe.UseCache()

// Set a value with TTL (in seconds)
err := cache.Set("key", "value", 60)

// Get a value
val, err := cache.Get("key")

// Delete a value
err := cache.Delete("key")

// Check if a key exists
exists, err := cache.Has("key")
```

### Working with Complex Data

The cache module can store and retrieve any data that can be serialized to JSON:

```go
type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

user := User{
    ID:   "1",
    Name: "John Doe",
    Age:  30,
}

// Store a struct
err := cache.Set("user:1", user, 3600)

// Retrieve a struct
var retrievedUser User
err := cache.Get("user:1", &retrievedUser)
```

## Implementation Details

The cache module uses Redis as the backend storage. It provides a simple interface for caching data with TTL support.

The module includes a logger that logs all cache operations for debugging purposes.