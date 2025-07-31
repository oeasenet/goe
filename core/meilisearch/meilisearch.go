package meilisearch

import (
	"context"
	"fmt"
	"github.com/meilisearch/meilisearch-go"
	"go.oease.dev/goe/v2/contract"
	"strings"
	"sync"
)

// MeiliSearch implements the contract.Meilisearch and contract.Module interfaces
type MeiliSearch struct {
	logger      contract.Logger
	config      contract.Config
	mu          sync.RWMutex
	connections map[string]meilisearch.ServiceManager
}

// NewMeiliSearch creates a new MeiliSearch instance
func NewMeiliSearch(config contract.Config, logger contract.Logger) *MeiliSearch {
	return &MeiliSearch{
		logger:      logger,
		config:      config,
		connections: make(map[string]meilisearch.ServiceManager),
	}
}

// Instance returns the underlying meilisearch instance for the default connection
func (ms *MeiliSearch) Instance() meilisearch.ServiceManager {
	defaultConnectionName := ms.config.GetString("MEILISEARCH_CONNECTION")
	if defaultConnectionName == "" {
		defaultConnectionName = "default"
	}

	conn, err := ms.Connection(defaultConnectionName)
	if err != nil {
		ms.logger.Error("Failed to get default meilisearch instance",
			"connection_name", defaultConnectionName,
			"error", err.Error(),
		)
		return nil
	}
	return conn
}

// Connection returns a specific meilisearch instance by name
func (ms *MeiliSearch) Connection(name string) (meilisearch.ServiceManager, error) {
	ms.mu.RLock()
	conn, ok := ms.connections[name]
	ms.mu.RUnlock()

	if !ok {
		// Decision: Do not connect on-demand here. Connections should be explicitly defined and set up OnStart.
		// If a connection is requested that wasn't configured/failed, it's an error.
		return nil, fmt.Errorf("meilisearch connection '%s' not found or not configured", name)
	}
	return conn, nil
}

// --- contract.Module interface implementation ---

// Name returns the unique name of the module
func (ms *MeiliSearch) Name() string {
	return "meilisearch"
}

// OnStart is called when the module starts
// This is where database connections will be established
func (ms *MeiliSearch) OnStart(ctx context.Context) error {
	ms.logger.Info("MONGO Database module OnStart")
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Get default connection name
	defaultConnectionName := ms.config.GetString("MEILISEARCH_CONNECTION")
	if defaultConnectionName == "" {
		defaultConnectionName = "default"
	}

	// Connect to default database
	ms.logger.Info("Attempting to connect to default meilisearch",
		"connection_config_name", defaultConnectionName,
	)
	sm, err := ms.connect(defaultConnectionName)
	if err != nil {
		ms.logger.Error("Failed to connect to default meilisearch",
			"connection_config_name", defaultConnectionName,
			"error", err.Error(),
		)
		// Allow app to start, Instance() will return nil.
	} else {
		// Store the connection using the name it will be requested by, which is defaultConnectionName.
		ms.connections[defaultConnectionName] = sm
		ms.logger.Info("Successfully connected to default meilisearch",
			"connection_config_name", defaultConnectionName,
		)
	}

	// Connect to additional meilisearch if configured
	connectionsList := ms.config.GetString("MEILISEARCH_CONNECTIONS")
	if connectionsList != "" {
		// Split the comma-separated list of connection names
		connectionNames := strings.Split(connectionsList, ",")
		for _, connName := range connectionNames {
			connName = strings.TrimSpace(connName)

			// Skip if it's the default connection (already connected)
			if connName == defaultConnectionName {
				continue
			}

			// Skip if empty
			if connName == "" {
				continue
			}

			ms.logger.Info("Attempting to connect to additional meilisearch",
				"connection_name", connName,
			)
			conn, err := ms.connect(connName)
			if err != nil {
				ms.logger.Error("Failed to connect to additional meilisearch",
					"connection_name", connName,
					"error", err.Error(),
				)
				// Continue with other connections
			} else {
				ms.connections[connName] = conn
				ms.logger.Info("Successfully connected to additional meilisearch",
					"connection_name", connName,
				)

			}
		}
	}

	return nil
}

// OnStop is called when the module stops
// This is where database connections will be closed
func (ms *MeiliSearch) OnStop(ctx context.Context) error {
	ms.logger.Info("Mongo Database module OnStop")
	ms.mu.Lock()
	defer ms.mu.Unlock()

	var lastErr error
	for name, conn := range ms.connections {
		ms.logger.Info("Closing meilisearch connection", "connection", name)
		conn.Close()
		delete(ms.connections, name)
	}
	return lastErr
}

// Provide returns the Meilisearch instance for Fx
// This will allow injecting contract.Meilisearch
func (ms *MeiliSearch) Provide() contract.Meilisearch {
	return ms
}
