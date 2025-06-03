package main

import (
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/fx"
)

// ModuleParams is a container for module dependencies used as input to New
type ModuleParams struct {
	fx.In

	// Core modules
	App    contract.App    `optional:"true"`
	Config contract.Config `optional:"true"`
	Log    contract.Log    `optional:"true"`
	Http   contract.Http   `optional:"true"`
	Event  contract.Event  `optional:"true"`
	Cache  contract.Cache  `optional:"true"`
}

// Module is a container for all module dependencies
type Module struct {
	// Core modules
	App    contract.App
	Config contract.Config
	Log    contract.Log
	Http   contract.Http
	Event  contract.Event
	Cache  contract.Cache
}

// New creates a new Module instance
func New(params ModuleParams) *Module {
	return &Module{
		App:    params.App,
		Config: params.Config,
		Log:    params.Log,
		Http:   params.Http,
		Event:  params.Event,
		Cache:  params.Cache,
	}
}

// Provider provides a Module instance
func Provider() interface{} {
	return New
}

// RegisterDI registers the DI module with the application
func RegisterDI(app contract.App) {
	app.RegisterProvider(Provider())
}
