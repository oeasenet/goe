package lock

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
)

func TestNewManager(t *testing.T) {
	logger := &testLogger{t: t}

	config := &Config{
		RedisURL:           "redis://" + testRedisAddr + "/0",
		DefaultExpiry:      8 * time.Second,
		DefaultTries:       32,
		DefaultRetryDelay:  500 * time.Millisecond,
		DefaultDriftFactor: 0.01,
		DefaultKeyPrefix:   "test:lock:",
		PoolSize:           10,
	}

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer manager.Close(context.Background())

	// Manager should be created successfully
	assert.NotNil(t, manager)
	assert.NotEmpty(t, manager.Pools())
	assert.Equal(t, "single", manager.Mode())
}

func TestManager_Health(t *testing.T) {
	logger := &testLogger{t: t}

	config := &Config{
		RedisURL:           "redis://" + testRedisAddr + "/0",
		DefaultExpiry:      8 * time.Second,
		DefaultTries:       32,
		DefaultRetryDelay:  500 * time.Millisecond,
		DefaultDriftFactor: 0.01,
		DefaultKeyPrefix:   "test:lock:",
	}

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer manager.Close(context.Background())

	ctx := context.Background()
	err = manager.Health(ctx)
	require.NoError(t, err)
}

func TestManager_NewMutex(t *testing.T) {
	logger := &testLogger{t: t}

	config := &Config{
		RedisURL:           "redis://" + testRedisAddr + "/0",
		DefaultExpiry:      8 * time.Second,
		DefaultTries:       32,
		DefaultRetryDelay:  500 * time.Millisecond,
		DefaultDriftFactor: 0.01,
		DefaultKeyPrefix:   "test:lock:",
	}

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer manager.Close(context.Background())

	// Clean up using the pool
	pool := manager.Pools()[0]
	if cp, ok := pool.(*ClientPool); ok {
		cleanupKeys(t, cp.Client(), "test:lock:*")
		defer cleanupKeys(t, cp.Client(), "test:lock:*")
	}

	// Create mutex
	mutex := manager.NewMutex("my-resource")
	require.NotNil(t, mutex)

	// Name should have prefix
	assert.Equal(t, "test:lock:my-resource", mutex.Name())

	// Should be able to lock
	ctx := context.Background()
	err = mutex.Lock(ctx)
	require.NoError(t, err)

	// Should be able to unlock
	err = mutex.Unlock(ctx)
	require.NoError(t, err)
}

func TestManager_NewMutexWithOptions(t *testing.T) {
	logger := &testLogger{t: t}

	config := &Config{
		RedisURL:           "redis://" + testRedisAddr + "/0",
		DefaultExpiry:      8 * time.Second,
		DefaultTries:       32,
		DefaultRetryDelay:  500 * time.Millisecond,
		DefaultDriftFactor: 0.01,
		DefaultKeyPrefix:   "test:lock:",
	}

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer manager.Close(context.Background())

	// Clean up using the pool
	pool := manager.Pools()[0]
	if cp, ok := pool.(*ClientPool); ok {
		cleanupKeys(t, cp.Client(), "test:lock:*")
		defer cleanupKeys(t, cp.Client(), "test:lock:*")
	}

	// Create mutex with custom options
	mutex := manager.NewMutex("custom-resource",
		contract.WithExpiry(10*time.Second),
		contract.WithTries(5),
		contract.WithRetryDelay(100*time.Millisecond),
	)
	require.NotNil(t, mutex)

	ctx := context.Background()
	err = mutex.Lock(ctx)
	require.NoError(t, err)

	// TTL should reflect custom expiry
	ttl := mutex.TTL()
	assert.True(t, ttl > 0)
	assert.True(t, ttl <= 10*time.Second)

	err = mutex.Unlock(ctx)
	require.NoError(t, err)
}

func TestManager_Stats(t *testing.T) {
	logger := &testLogger{t: t}

	config := &Config{
		RedisURL:           "redis://" + testRedisAddr + "/0",
		DefaultExpiry:      8 * time.Second,
		DefaultTries:       32,
		DefaultRetryDelay:  500 * time.Millisecond,
		DefaultDriftFactor: 0.01,
		DefaultKeyPrefix:   "test:lock:",
	}

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer manager.Close(context.Background())

	// Clean up using the pool
	pool := manager.Pools()[0]
	if cp, ok := pool.(*ClientPool); ok {
		cleanupKeys(t, cp.Client(), "test:lock:*")
		defer cleanupKeys(t, cp.Client(), "test:lock:*")
	}

	ctx := context.Background()

	// Initial stats should be zero
	stats, err := manager.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), stats.ActiveLocks)
	assert.Equal(t, int64(0), stats.TotalAcquired)
	assert.Equal(t, int64(0), stats.TotalReleased)
	assert.True(t, stats.BackendLatency > 0)

	// Acquire a lock
	mutex := manager.NewMutex("stats-test")
	err = mutex.Lock(ctx)
	require.NoError(t, err)

	// Stats should reflect the acquired lock
	stats, err = manager.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), stats.ActiveLocks)
	assert.Equal(t, int64(1), stats.TotalAcquired)

	// Release the lock
	err = mutex.Unlock(ctx)
	require.NoError(t, err)

	// Stats should reflect the released lock
	stats, err = manager.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), stats.ActiveLocks)
	assert.Equal(t, int64(1), stats.TotalReleased)
}

func TestManager_Close(t *testing.T) {
	logger := &testLogger{t: t}

	config := &Config{
		RedisURL:           "redis://" + testRedisAddr + "/0",
		DefaultExpiry:      8 * time.Second,
		DefaultTries:       32,
		DefaultRetryDelay:  500 * time.Millisecond,
		DefaultDriftFactor: 0.01,
		DefaultKeyPrefix:   "test:lock:",
	}

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	ctx := context.Background()

	// Close manager
	err = manager.Close(ctx)
	require.NoError(t, err)

	// Operations should fail after close
	err = manager.Health(ctx)
	assert.Error(t, err)
}

func TestManager_InvalidConnection(t *testing.T) {
	logger := &testLogger{t: t}

	config := &Config{
		RedisURL: "redis://invalid-host:6379/0",
	}

	_, err := NewManager(config, logger)
	assert.Error(t, err)
}

func TestManager_RedlockMode(t *testing.T) {
	logger := &testLogger{t: t}

	// Test Redlock mode with multiple URLs
	// Note: For true Redlock, these should be independent Redis instances.
	// This test verifies the configuration and pool creation.
	config := &Config{
		RedisURLs: []string{
			"redis://" + testRedisAddr + "/0",
			"redis://" + testRedisAddr + "/1",
			"redis://" + testRedisAddr + "/2",
		},
		DefaultExpiry:      8 * time.Second,
		DefaultTries:       32,
		DefaultRetryDelay:  500 * time.Millisecond,
		DefaultDriftFactor: 0.01,
		DefaultKeyPrefix:   "test:lock:",
	}

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer manager.Close(context.Background())

	// Should have 3 pools (Redlock mode)
	assert.Len(t, manager.Pools(), 3)
	assert.Equal(t, "redlock", manager.Mode())

	ctx := context.Background()

	// Health check should pass if quorum is healthy
	err = manager.Health(ctx)
	require.NoError(t, err)

	// Verify all pools are functional
	for i, pool := range manager.Pools() {
		err := pool.Ping(ctx)
		assert.NoError(t, err, "Pool %d should be healthy", i)
	}

	// Clean up
	cleanupPoolKeys(t, manager.Pools(), "test:lock:*")
	defer cleanupPoolKeys(t, manager.Pools(), "test:lock:*")

	// Test lock acquisition with Redlock (using different DBs simulates independent instances)
	mutex := manager.NewMutex("redlock-test")
	err = mutex.Lock(ctx)
	require.NoError(t, err)

	// Verify lock is held
	ttl := mutex.TTL()
	assert.True(t, ttl > 0, "Lock should have positive TTL")

	err = mutex.Unlock(ctx)
	require.NoError(t, err)
}

func TestManager_SinglePoolLock(t *testing.T) {
	logger := &testLogger{t: t}

	config := &Config{
		RedisURL:           "redis://" + testRedisAddr + "/0",
		DefaultExpiry:      8 * time.Second,
		DefaultTries:       32,
		DefaultRetryDelay:  500 * time.Millisecond,
		DefaultDriftFactor: 0.01,
		DefaultKeyPrefix:   "test:lock:single:",
	}

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer manager.Close(context.Background())

	ctx := context.Background()

	// Clean up
	pool := manager.Pools()[0]
	if cp, ok := pool.(*ClientPool); ok {
		cleanupKeys(t, cp.Client(), "test:lock:single:*")
		defer cleanupKeys(t, cp.Client(), "test:lock:single:*")
	}

	// Test single-pool lock acquisition
	mutex := manager.NewMutex("single-pool-test")
	err = mutex.Lock(ctx)
	require.NoError(t, err)

	// Verify lock is held
	ttl := mutex.TTL()
	assert.True(t, ttl > 0, "Lock should have positive TTL")

	err = mutex.Unlock(ctx)
	require.NoError(t, err)
}

func TestCreatePoolFromURL_SingleRedis(t *testing.T) {
	config := &Config{
		RedisURL: "redis://" + testRedisAddr + "/0",
	}

	pools, mode, err := createPools(config)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer func() {
		for _, p := range pools {
			_ = p.Close()
		}
	}()

	assert.Len(t, pools, 1)
	assert.Equal(t, "single", mode)
}

func TestCreatePoolFromURL_Redlock(t *testing.T) {
	config := &Config{
		RedisURLs: []string{
			"redis://" + testRedisAddr + "/0",
			"redis://" + testRedisAddr + "/1",
		},
	}

	pools, mode, err := createPools(config)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer func() {
		for _, p := range pools {
			_ = p.Close()
		}
	}()

	assert.Len(t, pools, 2)
	assert.Equal(t, "redlock", mode)
}

func TestCreatePoolFromURL_InvalidScheme(t *testing.T) {
	config := &Config{
		RedisURL: "http://localhost:6379",
	}

	_, _, err := createPools(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported URL scheme")
}

func TestCreatePoolFromURL_ClusterScheme(t *testing.T) {
	config := &Config{
		RedisURL: "redis-cluster://node1:6379,node2:6379",
	}

	// This will fail to connect, but tests URL parsing
	pools, mode, err := createPools(config)
	if err == nil {
		// If somehow it works (unlikely without real cluster)
		assert.Equal(t, "cluster", mode)
		for _, p := range pools {
			_ = p.Close()
		}
	}
	// Error is expected since no cluster is available
}

func TestCreatePoolFromURL_SentinelScheme(t *testing.T) {
	config := &Config{
		RedisURL: "redis-sentinel://mymaster@sentinel1:26379,sentinel2:26379/0",
	}

	// This will fail to connect, but tests URL parsing
	pools, mode, err := createPools(config)
	if err == nil {
		// If somehow it works (unlikely without real sentinel)
		assert.Equal(t, "sentinel", mode)
		for _, p := range pools {
			_ = p.Close()
		}
	}
	// Error is expected since no sentinel is available
}

func TestCreatePoolFromURL_SentinelMissingMaster(t *testing.T) {
	config := &Config{
		RedisURL: "redis-sentinel://sentinel1:26379,sentinel2:26379/0",
	}

	_, _, err := createPools(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sentinel URL must include master name")
}

func TestCreatePoolFromURL_RedlockWithNonSingleURL(t *testing.T) {
	config := &Config{
		RedisURLs: []string{
			"redis://" + testRedisAddr + "/0",
			"redis-cluster://node1:6379,node2:6379", // Invalid for Redlock
		},
	}

	_, _, err := createPools(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Redlock mode requires redis:// or rediss:// URLs")
}

func TestManager_PoolsMethod(t *testing.T) {
	logger := &testLogger{t: t}

	config := &Config{
		RedisURL:           "redis://" + testRedisAddr + "/0",
		DefaultExpiry:      8 * time.Second,
		DefaultTries:       32,
		DefaultRetryDelay:  500 * time.Millisecond,
		DefaultDriftFactor: 0.01,
		DefaultKeyPrefix:   "test:lock:",
	}

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer manager.Close(context.Background())

	pools := manager.Pools()
	assert.NotEmpty(t, pools)
	assert.Len(t, pools, 1)

	// Verify pool is functional
	ctx := context.Background()
	err = pools[0].Ping(ctx)
	assert.NoError(t, err)
}

func TestManager_ModeMethod(t *testing.T) {
	logger := &testLogger{t: t}

	// Test single mode
	config := &Config{
		RedisURL: "redis://" + testRedisAddr + "/0",
	}

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer manager.Close(context.Background())

	assert.Equal(t, "single", manager.Mode())
}

func TestManager_DefaultURL(t *testing.T) {
	logger := &testLogger{t: t}

	// Test with empty URL (should use default)
	config := &Config{
		RedisURL:          "", // Will default to localhost:6379
		DefaultExpiry:     8 * time.Second,
		DefaultTries:      32,
		DefaultRetryDelay: 500 * time.Millisecond,
		DefaultKeyPrefix:  "test:lock:",
	}

	// This may fail if localhost:6379 isn't the test Redis
	_, err := NewManager(config, logger)
	// Just verify it doesn't panic - connection error is expected if not localhost
	_ = err
}
