package log

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// MockLoggerConfig for testing
type MockLoggerConfig struct{}

func (m *MockLoggerConfig) Level() string          { return "info" }
func (m *MockLoggerConfig) Format() string         { return "json" }
func (m *MockLoggerConfig) Output() []string       { return []string{"stdout"} }
func (m *MockLoggerConfig) EnableCaller() bool     { return true }
func (m *MockLoggerConfig) EnableStacktrace() bool { return true }

// MockConfig for testing
type MockConfig struct{}

func (m *MockConfig) Get(key string) any                     { return nil }
func (m *MockConfig) GetString(key string) string            { return "" }
func (m *MockConfig) GetInt(key string) int                  { return 0 }
func (m *MockConfig) GetInt64(key string) int64              { return 0 }
func (m *MockConfig) GetFloat64(key string) float64          { return 0.0 }
func (m *MockConfig) GetBool(key string) bool                { return false }
func (m *MockConfig) GetDuration(key string) time.Duration   { return 0 }
func (m *MockConfig) GetStringSlice(key string) []string     { return []string{} }
func (m *MockConfig) GetStringMap(key string) map[string]any { return map[string]any{} }
func (m *MockConfig) Set(key string, value any)              {}
func (m *MockConfig) Has(key string) bool                    { return false }
func (m *MockConfig) All() map[string]any                    { return map[string]any{} }
func (m *MockConfig) Reload() error                          { return nil }

func TestLog_New(t *testing.T) {
	config := &MockLoggerConfig{}
	logger := New(config)

	assert.NotNil(t, logger)
	assert.NotNil(t, logger.GetLogger())
}

func TestLogModule_New(t *testing.T) {
	config := &MockConfig{}
	module := NewModule(config)

	assert.NotNil(t, module)
	assert.Equal(t, "log", module.Name())
	assert.NotNil(t, module.Provide())
}

func TestLogModule_Lifecycle(t *testing.T) {
	config := &MockConfig{}
	module := NewModule(config)

	ctx := context.Background()

	t.Run("start module", func(t *testing.T) {
		err := module.OnStart(ctx)
		assert.NoError(t, err)
	})

	t.Run("stop module", func(t *testing.T) {
		err := module.OnStop(ctx)
		assert.NoError(t, err)
	})
}

func TestLogger_Methods(t *testing.T) {
	config := &MockLoggerConfig{}
	logger := New(config)

	t.Run("debug", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.Debug("test debug message")
		})
	})

	t.Run("info", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.Info("test info message")
		})
	})

	t.Run("warn", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.Warn("test warn message")
		})
	})

	t.Run("error", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.Error("test error message")
		})
	})

	t.Run("debugf", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.Debugf("test debug message %s", "formatted")
		})
	})

	t.Run("infof", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.Infof("test info message %s", "formatted")
		})
	})

	t.Run("warnf", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.Warnf("test warn message %s", "formatted")
		})
	})

	t.Run("errorf", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.Errorf("test error message %s", "formatted")
		})
	})

	t.Run("debugw", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.Debugw("test debug message", "key", "value")
		})
	})

	t.Run("infow", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.Infow("test info message", "key", "value")
		})
	})

	t.Run("warnw", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.Warnw("test warn message", "key", "value")
		})
	})

	t.Run("errorw", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger.Errorw("test error message", "key", "value")
		})
	})

	t.Run("with", func(t *testing.T) {
		childLogger := logger.With("key", "value")
		assert.NotNil(t, childLogger)
		// Note: With() returns a new logger instance
		assert.IsType(t, logger, childLogger)
	})

	t.Run("with context", func(t *testing.T) {
		ctx := context.Background()
		childLogger := logger.WithContext(ctx)
		assert.NotNil(t, childLogger)
	})

	t.Run("with error", func(t *testing.T) {
		err := assert.AnError
		childLogger := logger.WithError(err)
		assert.NotNil(t, childLogger)
	})

	t.Run("get logger", func(t *testing.T) {
		zapLogger := logger.GetLogger()
		assert.NotNil(t, zapLogger)
		assert.IsType(t, &zap.SugaredLogger{}, zapLogger)
	})
}
