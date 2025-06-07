package tests

import (
	"testing"
	"time"

	"go.oease.dev/goe/v2/core/cache"
	"go.oease.dev/goe/v2/core/config"
	"go.oease.dev/goe/v2/core/log"
)

func TestMemoryCache(t *testing.T) {
	// Create config and logger
	cfg := config.New()
	logger := log.NewDefault()

	// Create cache module
	cacheModule := cache.NewModule(cfg, logger)
	manager := cacheModule.Provide()

	// Get default cache store
	store := manager.Store()

	t.Run("Set and Get", func(t *testing.T) {
		// Test string value
		err := store.Set("test_key", "test_value", 1*time.Hour)
		if err != nil {
			t.Fatalf("Failed to set value: %v", err)
		}

		value, err := store.Get("test_key")
		if err != nil {
			t.Fatalf("Failed to get value: %v", err)
		}

		if value != "test_value" {
			t.Errorf("Expected 'test_value', got %v", value)
		}
	})

	t.Run("GetT with type", func(t *testing.T) {
		// Test struct value
		type User struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}

		user := User{ID: 1, Name: "John Doe"}
		err := store.Set("user:1", user, 1*time.Hour)
		if err != nil {
			t.Fatalf("Failed to set user: %v", err)
		}

		retrievedUser, err := cache.GetT[User](store, "user:1")
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if retrievedUser.ID != user.ID || retrievedUser.Name != user.Name {
			t.Errorf("Expected %+v, got %+v", user, retrievedUser)
		}
	})

	t.Run("Has", func(t *testing.T) {
		err := store.Set("exists_key", "value", 1*time.Hour)
		if err != nil {
			t.Fatalf("Failed to set value: %v", err)
		}

		if !store.Has("exists_key") {
			t.Error("Expected key to exist")
		}

		if store.Has("non_existent_key") {
			t.Error("Expected key to not exist")
		}
	})

	t.Run("Forever", func(t *testing.T) {
		err := store.Forever("forever_key", "forever_value")
		if err != nil {
			t.Fatalf("Failed to set forever value: %v", err)
		}

		value, err := store.Get("forever_key")
		if err != nil {
			t.Fatalf("Failed to get forever value: %v", err)
		}

		if value != "forever_value" {
			t.Errorf("Expected 'forever_value', got %v", value)
		}
	})

	t.Run("Forget", func(t *testing.T) {
		err := store.Set("delete_key", "value", 1*time.Hour)
		if err != nil {
			t.Fatalf("Failed to set value: %v", err)
		}

		err = store.Forget("delete_key")
		if err != nil {
			t.Fatalf("Failed to forget key: %v", err)
		}

		if store.Has("delete_key") {
			t.Error("Expected key to be deleted")
		}
	})

	t.Run("Remember", func(t *testing.T) {
		callCount := 0
		callback := func() (any, error) {
			callCount++
			return "computed_value", nil
		}

		// First call should compute
		value, err := store.Remember("remember_key", 1*time.Hour, callback)
		if err != nil {
			t.Fatalf("Failed to remember: %v", err)
		}
		if value != "computed_value" {
			t.Errorf("Expected 'computed_value', got %v", value)
		}
		if callCount != 1 {
			t.Errorf("Expected callback to be called once, called %d times", callCount)
		}

		// Second call should get from cache
		value, err = store.Remember("remember_key", 1*time.Hour, callback)
		if err != nil {
			t.Fatalf("Failed to remember: %v", err)
		}
		if value != "computed_value" {
			t.Errorf("Expected 'computed_value', got %v", value)
		}
		if callCount != 1 {
			t.Errorf("Expected callback to still be called once, called %d times", callCount)
		}
	})

	t.Run("RememberT with type", func(t *testing.T) {
		type Product struct {
			ID    int     `json:"id"`
			Name  string  `json:"name"`
			Price float64 `json:"price"`
		}

		callCount := 0
		callback := func() (Product, error) {
			callCount++
			return Product{ID: 1, Name: "Widget", Price: 9.99}, nil
		}

		// First call should compute
		product, err := cache.RememberT(store, "product:1", 1*time.Hour, callback)
		if err != nil {
			t.Fatalf("Failed to remember product: %v", err)
		}
		if product.ID != 1 || product.Name != "Widget" || product.Price != 9.99 {
			t.Errorf("Unexpected product: %+v", product)
		}
		if callCount != 1 {
			t.Errorf("Expected callback to be called once, called %d times", callCount)
		}

		// Second call should get from cache
		product, err = cache.RememberT(store, "product:1", 1*time.Hour, callback)
		if err != nil {
			t.Fatalf("Failed to remember product: %v", err)
		}
		if callCount != 1 {
			t.Errorf("Expected callback to still be called once, called %d times", callCount)
		}
	})

	t.Run("Pull", func(t *testing.T) {
		err := store.Set("pull_key", "pull_value", 1*time.Hour)
		if err != nil {
			t.Fatalf("Failed to set value: %v", err)
		}

		value, err := store.Pull("pull_key")
		if err != nil {
			t.Fatalf("Failed to pull value: %v", err)
		}

		if value != "pull_value" {
			t.Errorf("Expected 'pull_value', got %v", value)
		}

		// Key should be deleted after pull
		if store.Has("pull_key") {
			t.Error("Expected key to be deleted after pull")
		}
	})

	t.Run("Add", func(t *testing.T) {
		// Add to non-existent key should succeed
		err := store.Add("new_key", "new_value", 1*time.Hour)
		if err != nil {
			t.Fatalf("Failed to add new key: %v", err)
		}

		// Add to existing key should fail
		err = store.Add("new_key", "another_value", 1*time.Hour)
		if err == nil {
			t.Error("Expected error when adding to existing key")
		}
	})

	t.Run("Increment and Decrement", func(t *testing.T) {
		// Test increment from zero
		val, err := store.Increment("counter")
		if err != nil {
			t.Fatalf("Failed to increment: %v", err)
		}
		if val != 1 {
			t.Errorf("Expected 1, got %d", val)
		}

		// Test increment by custom value
		val, err = store.Increment("counter", 5)
		if err != nil {
			t.Fatalf("Failed to increment by 5: %v", err)
		}
		if val != 6 {
			t.Errorf("Expected 6, got %d", val)
		}

		// Test decrement
		val, err = store.Decrement("counter", 2)
		if err != nil {
			t.Fatalf("Failed to decrement: %v", err)
		}
		if val != 4 {
			t.Errorf("Expected 4, got %d", val)
		}
	})

	t.Run("Flush", func(t *testing.T) {
		// Add some keys
		_ = store.Set("flush_key1", "value1", 1*time.Hour)
		_ = store.Set("flush_key2", "value2", 1*time.Hour)
		_ = store.Set("flush_key3", "value3", 1*time.Hour)

		// Flush all
		err := store.Flush()
		if err != nil {
			t.Fatalf("Failed to flush: %v", err)
		}

		// Check all keys are gone
		if store.Has("flush_key1") || store.Has("flush_key2") || store.Has("flush_key3") {
			t.Error("Expected all keys to be flushed")
		}
	})

	t.Run("Expiration", func(t *testing.T) {
		// Set with short TTL
		err := store.Set("expire_key", "expire_value", 1*time.Second)
		if err != nil {
			t.Fatalf("Failed to set value: %v", err)
		}

		// Should exist immediately
		value, err := store.Get("expire_key")
		if err != nil || value == nil {
			t.Error("Expected key to exist immediately after set")
		}

		// Wait for expiration
		time.Sleep(2 * time.Second)

		// Should not exist after expiration
		value, err = store.Get("expire_key")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if value != nil {
			t.Error("Expected key to expire and return nil")
		}
	})
}

func TestCacheManager(t *testing.T) {
	// Set up configuration with multiple stores
	cfg := config.New()
	cfg.Set("CACHE_STORE", "primary")
	cfg.Set("CACHE_primary_DRIVER", "memory")
	cfg.Set("CACHE_secondary_DRIVER", "memory")

	logger := log.NewDefault()
	cacheModule := cache.NewModule(cfg, logger)
	manager := cacheModule.Provide()

	t.Run("Multiple Stores", func(t *testing.T) {
		// Get primary store
		primary := manager.Store("primary")
		err := primary.Set("key", "primary_value", 1*time.Hour)
		if err != nil {
			t.Fatalf("Failed to set in primary store: %v", err)
		}

		// Get secondary store
		secondary := manager.Store("secondary")
		err = secondary.Set("key", "secondary_value", 1*time.Hour)
		if err != nil {
			t.Fatalf("Failed to set in secondary store: %v", err)
		}

		// Values should be isolated
		primaryVal, _ := primary.Get("key")
		if primaryVal != "primary_value" {
			t.Errorf("Expected 'primary_value', got %v", primaryVal)
		}

		secondaryVal, _ := secondary.Get("key")
		if secondaryVal != "secondary_value" {
			t.Errorf("Expected 'secondary_value', got %v", secondaryVal)
		}
	})

	t.Run("Default Store", func(t *testing.T) {
		// Should use primary as default
		defaultStore := manager.Store()
		err := defaultStore.Set("default_key", "default_value", 1*time.Hour)
		if err != nil {
			t.Fatalf("Failed to set in default store: %v", err)
		}

		// Primary store should have the same value
		primary := manager.Store("primary")
		val, _ := primary.Get("default_key")
		if val != "default_value" {
			t.Errorf("Expected default store to be primary store")
		}
	})

	t.Run("Driver Name", func(t *testing.T) {
		if manager.Driver() != "memory" {
			t.Errorf("Expected driver to be 'memory', got %s", manager.Driver())
		}
	})
}

func TestCachePrefix(t *testing.T) {
	cfg := config.New()
	cfg.Set("CACHE_PREFIX", "myapp")
	cfg.Set("CACHE_secondary_PREFIX", "secondary")

	logger := log.NewDefault()
	cacheModule := cache.NewModule(cfg, logger)
	manager := cacheModule.Provide()

	t.Run("Global Prefix", func(t *testing.T) {
		store := manager.Store()
		if store.GetPrefix() != "myapp" {
			t.Errorf("Expected prefix 'myapp', got %s", store.GetPrefix())
		}
	})

	t.Run("Store-specific Prefix", func(t *testing.T) {
		store := manager.Store("secondary")
		if store.GetPrefix() != "secondary" {
			t.Errorf("Expected prefix 'secondary', got %s", store.GetPrefix())
		}
	})
}
