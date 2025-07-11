package event

import (
	"context"
	"time"

	"go.oease.dev/goe/v2/contract"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// MetricsWrapper wraps an EventManager with metrics collection
type MetricsWrapper struct {
	manager contract.EventManager
	metrics contract.MetricsManager
	tracing contract.TracingManager

	// Metrics
	publishCounter  contract.Counter
	publishDuration contract.Histogram
	consumeCounter  contract.Counter
	consumeDuration contract.Histogram
	errorCounter    contract.Counter
	queueSizeGauge  contract.Gauge
}

// NewMetricsWrapper creates a new EventManager wrapper with metrics
func NewMetricsWrapper(manager contract.EventManager, metrics contract.MetricsManager, tracing contract.TracingManager) contract.EventManager {
	if metrics == nil || tracing == nil {
		return manager // Return unwrapped manager if observability is not available
	}

	wrapper := &MetricsWrapper{
		manager: manager,
		metrics: metrics,
		tracing: tracing,
	}

	// Initialize metrics
	wrapper.publishCounter = metrics.Counter("event_publish_total",
		contract.WithDescription("Total number of events published"),
		contract.WithUnit("events"),
	)

	wrapper.publishDuration = metrics.Histogram("event_publish_duration_seconds",
		contract.WithDescription("Event publish duration in seconds"),
		contract.WithUnit("seconds"),
	)

	wrapper.consumeCounter = metrics.Counter("event_consume_total",
		contract.WithDescription("Total number of events consumed"),
		contract.WithUnit("events"),
	)

	wrapper.consumeDuration = metrics.Histogram("event_consume_duration_seconds",
		contract.WithDescription("Event consume duration in seconds"),
		contract.WithUnit("seconds"),
	)

	wrapper.errorCounter = metrics.Counter("event_errors_total",
		contract.WithDescription("Total number of event processing errors"),
		contract.WithUnit("errors"),
	)

	wrapper.queueSizeGauge = metrics.Gauge("event_queue_size",
		contract.WithDescription("Current event queue size"),
		contract.WithUnit("events"),
	)

	return wrapper
}

// Publish publishes an event with metrics
func (w *MetricsWrapper) Publish(ctx context.Context, topic string, event contract.Event) error {
	start := time.Now()

	// Start tracing span
	ctx, span := w.tracing.StartSpan(ctx, "event.publish",
		contract.WithSpanKind(trace.SpanKindProducer),
		contract.WithSpanAttributes(
			attribute.String("event.type", event.Name()),
			attribute.String("event.operation", "publish"),
		),
	)
	defer span.End()

	// Perform operation
	err := w.manager.Publish(ctx, topic, event)
	duration := time.Since(start).Seconds()

	// Common attributes
	attrs := []attribute.KeyValue{
		attribute.String("operation", "publish"),
		attribute.String("event_type", event.Name()),
	}

	// Record metrics
	w.publishCounter.Inc(ctx, attrs...)
	w.publishDuration.Record(ctx, duration, attrs...)

	if err != nil {
		errorAttrs := append(attrs, attribute.String("error", err.Error()))
		w.errorCounter.Inc(ctx, errorAttrs...)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Event publish failed")
	} else {
		span.SetStatus(codes.Ok, "Event published successfully")
	}

	return err
}

// Subscribe subscribes to events with metrics
func (w *MetricsWrapper) Subscribe(ctx context.Context, topic string, consumerGroup string, handler contract.EventHandler) error {
	// Wrap the handler with metrics
	wrappedHandler := w.wrapHandler(handler)
	return w.manager.Subscribe(ctx, topic, consumerGroup, wrappedHandler)
}

// wrapHandler wraps an event handler with metrics collection
func (w *MetricsWrapper) wrapHandler(handler contract.EventHandler) contract.EventHandler {
	return contract.EventHandlerFunc(func(ctx context.Context, event contract.Event) error {
		start := time.Now()

		// Start tracing span
		ctx, span := w.tracing.StartSpan(ctx, "event.consume",
			contract.WithSpanKind(trace.SpanKindConsumer),
			contract.WithSpanAttributes(
				attribute.String("event.type", event.Name()),
				attribute.String("event.operation", "consume"),
			),
		)
		defer span.End()

		// Perform operation
		err := handler.Handle(ctx, event)
		duration := time.Since(start).Seconds()

		// Common attributes
		attrs := []attribute.KeyValue{
			attribute.String("operation", "consume"),
			attribute.String("event_type", event.Name()),
		}

		// Record metrics
		w.consumeCounter.Inc(ctx, attrs...)
		w.consumeDuration.Record(ctx, duration, attrs...)

		if err != nil {
			errorAttrs := append(attrs, attribute.String("error", err.Error()))
			w.errorCounter.Inc(ctx, errorAttrs...)
			span.RecordError(err)
			span.SetStatus(codes.Error, "Event consume failed")
		} else {
			span.SetStatus(codes.Ok, "Event consumed successfully")
		}

		return err
	})
}

// Unsubscribe unsubscribes from events
func (w *MetricsWrapper) Unsubscribe(ctx context.Context, topic string, consumerGroup string) error {
	ctx, span := w.tracing.StartSpan(ctx, "event.unsubscribe",
		contract.WithSpanKind(trace.SpanKindInternal),
		contract.WithSpanAttributes(
			attribute.String("event.topic", topic),
			attribute.String("event.operation", "unsubscribe"),
		),
	)
	defer span.End()

	err := w.manager.Unsubscribe(ctx, topic, consumerGroup)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Event unsubscribe failed")
	} else {
		span.SetStatus(codes.Ok, "Event unsubscribed successfully")
	}

	return err
}

// Health checks the health of the event system with metrics
func (w *MetricsWrapper) Health(ctx context.Context) error {
	start := time.Now()

	ctx, span := w.tracing.StartSpan(ctx, "event.health",
		contract.WithSpanKind(trace.SpanKindInternal),
		contract.WithSpanAttributes(
			attribute.String("event.operation", "health"),
		),
	)
	defer span.End()

	err := w.manager.Health(ctx)
	duration := time.Since(start).Seconds()

	attrs := []attribute.KeyValue{
		attribute.String("operation", "health"),
	}

	// Create a health check metric
	healthDuration := w.metrics.Histogram("event_health_check_duration_seconds",
		contract.WithDescription("Event health check duration in seconds"),
		contract.WithUnit("seconds"),
	)
	healthDuration.Record(ctx, duration, attrs...)

	if err != nil {
		w.errorCounter.Inc(ctx, append(attrs, attribute.String("error", err.Error()))...)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Event health check failed")
	} else {
		span.SetStatus(codes.Ok, "Event health check passed")
	}

	return err
}

// PublishWithDelay publishes an event with a delay
func (w *MetricsWrapper) PublishWithDelay(ctx context.Context, topic string, event contract.Event, delay time.Duration) error {
	return w.manager.PublishWithDelay(ctx, topic, event, delay)
}

// PublishBatch publishes multiple events in a batch
func (w *MetricsWrapper) PublishBatch(ctx context.Context, topic string, events []contract.Event) error {
	return w.manager.PublishBatch(ctx, topic, events)
}

// Acknowledge acknowledges message processing
func (w *MetricsWrapper) Acknowledge(ctx context.Context, topic string, consumerGroup string, messageID string) error {
	return w.manager.Acknowledge(ctx, topic, consumerGroup, messageID)
}

// Reject rejects a message and optionally requeues it
func (w *MetricsWrapper) Reject(ctx context.Context, topic string, consumerGroup string, messageID string, requeue bool) error {
	return w.manager.Reject(ctx, topic, consumerGroup, messageID, requeue)
}

// Stats returns event system statistics
func (w *MetricsWrapper) Stats(ctx context.Context) (*contract.EventStats, error) {
	return w.manager.Stats(ctx)
}

// Close closes the event manager and all connections
func (w *MetricsWrapper) Close(ctx context.Context) error {
	return w.manager.Close(ctx)
}

// GetDeadLetterQueue returns the dead letter queue manager
func (w *MetricsWrapper) GetDeadLetterQueue() contract.DeadLetterQueueManager {
	return w.manager.GetDeadLetterQueue()
}
