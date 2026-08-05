package cache

import (
	"errors"
	"fmt"
	"time"
)

// Option configures the cache module from Go code.
//
// Options are applied after environment variables, so a value set in code
// always wins over the matching CACHE_* variable. Environment variables still
// supply everything code does not set, which means an application that passes
// no options behaves exactly as it did before.
//
// The naming rule is mechanical: every CACHE_* environment key is exposed as
// With<Key> with the CACHE_ prefix dropped and the rest CamelCased —
// CACHE_REDIS_POOL_SIZE becomes WithRedisPoolSize. Store-scoped keys
// (CACHE_{NAME}_DRIVER and friends) take the store name as their first
// argument: WithStoreDriver("sessions", "redis") is CACHE_sessions_DRIVER.
//
// Credentials are the deliberate exception. CACHE_REDIS_URL (which can embed
// user:pass), CACHE_REDIS_USERNAME and CACHE_REDIS_PASSWORD have no option:
// secrets belong in the environment, not in source. The two compose —
// WithRedisHost and friends in code pick the endpoint while
// CACHE_REDIS_USERNAME and CACHE_REDIS_PASSWORD from the environment supply
// the credentials.
//
// Under the hood options become a configuration overlay, so custom drivers
// registered through Extend read code-configured values exactly as they read
// environment variables — no driver changes required.
//
// Options are applied in the order given, so the last write wins. If any
// option returns an error, every option is discarded, the environment-only
// configuration stays in effect, and Module.ValidateConfig reports all
// collected errors, which fails application startup before any store is
// served.
type Option func(*settings) error

// settings is the mutable state options write to. It is deliberately
// unexported: options are the only way in, which is what keeps credential
// keys out of reach of code configuration.
type settings struct {
	overrides map[string]any
}

// resolveOverrides applies opts and returns the configuration overlay. It
// reports every option error rather than stopping at the first, so a
// developer sees all of their mistakes in one startup failure instead of one
// per run.
func resolveOverrides(opts []Option) (map[string]any, []error) {
	s := &settings{overrides: make(map[string]any)}

	var errs []error
	for i, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(s); err != nil {
			errs = append(errs, fmt.Errorf("cache option %d: %w", i+1, err))
		}
	}

	return s.overrides, errs
}

// WithStore sets the default store name. Replaces CACHE_STORE.
//
// A store named after a registered driver ("memory", "redis", or a custom
// driver added through Extend) uses that driver directly; any other name
// needs its driver configured with WithStoreDriver or WithDriver.
func WithStore(name string) Option {
	return func(s *settings) error {
		if name == "" {
			return errors.New("WithStore: name must not be empty")
		}
		s.overrides["CACHE_STORE"] = name
		return nil
	}
}

// WithDriver sets the driver used by stores that have no store-specific
// driver configured. Replaces CACHE_DRIVER.
func WithDriver(driver string) Option {
	return func(s *settings) error {
		if driver == "" {
			return errors.New("WithDriver: driver must not be empty")
		}
		s.overrides["CACHE_DRIVER"] = driver
		return nil
	}
}

// WithPrefix sets the key prefix for all stores. Replaces CACHE_PREFIX,
// which itself falls back to APP_NAME.
func WithPrefix(prefix string) Option {
	return func(s *settings) error {
		if prefix == "" {
			return errors.New("WithPrefix: prefix must not be empty")
		}
		s.overrides["CACHE_PREFIX"] = prefix
		return nil
	}
}

// WithTTL sets the default cache TTL. Replaces CACHE_TTL.
func WithTTL(ttl time.Duration) Option {
	return func(s *settings) error {
		if ttl <= 0 {
			return fmt.Errorf("WithTTL: %v must be positive", ttl)
		}
		s.overrides["CACHE_TTL"] = ttl
		return nil
	}
}

// WithMemoryGCInterval sets the garbage-collection interval of the memory
// driver. Replaces CACHE_MEMORY_GC_INTERVAL.
func WithMemoryGCInterval(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithMemoryGCInterval: %v must be positive", d)
		}
		s.overrides["CACHE_MEMORY_GC_INTERVAL"] = d
		return nil
	}
}

// WithRedisHost sets the Redis host for the redis driver. Replaces
// CACHE_REDIS_HOST.
//
// Credentials never ride along here: CACHE_REDIS_USERNAME and
// CACHE_REDIS_PASSWORD from the environment are still applied to this
// connection.
func WithRedisHost(host string) Option {
	return func(s *settings) error {
		if host == "" {
			return errors.New("WithRedisHost: host must not be empty")
		}
		s.overrides["CACHE_REDIS_HOST"] = host
		return nil
	}
}

// WithRedisPort sets the Redis port for the redis driver. Replaces
// CACHE_REDIS_PORT.
func WithRedisPort(port int) Option {
	return func(s *settings) error {
		if port < 1 || port > 65535 {
			return fmt.Errorf("WithRedisPort: port %d is out of range 1-65535", port)
		}
		s.overrides["CACHE_REDIS_PORT"] = port
		return nil
	}
}

// WithRedisDatabase sets the Redis database number for the redis driver.
// Replaces CACHE_REDIS_DATABASE.
func WithRedisDatabase(db int) Option {
	return func(s *settings) error {
		if db < 0 {
			return fmt.Errorf("WithRedisDatabase: database %d must not be negative", db)
		}
		s.overrides["CACHE_REDIS_DATABASE"] = db
		return nil
	}
}

// WithRedisAddrs sets multiple Redis addresses for cluster or failover
// setups. Replaces CACHE_REDIS_ADDRS.
func WithRedisAddrs(addrs ...string) Option {
	return func(s *settings) error {
		if len(addrs) == 0 {
			return errors.New("WithRedisAddrs: at least one address is required")
		}
		for _, a := range addrs {
			if a == "" {
				return errors.New("WithRedisAddrs: address must not be empty")
			}
		}
		s.overrides["CACHE_REDIS_ADDRS"] = addrs
		return nil
	}
}

// WithRedisMasterName sets the Sentinel master name for failover setups.
// Replaces CACHE_REDIS_MASTER_NAME.
func WithRedisMasterName(name string) Option {
	return func(s *settings) error {
		if name == "" {
			return errors.New("WithRedisMasterName: name must not be empty")
		}
		s.overrides["CACHE_REDIS_MASTER_NAME"] = name
		return nil
	}
}

// WithRedisClientName sets the client name reported to Redis. Replaces
// CACHE_REDIS_CLIENT_NAME.
func WithRedisClientName(name string) Option {
	return func(s *settings) error {
		if name == "" {
			return errors.New("WithRedisClientName: name must not be empty")
		}
		s.overrides["CACHE_REDIS_CLIENT_NAME"] = name
		return nil
	}
}

// WithRedisPoolSize sets the Redis connection pool size. Replaces
// CACHE_REDIS_POOL_SIZE.
func WithRedisPoolSize(size int) Option {
	return func(s *settings) error {
		if size <= 0 {
			return fmt.Errorf("WithRedisPoolSize: size %d must be positive", size)
		}
		s.overrides["CACHE_REDIS_POOL_SIZE"] = size
		return nil
	}
}

// WithRedisClusterMode enables Redis cluster mode. Replaces
// CACHE_REDIS_IS_CLUSTER_MODE.
func WithRedisClusterMode(enabled bool) Option {
	return func(s *settings) error {
		s.overrides["CACHE_REDIS_IS_CLUSTER_MODE"] = enabled
		return nil
	}
}

// WithRedisReset clears all keys in the Redis store on startup. Replaces
// CACHE_REDIS_RESET. Almost never what you want outside tests.
func WithRedisReset(reset bool) Option {
	return func(s *settings) error {
		s.overrides["CACHE_REDIS_RESET"] = reset
		return nil
	}
}

// WithBadgerDatabase sets the directory the badger driver stores its data
// in. Replaces CACHE_BADGER_DATABASE. Defaults to ./fiber.badger.
func WithBadgerDatabase(path string) Option {
	return func(s *settings) error {
		if path == "" {
			return errors.New("WithBadgerDatabase: path must not be empty")
		}
		s.overrides["CACHE_BADGER_DATABASE"] = path
		return nil
	}
}

// WithBadgerReset clears all keys in the badger store on startup. Replaces
// CACHE_BADGER_RESET. Almost never what you want outside tests.
func WithBadgerReset(reset bool) Option {
	return func(s *settings) error {
		s.overrides["CACHE_BADGER_RESET"] = reset
		return nil
	}
}

// WithBadgerGCInterval sets how often the badger driver garbage-collects
// expired keys. Replaces CACHE_BADGER_GC_INTERVAL. Defaults to 10s.
func WithBadgerGCInterval(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithBadgerGCInterval: %v must be positive", d)
		}
		s.overrides["CACHE_BADGER_GC_INTERVAL"] = d
		return nil
	}
}

// WithBboltDatabase sets the file the bbolt driver stores its data in.
// Replaces CACHE_BBOLT_DATABASE. Defaults to fiber.db.
func WithBboltDatabase(path string) Option {
	return func(s *settings) error {
		if path == "" {
			return errors.New("WithBboltDatabase: path must not be empty")
		}
		s.overrides["CACHE_BBOLT_DATABASE"] = path
		return nil
	}
}

// WithBboltBucket sets the bbolt bucket keys are stored in. Replaces
// CACHE_BBOLT_BUCKET. Defaults to fiber_storage.
func WithBboltBucket(name string) Option {
	return func(s *settings) error {
		if name == "" {
			return errors.New("WithBboltBucket: name must not be empty")
		}
		s.overrides["CACHE_BBOLT_BUCKET"] = name
		return nil
	}
}

// WithBboltTimeout sets how long the bbolt driver waits to obtain the
// database file lock. Replaces CACHE_BBOLT_TIMEOUT. Defaults to 60s.
func WithBboltTimeout(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithBboltTimeout: %v must be positive", d)
		}
		s.overrides["CACHE_BBOLT_TIMEOUT"] = d
		return nil
	}
}

// WithBboltReset clears all keys in the bbolt bucket on startup. Replaces
// CACHE_BBOLT_RESET. Almost never what you want outside tests.
func WithBboltReset(reset bool) Option {
	return func(s *settings) error {
		s.overrides["CACHE_BBOLT_RESET"] = reset
		return nil
	}
}

// WithStoreDriver sets the driver for a named store, enabling multi-store
// setups from code. Replaces CACHE_{store}_DRIVER.
func WithStoreDriver(store, driver string) Option {
	return func(s *settings) error {
		if store == "" {
			return errors.New("WithStoreDriver: store must not be empty")
		}
		if driver == "" {
			return errors.New("WithStoreDriver: driver must not be empty")
		}
		s.overrides[fmt.Sprintf("CACHE_%s_DRIVER", store)] = driver
		return nil
	}
}

// WithStorePrefix sets the key prefix for a named store. Replaces
// CACHE_{store}_PREFIX.
func WithStorePrefix(store, prefix string) Option {
	return func(s *settings) error {
		if store == "" {
			return errors.New("WithStorePrefix: store must not be empty")
		}
		if prefix == "" {
			return errors.New("WithStorePrefix: prefix must not be empty")
		}
		s.overrides[fmt.Sprintf("CACHE_%s_PREFIX", store)] = prefix
		return nil
	}
}

// WithStoreTTL sets the default TTL for a named store. Replaces
// CACHE_{store}_TTL.
func WithStoreTTL(store string, ttl time.Duration) Option {
	return func(s *settings) error {
		if store == "" {
			return errors.New("WithStoreTTL: store must not be empty")
		}
		if ttl <= 0 {
			return fmt.Errorf("WithStoreTTL: %v must be positive", ttl)
		}
		s.overrides[fmt.Sprintf("CACHE_%s_TTL", store)] = ttl
		return nil
	}
}
