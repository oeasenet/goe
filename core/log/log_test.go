package log_test

import (
	"context"
	"os"
	"testing"

	"go.oease.dev/goe/v2/core/log"
)

func TestLogger(t *testing.T) {
	// Set environment to development
	os.Setenv("GOE_ENV", "dev")

	// Create a new logger
	logger := log.New()

	// Test basic logging
	t.Run("Basic Logging", func(t *testing.T) {
		// These should not panic
		logger.Debug("This is a debug message")
		logger.Info("This is an info message")
		logger.Warn("This is a warning message")
		logger.Error("This is an error message")
		// Don't test Fatal as it would exit the program
	})

	// Test logging with fields
	t.Run("Logging with Fields", func(t *testing.T) {
		// These should not panic
		logger.Info("User logged in", "user_id", 123, "username", "john")
	})

	// Test named logger
	t.Run("Named Logger", func(t *testing.T) {
		// Create a named logger
		namedLogger := logger.Named("test")
		// This should not panic
		namedLogger.Info("This is a message from the named logger")
	})

	// Test logger with context
	t.Run("Logger with Context", func(t *testing.T) {
		// Create a context
		ctx := context.Background()
		// Create a logger with context
		ctxLogger := logger.WithContext(ctx)
		// This should not panic
		ctxLogger.Info("This is a message from the context logger")
	})

	// Test logger with fields
	t.Run("Logger with Fields", func(t *testing.T) {
		// Create a logger with fields
		fieldsLogger := logger.WithFields(map[string]interface{}{
			"user_id":  123,
			"username": "john",
		})
		// This should not panic
		fieldsLogger.Info("This is a message from the fields logger")
	})

	// Test logger with field
	t.Run("Logger with Field", func(t *testing.T) {
		// Create a logger with a field
		fieldLogger := logger.WithField("request_id", "abc123")
		// This should not panic
		fieldLogger.Info("This is a message from the field logger")
	})

	// Test setting and getting log level
	t.Run("Log Level", func(t *testing.T) {
		// Set log level to info
		err := logger.SetLevel("info")
		if err != nil {
			t.Errorf("Failed to set log level: %v", err)
		}

		// Get log level
		level := logger.GetLevel()
		if level != "info" {
			t.Errorf("Expected log level to be 'info', got '%s'", level)
		}

		// Set log level to debug
		err = logger.SetLevel("debug")
		if err != nil {
			t.Errorf("Failed to set log level: %v", err)
		}

		// Get log level
		level = logger.GetLevel()
		if level != "debug" {
			t.Errorf("Expected log level to be 'debug', got '%s'", level)
		}

		// Set log level to an invalid value
		err = logger.SetLevel("invalid")
		if err == nil {
			t.Errorf("Expected error when setting invalid log level")
		}
	})

	// Test module lifecycle
	t.Run("Module Lifecycle", func(t *testing.T) {
		// Initialize
		err := logger.Initialize(context.Background())
		if err != nil {
			t.Errorf("Failed to initialize logger: %v", err)
		}

		// Start
		err = logger.Start(context.Background())
		if err != nil {
			t.Errorf("Failed to start logger: %v", err)
		}

		// Stop
		err = logger.Stop(context.Background())
		if err != nil {
			t.Errorf("Failed to stop logger: %v", err)
		}
	})

	// Test module name
	t.Run("Module Name", func(t *testing.T) {
		name := logger.Name()
		if name != "log" {
			t.Errorf("Expected module name to be 'log', got '%s'", name)
		}
	})

	// Print a message to show that colorful logging is working
	t.Log("[DEBUG_LOG] The following logs should be colorful in development environment:")
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")
}
