package job

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"go.oease.dev/goe/v2/contract"
)

// Option configures the job system from Go code.
//
// Options are applied after environment variables, so a value set in code
// always wins over the matching JOB_* variable. Environment variables still
// supply everything code does not set, which means an application that passes
// no options behaves exactly as it did before.
//
// The naming rule is mechanical: every field of Config is exposed as
// With<FieldName>. Read the Config documentation, prepend "With", and that is
// the option.
//
// Credentials are the deliberate exception. JOB_REDIS_URL (which can embed
// user:pass), JOB_REDIS_USERNAME and JOB_REDIS_PASSWORD have no option:
// secrets belong in the environment, not in source. The two compose —
// WithRedisHosts in code picks the endpoint while JOB_REDIS_USERNAME and
// JOB_REDIS_PASSWORD from the environment supply the credentials.
//
// Options are applied in the order given, so the last write wins. If any
// option returns an error, the module is not created and goe.New aborts
// startup with every collected error, before any Redis connection is
// attempted.
type Option func(*settings) error

// settings is the mutable state options write to. It is deliberately
// unexported: options are the only way in, which is what keeps credential
// fields out of reach of code configuration.
type settings struct {
	cfg *Config

	// connectionSetInCode records that WithRedisHosts supplied the endpoint,
	// which satisfies ValidateConfig's "connection explicitly configured"
	// requirement the same way a JOB_REDIS_* variable would.
	connectionSetInCode bool
}

// resolveConfig runs the full pipeline: defaults -> environment -> options.
// It reports every option error rather than stopping at the first, so a
// developer sees all of their mistakes in one startup failure instead of one
// per run.
func resolveConfig(config contract.Config, opts []Option) (*Config, bool, []error) {
	s := &settings{cfg: LoadConfig(config)}

	var errs []error
	for i, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(s); err != nil {
			errs = append(errs, fmt.Errorf("job option %d: %w", i+1, err))
		}
	}

	return s.cfg, s.connectionSetInCode, errs
}

// WithRedisHosts sets the Redis endpoints the job system connects to.
// Replaces JOB_REDIS_HOSTS, and overrides JOB_REDIS_URL when both are
// present, because code wins over the environment.
//
// Credentials never ride along here: JOB_REDIS_USERNAME and
// JOB_REDIS_PASSWORD from the environment are still applied to this
// connection.
func WithRedisHosts(hosts ...string) Option {
	return func(s *settings) error {
		if len(hosts) == 0 {
			return errors.New("WithRedisHosts: at least one host is required")
		}
		if slices.Contains(hosts, "") {
			return errors.New("WithRedisHosts: host must not be empty")
		}
		s.cfg.RedisHosts = hosts
		s.cfg.RedisAddr = hosts[0]
		// The manager prefers RedisURL over RedisHosts; an environment URL
		// must not silently beat an endpoint chosen in code.
		s.cfg.RedisURL = ""
		s.connectionSetInCode = true
		return nil
	}
}

// WithRedisDB sets the Redis database number. Replaces JOB_REDIS_DB.
func WithRedisDB(db int) Option {
	return func(s *settings) error {
		if db < 0 {
			return fmt.Errorf("WithRedisDB: database %d must not be negative", db)
		}
		s.cfg.RedisDB = db
		return nil
	}
}

// WithRedisPoolSize sets the Redis connection pool size. Replaces
// JOB_REDIS_POOL_SIZE.
func WithRedisPoolSize(size int) Option {
	return func(s *settings) error {
		if size <= 0 {
			return fmt.Errorf("WithRedisPoolSize: size %d must be positive", size)
		}
		s.cfg.RedisPoolSize = size
		return nil
	}
}

// WithKeyPrefix sets the prefix for all job-related Redis keys. Replaces
// JOB_KEY_PREFIX.
func WithKeyPrefix(prefix string) Option {
	return func(s *settings) error {
		if prefix == "" {
			return errors.New("WithKeyPrefix: prefix must not be empty")
		}
		s.cfg.KeyPrefix = prefix
		return nil
	}
}

// WithConcurrency sets the number of concurrent workers per queue. Replaces
// JOB_CONCURRENCY.
func WithConcurrency(n int) Option {
	return func(s *settings) error {
		if n <= 0 {
			return fmt.Errorf("WithConcurrency: %d must be positive", n)
		}
		s.cfg.Concurrency = n
		return nil
	}
}

// WithMaxConcurrency sets the maximum total concurrent workers across all
// queues. Replaces JOB_MAX_CONCURRENCY.
func WithMaxConcurrency(n int) Option {
	return func(s *settings) error {
		if n <= 0 {
			return fmt.Errorf("WithMaxConcurrency: %d must be positive", n)
		}
		s.cfg.MaxConcurrency = n
		return nil
	}
}

// WithPollInterval sets how often workers poll for new jobs. Replaces
// JOB_POLL_INTERVAL.
func WithPollInterval(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithPollInterval: %v must be positive", d)
		}
		s.cfg.PollInterval = d
		return nil
	}
}

// WithShutdownTimeout sets the timeout for graceful worker shutdown. Replaces
// JOB_SHUTDOWN_TIMEOUT.
func WithShutdownTimeout(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithShutdownTimeout: %v must be positive", d)
		}
		s.cfg.ShutdownTimeout = d
		return nil
	}
}

// WithHeartbeatInterval sets the worker heartbeat interval. Replaces
// JOB_HEARTBEAT_INTERVAL.
func WithHeartbeatInterval(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithHeartbeatInterval: %v must be positive", d)
		}
		s.cfg.HeartbeatInterval = d
		return nil
	}
}

// WithDefaultQueue sets the queue jobs are dispatched to when none is named.
// Replaces JOB_DEFAULT_QUEUE.
func WithDefaultQueue(name string) Option {
	return func(s *settings) error {
		if name == "" {
			return errors.New("WithDefaultQueue: name must not be empty")
		}
		s.cfg.DefaultQueue = name
		return nil
	}
}

// WithDefaultMaxAttempts sets the default maximum retry attempts. Replaces
// JOB_DEFAULT_MAX_ATTEMPTS.
func WithDefaultMaxAttempts(n int) Option {
	return func(s *settings) error {
		if n <= 0 {
			return fmt.Errorf("WithDefaultMaxAttempts: %d must be positive", n)
		}
		s.cfg.DefaultMaxAttempts = n
		return nil
	}
}

// WithDefaultTimeout sets the default job execution timeout. Replaces
// JOB_DEFAULT_TIMEOUT.
func WithDefaultTimeout(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithDefaultTimeout: %v must be positive", d)
		}
		s.cfg.DefaultTimeout = d
		return nil
	}
}

// WithRetryBackoff sets the initial retry backoff duration. Replaces
// JOB_RETRY_BACKOFF.
func WithRetryBackoff(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithRetryBackoff: %v must be positive", d)
		}
		s.cfg.RetryBackoff = d
		return nil
	}
}

// WithMaxRetryBackoff sets the maximum retry backoff duration. Replaces
// JOB_MAX_RETRY_BACKOFF.
func WithMaxRetryBackoff(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithMaxRetryBackoff: %v must be positive", d)
		}
		s.cfg.MaxRetryBackoff = d
		return nil
	}
}

// WithRetryBackoffFactor sets the retry backoff multiplier. Replaces
// JOB_RETRY_BACKOFF_FACTOR.
func WithRetryBackoffFactor(f float64) Option {
	return func(s *settings) error {
		if f <= 0 {
			return fmt.Errorf("WithRetryBackoffFactor: %v must be positive", f)
		}
		s.cfg.RetryBackoffFactor = f
		return nil
	}
}

// WithSchedulerEnabled enables or disables the scheduler that promotes due
// jobs. Replaces JOB_SCHEDULER_ENABLED. Enabled by default.
func WithSchedulerEnabled(enabled bool) Option {
	return func(s *settings) error { s.cfg.SchedulerEnabled = enabled; return nil }
}

// WithSchedulerInterval sets how often the scheduler checks for due jobs.
// Replaces JOB_SCHEDULER_INTERVAL.
func WithSchedulerInterval(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithSchedulerInterval: %v must be positive", d)
		}
		s.cfg.SchedulerInterval = d
		return nil
	}
}

// WithDefaultUniqueTTL sets the default TTL for job uniqueness constraints.
// Replaces JOB_DEFAULT_UNIQUE_TTL.
func WithDefaultUniqueTTL(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithDefaultUniqueTTL: %v must be positive", d)
		}
		s.cfg.DefaultUniqueTTL = d
		return nil
	}
}

// WithDLQEnabled enables or disables the dead letter queue. Replaces
// JOB_DLQ_ENABLED. Enabled by default.
func WithDLQEnabled(enabled bool) Option {
	return func(s *settings) error { s.cfg.DLQEnabled = enabled; return nil }
}

// WithDLQTTL sets how long jobs are kept in the dead letter queue. Replaces
// JOB_DLQ_TTL.
func WithDLQTTL(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithDLQTTL: %v must be positive", d)
		}
		s.cfg.DLQTTL = d
		return nil
	}
}

// WithMetricsEnabled enables or disables job metrics collection. Replaces
// JOB_METRICS_ENABLED. Enabled by default.
func WithMetricsEnabled(enabled bool) Option {
	return func(s *settings) error { s.cfg.MetricsEnabled = enabled; return nil }
}
