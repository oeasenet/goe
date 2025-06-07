package contract

import "time"

// Cache defines the caching interface
type Cache interface {
	// Get retrieves a value from cache
	Get(key string) (any, error)

	// Set stores a value in cache with TTL
	Set(key string, value any, ttl time.Duration) error

	// Forever stores a value in cache forever
	Forever(key string, value any) error

	// Forget removes a value from cache
	Forget(key string) error

	// Flush removes all values from cache
	Flush() error

	// Has checks if a key exists in cache
	Has(key string) bool

	// Remember gets a value from cache or computes it
	Remember(key string, ttl time.Duration, callback func() (any, error)) (any, error)

	// RememberForever gets a value from cache or computes it forever
	RememberForever(key string, callback func() (any, error)) (any, error)

	// Pull retrieves and removes a value from cache
	Pull(key string) (any, error)

	// Add stores a value only if key doesn't exist
	Add(key string, value any, ttl time.Duration) error

	// Increment increments an integer value
	Increment(key string, value ...int64) (int64, error)

	// Decrement decrements an integer value
	Decrement(key string, value ...int64) (int64, error)

	// Store returns the underlying cache store
	Store() CacheStore

	// GetPrefix returns the cache key prefix
	GetPrefix() string
}

// CacheStore represents a cache storage backend (compatible with Fiber's Storage interface)
type CacheStore interface {
	// Get gets the value for the given key.
	// `nil, nil` is returned when the key does not exist
	Get(key string) ([]byte, error)

	// Set stores the given value for the given key along
	// with an expiration value, 0 means no expiration.
	// Empty key or value will be ignored without an error.
	Set(key string, val []byte, exp time.Duration) error

	// Delete deletes the value for the given key.
	// It returns no error if the storage does not contain the key,
	Delete(key string) error

	// Reset resets the storage and delete all keys.
	Reset() error

	// Close closes the storage and will stop any running garbage
	// collectors and open connections.
	Close() error
}

// CacheManager manages multiple cache stores
type CacheManager interface {
	// Store returns a cache instance by name
	Store(name ...string) Cache

	// Driver returns the default driver name
	Driver() string

	// Extend registers a custom cache driver
	Extend(driver string, factory CacheStoreFactory)
}

// CacheStoreFactory creates cache store instances
type CacheStoreFactory func(config Config) (CacheStore, error)

// CacheConfig defines cache-specific configuration
type CacheConfig interface {
	// DefaultStore returns the default cache store name
	DefaultStore() string

	// Stores returns all configured stores
	Stores() map[string]CacheStoreConfig

	// Prefix returns the cache key prefix
	Prefix() string
}

// CacheStoreConfig defines configuration for a cache store
type CacheStoreConfig interface {
	// Driver returns the driver name (memory, redis, etc.)
	Driver() string

	// Connection returns connection parameters
	Connection() map[string]any

	// Prefix returns store-specific prefix
	Prefix() string

	// TTL returns default TTL
	TTL() time.Duration
}
