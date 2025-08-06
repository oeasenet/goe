package goe

import (
	"context"
)

// OnStart is deprecated. Use Options.OnStart when creating the application instead.
// This function exists for backward compatibility but will panic as Fx doesn't support
// adding lifecycle hooks after the container is built.
//
// Instead, use:
//
//	app := goe.New(goe.Options{
//	    OnStart: []func(context.Context) error{myStartupFunc},
//	})
func OnStart(fn func(context.Context) error) {
	panic("OnStart cannot be called after goe.New(). Use Options.OnStart instead")
}

// OnStop is deprecated. Use Options.OnStop when creating the application instead.
// This function exists for backward compatibility but will panic as Fx doesn't support
// adding lifecycle hooks after the container is built.
//
// Instead, use:
//
//	app := goe.New(goe.Options{
//	    OnStop: []func(context.Context) error{myCleanupFunc},
//	})
func OnStop(fn func(context.Context) error) {
	panic("OnStop cannot be called after goe.New(). Use Options.OnStop instead")
}

// WithHooks is a helper function to create Options with lifecycle hooks.
// This provides a fluent API for registering hooks when creating the application.
//
// Example:
//
//	app := goe.New(
//	    goe.WithHooks(
//	        goe.WithOnStart(initializeData),
//	        goe.WithOnStop(cleanup),
//	    ),
//	    goe.Options{
//	        WithHTTP: true,
//	        WithDB: true,
//	    },
//	)
func WithHooks(hooks ...HookOption) Options {
	opts := Options{}
	for _, hook := range hooks {
		hook(&opts)
	}
	return opts
}

// HookOption is a function that modifies Options to add hooks
type HookOption func(*Options)

// WithOnStart returns a HookOption that adds an OnStart hook
func WithOnStart(fn func(context.Context) error) HookOption {
	return func(opts *Options) {
		opts.OnStart = append(opts.OnStart, fn)
	}
}

// WithOnStop returns a HookOption that adds an OnStop hook
func WithOnStop(fn func(context.Context) error) HookOption {
	return func(opts *Options) {
		opts.OnStop = append(opts.OnStop, fn)
	}
}
