package observability

import (
	"context"
	"sync"

	"go.oease.dev/goe/v2/contract"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// observability implements the main Observability interface
type observability struct {
	config         contract.Config
	logger         contract.Logger
	metricsManager contract.MetricsManager
	tracingManager contract.TracingManager
	sdkManager     *SDKManager
	mu             sync.RWMutex
}

// New creates a new observability instance
func New(config contract.Config, logger contract.Logger) contract.Observability {
	obs := &observability{
		config: config,
		logger: logger,
	}

	// Initialize SDK manager
	sdkManager, err := NewSDKManager(config, logger)
	if err != nil {
		logger.Error("Failed to initialize OpenTelemetry SDK", "error", err)
		// Continue with noop implementations
		obs.metricsManager = NewMetricsManager(config, logger, func(func(context.Context) error) {})
		obs.tracingManager = NewTracingManager(config, logger, func(func(context.Context) error) {})
		return obs
	}

	obs.sdkManager = sdkManager

	// Initialize metrics manager
	obs.metricsManager = NewMetricsManager(config, logger, func(func(context.Context) error) {})

	// Initialize tracing manager
	obs.tracingManager = NewTracingManager(config, logger, func(func(context.Context) error) {})

	return obs
}

// Metrics returns the metrics manager
func (o *observability) Metrics() contract.MetricsManager {
	return o.metricsManager
}

// Tracing returns the tracing manager
func (o *observability) Tracing() contract.TracingManager {
	return o.tracingManager
}

// Shutdown gracefully shuts down all observability components
func (o *observability) Shutdown(ctx context.Context) error {
	if o.sdkManager != nil {
		return o.sdkManager.Shutdown(ctx)
	}
	return nil
}

// metricsManager implements the MetricsManager interface
type metricsManager struct {
	meter          metric.Meter
	config         contract.Config
	logger         contract.Logger
	counters       map[string]contract.Counter
	histograms     map[string]contract.Histogram
	gauges         map[string]contract.Gauge
	upDownCounters map[string]contract.UpDownCounter
	mu             sync.RWMutex
}

// NewMetricsManager creates a new metrics manager
func NewMetricsManager(config contract.Config, logger contract.Logger, addShutdownFunc func(func(context.Context) error)) contract.MetricsManager {
	serviceName := config.GetString("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = config.GetString("APP_NAME")
		if serviceName == "" {
			serviceName = "goe-app"
		}
	}

	meter := otel.Meter(serviceName)

	return &metricsManager{
		meter:          meter,
		config:         config,
		logger:         logger,
		counters:       make(map[string]contract.Counter),
		histograms:     make(map[string]contract.Histogram),
		gauges:         make(map[string]contract.Gauge),
		upDownCounters: make(map[string]contract.UpDownCounter),
	}
}

// Counter creates or retrieves a counter metric
func (m *metricsManager) Counter(name string, opts ...contract.MetricOption) contract.Counter {
	m.mu.RLock()
	if counter, exists := m.counters[name]; exists {
		m.mu.RUnlock()
		return counter
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double check
	if counter, exists := m.counters[name]; exists {
		return counter
	}

	// Apply options
	config := &contract.MetricConfig{}
	for _, opt := range opts {
		opt.Apply(config)
	}

	// Create OpenTelemetry counter
	otelCounter, err := m.meter.Int64Counter(name,
		metric.WithDescription(config.Description),
		metric.WithUnit(config.Unit),
	)
	if err != nil {
		m.logger.Error("Failed to create counter", "name", name, "error", err)
		return &noopCounter{}
	}

	counter := &counter{otelCounter: otelCounter}
	m.counters[name] = counter
	return counter
}

// Histogram creates or retrieves a histogram metric
func (m *metricsManager) Histogram(name string, opts ...contract.MetricOption) contract.Histogram {
	m.mu.RLock()
	if histogram, exists := m.histograms[name]; exists {
		m.mu.RUnlock()
		return histogram
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double check
	if histogram, exists := m.histograms[name]; exists {
		return histogram
	}

	// Apply options
	config := &contract.MetricConfig{}
	for _, opt := range opts {
		opt.Apply(config)
	}

	// Create OpenTelemetry histogram
	otelHistogram, err := m.meter.Float64Histogram(name,
		metric.WithDescription(config.Description),
		metric.WithUnit(config.Unit),
	)
	if err != nil {
		m.logger.Error("Failed to create histogram", "name", name, "error", err)
		return &noopHistogram{}
	}

	histogram := &histogram{otelHistogram: otelHistogram}
	m.histograms[name] = histogram
	return histogram
}

// Gauge creates or retrieves a gauge metric
func (m *metricsManager) Gauge(name string, opts ...contract.MetricOption) contract.Gauge {
	m.mu.RLock()
	if gauge, exists := m.gauges[name]; exists {
		m.mu.RUnlock()
		return gauge
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double check
	if gauge, exists := m.gauges[name]; exists {
		return gauge
	}

	// Apply options
	config := &contract.MetricConfig{}
	for _, opt := range opts {
		opt.Apply(config)
	}

	// Create OpenTelemetry gauge
	otelGauge, err := m.meter.Float64Gauge(name,
		metric.WithDescription(config.Description),
		metric.WithUnit(config.Unit),
	)
	if err != nil {
		m.logger.Error("Failed to create gauge", "name", name, "error", err)
		return &noopGauge{}
	}

	gauge := &gauge{otelGauge: otelGauge}
	m.gauges[name] = gauge
	return gauge
}

// UpDownCounter creates or retrieves an up-down counter metric
func (m *metricsManager) UpDownCounter(name string, opts ...contract.MetricOption) contract.UpDownCounter {
	m.mu.RLock()
	if upDownCounter, exists := m.upDownCounters[name]; exists {
		m.mu.RUnlock()
		return upDownCounter
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double check
	if upDownCounter, exists := m.upDownCounters[name]; exists {
		return upDownCounter
	}

	// Apply options
	config := &contract.MetricConfig{}
	for _, opt := range opts {
		opt.Apply(config)
	}

	// Create OpenTelemetry up-down counter
	otelUpDownCounter, err := m.meter.Int64UpDownCounter(name,
		metric.WithDescription(config.Description),
		metric.WithUnit(config.Unit),
	)
	if err != nil {
		m.logger.Error("Failed to create up-down counter", "name", name, "error", err)
		return &noopUpDownCounter{}
	}

	upDownCounter := &upDownCounter{otelUpDownCounter: otelUpDownCounter}
	m.upDownCounters[name] = upDownCounter
	return upDownCounter
}

// GetMeter returns the underlying OpenTelemetry meter
func (m *metricsManager) GetMeter() metric.Meter {
	return m.meter
}

// tracingManager implements the TracingManager interface
type tracingManager struct {
	tracer trace.Tracer
	config contract.Config
	logger contract.Logger
}

// NewTracingManager creates a new tracing manager
func NewTracingManager(config contract.Config, logger contract.Logger, addShutdownFunc func(func(context.Context) error)) contract.TracingManager {
	serviceName := config.GetString("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = config.GetString("APP_NAME")
		if serviceName == "" {
			serviceName = "goe-app"
		}
	}

	tracer := otel.Tracer(serviceName)

	return &tracingManager{
		tracer: tracer,
		config: config,
		logger: logger,
	}
}

// StartSpan starts a new span with the given name
func (t *tracingManager) StartSpan(ctx context.Context, name string, opts ...contract.SpanOption) (context.Context, contract.Span) {
	// Apply options
	config := &contract.SpanConfig{}
	for _, opt := range opts {
		opt.Apply(config)
	}

	// Create span options
	spanOpts := []trace.SpanStartOption{
		trace.WithSpanKind(config.Kind),
	}

	if len(config.Attributes) > 0 {
		spanOpts = append(spanOpts, trace.WithAttributes(config.Attributes...))
	}

	// Start span
	ctx, otelSpan := t.tracer.Start(ctx, name, spanOpts...)

	span := &span{otelSpan: otelSpan}
	return ctx, span
}

// GetTracer returns the underlying OpenTelemetry tracer
func (t *tracingManager) GetTracer() trace.Tracer {
	return t.tracer
}
