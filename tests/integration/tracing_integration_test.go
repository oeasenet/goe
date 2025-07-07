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
	"gorm.io/gorm"
)

func TestHTTPTracingIntegration(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_SERVICE_NAME", "http-tracing-test")
	t.Setenv("OTEL_TRACING_ENABLED", "true")
	t.Setenv("OTEL_TRACING_EXPORTERS", "otlp")
	t.Setenv("HTTP_PORT", "8083") // Use different port to avoid conflicts

	var tracingManager contract.TracingManager

	// Create application with HTTP and observability enabled
	app := goe.New(goe.Options{
		WithHTTP:          true,
		WithObservability: true,
		Invokers: []any{
			func(tm contract.TracingManager, hk contract.HTTPKernel) {
				tracingManager = tm

				// Set up a test route with tracing
				hk.App().Get("/test-tracing", func(c contract.Context) error {
					ctx := c.Context()

					// Start a span for this request
					ctx, span := tm.StartSpan(ctx, "test_http_request",
						contract.WithSpanKind(trace.SpanKindServer),
						contract.WithSpanAttributes(
							attribute.String("http.method", "GET"),
							attribute.String("http.route", "/test-tracing"),
						),
					)
					defer span.End()

					// Set some attributes
					span.SetAttributes(
						attribute.String("user.id", "test-user"),
						attribute.String("request.type", "test"),
					)

					// Add an event
					span.AddEvent("Processing request")

					// Simulate some work
					time.Sleep(10 * time.Millisecond)

					// Set success status
					span.SetStatus(codes.Ok, "Request processed successfully")

					return c.JSON(map[string]string{"message": "tracing test"})
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

	// Test that HTTP tracing is available
	assert.NotNil(t, tracingManager)

	// Test creating spans manually
	testCtx := context.Background()
	spanCtx, span := tracingManager.StartSpan(testCtx, "test_manual_span",
		contract.WithSpanKind(trace.SpanKindInternal),
		contract.WithSpanAttributes(
			attribute.String("component", "test"),
		),
	)
	defer span.End()

	// Test span operations
	span.SetAttributes(attribute.String("test.operation", "manual"))
	span.AddEvent("Manual span created")
	span.SetStatus(codes.Ok, "Manual span completed")

	// Test nested spans
	_, childSpan := tracingManager.StartSpan(spanCtx, "test_child_span",
		contract.WithSpanKind(trace.SpanKindInternal),
	)
	childSpan.SetAttributes(attribute.String("span.type", "child"))
	childSpan.End()

	// Verify tracer is accessible
	tracer := tracingManager.GetTracer()
	assert.NotNil(t, tracer)

	// Stop the application
	app.Stop(ctx)
}

func TestCacheTracingIntegration(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_SERVICE_NAME", "cache-tracing-test")
	t.Setenv("OTEL_TRACING_ENABLED", "true")
	t.Setenv("CACHE_DRIVER", "memory")

	var cache contract.Cache
	var tracingManager contract.TracingManager

	// Create application with Cache and observability enabled
	app := goe.New(goe.Options{
		WithCache:         true,
		WithObservability: true,
		Invokers: []any{
			func(c contract.Cache, tm contract.TracingManager) {
				cache = c
				tracingManager = tm

				// Test cache operations with tracing
				ctx := context.Background()

				// Start a span for cache operations
				ctx, span := tm.StartSpan(ctx, "cache_operations_test",
					contract.WithSpanKind(trace.SpanKindInternal),
					contract.WithSpanAttributes(
						attribute.String("component", "cache"),
						attribute.String("operation", "test"),
					),
				)
				defer span.End()

				// Test cache operations (these should automatically create spans if wrapped)
				err := cache.Set("test_key", "test_value", time.Minute)
				assert.NoError(t, err)
				span.AddEvent("Cache set operation completed")

				value, err := cache.Get("test_key")
				assert.NoError(t, err)
				assert.Equal(t, "test_value", value)
				span.AddEvent("Cache get operation completed")

				exists := cache.Has("test_key")
				assert.True(t, exists)
				span.AddEvent("Cache has operation completed")

				err = cache.Forget("test_key")
				assert.NoError(t, err)
				span.AddEvent("Cache forget operation completed")

				// Test Remember operation with tracing
				_, rememberSpan := tm.StartSpan(ctx, "cache_remember_test")
				defer rememberSpan.End()

				value, err = cache.Remember("remember_key", time.Minute, func() (any, error) {
					rememberSpan.AddEvent("Cache callback executed")
					return "remembered_value", nil
				})
				assert.NoError(t, err)
				assert.Equal(t, "remembered_value", value)

				rememberSpan.SetStatus(codes.Ok, "Remember operation completed")
				span.SetStatus(codes.Ok, "All cache operations completed")
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

	// Verify cache and tracing are available
	assert.NotNil(t, cache)
	assert.NotNil(t, tracingManager)

	// Stop the application
	app.Stop(ctx)
}

func TestDBTracingIntegration(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_SERVICE_NAME", "db-tracing-test")
	t.Setenv("OTEL_TRACING_ENABLED", "true")
	t.Setenv("DB_CONNECTION", "default")
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_DATABASE", ":memory:")

	var db contract.DB
	var tracingManager contract.TracingManager

	// Create application with DB and observability enabled
	app := goe.New(goe.Options{
		WithDB:            true,
		WithObservability: true,
		Invokers: []any{
			func(d contract.DB, tm contract.TracingManager) {
				// Just store the references, don't use them yet
				db = d
				tracingManager = tm
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

	// Verify DB and tracing are available
	assert.NotNil(t, db)
	assert.NotNil(t, tracingManager)

	// Now test DB operations with tracing after the application is started
	testCtx := context.Background()

	// Start a span for DB operations
	testCtx, span := tracingManager.StartSpan(testCtx, "db_operations_test",
		contract.WithSpanKind(trace.SpanKindClient),
		contract.WithSpanAttributes(
			attribute.String("db.system", "sqlite"),
			attribute.String("db.operation", "test"),
		),
	)
	defer span.End()

	instance := db.Instance()
	assert.NotNil(t, instance)

	// Create a simple test table with context
	err = instance.WithContext(testCtx).Exec("CREATE TABLE test_tracing (id INTEGER PRIMARY KEY, name TEXT)").Error
	assert.NoError(t, err)
	span.AddEvent("Table created")

	// Insert test data with context
	err = instance.WithContext(testCtx).Exec("INSERT INTO test_tracing (name) VALUES (?)", "test").Error
	assert.NoError(t, err)
	span.AddEvent("Data inserted")

	// Query test data with context
	var count int64
	err = instance.WithContext(testCtx).Model(&struct{}{}).Table("test_tracing").Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
	span.AddEvent("Data queried")

	// Test transaction with tracing
	_, txSpan := tracingManager.StartSpan(testCtx, "db_transaction_test",
		contract.WithSpanKind(trace.SpanKindClient),
		contract.WithSpanAttributes(
			attribute.String("db.operation", "transaction"),
		),
	)
	defer txSpan.End()

	err = instance.WithContext(testCtx).Transaction(func(tx *gorm.DB) error {
		txSpan.AddEvent("Transaction started")

		err := tx.Exec("INSERT INTO test_tracing (name) VALUES (?)", "tx_test").Error
		if err != nil {
			txSpan.RecordError(err)
			return err
		}

		txSpan.AddEvent("Transaction operation completed")
		return nil
	})
	assert.NoError(t, err)
	txSpan.SetStatus(codes.Ok, "Transaction completed successfully")

	span.SetStatus(codes.Ok, "All DB operations completed")

	// Stop the application
	app.Stop(ctx)
}

func TestObservabilityTracingIntegration(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_SERVICE_NAME", "observability-tracing-test")
	t.Setenv("OTEL_TRACING_ENABLED", "true")

	var observability contract.Observability

	// Create application with observability enabled
	app := goe.New(goe.Options{
		WithObservability: true,
		Invokers: []any{
			func(obs contract.Observability) {
				observability = obs

				// Test observability self-monitoring with tracing
				tracingManager := obs.Tracing()
				assert.NotNil(t, tracingManager)

				ctx := context.Background()

				// Test various span operations
				ctx, rootSpan := tracingManager.StartSpan(ctx, "observability_test",
					contract.WithSpanKind(trace.SpanKindServer),
					contract.WithSpanAttributes(
						attribute.String("component", "observability"),
						attribute.String("test.type", "integration"),
					),
				)
				defer rootSpan.End()

				// Test span attributes
				rootSpan.SetAttributes(
					attribute.String("service.name", "observability-tracing-test"),
					attribute.String("service.version", "1.0.0"),
					attribute.Int("test.iteration", 1),
				)

				// Test span events
				rootSpan.AddEvent("Test started")
				rootSpan.AddEvent("Processing observability features")

				// Test nested spans
				_, childSpan := tracingManager.StartSpan(ctx, "observability_child_operation",
					contract.WithSpanKind(trace.SpanKindInternal),
					contract.WithSpanAttributes(
						attribute.String("operation.type", "child"),
					),
				)

				childSpan.AddEvent("Child operation started")
				time.Sleep(5 * time.Millisecond) // Simulate work
				childSpan.AddEvent("Child operation completed")
				childSpan.SetStatus(codes.Ok, "Child operation successful")
				childSpan.End()

				// Test error recording
				_, errorSpan := tracingManager.StartSpan(ctx, "observability_error_test")
				defer errorSpan.End()

				testError := assert.AnError
				errorSpan.RecordError(testError)
				errorSpan.SetStatus(codes.Error, "Test error recorded")

				// Test that tracer is accessible
				tracer := tracingManager.GetTracer()
				assert.NotNil(t, tracer)

				rootSpan.AddEvent("Test completed")
				rootSpan.SetStatus(codes.Ok, "Observability test completed successfully")
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

func TestCrossModuleTracingPropagation(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_SERVICE_NAME", "cross-module-tracing-test")
	t.Setenv("OTEL_TRACING_ENABLED", "true")
	t.Setenv("HTTP_PORT", "8084")
	t.Setenv("CACHE_DRIVER", "memory")
	t.Setenv("DB_CONNECTION", "default")
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_DATABASE", ":memory:")

	var allComponentsReceived bool
	var httpKernel contract.HTTPKernel
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
				hk contract.HTTPKernel,
				c contract.Cache,
				d contract.DB,
				o contract.Observability,
			) {
				allComponentsReceived = true

				// Store references for later use
				httpKernel = hk
				cache = c
				db = d
				obs = o

				// Verify all components are available
				assert.NotNil(t, httpKernel)
				assert.NotNil(t, cache)
				assert.NotNil(t, db)
				assert.NotNil(t, obs)
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

	// Stop the application
	app.Stop(ctx)
}
