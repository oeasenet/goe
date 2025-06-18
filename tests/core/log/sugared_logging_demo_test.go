package log

import (
	"go.oease.dev/goe/v2/core/log"
	"testing"
)

func TestSugaredLoggingDemo(t *testing.T) {
	// Create a logger with default configuration
	logger := log.NewDefault()

	// Test printf-style logging methods
	t.Run("Printf-style logging", func(t *testing.T) {
		logger.Infof("This is a printf-style log message with number: %d and string: %s", 42, "hello")
		logger.Debugf("Debug message with value: %v", map[string]int{"count": 5})
		logger.Warnf("Warning message with percentage: %.2f%%", 85.5)
		logger.Errorf("Error occurred: %s", "connection timeout")

		// Test passes if no panic occurs
		t.Log("Printf-style logging methods work correctly")
	})

	// Test key-value logging methods
	t.Run("Key-value logging", func(t *testing.T) {
		logger.Infow("This is a key-value log message", "user_id", 123, "action", "login", "success", true)
		logger.Debugw("Debug with context", "request_id", "abc-123", "duration_ms", 250)
		logger.Warnw("Warning message", "reason", "rate limit exceeded", "limit", 100)
		logger.Errorw("Error with details", "error_code", 500, "message", "internal server error")

		// Test passes if no panic occurs
		t.Log("Key-value logging methods work correctly")
	})

	// Test that the logger still supports structured logging
	t.Run("Structured logging compatibility", func(t *testing.T) {
		field := log.NewField("test_key", "test_value")
		logger.Info("Structured log message", field)

		// Test passes if no panic occurs
		t.Log("Structured logging still works correctly")
	})
}
