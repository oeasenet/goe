package contract

import (
	"context"

	"go.uber.org/zap"
)

// Log represents the logging module interface
type Log interface {
	Module

	// Debug logs a message at debug level
	Debug(msg string, fields ...interface{})

	// Info logs a message at info level
	Info(msg string, fields ...interface{})

	// Warn logs a message at warn level
	Warn(msg string, fields ...interface{})

	// Error logs a message at error level
	Error(msg string, fields ...interface{})

	// Fatal logs a message at fatal level and then exits
	Fatal(msg string, fields ...interface{})

	// WithContext returns a logger with context
	WithContext(ctx context.Context) Log

	// WithFields returns a logger with fields
	WithFields(fields map[string]interface{}) Log

	// WithField returns a logger with a field
	WithField(key string, value interface{}) Log

	// Named returns a logger with the specified name
	Named(name string) Log

	// With returns a logger with the specified fields
	With(fields ...zap.Field) Log

	// Zap returns the underlying zap logger
	Zap() *zap.Logger

	// SetLevel sets the logging level
	SetLevel(level string) error

	// GetLevel gets the current logging level
	GetLevel() string
}

// LogProvider is a function that provides a Log instance
type LogProvider func() Log
