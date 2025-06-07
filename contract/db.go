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
	AutoMigrate(dst ...interface{}) error

	// AutoMigrateOnConnection performs auto migration for the given GORM models on a specific connection
	AutoMigrateOnConnection(connectionName string, dst ...interface{}) error
}
