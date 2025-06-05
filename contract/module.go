package contract

import "context"

// Module defines the interface that all Goe modules must implement
type Module interface {
	// Name returns the unique name of the module
	Name() string

	// OnStart is called when the module starts
	OnStart(ctx context.Context) error

	// OnStop is called when the module stops
	OnStop(ctx context.Context) error
}

// Provider represents a function that provides dependencies
type Provider interface{}

// Invoker represents a function that should be invoked after dependencies are provided
type Invoker interface{}
