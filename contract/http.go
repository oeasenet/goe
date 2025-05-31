package contract

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

// Http represents the HTTP server module interface
type Http interface {
	Module

	// Engine returns the underlying HTTP engine (Fiber in the default implementation)
	Engine() interface{}

	// Fiber returns the Fiber app instance (convenience method for the default implementation)
	Fiber() *fiber.App

	// Get registers a route for GET requests
	Get(path string, handlers ...interface{}) Http

	// Post registers a route for POST requests
	Post(path string, handlers ...interface{}) Http

	// Put registers a route for PUT requests
	Put(path string, handlers ...interface{}) Http

	// Delete registers a route for DELETE requests
	Delete(path string, handlers ...interface{}) Http

	// Patch registers a route for PATCH requests
	Patch(path string, handlers ...interface{}) Http

	// Options registers a route for OPTIONS requests
	Options(path string, handlers ...interface{}) Http

	// Head registers a route for HEAD requests
	Head(path string, handlers ...interface{}) Http

	// All registers a route for all HTTP methods
	All(path string, handlers ...interface{}) Http

	// Group creates a new route group with prefix
	Group(prefix string, handlers ...interface{}) RouteGroup

	// Use registers middleware
	Use(handlers ...interface{}) Http

	// Static serves static files
	Static(prefix, root string, config ...interface{}) Http

	// Listen starts the HTTP server
	Listen(address string) error

	// Shutdown gracefully shuts down the HTTP server
	Shutdown(ctx context.Context) error
}

// RouteGroup represents a group of routes with a common prefix
type RouteGroup interface {
	// Get registers a route for GET requests within this group
	Get(path string, handlers ...interface{}) RouteGroup

	// Post registers a route for POST requests within this group
	Post(path string, handlers ...interface{}) RouteGroup

	// Put registers a route for PUT requests within this group
	Put(path string, handlers ...interface{}) RouteGroup

	// Delete registers a route for DELETE requests within this group
	Delete(path string, handlers ...interface{}) RouteGroup

	// Patch registers a route for PATCH requests within this group
	Patch(path string, handlers ...interface{}) RouteGroup

	// Options registers a route for OPTIONS requests within this group
	Options(path string, handlers ...interface{}) RouteGroup

	// Head registers a route for HEAD requests within this group
	Head(path string, handlers ...interface{}) RouteGroup

	// All registers a route for all HTTP methods within this group
	All(path string, handlers ...interface{}) RouteGroup

	// Group creates a new route group with prefix
	Group(prefix string, handlers ...interface{}) RouteGroup

	// Use registers middleware for this group
	Use(handlers ...interface{}) RouteGroup
}

// HttpProvider is a function that provides an Http instance
type HttpProvider func() Http
