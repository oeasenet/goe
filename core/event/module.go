package event

import (
	"context"

	"go.oease.dev/goe/v2/contract"
)

// Module represents the event module for Fx
type Module struct {
	manager contract.EventManager
	logger  contract.Logger
	config  contract.Config
}

// NewModule creates a new event module
func NewModule(config contract.Config, logger contract.Logger) (*Module, error) {
	// Load event configuration
	eventConfig := LoadConfig(config)

	// Create Redis manager
	manager, err := NewRedisManager(eventConfig, logger)
	if err != nil {
		return nil, err
	}

	return &Module{
		manager: manager,
		logger:  logger,
		config:  config,
	}, nil
}

// Name returns the module name
func (m *Module) Name() string {
	return "event"
}

// OnStart is called when the module starts
func (m *Module) OnStart(ctx context.Context) error {
	// Test connection
	if err := m.manager.Health(ctx); err != nil {
		return err
	}

	m.logger.Info("Event module started",
		"redis_addr", m.config.GetString("EVENT_REDIS_ADDR"),
		"redis_db", m.config.GetInt("EVENT_REDIS_DB"),
	)

	return nil
}

// OnStop is called when the module stops
func (m *Module) OnStop(ctx context.Context) error {
	m.logger.Info("Event module stopping")

	// Close event manager
	if err := m.manager.Close(ctx); err != nil {
		m.logger.Error("Error closing event manager", "error", err)
		return err
	}

	m.logger.Info("Event module stopped")
	return nil
}

// Provide returns the event manager instance for Fx
func (m *Module) Provide() contract.EventManager {
	return m.manager
}

// ProvideEventPublisher returns the event publisher for Fx
func (m *Module) ProvideEventPublisher() contract.EventPublisher {
	return m.manager
}

// ProvideEventConsumer returns the event consumer for Fx
func (m *Module) ProvideEventConsumer() contract.EventConsumer {
	return m.manager
}

// ProvideDeadLetterQueue returns the dead letter queue manager for Fx
func (m *Module) ProvideDeadLetterQueue() contract.DeadLetterQueueManager {
	return m.manager.GetDeadLetterQueue()
}

// ProvideEventManagerWithMetrics returns the event manager wrapped with metrics if observability is available
func ProvideEventManagerWithMetrics(
	manager contract.EventManager,
	metrics contract.MetricsManager,
	tracing contract.TracingManager,
) contract.EventManager {
	// TODO: Implement metrics wrapper when observability API is stable
	return manager
}
