package observability

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/contract"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// MockConfig for testing
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

// MockLogger for testing
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

func TestObservability_New(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	// Mock expected calls
	config.On("Has", "OTEL_METRICS_ENABLED").Return(true)
	config.On("Has", "OTEL_TRACING_ENABLED").Return(true)
	config.On("GetString", "OTEL_SERVICE_NAME").Return("")
	config.On("GetString", "APP_NAME").Return("test-app")
	config.On("GetBool", "OTEL_ENABLED").Return(true)
	config.On("GetBool", "OTEL_METRICS_ENABLED").Return(true)
	config.On("GetBool", "OTEL_TRACING_ENABLED").Return(true)
	config.On("GetString", "OTEL_SERVICE_VERSION").Return("1.0.0")
	config.On("GetString", "OTEL_ENVIRONMENT").Return("test")
	config.On("GetStringSlice", "OTEL_METRICS_EXPORTERS").Return([]string{"prometheus"})
	config.On("GetStringSlice", "OTEL_TRACING_EXPORTERS").Return([]string{"otlp"})
	config.On("GetString", "OTEL_METRICS_EXPORTERS").Return("")
	config.On("GetString", "OTEL_TRACING_EXPORTERS").Return("")
	config.On("GetString", "OTEL_TRACING_ENDPOINT").Return("localhost:4318")
	config.On("GetFloat64", "OTEL_TRACING_SAMPLING_RATIO").Return(1.0)
	config.On("GetDuration", "OTEL_METRICS_INTERVAL").Return(30 * time.Second)
	config.On("GetInt", "OTEL_METRICS_PORT").Return(9090)
	config.On("GetString", "OTEL_METRICS_PATH").Return("/metrics")
	config.On("GetString", "OTEL_METRICS_ENDPOINT").Return("")
	config.On("GetString", "OTEL_EXPORTER_OTLP_ENDPOINT").Return("")
	config.On("GetString", "OTEL_EXPORTER_OTLP_HEADERS").Return("")
	config.On("All").Return(map[string]any{})

	// Mock GetLogger to return a proper *zap.SugaredLogger
	logger.On("GetLogger").Return(zap.NewNop().Sugar())

	// Mock all logger methods that might be called during SDK initialization
	logger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	logger.On("Debug", mock.AnythingOfType("string"), mock.Anything).Return()
	logger.On("Warn", mock.AnythingOfType("string"), mock.Anything).Return()
	logger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	logger.On("Fatal", mock.AnythingOfType("string"), mock.Anything).Return()

	obs := New(config, logger)

	assert.NotNil(t, obs)
	assert.NotNil(t, obs.Metrics())
	assert.NotNil(t, obs.Tracing())
}

func TestObservability_Shutdown(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	// Mock expected calls
	config.On("Has", "OTEL_METRICS_ENABLED").Return(true)
	config.On("Has", "OTEL_TRACING_ENABLED").Return(true)
	config.On("GetString", "OTEL_SERVICE_NAME").Return("")
	config.On("GetString", "APP_NAME").Return("test-app")
	config.On("GetBool", "OTEL_ENABLED").Return(true)
	config.On("GetBool", "OTEL_METRICS_ENABLED").Return(true)
	config.On("GetBool", "OTEL_TRACING_ENABLED").Return(true)
	config.On("GetString", "OTEL_SERVICE_VERSION").Return("1.0.0")
	config.On("GetString", "OTEL_ENVIRONMENT").Return("test")
	config.On("GetStringSlice", "OTEL_METRICS_EXPORTERS").Return([]string{"prometheus"})
	config.On("GetStringSlice", "OTEL_TRACING_EXPORTERS").Return([]string{"otlp"})
	config.On("GetString", "OTEL_METRICS_EXPORTERS").Return("")
	config.On("GetString", "OTEL_TRACING_EXPORTERS").Return("")
	config.On("GetString", "OTEL_TRACING_ENDPOINT").Return("localhost:4318")
	config.On("GetFloat64", "OTEL_TRACING_SAMPLING_RATIO").Return(1.0)
	config.On("GetDuration", "OTEL_METRICS_INTERVAL").Return(30 * time.Second)
	config.On("GetInt", "OTEL_METRICS_PORT").Return(9090)
	config.On("GetString", "OTEL_METRICS_PATH").Return("/metrics")
	config.On("GetString", "OTEL_METRICS_ENDPOINT").Return("")
	config.On("GetString", "OTEL_EXPORTER_OTLP_ENDPOINT").Return("")
	config.On("GetString", "OTEL_EXPORTER_OTLP_HEADERS").Return("")
	config.On("All").Return(map[string]any{})

	logger.On("GetLogger").Return(zap.NewNop().Sugar())

	// Mock all logger methods that might be called during SDK initialization
	logger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	logger.On("Debug", mock.AnythingOfType("string"), mock.Anything).Return()
	logger.On("Warn", mock.AnythingOfType("string"), mock.Anything).Return()
	logger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	logger.On("Fatal", mock.AnythingOfType("string"), mock.Anything).Return()

	obs := New(config, logger)

	ctx := context.Background()
	err := obs.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestMetricsManager_Counter(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	config.On("GetString", "OTEL_SERVICE_NAME").Return("")
	config.On("GetString", "APP_NAME").Return("test-app")
	logger.On("GetLogger").Return(zap.NewNop().Sugar())

	manager := NewMetricsManager(config, logger, func(func(context.Context) error) {})

	t.Run("create new counter", func(t *testing.T) {
		counter := manager.Counter("test_counter")
		assert.NotNil(t, counter)
	})

	t.Run("retrieve existing counter", func(t *testing.T) {
		counter1 := manager.Counter("test_counter_2")
		counter2 := manager.Counter("test_counter_2")
		assert.Equal(t, counter1, counter2)
	})

	t.Run("counter with options", func(t *testing.T) {
		opts := []contract.MetricOption{
			contract.WithDescription("Test counter"),
			contract.WithUnit("count"),
		}
		counter := manager.Counter("test_counter_with_opts", opts...)
		assert.NotNil(t, counter)
	})
}

func TestMetricsManager_Histogram(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	config.On("GetString", "OTEL_SERVICE_NAME").Return("")
	config.On("GetString", "APP_NAME").Return("test-app")
	logger.On("GetLogger").Return(zap.NewNop().Sugar())

	manager := NewMetricsManager(config, logger, func(func(context.Context) error) {})

	t.Run("create new histogram", func(t *testing.T) {
		histogram := manager.Histogram("test_histogram")
		assert.NotNil(t, histogram)
	})

	t.Run("retrieve existing histogram", func(t *testing.T) {
		histogram1 := manager.Histogram("test_histogram_2")
		histogram2 := manager.Histogram("test_histogram_2")
		assert.Equal(t, histogram1, histogram2)
	})
}

func TestMetricsManager_Gauge(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	config.On("GetString", "OTEL_SERVICE_NAME").Return("")
	config.On("GetString", "APP_NAME").Return("test-app")
	logger.On("GetLogger").Return(zap.NewNop().Sugar())

	manager := NewMetricsManager(config, logger, func(func(context.Context) error) {})

	t.Run("create new gauge", func(t *testing.T) {
		gauge := manager.Gauge("test_gauge")
		assert.NotNil(t, gauge)
	})

	t.Run("retrieve existing gauge", func(t *testing.T) {
		gauge1 := manager.Gauge("test_gauge_2")
		gauge2 := manager.Gauge("test_gauge_2")
		assert.Equal(t, gauge1, gauge2)
	})
}

func TestMetricsManager_UpDownCounter(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	config.On("GetString", "OTEL_SERVICE_NAME").Return("")
	config.On("GetString", "APP_NAME").Return("test-app")
	logger.On("GetLogger").Return(zap.NewNop().Sugar())

	manager := NewMetricsManager(config, logger, func(func(context.Context) error) {})

	t.Run("create new up-down counter", func(t *testing.T) {
		counter := manager.UpDownCounter("test_updown_counter")
		assert.NotNil(t, counter)
	})

	t.Run("retrieve existing up-down counter", func(t *testing.T) {
		counter1 := manager.UpDownCounter("test_updown_counter_2")
		counter2 := manager.UpDownCounter("test_updown_counter_2")
		assert.Equal(t, counter1, counter2)
	})
}

func TestMetricsManager_GetMeter(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	config.On("GetString", "OTEL_SERVICE_NAME").Return("")
	config.On("GetString", "APP_NAME").Return("test-app")
	logger.On("GetLogger").Return(zap.NewNop().Sugar())

	manager := NewMetricsManager(config, logger, func(func(context.Context) error) {})

	meter := manager.GetMeter()
	assert.NotNil(t, meter)
}

func TestTracingManager_StartSpan(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	config.On("GetString", "OTEL_SERVICE_NAME").Return("")
	config.On("GetString", "APP_NAME").Return("test-app")
	logger.On("GetLogger").Return(zap.NewNop().Sugar())

	manager := NewTracingManager(config, logger, func(func(context.Context) error) {})

	t.Run("start span without options", func(t *testing.T) {
		ctx := context.Background()
		spanCtx, span := manager.StartSpan(ctx, "test_span")

		assert.NotNil(t, spanCtx)
		assert.NotNil(t, span)
	})

	t.Run("start span with options", func(t *testing.T) {
		ctx := context.Background()
		opts := []contract.SpanOption{
			contract.WithSpanKind(trace.SpanKindServer),
			contract.WithSpanAttributes(attribute.String("key", "value")),
		}
		spanCtx, span := manager.StartSpan(ctx, "test_span_with_opts", opts...)

		assert.NotNil(t, spanCtx)
		assert.NotNil(t, span)
	})
}

func TestTracingManager_GetTracer(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	config.On("GetString", "OTEL_SERVICE_NAME").Return("")
	config.On("GetString", "APP_NAME").Return("test-app")
	logger.On("GetLogger").Return(zap.NewNop().Sugar())

	manager := NewTracingManager(config, logger, func(func(context.Context) error) {})

	tracer := manager.GetTracer()
	assert.NotNil(t, tracer)
}

func TestModule_New(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	// Mock expected calls
	config.On("Has", "OTEL_METRICS_ENABLED").Return(true)
	config.On("Has", "OTEL_TRACING_ENABLED").Return(true)
	config.On("GetString", "OTEL_SERVICE_NAME").Return("")
	config.On("GetString", "APP_NAME").Return("test-app")
	config.On("GetBool", "OTEL_ENABLED").Return(true)
	config.On("GetBool", "OTEL_METRICS_ENABLED").Return(true)
	config.On("GetBool", "OTEL_TRACING_ENABLED").Return(true)
	config.On("GetString", "OTEL_SERVICE_VERSION").Return("1.0.0")
	config.On("GetString", "OTEL_ENVIRONMENT").Return("test")
	config.On("GetStringSlice", "OTEL_METRICS_EXPORTERS").Return([]string{"prometheus"})
	config.On("GetStringSlice", "OTEL_TRACING_EXPORTERS").Return([]string{"otlp"})
	config.On("GetString", "OTEL_METRICS_EXPORTERS").Return("")
	config.On("GetString", "OTEL_TRACING_EXPORTERS").Return("")
	config.On("GetString", "OTEL_TRACING_ENDPOINT").Return("localhost:4318")
	config.On("GetFloat64", "OTEL_TRACING_SAMPLING_RATIO").Return(1.0)
	config.On("GetDuration", "OTEL_METRICS_INTERVAL").Return(30 * time.Second)
	config.On("GetInt", "OTEL_METRICS_PORT").Return(9090)
	config.On("GetString", "OTEL_METRICS_PATH").Return("/metrics")
	config.On("GetString", "OTEL_METRICS_ENDPOINT").Return("")
	config.On("GetString", "OTEL_EXPORTER_OTLP_ENDPOINT").Return("")
	config.On("GetString", "OTEL_EXPORTER_OTLP_HEADERS").Return("")
	config.On("All").Return(map[string]any{})

	logger.On("GetLogger").Return(zap.NewNop().Sugar())

	// Mock all logger methods that might be called during SDK initialization
	logger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	logger.On("Debug", mock.AnythingOfType("string"), mock.Anything).Return()
	logger.On("Warn", mock.AnythingOfType("string"), mock.Anything).Return()
	logger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	logger.On("Fatal", mock.AnythingOfType("string"), mock.Anything).Return()

	module := NewModule(config, logger)

	assert.NotNil(t, module)
	assert.Equal(t, "observability", module.Name())
	assert.NotNil(t, module.Provide())
	assert.NotNil(t, module.ProvideMetrics())
	assert.NotNil(t, module.ProvideTracing())
}

func TestModule_Lifecycle(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	// Mock expected calls for module creation
	config.On("Has", "OTEL_METRICS_ENABLED").Return(true)
	config.On("Has", "OTEL_TRACING_ENABLED").Return(true)
	config.On("GetString", "OTEL_SERVICE_NAME").Return("")
	config.On("GetString", "APP_NAME").Return("test-app")
	config.On("GetBool", "OTEL_ENABLED").Return(true)
	config.On("GetBool", "OTEL_METRICS_ENABLED").Return(true)
	config.On("GetBool", "OTEL_TRACING_ENABLED").Return(true)
	config.On("GetString", "OTEL_SERVICE_VERSION").Return("1.0.0")
	config.On("GetString", "OTEL_ENVIRONMENT").Return("test")
	config.On("GetStringSlice", "OTEL_METRICS_EXPORTERS").Return([]string{"prometheus"})
	config.On("GetStringSlice", "OTEL_TRACING_EXPORTERS").Return([]string{"otlp"})
	config.On("GetString", "OTEL_METRICS_EXPORTERS").Return("")
	config.On("GetString", "OTEL_TRACING_EXPORTERS").Return("")
	config.On("GetString", "OTEL_TRACING_ENDPOINT").Return("localhost:4318")
	config.On("GetFloat64", "OTEL_TRACING_SAMPLING_RATIO").Return(1.0)
	config.On("GetDuration", "OTEL_METRICS_INTERVAL").Return(30 * time.Second)
	config.On("GetInt", "OTEL_METRICS_PORT").Return(9090)
	config.On("GetString", "OTEL_METRICS_PATH").Return("/metrics")
	config.On("GetString", "OTEL_METRICS_ENDPOINT").Return("")
	config.On("GetString", "OTEL_EXPORTER_OTLP_ENDPOINT").Return("")
	config.On("GetString", "OTEL_EXPORTER_OTLP_HEADERS").Return("")
	config.On("All").Return(map[string]any{})

	logger.On("GetLogger").Return(zap.NewNop().Sugar())

	// Mock all logger methods that might be called during SDK initialization
	logger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	logger.On("Debug", mock.AnythingOfType("string"), mock.Anything).Return()
	logger.On("Warn", mock.AnythingOfType("string"), mock.Anything).Return()
	logger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	logger.On("Fatal", mock.AnythingOfType("string"), mock.Anything).Return()

	module := NewModule(config, logger)

	ctx := context.Background()

	t.Run("start module", func(t *testing.T) {
		err := module.OnStart(ctx)
		assert.NoError(t, err)
	})

	t.Run("stop module", func(t *testing.T) {
		err := module.OnStop(ctx)
		assert.NoError(t, err)
	})
}

func TestConfig_New(t *testing.T) {
	baseConfig := &MockConfig{}

	config := NewConfig(baseConfig)
	assert.NotNil(t, config)
}

func TestConfig_Enabled(t *testing.T) {
	baseConfig := &MockConfig{}
	config := NewConfig(baseConfig)

	baseConfig.On("GetBool", "OTEL_ENABLED").Return(true)

	assert.True(t, config.Enabled())
}

func TestConfig_ServiceName(t *testing.T) {
	t.Run("from OTEL_SERVICE_NAME", func(t *testing.T) {
		baseConfig := &MockConfig{}
		config := NewConfig(baseConfig)
		baseConfig.On("GetString", "OTEL_SERVICE_NAME").Return("test-service")

		assert.Equal(t, "test-service", config.ServiceName())
	})

	t.Run("from APP_NAME", func(t *testing.T) {
		baseConfig := &MockConfig{}
		config := NewConfig(baseConfig)
		baseConfig.On("GetString", "OTEL_SERVICE_NAME").Return("")
		baseConfig.On("GetString", "APP_NAME").Return("test-app")

		assert.Equal(t, "test-app", config.ServiceName())
	})

	t.Run("default", func(t *testing.T) {
		baseConfig := &MockConfig{}
		config := NewConfig(baseConfig)
		baseConfig.On("GetString", "OTEL_SERVICE_NAME").Return("")
		baseConfig.On("GetString", "APP_NAME").Return("")

		assert.Equal(t, "goe-app", config.ServiceName())
	})
}

func TestConfig_ServiceVersion(t *testing.T) {
	t.Run("from OTEL_SERVICE_VERSION", func(t *testing.T) {
		baseConfig := &MockConfig{}
		config := NewConfig(baseConfig)
		baseConfig.On("GetString", "OTEL_SERVICE_VERSION").Return("2.0.0")

		assert.Equal(t, "2.0.0", config.ServiceVersion())
	})

	t.Run("from APP_VERSION", func(t *testing.T) {
		baseConfig := &MockConfig{}
		config := NewConfig(baseConfig)
		baseConfig.On("GetString", "OTEL_SERVICE_VERSION").Return("")
		baseConfig.On("GetString", "APP_VERSION").Return("1.5.0")

		assert.Equal(t, "1.5.0", config.ServiceVersion())
	})

	t.Run("default", func(t *testing.T) {
		baseConfig := &MockConfig{}
		config := NewConfig(baseConfig)
		baseConfig.On("GetString", "OTEL_SERVICE_VERSION").Return("")
		baseConfig.On("GetString", "APP_VERSION").Return("")

		assert.Equal(t, "1.0.0", config.ServiceVersion())
	})
}

func TestConfig_Environment(t *testing.T) {
	t.Run("from OTEL_ENVIRONMENT", func(t *testing.T) {
		baseConfig := &MockConfig{}
		config := NewConfig(baseConfig)
		baseConfig.On("GetString", "OTEL_ENVIRONMENT").Return("production")

		assert.Equal(t, "production", config.Environment())
	})

	t.Run("from GOE_ENV", func(t *testing.T) {
		baseConfig := &MockConfig{}
		config := NewConfig(baseConfig)
		baseConfig.On("GetString", "OTEL_ENVIRONMENT").Return("")
		baseConfig.On("GetString", "GOE_ENV").Return("staging")

		assert.Equal(t, "staging", config.Environment())
	})

	t.Run("default", func(t *testing.T) {
		baseConfig := &MockConfig{}
		config := NewConfig(baseConfig)
		baseConfig.On("GetString", "OTEL_ENVIRONMENT").Return("")
		baseConfig.On("GetString", "GOE_ENV").Return("")

		assert.Equal(t, "dev", config.Environment())
	})
}

func TestMetricsConfig(t *testing.T) {
	baseConfig := &MockConfig{}
	config := NewConfig(baseConfig)
	metricsConfig := config.Metrics()

	t.Run("enabled", func(t *testing.T) {
		baseConfig.On("Has", "OTEL_METRICS_ENABLED").Return(true)
		baseConfig.On("GetBool", "OTEL_METRICS_ENABLED").Return(true)

		assert.True(t, metricsConfig.Enabled())
	})

	t.Run("port", func(t *testing.T) {
		baseConfig.On("GetInt", "OTEL_METRICS_PORT").Return(9090)

		assert.Equal(t, 9090, metricsConfig.Port())
	})

	t.Run("path", func(t *testing.T) {
		baseConfig.On("GetString", "OTEL_METRICS_PATH").Return("/metrics")

		assert.Equal(t, "/metrics", metricsConfig.Path())
	})

	t.Run("interval", func(t *testing.T) {
		baseConfig.On("GetDuration", "OTEL_METRICS_INTERVAL").Return(30 * time.Second)

		assert.Equal(t, 30*time.Second, metricsConfig.Interval())
	})

	t.Run("exporters", func(t *testing.T) {
		baseConfig.On("GetStringSlice", "OTEL_METRICS_EXPORTERS").Return([]string{"prometheus"})

		exporters := metricsConfig.Exporters()
		assert.Equal(t, []string{"prometheus"}, exporters)
	})
}

func TestTracingConfig(t *testing.T) {
	baseConfig := &MockConfig{}
	config := NewConfig(baseConfig)
	tracingConfig := config.Tracing()

	t.Run("enabled", func(t *testing.T) {
		baseConfig.On("Has", "OTEL_TRACING_ENABLED").Return(true)
		baseConfig.On("GetBool", "OTEL_TRACING_ENABLED").Return(true)

		assert.True(t, tracingConfig.Enabled())
	})

	t.Run("endpoint", func(t *testing.T) {
		baseConfig.On("GetString", "OTEL_TRACING_ENDPOINT").Return("localhost:4318")

		assert.Equal(t, "localhost:4318", tracingConfig.Endpoint())
	})

	t.Run("sampling ratio", func(t *testing.T) {
		baseConfig.On("GetFloat64", "OTEL_TRACING_SAMPLING_RATIO").Return(0.5)

		assert.Equal(t, 0.5, tracingConfig.SamplingRatio())
	})

	t.Run("exporters", func(t *testing.T) {
		baseConfig.On("GetStringSlice", "OTEL_TRACING_EXPORTERS").Return([]string{"otlp"})

		exporters := tracingConfig.Exporters()
		assert.Equal(t, []string{"otlp"}, exporters)
	})

	t.Run("headers", func(t *testing.T) {
		baseConfig.On("All").Return(map[string]any{
			"OTEL_TRACING_HEADERS_API_KEY":   "secret123",
			"OTEL_TRACING_HEADERS_CLIENT_ID": "client456",
		})

		headers := tracingConfig.Headers()
		assert.Equal(t, "secret123", headers["api-key"])
		assert.Equal(t, "client456", headers["client-id"])
	})
}
