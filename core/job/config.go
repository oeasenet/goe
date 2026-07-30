package job

import (
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.oease.dev/goe/v2/contract"
)

// Config holds the job system configuration
type Config struct {
	// Redis connection settings
	RedisURL      string   // Redis connection URL (redis://username:password@host:port/database)
	RedisHosts    []string // List of Redis hosts (fallback if URL not provided)
	RedisUsername string   // Redis username (fallback if URL not provided)
	RedisPassword string   // Redis password (fallback if URL not provided)
	RedisDB       int      // Redis database number (fallback if URL not provided)
	RedisPoolSize int      // Redis connection pool size

	// Legacy field for backward compatibility
	RedisAddr string // Deprecated: use RedisURL or RedisHosts

	// Key prefix for all job-related keys
	KeyPrefix string

	// Worker settings
	Concurrency       int           // Number of concurrent workers per queue
	MaxConcurrency    int           // Maximum total concurrent workers
	PollInterval      time.Duration // How often to poll for new jobs
	ShutdownTimeout   time.Duration // Timeout for graceful shutdown
	HeartbeatInterval time.Duration // Worker heartbeat interval

	// Job settings
	DefaultQueue       string        // Default queue name
	DefaultMaxAttempts int           // Default maximum retry attempts
	DefaultTimeout     time.Duration // Default job execution timeout
	RetryBackoff       time.Duration // Initial retry backoff duration
	MaxRetryBackoff    time.Duration // Maximum retry backoff duration
	RetryBackoffFactor float64       // Retry backoff multiplier

	// Scheduler settings
	SchedulerEnabled  bool          // Enable the scheduler
	SchedulerInterval time.Duration // How often to check for scheduled jobs

	// Uniqueness settings
	DefaultUniqueTTL time.Duration // Default TTL for job uniqueness

	// Dead letter queue settings
	DLQEnabled bool          // Enable dead letter queue
	DLQTTL     time.Duration // How long to keep jobs in DLQ

	// Metrics settings
	MetricsEnabled bool // Enable job metrics collection
}

// DefaultConfig returns the default job configuration
func DefaultConfig() *Config {
	return &Config{
		RedisHosts:         []string{"localhost:6379"},
		RedisUsername:      "",
		RedisPassword:      "",
		RedisDB:            0,
		RedisPoolSize:      10,
		RedisAddr:          "localhost:6379", // Legacy field
		KeyPrefix:          "goe:job:",
		Concurrency:        5,
		MaxConcurrency:     100,
		PollInterval:       time.Second,
		ShutdownTimeout:    30 * time.Second,
		HeartbeatInterval:  15 * time.Second,
		DefaultQueue:       "default",
		DefaultMaxAttempts: 3,
		DefaultTimeout:     30 * time.Minute,
		RetryBackoff:       time.Second,
		MaxRetryBackoff:    5 * time.Minute,
		RetryBackoffFactor: 2.0,
		SchedulerEnabled:   true,
		SchedulerInterval:  time.Second,
		DefaultUniqueTTL:   time.Hour,
		DLQEnabled:         true,
		DLQTTL:             7 * 24 * time.Hour, // 7 days
		MetricsEnabled:     true,
	}
}

// RedactedTarget returns the Redis endpoint the manager actually connects to,
// safe for logging. Priority mirrors NewManager: URL, then hosts, then the
// legacy addr. The URL form is reduced to host:port so credentials embedded in
// JOB_REDIS_URL never reach the logs; an unparseable URL is reported as such
// rather than echoed, for the same reason.
func (c *Config) RedactedTarget() string {
	if c.RedisURL != "" {
		opt, err := redis.ParseURL(c.RedisURL)
		if err != nil {
			return "<invalid JOB_REDIS_URL>"
		}
		return opt.Addr
	}
	if len(c.RedisHosts) > 0 {
		return strings.Join(c.RedisHosts, ",")
	}
	return c.RedisAddr
}

// LoadConfig loads configuration from GOE config
func LoadConfig(config contract.Config) *Config {
	cfg := DefaultConfig()

	// Redis connection - priority: URL > individual parameters > legacy addr
	if url := config.GetString("JOB_REDIS_URL"); url != "" {
		cfg.RedisURL = url
	} else {
		// Use individual parameters
		if hosts := config.GetStringSlice("JOB_REDIS_HOSTS"); len(hosts) > 0 {
			cfg.RedisHosts = hosts
		}
		if username := config.GetString("JOB_REDIS_USERNAME"); username != "" {
			cfg.RedisUsername = username
		}
		if password := config.GetString("JOB_REDIS_PASSWORD"); password != "" {
			cfg.RedisPassword = password
		}
		if db := config.GetInt("JOB_REDIS_DB"); db != 0 {
			cfg.RedisDB = db
		}

		// Legacy support - fallback to JOB_REDIS_ADDR if no hosts specified
		if len(cfg.RedisHosts) == 0 || (len(cfg.RedisHosts) == 1 && cfg.RedisHosts[0] == "localhost:6379") {
			if addr := config.GetString("JOB_REDIS_ADDR"); addr != "" {
				cfg.RedisAddr = addr
				cfg.RedisHosts = []string{addr}
			}
		}
	}

	if poolSize := config.GetInt("JOB_REDIS_POOL_SIZE"); poolSize > 0 {
		cfg.RedisPoolSize = poolSize
	}

	// Key prefix
	if prefix := config.GetString("JOB_KEY_PREFIX"); prefix != "" {
		cfg.KeyPrefix = prefix
	}

	// Worker settings
	if concurrency := config.GetInt("JOB_CONCURRENCY"); concurrency > 0 {
		cfg.Concurrency = concurrency
	}
	if maxConcurrency := config.GetInt("JOB_MAX_CONCURRENCY"); maxConcurrency > 0 {
		cfg.MaxConcurrency = maxConcurrency
	}
	if pollInterval := config.GetDuration("JOB_POLL_INTERVAL"); pollInterval > 0 {
		cfg.PollInterval = pollInterval
	}
	if shutdownTimeout := config.GetDuration("JOB_SHUTDOWN_TIMEOUT"); shutdownTimeout > 0 {
		cfg.ShutdownTimeout = shutdownTimeout
	}
	if heartbeatInterval := config.GetDuration("JOB_HEARTBEAT_INTERVAL"); heartbeatInterval > 0 {
		cfg.HeartbeatInterval = heartbeatInterval
	}

	// Job settings
	if defaultQueue := config.GetString("JOB_DEFAULT_QUEUE"); defaultQueue != "" {
		cfg.DefaultQueue = defaultQueue
	}
	if defaultMaxAttempts := config.GetInt("JOB_DEFAULT_MAX_ATTEMPTS"); defaultMaxAttempts > 0 {
		cfg.DefaultMaxAttempts = defaultMaxAttempts
	}
	if defaultTimeout := config.GetDuration("JOB_DEFAULT_TIMEOUT"); defaultTimeout > 0 {
		cfg.DefaultTimeout = defaultTimeout
	}
	if retryBackoff := config.GetDuration("JOB_RETRY_BACKOFF"); retryBackoff > 0 {
		cfg.RetryBackoff = retryBackoff
	}
	if maxRetryBackoff := config.GetDuration("JOB_MAX_RETRY_BACKOFF"); maxRetryBackoff > 0 {
		cfg.MaxRetryBackoff = maxRetryBackoff
	}
	if retryBackoffFactor := config.GetFloat64("JOB_RETRY_BACKOFF_FACTOR"); retryBackoffFactor > 0 {
		cfg.RetryBackoffFactor = retryBackoffFactor
	}

	// Scheduler settings
	if !config.GetBool("JOB_SCHEDULER_ENABLED") && config.Has("JOB_SCHEDULER_ENABLED") {
		cfg.SchedulerEnabled = false
	}
	if schedulerInterval := config.GetDuration("JOB_SCHEDULER_INTERVAL"); schedulerInterval > 0 {
		cfg.SchedulerInterval = schedulerInterval
	}

	// Uniqueness settings
	if uniqueTTL := config.GetDuration("JOB_DEFAULT_UNIQUE_TTL"); uniqueTTL > 0 {
		cfg.DefaultUniqueTTL = uniqueTTL
	}

	// Dead letter queue settings
	if !config.GetBool("JOB_DLQ_ENABLED") && config.Has("JOB_DLQ_ENABLED") {
		cfg.DLQEnabled = false
	}
	if dlqTTL := config.GetDuration("JOB_DLQ_TTL"); dlqTTL > 0 {
		cfg.DLQTTL = dlqTTL
	}

	// Metrics settings
	if !config.GetBool("JOB_METRICS_ENABLED") && config.Has("JOB_METRICS_ENABLED") {
		cfg.MetricsEnabled = false
	}

	return cfg
}
