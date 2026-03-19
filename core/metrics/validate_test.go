package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/zap"
)

// validateLogger implements contract.Logger for validation
type validateLogger struct{}

func (l *validateLogger) Debug(msg string, args ...any)                   {}
func (l *validateLogger) Info(msg string, args ...any)                    {}
func (l *validateLogger) Warn(msg string, args ...any)                    {}
func (l *validateLogger) Error(msg string, args ...any)                   {}
func (l *validateLogger) Fatal(msg string, args ...any)                   {}
func (l *validateLogger) Debugf(template string, args ...any)             {}
func (l *validateLogger) Infof(template string, args ...any)              {}
func (l *validateLogger) Warnf(template string, args ...any)              {}
func (l *validateLogger) Errorf(template string, args ...any)             {}
func (l *validateLogger) Fatalf(template string, args ...any)             {}
func (l *validateLogger) Debugw(msg string, keysAndValues ...any)         {}
func (l *validateLogger) Infow(msg string, keysAndValues ...any)          {}
func (l *validateLogger) Warnw(msg string, keysAndValues ...any)          {}
func (l *validateLogger) Errorw(msg string, keysAndValues ...any)         {}
func (l *validateLogger) Fatalw(msg string, keysAndValues ...any)         {}
func (l *validateLogger) With(keysAndValues ...any) contract.Logger       { return l }
func (l *validateLogger) WithContext(ctx context.Context) contract.Logger { return l }
func (l *validateLogger) WithError(err error) contract.Logger             { return l }
func (l *validateLogger) GetLogger() *zap.SugaredLogger                   { return nil }

// validateConfig implements contract.Config for validation
type validateConfig struct {
	values map[string]any
}

func (c *validateConfig) Get(key string) any                     { return c.values[key] }
func (c *validateConfig) GetString(key string) string            { v, _ := c.values[key].(string); return v }
func (c *validateConfig) GetInt(key string) int                  { v, _ := c.values[key].(int); return v }
func (c *validateConfig) GetInt64(key string) int64              { v, _ := c.values[key].(int64); return v }
func (c *validateConfig) GetBool(key string) bool                { v, _ := c.values[key].(bool); return v }
func (c *validateConfig) GetFloat64(key string) float64          { v, _ := c.values[key].(float64); return v }
func (c *validateConfig) GetDuration(key string) time.Duration   { return 0 }
func (c *validateConfig) GetStringSlice(key string) []string     { return nil }
func (c *validateConfig) GetStringMap(key string) map[string]any { return nil }
func (c *validateConfig) Set(key string, value any)              { c.values[key] = value }
func (c *validateConfig) Has(key string) bool                    { _, ok := c.values[key]; return ok }
func (c *validateConfig) All() map[string]any                    { return c.values }
func (c *validateConfig) Reload() error                          { return nil }

func TestMetricsValidation_FullOutput(t *testing.T) {
	cfg := &validateConfig{values: map[string]any{"APP_NAME": "test-app"}}
	logger := &validateLogger{}
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	// Create metrics with various types and labels
	counter := manager.Counter("http_requests_total", "Total HTTP requests", "method", "status")
	gauge := manager.Gauge("active_connections", "Active connections")
	histogram := manager.Histogram("request_duration_seconds", "Request duration", []float64{0.1, 0.5, 1.0, 5.0}, "endpoint")

	// Record some values
	counter.WithLabelValues("GET", "200").Inc()
	counter.WithLabelValues("GET", "200").Add(10)
	counter.WithLabelValues("POST", "201").Inc()
	counter.WithLabelValues("GET", "500").Add(3)

	gauge.Set(42)
	gauge.Inc()
	gauge.Dec()

	histogram.WithLabelValues("/api/users").Observe(0.25)
	histogram.WithLabelValues("/api/users").Observe(0.75)
	histogram.WithLabelValues("/api/products").Observe(2.5)

	// Get metrics output
	handler := manager.Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()

	// Log the output for inspection
	t.Logf("=== Metrics Output ===\n%s", body)

	// Validate counter metrics
	assert.Contains(t, body, "goe_http_requests_total", "Counter should be present")
	assert.Contains(t, body, `method="GET"`, "GET method label should be present")
	assert.Contains(t, body, `method="POST"`, "POST method label should be present")
	assert.Contains(t, body, `status="200"`, "200 status label should be present")
	assert.Contains(t, body, `status="201"`, "201 status label should be present")
	assert.Contains(t, body, `status="500"`, "500 status label should be present")

	// Validate gauge metrics
	assert.Contains(t, body, "goe_active_connections", "Gauge should be present")
	assert.Contains(t, body, "goe_active_connections 42", "Gauge should have value 42")

	// Validate histogram metrics
	assert.Contains(t, body, "goe_request_duration_seconds_bucket", "Histogram buckets should be present")
	assert.Contains(t, body, "goe_request_duration_seconds_sum", "Histogram sum should be present")
	assert.Contains(t, body, "goe_request_duration_seconds_count", "Histogram count should be present")
	assert.Contains(t, body, `endpoint="/api/users"`, "Endpoint label should be present")
	assert.Contains(t, body, `endpoint="/api/products"`, "Endpoint label should be present")

	// Validate HELP and TYPE comments (Prometheus format)
	assert.Contains(t, body, "# HELP goe_http_requests_total Total HTTP requests")
	assert.Contains(t, body, "# TYPE goe_http_requests_total counter")
	assert.Contains(t, body, "# HELP goe_active_connections Active connections")
	assert.Contains(t, body, "# TYPE goe_active_connections gauge")
	assert.Contains(t, body, "# HELP goe_request_duration_seconds Request duration")
	assert.Contains(t, body, "# TYPE goe_request_duration_seconds histogram")

	// Validate counter values
	lines := strings.Split(body, "\n")
	foundGetCounter := false
	for _, line := range lines {
		if strings.Contains(line, `goe_http_requests_total{method="GET",status="200"}`) {
			foundGetCounter = true
			assert.Contains(t, line, "11", "GET/200 counter should be 11 (1 + 10)")
		}
	}
	assert.True(t, foundGetCounter, "Should find GET/200 counter line")
}

func TestMetricsValidation_GoRuntimeMetrics(t *testing.T) {
	cfg := &validateConfig{values: map[string]any{
		"METRICS_GO_ENABLED":      true,
		"METRICS_PROCESS_ENABLED": true,
	}}
	logger := &validateLogger{}
	config := NewConfig(cfg)
	manager := NewManager(config, logger)

	handler := manager.Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Go runtime metrics should be present
	assert.Contains(t, body, "go_goroutines", "Go goroutines metric should be present")
	assert.Contains(t, body, "go_gc_duration_seconds", "Go GC duration metric should be present")
	assert.Contains(t, body, "go_memstats_alloc_bytes", "Go memory stats should be present")

	// Process metrics should be present
	assert.Contains(t, body, "process_cpu_seconds_total", "Process CPU metric should be present")
	assert.Contains(t, body, "process_resident_memory_bytes", "Process memory metric should be present")
}
