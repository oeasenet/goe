package observability

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.oease.dev/goe/v2/contract"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// SDKManager manages OpenTelemetry SDK initialization and shutdown
type SDKManager struct {
	config           contract.ObservabilityConfig
	logger           contract.Logger
	meterProvider    *metric.MeterProvider
	tracerProvider   *trace.TracerProvider
	prometheusServer *http.Server
	shutdownFuncs    []func(context.Context) error
}

// NewSDKManager creates a new SDK manager
func NewSDKManager(config contract.Config, logger contract.Logger) (*SDKManager, error) {
	obsConfig := NewConfig(config)

	if !obsConfig.Enabled() {
		logger.Info("OpenTelemetry SDK disabled")
		return &SDKManager{
			config: obsConfig,
			logger: logger,
		}, nil
	}

	sdk := &SDKManager{
		config:        obsConfig,
		logger:        logger,
		shutdownFuncs: make([]func(context.Context) error, 0),
	}

	// Initialize resource
	res, err := sdk.initResource()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize resource: %w", err)
	}

	// Initialize metrics
	if obsConfig.Metrics().Enabled() {
		if err := sdk.initMetrics(res); err != nil {
			return nil, fmt.Errorf("failed to initialize metrics: %w", err)
		}
	}

	// Initialize tracing
	if obsConfig.Tracing().Enabled() {
		if err := sdk.initTracing(res); err != nil {
			return nil, fmt.Errorf("failed to initialize tracing: %w", err)
		}
	}

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Info("OpenTelemetry SDK initialized",
		"service_name", obsConfig.ServiceName(),
		"service_version", obsConfig.ServiceVersion(),
		"environment", obsConfig.Environment(),
	)

	return sdk, nil
}

// initResource creates the OpenTelemetry resource
func (s *SDKManager) initResource() (*resource.Resource, error) {
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(s.config.ServiceName()),
		semconv.ServiceVersion(s.config.ServiceVersion()),
		semconv.DeploymentEnvironment(s.config.Environment()),
	)
	return res, nil
}

// initMetrics initializes the metrics SDK
func (s *SDKManager) initMetrics(res *resource.Resource) error {
	metricsConfig := s.config.Metrics()
	exporters := metricsConfig.Exporters()

	var readers []metric.Reader

	for _, exporterName := range exporters {
		switch exporterName {
		case "prometheus":
			reader, server, err := s.createPrometheusReader(metricsConfig)
			if err != nil {
				return fmt.Errorf("failed to create Prometheus reader: %w", err)
			}
			readers = append(readers, reader)
			s.prometheusServer = server

		case "otlp":
			reader, err := s.createOTLPMetricsReader(metricsConfig)
			if err != nil {
				return fmt.Errorf("failed to create OTLP metrics reader: %w", err)
			}
			readers = append(readers, reader)

		default:
			s.logger.Warn("Unknown metrics exporter", "exporter", exporterName)
		}
	}

	if len(readers) == 0 {
		s.logger.Warn("No metrics readers configured")
		return nil
	}

	// Create meter provider options
	opts := []metric.Option{
		metric.WithResource(res),
	}
	for _, reader := range readers {
		opts = append(opts, metric.WithReader(reader))
	}

	s.meterProvider = metric.NewMeterProvider(opts...)

	// Set global meter provider
	otel.SetMeterProvider(s.meterProvider)

	// Add shutdown function
	s.shutdownFuncs = append(s.shutdownFuncs, func(ctx context.Context) error {
		return s.meterProvider.Shutdown(ctx)
	})

	return nil
}

// createPrometheusReader creates a Prometheus metrics reader
func (s *SDKManager) createPrometheusReader(config contract.MetricsConfig) (metric.Reader, *http.Server, error) {
	// Create Prometheus exporter
	exporter, err := prometheus.New()
	if err != nil {
		return nil, nil, err
	}

	// Create HTTP server for Prometheus metrics
	mux := http.NewServeMux()
	mux.Handle(config.Path(), promhttp.Handler())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port()),
		Handler: mux,
	}

	// Start server in background
	go func() {
		s.logger.Info("Starting Prometheus metrics server",
			"port", config.Port(),
			"path", config.Path(),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Prometheus server error", "error", err)
		}
	}()

	// Add shutdown function for server
	s.shutdownFuncs = append(s.shutdownFuncs, func(ctx context.Context) error {
		return server.Shutdown(ctx)
	})

	return exporter, server, nil
}

// createOTLPMetricsReader creates an OTLP metrics reader
func (s *SDKManager) createOTLPMetricsReader(config contract.MetricsConfig) (metric.Reader, error) {
	// Create OTLP HTTP exporter
	exporter, err := otlpmetrichttp.New(context.Background(),
		otlpmetrichttp.WithEndpoint(config.Endpoint()),
		otlpmetrichttp.WithInsecure(), // TODO: Make this configurable
	)
	if err != nil {
		return nil, err
	}

	// Create periodic reader
	reader := metric.NewPeriodicReader(
		exporter,
		metric.WithInterval(config.Interval()),
	)

	// Add shutdown function
	s.shutdownFuncs = append(s.shutdownFuncs, func(ctx context.Context) error {
		return reader.Shutdown(ctx)
	})

	return reader, nil
}

// initTracing initializes the tracing SDK
func (s *SDKManager) initTracing(res *resource.Resource) error {
	tracingConfig := s.config.Tracing()
	exporters := tracingConfig.Exporters()

	var spanExporters []trace.SpanExporter

	for _, exporterName := range exporters {
		switch exporterName {
		case "otlp":
			exporter, err := s.createOTLPTraceExporter(tracingConfig)
			if err != nil {
				return fmt.Errorf("failed to create OTLP trace exporter: %w", err)
			}
			spanExporters = append(spanExporters, exporter)

		default:
			s.logger.Warn("Unknown tracing exporter", "exporter", exporterName)
		}
	}

	if len(spanExporters) == 0 {
		s.logger.Warn("No trace exporters configured")
		return nil
	}

	// Create span processors
	var spanProcessors []trace.SpanProcessor
	for _, exporter := range spanExporters {
		processor := trace.NewBatchSpanProcessor(exporter)
		spanProcessors = append(spanProcessors, processor)

		// Add shutdown function
		s.shutdownFuncs = append(s.shutdownFuncs, func(ctx context.Context) error {
			return processor.Shutdown(ctx)
		})
	}

	// Create sampler
	sampler := trace.TraceIDRatioBased(tracingConfig.SamplingRatio())

	// Create tracer provider options
	opts := []trace.TracerProviderOption{
		trace.WithResource(res),
		trace.WithSampler(sampler),
	}
	for _, processor := range spanProcessors {
		opts = append(opts, trace.WithSpanProcessor(processor))
	}

	s.tracerProvider = trace.NewTracerProvider(opts...)

	// Set global tracer provider
	otel.SetTracerProvider(s.tracerProvider)

	// Add shutdown function
	s.shutdownFuncs = append(s.shutdownFuncs, func(ctx context.Context) error {
		return s.tracerProvider.Shutdown(ctx)
	})

	return nil
}

// createOTLPTraceExporter creates an OTLP trace exporter
func (s *SDKManager) createOTLPTraceExporter(config contract.TracingConfig) (trace.SpanExporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(config.Endpoint()),
		otlptracehttp.WithInsecure(), // TODO: Make this configurable
	}

	// Add headers if configured
	headers := config.Headers()
	if len(headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(headers))
	}

	return otlptracehttp.New(context.Background(), opts...)
}

// Shutdown gracefully shuts down the SDK
func (s *SDKManager) Shutdown(ctx context.Context) error {
	var errors []error

	// Create a context with timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for _, shutdownFunc := range s.shutdownFuncs {
		if err := shutdownFunc(shutdownCtx); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	s.logger.Info("OpenTelemetry SDK shutdown completed")
	return nil
}
