package db

import (
	"context"
	"fmt"
	"time"

	"go.oease.dev/goe/v2/contract"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

// MetricsPlugin is a GORM plugin that adds metrics collection
type MetricsPlugin struct {
	metrics contract.MetricsManager
	tracing contract.TracingManager

	// Metrics
	queryCounter    contract.Counter
	queryDuration   contract.Histogram
	connectionGauge contract.Gauge
	errorCounter    contract.Counter
}

// NewMetricsPlugin creates a new GORM metrics plugin
func NewMetricsPlugin(metrics contract.MetricsManager, tracing contract.TracingManager) *MetricsPlugin {
	if metrics == nil || tracing == nil {
		return nil // Return nil if observability is not available
	}

	plugin := &MetricsPlugin{
		metrics: metrics,
		tracing: tracing,
	}

	// Initialize metrics
	plugin.queryCounter = metrics.Counter("db_queries_total",
		contract.WithDescription("Total number of database queries"),
		contract.WithUnit("queries"),
	)

	plugin.queryDuration = metrics.Histogram("db_query_duration_seconds",
		contract.WithDescription("Database query duration in seconds"),
		contract.WithUnit("seconds"),
	)

	plugin.connectionGauge = metrics.Gauge("db_connections_active",
		contract.WithDescription("Number of active database connections"),
		contract.WithUnit("connections"),
	)

	plugin.errorCounter = metrics.Counter("db_errors_total",
		contract.WithDescription("Total number of database errors"),
		contract.WithUnit("errors"),
	)

	return plugin
}

// Name returns the plugin name
func (p *MetricsPlugin) Name() string {
	return "goe:metrics"
}

// Initialize initializes the plugin
func (p *MetricsPlugin) Initialize(db *gorm.DB) error {
	if p == nil {
		return nil // Skip if plugin is nil (observability not available)
	}

	// Register callbacks for different operations
	err := db.Callback().Create().Before("gorm:create").Register("goe:metrics:before_create", p.beforeCallback)
	if err != nil {
		return err
	}
	err = db.Callback().Create().After("gorm:create").Register("goe:metrics:after_create", p.afterCallback)
	if err != nil {
		return err
	}

	err = db.Callback().Query().Before("gorm:query").Register("goe:metrics:before_query", p.beforeCallback)
	if err != nil {
		return err
	}
	err = db.Callback().Query().After("gorm:query").Register("goe:metrics:after_query", p.afterCallback)
	if err != nil {
		return err
	}

	err = db.Callback().Update().Before("gorm:update").Register("goe:metrics:before_update", p.beforeCallback)
	if err != nil {
		return err
	}
	err = db.Callback().Update().After("gorm:update").Register("goe:metrics:after_update", p.afterCallback)
	if err != nil {
		return err
	}

	err = db.Callback().Delete().Before("gorm:delete").Register("goe:metrics:before_delete", p.beforeCallback)
	if err != nil {
		return err
	}
	err = db.Callback().Delete().After("gorm:delete").Register("goe:metrics:after_delete", p.afterCallback)
	if err != nil {
		return err
	}

	return nil
}

// beforeCallback is called before database operations
func (p *MetricsPlugin) beforeCallback(db *gorm.DB) {
	if p == nil {
		return
	}

	// Start timing
	startTime := time.Now()
	db.Set("goe:metrics:start_time", startTime)

	// Get operation type
	operation := getOperationType(db)

	// Start tracing span
	ctx := db.Statement.Context
	if ctx == nil {
		ctx = context.Background()
	}

	spanName := fmt.Sprintf("db.%s", operation)
	ctx, span := p.tracing.StartSpan(ctx, spanName,
		contract.WithSpanKind(trace.SpanKindClient),
		contract.WithSpanAttributes(
			attribute.String("db.operation", operation),
			attribute.String("db.table", db.Statement.Table),
			attribute.String("db.system", "gorm"),
		),
	)

	// Store span in context
	db.Set("goe:metrics:span", span)
	db.Statement.Context = ctx
}

// afterCallback is called after database operations
func (p *MetricsPlugin) afterCallback(db *gorm.DB) {
	if p == nil {
		return
	}

	// Get start time
	startTimeValue, exists := db.Get("goe:metrics:start_time")
	if !exists {
		return
	}
	startTime, ok := startTimeValue.(time.Time)
	if !ok {
		return
	}

	// Calculate duration
	duration := time.Since(startTime).Seconds()

	// Get operation type
	operation := getOperationType(db)

	// Get span
	spanValue, exists := db.Get("goe:metrics:span")
	var span contract.Span
	if exists {
		span, _ = spanValue.(contract.Span)
	}

	// Common attributes
	attrs := []attribute.KeyValue{
		attribute.String("operation", operation),
		attribute.String("table", db.Statement.Table),
	}

	ctx := db.Statement.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// Record metrics
	p.queryCounter.Inc(ctx, attrs...)
	p.queryDuration.Record(ctx, duration, attrs...)

	// Handle errors
	if db.Error != nil {
		errorAttrs := append(attrs, attribute.String("error", db.Error.Error()))
		p.errorCounter.Inc(ctx, errorAttrs...)

		if span != nil {
			span.RecordError(db.Error)
			span.SetStatus(codes.Error, "Database operation failed")
		}
	} else {
		if span != nil {
			span.SetStatus(codes.Ok, "Database operation successful")
		}
	}

	// Update span with additional information
	if span != nil {
		span.SetAttributes(
			attribute.Float64("db.duration", duration),
			attribute.Int64("db.rows_affected", db.RowsAffected),
		)

		// Add SQL query if available (be careful with sensitive data)
		if db.Statement.SQL.String() != "" {
			// Only log the SQL structure, not the actual values for security
			span.SetAttributes(attribute.String("db.statement", db.Statement.SQL.String()))
		}

		span.End()
	}
}

// getOperationType determines the type of database operation
func getOperationType(db *gorm.DB) string {
	switch {
	case db.Statement.SQL.String() != "":
		// Raw SQL
		sql := db.Statement.SQL.String()
		if len(sql) > 6 {
			return fmt.Sprintf("raw_%s", sql[:6])
		}
		return "raw"
	case db.Statement.Schema != nil:
		// Determine operation based on context
		if db.Statement.Context != nil {
			if operation := db.Statement.Context.Value("gorm:operation"); operation != nil {
				return fmt.Sprintf("%v", operation)
			}
		}
		// Fallback to checking the statement
		return "query"
	default:
		return "unknown"
	}
}

// AddMetricsToGORM adds metrics plugin to a GORM instance
func AddMetricsToGORM(db *gorm.DB, metrics contract.MetricsManager, tracing contract.TracingManager) error {
	plugin := NewMetricsPlugin(metrics, tracing)
	if plugin == nil {
		return nil // Skip if observability is not available
	}

	return db.Use(plugin)
}

// MetricsWrapper wraps a DB instance with metrics collection
type MetricsWrapper struct {
	db      contract.DB
	metrics contract.MetricsManager
	tracing contract.TracingManager
}

// NewMetricsWrapper creates a new DB wrapper with metrics
func NewMetricsWrapper(db contract.DB, metrics contract.MetricsManager, tracing contract.TracingManager) contract.DB {
	if metrics == nil || tracing == nil {
		return db // Return unwrapped DB if observability is not available
	}

	wrapper := &MetricsWrapper{
		db:      db,
		metrics: metrics,
		tracing: tracing,
	}

	// Note: We don't add metrics here immediately because the database
	// might not be connected yet. Metrics will be added lazily when
	// the instance is first accessed.

	return wrapper
}

// Instance returns the underlying GORM DB instance with metrics
func (w *MetricsWrapper) Instance() *gorm.DB {
	instance := w.db.Instance()
	if instance != nil {
		// Ensure metrics are added lazily
		AddMetricsToGORM(instance, w.metrics, w.tracing)
	}
	return instance
}

// Connection returns a specific GORM DB instance by name with metrics
func (w *MetricsWrapper) Connection(name string) (*gorm.DB, error) {
	conn, err := w.db.Connection(name)
	if err != nil {
		return nil, err
	}
	if conn != nil {
		// Ensure metrics are added
		AddMetricsToGORM(conn, w.metrics, w.tracing)
	}
	return conn, nil
}

// AutoMigrate performs auto migration with metrics
func (w *MetricsWrapper) AutoMigrate(dst ...interface{}) error {
	start := time.Now()
	ctx := context.Background()

	// Start tracing span
	ctx, span := w.tracing.StartSpan(ctx, "db.auto_migrate",
		contract.WithSpanKind(trace.SpanKindClient),
		contract.WithSpanAttributes(
			attribute.String("db.operation", "auto_migrate"),
			attribute.Int("db.models_count", len(dst)),
		),
	)
	defer span.End()

	// Perform migration
	err := w.db.AutoMigrate(dst...)
	duration := time.Since(start).Seconds()

	// Record metrics
	migrationCounter := w.metrics.Counter("db_migrations_total",
		contract.WithDescription("Total number of database migrations"),
		contract.WithUnit("migrations"),
	)
	migrationDuration := w.metrics.Histogram("db_migration_duration_seconds",
		contract.WithDescription("Database migration duration in seconds"),
		contract.WithUnit("seconds"),
	)

	attrs := []attribute.KeyValue{
		attribute.String("operation", "auto_migrate"),
		attribute.Int("models_count", len(dst)),
	}

	migrationCounter.Inc(ctx, attrs...)
	migrationDuration.Record(ctx, duration, attrs...)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Migration failed")
	} else {
		span.SetStatus(codes.Ok, "Migration successful")
	}

	return err
}

// AutoMigrateOnConnection performs auto migration on a specific connection with metrics
func (w *MetricsWrapper) AutoMigrateOnConnection(connectionName string, dst ...interface{}) error {
	start := time.Now()
	ctx := context.Background()

	// Start tracing span
	ctx, span := w.tracing.StartSpan(ctx, "db.auto_migrate_on_connection",
		contract.WithSpanKind(trace.SpanKindClient),
		contract.WithSpanAttributes(
			attribute.String("db.operation", "auto_migrate_on_connection"),
			attribute.String("db.connection", connectionName),
			attribute.Int("db.models_count", len(dst)),
		),
	)
	defer span.End()

	// Perform migration
	err := w.db.AutoMigrateOnConnection(connectionName, dst...)
	duration := time.Since(start).Seconds()

	// Record metrics
	migrationCounter := w.metrics.Counter("db_migrations_total",
		contract.WithDescription("Total number of database migrations"),
		contract.WithUnit("migrations"),
	)
	migrationDuration := w.metrics.Histogram("db_migration_duration_seconds",
		contract.WithDescription("Database migration duration in seconds"),
		contract.WithUnit("seconds"),
	)

	attrs := []attribute.KeyValue{
		attribute.String("operation", "auto_migrate_on_connection"),
		attribute.String("connection", connectionName),
		attribute.Int("models_count", len(dst)),
	}

	migrationCounter.Inc(ctx, attrs...)
	migrationDuration.Record(ctx, duration, attrs...)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Migration failed")
	} else {
		span.SetStatus(codes.Ok, "Migration successful")
	}

	return err
}

// RegisterModelsForMigration pre-registers models for automatic migration
func (w *MetricsWrapper) RegisterModelsForMigration(dst ...interface{}) {
	w.db.RegisterModelsForMigration(dst...)
}

// RegisterModelsForMigrationOnConnection pre-registers models for automatic migration on a specific connection
func (w *MetricsWrapper) RegisterModelsForMigrationOnConnection(connectionName string, dst ...interface{}) {
	w.db.RegisterModelsForMigrationOnConnection(connectionName, dst...)
}
