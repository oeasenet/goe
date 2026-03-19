package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/fx"
)

// Module represents the health module for Fx
type Module struct {
	manager *Manager
	config  *Config
	logger  contract.Logger
}

// NewModule creates a new health module
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
	return "health"
}

// OnStart is called when the module starts
func (m *Module) OnStart(ctx context.Context) error {
	m.logger.Info("Health module started",
		"path", m.config.Path(),
		"liveness", m.config.LivenessPath(),
		"readiness", m.config.ReadinessPath(),
	)
	return nil
}

// OnStop is called when the module stops
func (m *Module) OnStop(ctx context.Context) error {
	m.logger.Debug("Health module stopped")
	return nil
}

// Manager returns the health manager
func (m *Module) Manager() contract.HealthManager {
	return m.manager
}

// RegisterRoutes registers health check routes on the Fiber app
func (m *Module) RegisterRoutes(app *fiber.App) {
	// Main health endpoint - comprehensive check
	app.Get(m.config.Path(), m.healthHandler)

	// Liveness probe - simple check that process is alive
	app.Get(m.config.LivenessPath(), m.livenessHandler)

	// Readiness probe - check if ready to accept traffic
	app.Get(m.config.ReadinessPath(), m.readinessHandler)
}

// healthHandler handles the main health endpoint
func (m *Module) healthHandler(c fiber.Ctx) error {
	ctx := c.Context()
	report := m.manager.HealthCheck(ctx)
	return m.sendHealthResponse(c, report)
}

// livenessHandler handles the liveness probe endpoint
func (m *Module) livenessHandler(c fiber.Ctx) error {
	ctx := c.Context()
	report := m.manager.LivenessCheck(ctx)
	return m.sendHealthResponse(c, report)
}

// readinessHandler handles the readiness probe endpoint
func (m *Module) readinessHandler(c fiber.Ctx) error {
	ctx := c.Context()
	report := m.manager.ReadinessCheck(ctx)
	return m.sendHealthResponse(c, report)
}

// sendHealthResponse sends the health report as JSON with appropriate status code
func (m *Module) sendHealthResponse(c fiber.Ctx, report contract.HealthReport) error {
	statusCode := http.StatusOK
	switch report.Status {
	case contract.HealthStatusDown:
		statusCode = http.StatusServiceUnavailable
	case contract.HealthStatusDegraded:
		statusCode = http.StatusOK // Still return 200 for degraded, just different status in body
	}

	// Convert latencies to milliseconds for JSON output
	response := healthResponse{
		Status:    string(report.Status),
		Timestamp: report.Timestamp.Format(time.RFC3339),
	}

	if len(report.Checks) > 0 {
		response.Checks = make(map[string]checkResponse, len(report.Checks))
		for name, check := range report.Checks {
			response.Checks[name] = checkResponse{
				Status:    string(check.Status),
				LatencyMs: float64(check.Latency.Microseconds()) / 1000.0,
				Message:   check.Message,
			}
		}
	}

	return c.Status(statusCode).JSON(response)
}

// healthResponse is the JSON response structure for health endpoints
type healthResponse struct {
	Status    string                   `json:"status"`
	Timestamp string                   `json:"timestamp"`
	Checks    map[string]checkResponse `json:"checks,omitempty"`
}

// checkResponse is the JSON response structure for individual checks
type checkResponse struct {
	Status    string  `json:"status"`
	LatencyMs float64 `json:"latency_ms"`
	Message   string  `json:"message,omitempty"`
}

// ModuleParams contains dependencies for the health module
type ModuleParams struct {
	fx.In

	Config contract.Config
	Logger contract.Logger
}

// ModuleResult contains the outputs provided by the health module
type ModuleResult struct {
	fx.Out

	Module  *Module `group:"modules"`
	Manager contract.HealthManager
}

// ProvideModule creates the Fx provider for the health module
func ProvideModule(params ModuleParams) ModuleResult {
	module := NewModule(params.Config, params.Logger)
	return ModuleResult{
		Module:  module,
		Manager: module.Manager(),
	}
}

// FxModule returns the Fx module options for the health module
func FxModule() fx.Option {
	return fx.Module("health",
		fx.Provide(ProvideModule),
	)
}
