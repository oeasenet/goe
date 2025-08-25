package contract

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// MongoDB defines the clean, simple mongo database interface
type MongoDB interface {
	// Client returns the underlying MongoDB client for advanced operations
	Client() *mongo.Client

	// DB Instance returns the default database instance
	DB() *mongo.Database

	// Connection returns a specific database instance by connection name
	Connection(name string) (*mongo.Database, error)

	// Col Collection returns a collection from the default database
	Col(name string) *mongo.Collection

	// ColFrom CollectionFrom returns a collection from a specific database connection
	ColFrom(connectionName, collectionName string) (*mongo.Collection, error)
}
