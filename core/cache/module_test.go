package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/config"
	"go.oease.dev/goe/v2/core/log"
)

func newTestModule(t *testing.T, setup func(*config.Module), opts ...Option) *Module {
	t.Helper()
	cfgModule := config.NewModule()
	if setup != nil {
		setup(cfgModule)
	}
	cfg := cfgModule.Provide()
	logger := log.NewModule(cfg).Provide()
	return NewModule(cfg, logger, opts...)
}

func TestModuleProvide(t *testing.T) {
	t.Run("zero config yields memory-backed cache with APP_NAME prefix", func(t *testing.T) {
		m := newTestModule(t, func(cfg *config.Module) {
			cfg.Provide().Set("APP_NAME", "myapp")
		})
		require.NoError(t, m.ValidateConfig())
		c := m.Provide()
		require.NotNil(t, c)
		assert.Equal(t, "myapp", c.GetPrefix())
		require.NoError(t, c.Set("k", "v", time.Minute))
		var out string
		require.NoError(t, c.Get("k", &out))
		assert.Equal(t, "v", out)
	})

	t.Run("Provide returns the same instance", func(t *testing.T) {
		m := newTestModule(t, nil)
		assert.Same(t, m.Provide(), m.Provide())
	})

	t.Run("CACHE_PREFIX wins over APP_NAME", func(t *testing.T) {
		m := newTestModule(t, func(cfg *config.Module) {
			cfg.Provide().Set("APP_NAME", "myapp")
			cfg.Provide().Set("CACHE_PREFIX", "cachepfx")
		})
		assert.Equal(t, "cachepfx", m.Provide().GetPrefix())
	})

	t.Run("unregistered driver panics on Provide", func(t *testing.T) {
		m := newTestModule(t, func(cfg *config.Module) {
			cfg.Provide().Set("CACHE_DRIVER", "nope")
		})
		assert.Panics(t, func() { m.Provide() })
	})

	t.Run("custom driver via option", func(t *testing.T) {
		store := newTTLRecordingStore()
		m := newTestModule(t, nil,
			WithDriver("fake"),
			WithCustomDriver("fake", func(contract.Config) (contract.CacheStore, error) { return store, nil }),
		)
		require.NoError(t, m.ValidateConfig())
		c := m.Provide()
		require.NoError(t, c.Set("k", "v", time.Minute))
		var out string
		require.NoError(t, c.Get("k", &out))
		assert.Equal(t, "v", out)
		assert.Len(t, store.data, 1)
	})

	t.Run("CACHE_TTL feeds the default TTL", func(t *testing.T) {
		store := newTTLRecordingStore()
		m := newTestModule(t, func(cfg *config.Module) {
			cfg.Provide().Set("CACHE_TTL", "90m")
		},
			WithDriver("fake"),
			WithCustomDriver("fake", func(contract.Config) (contract.CacheStore, error) { return store, nil }),
		)
		require.NoError(t, m.Provide().Set("k", "v", 0))
		assert.Equal(t, 90*time.Minute, store.lastExp)
	})

	t.Run("OnStop closes the built store and tolerates never-built", func(t *testing.T) {
		m := newTestModule(t, nil)
		require.NoError(t, m.OnStop(t.Context()))
		_ = m.Provide()
		require.NoError(t, m.OnStop(t.Context()))
	})
}
