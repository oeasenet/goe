package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.oease.dev/goe/v2/contract"
)

// buildClientOptions maps the MONGO_* configuration onto the driver's client
// options and returns them with the database name.
//
// Credentials resolve URI-first: userinfo embedded in the URI wins, and the
// separate MONGO_USERNAME/MONGO_PASSWORD variables fill in whenever the URI
// carries none — the only way to pair a secret with a credential-free URI,
// since code options deliberately cannot set either.
func buildClientOptions(config contract.Config, customMonitor *event.CommandMonitor, logger contract.Logger) (*options.ClientOptions, string, error) {
	// get uri and database
	uri := config.GetString("MONGO_URI")
	dbName := config.GetString("MONGO_DB_NAME")

	if uri == "" {
		return nil, "", fmt.Errorf("no MONGO_URI configured")
	}
	if dbName == "" {
		return nil, "", fmt.Errorf("no MONGO_DB_NAME configured")
	}

	opt := options.Client()
	opt.ApplyURI(uri)

	// Environment credentials, applied only when the URI embeds none.
	if opt.Auth == nil {
		username := config.GetString("MONGO_USERNAME")
		password := config.GetString("MONGO_PASSWORD")
		if username != "" || password != "" {
			opt.SetAuth(options.Credential{Username: username, Password: password})
		}
	}

	if customMonitor != nil {
		opt.SetMonitor(customMonitor)
	} else if config.GetBool("MONGO_DEBUG") {
		// Only enable command logging when debug is explicitly enabled
		opt.SetMonitor(defaultMonitor(logger))
	}

	// Configure connection
	if config.Has("MONGO_MIN_POOL_SIZE") {
		minPoolSize := config.GetInt("MONGO_MIN_POOL_SIZE")
		if minPoolSize > 0 {
			opt.SetMinPoolSize(uint64(minPoolSize))
		}
	}
	if config.Has("MONGO_MAX_POOL_SIZE") {
		maxPoolSize := config.GetInt("MONGO_MAX_POOL_SIZE")
		if maxPoolSize > 0 {
			opt.SetMaxPoolSize(uint64(maxPoolSize))
		}
	}

	if config.Has("MONGO_MAX_CONN_IDLE_TIME") {
		maxIdleTime := config.GetDuration("MONGO_MAX_CONN_IDLE_TIME")
		if maxIdleTime > 0 {
			opt.SetMaxConnIdleTime(maxIdleTime)
		}
	}

	return opt, dbName, nil
}

// connect builds the client. The driver performs no I/O here; reachability is
// proven by the startup ping in connectAndVerify.
func (dbm *DatabaseModule) connect() (*mongo.Database, error) {
	opt, dbName, err := buildClientOptions(dbm.config, dbm.customMonitor, dbm.logger)
	if err != nil {
		return nil, err
	}

	client, err := mongo.Connect(opt)
	if err != nil {
		return nil, err
	}
	return client.Database(dbName), nil
}

// defaultMonitor returns a default CommandMonitor that logs MongoDB command events.
//
// This monitor logs all MongoDB operations with different log levels:
//   - Debug: when a command succeeds, including duration
//   - Error: when a command fails, including the error message
func defaultMonitor(logger contract.Logger) *event.CommandMonitor {
	return &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			logger.Debug("[MONGO START]",
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
