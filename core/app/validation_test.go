package app

import (
	"context"
	"testing"
	"time"

	"go.oease.dev/goe/v2/contract"
	"go.uber.org/zap"
)

func TestApplicationValidation(t *testing.T) {
	// Mock config
	config := &mockConfig{
		values: map[string]any{
			"DISABLE_CONFIG_VALIDATION": "false",
		},
	}

	// Mock logger
	logger := &mockLogger{}

	// Create application
	testApp := New("test-app", "1.0.0", "test").(*app)

	// Set config and logger
	testApp.SetConfig(config)
	testApp.SetLogger(logger)

	// Add a mock module that validates successfully
	validModule := &mockValidatableModule{
		name: "valid",
		validateFunc: func() error {
			return nil
		},
	}
	_ = testApp.AddModule(validModule)

	// Test validation
	ctx := context.Background()
	err := testApp.Start(ctx)

	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	} else {
		_ = testApp.Stop(ctx)
	}

	// Test with invalid module
	invalidModule := &mockValidatableModule{
		name: "invalid",
		validateFunc: func() error {
			return &contract.ConfigValidationError{
				Module:      "invalid",
				MissingKeys: []string{"REQUIRED_KEY"},
			}
		},
	}

	testApp2 := New("test-app-2", "1.0.0", "test").(*app)
	testApp2.SetConfig(config)
	testApp2.SetLogger(logger)
	_ = testApp2.AddModule(invalidModule)

	err = testApp2.Start(ctx)
	if err == nil {
		t.Error("Expected validation to fail but it passed")
		_ = testApp2.Stop(ctx)
	}
}

// Mock implementations
type mockValidatableModule struct {
	name         string
	validateFunc func() error
}

func (m *mockValidatableModule) Name() string {
	return m.name
}

func (m *mockValidatableModule) OnStart(ctx context.Context) error {
	return nil
}

func (m *mockValidatableModule) OnStop(ctx context.Context) error {
	return nil
}

func (m *mockValidatableModule) ValidateConfig() error {
	if m.validateFunc != nil {
		return m.validateFunc()
	}
	return nil
}

type mockConfig struct {
	values map[string]any
}

func (m *mockConfig) Get(key string) any {
	return m.values[key]
}

func (m *mockConfig) GetString(key string) string {
	if val, ok := m.values[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (m *mockConfig) GetInt(key string) int         { return 0 }
func (m *mockConfig) GetInt64(key string) int64     { return 0 }
func (m *mockConfig) GetFloat64(key string) float64 { return 0 }

func (m *mockConfig) GetBool(key string) bool {
	if val, ok := m.values[key]; ok {
		if str, ok := val.(string); ok {
			return str == "true"
		}
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func (m *mockConfig) GetDuration(key string) time.Duration   { return 0 }
func (m *mockConfig) GetStringSlice(key string) []string     { return nil }
func (m *mockConfig) GetStringMap(key string) map[string]any { return nil }
func (m *mockConfig) Set(key string, value any)              { m.values[key] = value }

func (m *mockConfig) Has(key string) bool {
	_, ok := m.values[key]
	return ok
}

func (m *mockConfig) All() map[string]any { return m.values }
func (m *mockConfig) Reload() error       { return nil }

type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields ...any) {}
func (m *mockLogger) Info(msg string, fields ...any)  {}
func (m *mockLogger) Warn(msg string, fields ...any)  {}
func (m *mockLogger) Error(msg string, fields ...any) {}
func (m *mockLogger) Fatal(msg string, fields ...any) {}
func (m *mockLogger) Panic(msg string, fields ...any) {}

// Printf-style methods
func (m *mockLogger) Debugf(template string, args ...any) {}
func (m *mockLogger) Infof(template string, args ...any)  {}
func (m *mockLogger) Warnf(template string, args ...any)  {}
func (m *mockLogger) Errorf(template string, args ...any) {}
func (m *mockLogger) Fatalf(template string, args ...any) {}
func (m *mockLogger) Panicf(template string, args ...any) {}

// Key-value methods
func (m *mockLogger) Debugw(msg string, keysAndValues ...any) {}
func (m *mockLogger) Infow(msg string, keysAndValues ...any)  {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)  {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any) {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any) {}

// Context methods
func (m *mockLogger) With(keysAndValues ...any) contract.Logger       { return m }
func (m *mockLogger) WithContext(ctx context.Context) contract.Logger { return m }
func (m *mockLogger) WithError(err error) contract.Logger             { return m }
func (m *mockLogger) GetLogger() *zap.SugaredLogger                   { return nil }

func TestBootstrap(t *testing.T) {
	// Test successful bootstrap
	app, err := Bootstrap("test-app", "1.0.0", "test")
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	if app == nil {
		t.Fatal("Bootstrap returned nil app")
	}
}

func TestBootstrapAndValidate(t *testing.T) {
	// Test successful bootstrap and validation
	app, err := BootstrapAndValidate("test-app", "1.0.0", "test")
	if err != nil {
		t.Fatalf("BootstrapAndValidate failed: %v", err)
	}

	if app == nil {
		t.Fatal("BootstrapAndValidate returned nil app")
	}

	// The app should not be running after validation
	if app.IsRunning() {
		t.Error("App should not be running after validation")
	}
}
