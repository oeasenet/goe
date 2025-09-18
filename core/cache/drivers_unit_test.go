package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.oease.dev/goe/v2/core/config"
)

func TestMemoryStoreFactory(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*config.Module)
	}{
		{
			name: "default configuration",
			setupFunc: func(cfg *config.Module) {
				// No special setup needed
			},
		},
		{
			name: "custom GC interval",
			setupFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_MEMORY_GC_INTERVAL", 30*time.Second)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config module
			cfg := config.NewModule()
			tt.setupFunc(cfg)

			// Create memory store
			store, err := MemoryStoreFactory(cfg.Provide())

			assert.NoError(t, err)
			assert.NotNil(t, store)

			// Test basic operation
			key := "test:key"
			value := []byte("test value")

			err = store.Set(key, value, 10*time.Second)
			assert.NoError(t, err)

			retrieved, err := store.Get(key)
			assert.NoError(t, err)
			assert.Equal(t, value, retrieved)
		})
	}
}

func TestGetDriverInfo(t *testing.T) {
	tests := []struct {
		driver   string
		expected string
	}{
		{
			driver:   "memory",
			expected: "Fiber in-memory storage with automatic garbage collection",
		},
		{
			driver:   "redis",
			expected: "Fiber Redis storage using go-redis client with connection pooling and cluster support",
		},
		{
			driver:   "unknown",
			expected: "Unknown driver: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			info := GetDriverInfo(tt.driver)
			assert.Equal(t, tt.expected, info)
		})
	}
}
