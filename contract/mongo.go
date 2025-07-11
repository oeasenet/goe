package contract

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// MongoDB defines the mongo database interface
type MongoDB interface {
	// Instance returns the underlying MongoDB instance for the default connection
	Instance() *mongo.Database

	// Connection returns a specific MongoDB instance by name
	Connection(name string) (*mongo.Database, error)
}
