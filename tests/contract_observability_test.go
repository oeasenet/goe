package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/contract"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// MockObservability is a mock implementation of the Observability interface
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

func TestObservabilityInterface(t *testing.T) {
	// This test verifies that MockObservability implements the Observability interface
	var _ contract.Observability = (*MockObservability)(nil)

	// Create a mock observability
	observability := new(MockObservability)
	metricsManager := new(MockMetricsManager)
	tracingManager := new(MockTracingManager)

	// Set up expectations
	observability.On("Metrics").Return(metricsManager)
	observability.On("Tracing").Return(tracingManager)
	observability.On("Shutdown", mock.Anything).Return(nil)

	// Test the methods
	assert.Equal(t, metricsManager, observability.Metrics())
	assert.Equal(t, tracingManager, observability.Tracing())
	assert.NoError(t, observability.Shutdown(context.Background()))

	// Verify expectations
	observability.AssertExpectations(t)
}

// MockMetricsManager is a mock implementation of the MetricsManager interface
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
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(metric.Meter)
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
	manager.On("UpDownCounter", "test_updown_counter", mock.Anything).Return(upDownCounter)
	manager.On("GetMeter").Return(nil)

	// Test the methods
	assert.Equal(t, counter, manager.Counter("test_counter"))
	assert.Equal(t, histogram, manager.Histogram("test_histogram"))
	assert.Equal(t, gauge, manager.Gauge("test_gauge"))
	assert.Equal(t, upDownCounter, manager.UpDownCounter("test_updown_counter"))
	assert.Nil(t, manager.GetMeter())

	// Verify expectations
	manager.AssertExpectations(t)
}

// MockTracingManager is a mock implementation of the TracingManager interface
type MockTracingManager struct {
	mock.Mock
}

func (m *MockTracingManager) StartSpan(ctx context.Context, name string, opts ...contract.SpanOption) (context.Context, contract.Span) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(context.Context), args.Get(1).(contract.Span)
}

func (m *MockTracingManager) GetTracer() trace.Tracer {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(trace.Tracer)
}

func TestTracingManagerInterface(t *testing.T) {
	// This test verifies that MockTracingManager implements the TracingManager interface
	var _ contract.TracingManager = (*MockTracingManager)(nil)

	// Create a mock tracing manager
	manager := new(MockTracingManager)
	span := new(MockSpan)

	// Set up expectations
	ctx := context.Background()
	newCtx := context.WithValue(ctx, "span", span)
	manager.On("StartSpan", ctx, "test_span", mock.Anything).Return(newCtx, span)
	manager.On("GetTracer").Return(nil)

	// Test the methods
	returnedCtx, returnedSpan := manager.StartSpan(ctx, "test_span")
	assert.Equal(t, newCtx, returnedCtx)
	assert.Equal(t, span, returnedSpan)
	assert.Nil(t, manager.GetTracer())

	// Verify expectations
	manager.AssertExpectations(t)
}

// MockCounter is a mock implementation of the Counter interface
type MockCounter struct {
	mock.Mock
}

func (m *MockCounter) Add(ctx context.Context, value int64, attrs ...attribute.KeyValue) {
	args := []interface{}{ctx, value}
	for _, attr := range attrs {
		args = append(args, attr)
	}
	m.Called(args...)
}

func (m *MockCounter) Inc(ctx context.Context, attrs ...attribute.KeyValue) {
	args := []interface{}{ctx}
	for _, attr := range attrs {
		args = append(args, attr)
	}
	m.Called(args...)
}

func TestCounterInterface(t *testing.T) {
	// This test verifies that MockCounter implements the Counter interface
	var _ contract.Counter = (*MockCounter)(nil)

	// Create a mock counter
	counter := new(MockCounter)

	// Set up expectations
	counter.On("Add", mock.Anything, int64(1), mock.Anything).Return()
	counter.On("Inc", mock.Anything, mock.Anything).Return()

	// Test the methods
	ctx := context.Background()
	counter.Add(ctx, 1, attribute.String("label1", "value1"))
	counter.Inc(ctx, attribute.String("label1", "value1"))

	// Verify expectations
	counter.AssertExpectations(t)
}

// MockHistogram is a mock implementation of the Histogram interface
type MockHistogram struct {
	mock.Mock
}

func (m *MockHistogram) Record(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	args := []interface{}{ctx, value}
	for _, attr := range attrs {
		args = append(args, attr)
	}
	m.Called(args...)
}

func TestHistogramInterface(t *testing.T) {
	// This test verifies that MockHistogram implements the Histogram interface
	var _ contract.Histogram = (*MockHistogram)(nil)

	// Create a mock histogram
	histogram := new(MockHistogram)

	// Set up expectations
	histogram.On("Record", mock.Anything, 1.5, mock.Anything).Return()

	// Test the methods
	ctx := context.Background()
	histogram.Record(ctx, 1.5, attribute.String("label1", "value1"))

	// Verify expectations
	histogram.AssertExpectations(t)
}

// MockGauge is a mock implementation of the Gauge interface
type MockGauge struct {
	mock.Mock
}

func (m *MockGauge) Set(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	args := []interface{}{ctx, value}
	for _, attr := range attrs {
		args = append(args, attr)
	}
	m.Called(args...)
}

func TestGaugeInterface(t *testing.T) {
	// This test verifies that MockGauge implements the Gauge interface
	var _ contract.Gauge = (*MockGauge)(nil)

	// Create a mock gauge
	gauge := new(MockGauge)

	// Set up expectations
	gauge.On("Set", mock.Anything, 42.0, mock.Anything).Return()

	// Test the methods
	ctx := context.Background()
	gauge.Set(ctx, 42.0, attribute.String("label1", "value1"))

	// Verify expectations
	gauge.AssertExpectations(t)
}

// MockUpDownCounter is a mock implementation of the UpDownCounter interface
type MockUpDownCounter struct {
	mock.Mock
}

func (m *MockUpDownCounter) Add(ctx context.Context, value int64, attrs ...attribute.KeyValue) {
	args := []interface{}{ctx, value}
	for _, attr := range attrs {
		args = append(args, attr)
	}
	m.Called(args...)
}

func TestUpDownCounterInterface(t *testing.T) {
	// This test verifies that MockUpDownCounter implements the UpDownCounter interface
	var _ contract.UpDownCounter = (*MockUpDownCounter)(nil)

	// Create a mock up down counter
	upDownCounter := new(MockUpDownCounter)

	// Set up expectations
	upDownCounter.On("Add", mock.Anything, int64(5), mock.Anything).Return()

	// Test the methods
	ctx := context.Background()
	upDownCounter.Add(ctx, 5, attribute.String("label1", "value1"))

	// Verify expectations
	upDownCounter.AssertExpectations(t)
}

// MockSpan is a mock implementation of the Span interface
type MockSpan struct {
	mock.Mock
}

func (m *MockSpan) SetAttributes(attrs ...attribute.KeyValue) {
	args := []interface{}{}
	for _, attr := range attrs {
		args = append(args, attr)
	}
	m.Called(args...)
}

func (m *MockSpan) SetStatus(code codes.Code, description string) {
	m.Called(code, description)
}

func (m *MockSpan) RecordError(err error, opts ...trace.EventOption) {
	args := []interface{}{err}
	for _, opt := range opts {
		args = append(args, opt)
	}
	m.Called(args...)
}

func (m *MockSpan) AddEvent(name string, opts ...trace.EventOption) {
	args := []interface{}{name}
	for _, opt := range opts {
		args = append(args, opt)
	}
	m.Called(args...)
}

func (m *MockSpan) End(opts ...trace.SpanEndOption) {
	args := []interface{}{}
	for _, opt := range opts {
		args = append(args, opt)
	}
	m.Called(args...)
}

func (m *MockSpan) GetSpan() trace.Span {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(trace.Span)
}

func TestSpanInterface(t *testing.T) {
	// This test verifies that MockSpan implements the Span interface
	var _ contract.Span = (*MockSpan)(nil)

	// Create a mock span
	span := new(MockSpan)

	// Set up expectations
	span.On("SetStatus", codes.Ok, "success").Return()
	span.On("SetAttributes", mock.Anything).Return()
	span.On("AddEvent", "test_event", mock.Anything).Return()
	span.On("End", mock.Anything).Return()
	span.On("GetSpan").Return(nil)

	// Test the methods
	span.SetStatus(codes.Ok, "success")
	span.SetAttributes(attribute.String("key", "value"))
	span.AddEvent("test_event")
	span.End()
	span.GetSpan()

	// Verify expectations
	span.AssertExpectations(t)
}

func TestMetricOptions(t *testing.T) {
	// Test MetricOption using the actual implementation
	config := &contract.MetricConfig{}

	desc := contract.WithDescription("Test metric")
	desc.Apply(config)
	assert.Equal(t, "Test metric", config.Description)

	unit := contract.WithUnit("count")
	unit.Apply(config)
	assert.Equal(t, "count", config.Unit)

	attrs := contract.WithMetricAttributes(attribute.String("env", "test"))
	attrs.Apply(config)
	assert.Equal(t, []attribute.KeyValue{attribute.String("env", "test")}, config.Attributes)
}

func TestSpanOptions(t *testing.T) {
	// Test SpanOption using the actual implementation
	config := &contract.SpanConfig{}

	kind := contract.WithSpanKind(trace.SpanKindClient)
	kind.Apply(config)
	assert.Equal(t, trace.SpanKindClient, config.Kind)

	attrs := contract.WithSpanAttributes(attribute.String("key", "value"))
	attrs.Apply(config)
	assert.Equal(t, []attribute.KeyValue{attribute.String("key", "value")}, config.Attributes)
}
