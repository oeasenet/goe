package tests

import (
	"fmt"
	"testing"
	"time"

	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/core/log"
)

func TestGoeBasicFunctionality(t *testing.T) {
	// Create a new application
	app := goe.New(goe.Options{
		Name:        "Test App",
		Version:     "1.0.0",
		Environment: "test",
	})

	// Test that app was created
	if app == nil {
		t.Fatal("Failed to create application")
	}

	// Test global accessors
	if goe.App() == nil {
		t.Fatal("Global App() accessor returned nil")
	}

	if goe.Config() == nil {
		t.Fatal("Global Config() accessor returned nil")
	}

	if goe.Log() == nil {
		t.Fatal("Global Log() accessor returned nil")
	}

	// Test app properties
	if app.Name() != "Test App" {
		t.Errorf("Expected app name 'Test App', got '%s'", app.Name())
	}

	if app.Version() != "1.0.0" {
		t.Errorf("Expected app version '1.0.0', got '%s'", app.Version())
	}

	if app.Environment() != "test" {
		t.Errorf("Expected app environment 'test', got '%s'", app.Environment())
	}
}

func TestConfigModule(t *testing.T) {
	// Create app
	_ = goe.New()

	config := goe.Config()

	// Test setting and getting values
	config.Set("test_key", "test_value")

	if !config.Has("test_key") {
		t.Error("Config should have 'test_key'")
	}

	val := config.GetString("test_key")
	if val != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", val)
	}

	// Test different types
	config.Set("test_int", "42")
	intVal := config.GetInt("test_int")
	if intVal != 42 {
		t.Errorf("Expected 42, got %d", intVal)
	}

	config.Set("test_bool", "true")
	boolVal := config.GetBool("test_bool")
	if !boolVal {
		t.Error("Expected true, got false")
	}

	config.Set("test_duration", "5s")
	durVal := config.GetDuration("test_duration")
	if durVal != 5*time.Second {
		t.Errorf("Expected 5s, got %v", durVal)
	}
}

func TestLogModule(t *testing.T) {
	// Create app
	_ = goe.New()

	logger := goe.Log()

	// Test basic logging (should not panic)
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message")

	// Test logging with fields
	logger.Info("Test with fields",
		log.NewField("field1", "value1"),
		log.NewField("field2", 123),
	)

	// Test logger with fields
	newLogger := logger.With(
		log.NewField("request_id", "12345"),
	)
	newLogger.Info("Message with request ID")

	// Test logger with error
	newLogger = logger.WithError(fmt.Errorf("test error"))
	newLogger.Error("Operation failed")
}
