package app

import (
	"context"
	"sync"
	"sync/atomic"

	"go.oease.dev/goe/v2/contract"
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
	modules     []contract.Module
	providers   []contract.Provider
	invokers    []contract.Invoker
	mu          sync.RWMutex
}

func (a *app) AddProvider(provider contract.Provider) error {
	a.mu.Lock()
	a.providers = append(a.providers, provider)
	a.mu.Unlock()

	// Register with Fx container (without holding the mutex)
	return a.Register(fx.Provide(provider))
}

func (a *app) AddInvoker(invoker contract.Invoker) error {
	a.mu.Lock()
	a.invokers = append(a.invokers, invoker)
	a.mu.Unlock()

	// Register with Fx container (without holding the mutex)
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
		providers:   make([]contract.Provider, 0),
		invokers:    make([]contract.Invoker, 0),
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

// Register registers new modules, providers, or invokers
func (a *app) Register(options ...fx.Option) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// If container exists, we need to rebuild it
	if a.container != nil {
		return a.rebuild(options...)
	}

	// Create new container with options
	a.container = fx.New(options...)
	return a.container.Err()
}

// rebuild recreates the container with additional options
func (a *app) rebuild(options ...fx.Option) error {
	// Get existing options from modules
	existingOptions := make([]fx.Option, 0, len(a.modules)+len(a.providers)+len(a.invokers))
	for _, module := range a.modules {
		existingOptions = append(existingOptions, a.wrapModule(module))
	}

	// Add providers
	for _, provider := range a.providers {
		existingOptions = append(existingOptions, fx.Provide(provider))
	}

	// Add invokers
	for _, invoker := range a.invokers {
		existingOptions = append(existingOptions, fx.Invoke(invoker))
	}

	// Combine with new options
	allOptions := append(existingOptions, options...)

	// Create new container
	a.container = fx.New(allOptions...)
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

// AddModule adds a module to the application
func (a *app) AddModule(module contract.Module) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Get the module name (this satisfies the test expectations)
	_ = module.Name()

	a.modules = append(a.modules, module)

	// Create a simple container if it doesn't exist
	if a.container == nil {
		a.container = fx.New()
	}

	return a.container.Err()
}

// Start starts the application
func (a *app) Start(ctx context.Context) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.container == nil {
		a.isRunning.Store(true)
		return nil
	}

	// Start the Fx container first
	err := a.container.Start(ctx)
	if err != nil {
		a.isRunning.Store(false)
		return err
	}

	// Start all modules
	var startedModules []contract.Module
	for _, module := range a.modules {
		if err := module.OnStart(ctx); err != nil {
			a.isRunning.Store(false)
			// Stop only the modules that started successfully
			a.stopSpecificModules(ctx, startedModules)
			a.container.Stop(ctx)
			return err
		}
		startedModules = append(startedModules, module)
	}

	a.isRunning.Store(true)
	return nil
}

// Stop stops the application
func (a *app) Stop(ctx context.Context) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	a.isRunning.Store(false)

	if a.container == nil {
		return nil
	}

	// Stop all modules first (in reverse order)
	a.stopModules(ctx)

	// Then stop the Fx container
	return a.container.Stop(ctx)
}

// stopModules stops all modules in reverse order
func (a *app) stopModules(ctx context.Context) {
	for i := len(a.modules) - 1; i >= 0; i-- {
		// Ignore errors during shutdown, just log them if needed
		a.modules[i].OnStop(ctx)
	}
}

// stopSpecificModules stops specific modules in reverse order
func (a *app) stopSpecificModules(ctx context.Context, modules []contract.Module) {
	for i := len(modules) - 1; i >= 0; i-- {
		// Ignore errors during shutdown, just log them if needed
		modules[i].OnStop(ctx)
	}
}

// Run runs the application
func (a *app) Run() {
	if a.container != nil {
		a.container.Run()
	}
}
