package migrate

import (
	"time"

	"go.oease.dev/goe/v2/contract"
)

// Config holds the migration system configuration.
// All fields can be configured via environment variables with the MONGODB_MIGRATE_ prefix.
type Config struct {
	// Collection is the name of the collection used to track migration state.
	// Env: MONGODB_MIGRATE_COLLECTION (default: _goe_migrations)
	Collection string

	// Timeout is the maximum duration for a single migration execution.
	// Env: MONGODB_MIGRATE_TIMEOUT (default: 5m)
	Timeout time.Duration

	// LockTimeout is the maximum duration to hold the distributed migration lock.
	// If a migration exceeds this duration, the lock may be stolen by another instance.
	// Env: MONGODB_MIGRATE_LOCK_TIMEOUT (default: 30m)
	LockTimeout time.Duration

	// LockHeartbeat is the interval for lock heartbeat updates.
	// The heartbeat extends the lock expiration to prevent premature release.
	// Env: MONGODB_MIGRATE_LOCK_HEARTBEAT (default: 30s)
	LockHeartbeat time.Duration

	// UseTransactions enables transaction support for migrations.
	// Requires MongoDB replica set. When enabled, migrations run within transactions.
	// Env: MONGODB_MIGRATE_USE_TRANSACTIONS (default: true)
	UseTransactions bool

	// VerifyChecksums enables checksum verification on startup.
	// Detects if migration code has been modified after being applied.
	// Env: MONGODB_MIGRATE_VERIFY_CHECKSUMS (default: true)
	VerifyChecksums bool

	// AutoMigrate enables automatic migration on application start.
	// When true, pending migrations are applied automatically during OnStart.
	// Env: MONGODB_MIGRATE_AUTO (default: false)
	AutoMigrate bool

	// VersionScheme determines how migration versions are interpreted.
	// "sequential" - versions are sequential integers: 1, 2, 3... (default)
	// "timestamp" - versions are Unix timestamps for ordering
	// Env: MONGODB_MIGRATE_VERSION_SCHEME (default: sequential)
	VersionScheme string

	// SchemaVersionField is the field name for document-level schema versioning.
	// This field is automatically updated by schema-changing helpers like AddField.
	// Env: MONGODB_MIGRATE_SCHEMA_VERSION_FIELD (default: _goe_sv)
	SchemaVersionField string

	// DryRunByDefault enables dry-run mode by default.
	// When true, migrations will only preview changes without applying them.
	// Env: MONGODB_MIGRATE_DRY_RUN (default: false)
	DryRunByDefault bool
}

// DefaultConfig returns the default migration configuration
func DefaultConfig() *Config {
	return &Config{
		Collection:         "_goe_migrations",
		Timeout:            5 * time.Minute,
		LockTimeout:        30 * time.Minute,
		LockHeartbeat:      30 * time.Second,
		UseTransactions:    true,
		VerifyChecksums:    true,
		AutoMigrate:        false,
		VersionScheme:      "sequential",
		SchemaVersionField: "_goe_sv",
		DryRunByDefault:    false,
	}
}

// LoadConfig loads configuration from GOE config using MONGODB_MIGRATE_* settings.
//
// Environment variables:
//   - MONGODB_MIGRATE_COLLECTION: State collection name (default: _goe_migrations)
//   - MONGODB_MIGRATE_TIMEOUT: Per-migration timeout (default: 5m)
//   - MONGODB_MIGRATE_LOCK_TIMEOUT: Distributed lock timeout (default: 30m)
//   - MONGODB_MIGRATE_LOCK_HEARTBEAT: Lock heartbeat interval (default: 30s)
//   - MONGODB_MIGRATE_USE_TRANSACTIONS: Use transactions if available (default: true)
//   - MONGODB_MIGRATE_VERIFY_CHECKSUMS: Verify checksums on startup (default: true)
//   - MONGODB_MIGRATE_AUTO: Auto-run pending migrations on app start (default: false)
//   - MONGODB_MIGRATE_VERSION_SCHEME: Version scheme - sequential or timestamp (default: sequential)
//   - MONGODB_MIGRATE_SCHEMA_VERSION_FIELD: Document schema version field (default: _goe_sv)
//   - MONGODB_MIGRATE_DRY_RUN: Enable dry-run mode by default (default: false)
func LoadConfig(config contract.Config) *Config {
	cfg := DefaultConfig()

	// Collection name
	if collection := config.GetString("MONGODB_MIGRATE_COLLECTION"); collection != "" {
		cfg.Collection = collection
	}

	// Timeout
	if timeout := config.GetDuration("MONGODB_MIGRATE_TIMEOUT"); timeout > 0 {
		cfg.Timeout = timeout
	}

	// Lock timeout
	if lockTimeout := config.GetDuration("MONGODB_MIGRATE_LOCK_TIMEOUT"); lockTimeout > 0 {
		cfg.LockTimeout = lockTimeout
	}

	// Lock heartbeat
	if lockHeartbeat := config.GetDuration("MONGODB_MIGRATE_LOCK_HEARTBEAT"); lockHeartbeat > 0 {
		cfg.LockHeartbeat = lockHeartbeat
	}

	// Use transactions - default true, explicit false to disable
	if config.Has("MONGODB_MIGRATE_USE_TRANSACTIONS") {
		cfg.UseTransactions = config.GetBool("MONGODB_MIGRATE_USE_TRANSACTIONS")
	}

	// Verify checksums - default true, explicit false to disable
	if config.Has("MONGODB_MIGRATE_VERIFY_CHECKSUMS") {
		cfg.VerifyChecksums = config.GetBool("MONGODB_MIGRATE_VERIFY_CHECKSUMS")
	}

	// Auto migrate - default false, explicit true to enable
	if config.Has("MONGODB_MIGRATE_AUTO") {
		cfg.AutoMigrate = config.GetBool("MONGODB_MIGRATE_AUTO")
	}

	// Version scheme
	if scheme := config.GetString("MONGODB_MIGRATE_VERSION_SCHEME"); scheme != "" {
		cfg.VersionScheme = scheme
	}

	// Schema version field
	if field := config.GetString("MONGODB_MIGRATE_SCHEMA_VERSION_FIELD"); field != "" {
		cfg.SchemaVersionField = field
	}

	// Dry run by default
	if config.Has("MONGODB_MIGRATE_DRY_RUN") {
		cfg.DryRunByDefault = config.GetBool("MONGODB_MIGRATE_DRY_RUN")
	}

	return cfg
}

// Validate validates the configuration and fills in defaults for missing values
func (c *Config) Validate() error {
	// Set defaults for empty values
	if c.Collection == "" {
		c.Collection = "_goe_migrations"
	}
	if c.Timeout <= 0 {
		c.Timeout = 5 * time.Minute
	}
	if c.LockTimeout <= 0 {
		c.LockTimeout = 30 * time.Minute
	}
	if c.LockHeartbeat <= 0 {
		c.LockHeartbeat = 30 * time.Second
	}
	if c.VersionScheme != "sequential" && c.VersionScheme != "timestamp" {
		c.VersionScheme = "sequential"
	}
	if c.SchemaVersionField == "" {
		c.SchemaVersionField = "_goe_sv"
	}
	return nil
}
