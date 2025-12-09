package lock

import (
	"context"

	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/internal/configvalidator"
)

// Module represents the lock module for Fx dependency injection.
type Module struct {
	manager contract.LockManager
	logger  contract.Logger
	config  contract.Config
}

// NewModule creates a new lock module.
func NewModule(config contract.Config, logger contract.Logger) (*Module, error) {
	// Load lock configuration
	lockConfig := LoadConfig(config)

	// Create manager
	manager, err := NewManager(lockConfig, logger)
	if err != nil {
		return nil, err
	}

	return &Module{
		manager: manager,
		logger:  logger,
		config:  config,
	}, nil
}

// Name returns the module name.
func (m *Module) Name() string {
	return "lock"
}

// OnStart is called when the module starts.
func (m *Module) OnStart(ctx context.Context) error {
	// Test connection
	if err := m.manager.Health(ctx); err != nil {
		return err
	}

	m.logger.Info("Lock module started",
		"key_prefix", m.config.GetString("LOCK_KEY_PREFIX"),
	)

	return nil
}

// OnStop is called when the module stops.
func (m *Module) OnStop(ctx context.Context) error {
	m.logger.Info("Lock module stopping")

	// Close lock manager
	if err := m.manager.Close(ctx); err != nil {
		m.logger.Error("Error closing lock manager", "error", err)
		return err
	}

	m.logger.Info("Lock module stopped")
	return nil
}

// Provide returns the lock manager instance for Fx.
func (m *Module) Provide() contract.LockManager {
	return m.manager
}

// ProvideLockManager returns the lock manager instance for Fx (alias for Provide).
func (m *Module) ProvideLockManager() contract.LockManager {
	return m.manager
}

// ValidateConfig validates the lock module configuration.
func (m *Module) ValidateConfig() error {
	v := configvalidator.NewConfigValidator(m.config, "lock")

	// Require at least one Redis configuration
	hasLockConfig := m.config.Has("LOCK_REDIS_URL") ||
		m.config.Has("LOCK_REDIS_ADDR") ||
		m.config.Has("LOCK_REDIS_HOST") ||
		m.config.Has("LOCK_REDIS_HOSTS")

	if !hasLockConfig {
		v.RequireWithValidator("LOCK_REDIS_ADDR", "Redis server address for lock system", configvalidator.ValidateHostPort)
	}

	// Validate LOCK_* specific settings if present
	if m.config.Has("LOCK_REDIS_ADDR") {
		v.Optional("LOCK_REDIS_ADDR", "Redis server address", configvalidator.ValidateHostPort)
	}

	if m.config.Has("LOCK_REDIS_DB") {
		v.Optional("LOCK_REDIS_DB", "Redis database number", configvalidator.ValidatePositiveInt)
	}

	if m.config.Has("LOCK_REDIS_POOL_SIZE") {
		v.Optional("LOCK_REDIS_POOL_SIZE", "Redis pool size", configvalidator.ValidatePositiveInt)
	}

	if m.config.Has("LOCK_DEFAULT_TRIES") {
		v.Optional("LOCK_DEFAULT_TRIES", "Default retry attempts", configvalidator.ValidatePositiveInt)
	}

	return v.Validate()
}
