package migrate

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

// mockConfig is a map-backed contract.Config for option-resolution tests.
type mockConfig struct {
	values map[string]any
}

func newMockConfig() *mockConfig {
	return &mockConfig{values: make(map[string]any)}
}

func (m *mockConfig) Get(key string) any { return m.values[key] }

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

func (m *mockConfig) Set(key string, value any) { m.values[key] = value }

func (m *mockConfig) Has(key string) bool {
	_, ok := m.values[key]
	return ok
}

func (m *mockConfig) All() map[string]any { return m.values }

func (m *mockConfig) Reload() error { return nil }

func TestOptions_Precedence(t *testing.T) {
	cfg := newMockConfig()
	cfg.Set("MONGODB_MIGRATE_TIMEOUT", time.Minute)
	cfg.Set("MONGODB_MIGRATE_COLLECTION", "env_migrations")

	resolved, errs := resolveConfig(cfg, []Option{
		WithTimeout(10 * time.Minute),
	})
	require.Empty(t, errs)

	// Code wins over the matching environment variable.
	assert.Equal(t, 10*time.Minute, resolved.Timeout)
	// The environment still supplies everything code does not set.
	assert.Equal(t, "env_migrations", resolved.Collection)
	// Defaults fill the rest.
	assert.Equal(t, 30*time.Minute, resolved.LockTimeout)
}

func TestOptions_NoOptionsMatchesLoadConfig(t *testing.T) {
	cfg := newMockConfig()
	cfg.Set("MONGODB_MIGRATE_AUTO", true)

	resolved, errs := resolveConfig(cfg, nil)
	require.Empty(t, errs)
	assert.Equal(t, LoadConfig(cfg), resolved)
}

func TestOptions_Ordering(t *testing.T) {
	resolved, errs := resolveConfig(newMockConfig(), []Option{
		WithCollection("first"),
		WithCollection("second"),
	})
	require.Empty(t, errs)
	assert.Equal(t, "second", resolved.Collection)
}

func TestOptions_NilOptionsSkipped(t *testing.T) {
	resolved, errs := resolveConfig(newMockConfig(), []Option{
		nil,
		WithCollection("named"),
	})
	require.Empty(t, errs)
	assert.Equal(t, "named", resolved.Collection)
}

func TestOptions_AllErrorsReportedTogether(t *testing.T) {
	_, errs := resolveConfig(newMockConfig(), []Option{
		WithCollection(""),
		WithTimeout(0),
	})
	require.Len(t, errs, 2)
	assert.Contains(t, errs[0].Error(), "migrate option 1")
	assert.Contains(t, errs[1].Error(), "migrate option 2")
}

func TestOptions_AllFieldsApply(t *testing.T) {
	resolved, errs := resolveConfig(newMockConfig(), []Option{
		WithCollection("my_migrations"),
		WithTimeout(10 * time.Minute),
		WithLockTimeout(time.Hour),
		WithLockHeartbeat(time.Minute),
		WithUseTransactions(false),
		WithVerifyChecksums(false),
		WithAutoMigrate(true),
		WithVersionScheme("timestamp"),
		WithSchemaVersionField("_sv"),
		WithDryRunByDefault(true),
	})
	require.Empty(t, errs)

	assert.Equal(t, "my_migrations", resolved.Collection)
	assert.Equal(t, 10*time.Minute, resolved.Timeout)
	assert.Equal(t, time.Hour, resolved.LockTimeout)
	assert.Equal(t, time.Minute, resolved.LockHeartbeat)
	assert.False(t, resolved.UseTransactions)
	assert.False(t, resolved.VerifyChecksums)
	assert.True(t, resolved.AutoMigrate)
	assert.Equal(t, "timestamp", resolved.VersionScheme)
	assert.Equal(t, "_sv", resolved.SchemaVersionField)
	assert.True(t, resolved.DryRunByDefault)
}

func TestOptions_Validation(t *testing.T) {
	cases := []struct {
		name string
		opt  Option
	}{
		{"WithCollection empty", WithCollection("")},
		{"WithTimeout zero", WithTimeout(0)},
		{"WithLockTimeout zero", WithLockTimeout(0)},
		{"WithLockHeartbeat zero", WithLockHeartbeat(0)},
		{"WithVersionScheme unknown", WithVersionScheme("weird")},
		{"WithVersionScheme empty", WithVersionScheme("")},
		{"WithSchemaVersionField empty", WithSchemaVersionField("")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, errs := resolveConfig(newMockConfig(), []Option{tc.opt})
			assert.NotEmpty(t, errs)
		})
	}
}

func TestNewModule_OptionErrorFailsFast(t *testing.T) {
	// An invalid option must abort module construction with the option named.
	_, err := NewModule(newMockConfig(), &nopLogger{}, nil, WithTimeout(-time.Second))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "migrate option 1")
	assert.Contains(t, err.Error(), "WithTimeout")
}
