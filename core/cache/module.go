package cache

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"
	"sync"

	"go.oease.dev/goe/v2/contract"
	goeconfig "go.oease.dev/goe/v2/core/config"
	"go.oease.dev/goe/v2/core/internal/configvalidator"
)

// Module represents the cache module for Fx. It owns exactly one cache
// backed by the configured driver; the multi-store manager was removed in
// v2.5 (stores never had per-store connections, so a named store could only
// ever alias a driver + prefix pair).
type Module struct {
	logger contract.Logger

	// config is the effective configuration: the base config with the Option
	// overlay applied. The driver factory reads through it, so
	// code-configured values behave exactly like environment variables.
	config contract.Config

	// optErrs holds failures from Option application. They are reported by
	// ValidateConfig so that startup aborts; the module fell back to the
	// environment-only configuration, which is never actually served.
	optErrs []error

	// drivers maps driver names to factories: the builtins plus anything
	// registered through WithCustomDriver.
	drivers map[string]contract.CacheStoreFactory

	// mu guards cache so Provide and OnStop are safe to call in any order:
	// the framework builds at registration, tests build on demand.
	mu    sync.Mutex
	cache contract.Cache
}

// NewModule creates a new cache module.
//
// Configuration resolves in layers: CACHE_* environment variables, then opts.
// Anything set through an Option wins over the environment. If any option
// fails, every option is discarded, the environment-only configuration stays
// in effect, and ValidateConfig aborts startup with all collected errors.
func NewModule(config contract.Config, logger contract.Logger, opts ...Option) *Module {
	overrides, customDrivers, optErrs := resolveOverrides(opts)
	if len(optErrs) == 0 && len(overrides) > 0 {
		config = goeconfig.NewConfigWrapper(config, overrides)
	}

	drivers := builtinDrivers()
	if len(optErrs) == 0 {
		maps.Copy(drivers, customDrivers)
	}

	return &Module{
		logger:  logger.With("module", "cache"),
		config:  config,
		optErrs: optErrs,
		drivers: drivers,
	}
}

// Name returns the module name
func (m *Module) Name() string {
	return "cache"
}

// driverName resolves the configured driver, defaulting to memory.
func (m *Module) driverName() string {
	if driver := m.config.GetString("CACHE_DRIVER"); driver != "" {
		return driver
	}
	return "memory"
}

// Provide returns the cache instance, building it on first call.
//
// The framework calls this once at module registration inside goe.New, so a
// misconfigured or unreachable backend stops startup immediately — the same
// fail-fast point at which the job and lock modules open their connections.
// The panics below carry the same guidance startup validation gives for the
// paths validation cannot reach first.
func (m *Module) Provide() contract.Cache {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cache != nil {
		return m.cache
	}

	driver := m.driverName()
	factory, ok := m.drivers[driver]
	if !ok {
		panic(fmt.Sprintf("cache driver [%s] is not registered", driver))
	}

	store, err := factory(m.config)
	if err != nil {
		panic(fmt.Sprintf("failed to create cache store with driver [%s]: %v", driver, err))
	}

	prefix := m.config.GetString("CACHE_PREFIX")
	if prefix == "" {
		prefix = m.config.GetString("APP_NAME")
	}

	m.cache = New(store, prefix, m.config.GetDuration("CACHE_TTL"))
	return m.cache
}

// OnStart is called when the module starts
func (m *Module) OnStart(ctx context.Context) error {
	m.logger.Info("Cache module started", "driver", m.driverName())
	return nil
}

// OnStop is called when the module stops
func (m *Module) OnStop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cache == nil {
		return nil
	}

	if err := m.cache.Store().Close(); err != nil {
		m.logger.Error("Error closing cache store", "error", err)
	}
	m.logger.Info("Cache module stopped")
	return nil
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

	// Multi-store configuration was removed in v2.5. Keys that used to
	// select or shape stores fail startup with a migration hint instead of
	// being silently ignored — an app that relied on them must not come up
	// with a differently-wired cache.
	if m.config.Has("CACHE_STORE") {
		v.Optional("CACHE_STORE", "Removed multi-store key", func(any) error {
			return fmt.Errorf("CACHE_STORE was removed in v2.5 along with multi-store support; "+
				"set CACHE_DRIVER=%s (or cache.WithDriver) instead", m.config.GetString("CACHE_STORE"))
		})
	}
	for key := range m.config.All() {
		if key == "CACHE_DRIVER" || !strings.HasPrefix(key, "CACHE_") || !strings.HasSuffix(key, "_DRIVER") {
			continue
		}
		v.Optional(key, "Removed per-store driver key", func(any) error {
			return errors.New("per-store driver keys were removed in v2.5 along with multi-store support; " +
				"configure the single driver with CACHE_DRIVER or cache.WithDriver")
		})
	}

	// The configured driver must be registered: a builtin or a
	// WithCustomDriver registration. Anything else would panic on first use,
	// so it is rejected at startup instead.
	driver := m.driverName()
	if _, ok := m.drivers[driver]; !ok {
		v.Optional("CACHE_DRIVER", "Cache driver", func(any) error {
			return fmt.Errorf("driver %q is not registered (builtins: memory, redis, badger, bbolt; "+
				"custom drivers register through cache.WithCustomDriver)", driver)
		})
	}

	// Driver-specific keys are validated only when set.
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

	// TTL is optional but should be positive if set
	if m.config.Has("CACHE_TTL") {
		v.Optional("CACHE_TTL", "Default cache TTL", func(value any) error {
			if m.config.GetDuration("CACHE_TTL") < 0 {
				return fmt.Errorf("TTL must be positive")
			}
			return nil
		})
	}

	return v.Validate()
}
