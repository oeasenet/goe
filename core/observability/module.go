package observability

import (
	"context"

	"go.oease.dev/goe/v2/contract"
)

// Module represents the observability module for Fx
type Module struct {
	observability contract.Observability
	logger        contract.Logger
	config        contract.Config
}

// NewModule creates a new observability module
func NewModule(config contract.Config, logger contract.Logger) *Module {
	// Create observability instance
	obs := New(config, logger)

	return &Module{
		observability: obs,
		logger:        logger,
		config:        config,
	}
}

// Name returns the module name
func (m *Module) Name() string {
	return "observability"
}

// OnStart is called when the module starts
func (m *Module) OnStart(ctx context.Context) error {
	enabled := m.config.GetBool("OTEL_ENABLED")
	if !enabled {
		m.logger.Info("Observability module started (disabled)")
		return nil
	}

	serviceName := m.config.GetString("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = m.config.GetString("APP_NAME")
		if serviceName == "" {
			serviceName = "goe-app"
		}
	}

	m.logger.Info("Observability module started",
		"service_name", serviceName,
		"metrics_enabled", m.config.GetBool("OTEL_METRICS_ENABLED"),
		"tracing_enabled", m.config.GetBool("OTEL_TRACING_ENABLED"),
	)
	return nil
}

// OnStop is called when the module stops
func (m *Module) OnStop(ctx context.Context) error {
	m.logger.Info("Observability module stopping")

	// Shutdown observability components
	if err := m.observability.Shutdown(ctx); err != nil {
		m.logger.Error("Failed to shutdown observability", "error", err)
		return err
	}

	m.logger.Info("Observability module stopped")
	return nil
}

// Provide returns the observability instance for Fx
func (m *Module) Provide() contract.Observability {
	return m.observability
}

// ProvideMetrics returns the metrics manager for Fx
func (m *Module) ProvideMetrics() contract.MetricsManager {
	return m.observability.Metrics()
}

// ProvideTracing returns the tracing manager for Fx
func (m *Module) ProvideTracing() contract.TracingManager {
	return m.observability.Tracing()
}
