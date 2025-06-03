package app

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"go.oease.dev/goe/v2/contract"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// App implements the contract.App interface
type App struct {
	mu        sync.RWMutex
	container *fx.App
	ctx       context.Context
	cancel    context.CancelFunc
	modules   map[string]contract.Module
	options   []fx.Option
	log       contract.Log
}

// New creates a new App instance
func New() *App {
	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		ctx:     ctx,
		cancel:  cancel,
		modules: make(map[string]contract.Module),
		options: []fx.Option{},
	}
}

// Name returns the name of the module
func (a *App) Name() string {
	return "app"
}

// Initialize initializes the app module
func (a *App) Initialize(ctx context.Context) error {
	// Initialize all registered modules
	for _, module := range a.modules {
		if err := module.Initialize(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Start starts the app module
func (a *App) Start(ctx context.Context) error {
	// Start all registered modules
	for _, module := range a.modules {
		if err := module.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Stop stops the app module
func (a *App) Stop(ctx context.Context) error {
	// Stop all registered modules in reverse order
	modules := make([]contract.Module, 0, len(a.modules))
	for _, module := range a.modules {
		modules = append(modules, module)
	}

	for i := len(modules) - 1; i >= 0; i-- {
		if err := modules[i].Stop(ctx); err != nil {
			return err
		}
	}

	// Cancel the context
	a.cancel()
	return nil
}

// Container returns the dependency injection container
func (a *App) Container() interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.container
}

// Context returns the application context
func (a *App) Context() context.Context {
	return a.ctx
}

// Run starts the application and blocks until it's stopped
func (a *App) Run() error {
	// Create the Fx application
	if err := a.createFxApp(); err != nil {
		return err
	}

	// Start the application and block until it's stopped
	// Fx's Run() method already handles signal handling and graceful shutdown
	a.container.Run()
	return nil
}

// createFxApp creates the Fx application with all registered modules and providers
func (a *App) createFxApp() error {
	// Get the logger before acquiring the lock to avoid deadlock
	logger := a.logger()

	a.mu.Lock()
	defer a.mu.Unlock()

	// Configure Fx to use our logger
	a.options = append(a.options, fx.WithLogger(func() fxevent.Logger {
		return &fxLogger{log: logger}
	}))

	// Enable Fx's dependency graph visualization (commented out as it's not available in this version)
	// a.options = append(a.options, fx.Visualize())

	// Add lifecycle hooks for all registered modules
	a.options = append(a.options, fx.Invoke(func(lifecycle fx.Lifecycle) {
		// Sort modules by name for consistent initialization order
		moduleNames := make([]string, 0, len(a.modules))
		for name := range a.modules {
			moduleNames = append(moduleNames, name)
		}
		sort.Strings(moduleNames)

		// Register lifecycle hooks for each module in order
		for _, name := range moduleNames {
			module := a.modules[name]
			moduleName := name       // Create a copy of the name for the closure
			moduleInstance := module // Create a copy of the module for the closure

			lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					a.logger().Info("Starting module", "module", moduleName)
					if err := moduleInstance.Initialize(ctx); err != nil {
						return fmt.Errorf("failed to initialize module %s: %w", moduleName, err)
					}
					return moduleInstance.Start(ctx)
				},
				OnStop: func(ctx context.Context) error {
					a.logger().Info("Stopping module", "module", moduleName)
					return moduleInstance.Stop(ctx)
				},
			})
		}
	}))

	// Provide all modules as dependencies to make them available for injection
	for name, module := range a.modules {
		moduleName := name
		moduleInstance := module

		// Use fx.Annotate to provide each module with a name tag for dependency injection
		a.options = append(a.options, fx.Provide(
			fx.Annotate(
				func() contract.Module { return moduleInstance },
				fx.ResultTags(`name:"`+moduleName+`"`),
			),
		))

		// Also provide the specific module type if it implements a known contract interface
		// This allows for both named and type-based injection
		switch m := moduleInstance.(type) {
		case contract.Config:
			a.options = append(a.options, fx.Provide(func() contract.Config { return m }))
		case contract.Http:
			a.options = append(a.options, fx.Provide(func() contract.Http { return m }))
		case contract.Log:
			a.options = append(a.options, fx.Provide(func() contract.Log { return m }))
		case contract.Event:
			a.options = append(a.options, fx.Provide(func() contract.Event { return m }))
		case contract.Cache:
			a.options = append(a.options, fx.Provide(func() contract.Cache { return m }))
		}
	}

	// Create the Fx application with all registered options
	a.container = fx.New(a.options...)
	return nil
}

// RunWithTimeout starts the application and returns after the specified timeout
func (a *App) RunWithTimeout(timeout time.Duration) error {
	// Create the Fx application
	if err := a.createFxApp(); err != nil {
		return err
	}

	// Start the application in a goroutine
	go a.container.Run()

	// Wait for the timeout
	time.Sleep(timeout)

	// Stop the application
	return a.Stop(context.Background())
}

// RegisterModule registers a module with the application
func (a *App) RegisterModule(module contract.Module) contract.App {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.modules[module.Name()] = module
	return a
}

// RegisterModules registers multiple modules with the application
func (a *App) RegisterModules(modules ...contract.Module) contract.App {
	for _, module := range modules {
		a.RegisterModule(module)
	}
	return a
}

// RegisterProvider registers a provider function with the application
func (a *App) RegisterProvider(provider interface{}, opts ...interface{}) contract.App {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Convert generic opts to fx.Option
	fxOpts := make([]fx.Option, 0, len(opts))
	for _, opt := range opts {
		if fxOpt, ok := opt.(fx.Option); ok {
			fxOpts = append(fxOpts, fxOpt)
		}
	}

	a.options = append(a.options, fx.Provide(provider))
	a.options = append(a.options, fxOpts...)
	return a
}

// RegisterProviders registers multiple provider functions with the application
func (a *App) RegisterProviders(providers ...interface{}) contract.App {
	for _, provider := range providers {
		a.RegisterProvider(provider)
	}
	return a
}

// Invoke executes a function after the application has started
func (a *App) Invoke(function interface{}, opts ...interface{}) contract.App {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Convert generic opts to fx.Option
	fxOpts := make([]fx.Option, 0, len(opts))
	for _, opt := range opts {
		if fxOpt, ok := opt.(fx.Option); ok {
			fxOpts = append(fxOpts, fxOpt)
		}
	}

	a.options = append(a.options, fx.Invoke(function))
	a.options = append(a.options, fxOpts...)
	return a
}

// OnStart registers a function to be called when the application starts
func (a *App) OnStart(function interface{}) contract.App {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.options = append(a.options, fx.Invoke(function))
	return a
}

// OnStop registers a function to be called when the application stops
func (a *App) OnStop(function interface{}) contract.App {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Use fx.Hook to register a stop hook
	a.options = append(a.options, fx.Invoke(func(lifecycle fx.Lifecycle) {
		lifecycle.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				if fn, ok := function.(func() error); ok {
					return fn()
				}
				if fn, ok := function.(func(context.Context) error); ok {
					return fn(ctx)
				}
				if fn, ok := function.(func()); ok {
					fn()
					return nil
				}
				return nil
			},
		})
	}))
	return a
}

// logger returns the logger for the app
func (a *App) logger() contract.Log {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// If no logger is set, create a default one
	if a.log == nil {
		// Try to find a logger in the modules
		for _, module := range a.modules {
			if logger, ok := module.(contract.Log); ok {
				a.log = logger
				break
			}
		}

		// If still no logger, create a default one
		if a.log == nil {
			// Use a simple zap logger
			logger, _ := zap.NewProduction()
			a.log = &defaultLogger{logger: logger}
		}
	}

	return a.log
}

// defaultLogger is a simple implementation of contract.Log
type defaultLogger struct {
	logger *zap.Logger
}

// fxLogger is an adapter that implements fxevent.Logger using our contract.Log
type fxLogger struct {
	log contract.Log
}

// LogEvent logs an Fx event using our logger
func (l *fxLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.log.Debug("OnStart hook executing",
			"callee", e.FunctionName,
			"caller", e.CallerName,
		)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.log.Error("OnStart hook failed",
				"callee", e.FunctionName,
				"caller", e.CallerName,
				"error", e.Err,
			)
		} else {
			l.log.Debug("OnStart hook executed",
				"callee", e.FunctionName,
				"caller", e.CallerName,
				"runtime", e.Runtime.String(),
			)
		}
	case *fxevent.OnStopExecuting:
		l.log.Debug("OnStop hook executing",
			"callee", e.FunctionName,
			"caller", e.CallerName,
		)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.log.Error("OnStop hook failed",
				"callee", e.FunctionName,
				"caller", e.CallerName,
				"error", e.Err,
			)
		} else {
			l.log.Debug("OnStop hook executed",
				"callee", e.FunctionName,
				"caller", e.CallerName,
				"runtime", e.Runtime.String(),
			)
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			l.log.Error("Error supplying value",
				"type", e.TypeName,
				"error", e.Err,
			)
		} else {
			l.log.Debug("Supplied value",
				"type", e.TypeName,
			)
		}
	case *fxevent.Provided:
		for _, rtype := range e.OutputTypeNames {
			l.log.Debug("Provided value",
				"constructor", e.ConstructorName,
				"type", rtype,
			)
		}
		if e.Err != nil {
			l.log.Error("Error providing value",
				"constructor", e.ConstructorName,
				"error", e.Err,
			)
		}
	case *fxevent.Invoking:
		l.log.Debug("Invoking function",
			"function", e.FunctionName,
		)
	case *fxevent.Invoked:
		if e.Err != nil {
			l.log.Error("Error invoking function",
				"function", e.FunctionName,
				"error", e.Err,
			)
		} else {
			l.log.Debug("Invoked function",
				"function", e.FunctionName,
			)
		}
	case *fxevent.Started:
		if e.Err != nil {
			l.log.Error("Error starting application",
				"error", e.Err,
			)
		} else {
			l.log.Info("Application started")
		}
	case *fxevent.Stopped:
		if e.Err != nil {
			l.log.Error("Error stopping application",
				"error", e.Err,
			)
		} else {
			l.log.Info("Application stopped")
		}
	}
}

func (l *defaultLogger) Name() string {
	return "default-logger"
}

func (l *defaultLogger) Initialize(ctx context.Context) error {
	return nil
}

func (l *defaultLogger) Start(ctx context.Context) error {
	return nil
}

func (l *defaultLogger) Stop(ctx context.Context) error {
	return l.logger.Sync()
}

func (l *defaultLogger) Debug(msg string, fields ...interface{}) {
	zapFields := convertToZapFields(fields)
	l.logger.Debug(msg, zapFields...)
}

func (l *defaultLogger) Info(msg string, fields ...interface{}) {
	zapFields := convertToZapFields(fields)
	l.logger.Info(msg, zapFields...)
}

func (l *defaultLogger) Warn(msg string, fields ...interface{}) {
	zapFields := convertToZapFields(fields)
	l.logger.Warn(msg, zapFields...)
}

func (l *defaultLogger) Error(msg string, fields ...interface{}) {
	zapFields := convertToZapFields(fields)
	l.logger.Error(msg, zapFields...)
}

func (l *defaultLogger) Fatal(msg string, fields ...interface{}) {
	zapFields := convertToZapFields(fields)
	l.logger.Fatal(msg, zapFields...)
}

func (l *defaultLogger) WithContext(ctx context.Context) contract.Log {
	return l
}

func (l *defaultLogger) WithFields(fields map[string]interface{}) contract.Log {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}

	return &defaultLogger{logger: l.logger.With(zapFields...)}
}

func (l *defaultLogger) WithField(key string, value interface{}) contract.Log {
	return &defaultLogger{logger: l.logger.With(zap.Any(key, value))}
}

func (l *defaultLogger) Named(name string) contract.Log {
	return &defaultLogger{logger: l.logger.Named(name)}
}

func (l *defaultLogger) With(fields ...zap.Field) contract.Log {
	return &defaultLogger{logger: l.logger.With(fields...)}
}

func (l *defaultLogger) Zap() *zap.Logger {
	return l.logger
}

func (l *defaultLogger) SetLevel(level string) error {
	// This is a simple implementation that doesn't actually change the level
	// since zap loggers have immutable configurations
	return nil
}

func (l *defaultLogger) GetLevel() string {
	return "info"
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

// Provider provides an App instance
func Provider() contract.App {
	return New()
}
