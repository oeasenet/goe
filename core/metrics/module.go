package metrics

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/fx"
)

// Module represents the metrics module for Fx
type Module struct {
	manager *Manager
	config  *Config
	logger  contract.Logger
}

// NewModule creates a new metrics module
func NewModule(cfg contract.Config, logger contract.Logger) *Module {
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	return &Module{
		manager: manager,
		config:  config,
		logger:  logger,
	}
}

// Name returns the module name
func (m *Module) Name() string {
	return "metrics"
}

// OnStart is called when the module starts
func (m *Module) OnStart(ctx context.Context) error {
	m.logger.Info("Metrics module started",
		"path", m.config.Path(),
		"namespace", m.config.Namespace(),
	)
	return nil
}

// OnStop is called when the module stops
func (m *Module) OnStop(ctx context.Context) error {
	m.logger.Debug("Metrics module stopped")
	return nil
}

// Manager returns the metrics manager
func (m *Module) Manager() contract.MetricsManager {
	return m.manager
}

// RegisterRoutes registers metrics routes on the Fiber app
func (m *Module) RegisterRoutes(app *fiber.App) {
	// Adapt the Prometheus HTTP handler to Fiber
	handler := fasthttpadaptor.NewFastHTTPHandler(m.manager.Handler())
	app.Get(m.config.Path(), func(c fiber.Ctx) error {
		handler(c.RequestCtx())
		return nil
	})
}

// RegisterMiddleware registers the HTTP metrics middleware
func (m *Module) RegisterMiddleware(app *fiber.App) {
	app.Use(Middleware(m.manager))
}

// adaptHandler adapts an http.Handler to a Fiber handler
func adaptHandler(h http.Handler) fiber.Handler {
	fastHandler := fasthttpadaptor.NewFastHTTPHandler(h)
	return func(c fiber.Ctx) error {
		fastHandler(c.RequestCtx())
		return nil
	}
}

// ModuleParams contains dependencies for the metrics module
type ModuleParams struct {
	fx.In

	Config contract.Config
	Logger contract.Logger
}

// ModuleResult contains the outputs provided by the metrics module
type ModuleResult struct {
	fx.Out

	Module  *Module `group:"modules"`
	Manager contract.MetricsManager
}

// ProvideModule creates the Fx provider for the metrics module
func ProvideModule(params ModuleParams) ModuleResult {
	module := NewModule(params.Config, params.Logger)
	return ModuleResult{
		Module:  module,
		Manager: module.Manager(),
	}
}

// FxModule returns the Fx module options for the metrics module
func FxModule() fx.Option {
	return fx.Module("metrics",
		fx.Provide(ProvideModule),
	)
}
