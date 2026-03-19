package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2/contract"
)

// Validator wraps go-playground/validator for HTTP request validation
type Validator struct {
	validator *validator.Validate
}

// New creates a new validator instance with default configuration
func New() *Validator {
	v := validator.New()

	// Register custom tag name function to use json tags
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register common custom validators
	registerCustomValidators(v)

	return &Validator{
		validator: v,
	}
}

// Validate validates a struct according to its tags
func (v *Validator) Validate(i any) error {
	if err := v.validator.Struct(i); err != nil {
		return NewValidationError(err)
	}
	return nil
}

// ValidateVar validates a single variable against a tag
func (v *Validator) ValidateVar(field any, tag string) error {
	return v.validator.Var(field, tag)
}

// RegisterValidation registers a custom validation function
func (v *Validator) RegisterValidation(tag string, fn validator.Func) error {
	return v.validator.RegisterValidation(tag, fn)
}

// RegisterAlias registers an alias for a validation tag
func (v *Validator) RegisterAlias(alias, tags string) {
	v.validator.RegisterAlias(alias, tags)
}

// RegisterStructValidation registers a custom struct validation function
func (v *Validator) RegisterStructValidation(fn validator.StructLevelFunc, types ...any) {
	v.validator.RegisterStructValidation(fn, types...)
}

// GetValidator returns the underlying validator instance for advanced usage
func (v *Validator) GetValidator() *validator.Validate {
	return v.validator
}

// ValidateRequest validates request body in Fiber context.
//
// Note: If StructValidator is set in Fiber config (goe does this by default),
// c.Bind().Body() already validates automatically. In that case, prefer using
// c.Bind().Body(&dst) directly in your handler instead of this method.
func (v *Validator) ValidateRequest(c fiber.Ctx, dst any) error {
	// c.Bind().Body() parses and validates (via StructValidator) in one step.
	if err := c.Bind().Body(dst); err != nil {
		return classifyBindError(err)
	}
	return nil
}

// ValidateQuery validates query parameters in Fiber context.
//
// Note: If StructValidator is set in Fiber config (goe does this by default),
// c.Bind().Query() already validates automatically. In that case, prefer using
// c.Bind().Query(&dst) directly in your handler instead of this method.
func (v *Validator) ValidateQuery(c fiber.Ctx, dst any) error {
	if err := c.Bind().Query(dst); err != nil {
		return classifyBindError(err)
	}
	return nil
}

// ValidateParams validates URL parameters in Fiber context.
//
// Note: If StructValidator is set in Fiber config (goe does this by default),
// c.Bind().URI() already validates automatically. In that case, prefer using
// c.Bind().URI(&dst) directly in your handler instead of this method.
func (v *Validator) ValidateParams(c fiber.Ctx, dst any) error {
	if err := c.Bind().URI(dst); err != nil {
		return classifyBindError(err)
	}
	return nil
}

// ValidateHeaders validates request headers in Fiber context.
//
// Note: If StructValidator is set in Fiber config (goe does this by default),
// c.Bind().Header() already validates automatically. In that case, prefer using
// c.Bind().Header(&dst) directly in your handler instead of this method.
func (v *Validator) ValidateHeaders(c fiber.Ctx, dst any) error {
	if err := c.Bind().Header(dst); err != nil {
		return classifyBindError(err)
	}
	return nil
}

// ValidateForm validates form data in Fiber context.
//
// Note: If StructValidator is set in Fiber config (goe does this by default),
// c.Bind().Form() already validates automatically. In that case, prefer using
// c.Bind().Form(&dst) directly in your handler instead of this method.
func (v *Validator) ValidateForm(c fiber.Ctx, dst any) error {
	if err := c.Bind().Form(dst); err != nil {
		return classifyBindError(err)
	}
	return nil
}

// classifyBindError classifies errors from Fiber's bind methods into
// validation errors or parse errors for proper error typing.
func classifyBindError(err error) error {
	// Already our validation error type (returned by Validator.Validate via StructValidator)
	var ve Error
	if errors.As(err, &ve) {
		return ve
	}
	// Raw validator errors (in case StructValidator wrapping is bypassed)
	var valErrs validator.ValidationErrors
	if errors.As(err, &valErrs) {
		return NewValidationError(err)
	}
	return NewParseError(err)
}

// registerCustomValidators registers common custom validators
func registerCustomValidators(v *validator.Validate) {
	// Phone number validation
	_ = v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		// Simple phone validation - can be enhanced
		if len(phone) < 10 || len(phone) > 15 {
			return false
		}
		for _, r := range phone {
			if r != '+' && r != '-' && r != ' ' && (r < '0' || r > '9') {
				return false
			}
		}
		return true
	})

	// Username validation (alphanumeric and underscore)
	_ = v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		username := fl.Field().String()
		if len(username) < 3 || len(username) > 30 {
			return false
		}
		for _, r := range username {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
				return false
			}
		}
		return true
	})

	// Password strength validation
	_ = v.RegisterValidation("strong_password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		if len(password) < 8 {
			return false
		}

		var hasUpper, hasLower, hasNumber, hasSpecial bool
		for _, r := range password {
			switch {
			case r >= 'A' && r <= 'Z':
				hasUpper = true
			case r >= 'a' && r <= 'z':
				hasLower = true
			case r >= '0' && r <= '9':
				hasNumber = true
			case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", r):
				hasSpecial = true
			}
		}

		return hasUpper && hasLower && hasNumber && hasSpecial
	})
}

// Helper functions for common validation scenarios

// IsEmail validates if a string is a valid email
func IsEmail(email string) bool {
	v := validator.New()
	err := v.Var(email, "required,email")
	return err == nil
}

// IsURL validates if a string is a valid URL
func IsURL(url string) bool {
	v := validator.New()
	err := v.Var(url, "required,url")
	return err == nil
}

// IsUUID validates if a string is a valid UUID
func IsUUID(uuid string) bool {
	v := validator.New()
	err := v.Var(uuid, "required,uuid")
	return err == nil
}

// IsAlpha validates if a string contains only alphabetic characters
func IsAlpha(str string) bool {
	v := validator.New()
	err := v.Var(str, "required,alpha")
	return err == nil
}

// IsAlphanumeric validates if a string contains only alphanumeric characters
func IsAlphanumeric(str string) bool {
	v := validator.New()
	err := v.Var(str, "required,alphanum")
	return err == nil
}

// IsNumeric validates if a string contains only numeric characters (0-9)
func IsNumeric(str string) bool {
	if str == "" {
		return false
	}
	for _, r := range str {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// ValidateStruct is a convenience function for one-off struct validation
func ValidateStruct(s any) error {
	v := New()
	return v.Validate(s)
}

// MustValidate validates a struct and panics if validation fails
func MustValidate(s any) {
	if err := ValidateStruct(s); err != nil {
		panic(fmt.Sprintf("validation failed: %v", err))
	}
}

// Provider implements the contract.ValidationProvider interface
type Provider struct {
	validator *Validator
}

// NewProvider creates a new validation provider
func NewProvider() *Provider {
	return &Provider{
		validator: New(),
	}
}

// GetValidator returns the validator instance
func (p *Provider) GetValidator() contract.HTTPValidator {
	return p.validator
}

// Compile-time interface compliance checks
var (
	_ contract.HTTPValidator      = (*Validator)(nil)
	_ contract.ValidationProvider = (*Provider)(nil)
)
