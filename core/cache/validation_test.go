package cache

import (
	"strings"
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
				// No configuration, should use the memory driver by default
			},
			expectError: false,
		},
		{
			name: "memory driver configuration",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_DRIVER", "memory")
			},
			expectError: false,
		},
		{
			name: "redis driver - no connection key needed (localhost defaults)",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_DRIVER", "redis")
			},
			expectError: false,
		},
		{
			name: "redis driver - with URL",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_DRIVER", "redis")
				cfg.Provide().Set("CACHE_REDIS_URL", "redis://localhost:6379/0")
			},
			expectError: false,
		},
		{
			name: "redis driver - valid database number (0 allowed)",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_DRIVER", "redis")
				cfg.Provide().Set("CACHE_REDIS_DATABASE", "0")
			},
			expectError: false,
		},
		{
			name: "redis driver - invalid database number",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_DRIVER", "redis")
				cfg.Provide().Set("CACHE_REDIS_DATABASE", "-1")
			},
			expectError: true,
			errorString: "non-negative",
		},
		{
			name: "redis driver - invalid port",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_DRIVER", "redis")
				cfg.Provide().Set("CACHE_REDIS_PORT", "99999")
			},
			expectError: true,
			errorString: "port must be between 1 and 65535",
		},
		// Only memory, redis, badger and bbolt drivers are registered. Every
		// other CACHE_DRIVER value is rejected by validation instead of
		// panicking on first use. The postgres case is set up fully
		// (host/db/user) to prove the rejection comes from the driver
		// registry, not from a missing key.
		{
			name: "postgres driver - rejected (not implemented)",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_DRIVER", "postgres")
				cfg.Provide().Set("CACHE_DB_HOST", "localhost")
				cfg.Provide().Set("CACHE_DB_DATABASE", "cache_db")
				cfg.Provide().Set("CACHE_DB_USERNAME", "user")
			},
			expectError: true,
			errorString: "is not registered",
		},
		{
			name:        "mysql driver - rejected (not implemented)",
			configFunc:  func(cfg *config.Module) { cfg.Provide().Set("CACHE_DRIVER", "mysql") },
			expectError: true,
			errorString: "is not registered",
		},
		{
			name:        "memcache driver - rejected (not implemented)",
			configFunc:  func(cfg *config.Module) { cfg.Provide().Set("CACHE_DRIVER", "memcache") },
			expectError: true,
			errorString: "is not registered",
		},
		{
			name:        "mongodb driver - rejected (not implemented)",
			configFunc:  func(cfg *config.Module) { cfg.Provide().Set("CACHE_DRIVER", "mongodb") },
			expectError: true,
			errorString: "is not registered",
		},
		{
			name:        "dynamodb driver - rejected (not implemented)",
			configFunc:  func(cfg *config.Module) { cfg.Provide().Set("CACHE_DRIVER", "dynamodb") },
			expectError: true,
			errorString: "is not registered",
		},
		{
			name:        "s3 driver - rejected (not implemented)",
			configFunc:  func(cfg *config.Module) { cfg.Provide().Set("CACHE_DRIVER", "s3") },
			expectError: true,
			errorString: "is not registered",
		},
		{
			name:        "sqlite3 driver - rejected (not implemented)",
			configFunc:  func(cfg *config.Module) { cfg.Provide().Set("CACHE_DRIVER", "sqlite3") },
			expectError: true,
			errorString: "is not registered",
		},
		{
			name:        "badger driver - accepted (embedded driver)",
			configFunc:  func(cfg *config.Module) { cfg.Provide().Set("CACHE_DRIVER", "badger") },
			expectError: false,
		},
		{
			name:        "bbolt driver - accepted (embedded driver)",
			configFunc:  func(cfg *config.Module) { cfg.Provide().Set("CACHE_DRIVER", "bbolt") },
			expectError: false,
		},
		{
			name:        "invalid driver - rejected",
			configFunc:  func(cfg *config.Module) { cfg.Provide().Set("CACHE_DRIVER", "invalid_store") },
			expectError: true,
			errorString: "is not registered",
		},
		{
			name: "valid TTL configuration",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_DRIVER", "memory")
				cfg.Provide().Set("CACHE_TTL", "5m")
			},
			expectError: false,
		},
		{
			name: "invalid TTL configuration",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_DRIVER", "memory")
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

				if tt.errorString != "" && !strings.Contains(err.Error(), tt.errorString) {
					t.Errorf("expected error message to contain %q but got: %v", tt.errorString, err)
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
