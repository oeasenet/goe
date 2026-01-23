package testutil

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaitForCondition(t *testing.T) {
	t.Run("condition met immediately", func(t *testing.T) {
		result := WaitForCondition(t, func() bool { return true }, 100*time.Millisecond, "immediate true")
		assert.True(t, result)
	})

	t.Run("condition met after delay", func(t *testing.T) {
		counter := 0
		result := WaitForCondition(t, func() bool {
			counter++
			return counter >= 3
		}, 100*time.Millisecond, "delayed true")
		assert.True(t, result)
	})
}

func TestWithTimeout(t *testing.T) {
	t.Run("completes within timeout", func(t *testing.T) {
		var executed bool
		WithTimeout(t, 100*time.Millisecond, func(ctx context.Context) {
			executed = true
			assert.NotNil(t, ctx)
		})
		assert.True(t, executed)
	})

	t.Run("context has deadline", func(t *testing.T) {
		WithTimeout(t, 100*time.Millisecond, func(ctx context.Context) {
			deadline, ok := ctx.Deadline()
			assert.True(t, ok)
			assert.True(t, time.Until(deadline) <= 100*time.Millisecond)
		})
	})
}

func TestConcurrentRunner(t *testing.T) {
	t.Run("runs multiple operations", func(t *testing.T) {
		runner := NewConcurrentRunner(10)
		var counter int32

		for i := 0; i < 5; i++ {
			runner.Run(func() error {
				atomic.AddInt32(&counter, 1)
				return nil
			})
		}

		errs := runner.Wait()
		assert.Empty(t, errs)
		assert.Equal(t, int32(5), counter)
	})

	t.Run("collects errors", func(t *testing.T) {
		runner := NewConcurrentRunner(10)

		runner.Run(func() error {
			return errors.New("error 1")
		})
		runner.Run(func() error {
			return errors.New("error 2")
		})
		runner.Run(func() error {
			return nil // no error
		})

		errs := runner.Wait()
		assert.Len(t, errs, 2)
	})
}

func TestAssertEventually(t *testing.T) {
	t.Run("condition becomes true", func(t *testing.T) {
		counter := 0
		result := AssertEventually(t, func() bool {
			counter++
			return counter >= 2
		}, 100*time.Millisecond)
		assert.True(t, result)
	})
}

func TestAssertNever(t *testing.T) {
	t.Run("condition stays false", func(t *testing.T) {
		result := AssertNever(t, func() bool {
			return false
		}, 50*time.Millisecond)
		assert.True(t, result)
	})
}

func TestRetryOperation(t *testing.T) {
	t.Run("succeeds on first attempt", func(t *testing.T) {
		attempts := 0
		err := RetryOperation(t, 3, 1*time.Millisecond, func() error {
			attempts++
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("succeeds on retry", func(t *testing.T) {
		attempts := 0
		err := RetryOperation(t, 3, 1*time.Millisecond, func() error {
			attempts++
			if attempts < 2 {
				return errors.New("transient error")
			}
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, attempts)
	})

	t.Run("fails after max attempts", func(t *testing.T) {
		attempts := 0
		err := RetryOperation(t, 3, 1*time.Millisecond, func() error {
			attempts++
			return errors.New("persistent error")
		})
		assert.Error(t, err)
		assert.Equal(t, 3, attempts)
	})
}

func TestCleanup(t *testing.T) {
	t.Run("registers cleanup", func(t *testing.T) {
		// This just tests that the function doesn't panic
		Cleanup(t, func() {
			// cleanup logic
		})
	})
}
