package contract

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// MongoDB defines the clean, simple mongo database interface.
//
// GOE manages a single MongoDB connection. Applications that need a second
// data source construct their own mongo.Client via the driver — the framework
// deliberately does not multiplex connections.
type MongoDB interface {
	// Client returns the underlying MongoDB client for advanced operations
	Client() *mongo.Client

	// DB Instance returns the database instance
	DB() *mongo.Database

	// Col Collection returns a collection from the database
	Col(name string) *mongo.Collection
}
