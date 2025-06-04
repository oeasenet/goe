package goe

import (
	"sync"
	"time"

	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/app"
	"go.oease.dev/goe/v2/core/cache"
	"go.oease.dev/goe/v2/core/config"
	"go.oease.dev/goe/v2/core/event"
	"go.oease.dev/goe/v2/core/http"
	"go.oease.dev/goe/v2/core/log"
)

const Version = "v2.0.0-dev-1"

var (
	instance     *Framework
	instanceOnce sync.Once
)

// Framework is the main framework struct
type Framework struct {
	app    contract.App
	config contract.Config
	http   contract.Http
	log    contract.Log
	event  contract.Event
	cache  contract.Cache
}

// ModuleProviders returns the core module providers for dependency injection
func ModuleProviders() []interface{} {
	return []interface{}{
		app.Provider,
		config.Provider,
		log.Provider,
		http.Provider,
		event.Provider,
		cache.Provider,
	}
}

// New creates a new framework instance
func New() *Framework {
	instanceOnce.Do(func() {
		// Create a new framework instance
		instance = &Framework{}

		// Initialize all modules directly
		instance.app = app.New()
		instance.config = config.New()
		instance.http = http.New()
		instance.log = log.New()
		instance.event = event.New()
		instance.cache = cache.New()
	})
	return instance
}

// App returns the app module
func (f *Framework) App() contract.App {
	return f.app
}

// Config returns the config module
func (f *Framework) Config() contract.Config {
	return f.config
}

// Http returns the http module
func (f *Framework) Http() contract.Http {
	return f.http
}

// Log returns the log module
func (f *Framework) Log() contract.Log {
	return f.log
}

// Event returns the event module
func (f *Framework) Event() contract.Event {
	return f.event
}

// Cache returns the cache module
func (f *Framework) Cache() contract.Cache {
	return f.cache
}

// Run initializes and runs the application
func (f *Framework) Run() error {
	// Run the application
	return f.app.Run()
}

// RunWithTimeout initializes and runs the application with a timeout
func (f *Framework) RunWithTimeout(timeout time.Duration) error {
	// Run the application with a timeout
	return f.app.RunWithTimeout(timeout)
}

// Global accessor functions

// App returns the global app module
func App() contract.App {
	return New().App()
}

// Config returns the global config module
func Config() contract.Config {
	return New().Config()
}

// Http returns the global http module
func Http() contract.Http {
	return New().Http()
}

// Log returns the global log module
func Log() contract.Log {
	return New().Log()
}

// Event returns the global event module
func Event() contract.Event {
	return New().Event()
}

// Cache returns the global cache module
func Cache() contract.Cache {
	return New().Cache()
}
