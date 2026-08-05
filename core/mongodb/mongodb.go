package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.oease.dev/goe/v2/contract"
	goeconfig "go.oease.dev/goe/v2/core/config"
)

// DatabaseModule implements the contract.MongoDB and contract.Module
// interfaces. It manages a single MongoDB connection; applications needing a
// second data source construct their own mongo.Client via the driver.
type DatabaseModule struct {
	logger contract.Logger

	// config is the effective configuration: the base config with the Option
	// overlay applied, so code-configured values behave exactly like
	// environment variables everywhere the module reads config.
	config        contract.Config
	customMonitor *event.CommandMonitor

	mu sync.RWMutex
	db *mongo.Database

	// optErrs holds failures from Option application. They are reported by
	// ValidateConfig so that startup aborts; the module fell back to the
	// environment-only configuration, which is never actually served.
	optErrs []error
}

// NewDBModule creates a new DatabaseModule instance.
//
// Configuration resolves in layers: MONGO_* environment variables, then opts.
// Anything set through an Option wins over the environment. If any option
// fails, every option is discarded, the environment-only configuration stays
// in effect, and ValidateConfig aborts startup with all collected errors.
func NewDBModule(config contract.Config, logger contract.Logger, opts ...Option) *DatabaseModule {
	overrides, monitor, optErrs := resolveSettings(opts)
	if len(optErrs) > 0 {
		// Every option is discarded, the monitor included.
		monitor = nil
	} else if len(overrides) > 0 {
		config = goeconfig.NewConfigWrapper(config, overrides)
	}

	return &DatabaseModule{
		logger:        logger.With("module", "mongo"),
		config:        config,
		customMonitor: monitor,
		optErrs:       optErrs,
	}
}

// DB returns the database instance, or nil before OnStart has run.
func (dbm *DatabaseModule) DB() *mongo.Database {
	dbm.mu.RLock()
	defer dbm.mu.RUnlock()
	return dbm.db
}

// --- contract.Module interface implementation ---

// Name returns the unique name of the module
func (dbm *DatabaseModule) Name() string {
	return "mongo_db"
}

// OnStart establishes and verifies the connection.
//
// A connection that cannot be built or does not answer a ping fails startup.
// Earlier versions logged the failure and continued with a nil database,
// which only deferred the crash to the first Col()/DB() use in a handler —
// far from the cause, at request time. mongo.Connect performs no I/O, so the
// ping (bounded by MONGO_PING_TIMEOUT, default 5s) is what actually proves
// the deployment can reach its database.
func (dbm *DatabaseModule) OnStart(ctx context.Context) error {
	dbm.logger.Debug("MongoDB module starting")
	dbm.mu.Lock()
	defer dbm.mu.Unlock()

	db, err := dbm.connectAndVerify(ctx)
	if err != nil {
		return fmt.Errorf("mongodb: %w", err)
	}
	dbm.db = db

	dbm.logger.Info("MongoDB module started", "database", db.Name())
	return nil
}

// connectAndVerify builds the client and proves it reachable with a bounded
// ping. The client is disconnected on ping failure so no resources leak from
// a failed startup.
func (dbm *DatabaseModule) connectAndVerify(ctx context.Context) (*mongo.Database, error) {
	db, err := dbm.connect()
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, dbm.pingTimeout())
	defer cancel()

	if err := db.Client().Ping(pingCtx, nil); err != nil {
		_ = db.Client().Disconnect(ctx)
		return nil, fmt.Errorf("ping failed: %w", err)
	}
	return db, nil
}

// pingTimeout returns the startup ping timeout: MONGO_PING_TIMEOUT, or 5s.
func (dbm *DatabaseModule) pingTimeout() time.Duration {
	if d := dbm.config.GetDuration("MONGO_PING_TIMEOUT"); d > 0 {
		return d
	}
	return 5 * time.Second
}

// OnStop is called when the module stops
// This is where the database connection will be closed
func (dbm *DatabaseModule) OnStop(ctx context.Context) error {
	dbm.logger.Debug("MongoDB module stopping")
	dbm.mu.Lock()
	defer dbm.mu.Unlock()

	if dbm.db != nil {
		if err := dbm.db.Client().Disconnect(ctx); err != nil {
			dbm.logger.Error("Failed to close database connection", "error", err.Error())
			return err
		}
		dbm.db = nil
	}

	dbm.logger.Info("MongoDB module stopped")
	return nil
}

// Client returns the MongoDB client, or nil before OnStart has run.
func (dbm *DatabaseModule) Client() *mongo.Client {
	db := dbm.DB()
	if db == nil {
		return nil
	}
	return db.Client()
}

// Col returns a collection from the database, or nil before OnStart has run.
func (dbm *DatabaseModule) Col(name string) *mongo.Collection {
	db := dbm.DB()
	if db == nil {
		dbm.logger.Error("Cannot get collection: database instance is nil", "collection", name)
		return nil
	}
	return db.Collection(name)
}

// Provide returns the MONGODB instance for Fx
// This will allow injecting contract.MongoDB
func (dbm *DatabaseModule) Provide() contract.MongoDB {
	return dbm
}
