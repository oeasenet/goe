package contract

import (
	"context"

	"go.uber.org/zap"
)

// Logger defines the logging interface
type Logger interface {
	// Debug logs a debug message with any arguments
	Debug(msg string, args ...any)

	// Info logs an info message with any arguments
	Info(msg string, args ...any)

	// Warn logs a warning message with any arguments
	Warn(msg string, args ...any)

	// Error logs an error message with any arguments
	Error(msg string, args ...any)

	// Fatal logs a fatal message and exits the application with any arguments
	Fatal(msg string, args ...any)

	// Debugf logs a debug message with printf-style formatting
	Debugf(template string, args ...any)

	// Infof logs an info message with printf-style formatting
	Infof(template string, args ...any)

	// Warnf logs a warning message with printf-style formatting
	Warnf(template string, args ...any)

	// Errorf logs an error message with printf-style formatting
	Errorf(template string, args ...any)

	// Fatalf logs a fatal message and exits the application with printf-style formatting
	Fatalf(template string, args ...any)

	// Debugw logs a debug message with key-value pairs
	Debugw(msg string, keysAndValues ...any)

	// Infow logs an info message with key-value pairs
	Infow(msg string, keysAndValues ...any)

	// Warnw logs a warning message with key-value pairs
	Warnw(msg string, keysAndValues ...any)

	// Errorw logs an error message with key-value pairs
	Errorw(msg string, keysAndValues ...any)

	// Fatalw logs a fatal message and exits the application with key-value pairs
	Fatalw(msg string, keysAndValues ...any)

	// With creates a new logger with additional key-value pairs
	With(keysAndValues ...any) Logger

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
