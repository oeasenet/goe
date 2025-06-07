package contract_test

import (
	"context"
	"errors"
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

func (m *MockLogger) Debug(msg string, fields ...contract.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...contract.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...contract.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...contract.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields ...contract.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) With(fields ...contract.Field) contract.Logger {
	args := m.Called(fields)
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

// MockField is a mock implementation of the Field interface
type MockField struct {
	mock.Mock
}

func (m *MockField) Key() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockField) Value() any {
	args := m.Called()
	return args.Get(0)
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

func TestLoggerInterface(t *testing.T) {
	// This test verifies that MockLogger implements the Logger interface
	var _ contract.Logger = (*MockLogger)(nil)

	// Create a mock logger
	logger := new(MockLogger)

	// Create a mock field
	field := new(MockField)
	field.On("Key").Return("test_key")
	field.On("Value").Return("test_value")

	// Set up expectations
	logger.On("Debug", "debug message", mock.Anything).Return()
	logger.On("Info", "info message", mock.Anything).Return()
	logger.On("Warn", "warn message", mock.Anything).Return()
	logger.On("Error", "error message", mock.Anything).Return()
	logger.On("Fatal", "fatal message", mock.Anything).Return()
	logger.On("With", mock.Anything).Return(logger)
	logger.On("WithContext", mock.Anything).Return(logger)
	logger.On("WithError", mock.Anything).Return(logger)
	logger.On("GetLogger").Return(&zap.SugaredLogger{})

	// Test the methods
	logger.Debug("debug message", field)
	logger.Info("info message", field)
	logger.Warn("warn message", field)
	logger.Error("error message", field)
	logger.Fatal("fatal message", field)

	assert.Equal(t, logger, logger.With(field))
	assert.Equal(t, logger, logger.WithContext(context.Background()))
	assert.Equal(t, logger, logger.WithError(errors.New("test error")))
	assert.NotNil(t, logger.GetLogger())

	// Call the Key() and Value() methods explicitly to satisfy the expectations
	assert.Equal(t, "test_key", field.Key())
	assert.Equal(t, "test_value", field.Value())

	// Verify expectations
	logger.AssertExpectations(t)
	field.AssertExpectations(t)
}

func TestFieldInterface(t *testing.T) {
	// This test verifies that MockField implements the Field interface
	var _ contract.Field = (*MockField)(nil)

	// Create a mock field
	field := new(MockField)

	// Set up expectations
	field.On("Key").Return("test_key")
	field.On("Value").Return("test_value")

	// Test the methods
	assert.Equal(t, "test_key", field.Key())
	assert.Equal(t, "test_value", field.Value())

	// Verify expectations
	field.AssertExpectations(t)
}

func TestLoggerConfigInterface(t *testing.T) {
	// This test verifies that MockLoggerConfig implements the LoggerConfig interface
	var _ contract.LoggerConfig = (*MockLoggerConfig)(nil)

	// Create a mock logger config
	config := new(MockLoggerConfig)

	// Set up expectations
	config.On("Level").Return("debug")
	config.On("Format").Return("json")
	config.On("Output").Return([]string{"console", "file"})
	config.On("EnableCaller").Return(true)
	config.On("EnableStacktrace").Return(true)

	// Test the methods
	assert.Equal(t, "debug", config.Level())
	assert.Equal(t, "json", config.Format())
	assert.Equal(t, []string{"console", "file"}, config.Output())
	assert.True(t, config.EnableCaller())
	assert.True(t, config.EnableStacktrace())

	// Verify expectations
	config.AssertExpectations(t)
}
