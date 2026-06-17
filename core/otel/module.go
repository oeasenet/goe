package otel

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/fx"
)

// Module represents the OpenTelemetry module for Fx
type Module struct {
	provider *Provider
	config   *Config
	logger   contract.Logger
}

// NewModule creates a new OpenTelemetry module
func NewModule(ctx context.Context, cfg contract.Config, logger contract.Logger) (*Module, error) {
	config := NewConfig(cfg)

	// Create provider
	provider, err := NewProvider(ctx, config, logger)
	if err != nil {
		return nil, err
	}

	return &Module{
		provider: provider,
		config:   config,
		logger:   logger.With("module", "otel"),
	}, nil
}

// Name returns the module name
func (m *Module) Name() string {
	return "otel"
}

// OnStart is called when the module starts
func (m *Module) OnStart(ctx context.Context) error {
	m.logger.Info("OpenTelemetry module started",
		"service", m.config.ServiceName(),
		"exporter", m.config.ExporterType(),
	)
	return nil
}

// OnStop is called when the module stops
func (m *Module) OnStop(ctx context.Context) error {
	m.logger.Debug("Shutting down OpenTelemetry module")
	return m.provider.Shutdown(ctx)
}

// Provider returns the OpenTelemetry provider
func (m *Module) Provider() contract.OTelProvider {
	return m.provider
}

// RegisterMiddleware registers the OpenTelemetry tracing middleware
func (m *Module) RegisterMiddleware(app *fiber.App) {
	cfg := DefaultMiddlewareConfig(m.config.ServiceName())
	cfg.Tracer = m.provider.Tracer(m.config.ServiceName())
	cfg.Propagators = m.provider.TextMapPropagator()

	// Skip health and metrics endpoints from tracing
	cfg.Skip = func(c fiber.Ctx) bool {
		path := c.Path()
		return path == "/health" ||
			path == "/health/live" ||
			path == "/health/ready" ||
			path == "/metrics"
	}

	app.Use(Middleware(cfg))
}

// ModuleParams contains dependencies for the OpenTelemetry module
type ModuleParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Config    contract.Config
	Logger    contract.Logger
}

// ModuleResult contains the outputs provided by the OpenTelemetry module
type ModuleResult struct {
	fx.Out

	Module   *Module `group:"modules"`
	Provider contract.OTelProvider
}

// ProvideModule creates the Fx provider for the OpenTelemetry module
func ProvideModule(params ModuleParams) (ModuleResult, error) {
	ctx := context.Background()
	module, err := NewModule(ctx, params.Config, params.Logger)
	if err != nil {
		return ModuleResult{}, err
	}

	return ModuleResult{
		Module:   module,
		Provider: module.Provider(),
	}, nil
}

// FxModule returns the Fx module options for the OpenTelemetry module
func FxModule() fx.Option {
	return fx.Module("otel",
		fx.Provide(ProvideModule),
	)
}
