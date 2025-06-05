package log

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.oease.dev/goe/v2/contract"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapLogger implements the Logger interface using zap
type zapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// New creates a new logger instance using zap
func New(config contract.LoggerConfig) contract.Logger {
	// Check if we're in production based on environment or explicit format
	env := os.Getenv("GOE_ENV")
	isProduction := env == "prod" || env == "production"

	// Create encoder config based on environment
	var encoderConfig zapcore.EncoderConfig

	if isProduction || config.Format() == "json" {
		// Production or JSON format: structured logging
		encoderConfig = zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
	} else {
		// Development: pretty console output
		encoderConfig = zapcore.EncoderConfig{
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    customLevelEncoder,
			EncodeTime:     customTimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   customCallerEncoder,
		}
	}

	// Create console encoder for text format
	var encoder zapcore.Encoder
	if config.Format() == "json" {
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // No color in JSON
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Parse log level
	level := zapcore.InfoLevel
	switch config.Level() {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	// Create outputs
	var cores []zapcore.Core
	for _, output := range config.Output() {
		switch output {
		case "console":
			core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
			cores = append(cores, core)
		case "file":
			file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			core := zapcore.NewCore(encoder, zapcore.AddSync(file), level)
			cores = append(cores, core)
		}
	}

	// Combine cores
	core := zapcore.NewTee(cores...)

	// Create logger options
	var opts []zap.Option
	if config.EnableCaller() {
		opts = append(opts, zap.AddCaller(), zap.AddCallerSkip(1))
	}
	if config.EnableStacktrace() {
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// Create the logger
	logger := zap.New(core, opts...)

	return &zapLogger{
		logger: logger,
		sugar:  logger.Sugar(),
	}
}

// NewDefault creates a new logger with default configuration
func NewDefault() contract.Logger {
	return New(&defaultLoggerConfig{
		level:      "info",
		format:     "text",
		output:     []string{"console"},
		caller:     true,
		stacktrace: false,
	})
}

// customTimeEncoder formats time in a readable way
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// customLevelEncoder adds colors and formatting to log levels
func customLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.DebugLevel:
		enc.AppendString("\x1b[36mDEBUG\x1b[0m") // Cyan
	case zapcore.InfoLevel:
		enc.AppendString("\x1b[32mINFO\x1b[0m") // Green
	case zapcore.WarnLevel:
		enc.AppendString("\x1b[33mWARN\x1b[0m") // Yellow
	case zapcore.ErrorLevel:
		enc.AppendString("\x1b[31mERROR\x1b[0m") // Red
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		enc.AppendString("\x1b[35mFATAL\x1b[0m") // Magenta
	}
}

// customCallerEncoder formats the caller in a short, readable way
func customCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	// Get just the filename without the full path
	short := caller.TrimmedPath()
	enc.AppendString("\x1b[90m" + short + "\x1b[0m") // Gray color
}

// Debug logs a debug message
func (l *zapLogger) Debug(msg string, fields ...contract.Field) {
	l.logger.Debug(msg, convertFields(fields)...)
}

// Info logs an info message
func (l *zapLogger) Info(msg string, fields ...contract.Field) {
	l.logger.Info(msg, convertFields(fields)...)
}

// Warn logs a warning message
func (l *zapLogger) Warn(msg string, fields ...contract.Field) {
	l.logger.Warn(msg, convertFields(fields)...)
}

// Error logs an error message
func (l *zapLogger) Error(msg string, fields ...contract.Field) {
	l.logger.Error(msg, convertFields(fields)...)
}

// Fatal logs a fatal message and exits the application
func (l *zapLogger) Fatal(msg string, fields ...contract.Field) {
	l.logger.Fatal(msg, convertFields(fields)...)
}

// With creates a new logger with additional fields
func (l *zapLogger) With(fields ...contract.Field) contract.Logger {
	return &zapLogger{
		logger: l.logger.With(convertFields(fields)...),
		sugar:  l.sugar.With(convertFieldsToArgs(fields)...),
	}
}

// WithContext creates a new logger with context
func (l *zapLogger) WithContext(ctx context.Context) contract.Logger {
	// Extract request ID or trace ID from context if available
	fields := make([]zap.Field, 0)

	// You can extract values from context here
	// Example: if requestID := ctx.Value("request_id"); requestID != nil {
	//     fields = append(fields, zap.String("request_id", requestID.(string)))
	// }

	return &zapLogger{
		logger: l.logger.With(fields...),
		sugar:  l.sugar.With(fieldsToArgs(fields)...),
	}
}

// WithError creates a new logger with an error field
func (l *zapLogger) WithError(err error) contract.Logger {
	return &zapLogger{
		logger: l.logger.With(zap.Error(err)),
		sugar:  l.sugar.With("error", err),
	}
}

// convertFields converts contract.Field to zap.Field
func convertFields(fields []contract.Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key(), f.Value())
	}
	return zapFields
}

// convertFieldsToArgs converts fields to args for sugared logger
func convertFieldsToArgs(fields []contract.Field) []interface{} {
	args := make([]interface{}, 0, len(fields)*2)
	for _, f := range fields {
		args = append(args, f.Key(), f.Value())
	}
	return args
}

// fieldsToArgs converts zap fields to args
func fieldsToArgs(fields []zap.Field) []interface{} {
	args := make([]interface{}, 0, len(fields)*2)
	for _, f := range fields {
		args = append(args, f.Key, f.Interface)
	}
	return args
}

// Module represents the log module for Fx
type Module struct {
	logger contract.Logger
	zap    *zap.Logger
}

// NewModule creates a new log module
func NewModule(config contract.Config) *Module {
	// Create logger config from app config
	logConfig := &defaultLoggerConfig{
		level:      config.GetString("LOG_LEVEL"),
		format:     config.GetString("LOG_FORMAT"),
		output:     config.GetStringSlice("LOG_OUTPUT"),
		caller:     config.GetBool("LOG_CALLER"),
		stacktrace: config.GetBool("LOG_STACKTRACE"),
	}

	// Set defaults
	if logConfig.level == "" {
		logConfig.level = "info"
	}
	if logConfig.format == "" {
		logConfig.format = "text"
	}
	if len(logConfig.output) == 0 {
		logConfig.output = []string{"console"}
	}

	logger := New(logConfig)

	// Get the underlying zap logger for Fx
	zapLogger := logger.(*zapLogger).logger

	return &Module{
		logger: logger,
		zap:    zapLogger,
	}
}

// Name returns the module name
func (m *Module) Name() string {
	return "log"
}

// OnStart is called when the module starts
func (m *Module) OnStart(ctx context.Context) error {
	m.logger.Info("Log module started")
	return nil
}

// OnStop is called when the module stops
func (m *Module) OnStop(ctx context.Context) error {
	m.logger.Info("Log module stopped")
	// Sync the logger
	_ = m.zap.Sync()
	return nil
}

// Provide returns the logger instance for Fx
func (m *Module) Provide() contract.Logger {
	return m.logger
}

// ProvideZap returns the zap logger for Fx to use
func (m *Module) ProvideZap() *zap.Logger {
	return m.zap
}

// defaultLoggerConfig implements LoggerConfig
type defaultLoggerConfig struct {
	level      string
	format     string
	output     []string
	caller     bool
	stacktrace bool
}

func (c *defaultLoggerConfig) Level() string          { return c.level }
func (c *defaultLoggerConfig) Format() string         { return c.format }
func (c *defaultLoggerConfig) Output() []string       { return c.output }
func (c *defaultLoggerConfig) EnableCaller() bool     { return c.caller }
func (c *defaultLoggerConfig) EnableStacktrace() bool { return c.stacktrace }

// field implements the Field interface
type field struct {
	key   string
	value any
}

// Key returns the field key
func (f *field) Key() string {
	return f.key
}

// Value returns the field value
func (f *field) Value() any {
	return f.value
}

// NewField creates a new field
func NewField(key string, value any) contract.Field {
	return &field{key: key, value: value}
}

// GetZapLogger extracts the underlying zap logger from a contract.Logger
// This is useful when you need to pass the logger to libraries that expect zap
func GetZapLogger(logger contract.Logger) (*zap.Logger, error) {
	if zl, ok := logger.(*zapLogger); ok {
		return zl.logger, nil
	}
	return nil, fmt.Errorf("logger is not a zap logger")
}
