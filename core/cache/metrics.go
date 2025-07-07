package cache

import (
	"context"
	"time"

	"go.oease.dev/goe/v2/contract"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// MetricsWrapper wraps a cache instance with metrics collection
type MetricsWrapper struct {
	cache   contract.Cache
	metrics contract.MetricsManager
	tracing contract.TracingManager

	// Metrics
	operationCounter  contract.Counter
	operationDuration contract.Histogram
	hitCounter        contract.Counter
	missCounter       contract.Counter
	sizeGauge         contract.Gauge
}

// NewMetricsWrapper creates a new cache wrapper with metrics
func NewMetricsWrapper(cache contract.Cache, metrics contract.MetricsManager, tracing contract.TracingManager) contract.Cache {
	if metrics == nil || tracing == nil {
		return cache // Return unwrapped cache if observability is not available
	}

	wrapper := &MetricsWrapper{
		cache:   cache,
		metrics: metrics,
		tracing: tracing,
	}

	// Initialize metrics
	wrapper.operationCounter = metrics.Counter("cache_operations_total",
		contract.WithDescription("Total number of cache operations"),
		contract.WithUnit("operations"),
	)

	wrapper.operationDuration = metrics.Histogram("cache_operation_duration_seconds",
		contract.WithDescription("Cache operation duration in seconds"),
		contract.WithUnit("seconds"),
	)

	wrapper.hitCounter = metrics.Counter("cache_hits_total",
		contract.WithDescription("Total number of cache hits"),
		contract.WithUnit("hits"),
	)

	wrapper.missCounter = metrics.Counter("cache_misses_total",
		contract.WithDescription("Total number of cache misses"),
		contract.WithUnit("misses"),
	)

	wrapper.sizeGauge = metrics.Gauge("cache_size_bytes",
		contract.WithDescription("Estimated cache size in bytes"),
		contract.WithUnit("bytes"),
	)

	return wrapper
}

// Get retrieves a value from cache with metrics
func (w *MetricsWrapper) Get(key string) (any, error) {
	start := time.Now()
	ctx := context.Background()

	// Start tracing span
	ctx, span := w.tracing.StartSpan(ctx, "cache.get",
		contract.WithSpanKind(trace.SpanKindInternal),
		contract.WithSpanAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.operation", "get"),
		),
	)
	defer span.End()

	// Perform operation
	value, err := w.cache.Get(key)
	duration := time.Since(start).Seconds()

	// Common attributes
	attrs := []attribute.KeyValue{
		attribute.String("operation", "get"),
		attribute.String("store", w.cache.GetPrefix()),
	}

	// Record metrics
	w.operationCounter.Inc(ctx, attrs...)
	w.operationDuration.Record(ctx, duration, attrs...)

	// Record hit/miss
	if err != nil || value == nil {
		w.missCounter.Inc(ctx, attrs...)
		span.SetAttributes(attribute.Bool("cache.hit", false))
		span.SetStatus(codes.Ok, "Cache miss")
	} else {
		w.hitCounter.Inc(ctx, attrs...)
		span.SetAttributes(attribute.Bool("cache.hit", true))
		span.SetStatus(codes.Ok, "Cache hit")
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Cache operation failed")
	}

	return value, err
}

// Set stores a value in cache with metrics
func (w *MetricsWrapper) Set(key string, value any, ttl time.Duration) error {
	start := time.Now()
	ctx := context.Background()

	// Start tracing span
	ctx, span := w.tracing.StartSpan(ctx, "cache.set",
		contract.WithSpanKind(trace.SpanKindInternal),
		contract.WithSpanAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.operation", "set"),
			attribute.String("cache.ttl", ttl.String()),
		),
	)
	defer span.End()

	// Perform operation
	err := w.cache.Set(key, value, ttl)
	duration := time.Since(start).Seconds()

	// Common attributes
	attrs := []attribute.KeyValue{
		attribute.String("operation", "set"),
		attribute.String("store", w.cache.GetPrefix()),
	}

	// Record metrics
	w.operationCounter.Inc(ctx, attrs...)
	w.operationDuration.Record(ctx, duration, attrs...)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Cache operation failed")
	} else {
		span.SetStatus(codes.Ok, "Cache set successful")
	}

	return err
}

// Forever stores a value in cache forever with metrics
func (w *MetricsWrapper) Forever(key string, value any) error {
	start := time.Now()
	ctx := context.Background()

	// Start tracing span
	ctx, span := w.tracing.StartSpan(ctx, "cache.forever",
		contract.WithSpanKind(trace.SpanKindInternal),
		contract.WithSpanAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.operation", "forever"),
		),
	)
	defer span.End()

	// Perform operation
	err := w.cache.Forever(key, value)
	duration := time.Since(start).Seconds()

	// Common attributes
	attrs := []attribute.KeyValue{
		attribute.String("operation", "forever"),
		attribute.String("store", w.cache.GetPrefix()),
	}

	// Record metrics
	w.operationCounter.Inc(ctx, attrs...)
	w.operationDuration.Record(ctx, duration, attrs...)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Cache operation failed")
	} else {
		span.SetStatus(codes.Ok, "Cache forever successful")
	}

	return err
}

// Forget removes a value from cache with metrics
func (w *MetricsWrapper) Forget(key string) error {
	start := time.Now()
	ctx := context.Background()

	// Start tracing span
	ctx, span := w.tracing.StartSpan(ctx, "cache.forget",
		contract.WithSpanKind(trace.SpanKindInternal),
		contract.WithSpanAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.operation", "forget"),
		),
	)
	defer span.End()

	// Perform operation
	err := w.cache.Forget(key)
	duration := time.Since(start).Seconds()

	// Common attributes
	attrs := []attribute.KeyValue{
		attribute.String("operation", "forget"),
		attribute.String("store", w.cache.GetPrefix()),
	}

	// Record metrics
	w.operationCounter.Inc(ctx, attrs...)
	w.operationDuration.Record(ctx, duration, attrs...)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Cache operation failed")
	} else {
		span.SetStatus(codes.Ok, "Cache forget successful")
	}

	return err
}

// Flush removes all values from cache with metrics
func (w *MetricsWrapper) Flush() error {
	start := time.Now()
	ctx := context.Background()

	// Start tracing span
	ctx, span := w.tracing.StartSpan(ctx, "cache.flush",
		contract.WithSpanKind(trace.SpanKindInternal),
		contract.WithSpanAttributes(
			attribute.String("cache.operation", "flush"),
		),
	)
	defer span.End()

	// Perform operation
	err := w.cache.Flush()
	duration := time.Since(start).Seconds()

	// Common attributes
	attrs := []attribute.KeyValue{
		attribute.String("operation", "flush"),
		attribute.String("store", w.cache.GetPrefix()),
	}

	// Record metrics
	w.operationCounter.Inc(ctx, attrs...)
	w.operationDuration.Record(ctx, duration, attrs...)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Cache operation failed")
	} else {
		span.SetStatus(codes.Ok, "Cache flush successful")
	}

	return err
}

// Has checks if a key exists in cache with metrics
func (w *MetricsWrapper) Has(key string) bool {
	start := time.Now()
	ctx := context.Background()

	// Start tracing span
	ctx, span := w.tracing.StartSpan(ctx, "cache.has",
		contract.WithSpanKind(trace.SpanKindInternal),
		contract.WithSpanAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.operation", "has"),
		),
	)
	defer span.End()

	// Perform operation
	exists := w.cache.Has(key)
	duration := time.Since(start).Seconds()

	// Common attributes
	attrs := []attribute.KeyValue{
		attribute.String("operation", "has"),
		attribute.String("store", w.cache.GetPrefix()),
	}

	// Record metrics
	w.operationCounter.Inc(ctx, attrs...)
	w.operationDuration.Record(ctx, duration, attrs...)

	span.SetAttributes(attribute.Bool("cache.exists", exists))
	span.SetStatus(codes.Ok, "Cache has check completed")

	return exists
}

// Remember gets a value from cache or computes it with metrics
func (w *MetricsWrapper) Remember(key string, ttl time.Duration, callback func() (any, error)) (any, error) {
	start := time.Now()
	ctx := context.Background()

	// Start tracing span
	ctx, span := w.tracing.StartSpan(ctx, "cache.remember",
		contract.WithSpanKind(trace.SpanKindInternal),
		contract.WithSpanAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.operation", "remember"),
			attribute.String("cache.ttl", ttl.String()),
		),
	)
	defer span.End()

	// Perform operation
	value, err := w.cache.Remember(key, ttl, callback)
	duration := time.Since(start).Seconds()

	// Common attributes
	attrs := []attribute.KeyValue{
		attribute.String("operation", "remember"),
		attribute.String("store", w.cache.GetPrefix()),
	}

	// Record metrics
	w.operationCounter.Inc(ctx, attrs...)
	w.operationDuration.Record(ctx, duration, attrs...)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Cache operation failed")
	} else {
		span.SetStatus(codes.Ok, "Cache remember successful")
	}

	return value, err
}

// RememberForever gets a value from cache or computes it forever with metrics
func (w *MetricsWrapper) RememberForever(key string, callback func() (any, error)) (any, error) {
	start := time.Now()
	ctx := context.Background()

	// Start tracing span
	ctx, span := w.tracing.StartSpan(ctx, "cache.remember_forever",
		contract.WithSpanKind(trace.SpanKindInternal),
		contract.WithSpanAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.operation", "remember_forever"),
		),
	)
	defer span.End()

	// Perform operation
	value, err := w.cache.RememberForever(key, callback)
	duration := time.Since(start).Seconds()

	// Common attributes
	attrs := []attribute.KeyValue{
		attribute.String("operation", "remember_forever"),
		attribute.String("store", w.cache.GetPrefix()),
	}

	// Record metrics
	w.operationCounter.Inc(ctx, attrs...)
	w.operationDuration.Record(ctx, duration, attrs...)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Cache operation failed")
	} else {
		span.SetStatus(codes.Ok, "Cache remember forever successful")
	}

	return value, err
}

// Pull retrieves and removes a value from cache with metrics
func (w *MetricsWrapper) Pull(key string) (any, error) {
	start := time.Now()
	ctx := context.Background()

	// Start tracing span
	ctx, span := w.tracing.StartSpan(ctx, "cache.pull",
		contract.WithSpanKind(trace.SpanKindInternal),
		contract.WithSpanAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.operation", "pull"),
		),
	)
	defer span.End()

	// Perform operation
	value, err := w.cache.Pull(key)
	duration := time.Since(start).Seconds()

	// Common attributes
	attrs := []attribute.KeyValue{
		attribute.String("operation", "pull"),
		attribute.String("store", w.cache.GetPrefix()),
	}

	// Record metrics
	w.operationCounter.Inc(ctx, attrs...)
	w.operationDuration.Record(ctx, duration, attrs...)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Cache operation failed")
	} else {
		span.SetStatus(codes.Ok, "Cache pull successful")
	}

	return value, err
}

// Add stores a value only if key doesn't exist with metrics
func (w *MetricsWrapper) Add(key string, value any, ttl time.Duration) error {
	start := time.Now()
	ctx := context.Background()

	// Start tracing span
	ctx, span := w.tracing.StartSpan(ctx, "cache.add",
		contract.WithSpanKind(trace.SpanKindInternal),
		contract.WithSpanAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.operation", "add"),
			attribute.String("cache.ttl", ttl.String()),
		),
	)
	defer span.End()

	// Perform operation
	err := w.cache.Add(key, value, ttl)
	duration := time.Since(start).Seconds()

	// Common attributes
	attrs := []attribute.KeyValue{
		attribute.String("operation", "add"),
		attribute.String("store", w.cache.GetPrefix()),
	}

	// Record metrics
	w.operationCounter.Inc(ctx, attrs...)
	w.operationDuration.Record(ctx, duration, attrs...)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Cache operation failed")
	} else {
		span.SetStatus(codes.Ok, "Cache add successful")
	}

	return err
}

// Increment increments an integer value with metrics
func (w *MetricsWrapper) Increment(key string, value ...int64) (int64, error) {
	start := time.Now()
	ctx := context.Background()

	// Start tracing span
	ctx, span := w.tracing.StartSpan(ctx, "cache.increment",
		contract.WithSpanKind(trace.SpanKindInternal),
		contract.WithSpanAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.operation", "increment"),
		),
	)
	defer span.End()

	// Perform operation
	result, err := w.cache.Increment(key, value...)
	duration := time.Since(start).Seconds()

	// Common attributes
	attrs := []attribute.KeyValue{
		attribute.String("operation", "increment"),
		attribute.String("store", w.cache.GetPrefix()),
	}

	// Record metrics
	w.operationCounter.Inc(ctx, attrs...)
	w.operationDuration.Record(ctx, duration, attrs...)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Cache operation failed")
	} else {
		span.SetAttributes(attribute.Int64("cache.result", result))
		span.SetStatus(codes.Ok, "Cache increment successful")
	}

	return result, err
}

// Decrement decrements an integer value with metrics
func (w *MetricsWrapper) Decrement(key string, value ...int64) (int64, error) {
	start := time.Now()
	ctx := context.Background()

	// Start tracing span
	ctx, span := w.tracing.StartSpan(ctx, "cache.decrement",
		contract.WithSpanKind(trace.SpanKindInternal),
		contract.WithSpanAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.operation", "decrement"),
		),
	)
	defer span.End()

	// Perform operation
	result, err := w.cache.Decrement(key, value...)
	duration := time.Since(start).Seconds()

	// Common attributes
	attrs := []attribute.KeyValue{
		attribute.String("operation", "decrement"),
		attribute.String("store", w.cache.GetPrefix()),
	}

	// Record metrics
	w.operationCounter.Inc(ctx, attrs...)
	w.operationDuration.Record(ctx, duration, attrs...)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Cache operation failed")
	} else {
		span.SetAttributes(attribute.Int64("cache.result", result))
		span.SetStatus(codes.Ok, "Cache decrement successful")
	}

	return result, err
}

// Store returns the underlying cache store
func (w *MetricsWrapper) Store() contract.CacheStore {
	return w.cache.Store()
}

// GetPrefix returns the cache key prefix
func (w *MetricsWrapper) GetPrefix() string {
	return w.cache.GetPrefix()
}
