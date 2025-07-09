package goe

import (
	"context"
	"os"
	"sync"

	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/app"
	"go.oease.dev/goe/v2/core/cache"
	"go.oease.dev/goe/v2/core/config"
	"go.oease.dev/goe/v2/core/db" // + Import the new db package
	"go.oease.dev/goe/v2/core/http"
	"go.oease.dev/goe/v2/core/log"
	"go.oease.dev/goe/v2/core/observability"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

var (
	// Global instance holder
	instance struct {
		app           contract.Application
		config        contract.Config
		logger        contract.Logger
		http          contract.HTTPKernel
		cacheManager  contract.CacheManager
		db            contract.DB // Database instance
		observability contract.Observability
		mu            sync.RWMutex
	}
)

// Options represents the application options
type Options struct {
	Modules           []contract.Module
	Providers         []any
	Invokers          []any
	WithHTTP          bool // Enable HTTP module
	WithCache         bool // Enable Cache module
	WithDB            bool // + Enable DB module
	WithObservability bool // Enable Observability module
}

// New creates a new Goe application
func New(opts ...Options) contract.Application {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	// Merge options
	opt := Options{}

	if len(opts) > 0 {
		o := opts[0]
		opt.Modules = o.Modules
		opt.Providers = o.Providers
		opt.Invokers = o.Invokers
		opt.WithHTTP = o.WithHTTP
		opt.WithCache = o.WithCache
		opt.WithDB = o.WithDB // + Assign WithDB
		opt.WithObservability = o.WithObservability
	}

	// Create config first to read application settings
	configModule := config.NewModule()
	instance.config = configModule.Provide()

	// Get application settings from config
	appName := instance.config.GetString("APP_NAME")
	if appName == "" {
		appName = "Goe Application"
	}

	appVersion := instance.config.GetString("APP_VERSION")
	if appVersion == "" {
		appVersion = "1.0.0"
	}

	// Get environment from GOE_ENV
	environment := os.Getenv("GOE_ENV")
	if environment == "" {
		environment = "dev"
	}

	// Create application
	instance.app = app.New(appName, appVersion, environment)

	// Create log module
	logModule := log.NewModule(instance.config)
	instance.logger = logModule.Provide()

	// Build Fx options
	fxOptions := []fx.Option{
		fx.StartTimeout(1 * time.Minute),
		// Configure Fx to use our custom logger that logs at debug level
		fx.WithLogger(func() fxevent.Logger {
			zapLogger := logModule.ProvideZap()
			return log.NewFxDebugLogger(zapLogger)
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

	instance.logger.Info("WithHTTP flag", "enabled", opt.WithHTTP)
	instance.logger.Info("WithCache flag", "enabled", opt.WithCache)
	instance.logger.Info("WithDB flag", "enabled", opt.WithDB)
	instance.logger.Info("WithObservability flag", "enabled", opt.WithObservability)

	// Add Cache module if enabled
	var cacheModule *cache.Module
	if opt.WithCache {
		cacheModule = cache.NewModule(instance.config, instance.logger)
		instance.cacheManager = cacheModule.Provide()

		instance.logger.Info("Registering Cache module")

		fxOptions = append(fxOptions,
			fx.Provide(func() contract.CacheManager { return instance.cacheManager }),
			fx.Module(cacheModule.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: cacheModule.OnStart,
						OnStop:  cacheModule.OnStop,
					})
				}),
			),
		)
	}

	// Add DB module if enabled
	var dbModule *db.DatabaseModule
	if opt.WithDB {
		dbModule = db.NewDBModule(instance.config, instance.logger) // Pass config and logger
		instance.db = dbModule.Provide()                            // Store the contract.DB instance

		instance.logger.Info("Registering DB module")

		fxOptions = append(fxOptions,
			fx.Provide(func() contract.DB { return instance.db }),
			// Register DB module with its lifecycle hooks
			fx.Module(dbModule.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: dbModule.OnStart,
						OnStop:  dbModule.OnStop,
					})
				}),
			),
		)
	}

	// Add Observability module if enabled
	var observabilityModule *observability.Module
	if opt.WithObservability {
		observabilityModule = observability.NewModule(instance.config, instance.logger)
		instance.observability = observabilityModule.Provide()

		instance.logger.Info("Registering Observability module")

		fxOptions = append(fxOptions,
			fx.Provide(func() contract.Observability { return instance.observability }),
			fx.Provide(func() contract.MetricsManager { return observabilityModule.ProvideMetrics() }),
			fx.Provide(func() contract.TracingManager { return observabilityModule.ProvideTracing() }),
			fx.Module(observabilityModule.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: observabilityModule.OnStart,
						OnStop:  observabilityModule.OnStop,
					})
				}),
			),
		)
	}

	// Add HTTP module if enabled
	var httpModule *http.Module
	if opt.WithHTTP {
		httpModule = http.NewModule(instance.config, instance.logger)
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

	// Add conditional cache provider with metrics support
	if opt.WithCache {
		if opt.WithObservability {
			// Provide cache with metrics when both cache and observability are enabled
			fxOptions = append(fxOptions, fx.Provide(cache.ProvideCacheWithMetrics))
		} else {
			// Provide plain cache when only cache is enabled
			fxOptions = append(fxOptions, fx.Provide(func() contract.Cache { return cacheModule.ProvideCache() }))
		}
	}

	// Add conditional DB provider with metrics support
	if opt.WithDB && opt.WithObservability {
		// Replace the plain DB provider with metrics-wrapped version when both DB and observability are enabled
		fxOptions = append(fxOptions, fx.Decorate(db.ProvideDBWithMetrics))
	}

	// Add custom providers
	for _, provider := range opt.Providers {
		fxOptions = append(fxOptions, fx.Provide(provider))
	}

	// Add HTTP service injection BEFORE custom invokers to ensure middleware is applied first
	if opt.WithHTTP {
		fxOptions = append(fxOptions, fx.Invoke(func(provider http.ServiceProvider) {
			// Set up service middleware with all available services (including observability if enabled)
			httpModule.Provide().App().Use(http.CreateServiceMiddleware(provider))
		}))
	}

	// Add custom invokers (which may register routes)
	for _, invoker := range opt.Invokers {
		fxOptions = append(fxOptions, fx.Invoke(invoker))
	}

	// Register all options with the application
	if err := instance.app.Register(fxOptions...); err != nil {
		instance.logger.Fatal("Failed to create application", "error", err)
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

// Cache returns the global cache manager instance
func Cache() contract.CacheManager {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.cacheManager == nil {
		panic("Cache module not initialized. Set WithCache: true in goe.New() options")
	}

	return instance.cacheManager
}

// DB returns the global DB instance
func DB() contract.DB {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.db == nil {
		panic("DB module not initialized. Set WithDB: true in goe.New() options, and ensure DB connection is configured.")
	}

	return instance.db
}

// Observability returns the global observability instance
func Observability() contract.Observability {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.observability == nil {
		panic("Observability module not initialized. Set WithObservability: true in goe.New() options")
	}

	return instance.observability
}

// Metrics returns the global metrics manager instance
func Metrics() contract.MetricsManager {
	return Observability().Metrics()
}

// Tracing returns the global tracing manager instance
func Tracing() contract.TracingManager {
	return Observability().Tracing()
}

// AddModule adds a module to the global application instance
func AddModule(module contract.Module) error {
	instance.mu.RLock()
	app := instance.app
	instance.mu.RUnlock()

	if app == nil {
		panic("Application not initialized. Call goe.New() first")
	}

	return app.AddModule(module)
}

// AddProvider adds a provider to the global application instance
func AddProvider(provider contract.Provider) error {
	instance.mu.RLock()
	app := instance.app
	instance.mu.RUnlock()

	if app == nil {
		panic("Application not initialized. Call goe.New() first")
	}

	return app.AddProvider(provider)
}

// AddInvoker adds an invoker to the global application instance
func AddInvoker(invoker contract.Invoker) error {
	instance.mu.RLock()
	app := instance.app
	instance.mu.RUnlock()

	if app == nil {
		panic("Application not initialized. Call goe.New() first")
	}

	return app.AddInvoker(invoker)
}
