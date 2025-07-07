package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	"go.opentelemetry.io/otel/attribute"
)

func TestHTTPMetricsIntegration(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_SERVICE_NAME", "http-metrics-test")
	t.Setenv("OTEL_METRICS_ENABLED", "true")
	t.Setenv("OTEL_METRICS_EXPORTERS", "prometheus")
	t.Setenv("HTTP_PORT", "8081") // Use different port to avoid conflicts

	var metricsManager contract.MetricsManager

	// Create application with HTTP and observability enabled
	app := goe.New(goe.Options{
		WithHTTP:          true,
		WithObservability: true,
		Invokers: []any{
			func(mm contract.MetricsManager, hk contract.HTTPKernel) {
				metricsManager = mm

				// Set up a test route
				hk.App().Get("/test-metrics", func(c contract.Context) error {
					return c.JSON(map[string]string{"message": "metrics test"})
				})
			},
		},
	})

	assert.NotNil(t, app)

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Start(ctx)
	assert.NoError(t, err)

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test that HTTP metrics are available
	assert.NotNil(t, metricsManager)

	// Create some HTTP metrics to verify they work
	httpCounter := metricsManager.Counter("test_http_requests_total",
		contract.WithDescription("Test HTTP requests"),
		contract.WithUnit("requests"),
	)

	httpDuration := metricsManager.Histogram("test_http_request_duration",
		contract.WithDescription("Test HTTP request duration"),
		contract.WithUnit("seconds"),
	)

	// Record some test metrics
	testCtx := context.Background()
	httpCounter.Inc(testCtx, attribute.String("method", "GET"))
	httpDuration.Record(testCtx, 0.123, attribute.String("endpoint", "/test-metrics"))

	// Verify metrics instances are the same (cached)
	httpCounter2 := metricsManager.Counter("test_http_requests_total")
	assert.Equal(t, httpCounter, httpCounter2)

	// Stop the application
	app.Stop(ctx)
}

func TestCacheMetricsIntegration(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_SERVICE_NAME", "cache-metrics-test")
	t.Setenv("OTEL_METRICS_ENABLED", "true")
	t.Setenv("CACHE_DRIVER", "memory")

	var cache contract.Cache
	var metricsManager contract.MetricsManager

	// Create application with Cache and observability enabled
	app := goe.New(goe.Options{
		WithCache:         true,
		WithObservability: true,
		Invokers: []any{
			func(c contract.Cache, mm contract.MetricsManager) {
				cache = c
				metricsManager = mm

				// Test cache operations to trigger metrics
				testCtx := context.Background()

				// Test cache operations
				err := cache.Set("test_key", "test_value", time.Minute)
				assert.NoError(t, err)

				value, err := cache.Get("test_key")
				assert.NoError(t, err)
				assert.Equal(t, "test_value", value)

				exists := cache.Has("test_key")
				assert.True(t, exists)

				err = cache.Forget("test_key")
				assert.NoError(t, err)

				// Verify we can create cache-related metrics
				cacheCounter := mm.Counter("test_cache_operations_total",
					contract.WithDescription("Test cache operations"),
					contract.WithUnit("operations"),
				)
				cacheCounter.Inc(testCtx, attribute.String("operation", "test"))
			},
		},
	})

	assert.NotNil(t, app)

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Start(ctx)
	assert.NoError(t, err)

	// Wait for operations to complete
	time.Sleep(100 * time.Millisecond)

	// Verify cache and metrics are available
	assert.NotNil(t, cache)
	assert.NotNil(t, metricsManager)

	// Stop the application
	app.Stop(ctx)
}

func TestDBMetricsIntegration(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_SERVICE_NAME", "db-metrics-test")
	t.Setenv("OTEL_METRICS_ENABLED", "true")
	t.Setenv("DB_CONNECTION", "default")
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_DATABASE", ":memory:")

	var db contract.DB
	var metricsManager contract.MetricsManager

	// Create application with DB and observability enabled
	app := goe.New(goe.Options{
		WithDB:            true,
		WithObservability: true,
		Invokers: []any{
			func(d contract.DB, mm contract.MetricsManager) {
				// Just store the references, don't use them yet
				db = d
				metricsManager = mm
			},
		},
	})

	assert.NotNil(t, app)

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Start(ctx)
	assert.NoError(t, err)

	// Wait for operations to complete
	time.Sleep(100 * time.Millisecond)

	// Verify DB and metrics are available
	assert.NotNil(t, db)
	assert.NotNil(t, metricsManager)

	// Now test DB operations after the application is started
	instance := db.Instance()
	assert.NotNil(t, instance)

	// Create a simple test table
	err = instance.Exec("CREATE TABLE test_metrics (id INTEGER PRIMARY KEY, name TEXT)").Error
	assert.NoError(t, err)

	// Insert test data
	err = instance.Exec("INSERT INTO test_metrics (name) VALUES (?)", "test").Error
	assert.NoError(t, err)

	// Query test data
	var count int64
	err = instance.Model(&struct{}{}).Table("test_metrics").Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Verify we can create DB-related metrics
	dbCounter := metricsManager.Counter("test_db_queries_total",
		contract.WithDescription("Test database queries"),
		contract.WithUnit("queries"),
	)
	dbCounter.Inc(context.Background(), attribute.String("operation", "test"))

	// Stop the application
	app.Stop(ctx)
}

func TestObservabilityMetricsIntegration(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_SERVICE_NAME", "observability-metrics-test")
	t.Setenv("OTEL_METRICS_ENABLED", "true")
	t.Setenv("OTEL_METRICS_PORT", "9091") // Use different port

	var observability contract.Observability

	// Create application with observability enabled
	app := goe.New(goe.Options{
		WithObservability: true,
		Invokers: []any{
			func(obs contract.Observability) {
				observability = obs

				// Test observability self-monitoring
				metricsManager := obs.Metrics()
				assert.NotNil(t, metricsManager)

				// Create various types of metrics
				ctx := context.Background()

				// Counter
				counter := metricsManager.Counter("observability_operations_total",
					contract.WithDescription("Total observability operations"),
					contract.WithUnit("operations"),
				)
				counter.Inc(ctx, attribute.String("component", "metrics"))
				counter.Add(ctx, 5, attribute.String("component", "tracing"))

				// Histogram
				histogram := metricsManager.Histogram("observability_operation_duration",
					contract.WithDescription("Observability operation duration"),
					contract.WithUnit("seconds"),
				)
				histogram.Record(ctx, 0.001, attribute.String("operation", "metric_creation"))

				// Gauge
				gauge := metricsManager.Gauge("observability_active_components",
					contract.WithDescription("Number of active observability components"),
					contract.WithUnit("components"),
				)
				gauge.Set(ctx, 3.0, attribute.String("type", "metrics"))

				// UpDownCounter
				upDownCounter := metricsManager.UpDownCounter("observability_resource_usage",
					contract.WithDescription("Observability resource usage"),
					contract.WithUnit("bytes"),
				)
				upDownCounter.Add(ctx, 1024)
				upDownCounter.Add(ctx, -512)

				// Test that meter is accessible
				meter := metricsManager.GetMeter()
				assert.NotNil(t, meter)
			},
		},
	})

	assert.NotNil(t, app)

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Start(ctx)
	assert.NoError(t, err)

	// Wait for operations to complete
	time.Sleep(100 * time.Millisecond)

	// Verify observability is available
	assert.NotNil(t, observability)

	// Stop the application
	app.Stop(ctx)
}

func TestMultiModuleMetricsIntegration(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_SERVICE_NAME", "multi-module-metrics-test")
	t.Setenv("OTEL_METRICS_ENABLED", "true")
	t.Setenv("HTTP_PORT", "8082")
	t.Setenv("CACHE_DRIVER", "memory")
	t.Setenv("DB_CONNECTION", "default")
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_DATABASE", ":memory:")

	var allComponentsReceived bool
	var cache contract.Cache
	var db contract.DB
	var obs contract.Observability

	// Create application with all modules enabled
	app := goe.New(goe.Options{
		WithHTTP:          true,
		WithCache:         true,
		WithDB:            true,
		WithObservability: true,
		Invokers: []any{
			func(
				httpKernel contract.HTTPKernel,
				c contract.Cache,
				d contract.DB,
				o contract.Observability,
			) {
				allComponentsReceived = true

				// Store references for later use
				cache = c
				db = d
				obs = o

				// Verify all components are available
				assert.NotNil(t, httpKernel)
				assert.NotNil(t, cache)
				assert.NotNil(t, db)
				assert.NotNil(t, obs)

				metricsManager := obs.Metrics()
				assert.NotNil(t, metricsManager)

				// Test cross-module metrics
				ctx := context.Background()

				// Create metrics for each module
				httpMetric := metricsManager.Counter("multi_module_http_requests",
					contract.WithDescription("HTTP requests in multi-module test"),
				)
				httpMetric.Inc(ctx, attribute.String("module", "http"))

				cacheMetric := metricsManager.Counter("multi_module_cache_operations",
					contract.WithDescription("Cache operations in multi-module test"),
				)
				cacheMetric.Inc(ctx, attribute.String("module", "cache"))

				dbMetric := metricsManager.Counter("multi_module_db_queries",
					contract.WithDescription("DB queries in multi-module test"),
				)
				dbMetric.Inc(ctx, attribute.String("module", "db"))

				// Test cache operations (safe to do during DI)
				err := cache.Set("multi_test", "value", time.Minute)
				assert.NoError(t, err)

				value, err := cache.Get("multi_test")
				assert.NoError(t, err)
				assert.Equal(t, "value", value)
			},
		},
	})

	assert.NotNil(t, app)

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Start(ctx)
	assert.NoError(t, err)

	// Wait for operations to complete
	time.Sleep(100 * time.Millisecond)

	// Verify all components were received
	assert.True(t, allComponentsReceived)

	// Now test DB operations after the application has started
	instance := db.Instance()
	assert.NotNil(t, instance)

	// Stop the application
	app.Stop(ctx)
}
