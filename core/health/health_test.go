package health

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/zap"
)

// testLogger implements contract.Logger for testing
type testLogger struct {
	t *testing.T
}

func (l *testLogger) Debug(msg string, args ...any) {
	if l.t != nil {
		l.t.Logf("[DEBUG] %s %v", msg, args)
	}
}
func (l *testLogger) Info(msg string, args ...any) {
	if l.t != nil {
		l.t.Logf("[INFO] %s %v", msg, args)
	}
}
func (l *testLogger) Warn(msg string, args ...any) {
	if l.t != nil {
		l.t.Logf("[WARN] %s %v", msg, args)
	}
}
func (l *testLogger) Error(msg string, args ...any) {
	if l.t != nil {
		l.t.Logf("[ERROR] %s %v", msg, args)
	}
}
func (l *testLogger) Fatal(msg string, args ...any) {
	if l.t != nil {
		l.t.Fatalf("[FATAL] %s %v", msg, args)
	}
}
func (l *testLogger) Debugf(template string, args ...any) {
	if l.t != nil {
		l.t.Logf("[DEBUG] "+template, args...)
	}
}
func (l *testLogger) Infof(template string, args ...any) {
	if l.t != nil {
		l.t.Logf("[INFO] "+template, args...)
	}
}
func (l *testLogger) Warnf(template string, args ...any) {
	if l.t != nil {
		l.t.Logf("[WARN] "+template, args...)
	}
}
func (l *testLogger) Errorf(template string, args ...any) {
	if l.t != nil {
		l.t.Logf("[ERROR] "+template, args...)
	}
}
func (l *testLogger) Fatalf(template string, args ...any) {
	if l.t != nil {
		l.t.Fatalf("[FATAL] "+template, args...)
	}
}
func (l *testLogger) Debugw(msg string, keysAndValues ...any) {
	if l.t != nil {
		l.t.Logf("[DEBUG] %s %v", msg, keysAndValues)
	}
}
func (l *testLogger) Infow(msg string, keysAndValues ...any) {
	if l.t != nil {
		l.t.Logf("[INFO] %s %v", msg, keysAndValues)
	}
}
func (l *testLogger) Warnw(msg string, keysAndValues ...any) {
	if l.t != nil {
		l.t.Logf("[WARN] %s %v", msg, keysAndValues)
	}
}
func (l *testLogger) Errorw(msg string, keysAndValues ...any) {
	if l.t != nil {
		l.t.Logf("[ERROR] %s %v", msg, keysAndValues)
	}
}
func (l *testLogger) Fatalw(msg string, keysAndValues ...any) {
	if l.t != nil {
		l.t.Fatalf("[FATAL] %s %v", msg, keysAndValues)
	}
}
func (l *testLogger) With(keysAndValues ...any) contract.Logger {
	return l
}
func (l *testLogger) WithContext(ctx context.Context) contract.Logger {
	return l
}
func (l *testLogger) WithError(err error) contract.Logger {
	return l
}
func (l *testLogger) GetLogger() *zap.SugaredLogger {
	return nil
}

// testConfig is a simple test config implementation
type testConfig struct {
	values map[string]any
}

func newTestConfig() *testConfig {
	return &testConfig{
		values: make(map[string]any),
	}
}

func (c *testConfig) Get(key string) any {
	return c.values[key]
}

func (c *testConfig) GetString(key string) string {
	if v, ok := c.values[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (c *testConfig) GetInt(key string) int {
	if v, ok := c.values[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return 0
}

func (c *testConfig) GetInt64(key string) int64 {
	if v, ok := c.values[key]; ok {
		if i, ok := v.(int64); ok {
			return i
		}
	}
	return 0
}

func (c *testConfig) GetBool(key string) bool {
	if v, ok := c.values[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func (c *testConfig) GetDuration(key string) time.Duration {
	if v, ok := c.values[key]; ok {
		if d, ok := v.(time.Duration); ok {
			return d
		}
	}
	return 0
}

func (c *testConfig) GetFloat64(key string) float64 {
	if v, ok := c.values[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

func (c *testConfig) GetStringSlice(key string) []string {
	if v, ok := c.values[key]; ok {
		if s, ok := v.([]string); ok {
			return s
		}
	}
	return nil
}

func (c *testConfig) GetStringMap(key string) map[string]any {
	if v, ok := c.values[key]; ok {
		if m, ok := v.(map[string]any); ok {
			return m
		}
	}
	return nil
}

func (c *testConfig) Set(key string, value any) {
	c.values[key] = value
}

func (c *testConfig) Has(key string) bool {
	_, ok := c.values[key]
	return ok
}

func (c *testConfig) All() map[string]any {
	return c.values
}

func (c *testConfig) Reload() error {
	return nil
}

// Mock health checker for testing
type mockChecker struct {
	name   string
	status contract.HealthStatus
	err    error
}

func (m *mockChecker) Name() string {
	return m.name
}

func (m *mockChecker) Check(ctx context.Context) contract.HealthCheckResult {
	result := contract.HealthCheckResult{
		Status:    m.status,
		Timestamp: time.Now(),
		Latency:   10 * time.Millisecond,
	}
	if m.err != nil {
		result.Message = m.err.Error()
	}
	return result
}

func TestNewManager(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	assert.NotNil(t, manager)
	assert.Empty(t, manager.GetCheckers())
}

func TestManager_RegisterChecker(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	checker := &mockChecker{name: "test", status: contract.HealthStatusUp}
	manager.RegisterChecker(checker)

	checkers := manager.GetCheckers()
	assert.Len(t, checkers, 1)
	assert.Equal(t, "test", checkers[0].Name())
}

func TestManager_UnregisterChecker(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	checker1 := &mockChecker{name: "checker1", status: contract.HealthStatusUp}
	checker2 := &mockChecker{name: "checker2", status: contract.HealthStatusUp}

	manager.RegisterChecker(checker1)
	manager.RegisterChecker(checker2)
	assert.Len(t, manager.GetCheckers(), 2)

	manager.UnregisterChecker("checker1")
	checkers := manager.GetCheckers()
	assert.Len(t, checkers, 1)
	assert.Equal(t, "checker2", checkers[0].Name())
}

func TestManager_HealthCheck_AllUp(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	manager.RegisterChecker(&mockChecker{name: "db", status: contract.HealthStatusUp})
	manager.RegisterChecker(&mockChecker{name: "cache", status: contract.HealthStatusUp})

	ctx := context.Background()
	report := manager.HealthCheck(ctx)

	assert.Equal(t, contract.HealthStatusUp, report.Status)
	assert.Len(t, report.Checks, 2)
	assert.True(t, report.Checks["db"].Status == contract.HealthStatusUp)
	assert.True(t, report.Checks["cache"].Status == contract.HealthStatusUp)
}

func TestManager_HealthCheck_OneDown(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	manager.RegisterChecker(&mockChecker{name: "db", status: contract.HealthStatusUp})
	manager.RegisterChecker(&mockChecker{name: "cache", status: contract.HealthStatusDown, err: errors.New("connection failed")})

	ctx := context.Background()
	report := manager.HealthCheck(ctx)

	assert.Equal(t, contract.HealthStatusDown, report.Status)
	assert.Len(t, report.Checks, 2)
	assert.True(t, report.Checks["db"].Status == contract.HealthStatusUp)
	assert.True(t, report.Checks["cache"].Status == contract.HealthStatusDown)
}

func TestManager_HealthCheck_OneDegraded(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	manager.RegisterChecker(&mockChecker{name: "db", status: contract.HealthStatusUp})
	manager.RegisterChecker(&mockChecker{name: "cache", status: contract.HealthStatusDegraded})

	ctx := context.Background()
	report := manager.HealthCheck(ctx)

	// When one is degraded but none down, overall should be degraded
	assert.Equal(t, contract.HealthStatusDegraded, report.Status)
}

func TestManager_LivenessCheck(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	// Liveness should always be up if the service is running
	ctx := context.Background()
	report := manager.LivenessCheck(ctx)

	assert.Equal(t, contract.HealthStatusUp, report.Status)
}

func TestManager_ReadinessCheck(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	manager.RegisterChecker(&mockChecker{name: "db", status: contract.HealthStatusUp})
	manager.RegisterChecker(&mockChecker{name: "cache", status: contract.HealthStatusUp})

	ctx := context.Background()
	report := manager.ReadinessCheck(ctx)

	assert.Equal(t, contract.HealthStatusUp, report.Status)
	assert.Len(t, report.Checks, 2)
}

func TestManager_HealthCheck_NoCheckers(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	ctx := context.Background()
	report := manager.HealthCheck(ctx)

	assert.Equal(t, contract.HealthStatusUp, report.Status)
	assert.Empty(t, report.Checks)
}

func TestNewCustomChecker(t *testing.T) {
	checker := NewCustomChecker("custom", func(ctx context.Context) contract.HealthCheckResult {
		return contract.HealthCheckResult{
			Status:    contract.HealthStatusUp,
			Timestamp: time.Now(),
		}
	})

	assert.Equal(t, "custom", checker.Name())

	ctx := context.Background()
	result := checker.Check(ctx)
	assert.Equal(t, contract.HealthStatusUp, result.Status)
}

func TestHealthStatus_String(t *testing.T) {
	assert.Equal(t, "up", string(contract.HealthStatusUp))
	assert.Equal(t, "down", string(contract.HealthStatusDown))
	assert.Equal(t, "degraded", string(contract.HealthStatusDegraded))
}

func TestModule(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()

	module := NewModule(cfg, logger)
	require.NotNil(t, module)

	assert.Equal(t, "health", module.Name())
	assert.NotNil(t, module.Manager())

	// Test lifecycle
	ctx := context.Background()
	err := module.OnStart(ctx)
	assert.NoError(t, err)

	err = module.OnStop(ctx)
	assert.NoError(t, err)
}

func TestConfig(t *testing.T) {
	cfg := newTestConfig()

	config := NewConfig(cfg)
	require.NotNil(t, config)

	// Test default timeout
	assert.Equal(t, 5*time.Second, config.Timeout())

	// Test custom timeout
	cfg.values["HEALTH_TIMEOUT"] = 10 * time.Second
	assert.Equal(t, 10*time.Second, config.Timeout())
}

func TestManager_PanicRecovery(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	// Create a checker that panics
	panicChecker := NewCustomChecker("panic-checker", func(ctx context.Context) contract.HealthCheckResult {
		panic("test panic")
	})

	manager.RegisterChecker(panicChecker)

	ctx := context.Background()
	report := manager.HealthCheck(ctx)

	// Should not panic and should return down status
	assert.Equal(t, contract.HealthStatusDown, report.Status)
	assert.Equal(t, contract.HealthStatusDown, report.Checks["panic-checker"].Status)
	assert.Equal(t, "check panicked", report.Checks["panic-checker"].Message)
}
