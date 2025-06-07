package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.oease.dev/goe/v2/contract"
)

// cache implements the Cache interface
type cache struct {
	store  contract.CacheStore
	prefix string
	mu     sync.RWMutex
}

// New creates a new cache instance with a given store
func New(store contract.CacheStore, prefix string) contract.Cache {
	return &cache{
		store:  store,
		prefix: prefix,
	}
}

// Get retrieves a value from cache
func (c *cache) Get(key string) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := c.store.Get(c.prefixKey(key))
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}

	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}

	return value, nil
}

// Set stores a value in cache with TTL
func (c *cache) Set(key string, value any, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.store.Set(c.prefixKey(key), data, ttl)
}

// Forever stores a value in cache forever
func (c *cache) Forever(key string, value any) error {
	return c.Set(key, value, 0)
}

// Forget removes a value from cache
func (c *cache) Forget(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.store.Delete(c.prefixKey(key))
}

// Flush removes all values from cache
func (c *cache) Flush() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.store.Reset()
}

// Has checks if a key exists in cache
func (c *cache) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := c.store.Get(c.prefixKey(key))
	return err == nil && data != nil
}

// Remember gets a value from cache or computes it
func (c *cache) Remember(key string, ttl time.Duration, callback func() (any, error)) (any, error) {
	// Try to get from cache first
	value, err := c.Get(key)
	if err == nil && value != nil {
		return value, nil
	}

	// Compute the value
	value, err = callback()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := c.Set(key, value, ttl); err != nil {
		return value, err
	}

	return value, nil
}

// RememberForever gets a value from cache or computes it forever
func (c *cache) RememberForever(key string, callback func() (any, error)) (any, error) {
	return c.Remember(key, 0, callback)
}

// Pull retrieves and removes a value from cache
func (c *cache) Pull(key string) (any, error) {
	value, err := c.Get(key)
	if err != nil {
		return nil, err
	}

	if value != nil {
		_ = c.Forget(key)
	}

	return value, nil
}

// Add stores a value only if key doesn't exist
func (c *cache) Add(key string, value any, ttl time.Duration) error {
	if c.Has(key) {
		return errors.New("key already exists")
	}

	return c.Set(key, value, ttl)
}

// Increment increments an integer value
func (c *cache) Increment(key string, value ...int64) (int64, error) {
	increment := int64(1)
	if len(value) > 0 {
		increment = value[0]
	}

	// Get current value
	val, err := c.Get(key)
	if err != nil {
		return 0, err
	}

	var current int64
	if val != nil {
		switch v := val.(type) {
		case int64:
			current = v
		case float64:
			current = int64(v)
		case int:
			current = int64(v)
		default:
			return 0, fmt.Errorf("value is not a number")
		}
	}

	// Increment
	newValue := current + increment

	// Store back
	if err := c.Set(key, newValue, 0); err != nil {
		return 0, err
	}

	return newValue, nil
}

// Decrement decrements an integer value
func (c *cache) Decrement(key string, value ...int64) (int64, error) {
	decrement := int64(1)
	if len(value) > 0 {
		decrement = value[0]
	}

	return c.Increment(key, -decrement)
}

// Store returns the underlying cache store
func (c *cache) Store() contract.CacheStore {
	return c.store
}

// GetPrefix returns the cache key prefix
func (c *cache) GetPrefix() string {
	return c.prefix
}

// prefixKey adds prefix to the key
func (c *cache) prefixKey(key string) string {
	if c.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", c.prefix, key)
}
