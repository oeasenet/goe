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
}

// CacheStore represents a cache storage backend
type CacheStore interface {
	// Get retrieves a value
	Get(key string) ([]byte, error)

	// Put stores a value
	Put(key string, value []byte, seconds int) error

	// Delete removes a value
	Delete(key string) error

	// Flush removes all values
	Flush() error
}
