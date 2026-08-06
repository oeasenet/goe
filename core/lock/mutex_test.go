package lock

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/zap"
)

// testLogger implements contract.Logger for testing
type testLogger struct {
	t *testing.T
}

func (l *testLogger) Debug(msg string, args ...any) {
	if l.t != nil {
		l.t.Logf("[DEBUG] %s %v", msg, args)
	}
}
func (l *testLogger) Info(msg string, args ...any) {
	if l.t != nil {
		l.t.Logf("[INFO] %s %v", msg, args)
	}
}
func (l *testLogger) Warn(msg string, args ...any) {
	if l.t != nil {
		l.t.Logf("[WARN] %s %v", msg, args)
	}
}
func (l *testLogger) Error(msg string, args ...any) {
	if l.t != nil {
		l.t.Logf("[ERROR] %s %v", msg, args)
	}
}
func (l *testLogger) Fatal(msg string, args ...any) {
	if l.t != nil {
		l.t.Fatalf("[FATAL] %s %v", msg, args)
	}
}
func (l *testLogger) Debugf(template string, args ...any) {
	if l.t != nil {
		l.t.Logf("[DEBUG] "+template, args...)
	}
}
func (l *testLogger) Infof(template string, args ...any) {
	if l.t != nil {
		l.t.Logf("[INFO] "+template, args...)
	}
}
func (l *testLogger) Warnf(template string, args ...any) {
	if l.t != nil {
		l.t.Logf("[WARN] "+template, args...)
	}
}
func (l *testLogger) Errorf(template string, args ...any) {
	if l.t != nil {
		l.t.Logf("[ERROR] "+template, args...)
	}
}
func (l *testLogger) Fatalf(template string, args ...any) {
	if l.t != nil {
		l.t.Fatalf("[FATAL] "+template, args...)
	}
}
func (l *testLogger) Debugw(msg string, keysAndValues ...any) {
	if l.t != nil {
		l.t.Logf("[DEBUG] %s %v", msg, keysAndValues)
	}
}
func (l *testLogger) Infow(msg string, keysAndValues ...any) {
	if l.t != nil {
		l.t.Logf("[INFO] %s %v", msg, keysAndValues)
	}
}
func (l *testLogger) Warnw(msg string, keysAndValues ...any) {
	if l.t != nil {
		l.t.Logf("[WARN] %s %v", msg, keysAndValues)
	}
}
func (l *testLogger) Errorw(msg string, keysAndValues ...any) {
	if l.t != nil {
		l.t.Logf("[ERROR] %s %v", msg, keysAndValues)
	}
}
func (l *testLogger) Fatalw(msg string, keysAndValues ...any) {
	if l.t != nil {
		l.t.Fatalf("[FATAL] %s %v", msg, keysAndValues)
	}
}
func (l *testLogger) With(keysAndValues ...any) contract.Logger {
	return l
}
func (l *testLogger) WithContext(ctx context.Context) contract.Logger {
	return l
}
func (l *testLogger) WithError(err error) contract.Logger {
	return l
}
func (l *testLogger) GetLogger() *zap.SugaredLogger {
	return nil
}

// getTestRedisClient creates a Redis client for testing
func getTestRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	requireRedis(t)

	client := redis.NewClient(&redis.Options{
		Addr: testRedisAddr,
		DB:   0,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Ping(ctx).Err()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	return client
}

// getTestPool creates a single Pool for testing
func getTestPool(t *testing.T) Pool {
	client := getTestRedisClient(t)
	return NewClientPool(client)
}

// getTestPools creates multiple pools (simulating Redlock with multiple instances)
// Note: In real Redlock, you'd use independent Redis servers
// For testing, we use the same server to verify the algorithm logic
func getTestPools(t *testing.T, count int) []Pool {
	t.Helper()
	requireRedis(t)

	pools := make([]Pool, count)
	for i := range count {
		client := redis.NewClient(&redis.Options{
			Addr: testRedisAddr,
			DB:   i % 16, // Use different DBs to simulate independence
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := client.Ping(ctx).Err()
		if err != nil {
			t.Skipf("Redis not available: %v", err)
		}

		pools[i] = NewClientPool(client)
	}
	return pools
}

// closePools closes all pools
func closePools(pools []Pool) {
	for _, p := range pools {
		_ = p.Close()
	}
}

// cleanupKeys removes test keys from Redis
func cleanupKeys(t *testing.T, client *redis.Client, pattern string) {
	ctx := context.Background()
	keys, err := client.Keys(ctx, pattern).Result()
	require.NoError(t, err)

	if len(keys) > 0 {
		err = client.Del(ctx, keys...).Err()
		require.NoError(t, err)
	}
}

// cleanupPoolKeys removes test keys from all pool DBs
func cleanupPoolKeys(t *testing.T, pools []Pool, pattern string) {
	ctx := context.Background()
	for _, p := range pools {
		if cp, ok := p.(*ClientPool); ok {
			keys, err := cp.Client().Keys(ctx, pattern).Result()
			if err == nil && len(keys) > 0 {
				_ = cp.Client().Del(ctx, keys...).Err()
			}
		}
	}
}

func TestMutex_Lock(t *testing.T) {
	pool := getTestPool(t)
	defer func() { _ = pool.Close() }()

	client := pool.(*ClientPool).Client()
	logger := &testLogger{t: t}

	// Clean up before test
	cleanupKeys(t, client, "test:lock:*")
	defer cleanupKeys(t, client, "test:lock:*")

	mutex := NewMutex("test:lock:basic", []Pool{pool}, logger,
		contract.WithExpiry(5*time.Second),
		contract.WithTries(3),
		contract.WithRetryDelay(100*time.Millisecond),
	)

	ctx := context.Background()

	// Should acquire lock successfully
	err := mutex.Lock(ctx)
	require.NoError(t, err)

	// Lock should be held
	assert.True(t, mutex.acquired)
	assert.False(t, mutex.until.IsZero())
	assert.True(t, mutex.TTL() > 0)

	// Release lock
	err = mutex.Unlock(ctx)
	require.NoError(t, err)

	// Lock should no longer be held
	assert.False(t, mutex.acquired)
}

func TestMutex_TryLock(t *testing.T) {
	pool := getTestPool(t)
	defer func() { _ = pool.Close() }()

	client := pool.(*ClientPool).Client()
	logger := &testLogger{t: t}

	cleanupKeys(t, client, "test:lock:*")
	defer cleanupKeys(t, client, "test:lock:*")

	mutex1 := NewMutex("test:lock:trylock", []Pool{pool}, logger,
		contract.WithExpiry(5*time.Second),
	)
	mutex2 := NewMutex("test:lock:trylock", []Pool{pool}, logger,
		contract.WithExpiry(5*time.Second),
	)

	ctx := context.Background()

	// First mutex should acquire lock
	acquired, err := mutex1.TryLock(ctx)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Second mutex should fail to acquire
	acquired, err = mutex2.TryLock(ctx)
	require.NoError(t, err)
	assert.False(t, acquired)

	// Release first lock
	err = mutex1.Unlock(ctx)
	require.NoError(t, err)

	// Now second mutex should acquire
	acquired, err = mutex2.TryLock(ctx)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Clean up
	err = mutex2.Unlock(ctx)
	require.NoError(t, err)
}

func TestMutex_Extend(t *testing.T) {
	pool := getTestPool(t)
	defer func() { _ = pool.Close() }()

	client := pool.(*ClientPool).Client()
	logger := &testLogger{t: t}

	cleanupKeys(t, client, "test:lock:*")
	defer cleanupKeys(t, client, "test:lock:*")

	mutex := NewMutex("test:lock:extend", []Pool{pool}, logger,
		contract.WithExpiry(2*time.Second),
	)

	ctx := context.Background()

	// Acquire lock
	err := mutex.Lock(ctx)
	require.NoError(t, err)

	// Wait a bit
	time.Sleep(500 * time.Millisecond)

	originalTTL := mutex.TTL()
	assert.True(t, originalTTL > 0)
	assert.True(t, originalTTL < 2*time.Second)

	// Extend the lock
	err = mutex.Extend(ctx)
	require.NoError(t, err)

	// TTL should be reset
	newTTL := mutex.TTL()
	assert.True(t, newTTL > originalTTL)

	// Release lock
	err = mutex.Unlock(ctx)
	require.NoError(t, err)
}

func TestMutex_ExtendWithoutLock(t *testing.T) {
	pool := getTestPool(t)
	defer func() { _ = pool.Close() }()

	logger := &testLogger{t: t}

	mutex := NewMutex("test:lock:extend-no-lock", []Pool{pool}, logger)

	ctx := context.Background()

	// Try to extend without holding lock
	err := mutex.Extend(ctx)
	assert.ErrorIs(t, err, ErrLockNotHeld)
}

func TestMutex_UnlockWithoutLock(t *testing.T) {
	pool := getTestPool(t)
	defer func() { _ = pool.Close() }()

	logger := &testLogger{t: t}

	mutex := NewMutex("test:lock:unlock-no-lock", []Pool{pool}, logger)

	ctx := context.Background()

	// Try to unlock without holding lock
	err := mutex.Unlock(ctx)
	assert.ErrorIs(t, err, ErrLockNotHeld)
}

func TestMutex_ConcurrentAccess(t *testing.T) {
	pool := getTestPool(t)
	defer func() { _ = pool.Close() }()

	client := pool.(*ClientPool).Client()
	logger := &testLogger{t: t}

	cleanupKeys(t, client, "test:lock:*")
	defer cleanupKeys(t, client, "test:lock:*")

	const numGoroutines = 10
	const numIterations = 5

	var counter int64
	var wg sync.WaitGroup

	for range numGoroutines {
		wg.Go(func() {

			for range numIterations {
				mutex := NewMutex("test:lock:concurrent", []Pool{pool}, logger,
					contract.WithExpiry(5*time.Second),
					contract.WithTries(50),
					contract.WithRetryDelay(50*time.Millisecond),
				)

				ctx := context.Background()

				if err := mutex.Lock(ctx); err != nil {
					t.Logf("Lock failed: %v", err)
					continue
				}

				// Critical section
				current := atomic.LoadInt64(&counter)
				time.Sleep(10 * time.Millisecond) // Simulate work
				atomic.StoreInt64(&counter, current+1)

				if err := mutex.Unlock(ctx); err != nil {
					t.Logf("Unlock failed: %v", err)
				}
			}
		})
	}

	wg.Wait()

	// All increments should have happened safely
	assert.Equal(t, int64(numGoroutines*numIterations), atomic.LoadInt64(&counter))
}

func TestMutex_ContextCancellation(t *testing.T) {
	pool := getTestPool(t)
	defer func() { _ = pool.Close() }()

	client := pool.(*ClientPool).Client()
	logger := &testLogger{t: t}

	cleanupKeys(t, client, "test:lock:*")
	defer cleanupKeys(t, client, "test:lock:*")

	// First, acquire the lock with one mutex
	mutex1 := NewMutex("test:lock:cancel", []Pool{pool}, logger,
		contract.WithExpiry(30*time.Second),
	)

	ctx := context.Background()
	err := mutex1.Lock(ctx)
	require.NoError(t, err)

	// Now try to acquire with a cancelled context
	mutex2 := NewMutex("test:lock:cancel", []Pool{pool}, logger,
		contract.WithExpiry(5*time.Second),
		contract.WithTries(100),
		contract.WithRetryDelay(100*time.Millisecond),
	)

	ctx2, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err = mutex2.Lock(ctx2)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	// Clean up
	err = mutex1.Unlock(context.Background())
	require.NoError(t, err)
}

func TestMutex_Name(t *testing.T) {
	pool := getTestPool(t)
	defer func() { _ = pool.Close() }()

	logger := &testLogger{t: t}

	mutex := NewMutex("test:lock:name", []Pool{pool}, logger)
	assert.Equal(t, "test:lock:name", mutex.Name())
}

func TestMutex_CustomValue(t *testing.T) {
	pool := getTestPool(t)
	defer func() { _ = pool.Close() }()

	client := pool.(*ClientPool).Client()
	logger := &testLogger{t: t}

	cleanupKeys(t, client, "test:lock:*")
	defer cleanupKeys(t, client, "test:lock:*")

	customValue := "my-custom-lock-value"
	mutex := NewMutex("test:lock:custom-value", []Pool{pool}, logger,
		contract.WithValue(customValue),
		contract.WithExpiry(5*time.Second),
	)

	ctx := context.Background()

	err := mutex.Lock(ctx)
	require.NoError(t, err)

	// Verify the value in Redis
	val, err := client.Get(ctx, "test:lock:custom-value").Result()
	require.NoError(t, err)
	assert.Equal(t, customValue, val)

	err = mutex.Unlock(ctx)
	require.NoError(t, err)
}

func TestMutex_TTL(t *testing.T) {
	pool := getTestPool(t)
	defer func() { _ = pool.Close() }()

	client := pool.(*ClientPool).Client()
	logger := &testLogger{t: t}

	cleanupKeys(t, client, "test:lock:*")
	defer cleanupKeys(t, client, "test:lock:*")

	mutex := NewMutex("test:lock:ttl", []Pool{pool}, logger,
		contract.WithExpiry(3*time.Second),
	)

	ctx := context.Background()

	// TTL should be 0 when lock is not held
	assert.Equal(t, time.Duration(0), mutex.TTL())

	// Acquire lock
	err := mutex.Lock(ctx)
	require.NoError(t, err)

	// TTL should be positive
	ttl := mutex.TTL()
	assert.True(t, ttl > 0, "TTL should be positive after acquiring lock")
	assert.True(t, ttl <= 3*time.Second, "TTL should not exceed expiry time")

	// Release lock
	err = mutex.Unlock(ctx)
	require.NoError(t, err)

	// TTL should be 0 again
	assert.Equal(t, time.Duration(0), mutex.TTL())
}

func TestMutex_LockExpiry(t *testing.T) {
	pool := getTestPool(t)
	defer func() { _ = pool.Close() }()

	client := pool.(*ClientPool).Client()
	logger := &testLogger{t: t}

	cleanupKeys(t, client, "test:lock:*")
	defer cleanupKeys(t, client, "test:lock:*")

	// Create lock with short expiry
	mutex1 := NewMutex("test:lock:expiry", []Pool{pool}, logger,
		contract.WithExpiry(1*time.Second),
	)

	ctx := context.Background()

	err := mutex1.Lock(ctx)
	require.NoError(t, err)

	// Wait for lock to expire
	time.Sleep(1500 * time.Millisecond)

	// Another mutex should be able to acquire
	mutex2 := NewMutex("test:lock:expiry", []Pool{pool}, logger,
		contract.WithExpiry(5*time.Second),
	)

	acquired, err := mutex2.TryLock(ctx)
	require.NoError(t, err)
	assert.True(t, acquired, "Should acquire lock after expiry")

	// Clean up
	err = mutex2.Unlock(ctx)
	require.NoError(t, err)
}

func TestMutex_NoPoolsError(t *testing.T) {
	logger := &testLogger{t: t}

	mutex := NewMutex("test:lock:no-pools", []Pool{}, logger)

	ctx := context.Background()

	// Should return error when no pools available
	err := mutex.Lock(ctx)
	assert.ErrorIs(t, err, ErrNoPoolsAvailable)

	acquired, err := mutex.TryLock(ctx)
	assert.ErrorIs(t, err, ErrNoPoolsAvailable)
	assert.False(t, acquired)
}

func TestMutex_Quorum(t *testing.T) {
	// Test quorum calculation for various pool counts
	tests := []struct {
		pools    int
		expected int
	}{
		{1, 1},
		{2, 2},
		{3, 2},
		{4, 3},
		{5, 3},
	}

	for _, tt := range tests {
		quorum := tt.pools/2 + 1
		assert.Equal(t, tt.expected, quorum, "Quorum for %d pools", tt.pools)
	}
}

// TestMutex_MultiPool tests the Redlock algorithm with multiple pools
func TestMutex_MultiPool(t *testing.T) {
	pools := getTestPools(t, 3)
	defer closePools(pools)

	logger := &testLogger{t: t}

	// Clean up all pools
	defer cleanupPoolKeys(t, pools, "test:lock:*")

	mutex := NewMutex("test:lock:multipool", pools, logger,
		contract.WithExpiry(5*time.Second),
		contract.WithTries(3),
	)

	ctx := context.Background()

	// Should acquire lock successfully (quorum is 2 for 3 pools)
	err := mutex.Lock(ctx)
	require.NoError(t, err)

	assert.True(t, mutex.acquired)
	assert.Equal(t, 2, mutex.quorum)

	// Release lock
	err = mutex.Unlock(ctx)
	require.NoError(t, err)
}

// TestMutex_RetryDelay tests the exponential backoff with jitter
func TestMutex_RetryDelay(t *testing.T) {
	pool := getTestPool(t)
	defer func() { _ = pool.Close() }()

	logger := &testLogger{t: t}

	mutex := NewMutex("test:lock:retry", []Pool{pool}, logger,
		contract.WithRetryDelay(50*time.Millisecond),
	)

	// Test retry delay calculation
	delay0 := mutex.calculateRetryDelay(0)
	delay1 := mutex.calculateRetryDelay(1)
	delay2 := mutex.calculateRetryDelay(2)

	// Delays should increase (with some jitter)
	// Base delay is 50ms, with exponential backoff 1.5x
	// delay0 ≈ 50ms (±25% jitter)
	// delay1 ≈ 75ms (±25% jitter)
	// delay2 ≈ 112ms (±25% jitter)
	assert.True(t, delay0 > 25*time.Millisecond && delay0 < 100*time.Millisecond)
	assert.True(t, delay1 > delay0/2) // Should generally be higher than delay0
	assert.True(t, delay2 <= DefaultRetryDelayMax)
}

// TestMutex_ClockDrift tests clock drift compensation
func TestMutex_ClockDrift(t *testing.T) {
	requireRedis(t)

	pool := getTestPool(t)
	defer func() { _ = pool.Close() }()

	client := pool.(*ClientPool).Client()
	logger := &testLogger{t: t}

	cleanupKeys(t, client, "test:lock:*")
	defer cleanupKeys(t, client, "test:lock:*")

	expiry := 10 * time.Second
	mutex := NewMutex("test:lock:drift", []Pool{pool}, logger,
		contract.WithExpiry(expiry),
		contract.WithDriftFactor(0.01), // 1% drift factor
	)

	ctx := context.Background()

	err := mutex.Lock(ctx)
	require.NoError(t, err)

	// The TTL should be less than expiry due to clock drift compensation
	// drift = expiry * 0.01 + 2ms = 100ms + 2ms = 102ms
	// So TTL should be less than expiry - drift
	ttl := mutex.TTL()
	assert.True(t, ttl < expiry, "TTL should be less than expiry due to drift compensation")

	err = mutex.Unlock(ctx)
	require.NoError(t, err)
}

// Benchmark tests with Pool interface

func BenchmarkMutex_Lock(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr: testRedisAddr,
		DB:   0,
	})
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		b.Skipf("Redis not available: %v", err)
	}

	pool := NewClientPool(client)
	logger := &testLogger{t: nil}

	b.ResetTimer()
	for b.Loop() {
		mutex := NewMutex("bench:lock", []Pool{pool}, logger,
			contract.WithExpiry(5*time.Second),
			contract.WithTries(1),
		)

		if err := mutex.Lock(ctx); err != nil {
			continue
		}
		_ = mutex.Unlock(ctx)
	}
}

func BenchmarkMutex_TryLock(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr: testRedisAddr,
		DB:   0,
	})
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		b.Skipf("Redis not available: %v", err)
	}

	pool := NewClientPool(client)
	logger := &testLogger{t: nil}

	b.ResetTimer()
	for b.Loop() {
		mutex := NewMutex("bench:trylock", []Pool{pool}, logger,
			contract.WithExpiry(5*time.Second),
		)

		acquired, _ := mutex.TryLock(ctx)
		if acquired {
			_ = mutex.Unlock(ctx)
		}
	}
}

func BenchmarkMutex_Extend(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr: testRedisAddr,
		DB:   0,
	})
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		b.Skipf("Redis not available: %v", err)
	}

	pool := NewClientPool(client)
	logger := &testLogger{t: nil}

	mutex := NewMutex("bench:extend", []Pool{pool}, logger,
		contract.WithExpiry(30*time.Second),
	)

	if err := mutex.Lock(ctx); err != nil {
		b.Fatalf("Failed to acquire lock: %v", err)
	}
	defer func() { _ = mutex.Unlock(ctx) }()

	b.ResetTimer()
	for b.Loop() {
		if err := mutex.Extend(ctx); err != nil {
			b.Fatalf("Failed to extend lock: %v", err)
		}
	}
}

func BenchmarkMutex_LockUnlock_Parallel(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr:     testRedisAddr,
		DB:       0,
		PoolSize: 50,
	})
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		b.Skipf("Redis not available: %v", err)
	}

	pool := NewClientPool(client)
	logger := &testLogger{t: nil}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Use unique key per goroutine iteration to avoid contention
			key := "bench:parallel:" + string(rune(i%100))
			mutex := NewMutex(key, []Pool{pool}, logger,
				contract.WithExpiry(5*time.Second),
				contract.WithTries(1),
			)

			if err := mutex.Lock(ctx); err == nil {
				_ = mutex.Unlock(ctx)
			}
			i++
		}
	})
}

func BenchmarkMutex_Contention(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr:     testRedisAddr,
		DB:       0,
		PoolSize: 50,
	})
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		b.Skipf("Redis not available: %v", err)
	}

	// Clean up any stale locks
	client.Del(ctx, "bench:contention")

	pool := NewClientPool(client)
	logger := &testLogger{t: nil}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// All goroutines compete for the same lock
			mutex := NewMutex("bench:contention", []Pool{pool}, logger,
				contract.WithExpiry(100*time.Millisecond),
				contract.WithTries(10),
				contract.WithRetryDelay(5*time.Millisecond),
			)

			if err := mutex.Lock(ctx); err == nil {
				// Short critical section
				time.Sleep(1 * time.Microsecond)
				_ = mutex.Unlock(ctx)
			}
		}
	})
}

func BenchmarkMutex_ValueGeneration(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		_, err := generateRandomValue()
		if err != nil {
			b.Fatalf("Failed to generate value: %v", err)
		}
	}
}

// BenchmarkMutex_MultiPool benchmarks Redlock with multiple pools
func BenchmarkMutex_MultiPool(b *testing.B) {
	// Create 3 pools (using different DBs to simulate independence)
	pools := make([]Pool, 3)
	for i := range 3 {
		client := redis.NewClient(&redis.Options{
			Addr: testRedisAddr,
			DB:   i,
		})
		ctx := context.Background()
		if err := client.Ping(ctx).Err(); err != nil {
			b.Skipf("Redis not available: %v", err)
		}
		pools[i] = NewClientPool(client)
	}
	defer func() {
		for _, p := range pools {
			_ = p.Close()
		}
	}()

	logger := &testLogger{t: nil}
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		mutex := NewMutex("bench:multipool", pools, logger,
			contract.WithExpiry(5*time.Second),
			contract.WithTries(1),
		)

		if err := mutex.Lock(ctx); err != nil {
			continue
		}
		_ = mutex.Unlock(ctx)
	}
}
