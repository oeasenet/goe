package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.oease.dev/goe/v2/contract"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// MetricsWrapper wraps a MongoDB instance with metrics and tracing
type MetricsWrapper struct {
	mongodb contract.MongoDB
	metrics contract.MetricsManager
	tracing contract.TracingManager
}

// NewMetricsWrapper creates a new MetricsWrapper
func NewMetricsWrapper(
	mongodb contract.MongoDB,
	metrics contract.MetricsManager,
	tracing contract.TracingManager,
) contract.MongoDB {
	if metrics == nil || tracing == nil {
		return mongodb
	}
	return &MetricsWrapper{
		mongodb: mongodb,
		metrics: metrics,
		tracing: tracing,
	}
}

// Instance wraps the Instance method with metrics and tracing
func (mw *MetricsWrapper) Instance() *mongo.Database {
	ctx, span := mw.tracing.StartSpan(context.Background(), "mongodb.get_instance",
		contract.WithSpanKind(trace.SpanKindClient),
		contract.WithSpanAttributes(
			attribute.String("db.system", "mongodb"),
			attribute.String("db.operation", "get_instance"),
			attribute.String("db.connection", "default"),
		),
	)
	defer span.End()

	start := time.Now()
	db := mw.mongodb.Instance()
	duration := time.Since(start)

	// Record metrics
	mw.recordOperationMetrics(ctx, "get_instance", "default", duration, db != nil)

	// Record span status
	if db != nil {
		span.SetStatus(codes.Ok, "Instance retrieved successfully")
	} else {
		span.SetStatus(codes.Error, "Failed to retrieve instance")
	}

	return db
}

// Connection wraps the Connection method with metrics and tracing
func (mw *MetricsWrapper) Connection(name string) (*mongo.Database, error) {
	ctx, span := mw.tracing.StartSpan(context.Background(), "mongodb.get_connection",
		contract.WithSpanKind(trace.SpanKindClient),
		contract.WithSpanAttributes(
			attribute.String("db.system", "mongodb"),
			attribute.String("db.operation", "get_connection"),
			attribute.String("db.connection", name),
		),
	)
	defer span.End()

	start := time.Now()
	db, err := mw.mongodb.Connection(name)
	duration := time.Since(start)

	// Record metrics
	mw.recordOperationMetrics(ctx, "get_connection", name, duration, err == nil)

	// Record span status
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to retrieve connection")
	} else {
		span.SetStatus(codes.Ok, "Connection retrieved successfully")
	}

	return db, err
}

// recordOperationMetrics records metrics for MongoDB operations
func (mw *MetricsWrapper) recordOperationMetrics(ctx context.Context, operation, connection string, duration time.Duration, success bool) {
	// Operation counter
	counter := mw.metrics.Counter("mongodb_operations_total", contract.WithDescription("Total MongoDB operations"))
	counter.Add(ctx, 1,
		attribute.String("operation", operation),
		attribute.String("connection", connection),
		attribute.String("status", func() string {
			if success {
				return "success"
			}
			return "error"
		}()),
	)

	// Operation duration histogram
	histogram := mw.metrics.Histogram("mongodb_operation_duration_seconds", contract.WithDescription("MongoDB operation duration"))
	histogram.Record(ctx, duration.Seconds(),
		attribute.String("operation", operation),
		attribute.String("connection", connection),
	)

	// Error counter (only if operation failed)
	if !success {
		errorCounter := mw.metrics.Counter("mongodb_errors_total", contract.WithDescription("Total MongoDB errors"))
		errorCounter.Add(ctx, 1,
			attribute.String("operation", operation),
			attribute.String("connection", connection),
		)
	}
}

// NewMetricsCommandMonitor creates a command monitor that integrates with metrics and tracing
func NewMetricsCommandMonitor(
	metrics contract.MetricsManager,
	tracing contract.TracingManager,
	logger contract.Logger,
) *event.CommandMonitor {
	return &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			// Log the command start
			logger.Info("[MONGO START]",
				"command", evt.CommandName,
				"details", evt.Command.String(),
			)

			// Start tracing span
			tracing.StartSpan(ctx, "mongodb.command."+evt.CommandName,
				contract.WithSpanKind(trace.SpanKindClient),
				contract.WithSpanAttributes(
					attribute.String("db.system", "mongodb"),
					attribute.String("db.operation", evt.CommandName),
					attribute.String("db.connection_id", evt.ConnectionID),
					attribute.Int64("db.request_id", evt.RequestID),
				),
			)
		},
		Succeeded: func(ctx context.Context, evt *event.CommandSucceededEvent) {
			// Log the command success
			logger.Debug("[MONGO SUCCEED]",
				"command", evt.CommandName,
				"duration", evt.Duration.String(),
			)

			// Record metrics
			counter := metrics.Counter("mongodb_commands_total", contract.WithDescription("Total MongoDB commands"))
			counter.Add(ctx, 1,
				attribute.String("command", evt.CommandName),
				attribute.String("status", "success"),
			)

			histogram := metrics.Histogram("mongodb_command_duration_seconds", contract.WithDescription("MongoDB command duration"))
			histogram.Record(ctx, evt.Duration.Seconds(),
				attribute.String("command", evt.CommandName),
			)

			// Complete tracing span
			span := trace.SpanFromContext(ctx)
			if span != nil {
				span.SetStatus(codes.Ok, "Command succeeded")
				span.End()
			}
		},
		Failed: func(ctx context.Context, evt *event.CommandFailedEvent) {
			// Log the command failure
			logger.Error("[MONGO FAILED]",
				"command", evt.CommandName,
				"error", evt.Failure.Error(),
			)

			// Record metrics
			counter := metrics.Counter("mongodb_commands_total", contract.WithDescription("Total MongoDB commands"))
			counter.Add(ctx, 1,
				attribute.String("command", evt.CommandName),
				attribute.String("status", "error"),
			)

			errorCounter := metrics.Counter("mongodb_command_errors_total", contract.WithDescription("Total MongoDB command errors"))
			errorCounter.Add(ctx, 1,
				attribute.String("command", evt.CommandName),
				attribute.String("error", evt.Failure.Error()),
			)

			histogram := metrics.Histogram("mongodb_command_duration_seconds", contract.WithDescription("MongoDB command duration"))
			histogram.Record(ctx, evt.Duration.Seconds(),
				attribute.String("command", evt.CommandName),
			)

			// Complete tracing span with error
			span := trace.SpanFromContext(ctx)
			if span != nil {
				span.RecordError(evt.Failure)
				span.SetStatus(codes.Error, "Command failed")
				span.End()
			}
		},
	}
}

// ProvideMongoDBWithMetrics provides a MongoDB instance wrapped with metrics and tracing
func ProvideMongoDBWithMetrics(
	mongodb contract.MongoDB,
	metrics contract.MetricsManager,
	tracing contract.TracingManager,
) contract.MongoDB {
	return NewMetricsWrapper(mongodb, metrics, tracing)
}
