# 9. Caching Strategies ⚡

Caching is a vital technique for improving application performance and reducing latency by storing frequently accessed
data in a fast, temporary storage. Goe's cache module provides a unified interface built
upon [Fiber's Storage package](https://docs.gofiber.io/storage), allowing easy integration with various caching
backends.

## Enabling the Cache Module

To use caching capabilities, enable the `Cache` module during application initialization:

```go
package main

import (
	"go.oease.dev/goe/v2"
)

func main() {
	app := goe.New(goe.Options{
		WithCache: true, // Enable the Cache module
		// ... other options
	})
	goe.Run()
}
```

## Configuration

Cache behavior is configured via environment variables, typically prefixed with `CACHE_`.

* **`CACHE_STORE`**: Specifies the default cache store name to use if you have multiple stores defined. Default:
  `"default"`.
* **`CACHE_DRIVER`**: The driver for the default cache store (or the store named by `CACHE_STORE` if it's "default" or
  not specifically configured with its own driver).
    * Supported drivers (via Fiber's Storage): `memory`, `redis`, `sqlite3`, `postgres`, `mysql`, `mongo`, `fs` (
      filesystem), etc. Check Fiber's documentation for the full list and their specific requirements.
    * Default: `memory`.
* **`CACHE_PREFIX`**: A global prefix added to all cache keys. If not set, it defaults to the `APP_NAME`. This helps
  avoid key collisions if multiple applications share the same cache backend.
    * Example: `CACHE_PREFIX=mygoeapp`

**Store-Specific Configuration:**

You can configure individual cache stores by prefixing keys with `CACHE_<STORE_NAME>_`. For example, for a store named
`redis_cache`:

* `CACHE_REDIS_CACHE_DRIVER=redis`
* `CACHE_REDIS_CACHE_PREFIX=my_redis_specific_prefix`
* `CACHE_REDIS_CACHE_TTL=1h` (Default TTL for this store)

**Driver-Specific Connection Parameters:**

Connection parameters for drivers like Redis are also set via environment variables, usually following a pattern like
`CACHE_REDIS_HOST`, `CACHE_REDIS_PORT`, `CACHE_REDIS_PASSWORD`, `CACHE_REDIS_DB`. The `core/cache/manager.go` attempts
to pick these up based on the store name or generic `CACHE_` prefixes.

* Example for a default Redis setup:
    * `CACHE_DRIVER=redis`
    * `CACHE_REDIS_HOSTS=localhost:6379`
    * `CACHE_REDIS_PASSWORD=`
    * `CACHE_REDIS_DB=0`

* **`CACHE_TTL`**: A default Time-To-Live for cache items if not specified during a `Set` operation for the default
  store. Example: `CACHE_TTL=30m`. Defaults to 2 hours if not set at any level.

The cache manager (`core/cache/manager.go`) resolves the configuration for a store by first looking for
`CACHE_<STORE_NAME>_DRIVER`, then `CACHE_DRIVER`. Connection parameters are collected from `CACHE_<STORE_NAME>_*` and
then from general `CACHE_*` keys if not found for the specific store.

## Accessing the Cache

### 1. Default Cache Store

* **Global Accessor**: `goe.Cache() contract.Cache`
* **Dependency Injection**: Inject `contract.Cache`

```go
import (
"fmt"
"time"
"go.oease.dev/goe/v2"
"go.oease.dev/goe/v2/contract"
)

// Global access
func GetSomethingFromCacheGlobally(key string) (any, error) {
if !goe.IsModuleEnabled(goe.ModuleCache) {
return nil, fmt.Errorf("cache module not enabled")
}
return goe.Cache().Get(key)
}

// Dependency Injection
type MyDataService struct {
cache  contract.Cache
logger contract.Logger
}

func NewMyDataService(cache contract.Cache, logger contract.Logger) *MyDataService {
return &MyDataService{cache: cache, logger: logger}
}

func (s *MyDataService) GetExpensiveData(id string) (string, error) {
cachedResult, err := s.cache.Get(id)
if err == nil && cachedResult != nil {
s.logger.Info("Cache hit!", contract.NewField("key", id))
return cachedResult.(string), nil // Type assertion needed for `any`
}

s.logger.Info("Cache miss. Fetching data...", contract.NewField("key", id))
// data := fetchFromDatabase(id) // Expensive operation
data := fmt.Sprintf("data_for_%s", id) // Simulate

// Store in cache for 10 minutes
err = s.cache.Set(id, data, 10*time.Minute)
if err != nil {
s.logger.Warn("Failed to set cache", contract.NewField("key", id), contract.NewField("error", err))
}
return data, nil
}
```

### 2. Named Cache Stores (Cache Manager)

If you have multiple cache stores configured (e.g., one for sessions, another for application data), you can use the
`contract.CacheManager`.

* **Global Accessor**: `goe.CacheManager() contract.CacheManager`
* **Dependency Injection**: Inject `contract.CacheManager`

```go
import (
"go.oease.dev/goe/v2"
"go.oease.dev/goe/v2/contract"
)

// Global access to Cache Manager
func GetFromSpecificStore(storeName, key string) (any, error) {
if !goe.IsModuleEnabled(goe.ModuleCache) {
return nil, fmt.Errorf("cache module not enabled")
}
cacheStore := goe.CacheManager().Store(storeName) // Returns contract.Cache
return cacheStore.Get(key)
}

// Dependency Injection of Cache Manager
type SessionService struct {
sessionCache contract.Cache
logger contract.Logger
}

func NewSessionService(cm contract.CacheManager, logger contract.Logger) *SessionService {
// Assuming "sessions" store is configured
sessionCacheInstance := cm.Store("sessions")
return &SessionService{sessionCache: sessionCacheInstance, logger: logger}
}

func (s *SessionService) GetSessionData(sessionID string) (map[string]any, error) {
data, err := s.sessionCache.Get(sessionID)
if err != nil {
return nil, err
}
if data == nil {
return nil, nil // Session not found
}
return data.(map[string]any), nil // Type assertion
}
```

If `Store()` is called with no arguments or an empty string, it returns the default cache store.

## Common Cache Operations

The `contract.Cache` interface provides a rich set of methods:

* **`Get(key string) (any, error)`**: Retrieves an item from the cache. Returns `nil, nil` if the item doesn't exist.
  The value is `any`, so type assertion is often needed.
* **`Set(key string, value any, ttl time.Duration) error`**: Stores an item in the cache for a specified duration (TTL).
  If `ttl` is `0`, it's stored "forever" (or per backend's interpretation). Values are typically JSON marshaled.
* **`Forever(key string, value any) error`**: Stores an item in the cache indefinitely. Equivalent to
  `Set(key, value, 0)`.
* **`Forget(key string) error`**: Removes an item from the cache.
* **`Flush() error`**: Removes all items from the cache store. Use with caution!
* **`Has(key string) bool`**: Checks if an item exists in the cache.
* **`Remember(key string, ttl time.Duration, callback func() (any, error)) (any, error)`**:
  Retrieves an item from the cache. If it doesn't exist, the `callback` function is executed, its result is stored in
  the cache for `ttl`, and then returned. This is useful for "cache-aside" patterns.
* **`RememberForever(key string, callback func() (any, error)) (any, error)`**: Same as `Remember`, but stores the item
  indefinitely.
* **`Pull(key string) (any, error)`**: Retrieves an item from thecache and then deletes it. Returns `nil, nil` if not
  found.
* **`Add(key string, value any, ttl time.Duration) error`**: Stores an item only if the key doesn't already exist.
  Returns an error if the key exists.
* **`Increment(key string, value ...int64) (int64, error)`**: Atomically increments an integer item in the cache. If the
  key doesn't exist, it's typically initialized to `0` before incrementing. The optional `value` argument specifies the
  increment amount (defaults to `1`).
* **`Decrement(key string, value ...int64) (int64, error)`**: Atomically decrements an integer item.

**Example using `Remember`:**

```go
func GetUserDetails(userID string, cacheSvc contract.Cache) (string, error) {
key := "user_details:" + userID

userDetails, err := cacheSvc.Remember(key, 15*time.Minute, func () (any, error) {
// This function is called only if 'key' is not in cache
// goe.Log().Info("Fetching user details from DB for Remember", contract.NewField("user_id", userID))
// actualDetails := fetchUserDetailsFromDB(userID)
actualDetails := fmt.Sprintf("DB details for %s", userID) // Simulate
return actualDetails, nil
})

if err != nil {
return "", err
}
return userDetails.(string), nil
}
```

## Typed Cache Operations (Generics)

Goe provides generic helper functions and a `TypedCache[T]` wrapper in `core/cache/typed.go` for more type-safe
interactions, reducing the need for manual type assertions.

```go
import (
"go.oease.dev/goe/v2/contract"
goecache "go.oease.dev/goe/v2/core/cache" // For typed helpers
"time"
)

type UserProfile struct {
Name  string `json:"name"`
Email string `json:"email"`
}

func GetCachedUserProfile(userID string, C contract.Cache) (*UserProfile, error) {
key := "user_profile:" + userID

// Using the generic GetT helper
profile, err := goecache.GetT[*UserProfile](C, key)
if err != nil {
return nil, err
}
// 'profile' is already of type *UserProfile, or nil if not found/error

if profile == nil { // Cache miss
// fetchedProfile := fetchFromDBAndModel(userID) // Simulate
fetchedProfile := &UserProfile{Name: "Jane Doe", Email: "jane@example.com"}

// Using generic SetT helper
errSet := goecache.SetT(C, key, fetchedProfile, 1*time.Hour)
if errSet != nil {
// Log error, but can still return fetchedProfile
}
return fetchedProfile, nil
}
return profile, nil
}

// Example with RememberT
func GetOrFetchSettings(settingsKey string, C contract.Cache) (map[string]string, error) {
return goecache.RememberT[map[string]string](C, settingsKey, 5*time.Minute, func () (map[string]string, error) {
// goe.Log().Info("Fetching settings from source for RememberT", contract.NewField("key", settingsKey))
// actualSettings := loadFromSource() // Simulate
actualSettings := map[string]string{"theme": "dark", "language": "en"}
return actualSettings, nil
})
}
```

Available typed helpers in `go.oease.dev/goe/v2/core/cache`:

* `GetT[T any](cache contract.Cache, key string) (T, error)`
* `SetT[T any](cache contract.Cache, key string, value T, ttl time.Duration) error`
* `RememberT[T any](cache contract.Cache, key string, ttl time.Duration, callback func() (T, error)) (T, error)`
* `RememberForeverT[T any](cache contract.Cache, key string, callback func() (T, error)) (T, error)`
* `PullT[T any](cache contract.Cache, key string) (T, error)`
* You can also wrap a `contract.Cache` with `NewTyped[T](cache)` to get a `*TypedCache[T]` instance with more methods.

## Cache Key Prefixes

Goe automatically prefixes cache keys.

* The global prefix is taken from `CACHE_PREFIX` or defaults to `APP_NAME`.
* Each store can also have its own prefix defined by `CACHE_<STORE_NAME>_PREFIX`, which overrides the global prefix for
  that store.
* The final key stored in the backend will be `resolved_prefix:your_key`.
  This helps in namespacing keys, especially if the same cache backend (e.g., a Redis instance) is shared by multiple
  applications or environments.

## Extending with Custom Drivers

The cache manager (`contract.CacheManager`) allows you to register custom cache store drivers using the `Extend` method
if Fiber's built-in drivers don't meet your needs. This is an advanced use case.

```go
// Example: Registering a hypothetical custom driver
// var myCustomDriverFactory contract.CacheStoreFactory = func(cfg contract.Config) (contract.CacheStore, error) {
//     // ... logic to create and return an instance of your custom store ...
//     // Your custom store must implement contract.CacheStore
//     return newMyCustomStore(cfg.GetString("MY_CUSTOM_DSN")), nil
// }
//
// // In an Fx invoker or early in app setup:
// func registerCustomCache(manager contract.CacheManager) {
//    manager.Extend("my_driver_name", myCustomDriverFactory)
//    // Now you can use "my_driver_name" in CACHE_DRIVER or CACHE_MYSTORE_DRIVER
// }
```

## Best Practices for Caching

* **Cache What's Expensive**: Cache data that is computationally expensive to generate or slow to retrieve (e.g.,
  complex database queries, external API calls).
* **Appropriate TTLs**: Choose sensible Time-To-Live (TTL) values. Too short, and the cache is ineffective; too long,
  and you risk serving stale data.
* **Cache Invalidation**: Develop a strategy for cache invalidation if data can change before its TTL expires. This can
  be complex (e.g., event-driven invalidation, explicit deletion on update).
* **Data Serialization**: Be mindful of data serialization. Goe's cache typically uses JSON. Ensure your cached types
  can be correctly marshaled and unmarshaled.
* **Key Naming Conventions**: Use clear and consistent cache key naming conventions. Include identifiers that make keys
  unique and understandable (e.g., `user:details:123`, `products:featured`).
* **Avoid Caching Large Objects**: Storing very large objects in cache can strain cache memory and impact performance.
  Cache only what's necessary.
* **Error Handling**: Operations like `Get` or `Set` can fail (e.g., network issues with Redis). Handle these errors
  gracefully; often, it means falling back to the original data source.
* **Cache Warming**: For critical, frequently accessed data, consider "warming" the cache (populating it proactively)
  when the application starts or during off-peak hours.
* **Monitoring**: If using a distributed cache like Redis, monitor its performance and memory usage.

Goe's cache module provides a flexible and powerful way to integrate caching into your applications, significantly
boosting performance when used effectively.

Next, we'll explore Goe's [Module System](10-modules.md) in more detail.
