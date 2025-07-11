# Observability

GOE provides comprehensive built-in observability features including automatic metrics collection, distributed tracing, and structured logging across all modules. This guide covers the complete observability implementation.

## Overview

GOE's observability features include:
- **Automatic Metrics**: Built-in metrics collection for all modules (HTTP, Database, Cache, Events, MongoDB)
- **Distributed Tracing**: OpenTelemetry-based tracing across all operations
- **Structured Logging**: Zap-based logging with correlation
- **Prometheus Export**: Standard Prometheus metrics exposition
- **Zero Configuration**: Works out-of-the-box with sensible defaults

## Quick Start

Enable observability in your GOE application:

```go
package main

import "go.oease.dev/goe/v2"

func main() {
    app := goe.New(goe.Options{
        WithHTTP:          true,
        WithDB:            true,
        WithCache:         true,
        WithEvent:         true,
        WithObservability: true, // 🎯 This enables comprehensive metrics!
    })
    
    goe.Run()
}
```

With observability enabled, metrics are automatically available at `http://localhost:9090/metrics`.

## Configuration

Configure observability through environment variables:

```bash
# Core Observability
OTEL_ENABLED=true
OTEL_SERVICE_NAME=my-goe-app
OTEL_SERVICE_VERSION=1.0.0
OTEL_ENVIRONMENT=production

# Metrics Configuration
OTEL_METRICS_ENABLED=true
OTEL_METRICS_PORT=9090
OTEL_METRICS_PATH=/metrics
OTEL_METRICS_EXPORTERS=prometheus

# Tracing Configuration
OTEL_TRACING_ENABLED=true
OTEL_TRACING_ENDPOINT=http://localhost:4318
OTEL_TRACING_SAMPLING_RATIO=1.0

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

## Automatic Module Metrics

GOE automatically collects comprehensive metrics for all enabled modules:

### HTTP Module Metrics

Automatically collected for all HTTP requests:

```prometheus
# Request tracking
http_requests_total{method="GET",route="/",status="200",status_class="2xx"} 42
http_request_duration_seconds_bucket{method="GET",route="/",le="0.1"} 40
http_request_size_bytes_bucket{method="GET",route="/",le="1024"} 35
http_response_size_bytes_bucket{method="GET",route="/",le="1024"} 40

# Connection tracking
http_active_connections{server="http"} 5
```

### Database Module Metrics

Automatically collected via GORM plugin:

```prometheus
# Query performance
db_queries_total{operation="query",table="users"} 150
db_query_duration_seconds_bucket{operation="query",table="users",le="0.01"} 145
db_errors_total{operation="query",table="users"} 2

# Connection monitoring
db_connections_active 10

# Migration tracking
db_migrations_total{operation="auto_migrate",models_count="5"} 1
db_migration_duration_seconds_bucket{operation="auto_migrate",le="1.0"} 1
```

### Cache Module Metrics

Automatically collected for all cache operations:

```prometheus
# Operation tracking
cache_operations_total{operation="get",store="redis"} 1000
cache_operation_duration_seconds_bucket{operation="get",store="redis",le="0.001"} 950

# Hit/miss tracking
cache_hits_total{operation="get",store="redis"} 850
cache_misses_total{operation="get",store="redis"} 150

# Size monitoring
cache_size_bytes{store="redis"} 1048576
```

### Event Module Metrics

Automatically collected for event publishing and consumption:

```prometheus
# Publishing metrics
event_publish_total{operation="publish",event_type="user.created"} 25
event_publish_duration_seconds_bucket{operation="publish",event_type="user.created",le="0.01"} 24

# Consumption metrics
event_consume_total{operation="consume",event_type="user.created"} 25
event_consume_duration_seconds_bucket{operation="consume",event_type="user.created",le="0.01"} 23

# Error tracking
event_errors_total{operation="publish",event_type="user.created"} 0

# Queue monitoring
event_queue_size 5
event_health_check_duration_seconds_bucket{operation="health",le="0.1"} 10
```

### MongoDB Module Metrics

Automatically collected for MongoDB operations:

```prometheus
# Operation tracking
mongodb_operations_total{operation="find",collection="users"} 75
mongodb_operation_duration_seconds_bucket{operation="find",collection="users",le="0.01"} 70

# Connection monitoring
mongodb_connections_active 8

# Error tracking
mongodb_errors_total{operation="find",collection="users"} 1
```

## Distributed Tracing

All modules include automatic OpenTelemetry tracing:

### Automatic Trace Attributes

- **HTTP**: `http.method`, `http.route`, `http.status_code`, `http.duration`
- **Database**: `db.operation`, `db.table`, `db.duration`, `db.rows_affected`
- **Cache**: `cache.key`, `cache.operation`, `cache.hit`, `cache.store`
- **Events**: `event.type`, `event.operation`
- **MongoDB**: `db.operation`, `db.collection`, `db.duration`

### Trace Propagation

Traces automatically propagate across module boundaries:

```
HTTP Request → Database Query → Cache Operation → Event Publish
     │              │               │                │
     └── Span ──→ Child Span ──→ Child Span ──→ Child Span
```

## Custom Metrics and Tracing

### Using the Metrics Manager

In your HTTP handlers, you can access the metrics manager:

```go
import (
    "go.oease.dev/goe/v2/core/http"
    "go.opentelemetry.io/otel/attribute"
)

func MyHandler(c fiber.Ctx) error {
    // Get metrics manager from context
    metrics := http.GetMetrics(c)
    if metrics != nil {
        // Create custom metrics
        customCounter := metrics.Counter("custom_operations_total",
            contract.WithDescription("Custom operation counter"),
            contract.WithUnit("operations"),
        )
        
        // Record metrics
        customCounter.Inc(c.Context(),
            attribute.String("operation", "custom"),
            attribute.String("status", "success"),
        )
    }
    
    return c.JSON(fiber.Map{"message": "success"})
}
```

### Custom Tracing

```go
func MyHandler(c fiber.Ctx) error {
    // Get tracing manager from context
    tracing := http.GetTracing(c)
    if tracing != nil {
        // Create custom spans
        ctx, span := tracing.StartSpan(c.Context(), "custom.operation",
            contract.WithSpanKind(trace.SpanKindInternal),
            contract.WithSpanAttributes(
                attribute.String("custom.param", "value"),
            ),
        )
        defer span.End()
        
        // Use the traced context for downstream operations
        // ... your business logic ...
        
        span.SetAttributes(attribute.String("result", "success"))
    }
    
    return c.JSON(fiber.Map{"message": "success"})
}
```

## Performance Considerations

### Overhead Measurements

| Module | Overhead per Operation | Impact |
|--------|----------------------|---------|
| HTTP | ~0.1-0.5ms | Minimal |
| Database | ~0.01-0.1ms | Negligible |
| Cache | ~0.01ms | Negligible |
| Events | ~0.1ms | Minimal |

### Optimization Tips

1. **Sampling**: Use `OTEL_TRACING_SAMPLING_RATIO` for high-traffic services
2. **Label Cardinality**: Avoid high-cardinality labels (user IDs, timestamps)
3. **Batch Operations**: Metrics are exported asynchronously
4. **Resource Limits**: Monitor memory usage in production

## Production Setup

### Docker Compose Example

```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8181:8181"
      - "9090:9090"  # Metrics endpoint
    environment:
      - OTEL_ENABLED=true
      - OTEL_METRICS_ENABLED=true
      - OTEL_TRACING_ENABLED=true
      - OTEL_SERVICE_NAME=my-goe-app
      - OTEL_ENVIRONMENT=production

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9091:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-storage:/var/lib/grafana

volumes:
  grafana-storage:
```

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'goe-app'
    static_configs:
      - targets: ['app:9090']
    metrics_path: '/metrics'
    scrape_interval: 5s
```

## Troubleshooting

### Common Issues

#### 1. No Metrics Endpoint

**Problem**: `curl http://localhost:9090/metrics` fails

**Solutions**:
- Ensure `WithObservability: true` in your app configuration
- Check `OTEL_ENABLED=true` and `OTEL_METRICS_ENABLED=true`
- Verify port isn't in use: `lsof -i :9090`

#### 2. No Module Metrics

**Problem**: Only basic Go metrics, no `http_`, `db_`, `cache_`, or `event_` metrics

**Solutions**:
- Generate activity: make HTTP requests, use database, cache, events
- Check logs for: `Creating metrics middleware - observability components available`
- Verify modules are enabled: `WithHTTP: true`, `WithDB: true`, etc.

#### 3. High Memory Usage

**Problem**: Memory usage increases over time

**Solutions**:
- Avoid high-cardinality labels (user IDs, timestamps)
- Reduce sampling: `OTEL_TRACING_SAMPLING_RATIO=0.1`
- Monitor metric series count

### Debug Steps

1. **Check application startup logs**:
   ```
   Creating metrics middleware - observability components available
   Prometheus server verified running status=200
   ```

2. **Test metrics endpoint**:
   ```bash
   curl -I http://localhost:9090/metrics
   ```

3. **Generate activity and check for module metrics**:
   ```bash
   # Generate HTTP traffic
   curl http://localhost:8181/
   
   # Check for module metrics
   curl -s http://localhost:9090/metrics | grep -E "^(http|db|cache|event)_"
   ```

4. **Verify configuration**:
   ```bash
   echo "OTEL_ENABLED: $OTEL_ENABLED"
   echo "OTEL_METRICS_ENABLED: $OTEL_METRICS_ENABLED"
   ```

## Monitoring Stack Integration

### Grafana Dashboards

GOE metrics work with standard Grafana dashboards. Key queries:

```promql
# Request rate
rate(http_requests_total[5m])

# Error rate
rate(http_requests_total{status=~"4..|5.."}[5m]) / rate(http_requests_total[5m]) * 100

# Response time percentiles
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Database query performance
rate(db_queries_total[5m])
histogram_quantile(0.95, rate(db_query_duration_seconds_bucket[5m]))

# Cache hit rate
rate(cache_hits_total[5m]) / rate(cache_operations_total[5m]) * 100
```

### Alerting Rules

```yaml
# prometheus-alerts.yml
groups:
  - name: goe-app
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        annotations:
          summary: "High error rate detected"

      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        annotations:
          summary: "High response time detected"

      - alert: DatabaseErrors
        expr: rate(db_errors_total[5m]) > 0.01
        for: 2m
        annotations:
          summary: "Database errors detected"
```

## Best Practices

1. **Enable observability from day one** - it's zero-configuration
2. **Use structured logging** with consistent field names
3. **Monitor business metrics** in addition to technical metrics
4. **Set up alerting** for critical application and infrastructure issues
5. **Use distributed tracing** to debug complex request flows
6. **Implement health checks** for external dependencies
7. **Monitor trends over time** to spot gradual degradation
8. **Test your monitoring** in staging environments

## Migration from v1.x

The v2.x observability system is a complete rewrite with automatic integration:

### Key Changes
- **Automatic**: Metrics are now automatically enabled with `WithObservability: true`
- **Comprehensive**: All modules have built-in metrics (HTTP, DB, Cache, Events, MongoDB)
- **Standards-Based**: Full OpenTelemetry compliance
- **Zero-Config**: Works out-of-the-box with sensible defaults

### Upgrade Steps
1. Add `WithObservability: true` to your app configuration
2. Remove any custom metrics code (now handled automatically)
3. Update environment variables (see Configuration section)
4. Update dashboards to use new metric names
5. Test the metrics endpoint after upgrade

## Next Steps

- [**Deployment**](./deployment.md) - Deploy with monitoring enabled
- [**Best Practices**](./best-practices.md) - Follow development best practices
- [**Testing**](./testing.md) - Test your application including metrics