package contract

import "github.com/gofiber/fiber/v3"

// HTTPValidator defines the interface for HTTP request validation
type HTTPValidator interface {
	// Validate validates a struct according to its tags
	Validate(i interface{}) error

	// ValidateRequest validates and parses JSON request body
	ValidateRequest(c fiber.Ctx, dst interface{}) error

	// ValidateQuery validates and parses query parameters
	ValidateQuery(c fiber.Ctx, dst interface{}) error

	// ValidateParams validates and parses URL parameters
	ValidateParams(c fiber.Ctx, dst interface{}) error

	// ValidateHeaders validates and parses request headers
	ValidateHeaders(c fiber.Ctx, dst interface{}) error

	// ValidateForm validates and parses form data
	ValidateForm(c fiber.Ctx, dst interface{}) error
}

// ValidationProvider defines the interface for validation providers
// This allows for dependency injection of validation services
type ValidationProvider interface {
	// GetValidator returns an HTTP validator instance
	GetValidator() HTTPValidator
}

// ValidationMiddleware defines the interface for validation middleware
type ValidationMiddleware interface {
	// ValidateBody creates middleware that validates request body
	ValidateBody(dst interface{}) fiber.Handler

	// ValidateQuery creates middleware that validates query parameters
	ValidateQuery(dst interface{}) fiber.Handler

	// ValidateParams creates middleware that validates URL parameters
	ValidateParams(dst interface{}) fiber.Handler
}
