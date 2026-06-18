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
			name: "redis store - no connection key needed (localhost defaults)",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "redis")
			},
			expectError: false,
		},
		{
			name: "redis store - with URL",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "redis")
				cfg.Provide().Set("CACHE_REDIS_URL", "redis://localhost:6379/0")
			},
			expectError: false,
		},
		{
			name: "redis store - valid database number (0 allowed)",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "redis")
				cfg.Provide().Set("CACHE_REDIS_DATABASE", "0")
			},
			expectError: false,
		},
		{
			name: "redis store - invalid database number",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "redis")
				cfg.Provide().Set("CACHE_REDIS_DATABASE", "-1")
			},
			expectError: true,
			errorString: "non-negative",
		},
		{
			name: "redis store - invalid port",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "redis")
				cfg.Provide().Set("CACHE_REDIS_PORT", "99999")
			},
			expectError: true,
			errorString: "port must be between 1 and 65535",
		},
		// Only memory and redis have registered drivers. Every other CACHE_STORE
		// value is now rejected by validation instead of silently using memory.
		// The postgres case is set up fully (host/db/user) to prove the rejection
		// comes from the store whitelist, not from a missing per-store key.
		{
			name: "postgres store - rejected (not implemented)",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_STORE", "postgres")
				cfg.Provide().Set("CACHE_DB_HOST", "localhost")
				cfg.Provide().Set("CACHE_DB_DATABASE", "cache_db")
				cfg.Provide().Set("CACHE_DB_USERNAME", "user")
			},
			expectError: true,
			errorString: "value must be one of",
		},
		{
			name:        "mysql store - rejected (not implemented)",
			configFunc:  func(cfg *config.Module) { cfg.Provide().Set("CACHE_STORE", "mysql") },
			expectError: true,
			errorString: "value must be one of",
		},
		{
			name:        "memcache store - rejected (not implemented)",
			configFunc:  func(cfg *config.Module) { cfg.Provide().Set("CACHE_STORE", "memcache") },
			expectError: true,
			errorString: "value must be one of",
		},
		{
			name:        "mongodb store - rejected (not implemented)",
			configFunc:  func(cfg *config.Module) { cfg.Provide().Set("CACHE_STORE", "mongodb") },
			expectError: true,
			errorString: "value must be one of",
		},
		{
			name:        "dynamodb store - rejected (not implemented)",
			configFunc:  func(cfg *config.Module) { cfg.Provide().Set("CACHE_STORE", "dynamodb") },
			expectError: true,
			errorString: "value must be one of",
		},
		{
			name:        "s3 store - rejected (not implemented)",
			configFunc:  func(cfg *config.Module) { cfg.Provide().Set("CACHE_STORE", "s3") },
			expectError: true,
			errorString: "value must be one of",
		},
		{
			name:        "sqlite3 store - rejected (not implemented)",
			configFunc:  func(cfg *config.Module) { cfg.Provide().Set("CACHE_STORE", "sqlite3") },
			expectError: true,
			errorString: "value must be one of",
		},
		{
			name:        "badger store - rejected (not implemented)",
			configFunc:  func(cfg *config.Module) { cfg.Provide().Set("CACHE_STORE", "badger") },
			expectError: true,
			errorString: "value must be one of",
		},
		{
			name:        "invalid store type - rejected",
			configFunc:  func(cfg *config.Module) { cfg.Provide().Set("CACHE_STORE", "invalid_store") },
			expectError: true,
			errorString: "value must be one of",
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
