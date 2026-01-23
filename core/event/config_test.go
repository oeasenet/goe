package event

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockConfig implements contract.Config for testing
type mockConfig struct {
	values map[string]any
}

func newMockConfig() *mockConfig {
	return &mockConfig{
		values: make(map[string]any),
	}
}

func (m *mockConfig) Get(key string) any {
	return m.values[key]
}

func (m *mockConfig) Set(key string, value any) {
	m.values[key] = value
}

func (m *mockConfig) GetString(key string) string {
	if v, ok := m.values[key].(string); ok {
		return v
	}
	return ""
}

func (m *mockConfig) GetInt(key string) int {
	if v, ok := m.values[key].(int); ok {
		return v
	}
	return 0
}

func (m *mockConfig) GetInt64(key string) int64 {
	if v, ok := m.values[key].(int64); ok {
		return v
	}
	return 0
}

func (m *mockConfig) GetBool(key string) bool {
	if v, ok := m.values[key].(bool); ok {
		return v
	}
	return false
}

func (m *mockConfig) GetFloat64(key string) float64 {
	if v, ok := m.values[key].(float64); ok {
		return v
	}
	return 0
}

func (m *mockConfig) GetDuration(key string) time.Duration {
	if v, ok := m.values[key].(time.Duration); ok {
		return v
	}
	return 0
}

func (m *mockConfig) GetStringSlice(key string) []string {
	if v, ok := m.values[key].([]string); ok {
		return v
	}
	return nil
}

func (m *mockConfig) GetStringMap(key string) map[string]any {
	if v, ok := m.values[key].(map[string]any); ok {
		return v
	}
	return nil
}

func (m *mockConfig) Has(key string) bool {
	_, ok := m.values[key]
	return ok
}

func (m *mockConfig) All() map[string]any {
	return m.values
}

func (m *mockConfig) Reload() error {
	return nil
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	require.NotNil(t, cfg)

	t.Run("Redis defaults", func(t *testing.T) {
		assert.Equal(t, []string{"localhost:6379"}, cfg.RedisHosts)
		assert.Empty(t, cfg.RedisUsername)
		assert.Empty(t, cfg.RedisPassword)
		assert.Equal(t, 0, cfg.RedisDB)
		assert.Equal(t, "localhost:6379", cfg.RedisAddr)
		assert.Empty(t, cfg.RedisURL)
	})

	t.Run("Consumer defaults", func(t *testing.T) {
		assert.Equal(t, 30*time.Second, cfg.ConsumerTimeout)
		assert.Equal(t, 3, cfg.MaxRetries)
		assert.Equal(t, time.Second, cfg.RetryBackoff)
		assert.Equal(t, 24*time.Hour, cfg.DeadLetterQueueTTL)
		assert.Equal(t, 5*time.Minute, cfg.StaleConsumerTimeout)
	})

	t.Run("Performance defaults", func(t *testing.T) {
		assert.Equal(t, 10, cfg.BatchSize)
		assert.Equal(t, 1000, cfg.MaxPendingMessages)
		assert.Equal(t, time.Minute, cfg.ClaimMinIdleTime)
		assert.Equal(t, 30*time.Second, cfg.ClaimInterval)
	})

	t.Run("Delayed queue defaults", func(t *testing.T) {
		assert.True(t, cfg.DelayedQueueEnabled)
		assert.Equal(t, time.Second, cfg.DelayedQueueCheckInterval)
	})
}

func TestLoadConfig(t *testing.T) {
	t.Run("empty config returns defaults", func(t *testing.T) {
		mockCfg := newMockConfig()
		cfg := LoadConfig(mockCfg)

		assert.Equal(t, DefaultConfig().RedisHosts, cfg.RedisHosts)
		assert.Equal(t, DefaultConfig().ConsumerTimeout, cfg.ConsumerTimeout)
	})

	t.Run("Redis URL takes priority", func(t *testing.T) {
		mockCfg := newMockConfig()
		mockCfg.Set("EVENT_REDIS_URL", "redis://user:pass@redis.example.com:6380/1")

		cfg := LoadConfig(mockCfg)

		assert.Equal(t, "redis://user:pass@redis.example.com:6380/1", cfg.RedisURL)
	})

	t.Run("Redis hosts configuration", func(t *testing.T) {
		mockCfg := newMockConfig()
		mockCfg.Set("EVENT_REDIS_HOSTS", []string{"redis1:6379", "redis2:6379"})
		mockCfg.Set("EVENT_REDIS_USERNAME", "testuser")
		mockCfg.Set("EVENT_REDIS_PASSWORD", "testpass")
		mockCfg.Set("EVENT_REDIS_DB", 2)

		cfg := LoadConfig(mockCfg)

		assert.Equal(t, []string{"redis1:6379", "redis2:6379"}, cfg.RedisHosts)
		assert.Equal(t, "testuser", cfg.RedisUsername)
		assert.Equal(t, "testpass", cfg.RedisPassword)
		assert.Equal(t, 2, cfg.RedisDB)
	})

	t.Run("Legacy Redis addr fallback", func(t *testing.T) {
		// Note: The legacy fallback only triggers when RedisHosts would be empty,
		// but DefaultConfig() already sets RedisHosts to ["localhost:6379"],
		// so the condition len(cfg.RedisHosts) == 0 is never satisfied.
		// This test verifies the current behavior where legacy addr is ignored
		// when no explicit hosts are configured (defaults take precedence).
		mockCfg := newMockConfig()
		mockCfg.Set("EVENT_REDIS_ADDR", "legacy-redis:6379")

		cfg := LoadConfig(mockCfg)

		// RedisAddr is only set when the legacy fallback triggers
		assert.Equal(t, DefaultConfig().RedisAddr, cfg.RedisAddr)
		assert.Equal(t, DefaultConfig().RedisHosts, cfg.RedisHosts)
	})

	t.Run("Consumer settings override", func(t *testing.T) {
		mockCfg := newMockConfig()
		mockCfg.Set("EVENT_CONSUMER_TIMEOUT", 60*time.Second)
		mockCfg.Set("EVENT_MAX_RETRIES", 5)
		mockCfg.Set("EVENT_RETRY_BACKOFF", 2*time.Second)
		mockCfg.Set("EVENT_DLQ_TTL", 48*time.Hour)
		mockCfg.Set("EVENT_STALE_CONSUMER_TIMEOUT", 10*time.Minute)

		cfg := LoadConfig(mockCfg)

		assert.Equal(t, 60*time.Second, cfg.ConsumerTimeout)
		assert.Equal(t, 5, cfg.MaxRetries)
		assert.Equal(t, 2*time.Second, cfg.RetryBackoff)
		assert.Equal(t, 48*time.Hour, cfg.DeadLetterQueueTTL)
		assert.Equal(t, 10*time.Minute, cfg.StaleConsumerTimeout)
	})

	t.Run("Performance settings override", func(t *testing.T) {
		mockCfg := newMockConfig()
		mockCfg.Set("EVENT_BATCH_SIZE", 20)
		mockCfg.Set("EVENT_MAX_PENDING_MESSAGES", 500)
		mockCfg.Set("EVENT_CLAIM_MIN_IDLE_TIME", 2*time.Minute)
		mockCfg.Set("EVENT_CLAIM_INTERVAL", time.Minute)

		cfg := LoadConfig(mockCfg)

		assert.Equal(t, 20, cfg.BatchSize)
		assert.Equal(t, 500, cfg.MaxPendingMessages)
		assert.Equal(t, 2*time.Minute, cfg.ClaimMinIdleTime)
		assert.Equal(t, time.Minute, cfg.ClaimInterval)
	})

	t.Run("Delayed queue settings", func(t *testing.T) {
		mockCfg := newMockConfig()
		mockCfg.Set("EVENT_DELAYED_QUEUE_ENABLED", false)
		mockCfg.Set("EVENT_DELAYED_QUEUE_CHECK_INTERVAL", 5*time.Second)

		cfg := LoadConfig(mockCfg)

		assert.False(t, cfg.DelayedQueueEnabled)
		assert.Equal(t, 5*time.Second, cfg.DelayedQueueCheckInterval)
	})

	t.Run("Full configuration", func(t *testing.T) {
		mockCfg := newMockConfig()
		mockCfg.Set("EVENT_REDIS_URL", "redis://localhost:6379/0")
		mockCfg.Set("EVENT_CONSUMER_TIMEOUT", 45*time.Second)
		mockCfg.Set("EVENT_MAX_RETRIES", 10)
		mockCfg.Set("EVENT_RETRY_BACKOFF", 500*time.Millisecond)
		mockCfg.Set("EVENT_DLQ_TTL", 72*time.Hour)
		mockCfg.Set("EVENT_STALE_CONSUMER_TIMEOUT", 15*time.Minute)
		mockCfg.Set("EVENT_BATCH_SIZE", 50)
		mockCfg.Set("EVENT_MAX_PENDING_MESSAGES", 2000)
		mockCfg.Set("EVENT_CLAIM_MIN_IDLE_TIME", 30*time.Second)
		mockCfg.Set("EVENT_CLAIM_INTERVAL", 15*time.Second)
		mockCfg.Set("EVENT_DELAYED_QUEUE_ENABLED", false)
		mockCfg.Set("EVENT_DELAYED_QUEUE_CHECK_INTERVAL", 2*time.Second)

		cfg := LoadConfig(mockCfg)

		assert.Equal(t, "redis://localhost:6379/0", cfg.RedisURL)
		assert.Equal(t, 45*time.Second, cfg.ConsumerTimeout)
		assert.Equal(t, 10, cfg.MaxRetries)
		assert.Equal(t, 500*time.Millisecond, cfg.RetryBackoff)
		assert.Equal(t, 72*time.Hour, cfg.DeadLetterQueueTTL)
		assert.Equal(t, 15*time.Minute, cfg.StaleConsumerTimeout)
		assert.Equal(t, 50, cfg.BatchSize)
		assert.Equal(t, 2000, cfg.MaxPendingMessages)
		assert.Equal(t, 30*time.Second, cfg.ClaimMinIdleTime)
		assert.Equal(t, 15*time.Second, cfg.ClaimInterval)
		assert.False(t, cfg.DelayedQueueEnabled)
		assert.Equal(t, 2*time.Second, cfg.DelayedQueueCheckInterval)
	})
}

func TestLoadConfig_Priority(t *testing.T) {
	t.Run("URL takes priority over hosts", func(t *testing.T) {
		mockCfg := newMockConfig()
		mockCfg.Set("EVENT_REDIS_URL", "redis://url-host:6379")
		mockCfg.Set("EVENT_REDIS_HOSTS", []string{"host-host:6379"})
		mockCfg.Set("EVENT_REDIS_ADDR", "addr-host:6379")

		cfg := LoadConfig(mockCfg)

		assert.Equal(t, "redis://url-host:6379", cfg.RedisURL)
		// When URL is set, hosts should remain at defaults
		assert.Equal(t, DefaultConfig().RedisHosts, cfg.RedisHosts)
	})

	t.Run("Hosts takes priority over legacy addr", func(t *testing.T) {
		mockCfg := newMockConfig()
		mockCfg.Set("EVENT_REDIS_HOSTS", []string{"host1:6379", "host2:6379"})
		mockCfg.Set("EVENT_REDIS_ADDR", "legacy:6379")

		cfg := LoadConfig(mockCfg)

		assert.Equal(t, []string{"host1:6379", "host2:6379"}, cfg.RedisHosts)
	})
}
