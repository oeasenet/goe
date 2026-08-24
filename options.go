package goe

import (
	"context"
	"time"

	"go.oease.dev/goe/v2/core/cache"
	"go.oease.dev/goe/v2/core/http"
	"go.oease.dev/goe/v2/core/job"
	"go.oease.dev/goe/v2/core/lock"
	"go.oease.dev/goe/v2/core/log"
	"go.oease.dev/goe/v2/core/mongodb"
	"go.oease.dev/goe/v2/core/mongodb/migrate"
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
	WithMigrate     bool           // Enable MongoDB Migration module (requires WithMongoDB)
	WithLock        bool           // Enable Lock module (distributed mutex)
	WithJob         bool           // Enable Job module (background job processing)
	WithHealth      bool           // Enable Health module (health checks)
	WithMetrics     bool           // Enable Metrics module (Prometheus metrics)
	WithOTel        bool           // Enable OpenTelemetry module (distributed tracing)
	ConfigOverrides map[string]any // Override any environment variables

	// HTTP configures the HTTP kernel from Go code instead of environment
	// variables. A non-nil value implies WithHTTP, so the module does not have
	// to be enabled separately.
	//
	// Options are applied after the environment, so anything set here wins over
	// the matching FIBER_*/HTTP_*/VIEWS_* variable while the environment still
	// supplies everything left unset.
	//
	//	goe.New(goe.Options{
	//	    HTTP: []goehttp.Option{
	//	        goehttp.WithPort(8080),
	//	        goehttp.WithBodyLimit(16 << 20),
	//	    },
	//	})
	//
	// See the core/http package for the full option list.
	HTTP []http.Option

	// Log configures the log module from Go code. The log module is always
	// enabled, so unlike the other module option slices this implies nothing.
	//
	// Options are applied after the environment, so anything set here wins
	// over the matching APP_* variable. The main use is carrying an
	// ldflags-stamped build version onto every JSON log line:
	//
	//	goe.New(goe.Options{
	//	    Log: []log.Option{log.WithVersion(buildinfo.Version)},
	//	})
	//
	// See the core/log package for the full option list.
	Log []log.Option

	// Cache configures the cache module from Go code instead of environment
	// variables. A non-nil value implies WithCache, so the module does not
	// have to be enabled separately.
	//
	// Options are applied after the environment, so anything set here wins
	// over the matching CACHE_* variable while the environment still supplies
	// everything left unset. Credentials are the exception: CACHE_REDIS_URL,
	// CACHE_REDIS_USERNAME and CACHE_REDIS_PASSWORD have no option and stay
	// environment-only.
	//
	//	goe.New(goe.Options{
	//	    Cache: []cache.Option{
	//	        cache.WithDriver("redis"),
	//	        cache.WithTTL(30 * time.Minute),
	//	    },
	//	})
	//
	// See the core/cache package for the full option list.
	Cache []cache.Option

	// Job configures the background job system from Go code instead of
	// environment variables. A non-nil value implies WithJob, so the module
	// does not have to be enabled separately.
	//
	// Options are applied after the environment, so anything set here wins
	// over the matching JOB_* variable while the environment still supplies
	// everything left unset. Credentials are the exception: JOB_REDIS_URL,
	// JOB_REDIS_USERNAME and JOB_REDIS_PASSWORD have no option and stay
	// environment-only.
	//
	//	goe.New(goe.Options{
	//	    Job: []job.Option{
	//	        job.WithConcurrency(10),
	//	        job.WithDefaultQueue("critical"),
	//	    },
	//	})
	//
	// See the core/job package for the full option list.
	Job []job.Option

	// Lock configures the distributed lock system from Go code instead of
	// environment variables. A non-nil value implies WithLock, so the module
	// does not have to be enabled separately.
	//
	// Options are applied after the environment, so anything set here wins
	// over the matching LOCK_* variable while the environment still supplies
	// everything left unset. Credentials are the exception: lock.WithRedisURL
	// rejects URLs that embed user:pass — LOCK_REDIS_USERNAME and
	// LOCK_REDIS_PASSWORD stay environment-only and apply to whichever
	// topology the URL declares.
	//
	//	goe.New(goe.Options{
	//	    Lock: []lock.Option{
	//	        lock.WithDefaultExpiry(10 * time.Second),
	//	        lock.WithKeyPrefix("myapp:lock:"),
	//	    },
	//	})
	//
	// See the core/lock package for the full option list.
	Lock []lock.Option

	// MongoDB configures the MongoDB module from Go code instead of
	// environment variables. A non-nil value implies WithMongoDB, so the
	// module does not have to be enabled separately.
	//
	// Options are applied after the environment, so anything set here wins
	// over the matching MONGO_* variable while the environment still supplies
	// everything left unset. Credentials are the exception: MONGO_URI,
	// MONGO_USERNAME and MONGO_PASSWORD have no option and stay
	// environment-only.
	//
	//	goe.New(goe.Options{
	//	    MongoDB: []mongodb.Option{
	//	        mongodb.WithDatabase("myapp"),
	//	        mongodb.WithMaxPoolSize(50),
	//	    },
	//	})
	//
	// See the core/mongodb package for the full option list.
	MongoDB []mongodb.Option

	// Migrate configures the MongoDB migration module from Go code instead
	// of environment variables. A non-nil value implies WithMigrate and
	// WithMongoDB, so neither module has to be enabled separately.
	//
	// Options are applied after the environment, so anything set here wins
	// over the matching MONGODB_MIGRATE_* variable while the environment
	// still supplies everything left unset.
	//
	//	goe.New(goe.Options{
	//	    Migrate: []migrate.Option{
	//	        migrate.WithAutoMigrate(true),
	//	        migrate.WithCollection("_migrations"),
	//	    },
	//	})
	//
	// See the core/mongodb/migrate package for the full option list.
	Migrate []migrate.Option

	// HTTPPort overrides the HTTP port.
	//
	// Deprecated: use HTTP with goehttp.WithPort instead. This field still
	// works, but it is applied as an environment override, so WithPort takes
	// precedence over it.
	HTTPPort int

	// Shutdown configuration
	ShutdownTimeout time.Duration // Total shutdown timeout (default: 30s)
	DrainTimeout    time.Duration // HTTP drain timeout (default: 5s)

	// Lifecycle hooks - these are executed through Fx's lifecycle system
	OnStart []func(context.Context) error // Functions to run after all modules start
	OnStop  []func(context.Context) error // Functions to run before modules stop
}
