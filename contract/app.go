package contract

import (
	"context"
	"go.uber.org/fx"
)

// Application defines the main application interface
type Application interface {
	// Name returns the application name
	Name() string

	// Version returns the application version
	Version() string

	// Environment returns the current environment (dev, prod, etc.)
	Environment() string

	// Context returns the application context
	Context() context.Context

	// Container returns the underlying Fx app instance
	Container() *fx.App

	// IsRunning returns true if the application is running
	IsRunning() bool

	// Register registers new modules, providers, or invokers
	Register(options ...fx.Option) error

	// AddModule adds a module to the application
	AddModule(module Module) error

	// AddProvider adds a provider to the application
	AddProvider(provider Provider) error

	// AddInvoker adds an invoker to the application
	AddInvoker(invoker Invoker) error

	// Start initializes and starts the application and all its modules.
	// This method:
	// 1. Starts the underlying Fx dependency injection container
	// 2. Calls OnStart() on all registered modules in dependency order
	// 3. Sets the application state to "running"
	//
	// The provided context can be used to set a timeout for the startup process.
	// If any module's OnStart() method returns an error, the startup process is aborted,
	// and all successfully started modules are stopped in reverse order.
	//
	// Start() is non-blocking - use it when you need programmatic control over the
	// application lifecycle. For simple applications, use Run() instead.
	//
	// Example:
	//   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	//   defer cancel()
	//   if err := app.Start(ctx); err != nil {
	//       log.Fatal("Failed to start application:", err)
	//   }
	//   defer app.Stop(context.Background())
	Start(ctx context.Context) error

	// Stop gracefully shuts down the application and all its modules.
	// This method:
	// 1. Sets the application state to "not running"
	// 2. Calls OnStop() on all registered modules in reverse order
	// 3. Stops the underlying Fx dependency injection container
	//
	// The provided context can be used to set a timeout for the shutdown process.
	// Module OnStop() errors are logged but don't prevent the shutdown process.
	//
	// Stop() should be called to ensure proper cleanup of resources like
	// database connections, HTTP servers, background goroutines, etc.
	//
	// Example:
	//   ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//   defer cancel()
	//   if err := app.Stop(ctx); err != nil {
	//       log.Error("Error during shutdown:", err)
	//   }
	Stop(ctx context.Context) error

	// Run starts and runs the application in a blocking manner.
	// This is a convenience method that combines Start() and blocks until the application
	// is shut down (typically via signal handling like Ctrl+C).
	//
	// Run() is ideal for simple applications where you don't need programmatic control
	// over the startup/shutdown process. For more control, use Start() and Stop() instead.
	//
	// Note: Run() handles its own context and signal handling internally.
	// If the application fails to start, Run() will panic or exit.
	//
	// Example (simple application):
	//   app := goe.New(goe.Options{WithHTTP: true})
	//   app.Run() // Blocks until shutdown signal
	//
	// Example (programmatic control):
	//   app := goe.New(goe.Options{WithHTTP: true})
	//   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	//   defer cancel()
	//   if err := app.Start(ctx); err != nil {
	//       log.Fatal("Failed to start:", err)
	//   }
	//   // Do other work...
	//   app.Stop(context.Background())
	Run()
}
