package contract_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/contract"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// MockObservability implements the Observability interface
type MockObservability struct {
	mock.Mock
}

func (m *MockObservability) Metrics() contract.MetricsManager {
	args := m.Called()
	return args.Get(0).(contract.MetricsManager)
}

func (m *MockObservability) Tracing() contract.TracingManager {
	args := m.Called()
	return args.Get(0).(contract.TracingManager)
}

func (m *MockObservability) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockMetricsManager implements the MetricsManager interface
type MockMetricsManager struct {
	mock.Mock
}

func (m *MockMetricsManager) Counter(name string, opts ...contract.MetricOption) contract.Counter {
	args := m.Called(name, opts)
	return args.Get(0).(contract.Counter)
}

func (m *MockMetricsManager) Histogram(name string, opts ...contract.MetricOption) contract.Histogram {
	args := m.Called(name, opts)
	return args.Get(0).(contract.Histogram)
}

func (m *MockMetricsManager) Gauge(name string, opts ...contract.MetricOption) contract.Gauge {
	args := m.Called(name, opts)
	return args.Get(0).(contract.Gauge)
}

func (m *MockMetricsManager) UpDownCounter(name string, opts ...contract.MetricOption) contract.UpDownCounter {
	args := m.Called(name, opts)
	return args.Get(0).(contract.UpDownCounter)
}

func (m *MockMetricsManager) GetMeter() metric.Meter {
	args := m.Called()
	return args.Get(0).(metric.Meter)
}

// MockTracingManager implements the TracingManager interface
type MockTracingManager struct {
	mock.Mock
}

func (m *MockTracingManager) StartSpan(ctx context.Context, name string, opts ...contract.SpanOption) (context.Context, contract.Span) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(context.Context), args.Get(1).(contract.Span)
}

func (m *MockTracingManager) GetTracer() trace.Tracer {
	args := m.Called()
	return args.Get(0).(trace.Tracer)
}

// MockCounter implements the Counter interface
type MockCounter struct {
	mock.Mock
}

func (m *MockCounter) Add(ctx context.Context, value int64, attrs ...attribute.KeyValue) {
	m.Called(ctx, value, attrs)
}

func (m *MockCounter) Inc(ctx context.Context, attrs ...attribute.KeyValue) {
	m.Called(ctx, attrs)
}

// MockHistogram implements the Histogram interface
type MockHistogram struct {
	mock.Mock
}

func (m *MockHistogram) Record(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	m.Called(ctx, value, attrs)
}

// MockGauge implements the Gauge interface
type MockGauge struct {
	mock.Mock
}

func (m *MockGauge) Set(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	m.Called(ctx, value, attrs)
}

// MockUpDownCounter implements the UpDownCounter interface
type MockUpDownCounter struct {
	mock.Mock
}

func (m *MockUpDownCounter) Add(ctx context.Context, value int64, attrs ...attribute.KeyValue) {
	m.Called(ctx, value, attrs)
}

// MockSpan implements the Span interface
type MockSpan struct {
	mock.Mock
}

func (m *MockSpan) SetAttributes(attrs ...attribute.KeyValue) {
	m.Called(attrs)
}

func (m *MockSpan) SetStatus(code codes.Code, description string) {
	m.Called(code, description)
}

func (m *MockSpan) RecordError(err error, opts ...trace.EventOption) {
	m.Called(err, opts)
}

func (m *MockSpan) AddEvent(name string, opts ...trace.EventOption) {
	m.Called(name, opts)
}

func (m *MockSpan) End(opts ...trace.SpanEndOption) {
	m.Called(opts)
}

func (m *MockSpan) GetSpan() trace.Span {
	args := m.Called()
	return args.Get(0).(trace.Span)
}

// MockObservabilityConfig implements the ObservabilityConfig interface
type MockObservabilityConfig struct {
	mock.Mock
}

func (m *MockObservabilityConfig) Enabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockObservabilityConfig) ServiceName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockObservabilityConfig) ServiceVersion() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockObservabilityConfig) Environment() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockObservabilityConfig) Metrics() contract.MetricsConfig {
	args := m.Called()
	return args.Get(0).(contract.MetricsConfig)
}

func (m *MockObservabilityConfig) Tracing() contract.TracingConfig {
	args := m.Called()
	return args.Get(0).(contract.TracingConfig)
}

// MockMetricsConfig implements the MetricsConfig interface
type MockMetricsConfig struct {
	mock.Mock
}

func (m *MockMetricsConfig) Enabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMetricsConfig) Endpoint() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockMetricsConfig) Port() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMetricsConfig) Path() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockMetricsConfig) Interval() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *MockMetricsConfig) Exporters() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

// MockTracingConfig implements the TracingConfig interface
type MockTracingConfig struct {
	mock.Mock
}

func (m *MockTracingConfig) Enabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTracingConfig) Endpoint() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTracingConfig) SamplingRatio() float64 {
	args := m.Called()
	return args.Get(0).(float64)
}

func (m *MockTracingConfig) Exporters() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockTracingConfig) Headers() map[string]string {
	args := m.Called()
	return args.Get(0).(map[string]string)
}

// Test functions

func TestObservabilityInterface(t *testing.T) {
	// This test verifies that MockObservability implements the Observability interface
	var _ contract.Observability = (*MockObservability)(nil)

	// Create a mock observability
	obs := new(MockObservability)
	metricsManager := new(MockMetricsManager)
	tracingManager := new(MockTracingManager)

	// Set up expectations
	obs.On("Metrics").Return(metricsManager)
	obs.On("Tracing").Return(tracingManager)
	obs.On("Shutdown", mock.Anything).Return(nil)

	// Test the methods
	assert.Equal(t, metricsManager, obs.Metrics())
	assert.Equal(t, tracingManager, obs.Tracing())
	assert.Nil(t, obs.Shutdown(context.Background()))

	// Verify expectations
	obs.AssertExpectations(t)
}

func TestMetricsManagerInterface(t *testing.T) {
	// This test verifies that MockMetricsManager implements the MetricsManager interface
	var _ contract.MetricsManager = (*MockMetricsManager)(nil)

	// Create a mock metrics manager
	manager := new(MockMetricsManager)
	counter := new(MockCounter)
	histogram := new(MockHistogram)
	gauge := new(MockGauge)
	upDownCounter := new(MockUpDownCounter)

	// Set up expectations
	manager.On("Counter", "test_counter", mock.Anything).Return(counter)
	manager.On("Histogram", "test_histogram", mock.Anything).Return(histogram)
	manager.On("Gauge", "test_gauge", mock.Anything).Return(gauge)
	manager.On("UpDownCounter", "test_updown", mock.Anything).Return(upDownCounter)
	// Test the methods
	assert.Equal(t, counter, manager.Counter("test_counter"))
	assert.Equal(t, histogram, manager.Histogram("test_histogram"))
	assert.Equal(t, gauge, manager.Gauge("test_gauge"))
	assert.Equal(t, upDownCounter, manager.UpDownCounter("test_updown"))

	// Verify expectations
	manager.AssertExpectations(t)
}

func TestTracingManagerInterface(t *testing.T) {
	// This test verifies that MockTracingManager implements the TracingManager interface
	var _ contract.TracingManager = (*MockTracingManager)(nil)

	// Create a mock tracing manager
	manager := new(MockTracingManager)
	span := new(MockSpan)
	ctx := context.Background()

	// Set up expectations
	manager.On("StartSpan", ctx, "test_span", mock.Anything).Return(ctx, span)

	// Test the methods
	newCtx, newSpan := manager.StartSpan(ctx, "test_span")
	assert.Equal(t, ctx, newCtx)
	assert.Equal(t, span, newSpan)

	// Verify expectations
	manager.AssertExpectations(t)
}

func TestCounterInterface(t *testing.T) {
	// This test verifies that MockCounter implements the Counter interface
	var _ contract.Counter = (*MockCounter)(nil)

	// Create a mock counter
	counter := new(MockCounter)
	ctx := context.Background()
	attrs := []attribute.KeyValue{attribute.String("key", "value")}

	// Set up expectations
	counter.On("Add", ctx, int64(5), attrs).Return()
	counter.On("Inc", ctx, attrs).Return()

	// Test the methods
	counter.Add(ctx, 5, attrs...)
	counter.Inc(ctx, attrs...)

	// Verify expectations
	counter.AssertExpectations(t)
}

func TestHistogramInterface(t *testing.T) {
	// This test verifies that MockHistogram implements the Histogram interface
	var _ contract.Histogram = (*MockHistogram)(nil)

	// Create a mock histogram
	histogram := new(MockHistogram)
	ctx := context.Background()
	attrs := []attribute.KeyValue{attribute.String("key", "value")}

	// Set up expectations
	histogram.On("Record", ctx, 1.5, attrs).Return()

	// Test the methods
	histogram.Record(ctx, 1.5, attrs...)

	// Verify expectations
	histogram.AssertExpectations(t)
}

func TestGaugeInterface(t *testing.T) {
	// This test verifies that MockGauge implements the Gauge interface
	var _ contract.Gauge = (*MockGauge)(nil)

	// Create a mock gauge
	gauge := new(MockGauge)
	ctx := context.Background()
	attrs := []attribute.KeyValue{attribute.String("key", "value")}

	// Set up expectations
	gauge.On("Set", ctx, 42.0, attrs).Return()

	// Test the methods
	gauge.Set(ctx, 42.0, attrs...)

	// Verify expectations
	gauge.AssertExpectations(t)
}

func TestUpDownCounterInterface(t *testing.T) {
	// This test verifies that MockUpDownCounter implements the UpDownCounter interface
	var _ contract.UpDownCounter = (*MockUpDownCounter)(nil)

	// Create a mock up-down counter
	upDownCounter := new(MockUpDownCounter)
	ctx := context.Background()
	attrs := []attribute.KeyValue{attribute.String("key", "value")}

	// Set up expectations
	upDownCounter.On("Add", ctx, int64(-3), attrs).Return()

	// Test the methods
	upDownCounter.Add(ctx, -3, attrs...)

	// Verify expectations
	upDownCounter.AssertExpectations(t)
}

func TestSpanInterface(t *testing.T) {
	// This test verifies that MockSpan implements the Span interface
	var _ contract.Span = (*MockSpan)(nil)

	// Create a mock span
	span := new(MockSpan)
	attrs := []attribute.KeyValue{attribute.String("key", "value")}

	// Set up expectations
	span.On("SetAttributes", attrs).Return()
	span.On("SetStatus", codes.Ok, "success").Return()
	span.On("RecordError", mock.AnythingOfType("*errors.errorString"), mock.Anything).Return()
	span.On("AddEvent", "test_event", mock.Anything).Return()
	span.On("End", mock.Anything).Return()

	// Test the methods
	span.SetAttributes(attrs...)
	span.SetStatus(codes.Ok, "success")
	span.RecordError(assert.AnError)
	span.AddEvent("test_event")
	span.End()

	// Verify expectations
	span.AssertExpectations(t)
}

func TestMetricOptions(t *testing.T) {
	// Test metric option implementations
	config := &contract.MetricConfig{}

	// Test WithDescription
	descOpt := contract.WithDescription("test description")
	descOpt.Apply(config)
	assert.Equal(t, "test description", config.Description)

	// Test WithUnit
	unitOpt := contract.WithUnit("bytes")
	unitOpt.Apply(config)
	assert.Equal(t, "bytes", config.Unit)

	// Test WithMetricAttributes
	attrs := []attribute.KeyValue{attribute.String("key", "value")}
	attrOpt := contract.WithMetricAttributes(attrs...)
	attrOpt.Apply(config)
	assert.Equal(t, attrs, config.Attributes)
}

func TestSpanOptions(t *testing.T) {
	// Test span option implementations
	config := &contract.SpanConfig{}

	// Test WithSpanKind
	kindOpt := contract.WithSpanKind(trace.SpanKindServer)
	kindOpt.Apply(config)
	assert.Equal(t, trace.SpanKindServer, config.Kind)

	// Test WithSpanAttributes
	attrs := []attribute.KeyValue{attribute.String("key", "value")}
	attrOpt := contract.WithSpanAttributes(attrs...)
	attrOpt.Apply(config)
	assert.Equal(t, attrs, config.Attributes)
}
