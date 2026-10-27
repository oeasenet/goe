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
	liveness map[string]contract.HealthChecker
	config   *Config
	logger   contract.Logger
	// stopping, when set and closed, makes readiness report down.
	stopping <-chan struct{}
}

// NewManager creates a new health manager
func NewManager(config *Config, logger contract.Logger) *Manager {
	return &Manager{
		checkers: make(map[string]contract.HealthChecker),
		liveness: make(map[string]contract.HealthChecker),
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

// RegisterLivenessChecker adds a checker to the liveness probe, which is
// otherwise always up. Keep it to conditions a restart fixes, such as a stalled
// control loop: a failing liveness probe gets the process restarted.
func (m *Manager) RegisterLivenessChecker(checker contract.HealthChecker) {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := checker.Name()
	m.liveness[name] = checker
	m.logger.Debug("Liveness checker registered", "checker", name)
}

// UnregisterChecker removes a health or liveness checker by name
func (m *Manager) UnregisterChecker(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.checkers, name)
	delete(m.liveness, name)
	m.logger.Debug("Health checker unregistered", "checker", name)
}

// LivenessCheck reports whether the process is alive: up unless a liveness
// checker reports down.
func (m *Manager) LivenessCheck(ctx context.Context) contract.HealthReport {
	return m.runChecks(ctx, m.snapshot(m.liveness))
}

// ReadinessCheck returns the readiness status by running all registered checks.
// Once the application is stopping it reports down without running them, so
// load balancers stop sending new requests while the server drains.
func (m *Manager) ReadinessCheck(ctx context.Context) contract.HealthReport {
	if m.isStopping() {
		now := time.Now()
		return contract.HealthReport{
			Status:    contract.HealthStatusDown,
			Timestamp: now,
			Checks: map[string]contract.HealthCheckResult{
				"shutdown": {Status: contract.HealthStatusDown, Message: "shutting down", Timestamp: now},
			},
		}
	}
	return m.runChecks(ctx, m.snapshot(m.checkers))
}

// HealthCheck returns a comprehensive health status with all check details
func (m *Manager) HealthCheck(ctx context.Context) contract.HealthReport {
	return m.runChecks(ctx, m.snapshot(m.checkers))
}

func (m *Manager) isStopping() bool {
	if m.stopping == nil {
		return false
	}
	select {
	case <-m.stopping:
		return true
	default:
		return false
	}
}

// snapshot copies a checker set under the lock.
func (m *Manager) snapshot(set map[string]contract.HealthChecker) map[string]contract.HealthChecker {
	m.mu.RLock()
	defer m.mu.RUnlock()
	checkers := make(map[string]contract.HealthChecker, len(set))
	maps.Copy(checkers, set)
	return checkers
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

// runChecks executes the given health checks concurrently
func (m *Manager) runChecks(ctx context.Context, checkers map[string]contract.HealthChecker) contract.HealthReport {
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
		wg.Go(func() {
			result := m.runSingleCheck(checkCtx, checker)

			resultsMu.Lock()
			results[name] = result
			resultsMu.Unlock()
		})
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
