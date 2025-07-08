package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/zap"
)

// MockLogger is a mock implementation of the Logger interface
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Info(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Warn(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Error(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Fatal(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Debugf(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Infof(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Warnf(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Errorf(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Fatalf(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Debugw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Infow(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Warnw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Errorw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Fatalw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) With(keysAndValues ...any) contract.Logger {
	args := m.Called(keysAndValues)
	return args.Get(0).(contract.Logger)
}

func (m *MockLogger) WithContext(ctx context.Context) contract.Logger {
	args := m.Called(ctx)
	return args.Get(0).(contract.Logger)
}

func (m *MockLogger) WithError(err error) contract.Logger {
	args := m.Called(err)
	return args.Get(0).(contract.Logger)
}

func (m *MockLogger) GetLogger() *zap.SugaredLogger {
	args := m.Called()
	return args.Get(0).(*zap.SugaredLogger)
}

func TestLoggerInterface(t *testing.T) {
	// This test verifies that MockLogger implements the Logger interface
	var _ contract.Logger = (*MockLogger)(nil)

	// Create a mock logger
	logger := new(MockLogger)

	// Set up expectations
	logger.On("Debug", "debug message", mock.Anything).Return()
	logger.On("Info", "info message", mock.Anything).Return()
	logger.On("Warn", "warn message", mock.Anything).Return()
	logger.On("Error", "error message", mock.Anything).Return()
	logger.On("Fatal", "fatal message", mock.Anything).Return()
	logger.On("Debugf", "debug %s", mock.Anything).Return()
	logger.On("Infof", "info %s", mock.Anything).Return()
	logger.On("Warnf", "warn %s", mock.Anything).Return()
	logger.On("Errorf", "error %s", mock.Anything).Return()
	logger.On("Fatalf", "fatal %s", mock.Anything).Return()
	logger.On("Debugw", "debug message", mock.Anything).Return()
	logger.On("Infow", "info message", mock.Anything).Return()
	logger.On("Warnw", "warn message", mock.Anything).Return()
	logger.On("Errorw", "error message", mock.Anything).Return()
	logger.On("Fatalw", "fatal message", mock.Anything).Return()
	logger.On("With", mock.Anything).Return(logger)
	logger.On("WithContext", mock.Anything).Return(logger)
	logger.On("WithError", mock.Anything).Return(logger)
	logger.On("GetLogger").Return(&zap.SugaredLogger{})

	// Test the methods
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")
	logger.Fatal("fatal message")
	logger.Debugf("debug %s", "test")
	logger.Infof("info %s", "test")
	logger.Warnf("warn %s", "test")
	logger.Errorf("error %s", "test")
	logger.Fatalf("fatal %s", "test")
	logger.Debugw("debug message", "key", "value")
	logger.Infow("info message", "key", "value")
	logger.Warnw("warn message", "key", "value")
	logger.Errorw("error message", "key", "value")
	logger.Fatalw("fatal message", "key", "value")

	withLogger := logger.With("key", "value")
	assert.Equal(t, logger, withLogger)

	withContextLogger := logger.WithContext(context.Background())
	assert.Equal(t, logger, withContextLogger)

	withErrorLogger := logger.WithError(assert.AnError)
	assert.Equal(t, logger, withErrorLogger)

	sugaredLogger := logger.GetLogger()
	assert.NotNil(t, sugaredLogger)

	// Verify expectations
	logger.AssertExpectations(t)
}

// MockField is a mock implementation of the Field interface
type MockField struct {
	mock.Mock
}

func (m *MockField) Key() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockField) Value() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockField) Type() string {
	args := m.Called()
	return args.String(0)
}

func TestFieldInterface(t *testing.T) {
	// This test verifies that MockField implements the Field interface
	var _ contract.Field = (*MockField)(nil)

	// Create a mock field
	field := new(MockField)

	// Set up expectations
	field.On("Key").Return("test_key")
	field.On("Value").Return("test_value")
	field.On("Type").Return("string")

	// Test the methods
	assert.Equal(t, "test_key", field.Key())
	assert.Equal(t, "test_value", field.Value())
	assert.Equal(t, "string", field.Type())

	// Verify expectations
	field.AssertExpectations(t)
}

// MockLoggerConfig is a mock implementation of the LoggerConfig interface
type MockLoggerConfig struct {
	mock.Mock
}

func (m *MockLoggerConfig) Level() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockLoggerConfig) Format() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockLoggerConfig) Output() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockLoggerConfig) EnableCaller() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockLoggerConfig) EnableStacktrace() bool {
	args := m.Called()
	return args.Bool(0)
}

func TestLoggerConfigInterface(t *testing.T) {
	// This test verifies that MockLoggerConfig implements the LoggerConfig interface
	var _ contract.LoggerConfig = (*MockLoggerConfig)(nil)

	// Create a mock logger config
	config := new(MockLoggerConfig)

	// Set up expectations
	config.On("Level").Return("info")
	config.On("Format").Return("json")
	config.On("Output").Return([]string{"stdout", "stderr"})
	config.On("EnableCaller").Return(true)
	config.On("EnableStacktrace").Return(true)

	// Test the methods
	assert.Equal(t, "info", config.Level())
	assert.Equal(t, "json", config.Format())
	assert.Equal(t, []string{"stdout", "stderr"}, config.Output())
	assert.True(t, config.EnableCaller())
	assert.True(t, config.EnableStacktrace())

	// Verify expectations
	config.AssertExpectations(t)
}
