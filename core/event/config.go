package event

import (
	"time"

	"go.oease.dev/goe/v2/contract"
)

// Config holds the event system configuration
type Config struct {
	// Redis connection settings
	RedisURL      string   // Redis connection URL (redis://username:password@host:port/database)
	RedisHosts    []string // List of Redis hosts (fallback if URL not provided)
	RedisUsername string   // Redis username (fallback if URL not provided)
	RedisPassword string   // Redis password (fallback if URL not provided)
	RedisDB       int      // Redis database number (fallback if URL not provided)

	// Legacy fields for backward compatibility
	RedisAddr string // Deprecated: use RedisURL or RedisHosts

	// Consumer settings
	ConsumerTimeout      time.Duration
	MaxRetries           int
	RetryBackoff         time.Duration
	DeadLetterQueueTTL   time.Duration
	StaleConsumerTimeout time.Duration

	// Performance settings
	BatchSize          int
	MaxPendingMessages int
	ClaimMinIdleTime   time.Duration
	ClaimInterval      time.Duration

	// Delayed queue settings
	DelayedQueueEnabled       bool
	DelayedQueueCheckInterval time.Duration
}

// DefaultConfig returns the default event configuration
func DefaultConfig() *Config {
	return &Config{
		RedisHosts:                []string{"localhost:6379"},
		RedisUsername:             "",
		RedisPassword:             "",
		RedisDB:                   0,
		RedisAddr:                 "localhost:6379", // Legacy field
		ConsumerTimeout:           30 * time.Second,
		MaxRetries:                3,
		RetryBackoff:              time.Second,
		DeadLetterQueueTTL:        24 * time.Hour,
		StaleConsumerTimeout:      5 * time.Minute,
		BatchSize:                 10,
		MaxPendingMessages:        1000,
		ClaimMinIdleTime:          time.Minute,
		ClaimInterval:             30 * time.Second,
		DelayedQueueEnabled:       true,
		DelayedQueueCheckInterval: time.Second,
	}
}

// LoadConfig loads configuration from GOE config
func LoadConfig(config contract.Config) *Config {
	cfg := DefaultConfig()

	// Redis connection - priority: URL > individual parameters > legacy addr
	if url := config.GetString("EVENT_REDIS_URL"); url != "" {
		cfg.RedisURL = url
	} else {
		// Use individual parameters
		if hosts := config.GetStringSlice("EVENT_REDIS_HOSTS"); len(hosts) > 0 {
			cfg.RedisHosts = hosts
		}
		if username := config.GetString("EVENT_REDIS_USERNAME"); username != "" {
			cfg.RedisUsername = username
		}
		if password := config.GetString("EVENT_REDIS_PASSWORD"); password != "" {
			cfg.RedisPassword = password
		}
		if db := config.GetInt("EVENT_REDIS_DB"); db != 0 {
			cfg.RedisDB = db
		}

		// Legacy support - fallback to EVENT_REDIS_ADDR if no hosts specified
		if len(cfg.RedisHosts) == 0 {
			if addr := config.GetString("EVENT_REDIS_ADDR"); addr != "" {
				cfg.RedisAddr = addr
				cfg.RedisHosts = []string{addr}
			}
		}
	}

	// Consumer settings
	if timeout := config.GetDuration("EVENT_CONSUMER_TIMEOUT"); timeout != 0 {
		cfg.ConsumerTimeout = timeout
	}
	if maxRetries := config.GetInt("EVENT_MAX_RETRIES"); maxRetries != 0 {
		cfg.MaxRetries = maxRetries
	}
	if backoff := config.GetDuration("EVENT_RETRY_BACKOFF"); backoff != 0 {
		cfg.RetryBackoff = backoff
	}
	if ttl := config.GetDuration("EVENT_DLQ_TTL"); ttl != 0 {
		cfg.DeadLetterQueueTTL = ttl
	}
	if staleTimeout := config.GetDuration("EVENT_STALE_CONSUMER_TIMEOUT"); staleTimeout != 0 {
		cfg.StaleConsumerTimeout = staleTimeout
	}

	// Performance settings
	if batchSize := config.GetInt("EVENT_BATCH_SIZE"); batchSize != 0 {
		cfg.BatchSize = batchSize
	}
	if maxPending := config.GetInt("EVENT_MAX_PENDING_MESSAGES"); maxPending != 0 {
		cfg.MaxPendingMessages = maxPending
	}
	if claimMinIdle := config.GetDuration("EVENT_CLAIM_MIN_IDLE_TIME"); claimMinIdle != 0 {
		cfg.ClaimMinIdleTime = claimMinIdle
	}
	if claimInterval := config.GetDuration("EVENT_CLAIM_INTERVAL"); claimInterval != 0 {
		cfg.ClaimInterval = claimInterval
	}

	// Delayed queue settings
	if delayedEnabled := config.GetBool("EVENT_DELAYED_QUEUE_ENABLED"); !delayedEnabled {
		cfg.DelayedQueueEnabled = delayedEnabled
	}
	if delayedInterval := config.GetDuration("EVENT_DELAYED_QUEUE_CHECK_INTERVAL"); delayedInterval != 0 {
		cfg.DelayedQueueCheckInterval = delayedInterval
	}

	return cfg
}
