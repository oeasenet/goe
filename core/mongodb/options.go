package mongodb

import (
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/event"
)

// Option configures the MongoDB module from Go code.
//
// Options are applied after environment variables, so a value set in code
// always wins over the matching MONGO_* variable. Environment variables still
// supply everything code does not set, which means an application that passes
// no options behaves exactly as it did before.
//
// The naming rule is mechanical: the environment key with the MONGO_ prefix
// dropped and the rest CamelCased — MONGO_MAX_POOL_SIZE becomes
// WithMaxPoolSize.
//
// Credentials are the deliberate exception. MONGO_URI (which can embed
// user:pass), MONGO_USERNAME and MONGO_PASSWORD have no option: secrets
// belong in the environment, not in source. The two compose — options tune
// pools and databases in code while the environment supplies the URI and,
// when the URI carries no userinfo, the separate credential variables.
//
// Options are applied in the order given, so the last write wins. If any
// option returns an error, every option is discarded, the environment-only
// configuration stays in effect, and Module.ValidateConfig reports all
// collected errors, which fails application startup before any connection is
// served.
type Option func(*settings) error

// settings is the mutable state options write to. It is deliberately
// unexported: options are the only way in, which is what keeps credential
// keys out of reach of code configuration.
type settings struct {
	overrides map[string]any

	// monitor is the custom command monitor; a live object, so it cannot
	// travel through the configuration overlay.
	monitor *event.CommandMonitor
}

// resolveSettings applies opts and returns the configuration overlay plus the
// custom monitor. It reports every option error rather than stopping at the
// first, so a developer sees all of their mistakes in one startup failure
// instead of one per run.
func resolveSettings(opts []Option) (map[string]any, *event.CommandMonitor, []error) {
	s := &settings{overrides: make(map[string]any)}

	var errs []error
	for i, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(s); err != nil {
			errs = append(errs, fmt.Errorf("mongodb option %d: %w", i+1, err))
		}
	}

	return s.overrides, s.monitor, errs
}

// WithDatabase sets the database name. Replaces MONGO_DB_NAME.
func WithDatabase(name string) Option {
	return func(s *settings) error {
		if name == "" {
			return errors.New("WithDatabase: name must not be empty")
		}
		s.overrides["MONGO_DB_NAME"] = name
		return nil
	}
}

// WithMinPoolSize sets the minimum connection pool size. Replaces
// MONGO_MIN_POOL_SIZE.
func WithMinPoolSize(size int) Option {
	return func(s *settings) error {
		if size < 0 {
			return fmt.Errorf("WithMinPoolSize: size %d must not be negative", size)
		}
		s.overrides["MONGO_MIN_POOL_SIZE"] = size
		return nil
	}
}

// WithMaxPoolSize sets the maximum connection pool size. Replaces
// MONGO_MAX_POOL_SIZE.
func WithMaxPoolSize(size int) Option {
	return func(s *settings) error {
		if size <= 0 {
			return fmt.Errorf("WithMaxPoolSize: size %d must be positive", size)
		}
		s.overrides["MONGO_MAX_POOL_SIZE"] = size
		return nil
	}
}

// WithMaxConnIdleTime sets how long a pooled connection may sit idle.
// Replaces MONGO_MAX_CONN_IDLE_TIME.
func WithMaxConnIdleTime(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithMaxConnIdleTime: %v must be positive", d)
		}
		s.overrides["MONGO_MAX_CONN_IDLE_TIME"] = d
		return nil
	}
}

// WithDebug enables command logging. Replaces MONGO_DEBUG.
func WithDebug(enabled bool) Option {
	return func(s *settings) error {
		s.overrides["MONGO_DEBUG"] = enabled
		return nil
	}
}

// WithPingTimeout sets the startup reachability ping timeout. Replaces
// MONGO_PING_TIMEOUT. Defaults to 5s.
func WithPingTimeout(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithPingTimeout: %v must be positive", d)
		}
		s.overrides["MONGO_PING_TIMEOUT"] = d
		return nil
	}
}

// WithCommandMonitor installs a custom command monitor. It wins over the debug monitor that MONGO_DEBUG/WithDebug would install.
// There is no environment equivalent — a monitor is code.
func WithCommandMonitor(monitor *event.CommandMonitor) Option {
	return func(s *settings) error {
		if monitor == nil {
			return errors.New("WithCommandMonitor: monitor must not be nil")
		}
		s.monitor = monitor
		return nil
	}
}
