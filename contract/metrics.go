package contract

import (
	"net/http"
)

// MetricsManager defines the interface for metrics collection and exposure
type MetricsManager interface {
	// Counter returns a counter metric by name, creating it if necessary
	// Labels are key-value pairs that will be applied to all observations
	Counter(name, help string, labels ...string) Counter

	// Gauge returns a gauge metric by name, creating it if necessary
	// Labels are key-value pairs that will be applied to all observations
	Gauge(name, help string, labels ...string) Gauge

	// Histogram returns a histogram metric by name, creating it if necessary
	// Labels are key-value pairs that will be applied to all observations
	Histogram(name, help string, buckets []float64, labels ...string) Histogram

	// Handler returns an HTTP handler for exposing metrics
	Handler() http.Handler

	// MustRegister registers custom collectors (must be prometheus.Collector compatible)
	MustRegister(collectors ...any)
}

// Counter represents a monotonically increasing counter metric
type Counter interface {
	// Inc increments the counter by 1
	Inc()
	// Add adds the given value to the counter (must be >= 0)
	Add(float64)
	// WithLabelValues returns a counter with the given label values
	WithLabelValues(lvs ...string) Counter
}

// Gauge represents a metric that can go up and down
type Gauge interface {
	// Set sets the gauge to the given value
	Set(float64)
	// Inc increments the gauge by 1
	Inc()
	// Dec decrements the gauge by 1
	Dec()
	// Add adds the given value to the gauge
	Add(float64)
	// Sub subtracts the given value from the gauge
	Sub(float64)
	// WithLabelValues returns a gauge with the given label values
	WithLabelValues(lvs ...string) Gauge
}

// Histogram represents a histogram metric that samples observations
type Histogram interface {
	// Observe adds a single observation to the histogram
	Observe(float64)
	// WithLabelValues returns a histogram with the given label values
	WithLabelValues(lvs ...string) Histogram
}

// MetricsConfig defines configuration for the metrics module
type MetricsConfig struct {
	// Enabled determines if metrics collection is active
	Enabled bool
	// Path is the path for the metrics endpoint (default: /metrics)
	Path string
	// Namespace is the prefix for all metrics (default: goe)
	Namespace string
	// Subsystem is an optional subsystem name
	Subsystem string
	// EnableGoMetrics enables Go runtime metrics
	EnableGoMetrics bool
	// EnableProcessMetrics enables process metrics
	EnableProcessMetrics bool
	// HTTPRequestDurationBuckets defines histogram buckets for HTTP request duration
	HTTPRequestDurationBuckets []float64
}

// DefaultHTTPRequestDurationBuckets returns default histogram buckets for HTTP request duration
func DefaultHTTPRequestDurationBuckets() []float64 {
	return []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
}
