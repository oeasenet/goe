# 19. Observability & Monitoring 📊

Observability is crucial for understanding the behavior and performance of your applications in production. Goe's observability module provides comprehensive metrics collection and distributed tracing capabilities built on [OpenTelemetry](https://opentelemetry.io/), supporting industry-standard exporters like Prometheus and OTLP.

## Enabling the Observability Module

To use observability capabilities, enable the `Observability` module during application initialization:

```go
package main

import (
	"go.oease.dev/goe/v2"
)

func main() {
	app := goe.New(goe.Options{
		WithObservability: true, // Enable the Observability module
		// ... other options
	})
	goe.Run()
}
```

## Configuration

Observability behavior is configured via environment variables, typically prefixed with `OTEL_`.

### Core Configuration

* **`OTEL_ENABLED`**: Master switch to enable/disable all observability features. Default: `false`.
* **`OTEL_SERVICE_NAME`**: The service name for metrics and tracing. Falls back to `APP_NAME` if not set. Default: `"goe-app"`.
* **`OTEL_SERVICE_VERSION`**: The service version. Falls back to `APP_VERSION` if not set. Default: `"1.0.0"`.
* **`OTEL_ENVIRONMENT`**: The deployment environment. Falls back to `GOE_ENV` if not set. Default: `"dev"`.

### Metrics Configuration

* **`OTEL_METRICS_ENABLED`**: Enable/disable metrics collection. Falls back to `OTEL_ENABLED` if not set.
* **`OTEL_METRICS_EXPORTERS`**: Comma-separated list of metrics exporters. Default: `"prometheus"`.
  * Supported exporters: `prometheus`, `otlp`
* **`OTEL_METRICS_PORT`**: Port for the Prometheus metrics server. Default: `9090`.
* **`OTEL_METRICS_PATH`**: Path for the Prometheus metrics endpoint. Default: `"/metrics"`.
* **`OTEL_METRICS_INTERVAL`**: Metrics collection interval for OTLP exporter. Default: `30s`.
* **`OTEL_METRICS_ENDPOINT`**: OTLP metrics endpoint URL. Default: `"http://localhost:4318"`.

### Tracing Configuration

* **`OTEL_TRACING_ENABLED`**: Enable/disable distributed tracing. Falls back to `OTEL_ENABLED` if not set.
* **`OTEL_TRACING_EXPORTERS`**: Comma-separated list of tracing exporters. Default: `"otlp"`.
  * Supported exporters: `otlp`
* **`OTEL_TRACING_ENDPOINT`**: OTLP tracing endpoint URL. Falls back to `OTEL_EXPORTER_OTLP_ENDPOINT`. Default: `"http://localhost:4318"`.
* **`OTEL_TRACING_SAMPLING_RATIO`**: Sampling ratio (0.0 to 1.0). Default: `1.0` (100% sampling).
* **`OTEL_TRACING_HEADERS_*`**: Additional headers for tracing export. Example: `OTEL_TRACING_HEADERS_AUTHORIZATION=Bearer token123`.

### Example Configuration

```bash
# Enable observability
OTEL_ENABLED=true
OTEL_SERVICE_NAME=my-goe-app
OTEL_SERVICE_VERSION=1.2.3
OTEL_ENVIRONMENT=production

# Metrics configuration
OTEL_METRICS_ENABLED=true
OTEL_METRICS_EXPORTERS=prometheus,otlp
OTEL_METRICS_PORT=9090
OTEL_METRICS_PATH=/metrics

# Tracing configuration
OTEL_TRACING_ENABLED=true
OTEL_TRACING_EXPORTERS=otlp
OTEL_TRACING_ENDPOINT=http://jaeger:4318
OTEL_TRACING_SAMPLING_RATIO=0.1
```

## Accessing Observability Components

### 1. Dependency Injection (Recommended)

```go
import (
	"context"
	"go.oease.dev/goe/v2/contract"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type UserService struct {
	metrics contract.MetricsManager
	tracing contract.TracingManager
	logger  contract.Logger
}

func NewUserService(
	metrics contract.MetricsManager,
	tracing contract.TracingManager,
	logger contract.Logger,
) *UserService {
	return &UserService{
		metrics: metrics,
		tracing: tracing,
		logger:  logger,
	}
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
	// Start a span for this operation
	ctx, span := s.tracing.StartSpan(ctx, "user.get",
		contract.WithSpanKind(trace.SpanKindServer),
		contract.WithSpanAttributes(
			attribute.String("user.id", userID),
		),
	)
	defer span.End()

	// Increment request counter
	requestCounter := s.metrics.Counter("user_requests_total",
		contract.WithDescription("Total user requests"),
		contract.WithUnit("requests"),
	)
	requestCounter.Inc(ctx, attribute.String("operation", "get"))

	// Record request duration
	requestDuration := s.metrics.Histogram("user_request_duration",
		contract.WithDescription("User request duration"),
		contract.WithUnit("seconds"),
	)

	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		requestDuration.Record(ctx, duration,
			attribute.String("operation", "get"),
		)
	}()

	// Your business logic here
	user, err := s.fetchUserFromDatabase(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to fetch user")
		return nil, err
	}

	span.SetAttributes(attribute.String("user.name", user.Name))
	span.SetStatus(codes.Ok, "User fetched successfully")

	return user, nil
}
```

### 2. Global Access

For quick access to observability components, you can use the global accessors:

```go
import (
	"context"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func TrackUserAction(ctx context.Context, userID string, action string) {
	// Access observability components globally
	metrics := goe.Metrics()
	tracing := goe.Tracing()

	// Start a span
	ctx, span := tracing.StartSpan(ctx, "user.action",
		contract.WithSpanAttributes(
			attribute.String("user.id", userID),
			attribute.String("action", action),
		),
	)
	defer span.End()

	// Record metrics
	actionCounter := metrics.Counter("user_actions_total",
		contract.WithDescription("Total user actions"),
	)
	actionCounter.Inc(ctx, 
		attribute.String("action", action),
		attribute.String("user.id", userID),
	)

	// Your business logic here
	span.SetStatus(codes.Ok, "Action completed")
}

// You can also access the full observability interface
func GetObservabilityInstance() contract.Observability {
	return goe.Observability()
}
```

**Note**: Global accessors will panic if the observability module is not enabled. Use dependency injection for better error handling and testability.

### 3. Full Observability Interface

```go
type MonitoringService struct {
	observability contract.Observability
}

func NewMonitoringService(obs contract.Observability) *MonitoringService {
	return &MonitoringService{observability: obs}
}

func (s *MonitoringService) TrackSystemMetrics(ctx context.Context) {
	metrics := s.observability.Metrics()

	// Create various metric types
	cpuGauge := metrics.Gauge("system_cpu_usage",
		contract.WithDescription("Current CPU usage percentage"),
		contract.WithUnit("percent"),
	)

	memoryGauge := metrics.Gauge("system_memory_usage",
		contract.WithDescription("Current memory usage"),
		contract.WithUnit("bytes"),
	)

	connectionCounter := metrics.UpDownCounter("active_connections",
		contract.WithDescription("Number of active connections"),
		contract.WithUnit("connections"),
	)

	// Update metrics
	cpuGauge.Set(ctx, 75.5, attribute.String("host", "server-1"))
	memoryGauge.Set(ctx, 1024*1024*512, attribute.String("host", "server-1"))
	connectionCounter.Add(ctx, 1) // New connection
}
```

## Metrics Types

### Counter
Monotonically increasing values (e.g., request count, error count):

```go
counter := metrics.Counter("http_requests_total",
	contract.WithDescription("Total HTTP requests"),
	contract.WithUnit("requests"),
)

// Increment by 1
counter.Inc(ctx, attribute.String("method", "GET"))

// Increment by specific value
counter.Add(ctx, 5, attribute.String("method", "POST"))
```

### Histogram
Distribution of values (e.g., request duration, response size):

```go
histogram := metrics.Histogram("http_request_duration",
	contract.WithDescription("HTTP request duration"),
	contract.WithUnit("seconds"),
)

histogram.Record(ctx, 0.123, 
	attribute.String("method", "GET"),
	attribute.String("status", "200"),
)
```

### Gauge
Current value that can go up or down (e.g., memory usage, active connections):

```go
gauge := metrics.Gauge("memory_usage",
	contract.WithDescription("Current memory usage"),
	contract.WithUnit("bytes"),
)

gauge.Set(ctx, 1024*1024*256, attribute.String("process", "worker"))
```

### UpDownCounter
Counter that can increase or decrease (e.g., queue size, active sessions):

```go
upDownCounter := metrics.UpDownCounter("queue_size",
	contract.WithDescription("Current queue size"),
	contract.WithUnit("items"),
)

upDownCounter.Add(ctx, 1)  // Item added to queue
upDownCounter.Add(ctx, -1) // Item removed from queue
```

## Distributed Tracing

### Creating Spans

```go
// Start a root span
ctx, span := tracing.StartSpan(ctx, "process_order",
	contract.WithSpanKind(trace.SpanKindServer),
	contract.WithSpanAttributes(
		attribute.String("order.id", orderID),
		attribute.String("customer.id", customerID),
	),
)
defer span.End()

// Add attributes during execution
span.SetAttributes(
	attribute.Float64("order.amount", order.Amount),
	attribute.String("payment.method", order.PaymentMethod),
)

// Record events
span.AddEvent("Order validation started")
span.AddEvent("Payment processing initiated")

// Handle errors
if err != nil {
	span.RecordError(err)
	span.SetStatus(codes.Error, "Order processing failed")
	return err
}

span.SetStatus(codes.Ok, "Order processed successfully")
```

### Nested Spans

```go
func (s *OrderService) ProcessOrder(ctx context.Context, order *Order) error {
	ctx, span := s.tracing.StartSpan(ctx, "order.process")
	defer span.End()

	// Validate order
	if err := s.validateOrder(ctx, order); err != nil {
		span.RecordError(err)
		return err
	}

	// Process payment
	if err := s.processPayment(ctx, order); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

func (s *OrderService) validateOrder(ctx context.Context, order *Order) error {
	ctx, span := s.tracing.StartSpan(ctx, "order.validate",
		contract.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	// Validation logic here
	span.SetAttributes(attribute.Bool("order.valid", true))
	return nil
}

func (s *OrderService) processPayment(ctx context.Context, order *Order) error {
	ctx, span := s.tracing.StartSpan(ctx, "payment.process",
		contract.WithSpanKind(trace.SpanKindClient),
		contract.WithSpanAttributes(
			attribute.String("payment.provider", "stripe"),
		),
	)
	defer span.End()

	// Payment processing logic here
	return nil
}
```

## Integration with HTTP Module

When both HTTP and Observability modules are enabled, observability components are automatically available in HTTP handlers through dependency injection. This provides seamless access to metrics and tracing without additional configuration.

### Automatic Integration

```go
package main

import (
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/core/http"
)

func main() {
	app := goe.New(goe.Options{
		WithHTTP:          true, // Enable HTTP module
		WithObservability: true, // Enable Observability module
		Invokers: []any{
			http.RegisterRoutes(func(r http.RouteRegistrar) {
				// Observability components are automatically available
				setupRoutes(r)
			}),
		},
	})
	goe.Run()
}

func setupRoutes(r http.RouteRegistrar) {
	r.HTTP.App().Get("/api/users/:id", func(c fiber.Ctx) error {
		// Access observability components from context
		metrics := http.GetMetrics(c)
		tracing := http.GetTracing(c)

		// Use metrics and tracing in your handler
		if metrics != nil {
			counter := metrics.Counter("api_requests_total")
			counter.Inc(c.Context())
		}

		if tracing != nil {
			ctx, span := tracing.StartSpan(c.Context(), "get_user")
			defer span.End()
			// Your handler logic here
		}

		return c.JSON(map[string]string{"message": "User data"})
	})
}
```

### Manual Integration

You can also manually instrument HTTP requests for more control:

```go
func SetupRoutes(app contract.HTTPKernel, metrics contract.MetricsManager, tracing contract.TracingManager) {
	// Create metrics for HTTP requests
	httpCounter := metrics.Counter("http_requests_total")
	httpDuration := metrics.Histogram("http_request_duration")

	app.Get("/api/users/:id", func(c contract.Context) error {
		ctx := c.Context()
		userID := c.Params("id")

		// Start span for this request
		ctx, span := tracing.StartSpan(ctx, "GET /api/users/:id",
			contract.WithSpanKind(trace.SpanKindServer),
			contract.WithSpanAttributes(
				attribute.String("http.method", "GET"),
				attribute.String("http.route", "/api/users/:id"),
				attribute.String("user.id", userID),
			),
		)
		defer span.End()

		start := time.Now()
		defer func() {
			duration := time.Since(start).Seconds()
			status := c.Response().StatusCode()

			// Record metrics
			httpCounter.Inc(ctx,
				attribute.String("method", "GET"),
				attribute.String("route", "/api/users/:id"),
				attribute.Int("status", status),
			)

			httpDuration.Record(ctx, duration,
				attribute.String("method", "GET"),
				attribute.String("route", "/api/users/:id"),
			)

			// Update span with response info
			span.SetAttributes(attribute.Int("http.status_code", status))
			if status >= 400 {
				span.SetStatus(codes.Error, "HTTP error")
			} else {
				span.SetStatus(codes.Ok, "Request completed")
			}
		}()

		// Your handler logic here
		user, err := getUserByID(ctx, userID)
		if err != nil {
			span.RecordError(err)
			return c.Status(500).JSON(map[string]string{"error": "Internal server error"})
		}

		return c.JSON(user)
	})
}
```

## Best Practices

### 1. Metric Naming
Follow OpenTelemetry semantic conventions:

```go
// Good: Descriptive names with units
httpRequestsTotal := metrics.Counter("http_requests_total",
	contract.WithDescription("Total number of HTTP requests"),
	contract.WithUnit("requests"),
)

dbConnectionsActive := metrics.Gauge("db_connections_active",
	contract.WithDescription("Number of active database connections"),
	contract.WithUnit("connections"),
)

// Avoid: Vague names without context
counter := metrics.Counter("requests") // Too generic
```

### 2. Attribute Usage
Use consistent and meaningful attributes:

```go
// Good: Consistent attribute naming
span.SetAttributes(
	attribute.String("service.name", "user-service"),
	attribute.String("service.version", "1.2.3"),
	attribute.String("user.id", userID),
	attribute.String("operation.type", "read"),
)

// Avoid: Inconsistent or high-cardinality attributes
span.SetAttributes(
	attribute.String("timestamp", time.Now().String()), // High cardinality
	attribute.String("UserID", userID),                 // Inconsistent naming
)
```

### 3. Error Handling
Always record errors in spans:

```go
user, err := s.userRepo.GetByID(ctx, userID)
if err != nil {
	span.RecordError(err)
	span.SetStatus(codes.Error, "Failed to fetch user")
	return nil, err
}
```

### 4. Resource Management
Ensure proper cleanup:

```go
func (s *Service) Shutdown(ctx context.Context) error {
	// Observability components are automatically shut down by the module
	// But you can add custom cleanup here if needed
	return nil
}
```

## Monitoring Setup

### Prometheus + Grafana

1. **Prometheus Configuration** (`prometheus.yml`):
```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'goe-app'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 5s
    metrics_path: /metrics
```

2. **Grafana Dashboard**: Import or create dashboards to visualize your metrics.

### Jaeger Tracing

1. **Start Jaeger**:
```bash
docker run -d --name jaeger \
  -p 16686:16686 \
  -p 14250:14250 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

2. **Configure your app**:
```bash
OTEL_TRACING_ENDPOINT=http://localhost:4318
```

## Troubleshooting

### Common Issues

1. **Metrics not appearing in Prometheus**:
   - Check that `OTEL_METRICS_ENABLED=true`
   - Verify the metrics port is accessible
   - Ensure Prometheus is scraping the correct endpoint

2. **Traces not appearing in Jaeger**:
   - Check that `OTEL_TRACING_ENABLED=true`
   - Verify the OTLP endpoint is correct
   - Check sampling ratio (set to 1.0 for testing)

3. **High memory usage**:
   - Reduce sampling ratio for tracing
   - Limit high-cardinality attributes
   - Configure appropriate metric retention

### Debug Mode

Enable debug logging to troubleshoot observability issues:

```bash
OTEL_LOG_LEVEL=debug
```

## Performance Considerations

- **Sampling**: Use appropriate sampling ratios for tracing in production
- **Cardinality**: Limit the number of unique attribute combinations
- **Buffering**: OpenTelemetry batches exports automatically
- **Resource Usage**: Monitor the observability overhead itself

The observability module is designed to have minimal performance impact while providing comprehensive insights into your application's behavior.
