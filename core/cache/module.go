package cache

import (
	"context"
	"errors"
	"fmt"

	"go.oease.dev/goe/v2/contract"
	goeconfig "go.oease.dev/goe/v2/core/config"
	"go.oease.dev/goe/v2/core/internal/configvalidator"
)

// Module represents the cache module for Fx
type Module struct {
	manager contract.CacheManager
	logger  contract.Logger

	// config is the effective configuration: the base config with the Option
	// overlay applied. The manager and every driver factory read through it,
	// so code-configured values behave exactly like environment variables.
	config contract.Config

	// optErrs holds failures from Option application. They are reported by
	// ValidateConfig so that startup aborts; the module fell back to the
	// environment-only configuration, which is never actually served.
	optErrs []error
}

// NewModule creates a new cache module.
//
// Configuration resolves in layers: CACHE_* environment variables, then opts.
// Anything set through an Option wins over the environment. If any option
// fails, every option is discarded, the environment-only configuration stays
// in effect, and ValidateConfig aborts startup with all collected errors.
func NewModule(config contract.Config, logger contract.Logger, opts ...Option) *Module {
	overrides, optErrs := resolveOverrides(opts)
	if len(optErrs) == 0 && len(overrides) > 0 {
		config = goeconfig.NewConfigWrapper(config, overrides)
	}

	// Create cache manager on the effective configuration
	manager := NewManager(config)

	// Register all built-in drivers (Fiber storage drivers)
	RegisterBuiltinDrivers(manager)

	return &Module{
		manager: manager,
		logger:  logger.With("module", "cache"),
		config:  config,
		optErrs: optErrs,
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
	// Option failures are reported first and abort startup. The module fell
	// back to the environment-only configuration when this happened, so
	// returning here guarantees a half-configured cache is never served.
	if len(m.optErrs) > 0 {
		return errors.Join(m.optErrs...)
	}

	v := configvalidator.NewConfigValidator(m.config, "cache")

	// Cache store is optional, defaults to memory
	store := m.config.GetString("CACHE_STORE")
	if store != "" {
		// The store must resolve to a registered driver: either the store
		// names a driver directly ("memory", "redis", or a custom driver
		// added through Extend), or a CACHE_{store}_DRIVER / CACHE_DRIVER
		// key configures one. Anything else would panic on first Store()
		// use, so it is rejected at startup instead.
		driver, explicit, registered := m.manager.(*manager).storeDriverStatus(store)
		switch {
		case !registered:
			v.Optional("CACHE_STORE", "Cache store type", func(any) error {
				return fmt.Errorf("store %q uses driver %q, which is not registered", store, driver)
			})
		case !explicit && driver != store:
			v.Optional("CACHE_STORE", "Cache store type", func(any) error {
				return fmt.Errorf("store %q does not name a registered driver and none is configured; "+
					"set CACHE_DRIVER/CACHE_%s_DRIVER or cache.WithDriver/cache.WithStoreDriver", store, store)
			})
		}

		// Driver-specific keys are validated only when set, keyed on the
		// resolved driver so named stores backed by it are covered too.
		switch driver {
		case "redis":
			if m.config.Has("CACHE_REDIS_PORT") {
				v.Optional("CACHE_REDIS_PORT", "Redis port", configvalidator.ValidatePort)
			}
			if m.config.Has("CACHE_REDIS_DATABASE") {
				v.Optional("CACHE_REDIS_DATABASE", "Redis database number", configvalidator.ValidateNonNegativeInt)
			}
		case "badger":
			if m.config.Has("CACHE_BADGER_GC_INTERVAL") {
				v.Optional("CACHE_BADGER_GC_INTERVAL", "Badger GC interval", func(any) error {
					if m.config.GetDuration("CACHE_BADGER_GC_INTERVAL") <= 0 {
						return fmt.Errorf("GC interval must be positive")
					}
					return nil
				})
			}
		case "bbolt":
			if m.config.Has("CACHE_BBOLT_TIMEOUT") {
				v.Optional("CACHE_BBOLT_TIMEOUT", "bbolt file-lock timeout", func(any) error {
					if m.config.GetDuration("CACHE_BBOLT_TIMEOUT") <= 0 {
						return fmt.Errorf("timeout must be positive")
					}
					return nil
				})
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
