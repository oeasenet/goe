package db

import (
	"context"
	"fmt"
	"strings" // Added for strings.ToUpper
	"sync"

	"go.oease.dev/goe/v2/contract"
	"gorm.io/gorm"
)

// DatabaseModule implements the contract.DB and contract.Module interfaces
type DatabaseModule struct {
	config contract.Config
	logger contract.Logger
	mu     sync.RWMutex
	// defaultConnectionName string // This can be derived from config when needed
	connections map[string]*gorm.DB
	// gormConfig *gorm.Config // To be added later for more GORM specific configs
}

// NewDBModule creates a new DatabaseModule instance
func NewDBModule(config contract.Config, logger contract.Logger) *DatabaseModule {
	return &DatabaseModule{
		config:      config,
		logger:      logger,
		connections: make(map[string]*gorm.DB),
	}
}

// --- contract.DB interface implementation ---

// Instance returns the underlying GORM DB instance for the default connection
func (dbm *DatabaseModule) Instance() *gorm.DB {
	defaultConnectionName := dbm.config.GetString("DB_CONNECTION")
	if defaultConnectionName == "" {
		defaultConnectionName = "default"
	}

	conn, err := dbm.Connection(defaultConnectionName)
	if err != nil {
		dbm.logger.Error("Failed to get default database instance",
			contract.NewField("connection_name", defaultConnectionName),
			contract.NewField("error", err),
		)
		return nil
	}
	return conn
}

// Connection returns a specific GORM DB instance by name
func (dbm *DatabaseModule) Connection(name string) (*gorm.DB, error) {
	dbm.mu.RLock()
	conn, ok := dbm.connections[name]
	dbm.mu.RUnlock()

	if !ok {
		// Decision: Do not connect on-demand here. Connections should be explicitly defined and set up OnStart.
		// If a connection is requested that wasn't configured/failed, it's an error.
		return nil, fmt.Errorf("database connection '%s' not found or not configured", name)
	}
	return conn, nil
}

// --- contract.Module interface implementation ---

// Name returns the unique name of the module
func (dbm *DatabaseModule) Name() string {
	return "db"
}

// OnStart is called when the module starts
// This is where database connections will be established
func (dbm *DatabaseModule) OnStart(ctx context.Context) error {
	dbm.logger.Info("Database module OnStart")
	dbm.mu.Lock()
	defer dbm.mu.Unlock()

	defaultConnectionName := dbm.config.GetString("DB_CONNECTION")
	if defaultConnectionName == "" {
		defaultConnectionName = "default"
	}

	dbm.logger.Info("Attempting to connect to default database", contract.NewField("connection_config_name", defaultConnectionName))

	// The 'connect' method uses 'default' for config keys like DB_DRIVER, DB_HOST
	// or DB_MYCONN_DRIVER if name is 'myconn'.
	// So, if defaultConnectionName from config is e.g. "main_db", we pass "main_db" to connect.
	// If defaultConnectionName is "default" (or was empty and defaulted to "default"), we pass "default".
	db, err := dbm.connect(defaultConnectionName)
	if err != nil {
		dbm.logger.Error("Failed to connect to default database",
			contract.NewField("connection_config_name", defaultConnectionName),
			contract.NewField("error", err.Error()),
		)
		// Allow app to start, Instance() will return nil.
	} else {
		// Store the connection using the name it will be requested by, which is defaultConnectionName.
		dbm.connections[defaultConnectionName] = db
		dbm.logger.Info("Successfully connected to default database", contract.NewField("connection_config_name", defaultConnectionName))
	}

	// TODO: Implement logic for multiple connections based on a configuration like DB_CONNECTIONS_LIST="secondary_db,tertiary_db"
	// Then iterate through these names, call dbm.connect(name), and store them in dbm.connections[name].

	// Auto-migration logic (optional, based on config)
	// Check general DB_AUTO_MIGRATE or specific DB_CONNNAME_AUTO_MIGRATE
	autoMigrateConfigKey := fmt.Sprintf("DB_%s_AUTO_MIGRATE", strings.ToUpper(defaultConnectionName))
	if nameKeyIsDefault := strings.ToLower(defaultConnectionName) == "default"; nameKeyIsDefault {
		autoMigrateConfigKey = "DB_AUTO_MIGRATE" // for "default" connection, use DB_AUTO_MIGRATE
	}

	if dbm.config.GetBool("DB_AUTO_MIGRATE_ANY") || dbm.config.GetBool(autoMigrateConfigKey) {
		if db != nil {
			dbm.logger.Info("Auto-migration is enabled for default connection. Models should be registered and migrated by the application.",
				contract.NewField("connection_config_name", defaultConnectionName),
				contract.NewField("checked_config_key", autoMigrateConfigKey),
			)
		} else {
			dbm.logger.Warn("Auto-migration enabled for default connection, but connection failed.",
				contract.NewField("connection_config_name", defaultConnectionName),
				contract.NewField("checked_config_key", autoMigrateConfigKey),
			)
		}
	}
	return nil
}

// OnStop is called when the module stops
// This is where database connections will be closed
func (dbm *DatabaseModule) OnStop(ctx context.Context) error {
	dbm.logger.Info("Database module OnStop")
	dbm.mu.Lock()
	defer dbm.mu.Unlock()

	var lastErr error
	for name, conn := range dbm.connections {
		dbm.logger.Info("Closing database connection", contract.NewField("connection", name))
		sqlDB, err := conn.DB()
		if err != nil {
			dbm.logger.Error("Failed to get SQL DB from GORM instance for closing", contract.NewField("connection", name), contract.NewField("error", err))
			lastErr = err
			continue
		}
		if err := sqlDB.Close(); err != nil {
			dbm.logger.Error("Failed to close database connection", contract.NewField("connection", name), contract.NewField("error", err))
			lastErr = err
		}
		delete(dbm.connections, name)
	}
	return lastErr
}

// AutoMigrate performs auto migration for the given GORM models on the default connection
func (dbm *DatabaseModule) AutoMigrate(dst ...interface{}) error {
	defaultDB := dbm.Instance() // This already resolves the default connection name
	if defaultDB == nil {
		defaultConnName := dbm.config.GetString("DB_CONNECTION")
		if defaultConnName == "" {
			defaultConnName = "default"
		}
		return fmt.Errorf("default database instance ('%s') is not available for auto-migration", defaultConnName)
	}
	return defaultDB.AutoMigrate(dst...)
}

// AutoMigrateOnConnection performs auto migration for the given GORM models on a specific connection
func (dbm *DatabaseModule) AutoMigrateOnConnection(connectionName string, dst ...interface{}) error {
	conn, err := dbm.Connection(connectionName)
	if err != nil {
		return err
	}
	return conn.AutoMigrate(dst...)
}

// Provide returns the DB instance for Fx
// This will allow injecting contract.DB
func (dbm *DatabaseModule) Provide() contract.DB {
	return dbm
}
