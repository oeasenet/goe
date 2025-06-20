package db_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TestModel is a simple model for testing auto migration
type TestModel struct {
	gorm.Model
	Name string `gorm:"size:255"`
	Age  int
}

// MockConfig is a mock implementation of the Config interface
type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) Get(key string) any {
	args := m.Called(key)
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

// setupTestConfig creates a mock config with SQLite settings
func setupTestConfig() *MockConfig {
	config := new(MockConfig)

	// Default connection settings
	config.On("GetString", "DB_CONNECTION").Return("default")
	config.On("GetString", "DB_DRIVER").Return("sqlite")
	config.On("GetString", "DB_DATABASE").Return(":memory:")
	config.On("GetString", "DB_DSN").Return("")
	config.On("GetString", "DB_HOST").Return("")
	config.On("GetString", "DB_PORT").Return("")
	config.On("GetString", "DB_USERNAME").Return("")
	config.On("GetString", "DB_PASSWORD").Return("")
	config.On("GetString", "DB_CHARSET").Return("")
	config.On("GetString", "DB_TIMEZONE").Return("")
	config.On("GetString", "DB_CONNECTIONS").Return("")

	// Connection pool settings
	config.On("GetInt", "DB_MAX_IDLE_CONNS").Return(10)
	config.On("GetInt", "DB_MAX_OPEN_CONNS").Return(100)
	config.On("GetDuration", "DB_CONN_MAX_LIFETIME").Return(time.Duration(0))
	config.On("GetDuration", "DB_CONN_MAX_IDLE_TIME").Return(time.Duration(0))

	// GORM settings
	config.On("GetBool", "DB_LOG_MODE").Return(false)
	config.On("GetBool", "DB_IGNORE_RECORD_NOT_FOUND_ERROR").Return(false)
	config.On("GetBool", "DB_DISABLE_FOREIGN_KEY_CONSTRAINT_WHEN_MIGRATING").Return(false)
	config.On("GetBool", "DB_AUTO_MIGRATE").Return(true)
	config.On("GetBool", "DB_AUTO_MIGRATE_ANY").Return(false)

	// For Has method
	config.On("Has", mock.Anything).Return(false)

	return config
}

// setupTestLogger creates a mock logger
func setupTestLogger() *MockLogger {
	logger := new(MockLogger)

	// Set up expectations for common logger methods
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()
	logger.On("GetLogger").Return(&zap.SugaredLogger{})

	return logger
}

// TestDatabaseModule_New tests the creation of a new DatabaseModule
func TestDatabaseModule_New(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := db.NewDBModule(config, logger)

	assert.NotNil(t, dbModule)
	assert.Equal(t, "db", dbModule.Name())
	assert.NotNil(t, dbModule.Provide())
}

// TestDatabaseModule_OnStart tests the OnStart method
func TestDatabaseModule_OnStart(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := db.NewDBModule(config, logger)

	err := dbModule.OnStart(context.Background())
	assert.Nil(t, err)

	// Verify that the default connection was established
	instance := dbModule.Instance()
	assert.NotNil(t, instance)
}

// TestDatabaseModule_OnStop tests the OnStop method
func TestDatabaseModule_OnStop(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := db.NewDBModule(config, logger)

	// Start the module first
	err := dbModule.OnStart(context.Background())
	assert.Nil(t, err)

	// Then stop it
	err = dbModule.OnStop(context.Background())
	assert.Nil(t, err)
}

// TestDatabaseModule_Instance tests the Instance method
func TestDatabaseModule_Instance(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := db.NewDBModule(config, logger)

	// Start the module first
	err := dbModule.OnStart(context.Background())
	assert.Nil(t, err)

	// Get the default instance
	instance := dbModule.Instance()
	assert.NotNil(t, instance)

	// Test that the instance is a valid GORM DB
	var count int64
	result := instance.Raw("SELECT 1").Count(&count)
	assert.Nil(t, result.Error)
	assert.Equal(t, int64(1), count)
}

// TestDatabaseModule_Connection tests the Connection method
func TestDatabaseModule_Connection(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := db.NewDBModule(config, logger)

	// Start the module first
	err := dbModule.OnStart(context.Background())
	assert.Nil(t, err)

	// Get the default connection
	conn, err := dbModule.Connection("default")
	assert.Nil(t, err)
	assert.NotNil(t, conn)

	// Test that the connection is a valid GORM DB
	var count int64
	result := conn.Raw("SELECT 1").Count(&count)
	assert.Nil(t, result.Error)
	assert.Equal(t, int64(1), count)

	// Test getting a non-existent connection
	conn, err = dbModule.Connection("non_existent")
	assert.Error(t, err)
	assert.Nil(t, conn)
}

// TestDatabaseModule_AutoMigrate tests the AutoMigrate method
func TestDatabaseModule_AutoMigrate(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := db.NewDBModule(config, logger)

	// Start the module first
	err := dbModule.OnStart(context.Background())
	assert.Nil(t, err)

	// Auto migrate the test model
	err = dbModule.AutoMigrate(&TestModel{})
	assert.Nil(t, err)

	// Verify that the table was created
	instance := dbModule.Instance()
	assert.NotNil(t, instance)

	// Check if the table exists
	var count int64
	result := instance.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='test_models'").Count(&count)
	assert.Nil(t, result.Error)
	assert.Equal(t, int64(1), count)
}

// TestDatabaseModule_AutoMigrateOnConnection tests the AutoMigrateOnConnection method
func TestDatabaseModule_AutoMigrateOnConnection(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := db.NewDBModule(config, logger)

	// Start the module first
	err := dbModule.OnStart(context.Background())
	assert.Nil(t, err)

	// Auto migrate the test model on the default connection
	err = dbModule.AutoMigrateOnConnection("default", &TestModel{})
	assert.Nil(t, err)

	// Verify that the table was created
	conn, err := dbModule.Connection("default")
	assert.Nil(t, err)
	assert.NotNil(t, conn)

	// Check if the table exists
	var count int64
	result := conn.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='test_models'").Count(&count)
	assert.Nil(t, result.Error)
	assert.Equal(t, int64(1), count)

	// Test auto migrating on a non-existent connection
	err = dbModule.AutoMigrateOnConnection("non_existent", &TestModel{})
	assert.Error(t, err)
}

// TestDatabaseModule_CRUD tests basic CRUD operations with the database
func TestDatabaseModule_CRUD(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := db.NewDBModule(config, logger)

	// Start the module first
	err := dbModule.OnStart(context.Background())
	assert.Nil(t, err)

	// Auto migrate the test model
	err = dbModule.AutoMigrate(&TestModel{})
	assert.Nil(t, err)

	// Get the database instance
	instance := dbModule.Instance()
	assert.NotNil(t, instance)

	// Create a test model
	testModel := &TestModel{
		Name: "Test User",
		Age:  30,
	}

	// Create
	result := instance.Create(testModel)
	assert.Nil(t, result.Error)
	assert.NotEqual(t, uint(0), testModel.ID)

	// Read
	var readModel TestModel
	result = instance.First(&readModel, testModel.ID)
	assert.Nil(t, result.Error)
	assert.Equal(t, testModel.Name, readModel.Name)
	assert.Equal(t, testModel.Age, readModel.Age)

	// Update
	readModel.Name = "Updated Name"
	result = instance.Save(&readModel)
	assert.Nil(t, result.Error)

	// Verify update
	var updatedModel TestModel
	result = instance.First(&updatedModel, testModel.ID)
	assert.Nil(t, result.Error)
	assert.Equal(t, "Updated Name", updatedModel.Name)

	// Delete
	result = instance.Delete(&updatedModel)
	assert.Nil(t, result.Error)

	// Verify delete
	var deletedModel TestModel
	result = instance.First(&deletedModel, testModel.ID)
	assert.Error(t, result.Error)
	assert.True(t, result.Error == gorm.ErrRecordNotFound)
}

// TestDatabaseModule_MultipleConnections tests using multiple database connections
func TestDatabaseModule_MultipleConnections(t *testing.T) {
	// Create a temporary file for the second SQLite database
	tempFile, err := os.CreateTemp("", "test_db_*.sqlite")
	require.Nil(t, err)
	tempFilePath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempFilePath)

	// Create a config with multiple connections
	config := new(MockConfig)

	// Default connection settings
	config.On("GetString", "DB_CONNECTION").Return("default")
	config.On("GetString", "DB_DRIVER").Return("sqlite")
	config.On("GetString", "DB_DATABASE").Return(":memory:")
	config.On("GetString", "DB_DSN").Return("")

	// Second connection settings
	config.On("GetString", "DB_CONNECTIONS").Return("second")
	config.On("GetString", "DB_SECOND_DRIVER").Return("sqlite")
	config.On("GetString", "DB_SECOND_DATABASE").Return(tempFilePath)
	config.On("GetString", "DB_SECOND_DSN").Return("")

	// Common settings for both connections
	config.On("GetString", mock.MatchedBy(func(key string) bool {
		return key != "DB_CONNECTION" &&
			key != "DB_DRIVER" &&
			key != "DB_DATABASE" &&
			key != "DB_DSN" &&
			key != "DB_CONNECTIONS" &&
			key != "DB_SECOND_DRIVER" &&
			key != "DB_SECOND_DATABASE" &&
			key != "DB_SECOND_DSN"
	})).Return("")

	config.On("GetInt", mock.Anything).Return(0)
	config.On("GetBool", mock.Anything).Return(false)
	config.On("GetDuration", mock.Anything).Return(time.Duration(0))
	config.On("Has", mock.Anything).Return(false)

	logger := setupTestLogger()

	dbModule := db.NewDBModule(config, logger)

	// Start the module
	err = dbModule.OnStart(context.Background())
	assert.Nil(t, err)

	// Get the default connection
	defaultConn := dbModule.Instance()
	assert.NotNil(t, defaultConn)

	// Get the second connection
	secondConn, err := dbModule.Connection("second")
	assert.Nil(t, err)
	assert.NotNil(t, secondConn)

	// Test that they are different connections
	// Migrate TestModel on default connection
	err = dbModule.AutoMigrateOnConnection("default", &TestModel{})
	assert.Nil(t, err)

	// Create a different model for the second connection
	type SecondModel struct {
		gorm.Model
		Title string
	}

	// Migrate SecondModel on second connection
	err = secondConn.AutoMigrate(&SecondModel{})
	assert.Nil(t, err)

	// Verify that TestModel exists only on default connection
	var count int64
	result := defaultConn.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='test_models'").Count(&count)
	assert.Nil(t, result.Error)
	assert.Equal(t, int64(1), count)

	result = secondConn.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='test_models'").Count(&count)
	assert.Nil(t, result.Error)
	assert.Equal(t, int64(0), count)

	// Verify that SecondModel exists only on second connection
	result = secondConn.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='second_models'").Count(&count)
	assert.Nil(t, result.Error)
	assert.Equal(t, int64(1), count)

	result = defaultConn.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='second_models'").Count(&count)
	assert.Nil(t, result.Error)
	assert.Equal(t, int64(0), count)
}

// TestDatabaseModule_Transactions tests transaction support
func TestDatabaseModule_Transactions(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := db.NewDBModule(config, logger)

	// Start the module first
	err := dbModule.OnStart(context.Background())
	assert.Nil(t, err)

	// Auto migrate the test model
	err = dbModule.AutoMigrate(&TestModel{})
	assert.Nil(t, err)

	// Get the database instance
	instance := dbModule.Instance()
	assert.NotNil(t, instance)

	// Test successful transaction
	err = instance.Transaction(func(tx *gorm.DB) error {
		// Create a test model within transaction
		testModel := &TestModel{
			Name: "Transaction User",
			Age:  25,
		}

		if err := tx.Create(testModel).Error; err != nil {
			return err
		}

		// Create another test model within the same transaction
		anotherModel := &TestModel{
			Name: "Another User",
			Age:  35,
		}

		if err := tx.Create(anotherModel).Error; err != nil {
			return err
		}

		return nil
	})

	assert.Nil(t, err)

	// Verify both models were created
	var count int64
	result := instance.Model(&TestModel{}).Count(&count)
	assert.Nil(t, result.Error)
	assert.Equal(t, int64(2), count)

	// Test failed transaction
	err = instance.Transaction(func(tx *gorm.DB) error {
		// Create a test model within transaction
		testModel := &TestModel{
			Name: "Will Rollback",
			Age:  40,
		}

		if err := tx.Create(testModel).Error; err != nil {
			return err
		}

		// Return an error to rollback the transaction
		return assert.AnError
	})

	assert.Error(t, err)

	// Verify the model was not created (count should still be 2)
	result = instance.Model(&TestModel{}).Count(&count)
	assert.Nil(t, result.Error)
	assert.Equal(t, int64(2), count)
}

// TestDatabaseModule_QueryBuilding tests query building capabilities
func TestDatabaseModule_QueryBuilding(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := db.NewDBModule(config, logger)

	// Start the module first
	err := dbModule.OnStart(context.Background())
	assert.Nil(t, err)

	// Auto migrate the test model
	err = dbModule.AutoMigrate(&TestModel{})
	assert.Nil(t, err)

	// Get the database instance
	instance := dbModule.Instance()
	assert.NotNil(t, instance)

	// Create test data
	testModels := []TestModel{
		{Name: "User 1", Age: 20},
		{Name: "User 2", Age: 25},
		{Name: "User 3", Age: 30},
		{Name: "User 4", Age: 35},
		{Name: "User 5", Age: 40},
	}

	result := instance.Create(&testModels)
	assert.Nil(t, result.Error)

	// Test WHERE clause
	var whereModels []TestModel
	result = instance.Where("age > ?", 30).Find(&whereModels)
	assert.Nil(t, result.Error)
	assert.Equal(t, 2, len(whereModels))

	// Test ORDER BY
	var orderedModels []TestModel
	result = instance.Order("age DESC").Find(&orderedModels)
	assert.Nil(t, result.Error)
	assert.Equal(t, 5, len(orderedModels))
	assert.Equal(t, "User 5", orderedModels[0].Name)
	assert.Equal(t, "User 1", orderedModels[4].Name)

	// Test LIMIT and OFFSET
	var limitModels []TestModel
	result = instance.Limit(2).Offset(1).Order("age ASC").Find(&limitModels)
	assert.Nil(t, result.Error)
	assert.Equal(t, 2, len(limitModels))
	assert.Equal(t, "User 2", limitModels[0].Name)
	assert.Equal(t, "User 3", limitModels[1].Name)

	// Test SELECT specific columns
	var selectedModels []TestModel
	result = instance.Select("name").Where("age > ?", 30).Find(&selectedModels)
	assert.Nil(t, result.Error)
	assert.Equal(t, 2, len(selectedModels))
	assert.NotEmpty(t, selectedModels[0].Name)
	assert.Equal(t, 0, selectedModels[0].Age) // Age should be zero as it wasn't selected

	// Test COUNT
	var count int64
	result = instance.Model(&TestModel{}).Where("age < ?", 30).Count(&count)
	assert.Nil(t, result.Error)
	assert.Equal(t, int64(2), count)
}

// TestDatabaseModule_RegisterModelsForMigration tests the new model registration functionality
func TestDatabaseModule_RegisterModelsForMigration(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	// Enable auto-migration
	config.On("GetBool", "DB_AUTO_MIGRATE_ANY").Return(false)
	config.On("GetBool", "DB_AUTO_MIGRATE").Return(true)

	dbModule := db.NewDBModule(config, logger)

	// Register models BEFORE starting the module
	dbModule.RegisterModelsForMigration(&TestModel{})

	// Start the module - this should automatically migrate registered models
	err := dbModule.OnStart(context.Background())
	assert.Nil(t, err)

	// Verify that the table was created automatically
	instance := dbModule.Instance()
	assert.NotNil(t, instance)

	// Check if the table exists
	var count int64
	result := instance.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='test_models'").Count(&count)
	assert.Nil(t, result.Error)
	assert.Equal(t, int64(1), count)

	// Test that we can use the migrated table
	testModel := &TestModel{Name: "Auto Migrated", Age: 25}
	result = instance.Create(testModel)
	assert.Nil(t, result.Error)
	assert.NotEqual(t, uint(0), testModel.ID)
}

// TestDatabaseModule_RegisterModelsForMigrationOnConnection tests registration for specific connections
func TestDatabaseModule_RegisterModelsForMigrationOnConnection(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	// Enable auto-migration for default connection
	config.On("GetBool", "DB_AUTO_MIGRATE_ANY").Return(false)
	config.On("GetBool", "DB_AUTO_MIGRATE").Return(true)

	dbModule := db.NewDBModule(config, logger)

	// Register models for specific connection BEFORE starting the module
	dbModule.RegisterModelsForMigrationOnConnection("default", &TestModel{})

	// Start the module - this should automatically migrate registered models
	err := dbModule.OnStart(context.Background())
	assert.Nil(t, err)

	// Verify that the table was created automatically on the specified connection
	conn, err := dbModule.Connection("default")
	assert.Nil(t, err)
	assert.NotNil(t, conn)

	// Check if the table exists
	var count int64
	result := conn.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='test_models'").Count(&count)
	assert.Nil(t, result.Error)
	assert.Equal(t, int64(1), count)
}

// TestDatabaseModule_RegisterModelsWithoutAutoMigration tests that registration works but doesn't migrate when auto-migration is disabled
func TestDatabaseModule_RegisterModelsWithoutAutoMigration(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	// Override the default auto-migration setting to disable it
	config.ExpectedCalls = nil // Clear existing expectations
	config.On("GetString", "DB_CONNECTION").Return("default")
	config.On("GetString", "DB_DRIVER").Return("sqlite")
	config.On("GetString", "DB_DATABASE").Return(":memory:")
	config.On("GetString", "DB_DSN").Return("")
	config.On("GetString", "DB_HOST").Return("")
	config.On("GetString", "DB_PORT").Return("")
	config.On("GetString", "DB_USERNAME").Return("")
	config.On("GetString", "DB_PASSWORD").Return("")
	config.On("GetString", "DB_CHARSET").Return("")
	config.On("GetString", "DB_TIMEZONE").Return("")
	config.On("GetString", "DB_CONNECTIONS").Return("")
	config.On("GetInt", "DB_MAX_IDLE_CONNS").Return(10)
	config.On("GetInt", "DB_MAX_OPEN_CONNS").Return(100)
	config.On("GetDuration", "DB_CONN_MAX_LIFETIME").Return(time.Duration(0))
	config.On("GetDuration", "DB_CONN_MAX_IDLE_TIME").Return(time.Duration(0))
	config.On("GetBool", "DB_LOG_MODE").Return(false)
	config.On("GetBool", "DB_IGNORE_RECORD_NOT_FOUND_ERROR").Return(false)
	config.On("GetBool", "DB_DISABLE_FOREIGN_KEY_CONSTRAINT_WHEN_MIGRATING").Return(false)
	config.On("Has", mock.Anything).Return(false)

	// Disable auto-migration
	config.On("GetBool", "DB_AUTO_MIGRATE_ANY").Return(false)
	config.On("GetBool", "DB_AUTO_MIGRATE").Return(false)

	dbModule := db.NewDBModule(config, logger)

	// Register models BEFORE starting the module
	dbModule.RegisterModelsForMigration(&TestModel{})

	// Start the module - this should NOT automatically migrate registered models
	err := dbModule.OnStart(context.Background())
	assert.Nil(t, err)

	// Verify that the table was NOT created automatically
	instance := dbModule.Instance()
	assert.NotNil(t, instance)

	// Check if the table exists - it should NOT exist
	var count int64
	result := instance.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='test_models'").Count(&count)
	assert.Nil(t, result.Error)
	assert.Equal(t, int64(0), count)

	// But we can still manually migrate the registered models
	err = dbModule.AutoMigrate(&TestModel{})
	assert.Nil(t, err)

	// Now the table should exist
	result = instance.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='test_models'").Count(&count)
	assert.Nil(t, result.Error)
	assert.Equal(t, int64(1), count)
}

// TestDatabaseModule_MultipleModelRegistration tests registering multiple models
func TestDatabaseModule_MultipleModelRegistration(t *testing.T) {
	// Define a second test model
	type SecondTestModel struct {
		gorm.Model
		Title       string `gorm:"size:100"`
		Description string `gorm:"size:500"`
	}

	config := setupTestConfig()
	logger := setupTestLogger()

	// Enable auto-migration
	config.On("GetBool", "DB_AUTO_MIGRATE_ANY").Return(false)
	config.On("GetBool", "DB_AUTO_MIGRATE").Return(true)

	dbModule := db.NewDBModule(config, logger)

	// Register multiple models BEFORE starting the module
	dbModule.RegisterModelsForMigration(&TestModel{}, &SecondTestModel{})

	// Start the module - this should automatically migrate all registered models
	err := dbModule.OnStart(context.Background())
	assert.Nil(t, err)

	// Verify that both tables were created automatically
	instance := dbModule.Instance()
	assert.NotNil(t, instance)

	// Check if the first table exists
	var count1 int64
	result := instance.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='test_models'").Count(&count1)
	assert.Nil(t, result.Error)
	assert.Equal(t, int64(1), count1)

	// Check if the second table exists
	var count2 int64
	result = instance.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='second_test_models'").Count(&count2)
	assert.Nil(t, result.Error)
	assert.Equal(t, int64(1), count2)
}
