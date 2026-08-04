// Package cache provides an optional caching module that can be registered with
// the application when required. The module exposes a CacheManager abstraction
// so applications can swap cache backends without changing business logic.
//
// # Configuration
//
// Configure via environment variables:
//
//	CACHE_STORE=redis                  # Default store; a registered driver name works directly
//	CACHE_PREFIX=myapp                 # Key prefix (falls back to APP_NAME)
//	CACHE_TTL=2h                       # Default TTL
//	CACHE_REDIS_URL=redis://host:6379  # Redis URL (environment-only, wins over host/port)
//	CACHE_REDIS_HOST=127.0.0.1         # Redis host (used when no URL is set)
//	CACHE_REDIS_PORT=6379              # Redis port
//	CACHE_REDIS_USERNAME=svc           # Redis username (environment-only)
//	CACHE_REDIS_PASSWORD=secret        # Redis password (environment-only)
//
// Or from Go code through goe.Options.Cache — every CACHE_* key has a
// matching Option (CACHE_REDIS_POOL_SIZE is cache.WithRedisPoolSize), and
// code wins over the environment:
//
//	goe.New(goe.Options{
//	    Cache: []cache.Option{
//	        cache.WithStore("redis"),
//	        cache.WithTTL(30 * time.Minute),
//	        cache.WithRedisHost("redis.internal"),
//	    },
//	})
//
// Credentials are the exception: CACHE_REDIS_URL, CACHE_REDIS_USERNAME and
// CACHE_REDIS_PASSWORD have no Option. Secrets stay in the environment and
// are applied to whichever endpoint wins — including one chosen in code with
// cache.WithRedisHost.
package cache
