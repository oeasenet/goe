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
	mu          sync.RWMutex
}

// New creates a new application instance
func New(name, version, environment string) contract.Application {
	return &app{
		name:        name,
		version:     version,
		environment: environment,
		ctx:         context.Background(),
		modules:     make([]contract.Module, 0),
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
	existingOptions := make([]fx.Option, 0, len(a.modules))
	for _, module := range a.modules {
		existingOptions = append(existingOptions, a.wrapModule(module))
	}

	// Combine with new options
	allOptions := append(existingOptions, options...)

	// Create new container
	a.container = fx.New(allOptions...)
	return a.container.Err()
}

// wrapModule wraps a module into Fx options
func (a *app) wrapModule(module contract.Module) fx.Option {
	return fx.Module(
		module.Name(),
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return module.OnStart(ctx)
				},
				OnStop: func(ctx context.Context) error {
					return module.OnStop(ctx)
				},
			})
		}),
	)
}

// AddModule adds a module to the application
func (a *app) AddModule(module contract.Module) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.modules = append(a.modules, module)
	return a.Register(a.wrapModule(module))
}

// Start starts the application
func (a *app) Start(ctx context.Context) error {
	if a.container == nil {
		return nil
	}

	a.isRunning.Store(true)
	return a.container.Start(ctx)
}

// Stop stops the application
func (a *app) Stop(ctx context.Context) error {
	if a.container == nil {
		return nil
	}

	a.isRunning.Store(false)
	return a.container.Stop(ctx)
}

// Run runs the application
func (a *app) Run() {
	if a.container != nil {
		a.container.Run()
	}
}
