package contract

import (
	"github.com/meilisearch/meilisearch-go"
)

// Meilisearch defines the meilisearch interface
type Meilisearch interface {
	// Instance returns the underlying Meilisearch instance for the default connection
	Instance() meilisearch.ServiceManager

	// Connection returns a specific Meilisearch instance by name
	Connection(name string) (meilisearch.ServiceManager, error)
}
