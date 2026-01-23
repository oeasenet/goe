// Package main demonstrates GOE's production-ready features.
//
// This example shows:
// - Health checks with liveness and readiness probes (Kubernetes-compatible)
// - Prometheus metrics with custom counters, gauges, and histograms
// - OpenTelemetry distributed tracing
// - Graceful shutdown with priority-based hooks
// - Custom health checkers for external dependencies
//
// Run:
//
//	# Basic run
//	go run main.go
//
//	# With custom configuration
//	OTEL_EXPORTER_TYPE=stdout go run main.go
//
// Test endpoints:
//
//	# Health endpoints
//	curl http://localhost:3000/health        # Full health check
//	curl http://localhost:3000/health/live   # Kubernetes liveness probe
//	curl http://localhost:3000/health/ready  # Kubernetes readiness probe
//
//	# Metrics endpoint (Prometheus format)
//	curl http://localhost:3000/metrics
//
//	# Application endpoint
//	curl http://localhost:3000/api/hello
//	curl -X POST http://localhost:3000/api/process -H "Content-Type: application/json" -d '{"name":"test"}'
package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/health"
	"go.oease.dev/goe/v2/core/shutdown"
	"go.oease.dev/goe/v2/webresult"
)

func main() {
	// Create a new GOE application with production essentials enabled
	_ = goe.New(goe.Options{
		WithHTTP:    true, // HTTP server
		WithHealth:  true, // Health checks at /health, /health/live, /health/ready
		WithMetrics: true, // Prometheus metrics at /metrics
		WithOTel:    true, // OpenTelemetry tracing

		// Register application components
		Invokers: []any{
			registerRoutes,
			registerCustomHealthChecks,
			registerShutdownHooks,
		},
	})

	// Run blocks until shutdown (SIGINT/SIGTERM)
	goe.Run()
}

// registerRoutes sets up API routes with metrics and tracing
func registerRoutes(
	kernel contract.HTTPKernel,
	logger contract.Logger,
	metrics contract.MetricsManager,
) {
	app := kernel.App()

	// Create custom metrics for the application
	requestCounter := metrics.Counter("api_requests_total", "Total API requests", "method", "endpoint", "status")
	activeRequests := metrics.Gauge("api_active_requests", "Currently active requests")
	requestDuration := metrics.Histogram("api_request_duration_seconds", "API request duration",
		[]float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0}, "endpoint")

	// API routes with instrumentation
	app.Get("/api/hello", func(c fiber.Ctx) error {
		start := time.Now()
		activeRequests.Inc()
		defer func() {
			activeRequests.Dec()
			requestDuration.WithLabelValues("/api/hello").Observe(time.Since(start).Seconds())
		}()

		requestCounter.WithLabelValues("GET", "/api/hello", "200").Inc()
		return webresult.SendSucceed(c, fiber.Map{
			"message":   "Hello from production-ready GOE!",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	app.Post("/api/process", func(c fiber.Ctx) error {
		start := time.Now()
		activeRequests.Inc()
		defer func() {
			activeRequests.Dec()
			requestDuration.WithLabelValues("/api/process").Observe(time.Since(start).Seconds())
		}()

		// Simulate processing with variable latency
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

		var body struct {
			Name string `json:"name"`
		}
		if err := c.Bind().JSON(&body); err != nil {
			requestCounter.WithLabelValues("POST", "/api/process", "400").Inc()
			return webresult.SendFailed(c, "Invalid request body")
		}

		requestCounter.WithLabelValues("POST", "/api/process", "200").Inc()
		return webresult.SendSucceed(c, fiber.Map{
			"processed": body.Name,
			"duration":  time.Since(start).String(),
		})
	})

	logger.Info("API routes registered with metrics instrumentation")
}

// registerCustomHealthChecks adds application-specific health checks
func registerCustomHealthChecks(healthManager contract.HealthManager, logger contract.Logger) {
	// Register a custom health checker for external API dependency
	externalAPIChecker := health.NewCustomChecker("external-api", func(ctx context.Context) contract.HealthCheckResult {
		// Simulate checking an external API
		// In a real application, you would make an HTTP request here
		start := time.Now()

		// Simulate random health status for demo purposes
		isHealthy := rand.Float32() > 0.1 // 90% healthy

		result := contract.HealthCheckResult{
			Timestamp: time.Now(),
			Latency:   time.Since(start),
			Metadata: map[string]any{
				"endpoint": "https://api.example.com/health",
			},
		}

		if isHealthy {
			result.Status = contract.HealthStatusUp
			result.Message = "External API is responding"
		} else {
			result.Status = contract.HealthStatusDegraded
			result.Message = "External API is slow"
		}

		return result
	})

	healthManager.RegisterChecker(externalAPIChecker)

	// Register a custom checker for background job queue
	queueChecker := health.NewCustomChecker("job-queue", func(ctx context.Context) contract.HealthCheckResult {
		// Simulate checking a job queue
		return contract.HealthCheckResult{
			Status:    contract.HealthStatusUp,
			Message:   "Job queue is operational",
			Timestamp: time.Now(),
			Latency:   5 * time.Millisecond,
			Metadata: map[string]any{
				"pending_jobs": 42,
				"workers":      4,
			},
		}
	})

	healthManager.RegisterChecker(queueChecker)

	logger.Info("Custom health checkers registered", "count", 2)
}

// registerShutdownHooks adds application-specific cleanup logic
func registerShutdownHooks(shutdownManager *shutdown.Manager, logger contract.Logger) {
	// Register a hook to flush pending metrics before shutdown
	shutdownManager.RegisterHook("flush-metrics", shutdown.PriorityTelemetry, func(ctx context.Context) error {
		logger.Info("Flushing pending metrics...")
		// In a real app, you might call: metricsManager.Flush(ctx)
		time.Sleep(100 * time.Millisecond) // Simulate flush
		return nil
	})

	// Register a hook to gracefully finish pending jobs
	shutdownManager.RegisterHook("finish-jobs", shutdown.PriorityWorkers, func(ctx context.Context) error {
		logger.Info("Waiting for pending jobs to complete...")
		// In a real app, you would wait for job workers to finish
		time.Sleep(200 * time.Millisecond) // Simulate job completion
		return nil
	})

	// Register a hook to notify external services about shutdown
	shutdownManager.RegisterHook("notify-services", shutdown.PriorityHTTP-10, func(ctx context.Context) error {
		logger.Info("Notifying external services about shutdown...")
		// In a real app, you might deregister from service discovery
		return nil
	})

	logger.Info("Shutdown hooks registered",
		"hooks", []string{"flush-metrics", "finish-jobs", "notify-services"})
}
