package goe

import (
	"context"

	"go.oease.dev/goe/v2/core/mongodb/migrate"

	"go.oease.dev/goe/v2/contract"
)

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

// Log returns the global logger instance.
// Panics if the application has not been initialized.
func Log() contract.Logger {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.logger == nil {
		panic("Application not initialized. Call goe.New() first")
	}

	return instance.logger
}

// LogOrNil returns the global logger instance, or nil if not yet initialized.
// Use this when logging is best-effort and a panic would be worse than silence.
func LogOrNil() contract.Logger {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

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

// Migrate returns the global MongoDB migration instance.
// Use this to run migrations, check status, or manage schema versions.
//
// Example:
//
//	// Run all pending migrations
//	result, err := goe.Migrate().Up(ctx)
//
//	// Check migration status
//	status, err := goe.Migrate().Status(ctx)
//
//	// Rollback last migration
//	result, err := goe.Migrate().Down(ctx, 1)
func Migrate() *migrate.Migrator {
	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.migrator == nil {
		panic("Migration module not initialized. Set WithMigrate: true and WithMongoDB: true in goe.New() options")
	}

	return instance.migrator
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
