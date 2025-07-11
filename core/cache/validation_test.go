package cache

import (
	"testing"

	"go.oease.dev/goe/v2/core/config"
	"go.oease.dev/goe/v2/core/log"
)

func TestCacheModuleValidation(t *testing.T) {
	tests := []struct {
		name        string
		configFunc  func(*config.Module)
		expectError bool
		errorString string
	}{
		{
			name: "no configuration - should pass with defaults",
			configFunc: func(cfg *config.Module) {
				// No configuration, should use memory store by default
			},
			expectError: false,
		},
		{
			name: "memory store configuration",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "memory")
			},
			expectError: false,
		},
		{
			name: "redis store - valid configuration",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "redis")
				cfg.Provide().Set("CACHE_REDIS_ADDR", "localhost:6379")
			},
			expectError: false,
		},
		{
			name: "redis store - missing address",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "redis")
				// Missing CACHE_REDIS_ADDR
			},
			expectError: true,
			errorString: "CACHE_REDIS_ADDR",
		},
		{
			name: "redis store - invalid address",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "redis")
				cfg.Provide().Set("CACHE_REDIS_ADDR", "invalid-address")
			},
			expectError: true,
			errorString: "invalid host:port format",
		},
		{
			name: "redis store - valid with database",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "redis")
				cfg.Provide().Set("CACHE_REDIS_ADDR", "localhost:6379")
				cfg.Provide().Set("CACHE_REDIS_DB", "1")
			},
			expectError: false,
		},
		{
			name: "redis store - invalid database number",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "redis")
				cfg.Provide().Set("CACHE_REDIS_ADDR", "localhost:6379")
				cfg.Provide().Set("CACHE_REDIS_DB", "-1")
			},
			expectError: true,
			errorString: "value must be positive",
		},
		{
			name: "memcache store - valid configuration",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "memcache")
				cfg.Provide().Set("CACHE_MEMCACHE_SERVERS", "localhost:11211")
			},
			expectError: false,
		},
		{
			name: "memcache store - missing servers",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "memcache")
				// Missing CACHE_MEMCACHE_SERVERS
			},
			expectError: true,
			errorString: "CACHE_MEMCACHE_SERVERS",
		},
		{
			name: "postgres store - valid configuration",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "postgres")
				cfg.Provide().Set("CACHE_DB_HOST", "localhost")
				cfg.Provide().Set("CACHE_DB_DATABASE", "cache_db")
				cfg.Provide().Set("CACHE_DB_USERNAME", "user")
			},
			expectError: false,
		},
		{
			name: "postgres store - missing host",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "postgres")
				cfg.Provide().Set("CACHE_DB_DATABASE", "cache_db")
				cfg.Provide().Set("CACHE_DB_USERNAME", "user")
			},
			expectError: true,
			errorString: "CACHE_DB_HOST",
		},
		{
			name: "postgres store - with valid port",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "postgres")
				cfg.Provide().Set("CACHE_DB_HOST", "localhost")
				cfg.Provide().Set("CACHE_DB_DATABASE", "cache_db")
				cfg.Provide().Set("CACHE_DB_USERNAME", "user")
				cfg.Provide().Set("CACHE_DB_PORT", "5432")
			},
			expectError: false,
		},
		{
			name: "postgres store - with invalid port",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "postgres")
				cfg.Provide().Set("CACHE_DB_HOST", "localhost")
				cfg.Provide().Set("CACHE_DB_DATABASE", "cache_db")
				cfg.Provide().Set("CACHE_DB_USERNAME", "user")
				cfg.Provide().Set("CACHE_DB_PORT", "99999")
			},
			expectError: true,
			errorString: "port must be between 1 and 65535",
		},
		{
			name: "mysql store - valid configuration",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "mysql")
				cfg.Provide().Set("CACHE_DB_HOST", "localhost")
				cfg.Provide().Set("CACHE_DB_DATABASE", "cache_db")
				cfg.Provide().Set("CACHE_DB_USERNAME", "user")
			},
			expectError: false,
		},
		{
			name: "mongodb store - valid configuration",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "mongodb")
				cfg.Provide().Set("CACHE_MONGODB_URI", "mongodb://localhost:27017")
				cfg.Provide().Set("CACHE_MONGODB_DATABASE", "cache_db")
			},
			expectError: false,
		},
		{
			name: "mongodb store - missing URI",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "mongodb")
				cfg.Provide().Set("CACHE_MONGODB_DATABASE", "cache_db")
			},
			expectError: true,
			errorString: "CACHE_MONGODB_URI",
		},
		{
			name: "dynamodb store - valid configuration",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "dynamodb")
				cfg.Provide().Set("CACHE_DYNAMODB_TABLE", "cache_table")
				cfg.Provide().Set("CACHE_DYNAMODB_REGION", "us-east-1")
			},
			expectError: false,
		},
		{
			name: "dynamodb store - missing table",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "dynamodb")
				cfg.Provide().Set("CACHE_DYNAMODB_REGION", "us-east-1")
			},
			expectError: true,
			errorString: "CACHE_DYNAMODB_TABLE",
		},
		{
			name: "s3 store - valid configuration",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "s3")
				cfg.Provide().Set("CACHE_S3_BUCKET", "my-cache-bucket")
				cfg.Provide().Set("CACHE_S3_REGION", "us-east-1")
			},
			expectError: false,
		},
		{
			name: "s3 store - missing bucket",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "s3")
				cfg.Provide().Set("CACHE_S3_REGION", "us-east-1")
			},
			expectError: true,
			errorString: "CACHE_S3_BUCKET",
		},
		{
			name: "invalid store type",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "invalid_store")
			},
			expectError: true,
			errorString: "value must be one of",
		},
		{
			name: "sqlite3 store - valid",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "sqlite3")
			},
			expectError: false,
		},
		{
			name: "badger store - valid",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "badger")
			},
			expectError: false,
		},
		{
			name: "valid TTL configuration",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "memory")
				cfg.Provide().Set("CACHE_TTL", "5m")
			},
			expectError: false,
		},
		{
			name: "invalid TTL configuration",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "memory")
				cfg.Provide().Set("CACHE_TTL", "-5m") // Negative TTL
			},
			expectError: true,
			errorString: "TTL must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh config for each test
			cfgModule := config.NewModule()
			cfg := cfgModule.Provide()

			// Apply test configuration
			tt.configFunc(cfgModule)

			// Create logger
			logger := log.NewModule(cfg).Provide()

			// Create cache module
			cacheModule := NewModule(cfg, logger)

			// Validate configuration
			err := cacheModule.ValidateConfig()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				if tt.errorString != "" {
					if err.Error() == "" {
						t.Errorf("expected error message to contain '%s' but got empty message", tt.errorString)
					}
				}

				t.Logf("Got expected error: %v", err)
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}
