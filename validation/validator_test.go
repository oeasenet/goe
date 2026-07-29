package validation

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	v := New()
	require.NotNil(t, v)
	require.NotNil(t, v.validator)
}

func TestValidator_Validate(t *testing.T) {
	v := New()

	type TestStruct struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"gte=0,lte=120"`
	}

	t.Run("valid struct", func(t *testing.T) {
		s := TestStruct{Name: "John", Email: "john@example.com", Age: 30}
		err := v.Validate(s)
		assert.NoError(t, err)
	})

	t.Run("invalid struct returns validation error", func(t *testing.T) {
		s := TestStruct{Name: "", Email: "invalid", Age: -1}
		err := v.Validate(s)
		assert.Error(t, err)

		// Check it's our Error type
		var validationErr Error
		assert.ErrorAs(t, err, &validationErr)
		assert.NotEmpty(t, validationErr.Errors)
	})

	t.Run("uses json tag for field names", func(t *testing.T) {
		type JsonTagStruct struct {
			FirstName string `json:"first_name" validate:"required"`
		}
		s := JsonTagStruct{FirstName: ""}
		err := v.Validate(s)

		var validationErr Error
		require.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "first_name", validationErr.Errors[0].Field)
	})
}

func TestValidator_ValidateVar(t *testing.T) {
	v := New()

	t.Run("valid email", func(t *testing.T) {
		err := v.ValidateVar("test@example.com", "required,email")
		assert.NoError(t, err)
	})

	t.Run("invalid email", func(t *testing.T) {
		err := v.ValidateVar("invalid-email", "required,email")
		assert.Error(t, err)
	})

	t.Run("valid min length", func(t *testing.T) {
		err := v.ValidateVar("hello", "min=3")
		assert.NoError(t, err)
	})

	t.Run("invalid min length", func(t *testing.T) {
		err := v.ValidateVar("hi", "min=3")
		assert.Error(t, err)
	})

	t.Run("valid numeric range", func(t *testing.T) {
		err := v.ValidateVar(50, "gte=0,lte=100")
		assert.NoError(t, err)
	})

	t.Run("invalid numeric range", func(t *testing.T) {
		err := v.ValidateVar(150, "gte=0,lte=100")
		assert.Error(t, err)
	})
}

func TestValidator_RegisterValidation(t *testing.T) {
	v := New()

	// Register custom validation for even numbers
	err := v.RegisterValidation("even", func(fl validator.FieldLevel) bool {
		return fl.Field().Int()%2 == 0
	})
	require.NoError(t, err)

	t.Run("custom validation passes", func(t *testing.T) {
		type TestStruct struct {
			Value int `validate:"even"`
		}
		err := v.Validate(TestStruct{Value: 4})
		assert.NoError(t, err)
	})

	t.Run("custom validation fails", func(t *testing.T) {
		type TestStruct struct {
			Value int `validate:"even"`
		}
		err := v.Validate(TestStruct{Value: 3})
		assert.Error(t, err)
	})
}

func TestValidator_RegisterAlias(t *testing.T) {
	v := New()

	// Register alias for common email+required
	v.RegisterAlias("required_email", "required,email")

	t.Run("alias validation passes", func(t *testing.T) {
		type TestStruct struct {
			Email string `validate:"required_email"`
		}
		err := v.Validate(TestStruct{Email: "test@example.com"})
		assert.NoError(t, err)
	})

	t.Run("alias validation fails on empty", func(t *testing.T) {
		type TestStruct struct {
			Email string `validate:"required_email"`
		}
		err := v.Validate(TestStruct{Email: ""})
		assert.Error(t, err)
	})

	t.Run("alias validation fails on invalid email", func(t *testing.T) {
		type TestStruct struct {
			Email string `validate:"required_email"`
		}
		err := v.Validate(TestStruct{Email: "not-an-email"})
		assert.Error(t, err)
	})
}

func TestValidator_RegisterStructValidation(t *testing.T) {
	v := New()

	type DateRange struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}

	// Register struct-level validation
	v.RegisterStructValidation(func(sl validator.StructLevel) {
		dr := sl.Current().Interface().(DateRange)
		if dr.StartDate > dr.EndDate {
			sl.ReportError(dr.StartDate, "start_date", "StartDate", "startbeforeend", "")
		}
	}, DateRange{})

	t.Run("struct validation passes", func(t *testing.T) {
		dr := DateRange{StartDate: "2024-01-01", EndDate: "2024-12-31"}
		err := v.Validate(dr)
		assert.NoError(t, err)
	})

	t.Run("struct validation fails", func(t *testing.T) {
		dr := DateRange{StartDate: "2024-12-31", EndDate: "2024-01-01"}
		err := v.Validate(dr)
		assert.Error(t, err)
	})
}

func TestValidator_GetValidator(t *testing.T) {
	v := New()
	underlying := v.GetValidator()

	require.NotNil(t, underlying)
	assert.IsType(t, &validator.Validate{}, underlying)
}

func TestValidateStruct(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"required"`
	}

	t.Run("valid struct", func(t *testing.T) {
		err := ValidateStruct(TestStruct{Name: "test"})
		assert.NoError(t, err)
	})

	t.Run("invalid struct", func(t *testing.T) {
		err := ValidateStruct(TestStruct{Name: ""})
		assert.Error(t, err)
	})
}

func TestMustValidate(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"required"`
	}

	t.Run("does not panic on valid struct", func(t *testing.T) {
		assert.NotPanics(t, func() {
			MustValidate(TestStruct{Name: "test"})
		})
	})

	t.Run("panics on invalid struct", func(t *testing.T) {
		assert.Panics(t, func() {
			MustValidate(TestStruct{Name: ""})
		})
	})
}

func TestCustomValidators(t *testing.T) {
	v := New()

	t.Run("phone validation", func(t *testing.T) {
		type TestStruct struct {
			Phone string `validate:"phone"`
		}

		validPhones := []string{"+1234567890", "123-456-7890", "1234567890123"}
		for _, phone := range validPhones {
			err := v.Validate(TestStruct{Phone: phone})
			assert.NoError(t, err, "phone %s should be valid", phone)
		}

		invalidPhones := []string{"123", "abc-def-ghij", "123456789012345678"}
		for _, phone := range invalidPhones {
			err := v.Validate(TestStruct{Phone: phone})
			assert.Error(t, err, "phone %s should be invalid", phone)
		}
	})

	t.Run("username validation", func(t *testing.T) {
		type TestStruct struct {
			Username string `validate:"username"`
		}

		validUsernames := []string{"john_doe", "user123", "JohnDoe", "a_b_c_123"}
		for _, username := range validUsernames {
			err := v.Validate(TestStruct{Username: username})
			assert.NoError(t, err, "username %s should be valid", username)
		}

		invalidUsernames := []string{"ab", "user-name", "user@name", "a very long username that exceeds the limit"}
		for _, username := range invalidUsernames {
			err := v.Validate(TestStruct{Username: username})
			assert.Error(t, err, "username %s should be invalid", username)
		}
	})

	t.Run("strong_password validation", func(t *testing.T) {
		type TestStruct struct {
			Password string `validate:"strong_password"`
		}

		validPasswords := []string{"SecureP@ss1", "Abc123!@#", "MyStr0ng!Pass"}
		for _, password := range validPasswords {
			err := v.Validate(TestStruct{Password: password})
			assert.NoError(t, err, "password %s should be valid", password)
		}

		invalidPasswords := []string{
			"short1!",        // too short
			"alllowercase1!", // no uppercase
			"ALLUPPERCASE1!", // no lowercase
			"NoNumbers!",     // no number
			"NoSpecial123",   // no special char
		}
		for _, password := range invalidPasswords {
			err := v.Validate(TestStruct{Password: password})
			assert.Error(t, err, "password %s should be invalid", password)
		}
	})
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"12345", true},
		{"0", true},
		{"999999", true},
		{"", false},
		{"12.34", false},
		{"12a34", false},
		{"-123", false},
		{" 123", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsNumeric(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJsonTagNameFunction(t *testing.T) {
	v := New()

	t.Run("uses json tag name", func(t *testing.T) {
		type TestStruct struct {
			FirstName string `json:"first_name" validate:"required"`
		}
		err := v.Validate(TestStruct{FirstName: ""})

		var validationErr Error
		require.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "first_name", validationErr.Errors[0].Field)
	})

	t.Run("ignores json omitted field names", func(t *testing.T) {
		type TestStruct struct {
			Internal string `json:"-" validate:"required"`
		}
		// When json tag is "-", validator uses empty string for field name
		// but the validation still works on the field
		err := v.Validate(TestStruct{Internal: ""})
		assert.Error(t, err)
	})
}
