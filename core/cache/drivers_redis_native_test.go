package cache

import (
	"context"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/config"
)

// The redis driver's native atomic operations are exercised in the default
// test lane the same way core/lock uses Redis: probe once per package, run
// wherever a Redis happens to be listening (CI provides one), skip cheaply
// where none is. The deeper end-to-end suites stay in the integration lane.

// testRedisAddr is the Redis instance these tests use, redirectable with
// GOE_TEST_REDIS_ADDR like every other probed suite.
var testRedisAddr = envOr("GOE_TEST_REDIS_ADDR", "127.0.0.1:6379")

// redisAvailable is probed once for the whole package, so a missing Redis
// costs a single short dial instead of a timeout per test.
var redisAvailable bool

func TestMain(m *testing.M) {
	redisAvailable = probeRedis(testRedisAddr, 500*time.Millisecond)
	os.Exit(m.Run())
}

// probeRedis reports whether something is accepting connections at addr. A
// plain TCP dial is deliberate: it is fast, and if the port is open but Redis
// is unhealthy the tests should fail loudly rather than quietly skip.
func probeRedis(addr string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// requireRedis skips the calling test immediately when Redis is unreachable.
func requireRedis(t *testing.T) {
	t.Helper()
	if !redisAvailable {
		t.Skipf("Redis not reachable at %s; start it (docker-compose.test.yml) "+
			"or set GOE_TEST_REDIS_ADDR", testRedisAddr)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func TestRedisStoreNativeAtomicOps(t *testing.T) {
	requireRedis(t)

	host, portStr, err := net.SplitHostPort(testRedisAddr)
	require.NoError(t, err)
	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)

	cfg := config.NewModule()
	cfg.Provide().Set("CACHE_REDIS_HOST", host)
	cfg.Provide().Set("CACHE_REDIS_PORT", port)
	// Database 14 keeps clear of the integration suite's database 15.
	cfg.Provide().Set("CACHE_REDIS_DATABASE", 14)

	store, err := RedisStoreFactory(cfg.Provide())
	require.NoError(t, err)

	native, ok := store.(contract.AtomicCacheStore)
	require.True(t, ok, "redis store must expose the native atomic capability")

	// Independent client so TTL assertions run against Redis itself. Cleanup
	// deletes only this test's keys — no FLUSHDB, the database may be shared.
	verify := goredis.NewClient(&goredis.Options{Addr: testRedisAddr, DB: 14})
	t.Cleanup(func() { _ = verify.Close() })
	ctx := context.Background()
	require.NoError(t, verify.Del(ctx, "native:new", "native:ttl", "native:add", "native:pull").Err())

	t.Run("increment creates without expiration and preserves an existing TTL", func(t *testing.T) {
		n, err := native.Increment("native:new", 2)
		require.NoError(t, err)
		assert.Equal(t, int64(2), n)

		ttl, err := verify.TTL(ctx, "native:new").Result()
		require.NoError(t, err)
		assert.Less(t, ttl, time.Duration(0), "key created by Increment must have no expiration")

		require.NoError(t, store.Set("native:ttl", []byte("5"), 30*time.Second))
		n, err = native.Increment("native:ttl", 1)
		require.NoError(t, err)
		assert.Equal(t, int64(6), n)

		ttl, err = verify.TTL(ctx, "native:ttl").Result()
		require.NoError(t, err)
		assert.Greater(t, ttl, time.Duration(0), "INCRBY must not strip the TTL")
	})

	t.Run("set-if-not-exists admits exactly once", func(t *testing.T) {
		stored, err := native.SetIfNotExists("native:add", []byte(`"first"`), 30*time.Second)
		require.NoError(t, err)
		assert.True(t, stored)

		stored, err = native.SetIfNotExists("native:add", []byte(`"second"`), 30*time.Second)
		require.NoError(t, err)
		assert.False(t, stored, "second write must be rejected")
	})

	t.Run("get-delete consumes and reports misses as nil, nil", func(t *testing.T) {
		require.NoError(t, store.Set("native:pull", []byte(`"v"`), 30*time.Second))

		data, err := native.GetDelete("native:pull")
		require.NoError(t, err)
		assert.Equal(t, []byte(`"v"`), data)

		data, err = native.GetDelete("native:pull")
		require.NoError(t, err)
		assert.Nil(t, data, "consumed key must miss")
	})
}
