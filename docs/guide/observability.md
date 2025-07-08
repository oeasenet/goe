# Observability

GOE provides built-in observability features including metrics, tracing, and logging. This guide covers setting up comprehensive monitoring for your application.

## Overview

GOE's observability features include:
- **Metrics**: Prometheus metrics for monitoring application performance
- **Tracing**: OpenTelemetry distributed tracing
- **Logging**: Structured logging with Zap
- **Health Checks**: Application health monitoring

## Enabling Observability

Enable the observability module in your GOE application:

```go
package main

import "go.oease.dev/goe/v2"

func main() {
    goe.New(goe.Options{
        WithHTTP:          true,
        WithObservability: true,
    })
    
    goe.Run()
}
```

## Configuration

Configure observability through environment variables:

```bash
# .env
# Observability
OTEL_ENABLED=true
OTEL_SERVICE_NAME=my-goe-app
OTEL_SERVICE_VERSION=1.0.0
OTEL_ENVIRONMENT=production

# Metrics
METRICS_ENABLED=true
METRICS_PORT=9090
METRICS_PATH=/metrics

# Tracing
TRACING_ENABLED=true
TRACING_ENDPOINT=http://jaeger:14268/api/traces
TRACING_SAMPLE_RATE=0.1

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

## Metrics

### Built-in Metrics

GOE automatically collects several metrics:

- **HTTP Request Metrics**:
  - `http_requests_total` - Total number of HTTP requests
  - `http_request_duration_seconds` - Request duration histogram
  - `http_requests_in_flight` - Current number of requests being processed

- **Database Metrics** (when DB module is enabled):
  - `db_connections_open` - Number of open database connections
  - `db_connections_idle` - Number of idle database connections
  - `db_query_duration_seconds` - Database query duration

- **Cache Metrics** (when cache module is enabled):
  - `cache_operations_total` - Total cache operations
  - `cache_hits_total` - Total cache hits
  - `cache_misses_total` - Total cache misses

### Custom Metrics

Create custom metrics for your application:

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    UserRegistrations = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "user_registrations_total",
            Help: "Total number of user registrations",
        },
        []string{"method", "success"},
    )
    
    OrderValue = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "order_value_dollars",
            Help:    "Value of orders in dollars",
            Buckets: []float64{10, 50, 100, 500, 1000, 5000},
        },
        []string{"currency", "payment_method"},
    )
    
    ActiveSessions = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_sessions",
            Help: "Number of active user sessions",
        },
    )
)

// Record metrics in your application
func RecordUserRegistration(method string, success bool) {
    UserRegistrations.WithLabelValues(method, fmt.Sprintf("%t", success)).Inc()
}

func RecordOrderValue(value float64, currency, paymentMethod string) {
    OrderValue.WithLabelValues(currency, paymentMethod).Observe(value)
}

func UpdateActiveSessions(count int) {
    ActiveSessions.Set(float64(count))
}
```

### Using Metrics in Services

```go
package service

import (
    "go.oease.dev/goe/v2/contract"
    "your-app/internal/metrics"
)

type UserService struct {
    repository UserRepository
    logger     contract.Logger
}

func NewUserService(repository UserRepository, logger contract.Logger) *UserService {
    return &UserService{
        repository: repository,
        logger:     logger,
    }
}

func (s *UserService) RegisterUser(email, password string) (*User, error) {
    s.logger.Info("Starting user registration", "email", email)
    
    user := &User{
        Email:    email,
        Password: hashPassword(password),
    }
    
    err := s.repository.Create(user)
    if err != nil {
        s.logger.Error("User registration failed", "email", email, "error", err)
        metrics.RecordUserRegistration("email", false)
        return nil, err
    }
    
    s.logger.Info("User registered successfully", "user_id", user.ID, "email", email)
    metrics.RecordUserRegistration("email", true)
    
    return user, nil
}
```

## Tracing

### Distributed Tracing

GOE integrates with OpenTelemetry for distributed tracing:

```go
package service

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
    "go.oease.dev/goe/v2/contract"
)

type UserService struct {
    repository UserRepository
    logger     contract.Logger
    tracer     trace.Tracer
}

func NewUserService(repository UserRepository, logger contract.Logger) *UserService {
    return &UserService{
        repository: repository,
        logger:     logger,
        tracer:     otel.Tracer("user-service"),
    }
}

func (s *UserService) GetUser(ctx context.Context, id int) (*User, error) {
    ctx, span := s.tracer.Start(ctx, "UserService.GetUser")
    defer span.End()
    
    span.SetAttributes(
        attribute.Int("user.id", id),
        attribute.String("service", "user-service"),
    )
    
    user, err := s.repository.FindByID(ctx, id)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(trace.StatusError, err.Error())
        return nil, err
    }
    
    span.SetAttributes(
        attribute.String("user.email", user.Email),
        attribute.Bool("user.active", user.Active),
    )
    
    return user, nil
}
```

### HTTP Tracing

HTTP requests are automatically traced when observability is enabled:

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/propagation"
)

func RegisterRoutes(httpKernel contract.HTTPKernel, userService *UserService) {
    app := httpKernel.App()
    
    // OpenTelemetry middleware is automatically added
    app.Get("/users/:id", func(c fiber.Ctx) error {
        // Extract tracing context from headers
        ctx := otel.GetTextMapPropagator().Extract(c.Context(), propagation.HeaderCarrier(c.GetReqHeaders()))
        
        id, err := c.ParamsInt("id")
        if err != nil {
            return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
        }
        
        user, err := userService.GetUser(ctx, id)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
        }
        
        return c.JSON(user)
    })
}
```

## Logging

### Structured Logging

Use structured logging for better observability:

```go
package service

import (
    "go.oease.dev/goe/v2/contract"
)

type OrderService struct {
    repository OrderRepository
    logger     contract.Logger
}

func (s *OrderService) CreateOrder(userID int, items []OrderItem) (*Order, error) {
    s.logger.Info("Creating order",
        "user_id", userID,
        "item_count", len(items),
        "operation", "create_order",
    )
    
    order := &Order{
        UserID: userID,
        Items:  items,
        Status: "pending",
    }
    
    total := s.calculateTotal(items)
    order.Total = total
    
    s.logger.Info("Order total calculated",
        "order_id", order.ID,
        "user_id", userID,
        "total", total,
        "currency", "USD",
    )
    
    if err := s.repository.Create(order); err != nil {
        s.logger.Error("Failed to create order",
            "user_id", userID,
            "error", err,
            "operation", "create_order",
        )
        return nil, err
    }
    
    s.logger.Info("Order created successfully",
        "order_id", order.ID,
        "user_id", userID,
        "total", total,
        "status", order.Status,
    )
    
    return order, nil
}
```

### Log Correlation

Correlate logs with traces using request IDs:

```go
import (
    "go.opentelemetry.io/otel/trace"
)

func (s *UserService) ProcessUser(ctx context.Context, id int) error {
    span := trace.SpanFromContext(ctx)
    traceID := span.SpanContext().TraceID().String()
    
    s.logger.Info("Processing user",
        "user_id", id,
        "trace_id", traceID,
        "operation", "process_user",
    )
    
    // Processing logic...
    
    return nil
}
```

## Health Checks

### Application Health

```go
package handler

import (
    "context"
    "time"
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2/contract"
)

type HealthHandler struct {
    db       contract.DB
    cache    contract.Cache
    logger   contract.Logger
}

func NewHealthHandler(db contract.DB, cache contract.Cache, logger contract.Logger) *HealthHandler {
    return &HealthHandler{
        db:     db,
        cache:  cache,
        logger: logger,
    }
}

func (h *HealthHandler) LivenessCheck(c fiber.Ctx) error {
    // Simple liveness check
    return c.JSON(fiber.Map{
        "status": "alive",
        "timestamp": time.Now().Unix(),
    })
}

func (h *HealthHandler) ReadinessCheck(c fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    checks := map[string]interface{}{
        "database": h.checkDatabase(ctx),
        "cache":    h.checkCache(ctx),
    }
    
    allHealthy := true
    for _, check := range checks {
        if status, ok := check.(map[string]interface{})["status"]; ok && status != "healthy" {
            allHealthy = false
            break
        }
    }
    
    response := fiber.Map{
        "status": "ready",
        "timestamp": time.Now().Unix(),
        "checks": checks,
    }
    
    if !allHealthy {
        response["status"] = "not_ready"
        return c.Status(503).JSON(response)
    }
    
    return c.JSON(response)
}

func (h *HealthHandler) checkDatabase(ctx context.Context) interface{} {
    if err := h.db.Instance().WithContext(ctx).Exec("SELECT 1").Error; err != nil {
        return map[string]interface{}{
            "status": "unhealthy",
            "error":  err.Error(),
        }
    }
    
    return map[string]interface{}{
        "status": "healthy",
    }
}

func (h *HealthHandler) checkCache(ctx context.Context) interface{} {
    if err := h.cache.Set("health_check", "ok", 10*time.Second); err != nil {
        return map[string]interface{}{
            "status": "unhealthy",
            "error":  err.Error(),
        }
    }
    
    return map[string]interface{}{
        "status": "healthy",
    }
}
```

## Monitoring Stack

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'goe-app'
    static_configs:
      - targets: ['app:9090']
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "GOE Application Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "type": "stat",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "Requests/sec"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "stat",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"4..|5..\"}[5m]) / rate(http_requests_total[5m]) * 100",
            "legendFormat": "Error %"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          }
        ]
      },
      {
        "title": "Database Connections",
        "type": "graph",
        "targets": [
          {
            "expr": "db_connections_open",
            "legendFormat": "Open"
          },
          {
            "expr": "db_connections_idle",
            "legendFormat": "Idle"
          }
        ]
      }
    ]
  }
}
```

## Alerting

### Prometheus Alerting Rules

```yaml
# alerts.yml
groups:
  - name: goe-app-alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} requests/sec"

      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High response time detected"
          description: "95th percentile response time is {{ $value }} seconds"

      - alert: DatabaseDown
        expr: up{job="goe-app"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Database is down"
          description: "Database has been down for more than 1 minute"

      - alert: HighMemoryUsage
        expr: (process_resident_memory_bytes / process_virtual_memory_bytes) * 100 > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Memory usage is {{ $value }}%"
```

### Alertmanager Configuration

```yaml
# alertmanager.yml
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'alerts@myapp.com'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
  - name: 'web.hook'
    email_configs:
      - to: 'admin@myapp.com'
        subject: 'GOE App Alert: {{ .GroupLabels.alertname }}'
        body: |
          {{ range .Alerts }}
          Alert: {{ .Annotations.summary }}
          Description: {{ .Annotations.description }}
          {{ end }}
    
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/...'
        channel: '#alerts'
        title: 'GOE App Alert'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
```

## Performance Monitoring

### APM Integration

```go
package main

import (
    "go.opentelemetry.io/contrib/instrumentation/github.com/gofiber/fiber/otelfiber"
    "go.opentelemetry.io/contrib/instrumentation/gorm.io/gorm/otelgorm"
)

func setupAPM(app *fiber.App, db *gorm.DB) {
    // Add OpenTelemetry middleware to Fiber
    app.Use(otelfiber.Middleware())
    
    // Add OpenTelemetry plugin to GORM
    if err := db.Use(otelgorm.NewPlugin()); err != nil {
        log.Fatal("Failed to setup GORM OpenTelemetry plugin", err)
    }
}
```

### Custom Instrumentation

```go
package service

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/metric"
)

type PaymentService struct {
    meter   metric.Meter
    counter metric.Int64Counter
}

func NewPaymentService() *PaymentService {
    meter := otel.Meter("payment-service")
    counter, _ := meter.Int64Counter("payments_processed")
    
    return &PaymentService{
        meter:   meter,
        counter: counter,
    }
}

func (s *PaymentService) ProcessPayment(ctx context.Context, amount float64, currency string) error {
    // Record payment
    s.counter.Add(ctx, 1, metric.WithAttributes(
        attribute.String("currency", currency),
        attribute.Float64("amount", amount),
    ))
    
    // Process payment logic...
    
    return nil
}
```

## Best Practices

1. **Use structured logging** with consistent field names
2. **Implement proper error handling** with appropriate log levels
3. **Monitor key business metrics** not just technical metrics
4. **Set up alerting** for critical issues
5. **Use distributed tracing** for complex operations
6. **Implement health checks** for all external dependencies
7. **Monitor performance trends** over time
8. **Set up automated incident response** where possible

## Next Steps

- [**Deployment**](./deployment.md) - Deploy with monitoring enabled
- [**Best Practices**](./best-practices.md) - Follow monitoring best practices
- [**Testing**](./testing.md) - Test your monitoring setup