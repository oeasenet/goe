package lock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptions_Precedence(t *testing.T) {
	cfg := newMockConfig()
	cfg.Set("LOCK_DEFAULT_EXPIRY", 2*time.Second)
	cfg.Set("LOCK_KEY_PREFIX", "env:")

	resolved, _, errs := resolveConfig(cfg, []Option{
		WithDefaultExpiry(10 * time.Second),
	})
	require.Empty(t, errs)

	// Code wins over the matching environment variable.
	assert.Equal(t, 10*time.Second, resolved.DefaultExpiry)
	// The environment still supplies everything code does not set.
	assert.Equal(t, "env:", resolved.DefaultKeyPrefix)
	// Defaults fill the rest.
	assert.Equal(t, 32, resolved.DefaultTries)
}

func TestOptions_NoOptionsMatchesLoadConfig(t *testing.T) {
	cfg := newMockConfig()
	cfg.Set("LOCK_DEFAULT_TRIES", 5)

	resolved, fromCode, errs := resolveConfig(cfg, nil)
	require.Empty(t, errs)
	assert.False(t, fromCode)
	assert.Equal(t, LoadConfig(cfg), resolved)
}

func TestOptions_Ordering(t *testing.T) {
	resolved, _, errs := resolveConfig(newMockConfig(), []Option{
		WithDefaultTries(3),
		WithDefaultTries(4),
	})
	require.Empty(t, errs)
	assert.Equal(t, 4, resolved.DefaultTries)
}

func TestOptions_NilOptionsSkipped(t *testing.T) {
	resolved, _, errs := resolveConfig(newMockConfig(), []Option{
		nil,
		WithDefaultTries(3),
	})
	require.Empty(t, errs)
	assert.Equal(t, 3, resolved.DefaultTries)
}

func TestOptions_AllErrorsReportedTogether(t *testing.T) {
	_, _, errs := resolveConfig(newMockConfig(), []Option{
		WithDefaultTries(0),
		WithPoolSize(0),
	})
	require.Len(t, errs, 2)
	assert.Contains(t, errs[0].Error(), "lock option 1")
	assert.Contains(t, errs[1].Error(), "lock option 2")
}

func TestOptions_AllFieldsApply(t *testing.T) {
	resolved, fromCode, errs := resolveConfig(newMockConfig(), []Option{
		WithRedisURL("redis-sentinel://mymaster@s1:26379,s2:26379/0"),
		WithDefaultExpiry(12 * time.Second),
		WithDefaultTries(16),
		WithDefaultRetryDelay(250 * time.Millisecond),
		WithDefaultDriftFactor(0.02),
		WithKeyPrefix("myapp:lock:"),
		WithPoolSize(25),
		WithTLSInsecureSkipVerify(true),
	})
	require.Empty(t, errs)
	assert.True(t, fromCode)

	assert.Equal(t, "redis-sentinel://mymaster@s1:26379,s2:26379/0", resolved.RedisURL)
	assert.Equal(t, 12*time.Second, resolved.DefaultExpiry)
	assert.Equal(t, 16, resolved.DefaultTries)
	assert.Equal(t, 250*time.Millisecond, resolved.DefaultRetryDelay)
	assert.Equal(t, 0.02, resolved.DefaultDriftFactor)
	assert.Equal(t, "myapp:lock:", resolved.DefaultKeyPrefix)
	assert.Equal(t, 25, resolved.PoolSize)
	assert.True(t, resolved.TLSInsecureSkipVerify)
}

func TestWithRedisURL_RejectsEmbeddedCredentials(t *testing.T) {
	cases := []struct {
		name string
		url  string
	}{
		{"single with user and password", "redis://user:pass@host:6379/0"},
		{"single with username only", "redis://user@host:6379/0"},
		{"tls with credentials", "rediss://user:pass@host:6379/0"},
		{"sentinel with credentials", "redis-sentinel://user:pass@mymaster@s1:26379/0"},
		{"cluster with credentials", "redis-cluster://user:pass@n1:6379,n2:6379"},
		{"cluster with username only", "redis-cluster://user@n1:6379,n2:6379"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, errs := resolveConfig(newMockConfig(), []Option{WithRedisURL(tc.url)})
			require.NotEmpty(t, errs)
			// The error must teach where credentials belong.
			assert.Contains(t, errs[0].Error(), "LOCK_REDIS_USERNAME")
			assert.Contains(t, errs[0].Error(), "LOCK_REDIS_PASSWORD")
		})
	}
}

func TestWithRedisURL_AcceptsCredentialFreeURLs(t *testing.T) {
	cases := []struct {
		name string
		url  string
	}{
		{"single", "redis://host:6379/0"},
		{"tls", "rediss://host:6380/1"},
		{"sentinel with master name", "redis-sentinel://mymaster@s1:26379,s2:26379/0"},
		{"cluster", "redis-cluster://n1:6379,n2:6379"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resolved, fromCode, errs := resolveConfig(newMockConfig(), []Option{WithRedisURL(tc.url)})
			require.Empty(t, errs)
			assert.True(t, fromCode)
			assert.Equal(t, tc.url, resolved.RedisURL)
		})
	}
}

func TestWithRedisURL_RejectsInvalidURLs(t *testing.T) {
	cases := []struct {
		name string
		url  string
	}{
		{"empty", ""},
		{"missing scheme", "host:6379"},
		{"unsupported scheme", "http://host:6379"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, errs := resolveConfig(newMockConfig(), []Option{WithRedisURL(tc.url)})
			assert.NotEmpty(t, errs)
		})
	}
}

func TestWithRedisURL_OverridesEnvURLs(t *testing.T) {
	// Code wins: a single URL set in code must beat LOCK_REDIS_URLS from the
	// environment, which the manager would otherwise prefer.
	cfg := newMockConfig()
	cfg.Set("LOCK_REDIS_URLS", []string{"redis://env-a:6379/0", "redis://env-b:6379/0"})

	resolved, _, errs := resolveConfig(cfg, []Option{
		WithRedisURL("redis://code-host:6379/0"),
	})
	require.Empty(t, errs)
	assert.Equal(t, "redis://code-host:6379/0", resolved.RedisURL)
	assert.Empty(t, resolved.RedisURLs)
}

func TestWithRedisURLs_RedlockRules(t *testing.T) {
	t.Run("accepts multiple single-instance URLs", func(t *testing.T) {
		resolved, fromCode, errs := resolveConfig(newMockConfig(), []Option{
			WithRedisURLs("redis://a:6379/0", "rediss://b:6380/0", "redis://c:6379/0"),
		})
		require.Empty(t, errs)
		assert.True(t, fromCode)
		assert.Len(t, resolved.RedisURLs, 3)
	})

	t.Run("rejects sentinel and cluster schemes", func(t *testing.T) {
		for _, url := range []string{
			"redis-sentinel://mymaster@s1:26379/0",
			"redis-cluster://n1:6379,n2:6379",
		} {
			_, _, errs := resolveConfig(newMockConfig(), []Option{
				WithRedisURLs("redis://a:6379/0", url),
			})
			require.NotEmpty(t, errs, "url %s should be rejected", url)
			assert.Contains(t, errs[0].Error(), "redis:// or rediss://")
		}
	})

	t.Run("rejects embedded credentials", func(t *testing.T) {
		_, _, errs := resolveConfig(newMockConfig(), []Option{
			WithRedisURLs("redis://u:p@a:6379/0"),
		})
		require.NotEmpty(t, errs)
		assert.Contains(t, errs[0].Error(), "LOCK_REDIS_PASSWORD")
	})

	t.Run("rejects empty list", func(t *testing.T) {
		_, _, errs := resolveConfig(newMockConfig(), []Option{WithRedisURLs()})
		assert.NotEmpty(t, errs)
	})
}

func TestOptions_Validation(t *testing.T) {
	cases := []struct {
		name string
		opt  Option
	}{
		{"WithDefaultExpiry zero", WithDefaultExpiry(0)},
		{"WithDefaultTries zero", WithDefaultTries(0)},
		{"WithDefaultRetryDelay zero", WithDefaultRetryDelay(0)},
		{"WithDefaultDriftFactor zero", WithDefaultDriftFactor(0)},
		{"WithDefaultDriftFactor one", WithDefaultDriftFactor(1)},
		{"WithKeyPrefix empty", WithKeyPrefix("")},
		{"WithPoolSize zero", WithPoolSize(0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, errs := resolveConfig(newMockConfig(), []Option{tc.opt})
			assert.NotEmpty(t, errs)
		})
	}
}

func TestLoadConfig_EnvCredentials(t *testing.T) {
	cfg := newMockConfig()
	cfg.Set("LOCK_REDIS_URL", "redis://redis.internal:6379/0")
	cfg.Set("LOCK_REDIS_USERNAME", "svc-user")
	cfg.Set("LOCK_REDIS_PASSWORD", "s3cret")

	loaded := LoadConfig(cfg)
	assert.Equal(t, "svc-user", loaded.RedisUsername)
	assert.Equal(t, "s3cret", loaded.RedisPassword)
}

func TestCreatePools_EnvURLPassthrough(t *testing.T) {
	// LOCK_REDIS_URL alone drives the connection exactly as it always has:
	// address and database from the URL, no credentials invented.
	cfg := newMockConfig()
	cfg.Set("LOCK_REDIS_URL", "redis://redis.internal:6399/3")

	pools, mode, err := createPools(LoadConfig(cfg))
	require.NoError(t, err)
	require.Equal(t, "single", mode)
	require.Len(t, pools, 1)
	defer pools[0].Close()

	opts := pools[0].(*ClientPool).Client().Options()
	assert.Equal(t, "redis.internal:6399", opts.Addr)
	assert.Equal(t, 3, opts.DB)
	assert.Empty(t, opts.Username)
	assert.Empty(t, opts.Password)
}

func TestCreatePools_EnvCredentialOverlay(t *testing.T) {
	t.Run("single pool picks up env credentials", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.RedisURL = "redis://redis.internal:6379/2"
		cfg.RedisUsername = "svc-user"
		cfg.RedisPassword = "s3cret"

		pools, mode, err := createPools(cfg)
		require.NoError(t, err)
		require.Equal(t, "single", mode)
		require.Len(t, pools, 1)
		defer pools[0].Close()

		opts := pools[0].(*ClientPool).Client().Options()
		assert.Equal(t, "svc-user", opts.Username)
		assert.Equal(t, "s3cret", opts.Password)
		assert.Equal(t, 2, opts.DB)
	})

	t.Run("url credentials win over env credentials", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.RedisURL = "redis://url-user:url-pass@redis.internal:6379/0"
		cfg.RedisUsername = "env-user"
		cfg.RedisPassword = "env-pass"

		pools, _, err := createPools(cfg)
		require.NoError(t, err)
		require.Len(t, pools, 1)
		defer pools[0].Close()

		opts := pools[0].(*ClientPool).Client().Options()
		assert.Equal(t, "url-user", opts.Username)
		assert.Equal(t, "url-pass", opts.Password)
	})

	t.Run("sentinel pool picks up env credentials", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.RedisUsername = "svc-user"
		cfg.RedisPassword = "s3cret"

		pools, mode, err := createSentinelPool("redis-sentinel://mymaster@s1:26379,s2:26379/1", cfg)
		require.NoError(t, err)
		require.Equal(t, "sentinel", mode)
		require.Len(t, pools, 1)
		defer pools[0].Close()

		opts := pools[0].(*FailoverPool).Client().Options()
		assert.Equal(t, "svc-user", opts.Username)
		assert.Equal(t, "s3cret", opts.Password)
	})

	t.Run("cluster pool picks up env credentials", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.RedisUsername = "svc-user"
		cfg.RedisPassword = "s3cret"

		pools, mode, err := createClusterPool("redis-cluster://n1:6379,n2:6379", cfg)
		require.NoError(t, err)
		require.Equal(t, "cluster", mode)
		require.Len(t, pools, 1)
		defer pools[0].Close()

		opts := pools[0].(*ClusterPool).Client().Options()
		assert.Equal(t, "svc-user", opts.Username)
		assert.Equal(t, "s3cret", opts.Password)
	})
}

func TestNewModule_OptionErrorFailsFastWithoutConnecting(t *testing.T) {
	// An invalid option must abort module construction before any Redis
	// connection (and its ping) is attempted.
	_, err := NewModule(newMockConfig(), &testLogger{t: t}, WithPoolSize(-1))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lock option 1")
	assert.Contains(t, err.Error(), "WithPoolSize")
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
		assert.Contains(t, err.Error(), "LOCK_REDIS_URL")
	})
}
