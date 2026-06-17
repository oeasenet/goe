package app

import (
	"context"
	"sync"
	"sync/atomic"

	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/internal/configvalidator"
	"go.uber.org/fx"
)

// app implements the Application interface
type app struct {
	name        string
	version     string
	environment string
	ctx         context.Context
	container   *fx.App
	isRunning   atomic.Bool
	modules     []contract.Module // tracking for validation
	fxOptions   []fx.Option       // accumulated FX options (single source of truth for lifecycle)
	mu          sync.RWMutex
	config      contract.Config
	logger      contract.Logger
}

func (a *app) AddProvider(provider contract.Provider) error {
	return a.Register(fx.Provide(provider))
}

func (a *app) AddInvoker(invoker contract.Invoker) error {
	return a.Register(fx.Invoke(invoker))
}

// New creates a new application instance
func New(name, version, environment string) contract.Application {
	return &app{
		name:        name,
		version:     version,
		environment: environment,
		ctx:         context.Background(),
		modules:     make([]contract.Module, 0),
		fxOptions:   make([]fx.Option, 0),
	}
}

// Name returns the application name
func (a *app) Name() string {
	return a.name
}

// Version returns the application version
func (a *app) Version() string {
	return a.version
}

// Environment returns the current environment
func (a *app) Environment() string {
	return a.environment
}

// Context returns the application context
func (a *app) Context() context.Context {
	return a.ctx
}

// Container returns the underlying Fx app instance
func (a *app) Container() *fx.App {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.container
}

// IsRunning returns true if the application is running
func (a *app) IsRunning() bool {
	return a.isRunning.Load()
}

// Register registers new modules, providers, or invokers.
// Options are accumulated and the FX container is (re)created from all of them.
func (a *app) Register(options ...fx.Option) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.fxOptions = append(a.fxOptions, options...)
	a.container = fx.New(a.fxOptions...)
	return a.container.Err()
}

// wrapModule wraps a module into Fx options
func (a *app) wrapModule(module contract.Module) fx.Option {
	return fx.Invoke(func(lc fx.Lifecycle) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return module.OnStart(ctx)
			},
			OnStop: func(ctx context.Context) error {
				return module.OnStop(ctx)
			},
		})
	})
}

// AddModule adds a module to the application.
// The module's lifecycle (OnStart/OnStop) is registered with FX immediately,
// so FX is the single owner of all module lifecycle management.
func (a *app) AddModule(module contract.Module) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	_ = module.Name()
	a.modules = append(a.modules, module)

	// Register lifecycle with FX and rebuild the container
	a.fxOptions = append(a.fxOptions, a.wrapModule(module))
	a.container = fx.New(a.fxOptions...)
	return a.container.Err()
}

// SetConfig sets the configuration
func (a *app) SetConfig(config contract.Config) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.config = config
}

// SetLogger sets the logger
func (a *app) SetLogger(logger contract.Logger) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.logger = logger.With("module", "app")
}

// validateModules validates all module configurations using the provided snapshots.
// Callers must pass snapshots taken under a.mu.RLock so this method does not
// need to acquire the lock itself.
func (a *app) validateModules(modules []contract.Module, cfg contract.Config, logger contract.Logger) error {
	// Skip validation if config or logger is not set
	if cfg == nil || logger == nil {
		return nil
	}

	// Create startup validator
	startupValidator := configvalidator.NewStartupValidator(cfg, logger)

	// Register all modules that support validation
	for _, module := range modules {
		if validatableModule, ok := module.(contract.ModuleWithValidator); ok {
			startupValidator.RegisterModule(validatableModule)
		}
	}

	// Validate all modules
	return startupValidator.ValidateAll()
}

// Start starts the application.
// All module lifecycle (OnStart/OnStop) is managed by FX hooks.
func (a *app) Start(ctx context.Context) error {
	a.mu.RLock()
	modules := make([]contract.Module, len(a.modules))
	copy(modules, a.modules)
	container := a.container
	cfg := a.config
	logger := a.logger
	a.mu.RUnlock()

	// Validate module configurations before starting (using snapshots)
	if err := a.validateModules(modules, cfg, logger); err != nil {
		a.isRunning.Store(false)
		return err
	}

	if container == nil {
		a.isRunning.Store(true)
		return nil
	}

	// FX owns all module lifecycle via hooks registered in AddModule/Register.
	// container.Start fires OnStart hooks in registration order and automatically
	// rolls back (calls OnStop) for any already-started hooks if one fails.
	if err := container.Start(ctx); err != nil {
		a.isRunning.Store(false)
		return err
	}

	a.isRunning.Store(true)
	return nil
}

// Stop stops the application.
// FX stops all lifecycle hooks in reverse registration order.
func (a *app) Stop(ctx context.Context) error {
	a.mu.RLock()
	container := a.container
	a.mu.RUnlock()

	a.isRunning.Store(false)

	if container == nil {
		return nil
	}

	return container.Stop(ctx)
}

// Run runs the application
func (a *app) Run() {
	if a.container != nil {
		a.container.Run()
	}
}
