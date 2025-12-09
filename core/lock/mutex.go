package lock

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"math"
	mrand "math/rand"
	"sync"
	"time"

	"go.oease.dev/goe/v2/contract"
)

var (
	// ErrLockNotHeld is returned when trying to unlock or extend a lock that is not held.
	ErrLockNotHeld = errors.New("lock: lock not held")

	// ErrLockAcquireFailed is returned when the lock cannot be acquired after all retries.
	ErrLockAcquireFailed = errors.New("lock: failed to acquire lock")

	// ErrLockExtendFailed is returned when the lock cannot be extended.
	ErrLockExtendFailed = errors.New("lock: failed to extend lock")

	// ErrNoPoolsAvailable is returned when no Redis pools are available.
	ErrNoPoolsAvailable = errors.New("lock: no pools available")
)

// Lua scripts for atomic operations
const (
	// deleteScript deletes the lock only if it's held by the current owner.
	// KEYS[1] = lock key
	// ARGV[1] = lock value (owner identifier)
	// Returns 1 if deleted, 0 if not held
	deleteScript = `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	// extendScript extends the lock TTL only if it's held by the current owner.
	// KEYS[1] = lock key
	// ARGV[1] = lock value (owner identifier)
	// ARGV[2] = new TTL in milliseconds
	// Returns 1 if extended, 0 if not held
	extendScript = `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`
)

// Default configuration values
const (
	// DefaultClockDriftFactor is the default clock drift factor.
	// This accounts for clock drift between different machines.
	DefaultClockDriftFactor = 0.01

	// DefaultRetryDelayBase is the base retry delay.
	DefaultRetryDelayBase = 50 * time.Millisecond

	// DefaultRetryDelayMax is the maximum retry delay.
	DefaultRetryDelayMax = 250 * time.Millisecond

	// MinLockValidity is the minimum lock validity time to be considered valid.
	MinLockValidity = 2 * time.Millisecond
)

// Mutex implements a distributed mutual exclusion lock using the Redlock algorithm.
// It supports multiple Redis instances for fault tolerance, requiring a quorum
// (N/2 + 1) of successful lock acquisitions for the lock to be considered held.
//
// For a single Redis instance, it degrades gracefully to a standard Redis lock.
//
// Reference: https://redis.io/docs/latest/develop/clients/patterns/distributed-locks/
type Mutex struct {
	name    string // Lock name (key in Redis)
	value   string // Unique value identifying this lock holder
	pools   []Pool // Redis pools (multiple for Redlock, single for simple mode)
	options *contract.MutexOptions
	logger  contract.Logger
	quorum  int // Number of pools required for quorum (N/2 + 1)

	mu       sync.Mutex // Protects the following fields
	until    time.Time  // When the lock expires (based on shortest validity)
	acquired bool       // Whether the lock is currently held
}

// NewMutex creates a new Mutex instance using the Redlock algorithm.
// For fault tolerance, provide multiple independent Redis pools.
// The quorum is automatically calculated as N/2 + 1.
func NewMutex(name string, pools []Pool, logger contract.Logger, opts ...contract.MutexOption) *Mutex {
	options := contract.DefaultMutexOptions()
	contract.ApplyMutexOptions(options, opts...)

	// Calculate quorum: majority of pools required
	quorum := len(pools)/2 + 1

	m := &Mutex{
		name:    name,
		pools:   pools,
		options: options,
		logger:  logger,
		quorum:  quorum,
	}

	return m
}

// Name returns the name/key of the mutex.
func (m *Mutex) Name() string {
	return m.name
}

// Lock acquires the lock using the Redlock algorithm.
// It blocks until the lock is acquired or the context is cancelled.
//
// The algorithm:
// 1. Get current timestamp
// 2. Try to acquire lock on all N Redis instances sequentially/parallel
// 3. Calculate elapsed time and remaining validity
// 4. Lock is acquired if we got it on quorum (N/2+1) instances AND validity > 0
// 5. If failed, release lock on all instances and retry with random delay
func (m *Mutex) Lock(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.pools) == 0 {
		return ErrNoPoolsAvailable
	}

	// Generate lock value if not already set
	if err := m.ensureValue(); err != nil {
		return err
	}

	expiry := m.options.Expiry
	tries := m.options.Tries

	for i := 0; i < tries; i++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Try to acquire the lock using Redlock algorithm
		start := time.Now()
		successCount, err := m.acquireOnPools(ctx, expiry)
		elapsed := time.Since(start)

		if err != nil {
			m.logger.Debug("Lock acquisition attempt failed",
				"name", m.name,
				"attempt", i+1,
				"error", err,
			)
		}

		// Calculate clock drift compensation
		// drift = expiry * driftFactor + 2ms (for network latency)
		driftFactor := m.options.DriftFactor
		if driftFactor == 0 {
			driftFactor = DefaultClockDriftFactor
		}
		drift := time.Duration(int64(float64(expiry)*driftFactor)) + 2*time.Millisecond

		// Calculate remaining validity time
		validityTime := expiry - elapsed - drift

		// Lock is acquired if:
		// 1. We acquired on quorum number of instances
		// 2. Remaining validity time is positive
		if successCount >= m.quorum && validityTime > MinLockValidity {
			m.until = time.Now().Add(validityTime)
			m.acquired = true

			m.logger.Debug("Lock acquired",
				"name", m.name,
				"value", m.value,
				"until", m.until,
				"quorum", m.quorum,
				"success_count", successCount,
				"validity", validityTime,
			)
			return nil
		}

		// Failed to acquire quorum - release all locks
		m.releaseOnPools(ctx)

		// Wait before retrying with random jitter
		if i < tries-1 {
			delay := m.calculateRetryDelay(i)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return ErrLockAcquireFailed
}

// TryLock attempts to acquire the lock without blocking.
// Returns true if the lock was acquired, false otherwise.
func (m *Mutex) TryLock(ctx context.Context) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.pools) == 0 {
		return false, ErrNoPoolsAvailable
	}

	// Generate lock value if not already set
	if err := m.ensureValue(); err != nil {
		return false, err
	}

	expiry := m.options.Expiry
	start := time.Now()

	successCount, err := m.acquireOnPools(ctx, expiry)
	if err != nil {
		// Release any partial acquisitions
		m.releaseOnPools(ctx)
		return false, err
	}

	elapsed := time.Since(start)

	// Calculate drift
	driftFactor := m.options.DriftFactor
	if driftFactor == 0 {
		driftFactor = DefaultClockDriftFactor
	}
	drift := time.Duration(int64(float64(expiry)*driftFactor)) + 2*time.Millisecond
	validityTime := expiry - elapsed - drift

	// Check quorum and validity
	if successCount >= m.quorum && validityTime > MinLockValidity {
		m.until = time.Now().Add(validityTime)
		m.acquired = true

		m.logger.Debug("Lock acquired (try)",
			"name", m.name,
			"value", m.value,
			"until", m.until,
			"quorum", m.quorum,
			"success_count", successCount,
		)
		return true, nil
	}

	// Failed - release all locks
	m.releaseOnPools(ctx)
	return false, nil
}

// Unlock releases the lock on all Redis instances.
func (m *Mutex) Unlock(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.acquired {
		return ErrLockNotHeld
	}

	// Release on all pools (best effort)
	m.releaseOnPools(ctx)

	m.logger.Debug("Lock released",
		"name", m.name,
		"value", m.value,
	)

	m.acquired = false
	m.until = time.Time{}
	m.value = "" // Clear value so next Lock() generates a fresh one

	return nil
}

// Extend extends the lock's expiration time on all instances.
// This should be called periodically for long-running operations.
func (m *Mutex) Extend(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.acquired {
		return ErrLockNotHeld
	}

	expiry := m.options.Expiry
	start := time.Now()

	successCount := m.extendOnPools(ctx, expiry)
	elapsed := time.Since(start)

	// Check quorum
	if successCount < m.quorum {
		// Lost lock on majority of instances
		m.acquired = false
		m.until = time.Time{}
		m.value = ""
		return ErrLockExtendFailed
	}

	// Calculate new expiry time
	driftFactor := m.options.DriftFactor
	if driftFactor == 0 {
		driftFactor = DefaultClockDriftFactor
	}
	drift := time.Duration(int64(float64(expiry)*driftFactor)) + 2*time.Millisecond
	m.until = time.Now().Add(expiry - elapsed - drift)

	m.logger.Debug("Lock extended",
		"name", m.name,
		"value", m.value,
		"until", m.until,
		"success_count", successCount,
	)

	return nil
}

// Until returns the time when the lock will expire.
func (m *Mutex) Until() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.until
}

// TTL returns the remaining time until the lock expires.
func (m *Mutex) TTL() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.acquired || m.until.IsZero() {
		return 0
	}

	ttl := time.Until(m.until)
	if ttl < 0 {
		return 0
	}
	return ttl
}

// acquireOnPools tries to acquire the lock on all pools in parallel.
// Returns the number of successful acquisitions.
func (m *Mutex) acquireOnPools(ctx context.Context, expiry time.Duration) (int, error) {
	if len(m.pools) == 1 {
		// Single pool - use simple acquisition
		ok, err := m.pools[0].SetNX(ctx, m.name, m.value, expiry)
		if err != nil {
			return 0, err
		}
		if ok {
			return 1, nil
		}
		return 0, nil
	}

	// Multiple pools - acquire in parallel
	type result struct {
		success bool
		err     error
	}

	results := make(chan result, len(m.pools))

	for _, pool := range m.pools {
		go func(p Pool) {
			// Use a short timeout per instance
			timeout := time.Duration(float64(expiry) * 0.1)
			if timeout < 5*time.Millisecond {
				timeout = 5 * time.Millisecond
			}
			if timeout > 50*time.Millisecond {
				timeout = 50 * time.Millisecond
			}

			acquireCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			ok, err := p.SetNX(acquireCtx, m.name, m.value, expiry)
			if err != nil {
				results <- result{success: false, err: err}
				return
			}
			results <- result{success: ok, err: nil}
		}(pool)
	}

	// Collect results
	successCount := 0
	var lastErr error

	for i := 0; i < len(m.pools); i++ {
		r := <-results
		if r.success {
			successCount++
		}
		if r.err != nil {
			lastErr = r.err
		}
	}

	return successCount, lastErr
}

// releaseOnPools releases the lock on all pools (best effort).
func (m *Mutex) releaseOnPools(ctx context.Context) {
	if len(m.pools) == 1 {
		// Single pool - simple release
		_, _ = m.pools[0].Eval(ctx, deleteScript, []string{m.name}, m.value)
		return
	}

	// Multiple pools - release in parallel
	var wg sync.WaitGroup
	wg.Add(len(m.pools))

	for _, pool := range m.pools {
		go func(p Pool) {
			defer wg.Done()
			// Use a short timeout for release
			releaseCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
			defer cancel()
			_, _ = p.Eval(releaseCtx, deleteScript, []string{m.name}, m.value)
		}(pool)
	}

	wg.Wait()
}

// extendOnPools extends the lock on all pools.
// Returns the number of successful extensions.
func (m *Mutex) extendOnPools(ctx context.Context, expiry time.Duration) int {
	if len(m.pools) == 1 {
		// Single pool - simple extend
		result, err := m.pools[0].Eval(ctx, extendScript, []string{m.name}, m.value, int64(expiry/time.Millisecond))
		if err != nil {
			return 0
		}
		if val, ok := result.(int64); ok && val == 1 {
			return 1
		}
		return 0
	}

	// Multiple pools - extend in parallel
	type result struct {
		success bool
	}

	results := make(chan result, len(m.pools))

	for _, pool := range m.pools {
		go func(p Pool) {
			extendCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
			defer cancel()

			res, err := p.Eval(extendCtx, extendScript, []string{m.name}, m.value, int64(expiry/time.Millisecond))
			if err != nil {
				results <- result{success: false}
				return
			}
			if val, ok := res.(int64); ok && val == 1 {
				results <- result{success: true}
				return
			}
			results <- result{success: false}
		}(pool)
	}

	// Collect results
	successCount := 0
	for i := 0; i < len(m.pools); i++ {
		r := <-results
		if r.success {
			successCount++
		}
	}

	return successCount
}

// calculateRetryDelay calculates the retry delay with random jitter.
// This helps prevent thundering herd problems when multiple clients
// are competing for the same lock.
func (m *Mutex) calculateRetryDelay(attempt int) time.Duration {
	// Use custom delay function if provided
	if m.options.RetryDelayFunc != nil {
		return m.options.RetryDelayFunc(attempt)
	}

	// Use configured retry delay as base
	baseDelay := m.options.RetryDelay
	if baseDelay == 0 {
		baseDelay = DefaultRetryDelayBase
	}

	// Add exponential backoff with cap
	delay := time.Duration(float64(baseDelay) * math.Pow(1.5, float64(attempt)))
	if delay > DefaultRetryDelayMax {
		delay = DefaultRetryDelayMax
	}

	// Add random jitter (±25%)
	jitter := time.Duration(mrand.Int63n(int64(delay / 2)))
	delay = delay - delay/4 + jitter

	return delay
}

// ensureValue generates a unique value for the lock if not already set.
func (m *Mutex) ensureValue() error {
	if m.value != "" {
		return nil
	}

	// Check if a specific value is set in options
	if m.options.Value != "" {
		m.value = m.options.Value
		return nil
	}

	// Use custom generator if provided
	if m.options.GenValueFunc != nil {
		value, err := m.options.GenValueFunc()
		if err != nil {
			return err
		}
		m.value = value
		return nil
	}

	// Generate random value
	value, err := generateRandomValue()
	if err != nil {
		return err
	}
	m.value = value
	return nil
}

// generateRandomValue generates a cryptographically random value for the lock.
// The value is 16 bytes (128 bits) encoded as base64, providing sufficient
// uniqueness to prevent accidental lock conflicts.
func generateRandomValue() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
