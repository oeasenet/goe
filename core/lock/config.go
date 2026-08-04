package lock

import (
	"time"

	"go.oease.dev/goe/v2/contract"
)

// Config holds the lock system configuration.
// The framework auto-detects connection mode from URL scheme:
//   - redis://host:port/db           → Single Redis instance
//   - rediss://host:port/db          → Single Redis with TLS
//   - redis-sentinel://master@host:port,host:port/db → Redis Sentinel
//   - redis-cluster://host:port,host:port           → Redis Cluster
//
// For Redlock algorithm with multiple independent Redis instances,
// provide multiple URLs in RedisURLs field.
type Config struct {
	// RedisURL is the primary connection URL.
	// Supported schemes:
	//   - redis://[user:pass@]host[:port][/db]
	//   - rediss://[user:pass@]host[:port][/db] (TLS)
	//   - redis-sentinel://[user:pass@]master@sentinel1[:port],sentinel2[:port][/db]
	//   - redis-cluster://[user:pass@]node1[:port],node2[:port]
	RedisURL string

	// RedisURLs allows multiple Redis URLs for the Redlock algorithm.
	// Each URL should be an independent redis:// or rediss:// instance.
	// When set, RedisURL is ignored and Redlock mode is used.
	RedisURLs []string

	// RedisUsername and RedisPassword are applied to every connection whose
	// URL does not embed its own userinfo (URL credentials win). They are
	// environment-only — LOCK_REDIS_USERNAME and LOCK_REDIS_PASSWORD have no
	// code Option, by design: secrets never belong in source, and code-side
	// URLs (WithRedisURL) reject embedded credentials for the same reason.
	RedisUsername string
	RedisPassword string

	// Lock defaults
	DefaultExpiry      time.Duration // Default lock expiry time (default: 8s)
	DefaultTries       int           // Default retry attempts (default: 32)
	DefaultRetryDelay  time.Duration // Default delay between retries (default: 500ms)
	DefaultDriftFactor float64       // Default clock drift factor (default: 0.01)
	DefaultKeyPrefix   string        // Prefix for all lock keys (default: "lock:")

	// Pool settings
	PoolSize int // Redis connection pool size (default: 10)

	// TLS settings (only used when scheme doesn't indicate TLS)
	TLSInsecureSkipVerify bool // Skip TLS certificate verification (not recommended)
}

// DefaultConfig returns the default lock configuration.
func DefaultConfig() *Config {
	return &Config{
		RedisURL:           "redis://localhost:6379/0",
		DefaultExpiry:      8 * time.Second,
		DefaultTries:       32,
		DefaultRetryDelay:  500 * time.Millisecond,
		DefaultDriftFactor: 0.01,
		DefaultKeyPrefix:   "lock:",
		PoolSize:           10,
	}
}

// LoadConfig loads configuration from GOE config using LOCK_* settings.
//
// Environment variables:
//   - LOCK_REDIS_URL: Single Redis URL (supports all schemes)
//   - LOCK_REDIS_URLS: Comma-separated URLs for Redlock
//   - LOCK_REDIS_USERNAME: Username applied when the URL embeds none (env-only)
//   - LOCK_REDIS_PASSWORD: Password applied when the URL embeds none (env-only)
//   - LOCK_DEFAULT_EXPIRY: Lock expiry duration (e.g., "8s")
//   - LOCK_DEFAULT_TRIES: Number of retry attempts
//   - LOCK_DEFAULT_RETRY_DELAY: Delay between retries (e.g., "500ms")
//   - LOCK_DEFAULT_DRIFT_FACTOR: Clock drift factor (e.g., "0.01")
//   - LOCK_KEY_PREFIX: Key prefix for locks
//   - LOCK_POOL_SIZE: Connection pool size
//   - LOCK_TLS_INSECURE_SKIP_VERIFY: Skip TLS verification (true/false)
func LoadConfig(config contract.Config) *Config {
	cfg := DefaultConfig()

	// Primary URL
	if url := config.GetString("LOCK_REDIS_URL"); url != "" {
		cfg.RedisURL = url
	}

	// Multiple URLs for Redlock
	if urls := config.GetStringSlice("LOCK_REDIS_URLS"); len(urls) > 0 {
		cfg.RedisURLs = urls
	}

	// Credentials, applied to any URL that does not embed its own. These are
	// environment-only on purpose; see the Config field documentation.
	if username := config.GetString("LOCK_REDIS_USERNAME"); username != "" {
		cfg.RedisUsername = username
	}
	if password := config.GetString("LOCK_REDIS_PASSWORD"); password != "" {
		cfg.RedisPassword = password
	}

	// Lock default settings
	if expiry := config.GetDuration("LOCK_DEFAULT_EXPIRY"); expiry != 0 {
		cfg.DefaultExpiry = expiry
	}
	if tries := config.GetInt("LOCK_DEFAULT_TRIES"); tries != 0 {
		cfg.DefaultTries = tries
	}
	if delay := config.GetDuration("LOCK_DEFAULT_RETRY_DELAY"); delay != 0 {
		cfg.DefaultRetryDelay = delay
	}
	if drift := config.GetFloat64("LOCK_DEFAULT_DRIFT_FACTOR"); drift != 0 {
		cfg.DefaultDriftFactor = drift
	}
	if prefix := config.GetString("LOCK_KEY_PREFIX"); prefix != "" {
		cfg.DefaultKeyPrefix = prefix
	}

	// Pool settings
	if poolSize := config.GetInt("LOCK_POOL_SIZE"); poolSize > 0 {
		cfg.PoolSize = poolSize
	}

	// TLS settings
	if config.GetBool("LOCK_TLS_INSECURE_SKIP_VERIFY") {
		cfg.TLSInsecureSkipVerify = true
	}

	return cfg
}
