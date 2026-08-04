package cache

import (
	"fmt"
	"net/url"
	"time"

	"github.com/gofiber/storage/memory/v2"
	"github.com/gofiber/storage/redis/v3"
	goredis "github.com/redis/go-redis/v9"
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
	return redis.New(buildRedisConfig(config)), nil
}

// buildRedisConfig maps CACHE_REDIS_* configuration onto the Fiber Redis
// storage config.
func buildRedisConfig(config contract.Config) redis.Config {
	// Build Redis configuration from environment
	redisConfig := redis.Config{
		Reset: config.GetBool("CACHE_REDIS_RESET"),
	}

	// Check if URL is provided
	if url := config.GetString("CACHE_REDIS_URL"); url != "" {
		// The URL flows through to the Fiber storage as-is — with one
		// exception. The storage ignores the discrete Username/Password
		// fields once a URL is set, which would silently drop credentials
		// supplied through CACHE_REDIS_USERNAME/CACHE_REDIS_PASSWORD (the
		// only way to pair a secret with a credential-free URL). When that
		// combination occurs the credentials are injected into the URL's
		// userinfo; everything else about the URL — scheme, host, database,
		// query parameters — is preserved.
		redisConfig.URL = injectURLCredentials(url,
			config.GetString("CACHE_REDIS_USERNAME"),
			config.GetString("CACHE_REDIS_PASSWORD"))
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

	return redisConfig
}

// injectURLCredentials returns rawURL with username/password embedded as
// userinfo, when there is something to embed and the URL carries none of its
// own. URL-embedded credentials win, and a URL go-redis cannot parse is
// returned untouched so the storage reports it exactly as before.
func injectURLCredentials(rawURL, username, password string) string {
	if username == "" && password == "" {
		return rawURL
	}

	opt, err := goredis.ParseURL(rawURL)
	if err != nil || opt.Username != "" || opt.Password != "" {
		return rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if password == "" {
		u.User = url.User(username)
	} else {
		u.User = url.UserPassword(username, password)
	}
	return u.String()
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
