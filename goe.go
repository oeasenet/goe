package goe

import (
	"context"
	"maps"
	"os"
	"sync"
	"time"

	"go.oease.dev/goe/v2/core/mongodb/migrate"

	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/app"
	"go.oease.dev/goe/v2/core/config"
	"go.oease.dev/goe/v2/core/health"
	"go.oease.dev/goe/v2/core/http"
	"go.oease.dev/goe/v2/core/log"
	"go.oease.dev/goe/v2/core/shutdown"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

var (
	// Global instance holder
	instance struct {
		app             contract.Application
		config          contract.Config
		logger          contract.Logger
		http            contract.HTTPKernel
		cacheManager    contract.CacheManager
		db              contract.DB             // Database instance
		mongoDB         contract.MongoDB        // MongoDB instance
		migrator        *migrate.Migrator       // MongoDB migration instance
		lockManager     contract.LockManager    // Lock manager instance
		jobManager      contract.JobManager     // Job manager instance
		healthManager   contract.HealthManager  // Health manager instance
		metricsManager  contract.MetricsManager // Metrics manager instance
		otelProvider    contract.OTelProvider   // OpenTelemetry provider instance
		shutdownManager *shutdown.Manager       // Shutdown manager instance
		mu              sync.RWMutex
	}
)

// New creates a new Goe application.
//
// The global goe.* accessors (Log, Config, DB, ...) are usable from the user
// code New runs — custom module constructors, providers, and invokers — not
// just after New returns. Accessors for modules that are not enabled panic
// with a descriptive message.
func New(opts ...Options) contract.Application {
	// Merge options
	opt := Options{}

	if len(opts) > 0 {
		o := opts[0]
		opt.Modules = o.Modules
		opt.Providers = o.Providers
		opt.Invokers = o.Invokers
		opt.WithHTTP = o.WithHTTP
		opt.WithCache = o.WithCache
		opt.WithDB = o.WithDB
		opt.WithMongoDB = o.WithMongoDB
		opt.WithMigrate = o.WithMigrate
		opt.WithLock = o.WithLock
		opt.WithJob = o.WithJob
		opt.WithHealth = o.WithHealth
		opt.WithMetrics = o.WithMetrics
		opt.WithOTel = o.WithOTel
		opt.HTTP = o.HTTP
		opt.HTTPPort = o.HTTPPort
		opt.ConfigOverrides = o.ConfigOverrides
		opt.ShutdownTimeout = o.ShutdownTimeout
		opt.DrainTimeout = o.DrainTimeout
		opt.OnStart = o.OnStart
		opt.OnStop = o.OnStop
	}

	// Supplying HTTP options is an unambiguous request for the HTTP module, so
	// enabling it separately would only be a way to get it wrong.
	if opt.HTTP != nil {
		opt.WithHTTP = true
	}

	// Assemble config, logger, app, and the built-in modules under the
	// instance lock, then release it. Everything after this call may execute
	// user code (module constructors, providers, invokers), and user code may
	// call the goe.* accessors: with the write lock still held, each of those
	// calls would block on instance.mu forever — a silent freeze with zero
	// diagnostics, because fx runs the whole graph inside app.Register below.
	application, cfg, logger, reg := buildCore(opt)

	fxOptions := reg.fxOptions

	// Add custom modules - Handle them exactly like built-in modules
	for _, moduleConstructor := range opt.Modules {
		constructor := moduleConstructor // Capture loop variable

		// Create module instance directly with dependencies (like built-in modules)
		var module contract.Module
		var moduleProviders []fx.Option

		// Handle different constructor signatures
		switch cons := constructor.(type) {
		case func(contract.Logger, contract.Config) contract.Module:
			module = cons(logger, cfg)
		case func(contract.Config, contract.Logger) contract.Module:
			module = cons(cfg, logger)
		default:
			logger.Fatal("Invalid module constructor signature",
				"expected", "func(contract.Logger, contract.Config) contract.Module",
				"or", "func(contract.Config, contract.Logger) contract.Module")
		}

		logger.Debug("Registering custom module", "name", module.Name())

		// If module provides services, register them to DI (like built-in modules)
		if provider, ok := module.(interface{ ProvideServices() []fx.Option }); ok {
			// Module can provide multiple services
			moduleProviders = append(moduleProviders, provider.ProvideServices()...)
		} else {
			// Check for common service provider methods
			moduleProviders = append(moduleProviders, checkAndProvideServices(module)...)
		}

		// Register module exactly like built-in modules
		allOptions := append(moduleProviders, fx.Module(module.Name(),
			fx.Invoke(func(lc fx.Lifecycle) {
				lc.Append(fx.Hook{
					OnStart: module.OnStart,
					OnStop:  module.OnStop,
				})
			}),
		))

		fxOptions = append(fxOptions, allOptions...)
	}

	// Add cache provider
	if opt.WithCache {
		fxOptions = append(fxOptions, fx.Provide(func() contract.Cache { return reg.cache.ProvideCache() }))
	}

	// Add custom providers
	for _, provider := range opt.Providers {
		fxOptions = append(fxOptions, fx.Provide(provider))
	}

	// Add HTTP service injection and module routes BEFORE the HTTP server starts
	if opt.WithHTTP {
		fxOptions = append(fxOptions, fx.Invoke(func(provider http.ServiceProvider) {
			fiberApp := reg.http.Provide().App()

			// Set up service middleware immediately when all dependencies are available
			// This ensures the middleware is registered before the HTTP server starts listening
			fiberApp.Use(http.CreateServiceMiddleware(provider))
			logger.Debug("HTTP service middleware registered")

			// Register OpenTelemetry tracing middleware (first, to capture all requests)
			if opt.WithOTel && reg.otel != nil {
				reg.otel.RegisterMiddleware(fiberApp)
				logger.Debug("OpenTelemetry middleware registered")
			}

			// Register Prometheus metrics middleware
			if opt.WithMetrics && reg.metrics != nil {
				reg.metrics.RegisterMiddleware(fiberApp)
				logger.Debug("Metrics middleware registered")
			}

			// Register health check routes
			if opt.WithHealth && reg.health != nil {
				reg.health.RegisterRoutes(fiberApp)
				logger.Debug("Health routes registered")

				// Auto-register health checkers for enabled modules
				if opt.WithDB && instance.db != nil {
					instance.healthManager.RegisterChecker(health.NewDatabaseChecker(instance.db.Instance()))
				}
				if opt.WithCache && instance.cacheManager != nil {
					instance.healthManager.RegisterChecker(health.NewCacheChecker(instance.cacheManager.Store()))
				}
				if opt.WithMongoDB && instance.mongoDB != nil {
					instance.healthManager.RegisterChecker(health.NewMongoDBChecker(instance.mongoDB))
				}
				if opt.WithJob && instance.jobManager != nil {
					instance.healthManager.RegisterChecker(health.NewJobChecker(instance.jobManager))
				}
			}

			// Register metrics endpoint
			if opt.WithMetrics && reg.metrics != nil {
				reg.metrics.RegisterRoutes(fiberApp)
				logger.Debug("Metrics routes registered")
			}
		}))
	}

	// Add custom invokers (which may register routes)
	for _, invoker := range opt.Invokers {
		fxOptions = append(fxOptions, fx.Invoke(invoker))
	}

	// Add lifecycle hooks
	if len(opt.OnStart) > 0 || len(opt.OnStop) > 0 {
		fxOptions = append(fxOptions, fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// Execute all OnStart hooks
					for i, hook := range opt.OnStart {
						if err := hook(ctx); err != nil {
							logger.Error("OnStart hook failed", "index", i, "error", err)
							return err
						}
					}
					return nil
				},
				OnStop: func(ctx context.Context) error {
					// Execute all OnStop hooks in reverse order
					for i := len(opt.OnStop) - 1; i >= 0; i-- {
						if err := opt.OnStop[i](ctx); err != nil {
							logger.Error("OnStop hook failed", "index", i, "error", err)
							// Continue with other hooks even if one fails
						}
					}
					return nil
				},
			})
		}))
	}

	// Register all options with the application. Register builds the fx graph
	// synchronously, so this is where providers and invokers actually run.
	if err := application.Register(fxOptions...); err != nil {
		logger.Fatal("Failed to create application", "error", err)
	}

	return application
}

// buildCore assembles the config, logger, application, shutdown manager, and
// all enabled built-in modules, holding instance.mu for the duration. It is
// the only phase of New that writes to the instance singleton.
//
// Nothing in here may call user code: the goe.* accessors take an RLock on
// instance.mu, so user code invoked under the write lock would deadlock.
// User-facing work (custom module constructors, providers, invokers) belongs
// in New, after this returns and the lock is released.
func buildCore(opt Options) (contract.Application, contract.Config, contract.Logger, *moduleRegistry) {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	// Create config first to read application settings
	configModule := config.NewModule()
	baseConfig := configModule.Provide()

	// Create configuration with overrides
	configOverrides := make(map[string]any)
	if opt.ConfigOverrides != nil {
		maps.Copy(configOverrides, opt.ConfigOverrides)
	}

	// Add HTTPPort override if specified
	if opt.HTTPPort > 0 {
		configOverrides["HTTP_PORT"] = opt.HTTPPort
	}

	// Wrap config with overrides if any exist
	if len(configOverrides) > 0 {
		instance.config = config.NewConfigWrapper(baseConfig, configOverrides)
	} else {
		instance.config = baseConfig
	}

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
		fx.StartTimeout(2 * time.Minute),
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

	// Which modules are enabled is wiring detail, not an operational event.
	// Ten Info lines on every boot drowned out messages that matter.
	instance.logger.Debug("Module flags",
		"http", opt.WithHTTP, "cache", opt.WithCache, "db", opt.WithDB,
		"mongodb", opt.WithMongoDB, "migrate", opt.WithMigrate, "lock", opt.WithLock,
		"job", opt.WithJob, "health", opt.WithHealth, "metrics", opt.WithMetrics,
		"otel", opt.WithOTel,
	)
	// Register the enabled built-in modules. Each call is a no-op when its
	// Options flag is off, and the order is the order Fx will run their
	// lifecycle hooks in — see moduleRegistry in modules.go.
	reg := &moduleRegistry{opt: opt, fxOptions: fxOptions}
	reg.addCache()
	reg.addDB()
	reg.addMongoDB()
	reg.addMigrate()
	reg.addLock()
	reg.addJob()

	// Create shutdown manager with configured timeouts
	shutdownTimeout := opt.ShutdownTimeout
	if shutdownTimeout == 0 {
		shutdownTimeout = instance.config.GetDuration("SHUTDOWN_TIMEOUT")
		if shutdownTimeout == 0 {
			shutdownTimeout = 30 * time.Second
		}
	}
	drainTimeout := opt.DrainTimeout
	if drainTimeout == 0 {
		drainTimeout = instance.config.GetDuration("SHUTDOWN_DRAIN_TIMEOUT")
		if drainTimeout == 0 {
			drainTimeout = 5 * time.Second
		}
	}
	instance.shutdownManager = shutdown.NewManager(instance.logger, shutdownTimeout, drainTimeout)

	// Always provide shutdown manager for dependency injection
	reg.fxOptions = append(reg.fxOptions,
		fx.Provide(func() *shutdown.Manager { return instance.shutdownManager }),
	)

	reg.addHealth()
	reg.addMetrics()
	reg.addOTel()
	reg.addHTTP()

	return instance.app, instance.config, instance.logger, reg
}
