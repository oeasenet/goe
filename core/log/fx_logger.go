package log

import (
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// FxDebugLogger wraps zap.Logger and logs all Fx events at debug level
type FxDebugLogger struct {
	logger *zap.Logger
}

// NewFxDebugLogger creates a new FxDebugLogger
func NewFxDebugLogger(logger *zap.Logger) fxevent.Logger {
	return &FxDebugLogger{logger: logger}
}

// LogEvent logs Fx events at debug level
func (l *FxDebugLogger) LogEvent(event fxevent.Event) {
	// Check if debug level is enabled
	if !l.logger.Core().Enabled(zap.DebugLevel) {
		// Only log errors and important events when debug is disabled
		switch e := event.(type) {
		case *fxevent.OnStartExecuted:
			if e.Err != nil {
				l.logger.Error("OnStart hook failed",
					zap.String("callee", e.FunctionName),
					zap.String("caller", e.CallerName),
					zap.Error(e.Err),
				)
			}
		case *fxevent.OnStopExecuted:
			if e.Err != nil {
				l.logger.Error("OnStop hook failed",
					zap.String("callee", e.FunctionName),
					zap.String("caller", e.CallerName),
					zap.Error(e.Err),
				)
			}
		case *fxevent.Supplied:
			if e.Err != nil {
				l.logger.Error("Error supplying",
					zap.String("type", e.TypeName),
					zap.Error(e.Err),
				)
			}
		case *fxevent.Provided:
			if e.Err != nil {
				l.logger.Error("Error providing",
					zap.String("constructor", e.ConstructorName),
					zap.Error(e.Err),
				)
			}
		case *fxevent.Invoked:
			if e.Err != nil {
				l.logger.Error("Invoke failed",
					zap.String("function", e.FunctionName),
					zap.Error(e.Err),
				)
			}
		case *fxevent.Started:
			if e.Err != nil {
				l.logger.Error("Start failed", zap.Error(e.Err))
			} else {
				l.logger.Info("Started")
			}
		case *fxevent.Stopping:
			l.logger.Info("Stopping application",
				zap.String("signal", e.Signal.String()),
			)
		case *fxevent.Stopped:
			if e.Err != nil {
				l.logger.Error("Stop failed", zap.Error(e.Err))
			}
		case *fxevent.RollingBack:
			l.logger.Error("Rolling back", zap.Error(e.StartErr))
		case *fxevent.RolledBack:
			if e.Err != nil {
				l.logger.Error("Rollback failed", zap.Error(e.Err))
			}
		}
		return
	}

	// When debug is enabled, log everything
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.logger.Debug("OnStart hook executing",
			zap.String("callee", e.FunctionName),
			zap.String("caller", e.CallerName),
		)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.logger.Error("OnStart hook failed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.Error(e.Err),
			)
		} else {
			l.logger.Debug("OnStart hook executed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.String("runtime", e.Runtime.String()),
			)
		}
	case *fxevent.OnStopExecuting:
		l.logger.Debug("OnStop hook executing",
			zap.String("callee", e.FunctionName),
			zap.String("caller", e.CallerName),
		)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.logger.Error("OnStop hook failed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.Error(e.Err),
			)
		} else {
			l.logger.Debug("OnStop hook executed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.String("runtime", e.Runtime.String()),
			)
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			l.logger.Error("Error supplying",
				zap.String("type", e.TypeName),
				zap.Strings("stacktrace", e.StackTrace),
				zap.Strings("moduletrace", e.ModuleTrace),
				zap.Error(e.Err),
			)
		} else {
			l.logger.Debug("Supplied",
				zap.String("type", e.TypeName),
				zap.Strings("stacktrace", e.StackTrace),
				zap.Strings("moduletrace", e.ModuleTrace),
			)
		}
	case *fxevent.Provided:
		if e.Err != nil {
			l.logger.Error("Error providing",
				zap.String("constructor", e.ConstructorName),
				zap.Strings("stacktrace", e.StackTrace),
				zap.Strings("moduletrace", e.ModuleTrace),
				zap.Strings("types", e.OutputTypeNames),
				zap.Error(e.Err),
			)
		} else {
			l.logger.Debug("Provided",
				zap.String("constructor", e.ConstructorName),
				zap.Strings("stacktrace", e.StackTrace),
				zap.Strings("moduletrace", e.ModuleTrace),
				zap.Strings("types", e.OutputTypeNames),
			)
		}
	case *fxevent.Replaced:
		if e.Err != nil {
			l.logger.Error("Error replacing",
				zap.Strings("stacktrace", e.StackTrace),
				zap.Strings("moduletrace", e.ModuleTrace),
				zap.Strings("types", e.OutputTypeNames),
				zap.Error(e.Err),
			)
		} else {
			l.logger.Debug("Replaced",
				zap.Strings("stacktrace", e.StackTrace),
				zap.Strings("moduletrace", e.ModuleTrace),
				zap.Strings("types", e.OutputTypeNames),
			)
		}
	case *fxevent.Decorated:
		if e.Err != nil {
			l.logger.Error("Error decorating",
				zap.String("decorator", e.DecoratorName),
				zap.Strings("stacktrace", e.StackTrace),
				zap.Strings("moduletrace", e.ModuleTrace),
				zap.Strings("types", e.OutputTypeNames),
				zap.Error(e.Err),
			)
		} else {
			l.logger.Debug("Decorated",
				zap.String("decorator", e.DecoratorName),
				zap.Strings("stacktrace", e.StackTrace),
				zap.Strings("moduletrace", e.ModuleTrace),
				zap.Strings("types", e.OutputTypeNames),
			)
		}
	case *fxevent.Run:
		if e.Err != nil {
			l.logger.Error("Error running",
				zap.String("name", e.Name),
				zap.String("kind", e.Kind),
				zap.Error(e.Err),
			)
		} else {
			l.logger.Debug("Run",
				zap.String("name", e.Name),
				zap.String("kind", e.Kind),
				zap.String("runtime", e.Runtime.String()),
			)
		}
	case *fxevent.Invoking:
		// Module names can be noisy, so only log them at debug
		l.logger.Debug("Invoking",
			zap.String("function", e.FunctionName),
			zap.String("module", e.ModuleName),
		)
	case *fxevent.Invoked:
		if e.Err != nil {
			l.logger.Error("Invoke failed",
				zap.String("function", e.FunctionName),
				zap.String("trace", e.Trace),
				zap.Error(e.Err),
			)
		}
		// Success case is not logged to reduce noise
	case *fxevent.Stopping:
		l.logger.Info("Stopping application",
			zap.String("signal", e.Signal.String()),
		)
	case *fxevent.Stopped:
		if e.Err != nil {
			l.logger.Error("Stop failed", zap.Error(e.Err))
		}
		// Success case is not logged as the app is shutting down
	case *fxevent.RollingBack:
		l.logger.Error("Rolling back", zap.Error(e.StartErr))
	case *fxevent.RolledBack:
		if e.Err != nil {
			l.logger.Error("Rollback failed", zap.Error(e.Err))
		}
		// Success case is already covered by the error state
	case *fxevent.Started:
		if e.Err != nil {
			l.logger.Error("Start failed", zap.Error(e.Err))
		} else {
			l.logger.Info("Started")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			l.logger.Error("Error initializing custom logger", zap.Error(e.Err))
		} else {
			l.logger.Debug("Initialized custom fxevent.Logger",
				zap.String("function", e.ConstructorName),
			)
		}
	}
}

// Test tests the logger with a message at debug level
func (l *FxDebugLogger) Test(msg string) {
	l.logger.Debug(msg)
}

// String returns the logger type
func (l *FxDebugLogger) String() string {
	return "FxDebugLogger"
}
