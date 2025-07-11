# Observability Troubleshooting

This guide provides step-by-step troubleshooting for common observability issues in GOE applications.

## Quick Diagnostics

### 1. Verify Observability is Enabled

```go
// ✅ Correct configuration
app := goe.New(goe.Options{
    WithHTTP:          true,
    WithDB:            true,
    WithObservability: true, // Required!
})

// ❌ Common mistake - missing observability
app := goe.New(goe.Options{
    WithHTTP: true,
    WithDB:   true,
    // Missing WithObservability: true
})
```

### 2. Check Environment Variables

```bash
# Required variables
export OTEL_ENABLED=true
export OTEL_METRICS_ENABLED=true
export OTEL_METRICS_PORT=9090

# Verify they're set
echo "OTEL_ENABLED: $OTEL_ENABLED"
echo "OTEL_METRICS_ENABLED: $OTEL_METRICS_ENABLED"
echo "OTEL_METRICS_PORT: $OTEL_METRICS_PORT"
```

### 3. Test Metrics Endpoint

```bash
# Test if Prometheus server is running
curl -I http://localhost:9090/metrics

# Should return:
# HTTP/1.1 200 OK
# Content-Type: text/plain; version=0.0.4; charset=utf-8
```

## Common Problems & Solutions

### Problem 1: "Connection Refused" on Metrics Endpoint

**Symptoms:**
```bash
curl: (7) Failed to connect to localhost port 9090: Connection refused
```

**Diagnosis Steps:**

1. **Check if port is in use:**
   ```bash
   lsof -i :9090
   netstat -an | grep 9090
   ```

2. **Check application logs for startup errors:**
   ```bash
   tail -f app.log | grep -E "(prometheus|metrics|observability)"
   ```

3. **Look for this log message:**
   ```
   Prometheus server verified running status=200 url=http://localhost:9090/metrics
   ```

**Solutions:**

- **Port conflict**: Change port
  ```bash
  export OTEL_METRICS_PORT=9091
  ```

- **Observability not enabled**: Add to configuration
  ```go
  goe.New(goe.Options{
      WithObservability: true, // Add this
  })
  ```

- **Binding issue**: Server binds to 0.0.0.0, check firewall rules

### Problem 2: No Module-Specific Metrics

**Symptoms:**
- Basic Go metrics appear (`go_goroutines`, `process_cpu_seconds_total`)
- No `http_`, `db_`, `cache_`, or `event_` metrics

**Diagnosis Steps:**

1. **Check for metrics middleware logs:**
   ```bash
   grep "Creating metrics middleware" app.log
   ```

2. **Generate activity:**
   ```bash
   # HTTP metrics
   curl http://localhost:8181/
   
   # Check for HTTP metrics
   curl -s http://localhost:9090/metrics | grep "http_requests_total"
   ```

3. **Verify module usage:**
   - HTTP: Make requests to your endpoints
   - Database: Perform database operations
   - Cache: Use cache operations
   - Events: Publish/consume events

**Solutions:**

- **Modules not enabled**: Enable required modules
  ```go
  goe.New(goe.Options{
      WithHTTP:  true, // For HTTP metrics
      WithDB:    true, // For database metrics
      WithCache: true, // For cache metrics
      WithEvent: true, // For event metrics
      WithObservability: true,
  })
  ```

- **No activity generated**: Use the application features that generate metrics

- **Timing issue**: Check logs for middleware creation messages

### Problem 3: High Memory Usage

**Symptoms:**
- Application memory usage increases over time
- Out of memory errors in production

**Diagnosis Steps:**

1. **Check metric cardinality:**
   ```bash
   curl -s http://localhost:9090/metrics | grep -c "^[a-z]"
   ```

2. **Monitor memory growth:**
   ```bash
   # Check current memory usage
   ps aux | grep your-app
   
   # Monitor over time
   watch "ps aux | grep your-app"
   ```

**Solutions:**

- **Reduce high-cardinality labels:**
  ```go
  // ❌ High cardinality - creates too many series
  counter.Inc(ctx, 
      attribute.String("user_id", userID), // Thousands of unique values
      attribute.String("timestamp", time.Now().String()),
  )
  
  // ✅ Low cardinality - limited unique values
  counter.Inc(ctx,
      attribute.String("method", "GET"),
      attribute.String("status_class", "2xx"),
  )
  ```

- **Reduce trace sampling:**
  ```bash
  export OTEL_TRACING_SAMPLING_RATIO=0.1  # Sample 10% instead of 100%
  ```

### Problem 4: Database Metrics Missing

**Symptoms:**
- HTTP and other metrics work
- No `db_queries_total` or related metrics

**Diagnosis Steps:**

1. **Check GORM plugin registration:**
   ```bash
   grep -E "(gorm|database)" app.log
   ```

2. **Verify database operations:**
   ```go
   // Make sure you're using the injected DB instance
   func MyHandler(c fiber.Ctx) error {
       services := http.GetServices(c)
       db := services.DB // This is the metrics-wrapped instance
       
       var user User
       db.Instance().Where("id = ?", 1).First(&user) // This will be tracked
       return c.JSON(user)
   }
   ```

**Solutions:**

- **Use wrapped DB instance**: Always use the dependency-injected database
- **Check for GORM errors**: Look for plugin registration errors in logs
- **Verify database operations**: Ensure you're actually making database calls

### Problem 5: Event Metrics Missing

**Symptoms:**
- Other metrics work but no event metrics
- Events are being published/consumed but not tracked

**Diagnosis Steps:**

1. **Verify event module is enabled:**
   ```go
   goe.New(goe.Options{
       WithEvent: true, // Required
       WithObservability: true,
   })
   ```

2. **Check event operations:**
   ```go
   // Publish events to generate metrics
   eventManager.Publish(ctx, &MyEvent{
       Type: "user.created",
       Data: userData,
   })
   ```

**Solutions:**

- **Enable event module**: Add `WithEvent: true` to configuration
- **Use event manager**: Publish and consume events through the framework
- **Check event handler registration**: Ensure handlers are properly registered

## Debug Tools and Commands

### Complete Diagnostic Script

```bash
#!/bin/bash
echo "=== GOE Observability Diagnostics ==="

echo "1. Environment Variables:"
env | grep -E "^OTEL_" | sort

echo -e "\n2. Metrics Endpoint Test:"
if curl -s -I http://localhost:9090/metrics | grep -q "200 OK"; then
    echo "✅ Metrics endpoint accessible"
else
    echo "❌ Metrics endpoint not accessible"
fi

echo -e "\n3. Port Usage:"
lsof -i :9090 2>/dev/null || echo "Port 9090 not in use"
lsof -i :8181 2>/dev/null || echo "Port 8181 not in use"

echo -e "\n4. Available Metrics Sample:"
curl -s http://localhost:9090/metrics | grep -E "^(http|db|cache|event)_" | head -10

echo -e "\n5. Basic Metrics Count:"
echo "Total metrics: $(curl -s http://localhost:9090/metrics | grep -c "^[a-z]")"
echo "HTTP metrics: $(curl -s http://localhost:9090/metrics | grep -c "^http_")"
echo "DB metrics: $(curl -s http://localhost:9090/metrics | grep -c "^db_")"
echo "Cache metrics: $(curl -s http://localhost:9090/metrics | grep -c "^cache_")"
echo "Event metrics: $(curl -s http://localhost:9090/metrics | grep -c "^event_")"
```

### Activity Generator Script

```bash
#!/bin/bash
echo "Generating activity to create metrics..."

echo "Making HTTP requests..."
for i in {1..5}; do
    curl -s http://localhost:8181/ > /dev/null
    echo "Request $i completed"
    sleep 1
done

echo -e "\nWaiting for metrics to be recorded..."
sleep 2

echo -e "\nChecking for HTTP metrics:"
curl -s http://localhost:9090/metrics | grep "http_requests_total"
```

### Memory Monitoring

```bash
#!/bin/bash
echo "Monitoring memory usage..."

while true; do
    MEMORY=$(ps aux | grep your-app | grep -v grep | awk '{print $6}')
    METRICS_COUNT=$(curl -s http://localhost:9090/metrics 2>/dev/null | grep -c "^[a-z]" || echo "N/A")
    echo "$(date): Memory: ${MEMORY}KB, Metrics: ${METRICS_COUNT}"
    sleep 10
done
```

## Best Practices for Troubleshooting

1. **Enable debug logging during development:**
   ```bash
   export LOG_LEVEL=debug
   ```

2. **Use structured logging:**
   ```go
   logger.Info("Processing request",
       "user_id", userID,
       "operation", "get_profile",
       "trace_id", span.SpanContext().TraceID().String(),
   )
   ```

3. **Monitor key indicators:**
   - Application startup logs
   - Metrics endpoint accessibility
   - Memory usage trends
   - Metric cardinality

4. **Test observability in staging:**
   - Verify metrics collection under load
   - Test alerting rules
   - Validate dashboard queries

5. **Keep metrics simple:**
   - Use low-cardinality labels
   - Follow Prometheus naming conventions
   - Document custom metrics

## Getting Help

When reporting observability issues, include:

1. **GOE version**: `go list -m go.oease.dev/goe/v2`
2. **Configuration**: Sanitized environment variables
3. **Logs**: Application startup and error logs
4. **Diagnostics**: Output from diagnostic scripts above
5. **Expected behavior**: What metrics you expect to see

## Recovery Procedures

### Reset Observability Configuration

```bash
# 1. Stop application
# 2. Clear problematic environment variables
unset OTEL_ENABLED OTEL_METRICS_ENABLED OTEL_TRACING_ENABLED

# 3. Set minimal configuration
export OTEL_ENABLED=true
export OTEL_METRICS_ENABLED=true
export OTEL_METRICS_PORT=9090

# 4. Restart application
# 5. Test basic functionality
curl -I http://localhost:9090/metrics
```

### Emergency Disable

If observability is causing issues in production:

```go
// Temporarily disable observability
app := goe.New(goe.Options{
    WithHTTP:          true,
    WithDB:            true,
    WithObservability: false, // Disable temporarily
})
```

Or via environment:

```bash
export OTEL_ENABLED=false
```

This will gracefully disable all metrics and tracing while keeping the application functional.