package job

import (
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clientOptions extracts the resolved options from a client built by
// newRedisClient. The URL path always builds a *redis.Client, and creating a
// client performs no I/O, so these tests never need a live Redis.
func clientOptions(t *testing.T, client redis.UniversalClient) *redis.Options {
	t.Helper()
	c, ok := client.(*redis.Client)
	require.True(t, ok, "expected *redis.Client, got %T", client)
	return c.Options()
}

// mockConfig is a map-backed contract.Config for option-resolution tests,
// mirroring the one the lock package tests use.
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
	cfg.Set("JOB_CONCURRENCY", 7)
	cfg.Set("JOB_DEFAULT_QUEUE", "emails")

	resolved, _, errs := resolveConfig(cfg, []Option{WithConcurrency(9)})
	require.Empty(t, errs)

	// Code wins over the matching environment variable.
	assert.Equal(t, 9, resolved.Concurrency)
	// The environment still supplies everything code does not set.
	assert.Equal(t, "emails", resolved.DefaultQueue)
	// Defaults fill the rest.
	assert.Equal(t, 100, resolved.MaxConcurrency)
}

func TestOptions_NoOptionsMatchesLoadConfig(t *testing.T) {
	cfg := newMockConfig()
	cfg.Set("JOB_CONCURRENCY", 7)

	resolved, fromCode, errs := resolveConfig(cfg, nil)
	require.Empty(t, errs)
	assert.False(t, fromCode)
	assert.Equal(t, LoadConfig(cfg), resolved)
}

func TestOptions_Ordering(t *testing.T) {
	resolved, _, errs := resolveConfig(newMockConfig(), []Option{
		WithConcurrency(3),
		WithConcurrency(4),
	})
	require.Empty(t, errs)
	assert.Equal(t, 4, resolved.Concurrency)
}

func TestOptions_NilOptionsSkipped(t *testing.T) {
	resolved, _, errs := resolveConfig(newMockConfig(), []Option{
		nil,
		WithConcurrency(3),
	})
	require.Empty(t, errs)
	assert.Equal(t, 3, resolved.Concurrency)
}

func TestOptions_AllErrorsReportedTogether(t *testing.T) {
	_, _, errs := resolveConfig(newMockConfig(), []Option{
		WithConcurrency(0),
		WithDefaultQueue(""),
	})
	require.Len(t, errs, 2)
	assert.Contains(t, errs[0].Error(), "job option 1")
	assert.Contains(t, errs[1].Error(), "job option 2")
}

func TestOptions_Validation(t *testing.T) {
	cases := []struct {
		name string
		opt  Option
	}{
		{"WithRedisHosts empty", WithRedisHosts()},
		{"WithRedisHosts blank host", WithRedisHosts("")},
		{"WithRedisDB negative", WithRedisDB(-1)},
		{"WithRedisPoolSize zero", WithRedisPoolSize(0)},
		{"WithKeyPrefix empty", WithKeyPrefix("")},
		{"WithConcurrency zero", WithConcurrency(0)},
		{"WithMaxConcurrency zero", WithMaxConcurrency(0)},
		{"WithPollInterval zero", WithPollInterval(0)},
		{"WithShutdownTimeout zero", WithShutdownTimeout(0)},
		{"WithHeartbeatInterval zero", WithHeartbeatInterval(0)},
		{"WithDefaultQueue empty", WithDefaultQueue("")},
		{"WithDefaultMaxAttempts zero", WithDefaultMaxAttempts(0)},
		{"WithDefaultTimeout zero", WithDefaultTimeout(0)},
		{"WithRetryBackoff zero", WithRetryBackoff(0)},
		{"WithMaxRetryBackoff zero", WithMaxRetryBackoff(0)},
		{"WithRetryBackoffFactor zero", WithRetryBackoffFactor(0)},
		{"WithSchedulerInterval zero", WithSchedulerInterval(0)},
		{"WithDefaultUniqueTTL zero", WithDefaultUniqueTTL(0)},
		{"WithDLQTTL zero", WithDLQTTL(0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, errs := resolveConfig(newMockConfig(), []Option{tc.opt})
			assert.NotEmpty(t, errs)
		})
	}
}

func TestOptions_AllFieldsApply(t *testing.T) {
	resolved, fromCode, errs := resolveConfig(newMockConfig(), []Option{
		WithRedisHosts("redis-a:6379", "redis-b:6379"),
		WithRedisDB(3),
		WithRedisPoolSize(20),
		WithKeyPrefix("myapp:job:"),
		WithConcurrency(10),
		WithMaxConcurrency(200),
		WithPollInterval(2 * time.Second),
		WithShutdownTimeout(time.Minute),
		WithHeartbeatInterval(30 * time.Second),
		WithDefaultQueue("critical"),
		WithDefaultMaxAttempts(5),
		WithDefaultTimeout(10 * time.Minute),
		WithRetryBackoff(2 * time.Second),
		WithMaxRetryBackoff(10 * time.Minute),
		WithRetryBackoffFactor(3.0),
		WithSchedulerEnabled(false),
		WithSchedulerInterval(5 * time.Second),
		WithDefaultUniqueTTL(2 * time.Hour),
		WithDLQEnabled(false),
		WithDLQTTL(48 * time.Hour),
		WithMetricsEnabled(false),
	})
	require.Empty(t, errs)
	assert.True(t, fromCode)

	assert.Equal(t, []string{"redis-a:6379", "redis-b:6379"}, resolved.RedisHosts)
	assert.Equal(t, 3, resolved.RedisDB)
	assert.Equal(t, 20, resolved.RedisPoolSize)
	assert.Equal(t, "myapp:job:", resolved.KeyPrefix)
	assert.Equal(t, 10, resolved.Concurrency)
	assert.Equal(t, 200, resolved.MaxConcurrency)
	assert.Equal(t, 2*time.Second, resolved.PollInterval)
	assert.Equal(t, time.Minute, resolved.ShutdownTimeout)
	assert.Equal(t, 30*time.Second, resolved.HeartbeatInterval)
	assert.Equal(t, "critical", resolved.DefaultQueue)
	assert.Equal(t, 5, resolved.DefaultMaxAttempts)
	assert.Equal(t, 10*time.Minute, resolved.DefaultTimeout)
	assert.Equal(t, 2*time.Second, resolved.RetryBackoff)
	assert.Equal(t, 10*time.Minute, resolved.MaxRetryBackoff)
	assert.Equal(t, 3.0, resolved.RetryBackoffFactor)
	assert.False(t, resolved.SchedulerEnabled)
	assert.Equal(t, 5*time.Second, resolved.SchedulerInterval)
	assert.Equal(t, 2*time.Hour, resolved.DefaultUniqueTTL)
	assert.False(t, resolved.DLQEnabled)
	assert.Equal(t, 48*time.Hour, resolved.DLQTTL)
	assert.False(t, resolved.MetricsEnabled)
}

func TestOptions_RedisHostsOverridesEnvURL(t *testing.T) {
	// Code wins: hosts set in code must beat JOB_REDIS_URL from the
	// environment, which the manager would otherwise prefer.
	cfg := newMockConfig()
	cfg.Set("JOB_REDIS_URL", "redis://env-host:6379/0")

	resolved, fromCode, errs := resolveConfig(cfg, []Option{
		WithRedisHosts("code-host:6379"),
	})
	require.Empty(t, errs)
	assert.True(t, fromCode)
	assert.Empty(t, resolved.RedisURL)
	assert.Equal(t, []string{"code-host:6379"}, resolved.RedisHosts)
	assert.Equal(t, "code-host:6379", resolved.RedactedTarget())
}

func TestOptions_EnvCredentialsApplyWithCodeHosts(t *testing.T) {
	// Credentials are environment-only. Setting the endpoint in code must not
	// lose the credentials supplied by the environment.
	cfg := newMockConfig()
	cfg.Set("JOB_REDIS_USERNAME", "svc-user")
	cfg.Set("JOB_REDIS_PASSWORD", "s3cret")

	resolved, _, errs := resolveConfig(cfg, []Option{
		WithRedisHosts("code-host:6379"),
	})
	require.Empty(t, errs)
	assert.Equal(t, "svc-user", resolved.RedisUsername)
	assert.Equal(t, "s3cret", resolved.RedisPassword)
}

func TestLoadConfig_CredentialsReadAlongsideURL(t *testing.T) {
	// JOB_REDIS_USERNAME/JOB_REDIS_PASSWORD must be honored even when
	// JOB_REDIS_URL is also set; previously they were silently ignored.
	cfg := newMockConfig()
	cfg.Set("JOB_REDIS_URL", "redis://redis.internal:6379/0")
	cfg.Set("JOB_REDIS_USERNAME", "svc-user")
	cfg.Set("JOB_REDIS_PASSWORD", "s3cret")

	loaded := LoadConfig(cfg)
	assert.Equal(t, "redis://redis.internal:6379/0", loaded.RedisURL)
	assert.Equal(t, "svc-user", loaded.RedisUsername)
	assert.Equal(t, "s3cret", loaded.RedisPassword)
}

func TestNewRedisClient_EnvURLPassthrough(t *testing.T) {
	// JOB_REDIS_URL alone drives the connection exactly as it always has:
	// address and database from the URL, no credentials invented.
	cfg := newMockConfig()
	cfg.Set("JOB_REDIS_URL", "redis://redis.internal:6399/3")

	client, err := newRedisClient(LoadConfig(cfg))
	require.NoError(t, err)
	defer client.Close()

	opts := clientOptions(t, client)
	assert.Equal(t, "redis.internal:6399", opts.Addr)
	assert.Equal(t, 3, opts.DB)
	assert.Empty(t, opts.Username)
	assert.Empty(t, opts.Password)
}

func TestNewRedisClient_URLCredentialFill(t *testing.T) {
	t.Run("env credentials fill a credential-free URL", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.RedisURL = "redis://redis.internal:6379/2"
		cfg.RedisUsername = "svc-user"
		cfg.RedisPassword = "s3cret"

		client, err := newRedisClient(cfg)
		require.NoError(t, err)
		defer client.Close()

		opts := clientOptions(t, client)
		assert.Equal(t, "svc-user", opts.Username)
		assert.Equal(t, "s3cret", opts.Password)
		assert.Equal(t, 2, opts.DB)
	})

	t.Run("credentials embedded in the URL win", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.RedisURL = "redis://url-user:url-pass@redis.internal:6379/0"
		cfg.RedisUsername = "env-user"
		cfg.RedisPassword = "env-pass"

		client, err := newRedisClient(cfg)
		require.NoError(t, err)
		defer client.Close()

		opts := clientOptions(t, client)
		assert.Equal(t, "url-user", opts.Username)
		assert.Equal(t, "url-pass", opts.Password)
	})
}

func TestNewModule_OptionErrorFailsFastWithoutConnecting(t *testing.T) {
	// An invalid option must abort module construction before any Redis
	// connection is attempted, and the error must identify the option.
	_, err := NewModule(newMockConfig(), &nopLogger{}, WithConcurrency(-1))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "job option 1")
	assert.Contains(t, err.Error(), "WithConcurrency")
}

func TestModule_ValidateConfig_CodeConnectionSatisfiesRequirement(t *testing.T) {
	t.Run("connection configured in code passes without env keys", func(t *testing.T) {
		m := &Module{config: newMockConfig(), connectionSetInCode: true}
		assert.NoError(t, m.ValidateConfig())
	})

	t.Run("no env keys and no code connection still fails", func(t *testing.T) {
		m := &Module{config: newMockConfig()}
		err := m.ValidateConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "JOB_REDIS_ADDR")
	})
}

func TestOptions_ErrorMessagesNameTheOption(t *testing.T) {
	_, _, errs := resolveConfig(newMockConfig(), []Option{WithRedisPoolSize(-5)})
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "WithRedisPoolSize")
	var joined error = errors.Join(errs...)
	assert.Contains(t, joined.Error(), "job option 1")
}
