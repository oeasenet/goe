package otel

import (
	"context"
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
	values map[string]interface{}
}

func newTestConfig() *testConfig {
	return &testConfig{
		values: map[string]interface{}{
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

// Config Tests

func TestNewConfig(t *testing.T) {
	cfg := newTestConfig()
	config := NewConfig(cfg)
	require.NotNil(t, config)
}

func TestConfig_Enabled(t *testing.T) {
	cfg := newTestConfig()
	config := NewConfig(cfg)

	// Default is enabled
	assert.True(t, config.Enabled())

	// Explicitly disabled
	cfg.values["OTEL_ENABLED"] = false
	config2 := NewConfig(cfg)
	assert.False(t, config2.Enabled())

	// Explicitly enabled
	cfg.values["OTEL_ENABLED"] = true
	config3 := NewConfig(cfg)
	assert.True(t, config3.Enabled())
}

func TestConfig_ServiceName(t *testing.T) {
	cfg := newTestConfig()

	// Use APP_NAME as fallback
	config := NewConfig(cfg)
	assert.Equal(t, "test-app", config.ServiceName())

	// Use OTEL_SERVICE_NAME if set
	cfg.values["OTEL_SERVICE_NAME"] = "otel-service"
	config2 := NewConfig(cfg)
	assert.Equal(t, "otel-service", config2.ServiceName())

	// Default when both are empty
	cfg2 := &testConfig{values: map[string]interface{}{}}
	config3 := NewConfig(cfg2)
	assert.Equal(t, "goe-app", config3.ServiceName())
}

func TestConfig_ServiceVersion(t *testing.T) {
	cfg := newTestConfig()
	config := NewConfig(cfg)

	// Default version
	assert.Equal(t, "1.0.0", config.ServiceVersion())

	// Use APP_VERSION
	cfg.values["APP_VERSION"] = "2.0.0"
	config2 := NewConfig(cfg)
	assert.Equal(t, "2.0.0", config2.ServiceVersion())

	// OTEL_SERVICE_VERSION takes precedence
	cfg.values["OTEL_SERVICE_VERSION"] = "3.0.0"
	config3 := NewConfig(cfg)
	assert.Equal(t, "3.0.0", config3.ServiceVersion())
}

func TestConfig_TracesEnabled(t *testing.T) {
	cfg := newTestConfig()
	config := NewConfig(cfg)

	// Default is enabled
	assert.True(t, config.TracesEnabled())

	// Can be disabled
	cfg.values["OTEL_TRACES_ENABLED"] = false
	config2 := NewConfig(cfg)
	assert.False(t, config2.TracesEnabled())
}

func TestConfig_MetricsEnabled(t *testing.T) {
	cfg := newTestConfig()
	config := NewConfig(cfg)

	// Default is disabled (use Prometheus instead)
	assert.False(t, config.MetricsEnabled())

	// Can be enabled
	cfg.values["OTEL_METRICS_ENABLED"] = true
	config2 := NewConfig(cfg)
	assert.True(t, config2.MetricsEnabled())
}

func TestConfig_TraceSampler(t *testing.T) {
	cfg := newTestConfig()
	config := NewConfig(cfg)

	// Default sampler
	assert.Equal(t, "parentbased_traceidratio", config.TraceSampler())

	// Custom sampler
	cfg.values["OTEL_TRACES_SAMPLER"] = "always_on"
	config2 := NewConfig(cfg)
	assert.Equal(t, "always_on", config2.TraceSampler())
}

func TestConfig_TraceSamplerArg(t *testing.T) {
	cfg := newTestConfig()
	config := NewConfig(cfg)

	// Default is 10%
	assert.Equal(t, 0.1, config.TraceSamplerArg())

	// Custom ratio
	cfg.values["OTEL_TRACES_SAMPLER_ARG"] = 0.5
	config2 := NewConfig(cfg)
	assert.Equal(t, 0.5, config2.TraceSamplerArg())
}

func TestConfig_ExporterType(t *testing.T) {
	cfg := newTestConfig()
	config := NewConfig(cfg)

	// Default is otlp
	assert.Equal(t, "otlp", config.ExporterType())

	// Can use OTEL_EXPORTER_TYPE
	cfg.values["OTEL_EXPORTER_TYPE"] = "stdout"
	config2 := NewConfig(cfg)
	assert.Equal(t, "stdout", config2.ExporterType())

	// Or OTEL_EXPORTER fallback
	delete(cfg.values, "OTEL_EXPORTER_TYPE")
	cfg.values["OTEL_EXPORTER"] = "none"
	config3 := NewConfig(cfg)
	assert.Equal(t, "none", config3.ExporterType())
}

func TestConfig_OTLPEndpoint(t *testing.T) {
	cfg := newTestConfig()
	config := NewConfig(cfg)

	// Default endpoint
	assert.Equal(t, "localhost:4317", config.OTLPEndpoint())

	// Custom endpoint
	cfg.values["OTEL_EXPORTER_OTLP_ENDPOINT"] = "collector.example.com:4317"
	config2 := NewConfig(cfg)
	assert.Equal(t, "collector.example.com:4317", config2.OTLPEndpoint())
}

func TestConfig_OTLPProtocol(t *testing.T) {
	cfg := newTestConfig()
	config := NewConfig(cfg)

	// Default is grpc
	assert.Equal(t, "grpc", config.OTLPProtocol())

	// Can be http
	cfg.values["OTEL_EXPORTER_OTLP_PROTOCOL"] = "http"
	config2 := NewConfig(cfg)
	assert.Equal(t, "http", config2.OTLPProtocol())
}

func TestConfig_OTLPInsecure(t *testing.T) {
	cfg := newTestConfig()
	config := NewConfig(cfg)

	// Default is true (insecure for local dev)
	assert.True(t, config.OTLPInsecure())

	// Can be set to false for TLS
	cfg.values["OTEL_EXPORTER_OTLP_INSECURE"] = false
	config2 := NewConfig(cfg)
	assert.False(t, config2.OTLPInsecure())
}

func TestConfig_OTLPHeaders(t *testing.T) {
	cfg := newTestConfig()
	config := NewConfig(cfg)

	// Default is nil
	assert.Nil(t, config.OTLPHeaders())

	// Parse header string
	cfg.values["OTEL_EXPORTER_OTLP_HEADERS"] = "Authorization=Bearer token,X-Custom=value"
	config2 := NewConfig(cfg)
	headers := config2.OTLPHeaders()
	require.NotNil(t, headers)
	assert.Equal(t, "Bearer token", headers["Authorization"])
	assert.Equal(t, "value", headers["X-Custom"])
}

func TestConfig_Propagators(t *testing.T) {
	cfg := newTestConfig()
	config := NewConfig(cfg)

	// Default propagators
	propagators := config.Propagators()
	assert.Contains(t, propagators, "tracecontext")
	assert.Contains(t, propagators, "baggage")

	// Custom propagators
	cfg.values["OTEL_PROPAGATORS"] = "tracecontext,b3"
	config2 := NewConfig(cfg)
	propagators2 := config2.Propagators()
	assert.Contains(t, propagators2, "tracecontext")
	assert.Contains(t, propagators2, "b3")
}

func TestConfig_ResourceAttributes(t *testing.T) {
	cfg := newTestConfig()
	config := NewConfig(cfg)

	// Default is nil
	assert.Nil(t, config.ResourceAttributes())

	// Parse attributes string
	cfg.values["OTEL_RESOURCE_ATTRIBUTES"] = "deployment.environment=production,cloud.region=us-east-1"
	config2 := NewConfig(cfg)
	attrs := config2.ResourceAttributes()
	require.NotNil(t, attrs)
	assert.Equal(t, "production", attrs["deployment.environment"])
	assert.Equal(t, "us-east-1", attrs["cloud.region"])
}

func TestConfig_ToContractConfig(t *testing.T) {
	cfg := newTestConfig()
	cfg.values["OTEL_SERVICE_NAME"] = "test-service"
	cfg.values["OTEL_SERVICE_VERSION"] = "1.2.3"
	cfg.values["OTEL_EXPORTER_TYPE"] = "stdout"

	config := NewConfig(cfg)
	contractConfig := config.ToContractConfig()

	assert.True(t, contractConfig.Enabled)
	assert.Equal(t, "test-service", contractConfig.ServiceName)
	assert.Equal(t, "1.2.3", contractConfig.ServiceVersion)
	assert.Equal(t, "stdout", contractConfig.ExporterType)
}

// Provider Tests

func TestNewProvider_WithStdout(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	cfg.values["OTEL_EXPORTER_TYPE"] = "stdout"

	config := NewConfig(cfg)
	ctx := context.Background()

	provider, err := NewProvider(ctx, config, logger)
	require.NoError(t, err)
	require.NotNil(t, provider)

	// Check that we can get a tracer
	tracer := provider.Tracer("test-tracer")
	assert.NotNil(t, tracer)

	// Check that we can get the tracer provider
	tp := provider.TracerProvider()
	assert.NotNil(t, tp)

	// Check propagator
	propagator := provider.TextMapPropagator()
	assert.NotNil(t, propagator)

	// Cleanup
	err = provider.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestNewProvider_WithNone(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	cfg.values["OTEL_EXPORTER_TYPE"] = "none"

	config := NewConfig(cfg)
	ctx := context.Background()

	provider, err := NewProvider(ctx, config, logger)
	require.NoError(t, err)
	require.NotNil(t, provider)

	// Cleanup
	err = provider.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestProvider_ForceFlush(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	cfg.values["OTEL_EXPORTER_TYPE"] = "none"

	config := NewConfig(cfg)
	ctx := context.Background()

	provider, err := NewProvider(ctx, config, logger)
	require.NoError(t, err)

	// ForceFlush should not error
	err = provider.ForceFlush(ctx)
	assert.NoError(t, err)

	// Cleanup
	_ = provider.Shutdown(ctx)
}

// Module Tests

func TestNewModule(t *testing.T) {
	logger := &testLogger{t: t}
	cfg := newTestConfig()
	cfg.values["OTEL_EXPORTER_TYPE"] = "none"

	ctx := context.Background()
	module, err := NewModule(ctx, cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, module)

	assert.Equal(t, "otel", module.Name())
	assert.NotNil(t, module.Provider())

	// Test lifecycle
	err = module.OnStart(ctx)
	assert.NoError(t, err)

	err = module.OnStop(ctx)
	assert.NoError(t, err)
}

// Sampler Tests

func TestCreateSampler_AlwaysOn(t *testing.T) {
	cfg := newTestConfig()
	cfg.values["OTEL_TRACES_SAMPLER"] = "always_on"
	config := NewConfig(cfg)

	sampler := createSampler(config)
	assert.NotNil(t, sampler)
}

func TestCreateSampler_AlwaysOff(t *testing.T) {
	cfg := newTestConfig()
	cfg.values["OTEL_TRACES_SAMPLER"] = "always_off"
	config := NewConfig(cfg)

	sampler := createSampler(config)
	assert.NotNil(t, sampler)
}

func TestCreateSampler_TraceIDRatio(t *testing.T) {
	cfg := newTestConfig()
	cfg.values["OTEL_TRACES_SAMPLER"] = "traceidratio"
	cfg.values["OTEL_TRACES_SAMPLER_ARG"] = 0.5
	config := NewConfig(cfg)

	sampler := createSampler(config)
	assert.NotNil(t, sampler)
}

func TestCreateSampler_ParentBased(t *testing.T) {
	testCases := []string{
		"parentbased_always_on",
		"parentbased_always_off",
		"parentbased_traceidratio",
	}

	for _, samplerType := range testCases {
		t.Run(samplerType, func(t *testing.T) {
			cfg := newTestConfig()
			cfg.values["OTEL_TRACES_SAMPLER"] = samplerType
			config := NewConfig(cfg)

			sampler := createSampler(config)
			assert.NotNil(t, sampler)
		})
	}
}

// Propagator Tests

func TestCreatePropagator_Default(t *testing.T) {
	cfg := newTestConfig()
	config := NewConfig(cfg)

	propagator := createPropagator(config)
	assert.NotNil(t, propagator)
}

func TestCreatePropagator_Custom(t *testing.T) {
	cfg := newTestConfig()
	cfg.values["OTEL_PROPAGATORS"] = "tracecontext,baggage"
	config := NewConfig(cfg)

	propagator := createPropagator(config)
	assert.NotNil(t, propagator)
}

func TestCreatePropagator_Empty(t *testing.T) {
	cfg := newTestConfig()
	cfg.values["OTEL_PROPAGATORS"] = "unknown"
	config := NewConfig(cfg)

	// Should return default propagators for unknown types
	propagator := createPropagator(config)
	assert.NotNil(t, propagator)
}

// Middleware Config Tests

func TestDefaultMiddlewareConfig(t *testing.T) {
	cfg := DefaultMiddlewareConfig("test-service")

	assert.NotNil(t, cfg.Tracer)
	assert.NotNil(t, cfg.Propagators)
	assert.Equal(t, "test-service", cfg.ServerName)
	assert.NotNil(t, cfg.SpanNameFormatter)
}

// NoopExporter Tests

func TestNoopExporter(t *testing.T) {
	exporter := &noopExporter{}
	ctx := context.Background()

	// ExportSpans should not error
	err := exporter.ExportSpans(ctx, nil)
	assert.NoError(t, err)

	// Shutdown should not error
	err = exporter.Shutdown(ctx)
	assert.NoError(t, err)
}
