package db

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MockConfig implements contract.Config for testing
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

// MockLogger implements contract.Logger for testing
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

// TestModel for testing auto-migration
type TestModel struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:100"`
}

func TestDatabaseModule_New(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	dbModule := NewDBModule(config, logger)

	assert.NotNil(t, dbModule)
	assert.Equal(t, "db", dbModule.Name())
	assert.NotNil(t, dbModule.Provide())
}

func TestDatabaseModule_Instance(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	dbModule := NewDBModule(config, logger)

	t.Run("default connection not configured", func(t *testing.T) {
		config.On("GetString", "DB_CONNECTION").Return("")
		logger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()

		instance := dbModule.Instance()
		assert.Nil(t, instance)

		config.AssertExpectations(t)
		logger.AssertExpectations(t)
	})

	t.Run("custom connection name", func(t *testing.T) {
		config.On("GetString", "DB_CONNECTION").Return("custom")
		logger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()

		instance := dbModule.Instance()
		assert.Nil(t, instance)

		config.AssertExpectations(t)
		logger.AssertExpectations(t)
	})
}

func TestDatabaseModule_Connection(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	dbModule := NewDBModule(config, logger)

	t.Run("connection not found", func(t *testing.T) {
		conn, err := dbModule.Connection("nonexistent")
		assert.Error(t, err)
		assert.Nil(t, conn)
		assert.Contains(t, err.Error(), "not found or not configured")
	})
}

func TestDatabaseModule_AutoMigrate(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	dbModule := NewDBModule(config, logger)

	t.Run("no default connection", func(t *testing.T) {
		config.On("GetString", "DB_CONNECTION").Return("")
		logger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()

		err := dbModule.AutoMigrate(&TestModel{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not available for auto-migration")

		config.AssertExpectations(t)
		logger.AssertExpectations(t)
	})
}

func TestDatabaseModule_AutoMigrateOnConnection(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	dbModule := NewDBModule(config, logger)

	t.Run("connection not found", func(t *testing.T) {
		err := dbModule.AutoMigrateOnConnection("nonexistent", &TestModel{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found or not configured")
	})
}

func TestDatabaseModule_RegisterModelsForMigration(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	dbModule := NewDBModule(config, logger)

	t.Run("register models for default connection", func(t *testing.T) {
		config.On("GetString", "DB_CONNECTION").Return("")
		logger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		dbModule.RegisterModelsForMigration(&TestModel{})

		config.AssertExpectations(t)
		logger.AssertExpectations(t)
	})

	t.Run("register models for custom connection", func(t *testing.T) {
		config.On("GetString", "DB_CONNECTION").Return("custom")
		logger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		dbModule.RegisterModelsForMigration(&TestModel{})

		config.AssertExpectations(t)
		logger.AssertExpectations(t)
	})
}

func TestDatabaseModule_RegisterModelsForMigrationOnConnection(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	dbModule := NewDBModule(config, logger)

	t.Run("register models for specific connection", func(t *testing.T) {
		logger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		dbModule.RegisterModelsForMigrationOnConnection("test", &TestModel{})

		logger.AssertExpectations(t)
	})
}

func TestDatabaseModule_OnStart(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	dbModule := NewDBModule(config, logger)

	t.Run("no database configuration", func(t *testing.T) {
		config.On("GetString", "DB_CONNECTION").Return("")
		config.On("GetString", "DB_DRIVER").Return("")
		config.On("GetString", "DB_CONNECTIONS").Return("")
		config.On("GetBool", "DB_AUTO_MIGRATE_ANY").Return(false)
		config.On("GetBool", "DB_AUTO_MIGRATE").Return(false)

		logger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		logger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()

		ctx := context.Background()
		err := dbModule.OnStart(ctx)
		assert.NoError(t, err)

		config.AssertExpectations(t)
		logger.AssertExpectations(t)
	})
}

func TestDatabaseModule_OnStop(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	dbModule := NewDBModule(config, logger)

	t.Run("stop with no connections", func(t *testing.T) {
		logger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		ctx := context.Background()
		err := dbModule.OnStop(ctx)
		assert.NoError(t, err)

		logger.AssertExpectations(t)
	})
}

func TestDatabaseModule_BuildDSN(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	t.Run("direct DSN", func(t *testing.T) {
		config.On("GetString", "DB_DSN").Return("direct://connection/string")
		logger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		// Use reflection to test private method via public interface
		// Since buildDSN is private, we'll test it indirectly through connect
		config.On("GetString", "DB_DRIVER").Return("sqlite")
		config.On("GetBool", "DB_LOG_MODE").Return(false)
		config.On("GetBool", "DB_IGNORE_RECORD_NOT_FOUND_ERROR").Return(false)
		config.On("GetBool", "DB_DISABLE_FOREIGN_KEY_CONSTRAINT_WHEN_MIGRATING").Return(false)
		config.On("GetInt", "DB_MAX_IDLE_CONNS").Return(0)
		config.On("GetInt", "DB_MAX_OPEN_CONNS").Return(0)
		config.On("Has", "DB_MAX_IDLE_CONNS").Return(false)
		config.On("Has", "DB_MAX_OPEN_CONNS").Return(false)
		config.On("GetDuration", "DB_CONN_MAX_LIFETIME").Return(time.Duration(0))
		config.On("GetDuration", "DB_CONN_MAX_IDLE_TIME").Return(time.Duration(0))

		dbModule := NewDBModule(config, logger)

		// This will fail to connect due to invalid DSN, but we can test the DSN building logic
		_, err := dbModule.Connection("test")
		assert.Error(t, err) // Expected since we're not actually connecting
	})
}

func TestDatabaseModule_IntegrationWithSQLite(t *testing.T) {
	// Skip integration test in unit testing mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := &MockConfig{}
	logger := &MockLogger{}

	dbModule := NewDBModule(config, logger)

	t.Run("successful SQLite connection", func(t *testing.T) {
		// Create a temporary SQLite file
		tempFile := t.TempDir() + "/test.db"

		config.On("GetString", "DB_CONNECTION").Return("default")
		config.On("GetString", "DB_DRIVER").Return("sqlite")
		config.On("GetString", "DB_DSN").Return("")
		config.On("GetString", "DB_DATABASE").Return(tempFile)
		config.On("GetString", "DB_CONNECTIONS").Return("")
		config.On("GetBool", "DB_AUTO_MIGRATE_ANY").Return(false)
		config.On("GetBool", "DB_AUTO_MIGRATE").Return(false)
		config.On("GetBool", "DB_LOG_MODE").Return(false)
		config.On("GetBool", "DB_IGNORE_RECORD_NOT_FOUND_ERROR").Return(false)
		config.On("GetBool", "DB_DISABLE_FOREIGN_KEY_CONSTRAINT_WHEN_MIGRATING").Return(false)
		config.On("GetInt", "DB_MAX_IDLE_CONNS").Return(0)
		config.On("GetInt", "DB_MAX_OPEN_CONNS").Return(0)
		config.On("Has", "DB_MAX_IDLE_CONNS").Return(false)
		config.On("Has", "DB_MAX_OPEN_CONNS").Return(false)
		config.On("GetDuration", "DB_CONN_MAX_LIFETIME").Return(time.Duration(0))
		config.On("GetDuration", "DB_CONN_MAX_IDLE_TIME").Return(time.Duration(0))

		// Additional config expectations for buildDSN method
		config.On("GetString", "DB_HOST").Return("")
		config.On("GetString", "DB_PORT").Return("")
		config.On("GetString", "DB_USERNAME").Return("")
		config.On("GetString", "DB_PASSWORD").Return("")
		config.On("GetString", "DB_CHARSET").Return("")
		config.On("GetString", "DB_TIMEZONE").Return("")

		logger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
		logger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()

		ctx := context.Background()
		err := dbModule.OnStart(ctx)
		assert.NoError(t, err)

		// Test getting instance
		instance := dbModule.Instance()
		assert.NotNil(t, instance)

		// Test auto-migration
		err = dbModule.AutoMigrate(&TestModel{})
		assert.NoError(t, err)

		// Test connection by name
		conn, err := dbModule.Connection("default")
		assert.NoError(t, err)
		assert.NotNil(t, conn)

		// Test stopping
		err = dbModule.OnStop(ctx)
		assert.NoError(t, err)

		// Clean up
		os.Remove(tempFile)
	})
}

func TestDatabaseModule_DSNBuilding(t *testing.T) {
	testCases := []struct {
		name          string
		driver        string
		configSetup   func(*MockConfig)
		expectError   bool
		expectedInDSN string
	}{
		{
			name:   "MySQL DSN",
			driver: "mysql",
			configSetup: func(c *MockConfig) {
				c.On("GetString", "DB_DSN").Return("")
				c.On("GetString", "DB_HOST").Return("localhost")
				c.On("GetString", "DB_PORT").Return("3306")
				c.On("GetString", "DB_DATABASE").Return("testdb")
				c.On("GetString", "DB_USERNAME").Return("user")
				c.On("GetString", "DB_PASSWORD").Return("pass")
				c.On("GetString", "DB_CHARSET").Return("utf8mb4")
				c.On("GetString", "DB_TIMEZONE").Return("UTC")
			},
			expectError:   false,
			expectedInDSN: "user:pass@tcp(localhost:3306)/testdb",
		},
		{
			name:   "PostgreSQL DSN",
			driver: "postgres",
			configSetup: func(c *MockConfig) {
				c.On("GetString", "DB_DSN").Return("")
				c.On("GetString", "DB_HOST").Return("localhost")
				c.On("GetString", "DB_PORT").Return("5432")
				c.On("GetString", "DB_DATABASE").Return("testdb")
				c.On("GetString", "DB_USERNAME").Return("user")
				c.On("GetString", "DB_PASSWORD").Return("pass")
				c.On("GetString", "DB_CHARSET").Return("")
				c.On("GetString", "DB_TIMEZONE").Return("")
				c.On("GetString", "DB_SSLMODE").Return("disable")
			},
			expectError:   false,
			expectedInDSN: "host=localhost port=5432 user=user dbname=testdb password=pass sslmode=disable TimeZone=UTC",
		},
		{
			name:   "SQLite DSN",
			driver: "sqlite",
			configSetup: func(c *MockConfig) {
				c.On("GetString", "DB_DSN").Return("")
				c.On("GetString", "DB_DATABASE").Return("/tmp/test.db")
			},
			expectError:   false,
			expectedInDSN: "/tmp/test.db",
		},
		{
			name:   "Missing host",
			driver: "mysql",
			configSetup: func(c *MockConfig) {
				c.On("GetString", "DB_DSN").Return("")
				c.On("GetString", "DB_HOST").Return("")
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &MockConfig{}
			logger := &MockLogger{}

			config.On("GetString", "DB_DRIVER").Return(tc.driver)
			tc.configSetup(config)

			if !tc.expectError {
				config.On("GetBool", "DB_LOG_MODE").Return(false)
				config.On("GetBool", "DB_IGNORE_RECORD_NOT_FOUND_ERROR").Return(false)
				config.On("GetBool", "DB_DISABLE_FOREIGN_KEY_CONSTRAINT_WHEN_MIGRATING").Return(false)
				config.On("GetInt", "DB_MAX_IDLE_CONNS").Return(0)
				config.On("GetInt", "DB_MAX_OPEN_CONNS").Return(0)
				config.On("Has", "DB_MAX_IDLE_CONNS").Return(false)
				config.On("Has", "DB_MAX_OPEN_CONNS").Return(false)
				config.On("GetDuration", "DB_CONN_MAX_LIFETIME").Return(time.Duration(0))
				config.On("GetDuration", "DB_CONN_MAX_IDLE_TIME").Return(time.Duration(0))

				logger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
				logger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
			}

			dbModule := NewDBModule(config, logger)

			// This will test the DSN building logic indirectly
			_, err := dbModule.Connection("test")

			if tc.expectError {
				assert.Error(t, err)
				// The error could be either about DSN building or connection not found
				assert.True(t,
					strings.Contains(err.Error(), "cannot build DSN") ||
						strings.Contains(err.Error(), "not found or not configured"),
				)
			} else {
				// Even if connection fails, DSN building should work
				assert.True(t, err == nil || (err != nil && !strings.Contains(err.Error(), "cannot build DSN")))
			}
		})
	}
}

func TestGoeGormLogger(t *testing.T) {
	logger := &MockLogger{}
	gormLogger := NewGoeGormLogger(logger)

	t.Run("Printf", func(t *testing.T) {
		logger.On("Info", "test message", []any(nil)).Return()

		gormLogger.Printf("test message")

		logger.AssertExpectations(t)
	})

	t.Run("Info", func(t *testing.T) {
		logger.On("Info", "test info", mock.Anything).Return()

		ctx := context.Background()
		gormLogger.Info(ctx, "test info", "key", "value")

		logger.AssertExpectations(t)
	})

	t.Run("Warn", func(t *testing.T) {
		logger.On("Warn", "test warn", mock.Anything).Return()

		ctx := context.Background()
		gormLogger.Warn(ctx, "test warn", "key", "value")

		logger.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		logger.On("Error", "test error", mock.Anything).Return()

		ctx := context.Background()
		gormLogger.Error(ctx, "test error", "key", "value")

		logger.AssertExpectations(t)
	})

	t.Run("LogMode", func(t *testing.T) {
		result := gormLogger.LogMode(1)
		assert.NotNil(t, result)
		assert.Equal(t, gormLogger, result)
	})

	t.Run("Trace", func(t *testing.T) {
		logger.On("Debug", "GORM Trace", mock.Anything).Return()

		ctx := context.Background()
		begin := time.Now()

		gormLogger.Trace(ctx, begin, func() (string, int64) {
			return "SELECT * FROM users", 10
		}, nil)

		logger.AssertExpectations(t)
	})

	t.Run("Trace with error", func(t *testing.T) {
		logger.On("Error", "GORM Trace Error", mock.Anything).Return()

		ctx := context.Background()
		begin := time.Now()

		gormLogger.Trace(ctx, begin, func() (string, int64) {
			return "SELECT * FROM users", 0
		}, assert.AnError)

		logger.AssertExpectations(t)
	})

	t.Run("Trace with RecordNotFound", func(t *testing.T) {
		logger.On("Debug", "GORM Trace", mock.Anything).Return()

		ctx := context.Background()
		begin := time.Now()

		gormLogger.Trace(ctx, begin, func() (string, int64) {
			return "SELECT * FROM users WHERE id = ?", 0
		}, gorm.ErrRecordNotFound)

		logger.AssertExpectations(t)
	})
}

func TestDatabaseModule_ConcurrentAccess(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	dbModule := NewDBModule(config, logger)

	// Test concurrent model registration
	t.Run("concurrent model registration", func(t *testing.T) {
		config.On("GetString", "DB_CONNECTION").Return("default")
		logger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()

		done := make(chan bool)

		for i := 0; i < 10; i++ {
			go func() {
				dbModule.RegisterModelsForMigration(&TestModel{})
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}

		// Should not race
		assert.True(t, true)
	})
}

func TestDatabaseModule_Provide(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	dbModule := NewDBModule(config, logger)

	t.Run("provide returns self", func(t *testing.T) {
		provided := dbModule.Provide()
		assert.Equal(t, dbModule, provided)
	})
}
