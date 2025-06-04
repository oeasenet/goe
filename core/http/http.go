package http

import (
	"context"
	"go.oease.dev/goe/v2"
	"sync"

	"github.com/gofiber/fiber/v3"
	static "github.com/gofiber/fiber/v3/middleware/static" // New import
	"go.oease.dev/goe/v2/contract"
)

// Http implements the contract.Http interface
type Http struct {
	mu     sync.RWMutex
	app    *fiber.App
	config *fiber.Config
}

// New creates a new Http instance
func New() *Http {
	config := &fiber.Config{
		ServerHeader:                 goe.Config().GetDefault("FIBER_SERVER_HEADER", "GOE Web Server/"+goe.Version),
		BodyLimit:                    goe.Config().GetIntDefault("FIBER_BODY_LIMIT", 2048*1024*1024),
		Concurrency:                  goe.Config().GetIntDefault("FIBER_CONCURRENCY", 256*1024),
		PassLocalsToViews:            true,
		ReadBufferSize:               4096,
		WriteBufferSize:              4096,
		ProxyHeader:                  goe.Config().GetDefault("FIBER_PROXY_HEADER", "X-Forwarded-For"),
		ErrorHandler:                 nil,
		DisableKeepalive:             false,
		DisableDefaultDate:           false,
		DisableDefaultContentType:    false,
		DisableHeaderNormalizing:     false,
		AppName:                      "",
		StreamRequestBody:            false,
		DisablePreParseMultipartForm: false,
		ReduceMemoryUsage:            false,
		JSONEncoder:                  nil,
		JSONDecoder:                  nil,
		CBOREncoder:                  nil,
		CBORDecoder:                  nil,
		XMLEncoder:                   nil,
		XMLDecoder:                   nil,
		TrustProxy:                   false,
		TrustProxyConfig:             fiber.TrustProxyConfig{},
		EnableIPValidation:           false,
		ColorScheme:                  fiber.Colors{},
		StructValidator:              nil,
		RequestMethods:               nil,
		EnableSplittingOnParsers:     false,
	}
	return &Http{
		config: config,
		app:    fiber.New(*config),
	}
}

// Name returns the name of the module
func (m *Http) Name() string {
	return "http"
}

// Initialize initializes the http module
func (m *Http) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.app == nil {
		m.app = fiber.New(*m.config)
	}
	return nil
}

// Start starts the http module
func (m *Http) Start(ctx context.Context) error {
	// We don't actually start the HTTP server here because it would block
	// Instead, we just make sure the app is initialized
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.app == nil {
		m.app = fiber.New(*m.config)
	}
	return nil
}

// Stop stops the http module
func (m *Http) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.app != nil {
		return m.app.Shutdown()
	}
	return nil
}

// Fiber returns the Fiber app instance
func (m *Http) Fiber() *fiber.App {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.app
}

// Get registers a route for GET requests
func (m *Http) Get(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	m.mu.Lock()
	defer m.mu.Unlock()

	if handler != nil {
		m.app.Get(path, handler, middlewares...)
	}

	return m
}

// Post registers a route for POST requests
func (m *Http) Post(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	m.mu.Lock()
	defer m.mu.Unlock()

	if handler != nil {
		m.app.Post(path, handler, middlewares...)
	}
	return m
}

// Put registers a route for PUT requests
func (m *Http) Put(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	m.mu.Lock()
	defer m.mu.Unlock()

	if handler != nil {
		m.app.Put(path, handler, middlewares...)
	}
	return m
}

// Delete registers a route for DELETE requests
func (m *Http) Delete(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	m.mu.Lock()
	defer m.mu.Unlock()

	if handler != nil {
		m.app.Delete(path, handler, middlewares...)
	}
	return m
}

// Patch registers a route for PATCH requests
func (m *Http) Patch(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	m.mu.Lock()
	defer m.mu.Unlock()

	if handler != nil {
		m.app.Patch(path, handler, middlewares...)
	}
	return m
}

// Options registers a route for OPTIONS requests
func (m *Http) Options(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	m.mu.Lock()
	defer m.mu.Unlock()

	if handler != nil {
		m.app.Options(path, handler, middlewares...)
	}
	return m
}

// Head registers a route for HEAD requests
func (m *Http) Head(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	m.mu.Lock()
	defer m.mu.Unlock()

	if handler != nil {
		m.app.Head(path, handler, middlewares...)
	}
	return m
}

// All registers a route for all HTTP methods
func (m *Http) All(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	m.mu.Lock()
	defer m.mu.Unlock()

	if handler != nil {
		m.app.All(path, handler, middlewares...)
	}
	return m
}

func (m *Http) Connect(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	m.mu.Lock()
	defer m.mu.Unlock()

	if handler != nil {
		m.app.Connect(path, handler, middlewares...)
	}
	return m
}

func (m *Http) Trace(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	m.mu.Lock()
	defer m.mu.Unlock()

	if handler != nil {
		m.app.Trace(path, handler, middlewares...)
	}
	return m
}

func (m *Http) Add(methods []string, path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	m.mu.Lock()
	defer m.mu.Unlock()

	if handler != nil {
		m.app.Add(methods, path, handler, middlewares...)
	}
	return m
}

// Group creates a new route group with prefix
func (m *Http) Group(prefix string, handlers ...contract.HttpHandler) contract.RouteGroup {
	m.mu.Lock()
	defer m.mu.Unlock()

	group := m.app.Group(prefix, handlers...)
	return &RouteGroup{group: group}
}

// Use registers middleware
func (m *Http) Use(handlers ...any) contract.Http {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.app.Use(handlers...)
	return m
}

// Static serves static files
func (m *Http) Static(prefix, root string, config ...static.Config) contract.Http {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Use the static middleware: app.Use(prefix, static.New(root, staticCfg))
	// If prefix is empty, it means serving from root, which is common for single page apps or root static.
	// The static.New middleware handles the case where prefix might be added to the paths it serves from.
	// For now, direct mapping:
	var cfg static.Config
	if len(config) > 0 {
		cfg = config[0]
	}

	if prefix == "" { // static.New often expects a root path to serve, prefix is handled by app.Use
		m.app.Use(static.New(root, cfg))
	} else {
		m.app.Use(prefix, static.New(root, cfg))
	}
	return m
}

// Listen starts the HTTP server
func (m *Http) Listen(address string) error {
	return m.app.Listen(address)
}

// Shutdown gracefully shuts down the HTTP server
func (m *Http) Shutdown(ctx context.Context) error {
	return m.app.Shutdown()
}

// RouteGroup implements the contract.RouteGroup interface
type RouteGroup struct {
	group fiber.Router
}

func (r *RouteGroup) Get(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) {
	r.group.Get(path, handler, middlewares...)
}

func (r *RouteGroup) Post(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) {
	r.group.Post(path, handler, middlewares...)
}

func (r *RouteGroup) Put(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) {
	r.group.Put(path, handler, middlewares...)
}

func (r *RouteGroup) Delete(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) {
	r.group.Delete(path, handler, middlewares...)
}

func (r *RouteGroup) Patch(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) {
	r.group.Patch(path, handler, middlewares...)
}

func (r *RouteGroup) Options(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) {
	r.group.Options(path, handler, middlewares...)
}

func (r *RouteGroup) Head(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) {
	r.group.Head(path, handler, middlewares...)
}

func (r *RouteGroup) All(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) {
	r.group.All(path, handler, middlewares...)
}

func (r *RouteGroup) Connect(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) {
	r.group.Connect(path, handler, middlewares...)
}

func (r *RouteGroup) Trace(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) {
	r.group.Trace(path, handler, middlewares...)
}

func (r *RouteGroup) Add(methods []string, path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) {
	r.group.Add(methods, path, handler, middlewares...)
}

func (r *RouteGroup) Group(prefix string, handlers ...contract.HttpHandler) contract.RouteGroup {
	group := r.group.Group(prefix, handlers...)
	return &RouteGroup{group: group}
}

func (r *RouteGroup) Use(handlers ...any) {
	r.group.Use(handlers...)
}

// Provider provides an Http instance
func Provider() contract.Http {
	return New()
}
