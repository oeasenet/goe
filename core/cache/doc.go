// Package cache provides an optional caching module that can be registered
// with the application when required. The module exposes a single
// contract.Cache backed by one configurable driver, so applications can swap
// cache backends without changing business logic.
//
// # Configuration
//
// Configure via environment variables:
//
//	CACHE_DRIVER=redis                 # memory (default), redis, badger, bbolt, or a custom driver
//	CACHE_PREFIX=myapp                 # Key prefix (falls back to APP_NAME)
//	CACHE_TTL=2h                       # Default TTL, applied when a caller passes ttl 0
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
//	        cache.WithDriver("redis"),
//	        cache.WithTTL(30 * time.Minute),
//	        cache.WithRedisHost("redis.internal"),
//	    },
//	})
//
// Credentials are the exception: CACHE_REDIS_URL, CACHE_REDIS_USERNAME and
// CACHE_REDIS_PASSWORD have no Option. Secrets stay in the environment and
// are applied to whichever endpoint wins — including one chosen in code with
// cache.WithRedisHost.
//
// # Default TTL
//
// CACHE_TTL / cache.WithTTL set the expiration used when a caller passes
// ttl == 0 to Set, Add or Remember. Without a configured default, ttl == 0
// keeps meaning "no expiration". Forever and RememberForever always store
// without expiration, and Increment/Decrement counters never expire.
//
// # Custom drivers
//
// Register additional backends with cache.WithCustomDriver and select them
// like any builtin. The factory reads its settings through the same layered
// configuration built-in drivers use:
//
//	goe.New(goe.Options{
//	    Cache: []cache.Option{
//	        cache.WithDriver("mystore"),
//	        cache.WithCustomDriver("mystore", NewMyStore),
//	    },
//	})
//
// # Migration from v2.4 (multi-store removal)
//
// v2.5 removed the multi-store layer: there is one cache per application,
// selected by CACHE_DRIVER. Stores never had per-store connections, so this
// removes aliases, not capability.
//
//   - CACHE_STORE=redis            -> CACHE_DRIVER=redis
//   - cache.WithStore("redis")     -> cache.WithDriver("redis")
//   - cache.WithStoreDriver/...    -> removed (per-store keys fail startup)
//   - contract.CacheManager        -> inject contract.Cache directly
//   - goe.Cache().Store().Set(...) -> goe.Cache().Set(...)
//   - CacheManager.Extend(...)     -> cache.WithCustomDriver(...)
//
// CACHE_STORE and CACHE_{name}_DRIVER in the environment fail startup
// validation with a migration message rather than being silently ignored.
package cache
