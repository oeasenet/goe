package db

import (
	"testing"

	"go.oease.dev/goe/v2/core/config"
	"go.oease.dev/goe/v2/core/log"
)

func TestDatabaseModuleValidation(t *testing.T) {
	tests := []struct {
		name        string
		configFunc  func(*config.Module)
		expectError bool
		errorString string
	}{
		{
			name: "no configuration - should require driver",
			configFunc: func(cfg *config.Module) {
				// No configuration
			},
			expectError: true,
			errorString: "DB_DRIVER",
		},
		{
			name: "valid MySQL configuration",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("DB_DRIVER", "mysql")
				cfg.Provide().Set("DB_HOST", "localhost")
				cfg.Provide().Set("DB_DATABASE", "test")
				cfg.Provide().Set("DB_USERNAME", "user")
				cfg.Provide().Set("DB_PASSWORD", "pass")
			},
			expectError: false,
		},
		{
			name: "valid PostgreSQL configuration",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("DB_DRIVER", "postgres")
				cfg.Provide().Set("DB_HOST", "localhost")
				cfg.Provide().Set("DB_DATABASE", "test")
				cfg.Provide().Set("DB_USERNAME", "user")
				cfg.Provide().Set("DB_PASSWORD", "pass")
			},
			expectError: false,
		},
		{
			name: "valid SQLite configuration",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("DB_DRIVER", "sqlite")
				cfg.Provide().Set("DB_DATABASE", "test.db")
			},
			expectError: false,
		},
		{
			name: "SQLite with minimal configuration",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("DB_DRIVER", "sqlite")
				// SQLite doesn't require HOST, DATABASE can be empty for in-memory
			},
			expectError: false,
		},
		{
			name: "invalid driver",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("DB_DRIVER", "invalid")
			},
			expectError: true,
			errorString: "unsupported database driver",
		},
		{
			name: "MySQL missing host",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("DB_DRIVER", "mysql")
				cfg.Provide().Set("DB_DATABASE", "test")
			},
			expectError: true,
			errorString: "DB_HOST",
		},
		{
			name: "MySQL missing database",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("DB_DRIVER", "mysql")
				cfg.Provide().Set("DB_HOST", "localhost")
			},
			expectError: true,
			errorString: "DB_DATABASE",
		},
		{
			name: "MySQL with invalid port",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("DB_DRIVER", "mysql")
				cfg.Provide().Set("DB_HOST", "localhost")
				cfg.Provide().Set("DB_DATABASE", "test")
				cfg.Provide().Set("DB_PORT", "99999") // Invalid port
			},
			expectError: true,
			errorString: "port must be between 1 and 65535",
		},
		{
			name: "PostgreSQL with valid port",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("DB_DRIVER", "postgres")
				cfg.Provide().Set("DB_HOST", "localhost")
				cfg.Provide().Set("DB_DATABASE", "test")
				cfg.Provide().Set("DB_PORT", "5432")
			},
			expectError: false,
		},
		{
			name: "multiple connections - valid",
			configFunc: func(cfg *config.Module) {
				// Default connection
				cfg.Provide().Set("DB_DRIVER", "mysql")
				cfg.Provide().Set("DB_HOST", "localhost")
				cfg.Provide().Set("DB_DATABASE", "test")

				// Additional connections
				cfg.Provide().Set("DB_CONNECTIONS", "secondary,tertiary")
				cfg.Provide().Set("DB_SECONDARY_DRIVER", "postgres")
				cfg.Provide().Set("DB_SECONDARY_HOST", "localhost")
				cfg.Provide().Set("DB_SECONDARY_DATABASE", "test2")
				cfg.Provide().Set("DB_TERTIARY_DRIVER", "sqlite")
				cfg.Provide().Set("DB_TERTIARY_DATABASE", "test3.db")
			},
			expectError: false,
		},
		{
			name: "multiple connections - invalid secondary",
			configFunc: func(cfg *config.Module) {
				// Default connection
				cfg.Provide().Set("DB_DRIVER", "mysql")
				cfg.Provide().Set("DB_HOST", "localhost")
				cfg.Provide().Set("DB_DATABASE", "test")

				// Invalid secondary connection
				cfg.Provide().Set("DB_CONNECTIONS", "secondary")
				cfg.Provide().Set("DB_SECONDARY_DRIVER", "postgres")
				// Missing host for postgres
			},
			expectError: true,
			errorString: "DB_SECONDARY_HOST",
		},
		{
			name: "named default connection",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("DB_CONNECTION", "primary")
				cfg.Provide().Set("DB_PRIMARY_DRIVER", "mysql")
				cfg.Provide().Set("DB_PRIMARY_HOST", "localhost")
				cfg.Provide().Set("DB_PRIMARY_DATABASE", "test")
			},
			expectError: false,
		},
		{
			name: "named default connection missing config",
			configFunc: func(cfg *config.Module) {
				cfg.Provide().Set("DB_CONNECTION", "primary")
				cfg.Provide().Set("DB_PRIMARY_DRIVER", "mysql")
				// Missing host
			},
			expectError: true,
			errorString: "DB_PRIMARY_HOST",
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

			// Create database module
			dbModule := NewDBModule(cfg, logger)

			// Validate configuration
			err := dbModule.ValidateConfig()

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
