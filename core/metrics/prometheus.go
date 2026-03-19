package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.oease.dev/goe/v2/contract"
)

// prometheusCounter wraps a Prometheus counter
type prometheusCounter struct {
	counter    prometheus.Counter
	counterVec *prometheus.CounterVec
}

// Inc increments the counter by 1
func (c *prometheusCounter) Inc() {
	if c.counter != nil {
		c.counter.Inc()
	}
}

// Add adds the given value to the counter
func (c *prometheusCounter) Add(v float64) {
	if c.counter != nil {
		c.counter.Add(v)
	}
}

// WithLabelValues returns a counter with the given label values
func (c *prometheusCounter) WithLabelValues(lvs ...string) contract.Counter {
	if c.counterVec == nil {
		return c
	}
	return &prometheusCounter{
		counter: c.counterVec.WithLabelValues(lvs...),
	}
}

// prometheusGauge wraps a Prometheus gauge
type prometheusGauge struct {
	gauge    prometheus.Gauge
	gaugeVec *prometheus.GaugeVec
}

// Set sets the gauge to the given value
func (g *prometheusGauge) Set(v float64) {
	if g.gauge != nil {
		g.gauge.Set(v)
	}
}

// Inc increments the gauge by 1
func (g *prometheusGauge) Inc() {
	if g.gauge != nil {
		g.gauge.Inc()
	}
}

// Dec decrements the gauge by 1
func (g *prometheusGauge) Dec() {
	if g.gauge != nil {
		g.gauge.Dec()
	}
}

// Add adds the given value to the gauge
func (g *prometheusGauge) Add(v float64) {
	if g.gauge != nil {
		g.gauge.Add(v)
	}
}

// Sub subtracts the given value from the gauge
func (g *prometheusGauge) Sub(v float64) {
	if g.gauge != nil {
		g.gauge.Sub(v)
	}
}

// WithLabelValues returns a gauge with the given label values
func (g *prometheusGauge) WithLabelValues(lvs ...string) contract.Gauge {
	if g.gaugeVec == nil {
		return g
	}
	return &prometheusGauge{
		gauge: g.gaugeVec.WithLabelValues(lvs...),
	}
}

// prometheusHistogram wraps a Prometheus histogram
type prometheusHistogram struct {
	histogram    prometheus.Histogram
	histogramVec *prometheus.HistogramVec
}

// Observe adds a single observation to the histogram
func (h *prometheusHistogram) Observe(v float64) {
	if h.histogram != nil {
		h.histogram.Observe(v)
	}
}

// WithLabelValues returns a histogram with the given label values
func (h *prometheusHistogram) WithLabelValues(lvs ...string) contract.Histogram {
	if h.histogramVec == nil {
		return h
	}
	return &prometheusHistogram{
		histogram: h.histogramVec.WithLabelValues(lvs...).(prometheus.Histogram),
	}
}
