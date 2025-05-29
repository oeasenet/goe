package contract

import (
	"context"
	"time"
)

// Cache represents the cache module interface
type Cache interface {
	Module

	// Get retrieves a value from the cache
	Get(ctx context.Context, key string) (interface{}, error)

	// GetString retrieves a string value from the cache
	GetString(ctx context.Context, key string) (string, error)

	// GetInt retrieves an int value from the cache
	GetInt(ctx context.Context, key string) (int, error)

	// GetInt64 retrieves an int64 value from the cache
	GetInt64(ctx context.Context, key string) (int64, error)

	// GetFloat64 retrieves a float64 value from the cache
	GetFloat64(ctx context.Context, key string) (float64, error)

	// GetBool retrieves a bool value from the cache
	GetBool(ctx context.Context, key string) (bool, error)

	// Set stores a value in the cache
	Set(ctx context.Context, key string, value interface{}, ttl ...time.Duration) error

	// SetWithTTL stores a value in the cache with a TTL
	SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error

	// Clear removes all values from the cache
	Clear(ctx context.Context) error

	// Has checks if a key exists in the cache
	Has(ctx context.Context, key string) (bool, error)

	// Increment increments a numeric value in the cache
	Increment(ctx context.Context, key string, value int64) (int64, error)

	// Decrement decrements a numeric value in the cache
	Decrement(ctx context.Context, key string, value int64) (int64, error)

	// Remember gets a value from the cache or stores the result of the callback
	Remember(ctx context.Context, key string, ttl time.Duration, callback func() (interface{}, error)) (interface{}, error)

	// Forever stores a value in the cache indefinitely
	Forever(ctx context.Context, key string, value interface{}) error

	// Forget removes a value from the cache (alias for Delete)
	Forget(ctx context.Context, key string) error

	// Pull retrieves a value from the cache and then removes it
	Pull(ctx context.Context, key string) (interface{}, error)

	// Tags returns a tagged cache instance
	Tags(tags ...string) TaggedCache
}

// TaggedCache represents a tagged cache instance
type TaggedCache interface {
	// Get retrieves a value from the cache
	Get(ctx context.Context, key string) (interface{}, error)

	// GetString retrieves a string value from the cache
	GetString(ctx context.Context, key string) (string, error)

	// GetInt retrieves an int value from the cache
	GetInt(ctx context.Context, key string) (int, error)

	// GetInt64 retrieves an int64 value from the cache
	GetInt64(ctx context.Context, key string) (int64, error)

	// GetFloat64 retrieves a float64 value from the cache
	GetFloat64(ctx context.Context, key string) (float64, error)

	// GetBool retrieves a bool value from the cache
	GetBool(ctx context.Context, key string) (bool, error)

	// Set stores a value in the cache
	Set(ctx context.Context, key string, value interface{}, ttl ...time.Duration) error

	// SetWithTTL stores a value in the cache with a TTL
	SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error

	// Clear removes all values from the cache with the given tags
	Clear(ctx context.Context) error

	// Has checks if a key exists in the cache
	Has(ctx context.Context, key string) (bool, error)

	// Increment increments a numeric value in the cache
	Increment(ctx context.Context, key string, value int64) (int64, error)

	// Decrement decrements a numeric value in the cache
	Decrement(ctx context.Context, key string, value int64) (int64, error)

	// Remember gets a value from the cache or stores the result of the callback
	Remember(ctx context.Context, key string, ttl time.Duration, callback func() (interface{}, error)) (interface{}, error)

	// Forever stores a value in the cache indefinitely
	Forever(ctx context.Context, key string, value interface{}) error

	// Forget removes a value from the cache (alias for Delete)
	Forget(ctx context.Context, key string) error

	// Pull retrieves a value from the cache and then removes it
	Pull(ctx context.Context, key string) (interface{}, error)
}

// CacheProvider is a function that provides a Cache instance
type CacheProvider func() Cache
