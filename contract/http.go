package contract

import (
	"context"
	"github.com/gofiber/fiber/v3/middleware/static"

	"github.com/gofiber/fiber/v3"
)

type HttpHandler = func(ctx fiber.Ctx) error

type HttpErrorHandler = func(fiber.Ctx, error) error

// Http represents the HTTP server module interface based off Fiber
type Http interface {
	Module

	// Fiber returns the Fiber app instance (convenience method for the default implementation)
	Fiber() *fiber.App

	// Get registers a route for GET requests
	Get(path string, handler HttpHandler, middleware ...HttpHandler) Http

	// Post registers a route for POST requests
	Post(path string, handler HttpHandler, middleware ...HttpHandler) Http

	// Put registers a route for PUT requests
	Put(path string, handler HttpHandler, middleware ...HttpHandler) Http

	// Delete registers a route for DELETE requests
	Delete(path string, handler HttpHandler, middleware ...HttpHandler) Http

	// Patch registers a route for PATCH requests
	Patch(path string, handler HttpHandler, middleware ...HttpHandler) Http

	// Options registers a route for OPTIONS requests
	Options(path string, handler HttpHandler, middleware ...HttpHandler) Http

	// Head registers a route for HEAD requests
	Head(path string, handler HttpHandler, middleware ...HttpHandler) Http

	// All registers a route for all HTTP methods
	All(path string, handler HttpHandler, middleware ...HttpHandler) Http

	// Connect registers a route for CONNECT requests
	Connect(path string, handler HttpHandler, middleware ...HttpHandler) Http

	// Trace registers a route for TRACE requests
	Trace(path string, handler HttpHandler, middleware ...HttpHandler) Http

	// Add allows you to specify multiple HTTP methods to register a route.
	Add(methods []string, path string, handler HttpHandler, middleware ...HttpHandler) Http

	// Group creates a new route group with prefix
	Group(prefix string, handlers ...HttpHandler) RouteGroup

	// Use registers middleware
	Use(handlers ...any) Http

	// Static serves static files
	Static(prefix, root string, config ...static.Config) Http

	// Listen starts the HTTP server
	Listen(address string) error

	// Shutdown gracefully shuts down the HTTP server
	Shutdown(ctx context.Context) error
}

// RouteGroup represents a group of routes with a common prefix
type RouteGroup interface {
	// Get registers a route for GET requests
	Get(path string, handler HttpHandler, middleware ...HttpHandler)

	// Post registers a route for POST requests
	Post(path string, handler HttpHandler, middleware ...HttpHandler)

	// Put registers a route for PUT requests
	Put(path string, handler HttpHandler, middleware ...HttpHandler)

	// Delete registers a route for DELETE requests
	Delete(path string, handler HttpHandler, middleware ...HttpHandler)

	// Patch registers a route for PATCH requests
	Patch(path string, handler HttpHandler, middleware ...HttpHandler)

	// Options registers a route for OPTIONS requests
	Options(path string, handler HttpHandler, middleware ...HttpHandler)

	// Head registers a route for HEAD requests
	Head(path string, handler HttpHandler, middleware ...HttpHandler)

	// All registers a route for all HTTP methods
	All(path string, handler HttpHandler, middleware ...HttpHandler)

	// Connect registers a route for CONNECT requests
	Connect(path string, handler HttpHandler, middleware ...HttpHandler)

	// Trace registers a route for TRACE requests
	Trace(path string, handler HttpHandler, middleware ...HttpHandler)

	// Add allows you to specify multiple HTTP methods to register a route.
	Add(methods []string, path string, handler HttpHandler, middleware ...HttpHandler)

	// Group creates a new route group with prefix
	Group(prefix string, handlers ...HttpHandler) RouteGroup

	// Use registers middleware
	Use(handlers ...any)
}

// HttpProvider is a function that provides an Http instance
type HttpProvider func() Http
