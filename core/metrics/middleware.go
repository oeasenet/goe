package metrics

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2/contract"
)

// Middleware creates a Fiber middleware that records HTTP metrics
func Middleware(manager contract.MetricsManager) fiber.Handler {
	// Create HTTP metrics
	requestsTotal := manager.Counter(
		"http_requests_total",
		"Total number of HTTP requests",
		"method", "path", "status",
	)

	requestDuration := manager.Histogram(
		"http_request_duration_seconds",
		"HTTP request duration in seconds",
		contract.DefaultHTTPRequestDurationBuckets(),
		"method", "path",
	)

	requestsInFlight := manager.Gauge(
		"http_requests_in_flight",
		"Current number of HTTP requests being processed",
	)

	return func(c fiber.Ctx) error {
		// Skip metrics endpoint to avoid infinite recursion
		if c.Path() == "/metrics" {
			return c.Next()
		}

		start := time.Now()

		// Track in-flight requests
		requestsInFlight.Inc()
		defer requestsInFlight.Dec()

		// Process request
		err := c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		method := c.Method()
		path := c.Route().Path // Use route pattern, not actual path
		if path == "" {
			path = c.Path()
		}
		status := strconv.Itoa(c.Response().StatusCode())

		requestsTotal.WithLabelValues(method, path, status).Inc()
		requestDuration.WithLabelValues(method, path).Observe(duration)

		return err
	}
}

// CacheMetrics provides cache operation metrics
type CacheMetrics struct {
	operations contract.Counter
	hits       contract.Counter
	misses     contract.Counter
}

// NewCacheMetrics creates cache metrics
func NewCacheMetrics(manager contract.MetricsManager) *CacheMetrics {
	return &CacheMetrics{
		operations: manager.Counter(
			"cache_operations_total",
			"Total number of cache operations",
			"operation", "status",
		),
		hits: manager.Counter(
			"cache_hits_total",
			"Total number of cache hits",
		),
		misses: manager.Counter(
			"cache_misses_total",
			"Total number of cache misses",
		),
	}
}

// RecordOperation records a cache operation
func (m *CacheMetrics) RecordOperation(operation, status string) {
	m.operations.WithLabelValues(operation, status).Inc()
}

// RecordHit records a cache hit
func (m *CacheMetrics) RecordHit() {
	m.hits.Inc()
}

// RecordMiss records a cache miss
func (m *CacheMetrics) RecordMiss() {
	m.misses.Inc()
}

// DatabaseMetrics provides database operation metrics
type DatabaseMetrics struct {
	queries  contract.Counter
	duration contract.Histogram
}

// NewDatabaseMetrics creates database metrics
func NewDatabaseMetrics(manager contract.MetricsManager) *DatabaseMetrics {
	return &DatabaseMetrics{
		queries: manager.Counter(
			"db_queries_total",
			"Total number of database queries",
			"operation", "table",
		),
		duration: manager.Histogram(
			"db_query_duration_seconds",
			"Database query duration in seconds",
			[]float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
			"operation",
		),
	}
}

// RecordQuery records a database query
func (m *DatabaseMetrics) RecordQuery(operation, table string) {
	m.queries.WithLabelValues(operation, table).Inc()
}

// RecordDuration records query duration
func (m *DatabaseMetrics) RecordDuration(operation string, duration float64) {
	m.duration.WithLabelValues(operation).Observe(duration)
}
