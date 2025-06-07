package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"go.oease.dev/goe/v2/contract"
)

// TypedCache provides generic cache operations
type TypedCache[T any] struct {
	cache contract.Cache
}

// NewTyped creates a new typed cache wrapper
func NewTyped[T any](cache contract.Cache) *TypedCache[T] {
	return &TypedCache[T]{cache: cache}
}

// Get retrieves a typed value from cache
func (tc *TypedCache[T]) Get(key string) (T, error) {
	var zero T
	value, err := tc.cache.Get(key)
	if err != nil {
		return zero, err
	}
	if value == nil {
		return zero, nil
	}

	// Try direct type assertion first
	if typed, ok := value.(T); ok {
		return typed, nil
	}

	// Fall back to JSON marshaling/unmarshaling
	data, err := json.Marshal(value)
	if err != nil {
		return zero, fmt.Errorf("failed to marshal value: %w", err)
	}

	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return zero, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return result, nil
}

// Set stores a typed value in cache with TTL
func (tc *TypedCache[T]) Set(key string, value T, ttl time.Duration) error {
	return tc.cache.Set(key, value, ttl)
}

// Forever stores a typed value in cache forever
func (tc *TypedCache[T]) Forever(key string, value T) error {
	return tc.cache.Forever(key, value)
}

// Has checks if a key exists in cache
func (tc *TypedCache[T]) Has(key string) bool {
	return tc.cache.Has(key)
}

// Forget removes a value from cache
func (tc *TypedCache[T]) Forget(key string) error {
	return tc.cache.Forget(key)
}

// Remember gets a typed value from cache or computes it
func (tc *TypedCache[T]) Remember(key string, ttl time.Duration, callback func() (T, error)) (T, error) {
	value, err := tc.Get(key)
	if err == nil && !isZero(value) {
		return value, nil
	}

	// Compute the value
	value, err = callback()
	if err != nil {
		return value, err
	}

	// Store in cache
	if err := tc.Set(key, value, ttl); err != nil {
		return value, err
	}

	return value, nil
}

// RememberForever gets a typed value from cache or computes it forever
func (tc *TypedCache[T]) RememberForever(key string, callback func() (T, error)) (T, error) {
	return tc.Remember(key, 0, callback)
}

// Pull retrieves and removes a typed value from cache
func (tc *TypedCache[T]) Pull(key string) (T, error) {
	value, err := tc.Get(key)
	if err != nil {
		return value, err
	}

	if !isZero(value) {
		_ = tc.cache.Forget(key)
	}

	return value, nil
}

// Add stores a typed value only if key doesn't exist
func (tc *TypedCache[T]) Add(key string, value T, ttl time.Duration) error {
	return tc.cache.Add(key, value, ttl)
}

// GetT is a generic function to get typed values from cache
func GetT[T any](cache contract.Cache, key string) (T, error) {
	typed := NewTyped[T](cache)
	return typed.Get(key)
}

// SetT is a generic function to set typed values in cache
func SetT[T any](cache contract.Cache, key string, value T, ttl time.Duration) error {
	return cache.Set(key, value, ttl)
}

// RememberT is a generic function to remember typed values
func RememberT[T any](cache contract.Cache, key string, ttl time.Duration, callback func() (T, error)) (T, error) {
	typed := NewTyped[T](cache)
	return typed.Remember(key, ttl, callback)
}

// RememberForeverT is a generic function to remember typed values forever
func RememberForeverT[T any](cache contract.Cache, key string, callback func() (T, error)) (T, error) {
	typed := NewTyped[T](cache)
	return typed.RememberForever(key, callback)
}

// PullT is a generic function to pull typed values from cache
func PullT[T any](cache contract.Cache, key string) (T, error) {
	typed := NewTyped[T](cache)
	return typed.Pull(key)
}

// isZero checks if a value is the zero value of its type
func isZero[T any](v T) bool {
	var zero T
	return fmt.Sprintf("%v", v) == fmt.Sprintf("%v", zero)
}
