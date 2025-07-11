package cache

import (
	"context"
	"fmt"

	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/validator"
)

// Module represents the cache module for Fx
type Module struct {
	manager contract.CacheManager
	logger  contract.Logger
	config  contract.Config
}

// NewModule creates a new cache module
func NewModule(config contract.Config, logger contract.Logger) *Module {
	// Create cache manager
	manager := NewManager(config)

	// Register all built-in drivers (Fiber storage drivers)
	RegisterBuiltinDrivers(manager)

	return &Module{
		manager: manager,
		logger:  logger,
		config:  config,
	}
}

// Name returns the module name
func (m *Module) Name() string {
	return "cache"
}

// OnStart is called when the module starts
func (m *Module) OnStart(ctx context.Context) error {
	m.logger.Info("Cache module started",
		"driver", m.manager.Driver(),
		"store", m.config.GetString("CACHE_STORE"),
	)
	return nil
}

// OnStop is called when the module stops
func (m *Module) OnStop(ctx context.Context) error {
	m.logger.Info("Cache module stopped")

	// Close all stores
	// This is handled by each store's Close method

	return nil
}

// Provide returns the cache manager instance for Fx
func (m *Module) Provide() contract.CacheManager {
	return m.manager
}

// ProvideCache returns the default cache instance for Fx
func (m *Module) ProvideCache() contract.Cache {
	return m.manager.Store()
}

// ValidateConfig validates the cache module configuration
func (m *Module) ValidateConfig() error {
	v := validator.NewConfigValidator(m.config, "cache")

	// Cache store is optional, defaults to memory
	store := m.config.GetString("CACHE_STORE")
	if store != "" {
		// Validate store type
		validStores := []string{"memory", "redis", "memcache", "badger", "sqlite3", "postgres", "mysql", "mongodb", "dynamodb", "s3"}
		v.Optional("CACHE_STORE", "Cache store type", validator.ValidateOneOf(validStores...))

		// Store-specific validations
		switch store {
		case "redis":
			v.RequireWithValidator("CACHE_REDIS_ADDR", "Redis server address", validator.ValidateHostPort)

			if m.config.Has("CACHE_REDIS_DB") {
				v.Optional("CACHE_REDIS_DB", "Redis database number", validator.ValidatePositiveInt)
			}

		case "memcache":
			v.Require("CACHE_MEMCACHE_SERVERS", "Memcache server addresses (comma-separated)")

		case "postgres", "mysql":
			v.Require("CACHE_DB_HOST", "Database host")
			v.Require("CACHE_DB_DATABASE", "Database name")
			v.Require("CACHE_DB_USERNAME", "Database username")

			if m.config.Has("CACHE_DB_PORT") {
				v.Optional("CACHE_DB_PORT", "Database port", validator.ValidatePort)
			}

		case "mongodb":
			v.Require("CACHE_MONGODB_URI", "MongoDB connection URI")
			v.Require("CACHE_MONGODB_DATABASE", "MongoDB database name")

		case "dynamodb":
			v.Require("CACHE_DYNAMODB_TABLE", "DynamoDB table name")
			v.Require("CACHE_DYNAMODB_REGION", "AWS region")

		case "s3":
			v.Require("CACHE_S3_BUCKET", "S3 bucket name")
			v.Require("CACHE_S3_REGION", "AWS region")

		case "sqlite3":
			// SQLite can use memory or file, both are optional
			// Default is usually ":memory:" or a file path

		case "badger":
			// Badger uses local storage, path is optional
			// Default is usually a temp directory
		}
	}

	// TTL is optional but should be positive if set
	if m.config.Has("CACHE_TTL") {
		v.Optional("CACHE_TTL", "Default cache TTL", func(value any) error {
			duration := m.config.GetDuration("CACHE_TTL")
			if duration < 0 {
				return fmt.Errorf("TTL must be positive")
			}
			return nil
		})
	}

	return v.Validate()
}
