package cache

import (
	"fmt"
	"time"

	"github.com/gofiber/storage/memory/v2"
	"github.com/gofiber/storage/rueidis"
	"go.oease.dev/goe/v2/contract"
)

// MemoryStoreFactory creates Fiber memory store instances
func MemoryStoreFactory(config contract.Config) (contract.CacheStore, error) {
	// Get GC interval from config
	gcInterval := config.GetDuration("CACHE_MEMORY_GC_INTERVAL")
	if gcInterval == 0 {
		gcInterval = 10 * time.Second
	}

	return memory.New(memory.Config{
		GCInterval: gcInterval,
	}), nil
}

// RedisStoreFactory creates Fiber Redis store instances using rueidis
func RedisStoreFactory(config contract.Config) (contract.CacheStore, error) {
	// Build Redis configuration from environment
	redisConfig := rueidis.Config{
		Reset: config.GetBool("CACHE_REDIS_RESET"),
	}

	// Check if URL is provided
	if url := config.GetString("CACHE_REDIS_URL"); url != "" {
		redisConfig.URL = url
	} else {
		// Use individual configuration
		hosts := config.GetStringSlice("CACHE_REDIS_HOSTS")
		if len(hosts) == 0 {
			// Default to localhost:6379
			hosts = []string{"localhost:6379"}
		}
		redisConfig.InitAddress = hosts

		if username := config.GetString("CACHE_REDIS_USERNAME"); username != "" {
			redisConfig.Username = username
		}

		if password := config.GetString("CACHE_REDIS_PASSWORD"); password != "" {
			redisConfig.Password = password
		}

		if clientName := config.GetString("CACHE_REDIS_CLIENT_NAME"); clientName != "" {
			redisConfig.ClientName = clientName
		}

		redisConfig.SelectDB = config.GetInt("CACHE_REDIS_DATABASE")
	}

	// Advanced configuration
	if cacheSize := config.GetInt("CACHE_REDIS_CACHE_SIZE"); cacheSize > 0 {
		redisConfig.CacheSizeEachConn = cacheSize
	}

	if blockingPoolSize := config.GetInt("CACHE_REDIS_BLOCKING_POOL_SIZE"); blockingPoolSize > 0 {
		redisConfig.BlockingPoolSize = blockingPoolSize
	}

	if pipelineMultiplex := config.GetInt("CACHE_REDIS_PIPELINE_MULTIPLEX"); pipelineMultiplex > 0 {
		redisConfig.PipelineMultiplex = pipelineMultiplex
	}

	redisConfig.DisableRetry = config.GetBool("CACHE_REDIS_DISABLE_RETRY")
	redisConfig.DisableCache = config.GetBool("CACHE_REDIS_DISABLE_CACHE")
	redisConfig.AlwaysPipelining = config.GetBool("CACHE_REDIS_ALWAYS_PIPELINING")
	if !config.Has("CACHE_REDIS_ALWAYS_PIPELINING") {
		redisConfig.AlwaysPipelining = true // Default to true
	}

	if cacheTTL := config.GetDuration("CACHE_REDIS_CACHE_TTL"); cacheTTL > 0 {
		redisConfig.CacheTTL = cacheTTL
	} else {
		redisConfig.CacheTTL = time.Minute // Default
	}

	return rueidis.New(redisConfig), nil
}

// RegisterBuiltinDrivers registers all built-in cache drivers
func RegisterBuiltinDrivers(manager contract.CacheManager) {
	// Register Fiber storage drivers
	manager.Extend("memory", MemoryStoreFactory)
	manager.Extend("redis", RedisStoreFactory)
	manager.Extend("rueidis", RedisStoreFactory) // Alias for redis
}

// GetDriverInfo returns information about available drivers
func GetDriverInfo(driver string) string {
	switch driver {
	case "memory":
		return "Fiber in-memory storage with automatic garbage collection"
	case "redis", "rueidis":
		return "Fiber Redis storage using rueidis client with auto-pipelining and client-side caching"
	default:
		return fmt.Sprintf("Unknown driver: %s", driver)
	}
}
