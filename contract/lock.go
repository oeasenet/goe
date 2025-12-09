package contract

import (
	"context"
	"time"
)

// Mutex represents a distributed mutual exclusion lock.
// It provides a way to coordinate access to shared resources across multiple
// processes or machines using Redis as the backend.
//
// Usage example:
//
//	mutex := lockManager.NewMutex("my-resource")
//	if err := mutex.Lock(ctx); err != nil {
//	    return err
//	}
//	defer mutex.Unlock(ctx)
//	// ... critical section ...
type Mutex interface {
	// Name returns the name/key of the mutex.
	Name() string

	// Lock acquires the lock, blocking until it's available or context is cancelled.
	// Returns an error if the lock cannot be acquired due to context cancellation,
	// timeout, or backend failure.
	Lock(ctx context.Context) error

	// TryLock attempts to acquire the lock without blocking.
	// Returns (true, nil) if the lock was acquired successfully.
	// Returns (false, nil) if the lock is held by another process.
	// Returns (false, error) if there was an error communicating with the backend.
	TryLock(ctx context.Context) (bool, error)

	// Unlock releases the lock.
	// Returns an error if the lock is not held by this instance or if there's
	// a backend communication failure.
	Unlock(ctx context.Context) error

	// Extend extends the lock's expiration time by the configured TTL.
	// This should be called periodically for long-running operations to prevent
	// the lock from expiring. Returns an error if the lock is not held by this
	// instance or if there's a backend communication failure.
	Extend(ctx context.Context) error

	// Until returns the time when the lock will expire.
	// Returns zero time if the lock is not held.
	Until() time.Time

	// TTL returns the remaining time until the lock expires.
	// Returns zero or negative duration if the lock is not held or has expired.
	TTL() time.Duration
}

// MutexOption configures a Mutex instance.
type MutexOption func(*MutexOptions)

// MutexOptions holds configuration options for a Mutex.
type MutexOptions struct {
	// Expiry is the duration for which the lock is valid.
	// After this duration, the lock will automatically expire.
	// Default: 8 seconds
	Expiry time.Duration

	// Tries is the maximum number of attempts to acquire the lock.
	// Default: 32
	Tries int

	// RetryDelay is the delay between retry attempts when trying to acquire the lock.
	// Default: 500ms
	RetryDelay time.Duration

	// RetryDelayFunc is a function that returns the delay before the next retry.
	// If set, this overrides RetryDelay.
	RetryDelayFunc func(tries int) time.Duration

	// DriftFactor is the clock drift factor used to calculate the validity time.
	// Default: 0.01 (1%)
	DriftFactor float64

	// TimeoutFactor is the factor applied to the expiry time for operation timeouts.
	// Default: 0.05 (5%)
	TimeoutFactor float64

	// GenValueFunc is a function that generates the unique value for the lock.
	// If not set, a random value will be generated.
	GenValueFunc func() (string, error)

	// Value is a pre-set value for the lock. If set, GenValueFunc is ignored.
	// This is useful for testing or when you need to set a specific lock value.
	Value string

	// Metadata is optional metadata associated with the lock.
	// This can be used to store additional information about the lock holder.
	Metadata map[string]string
}

// LockManager manages distributed locks using Redis as the backend.
// It provides a factory for creating Mutex instances and manages the
// underlying Redis connection pool.
//
// Usage example:
//
//	// Via dependency injection
//	func MyService(lockManager contract.LockManager) {
//	    mutex := lockManager.NewMutex("resource-key")
//	    // ...
//	}
//
//	// Via global accessor
//	mutex := goe.Lock().NewMutex("resource-key")
type LockManager interface {
	// NewMutex creates a new distributed mutex with the given name.
	// The name is used as the Redis key for the lock.
	// Options can be provided to customize the mutex behavior.
	NewMutex(name string, opts ...MutexOption) Mutex

	// Health checks the health of the Redis backend.
	// Returns nil if the backend is healthy, or an error describing the issue.
	Health(ctx context.Context) error

	// Stats returns statistics about the lock manager.
	Stats(ctx context.Context) (*LockStats, error)

	// Close closes the lock manager and releases all resources.
	// After Close is called, the manager should not be used.
	Close(ctx context.Context) error
}

// LockStats contains statistics about the lock manager.
type LockStats struct {
	// ActiveLocks is the number of locks currently held by this manager.
	ActiveLocks int64

	// TotalAcquired is the total number of locks acquired since startup.
	TotalAcquired int64

	// TotalReleased is the total number of locks released since startup.
	TotalReleased int64

	// TotalFailed is the total number of failed lock attempts since startup.
	TotalFailed int64

	// BackendLatency is the average latency to the Redis backend.
	BackendLatency time.Duration
}

// WithExpiry sets the lock expiry duration.
func WithExpiry(expiry time.Duration) MutexOption {
	return func(o *MutexOptions) {
		o.Expiry = expiry
	}
}

// WithTries sets the maximum number of lock acquisition attempts.
func WithTries(tries int) MutexOption {
	return func(o *MutexOptions) {
		o.Tries = tries
	}
}

// WithRetryDelay sets the delay between retry attempts.
func WithRetryDelay(delay time.Duration) MutexOption {
	return func(o *MutexOptions) {
		o.RetryDelay = delay
	}
}

// WithRetryDelayFunc sets a custom function for calculating retry delays.
func WithRetryDelayFunc(fn func(tries int) time.Duration) MutexOption {
	return func(o *MutexOptions) {
		o.RetryDelayFunc = fn
	}
}

// WithDriftFactor sets the clock drift factor.
func WithDriftFactor(factor float64) MutexOption {
	return func(o *MutexOptions) {
		o.DriftFactor = factor
	}
}

// WithTimeoutFactor sets the timeout factor.
func WithTimeoutFactor(factor float64) MutexOption {
	return func(o *MutexOptions) {
		o.TimeoutFactor = factor
	}
}

// WithValue sets a specific value for the lock.
func WithValue(value string) MutexOption {
	return func(o *MutexOptions) {
		o.Value = value
	}
}

// WithGenValueFunc sets a custom function for generating lock values.
func WithGenValueFunc(fn func() (string, error)) MutexOption {
	return func(o *MutexOptions) {
		o.GenValueFunc = fn
	}
}

// WithMetadata sets metadata for the lock.
func WithMetadata(metadata map[string]string) MutexOption {
	return func(o *MutexOptions) {
		o.Metadata = metadata
	}
}

// DefaultMutexOptions returns the default mutex options.
func DefaultMutexOptions() *MutexOptions {
	return &MutexOptions{
		Expiry:        8 * time.Second,
		Tries:         32,
		RetryDelay:    500 * time.Millisecond,
		DriftFactor:   0.01,
		TimeoutFactor: 0.05,
	}
}

// ApplyMutexOptions applies the given options to a MutexOptions struct.
func ApplyMutexOptions(opts *MutexOptions, options ...MutexOption) {
	for _, opt := range options {
		opt(opts)
	}
}
