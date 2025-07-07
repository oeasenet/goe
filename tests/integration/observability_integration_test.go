package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func TestObservabilityIntegration(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_SERVICE_NAME", "test-service")
	t.Setenv("OTEL_SERVICE_VERSION", "1.0.0")
	t.Setenv("OTEL_ENVIRONMENT", "test")
	t.Setenv("OTEL_METRICS_ENABLED", "true")
	t.Setenv("OTEL_TRACING_ENABLED", "true")
	t.Setenv("OTEL_METRICS_EXPORTERS", "prometheus")
	t.Setenv("OTEL_TRACING_EXPORTERS", "otlp")

	// Track if observability components are accessible
	var (
		observabilityReceived bool
		metricsReceived       bool
		tracingReceived       bool
	)

	// Create application with observability enabled
	app := goe.New(goe.Options{
		WithObservability: true,
		Invokers: []any{
			func(obs contract.Observability, mm contract.MetricsManager, tm contract.TracingManager) {
				observabilityReceived = true
				metricsReceived = true
				tracingReceived = true

				// Test basic functionality
				assert.NotNil(t, obs)
				assert.NotNil(t, mm)
				assert.NotNil(t, tm)

				// Test metrics
				testMetricsWithManager(t, mm)

				// Test tracing
				testTracingWithManager(t, tm)
			},
		},
	})

	// Test that the application was created successfully
	assert.NotNil(t, app)

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Start(ctx)
	assert.NoError(t, err)

	// Test that we can access observability components
	assert.True(t, app.IsRunning())
	assert.True(t, observabilityReceived)
	assert.True(t, metricsReceived)
	assert.True(t, tracingReceived)

	// Stop the application
	app.Stop(ctx)
}

func testMetricsWithManager(t *testing.T, metricsManager contract.MetricsManager) {
	// Create and use metrics
	ctx := context.Background()

	// Test counter
	counter := metricsManager.Counter("test_requests_total",
		contract.WithDescription("Total test requests"),
		contract.WithUnit("requests"),
	)
	counter.Inc(ctx, attribute.String("method", "GET"))
	counter.Add(ctx, 5, attribute.String("method", "POST"))

	// Test histogram
	histogram := metricsManager.Histogram("test_request_duration",
		contract.WithDescription("Test request duration"),
		contract.WithUnit("seconds"),
	)
	histogram.Record(ctx, 0.123, attribute.String("endpoint", "/api/test"))

	// Test gauge
	gauge := metricsManager.Gauge("test_active_connections",
		contract.WithDescription("Active test connections"),
		contract.WithUnit("connections"),
	)
	gauge.Set(ctx, 42.0, attribute.String("server", "test-server"))

	// Test up-down counter
	upDownCounter := metricsManager.UpDownCounter("test_queue_size",
		contract.WithDescription("Test queue size"),
		contract.WithUnit("items"),
	)
	upDownCounter.Add(ctx, 10, attribute.String("queue", "test-queue"))
	upDownCounter.Add(ctx, -3, attribute.String("queue", "test-queue"))

	// Verify that the same metric names return the same instances
	counter2 := metricsManager.Counter("test_requests_total")
	assert.Equal(t, counter, counter2)
}

func testTracingWithManager(t *testing.T, tracingManager contract.TracingManager) {
	// Create and use spans
	ctx := context.Background()

	// Start a root span
	rootCtx, rootSpan := tracingManager.StartSpan(ctx, "test_operation",
		contract.WithSpanKind(trace.SpanKindServer),
		contract.WithSpanAttributes(
			attribute.String("operation", "test"),
			attribute.String("version", "1.0.0"),
		),
	)
	defer rootSpan.End()

	// Set span attributes
	rootSpan.SetAttributes(
		attribute.String("user_id", "test-user"),
		attribute.Int("request_size", 1024),
	)

	// Set span status
	rootSpan.SetStatus(codes.Ok, "Operation completed successfully")

	// Add events
	rootSpan.AddEvent("Processing started")

	// Start a child span
	childCtx, childSpan := tracingManager.StartSpan(rootCtx, "test_sub_operation",
		contract.WithSpanKind(trace.SpanKindInternal),
		contract.WithSpanAttributes(
			attribute.String("sub_operation", "database_query"),
		),
	)

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	// Add event to child span
	childSpan.AddEvent("Database query executed")
	childSpan.SetStatus(codes.Ok, "Query successful")
	childSpan.End()

	// Add final event to root span
	rootSpan.AddEvent("Processing completed")

	// Test error recording
	_, errorSpan := tracingManager.StartSpan(childCtx, "test_error_operation")
	defer errorSpan.End()

	// Simulate an error
	testError := assert.AnError
	errorSpan.RecordError(testError)
	errorSpan.SetStatus(codes.Error, "Operation failed")

	// Verify that spans are properly created
	assert.NotNil(t, rootSpan.GetSpan())
	assert.NotNil(t, childSpan.GetSpan())
	assert.NotNil(t, errorSpan.GetSpan())
}

func TestObservabilityDisabled(t *testing.T) {
	// Set environment variables for disabled observability
	t.Setenv("OTEL_ENABLED", "false")

	// Track if observability components are accessible even when disabled
	var (
		observabilityReceived bool
		metricsReceived       bool
		tracingReceived       bool
	)

	// Create application with observability enabled but OTEL disabled
	app := goe.New(goe.Options{
		WithObservability: true,
		Invokers: []any{
			func(obs contract.Observability, mm contract.MetricsManager, tm contract.TracingManager) {
				observabilityReceived = true
				metricsReceived = true
				tracingReceived = true

				// Test that components are accessible even when disabled
				assert.NotNil(t, obs)
				assert.NotNil(t, mm)
				assert.NotNil(t, tm)

				// These should not panic even when disabled
				counter := mm.Counter("disabled_counter")
				counter.Inc(context.Background())

				// These should not panic even when disabled
				_, span := tm.StartSpan(context.Background(), "disabled_span")
				span.SetStatus(codes.Ok, "test")
				span.End()
			},
		},
	})

	// Test that the application was created successfully
	assert.NotNil(t, app)

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Start(ctx)
	assert.NoError(t, err)

	// Test that we can still access observability components (they should be no-ops)
	assert.True(t, app.IsRunning())
	assert.True(t, observabilityReceived)
	assert.True(t, metricsReceived)
	assert.True(t, tracingReceived)

	// Stop the application
	app.Stop(ctx)
}

func TestObservabilityConfiguration(t *testing.T) {
	// Test various configuration scenarios
	testCases := []struct {
		name    string
		envVars map[string]string
	}{
		{
			name: "Metrics only",
			envVars: map[string]string{
				"OTEL_ENABLED":         "true",
				"OTEL_METRICS_ENABLED": "true",
				"OTEL_TRACING_ENABLED": "false",
				"OTEL_SERVICE_NAME":    "metrics-only-service",
			},
		},
		{
			name: "Tracing only",
			envVars: map[string]string{
				"OTEL_ENABLED":         "true",
				"OTEL_METRICS_ENABLED": "false",
				"OTEL_TRACING_ENABLED": "true",
				"OTEL_SERVICE_NAME":    "tracing-only-service",
			},
		},
		{
			name: "Custom configuration",
			envVars: map[string]string{
				"OTEL_ENABLED":                "true",
				"OTEL_SERVICE_NAME":           "custom-service",
				"OTEL_SERVICE_VERSION":        "2.0.0",
				"OTEL_ENVIRONMENT":            "production",
				"OTEL_METRICS_PORT":           "9091",
				"OTEL_METRICS_PATH":           "/custom-metrics",
				"OTEL_TRACING_SAMPLING_RATIO": "0.5",
				"OTEL_TRACING_ENDPOINT":       "http://custom-endpoint:4318",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tc.envVars {
				t.Setenv(key, value)
			}

			// Track if observability components are accessible
			var observabilityReceived bool

			// Create application
			app := goe.New(goe.Options{
				WithObservability: true,
				Invokers: []any{
					func(obs contract.Observability) {
						observabilityReceived = true
						assert.NotNil(t, obs)
						assert.NotNil(t, obs.Metrics())
						assert.NotNil(t, obs.Tracing())
					},
				},
			})
			assert.NotNil(t, app)

			// Start and stop the application quickly
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := app.Start(ctx)
			assert.NoError(t, err)

			// Test that observability components are accessible
			assert.True(t, observabilityReceived)

			app.Stop(ctx)
		})
	}
}

func TestObservabilityWithOtherModules(t *testing.T) {
	// Test observability integration with other modules
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_SERVICE_NAME", "multi-module-service")

	// Track if all modules are accessible
	var (
		observabilityReceived bool
		httpReceived          bool
		cacheReceived         bool
	)

	// Create application with multiple modules
	app := goe.New(goe.Options{
		WithObservability: true,
		WithHTTP:          true,
		WithCache:         true,
		Invokers: []any{
			func(
				obs contract.Observability,
				httpKernel contract.HTTPKernel,
				cacheManager contract.CacheManager,
			) {
				observabilityReceived = true
				httpReceived = true
				cacheReceived = true

				assert.NotNil(t, obs)
				assert.NotNil(t, httpKernel)
				assert.NotNil(t, cacheManager)

				// Test that observability can be used to instrument other modules
				metricsManager := obs.Metrics()
				tracingManager := obs.Tracing()

				// Create metrics for HTTP requests
				httpCounter := metricsManager.Counter("http_requests_total",
					contract.WithDescription("Total HTTP requests"),
				)
				httpCounter.Inc(context.Background(), attribute.String("method", "GET"))

				// Create metrics for cache operations
				cacheCounter := metricsManager.Counter("cache_operations_total",
					contract.WithDescription("Total cache operations"),
				)
				cacheCounter.Inc(context.Background(), attribute.String("operation", "get"))

				// Create traces for operations
				_, span := tracingManager.StartSpan(context.Background(), "multi_module_operation")
				span.SetAttributes(attribute.String("modules", "http,cache,observability"))
				span.End()
			},
		},
	})

	assert.NotNil(t, app)

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := app.Start(ctx)
	assert.NoError(t, err)

	// Test that all modules are accessible
	assert.True(t, observabilityReceived)
	assert.True(t, httpReceived)
	assert.True(t, cacheReceived)

	app.Stop(ctx)
}
