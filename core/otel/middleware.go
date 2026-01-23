package otel

import (
	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// MiddlewareConfig defines the config for OpenTelemetry middleware
type MiddlewareConfig struct {
	// Tracer is the tracer to use for creating spans
	Tracer trace.Tracer
	// Propagators is the propagator to use for context extraction
	Propagators propagation.TextMapPropagator
	// ServerName is the name of the server
	ServerName string
	// SpanNameFormatter is a function to format the span name
	SpanNameFormatter func(fiber.Ctx) string
	// Skip is a function to skip the middleware for certain requests
	Skip func(fiber.Ctx) bool
}

// DefaultMiddlewareConfig returns the default middleware config
func DefaultMiddlewareConfig(serviceName string) MiddlewareConfig {
	return MiddlewareConfig{
		Tracer:      otel.Tracer(serviceName),
		Propagators: otel.GetTextMapPropagator(),
		ServerName:  serviceName,
		SpanNameFormatter: func(c fiber.Ctx) string {
			return c.Method() + " " + c.Route().Path
		},
	}
}

// Middleware creates a Fiber middleware for OpenTelemetry tracing
func Middleware(config ...MiddlewareConfig) fiber.Handler {
	cfg := MiddlewareConfig{}
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.Tracer == nil {
		cfg.Tracer = otel.Tracer("fiber-server")
	}
	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}
	if cfg.SpanNameFormatter == nil {
		cfg.SpanNameFormatter = func(c fiber.Ctx) string {
			return c.Method() + " " + c.Route().Path
		}
	}

	return func(c fiber.Ctx) error {
		// Skip if configured
		if cfg.Skip != nil && cfg.Skip(c) {
			return c.Next()
		}

		// Extract context from incoming request headers
		carrier := &headerCarrier{ctx: c}
		ctx := cfg.Propagators.Extract(c.Context(), carrier)

		// Create span name
		spanName := cfg.SpanNameFormatter(c)
		if spanName == "" {
			spanName = c.Method() + " " + c.Path()
		}

		// Start span
		ctx, span := cfg.Tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPRequestMethodKey.String(c.Method()),
				semconv.URLPath(c.Path()),
				semconv.URLScheme(c.Protocol()),
				semconv.ServerAddress(c.Hostname()),
				semconv.UserAgentOriginal(string(c.Request().Header.UserAgent())),
				semconv.ClientAddress(c.IP()),
			),
		)
		defer span.End()

		// Store span in context for downstream use
		c.SetContext(ctx)

		// Process request
		err := c.Next()

		// Record response attributes
		statusCode := c.Response().StatusCode()
		span.SetAttributes(semconv.HTTPResponseStatusCode(statusCode))

		// Record error if present
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.Bool("error", true))
		}

		// Set span status based on HTTP status code
		if statusCode >= 500 {
			span.SetAttributes(attribute.Bool("error", true))
		}

		return err
	}
}

// headerCarrier adapts Fiber context to TextMapCarrier interface
type headerCarrier struct {
	ctx fiber.Ctx
}

// Get returns the value for a given key from the request headers
func (c *headerCarrier) Get(key string) string {
	return c.ctx.Get(key)
}

// Set sets a key-value pair in the request headers
func (c *headerCarrier) Set(key, value string) {
	c.ctx.Request().Header.Set(key, value)
}

// Keys returns all keys in the carrier
func (c *headerCarrier) Keys() []string {
	headers := c.ctx.GetReqHeaders()
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	return keys
}

// responseHeaderCarrier adapts Fiber response headers to TextMapCarrier interface
type responseHeaderCarrier struct {
	ctx fiber.Ctx
}

// Get returns the value for a given key from the response headers
func (c *responseHeaderCarrier) Get(key string) string {
	return string(c.ctx.Response().Header.Peek(key))
}

// Set sets a key-value pair in the response headers
func (c *responseHeaderCarrier) Set(key, value string) {
	c.ctx.Response().Header.Set(key, value)
}

// Keys returns all keys in the carrier
func (c *responseHeaderCarrier) Keys() []string {
	headers := c.ctx.GetRespHeaders()
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	return keys
}
