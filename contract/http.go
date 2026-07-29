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

	// Validator returns the struct validator for request validation (legacy)
	Validator() any

	// HTTPValidator returns the HTTP validator interface
	HTTPValidator() HTTPValidator
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

// HTTP server configuration lives in the core/http package as Options, which
// are applied over the FIBER_*/HTTP_* environment variables. See
// http.New and the With* constructors there.
