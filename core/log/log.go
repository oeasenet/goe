package log

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v3/middleware/requestid"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/internal/configvalidator"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapLogger implements the Logger interface using zap
type zapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

func (l *zapLogger) GetLogger() *zap.SugaredLogger {
	return l.sugar
}

// New creates a new logger instance using zap (no per-module overrides).
func New(config contract.LoggerConfig) contract.Logger {
	return buildLogger(config, nil)
}

// buildLogger constructs the zap-backed logger. moduleOverrides may be nil, in
// which case every module follows the global level (identical to prior behavior).
func buildLogger(config contract.LoggerConfig, moduleOverrides map[string]zapcore.Level) contract.Logger {
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
			TimeKey:          "T",
			LevelKey:         "L",
			NameKey:          "N",
			CallerKey:        "C",
			FunctionKey:      zapcore.OmitKey,
			MessageKey:       "M",
			StacktraceKey:    "S",
			LineEnding:       zapcore.DefaultLineEnding,
			EncodeLevel:      customLevelEncoder,
			EncodeTime:       customTimeEncoder,
			EncodeDuration:   zapcore.StringDurationEncoder,
			EncodeCaller:     customCallerEncoder,
			ConsoleSeparator: "    ", // Add spacing between fields
		}
	}

	// Create console encoder for text format
	var encoder zapcore.Encoder
	if config.Format() == "json" {
		// JSON is the machine-readable path; the log contract (and most log
		// tooling) expects lowercase levels there. Console output keeps the
		// capitalized/colored levels for humans.
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Parse the global (root) level. Unknown values fall back to Info.
	globalLevel, ok := parseLevel(config.Level())
	if !ok {
		globalLevel = zapcore.InfoLevel
	}

	// Leaf cores are enabled at Debug; the moduleLevelCore wrapper performs all
	// real level gating so per-module overrides can go above OR below global.
	var cores []zapcore.Core
	for _, output := range config.Output() {
		switch output {
		case "console":
			cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel))
		case "file":
			file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(file), zapcore.DebugLevel))
		}
	}

	// Wrap the combined core with per-module level gating.
	levels := &moduleLevels{global: globalLevel, overrides: moduleOverrides}
	core := newModuleLevelCore(zapcore.NewTee(cores...), levels)

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

// Debug logs a debug message with any arguments
func (l *zapLogger) Debug(msg string, args ...any) {
	l.sugar.Debugw(msg, args...)
}

// Info logs an info message with any arguments
func (l *zapLogger) Info(msg string, args ...any) {
	l.sugar.Infow(msg, args...)
}

// Warn logs a warning message with any arguments
func (l *zapLogger) Warn(msg string, args ...any) {
	l.sugar.Warnw(msg, args...)
}

// Error logs an error message with any arguments
func (l *zapLogger) Error(msg string, args ...any) {
	l.sugar.Errorw(msg, args...)
}

// Fatal logs a fatal message and exits the application with any arguments
func (l *zapLogger) Fatal(msg string, args ...any) {
	l.sugar.Fatalw(msg, args...)
}

// Debugf logs a debug message with printf-style formatting
func (l *zapLogger) Debugf(template string, args ...any) {
	l.sugar.Debugf(template, args...)
}

// Infof logs an info message with printf-style formatting
func (l *zapLogger) Infof(template string, args ...any) {
	l.sugar.Infof(template, args...)
}

// Warnf logs a warning message with printf-style formatting
func (l *zapLogger) Warnf(template string, args ...any) {
	l.sugar.Warnf(template, args...)
}

// Errorf logs an error message with printf-style formatting
func (l *zapLogger) Errorf(template string, args ...any) {
	l.sugar.Errorf(template, args...)
}

// Fatalf logs a fatal message and exits the application with printf-style formatting
func (l *zapLogger) Fatalf(template string, args ...any) {
	l.sugar.Fatalf(template, args...)
}

// Debugw logs a debug message with key-value pairs
func (l *zapLogger) Debugw(msg string, keysAndValues ...any) {
	l.sugar.Debugw(msg, keysAndValues...)
}

// Infow logs an info message with key-value pairs
func (l *zapLogger) Infow(msg string, keysAndValues ...any) {
	l.sugar.Infow(msg, keysAndValues...)
}

// Warnw logs a warning message with key-value pairs
func (l *zapLogger) Warnw(msg string, keysAndValues ...any) {
	l.sugar.Warnw(msg, keysAndValues...)
}

// Errorw logs an error message with key-value pairs
func (l *zapLogger) Errorw(msg string, keysAndValues ...any) {
	l.sugar.Errorw(msg, keysAndValues...)
}

// Fatalw logs a fatal message and exits the application with key-value pairs
func (l *zapLogger) Fatalw(msg string, keysAndValues ...any) {
	l.sugar.Fatalw(msg, keysAndValues...)
}

// With creates a new logger with additional key-value pairs
func (l *zapLogger) With(keysAndValues ...any) contract.Logger {
	return &zapLogger{
		logger: l.logger.With(convertArgsToFields(keysAndValues)...),
		sugar:  l.sugar.With(keysAndValues...),
	}
}

// WithContext returns a request-scoped logger enriched from ctx: request_id
// from the fiber requestid middleware (which stores into the request context
// when PassLocalsToContext is on — GOE's default) and, when a valid
// OpenTelemetry span is present, trace_id/span_id. It gives service-layer code
// that holds only a context.Context the same correlation fields that
// http.WithReqCtx provides inside handlers. A context carrying neither returns
// the receiver unchanged.
func (l *zapLogger) WithContext(ctx context.Context) contract.Logger {
	if ctx == nil {
		return l
	}

	fields := make([]zap.Field, 0, 3)
	if rid := requestid.FromContext(ctx); rid != "" {
		fields = append(fields, zap.String("request_id", rid))
	}
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		fields = append(fields,
			zap.String("trace_id", sc.TraceID().String()),
			zap.String("span_id", sc.SpanID().String()),
		)
	}
	if len(fields) == 0 {
		return l
	}

	logger := l.logger.With(fields...)
	return &zapLogger{
		logger: logger,
		sugar:  logger.Sugar(),
	}
}

// WithError creates a new logger with an error field
func (l *zapLogger) WithError(err error) contract.Logger {
	return &zapLogger{
		logger: l.logger.With(zap.Error(err)),
		sugar:  l.sugar.With("error", err),
	}
}

// convertArgsToFields converts key-value pairs to zap.Field slice
func convertArgsToFields(keysAndValues []any) []zap.Field {
	if len(keysAndValues)%2 != 0 {
		// If odd number of arguments, add a placeholder for the last value
		keysAndValues = append(keysAndValues, "MISSING_VALUE")
	}

	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			key = fmt.Sprintf("key_%d", i/2)
		}
		value := keysAndValues[i+1]
		fields = append(fields, zap.Any(key, value))
	}
	return fields
}

// Module represents the log module for Fx
type Module struct {
	logger contract.Logger
	zap    *zap.Logger
	config contract.Config

	// optErrs holds failures from Option application. They are reported by
	// ValidateConfig so that startup aborts before the application serves.
	optErrs []error
}

// NewModule creates a new log module.
//
// Configuration resolves in layers: LOG_*/APP_* environment variables first,
// then opts, so anything set through an Option wins over the environment.
func NewModule(config contract.Config, opts ...Option) *Module {
	// Apply code-first options. A failed option applies nothing; its error is
	// collected and aborts startup through OnStart (and ValidateConfig).
	var s settings
	var optErrs []error
	for _, opt := range opts {
		if err := opt(&s); err != nil {
			optErrs = append(optErrs, err)
		}
	}

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

	overrides, invalid := parseModuleLevels(config.GetString("LOG_MODULE_LEVELS"))
	overrides = withDefaultModuleLevels(overrides)
	logger := buildLogger(logConfig, overrides)

	// Base identity fields (service/env/version, plus WithBaseFields extras)
	// are stamped once at construction so every JSON line carries them,
	// including the fx logger derived below. Text output is for humans and
	// stays clean.
	if logConfig.format == "json" {
		base := append(baseIdentityFields(config, s.version), s.baseFields...)
		if len(base) > 0 {
			logger = logger.With(base...)
		}
	}
	if len(invalid) > 0 {
		logger.With(moduleFieldKey, "log").Warn(
			"Ignored invalid LOG_MODULE_LEVELS entries",
			"entries", invalid,
		)
	}

	// Get the underlying zap logger for Fx, tagged so that its entries are
	// subject to per-module gating like every other module's.
	zapLogger := logger.(*zapLogger).logger.With(zap.String(moduleFieldKey, fxModuleName))

	return &Module{
		logger:  logger,
		zap:     zapLogger,
		config:  config,
		optErrs: optErrs,
	}
}

// baseIdentityFields resolves the identity fields stamped on every JSON log
// line: service ← APP_NAME, env ← OEASE_ENV (falling back to GOE_ENV),
// version ← versionOverride (from WithVersion) falling back to APP_VERSION.
// Unset values are omitted rather than emitted empty — the log pipeline owns
// service/env from deployment metadata, but version can only come from the
// process.
func baseIdentityFields(config contract.Config, versionOverride string) []any {
	kv := make([]any, 0, 6)
	if service := config.GetString("APP_NAME"); service != "" {
		kv = append(kv, "service", service)
	}
	env := config.GetString("OEASE_ENV")
	if env == "" {
		env = config.GetString("GOE_ENV")
	}
	if env != "" {
		kv = append(kv, "env", env)
	}
	version := versionOverride
	if version == "" {
		version = config.GetString("APP_VERSION")
	}
	if version != "" {
		kv = append(kv, "version", version)
	}
	return kv
}

// Name returns the module name
func (m *Module) Name() string {
	return "log"
}

// OnStart is called when the module starts
func (m *Module) OnStart(ctx context.Context) error {
	// Built-in modules are not registered with the startup validator, so a
	// failed Option aborts startup here — the job module follows the same
	// pattern. Bootstrap logging has already worked on the environment-only
	// configuration, so the operator sees these errors on a working logger.
	if len(m.optErrs) > 0 {
		return errors.Join(m.optErrs...)
	}

	m.logger.With(moduleFieldKey, "log").Info("Log module started")
	return nil
}

// OnStop is called when the module stops
func (m *Module) OnStop(ctx context.Context) error {
	m.logger.With(moduleFieldKey, "log").Info("Log module stopped")
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

// ValidateConfig validates the log module configuration
func (m *Module) ValidateConfig() error {
	// Option failures are reported first and abort startup, mirroring the
	// http module's handling of its Options.
	if len(m.optErrs) > 0 {
		return errors.Join(m.optErrs...)
	}

	v := configvalidator.NewConfigValidator(m.config, "log")

	// Log level validation
	if m.config.Has("LOG_LEVEL") {
		// Only these are honored by parseLevel; panic/fatal are not valid thresholds.
		validLevels := []string{"debug", "info", "warn", "error"}
		v.Optional("LOG_LEVEL", "Log level", configvalidator.ValidateOneOf(validLevels...))
	}

	// Log format validation. "text" is the module's own default; "console" is
	// accepted as its historical alias.
	if m.config.Has("LOG_FORMAT") {
		validFormats := []string{"json", "text", "console"}
		v.Optional("LOG_FORMAT", "Log format", configvalidator.ValidateOneOf(validFormats...))
	}

	return v.Validate()
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
