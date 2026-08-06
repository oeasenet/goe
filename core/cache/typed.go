package cache

import (
	"encoding/json"
	"time"

	"go.oease.dev/goe/v2/contract"
)

// Typed generic helpers over contract.Cache. Go interfaces cannot carry
// type-parameterized methods, so these live as package functions; they wrap
// the untyped methods and therefore keep key prefixing, the configured
// default TTL and Remember's singleflight deduplication.

// Get retrieves the value for key. It reports found=false with a zero T and
// a nil error on a cache miss, removing the untyped API's pointer dance and
// its miss-versus-stored-zero ambiguity.
func Get[T any](c contract.Cache, key string) (T, bool, error) {
	var zero T

	// The untyped Get leaves the pointer untouched and returns nil on a
	// miss, so binding through json.RawMessage separates "absent" (raw
	// stays nil) from "stored zero value" without extra round trips.
	var raw json.RawMessage
	if err := c.Get(key, &raw); err != nil {
		return zero, false, err
	}
	if raw == nil {
		return zero, false, nil
	}

	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return zero, false, err
	}
	return v, true, nil
}

// GetOr retrieves the value for key, returning fallback on a cache miss.
func GetOr[T any](c contract.Cache, key string, fallback T) (T, error) {
	v, found, err := Get[T](c, key)
	if err != nil {
		return fallback, err
	}
	if !found {
		return fallback, nil
	}
	return v, nil
}

// Remember returns the cached value for key, computing and storing it on a
// miss. Concurrent computations for the same key are deduplicated by the
// underlying Cache (singleflight), and a ttl of 0 uses the module's
// configured default TTL.
func Remember[T any](c contract.Cache, key string, ttl time.Duration, compute func() (T, error)) (T, error) {
	var v T
	err := c.Remember(key, &v, ttl, func() (any, error) { return compute() })
	return v, err
}

// Pull retrieves and removes the value for key. It reports found=false with
// a zero T and a nil error when the key does not exist.
func Pull[T any](c contract.Cache, key string) (T, bool, error) {
	var zero T

	var raw json.RawMessage
	if err := c.Pull(key, &raw); err != nil {
		return zero, false, err
	}
	if raw == nil {
		return zero, false, nil
	}

	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return zero, false, err
	}
	return v, true, nil
}
