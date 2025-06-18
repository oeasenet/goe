package http

import (
	"github.com/go-playground/validator/v10"
	"go.oease.dev/goe/v2/contract"
)

// CustomValidator wraps go-playground/validator
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator creates a new custom validator
func NewValidator() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}

// Validate validates a struct
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// RegisterValidation registers a custom validation function
func (cv *CustomValidator) RegisterValidation(tag string, fn validator.Func) error {
	return cv.validator.RegisterValidation(tag, fn)
}

// RegisterAlias registers an alias for a validation tag
func (cv *CustomValidator) RegisterAlias(alias, tags string) {
	cv.validator.RegisterAlias(alias, tags)
}

// RegisterStructValidation registers a custom struct validation function
func (cv *CustomValidator) RegisterStructValidation(fn validator.StructLevelFunc, types ...interface{}) {
	cv.validator.RegisterStructValidation(fn, types...)
}

// GetValidator returns the underlying validator instance for advanced usage
func (cv *CustomValidator) GetValidator() *validator.Validate {
	return cv.validator
}

// ValidatorProvider provides the custom validator for dependency injection
type ValidatorProvider struct {
	validator *CustomValidator
	logger    contract.Logger
}

// NewValidatorProvider creates a new validator provider
func NewValidatorProvider(logger contract.Logger) *ValidatorProvider {
	v := NewValidator()

	// Log validator creation
	logger.Debug("Created custom validator for HTTP module")

	return &ValidatorProvider{
		validator: v,
		logger:    logger,
	}
}

// Provide returns the validator for Fx
func (vp *ValidatorProvider) Provide() *CustomValidator {
	return vp.validator
}

// RegisterCustomValidation allows modules to register custom validations
func (vp *ValidatorProvider) RegisterCustomValidation(tag string, fn validator.Func) error {
	if err := vp.validator.RegisterValidation(tag, fn); err != nil {
		vp.logger.Error("Failed to register custom validation",
			"tag", tag,
			"error", err.Error(),
		)
		return err
	}

	vp.logger.Debug("Registered custom validation",
		"tag", tag,
	)
	return nil
}
