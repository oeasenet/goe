package validator

import (
	"context"
	"time"

	"go.oease.dev/goe/v2/contract"
	"go.uber.org/zap"
)

// Mock config for testing
type mockConfig struct {
	values map[string]interface{}
}

func (m *mockConfig) Get(key string) interface{} {
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

func (m *mockConfig) GetInt(key string) int {
	if val, ok := m.values[key]; ok {
		if i, ok := val.(int); ok {
			return i
		}
	}
	return 0
}

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

func (m *mockConfig) GetDuration(key string) time.Duration           { return 0 }
func (m *mockConfig) GetStringSlice(key string) []string             { return nil }
func (m *mockConfig) GetStringMap(key string) map[string]interface{} { return nil }
func (m *mockConfig) Set(key string, value interface{})              { m.values[key] = value }

func (m *mockConfig) Has(key string) bool {
	_, ok := m.values[key]
	return ok
}

func (m *mockConfig) All() map[string]interface{} { return m.values }
func (m *mockConfig) Reload() error               { return nil }

type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields ...interface{}) {}
func (m *mockLogger) Info(msg string, fields ...interface{})  {}
func (m *mockLogger) Warn(msg string, fields ...interface{})  {}
func (m *mockLogger) Error(msg string, fields ...interface{}) {}
func (m *mockLogger) Fatal(msg string, fields ...interface{}) {}
func (m *mockLogger) Panic(msg string, fields ...interface{}) {}

// Printf-style methods
func (m *mockLogger) Debugf(template string, args ...interface{}) {}
func (m *mockLogger) Infof(template string, args ...interface{})  {}
func (m *mockLogger) Warnf(template string, args ...interface{})  {}
func (m *mockLogger) Errorf(template string, args ...interface{}) {}
func (m *mockLogger) Fatalf(template string, args ...interface{}) {}
func (m *mockLogger) Panicf(template string, args ...interface{}) {}

// Key-value methods
func (m *mockLogger) Debugw(msg string, keysAndValues ...interface{}) {}
func (m *mockLogger) Infow(msg string, keysAndValues ...interface{})  {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...interface{})  {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...interface{}) {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...interface{}) {}

// Context methods
func (m *mockLogger) With(keysAndValues ...interface{}) contract.Logger { return m }
func (m *mockLogger) WithContext(ctx context.Context) contract.Logger   { return m }
func (m *mockLogger) WithError(err error) contract.Logger               { return m }
func (m *mockLogger) GetLogger() *zap.SugaredLogger                     { return nil }
