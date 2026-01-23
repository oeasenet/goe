package contract

import (
	"context"
	"time"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	// HealthStatusUp indicates the component is healthy
	HealthStatusUp HealthStatus = "up"
	// HealthStatusDown indicates the component is unhealthy
	HealthStatusDown HealthStatus = "down"
	// HealthStatusDegraded indicates the component is partially healthy
	HealthStatusDegraded HealthStatus = "degraded"
)

// HealthCheckResult contains the result of a single health check
type HealthCheckResult struct {
	// Status is the health status of the component
	Status HealthStatus `json:"status"`
	// Latency is the time taken to perform the health check
	Latency time.Duration `json:"latency_ms"`
	// Message provides additional details about the health status
	Message string `json:"message,omitempty"`
	// Timestamp is when the check was performed
	Timestamp time.Time `json:"timestamp"`
	// Metadata contains additional key-value pairs about the health check
	Metadata map[string]any `json:"metadata,omitempty"`
}

// HealthReport contains the overall health status and individual check results
type HealthReport struct {
	// Status is the overall health status
	Status HealthStatus `json:"status"`
	// Timestamp is when the report was generated
	Timestamp time.Time `json:"timestamp"`
	// Checks contains results for each registered health check
	Checks map[string]HealthCheckResult `json:"checks,omitempty"`
}

// HealthChecker defines the interface for individual health checks
type HealthChecker interface {
	// Name returns the unique name of this health checker
	Name() string

	// Check performs the health check and returns the result
	// The context can be used to set a timeout for the check
	Check(ctx context.Context) HealthCheckResult
}

// HealthManager manages health checks for the application
type HealthManager interface {
	// RegisterChecker adds a health checker to the manager
	RegisterChecker(checker HealthChecker)

	// UnregisterChecker removes a health checker by name
	UnregisterChecker(name string)

	// LivenessCheck returns a simple liveness status
	// This should be a fast check indicating if the process is alive
	// Used by Kubernetes liveness probes
	LivenessCheck(ctx context.Context) HealthReport

	// ReadinessCheck returns the readiness status
	// This checks if the application is ready to accept traffic
	// Used by Kubernetes readiness probes
	ReadinessCheck(ctx context.Context) HealthReport

	// HealthCheck returns a comprehensive health status
	// This runs all registered health checks
	HealthCheck(ctx context.Context) HealthReport

	// GetCheckers returns all registered health checkers
	GetCheckers() []HealthChecker
}

// HealthConfig defines configuration for the health module
type HealthConfig struct {
	// Enabled determines if health checks are active
	Enabled bool
	// Path is the base path for health endpoints (default: /health)
	Path string
	// LivenessPath is the path for liveness probe (default: /health/live)
	LivenessPath string
	// ReadinessPath is the path for readiness probe (default: /health/ready)
	ReadinessPath string
	// Timeout is the maximum time for a health check (default: 5s)
	Timeout time.Duration
}
