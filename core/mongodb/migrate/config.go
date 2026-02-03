package migrate

import (
	"time"

	"go.oease.dev/goe/v2/contract"
)

// Config holds the migration system configuration
type Config struct {
	// Collection is the name of the collection used to track migration state
	Collection string

	// Timeout is the maximum duration for a single migration
	Timeout time.Duration

	// LockTimeout is the maximum duration to hold the migration lock
	LockTimeout time.Duration

	// LockHeartbeat is the interval for lock heartbeat updates
	LockHeartbeat time.Duration

	// UseTransactions enables transaction support for migrations
	// Requires MongoDB replica set
	UseTransactions bool

	// VerifyChecksums enables checksum verification on startup
	VerifyChecksums bool

	// AutoMigrate enables automatic migration on application start
	AutoMigrate bool

	// VersionScheme determines how versions are generated
	// "sequential" - 1, 2, 3... (default)
	// "timestamp" - Unix timestamp based
	VersionScheme string

	// SchemaVersionField is the field name for document-level schema versioning
	SchemaVersionField string

	// DryRunByDefault enables dry-run mode by default
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

// LoadConfig loads configuration from GOE config
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
