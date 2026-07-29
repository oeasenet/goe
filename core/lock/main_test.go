package lock

import (
	"net"
	"os"
	"testing"
	"time"
)

// testRedisAddr is the Redis instance the lock tests use.
//
// This was previously a hardcoded LAN address, which meant that on any machine
// other than the one it belonged to, every test in this package waited for a
// TCP timeout before skipping — roughly two minutes for the package as a whole.
// It now defaults to the address examples/docker-compose.yml publishes, and can
// be redirected with GOE_TEST_REDIS_ADDR.
var testRedisAddr = envOr("GOE_TEST_REDIS_ADDR", "127.0.0.1:6379")

// redisAvailable is probed once for the whole package rather than once per test,
// so a missing Redis costs a single short dial instead of a multi-second
// connection attempt in every test that needs it.
var redisAvailable bool

func TestMain(m *testing.M) {
	redisAvailable = probeRedis(testRedisAddr, 500*time.Millisecond)
	os.Exit(m.Run())
}

// probeRedis reports whether something is accepting connections at addr.
//
// A plain TCP dial is deliberate: it is fast, and if the port is open but Redis
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
		t.Skipf("Redis not reachable at %s; start it (examples/docker-compose.yml) "+
			"or set GOE_TEST_REDIS_ADDR", testRedisAddr)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
