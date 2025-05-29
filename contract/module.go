package contract

import (
	"context"
)

// Module represents a framework module with lifecycle hooks
type Module interface {
	// Name returns the name of the module
	Name() string

	// Initialize is called when the module is being initialized
	// It should set up any necessary resources
	Initialize(ctx context.Context) error

	// Start is called when the application is starting
	// It should start any background processes
	Start(ctx context.Context) error

	// Stop is called when the application is stopping
	// It should gracefully shut down any background processes
	Stop(ctx context.Context) error
}

// ModuleProvider is a function that provides a module instance
type ModuleProvider[T Module] func() T
