package lock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func (m *mockConfig) GetFloat64(key string) float64 {
	if v, ok := m.values[key].(float64); ok {
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

func (m *mockConfig) Set(key string, value any) {
	m.values[key] = value
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

	assert.Equal(t, "redis://localhost:6379/0", cfg.RedisURL)
	assert.Empty(t, cfg.RedisURLs)
	assert.Equal(t, 8*time.Second, cfg.DefaultExpiry)
	assert.Equal(t, 32, cfg.DefaultTries)
	assert.Equal(t, 500*time.Millisecond, cfg.DefaultRetryDelay)
	assert.Equal(t, 0.01, cfg.DefaultDriftFactor)
	assert.Equal(t, "lock:", cfg.DefaultKeyPrefix)
	assert.Equal(t, 10, cfg.PoolSize)
	assert.False(t, cfg.TLSInsecureSkipVerify)
}

func TestLoadConfig_WithURL(t *testing.T) {
	mockCfg := newMockConfig()
	mockCfg.Set("LOCK_REDIS_URL", "redis://user:pass@myredis:6380/2")

	cfg := LoadConfig(mockCfg)

	assert.Equal(t, "redis://user:pass@myredis:6380/2", cfg.RedisURL)
}

func TestLoadConfig_WithMultipleURLs(t *testing.T) {
	mockCfg := newMockConfig()
	mockCfg.Set("LOCK_REDIS_URLS", []string{
		"redis://redis1:6379/0",
		"redis://redis2:6379/0",
		"redis://redis3:6379/0",
	})

	cfg := LoadConfig(mockCfg)

	assert.Len(t, cfg.RedisURLs, 3)
	assert.Equal(t, "redis://redis1:6379/0", cfg.RedisURLs[0])
	assert.Equal(t, "redis://redis2:6379/0", cfg.RedisURLs[1])
	assert.Equal(t, "redis://redis3:6379/0", cfg.RedisURLs[2])
}

func TestLoadConfig_WithLockSettings(t *testing.T) {
	mockCfg := newMockConfig()
	mockCfg.Set("LOCK_REDIS_URL", "redis://myredis:6379/5")
	mockCfg.Set("LOCK_DEFAULT_EXPIRY", 10*time.Second)
	mockCfg.Set("LOCK_DEFAULT_TRIES", 50)
	mockCfg.Set("LOCK_DEFAULT_RETRY_DELAY", 200*time.Millisecond)
	mockCfg.Set("LOCK_DEFAULT_DRIFT_FACTOR", 0.02)
	mockCfg.Set("LOCK_KEY_PREFIX", "myapp:lock:")
	mockCfg.Set("LOCK_POOL_SIZE", 20)

	cfg := LoadConfig(mockCfg)

	assert.Equal(t, "redis://myredis:6379/5", cfg.RedisURL)
	assert.Equal(t, 10*time.Second, cfg.DefaultExpiry)
	assert.Equal(t, 50, cfg.DefaultTries)
	assert.Equal(t, 200*time.Millisecond, cfg.DefaultRetryDelay)
	assert.Equal(t, 0.02, cfg.DefaultDriftFactor)
	assert.Equal(t, "myapp:lock:", cfg.DefaultKeyPrefix)
	assert.Equal(t, 20, cfg.PoolSize)
}

func TestLoadConfig_TLSSettings(t *testing.T) {
	mockCfg := newMockConfig()
	mockCfg.Set("LOCK_TLS_INSECURE_SKIP_VERIFY", true)

	cfg := LoadConfig(mockCfg)

	assert.True(t, cfg.TLSInsecureSkipVerify)
}

func TestLoadConfig_DefaultsWhenEmpty(t *testing.T) {
	mockCfg := newMockConfig()
	// No settings provided

	cfg := LoadConfig(mockCfg)

	// Should use defaults
	assert.Equal(t, "redis://localhost:6379/0", cfg.RedisURL)
	assert.Empty(t, cfg.RedisURLs)
	assert.Equal(t, 8*time.Second, cfg.DefaultExpiry)
	assert.Equal(t, 32, cfg.DefaultTries)
	assert.Equal(t, 500*time.Millisecond, cfg.DefaultRetryDelay)
	assert.Equal(t, 0.01, cfg.DefaultDriftFactor)
	assert.Equal(t, "lock:", cfg.DefaultKeyPrefix)
	assert.Equal(t, 10, cfg.PoolSize)
}

func TestLoadConfig_SentinelURL(t *testing.T) {
	mockCfg := newMockConfig()
	mockCfg.Set("LOCK_REDIS_URL", "redis-sentinel://mymaster@sentinel1:26379,sentinel2:26379/0")

	cfg := LoadConfig(mockCfg)

	assert.Equal(t, "redis-sentinel://mymaster@sentinel1:26379,sentinel2:26379/0", cfg.RedisURL)
}

func TestLoadConfig_ClusterURL(t *testing.T) {
	mockCfg := newMockConfig()
	mockCfg.Set("LOCK_REDIS_URL", "redis-cluster://node1:6379,node2:6379,node3:6379")

	cfg := LoadConfig(mockCfg)

	assert.Equal(t, "redis-cluster://node1:6379,node2:6379,node3:6379", cfg.RedisURL)
}

func TestLoadConfig_TLSviaScheme(t *testing.T) {
	mockCfg := newMockConfig()
	mockCfg.Set("LOCK_REDIS_URL", "rediss://secure-redis:6379/0")

	cfg := LoadConfig(mockCfg)

	assert.Equal(t, "rediss://secure-redis:6379/0", cfg.RedisURL)
}
