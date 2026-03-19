package health

import (
	"context"
	"maps"
	"sync"
	"time"

	"go.oease.dev/goe/v2/contract"
)

// Manager implements the HealthManager interface
type Manager struct {
	mu       sync.RWMutex
	checkers map[string]contract.HealthChecker
	config   *Config
	logger   contract.Logger
}

// NewManager creates a new health manager
func NewManager(config *Config, logger contract.Logger) *Manager {
	return &Manager{
		checkers: make(map[string]contract.HealthChecker),
		config:   config,
		logger:   logger,
	}
}

// RegisterChecker adds a health checker to the manager
func (m *Manager) RegisterChecker(checker contract.HealthChecker) {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := checker.Name()
	m.checkers[name] = checker
	m.logger.Debug("Health checker registered", "checker", name)
}

// UnregisterChecker removes a health checker by name
func (m *Manager) UnregisterChecker(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.checkers, name)
	m.logger.Debug("Health checker unregistered", "checker", name)
}

// LivenessCheck returns a simple liveness status
// This is a fast check indicating if the process is alive
func (m *Manager) LivenessCheck(ctx context.Context) contract.HealthReport {
	return contract.HealthReport{
		Status:    contract.HealthStatusUp,
		Timestamp: time.Now(),
	}
}

// ReadinessCheck returns the readiness status by running all registered checks
func (m *Manager) ReadinessCheck(ctx context.Context) contract.HealthReport {
	return m.runChecks(ctx)
}

// HealthCheck returns a comprehensive health status with all check details
func (m *Manager) HealthCheck(ctx context.Context) contract.HealthReport {
	return m.runChecks(ctx)
}

// GetCheckers returns all registered health checkers
func (m *Manager) GetCheckers() []contract.HealthChecker {
	m.mu.RLock()
	defer m.mu.RUnlock()

	checkers := make([]contract.HealthChecker, 0, len(m.checkers))
	for _, checker := range m.checkers {
		checkers = append(checkers, checker)
	}
	return checkers
}

// runChecks executes all registered health checks concurrently
func (m *Manager) runChecks(ctx context.Context) contract.HealthReport {
	m.mu.RLock()
	checkers := make(map[string]contract.HealthChecker, len(m.checkers))
	maps.Copy(checkers, m.checkers)
	m.mu.RUnlock()

	// If no checkers registered, return up status
	if len(checkers) == 0 {
		return contract.HealthReport{
			Status:    contract.HealthStatusUp,
			Timestamp: time.Now(),
		}
	}

	// Create timeout context if not already set
	timeout := m.config.Timeout()
	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Run all checks concurrently
	results := make(map[string]contract.HealthCheckResult, len(checkers))
	var resultsMu sync.Mutex
	var wg sync.WaitGroup

	for name, checker := range checkers {
		wg.Add(1)
		go func(name string, checker contract.HealthChecker) {
			defer wg.Done()

			result := m.runSingleCheck(checkCtx, checker)

			resultsMu.Lock()
			results[name] = result
			resultsMu.Unlock()
		}(name, checker)
	}

	wg.Wait()

	// Determine overall status
	overallStatus := contract.HealthStatusUp
	for _, result := range results {
		if result.Status == contract.HealthStatusDown {
			overallStatus = contract.HealthStatusDown
			break
		}
		if result.Status == contract.HealthStatusDegraded && overallStatus != contract.HealthStatusDown {
			overallStatus = contract.HealthStatusDegraded
		}
	}

	return contract.HealthReport{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Checks:    results,
	}
}

// runSingleCheck runs a single health check with panic recovery
func (m *Manager) runSingleCheck(ctx context.Context, checker contract.HealthChecker) (result contract.HealthCheckResult) {
	start := time.Now()

	// Recover from panics in health checks
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("Health check panicked",
				"checker", checker.Name(),
				"panic", r,
			)
			result = contract.HealthCheckResult{
				Status:    contract.HealthStatusDown,
				Latency:   time.Since(start),
				Message:   "check panicked",
				Timestamp: time.Now(),
			}
		}
	}()

	// Run the check
	result = checker.Check(ctx)
	if result.Latency == 0 {
		result.Latency = time.Since(start)
	}
	if result.Timestamp.IsZero() {
		result.Timestamp = time.Now()
	}

	return result
}
