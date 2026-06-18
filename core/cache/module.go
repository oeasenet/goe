package cache

import (
	"context"
	"fmt"

	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/internal/configvalidator"
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
		logger:  logger.With("module", "cache"),
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
	// Close all initialized cache stores to release connections
	mgr := m.manager.(*manager)
	mgr.mu.RLock()
	stores := make(map[string]contract.Cache, len(mgr.stores))
	for k, v := range mgr.stores {
		stores[k] = v
	}
	mgr.mu.RUnlock()

	for name, store := range stores {
		if err := store.Store().Close(); err != nil {
			m.logger.Error("Error closing cache store", "store", name, "error", err)
		}
	}

	m.logger.Info("Cache module stopped", "stores_closed", len(stores))
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

// ValidateConfig validates the cache module configuration
func (m *Module) ValidateConfig() error {
	v := configvalidator.NewConfigValidator(m.config, "cache")

	// Cache store is optional, defaults to memory
	store := m.config.GetString("CACHE_STORE")
	if store != "" {
		// Only memory and redis have registered drivers (see RegisterBuiltinDrivers).
		validStores := []string{"memory", "redis"}
		v.Optional("CACHE_STORE", "Cache store type", configvalidator.ValidateOneOf(validStores...))

		// Redis works with localhost defaults; validate the real keys only when set.
		if store == "redis" {
			if m.config.Has("CACHE_REDIS_PORT") {
				v.Optional("CACHE_REDIS_PORT", "Redis port", configvalidator.ValidatePort)
			}
			if m.config.Has("CACHE_REDIS_DATABASE") {
				v.Optional("CACHE_REDIS_DATABASE", "Redis database number", configvalidator.ValidateNonNegativeInt)
			}
		}
	}

	// TTL is optional but should be positive if set
	if m.config.Has("CACHE_TTL") {
		v.Optional("CACHE_TTL", "Default cache TTL", func(value any) error {
			duration := m.config.GetDuration("CACHE_TTL")
			if duration < 0 {
				return fmt.Errorf("TTL must be positive")
			}
			return nil
		})
	}

	return v.Validate()
}
