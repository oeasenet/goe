package http

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2/contract"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// MetricsMiddleware creates a middleware that automatically collects HTTP metrics
func MetricsMiddleware(metrics contract.MetricsManager, tracing contract.TracingManager) fiber.Handler {
	// Create metrics
	requestCounter := metrics.Counter("http_requests_total",
		contract.WithDescription("Total number of HTTP requests"),
		contract.WithUnit("requests"),
	)

	requestDuration := metrics.Histogram("http_request_duration_seconds",
		contract.WithDescription("HTTP request duration in seconds"),
		contract.WithUnit("seconds"),
	)

	requestSize := metrics.Histogram("http_request_size_bytes",
		contract.WithDescription("HTTP request size in bytes"),
		contract.WithUnit("bytes"),
	)

	responseSize := metrics.Histogram("http_response_size_bytes",
		contract.WithDescription("HTTP response size in bytes"),
		contract.WithUnit("bytes"),
	)

	activeConnections := metrics.UpDownCounter("http_active_connections",
		contract.WithDescription("Number of active HTTP connections"),
		contract.WithUnit("connections"),
	)

	return func(c fiber.Ctx) error {
		start := time.Now()
		ctx := c.Context()

		// Start tracing span
		ctx, span := tracing.StartSpan(ctx, "HTTP "+c.Method()+" "+c.Route().Path,
			contract.WithSpanKind(trace.SpanKindServer),
			contract.WithSpanAttributes(
				attribute.String("http.method", c.Method()),
				attribute.String("http.route", c.Route().Path),
				attribute.String("http.url", c.OriginalURL()),
				attribute.String("http.scheme", c.Protocol()),
				attribute.String("http.host", c.Hostname()),
				attribute.String("http.user_agent", c.Get("User-Agent")),
				attribute.String("http.remote_addr", c.IP()),
			),
		)
		defer span.End()

		// Update context in fiber
		c.SetContext(ctx)

		// Record request size
		if contentLength := c.Get("Content-Length"); contentLength != "" {
			if size, err := strconv.ParseFloat(contentLength, 64); err == nil {
				requestSize.Record(ctx, size,
					attribute.String("method", c.Method()),
					attribute.String("route", c.Route().Path),
				)
			}
		}

		// Increment active connections
		activeConnections.Add(ctx, 1, attribute.String("server", "http"))

		// Continue to next middleware
		err := c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()
		status := c.Response().StatusCode()
		statusClass := strconv.Itoa(status/100) + "xx"

		// Common attributes
		attrs := []attribute.KeyValue{
			attribute.String("method", c.Method()),
			attribute.String("route", c.Route().Path),
			attribute.String("status", strconv.Itoa(status)),
			attribute.String("status_class", statusClass),
		}

		// Record metrics
		requestCounter.Inc(ctx, attrs...)
		requestDuration.Record(ctx, duration, attrs...)

		// Record response size
		responseSize.Record(ctx, float64(len(c.Response().Body())), attrs...)

		// Update span with response information
		span.SetAttributes(
			attribute.Int("http.status_code", status),
			attribute.Int("http.response_size", len(c.Response().Body())),
			attribute.Float64("http.duration", duration),
		)

		// Set span status based on HTTP status
		if status >= 400 {
			span.SetStatus(codes.Error, "HTTP "+strconv.Itoa(status))
			if err != nil {
				span.RecordError(err)
			}
		} else {
			span.SetStatus(codes.Ok, "HTTP "+strconv.Itoa(status))
		}

		// Decrement active connections
		activeConnections.Add(ctx, -1, attribute.String("server", "http"))

		return err
	}
}

// CreateMetricsMiddleware creates metrics middleware if observability components are available
func CreateMetricsMiddleware(services Services) fiber.Handler {
	if services.Metrics != nil && services.Tracing != nil {
		return MetricsMiddleware(services.Metrics, services.Tracing)
	}
	// Return no-op middleware if observability is not available
	return func(c fiber.Ctx) error {
		return c.Next()
	}
}
