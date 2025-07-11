package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.oease.dev/goe/v2/contract"
	"strings"
)

// Connect initializes a Mongo DB connection based on the provided configuration prefix.
// The configuration keys are expected to be like:
// MONGO_DB_URI, MONGO_DB_DB
// For a named connection "foo", the keys would be:
// MONGO_DB_FOO_URI, MONGO_DB_FOO_DB_NAME
func (dbm *DatabaseModule) connect(name string) (*mongo.Database, error) {
	configPrefix := "MONGO_DB_"
	if name != "default" && name != "" {
		configPrefix = fmt.Sprintf("MONGO_DB_%s_", strings.ToUpper(name))
	}

	// get uri and database
	uri := dbm.config.GetString(configPrefix + "URI")
	dbName := dbm.config.GetString(configPrefix + "DB_NAME")

	if uri == "" {
		return nil, fmt.Errorf("no URI for %s mongo connection", name)
	}
	if dbName == "" {
		return nil, fmt.Errorf("no database name for %s mongo connection", name)
	}

	opt := options.Client()
	opt.ApplyURI(uri)
	if dbm.customMonitor != nil {
		opt.SetMonitor(dbm.customMonitor)
	} else {
		opt.SetMonitor(defaultMonitor(dbm.logger))
	}

	// Configure connection
	if dbm.config.Has(configPrefix + "MIN_POOL_SIZE") {
		minPoolSize := dbm.config.GetInt(configPrefix + "MIN_POOL_SIZE")
		if minPoolSize > 0 {
			opt.SetMinPoolSize(uint64(minPoolSize))
		}
	}
	if dbm.config.Has(configPrefix + "MAX_POOL_SIZE") {
		maxPoolSize := dbm.config.GetInt(configPrefix + "MAX_POOL_SIZE")
		if maxPoolSize > 0 {
			opt.SetMaxPoolSize(uint64(maxPoolSize))
		}
	}

	if dbm.config.Has(configPrefix + "MAX_CONN_IDLE_TIME") {
		maxIdleTime := dbm.config.GetDuration(configPrefix + "MAX_CONN_IDLE_TIME")
		if maxIdleTime > 0 {
			opt.SetMaxConnIdleTime(maxIdleTime)
		}
	}

	dbm.logger.Info("MONGO Database connection established successfully",
		"name", name,
	)

	client, err := mongo.Connect(opt)
	if err != nil {
		return nil, err
	}

	dbm.logger.Info("MONGO Database connection established successfully",
		"name", name,
	)
	return client.Database(dbName), nil
}

// defaultMonitor returns a default CommandMonitor that logs MongoDB command events.
//
// This monitor logs all MongoDB operations with different log levels:
//   - Info: when a command starts
//   - Debug: when a command succeeds, including duration
//   - Error: when a command fails, including the error message
func defaultMonitor(logger contract.Logger) *event.CommandMonitor {
	return &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			logger.Info("[MONGO START]",
				"command", evt.CommandName,
				"details", evt.Command.String(),
			)
		},
		Succeeded: func(ctx context.Context, evt *event.CommandSucceededEvent) {
			logger.Debug("[MONGO SUCCEED]",
				"command", evt.CommandName,
				"duration", evt.Duration.String(),
			)
		},
		Failed: func(ctx context.Context, evt *event.CommandFailedEvent) {
			logger.Error("[MONGO FAILED]",
				"command", evt.CommandName,
				"error", evt.Failure.Error(),
			)
		},
	}
}
