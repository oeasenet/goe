package http

import (
	"context"
	"sync"

	"github.com/gofiber/fiber/v2"
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
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		h.app.Get(path, fiberHandlers...)
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
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		h.app.Post(path, fiberHandlers...)
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
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		h.app.Put(path, fiberHandlers...)
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
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		h.app.Delete(path, fiberHandlers...)
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
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		h.app.Patch(path, fiberHandlers...)
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
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		h.app.Options(path, fiberHandlers...)
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
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		h.app.Head(path, fiberHandlers...)
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
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		h.app.All(path, fiberHandlers...)
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
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
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

	// Convert config to fiber.Static
	var fiberConfig fiber.Static
	if len(config) > 0 {
		if staticConfig, ok := config[0].(fiber.Static); ok {
			fiberConfig = staticConfig
		}
	}

	h.app.Static(prefix, root, fiberConfig)
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
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		g.group.Get(path, fiberHandlers...)
	}
	return g
}

// Post registers a route for POST requests within this group
func (g *RouteGroup) Post(path string, handlers ...interface{}) contract.RouteGroup {
	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		g.group.Post(path, fiberHandlers...)
	}
	return g
}

// Put registers a route for PUT requests within this group
func (g *RouteGroup) Put(path string, handlers ...interface{}) contract.RouteGroup {
	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		g.group.Put(path, fiberHandlers...)
	}
	return g
}

// Delete registers a route for DELETE requests within this group
func (g *RouteGroup) Delete(path string, handlers ...interface{}) contract.RouteGroup {
	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		g.group.Delete(path, fiberHandlers...)
	}
	return g
}

// Patch registers a route for PATCH requests within this group
func (g *RouteGroup) Patch(path string, handlers ...interface{}) contract.RouteGroup {
	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		g.group.Patch(path, fiberHandlers...)
	}
	return g
}

// Options registers a route for OPTIONS requests within this group
func (g *RouteGroup) Options(path string, handlers ...interface{}) contract.RouteGroup {
	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		g.group.Options(path, fiberHandlers...)
	}
	return g
}

// Head registers a route for HEAD requests within this group
func (g *RouteGroup) Head(path string, handlers ...interface{}) contract.RouteGroup {
	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		g.group.Head(path, fiberHandlers...)
	}
	return g
}

// All registers a route for all HTTP methods within this group
func (g *RouteGroup) All(path string, handlers ...interface{}) contract.RouteGroup {
	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
			})
		}
	}

	if len(fiberHandlers) > 0 {
		g.group.All(path, fiberHandlers...)
	}
	return g
}

// Group creates a new route group with prefix
func (g *RouteGroup) Group(prefix string, handlers ...interface{}) contract.RouteGroup {
	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
			})
		}
	}

	group := g.group.Group(prefix, fiberHandlers...)
	return &RouteGroup{group: group}
}

// Use registers middleware for this group
func (g *RouteGroup) Use(handlers ...interface{}) contract.RouteGroup {
	// Convert handlers to fiber.Handler
	fiberHandlers := make([]fiber.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if fiberHandler, ok := handler.(fiber.Handler); ok {
			fiberHandlers = append(fiberHandlers, fiberHandler)
		} else if fn, ok := handler.(func(interface{}) error); ok {
			// Convert func(interface{}) error to fiber.Handler
			fiberHandlers = append(fiberHandlers, func(c *fiber.Ctx) error {
				return fn(c)
			})
		}
	}

	for _, handler := range fiberHandlers {
		g.group.Use(handler)
	}
	return g
}

// Provider provides an Http instance
func Provider() contract.Http {
	return New()
}
