package goe

import (
	"context"
	"os"
	"reflect"
	"sync"
	"time"

	"go.oease.dev/goe/v2/core/mongodb"

	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/app"
	"go.oease.dev/goe/v2/core/cache"
	"go.oease.dev/goe/v2/core/config"
	"go.oease.dev/goe/v2/core/db"
	"go.oease.dev/goe/v2/core/event"
	"go.oease.dev/goe/v2/core/health"
	"go.oease.dev/goe/v2/core/http"
	"go.oease.dev/goe/v2/core/job"
	"go.oease.dev/goe/v2/core/lock"
	"go.oease.dev/goe/v2/core/log"
	"go.oease.dev/goe/v2/core/metrics"
	"go.oease.dev/goe/v2/core/otel"
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
		eventManager    contract.EventManager   // Event manager instance
		lockManager     contract.LockManager    // Lock manager instance
		jobManager      contract.JobManager     // Job manager instance
		healthManager   contract.HealthManager  // Health manager instance
		metricsManager  contract.MetricsManager // Metrics manager instance
		otelProvider    contract.OTelProvider   // OpenTelemetry provider instance
		shutdownManager *shutdown.Manager       // Shutdown manager instance
		mu              sync.RWMutex
	}
)

// Options represents the application options
type Options struct {
	Modules         []any // Module constructors (functions that return contract.Module)
	Providers       []any
	Invokers        []any
	WithHTTP        bool           // Enable HTTP module
	WithCache       bool           // Enable Cache module
	WithDB          bool           // Enable DB module
	WithMongoDB     bool           // Enable Mongo DB module
	WithEvent       bool           // Enable Event module
	WithLock        bool           // Enable Lock module (distributed mutex)
	WithJob         bool           // Enable Job module (background job processing)
	WithHealth      bool           // Enable Health module (health checks)
	WithMetrics     bool           // Enable Metrics module (Prometheus metrics)
	WithOTel        bool           // Enable OpenTelemetry module (distributed tracing)
	HTTPPort        int            // Override HTTP port (overrides HTTP_PORT env var)
	ConfigOverrides map[string]any // Override any environment variables

	// Shutdown configuration
	ShutdownTimeout time.Duration // Total shutdown timeout (default: 30s)
	DrainTimeout    time.Duration // HTTP drain timeout (default: 5s)

	// Lifecycle hooks - these are executed through Fx's lifecycle system
	OnStart []func(context.Context) error // Functions to run after all modules start
	OnStop  []func(context.Context) error // Functions to run before modules stop
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
		opt.WithDB = o.WithDB
		opt.WithMongoDB = o.WithMongoDB
		opt.WithEvent = o.WithEvent
		opt.WithLock = o.WithLock
		opt.WithJob = o.WithJob
		opt.WithHealth = o.WithHealth
		opt.WithMetrics = o.WithMetrics
		opt.WithOTel = o.WithOTel
		opt.HTTPPort = o.HTTPPort
		opt.ConfigOverrides = o.ConfigOverrides
		opt.ShutdownTimeout = o.ShutdownTimeout
		opt.DrainTimeout = o.DrainTimeout
		opt.OnStart = o.OnStart
		opt.OnStop = o.OnStop
	}

	// Create config first to read application settings
	configModule := config.NewModule()
	baseConfig := configModule.Provide()

	// Create configuration with overrides
	configOverrides := make(map[string]any)
	if opt.ConfigOverrides != nil {
		for key, value := range opt.ConfigOverrides {
			configOverrides[key] = value
		}
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

	instance.logger.Info("WithHTTP flag", "enabled", opt.WithHTTP)
	instance.logger.Info("WithCache flag", "enabled", opt.WithCache)
	instance.logger.Info("WithDB flag", "enabled", opt.WithDB)
	instance.logger.Info("WithEvent flag", "enabled", opt.WithEvent)
	instance.logger.Info("WithMongoDB flag", "enabled", opt.WithMongoDB)
	instance.logger.Info("WithLock flag", "enabled", opt.WithLock)
	instance.logger.Info("WithJob flag", "enabled", opt.WithJob)
	instance.logger.Info("WithHealth flag", "enabled", opt.WithHealth)
	instance.logger.Info("WithMetrics flag", "enabled", opt.WithMetrics)
	instance.logger.Info("WithOTel flag", "enabled", opt.WithOTel)

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

	// Add Event module if enabled
	var eventModule *event.Module
	if opt.WithEvent {
		var err error
		eventModule, err = event.NewModule(instance.config, instance.logger)
		if err != nil {
			instance.logger.Fatal("Failed to create event module", "error", err)
		}
		instance.eventManager = eventModule.Provide()

		instance.logger.Info("Registering Event module")

		fxOptions = append(fxOptions,
			fx.Provide(func() contract.EventManager { return instance.eventManager }),
			fx.Provide(func() contract.EventPublisher { return eventModule.ProvideEventPublisher() }),
			fx.Provide(func() contract.EventConsumer { return eventModule.ProvideEventConsumer() }),
			fx.Provide(func() contract.DeadLetterQueueManager { return eventModule.ProvideDeadLetterQueue() }),
			fx.Module(eventModule.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: eventModule.OnStart,
						OnStop:  eventModule.OnStop,
					})
				}),
			),
		)
	}

	// Add MongoDB module if enabled
	var mongodbModule *mongodb.DatabaseModule
	if opt.WithMongoDB {
		mongodbModule = mongodb.NewDBModule(instance.config, instance.logger)
		instance.mongoDB = mongodbModule.Provide()

		instance.logger.Info("Registering MongoDB module")

		fxOptions = append(fxOptions,
			// Provide contract.MongoDB for dependency injection
			fx.Provide(func() contract.MongoDB { return instance.mongoDB }),
			// Register MongoDB module with its lifecycle hooks
			fx.Module(mongodbModule.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: mongodbModule.OnStart,
						OnStop:  mongodbModule.OnStop,
					})
				}),
			),
		)
	}

	// Add Lock module if enabled
	var lockModule *lock.Module
	if opt.WithLock {
		var err error
		lockModule, err = lock.NewModule(instance.config, instance.logger)
		if err != nil {
			instance.logger.Fatal("Failed to create lock module", "error", err)
		}
		instance.lockManager = lockModule.Provide()

		instance.logger.Info("Registering Lock module")

		fxOptions = append(fxOptions,
			fx.Provide(func() contract.LockManager { return instance.lockManager }),
			fx.Module(lockModule.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: lockModule.OnStart,
						OnStop:  lockModule.OnStop,
					})
				}),
			),
		)
	}

	// Add Job module if enabled
	var jobModule *job.Module
	if opt.WithJob {
		var err error
		jobModule, err = job.NewModule(instance.config, instance.logger)
		if err != nil {
			instance.logger.Fatal("Failed to create job module", "error", err)
		}
		instance.jobManager = jobModule.Provide()

		instance.logger.Info("Registering Job module")

		fxOptions = append(fxOptions,
			fx.Provide(func() contract.JobManager { return instance.jobManager }),
			fx.Module(jobModule.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: jobModule.OnStart,
						OnStop:  jobModule.OnStop,
					})
				}),
			),
		)
	}

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
	fxOptions = append(fxOptions,
		fx.Provide(func() *shutdown.Manager { return instance.shutdownManager }),
	)

	// Add Health module if enabled
	var healthModule *health.Module
	if opt.WithHealth {
		healthModule = health.NewModule(instance.config, instance.logger)
		instance.healthManager = healthModule.Manager()

		instance.logger.Info("Registering Health module")

		fxOptions = append(fxOptions,
			fx.Provide(func() contract.HealthManager { return instance.healthManager }),
			fx.Module(healthModule.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: healthModule.OnStart,
						OnStop:  healthModule.OnStop,
					})
				}),
			),
		)
	}

	// Add Metrics module if enabled
	var metricsModule *metrics.Module
	if opt.WithMetrics {
		metricsModule = metrics.NewModule(instance.config, instance.logger)
		instance.metricsManager = metricsModule.Manager()

		instance.logger.Info("Registering Metrics module")

		fxOptions = append(fxOptions,
			fx.Provide(func() contract.MetricsManager { return instance.metricsManager }),
			fx.Module(metricsModule.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: metricsModule.OnStart,
						OnStop:  metricsModule.OnStop,
					})
				}),
			),
		)
	}

	// Add OpenTelemetry module if enabled
	var otelModule *otel.Module
	if opt.WithOTel {
		var err error
		otelModule, err = otel.NewModule(context.Background(), instance.config, instance.logger)
		if err != nil {
			instance.logger.Fatal("Failed to create OpenTelemetry module", "error", err)
		}
		instance.otelProvider = otelModule.Provider()

		instance.logger.Info("Registering OpenTelemetry module")

		fxOptions = append(fxOptions,
			fx.Provide(func() contract.OTelProvider { return instance.otelProvider }),
			fx.Module(otelModule.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: otelModule.OnStart,
						OnStop:  otelModule.OnStop,
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

	// Add custom modules - Handle them exactly like built-in modules
	for _, moduleConstructor := range opt.Modules {
		constructor := moduleConstructor // Capture loop variable

		// Create module instance directly with dependencies (like built-in modules)
		var module contract.Module
		var moduleProviders []fx.Option

		// Handle different constructor signatures
		switch cons := constructor.(type) {
		case func(contract.Logger, contract.Config) contract.Module:
			module = cons(instance.logger, instance.config)
		case func(contract.Config, contract.Logger) contract.Module:
			module = cons(instance.config, instance.logger)
		default:
			instance.logger.Fatal("Invalid module constructor signature",
				"expected", "func(contract.Logger, contract.Config) contract.Module",
				"or", "func(contract.Config, contract.Logger) contract.Module")
		}

		instance.logger.Info("Registering custom module", "name", module.Name())

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
		fxOptions = append(fxOptions, fx.Provide(func() contract.Cache { return cacheModule.ProvideCache() }))
	}

	// Add custom providers
	for _, provider := range opt.Providers {
		fxOptions = append(fxOptions, fx.Provide(provider))
	}

	// Add HTTP service injection and module routes BEFORE the HTTP server starts
	if opt.WithHTTP {
		fxOptions = append(fxOptions, fx.Invoke(func(provider http.ServiceProvider) {
			fiberApp := httpModule.Provide().App()

			// Set up service middleware immediately when all dependencies are available
			// This ensures the middleware is registered before the HTTP server starts listening
			fiberApp.Use(http.CreateServiceMiddleware(provider))
			instance.logger.Debug("HTTP service middleware registered")

			// Register OpenTelemetry tracing middleware (first, to capture all requests)
			if opt.WithOTel && otelModule != nil {
				otelModule.RegisterMiddleware(fiberApp)
				instance.logger.Debug("OpenTelemetry middleware registered")
			}

			// Register Prometheus metrics middleware
			if opt.WithMetrics && metricsModule != nil {
				metricsModule.RegisterMiddleware(fiberApp)
				instance.logger.Debug("Metrics middleware registered")
			}

			// Register health check routes
			if opt.WithHealth && healthModule != nil {
				healthModule.RegisterRoutes(fiberApp)
				instance.logger.Debug("Health routes registered")

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
				if opt.WithEvent && instance.eventManager != nil {
					instance.healthManager.RegisterChecker(health.NewEventChecker(instance.eventManager))
				}
				if opt.WithJob && instance.jobManager != nil {
					instance.healthManager.RegisterChecker(health.NewJobChecker(instance.jobManager))
				}
			}

			// Register metrics endpoint
			if opt.WithMetrics && metricsModule != nil {
				metricsModule.RegisterRoutes(fiberApp)
				instance.logger.Debug("Metrics routes registered")
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
							instance.logger.Error("OnStart hook failed", "index", i, "error", err)
							return err
						}
					}
					return nil
				},
				OnStop: func(ctx context.Context) error {
					// Execute all OnStop hooks in reverse order
					for i := len(opt.OnStop) - 1; i >= 0; i-- {
						if err := opt.OnStop[i](ctx); err != nil {
							instance.logger.Error("OnStop hook failed", "index", i, "error", err)
							// Continue with other hooks even if one fails
						}
					}
					return nil
				},
			})
		}))
	}

	// Register all options with the application
	if err := instance.app.Register(fxOptions...); err != nil {
		instance.logger.Fatal("Failed to create application", "error", err)
	}

	return instance.app
}

// checkAndProvideServices inspects a module for common service provider methods
// and automatically registers them with Fx DI (like built-in modules do)
func checkAndProvideServices(module contract.Module) []fx.Option {
	var providers []fx.Option
	moduleValue := reflect.ValueOf(module)

	// Common service provider method patterns used by GOE modules
	serviceProviderMethods := []string{
		"Provide",               // Generic service provider
		"ProvideService",        // Generic service provider
		"ProvideClient",         // For client modules (like gRPC, HTTP clients)
		"ProvideManager",        // For manager services
		"ProvideHandler",        // For handler services
		"ProvideRepository",     // For data access modules
		"ProvideCache",          // For cache services
		"ProvideDB",             // For database services
		"ProvideLogger",         // For logger services
		"ProvideConfig",         // For config services
		"ProvideEventPublisher", // For event services
		"ProvideEventConsumer",  // For event services
	}

	// Check each potential service provider method
	for _, methodName := range serviceProviderMethods {
		if method := moduleValue.MethodByName(methodName); method.IsValid() {
			methodType := method.Type()

			// Method should have no parameters and return one value (the service)
			if methodType.NumIn() == 0 && methodType.NumOut() == 1 {
				// Create a provider function with the correct return type
				returnType := methodType.Out(0)

				// Create a function with the correct signature using reflection
				providerFunc := reflect.MakeFunc(
					reflect.FuncOf([]reflect.Type{}, []reflect.Type{returnType}, false),
					func(args []reflect.Value) []reflect.Value {
						return method.Call([]reflect.Value{})
					},
				)

				providers = append(providers, fx.Provide(providerFunc.Interface()))
			}
		}
	}

	return providers
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

// EventManager returns the global event manager instance
func EventManager() contract.EventManager {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.eventManager == nil {
		panic("Event module not initialized. Set WithEvent: true in goe.New() options")
	}

	return instance.eventManager
}

// EventPublisher returns the global event publisher instance
func EventPublisher() contract.EventPublisher {
	return EventManager()
}

// EventConsumer returns the global event consumer instance
func EventConsumer() contract.EventConsumer {
	return EventManager()
}

// DeadLetterQueue returns the global dead letter queue manager instance
func DeadLetterQueue() contract.DeadLetterQueueManager {
	return EventManager().GetDeadLetterQueue()
}

// MongoDB returns the global MongoDB instance
func MongoDB() contract.MongoDB {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.mongoDB == nil {
		panic("MongoDB module not initialized. Set WithMongoDB: true in goe.New() options, and ensure MongoDB connection is configured.")
	}

	return instance.mongoDB
}

// Mongo is a convenient alias for MongoDB() for shorter access
func Mongo() contract.MongoDB {
	return MongoDB()
}

// Health returns the global health manager instance
func Health() contract.HealthManager {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.healthManager == nil {
		panic("Health module not initialized. Set WithHealth: true in goe.New() options")
	}

	return instance.healthManager
}

// Metrics returns the global metrics manager instance
func Metrics() contract.MetricsManager {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.metricsManager == nil {
		panic("Metrics module not initialized. Set WithMetrics: true in goe.New() options")
	}

	return instance.metricsManager
}

// OTel returns the global OpenTelemetry provider instance
func OTel() contract.OTelProvider {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.otelProvider == nil {
		panic("OpenTelemetry module not initialized. Set WithOTel: true in goe.New() options")
	}

	return instance.otelProvider
}

// Lock returns the global lock manager instance.
// Use this to create distributed mutex locks for coordinating access to
// shared resources across multiple processes or machines.
//
// Example:
//
//	mutex := goe.Lock().NewMutex("my-resource")
//	if err := mutex.Lock(ctx); err != nil {
//	    return err
//	}
//	defer mutex.Unlock(ctx)
//	// ... critical section ...
func Lock() contract.LockManager {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.lockManager == nil {
		panic("Lock module not initialized. Set WithLock: true in goe.New() options")
	}

	return instance.lockManager
}

// Job returns the global job manager instance.
// Use this to dispatch background jobs, register handlers, and manage job schedules.
//
// Example - Dispatching a job:
//
//	job := contract.NewJobDefinition("send-email", map[string]string{
//	    "to": "user@example.com",
//	    "subject": "Welcome!",
//	})
//	jobID, err := goe.Job().Dispatch(ctx, job)
//
// Example - Registering a handler:
//
//	goe.Job().RegisterHandler("send-email", contract.JobHandlerFunc(func(ctx context.Context, job contract.Job) error {
//	    payload := job.Payload().(map[string]string)
//	    // Send the email...
//	    return nil
//	}))
//
// Example - Scheduling a recurring job:
//
//	goe.Job().RegisterSchedule(&contract.ScheduledJob{
//	    Name:     "cleanup-expired",
//	    Schedule: job.Daily(),
//	    Handler:  myCleanupHandler,
//	})
func Job() contract.JobManager {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.jobManager == nil {
		panic("Job module not initialized. Set WithJob: true in goe.New() options")
	}

	return instance.jobManager
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
