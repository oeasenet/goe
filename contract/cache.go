package contract

import "time"

// Cache defines the caching interface
type Cache interface {
	// Get retrieves a value from cache and binds it to the provided pointer
	// The value parameter must be a pointer to the type you want to retrieve
	// Returns nil if the key doesn't exist (no error)
	// Returns an error only if value is not a pointer or if unmarshaling fails
	Get(key string, value any) error

	// GetWithDefault retrieves a value from cache or sets a default if not found
	// The value parameter must be a pointer to the type you want to retrieve
	// If the key doesn't exist, the default value is assigned to the pointer
	// Returns an error only if value is not a pointer or if unmarshaling fails
	GetWithDefault(key string, value any, defaultValue any) error

	// Set stores a value in cache with TTL
	// A ttl of 0 uses the module's configured default TTL when one is set
	// (CACHE_TTL / cache.WithTTL), otherwise the value never expires
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
	// The value parameter must be a pointer to the type you want to retrieve
	// A ttl of 0 uses the module's configured default TTL when one is set
	// Returns an error only if value is not a pointer, callback fails, or unmarshaling fails
	Remember(key string, value any, ttl time.Duration, callback func() (any, error)) error

	// RememberForever gets a value from cache or computes it forever
	// The value parameter must be a pointer to the type you want to retrieve
	// Returns an error only if value is not a pointer, callback fails, or unmarshaling fails
	RememberForever(key string, value any, callback func() (any, error)) error

	// Pull retrieves and removes a value from cache
	// The value parameter must be a pointer to the type you want to retrieve
	// Returns nil if the key doesn't exist (no error)
	Pull(key string, value any) error

	// Add stores a value only if key doesn't exist
	// A ttl of 0 uses the module's configured default TTL when one is set
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

// CacheStoreFactory creates cache store instances. Custom backends register
// a factory through cache.WithCustomDriver and are selected with
// CACHE_DRIVER / cache.WithDriver like any builtin.
type CacheStoreFactory func(config Config) (CacheStore, error)
