// Package main demonstrates creating custom GOE modules with lifecycle hooks.
//
// This example shows:
// - Creating custom modules implementing contract.Module
// - Module lifecycle hooks (OnStart, OnStop)
// - Configuration validation with ConfigValidator
// - Module dependency injection
// - Background workers with graceful shutdown
// - Health check patterns
//
// Run:
//
//	go run .
//
// Test:
//
//	curl http://localhost:3000/health
//	curl http://localhost:3000/metrics
//	curl http://localhost:3000/stats
package main

import (
	"context"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/webresult"
	"go.uber.org/fx"
)

func main() {
	_ = goe.New(goe.Options{
		WithHTTP: true,

		// Custom modules are registered as providers
		// They can provide services to other parts of the application
		Modules: []any{
			ProvideMetricsModule,
			ProvideHealthModule,
		},

		Invokers: []any{
			RegisterRoutes,
		},
	})

	goe.Run()
}

// RegisterRoutes sets up HTTP endpoints.
func RegisterRoutes(
	kernel contract.HTTPKernel,
	metrics *MetricsModule,
	health *HealthModule,
	logger contract.Logger,
) {
	app := kernel.App()

	// Metrics endpoint
	app.Get("/metrics", func(c fiber.Ctx) error {
		return webresult.SendSucceed(c, metrics.GetMetrics())
	})

	// Stats endpoint
	app.Get("/stats", func(c fiber.Ctx) error {
		return webresult.SendSucceed(c, fiber.Map{
			"requests_total":  metrics.GetMetrics().RequestCount,
			"uptime_seconds":  time.Since(metrics.startTime).Seconds(),
			"collectors":      len(health.checkers),
			"last_collection": metrics.GetMetrics().LastCollected,
		})
	})

	// Health check endpoint
	app.Get("/health", func(c fiber.Ctx) error {
		results := health.RunChecks(c.Context())
		allHealthy := true
		for _, r := range results {
			if !r.Healthy {
				allHealthy = false
				break
			}
		}

		response := fiber.Map{
			"status": "healthy",
			"checks": results,
		}

		if !allHealthy {
			response["status"] = "unhealthy"
			c.Status(fiber.StatusServiceUnavailable)
		}

		return c.JSON(response)
	})

	// Demo endpoint that records metrics
	app.Get("/demo", func(c fiber.Ctx) error {
		metrics.RecordRequest()
		return webresult.SendSucceed(c, fiber.Map{
			"message": "Request recorded",
			"count":   metrics.GetMetrics().RequestCount,
		})
	})

	logger.Info("Custom module routes registered")
}

// =============================================================================
// Metrics Module - Demonstrates background worker with periodic collection
// =============================================================================

// Metrics represents collected metrics.
type Metrics struct {
	RequestCount  int64     `json:"request_count"`
	LastCollected time.Time `json:"last_collected"`
	CollectCount  int64     `json:"collect_count"`
}

// MetricsModule is a custom module that collects and exposes metrics.
type MetricsModule struct {
	config    contract.Config
	logger    contract.Logger
	startTime time.Time
	metrics   Metrics
	mu        sync.RWMutex
	stopCh    chan struct{}
	wg        sync.WaitGroup
}

// ProvideMetricsModule creates and registers the MetricsModule.
func ProvideMetricsModule(lc fx.Lifecycle, config contract.Config, logger contract.Logger) *MetricsModule {
	m := &MetricsModule{
		config:    config,
		logger:    logger,
		startTime: time.Now(),
		stopCh:    make(chan struct{}),
	}

	// Register lifecycle hooks with Fx
	lc.Append(fx.Hook{
		OnStart: m.OnStart,
		OnStop:  m.OnStop,
	})

	return m
}

// Name returns the module name.
func (m *MetricsModule) Name() string {
	return "metrics"
}

// ValidateConfig validates the module configuration.
func (m *MetricsModule) ValidateConfig() error {
	// Validate required configuration
	interval := m.config.GetDuration("METRICS_COLLECT_INTERVAL")
	if interval <= 0 {
		// Use default if not set
		m.logger.Info("METRICS_COLLECT_INTERVAL not set, using default 30s")
	}
	return nil
}

// OnStart is called when the module starts.
func (m *MetricsModule) OnStart(ctx context.Context) error {
	// Validate configuration
	if err := m.ValidateConfig(); err != nil {
		return err
	}

	m.logger.Info("Metrics module starting")

	// Get collection interval from config
	interval := m.config.GetDuration("METRICS_COLLECT_INTERVAL")
	if interval <= 0 {
		interval = 30 * time.Second
	}

	// Start background collection worker
	m.wg.Add(1)
	go m.collectLoop(interval)

	m.logger.Infow("Metrics module started",
		"collect_interval", interval.String(),
	)

	return nil
}

// OnStop is called when the module stops.
func (m *MetricsModule) OnStop(ctx context.Context) error {
	m.logger.Info("Metrics module stopping")

	// Signal the background worker to stop
	close(m.stopCh)

	// Wait for the worker to finish with timeout
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.logger.Info("Metrics module stopped gracefully")
	case <-ctx.Done():
		m.logger.Warn("Metrics module stop timed out")
	}

	return nil
}

// collectLoop runs the periodic metrics collection.
func (m *MetricsModule) collectLoop(interval time.Duration) {
	defer m.wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.collect()
		}
	}
}

// collect performs a metrics collection.
func (m *MetricsModule) collect() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics.CollectCount++
	m.metrics.LastCollected = time.Now()

	m.logger.Debugw("Metrics collected",
		"collect_count", m.metrics.CollectCount,
		"request_count", m.metrics.RequestCount,
	)
}

// RecordRequest increments the request counter.
func (m *MetricsModule) RecordRequest() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics.RequestCount++
}

// GetMetrics returns the current metrics.
func (m *MetricsModule) GetMetrics() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metrics
}

// =============================================================================
// Health Module - Demonstrates health check aggregation
// =============================================================================

// HealthChecker is a function that performs a health check.
type HealthChecker func(ctx context.Context) error

// HealthCheckResult represents the result of a health check.
type HealthCheckResult struct {
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency"`
}

// HealthModule aggregates and runs health checks.
type HealthModule struct {
	config   contract.Config
	logger   contract.Logger
	checkers map[string]HealthChecker
	mu       sync.RWMutex
}

// ProvideHealthModule creates and registers the HealthModule.
func ProvideHealthModule(lc fx.Lifecycle, config contract.Config, logger contract.Logger) *HealthModule {
	h := &HealthModule{
		config:   config,
		logger:   logger,
		checkers: make(map[string]HealthChecker),
	}

	// Register lifecycle hooks
	lc.Append(fx.Hook{
		OnStart: h.OnStart,
		OnStop:  h.OnStop,
	})

	return h
}

// Name returns the module name.
func (h *HealthModule) Name() string {
	return "health"
}

// OnStart is called when the module starts.
func (h *HealthModule) OnStart(ctx context.Context) error {
	h.logger.Info("Health module starting")

	// Register built-in health checks
	h.RegisterChecker("self", func(ctx context.Context) error {
		// Self check always passes
		return nil
	})

	h.logger.Infow("Health module started",
		"checkers", len(h.checkers),
	)

	return nil
}

// OnStop is called when the module stops.
func (h *HealthModule) OnStop(ctx context.Context) error {
	h.logger.Info("Health module stopped")
	return nil
}

// RegisterChecker adds a health checker.
func (h *HealthModule) RegisterChecker(name string, checker HealthChecker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers[name] = checker
	h.logger.Infow("Health checker registered", "name", name)
}

// RunChecks executes all health checks.
func (h *HealthModule) RunChecks(ctx context.Context) []HealthCheckResult {
	h.mu.RLock()
	defer h.mu.RUnlock()

	results := make([]HealthCheckResult, 0, len(h.checkers))

	for name, checker := range h.checkers {
		start := time.Now()
		err := checker(ctx)
		latency := time.Since(start)

		result := HealthCheckResult{
			Name:    name,
			Healthy: err == nil,
			Latency: latency.String(),
		}

		if err != nil {
			result.Message = err.Error()
		}

		results = append(results, result)
	}

	return results
}
