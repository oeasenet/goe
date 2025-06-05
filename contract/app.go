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
}
