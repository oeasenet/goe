//go:build integration

package cache

import (
	"context"
	"sync"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/config"
)

func TestRedisStoreFactory(t *testing.T) {
	// This test only verifies that the configuration is properly mapped
	// Actual connection tests are in TestRedisStoreIntegration

	tests := []struct {
		name      string
		setupFunc func(*config.Module)
	}{
		{
			name: "default configuration",
			setupFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_REDIS_HOST", "localhost")
				cfg.Provide().Set("CACHE_REDIS_PORT", 6379)
			},
		},
		{
			name: "with URL configuration",
			setupFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_REDIS_URL", "redis://localhost:6379/0")
			},
		},
		{
			name: "with database selection",
			setupFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_REDIS_HOST", "localhost")
				cfg.Provide().Set("CACHE_REDIS_PORT", 6379)
				cfg.Provide().Set("CACHE_REDIS_DATABASE", 1)
			},
		},
		{
			name: "with client name",
			setupFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_REDIS_HOST", "localhost")
				cfg.Provide().Set("CACHE_REDIS_PORT", 6379)
				cfg.Provide().Set("CACHE_REDIS_CLIENT_NAME", "test-client")
			},
		},
		{
			name: "with pool size",
			setupFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_REDIS_HOST", "localhost")
				cfg.Provide().Set("CACHE_REDIS_PORT", 6379)
				cfg.Provide().Set("CACHE_REDIS_POOL_SIZE", 20)
			},
		},
		{
			name: "with reset flag",
			setupFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_REDIS_HOST", "localhost")
				cfg.Provide().Set("CACHE_REDIS_PORT", 6379)
				cfg.Provide().Set("CACHE_REDIS_RESET", true)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config module
			cfg := config.NewModule()
			tt.setupFunc(cfg)

			// Create Redis store - this will actually try to connect
			store, err := RedisStoreFactory(cfg.Provide())

			// We expect these to work since we have Redis running
			assert.NoError(t, err)
			assert.NotNil(t, store)
		})
	}
}

func TestRedisStoreIntegration(t *testing.T) {

	// Create config for Redis
	cfg := config.NewModule()
	cfg.Provide().Set("CACHE_REDIS_HOST", "localhost")
	cfg.Provide().Set("CACHE_REDIS_PORT", 6379)
	cfg.Provide().Set("CACHE_REDIS_DATABASE", 15) // Use database 15 for testing

	// Create Redis store
	store, err := RedisStoreFactory(cfg.Provide())
	require.NoError(t, err)
	require.NotNil(t, store)

	// Clear any existing data
	err = store.Reset()
	assert.NoError(t, err)

	t.Run("basic operations", func(t *testing.T) {
		key := "test:key"
		value := []byte("test value")

		// Set a value
		err := store.Set(key, value, 10*time.Second)
		assert.NoError(t, err)

		// Get the value
		retrieved, err := store.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, value, retrieved)

		// Delete the value
		err = store.Delete(key)
		assert.NoError(t, err)

		// Verify it's deleted
		retrieved, err = store.Get(key)
		assert.NoError(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("expiration", func(t *testing.T) {
		key := "test:expiring"
		value := []byte("expiring value")

		// Set with short TTL
		err := store.Set(key, value, 100*time.Millisecond)
		assert.NoError(t, err)

		// Should exist immediately
		retrieved, err := store.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, value, retrieved)

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Should be gone
		retrieved, err = store.Get(key)
		assert.NoError(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("reset", func(t *testing.T) {
		// Set multiple values
		for i := 0; i < 5; i++ {
			key := "test:reset:" + string(rune('a'+i))
			err := store.Set(key, []byte("value"), 10*time.Second)
			assert.NoError(t, err)
		}

		// Reset should clear all
		err := store.Reset()
		assert.NoError(t, err)

		// Verify all are gone
		for i := 0; i < 5; i++ {
			key := "test:reset:" + string(rune('a'+i))
			retrieved, err := store.Get(key)
			assert.NoError(t, err)
			assert.Nil(t, retrieved)
		}
	})
}

// newAtomicTestStore builds a redis store on the test database and resets it.
func newAtomicTestStore(t *testing.T) contract.CacheStore {
	t.Helper()

	cfg := config.NewModule()
	cfg.Provide().Set("CACHE_REDIS_HOST", "localhost")
	cfg.Provide().Set("CACHE_REDIS_PORT", 6379)
	cfg.Provide().Set("CACHE_REDIS_DATABASE", 15)

	store, err := RedisStoreFactory(cfg.Provide())
	require.NoError(t, err)
	require.NoError(t, store.Reset())
	return store
}

// newVerifyClient returns an independent go-redis client so TTL and value
// assertions run against Redis itself rather than through the code under test.
func newVerifyClient(t *testing.T) *goredis.Client {
	t.Helper()

	verify := goredis.NewClient(&goredis.Options{Addr: "localhost:6379", DB: 15})
	t.Cleanup(func() { _ = verify.Close() })
	return verify
}

func TestRedisStoreAtomicIntegration(t *testing.T) {
	store := newAtomicTestStore(t)
	verify := newVerifyClient(t)
	ctx := context.Background()

	native, ok := store.(contract.AtomicCacheStore)
	require.True(t, ok, "redis store must expose the native atomic capability")

	t.Run("increment creates a missing key without expiration", func(t *testing.T) {
		n, err := native.Increment("atomic:new", 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), n)

		ttl, err := verify.TTL(ctx, "atomic:new").Result()
		require.NoError(t, err)
		assert.Less(t, ttl, time.Duration(0), "key created by Increment must have no expiration")
	})

	t.Run("increment preserves an existing key's TTL", func(t *testing.T) {
		require.NoError(t, store.Set("atomic:ttl", []byte("5"), 30*time.Second))

		n, err := native.Increment("atomic:ttl", 2)
		require.NoError(t, err)
		assert.Equal(t, int64(7), n)

		ttl, err := verify.TTL(ctx, "atomic:ttl").Result()
		require.NoError(t, err)
		assert.Greater(t, ttl, time.Duration(0), "increment must not strip the TTL")
		assert.LessOrEqual(t, ttl, 30*time.Second)
	})

	t.Run("increment rejects non-integer values", func(t *testing.T) {
		require.NoError(t, store.Set("atomic:str", []byte(`"nope"`), 30*time.Second))

		_, err := native.Increment("atomic:str", 1)
		assert.Error(t, err)
	})

	t.Run("set-if-not-exists stores only once and applies the TTL", func(t *testing.T) {
		stored, err := native.SetIfNotExists("atomic:add", []byte(`"first"`), 30*time.Second)
		require.NoError(t, err)
		assert.True(t, stored)

		stored, err = native.SetIfNotExists("atomic:add", []byte(`"second"`), 30*time.Second)
		require.NoError(t, err)
		assert.False(t, stored, "second write must be rejected")

		data, err := store.Get("atomic:add")
		require.NoError(t, err)
		assert.Equal(t, []byte(`"first"`), data)

		ttl, err := verify.TTL(ctx, "atomic:add").Result()
		require.NoError(t, err)
		assert.Greater(t, ttl, time.Duration(0))
	})

	t.Run("set-if-not-exists with zero exp stores without expiration", func(t *testing.T) {
		stored, err := native.SetIfNotExists("atomic:forever", []byte(`1`), 0)
		require.NoError(t, err)
		assert.True(t, stored)

		ttl, err := verify.TTL(ctx, "atomic:forever").Result()
		require.NoError(t, err)
		assert.Less(t, ttl, time.Duration(0))
	})

	t.Run("get-delete consumes the key in one step", func(t *testing.T) {
		require.NoError(t, store.Set("atomic:pull", []byte(`"v"`), 30*time.Second))

		data, err := native.GetDelete("atomic:pull")
		require.NoError(t, err)
		assert.Equal(t, []byte(`"v"`), data)

		gone, err := store.Get("atomic:pull")
		require.NoError(t, err)
		assert.Nil(t, gone)
	})

	t.Run("get-delete misses with nil, nil", func(t *testing.T) {
		data, err := native.GetDelete("atomic:absent")
		require.NoError(t, err)
		assert.Nil(t, data)
	})
}

// End-to-end through contract.Cache: the compound operations must ride the
// native capability, which shows up as behavior the emulated path cannot
// provide — most importantly Increment leaving a seeded TTL intact.
func TestRedisCacheAtomicEndToEndIntegration(t *testing.T) {
	store := newAtomicTestStore(t)
	verify := newVerifyClient(t)
	ctx := context.Background()

	c := New(store, "e2e", 0)

	t.Run("rate limiter pattern: Add seeds the TTL, Increment counts within it", func(t *testing.T) {
		require.NoError(t, c.Add("attempts", 0, 30*time.Second))

		for want := int64(1); want <= 3; want++ {
			n, err := c.Increment("attempts")
			require.NoError(t, err)
			assert.Equal(t, want, n)
		}

		ttl, err := verify.TTL(ctx, "e2e:attempts").Result()
		require.NoError(t, err)
		assert.Greater(t, ttl, time.Duration(0), "lockout window must survive the increments")
		assert.LessOrEqual(t, ttl, 30*time.Second)
	})

	t.Run("concurrent increments never lose an update", func(t *testing.T) {
		var wg sync.WaitGroup
		for range 50 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := c.Increment("hits")
				assert.NoError(t, err)
			}()
		}
		wg.Wait()

		var got int64
		require.NoError(t, c.Get("hits", &got))
		assert.Equal(t, int64(50), got)
	})

	t.Run("typed Pull consumes a one-time value exactly once", func(t *testing.T) {
		require.NoError(t, c.Set("otp", map[string]string{"purpose": "signup"}, 30*time.Second))

		v, found, err := Pull[map[string]string](c, "otp")
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, "signup", v["purpose"])

		_, found, err = Pull[map[string]string](c, "otp")
		require.NoError(t, err)
		assert.False(t, found, "second pull must miss")
	})
}
