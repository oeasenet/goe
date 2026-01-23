package contract

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// OTelProvider defines the interface for OpenTelemetry instrumentation
type OTelProvider interface {
	// TracerProvider returns the OpenTelemetry tracer provider
	TracerProvider() trace.TracerProvider

	// Tracer returns a named tracer for creating spans
	Tracer(name string, opts ...trace.TracerOption) trace.Tracer

	// TextMapPropagator returns the text map propagator for context propagation
	TextMapPropagator() propagation.TextMapPropagator

	// Shutdown gracefully shuts down the OpenTelemetry provider
	// This flushes any pending spans and releases resources
	Shutdown(ctx context.Context) error

	// ForceFlush forces a flush of pending spans
	ForceFlush(ctx context.Context) error
}

// OTelConfig defines configuration for the OpenTelemetry module
type OTelConfig struct {
	// Enabled determines if OpenTelemetry is active
	Enabled bool

	// ServiceName is the name of the service (default: APP_NAME)
	ServiceName string
	// ServiceVersion is the version of the service (default: APP_VERSION)
	ServiceVersion string

	// Tracing configuration
	TracesEnabled bool
	// TraceSampler is the trace sampling strategy (always_on, always_off, traceidratio, parentbased_always_on, parentbased_always_off, parentbased_traceidratio)
	TraceSampler string
	// TraceSamplerArg is the argument for ratio-based samplers (0.0-1.0)
	TraceSamplerArg float64

	// Metrics configuration
	MetricsEnabled bool

	// Exporter configuration
	// ExporterType is the exporter type (otlp, stdout, none)
	ExporterType string
	// OTLPEndpoint is the OTLP collector endpoint
	OTLPEndpoint string
	// OTLPProtocol is the OTLP protocol (grpc, http)
	OTLPProtocol string
	// OTLPHeaders are additional headers for OTLP requests
	OTLPHeaders map[string]string
	// OTLPInsecure determines if TLS should be disabled
	OTLPInsecure bool

	// Propagation
	// Propagators is the list of propagators to use (tracecontext, baggage, b3, jaeger, xray, ottrace)
	Propagators []string

	// Resource attributes
	// ResourceAttributes are additional resource attributes
	ResourceAttributes map[string]string
}

// SpanContext provides access to the current span context
type SpanContext interface {
	// TraceID returns the trace ID as a string
	TraceID() string
	// SpanID returns the span ID as a string
	SpanID() string
	// IsSampled returns true if the span is sampled
	IsSampled() bool
}

// DefaultOTelConfig returns the default OpenTelemetry configuration
func DefaultOTelConfig() OTelConfig {
	return OTelConfig{
		Enabled:         false,
		TracesEnabled:   true,
		MetricsEnabled:  true,
		TraceSampler:    "parentbased_traceidratio",
		TraceSamplerArg: 0.1,
		ExporterType:    "otlp",
		OTLPProtocol:    "grpc",
		OTLPEndpoint:    "localhost:4317",
		OTLPInsecure:    true,
		Propagators:     []string{"tracecontext", "baggage"},
	}
}
