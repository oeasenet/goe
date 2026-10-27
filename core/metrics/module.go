package metrics

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/fx"
)

// Module represents the metrics module for Fx
type Module struct {
	manager  *Manager
	config   *Config
	logger   contract.Logger
	settings settings
	optErrs  []error
	server   *http.Server
}

// NewModule creates a new metrics module. Options are applied after the
// METRICS_* environment variables; an invalid one fails OnStart.
func NewModule(cfg contract.Config, logger contract.Logger, opts ...Option) *Module {
	config := NewConfig(cfg)

	var s settings
	s.applyEnv(cfg)
	var optErrs []error
	for i, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&s); err != nil {
			optErrs = append(optErrs, fmt.Errorf("metrics option %d: %w", i+1, err))
		}
	}

	registry := s.registry
	if registry == nil {
		registry = prometheus.NewRegistry()
	}

	return &Module{
		manager:  newManager(config, logger, registry),
		config:   config,
		logger:   logger.With("module", "metrics"),
		settings: s,
		optErrs:  optErrs,
	}
}

// Name returns the module name
func (m *Module) Name() string {
	return "metrics"
}

// OnStart starts the separate metrics listener when a port is configured.
// It fails when an option was invalid or the port cannot be bound.
func (m *Module) OnStart(ctx context.Context) error {
	if len(m.optErrs) > 0 {
		return errors.Join(m.optErrs...)
	}
	port := m.settings.port
	if port < 0 || port > 65535 {
		return fmt.Errorf("metrics: port %d is out of range 1-65535", port)
	}
	if port == 0 {
		if m.settings.pprof {
			m.logger.Warn("Metrics pprof needs a separate port (METRICS_PORT); not served")
		}
		m.logger.Info("Metrics module started",
			"path", m.config.Path(),
			"namespace", m.config.Namespace(),
		)
		return nil
	}

	mux := http.NewServeMux()
	mux.Handle(m.config.Path(), m.manager.Handler())
	if m.settings.pprof {
		registerPprof(mux)
	}
	addr := net.JoinHostPort(m.settings.host, strconv.Itoa(port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("metrics: listening on %s: %w", addr, err)
	}
	m.server = &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = m.server.Serve(ln) }()

	m.logger.Info("Metrics server started",
		"address", ln.Addr().String(),
		"path", m.config.Path(),
		"namespace", m.config.Namespace(),
		"pprof", m.settings.pprof,
	)
	return nil
}

// OnStop stops the separate metrics listener, if any. A scrape still running
// when ctx ends is cut off; that never fails the stop.
func (m *Module) OnStop(ctx context.Context) error {
	if m.server != nil {
		if err := m.server.Shutdown(ctx); err != nil {
			_ = m.server.Close()
		}
	}
	m.logger.Debug("Metrics module stopped")
	return nil
}

// Manager returns the metrics manager
func (m *Module) Manager() contract.MetricsManager {
	return m.manager
}

// RegisterRoutes registers metrics routes on the Fiber app, unless metrics
// are served on their own port.
func (m *Module) RegisterRoutes(app *fiber.App) {
	if m.settings.port != 0 {
		return
	}
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
