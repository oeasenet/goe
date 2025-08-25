package db

import (
	"context"
	"fmt"
	"strings" // Added for strings.ToUpper
	"sync"

	"gorm.io/gorm"

	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/internal/configvalidator"
)

// DatabaseModule implements the contract.DB and contract.Module interfaces
type DatabaseModule struct {
	config contract.Config
	logger contract.Logger
	mu     sync.RWMutex
	// defaultConnectionName string // This can be derived from config when needed
	connections map[string]*gorm.DB
	// gormConfig *gorm.Config // To be added later for more GORM specific configs

	// Migration registration storage
	registeredModels map[string][]interface{} // connectionName -> models slice
}

// NewDBModule creates a new DatabaseModule instance
func NewDBModule(config contract.Config, logger contract.Logger) *DatabaseModule {
	return &DatabaseModule{
		config:           config,
		logger:           logger,
		connections:      make(map[string]*gorm.DB),
		registeredModels: make(map[string][]interface{}),
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
		// Only log this as debug during startup, not as error
		dbm.logger.Debug("Database instance not yet available",
			"connection_name", defaultConnectionName,
			"error", err,
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

	// Get default connection name
	defaultConnectionName := dbm.config.GetString("DB_CONNECTION")
	if defaultConnectionName == "" {
		defaultConnectionName = "default"
	}

	// Connect to default database
	dbm.logger.Info("Attempting to connect to default database", "connection_config_name", defaultConnectionName)
	db, err := dbm.connect(defaultConnectionName)
	if err != nil {
		dbm.logger.Error("Failed to connect to default database",
			"connection_config_name", defaultConnectionName,
			"error", err.Error(),
		)
		// Allow app to start, DB() will return nil.
	} else {
		// Store the connection using the name it will be requested by, which is defaultConnectionName.
		dbm.connections[defaultConnectionName] = db
		dbm.logger.Info("Successfully connected to default database", "connection_config_name", defaultConnectionName)
	}

	// Connect to additional databases if configured
	connectionsList := dbm.config.GetString("DB_CONNECTIONS")
	if connectionsList != "" {
		// Split the comma-separated list of connection names
		connectionNames := strings.Split(connectionsList, ",")
		for _, connName := range connectionNames {
			connName = strings.TrimSpace(connName)

			// Skip if it's the default connection (already connected)
			if connName == defaultConnectionName {
				continue
			}

			// Skip if empty
			if connName == "" {
				continue
			}

			dbm.logger.Info("Attempting to connect to additional database", "connection_name", connName)
			conn, err := dbm.connect(connName)
			if err != nil {
				dbm.logger.Error("Failed to connect to additional database",
					"connection_name", connName,
					"error", err.Error(),
				)
				// Continue with other connections
			} else {
				dbm.connections[connName] = conn
				dbm.logger.Info("Successfully connected to additional database", "connection_name", connName)

				// Check for auto-migration for this connection
				autoMigrateKey := fmt.Sprintf("DB_%s_AUTO_MIGRATE", strings.ToUpper(connName))
				if dbm.config.GetBool("DB_AUTO_MIGRATE_ANY") || dbm.config.GetBool(autoMigrateKey) {
					dbm.logger.Info("Auto-migration is enabled for connection",
						"connection_name", connName,
						"checked_config_key", autoMigrateKey,
					)

					// Perform auto-migration for registered models
					if models, exists := dbm.registeredModels[connName]; exists && len(models) > 0 {
						dbm.logger.Info("Performing database migration...",
							"connection_name", connName,
							"model_count", len(models),
						)
						if err := conn.AutoMigrate(models...); err != nil {
							dbm.logger.Error("Database migration failed",
								"connection_name", connName,
								"error", err,
							)
						} else {
							dbm.logger.Debug("Database migration completed successfully",
								"connection_name", connName,
								"model_count", len(models),
							)
						}
					} else {
						dbm.logger.Debug("No models registered for auto-migration on this connection",
							"connection_name", connName,
						)
					}
				}
			}
		}
	}

	// Auto-migration logic for default connection (optional, based on config)
	autoMigrateConfigKey := fmt.Sprintf("DB_%s_AUTO_MIGRATE", strings.ToUpper(defaultConnectionName))
	if nameKeyIsDefault := strings.ToLower(defaultConnectionName) == "default"; nameKeyIsDefault {
		autoMigrateConfigKey = "DB_AUTO_MIGRATE" // for "default" connection, use DB_AUTO_MIGRATE
	}

	if dbm.config.GetBool("DB_AUTO_MIGRATE_ANY") || dbm.config.GetBool(autoMigrateConfigKey) {
		if db != nil {
			dbm.logger.Info("Auto-migration is enabled for default connection",
				"connection_config_name", defaultConnectionName,
				"checked_config_key", autoMigrateConfigKey,
			)

			// Perform auto-migration for registered models on default connection
			if models, exists := dbm.registeredModels[defaultConnectionName]; exists && len(models) > 0 {
				dbm.logger.Info("Performing database migration...",
					"connection_name", defaultConnectionName,
					"model_count", len(models),
				)
				if err := db.AutoMigrate(models...); err != nil {
					dbm.logger.Error("Database migration failed",
						"connection_name", defaultConnectionName,
						"error", err,
					)
				} else {
					dbm.logger.Debug("Database migration completed successfully",
						"connection_name", defaultConnectionName,
						"model_count", len(models),
					)
				}
			} else {
				dbm.logger.Debug("No models registered for auto-migration on default connection",
					"connection_name", defaultConnectionName,
				)
			}
		} else {
			dbm.logger.Warn("Auto-migration enabled for default connection, but connection failed.",
				"connection_config_name", defaultConnectionName,
				"checked_config_key", autoMigrateConfigKey,
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
		dbm.logger.Info("Closing database connection", "connection", name)
		sqlDB, err := conn.DB()
		if err != nil {
			dbm.logger.Error("Failed to get SQL DB from GORM instance for closing", "connection", name, "error", err)
			lastErr = err
			continue
		}
		if err := sqlDB.Close(); err != nil {
			dbm.logger.Error("Failed to close database connection", "connection", name, "error", err)
			lastErr = err
		}
		delete(dbm.connections, name)
	}
	return lastErr
}

// AutoMigrate performs auto migration for the given GORM models on the default connection
// Use RegisterModelsForMigration() for fully automated migration unless you want to do everything manually after the app successfully started.
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

// RegisterModelsForMigration pre-registers models for automatic migration on the default connection
func (dbm *DatabaseModule) RegisterModelsForMigration(dst ...interface{}) {
	dbm.mu.Lock()
	defer dbm.mu.Unlock()

	defaultConnectionName := dbm.config.GetString("DB_CONNECTION")
	if defaultConnectionName == "" {
		defaultConnectionName = "default"
	}

	dbm.registeredModels[defaultConnectionName] = append(dbm.registeredModels[defaultConnectionName], dst...)
	dbm.logger.Info("Registered models for auto-migration on default connection",
		"connection_name", defaultConnectionName,
		"model_count", len(dst),
	)
}

// RegisterModelsForMigrationOnConnection pre-registers models for automatic migration on a specific connection
func (dbm *DatabaseModule) RegisterModelsForMigrationOnConnection(connectionName string, dst ...interface{}) {
	dbm.mu.Lock()
	defer dbm.mu.Unlock()

	dbm.registeredModels[connectionName] = append(dbm.registeredModels[connectionName], dst...)
	dbm.logger.Info("Registered models for auto-migration on specific connection",
		"connection_name", connectionName,
		"model_count", len(dst),
	)
}

// Provide returns the DB instance for Fx
// This will allow injecting contract.DB
func (dbm *DatabaseModule) Provide() contract.DB {
	return dbm
}

// ValidateConfig validates the database module configuration
func (dbm *DatabaseModule) ValidateConfig() error {
	v := configvalidator.NewConfigValidator(dbm.config, "db")

	// Get default connection name
	defaultConnectionName := dbm.config.GetString("DB_CONNECTION")
	if defaultConnectionName == "" {
		defaultConnectionName = "default"
	}

	// For default connection
	configPrefix := "DB_"
	if defaultConnectionName != "default" {
		configPrefix = fmt.Sprintf("DB_%s_", strings.ToUpper(defaultConnectionName))
	}

	// Driver is always required (except for SQLite which can work with defaults)
	driver := dbm.config.GetString(configPrefix + "DRIVER")
	if driver != "" {
		v.RequireWithValidator(configPrefix+"DRIVER", "Database driver type", configvalidator.ValidateDatabaseDriver)

		// For non-SQLite databases, host is required
		if !strings.Contains(strings.ToLower(driver), "sqlite") {
			v.Require(configPrefix+"HOST", "Database host")

			// Database name is required for most drivers
			v.Require(configPrefix+"DATABASE", "Database name")
		}

		// Optional validations
		if dbm.config.Has(configPrefix + "PORT") {
			v.Optional(configPrefix+"PORT", "Database port", configvalidator.ValidatePort)
		}
	} else {
		// If no driver specified, it's required
		v.Require(configPrefix+"DRIVER", "Database driver type")
	}

	// Check additional connections if specified
	connectionsList := dbm.config.GetString("DB_CONNECTIONS")
	if connectionsList != "" {
		connections := strings.Split(connectionsList, ",")
		for _, connName := range connections {
			connName = strings.TrimSpace(connName)
			if connName == "" || connName == defaultConnectionName {
				continue
			}

			connPrefix := fmt.Sprintf("DB_%s_", strings.ToUpper(connName))

			// Each additional connection must have a driver
			v.RequireWithValidator(connPrefix+"DRIVER", fmt.Sprintf("Database driver for connection '%s'", connName), configvalidator.ValidateDatabaseDriver)

			connDriver := dbm.config.GetString(connPrefix + "DRIVER")
			if connDriver != "" && !strings.Contains(strings.ToLower(connDriver), "sqlite") {
				v.Require(connPrefix+"HOST", fmt.Sprintf("Database host for connection '%s'", connName))
				v.Require(connPrefix+"DATABASE", fmt.Sprintf("Database name for connection '%s'", connName))

				if dbm.config.Has(connPrefix + "PORT") {
					v.Optional(connPrefix+"PORT", fmt.Sprintf("Database port for connection '%s'", connName), configvalidator.ValidatePort)
				}
			}
		}
	}

	return v.Validate()
}
