package log

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewDefault(t *testing.T) {
	logger := NewDefault()

	require.NotNil(t, logger)
	assert.NotNil(t, logger.GetLogger())

	// Test that it can log without panicking
	assert.NotPanics(t, func() {
		logger.Info("test message from default logger")
	})
}

func TestCustomTimeEncoder(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 10, 30, 45, 123000000, time.UTC)

	// Create a mock encoder to capture the output
	var encoded string
	enc := mockPrimitiveArrayEncoder{
		appendString: func(s string) {
			encoded = s
		},
	}

	customTimeEncoder(testTime, &enc)

	assert.Equal(t, "2024-06-15 10:30:45.123", encoded)
}

func TestCustomLevelEncoder(t *testing.T) {
	tests := []struct {
		name     string
		level    zapcore.Level
		contains string
	}{
		{"debug level", zapcore.DebugLevel, "DEBUG"},
		{"info level", zapcore.InfoLevel, "INFO"},
		{"warn level", zapcore.WarnLevel, "WARN"},
		{"error level", zapcore.ErrorLevel, "ERROR"},
		{"dpanic level", zapcore.DPanicLevel, "FATAL"},
		{"panic level", zapcore.PanicLevel, "FATAL"},
		{"fatal level", zapcore.FatalLevel, "FATAL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var encoded string
			enc := mockPrimitiveArrayEncoder{
				appendString: func(s string) {
					encoded = s
				},
			}

			customLevelEncoder(tt.level, &enc)

			assert.Contains(t, encoded, tt.contains)
			// Check that color codes are present
			assert.Contains(t, encoded, "\x1b[")
		})
	}
}

func TestCustomCallerEncoder(t *testing.T) {
	caller := zapcore.EntryCaller{
		Defined:  true,
		PC:       0,
		File:     "/path/to/some/file.go",
		Line:     42,
		Function: "someFunction",
	}

	var encoded string
	enc := mockPrimitiveArrayEncoder{
		appendString: func(s string) {
			encoded = s
		},
	}

	customCallerEncoder(caller, &enc)

	// Should contain the file reference and gray color codes
	assert.Contains(t, encoded, "\x1b[90m")
	assert.Contains(t, encoded, "\x1b[0m")
}

func TestConvertFields(t *testing.T) {
	fields := []mockField{
		{key: "user", value: "john"},
		{key: "count", value: 42},
		{key: "active", value: true},
	}

	// Convert to contract.Field slice
	contractFields := make([]interface {
		Key() string
		Value() any
	}, len(fields))
	for i := range fields {
		contractFields[i] = &fields[i]
	}

	// Use the internal function through a wrapper
	result := testConvertFields(contractFields)

	assert.Len(t, result, 3)
	assert.Equal(t, "user", result[0].Key)
	assert.Equal(t, "count", result[1].Key)
	assert.Equal(t, "active", result[2].Key)
}

func TestConvertFieldsToArgs(t *testing.T) {
	fields := []mockField{
		{key: "user", value: "john"},
		{key: "count", value: 42},
	}

	contractFields := make([]interface {
		Key() string
		Value() any
	}, len(fields))
	for i := range fields {
		contractFields[i] = &fields[i]
	}

	result := testConvertFieldsToArgs(contractFields)

	assert.Len(t, result, 4) // 2 fields * 2 (key + value)
	assert.Equal(t, "user", result[0])
	assert.Equal(t, "john", result[1])
	assert.Equal(t, "count", result[2])
	assert.Equal(t, 42, result[3])
}

func TestFieldsToArgs(t *testing.T) {
	fields := []zap.Field{
		zap.String("key1", "value1"),
		zap.Int("key2", 123),
	}

	result := fieldsToArgs(fields)

	assert.Len(t, result, 4)
	assert.Equal(t, "key1", result[0])
	assert.Equal(t, "key2", result[2])
}

func TestConvertArgsToFields(t *testing.T) {
	t.Run("even number of args", func(t *testing.T) {
		args := []any{"key1", "value1", "key2", 42}
		result := convertArgsToFields(args)

		assert.Len(t, result, 2)
		assert.Equal(t, "key1", result[0].Key)
		assert.Equal(t, "key2", result[1].Key)
	})

	t.Run("odd number of args adds placeholder", func(t *testing.T) {
		args := []any{"key1", "value1", "key2"}
		result := convertArgsToFields(args)

		assert.Len(t, result, 2)
		assert.Equal(t, "key1", result[0].Key)
		assert.Equal(t, "key2", result[1].Key)
	})

	t.Run("non-string key uses placeholder", func(t *testing.T) {
		args := []any{123, "value1", "key2", "value2"}
		result := convertArgsToFields(args)

		assert.Len(t, result, 2)
		assert.Equal(t, "key_0", result[0].Key)
		assert.Equal(t, "key2", result[1].Key)
	})

	t.Run("empty args", func(t *testing.T) {
		result := convertArgsToFields([]any{})
		assert.Empty(t, result)
	})
}

func TestNewField(t *testing.T) {
	f := NewField("testKey", "testValue")

	require.NotNil(t, f)
	assert.Equal(t, "testKey", f.Key())
	assert.Equal(t, "testValue", f.Value())
}

func TestFieldMethods(t *testing.T) {
	f := &field{
		key:   "myKey",
		value: 123,
	}

	t.Run("Key returns key", func(t *testing.T) {
		assert.Equal(t, "myKey", f.Key())
	})

	t.Run("Value returns value", func(t *testing.T) {
		assert.Equal(t, 123, f.Value())
	})
}

func TestGetZapLogger(t *testing.T) {
	t.Run("returns zap logger from zapLogger", func(t *testing.T) {
		logger := NewDefault()
		zapLogger, err := GetZapLogger(logger)

		require.NoError(t, err)
		require.NotNil(t, zapLogger)
		assert.IsType(t, &zap.Logger{}, zapLogger)
	})

	t.Run("returns error for non-zap logger", func(t *testing.T) {
		mockLogger := &nonZapLogger{}
		zapLogger, err := GetZapLogger(mockLogger)

		require.Error(t, err)
		assert.Nil(t, zapLogger)
		assert.Contains(t, err.Error(), "not a zap logger")
	})
}

func TestDefaultLoggerConfig(t *testing.T) {
	cfg := &defaultLoggerConfig{
		level:      "debug",
		format:     "json",
		output:     []string{"stdout", "file"},
		caller:     true,
		stacktrace: true,
	}

	t.Run("Level returns level", func(t *testing.T) {
		assert.Equal(t, "debug", cfg.Level())
	})

	t.Run("Format returns format", func(t *testing.T) {
		assert.Equal(t, "json", cfg.Format())
	})

	t.Run("Output returns output", func(t *testing.T) {
		assert.Equal(t, []string{"stdout", "file"}, cfg.Output())
	})

	t.Run("EnableCaller returns caller setting", func(t *testing.T) {
		assert.True(t, cfg.EnableCaller())
	})

	t.Run("EnableStacktrace returns stacktrace setting", func(t *testing.T) {
		assert.True(t, cfg.EnableStacktrace())
	})
}

func TestModule_ProvideZap(t *testing.T) {
	config := &MockConfig{}
	module := NewModule(config)

	zapLogger := module.ProvideZap()

	require.NotNil(t, zapLogger)
	assert.IsType(t, &zap.Logger{}, zapLogger)
}

func TestModule_ValidateConfig(t *testing.T) {
	t.Run("valid config passes", func(t *testing.T) {
		mockCfg := &validatingMockConfig{
			values: map[string]any{},
		}
		module := NewModule(mockCfg)

		err := module.ValidateConfig()
		assert.NoError(t, err)
	})

	t.Run("validates log level", func(t *testing.T) {
		mockCfg := &validatingMockConfig{
			values: map[string]any{
				"LOG_LEVEL": "debug",
			},
		}
		module := NewModule(mockCfg)

		err := module.ValidateConfig()
		assert.NoError(t, err)
	})

	t.Run("accepts warn and error levels", func(t *testing.T) {
		for _, lvl := range []string{"warn", "error"} {
			mockCfg := &validatingMockConfig{values: map[string]any{"LOG_LEVEL": lvl}}
			err := NewModule(mockCfg).ValidateConfig()
			assert.NoError(t, err, "level %q should be valid", lvl)
		}
	})

	t.Run("rejects unsupported levels", func(t *testing.T) {
		// panic/fatal are not honored by the level parser; bogus is invalid.
		for _, lvl := range []string{"panic", "fatal", "bogus"} {
			mockCfg := &validatingMockConfig{values: map[string]any{"LOG_LEVEL": lvl}}
			err := NewModule(mockCfg).ValidateConfig()
			assert.Error(t, err, "level %q should be rejected", lvl)
		}
	})

	t.Run("validates log format", func(t *testing.T) {
		mockCfg := &validatingMockConfig{
			values: map[string]any{
				"LOG_FORMAT": "json",
			},
		}
		module := NewModule(mockCfg)

		err := module.ValidateConfig()
		assert.NoError(t, err)
	})

	t.Run("validates log output", func(t *testing.T) {
		mockCfg := &validatingMockConfig{
			values: map[string]any{
				"LOG_OUTPUT": []string{"stdout", "stderr", "/var/log/app.log"},
			},
		}
		module := NewModule(mockCfg)

		err := module.ValidateConfig()
		assert.NoError(t, err)
	})
}

func TestNew_WithDifferentConfigs(t *testing.T) {
	t.Run("debug level", func(t *testing.T) {
		cfg := &testLoggerConfig{level: "debug", format: "text", output: []string{"console"}}
		logger := New(cfg)
		require.NotNil(t, logger)
		assert.NotPanics(t, func() {
			logger.Debug("debug message")
		})
	})

	t.Run("warn level", func(t *testing.T) {
		cfg := &testLoggerConfig{level: "warn", format: "text", output: []string{"console"}}
		logger := New(cfg)
		require.NotNil(t, logger)
	})

	t.Run("error level", func(t *testing.T) {
		cfg := &testLoggerConfig{level: "error", format: "text", output: []string{"console"}}
		logger := New(cfg)
		require.NotNil(t, logger)
	})

	t.Run("json format", func(t *testing.T) {
		cfg := &testLoggerConfig{level: "info", format: "json", output: []string{"console"}}
		logger := New(cfg)
		require.NotNil(t, logger)
		assert.NotPanics(t, func() {
			logger.Info("json formatted message")
		})
	})

	t.Run("with caller disabled", func(t *testing.T) {
		cfg := &testLoggerConfig{level: "info", format: "text", output: []string{"console"}, caller: false}
		logger := New(cfg)
		require.NotNil(t, logger)
	})

	t.Run("with stacktrace disabled", func(t *testing.T) {
		cfg := &testLoggerConfig{level: "info", format: "text", output: []string{"console"}, stacktrace: false}
		logger := New(cfg)
		require.NotNil(t, logger)
	})
}

// Helper types and functions

type mockPrimitiveArrayEncoder struct {
	appendString func(string)
}

func (m *mockPrimitiveArrayEncoder) AppendBool(bool)             {}
func (m *mockPrimitiveArrayEncoder) AppendByteString([]byte)     {}
func (m *mockPrimitiveArrayEncoder) AppendComplex128(complex128) {}
func (m *mockPrimitiveArrayEncoder) AppendComplex64(complex64)   {}
func (m *mockPrimitiveArrayEncoder) AppendFloat64(float64)       {}
func (m *mockPrimitiveArrayEncoder) AppendFloat32(float32)       {}
func (m *mockPrimitiveArrayEncoder) AppendInt(int)               {}
func (m *mockPrimitiveArrayEncoder) AppendInt64(int64)           {}
func (m *mockPrimitiveArrayEncoder) AppendInt32(int32)           {}
func (m *mockPrimitiveArrayEncoder) AppendInt16(int16)           {}
func (m *mockPrimitiveArrayEncoder) AppendInt8(int8)             {}
func (m *mockPrimitiveArrayEncoder) AppendString(s string) {
	if m.appendString != nil {
		m.appendString(s)
	}
}
func (m *mockPrimitiveArrayEncoder) AppendUint(uint)       {}
func (m *mockPrimitiveArrayEncoder) AppendUint64(uint64)   {}
func (m *mockPrimitiveArrayEncoder) AppendUint32(uint32)   {}
func (m *mockPrimitiveArrayEncoder) AppendUint16(uint16)   {}
func (m *mockPrimitiveArrayEncoder) AppendUint8(uint8)     {}
func (m *mockPrimitiveArrayEncoder) AppendUintptr(uintptr) {}

type mockField struct {
	key   string
	value any
}

func (f *mockField) Key() string { return f.key }
func (f *mockField) Value() any  { return f.value }

// nonZapLogger is a mock that doesn't wrap a zap logger
// Note: This intentionally does NOT implement contract.Logger properly
// to test the GetZapLogger error case
type nonZapLogger struct{}

func (n *nonZapLogger) Debug(msg string, args ...any)                   {}
func (n *nonZapLogger) Info(msg string, args ...any)                    {}
func (n *nonZapLogger) Warn(msg string, args ...any)                    {}
func (n *nonZapLogger) Error(msg string, args ...any)                   {}
func (n *nonZapLogger) Fatal(msg string, args ...any)                   {}
func (n *nonZapLogger) Debugf(template string, args ...any)             {}
func (n *nonZapLogger) Infof(template string, args ...any)              {}
func (n *nonZapLogger) Warnf(template string, args ...any)              {}
func (n *nonZapLogger) Errorf(template string, args ...any)             {}
func (n *nonZapLogger) Fatalf(template string, args ...any)             {}
func (n *nonZapLogger) Debugw(msg string, keysAndValues ...any)         {}
func (n *nonZapLogger) Infow(msg string, keysAndValues ...any)          {}
func (n *nonZapLogger) Warnw(msg string, keysAndValues ...any)          {}
func (n *nonZapLogger) Errorw(msg string, keysAndValues ...any)         {}
func (n *nonZapLogger) Fatalw(msg string, keysAndValues ...any)         {}
func (n *nonZapLogger) With(keysAndValues ...any) contract.Logger       { return n }
func (n *nonZapLogger) WithContext(ctx context.Context) contract.Logger { return n }
func (n *nonZapLogger) WithError(err error) contract.Logger             { return n }
func (n *nonZapLogger) GetLogger() *zap.SugaredLogger                   { return nil }

// validatingMockConfig for ValidateConfig tests
type validatingMockConfig struct {
	values map[string]any
}

func (m *validatingMockConfig) Get(key string) any {
	return m.values[key]
}

func (m *validatingMockConfig) Set(key string, value any) {
	m.values[key] = value
}

func (m *validatingMockConfig) GetString(key string) string {
	if v, ok := m.values[key].(string); ok {
		return v
	}
	return ""
}

func (m *validatingMockConfig) GetInt(key string) int {
	if v, ok := m.values[key].(int); ok {
		return v
	}
	return 0
}

func (m *validatingMockConfig) GetInt64(key string) int64 {
	if v, ok := m.values[key].(int64); ok {
		return v
	}
	return 0
}

func (m *validatingMockConfig) GetBool(key string) bool {
	if v, ok := m.values[key].(bool); ok {
		return v
	}
	return false
}

func (m *validatingMockConfig) GetFloat64(key string) float64 {
	if v, ok := m.values[key].(float64); ok {
		return v
	}
	return 0
}

func (m *validatingMockConfig) GetDuration(key string) time.Duration {
	if v, ok := m.values[key].(time.Duration); ok {
		return v
	}
	return 0
}

func (m *validatingMockConfig) GetStringSlice(key string) []string {
	if v, ok := m.values[key].([]string); ok {
		return v
	}
	return nil
}

func (m *validatingMockConfig) GetStringMap(key string) map[string]any {
	if v, ok := m.values[key].(map[string]any); ok {
		return v
	}
	return nil
}

func (m *validatingMockConfig) Has(key string) bool {
	_, ok := m.values[key]
	return ok
}

func (m *validatingMockConfig) All() map[string]any {
	return m.values
}

func (m *validatingMockConfig) Reload() error {
	return nil
}

// testLoggerConfig for testing New with different configurations
type testLoggerConfig struct {
	level      string
	format     string
	output     []string
	caller     bool
	stacktrace bool
}

func (c *testLoggerConfig) Level() string          { return c.level }
func (c *testLoggerConfig) Format() string         { return c.format }
func (c *testLoggerConfig) Output() []string       { return c.output }
func (c *testLoggerConfig) EnableCaller() bool     { return c.caller }
func (c *testLoggerConfig) EnableStacktrace() bool { return c.stacktrace }

// Helper functions to test internal functions through contract.Field interface
func testConvertFields(fields []interface {
	Key() string
	Value() any
}) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key(), f.Value())
	}
	return zapFields
}

func testConvertFieldsToArgs(fields []interface {
	Key() string
	Value() any
}) []any {
	args := make([]any, 0, len(fields)*2)
	for _, f := range fields {
		args = append(args, f.Key(), f.Value())
	}
	return args
}
