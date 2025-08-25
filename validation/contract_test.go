package validation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/validation"
)

func TestValidationContracts(t *testing.T) {
	t.Run("Validator implements HTTPValidator contract", func(t *testing.T) {
		validator := validation.New()

		// Test that validator implements the contract
		var _ contract.HTTPValidator = validator

		// Test basic validation functionality
		type TestStruct struct {
			Name  string `json:"name" validate:"required"`
			Email string `json:"email" validate:"required,email"`
		}

		valid := TestStruct{
			Name:  "John Doe",
			Email: "john@example.com",
		}

		err := validator.Validate(valid)
		assert.NoError(t, err)

		invalid := TestStruct{
			Name:  "",
			Email: "invalid-email",
		}

		err = validator.Validate(invalid)
		assert.Error(t, err)
	})

	t.Run("Provider implements ValidationProvider contract", func(t *testing.T) {
		provider := validation.NewProvider()

		// Test that provider implements the contract
		var _ contract.ValidationProvider = provider

		// Test that GetValidator returns HTTPValidator
		httpValidator := provider.GetValidator()
		assert.NotNil(t, httpValidator)

		// Test that returned validator works
		type TestStruct struct {
			Value string `json:"value" validate:"required"`
		}

		valid := TestStruct{Value: "test"}
		err := httpValidator.Validate(valid)
		assert.NoError(t, err)

		invalid := TestStruct{Value: ""}
		err = httpValidator.Validate(invalid)
		assert.Error(t, err)
	})

	t.Run("MiddlewareProvider implements ValidationMiddleware contract", func(t *testing.T) {
		middleware := validation.NewMiddlewareProvider()

		// Test that middleware implements the contract
		var _ contract.ValidationMiddleware = middleware

		// Test that middleware methods return handlers
		type TestStruct struct {
			Name string `json:"name" validate:"required"`
		}

		bodyHandler := middleware.ValidateBody(&TestStruct{})
		assert.NotNil(t, bodyHandler)

		queryHandler := middleware.ValidateQuery(&TestStruct{})
		assert.NotNil(t, queryHandler)

		paramsHandler := middleware.ValidateParams(&TestStruct{})
		assert.NotNil(t, paramsHandler)
	})
}
