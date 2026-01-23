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
		assert.Contains(t, err.Error(), "LOCK_REDIS_ADDR")
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

	t.Run("with LOCK_REDIS_ADDR passes validation", func(t *testing.T) {
		mockCfg := newMockConfig()
		mockCfg.Set("LOCK_REDIS_ADDR", "localhost:6379")

		m := &Module{
			config: mockCfg,
			logger: logger,
		}

		err := m.ValidateConfig()
		assert.NoError(t, err)
	})

	t.Run("with LOCK_REDIS_HOST passes validation", func(t *testing.T) {
		mockCfg := newMockConfig()
		mockCfg.Set("LOCK_REDIS_HOST", "localhost")

		m := &Module{
			config: mockCfg,
			logger: logger,
		}

		err := m.ValidateConfig()
		assert.NoError(t, err)
	})

	t.Run("with LOCK_REDIS_HOSTS passes validation", func(t *testing.T) {
		mockCfg := newMockConfig()
		mockCfg.Set("LOCK_REDIS_HOSTS", []string{"localhost:6379", "localhost:6380"})

		m := &Module{
			config: mockCfg,
			logger: logger,
		}

		err := m.ValidateConfig()
		assert.NoError(t, err)
	})

	t.Run("validates LOCK_REDIS_DB format", func(t *testing.T) {
		mockCfg := newMockConfig()
		mockCfg.Set("LOCK_REDIS_URL", "redis://localhost:6379/0")
		mockCfg.Set("LOCK_REDIS_DB", 5) // valid positive int

		m := &Module{
			config: mockCfg,
			logger: logger,
		}

		err := m.ValidateConfig()
		assert.NoError(t, err)
	})

	t.Run("validates LOCK_REDIS_POOL_SIZE", func(t *testing.T) {
		mockCfg := newMockConfig()
		mockCfg.Set("LOCK_REDIS_URL", "redis://localhost:6379/0")
		mockCfg.Set("LOCK_REDIS_POOL_SIZE", 20)

		m := &Module{
			config: mockCfg,
			logger: logger,
		}

		err := m.ValidateConfig()
		assert.NoError(t, err)
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
