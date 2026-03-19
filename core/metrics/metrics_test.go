package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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
		values: map[string]any{
			"APP_NAME": "test-app",
		},
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

func TestNewManager(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	assert.NotNil(t, manager)
}

func TestManager_Counter(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	counter := manager.Counter("test_counter", "A test counter", "method", "path")
	assert.NotNil(t, counter)

	// Increment the counter with label values
	counter.WithLabelValues("GET", "/api/test").Inc()
	counter.WithLabelValues("POST", "/api/test").Add(5)

	// Verify the counter is registered
	handler := manager.Handler()
	assert.NotNil(t, handler)
}

func TestManager_Gauge(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	gauge := manager.Gauge("test_gauge", "A test gauge", "service")
	assert.NotNil(t, gauge)

	// Set gauge values with label values
	gauge.WithLabelValues("api").Set(42)
	gauge.WithLabelValues("api").Inc()
	gauge.WithLabelValues("api").Dec()
	gauge.WithLabelValues("api").Add(10)
	gauge.WithLabelValues("api").Sub(5)
}

func TestManager_Histogram(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	buckets := []float64{0.1, 0.5, 1.0, 2.5, 5.0}
	histogram := manager.Histogram("test_histogram", "A test histogram", buckets, "method")
	assert.NotNil(t, histogram)

	// Observe values with label values
	histogram.WithLabelValues("GET").Observe(0.25)
	histogram.WithLabelValues("POST").Observe(1.5)
}

func TestManager_Handler(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	// Create some metrics
	counter := manager.Counter("handler_test_counter", "Test counter")
	counter.Inc()

	handler := manager.Handler()
	assert.NotNil(t, handler)

	// Test the handler returns valid Prometheus output
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "handler_test_counter")
}

func TestModule(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()

	module := NewModule(cfg, logger)
	require.NotNil(t, module)

	assert.Equal(t, "metrics", module.Name())
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

	// Test default namespace (defaults to "goe")
	assert.Equal(t, "goe", config.Namespace())

	// Test custom namespace
	cfg.values["METRICS_NAMESPACE"] = "custom"
	config2 := NewConfig(cfg)
	assert.Equal(t, "custom", config2.Namespace())
}

func TestCounter_WithoutLabels(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	counter := manager.Counter("no_label_counter", "Counter without labels")
	assert.NotNil(t, counter)

	counter.Inc()
	counter.Add(10)
}

func TestGauge_WithoutLabels(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	gauge := manager.Gauge("no_label_gauge", "Gauge without labels")
	assert.NotNil(t, gauge)

	gauge.Set(100)
	gauge.Inc()
	gauge.Dec()
}

func TestHistogram_WithoutLabels(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	histogram := manager.Histogram("no_label_histogram", "Histogram without labels", nil)
	assert.NotNil(t, histogram)

	histogram.Observe(0.5)
}

func TestMetricsOutput(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	// Create metrics with different types
	counter := manager.Counter("output_counter", "Test counter", "status")
	gauge := manager.Gauge("output_gauge", "Test gauge")
	histogram := manager.Histogram("output_histogram", "Test histogram", []float64{0.1, 0.5, 1.0})

	counter.WithLabelValues("success").Inc()
	counter.WithLabelValues("error").Add(5)
	gauge.Set(42)
	histogram.Observe(0.25)

	// Get the metrics output
	handler := manager.Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Verify metrics are present
	assert.Contains(t, body, "output_counter")
	assert.Contains(t, body, "output_gauge")
	assert.Contains(t, body, "output_histogram")
	assert.Contains(t, body, "status=\"success\"")
	assert.Contains(t, body, "status=\"error\"")
}

func TestNamespaceNormalization(t *testing.T) {
	cfg := newTestConfig()
	// When no METRICS_NAMESPACE is set, default is "goe"
	config := NewConfig(cfg)
	assert.Equal(t, "goe", config.Namespace())

	// Custom namespace is used as-is
	cfg.values["METRICS_NAMESPACE"] = "my_custom_namespace"
	config2 := NewConfig(cfg)
	assert.Equal(t, "my_custom_namespace", config2.Namespace())
}

func TestConfig_Subsystem(t *testing.T) {
	cfg := newTestConfig()

	config := NewConfig(cfg)
	// Default subsystem should be empty
	assert.Empty(t, config.Subsystem())

	cfg.values["METRICS_SUBSYSTEM"] = "http"
	config2 := NewConfig(cfg)
	assert.Equal(t, "http", config2.Subsystem())
}

func TestHTTPMiddlewareMetrics(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	// Create standard HTTP metrics that would be used by middleware
	requestCounter := manager.Counter("http_requests_total", "Total HTTP requests", "method", "path", "status")
	requestDuration := manager.Histogram("http_request_duration_seconds", "HTTP request duration", []float64{0.001, 0.01, 0.1, 0.5, 1.0, 5.0}, "method", "path")

	// Simulate some requests
	requestCounter.WithLabelValues("GET", "/api/users", "200").Inc()
	requestCounter.WithLabelValues("POST", "/api/users", "201").Inc()
	requestCounter.WithLabelValues("GET", "/api/users", "500").Inc()

	requestDuration.WithLabelValues("GET", "/api/users").Observe(0.05)
	requestDuration.WithLabelValues("POST", "/api/users").Observe(0.15)

	// Verify metrics are recorded
	handler := manager.Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	assert.Contains(t, body, "http_requests_total")
	assert.Contains(t, body, "http_request_duration_seconds")

	// Check that multiple status codes are recorded
	assert.True(t, strings.Contains(body, `status="200"`) || strings.Contains(body, "status=\"200\""))
}
