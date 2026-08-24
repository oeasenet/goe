package log

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap/zapcore"
)

func TestWithContext_RequestID(t *testing.T) {
	logger, logs := newObservedLogger(zapcore.DebugLevel, nil)

	// The same setup GOE's http kernel uses: requestid middleware storing into
	// the request context via PassLocalsToContext.
	app := fiber.New(fiber.Config{PassLocalsToContext: true})
	app.Use(requestid.New())
	app.Get("/", func(c fiber.Ctx) error {
		// Service-layer shape: only a context.Context crosses the boundary.
		logInService(c.Context(), logger)
		return nil
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(fiber.HeaderXRequestID, "req-42")
	_, err := app.Test(req)
	require.NoError(t, err)

	entries := logs.All()
	require.Len(t, entries, 1)
	fields := entries[0].ContextMap()
	assert.Equal(t, "req-42", fields["request_id"])
}

// logInService stands in for service-layer code that holds only a
// context.Context, not the fiber ctx.
func logInService(ctx context.Context, logger *zapLogger) {
	logger.WithContext(ctx).Info("from the service layer")
}

func TestWithContext_TraceAndSpan(t *testing.T) {
	logger, logs := newObservedLogger(zapcore.DebugLevel, nil)

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
		SpanID:     trace.SpanID{0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11},
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	logger.WithContext(ctx).Info("with span")

	entries := logs.All()
	require.Len(t, entries, 1)
	fields := entries[0].ContextMap()
	assert.Equal(t, sc.TraceID().String(), fields["trace_id"])
	assert.Equal(t, sc.SpanID().String(), fields["span_id"])
}

func TestWithContext_SugaredPathCarriesFields(t *testing.T) {
	// The derived logger's sugared side must carry the fields too — Infow goes
	// through the sugar, not the structured logger.
	logger, logs := newObservedLogger(zapcore.DebugLevel, nil)

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{0xff, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
		SpanID:     trace.SpanID{0xfa, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11},
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	logger.WithContext(ctx).Infow("sugared", "extra", "kv")

	entries := logs.All()
	require.Len(t, entries, 1)
	fields := entries[0].ContextMap()
	assert.Equal(t, sc.TraceID().String(), fields["trace_id"])
	assert.Equal(t, "kv", fields["extra"])
}

func TestWithContext_EmptyContext_NoFields(t *testing.T) {
	logger, logs := newObservedLogger(zapcore.DebugLevel, nil)

	logger.WithContext(context.Background()).Info("plain")

	entries := logs.All()
	require.Len(t, entries, 1)
	fields := entries[0].ContextMap()
	assert.NotContains(t, fields, "request_id")
	assert.NotContains(t, fields, "trace_id")
	assert.NotContains(t, fields, "span_id")
}

func TestWithContext_NilContext_DoesNotPanic(t *testing.T) {
	logger, logs := newObservedLogger(zapcore.DebugLevel, nil)

	assert.NotPanics(t, func() {
		//nolint:staticcheck // deliberately exercising the nil-context guard
		logger.WithContext(nil).Info("nil ctx")
	})
	assert.Equal(t, 1, logs.Len())
}
