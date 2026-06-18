package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/validation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

// ContextKey represents keys used in fiber context locals
type ContextKey string

const (
	// ServicesKey is the key for DI services in context
	ServicesKey ContextKey = "services"
)

// Services holds all the services that can be injected into handlers
type Services struct {
	App       contract.Application
	Config    contract.Config
	Logger    contract.Logger
	Validator *validation.Validator
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

// WithReqCtx returns a request-scoped logger: the application logger enriched
// with request_id (from the requestid middleware or the raw header) and, when an
// OpenTelemetry span is active, trace_id/span_id. Call it at the top of a handler
// so every subsequent log line is correlated to the request.
func WithReqCtx(c fiber.Ctx) contract.Logger {
	logger := GetServices(c).Logger
	if logger == nil {
		return logger
	}
	if kv := requestLogFields(c); len(kv) > 0 {
		logger = logger.With(kv...)
	}
	return logger
}

// GetLogger is a backward-compatible alias for WithReqCtx.
func GetLogger(c fiber.Ctx) contract.Logger {
	return WithReqCtx(c)
}

// requestLogFields returns the request-scoped log fields for c: request_id, plus
// trace_id/span_id when a valid OpenTelemetry span is in the context.
func requestLogFields(c fiber.Ctx) []any {
	var kv []any
	if rid := requestIDFrom(c, requestIDHeader(c)); rid != "" {
		kv = append(kv, "request_id", rid)
	}
	if sc := trace.SpanContextFromContext(c.Context()); sc.IsValid() {
		kv = append(kv, "trace_id", sc.TraceID().String(), "span_id", sc.SpanID().String())
	}
	return kv
}

// requestIDFrom returns the request id from the requestid middleware, falling
// back to the raw header so an upstream id is captured even when the middleware
// is disabled. An empty header defaults to X-Request-ID.
func requestIDFrom(c fiber.Ctx, header string) string {
	if rid := requestid.FromContext(c); rid != "" {
		return rid
	}
	if header == "" {
		header = fiber.HeaderXRequestID
	}
	return c.Get(header)
}

// requestIDHeader returns the configured request-id header (HTTP_REQUEST_ID_HEADER),
// defaulting to X-Request-ID.
func requestIDHeader(c fiber.Ctx) string {
	if cfg := GetServices(c).Config; cfg != nil {
		if h := cfg.GetString("HTTP_REQUEST_ID_HEADER"); h != "" {
			return h
		}
	}
	return fiber.HeaderXRequestID
}

// requestIDEnabled reports whether GOE should register the default requestid
// middleware. It is on unless HTTP_REQUEST_ID is explicitly set to false.
func requestIDEnabled(config contract.Config) bool {
	return !config.Has("HTTP_REQUEST_ID") || config.GetBool("HTTP_REQUEST_ID")
}

// GetApp retrieves application from the context
func GetApp(c fiber.Ctx) contract.Application {
	return GetServices(c).App
}

// GetValidator retrieves validator from the context
func GetValidator(c fiber.Ctx) *validation.Validator {
	return GetServices(c).Validator
}

// GetCache retrieves cache from the context
func GetCache(c fiber.Ctx) contract.Cache {
	return GetServices(c).Cache
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

	App       contract.Application
	Config    contract.Config
	Logger    contract.Logger
	Cache     contract.Cache        `optional:"true"`
	Validator *validation.Validator `optional:"true"`
}

// CreateServiceMiddleware creates a middleware that injects services into the context
func CreateServiceMiddleware(provider ServiceProvider) fiber.Handler {
	services := Services{
		App:       provider.App,
		Config:    provider.Config,
		Logger:    provider.Logger,
		Cache:     provider.Cache,
		Validator: provider.Validator,
	}

	provider.Logger.Debug("CreateServiceMiddleware called")

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
