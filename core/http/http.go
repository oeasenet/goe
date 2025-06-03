package http

import (
	"context"
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
	config := &fiber.Config{}
	return &Http{
		config: config,
		app:    fiber.New(*config),
	}
}

// Name returns the name of the module
func (h *Http) Name() string {
	return "http"
}

// Initialize initializes the http module
func (h *Http) Initialize(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.app == nil {
		h.app = fiber.New(*h.config)
	}
	return nil
}

// Start starts the http module
func (h *Http) Start(ctx context.Context) error {
	// We don't actually start the HTTP server here because it would block
	// Instead, we just make sure the app is initialized
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.app == nil {
		h.app = fiber.New(*h.config)
	}
	return nil
}

// Stop stops the http module
func (h *Http) Stop(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.app != nil {
		return h.app.Shutdown()
	}
	return nil
}

// Fiber returns the Fiber app instance
func (h *Http) Fiber() *fiber.App {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.app
}

// Get registers a route for GET requests
func (h *Http) Get(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	if handler != nil {
		h.app.Get(path, handler, middlewares...)
	}

	return h
}

// Post registers a route for POST requests
func (h *Http) Post(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	if handler != nil {
		h.app.Post(path, handler, middlewares...)
	}
	return h
}

// Put registers a route for PUT requests
func (h *Http) Put(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	if handler != nil {
		h.app.Put(path, handler, middlewares...)
	}
	return h
}

// Delete registers a route for DELETE requests
func (h *Http) Delete(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	if handler != nil {
		h.app.Delete(path, handler, middlewares...)
	}
	return h
}

// Patch registers a route for PATCH requests
func (h *Http) Patch(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	if handler != nil {
		h.app.Patch(path, handler, middlewares...)
	}
	return h
}

// Options registers a route for OPTIONS requests
func (h *Http) Options(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	if handler != nil {
		h.app.Options(path, handler, middlewares...)
	}
	return h
}

// Head registers a route for HEAD requests
func (h *Http) Head(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	if handler != nil {
		h.app.Head(path, handler, middlewares...)
	}
	return h
}

// All registers a route for all HTTP methods
func (h *Http) All(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	if handler != nil {
		h.app.All(path, handler, middlewares...)
	}
	return h
}

func (h *Http) Connect(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	if handler != nil {
		h.app.Connect(path, handler, middlewares...)
	}
	return h
}

func (h *Http) Trace(path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	if handler != nil {
		h.app.Trace(path, handler, middlewares...)
	}
	return h
}

func (h *Http) Add(methods []string, path string, handler contract.HttpHandler, middlewares ...contract.HttpHandler) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	if handler != nil {
		h.app.Add(methods, path, handler, middlewares...)
	}
	return h
}

// Group creates a new route group with prefix
func (h *Http) Group(prefix string, handlers ...contract.HttpHandler) contract.RouteGroup {
	h.mu.Lock()
	defer h.mu.Unlock()

	group := h.app.Group(prefix, handlers...)
	return &RouteGroup{group: group}
}

// Use registers middleware
func (h *Http) Use(handlers ...any) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.app.Use(handlers...)
	return h
}

// Static serves static files
func (h *Http) Static(prefix, root string, config ...static.Config) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Use the static middleware: app.Use(prefix, static.New(root, staticCfg))
	// If prefix is empty, it means serving from root, which is common for single page apps or root static.
	// The static.New middleware handles the case where prefix might be added to the paths it serves from.
	// For now, direct mapping:
	var cfg static.Config
	if len(config) > 0 {
		cfg = config[0]
	}

	if prefix == "" { // static.New often expects a root path to serve, prefix is handled by app.Use
		h.app.Use(static.New(root, cfg))
	} else {
		h.app.Use(prefix, static.New(root, cfg))
	}
	return h
}

// Listen starts the HTTP server
func (h *Http) Listen(address string) error {
	return h.app.Listen(address)
}

// Shutdown gracefully shuts down the HTTP server
func (h *Http) Shutdown(ctx context.Context) error {
	return h.app.Shutdown()
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
