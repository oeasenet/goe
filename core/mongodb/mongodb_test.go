package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/zap"
)

// Removed unused TestModel - not needed for unit tests

// MockConfig is a mock implementation of the Config interface
type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) Get(key string) any {
	args := m.Called(key)
	if len(args) == 0 {
		return nil
	}
	return args.Get(0)
}

func (m *MockConfig) GetString(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *MockConfig) GetInt(key string) int {
	args := m.Called(key)
	return args.Int(0)
}

func (m *MockConfig) GetInt64(key string) int64 {
	args := m.Called(key)
	return args.Get(0).(int64)
}

func (m *MockConfig) GetFloat64(key string) float64 {
	args := m.Called(key)
	return args.Get(0).(float64)
}

func (m *MockConfig) GetBool(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockConfig) GetDuration(key string) time.Duration {
	args := m.Called(key)
	return args.Get(0).(time.Duration)
}

func (m *MockConfig) GetStringSlice(key string) []string {
	args := m.Called(key)
	return args.Get(0).([]string)
}

func (m *MockConfig) GetStringMap(key string) map[string]any {
	args := m.Called(key)
	return args.Get(0).(map[string]any)
}

func (m *MockConfig) Set(key string, value any) {
	m.Called(key, value)
}

func (m *MockConfig) Has(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockConfig) All() map[string]any {
	args := m.Called()
	return args.Get(0).(map[string]any)
}

func (m *MockConfig) Reload() error {
	args := m.Called()
	return args.Error(0)
}

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
	// Module tagging calls With during construction; return self so the same
	// mock receives subsequent calls.
	return m
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

// setupTestConfig creates a mock config with MongoDB settings
func setupTestConfig() *MockConfig {
	config := new(MockConfig)

	// Default connection settings
	config.On("GetString", "MONGO_CONNECTION").Return("")
	config.On("GetString", "MONGO_CONNECTIONS").Return("")
	config.On("GetString", "MONGO_URI").Return("mongodb://localhost:27017/")
	config.On("GetString", "MONGO_DB_NAME").Return("goe_test")
	config.On("GetString", "MONGO_MIN_POOL_SIZE").Return("")
	config.On("GetString", "MONGO_MAX_POOL_SIZE").Return("")
	config.On("GetString", "MONGO_MAX_CONN_IDLE_TIME").Return("")

	// For Has method - return false for all config keys
	config.On("Has", mock.Anything).Return(false)

	// For Get method - return nil for all config keys
	config.On("Get", mock.Anything).Return(nil)

	return config
}

// setupTestLogger creates a mock logger
func setupTestLogger() *MockLogger {
	logger := new(MockLogger)

	// Set up expectations for common logger methods
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Infof", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()
	logger.On("Warnf", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()
	logger.On("Errorf", mock.Anything, mock.Anything).Return()
	logger.On("GetLogger").Return(&zap.SugaredLogger{})

	return logger
}

// TestDatabaseModule_New tests the creation of a new DatabaseModule
func TestDatabaseModule_New(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := NewDBModule(config, logger)

	assert.NotNil(t, dbModule)
	assert.Equal(t, "mongo_db", dbModule.Name())
	assert.NotNil(t, dbModule.Provide())
}

// TestDatabaseModule_OnStart tests the OnStart method
func TestDatabaseModule_OnStart(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := NewDBModule(config, logger)

	// Test module creation
	assert.NotNil(t, dbModule)
	assert.Equal(t, "mongo_db", dbModule.Name())

	// Since we can't test actual connection without MongoDB,
	// we just verify the module is created properly
}

// TestDatabaseModule_OnStop tests the OnStop method
func TestDatabaseModule_OnStop(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := NewDBModule(config, logger)

	// Test that OnStop can be called without error on unstarted module
	ctx, cancel := DefaultContext()
	defer cancel()
	err := dbModule.OnStop(ctx)
	assert.Nil(t, err)
}

// TestDatabaseModule_Instance tests the DB method
func TestDatabaseModule_Instance(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := NewDBModule(config, logger)

	// Test that DB returns nil before connection is established
	instance := dbModule.DB()
	assert.Nil(t, instance)
}

// TestDatabaseModule_SetMonitor tests the SetMonitor method
func TestDatabaseModule_SetMonitor(t *testing.T) {
	// Mock monitor
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			// Mock function
		},
		Succeeded: func(ctx context.Context, evt *event.CommandSucceededEvent) {
			// Mock function
		},
		Failed: func(ctx context.Context, evt *event.CommandFailedEvent) {
			// Mock function
		},
	}

	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := NewDBModule(config, logger)

	// Test that setMonitor doesn't panic
	dbModule.setMonitor(monitor)

	// Verify module is still functional
	assert.NotNil(t, dbModule)
	assert.Equal(t, "mongo_db", dbModule.Name())
}

// TestDatabaseModule_Connection tests the Connection method
func TestDatabaseModule_Connection(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := NewDBModule(config, logger)

	// Test that Connection returns error for non-existent connection
	_, err := dbModule.Connection("nonexistent")
	assert.Error(t, err)
}

// TestDatabaseModule_ConfigValidation tests configuration validation
func TestDatabaseModule_ConfigValidation(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := NewDBModule(config, logger)

	// Test basic module properties
	assert.NotNil(t, dbModule)
	assert.Equal(t, "mongo_db", dbModule.Name())
	assert.NotNil(t, dbModule.Provide())
}

// TestDatabaseModule_Collection tests the Col method
func TestDatabaseModule_Collection(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := NewDBModule(config, logger)

	// Test that Col returns nil when no connection is established
	col := dbModule.Col("test_collection")
	assert.Nil(t, col)
}

// TestDatabaseModule_CollectionFrom tests the ColFrom method
func TestDatabaseModule_CollectionFrom(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := NewDBModule(config, logger)

	// Test that ColFrom returns error for non-existent connection
	_, err := dbModule.ColFrom("nonexistent", "test_collection")
	assert.Error(t, err)
}

// TestDatabaseModule_Client tests the Client method
func TestDatabaseModule_Client(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := NewDBModule(config, logger)

	// Test that Client returns nil when no connection is established
	client := dbModule.Client()
	assert.Nil(t, client)
}

// TestDatabaseModule_DatabaseFrom is an alias test for Connection method
func TestDatabaseModule_DatabaseFrom(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := NewDBModule(config, logger)

	// Test that Connection returns error for non-existent connection
	_, err := dbModule.Connection("nonexistent")
	assert.Error(t, err)
}
