package contract

import (
	"gorm.io/gorm"
)

// DB defines the database interface
type DB interface {
	// Instance returns the underlying GORM DB instance for the default connection
	Instance() *gorm.DB

	// Connection returns a specific GORM DB instance by name
	// This allows for multiple database connections if needed in the future
	Connection(name string) (*gorm.DB, error)

	// AutoMigrate performs auto migration for the given GORM models on the default connection
	// Note: This requires database connections to be established first
	AutoMigrate(dst ...any) error

	// AutoMigrateOnConnection performs auto migration for the given GORM models on a specific connection
	// Note: This requires database connections to be established first
	AutoMigrateOnConnection(connectionName string, dst ...any) error

	// RegisterModelsForMigration pre-registers models for automatic migration on the default connection
	// These models will be migrated automatically during startup if auto-migration is enabled
	RegisterModelsForMigration(dst ...any)

	// RegisterModelsForMigrationOnConnection pre-registers models for automatic migration on a specific connection
	// These models will be migrated automatically during startup if auto-migration is enabled for that connection
	RegisterModelsForMigrationOnConnection(connectionName string, dst ...any)
}

type DBModel struct {
	ID        string         `gorm:"primaryKey"`
	CreatedAt int64          `gorm:"autoCreateTime:milli"`
	UpdatedAt int64          `gorm:"autoUpdateTime:milli"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
