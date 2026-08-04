package cache

import (
	"testing"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestBuildRedisConfig_EnvURLPassthrough(t *testing.T) {
	// The URL from the environment drives the store exactly as it always
	// has: untouched, query parameters and all.
	cfg := newMapConfig()
	cfg.Set("CACHE_REDIS_URL", "redis://redis.internal:6399/3?dial_timeout=3s")

	rc := buildRedisConfig(cfg)
	assert.Equal(t, "redis://redis.internal:6399/3?dial_timeout=3s", rc.URL)
	assert.Empty(t, rc.Username)
	assert.Empty(t, rc.Password)
}

func TestBuildRedisConfig_EnvCredentialsMergeIntoURL(t *testing.T) {
	// CACHE_REDIS_USERNAME/CACHE_REDIS_PASSWORD must be honored even when
	// CACHE_REDIS_URL is set: the Fiber storage ignores discrete credentials
	// once a URL is present, so the credentials are injected into the URL
	// itself — everything else about the URL is preserved.
	cfg := newMapConfig()
	cfg.Set("CACHE_REDIS_URL", "redis://redis.internal:6399/3")
	cfg.Set("CACHE_REDIS_USERNAME", "svc-user")
	cfg.Set("CACHE_REDIS_PASSWORD", "s3cret")

	rc := buildRedisConfig(cfg)
	assert.Equal(t, "redis://svc-user:s3cret@redis.internal:6399/3", rc.URL)
}

func TestBuildRedisConfig_PasswordOnlyMergesIntoURL(t *testing.T) {
	cfg := newMapConfig()
	cfg.Set("CACHE_REDIS_URL", "redis://redis.internal:6379/0")
	cfg.Set("CACHE_REDIS_PASSWORD", "s3cret")

	rc := buildRedisConfig(cfg)
	assert.Equal(t, "redis://:s3cret@redis.internal:6379/0", rc.URL)
}

func TestBuildRedisConfig_MergePreservesQueryAndScheme(t *testing.T) {
	// Injection must not disturb the rest of the URL: rediss:// keeps its
	// TLS-selecting scheme and query parameters survive.
	cfg := newMapConfig()
	cfg.Set("CACHE_REDIS_URL", "rediss://redis.internal:6380/1?dial_timeout=3s")
	cfg.Set("CACHE_REDIS_PASSWORD", "s3cret")

	rc := buildRedisConfig(cfg)
	assert.Equal(t, "rediss://:s3cret@redis.internal:6380/1?dial_timeout=3s", rc.URL)
}

func TestBuildRedisConfig_URLCredentialsWin(t *testing.T) {
	cfg := newMapConfig()
	cfg.Set("CACHE_REDIS_URL", "redis://url-user:url-pass@redis.internal:6379/0")
	cfg.Set("CACHE_REDIS_USERNAME", "env-user")
	cfg.Set("CACHE_REDIS_PASSWORD", "env-pass")

	// A URL that embeds its own credentials is passed through untouched.
	rc := buildRedisConfig(cfg)
	assert.Equal(t, "redis://url-user:url-pass@redis.internal:6379/0", rc.URL)
}

func TestBuildRedisConfig_SpecialCharactersAreEncoded(t *testing.T) {
	cfg := newMapConfig()
	cfg.Set("CACHE_REDIS_URL", "redis://redis.internal:6379/0")
	cfg.Set("CACHE_REDIS_PASSWORD", "p@ss/w:rd")

	// The injected credentials must be percent-encoded so the URL stays
	// parseable and round-trips to the original secret.
	rc := buildRedisConfig(cfg)
	opt, err := goredis.ParseURL(rc.URL)
	assert.NoError(t, err)
	assert.Equal(t, "p@ss/w:rd", opt.Password)
}

func TestBuildRedisConfig_DiscreteFieldsWithCredentials(t *testing.T) {
	cfg := newMapConfig()
	cfg.Set("CACHE_REDIS_HOST", "redis.internal")
	cfg.Set("CACHE_REDIS_PORT", 6380)
	cfg.Set("CACHE_REDIS_USERNAME", "svc-user")
	cfg.Set("CACHE_REDIS_PASSWORD", "s3cret")

	rc := buildRedisConfig(cfg)
	assert.Equal(t, "redis.internal", rc.Host)
	assert.Equal(t, 6380, rc.Port)
	assert.Equal(t, "svc-user", rc.Username)
	assert.Equal(t, "s3cret", rc.Password)
}

func TestBuildRedisConfig_InvalidURLPassedThrough(t *testing.T) {
	// An unparseable URL is left for the Fiber storage to reject with its own
	// error, exactly as before.
	cfg := newMapConfig()
	cfg.Set("CACHE_REDIS_URL", "redis://not a url::")
	cfg.Set("CACHE_REDIS_PASSWORD", "s3cret")

	rc := buildRedisConfig(cfg)
	assert.Equal(t, "redis://not a url::", rc.URL)
}
