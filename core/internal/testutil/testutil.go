// Package testutil provides shared testing utilities for the GOE framework.
// These utilities help reduce boilerplate in test files across packages.
package testutil

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

// WaitForCondition waits for a condition to become true within a timeout.
// Returns true if the condition was met, false if timeout occurred.
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, description string) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("Condition not met within %v: %s", timeout, description)
	return false
}

// WithTimeout creates a context with the specified timeout and calls cancel after the function completes.
func WithTimeout(t *testing.T, timeout time.Duration, fn func(ctx context.Context)) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	fn(ctx)
}

// ConcurrentRunner helps run concurrent test operations and collect results.
type ConcurrentRunner struct {
	wg      sync.WaitGroup
	errChan chan error
	mu      sync.Mutex
	errors  []error
}

// NewConcurrentRunner creates a new concurrent runner with the specified error channel buffer size.
func NewConcurrentRunner(bufferSize int) *ConcurrentRunner {
	return &ConcurrentRunner{
		errChan: make(chan error, bufferSize),
		errors:  make([]error, 0),
	}
}

// Run executes the given function concurrently.
func (r *ConcurrentRunner) Run(fn func() error) {
	r.wg.Go(func() {
		if err := fn(); err != nil {
			r.mu.Lock()
			r.errors = append(r.errors, err)
			r.mu.Unlock()
		}
	})
}

// Wait waits for all goroutines to complete and returns any errors.
func (r *ConcurrentRunner) Wait() []error {
	r.wg.Wait()
	return r.errors
}

// RequireNoErrors fails the test if there are any errors in the runner.
func (r *ConcurrentRunner) RequireNoErrors(t *testing.T) {
	t.Helper()
	r.wg.Wait()
	if len(r.errors) > 0 {
		t.Errorf("Concurrent operations had %d errors:", len(r.errors))
		for i, err := range r.errors {
			t.Errorf("  Error %d: %v", i+1, err)
		}
		t.FailNow()
	}
}

// AssertEventually asserts that a condition eventually becomes true within the timeout.
func AssertEventually(t *testing.T, condition func() bool, timeout time.Duration, msgAndArgs ...any) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(msgAndArgs) > 0 {
		t.Errorf("Condition not met within %v: %v", timeout, msgAndArgs)
	} else {
		t.Errorf("Condition not met within %v", timeout)
	}
	return false
}

// AssertNever asserts that a condition never becomes true within the duration.
func AssertNever(t *testing.T, condition func() bool, duration time.Duration, msgAndArgs ...any) bool {
	t.Helper()
	deadline := time.Now().Add(duration)
	for time.Now().Before(deadline) {
		if condition() {
			if len(msgAndArgs) > 0 {
				t.Errorf("Condition unexpectedly became true: %v", msgAndArgs)
			} else {
				t.Errorf("Condition unexpectedly became true within %v", duration)
			}
			return false
		}
		time.Sleep(10 * time.Millisecond)
	}
	return true
}

// RetryOperation retries an operation with exponential backoff.
func RetryOperation(t *testing.T, maxAttempts int, baseDelay time.Duration, op func() error) error {
	t.Helper()
	var lastErr error
	delay := baseDelay
	for range maxAttempts {
		if err := op(); err != nil {
			lastErr = err
			time.Sleep(delay)
			delay *= 2
			continue
		}
		return nil
	}
	return lastErr
}

// Cleanup registers a cleanup function that will be called after the test completes.
// This is a helper for tests that need to ensure resources are cleaned up.
func Cleanup(t *testing.T, cleanupFn func()) {
	t.Helper()
	t.Cleanup(cleanupFn)
}

// GetFreePort returns an available port on the system.
// It binds to port 0, gets the assigned port, and closes the listener.
// The port is then available for use.
func GetFreePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to get free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()
	return port
}
