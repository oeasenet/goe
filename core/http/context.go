package http

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/fx"
)

// ContextKey represents keys used in fiber context locals
type ContextKey string

const (
	// RequestIDKey is the key for request ID in context
	RequestIDKey ContextKey = "requestID"

	// ServicesKey is the key for DI services in context
	ServicesKey ContextKey = "services"
)

// Services holds all the services that can be injected into handlers
type Services struct {
	App       contract.Application
	Config    contract.Config
	Logger    contract.Logger
	Validator *CustomValidator
	// Add more services as needed
	Cache contract.Cache
}

// InjectServices creates a middleware that injects services into the context
func InjectServices(services Services) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals(string(ServicesKey), services)
		return c.Next()
	}
}

// GetServices retrieves services from the context
func GetServices(c fiber.Ctx) Services {
	if services, ok := c.Locals(string(ServicesKey)).(Services); ok {
		return services
	}
	return Services{}
}

// GetConfig retrieves config from the context
func GetConfig(c fiber.Ctx) contract.Config {
	return GetServices(c).Config
}

// GetLogger retrieves logger from the context
func GetLogger(c fiber.Ctx) contract.Logger {
	services := GetServices(c)
	logger := services.Logger

	// Add request ID to logger if available
	if requestID, ok := c.Locals(string(RequestIDKey)).(string); ok && requestID != "" {
		logger = logger.With("request_id", requestID)
	}

	return logger
}

// GetApp retrieves application from the context
func GetApp(c fiber.Ctx) contract.Application {
	return GetServices(c).App
}

// GetValidator retrieves validator from the context
func GetValidator(c fiber.Ctx) *CustomValidator {
	return GetServices(c).Validator
}

// Handler creates a handler with dependency injection
// This is a more convenient way to create handlers with DI
type Handler[T any] func(c fiber.Ctx, deps T) error

// AsHandler converts a Handler[T] to a fiber.Handler
func AsHandler[T any](h Handler[T], deps T) fiber.Handler {
	return func(c fiber.Ctx) error {
		return h(c, deps)
	}
}

// ServiceProvider is used to provide services to the HTTP module
type ServiceProvider struct {
	fx.In

	App    contract.Application
	Config contract.Config
	Logger contract.Logger
	Cache  contract.Cache
}

// CreateServiceMiddleware creates a middleware that injects services
func CreateServiceMiddleware(provider ServiceProvider) fiber.Handler {
	services := Services{
		App:    provider.App,
		Config: provider.Config,
		Logger: provider.Logger,
		Cache:  provider.Cache,
	}
	return InjectServices(services)
}

// Group represents a route group with DI support
type Group struct {
	fiber.Router
	services Services
}

// NewGroup creates a new route group with DI support
func NewGroup(router fiber.Router, services Services) *Group {
	return &Group{
		Router:   router,
		services: services,
	}
}

// Handle registers a route with dependency injection
func (g *Group) Handle(method, path string, handler Handler[Services]) fiber.Router {
	return g.Add([]string{method}, path, AsHandler(handler, g.services))
}

// GET registers a GET route with dependency injection
func (g *Group) GET(path string, handler Handler[Services]) fiber.Router {
	return g.Handle("GET", path, handler)
}

// POST registers a POST route with dependency injection
func (g *Group) POST(path string, handler Handler[Services]) fiber.Router {
	return g.Handle("POST", path, handler)
}

// PUT registers a PUT route with dependency injection
func (g *Group) PUT(path string, handler Handler[Services]) fiber.Router {
	return g.Handle("PUT", path, handler)
}

// DELETE registers a DELETE route with dependency injection
func (g *Group) DELETE(path string, handler Handler[Services]) fiber.Router {
	return g.Handle("DELETE", path, handler)
}

// PATCH registers a PATCH route with dependency injection
func (g *Group) PATCH(path string, handler Handler[Services]) fiber.Router {
	return g.Handle("PATCH", path, handler)
}
