package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
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

// Get retrieves a value from cache and binds it to the provided pointer
func (c *cache) Get(key string, value any) error {
	// Validate that value is a pointer
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("value must be a non-nil pointer")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := c.store.Get(c.prefixKey(key))
	if err != nil {
		return err
	}
	if data == nil {
		// Cache miss is not an error, just return nil
		return nil
	}

	return json.Unmarshal(data, value)
}

// GetWithDefault retrieves a value from cache or sets a default if not found
func (c *cache) GetWithDefault(key string, value any, defaultValue any) error {
	// Validate that value is a pointer
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("value must be a non-nil pointer")
	}

	c.mu.RLock()
	data, err := c.store.Get(c.prefixKey(key))
	c.mu.RUnlock()

	if err != nil {
		return err
	}
	if data == nil {
		// Cache miss - use the default value
		rdv := reflect.ValueOf(defaultValue)
		rv.Elem().Set(rdv)
		return nil
	}

	return json.Unmarshal(data, value)
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
func (c *cache) Remember(key string, value any, ttl time.Duration, callback func() (any, error)) error {
	// Validate that value is a pointer
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("value must be a non-nil pointer")
	}

	c.mu.RLock()
	data, err := c.store.Get(c.prefixKey(key))
	c.mu.RUnlock()

	if err != nil {
		return err
	}
	if data != nil {
		// Cache hit - unmarshal and return
		return json.Unmarshal(data, value)
	}

	// Cache miss - compute the value
	computedValue, err := callback()
	if err != nil {
		return err
	}

	// Store in cache
	if err := c.Set(key, computedValue, ttl); err != nil {
		return err
	}

	// Unmarshal the computed value into the provided pointer
	computedData, err := json.Marshal(computedValue)
	if err != nil {
		return err
	}

	return json.Unmarshal(computedData, value)
}

// RememberForever gets a value from cache or computes it forever
func (c *cache) RememberForever(key string, value any, callback func() (any, error)) error {
	return c.Remember(key, value, 0, callback)
}

// Pull retrieves and removes a value from cache
func (c *cache) Pull(key string, value any) error {
	// Validate that value is a pointer
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("value must be a non-nil pointer")
	}

	c.mu.RLock()
	data, err := c.store.Get(c.prefixKey(key))
	c.mu.RUnlock()

	if err != nil {
		return err
	}
	if data == nil {
		// Cache miss is not an error, just return nil
		return nil
	}

	// Unmarshal the value
	err = json.Unmarshal(data, value)
	if err != nil {
		return err
	}

	// Remove the key after successful retrieval
	_ = c.Forget(key)

	return nil
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
	c.mu.RLock()
	data, err := c.store.Get(c.prefixKey(key))
	c.mu.RUnlock()

	if err != nil {
		return 0, err
	}

	var current int64
	if data == nil {
		// Key doesn't exist, start from 0
		current = 0
	} else {
		// Try to unmarshal as different numeric types
		if err := json.Unmarshal(data, &current); err != nil {
			// Try as float64
			var floatVal float64
			if err2 := json.Unmarshal(data, &floatVal); err2 == nil {
				current = int64(floatVal)
			} else {
				// Try as int
				var intVal int
				if err3 := json.Unmarshal(data, &intVal); err3 == nil {
					current = int64(intVal)
				} else {
					return 0, fmt.Errorf("value is not a number")
				}
			}
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
