package contract

import (
	"context"
	"go.uber.org/zap"
)

// Logger defines the logging interface
type Logger interface {
	// Debug logs a debug message
	Debug(msg string, fields ...Field)

	// Info logs an info message
	Info(msg string, fields ...Field)

	// Warn logs a warning message
	Warn(msg string, fields ...Field)

	// Error logs an error message
	Error(msg string, fields ...Field)

	// Fatal logs a fatal message and exits the application
	Fatal(msg string, fields ...Field)

	// With creates a new logger with additional fields
	With(fields ...Field) Logger

	// WithContext creates a new logger with context
	WithContext(ctx context.Context) Logger

	// WithError creates a new logger with an error field
	WithError(err error) Logger

	GetLogger() *zap.SugaredLogger
}

// Field represents a logging field
type Field interface {
	Key() string
	Value() any
}

// LoggerConfig defines logger configuration
type LoggerConfig interface {
	// Level returns the logging level
	Level() string

	// Format returns the logging format (json, text)
	Format() string

	// Output returns the output destinations
	Output() []string

	// EnableCaller returns whether to include caller information
	EnableCaller() bool

	// EnableStacktrace returns whether to include stack traces
	EnableStacktrace() bool
}
