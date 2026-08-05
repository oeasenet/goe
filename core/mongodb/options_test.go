package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/event"
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

// mapConfig is a map-backed contract.Config; the layering tests need real
// per-key lookups rather than testify expectations.
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

func TestOptions_Precedence(t *testing.T) {
	env := newMapConfig()
	env.Set("MONGO_URI", "mongodb://env-host:27017")
	env.Set("MONGO_DB_NAME", "envdb")

	m := NewDBModule(env, &nopLogger{}, WithDatabase("codedb"))
	require.NoError(t, m.ValidateConfig())

	// Code wins over the matching environment variable.
	assert.Equal(t, "codedb", m.config.GetString("MONGO_DB_NAME"))
	// The environment still supplies everything code does not set.
	assert.Equal(t, "mongodb://env-host:27017", m.config.GetString("MONGO_URI"))
}

func TestOptions_AllFieldsApply(t *testing.T) {
	env := newMapConfig()
	env.Set("MONGO_URI", "mongodb://env-host:27017")

	m := NewDBModule(env, &nopLogger{},
		WithDatabase("mydb"),
		WithMinPoolSize(2),
		WithMaxPoolSize(50),
		WithMaxConnIdleTime(3*time.Minute),
		WithDebug(true),
		WithPingTimeout(2*time.Second),
	)

	assert.Equal(t, "mydb", m.config.GetString("MONGO_DB_NAME"))
	assert.Equal(t, 2, m.config.GetInt("MONGO_MIN_POOL_SIZE"))
	assert.Equal(t, 50, m.config.GetInt("MONGO_MAX_POOL_SIZE"))
	assert.Equal(t, 3*time.Minute, m.config.GetDuration("MONGO_MAX_CONN_IDLE_TIME"))
	assert.True(t, m.config.GetBool("MONGO_DEBUG"))
	assert.Equal(t, 2*time.Second, m.config.GetDuration("MONGO_PING_TIMEOUT"))
}

func TestOptions_CommandMonitor(t *testing.T) {
	monitor := &event.CommandMonitor{}
	m := NewDBModule(newMapConfig(), &nopLogger{}, WithCommandMonitor(monitor))
	assert.Same(t, monitor, m.customMonitor)
}

func TestOptions_Validation(t *testing.T) {
	cases := []struct {
		name string
		opt  Option
	}{
		{"WithDatabase empty", WithDatabase("")},
		{"WithMinPoolSize negative", WithMinPoolSize(-1)},
		{"WithMaxPoolSize zero", WithMaxPoolSize(0)},
		{"WithMaxConnIdleTime zero", WithMaxConnIdleTime(0)},
		{"WithPingTimeout zero", WithPingTimeout(0)},
		{"WithCommandMonitor nil", WithCommandMonitor(nil)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, errs := resolveSettings([]Option{tc.opt})
			assert.NotEmpty(t, errs)
		})
	}
}

func TestOptions_ErrorsFallBackToEnvironmentOnly(t *testing.T) {
	env := newMapConfig()
	env.Set("MONGO_URI", "mongodb://env-host:27017")
	env.Set("MONGO_DB_NAME", "envdb")

	m := NewDBModule(env, &nopLogger{},
		WithDatabase("codedb"),
		WithMaxPoolSize(-5),
	)

	// Any option error discards every option: the environment-only
	// configuration remains in effect.
	assert.Equal(t, "envdb", m.config.GetString("MONGO_DB_NAME"))

	err := m.ValidateConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mongodb option 2")
	assert.Contains(t, err.Error(), "WithMaxPoolSize")
}

func TestBuildClientOptions_PoolSettings(t *testing.T) {
	cfg := newMapConfig()
	cfg.Set("MONGO_URI", "mongodb://db.internal:27017")
	cfg.Set("MONGO_DB_NAME", "mydb")
	cfg.Set("MONGO_MIN_POOL_SIZE", 2)
	cfg.Set("MONGO_MAX_POOL_SIZE", 50)
	cfg.Set("MONGO_MAX_CONN_IDLE_TIME", 3*time.Minute)

	opts, dbName, err := buildClientOptions(cfg, nil, &nopLogger{})
	require.NoError(t, err)
	assert.Equal(t, "mydb", dbName)
	require.NotNil(t, opts.MinPoolSize)
	assert.Equal(t, uint64(2), *opts.MinPoolSize)
	require.NotNil(t, opts.MaxPoolSize)
	assert.Equal(t, uint64(50), *opts.MaxPoolSize)
	require.NotNil(t, opts.MaxConnIdleTime)
	assert.Equal(t, 3*time.Minute, *opts.MaxConnIdleTime)
}

func TestBuildClientOptions_EnvCredentialsFillCredentialFreeURI(t *testing.T) {
	cfg := newMapConfig()
	cfg.Set("MONGO_URI", "mongodb://db.internal:27017")
	cfg.Set("MONGO_DB_NAME", "mydb")
	cfg.Set("MONGO_USERNAME", "svc-user")
	cfg.Set("MONGO_PASSWORD", "s3cret")

	opts, _, err := buildClientOptions(cfg, nil, &nopLogger{})
	require.NoError(t, err)
	require.NotNil(t, opts.Auth)
	assert.Equal(t, "svc-user", opts.Auth.Username)
	assert.Equal(t, "s3cret", opts.Auth.Password)
}

func TestBuildClientOptions_URICredentialsWin(t *testing.T) {
	cfg := newMapConfig()
	cfg.Set("MONGO_URI", "mongodb://uri-user:uri-pass@db.internal:27017")
	cfg.Set("MONGO_DB_NAME", "mydb")
	cfg.Set("MONGO_USERNAME", "env-user")
	cfg.Set("MONGO_PASSWORD", "env-pass")

	opts, _, err := buildClientOptions(cfg, nil, &nopLogger{})
	require.NoError(t, err)
	require.NotNil(t, opts.Auth)
	assert.Equal(t, "uri-user", opts.Auth.Username)
	assert.Equal(t, "uri-pass", opts.Auth.Password)
}

func TestBuildClientOptions_NoCredentialsAnywhere(t *testing.T) {
	cfg := newMapConfig()
	cfg.Set("MONGO_URI", "mongodb://db.internal:27017")
	cfg.Set("MONGO_DB_NAME", "mydb")

	opts, _, err := buildClientOptions(cfg, nil, &nopLogger{})
	require.NoError(t, err)
	assert.Nil(t, opts.Auth)
}

func TestBuildClientOptions_CustomMonitorWinsOverDebug(t *testing.T) {
	cfg := newMapConfig()
	cfg.Set("MONGO_URI", "mongodb://db.internal:27017")
	cfg.Set("MONGO_DB_NAME", "mydb")
	cfg.Set("MONGO_DEBUG", true)

	custom := &event.CommandMonitor{}
	opts, _, err := buildClientOptions(cfg, custom, &nopLogger{})
	require.NoError(t, err)
	assert.Same(t, custom, opts.Monitor)
}

func TestBuildClientOptions_MissingURIOrDBName(t *testing.T) {
	t.Run("missing uri", func(t *testing.T) {
		cfg := newMapConfig()
		cfg.Set("MONGO_DB_NAME", "mydb")
		_, _, err := buildClientOptions(cfg, nil, &nopLogger{})
		assert.Error(t, err)
	})
	t.Run("missing db name", func(t *testing.T) {
		cfg := newMapConfig()
		cfg.Set("MONGO_URI", "mongodb://db.internal:27017")
		_, _, err := buildClientOptions(cfg, nil, &nopLogger{})
		assert.Error(t, err)
	})
}

func TestOnStart_FailsFastWhenUnreachable(t *testing.T) {
	// Port 1 refuses connections immediately: the startup ping must fail and
	// OnStart must abort instead of leaving a nil DB behind for handlers to
	// panic on later.
	env := newMapConfig()
	env.Set("MONGO_URI", "mongodb://127.0.0.1:1")
	env.Set("MONGO_DB_NAME", "mydb")
	env.Set("MONGO_PING_TIMEOUT", 500*time.Millisecond)

	m := NewDBModule(env, &nopLogger{})
	err := m.OnStart(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ping failed")
}
