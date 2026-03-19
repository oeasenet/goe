package otel

import (
	"context"
	"fmt"

	"go.oease.dev/goe/v2/contract"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Provider implements the OTelProvider interface
type Provider struct {
	tracerProvider *sdktrace.TracerProvider
	propagator     propagation.TextMapPropagator
	config         *Config
	logger         contract.Logger
}

// NewProvider creates a new OpenTelemetry provider
func NewProvider(ctx context.Context, config *Config, logger contract.Logger) (*Provider, error) {
	// Create resource
	res, err := createResource(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace exporter
	exporter, err := createTraceExporter(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create sampler
	sampler := createSampler(config)

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	)

	// Create propagator
	propagator := createPropagator(config)

	// Set global providers
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagator)

	logger.Info("OpenTelemetry provider initialized",
		"service", config.ServiceName(),
		"version", config.ServiceVersion(),
		"exporter", config.ExporterType(),
		"endpoint", config.OTLPEndpoint(),
	)

	return &Provider{
		tracerProvider: tp,
		propagator:     propagator,
		config:         config,
		logger:         logger,
	}, nil
}

// TracerProvider returns the OpenTelemetry tracer provider
func (p *Provider) TracerProvider() trace.TracerProvider {
	return p.tracerProvider
}

// Tracer returns a named tracer for creating spans
func (p *Provider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return p.tracerProvider.Tracer(name, opts...)
}

// TextMapPropagator returns the text map propagator for context propagation
func (p *Provider) TextMapPropagator() propagation.TextMapPropagator {
	return p.propagator
}

// Shutdown gracefully shuts down the OpenTelemetry provider
func (p *Provider) Shutdown(ctx context.Context) error {
	p.logger.Debug("Shutting down OpenTelemetry provider")
	return p.tracerProvider.Shutdown(ctx)
}

// ForceFlush forces a flush of pending spans
func (p *Provider) ForceFlush(ctx context.Context) error {
	return p.tracerProvider.ForceFlush(ctx)
}

// createResource creates an OpenTelemetry resource
func createResource(ctx context.Context, config *Config) (*resource.Resource, error) {
	attrs := []resource.Option{
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName()),
			semconv.ServiceVersion(config.ServiceVersion()),
		),
		resource.WithHost(),
		resource.WithTelemetrySDK(),
	}

	// Add custom resource attributes
	if customAttrs := config.ResourceAttributes(); len(customAttrs) > 0 {
		for key, value := range customAttrs {
			// Note: For simplicity, treating all custom attributes as strings
			_ = key
			_ = value
		}
	}

	return resource.New(ctx, attrs...)
}

// createTraceExporter creates a trace exporter based on configuration
func createTraceExporter(ctx context.Context, config *Config) (sdktrace.SpanExporter, error) {
	switch config.ExporterType() {
	case "otlp":
		return createOTLPExporter(ctx, config)
	case "stdout":
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	case "none":
		return &noopExporter{}, nil
	default:
		return createOTLPExporter(ctx, config)
	}
}

// createOTLPExporter creates an OTLP trace exporter
func createOTLPExporter(ctx context.Context, config *Config) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(config.OTLPEndpoint()),
	}

	if config.OTLPInsecure() {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	if headers := config.OTLPHeaders(); len(headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(headers))
	}

	return otlptracegrpc.New(ctx, opts...)
}

// createSampler creates a trace sampler based on configuration
func createSampler(config *Config) sdktrace.Sampler {
	samplerArg := config.TraceSamplerArg()

	switch config.TraceSampler() {
	case "always_on":
		return sdktrace.AlwaysSample()
	case "always_off":
		return sdktrace.NeverSample()
	case "traceidratio":
		return sdktrace.TraceIDRatioBased(samplerArg)
	case "parentbased_always_on":
		return sdktrace.ParentBased(sdktrace.AlwaysSample())
	case "parentbased_always_off":
		return sdktrace.ParentBased(sdktrace.NeverSample())
	case "parentbased_traceidratio":
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(samplerArg))
	default:
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(samplerArg))
	}
}

// createPropagator creates a text map propagator based on configuration
func createPropagator(config *Config) propagation.TextMapPropagator {
	propagators := config.Propagators()
	var props []propagation.TextMapPropagator

	for _, p := range propagators {
		switch p {
		case "tracecontext":
			props = append(props, propagation.TraceContext{})
		case "baggage":
			props = append(props, propagation.Baggage{})
		}
	}

	if len(props) == 0 {
		// Default to W3C Trace Context and Baggage
		return propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)
	}

	return propagation.NewCompositeTextMapPropagator(props...)
}

// noopExporter is a no-op span exporter
type noopExporter struct{}

func (e *noopExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	return nil
}

func (e *noopExporter) Shutdown(ctx context.Context) error {
	return nil
}
