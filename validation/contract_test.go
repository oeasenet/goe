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

}
