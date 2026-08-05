package migrate

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.oease.dev/goe/v2/contract"
)

// Module represents the migration module for Fx dependency injection
type Module struct {
	migrator *Migrator
	logger   contract.Logger
	config   contract.Config
	mongodb  contract.MongoDB
	cfg      *Config
}

// NewModule creates a new migration module.
// The migrator is not created here because MongoDB connections are established
// during the Fx OnStart lifecycle phase. The migrator is created in OnStart
// after the MongoDB connection is available.
//
// Configuration resolves in layers: defaults, then MONGODB_MIGRATE_*
// environment variables, then opts. Anything set through an Option wins over
// the environment. Option errors abort construction with every error
// reported at once.
func NewModule(config contract.Config, logger contract.Logger, mongodb contract.MongoDB, opts ...Option) (*Module, error) {
	// Resolve migration configuration: defaults -> environment -> code.
	cfg, optErrs := resolveConfig(config, opts)
	if len(optErrs) > 0 {
		return nil, errors.Join(optErrs...)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid migration configuration: %w", err)
	}

	return &Module{
		logger:  logger.With("module", "mongo-migrate"),
		config:  config,
		mongodb: mongodb,
		cfg:     cfg,
	}, nil
}

// Name returns the module name
func (m *Module) Name() string {
	return "mongodb_migrate"
}

// OnStart is called when the module starts.
// The migrator is created here because MongoDB connections are only available
// after the MongoDB module's OnStart has completed.
func (m *Module) OnStart(ctx context.Context) error {
	m.logger.Info("MongoDB migration module starting",
		"collection", m.cfg.Collection,
		"auto_migrate", m.cfg.AutoMigrate,
		"verify_checksums", m.cfg.VerifyChecksums,
	)

	// Get the database - now available because MongoDB OnStart has already run
	db := m.mongodb.DB()
	if db == nil {
		return fmt.Errorf("mongodb database is nil - ensure MongoDB module is started and connected")
	}

	// Create migrator now that DB is available
	m.migrator = NewMigrator(db,
		WithLogger(m.logger),
		WithConfig(m.cfg),
		WithDryRun(m.cfg.DryRunByDefault),
	)

	// Initialize the migrator (create collections, indexes)
	if err := m.migrator.Init(ctx); err != nil {
		m.logger.Error("Failed to initialize migrator", "error", err)
		return err
	}

	// Verify checksums if enabled
	if m.cfg.VerifyChecksums && HasMigrations() {
		m.logger.Debug("Verifying migration checksums")
		if err := m.migrator.VerifyChecksums(ctx); err != nil {
			m.logger.Error("Migration checksum verification failed", "error", err)
			return err
		}
		m.logger.Debug("Migration checksums verified")
	}

	// Run auto-migration if enabled
	if m.cfg.AutoMigrate && HasMigrations() {
		m.logger.Info("Running auto-migration")

		result, err := m.migrator.Up(ctx)
		if err != nil {
			m.logger.Error("Auto-migration failed", "error", err)
			return err
		}

		if len(result.Applied) > 0 {
			m.logger.Info("Auto-migration completed",
				"applied", len(result.Applied),
				"skipped", len(result.Skipped),
				"duration", result.Duration.String(),
			)
		} else {
			m.logger.Info("No pending migrations to apply")
		}
	}

	// Log current state
	version, err := m.migrator.Version(ctx)
	if err == nil {
		m.logger.Info("Current migration version", "version", version)
	}

	pending, err := m.migrator.Pending(ctx)
	if err == nil && len(pending) > 0 {
		m.logger.Info("Pending migrations", "count", len(pending))
	}

	m.logger.Info("MongoDB migration module started")
	return nil
}

// OnStop is called when the module stops
func (m *Module) OnStop(ctx context.Context) error {
	m.logger.Info("MongoDB migration module stopping")

	// Release any held locks (migrator may be nil if OnStart failed)
	if m.migrator != nil && m.migrator.lock.IsHeld() {
		if err := m.migrator.lock.Release(ctx); err != nil {
			m.logger.Error("Failed to release migration lock", "error", err)
		}
	}

	m.logger.Info("MongoDB migration module stopped")
	return nil
}

// Provide returns the Migrator instance for Fx
func (m *Module) Provide() *Migrator {
	return m.migrator
}

// ProvideMigrator returns the Migrator instance (alias for Provide)
func (m *Module) ProvideMigrator() *Migrator {
	return m.migrator
}

// Migrator returns the underlying migrator
func (m *Module) Migrator() *Migrator {
	return m.migrator
}

// Config returns the migration configuration
func (m *Module) Config() *Config {
	return m.cfg
}

// --- Helper functions for creating modules with different configurations ---

// NewModuleWithDB creates a migration module with a specific database
func NewModuleWithDB(config contract.Config, logger contract.Logger, db *mongo.Database) (*Module, error) {
	logger = logger.With("module", "mongo-migrate")
	cfg := LoadConfig(config)

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid migration configuration: %w", err)
	}

	migrator := NewMigrator(db,
		WithLogger(logger),
		WithConfig(cfg),
		WithDryRun(cfg.DryRunByDefault),
	)

	return &Module{
		migrator: migrator,
		logger:   logger,
		config:   config,
		cfg:      cfg,
	}, nil
}

// NewModuleWithMigrator creates a module with a pre-configured migrator
func NewModuleWithMigrator(migrator *Migrator, logger contract.Logger) *Module {
	return &Module{
		migrator: migrator,
		logger:   logger,
		cfg:      migrator.config,
	}
}
