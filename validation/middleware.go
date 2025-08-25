package validation

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2/contract"
)

// Config defines the config for validation middleware
type Config struct {
	// Validator instance to use
	Validator *Validator

	// ErrorHandler is executed when validation fails
	ErrorHandler fiber.ErrorHandler

	// ContextKey is the key used to store validator in context
	ContextKey string
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	Validator:    nil,
	ErrorHandler: defaultErrorHandler,
	ContextKey:   "validator",
}

// defaultErrorHandler handles validation errors
func defaultErrorHandler(c fiber.Ctx, err error) error {
	// Check if it's a validation error
	var ve Error
	if errors.As(err, &ve) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": ve.Errors,
		})
	}

	// Check if it's a parse error
	var pe ParseError
	if errors.As(err, &pe) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": pe.Message,
		})
	}

	// Default error handling
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": err.Error(),
	})
}

// NewMiddleware New creates a new validation middleware
func NewMiddleware(config ...Config) fiber.Handler {
	// Set default config
	cfg := ConfigDefault

	// Override config if provided
	if len(config) > 0 {
		cfg = config[0]

		// Set default values if not provided
		if cfg.Validator == nil {
			cfg.Validator = New()
		}
		if cfg.ErrorHandler == nil {
			cfg.ErrorHandler = ConfigDefault.ErrorHandler
		}
		if cfg.ContextKey == "" {
			cfg.ContextKey = ConfigDefault.ContextKey
		}
	} else {
		cfg.Validator = New()
	}

	// Return middleware handler
	return func(c fiber.Ctx) error {
		// Store validator in context for use in handlers
		c.Locals(cfg.ContextKey, cfg.Validator)

		// Continue to next middleware
		return c.Next()
	}
}

// GetValidator retrieves the validator from context
func GetValidator(c fiber.Ctx) *Validator {
	if v, ok := c.Locals("validator").(*Validator); ok {
		return v
	}
	return nil
}

// ValidateBody is a middleware that validates request body
func ValidateBody(dst interface{}) fiber.Handler {
	return func(c fiber.Ctx) error {
		v := GetValidator(c)
		if v == nil {
			v = New()
		}

		if err := v.ValidateRequest(c, dst); err != nil {
			return defaultErrorHandler(c, err)
		}

		// Store validated data in context
		c.Locals("validatedBody", dst)

		return c.Next()
	}
}

// ValidateQuery is a middleware that validates query parameters
func ValidateQuery(dst interface{}) fiber.Handler {
	return func(c fiber.Ctx) error {
		v := GetValidator(c)
		if v == nil {
			v = New()
		}

		if err := v.ValidateQuery(c, dst); err != nil {
			return defaultErrorHandler(c, err)
		}

		// Store validated data in context
		c.Locals("validatedQuery", dst)

		return c.Next()
	}
}

// ValidateParams is a middleware that validates URL parameters
func ValidateParams(dst interface{}) fiber.Handler {
	return func(c fiber.Ctx) error {
		v := GetValidator(c)
		if v == nil {
			v = New()
		}

		if err := v.ValidateParams(c, dst); err != nil {
			return defaultErrorHandler(c, err)
		}

		// Store validated data in context
		c.Locals("validatedParams", dst)

		return c.Next()
	}
}

// GetValidatedBody retrieves validated body from context
func GetValidatedBody(c fiber.Ctx) interface{} {
	return c.Locals("validatedBody")
}

// GetValidatedQuery retrieves validated query from context
func GetValidatedQuery(c fiber.Ctx) interface{} {
	return c.Locals("validatedQuery")
}

// GetValidatedParams retrieves validated params from context
func GetValidatedParams(c fiber.Ctx) interface{} {
	return c.Locals("validatedParams")
}

// MiddlewareProvider implements the contract.ValidationMiddleware interface
type MiddlewareProvider struct{}

// NewMiddlewareProvider creates a new middleware provider
func NewMiddlewareProvider() *MiddlewareProvider {
	return &MiddlewareProvider{}
}

// ValidateBody creates middleware that validates request body
func (mp *MiddlewareProvider) ValidateBody(dst interface{}) fiber.Handler {
	return ValidateBody(dst)
}

// ValidateQuery creates middleware that validates query parameters
func (mp *MiddlewareProvider) ValidateQuery(dst interface{}) fiber.Handler {
	return ValidateQuery(dst)
}

// ValidateParams creates middleware that validates URL parameters
func (mp *MiddlewareProvider) ValidateParams(dst interface{}) fiber.Handler {
	return ValidateParams(dst)
}

// Compile-time interface compliance check
var _ contract.ValidationMiddleware = (*MiddlewareProvider)(nil)
