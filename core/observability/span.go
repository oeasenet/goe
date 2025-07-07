package observability

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// span implements the Span interface
type span struct {
	otelSpan trace.Span
}

// SetAttributes sets attributes on the span
func (s *span) SetAttributes(attrs ...attribute.KeyValue) {
	s.otelSpan.SetAttributes(attrs...)
}

// SetStatus sets the status of the span
func (s *span) SetStatus(code codes.Code, description string) {
	s.otelSpan.SetStatus(code, description)
}

// RecordError records an error on the span
func (s *span) RecordError(err error, opts ...trace.EventOption) {
	s.otelSpan.RecordError(err, opts...)
}

// AddEvent adds an event to the span
func (s *span) AddEvent(name string, opts ...trace.EventOption) {
	s.otelSpan.AddEvent(name, opts...)
}

// End ends the span
func (s *span) End(opts ...trace.SpanEndOption) {
	s.otelSpan.End(opts...)
}

// GetSpan returns the underlying OpenTelemetry span
func (s *span) GetSpan() trace.Span {
	return s.otelSpan
}
