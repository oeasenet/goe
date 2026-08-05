package migrate

import (
	"errors"
	"fmt"
	"time"

	"go.oease.dev/goe/v2/contract"
)

// Option configures the migration module from Go code.
//
// Options are applied after environment variables, so a value set in code
// always wins over the matching MONGODB_MIGRATE_* variable. Environment
// variables still supply everything code does not set, which means an
// application that passes no options behaves exactly as it did before.
//
// The naming rule is mechanical: every field of Config is exposed as
// With<FieldName> — MONGODB_MIGRATE_LOCK_TIMEOUT is WithLockTimeout,
// MONGODB_MIGRATE_DRY_RUN is WithDryRunByDefault. (These are module options;
// the lower-level MigratorOption family — WithLogger, WithConfig, WithDryRun,
// WithHostname — keeps configuring a hand-built Migrator as before.)
//
// Options are applied in the order given, so the last write wins. If any
// option returns an error, the module is not created and goe.New aborts
// startup with every collected error.
type Option func(*settings) error

// settings is the mutable state options write to. It is deliberately
// unexported: options are the only way in.
type settings struct {
	cfg *Config
}

// resolveConfig runs the full pipeline: defaults -> environment -> options.
// It reports every option error rather than stopping at the first, so a
// developer sees all of their mistakes in one startup failure instead of one
// per run.
func resolveConfig(config contract.Config, opts []Option) (*Config, []error) {
	s := &settings{cfg: LoadConfig(config)}

	var errs []error
	for i, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(s); err != nil {
			errs = append(errs, fmt.Errorf("migrate option %d: %w", i+1, err))
		}
	}

	return s.cfg, errs
}

// WithCollection sets the collection used to track migration state. Replaces
// MONGODB_MIGRATE_COLLECTION.
func WithCollection(name string) Option {
	return func(s *settings) error {
		if name == "" {
			return errors.New("WithCollection: name must not be empty")
		}
		s.cfg.Collection = name
		return nil
	}
}

// WithTimeout sets the maximum duration for a single migration execution.
// Replaces MONGODB_MIGRATE_TIMEOUT.
func WithTimeout(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithTimeout: %v must be positive", d)
		}
		s.cfg.Timeout = d
		return nil
	}
}

// WithLockTimeout sets the maximum duration to hold the distributed
// migration lock. Replaces MONGODB_MIGRATE_LOCK_TIMEOUT.
func WithLockTimeout(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithLockTimeout: %v must be positive", d)
		}
		s.cfg.LockTimeout = d
		return nil
	}
}

// WithLockHeartbeat sets the interval for lock heartbeat updates. Replaces
// MONGODB_MIGRATE_LOCK_HEARTBEAT.
func WithLockHeartbeat(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithLockHeartbeat: %v must be positive", d)
		}
		s.cfg.LockHeartbeat = d
		return nil
	}
}

// WithUseTransactions enables or disables running migrations inside
// transactions (requires a replica set). Replaces
// MONGODB_MIGRATE_USE_TRANSACTIONS. Enabled by default.
func WithUseTransactions(enabled bool) Option {
	return func(s *settings) error { s.cfg.UseTransactions = enabled; return nil }
}

// WithVerifyChecksums enables or disables checksum verification on startup,
// which detects migrations edited after being applied. Replaces
// MONGODB_MIGRATE_VERIFY_CHECKSUMS. Enabled by default.
func WithVerifyChecksums(enabled bool) Option {
	return func(s *settings) error { s.cfg.VerifyChecksums = enabled; return nil }
}

// WithAutoMigrate enables or disables applying pending migrations
// automatically during startup. Replaces MONGODB_MIGRATE_AUTO. Disabled by
// default.
func WithAutoMigrate(enabled bool) Option {
	return func(s *settings) error { s.cfg.AutoMigrate = enabled; return nil }
}

// WithVersionScheme sets how migration versions are interpreted:
// "sequential" (1, 2, 3...) or "timestamp" (Unix timestamps). Replaces
// MONGODB_MIGRATE_VERSION_SCHEME.
func WithVersionScheme(scheme string) Option {
	return func(s *settings) error {
		if scheme != "sequential" && scheme != "timestamp" {
			return fmt.Errorf("WithVersionScheme: %q must be \"sequential\" or \"timestamp\"", scheme)
		}
		s.cfg.VersionScheme = scheme
		return nil
	}
}

// WithSchemaVersionField sets the field name used for document-level schema
// versioning. Replaces MONGODB_MIGRATE_SCHEMA_VERSION_FIELD.
func WithSchemaVersionField(field string) Option {
	return func(s *settings) error {
		if field == "" {
			return errors.New("WithSchemaVersionField: field must not be empty")
		}
		s.cfg.SchemaVersionField = field
		return nil
	}
}

// WithDryRunByDefault enables dry-run mode by default, previewing changes
// without applying them. Replaces MONGODB_MIGRATE_DRY_RUN.
func WithDryRunByDefault(enabled bool) Option {
	return func(s *settings) error { s.cfg.DryRunByDefault = enabled; return nil }
}
