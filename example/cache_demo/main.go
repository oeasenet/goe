package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/cache"
	"go.uber.org/fx"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	// Create environment file for demo
	createEnvFile()

	app := goe.New(goe.Options{
		WithHTTP:  false,
		WithCache: true,
		Invokers: []any{
			runCacheDemo,
		},
	})

	app.Container().Run()
}

func runCacheDemo(lc fx.Lifecycle, cacheManager contract.CacheManager) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			goe.Log().Info("Starting cache demo")

			// Get default cache store
			store := cacheManager.Store()

			// Basic string operations
			fmt.Println("\n=== Basic String Operations ===")
			err := store.Set("greeting", "Hello, Cache!", 5*time.Minute)
			if err != nil {
				return fmt.Errorf("failed to set greeting: %w", err)
			}

			greeting, err := store.Get("greeting")
			if err != nil {
				return fmt.Errorf("failed to get greeting: %w", err)
			}
			fmt.Printf("Greeting: %v\n", greeting)

			// Working with structs
			fmt.Println("\n=== Working with Structs ===")
			user := User{
				ID:        1,
				Name:      "John Doe",
				Email:     "john@example.com",
				CreatedAt: time.Now(),
			}

			err = store.Set("user:1", user, 10*time.Minute)
			if err != nil {
				return fmt.Errorf("failed to cache user: %w", err)
			}

			// Using typed cache operations
			cachedUser, err := cache.GetT[User](store, "user:1")
			if err != nil {
				return fmt.Errorf("failed to get cached user: %w", err)
			}
			fmt.Printf("Cached User: %+v\n", cachedUser)

			// Remember pattern
			fmt.Println("\n=== Remember Pattern ===")
			product, err := cache.RememberT(store, "product:1", 1*time.Hour, func() (map[string]any, error) {
				fmt.Println("Computing product data...")
				return map[string]any{
					"id":    1,
					"name":  "Laptop",
					"price": 999.99,
					"stock": 10,
				}, nil
			})
			if err != nil {
				return fmt.Errorf("failed to remember product: %w", err)
			}
			fmt.Printf("Product: %+v\n", product)

			// Second call should not compute
			product2, err := cache.RememberT(store, "product:1", 1*time.Hour, func() (map[string]any, error) {
				fmt.Println("This should not be printed!")
				return nil, nil
			})
			if err != nil {
				return fmt.Errorf("failed to remember product again: %w", err)
			}
			fmt.Printf("Product (from cache): %+v\n", product2)

			// Counter operations
			fmt.Println("\n=== Counter Operations ===")
			views, err := store.Increment("page:home:views")
			if err != nil {
				return fmt.Errorf("failed to increment views: %w", err)
			}
			fmt.Printf("Page views: %d\n", views)

			views, err = store.Increment("page:home:views", 5)
			if err != nil {
				return fmt.Errorf("failed to increment views by 5: %w", err)
			}
			fmt.Printf("Page views after +5: %d\n", views)

			views, err = store.Decrement("page:home:views", 2)
			if err != nil {
				return fmt.Errorf("failed to decrement views: %w", err)
			}
			fmt.Printf("Page views after -2: %d\n", views)

			// Multiple stores
			fmt.Println("\n=== Multiple Stores ===")
			// Using default store
			defaultStore := cacheManager.Store()
			err = defaultStore.Set("default:key", "I'm in default store", 5*time.Minute)
			if err != nil {
				return fmt.Errorf("failed to set in default store: %w", err)
			}

			// Using secondary store
			secondaryStore := cacheManager.Store("secondary")
			err = secondaryStore.Set("secondary:key", "I'm in secondary store", 5*time.Minute)
			if err != nil {
				return fmt.Errorf("failed to set in secondary store: %w", err)
			}

			fmt.Printf("Default store driver: %s\n", cacheManager.Driver())
			fmt.Printf("Default store prefix: %s\n", defaultStore.GetPrefix())
			fmt.Printf("Secondary store prefix: %s\n", secondaryStore.GetPrefix())

			// Checking existence
			fmt.Println("\n=== Checking Existence ===")
			if store.Has("greeting") {
				fmt.Println("✓ 'greeting' key exists")
			}

			if !store.Has("non-existent") {
				fmt.Println("✓ 'non-existent' key does not exist")
			}

			// Pull operation
			fmt.Println("\n=== Pull Operation ===")
			tempValue := "This will be removed after retrieval"
			_ = store.Set("temp:key", tempValue, 1*time.Hour)

			pulled, err := store.Pull("temp:key")
			if err != nil {
				return fmt.Errorf("failed to pull: %w", err)
			}
			fmt.Printf("Pulled value: %v\n", pulled)

			if !store.Has("temp:key") {
				fmt.Println("✓ Key was removed after pull")
			}

			// Add operation (only if not exists)
			fmt.Println("\n=== Add Operation ===")
			err = store.Add("unique:key", "First value", 1*time.Hour)
			if err != nil {
				return fmt.Errorf("failed to add unique key: %w", err)
			}
			fmt.Println("✓ Successfully added unique key")

			err = store.Add("unique:key", "Second value", 1*time.Hour)
			if err != nil {
				fmt.Printf("✓ Expected error when adding existing key: %v\n", err)
			}

			// Typed cache wrapper
			fmt.Println("\n=== Typed Cache Wrapper ===")
			userCache := cache.NewTyped[User](store)

			err = userCache.Set("typed:user:2", User{
				ID:        2,
				Name:      "Jane Smith",
				Email:     "jane@example.com",
				CreatedAt: time.Now(),
			}, 30*time.Minute)
			if err != nil {
				return fmt.Errorf("failed to set typed user: %w", err)
			}

			typedUser, err := userCache.Get("typed:user:2")
			if err != nil {
				return fmt.Errorf("failed to get typed user: %w", err)
			}
			fmt.Printf("Typed User: %+v\n", typedUser)

			goe.Log().Info("Cache demo completed successfully")

			return nil
		},
	})
}

func createEnvFile() {
	envContent := `# Cache Configuration
APP_NAME=CacheDemo
APP_VERSION=1.0.0

# Default cache store using Fiber memory driver
CACHE_STORE=default
CACHE_DRIVER=memory
CACHE_PREFIX=demo
CACHE_MEMORY_GC_INTERVAL=10s

# Secondary cache store
CACHE_secondary_DRIVER=memory
CACHE_secondary_PREFIX=secondary

# Example Redis configuration (commented out)
# CACHE_STORE=primary
# CACHE_DRIVER=redis
# CACHE_REDIS_URL=redis://localhost:6379/0
# Or use individual settings:
# CACHE_REDIS_HOSTS=localhost:6379
# CACHE_REDIS_DATABASE=0
# CACHE_REDIS_PASSWORD=
# CACHE_REDIS_CLIENT_NAME=cache-demo

# Logging
LOG_LEVEL=info
LOG_FORMAT=text
`
	_ = os.WriteFile(".env", []byte(envContent), 0644)
}