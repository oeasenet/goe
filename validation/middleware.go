package validation

import (
	"errors"
	"reflect"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2/contract"
)

const (
	// DefaultValidatorContextKey is the default name used to store the validator in Fiber locals.
	DefaultValidatorContextKey = "validator"
	validatorContextKeyKey     = "_goe_validator_context_key"
)

// Config defines the config for validation middleware.
type Config struct {
	// Validator instance to use
	Validator *Validator

	// ErrorHandler is executed when validation fails
	ErrorHandler fiber.ErrorHandler

	// ContextKey is the key used to store validator in context
	ContextKey string
}

// ConfigDefault is the default config.
var ConfigDefault = Config{
	Validator:    nil,
	ErrorHandler: defaultErrorHandler,
	ContextKey:   DefaultValidatorContextKey,
}

// defaultErrorHandler handles validation errors.
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

// NewMiddleware creates a new validation middleware.
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

	contextKey := cfg.ContextKey
	if contextKey == "" {
		contextKey = DefaultValidatorContextKey
	}

	// Return middleware handler
	return func(c fiber.Ctx) error {
		// Store validator in context for use in handlers
		c.Locals(contextKey, cfg.Validator)
		c.Locals(validatorContextKeyKey, contextKey)

		// Continue to next middleware
		return c.Next()
	}
}

// GetValidator retrieves the validator from context.
func GetValidator(c fiber.Ctx) *Validator {
	storedKey, ok := c.Locals(validatorContextKeyKey).(string)
	if !ok || storedKey == "" {
		storedKey = DefaultValidatorContextKey
	}

	if v, ok := c.Locals(storedKey).(*Validator); ok {
		return v
	}
	return nil
}

// ValidateBody is a middleware that validates request body.
// A fresh instance of the dst type is allocated per request to prevent
// concurrent requests from sharing and overwriting the same struct.
func ValidateBody(dst any) fiber.Handler {
	dstType := reflect.TypeOf(dst)
	if dstType.Kind() == reflect.Pointer {
		dstType = dstType.Elem()
	}

	return func(c fiber.Ctx) error {
		v := GetValidator(c)
		if v == nil {
			v = New()
		}

		// Allocate a fresh struct per request to avoid shared-pointer data races
		reqDst := reflect.New(dstType).Interface()

		if err := v.ValidateRequest(c, reqDst); err != nil {
			return defaultErrorHandler(c, err)
		}

		// Store validated data in context
		c.Locals("validatedBody", reqDst)

		return c.Next()
	}
}

// ValidateQuery is a middleware that validates query parameters.
// A fresh instance of the dst type is allocated per request to prevent
// concurrent requests from sharing and overwriting the same struct.
func ValidateQuery(dst any) fiber.Handler {
	dstType := reflect.TypeOf(dst)
	if dstType.Kind() == reflect.Pointer {
		dstType = dstType.Elem()
	}

	return func(c fiber.Ctx) error {
		v := GetValidator(c)
		if v == nil {
			v = New()
		}

		reqDst := reflect.New(dstType).Interface()

		if err := v.ValidateQuery(c, reqDst); err != nil {
			return defaultErrorHandler(c, err)
		}

		// Store validated data in context
		c.Locals("validatedQuery", reqDst)

		return c.Next()
	}
}

// ValidateParams is a middleware that validates URL parameters.
// A fresh instance of the dst type is allocated per request to prevent
// concurrent requests from sharing and overwriting the same struct.
func ValidateParams(dst any) fiber.Handler {
	dstType := reflect.TypeOf(dst)
	if dstType.Kind() == reflect.Pointer {
		dstType = dstType.Elem()
	}

	return func(c fiber.Ctx) error {
		v := GetValidator(c)
		if v == nil {
			v = New()
		}

		reqDst := reflect.New(dstType).Interface()

		if err := v.ValidateParams(c, reqDst); err != nil {
			return defaultErrorHandler(c, err)
		}

		// Store validated data in context
		c.Locals("validatedParams", reqDst)

		return c.Next()
	}
}

// GetValidatedBody retrieves validated body from context
func GetValidatedBody(c fiber.Ctx) any {
	return c.Locals("validatedBody")
}

// GetValidatedQuery retrieves validated query from context
func GetValidatedQuery(c fiber.Ctx) any {
	return c.Locals("validatedQuery")
}

// GetValidatedParams retrieves validated params from context
func GetValidatedParams(c fiber.Ctx) any {
	return c.Locals("validatedParams")
}

// MiddlewareProvider implements the contract.ValidationMiddleware interface
type MiddlewareProvider struct{}

// NewMiddlewareProvider creates a new middleware provider
func NewMiddlewareProvider() *MiddlewareProvider {
	return &MiddlewareProvider{}
}

// ValidateBody creates middleware that validates request body
func (mp *MiddlewareProvider) ValidateBody(dst any) fiber.Handler {
	return ValidateBody(dst)
}

// ValidateQuery creates middleware that validates query parameters
func (mp *MiddlewareProvider) ValidateQuery(dst any) fiber.Handler {
	return ValidateQuery(dst)
}

// ValidateParams creates middleware that validates URL parameters
func (mp *MiddlewareProvider) ValidateParams(dst any) fiber.Handler {
	return ValidateParams(dst)
}

// Compile-time interface compliance check
var _ contract.ValidationMiddleware = (*MiddlewareProvider)(nil)
