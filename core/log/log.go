package log

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"go.oease.dev/goe/v2/contract"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger implements the contract.Log interface
type Logger struct {
	mu     sync.RWMutex
	logger *zap.Logger
	level  zapcore.Level
	env    string
}

// New creates a new Logger instance
func New() contract.Log {
	// Determine environment
	env := os.Getenv("GOE_ENV")
	if env == "" {
		env = "dev" // Default to development environment
	}

	// Create logger based on environment
	var logger *zap.Logger
	var level zapcore.Level

	if strings.ToLower(env) == "dev" || strings.ToLower(env) == "development" {
		// Development logger with colors
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		logger, _ = config.Build(zap.AddCallerSkip(1))
		level = zapcore.DebugLevel
	} else {
		// Production logger
		config := zap.NewProductionConfig()
		logger, _ = config.Build(zap.AddCallerSkip(1))
		level = zapcore.InfoLevel
	}
	return &Logger{
		logger: logger,
		level:  level,
		env:    env,
	}
}

// Name returns the name of the module
func (l *Logger) Name() string {
	return "log"
}

// Initialize initializes the log module
func (l *Logger) Initialize(ctx context.Context) error {
	return nil
}

// Start starts the log module
func (l *Logger) Start(ctx context.Context) error {
	return nil
}

// Stop stops the log module
func (l *Logger) Stop(ctx context.Context) error {
	// Ignore "sync /dev/stderr: inappropriate ioctl for device" error
	// This is a known issue with zap logger when running in certain environments
	err := l.logger.Sync()
	if err != nil && strings.Contains(err.Error(), "inappropriate ioctl for device") {
		return nil
	}
	return err
}

// Debug logs a message at debug level
func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	zapFields := convertToZapFields(fields)
	l.logger.Debug(msg, zapFields...)
}

// Info logs a message at info level
func (l *Logger) Info(msg string, fields ...interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	zapFields := convertToZapFields(fields)
	l.logger.Info(msg, zapFields...)
}

// Warn logs a message at warn level
func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	zapFields := convertToZapFields(fields)
	l.logger.Warn(msg, zapFields...)
}

// Error logs a message at error level
func (l *Logger) Error(msg string, fields ...interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	zapFields := convertToZapFields(fields)
	l.logger.Error(msg, zapFields...)
}

// Fatal logs a message at fatal level and then exits
func (l *Logger) Fatal(msg string, fields ...interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	zapFields := convertToZapFields(fields)
	l.logger.Fatal(msg, zapFields...)
}

// WithContext returns a logger with context
func (l *Logger) WithContext(ctx context.Context) contract.Log {
	return l
}

// WithFields returns a logger with fields
func (l *Logger) WithFields(fields map[string]interface{}) contract.Log {
	l.mu.RLock()
	defer l.mu.RUnlock()

	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}

	newLogger := &Logger{
		logger: l.logger.With(zapFields...),
		level:  l.level,
		env:    l.env,
	}

	return newLogger
}

// WithField returns a logger with a field
func (l *Logger) WithField(key string, value interface{}) contract.Log {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newLogger := &Logger{
		logger: l.logger.With(zap.Any(key, value)),
		level:  l.level,
		env:    l.env,
	}

	return newLogger
}

// Named returns a logger with the specified name
func (l *Logger) Named(name string) contract.Log {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newLogger := &Logger{
		logger: l.logger.Named(name),
		level:  l.level,
		env:    l.env,
	}

	return newLogger
}

// With returns a logger with the specified fields
func (l *Logger) With(fields ...zap.Field) contract.Log {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newLogger := &Logger{
		logger: l.logger.With(fields...),
		level:  l.level,
		env:    l.env,
	}

	return newLogger
}

// Zap returns the underlying zap logger
func (l *Logger) Zap() *zap.Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.logger
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Convert string level to zapcore.Level
	var zapLevel zapcore.Level
	switch strings.ToLower(level) {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn", "warning":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	case "fatal":
		zapLevel = zapcore.FatalLevel
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}

	// Update the logger's level
	l.level = zapLevel

	return nil
}

// GetLevel gets the current logging level
func (l *Logger) GetLevel() string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Convert zapcore.Level to string
	switch l.level {
	case zapcore.DebugLevel:
		return "debug"
	case zapcore.InfoLevel:
		return "info"
	case zapcore.WarnLevel:
		return "warn"
	case zapcore.ErrorLevel:
		return "error"
	case zapcore.FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}

// convertToZapFields converts interface{} fields to zap.Field
func convertToZapFields(fields []interface{}) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))

	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if !ok {
				continue
			}
			zapFields = append(zapFields, zap.Any(key, fields[i+1]))
		}
	}

	return zapFields
}

// Provider provides a Log instance
func Provider() contract.Log {
	return New()
}
