package mongodb

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.oease.dev/goe/v2/contract"
	"strings"
	"sync"
	"time"
)

// DatabaseModule implements the contract.DB and contract.Module interfaces
type DatabaseModule struct {
	logger      contract.Logger
	config      contract.Config
	mu          sync.RWMutex
	connections map[string]*mongo.Database
}

// NewDBModule creates a new DatabaseModule instance
func NewDBModule(config contract.Config, logger contract.Logger) *DatabaseModule {
	return &DatabaseModule{
		logger:      logger,
		config:      config,
		connections: make(map[string]*mongo.Database),
	}
}

// Instance returns the underlying MONGO DB instance for the default connection
func (dbm *DatabaseModule) Instance() *mongo.Database {
	defaultConnectionName := dbm.config.GetString("MONGO_DB_CONNECTION")
	if defaultConnectionName == "" {
		defaultConnectionName = "default"
	}

	conn, err := dbm.Connection(defaultConnectionName)
	if err != nil {
		dbm.logger.Errorf("Failed to get default database instance, connection_name: %s, error: %s", defaultConnectionName, err)
		return nil
	}
	return conn
}

// Connection returns a specific Mongo DB instance by name
func (dbm *DatabaseModule) Connection(name string) (*mongo.Database, error) {
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

func (dbm *DatabaseModule) IsNoDocumentsError(err error) bool {
	return errors.Is(err, mongo.ErrNoDocuments)
}

func (dbm *DatabaseModule) Ctx() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	go func() {
		<-ctx.Done() // Wait until timeout or manual cancellation
		cancel()     // Release resources
	}()
	return ctx
}

// --- contract.Module interface implementation ---

// Name returns the unique name of the module
func (dbm *DatabaseModule) Name() string {
	return "mongo_db"
}

// OnStart is called when the module starts
// This is where database connections will be established
func (dbm *DatabaseModule) OnStart(ctx context.Context) error {
	dbm.logger.Info("MONGO Database module OnStart")
	dbm.mu.Lock()
	defer dbm.mu.Unlock()

	// Get default connection name
	defaultConnectionName := dbm.config.GetString("MONGO_DB_CONNECTION")
	if defaultConnectionName == "" {
		defaultConnectionName = "default"
	}

	// Connect to default database
	dbm.logger.Infof("Attempting to connect to default mongo database, connection_config_name: %s", defaultConnectionName)
	db, err := dbm.connect(defaultConnectionName)
	if err != nil {
		dbm.logger.Errorf("Failed to connect to default mongo database, connection_config_name: %s, error: %s", defaultConnectionName, err.Error())
		// Allow app to start, Instance() will return nil.
	} else {
		// Store the connection using the name it will be requested by, which is defaultConnectionName.
		dbm.connections[defaultConnectionName] = db
		dbm.logger.Infof("Successfully connected to default mongo database, connection_config_name: %s", defaultConnectionName)
	}

	// Connect to additional databases if configured
	connectionsList := dbm.config.GetString("MONGO_DB_CONNECTIONS")
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

			dbm.logger.Infof("Attempting to connect to additional database, connection_name: %s", connName)
			conn, err := dbm.connect(connName)
			if err != nil {
				dbm.logger.Errorf("Failed to connect to additional database, connection_name: %s,  error: %s", connName, err.Error())
				// Continue with other connections
			} else {
				dbm.connections[connName] = conn
				dbm.logger.Infof("Successfully connected to additional database", "connection_name", connName)

			}
		}
	}

	return nil
}

// OnStop is called when the module stops
// This is where database connections will be closed
func (dbm *DatabaseModule) OnStop(ctx context.Context) error {
	dbm.logger.Info("Mongo Database module OnStop")
	dbm.mu.Lock()
	defer dbm.mu.Unlock()

	var lastErr error
	for name, conn := range dbm.connections {
		dbm.logger.Info("Closing mongo database connection", "connection", name)
		if err := conn.Client().Disconnect(ctx); err != nil {
			dbm.logger.Errorf("Failed to close database connection, connection: %s, error: %s", name, err.Error())
			lastErr = err
		}
		delete(dbm.connections, name)
	}
	return lastErr
}

// Provide returns the MONGODB instance for Fx
// This will allow injecting contract.MongoDB
func (dbm *DatabaseModule) Provide() contract.MongoDB {
	return dbm
}
