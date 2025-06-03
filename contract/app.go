package contract

import (
	"context"
	"time"
)

// App represents the application module interface
type App interface {
	Module

	// Container returns the dependency injection container
	Container() interface{}

	// Context returns the application context
	Context() context.Context

	// Run starts the application and blocks until it's stopped
	Run() error

	// RunWithTimeout starts the application and returns after the specified timeout
	RunWithTimeout(timeout time.Duration) error

	// RegisterModule registers a module with the application
	RegisterModule(module Module) App

	// RegisterModules registers multiple modules with the application
	RegisterModules(modules ...Module) App

	// RegisterProvider registers a provider function with the application
	RegisterProvider(provider interface{}, opts ...interface{}) App

	// RegisterProviders registers multiple provider functions with the application
	RegisterProviders(providers ...interface{}) App

	// Invoke executes a function after the application has started
	Invoke(function interface{}, opts ...interface{}) App

	// OnStart registers a function to be called when the application starts
	OnStart(function interface{}) App

	// OnStop registers a function to be called when the application stops
	OnStop(function interface{}) App
}

// AppProvider is a function that provides an App instance
type AppProvider func() App
