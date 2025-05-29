package cache

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.oease.dev/goe/v2/contract"
)

// Cache implements the contract.Cache interface
type Cache struct {
	mu      sync.RWMutex
	items   map[string]cacheItem
	tags    map[string]map[string]struct{}
	tagKeys map[string]map[string]struct{}
}

// cacheItem represents a cached item
type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// New creates a new Cache instance
func New() *Cache {
	return &Cache{
		items:   make(map[string]cacheItem),
		tags:    make(map[string]map[string]struct{}),
		tagKeys: make(map[string]map[string]struct{}),
	}
}

// Name returns the name of the module
func (c *Cache) Name() string {
	return "cache"
}

// Initialize initializes the cache module
func (c *Cache) Initialize(ctx context.Context) error {
	return nil
}

// Start starts the cache module
func (c *Cache) Start(ctx context.Context) error {
	// Start a goroutine to clean up expired items
	go c.cleanupLoop(ctx)
	return nil
}

// Stop stops the cache module
func (c *Cache) Stop(ctx context.Context) error {
	return nil
}

// Get retrieves a value from the cache
func (c *Cache) Get(ctx context.Context, key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return nil, errors.New("key not found")
	}

	// Check if the item has expired
	if !item.expiration.IsZero() && item.expiration.Before(time.Now()) {
		return nil, errors.New("key expired")
	}

	return item.value, nil
}

// GetString retrieves a string value from the cache
func (c *Cache) GetString(ctx context.Context, key string) (string, error) {
	value, err := c.Get(ctx, key)
	if err != nil {
		return "", err
	}

	str, ok := value.(string)
	if !ok {
		return "", errors.New("value is not a string")
	}

	return str, nil
}

// GetInt retrieves an int value from the cache
func (c *Cache) GetInt(ctx context.Context, key string) (int, error) {
	value, err := c.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	intVal, ok := value.(int)
	if !ok {
		return 0, errors.New("value is not an int")
	}

	return intVal, nil
}

// GetInt64 retrieves an int64 value from the cache
func (c *Cache) GetInt64(ctx context.Context, key string) (int64, error) {
	value, err := c.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	int64Val, ok := value.(int64)
	if !ok {
		return 0, errors.New("value is not an int64")
	}

	return int64Val, nil
}

// GetFloat64 retrieves a float64 value from the cache
func (c *Cache) GetFloat64(ctx context.Context, key string) (float64, error) {
	value, err := c.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	float64Val, ok := value.(float64)
	if !ok {
		return 0, errors.New("value is not a float64")
	}

	return float64Val, nil
}

// GetBool retrieves a bool value from the cache
func (c *Cache) GetBool(ctx context.Context, key string) (bool, error) {
	value, err := c.Get(ctx, key)
	if err != nil {
		return false, err
	}

	boolVal, ok := value.(bool)
	if !ok {
		return false, errors.New("value is not a bool")
	}

	return boolVal, nil
}

// Set stores a value in the cache
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl ...time.Duration) error {
	var expiration time.Time
	if len(ttl) > 0 && ttl[0] > 0 {
		expiration = time.Now().Add(ttl[0])
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheItem{
		value:      value,
		expiration: expiration,
	}

	return nil
}

// SetWithTTL stores a value in the cache with a TTL
func (c *Cache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.Set(ctx, key, value, ttl)
}

// Delete removes a value from the cache
func (c *Cache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)

	// Remove key from tags
	for _, keys := range c.tags {
		delete(keys, key)
	}

	// Remove key from tagKeys
	if tags, ok := c.tagKeys[key]; ok {
		for tag := range tags {
			if keys, ok := c.tags[tag]; ok {
				delete(keys, key)
			}
		}
		delete(c.tagKeys, key)
	}

	return nil
}

// Clear removes all values from the cache
func (c *Cache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]cacheItem)
	c.tags = make(map[string]map[string]struct{})
	c.tagKeys = make(map[string]map[string]struct{})

	return nil
}

// Has checks if a key exists in the cache
func (c *Cache) Has(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return false, nil
	}

	// Check if the item has expired
	if !item.expiration.IsZero() && item.expiration.Before(time.Now()) {
		return false, nil
	}

	return true, nil
}

// Increment increments a numeric value in the cache
func (c *Cache) Increment(ctx context.Context, key string, value int64) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[key]
	if !ok {
		// If the key doesn't exist, create it with the value
		c.items[key] = cacheItem{
			value:      value,
			expiration: time.Time{},
		}
		return value, nil
	}

	// Check if the item has expired
	if !item.expiration.IsZero() && item.expiration.Before(time.Now()) {
		// If the item has expired, create it with the value
		c.items[key] = cacheItem{
			value:      value,
			expiration: time.Time{},
		}
		return value, nil
	}

	// Try to increment the value
	switch v := item.value.(type) {
	case int:
		newValue := int64(v) + value
		c.items[key] = cacheItem{
			value:      newValue,
			expiration: item.expiration,
		}
		return newValue, nil
	case int64:
		newValue := v + value
		c.items[key] = cacheItem{
			value:      newValue,
			expiration: item.expiration,
		}
		return newValue, nil
	case float64:
		newValue := int64(v) + value
		c.items[key] = cacheItem{
			value:      newValue,
			expiration: item.expiration,
		}
		return newValue, nil
	default:
		return 0, errors.New("value is not numeric")
	}
}

// Decrement decrements a numeric value in the cache
func (c *Cache) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return c.Increment(ctx, key, -value)
}

// Remember gets a value from the cache or stores the result of the callback
func (c *Cache) Remember(ctx context.Context, key string, ttl time.Duration, callback func() (interface{}, error)) (interface{}, error) {
	// Check if the key exists
	if value, err := c.Get(ctx, key); err == nil {
		return value, nil
	}

	// Call the callback
	value, err := callback()
	if err != nil {
		return nil, err
	}

	// Store the value
	if err := c.Set(ctx, key, value, ttl); err != nil {
		return nil, err
	}

	return value, nil
}

// Forever stores a value in the cache indefinitely
func (c *Cache) Forever(ctx context.Context, key string, value interface{}) error {
	return c.Set(ctx, key, value)
}

// Forget removes a value from the cache (alias for Delete)
func (c *Cache) Forget(ctx context.Context, key string) error {
	return c.Delete(ctx, key)
}

// Pull retrieves a value from the cache and then removes it
func (c *Cache) Pull(ctx context.Context, key string) (interface{}, error) {
	value, err := c.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if err := c.Delete(ctx, key); err != nil {
		return nil, err
	}

	return value, nil
}

// Tags returns a tagged cache instance
func (c *Cache) Tags(tags ...string) contract.TaggedCache {
	return &TaggedCache{
		parent: c,
		tags:   tags,
	}
}

// cleanupLoop periodically cleans up expired items
func (c *Cache) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.cleanup()
		}
	}
}

// cleanup removes expired items
func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if !item.expiration.IsZero() && item.expiration.Before(now) {
			delete(c.items, key)

			// Remove key from tags
			for _, keys := range c.tags {
				delete(keys, key)
			}

			// Remove key from tagKeys
			if tags, ok := c.tagKeys[key]; ok {
				for tag := range tags {
					if keys, ok := c.tags[tag]; ok {
						delete(keys, key)
					}
				}
				delete(c.tagKeys, key)
			}
		}
	}
}

// TaggedCache implements the contract.TaggedCache interface
type TaggedCache struct {
	parent *Cache
	tags   []string
}

// Get retrieves a value from the cache
func (t *TaggedCache) Get(ctx context.Context, key string) (interface{}, error) {
	taggedKey := t.taggedKey(key)
	return t.parent.Get(ctx, taggedKey)
}

// GetString retrieves a string value from the cache
func (t *TaggedCache) GetString(ctx context.Context, key string) (string, error) {
	taggedKey := t.taggedKey(key)
	return t.parent.GetString(ctx, taggedKey)
}

// GetInt retrieves an int value from the cache
func (t *TaggedCache) GetInt(ctx context.Context, key string) (int, error) {
	taggedKey := t.taggedKey(key)
	return t.parent.GetInt(ctx, taggedKey)
}

// GetInt64 retrieves an int64 value from the cache
func (t *TaggedCache) GetInt64(ctx context.Context, key string) (int64, error) {
	taggedKey := t.taggedKey(key)
	return t.parent.GetInt64(ctx, taggedKey)
}

// GetFloat64 retrieves a float64 value from the cache
func (t *TaggedCache) GetFloat64(ctx context.Context, key string) (float64, error) {
	taggedKey := t.taggedKey(key)
	return t.parent.GetFloat64(ctx, taggedKey)
}

// GetBool retrieves a bool value from the cache
func (t *TaggedCache) GetBool(ctx context.Context, key string) (bool, error) {
	taggedKey := t.taggedKey(key)
	return t.parent.GetBool(ctx, taggedKey)
}

// Set stores a value in the cache
func (t *TaggedCache) Set(ctx context.Context, key string, value interface{}, ttl ...time.Duration) error {
	taggedKey := t.taggedKey(key)

	// Store the value
	if err := t.parent.Set(ctx, taggedKey, value, ttl...); err != nil {
		return err
	}

	// Associate the key with the tags
	t.parent.mu.Lock()
	defer t.parent.mu.Unlock()

	for _, tag := range t.tags {
		if _, ok := t.parent.tags[tag]; !ok {
			t.parent.tags[tag] = make(map[string]struct{})
		}
		t.parent.tags[tag][taggedKey] = struct{}{}

		if _, ok := t.parent.tagKeys[taggedKey]; !ok {
			t.parent.tagKeys[taggedKey] = make(map[string]struct{})
		}
		t.parent.tagKeys[taggedKey][tag] = struct{}{}
	}

	return nil
}

// SetWithTTL stores a value in the cache with a TTL
func (t *TaggedCache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return t.Set(ctx, key, value, ttl)
}

// Delete removes a value from the cache
func (t *TaggedCache) Delete(ctx context.Context, key string) error {
	taggedKey := t.taggedKey(key)
	return t.parent.Delete(ctx, taggedKey)
}

// Clear removes all values from the cache with the given tags
func (t *TaggedCache) Clear(ctx context.Context) error {
	t.parent.mu.Lock()
	defer t.parent.mu.Unlock()

	// Remove all keys associated with the tags
	for _, tag := range t.tags {
		if keys, ok := t.parent.tags[tag]; ok {
			for key := range keys {
				delete(t.parent.items, key)

				// Remove key from tagKeys
				if tags, ok := t.parent.tagKeys[key]; ok {
					for tag := range tags {
						if keys, ok := t.parent.tags[tag]; ok {
							delete(keys, key)
						}
					}
					delete(t.parent.tagKeys, key)
				}
			}
			delete(t.parent.tags, tag)
		}
	}

	return nil
}

// Has checks if a key exists in the cache
func (t *TaggedCache) Has(ctx context.Context, key string) (bool, error) {
	taggedKey := t.taggedKey(key)
	return t.parent.Has(ctx, taggedKey)
}

// Increment increments a numeric value in the cache
func (t *TaggedCache) Increment(ctx context.Context, key string, value int64) (int64, error) {
	taggedKey := t.taggedKey(key)
	return t.parent.Increment(ctx, taggedKey, value)
}

// Decrement decrements a numeric value in the cache
func (t *TaggedCache) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	taggedKey := t.taggedKey(key)
	return t.parent.Decrement(ctx, taggedKey, value)
}

// Remember gets a value from the cache or stores the result of the callback
func (t *TaggedCache) Remember(ctx context.Context, key string, ttl time.Duration, callback func() (interface{}, error)) (interface{}, error) {
	taggedKey := t.taggedKey(key)

	// Check if the key exists
	if value, err := t.parent.Get(ctx, taggedKey); err == nil {
		return value, nil
	}

	// Call the callback
	value, err := callback()
	if err != nil {
		return nil, err
	}

	// Store the value
	if err := t.Set(ctx, key, value, ttl); err != nil {
		return nil, err
	}

	return value, nil
}

// Forever stores a value in the cache indefinitely
func (t *TaggedCache) Forever(ctx context.Context, key string, value interface{}) error {
	return t.Set(ctx, key, value)
}

// Forget removes a value from the cache (alias for Delete)
func (t *TaggedCache) Forget(ctx context.Context, key string) error {
	return t.Delete(ctx, key)
}

// Pull retrieves a value from the cache and then removes it
func (t *TaggedCache) Pull(ctx context.Context, key string) (interface{}, error) {
	taggedKey := t.taggedKey(key)
	return t.parent.Pull(ctx, taggedKey)
}

// taggedKey generates a key with the tags
func (t *TaggedCache) taggedKey(key string) string {
	// Simple implementation for now
	return "tags:" + key
}

// Provider provides a Cache instance
func Provider() contract.Cache {
	return New()
}
