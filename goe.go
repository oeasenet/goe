package goe

import (
	"context"
	"os"
	"sync"

	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/app"
	"go.oease.dev/goe/v2/core/config"
	"go.oease.dev/goe/v2/core/http"
	"go.oease.dev/goe/v2/core/log"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

var (
	// Global instance holder
	instance struct {
		app    contract.Application
		config contract.Config
		logger contract.Logger
		http   contract.HTTPKernel
		mu     sync.RWMutex
	}
)

// Options represents the application options
type Options struct {
	Name        string
	Version     string
	Environment string
	Modules     []contract.Module
	Providers   []any
	Invokers    []any
	WithHTTP    bool // Enable HTTP module
}

// New creates a new Goe application
func New(opts ...Options) contract.Application {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	// Merge options
	opt := Options{
		Name:        "Goe Application",
		Version:     "1.0.0",
		Environment: "dev",
	}

	if len(opts) > 0 {
		o := opts[0]
		if o.Name != "" {
			opt.Name = o.Name
		}
		if o.Version != "" {
			opt.Version = o.Version
		}
		if o.Environment != "" {
			opt.Environment = o.Environment
		}
		opt.Modules = o.Modules
		opt.Providers = o.Providers
		opt.Invokers = o.Invokers
		opt.WithHTTP = o.WithHTTP
	}

	// Get environment from GOE_ENV if not set
	if env := os.Getenv("GOE_ENV"); env != "" {
		opt.Environment = env
	}

	// Create application
	instance.app = app.New(opt.Name, opt.Version, opt.Environment)

	// Create core modules
	configModule := config.NewModule()
	instance.config = configModule.Provide()

	logModule := log.NewModule(instance.config)
	instance.logger = logModule.Provide()

	// Build Fx options
	fxOptions := []fx.Option{
		// Configure Fx to use our custom logger
		fx.WithLogger(func() fxevent.Logger {
			zapLogger := logModule.ProvideZap()
			return &fxevent.ZapLogger{Logger: zapLogger}
		}),

		// Provide core services
		fx.Provide(func() contract.Application { return instance.app }),
		fx.Provide(func() contract.Config { return instance.config }),
		fx.Provide(func() contract.Logger { return instance.logger }),
		fx.Provide(func() *zap.Logger { return logModule.ProvideZap() }),

		// Core modules
		fx.Module(configModule.Name(),
			fx.Invoke(func(lc fx.Lifecycle) {
				lc.Append(fx.Hook{
					OnStart: configModule.OnStart,
					OnStop:  configModule.OnStop,
				})
			}),
		),
		fx.Module(logModule.Name(),
			fx.Invoke(func(lc fx.Lifecycle) {
				lc.Append(fx.Hook{
					OnStart: logModule.OnStart,
					OnStop:  logModule.OnStop,
				})
			}),
		),
	}

	instance.logger.Info("WithHTTP flag", log.NewField("enabled", opt.WithHTTP))

	// Add HTTP module if enabled
	if opt.WithHTTP {
		httpModule := http.NewModule(instance.config, instance.logger)
		instance.http = httpModule.Provide()

		instance.logger.Info("Registering HTTP module")

		fxOptions = append(fxOptions,
			fx.Provide(func() contract.HTTPKernel { return instance.http }),
			fx.Module(httpModule.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: httpModule.OnStart,
						OnStop:  httpModule.OnStop,
					})
				}),
			),
		)
	}

	// Add custom modules
	for _, module := range opt.Modules {
		mod := module // Capture loop variable
		fxOptions = append(fxOptions, fx.Module(mod.Name(),
			fx.Invoke(func(lc fx.Lifecycle) {
				lc.Append(fx.Hook{
					OnStart: mod.OnStart,
					OnStop:  mod.OnStop,
				})
			}),
		))
	}

	// Add custom providers
	for _, provider := range opt.Providers {
		fxOptions = append(fxOptions, fx.Provide(provider))
	}

	// Add HTTP service injection BEFORE custom invokers to ensure middleware is applied first
	if opt.WithHTTP {
		fxOptions = append(fxOptions, fx.Invoke(func(provider http.ServiceProvider, httpKernel contract.HTTPKernel) {
			// Use the injected httpKernel instead of instance.http to avoid timing issues
			httpKernel.App().Use(http.CreateServiceMiddleware(provider))
		}))
	}

	// Add custom invokers (which may register routes)
	for _, invoker := range opt.Invokers {
		fxOptions = append(fxOptions, fx.Invoke(invoker))
	}

	// Register all options with the application
	if err := instance.app.Register(fxOptions...); err != nil {
		instance.logger.Fatal("Failed to create application", log.NewField("error", err))
	}

	return instance.app
}

// Run runs the application
func Run() {
	instance.mu.RLock()
	app := instance.app
	instance.mu.RUnlock()

	if app == nil {
		panic("Application not initialized. Call goe.New() first")
	}

	app.Container().Run()
}

// App returns the global application instance
func App() contract.Application {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.app == nil {
		panic("Application not initialized. Call goe.New() first")
	}

	return instance.app
}

// Config returns the global config instance
func Config() contract.Config {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.config == nil {
		panic("Application not initialized. Call goe.New() first")
	}

	return instance.config
}

// Log returns the global logger instance
func Log() contract.Logger {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.logger == nil {
		panic("Application not initialized. Call goe.New() first")
	}

	return instance.logger
}

// Context returns the application context
func Context() context.Context {
	return App().Context()
}

// IsRunning returns true if the application is running
func IsRunning() bool {
	return App().IsRunning()
}

// GetEnvironment returns the current environment
func GetEnvironment() string {
	return App().Environment()
}

// HTTP returns the global HTTP kernel instance
func HTTP() contract.HTTPKernel {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.http == nil {
		panic("HTTP module not initialized. Set WithHTTP: true in goe.New() options")
	}

	return instance.http
}

// httpAccessor is used internally to access HTTP without locking
func httpAccessor() contract.HTTPKernel {
	return instance.http
}
