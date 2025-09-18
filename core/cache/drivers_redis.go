package cache

import (
	"fmt"
	"time"

	"github.com/gofiber/storage/memory/v2"
	"github.com/gofiber/storage/redis/v3"
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

// RedisStoreFactory creates Fiber Redis store instances
func RedisStoreFactory(config contract.Config) (contract.CacheStore, error) {
	// Build Redis configuration from environment
	redisConfig := redis.Config{
		Reset: config.GetBool("CACHE_REDIS_RESET"),
	}

	// Check if URL is provided
	if url := config.GetString("CACHE_REDIS_URL"); url != "" {
		redisConfig.URL = url
	} else {
		// Use individual configuration
		host := config.GetString("CACHE_REDIS_HOST")
		if host == "" {
			host = "127.0.0.1"
		}
		redisConfig.Host = host

		port := config.GetInt("CACHE_REDIS_PORT")
		if port == 0 {
			port = 6379
		}
		redisConfig.Port = port

		if username := config.GetString("CACHE_REDIS_USERNAME"); username != "" {
			redisConfig.Username = username
		}

		if password := config.GetString("CACHE_REDIS_PASSWORD"); password != "" {
			redisConfig.Password = password
		}

		if clientName := config.GetString("CACHE_REDIS_CLIENT_NAME"); clientName != "" {
			redisConfig.ClientName = clientName
		}

		redisConfig.Database = config.GetInt("CACHE_REDIS_DATABASE")

		// Check for cluster mode with multiple hosts
		hosts := config.GetStringSlice("CACHE_REDIS_ADDRS")
		if len(hosts) > 0 {
			redisConfig.Addrs = hosts
		}

		// Failover configuration
		if masterName := config.GetString("CACHE_REDIS_MASTER_NAME"); masterName != "" {
			redisConfig.MasterName = masterName
		}
	}

	// Pool size configuration
	if poolSize := config.GetInt("CACHE_REDIS_POOL_SIZE"); poolSize > 0 {
		redisConfig.PoolSize = poolSize
	}

	// Cluster mode
	redisConfig.IsClusterMode = config.GetBool("CACHE_REDIS_IS_CLUSTER_MODE")

	return redis.New(redisConfig), nil
}

// RegisterBuiltinDrivers registers all built-in cache drivers
func RegisterBuiltinDrivers(manager contract.CacheManager) {
	// Register Fiber storage drivers
	manager.Extend("memory", MemoryStoreFactory)
	manager.Extend("redis", RedisStoreFactory)
}

// GetDriverInfo returns information about available drivers
func GetDriverInfo(driver string) string {
	switch driver {
	case "memory":
		return "Fiber in-memory storage with automatic garbage collection"
	case "redis":
		return "Fiber Redis storage using go-redis client with connection pooling and cluster support"
	default:
		return fmt.Sprintf("Unknown driver: %s", driver)
	}
}
