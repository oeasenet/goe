package contract

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// MongoDB defines the mongo database interface
type MongoDB interface {
	// Instance returns the underlying MONGO DB instance for the default connection
	Instance() *mongo.Database

	// Connection returns a specific MONGO DB instance by name
	// This allows for multiple database connections if needed in the future
	Connection(name string) (*mongo.Database, error)

	// SetMonitor sets a custom CommandMonitor for the DatabaseModule.
	SetMonitor(monitor *event.CommandMonitor)

	// IsNoDocumentsError checks if the given error is mongo.ErrNoDocuments.
	IsNoDocumentsError(err error) bool

	// Ctx creates and returns a new context with a 10-second timeout.
	Ctx() context.Context

	// AutoMigrate performs auto migration for the given GORM models on the default connection
	//AutoMigrate(dst ...interface{}) error

	// AutoMigrateOnConnection performs auto migration for the given GORM models on a specific connection
	//AutoMigrateOnConnection(connectionName string, dst ...interface{}) error
}
