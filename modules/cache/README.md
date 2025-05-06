# Cache Module

The Cache module provides a simple interface for caching data using Redis. It implements the [`contracts.Cache`](https://github.com/oeasenet/goe/blob/main/contracts/cache.go) interface and offers a high-performance caching solution for your application.

## Features

- Redis-based caching
- Simple key-value storage
- TTL (Time-To-Live) support
- JSON serialization for complex data types
- Automatic connection pooling
- Structured logging of cache operations
- Thread-safe operations

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
err := cache.Set("key", []byte("value"), 60*time.Second)
if err != nil {
    // Handle error
}

// Get a value
data := cache.Get("key")
if data == nil {
    // Key not found or error occurred
}
fmt.Println(string(data)) // "value"

// Delete a value
err := cache.Delete("key")
if err != nil {
    // Handle error
}
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
err := cache.SetBind("user:1", user, 3600*time.Second)
if err != nil {
    // Handle error
}

// Retrieve a struct
var retrievedUser User
err := cache.GetBind("user:1", &retrievedUser)
if err != nil {
    // Handle error
}
fmt.Printf("User: %s, Age: %d\n", retrievedUser.Name, retrievedUser.Age)
```

### Caching Patterns

#### Cache-Aside Pattern

```go
func GetUserWithCache(userID string) (*User, error) {
    // Try to get from cache first
    var user User
    err := cache.GetBind("user:"+userID, &user)
    if err == nil {
        return &user, nil
    }
    
    // Not in cache, get from database
    user, err = getUserFromDatabase(userID)
    if err != nil {
        return nil, err
    }
    
    // Store in cache for future requests
    err = cache.SetBind("user:"+userID, user, 30*time.Minute)
    if err != nil {
        // Log error but continue
        log.Printf("Failed to cache user: %v", err)
    }
    
    return &user, nil
}
```

#### Cache Invalidation

```go
func UpdateUser(user *User) error {
    // Update in database
    err := updateUserInDatabase(user)
    if err != nil {
        return err
    }
    
    // Invalidate cache
    err = cache.Delete("user:" + user.ID)
    if err != nil {
        // Log error but continue
        log.Printf("Failed to invalidate user cache: %v", err)
    }
    
    return nil
}
```

## API Reference

The Cache module implements the [`contracts.Cache`](https://github.com/oeasenet/goe/blob/main/contracts/cache.go) interface:

```go
type Cache interface {
    // Get retrieves a value from the cache by key
    Get(key string) []byte
    
    // GetBind retrieves a value from the cache and unmarshals it into the provided struct
    GetBind(key string, bindPtr any) error
    
    // Set stores a value in the cache with the specified expiration time
    Set(key string, value []byte, expire time.Duration) error
    
    // SetBind marshals a struct to JSON and stores it in the cache
    SetBind(key string, bindPtr any, expire time.Duration) error
    
    // Delete removes a value from the cache by key
    Delete(key string) error
}
```

## Implementation Details

The cache module uses Redis as the backend storage through the [`RedisCache`](https://github.com/oeasenet/goe/blob/main/modules/cache/cache_redis.go) implementation. It provides:

- Connection pooling for optimal performance
- JSON serialization for complex data types
- Automatic error logging
- Thread-safe operations

The module is built on top of the [Fiber Redis Storage](https://github.com/gofiber/storage/tree/main/redis) package, which provides a reliable and efficient Redis client.

### Configuration Options

The Redis cache can be configured with the following options:

- **Host**: Redis server hostname
- **Port**: Redis server port
- **Username**: Redis username (optional)
- **Password**: Redis password (optional)
- **Database**: Redis database number
- **ClientName**: Name of the Redis client
- **PoolSize**: Number of connections in the pool (defaults to 10 * GOMAXPROCS)

These options are automatically set based on the environment variables or configuration files provided to the GOE framework.