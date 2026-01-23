package validation

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError_Error(t *testing.T) {
	t.Run("single field error", func(t *testing.T) {
		err := Error{
			Errors: []FieldError{
				{Field: "name", Message: "name is required"},
			},
		}
		result := err.Error()
		assert.Contains(t, result, "validation failed")
		assert.Contains(t, result, "name: name is required")
	})

	t.Run("multiple field errors", func(t *testing.T) {
		err := Error{
			Errors: []FieldError{
				{Field: "name", Message: "name is required"},
				{Field: "email", Message: "email must be a valid email address"},
			},
		}
		result := err.Error()
		assert.Contains(t, result, "validation failed")
		assert.Contains(t, result, "name: name is required")
		assert.Contains(t, result, "email: email must be a valid email address")
	})

	t.Run("empty errors", func(t *testing.T) {
		err := Error{Errors: []FieldError{}}
		result := err.Error()
		assert.Equal(t, "validation failed: ", result)
	})
}

func TestNewValidationError(t *testing.T) {
	v := validator.New()

	t.Run("converts validator errors to Error", func(t *testing.T) {
		type TestStruct struct {
			Name  string `validate:"required"`
			Email string `validate:"required,email"`
		}

		err := v.Struct(TestStruct{Name: "", Email: "invalid"})
		require.Error(t, err)

		validationErr := NewValidationError(err)
		require.NotNil(t, validationErr)

		var verr Error
		require.ErrorAs(t, validationErr, &verr)
		assert.Len(t, verr.Errors, 2)
	})

	t.Run("handles non-validator errors gracefully", func(t *testing.T) {
		err := errors.New("some other error")
		validationErr := NewValidationError(err)

		var verr Error
		require.ErrorAs(t, validationErr, &verr)
		// Should have empty errors slice when not validator.ValidationErrors
		assert.Empty(t, verr.Errors)
	})
}

func TestParseError_Error(t *testing.T) {
	err := ParseError{Message: "invalid JSON format"}
	result := err.Error()

	assert.Equal(t, "parse error: invalid JSON format", result)
}

func TestNewParseError(t *testing.T) {
	originalErr := errors.New("unexpected EOF")
	parseErr := NewParseError(originalErr)

	var perr ParseError
	require.ErrorAs(t, parseErr, &perr)
	assert.Equal(t, "unexpected EOF", perr.Message)
}

func TestGetErrorMessage(t *testing.T) {
	v := validator.New()

	// Helper to get FieldError from validation
	getFieldError := func(tag, param string) validator.FieldError {
		type TestStruct struct {
			Field string `validate:"TAG"`
		}

		// Create dynamic struct with the tag we want to test
		type DynamicStruct struct {
			Field string
		}

		var fe validator.FieldError

		switch tag {
		case "required":
			type S struct {
				Field string `validate:"required"`
			}
			if errs := v.Struct(S{Field: ""}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "email":
			type S struct {
				Field string `validate:"email"`
			}
			if errs := v.Struct(S{Field: "invalid"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "min":
			type S struct {
				Field string `validate:"min=5"`
			}
			if errs := v.Struct(S{Field: "abc"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "max":
			type S struct {
				Field string `validate:"max=3"`
			}
			if errs := v.Struct(S{Field: "toolong"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "len":
			type S struct {
				Field string `validate:"len=5"`
			}
			if errs := v.Struct(S{Field: "abc"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "gt":
			type S struct {
				Field int `validate:"gt=10"`
			}
			if errs := v.Struct(S{Field: 5}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "gte":
			type S struct {
				Field int `validate:"gte=10"`
			}
			if errs := v.Struct(S{Field: 5}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "lt":
			type S struct {
				Field int `validate:"lt=10"`
			}
			if errs := v.Struct(S{Field: 15}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "lte":
			type S struct {
				Field int `validate:"lte=10"`
			}
			if errs := v.Struct(S{Field: 15}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "alpha":
			type S struct {
				Field string `validate:"alpha"`
			}
			if errs := v.Struct(S{Field: "abc123"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "alphanum":
			type S struct {
				Field string `validate:"alphanum"`
			}
			if errs := v.Struct(S{Field: "abc-123"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "numeric":
			type S struct {
				Field string `validate:"numeric"`
			}
			if errs := v.Struct(S{Field: "abc"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "url":
			type S struct {
				Field string `validate:"url"`
			}
			if errs := v.Struct(S{Field: "not-a-url"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "uri":
			type S struct {
				Field string `validate:"uri"`
			}
			if errs := v.Struct(S{Field: ":::invalid"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "uuid":
			type S struct {
				Field string `validate:"uuid"`
			}
			if errs := v.Struct(S{Field: "not-a-uuid"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "uuid4":
			type S struct {
				Field string `validate:"uuid4"`
			}
			if errs := v.Struct(S{Field: "not-a-uuid"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "ascii":
			type S struct {
				Field string `validate:"ascii"`
			}
			if errs := v.Struct(S{Field: "日本語"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "contains":
			type S struct {
				Field string `validate:"contains=test"`
			}
			if errs := v.Struct(S{Field: "hello"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "containsany":
			type S struct {
				Field string `validate:"containsany=!@#"`
			}
			if errs := v.Struct(S{Field: "hello"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "excludes":
			type S struct {
				Field string `validate:"excludes=bad"`
			}
			if errs := v.Struct(S{Field: "this is bad"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "excludesall":
			type S struct {
				Field string `validate:"excludesall=!@#"`
			}
			if errs := v.Struct(S{Field: "hello!"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "oneof":
			type S struct {
				Field string `validate:"oneof=red green blue"`
			}
			if errs := v.Struct(S{Field: "yellow"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "ip":
			type S struct {
				Field string `validate:"ip"`
			}
			if errs := v.Struct(S{Field: "not-an-ip"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "ipv4":
			type S struct {
				Field string `validate:"ipv4"`
			}
			if errs := v.Struct(S{Field: "not-an-ip"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "ipv6":
			type S struct {
				Field string `validate:"ipv6"`
			}
			if errs := v.Struct(S{Field: "not-an-ip"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "json":
			type S struct {
				Field string `validate:"json"`
			}
			if errs := v.Struct(S{Field: "not-json"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "base64":
			type S struct {
				Field string `validate:"base64"`
			}
			if errs := v.Struct(S{Field: "not@base64!"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "hexcolor":
			type S struct {
				Field string `validate:"hexcolor"`
			}
			if errs := v.Struct(S{Field: "not-hex"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "latitude":
			type S struct {
				Field float64 `validate:"latitude"`
			}
			if errs := v.Struct(S{Field: 200.0}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "longitude":
			type S struct {
				Field float64 `validate:"longitude"`
			}
			if errs := v.Struct(S{Field: 400.0}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "hostname":
			type S struct {
				Field string `validate:"hostname"`
			}
			if errs := v.Struct(S{Field: "invalid..hostname"}); errs != nil {
				fe = errs.(validator.ValidationErrors)[0]
			}
		case "unknown_tag":
			// This won't produce a validation error through the validator
			// We need to test the default case differently
			return nil
		}
		return fe
	}

	tests := []struct {
		tag         string
		wantContain string
	}{
		{"required", "is required"},
		{"email", "must be a valid email address"},
		{"min", "must be at least"},
		{"max", "must be at most"},
		{"len", "must be exactly"},
		{"gt", "must be greater than"},
		{"gte", "must be greater than or equal to"},
		{"lt", "must be less than"},
		{"lte", "must be less than or equal to"},
		{"alpha", "must contain only alphabetic characters"},
		{"alphanum", "must contain only alphanumeric characters"},
		{"numeric", "must be a valid number"},
		{"url", "must be a valid URL"},
		{"uri", "must be a valid URI"},
		{"uuid", "must be a valid UUID"},
		{"uuid4", "must be a valid UUID v4"},
		{"ascii", "must contain only ASCII characters"},
		{"contains", "must contain"},
		{"containsany", "must contain at least one of"},
		{"excludes", "must not contain"},
		{"excludesall", "must not contain any of"},
		{"oneof", "must be one of"},
		{"ip", "must be a valid IP address"},
		{"ipv4", "must be a valid IPv4 address"},
		{"ipv6", "must be a valid IPv6 address"},
		{"json", "must be valid JSON"},
		{"base64", "must be valid base64"},
		{"hexcolor", "must be a valid hex color"},
		{"latitude", "must be a valid latitude"},
		{"longitude", "must be a valid longitude"},
		{"hostname", "must be a valid hostname"},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			fe := getFieldError(tt.tag, "")
			if fe == nil {
				t.Skip("Could not generate field error for tag")
				return
			}

			msg := getErrorMessage(fe)
			assert.Contains(t, msg, tt.wantContain, "message for tag %s should contain %s, got: %s", tt.tag, tt.wantContain, msg)
		})
	}
}

func TestFieldError(t *testing.T) {
	fe := FieldError{
		Field:   "email",
		Value:   "invalid-email",
		Tag:     "email",
		Message: "email must be a valid email address",
	}

	assert.Equal(t, "email", fe.Field)
	assert.Equal(t, "invalid-email", fe.Value)
	assert.Equal(t, "email", fe.Tag)
	assert.Equal(t, "email must be a valid email address", fe.Message)
}

func TestErrorTypes(t *testing.T) {
	t.Run("Error implements error interface", func(t *testing.T) {
		var _ error = Error{}
		var _ error = &Error{}
	})

	t.Run("ParseError implements error interface", func(t *testing.T) {
		var _ error = ParseError{}
		var _ error = &ParseError{}
	})
}
