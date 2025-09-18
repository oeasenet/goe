//go:build integration
// +build integration

package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
