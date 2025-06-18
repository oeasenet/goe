package cache

import (
	"context"

	"go.oease.dev/goe/v2/contract"
)

// Module represents the cache module for Fx
type Module struct {
	manager contract.CacheManager
	logger  contract.Logger
	config  contract.Config
}

// NewModule creates a new cache module
func NewModule(config contract.Config, logger contract.Logger) *Module {
	// Create cache manager
	manager := NewManager(config)

	// Register all built-in drivers (Fiber storage drivers)
	RegisterBuiltinDrivers(manager)

	return &Module{
		manager: manager,
		logger:  logger,
		config:  config,
	}
}

// Name returns the module name
func (m *Module) Name() string {
	return "cache"
}

// OnStart is called when the module starts
func (m *Module) OnStart(ctx context.Context) error {
	m.logger.Info("Cache module started",
		"driver", m.manager.Driver(),
		"store", m.config.GetString("CACHE_STORE"),
	)
	return nil
}

// OnStop is called when the module stops
func (m *Module) OnStop(ctx context.Context) error {
	m.logger.Info("Cache module stopped")

	// Close all stores
	// This is handled by each store's Close method

	return nil
}

// Provide returns the cache manager instance for Fx
func (m *Module) Provide() contract.CacheManager {
	return m.manager
}

// ProvideCache returns the default cache instance for Fx
func (m *Module) ProvideCache() contract.Cache {
	return m.manager.Store()
}
