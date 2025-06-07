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

func main() {
	// Show different driver configurations
	fmt.Println("Cache Drivers Demo")
	fmt.Println("==================")
	fmt.Println()
	fmt.Println("This demo shows how to configure different Fiber storage drivers.")
	fmt.Println()

	// Let user choose which demo to run
	if len(os.Args) > 1 && os.Args[1] == "redis" {
		fmt.Println("Running with Redis driver...")
		runRedisDemo()
	} else {
		fmt.Println("Running with Memory driver (default)...")
		runMemoryDemo()
	}
}

func runMemoryDemo() {
	// Create env file for memory driver
	envContent := `# Memory Driver Configuration
APP_NAME=CacheDriversDemo
APP_VERSION=1.0.0

# Using Fiber's memory driver
CACHE_DRIVER=memory
CACHE_PREFIX=demo
CACHE_MEMORY_GC_INTERVAL=5s

# Multiple stores with memory driver
CACHE_STORE=primary
CACHE_primary_DRIVER=memory
CACHE_primary_PREFIX=primary

CACHE_secondary_DRIVER=memory
CACHE_secondary_PREFIX=secondary
CACHE_secondary_MEMORY_GC_INTERVAL=10s

# Logging
LOG_LEVEL=info
LOG_FORMAT=text
`
	_ = os.WriteFile(".env", []byte(envContent), 0644)
	defer os.Remove(".env")

	app := goe.New(goe.Options{
		WithCache: true,
		Invokers: []any{
			runCacheDemo,
		},
	})

	app.Container().Run()
}

func runRedisDemo() {
	// Create env file for Redis driver
	envContent := `# Redis Driver Configuration
APP_NAME=CacheDriversDemo
APP_VERSION=1.0.0

# Using Fiber's Redis driver (rueidis)
CACHE_DRIVER=redis
CACHE_PREFIX=demo

# Redis connection settings
# Option 1: Using URL
CACHE_REDIS_URL=redis://localhost:6379/0

# Option 2: Using individual settings (uncomment to use)
# CACHE_REDIS_HOSTS=localhost:6379
# CACHE_REDIS_DATABASE=0
# CACHE_REDIS_PASSWORD=yourpassword
# CACHE_REDIS_USERNAME=
# CACHE_REDIS_CLIENT_NAME=cache-demo

# Advanced Redis settings
CACHE_REDIS_CACHE_SIZE=134217728       # 128MB client-side cache
CACHE_REDIS_ALWAYS_PIPELINING=true     # Enable auto-pipelining
CACHE_REDIS_DISABLE_CACHE=false        # Enable client-side caching
CACHE_REDIS_CACHE_TTL=1m              # Client cache TTL

# Multiple stores
CACHE_STORE=primary
CACHE_primary_DRIVER=redis
CACHE_primary_REDIS_DATABASE=0
CACHE_primary_PREFIX=primary

CACHE_secondary_DRIVER=redis  
CACHE_secondary_REDIS_DATABASE=1
CACHE_secondary_PREFIX=secondary

# Logging
LOG_LEVEL=info
LOG_FORMAT=text
`
	_ = os.WriteFile(".env", []byte(envContent), 0644)
	defer os.Remove(".env")

	fmt.Println("\nNOTE: Make sure Redis is running on localhost:6379")
	fmt.Println("You can start Redis with: docker run -d -p 6379:6379 redis:latest")
	fmt.Println()

	app := goe.New(goe.Options{
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
			goe.Log().Info("Starting cache drivers demo")

			// Show driver info
			fmt.Printf("\nDriver Information:\n")
			fmt.Printf("Default Driver: %s\n", cacheManager.Driver())
			fmt.Printf("Driver Info: %s\n\n", cache.GetDriverInfo(cacheManager.Driver()))

			// Get primary cache
			primaryCache := cacheManager.Store("primary")
			fmt.Printf("Primary Cache:\n")
			fmt.Printf("- Driver: %s\n", cacheManager.Driver())
			fmt.Printf("- Prefix: %s\n\n", primaryCache.GetPrefix())

			// Basic operations
			fmt.Println("=== Basic Operations ===")
			
			// Set value
			key := "demo:key"
			value := map[string]interface{}{
				"message": "Hello from Fiber cache driver!",
				"driver":  cacheManager.Driver(),
				"time":    time.Now().Format(time.RFC3339),
			}
			
			err := primaryCache.Set(key, value, 5*time.Minute)
			if err != nil {
				return fmt.Errorf("failed to set value: %w", err)
			}
			fmt.Printf("✓ Set value with key '%s'\n", key)

			// Get value
			retrieved, err := primaryCache.Get(key)
			if err != nil {
				return fmt.Errorf("failed to get value: %w", err)
			}
			fmt.Printf("✓ Retrieved value: %v\n", retrieved)

			// Type-safe operations
			fmt.Println("\n=== Type-safe Operations ===")
			type Config struct {
				AppName    string    `json:"app_name"`
				Version    string    `json:"version"`
				Driver     string    `json:"driver"`
				UpdatedAt  time.Time `json:"updated_at"`
			}

			config := Config{
				AppName:   "CacheDriversDemo",
				Version:   "1.0.0", 
				Driver:    cacheManager.Driver(),
				UpdatedAt: time.Now(),
			}

			// Store typed value
			err = primaryCache.Set("config", config, 10*time.Minute)
			if err != nil {
				return fmt.Errorf("failed to set config: %w", err)
			}
			fmt.Println("✓ Stored typed config object")

			// Retrieve with type safety
			retrievedConfig, err := cache.GetT[Config](primaryCache, "config")
			if err != nil {
				return fmt.Errorf("failed to get typed config: %w", err)
			}
			fmt.Printf("✓ Retrieved typed config: %+v\n", retrievedConfig)

			// Benchmark cache performance
			fmt.Println("\n=== Performance Test ===")
			iterations := 1000
			
			// Write performance
			start := time.Now()
			for i := 0; i < iterations; i++ {
				key := fmt.Sprintf("perf:key:%d", i)
				_ = primaryCache.Set(key, i, 1*time.Minute)
			}
			writeTime := time.Since(start)
			fmt.Printf("✓ Write %d keys: %v (%.2f ops/sec)\n", 
				iterations, writeTime, float64(iterations)/writeTime.Seconds())

			// Read performance
			start = time.Now()
			for i := 0; i < iterations; i++ {
				key := fmt.Sprintf("perf:key:%d", i)
				_, _ = primaryCache.Get(key)
			}
			readTime := time.Since(start)
			fmt.Printf("✓ Read %d keys: %v (%.2f ops/sec)\n", 
				iterations, readTime, float64(iterations)/readTime.Seconds())

			// Multiple stores test
			fmt.Println("\n=== Multiple Stores ===")
			secondaryCache := cacheManager.Store("secondary")
			
			err = primaryCache.Set("store:test", "primary value", 1*time.Minute)
			if err != nil {
				return fmt.Errorf("failed to set in primary: %w", err)
			}
			
			err = secondaryCache.Set("store:test", "secondary value", 1*time.Minute)
			if err != nil {
				return fmt.Errorf("failed to set in secondary: %w", err)
			}

			primaryVal, _ := primaryCache.Get("store:test")
			secondaryVal, _ := secondaryCache.Get("store:test")
			
			fmt.Printf("✓ Primary store value: %v\n", primaryVal)
			fmt.Printf("✓ Secondary store value: %v\n", secondaryVal)
			fmt.Println("✓ Stores are properly isolated")

			// Remember pattern performance
			fmt.Println("\n=== Remember Pattern ===")
			computeCount := 0
			
			// First call - computes value
			start = time.Now()
			result, err := cache.RememberT(primaryCache, "expensive:operation", 5*time.Minute, func() (string, error) {
				computeCount++
				time.Sleep(100 * time.Millisecond) // Simulate expensive operation
				return "Computed expensive result", nil
			})
			if err != nil {
				return fmt.Errorf("remember failed: %w", err)
			}
			firstCallTime := time.Since(start)
			fmt.Printf("✓ First call (computed): %v, Result: %s\n", firstCallTime, result)

			// Second call - from cache
			start = time.Now()
			result, err = cache.RememberT(primaryCache, "expensive:operation", 5*time.Minute, func() (string, error) {
				computeCount++
				time.Sleep(100 * time.Millisecond)
				return "Should not compute", nil
			})
			if err != nil {
				return fmt.Errorf("remember failed: %w", err)
			}
			secondCallTime := time.Since(start)
			fmt.Printf("✓ Second call (cached): %v, Result: %s\n", secondCallTime, result)
			fmt.Printf("✓ Compute function called %d time(s)\n", computeCount)
			fmt.Printf("✓ Cache speedup: %.0fx faster\n", firstCallTime.Seconds()/secondCallTime.Seconds())

			// Driver-specific features
			if cacheManager.Driver() == "redis" {
				fmt.Println("\n=== Redis-specific Features ===")
				fmt.Println("✓ Auto-pipelining enabled for better performance")
				fmt.Println("✓ Client-side caching enabled for reduced latency")
				fmt.Println("✓ Cluster support available via multiple hosts")
			}

			fmt.Println("\n✅ Cache drivers demo completed successfully!")
			
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Cleanup performance test keys
			primaryCache := cacheManager.Store("primary")
			for i := 0; i < 1000; i++ {
				key := fmt.Sprintf("perf:key:%d", i)
				_ = primaryCache.Forget(key)
			}
			return nil
		},
	})
}

func init() {
	// Show usage if needed
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Println("Usage: go run main.go [driver]")
		fmt.Println()
		fmt.Println("Drivers:")
		fmt.Println("  memory  - Use Fiber's in-memory driver (default)")
		fmt.Println("  redis   - Use Fiber's Redis driver (requires Redis server)")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run main.go         # Uses memory driver")
		fmt.Println("  go run main.go memory  # Explicitly use memory driver")  
		fmt.Println("  go run main.go redis   # Use Redis driver")
		os.Exit(0)
	}
}