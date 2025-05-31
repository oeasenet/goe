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

// Engine returns the underlying HTTP engine
func (h *Http) Engine() interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.app
}

// Fiber returns the Fiber app instance
func (h *Http) Fiber() *fiber.App {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.app
}

// Get registers a route for GET requests
func (h *Http) Get(path string, handlers ...interface{}) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fnOld, ok := handler.(func(interface{}) error); ok { // Legacy generic handler
			fiberHandlers = append(fiberHandlers, func(c fiber.Ctx) error { // Wrapper now uses fiber.Ctx
				return fnOld(c) // c is fiber.Ctx (interface)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		if len(fiberHandlers) == 1 {
			h.app.Get(path, fiberHandlers[0])
		} else {
			h.app.Get(path, fiberHandlers[0], fiberHandlers[1:]...)
		}
	}
	return h
}

// Post registers a route for POST requests
func (h *Http) Post(path string, handlers ...interface{}) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fnOld, ok := handler.(func(interface{}) error); ok { // Legacy generic handler
			fiberHandlers = append(fiberHandlers, func(c fiber.Ctx) error { // Wrapper now uses fiber.Ctx
				return fnOld(c) // c is fiber.Ctx (interface)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		if len(fiberHandlers) == 1 {
			h.app.Post(path, fiberHandlers[0])
		} else {
			h.app.Post(path, fiberHandlers[0], fiberHandlers[1:]...)
		}
	}
	return h
}

// Put registers a route for PUT requests
func (h *Http) Put(path string, handlers ...interface{}) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fnOld, ok := handler.(func(interface{}) error); ok { // Legacy generic handler
			fiberHandlers = append(fiberHandlers, func(c fiber.Ctx) error { // Wrapper now uses fiber.Ctx
				return fnOld(c) // c is fiber.Ctx (interface)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		if len(fiberHandlers) == 1 {
			h.app.Put(path, fiberHandlers[0])
		} else {
			h.app.Put(path, fiberHandlers[0], fiberHandlers[1:]...)
		}
	}
	return h
}

// Delete registers a route for DELETE requests
func (h *Http) Delete(path string, handlers ...interface{}) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fnOld, ok := handler.(func(interface{}) error); ok { // Legacy generic handler
			fiberHandlers = append(fiberHandlers, func(c fiber.Ctx) error { // Wrapper now uses fiber.Ctx
				return fnOld(c) // c is fiber.Ctx (interface)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		if len(fiberHandlers) == 1 {
			h.app.Delete(path, fiberHandlers[0])
		} else {
			h.app.Delete(path, fiberHandlers[0], fiberHandlers[1:]...)
		}
	}
	return h
}

// Patch registers a route for PATCH requests
func (h *Http) Patch(path string, handlers ...interface{}) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fnOld, ok := handler.(func(interface{}) error); ok { // Legacy generic handler
			fiberHandlers = append(fiberHandlers, func(c fiber.Ctx) error { // Wrapper now uses fiber.Ctx
				return fnOld(c) // c is fiber.Ctx (interface)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		if len(fiberHandlers) == 1 {
			h.app.Patch(path, fiberHandlers[0])
		} else {
			h.app.Patch(path, fiberHandlers[0], fiberHandlers[1:]...)
		}
	}
	return h
}

// Options registers a route for OPTIONS requests
func (h *Http) Options(path string, handlers ...interface{}) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fnOld, ok := handler.(func(interface{}) error); ok { // Legacy generic handler
			fiberHandlers = append(fiberHandlers, func(c fiber.Ctx) error { // Wrapper now uses fiber.Ctx
				return fnOld(c) // c is fiber.Ctx (interface)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		if len(fiberHandlers) == 1 {
			h.app.Options(path, fiberHandlers[0])
		} else {
			h.app.Options(path, fiberHandlers[0], fiberHandlers[1:]...)
		}
	}
	return h
}

// Head registers a route for HEAD requests
func (h *Http) Head(path string, handlers ...interface{}) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fnOld, ok := handler.(func(interface{}) error); ok { // Legacy generic handler
			fiberHandlers = append(fiberHandlers, func(c fiber.Ctx) error { // Wrapper now uses fiber.Ctx
				return fnOld(c) // c is fiber.Ctx (interface)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		if len(fiberHandlers) == 1 {
			h.app.Head(path, fiberHandlers[0])
		} else {
			h.app.Head(path, fiberHandlers[0], fiberHandlers[1:]...)
		}
	}
	return h
}

// All registers a route for all HTTP methods
func (h *Http) All(path string, handlers ...interface{}) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fnOld, ok := handler.(func(interface{}) error); ok { // Legacy generic handler
			fiberHandlers = append(fiberHandlers, func(c fiber.Ctx) error { // Wrapper now uses fiber.Ctx
				return fnOld(c) // c is fiber.Ctx (interface)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		if len(fiberHandlers) == 1 {
			h.app.All(path, fiberHandlers[0])
		} else {
			h.app.All(path, fiberHandlers[0], fiberHandlers[1:]...)
		}
	}
	return h
}

// Group creates a new route group with prefix
func (h *Http) Group(prefix string, handlers ...interface{}) contract.RouteGroup {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fnOld, ok := handler.(func(interface{}) error); ok { // Legacy generic handler
			fiberHandlers = append(fiberHandlers, func(c fiber.Ctx) error { // Wrapper now uses fiber.Ctx
				return fnOld(c) // c is fiber.Ctx (interface)
			})
		}
	}

	group := h.app.Group(prefix, fiberHandlers...)
	return &RouteGroup{group: group}
}

// Use registers middleware
func (h *Http) Use(handlers ...interface{}) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		}
	}

	for _, handler := range fiberHandlers {
		h.app.Use(handler)
	}
	return h
}

// Static serves static files
func (h *Http) Static(prefix, root string, config ...interface{}) contract.Http {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Convert config to static.Config
	var staticCfg static.Config // Use aliased import
	if len(config) > 0 {
		if sc, ok := config[0].(static.Config); ok { // Cast to aliased type
			staticCfg = sc
		}
		// Not attempting to map from a hypothetical old fiber.Static struct here,
		// as the type itself was undefined. Users must pass static.Config if they want custom config.
	}
	// Use the static middleware: app.Use(prefix, static.New(root, staticCfg))
	// If prefix is empty, it means serving from root, which is common for single page apps or root static.
	// The static.New middleware handles the case where prefix might be added to the paths it serves from.
	// For now, direct mapping:
	if prefix == "" { // static.New often expects a root path to serve, prefix is handled by app.Use
		h.app.Use(static.New(root, staticCfg))
	} else {
		h.app.Use(prefix, static.New(root, staticCfg))
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

// Get registers a route for GET requests within this group
func (g *RouteGroup) Get(path string, handlers ...interface{}) contract.RouteGroup {
	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fnOld, ok := handler.(func(interface{}) error); ok { // Legacy generic handler
			fiberHandlers = append(fiberHandlers, func(c fiber.Ctx) error { // Wrapper now uses fiber.Ctx
				return fnOld(c) // c is fiber.Ctx (interface)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		if len(fiberHandlers) == 1 {
			g.group.Get(path, fiberHandlers[0])
		} else {
			g.group.Get(path, fiberHandlers[0], fiberHandlers[1:]...)
		}
	}
	return g
}

// Post registers a route for POST requests within this group
func (g *RouteGroup) Post(path string, handlers ...interface{}) contract.RouteGroup {
	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fnOld, ok := handler.(func(interface{}) error); ok { // Legacy generic handler
			fiberHandlers = append(fiberHandlers, func(c fiber.Ctx) error { // Wrapper now uses fiber.Ctx
				return fnOld(c) // c is fiber.Ctx (interface)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		if len(fiberHandlers) == 1 {
			g.group.Post(path, fiberHandlers[0])
		} else {
			g.group.Post(path, fiberHandlers[0], fiberHandlers[1:]...)
		}
	}
	return g
}

// Put registers a route for PUT requests within this group
func (g *RouteGroup) Put(path string, handlers ...interface{}) contract.RouteGroup {
	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fnOld, ok := handler.(func(interface{}) error); ok { // Legacy generic handler
			fiberHandlers = append(fiberHandlers, func(c fiber.Ctx) error { // Wrapper now uses fiber.Ctx
				return fnOld(c) // c is fiber.Ctx (interface)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		if len(fiberHandlers) == 1 {
			g.group.Put(path, fiberHandlers[0])
		} else {
			g.group.Put(path, fiberHandlers[0], fiberHandlers[1:]...)
		}
	}
	return g
}

// Delete registers a route for DELETE requests within this group
func (g *RouteGroup) Delete(path string, handlers ...interface{}) contract.RouteGroup {
	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fnOld, ok := handler.(func(interface{}) error); ok { // Legacy generic handler
			fiberHandlers = append(fiberHandlers, func(c fiber.Ctx) error { // Wrapper now uses fiber.Ctx
				return fnOld(c) // c is fiber.Ctx (interface)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		if len(fiberHandlers) == 1 {
			g.group.Delete(path, fiberHandlers[0])
		} else {
			g.group.Delete(path, fiberHandlers[0], fiberHandlers[1:]...)
		}
	}
	return g
}

// Patch registers a route for PATCH requests within this group
func (g *RouteGroup) Patch(path string, handlers ...interface{}) contract.RouteGroup {
	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fnOld, ok := handler.(func(interface{}) error); ok { // Legacy generic handler
			fiberHandlers = append(fiberHandlers, func(c fiber.Ctx) error { // Wrapper now uses fiber.Ctx
				return fnOld(c) // c is fiber.Ctx (interface)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		if len(fiberHandlers) == 1 {
			g.group.Patch(path, fiberHandlers[0])
		} else {
			g.group.Patch(path, fiberHandlers[0], fiberHandlers[1:]...)
		}
	}
	return g
}

// Options registers a route for OPTIONS requests within this group
func (g *RouteGroup) Options(path string, handlers ...interface{}) contract.RouteGroup {
	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fnOld, ok := handler.(func(interface{}) error); ok { // Legacy generic handler
			fiberHandlers = append(fiberHandlers, func(c fiber.Ctx) error { // Wrapper now uses fiber.Ctx
				return fnOld(c) // c is fiber.Ctx (interface)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		if len(fiberHandlers) == 1 {
			g.group.Options(path, fiberHandlers[0])
		} else {
			g.group.Options(path, fiberHandlers[0], fiberHandlers[1:]...)
		}
	}
	return g
}

// Head registers a route for HEAD requests within this group
func (g *RouteGroup) Head(path string, handlers ...interface{}) contract.RouteGroup {
	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fnOld, ok := handler.(func(interface{}) error); ok { // Legacy generic handler
			fiberHandlers = append(fiberHandlers, func(c fiber.Ctx) error { // Wrapper now uses fiber.Ctx
				return fnOld(c) // c is fiber.Ctx (interface)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		if len(fiberHandlers) == 1 {
			g.group.Head(path, fiberHandlers[0])
		} else {
			g.group.Head(path, fiberHandlers[0], fiberHandlers[1:]...)
		}
	}
	return g
}

// All registers a route for all HTTP methods within this group
func (g *RouteGroup) All(path string, handlers ...interface{}) contract.RouteGroup {
	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fnOld, ok := handler.(func(interface{}) error); ok { // Legacy generic handler
			fiberHandlers = append(fiberHandlers, func(c fiber.Ctx) error { // Wrapper now uses fiber.Ctx
				return fnOld(c) // c is fiber.Ctx (interface)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		if len(fiberHandlers) == 1 {
			g.group.All(path, fiberHandlers[0])
		} else {
			g.group.All(path, fiberHandlers[0], fiberHandlers[1:]...)
		}
	}
	return g
}

// Group creates a new route group with prefix
func (g *RouteGroup) Group(prefix string, handlers ...interface{}) contract.RouteGroup {
	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fnOld, ok := handler.(func(interface{}) error); ok { // Legacy generic handler
			fiberHandlers = append(fiberHandlers, func(c fiber.Ctx) error { // Wrapper now uses fiber.Ctx
				return fnOld(c) // c is fiber.Ctx (interface)
			})
		}
	}

	group := g.group.Group(prefix, fiberHandlers...)
	return &RouteGroup{group: group}
}

// Use registers middleware for this group
func (g *RouteGroup) Use(handlers ...interface{}) contract.RouteGroup {
	finalHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, h := range handlers {
		if fiberHandler, ok := h.(fiber.Handler); ok { // fiber.Handler is func(fiber.Ctx) error in v3
			finalHandlers = append(finalHandlers, fiberHandler)
		}
		// Note: Http.Use only processes fiber.Handler. This RouteGroup.Use is now consistent.
	}

	for _, handler := range finalHandlers {
		g.group.Use(handler)
	}
	return g
}

// Provider provides an Http instance
func Provider() contract.Http {
	return New()
}
