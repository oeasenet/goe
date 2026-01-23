package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.oease.dev/goe/v2/contract"
)

// Manager implements the MetricsManager interface using Prometheus
type Manager struct {
	mu         sync.RWMutex
	registry   *prometheus.Registry
	config     *Config
	logger     contract.Logger
	counters   map[string]*prometheusCounter
	gauges     map[string]*prometheusGauge
	histograms map[string]*prometheusHistogram
}

// NewManager creates a new metrics manager
func NewManager(config *Config, logger contract.Logger) *Manager {
	registry := prometheus.NewRegistry()

	m := &Manager{
		registry:   registry,
		config:     config,
		logger:     logger,
		counters:   make(map[string]*prometheusCounter),
		gauges:     make(map[string]*prometheusGauge),
		histograms: make(map[string]*prometheusHistogram),
	}

	// Register Go runtime metrics
	if config.EnableGoMetrics() {
		registry.MustRegister(collectors.NewGoCollector())
	}

	// Register process metrics
	if config.EnableProcessMetrics() {
		registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}

	return m
}

// Counter returns a counter metric by name, creating it if necessary
func (m *Manager) Counter(name, help string, labels ...string) contract.Counter {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.metricKey(name, labels)
	if counter, exists := m.counters[key]; exists {
		return counter
	}

	opts := prometheus.CounterOpts{
		Namespace: m.config.Namespace(),
		Subsystem: m.config.Subsystem(),
		Name:      name,
		Help:      help,
	}

	var promCounter prometheus.Counter
	var promCounterVec *prometheus.CounterVec

	if len(labels) > 0 {
		promCounterVec = prometheus.NewCounterVec(opts, labels)
		m.registry.MustRegister(promCounterVec)
	} else {
		promCounter = prometheus.NewCounter(opts)
		m.registry.MustRegister(promCounter)
	}

	counter := &prometheusCounter{
		counter:    promCounter,
		counterVec: promCounterVec,
	}
	m.counters[key] = counter
	return counter
}

// Gauge returns a gauge metric by name, creating it if necessary
func (m *Manager) Gauge(name, help string, labels ...string) contract.Gauge {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.metricKey(name, labels)
	if gauge, exists := m.gauges[key]; exists {
		return gauge
	}

	opts := prometheus.GaugeOpts{
		Namespace: m.config.Namespace(),
		Subsystem: m.config.Subsystem(),
		Name:      name,
		Help:      help,
	}

	var promGauge prometheus.Gauge
	var promGaugeVec *prometheus.GaugeVec

	if len(labels) > 0 {
		promGaugeVec = prometheus.NewGaugeVec(opts, labels)
		m.registry.MustRegister(promGaugeVec)
	} else {
		promGauge = prometheus.NewGauge(opts)
		m.registry.MustRegister(promGauge)
	}

	gauge := &prometheusGauge{
		gauge:    promGauge,
		gaugeVec: promGaugeVec,
	}
	m.gauges[key] = gauge
	return gauge
}

// Histogram returns a histogram metric by name, creating it if necessary
func (m *Manager) Histogram(name, help string, buckets []float64, labels ...string) contract.Histogram {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.metricKey(name, labels)
	if histogram, exists := m.histograms[key]; exists {
		return histogram
	}

	if buckets == nil {
		buckets = prometheus.DefBuckets
	}

	opts := prometheus.HistogramOpts{
		Namespace: m.config.Namespace(),
		Subsystem: m.config.Subsystem(),
		Name:      name,
		Help:      help,
		Buckets:   buckets,
	}

	var promHistogram prometheus.Histogram
	var promHistogramVec *prometheus.HistogramVec

	if len(labels) > 0 {
		promHistogramVec = prometheus.NewHistogramVec(opts, labels)
		m.registry.MustRegister(promHistogramVec)
	} else {
		promHistogram = prometheus.NewHistogram(opts)
		m.registry.MustRegister(promHistogram)
	}

	histogram := &prometheusHistogram{
		histogram:    promHistogram,
		histogramVec: promHistogramVec,
	}
	m.histograms[key] = histogram
	return histogram
}

// Handler returns an HTTP handler for exposing metrics
func (m *Manager) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// MustRegister registers custom collectors (must be prometheus.Collector compatible)
func (m *Manager) MustRegister(collectors ...any) {
	for _, c := range collectors {
		if pc, ok := c.(prometheus.Collector); ok {
			m.registry.MustRegister(pc)
		}
	}
}

// Registry returns the underlying Prometheus registry
func (m *Manager) Registry() *prometheus.Registry {
	return m.registry
}

// metricKey creates a unique key for a metric
func (m *Manager) metricKey(name string, labels []string) string {
	key := name
	for _, label := range labels {
		key += "_" + label
	}
	return key
}
