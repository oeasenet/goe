package contract

import (
	"github.com/gofiber/fiber/v3"
)

// HTTPKernel defines the HTTP kernel interface
type HTTPKernel interface {
	// App returns the underlying Fiber app
	App() *fiber.App

	// Listen starts the HTTP server
	Listen(addr string) error

	// Shutdown gracefully shuts down the server
	Shutdown() error
}

// Router defines the HTTP router interface
// We use fiber.Router directly as it's well-standardized
type Router interface {
	fiber.Router
}

// Context represents the HTTP request context
// We use fiber.Ctx directly as it's the standard in Fiber
type Context = fiber.Ctx

// Handler represents an HTTP handler function
// We use fiber.Handler directly
type Handler = fiber.Handler

// Middleware represents HTTP middleware
// We use fiber.Handler for middleware as well
type Middleware = fiber.Handler

// ErrorHandler handles HTTP errors
type ErrorHandler = fiber.ErrorHandler

// HTTPConfig represents HTTP configuration
type HTTPConfig struct {
	// Host to bind the server to
	Host string

	// Port to bind the server to
	Port int

	// Prefork enables use of SO_REUSEPORT socket option
	Prefork bool

	// ServerHeader sets the Server header
	ServerHeader string

	// StrictRouting enables strict routing
	StrictRouting bool

	// CaseSensitive enables case sensitive routing
	CaseSensitive bool

	// BodyLimit sets the maximum allowed size for a request body
	BodyLimit int

	// ReadTimeout is the amount of time allowed to read the full request
	ReadTimeout string

	// WriteTimeout is the amount of time allowed to write the full response
	WriteTimeout string

	// IdleTimeout is the maximum amount of time to wait for the next request
	IdleTimeout string

	// TrustedProxies contains the list of trusted proxy IPs
	TrustedProxies []string
}

// RouteInfo contains information about a registered route
type RouteInfo struct {
	Method  string
	Path    string
	Name    string
	Handler Handler
}
