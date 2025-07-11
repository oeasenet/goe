package observability

import (
	"context"

	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/validator"
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

// ValidateConfig validates the observability module configuration
func (m *Module) ValidateConfig() error {
	v := validator.NewConfigValidator(m.config, "observability")

	// Only validate if observability is enabled
	if m.config.GetBool("OTEL_ENABLED") {
		// Service name is required when enabled
		serviceName := m.config.GetString("OTEL_SERVICE_NAME")
		if serviceName == "" {
			serviceName = m.config.GetString("APP_NAME")
		}
		if serviceName == "" {
			v.Require("OTEL_SERVICE_NAME", "Service name for observability (or set APP_NAME)")
		}

		// If metrics are enabled, validate metrics-specific config
		if m.config.GetBool("OTEL_METRICS_ENABLED") {
			if m.config.Has("OTEL_METRICS_PORT") {
				v.Optional("OTEL_METRICS_PORT", "Metrics server port", validator.ValidatePort)
			}
		}

		// If tracing is enabled, validate tracing-specific config
		if m.config.GetBool("OTEL_TRACING_ENABLED") {
			// Endpoint is required for tracing
			endpoint := m.config.GetString("OTEL_TRACING_ENDPOINT")
			if endpoint == "" {
				endpoint = m.config.GetString("OTEL_EXPORTER_OTLP_ENDPOINT")
			}
			if endpoint == "" {
				v.Require("OTEL_TRACING_ENDPOINT", "Tracing endpoint (or set OTEL_EXPORTER_OTLP_ENDPOINT)")
			} else {
				v.Optional("OTEL_TRACING_ENDPOINT", "Tracing endpoint", validator.ValidateHostPort)
			}

			// Validate sampling ratio if set
			if m.config.Has("OTEL_TRACING_SAMPLING_RATIO") {
				v.Optional("OTEL_TRACING_SAMPLING_RATIO", "Sampling ratio (0.0-1.0)", func(value any) error {
					ratio := m.config.GetFloat64("OTEL_TRACING_SAMPLING_RATIO")
					if ratio < 0 || ratio > 1 {
						return validator.ValidateOneOf("value between 0.0 and 1.0")(value)
					}
					return nil
				})
			}
		}
	}

	return v.Validate()
}
