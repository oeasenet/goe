package goe

import (
	"context"
	"reflect"

	"go.oease.dev/goe/v2/core/mongodb"
	"go.oease.dev/goe/v2/core/mongodb/migrate"

	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/cache"
	"go.oease.dev/goe/v2/core/db"
	"go.oease.dev/goe/v2/core/health"
	"go.oease.dev/goe/v2/core/http"
	"go.oease.dev/goe/v2/core/job"
	"go.oease.dev/goe/v2/core/lock"
	"go.oease.dev/goe/v2/core/metrics"
	"go.oease.dev/goe/v2/core/otel"
	"go.uber.org/fx"
)

// moduleRegistry accumulates the Fx options for the enabled built-in modules.
//
// It exists so that New reads as a sequence of "enable this module" steps
// instead of one long function. Each method checks its own Options flag and does
// nothing when the module is disabled, which keeps the call site flat.
//
// Registration order is significant: Fx runs lifecycle hooks in the order they
// are appended, so these methods must be called in the same order the single
// function used to build them.
//
// The methods write to the package-level instance singleton, exactly as the
// inlined code did. buildCore holds instance.mu while they run, and they must
// not call user code: an accessor call under the write lock deadlocks.
type moduleRegistry struct {
	opt       Options
	fxOptions []fx.Option

	cache   *cache.Module
	db      *db.DatabaseModule
	mongodb *mongodb.DatabaseModule
	migrate *migrate.Module
	lock    *lock.Module
	job     *job.Module
	health  *health.Module
	metrics *metrics.Module
	otel    *otel.Module
	http    *http.Module
}

func (r *moduleRegistry) addCache() {
	// Add Cache module if enabled

	if r.opt.WithCache {
		r.cache = cache.NewModule(instance.config, instance.logger, r.opt.Cache...)
		// A method value, not a call: the cache is built on first use so
		// driver connections open only after startup validation has passed.
		instance.cacheProvider = r.cache.Provide

		instance.logger.Debug("Registering Cache module")

		r.fxOptions = append(r.fxOptions,
			fx.Module(r.cache.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: r.cache.OnStart,
						OnStop:  r.cache.OnStop,
					})
				}),
			),
		)
	}
}

func (r *moduleRegistry) addDB() {
	// Add DB module if enabled

	if r.opt.WithDB {
		r.db = db.NewDBModule(instance.config, instance.logger) // Pass config and logger
		instance.db = r.db.Provide()                            // Store the contract.DB instance

		instance.logger.Debug("Registering DB module")

		r.fxOptions = append(r.fxOptions,
			fx.Provide(func() contract.DB { return instance.db }),
			// Register DB module with its lifecycle hooks
			fx.Module(r.db.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: r.db.OnStart,
						OnStop:  r.db.OnStop,
					})
				}),
			),
		)
	}
}

func (r *moduleRegistry) addMongoDB() {
	// Add MongoDB module if enabled

	if r.opt.WithMongoDB {
		r.mongodb = mongodb.NewDBModule(instance.config, instance.logger, r.opt.MongoDB...)
		instance.mongoDB = r.mongodb.Provide()

		instance.logger.Debug("Registering MongoDB module")

		r.fxOptions = append(r.fxOptions,
			// Provide contract.MongoDB for dependency injection
			fx.Provide(func() contract.MongoDB { return instance.mongoDB }),
			// Register MongoDB module with its lifecycle hooks
			fx.Module(r.mongodb.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: r.mongodb.OnStart,
						OnStop:  r.mongodb.OnStop,
					})
				}),
			),
		)
	}
}

func (r *moduleRegistry) addMigrate() {
	// Add MongoDB Migration module if enabled (requires WithMongoDB)

	if r.opt.WithMigrate {
		if !r.opt.WithMongoDB {
			instance.logger.Fatal("Migration module requires MongoDB. Set WithMongoDB: true")
		}

		var err error
		r.migrate, err = migrate.NewModule(instance.config, instance.logger, instance.mongoDB, r.opt.Migrate...)
		if err != nil {
			instance.logger.Fatal("Failed to create migration module", "error", err)
		}

		instance.logger.Debug("Registering MongoDB Migration module")

		r.fxOptions = append(r.fxOptions,
			fx.Module(r.migrate.Name(),
				// Declare explicit Fx dependency on contract.MongoDB to guarantee
				// MongoDB's OnStart (which establishes connections) runs before
				// the migration module's OnStart (which needs the DB connection).
				fx.Invoke(func(lc fx.Lifecycle, _ contract.MongoDB) {
					lc.Append(fx.Hook{
						OnStart: func(ctx context.Context) error {
							if err := r.migrate.OnStart(ctx); err != nil {
								return err
							}
							// Set instance.migrator after OnStart creates the migrator
							// (migrator is created in OnStart because MongoDB connections
							// are only available after MongoDB module's OnStart).
							// This runs at app start, long after New released the write
							// lock, and concurrent goroutines may already be reading via
							// goe.Migrate() — so the write must take the lock.
							migrator := r.migrate.Provide()
							instance.mu.Lock()
							instance.migrator = migrator
							instance.mu.Unlock()
							return nil
						},
						OnStop: r.migrate.OnStop,
					})
				}),
			),
		)
	}
}

func (r *moduleRegistry) addLock() {
	// Add Lock module if enabled

	if r.opt.WithLock {
		var err error
		r.lock, err = lock.NewModule(instance.config, instance.logger, r.opt.Lock...)
		if err != nil {
			instance.logger.Fatal("Failed to create lock module", "error", err)
		}
		instance.lockManager = r.lock.Provide()

		instance.logger.Debug("Registering Lock module")

		r.fxOptions = append(r.fxOptions,
			fx.Provide(func() contract.LockManager { return instance.lockManager }),
			fx.Module(r.lock.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: r.lock.OnStart,
						OnStop:  r.lock.OnStop,
					})
				}),
			),
		)
	}
}

func (r *moduleRegistry) addJob() {
	// Add Job module if enabled

	if r.opt.WithJob {
		var err error
		r.job, err = job.NewModule(instance.config, instance.logger, r.opt.Job...)
		if err != nil {
			instance.logger.Fatal("Failed to create job module", "error", err)
		}
		instance.jobManager = r.job.Provide()

		instance.logger.Debug("Registering Job module")

		r.fxOptions = append(r.fxOptions,
			fx.Provide(func() contract.JobManager { return instance.jobManager }),
			fx.Module(r.job.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: r.job.OnStart,
						OnStop:  r.job.OnStop,
					})
				}),
			),
		)
	}
}

func (r *moduleRegistry) addHealth() {
	// Add Health module if enabled

	if r.opt.WithHealth {
		r.health = health.NewModule(instance.config, instance.logger)
		instance.healthManager = r.health.Manager()

		instance.logger.Debug("Registering Health module")

		r.fxOptions = append(r.fxOptions,
			fx.Provide(func() contract.HealthManager { return instance.healthManager }),
			fx.Module(r.health.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: r.health.OnStart,
						OnStop:  r.health.OnStop,
					})
				}),
			),
		)
	}
}

func (r *moduleRegistry) addMetrics() {
	// Add Metrics module if enabled

	if r.opt.WithMetrics {
		r.metrics = metrics.NewModule(instance.config, instance.logger)
		instance.metricsManager = r.metrics.Manager()

		instance.logger.Debug("Registering Metrics module")

		r.fxOptions = append(r.fxOptions,
			fx.Provide(func() contract.MetricsManager { return instance.metricsManager }),
			fx.Module(r.metrics.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: r.metrics.OnStart,
						OnStop:  r.metrics.OnStop,
					})
				}),
			),
		)
	}
}

func (r *moduleRegistry) addOTel() {
	// Add OpenTelemetry module if enabled

	if r.opt.WithOTel {
		var err error
		r.otel, err = otel.NewModule(context.Background(), instance.config, instance.logger)
		if err != nil {
			instance.logger.Fatal("Failed to create OpenTelemetry module", "error", err)
		}
		instance.otelProvider = r.otel.Provider()

		instance.logger.Debug("Registering OpenTelemetry module")

		r.fxOptions = append(r.fxOptions,
			fx.Provide(func() contract.OTelProvider { return instance.otelProvider }),
			fx.Module(r.otel.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: r.otel.OnStart,
						OnStop:  r.otel.OnStop,
					})
				}),
			),
		)
	}
}

func (r *moduleRegistry) addHTTP() {
	// Add HTTP module if enabled

	if r.opt.WithHTTP {
		r.http = http.NewModule(instance.config, instance.logger, r.opt.HTTP...)
		instance.http = r.http.Provide()

		instance.logger.Debug("Registering HTTP module")

		r.fxOptions = append(r.fxOptions,
			fx.Provide(func() contract.HTTPKernel { return instance.http }),
			fx.Module(r.http.Name(),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: r.http.OnStart,
						OnStop:  r.http.OnStop,
					})
				}),
			),
		)
	}
}

// checkAndProvideServices inspects a module for common service provider methods
// and automatically registers them with Fx DI (like built-in modules do)
func checkAndProvideServices(module contract.Module) []fx.Option {
	var providers []fx.Option
	moduleValue := reflect.ValueOf(module)

	// Common service provider method patterns used by GOE modules
	serviceProviderMethods := []string{
		"Provide",           // Generic service provider
		"ProvideService",    // Generic service provider
		"ProvideClient",     // For client modules (like gRPC, HTTP clients)
		"ProvideManager",    // For manager services
		"ProvideHandler",    // For handler services
		"ProvideRepository", // For data access modules
		"ProvideCache",      // For cache services
		"ProvideDB",         // For database services
		"ProvideLogger",     // For logger services
		"ProvideConfig",     // For config services
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
