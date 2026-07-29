package goe

import (
	"context"
	"time"

	"go.oease.dev/goe/v2/core/http"
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
