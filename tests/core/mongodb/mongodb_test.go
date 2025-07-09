package mongodb

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/mongodb"
	"go.uber.org/zap"
	"testing"
	"time"
)

type TestModel struct {
	Name string
	Age  int
}

func (*TestModel) colName() string {
	return "test_models"
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
	config.On("GetString", "MONGO_DB_CONNECTION").Return("")
	config.On("GetString", "MONGO_DB_CONNECTIONS").Return("")
	config.On("GetString", "MONGO_DB_URI").Return("mongodb://localhost:27017/")
	config.On("GetString", "MONGO_DB_DB_NAME").Return("goe_test")
	config.On("GetString", "MONGO_DB_MIN_POOL_SIZE").Return("")
	config.On("GetString", "MONGO_DB_MAX_POOL_SIZE").Return("")
	config.On("GetString", "MONGO_DB_MAX_CONN_IDLE_TIME").Return("")

	// For Has method
	config.On("Has", mock.Anything).Return(false)

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

	dbModule := mongodb.NewDBModule(config, logger)

	assert.NotNil(t, dbModule)
	assert.Equal(t, "mongo_db", dbModule.Name())
	assert.NotNil(t, dbModule.Provide())
}

// TestDatabaseModule_OnStart tests the OnStart method
func TestDatabaseModule_OnStart(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := mongodb.NewDBModule(config, logger)

	err := dbModule.OnStart(dbModule.Ctx())
	assert.Nil(t, err)

	// Verify that the default connection was established
	instance := dbModule.Instance()
	assert.NotNil(t, instance)
}

// TestDatabaseModule_OnStop tests the OnStop method
func TestDatabaseModule_OnStop(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := mongodb.NewDBModule(config, logger)

	// Start the module first
	err := dbModule.OnStart(dbModule.Ctx())
	assert.Nil(t, err)

	// Then stop it
	err = dbModule.OnStop(dbModule.Ctx())
	assert.Nil(t, err)
}

// TestDatabaseModule_Instance tests the Instance method
func TestDatabaseModule_Instance(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := mongodb.NewDBModule(config, logger)

	// Start the module first
	err := dbModule.OnStart(dbModule.Ctx())
	assert.Nil(t, err)

	// Get the default instance
	instance := dbModule.Instance()
	assert.NotNil(t, instance)

	// Test that the instance is a valid MONGO DB
	ctx, cancel := context.WithTimeout(dbModule.Ctx(), 2*time.Second)
	defer cancel()

	err = instance.Client().Ping(ctx, nil)
	assert.Nil(t, err)
}

// TestDatabaseModule_SetMonitor tests the SetMonitor method
func TestDatabaseModule_SetMonitor(t *testing.T) {
	var startedCalled, succeededCalled, failedCalled bool
	// Mock monitor
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			startedCalled = true
		},
		Succeeded: func(ctx context.Context, evt *event.CommandSucceededEvent) {
			succeededCalled = true
		},
		Failed: func(ctx context.Context, evt *event.CommandFailedEvent) {
			failedCalled = true
		},
	}

	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := mongodb.NewDBModule(config, logger)
	dbModule.SetMonitor(monitor)

	// Create a test model
	testModel := &TestModel{
		Name: "Test User",
		Age:  30,
	}

	// Start the module first
	err := dbModule.OnStart(dbModule.Ctx())
	assert.Nil(t, err)

	Conn := dbModule.Instance()

	// drop collection
	defer Conn.Collection(testModel.colName()).Drop(dbModule.Ctx())

	// Insert
	_, err = Conn.Collection(testModel.colName()).InsertOne(dbModule.Ctx(), testModel)
	assert.Nil(t, err)

	// Assertions
	assert.True(t, startedCalled)
	assert.True(t, succeededCalled)
	assert.False(t, failedCalled)
}

// TestDatabaseModule_CRUD tests basic CRUD operations with the database
func TestDatabaseModule_CRUD(t *testing.T) {
	config := setupTestConfig()
	logger := setupTestLogger()

	dbModule := mongodb.NewDBModule(config, logger)

	// Start the module first
	err := dbModule.OnStart(dbModule.Ctx())
	assert.Nil(t, err)

	// Get the database instance
	instance := dbModule.Instance()
	assert.NotNil(t, instance)

	// Create a test model
	testModel := &TestModel{
		Name: "Test User",
		Age:  30,
	}

	// drop collection
	defer instance.Collection(testModel.colName()).Drop(dbModule.Ctx())

	// Insert
	insertResult, err := instance.Collection(testModel.colName()).InsertOne(dbModule.Ctx(), testModel)
	assert.Nil(t, err)
	objectID := insertResult.InsertedID.(bson.ObjectID)
	assert.NotEqual(t, bson.NilObjectID, objectID)

	// Find
	var readModel TestModel
	err = instance.Collection(testModel.colName()).FindOne(dbModule.Ctx(), bson.M{}).Decode(&readModel)
	assert.Nil(t, err)
	assert.Equal(t, testModel.Name, readModel.Name)
	assert.Equal(t, testModel.Age, readModel.Age)

	// Update
	readModel.Name = "Updated Name"
	updateResult, err := instance.Collection(testModel.colName()).UpdateOne(dbModule.Ctx(), bson.M{}, bson.M{"$set": &readModel})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), updateResult.ModifiedCount)

	// Verify update
	var updatedModel TestModel
	err = instance.Collection(testModel.colName()).FindOne(dbModule.Ctx(), bson.M{}).Decode(&updatedModel)
	assert.Nil(t, err)
	assert.Equal(t, "Updated Name", updatedModel.Name)

	// Delete
	deleteResult, err := instance.Collection(testModel.colName()).DeleteOne(dbModule.Ctx(), bson.M{"_id": objectID})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), deleteResult.DeletedCount)

	// Verify delete
	findResult := instance.Collection(testModel.colName()).FindOne(dbModule.Ctx(), bson.M{})
	assert.Error(t, findResult.Err())
	assert.True(t, dbModule.IsNoDocumentsError(findResult.Err()))
}

// TestDatabaseModule_MultipleConnections tests using multiple database connections
func TestDatabaseModule_MultipleConnections(t *testing.T) {
	// Create a config with multiple connections
	config := new(MockConfig)

	// Default connection settings
	config.On("GetString", "MONGO_DB_CONNECTION").Return("")
	config.On("GetString", "MONGO_DB_URI").Return("mongodb://localhost:27017/")
	config.On("GetString", "MONGO_DB_DB_NAME").Return("goe_test")
	config.On("GetString", "MONGO_DB_MIN_POOL_SIZE").Return("")
	config.On("GetString", "MONGO_DB_MAX_POOL_SIZE").Return("")
	config.On("GetString", "MONGO_DB_MAX_CONN_IDLE_TIME").Return("")

	//Second connection settings
	config.On("GetString", "MONGO_DB_CONNECTIONS").Return("second")
	config.On("GetString", "MONGO_DB_SECOND_URI").Return("MONGO_DB_SECOND_URI")
	config.On("GetString", "MONGO_DB_SECOND_DB_NAME").Return("goe_test")
	config.On("GetString", "MONGO_DB_SECOND_MIN_POOL_SIZE").Return()
	config.On("GetString", "MONGO_DB_SECOND_MAX_POOL_SIZE").Return()
	config.On("GetString", "MONGO_DB_SECOND_MAX_CONN_IDLE_TIME").Return()

	// For Has method
	config.On("Has", mock.Anything).Return(false)

	logger := setupTestLogger()

	dbModule := mongodb.NewDBModule(config, logger)

	// Start the module
	err := dbModule.OnStart(context.Background())
	assert.Nil(t, err)

	// Get the default connection
	defaultConn := dbModule.Instance()
	assert.NotNil(t, defaultConn)

	// Get the second connection
	secondConn, err := dbModule.Connection("second")
	assert.Nil(t, err)
	assert.NotNil(t, secondConn)

	// Test that they are different connections
	// Create a different model for the second connection
	type SecondModel struct {
		Title string
	}
	secondModel := &SecondModel{
		Title: "test_title",
	}
	secondModelColName := "test_models"

	testModel := &TestModel{
		Name: "Student",
		Age:  18,
	}

	_, err = defaultConn.Collection(testModel.colName()).InsertOne(dbModule.Ctx(), testModel)
	assert.Nil(t, err)
	_, err = secondConn.Collection(secondModelColName).InsertOne(dbModule.Ctx(), secondModel)
	assert.Nil(t, err)

	defer defaultConn.Collection(testModel.colName()).Drop(dbModule.Ctx())
	defer secondConn.Collection(secondModelColName).Drop(dbModule.Ctx())

	// Verify that TestModel exists only on default connection
	var count int64
	count, err = defaultConn.Collection(testModel.colName()).CountDocuments(dbModule.Ctx(), bson.M{})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), count)

	// Verify that SecondModel exists only on second connection
	count, err = secondConn.Collection(secondModelColName).CountDocuments(dbModule.Ctx(), bson.M{})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), count)
}
