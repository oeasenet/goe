package observability

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// counter implements the Counter interface
type counter struct {
	otelCounter metric.Int64Counter
}

// Add increments the counter by the given value
func (c *counter) Add(ctx context.Context, value int64, attrs ...attribute.KeyValue) {
	c.otelCounter.Add(ctx, value, metric.WithAttributes(attrs...))
}

// Inc increments the counter by 1
func (c *counter) Inc(ctx context.Context, attrs ...attribute.KeyValue) {
	c.Add(ctx, 1, attrs...)
}

// histogram implements the Histogram interface
type histogram struct {
	otelHistogram metric.Float64Histogram
}

// Record records a value in the histogram
func (h *histogram) Record(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	h.otelHistogram.Record(ctx, value, metric.WithAttributes(attrs...))
}

// gauge implements the Gauge interface
type gauge struct {
	otelGauge metric.Float64Gauge
}

// Set sets the gauge to the given value
func (g *gauge) Set(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	g.otelGauge.Record(ctx, value, metric.WithAttributes(attrs...))
}

// upDownCounter implements the UpDownCounter interface
type upDownCounter struct {
	otelUpDownCounter metric.Int64UpDownCounter
}

// Add adds the given value to the counter (can be negative)
func (u *upDownCounter) Add(ctx context.Context, value int64, attrs ...attribute.KeyValue) {
	u.otelUpDownCounter.Add(ctx, value, metric.WithAttributes(attrs...))
}

// Noop implementations for error cases

// noopCounter is a no-op implementation of Counter
type noopCounter struct{}

func (n *noopCounter) Add(ctx context.Context, value int64, attrs ...attribute.KeyValue) {}
func (n *noopCounter) Inc(ctx context.Context, attrs ...attribute.KeyValue)              {}

// noopHistogram is a no-op implementation of Histogram
type noopHistogram struct{}

func (n *noopHistogram) Record(ctx context.Context, value float64, attrs ...attribute.KeyValue) {}

// noopGauge is a no-op implementation of Gauge
type noopGauge struct{}

func (n *noopGauge) Set(ctx context.Context, value float64, attrs ...attribute.KeyValue) {}

// noopUpDownCounter is a no-op implementation of UpDownCounter
type noopUpDownCounter struct{}

func (n *noopUpDownCounter) Add(ctx context.Context, value int64, attrs ...attribute.KeyValue) {}
