package contract

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Observability defines the main observability interface
type Observability interface {
	// Metrics returns the metrics manager
	Metrics() MetricsManager

	// Tracing returns the tracing manager
	Tracing() TracingManager

	// Shutdown gracefully shuts down all observability components
	Shutdown(ctx context.Context) error
}

// MetricsManager manages metrics collection and export
type MetricsManager interface {
	// Counter creates or retrieves a counter metric
	Counter(name string, opts ...MetricOption) Counter

	// Histogram creates or retrieves a histogram metric
	Histogram(name string, opts ...MetricOption) Histogram

	// Gauge creates or retrieves a gauge metric
	Gauge(name string, opts ...MetricOption) Gauge

	// UpDownCounter creates or retrieves an up-down counter metric
	UpDownCounter(name string, opts ...MetricOption) UpDownCounter

	// GetMeter returns the underlying OpenTelemetry meter
	GetMeter() metric.Meter
}

// TracingManager manages distributed tracing
type TracingManager interface {
	// StartSpan starts a new span with the given name
	StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)

	// GetTracer returns the underlying OpenTelemetry tracer
	GetTracer() trace.Tracer
}

// Counter represents a counter metric
type Counter interface {
	// Add increments the counter by the given value
	Add(ctx context.Context, value int64, attrs ...attribute.KeyValue)

	// Inc increments the counter by 1
	Inc(ctx context.Context, attrs ...attribute.KeyValue)
}

// Histogram represents a histogram metric
type Histogram interface {
	// Record records a value in the histogram
	Record(ctx context.Context, value float64, attrs ...attribute.KeyValue)
}

// Gauge represents a gauge metric
type Gauge interface {
	// Set sets the gauge to the given value
	Set(ctx context.Context, value float64, attrs ...attribute.KeyValue)
}

// UpDownCounter represents an up-down counter metric
type UpDownCounter interface {
	// Add adds the given value to the counter (can be negative)
	Add(ctx context.Context, value int64, attrs ...attribute.KeyValue)
}

// Span represents a tracing span
type Span interface {
	// SetAttributes sets attributes on the span
	SetAttributes(attrs ...attribute.KeyValue)

	// SetStatus sets the status of the span
	SetStatus(code codes.Code, description string)

	// RecordError records an error on the span
	RecordError(err error, opts ...trace.EventOption)

	// AddEvent adds an event to the span
	AddEvent(name string, opts ...trace.EventOption)

	// End ends the span
	End(opts ...trace.SpanEndOption)

	// GetSpan returns the underlying OpenTelemetry span
	GetSpan() trace.Span
}

// MetricOption configures metric creation
type MetricOption interface {
	Apply(*MetricConfig)
}

// SpanOption configures span creation
type SpanOption interface {
	Apply(*SpanConfig)
}

// MetricConfig holds configuration for metrics
type MetricConfig struct {
	Description string
	Unit        string
	Attributes  []attribute.KeyValue
}

// SpanConfig holds configuration for spans
type SpanConfig struct {
	Kind       trace.SpanKind
	Attributes []attribute.KeyValue
}

// ObservabilityConfig defines observability-specific configuration
type ObservabilityConfig interface {
	// Enabled returns whether observability is enabled
	Enabled() bool

	// ServiceName returns the service name for tracing
	ServiceName() string

	// ServiceVersion returns the service version
	ServiceVersion() string

	// Environment returns the environment name
	Environment() string

	// Metrics returns metrics configuration
	Metrics() MetricsConfig

	// Tracing returns tracing configuration
	Tracing() TracingConfig
}

// MetricsConfig defines metrics-specific configuration
type MetricsConfig interface {
	// Enabled returns whether metrics are enabled
	Enabled() bool

	// Endpoint returns the metrics endpoint (for Prometheus)
	Endpoint() string

	// Port returns the metrics server port
	Port() int

	// Path returns the metrics endpoint path
	Path() string

	// Interval returns the metrics collection interval
	Interval() time.Duration

	// Exporters returns the list of enabled exporters
	Exporters() []string
}

// TracingConfig defines tracing-specific configuration
type TracingConfig interface {
	// Enabled returns whether tracing is enabled
	Enabled() bool

	// Endpoint returns the tracing endpoint (OTLP)
	Endpoint() string

	// SamplingRatio returns the sampling ratio (0.0 to 1.0)
	SamplingRatio() float64

	// Exporters returns the list of enabled exporters
	Exporters() []string

	// Headers returns additional headers for tracing export
	Headers() map[string]string
}

// Metric option implementations
type metricDescription string

func (d metricDescription) Apply(config *MetricConfig) {
	config.Description = string(d)
}

// WithDescription sets the metric description
func WithDescription(desc string) MetricOption {
	return metricDescription(desc)
}

type metricUnit string

func (u metricUnit) Apply(config *MetricConfig) {
	config.Unit = string(u)
}

// WithUnit sets the metric unit
func WithUnit(unit string) MetricOption {
	return metricUnit(unit)
}

type metricAttributes []attribute.KeyValue

func (a metricAttributes) Apply(config *MetricConfig) {
	config.Attributes = []attribute.KeyValue(a)
}

// WithMetricAttributes sets the metric attributes
func WithMetricAttributes(attrs ...attribute.KeyValue) MetricOption {
	return metricAttributes(attrs)
}

// Span option implementations
type spanKind trace.SpanKind

func (k spanKind) Apply(config *SpanConfig) {
	config.Kind = trace.SpanKind(k)
}

// WithSpanKind sets the span kind
func WithSpanKind(kind trace.SpanKind) SpanOption {
	return spanKind(kind)
}

type spanAttributes []attribute.KeyValue

func (a spanAttributes) Apply(config *SpanConfig) {
	config.Attributes = []attribute.KeyValue(a)
}

// WithSpanAttributes sets the span attributes
func WithSpanAttributes(attrs ...attribute.KeyValue) SpanOption {
	return spanAttributes(attrs)
}
