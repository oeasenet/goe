package otel_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// MockConfig implements the Config interface for testing
type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) Get(key string) any {
	args := m.Called(key)
	return args.Get(0)
}

func (m *MockConfig) GetString(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *MockConfig) GetInt(key string) int {
	args := m.Called(key)
	return args.Int(0)
}

func (m *MockConfig) GetInt64(key string) int64 {
	args := m.Called(key)
	return args.Get(0).(int64)
}

func (m *MockConfig) GetFloat64(key string) float64 {
	args := m.Called(key)
	return args.Get(0).(float64)
}

func (m *MockConfig) GetBool(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockConfig) GetDuration(key string) time.Duration {
	args := m.Called(key)
	return args.Get(0).(time.Duration)
}

func (m *MockConfig) GetStringSlice(key string) []string {
	args := m.Called(key)
	return args.Get(0).([]string)
}

func (m *MockConfig) GetStringMap(key string) map[string]any {
	args := m.Called(key)
	return args.Get(0).(map[string]any)
}

func (m *MockConfig) Set(key string, value any) {
	m.Called(key, value)
}

func (m *MockConfig) Has(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockConfig) All() map[string]any {
	args := m.Called()
	return args.Get(0).(map[string]any)
}

func (m *MockConfig) Reload() error {
	args := m.Called()
	return args.Error(0)
}

// MockLogger implements the Logger interface for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Info(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Warn(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Error(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Fatal(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Debugf(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Infof(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Warnf(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Errorf(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Fatalf(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Debugw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Infow(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Warnw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Errorw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Fatalw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) With(keysAndValues ...any) contract.Logger {
	args := m.Called(keysAndValues)
	return args.Get(0).(contract.Logger)
}

func (m *MockLogger) WithContext(ctx context.Context) contract.Logger {
	args := m.Called(ctx)
	return args.Get(0).(contract.Logger)
}

func (m *MockLogger) WithError(err error) contract.Logger {
	args := m.Called(err)
	return args.Get(0).(contract.Logger)
}

func (m *MockLogger) GetLogger() *zap.SugaredLogger {
	args := m.Called()
	return args.Get(0).(*zap.SugaredLogger)
}

func TestNewObservability(t *testing.T) {
	// Create mock dependencies
	config := new(MockConfig)
	logger := new(MockLogger)

	// Set up expectations for disabled observability
	config.On("GetBool", "OTEL_ENABLED").Return(false)
	config.On("GetString", "OTEL_SERVICE_NAME").Return("")
	config.On("GetString", "APP_NAME").Return("test-app")
	logger.On("Error", mock.AnythingOfType("string"), mock.Anything).Maybe()
	logger.On("Info", mock.AnythingOfType("string"), mock.Anything).Maybe()

	// Create observability instance
	obs := observability.New(config, logger)

	// Test that it implements the interface
	assert.NotNil(t, obs)
	assert.NotNil(t, obs.Metrics())
	assert.NotNil(t, obs.Tracing())

	// Test shutdown
	err := obs.Shutdown(context.Background())
	assert.Nil(t, err)

	// Verify expectations
	config.AssertExpectations(t)
}

func TestNewConfig(t *testing.T) {
	// Create mock base config
	baseConfig := new(MockConfig)

	// Set up expectations
	baseConfig.On("GetBool", "OTEL_ENABLED").Return(true)
	baseConfig.On("GetString", "OTEL_SERVICE_NAME").Return("test-service")
	baseConfig.On("GetString", "OTEL_SERVICE_VERSION").Return("1.0.0")
	baseConfig.On("GetString", "OTEL_ENVIRONMENT").Return("test")

	// Create observability config
	obsConfig := observability.NewConfig(baseConfig)

	// Test the methods
	assert.True(t, obsConfig.Enabled())
	assert.Equal(t, "test-service", obsConfig.ServiceName())
	assert.Equal(t, "1.0.0", obsConfig.ServiceVersion())
	assert.Equal(t, "test", obsConfig.Environment())

	// Test metrics config
	metricsConfig := obsConfig.Metrics()
	assert.NotNil(t, metricsConfig)

	// Test tracing config
	tracingConfig := obsConfig.Tracing()
	assert.NotNil(t, tracingConfig)

	// Verify expectations
	baseConfig.AssertExpectations(t)
}

func TestMetricsConfig(t *testing.T) {
	// Create mock base config
	baseConfig := new(MockConfig)

	// Set up expectations
	baseConfig.On("Has", "OTEL_METRICS_ENABLED").Return(true)
	baseConfig.On("GetBool", "OTEL_METRICS_ENABLED").Return(true)
	baseConfig.On("GetString", "OTEL_METRICS_ENDPOINT").Return("http://localhost:9090")
	baseConfig.On("GetInt", "OTEL_METRICS_PORT").Return(9090)
	baseConfig.On("GetString", "OTEL_METRICS_PATH").Return("/metrics")
	baseConfig.On("GetDuration", "OTEL_METRICS_INTERVAL").Return(30 * time.Second)
	baseConfig.On("GetStringSlice", "OTEL_METRICS_EXPORTERS").Return([]string{"prometheus"})

	// Create observability config
	obsConfig := observability.NewConfig(baseConfig)
	metricsConfig := obsConfig.Metrics()

	// Test the methods
	assert.True(t, metricsConfig.Enabled())
	assert.Equal(t, "http://localhost:9090", metricsConfig.Endpoint())
	assert.Equal(t, 9090, metricsConfig.Port())
	assert.Equal(t, "/metrics", metricsConfig.Path())
	assert.Equal(t, 30*time.Second, metricsConfig.Interval())
	assert.Equal(t, []string{"prometheus"}, metricsConfig.Exporters())

	// Verify expectations
	baseConfig.AssertExpectations(t)
}

func TestTracingConfig(t *testing.T) {
	// Create mock base config
	baseConfig := new(MockConfig)

	// Set up expectations
	baseConfig.On("Has", "OTEL_TRACING_ENABLED").Return(true)
	baseConfig.On("GetBool", "OTEL_TRACING_ENABLED").Return(true)
	baseConfig.On("GetString", "OTEL_TRACING_ENDPOINT").Return("http://localhost:4318")
	baseConfig.On("GetFloat64", "OTEL_TRACING_SAMPLING_RATIO").Return(1.0)
	baseConfig.On("GetStringSlice", "OTEL_TRACING_EXPORTERS").Return([]string{"otlp"})
	baseConfig.On("All").Return(map[string]any{
		"OTEL_TRACING_HEADERS_AUTHORIZATION": "Bearer token123",
	})

	// Create observability config
	obsConfig := observability.NewConfig(baseConfig)
	tracingConfig := obsConfig.Tracing()

	// Test the methods
	assert.True(t, tracingConfig.Enabled())
	assert.Equal(t, "http://localhost:4318", tracingConfig.Endpoint())
	assert.Equal(t, 1.0, tracingConfig.SamplingRatio())
	assert.Equal(t, []string{"otlp"}, tracingConfig.Exporters())

	headers := tracingConfig.Headers()
	assert.Contains(t, headers, "authorization")
	assert.Equal(t, "Bearer token123", headers["authorization"])

	// Verify expectations
	baseConfig.AssertExpectations(t)
}

func TestMetricsManager(t *testing.T) {
	// Create mock dependencies
	config := new(MockConfig)
	logger := new(MockLogger)

	// Set up expectations
	config.On("GetString", "OTEL_SERVICE_NAME").Return("test-service")
	logger.On("Error", mock.AnythingOfType("string"), mock.Anything, mock.Anything).Maybe()

	// Create metrics manager
	metricsManager := observability.NewMetricsManager(config, logger, func(func(context.Context) error) {})

	// Test counter creation
	counter := metricsManager.Counter("test_counter",
		contract.WithDescription("Test counter"),
		contract.WithUnit("requests"),
	)
	assert.NotNil(t, counter)

	// Test histogram creation
	histogram := metricsManager.Histogram("test_histogram",
		contract.WithDescription("Test histogram"),
		contract.WithUnit("seconds"),
	)
	assert.NotNil(t, histogram)

	// Test gauge creation
	gauge := metricsManager.Gauge("test_gauge",
		contract.WithDescription("Test gauge"),
		contract.WithUnit("bytes"),
	)
	assert.NotNil(t, gauge)

	// Test up-down counter creation
	upDownCounter := metricsManager.UpDownCounter("test_updown",
		contract.WithDescription("Test up-down counter"),
		contract.WithUnit("connections"),
	)
	assert.NotNil(t, upDownCounter)

	// Test meter retrieval
	meter := metricsManager.GetMeter()
	assert.NotNil(t, meter)

	// Test that same metric names return same instances
	counter2 := metricsManager.Counter("test_counter")
	assert.Equal(t, counter, counter2)

	// Verify expectations
	config.AssertExpectations(t)
}

func TestTracingManager(t *testing.T) {
	// Create mock dependencies
	config := new(MockConfig)
	logger := new(MockLogger)

	// Set up expectations
	config.On("GetString", "OTEL_SERVICE_NAME").Return("test-service")

	// Create tracing manager
	tracingManager := observability.NewTracingManager(config, logger, func(func(context.Context) error) {})

	// Test span creation
	ctx := context.Background()
	newCtx, span := tracingManager.StartSpan(ctx, "test_span",
		contract.WithSpanKind(trace.SpanKindServer),
		contract.WithSpanAttributes(attribute.String("key", "value")),
	)

	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)

	// Test span operations
	span.SetAttributes(attribute.String("test", "value"))
	span.SetStatus(codes.Ok, "success")
	span.AddEvent("test_event")
	span.End()

	// Test tracer retrieval
	tracer := tracingManager.GetTracer()
	assert.NotNil(t, tracer)

	// Verify expectations
	config.AssertExpectations(t)
}

func TestModule(t *testing.T) {
	// Create mock dependencies
	config := new(MockConfig)
	logger := new(MockLogger)

	// Set up expectations for module lifecycle
	// NewSDKManager -> NewConfig -> obsConfig.Enabled()
	config.On("GetBool", "OTEL_ENABLED").Return(true)

	// SDK manager initialization calls (when enabled)
	// These are called during SDK initialization for resource creation and config
	config.On("GetString", "OTEL_SERVICE_NAME").Return("test-service")
	config.On("GetString", "OTEL_SERVICE_VERSION").Return("1.0.0")
	config.On("GetString", "OTEL_ENVIRONMENT").Return("test")

	// Metrics configuration calls during SDK initialization
	config.On("Has", "OTEL_METRICS_ENABLED").Return(true)
	config.On("GetBool", "OTEL_METRICS_ENABLED").Return(true)
	config.On("GetStringSlice", "OTEL_METRICS_EXPORTERS").Return([]string{"prometheus"})
	config.On("GetInt", "OTEL_METRICS_PORT").Return(9090)
	config.On("GetString", "OTEL_METRICS_PATH").Return("/metrics")

	// Tracing configuration calls during SDK initialization
	config.On("Has", "OTEL_TRACING_ENABLED").Return(true)
	config.On("GetBool", "OTEL_TRACING_ENABLED").Return(true)
	config.On("GetStringSlice", "OTEL_TRACING_EXPORTERS").Return([]string{"otlp"})
	config.On("GetString", "OTEL_TRACING_ENDPOINT").Return("http://localhost:4318")
	config.On("GetFloat64", "OTEL_TRACING_SAMPLING_RATIO").Return(1.0)
	config.On("All").Return(map[string]any{})
	config.On("GetString", "OTEL_EXPORTER_OTLP_HEADERS").Return("")

	// Logger expectations - SDK manager logs when initialized
	logger.On("Info", "OpenTelemetry SDK initialized", mock.Anything).Return()
	logger.On("Info", "Starting Prometheus metrics server", mock.Anything).Return()
	logger.On("Info", "Observability module started", mock.Anything).Return()
	logger.On("Info", "Observability module stopping", mock.Anything).Return()
	logger.On("Info", "OpenTelemetry SDK shutdown completed", mock.Anything).Return()
	logger.On("Info", "Observability module stopped", mock.Anything).Return()
	logger.On("Error", mock.AnythingOfType("string"), mock.Anything).Maybe()

	// Create module
	module := observability.NewModule(config, logger)

	// Test module interface
	assert.Equal(t, "observability", module.Name())

	// Test module lifecycle
	ctx := context.Background()
	err := module.OnStart(ctx)
	assert.Nil(t, err)

	err = module.OnStop(ctx)
	assert.Nil(t, err)

	// Test providers
	obs := module.Provide()
	assert.NotNil(t, obs)

	metricsManager := module.ProvideMetrics()
	assert.NotNil(t, metricsManager)

	tracingManager := module.ProvideTracing()
	assert.NotNil(t, tracingManager)

	// Verify expectations
	config.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestModuleDisabled(t *testing.T) {
	// Create mock dependencies
	config := new(MockConfig)
	logger := new(MockLogger)

	// Set up expectations for disabled module
	config.On("GetBool", "OTEL_ENABLED").Return(false)

	// NewMetricsManager and NewTracingManager both call these
	config.On("GetString", "OTEL_SERVICE_NAME").Return("").Times(2)
	config.On("GetString", "APP_NAME").Return("test-app").Times(2)

	// Expect the SDK manager to log that it's disabled
	logger.On("Info", "OpenTelemetry SDK disabled", mock.Anything).Return()

	// Expect the module lifecycle logs
	logger.On("Info", "Observability module started (disabled)", mock.Anything).Return()
	logger.On("Info", "Observability module stopping", mock.Anything).Return()
	logger.On("Info", "OpenTelemetry SDK shutdown completed", mock.Anything).Return()
	logger.On("Info", "Observability module stopped", mock.Anything).Return()

	// Create module
	module := observability.NewModule(config, logger)

	// Test module lifecycle with disabled observability
	ctx := context.Background()
	err := module.OnStart(ctx)
	assert.Nil(t, err)

	err = module.OnStop(ctx)
	assert.Nil(t, err)

	// Verify expectations
	config.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestMetricInstrumentation(t *testing.T) {
	// Create mock dependencies
	config := new(MockConfig)
	logger := new(MockLogger)

	// Set up expectations
	config.On("GetString", "OTEL_SERVICE_NAME").Return("test-service")
	logger.On("Error", mock.AnythingOfType("string"), mock.Anything, mock.Anything).Maybe()

	// Create metrics manager
	metricsManager := observability.NewMetricsManager(config, logger, func(func(context.Context) error) {})

	// Create metrics
	counter := metricsManager.Counter("http_requests_total",
		contract.WithDescription("Total HTTP requests"),
		contract.WithUnit("requests"),
	)

	histogram := metricsManager.Histogram("http_request_duration",
		contract.WithDescription("HTTP request duration"),
		contract.WithUnit("seconds"),
	)

	gauge := metricsManager.Gauge("active_connections",
		contract.WithDescription("Active connections"),
		contract.WithUnit("connections"),
	)

	upDownCounter := metricsManager.UpDownCounter("queue_size",
		contract.WithDescription("Queue size"),
		contract.WithUnit("items"),
	)

	// Test metric operations
	ctx := context.Background()
	attrs := []attribute.KeyValue{
		attribute.String("method", "GET"),
		attribute.String("status", "200"),
	}

	// Test counter
	counter.Inc(ctx, attrs...)
	counter.Add(ctx, 5, attrs...)

	// Test histogram
	histogram.Record(ctx, 0.123, attrs...)

	// Test gauge
	gauge.Set(ctx, 42.0, attrs...)

	// Test up-down counter
	upDownCounter.Add(ctx, 1, attrs...)
	upDownCounter.Add(ctx, -1, attrs...)

	// Verify expectations
	config.AssertExpectations(t)
}
