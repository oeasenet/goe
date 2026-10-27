package metrics

import (
	"time"

	"go.oease.dev/goe/v2/contract"
)

// streamDurationBuckets cover a second-long reply up to a ten-minute agent run.
var streamDurationBuckets = []float64{0.1, 0.5, 1, 5, 15, 30, 60, 120, 300, 600}

// StreamMetrics records the streams served by goehttp.SSE and goehttp.SendStream
// as <namespace>_http_streams_active{route} and
// <namespace>_http_stream_duration_seconds{route,outcome}. It satisfies
// goehttp.StreamObserver; GOE wires it when both modules are enabled.
type StreamMetrics struct {
	active   contract.Gauge
	duration contract.Histogram
}

// NewStreamMetrics registers the stream metrics with manager.
func NewStreamMetrics(manager contract.MetricsManager) *StreamMetrics {
	return &StreamMetrics{
		active:   manager.Gauge("http_streams_active", "Streams being served.", "route"),
		duration: manager.Histogram("http_stream_duration_seconds", "How long streams lasted, by outcome.", streamDurationBuckets, "route", "outcome"),
	}
}

// StreamStarted records a stream that began.
func (s *StreamMetrics) StreamStarted(route string) {
	s.active.WithLabelValues(route).Inc()
}

// StreamEnded records a stream that ended, with its outcome and duration.
func (s *StreamMetrics) StreamEnded(route, outcome string, d time.Duration) {
	s.active.WithLabelValues(route).Dec()
	s.duration.WithLabelValues(route, outcome).Observe(d.Seconds())
}
