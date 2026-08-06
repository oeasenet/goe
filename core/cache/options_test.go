package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/zap"
)

// nopLogger implements contract.Logger for unit tests that don't need output.
type nopLogger struct{}

func (l *nopLogger) Debug(msg string, args ...any)                   {}
func (l *nopLogger) Info(msg string, args ...any)                    {}
func (l *nopLogger) Warn(msg string, args ...any)                    {}
func (l *nopLogger) Error(msg string, args ...any)                   {}
func (l *nopLogger) Fatal(msg string, args ...any)                   {}
func (l *nopLogger) Debugf(template string, args ...any)             {}
func (l *nopLogger) Infof(template string, args ...any)              {}
func (l *nopLogger) Warnf(template string, args ...any)              {}
func (l *nopLogger) Errorf(template string, args ...any)             {}
func (l *nopLogger) Fatalf(template string, args ...any)             {}
func (l *nopLogger) Panicf(template string, args ...any)             {}
func (l *nopLogger) Debugw(msg string, keysAndValues ...any)         {}
func (l *nopLogger) Infow(msg string, keysAndValues ...any)          {}
func (l *nopLogger) Warnw(msg string, keysAndValues ...any)          {}
func (l *nopLogger) Errorw(msg string, keysAndValues ...any)         {}
func (l *nopLogger) Fatalw(msg string, keysAndValues ...any)         {}
func (l *nopLogger) With(keysAndValues ...any) contract.Logger       { return l }
func (l *nopLogger) WithContext(ctx context.Context) contract.Logger { return l }
func (l *nopLogger) WithError(err error) contract.Logger             { return l }
func (l *nopLogger) GetLogger() *zap.SugaredLogger                   { return nil }

// mapConfig is a map-backed contract.Config. The layering tests need real
// per-key lookups, which the testify-mock MockConfig cannot express cleanly.
type mapConfig struct {
	values map[string]any
}

func newMapConfig() *mapConfig {
	return &mapConfig{values: make(map[string]any)}
}

func (m *mapConfig) Get(key string) any { return m.values[key] }

func (m *mapConfig) GetString(key string) string {
	if v, ok := m.values[key].(string); ok {
		return v
	}
	return ""
}

func (m *mapConfig) GetInt(key string) int {
	if v, ok := m.values[key].(int); ok {
		return v
	}
	return 0
}

func (m *mapConfig) GetInt64(key string) int64 {
	if v, ok := m.values[key].(int64); ok {
		return v
	}
	return 0
}

func (m *mapConfig) GetFloat64(key string) float64 {
	if v, ok := m.values[key].(float64); ok {
		return v
	}
	return 0
}

func (m *mapConfig) GetBool(key string) bool {
	if v, ok := m.values[key].(bool); ok {
		return v
	}
	return false
}

func (m *mapConfig) GetDuration(key string) time.Duration {
	if v, ok := m.values[key].(time.Duration); ok {
		return v
	}
	return 0
}

func (m *mapConfig) GetStringSlice(key string) []string {
	if v, ok := m.values[key].([]string); ok {
		return v
	}
	return nil
}

func (m *mapConfig) GetStringMap(key string) map[string]any {
	if v, ok := m.values[key].(map[string]any); ok {
		return v
	}
	return nil
}

func (m *mapConfig) Set(key string, value any) { m.values[key] = value }

func (m *mapConfig) Has(key string) bool {
	_, ok := m.values[key]
	return ok
}

func (m *mapConfig) All() map[string]any { return m.values }

func (m *mapConfig) Reload() error { return nil }

// probeStore is a minimal CacheStore whose factory records the config it was
// handed, proving custom drivers see code-configured values.
type probeStore struct{}

func (s *probeStore) Get(key string) ([]byte, error)                      { return nil, nil }
func (s *probeStore) Set(key string, val []byte, exp time.Duration) error { return nil }
func (s *probeStore) Delete(key string) error                             { return nil }
func (s *probeStore) Reset() error                                        { return nil }
func (s *probeStore) Close() error                                        { return nil }

func TestOptions_Precedence(t *testing.T) {
	env := newMapConfig()
	env.Set("CACHE_TTL", time.Hour)
	env.Set("CACHE_PREFIX", "envprefix")

	m := NewModule(env, &nopLogger{}, WithTTL(30*time.Minute))
	require.NoError(t, m.ValidateConfig())

	// Code wins over the matching environment variable.
	assert.Equal(t, 30*time.Minute, m.config.GetDuration("CACHE_TTL"))
	// The environment still supplies everything code does not set.
	assert.Equal(t, "envprefix", m.config.GetString("CACHE_PREFIX"))
}

func TestOptions_Ordering(t *testing.T) {
	m := NewModule(newMapConfig(), &nopLogger{},
		WithTTL(time.Minute),
		WithTTL(2*time.Minute),
	)
	assert.Equal(t, 2*time.Minute, m.config.GetDuration("CACHE_TTL"))
}

func TestOptions_NilOptionsSkipped(t *testing.T) {
	m := NewModule(newMapConfig(), &nopLogger{}, nil, WithTTL(time.Minute))
	require.NoError(t, m.ValidateConfig())
	assert.Equal(t, time.Minute, m.config.GetDuration("CACHE_TTL"))
}

func TestDriverResolution_DefaultsToMemory(t *testing.T) {
	m := NewModule(newMapConfig(), &nopLogger{})
	assert.Equal(t, "memory", m.driverName())
	require.NoError(t, m.ValidateConfig())
}

func TestDriverResolution_WithDriverSelectsDriver(t *testing.T) {
	m := NewModule(newMapConfig(), &nopLogger{}, WithDriver("redis"))
	assert.Equal(t, "redis", m.driverName())
	require.NoError(t, m.ValidateConfig())
}

func TestDriverResolution_EnvDriverSelectsDriver(t *testing.T) {
	env := newMapConfig()
	env.Set("CACHE_DRIVER", "redis")

	m := NewModule(env, &nopLogger{})
	assert.Equal(t, "redis", m.driverName())
	require.NoError(t, m.ValidateConfig())
}

func TestDriverResolution_UnknownDriverRejected(t *testing.T) {
	env := newMapConfig()
	env.Set("CACHE_DRIVER", "bogus")

	m := NewModule(env, &nopLogger{})
	err := m.ValidateConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not registered")
}

func TestOptions_ErrorsFallBackToEnvironmentOnly(t *testing.T) {
	env := newMapConfig()
	env.Set("CACHE_TTL", time.Hour)

	m := NewModule(env, &nopLogger{},
		WithTTL(30*time.Minute),
		WithRedisPort(-1),
	)

	// Any option error discards every option: the environment-only
	// configuration remains in effect.
	assert.Equal(t, time.Hour, m.config.GetDuration("CACHE_TTL"))

	// ValidateConfig reports the failures and aborts startup.
	err := m.ValidateConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cache option 2")
	assert.Contains(t, err.Error(), "WithRedisPort")
}

func TestOptions_AllErrorsReportedTogether(t *testing.T) {
	m := NewModule(newMapConfig(), &nopLogger{},
		WithDriver(""),
		WithRedisPort(70000),
	)
	err := m.ValidateConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cache option 1")
	assert.Contains(t, err.Error(), "cache option 2")
}

func TestOptions_Validation(t *testing.T) {
	cases := []struct {
		name string
		opt  Option
	}{
		{"WithDriver empty", WithDriver("")},
		{"WithPrefix empty", WithPrefix("")},
		{"WithTTL zero", WithTTL(0)},
		{"WithMemoryGCInterval zero", WithMemoryGCInterval(0)},
		{"WithRedisHost empty", WithRedisHost("")},
		{"WithRedisPort zero", WithRedisPort(0)},
		{"WithRedisPort too large", WithRedisPort(65536)},
		{"WithRedisDatabase negative", WithRedisDatabase(-1)},
		{"WithRedisAddrs empty", WithRedisAddrs()},
		{"WithRedisAddrs blank entry", WithRedisAddrs("")},
		{"WithRedisMasterName empty", WithRedisMasterName("")},
		{"WithRedisClientName empty", WithRedisClientName("")},
		{"WithRedisPoolSize zero", WithRedisPoolSize(0)},
		{"WithCustomDriver empty name", WithCustomDriver("", func(contract.Config) (contract.CacheStore, error) { return &probeStore{}, nil })},
		{"WithCustomDriver nil factory", WithCustomDriver("probe", nil)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, errs := resolveOverrides([]Option{tc.opt})
			assert.NotEmpty(t, errs)
		})
	}
}

func TestOptions_RedisConnectionKeysApply(t *testing.T) {
	m := NewModule(newMapConfig(), &nopLogger{},
		WithDriver("redis"),
		WithRedisHost("redis.internal"),
		WithRedisPort(6380),
		WithRedisDatabase(2),
		WithRedisPoolSize(15),
		WithRedisClientName("myapp"),
		WithRedisClusterMode(true),
		WithRedisAddrs("n1:6379", "n2:6379"),
		WithRedisMasterName("mymaster"),
		WithRedisReset(true),
	)
	require.NoError(t, m.ValidateConfig())

	assert.Equal(t, "redis.internal", m.config.GetString("CACHE_REDIS_HOST"))
	assert.Equal(t, 6380, m.config.GetInt("CACHE_REDIS_PORT"))
	assert.Equal(t, 2, m.config.GetInt("CACHE_REDIS_DATABASE"))
	assert.Equal(t, 15, m.config.GetInt("CACHE_REDIS_POOL_SIZE"))
	assert.Equal(t, "myapp", m.config.GetString("CACHE_REDIS_CLIENT_NAME"))
	assert.True(t, m.config.GetBool("CACHE_REDIS_IS_CLUSTER_MODE"))
	assert.Equal(t, []string{"n1:6379", "n2:6379"}, m.config.GetStringSlice("CACHE_REDIS_ADDRS"))
	assert.Equal(t, "mymaster", m.config.GetString("CACHE_REDIS_MASTER_NAME"))
	assert.True(t, m.config.GetBool("CACHE_REDIS_RESET"))
}

func TestOptions_CustomDriverSeesCodeConfig(t *testing.T) {
	// A custom driver selected with WithDriver must be used, and its factory
	// must receive the merged (env + code) configuration.
	env := newMapConfig()
	env.Set("CACHE_PREFIX", "envprefix")

	var seen contract.Config
	m := NewModule(env, &nopLogger{},
		WithDriver("probe"),
		WithRedisHost("code-host"),
		WithCustomDriver("probe", func(config contract.Config) (contract.CacheStore, error) {
			seen = config
			return &probeStore{}, nil
		}),
	)
	require.NoError(t, m.ValidateConfig())

	c := m.Provide()
	require.NotNil(t, c)
	require.NotNil(t, seen, "factory was not invoked")
	assert.Equal(t, "code-host", seen.GetString("CACHE_REDIS_HOST"))
	assert.Equal(t, "envprefix", seen.GetString("CACHE_PREFIX"))
	assert.Equal(t, "envprefix", c.GetPrefix())
}

func TestOptions_CustomDriverDiscardedOnOptionError(t *testing.T) {
	// A failing option discards every option, including custom driver
	// registrations — otherwise a half-applied configuration could serve a
	// code-registered driver against environment-only settings.
	m := NewModule(newMapConfig(), &nopLogger{},
		WithCustomDriver("probe", func(contract.Config) (contract.CacheStore, error) { return &probeStore{}, nil }),
		WithRedisPort(-1),
	)
	_, registered := m.drivers["probe"]
	assert.False(t, registered)
	assert.Error(t, m.ValidateConfig())
}

func TestValidateConfig_RemovedMultiStoreKeysRejected(t *testing.T) {
	t.Run("CACHE_STORE", func(t *testing.T) {
		env := newMapConfig()
		env.Set("CACHE_STORE", "redis")

		m := NewModule(env, &nopLogger{})
		err := m.ValidateConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "CACHE_STORE was removed")
		assert.Contains(t, err.Error(), "CACHE_DRIVER=redis")
	})

	t.Run("per-store driver key", func(t *testing.T) {
		env := newMapConfig()
		env.Set("CACHE_sessions_DRIVER", "redis")

		m := NewModule(env, &nopLogger{})
		err := m.ValidateConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "per-store driver keys were removed")
	})
}
