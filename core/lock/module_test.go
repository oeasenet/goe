package lock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModule_Name(t *testing.T) {
	// Create a module with nil manager (just to test Name)
	m := &Module{}
	assert.Equal(t, "lock", m.Name())
}

func TestModule_Provide(t *testing.T) {
	// Create a module with nil manager
	m := &Module{manager: nil}
	assert.Nil(t, m.Provide())
	assert.Nil(t, m.ProvideLockManager())
}

func TestModule_ValidateConfig(t *testing.T) {
	logger := &testLogger{t: t}

	t.Run("no redis config returns error", func(t *testing.T) {
		mockCfg := newMockConfig()
		// No Redis configuration set

		m := &Module{
			config: mockCfg,
			logger: logger,
		}

		err := m.ValidateConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LOCK_REDIS_URL")
	})

	t.Run("with LOCK_REDIS_URL passes validation", func(t *testing.T) {
		mockCfg := newMockConfig()
		mockCfg.Set("LOCK_REDIS_URL", "redis://localhost:6379/0")

		m := &Module{
			config: mockCfg,
			logger: logger,
		}

		err := m.ValidateConfig()
		assert.NoError(t, err)
	})

	t.Run("LOCK_REDIS_ADDR alone is rejected (URL required)", func(t *testing.T) {
		mockCfg := newMockConfig()
		mockCfg.Set("LOCK_REDIS_ADDR", "localhost:6379")

		m := &Module{config: mockCfg, logger: logger}

		err := m.ValidateConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LOCK_REDIS_URL")
	})

	t.Run("LOCK_REDIS_HOST alone is rejected (URL required)", func(t *testing.T) {
		mockCfg := newMockConfig()
		mockCfg.Set("LOCK_REDIS_HOST", "localhost")

		m := &Module{config: mockCfg, logger: logger}

		err := m.ValidateConfig()
		assert.Error(t, err)
	})

	t.Run("with LOCK_REDIS_URLS passes validation", func(t *testing.T) {
		mockCfg := newMockConfig()
		mockCfg.Set("LOCK_REDIS_URLS", []string{"redis://r1:6379/0", "redis://r2:6379/0"})

		m := &Module{config: mockCfg, logger: logger}

		err := m.ValidateConfig()
		assert.NoError(t, err)
	})

	t.Run("validates LOCK_POOL_SIZE", func(t *testing.T) {
		ok := newMockConfig()
		ok.Set("LOCK_REDIS_URL", "redis://localhost:6379/0")
		ok.Set("LOCK_POOL_SIZE", 20)
		assert.NoError(t, (&Module{config: ok, logger: logger}).ValidateConfig())

		bad := newMockConfig()
		bad.Set("LOCK_REDIS_URL", "redis://localhost:6379/0")
		bad.Set("LOCK_POOL_SIZE", -1)
		assert.Error(t, (&Module{config: bad, logger: logger}).ValidateConfig())
	})

	t.Run("validates LOCK_DEFAULT_TRIES", func(t *testing.T) {
		mockCfg := newMockConfig()
		mockCfg.Set("LOCK_REDIS_URL", "redis://localhost:6379/0")
		mockCfg.Set("LOCK_DEFAULT_TRIES", 50)

		m := &Module{
			config: mockCfg,
			logger: logger,
		}

		err := m.ValidateConfig()
		assert.NoError(t, err)
	})
}
