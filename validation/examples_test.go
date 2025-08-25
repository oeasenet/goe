package validation_test

import (
	"fmt"
	"testing"

	"go.oease.dev/goe/v2/validation"
)

// Example struct with validation tags
type User struct {
	ID       int    `json:"id" validate:"required,min=1"`
	Name     string `json:"name" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"gte=18,lte=120"`
	Username string `json:"username" validate:"required,username"`
	Password string `json:"password" validate:"required,strong_password"`
	Phone    string `json:"phone" validate:"omitempty,phone"`
	Website  string `json:"website" validate:"omitempty,url"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type PaginationQuery struct {
	Page  int `query:"page" validate:"min=1"`
	Limit int `query:"limit" validate:"min=1,max=100"`
}

func ExampleValidator_Validate() {
	v := validation.New()

	user := User{
		ID:       1,
		Name:     "John Doe",
		Email:    "john@example.com",
		Age:      25,
		Username: "john_doe",
		Password: "SecureP@ss123",
		Phone:    "+1234567890",
		Website:  "https://example.com",
	}

	if err := v.Validate(user); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Println("Validation passed")
	}
	// Output: Validation passed
}

func ExampleValidator_RegisterValidation() {
	v := validation.New()

	// Register custom validation for postal code
	// Note: This is a simplified example
	// In reality, you would use the validator.FieldLevel interface

	type Address struct {
		PostalCode string `json:"postal_code" validate:"required,len=5"`
	}

	addr := Address{PostalCode: "12345"}
	if err := v.Validate(addr); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Println("Valid postal code")
	}
	// Output: Valid postal code
}

func TestValidation(t *testing.T) {
	v := validation.New()

	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
	}{
		{
			name: "valid user",
			data: User{
				ID:       1,
				Name:     "John Doe",
				Email:    "john@example.com",
				Age:      25,
				Username: "john_doe",
				Password: "SecureP@ss123",
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			data: User{
				ID:       1,
				Name:     "John Doe",
				Email:    "invalid-email",
				Age:      25,
				Username: "john_doe",
				Password: "SecureP@ss123",
			},
			wantErr: true,
		},
		{
			name: "age too young",
			data: User{
				ID:       1,
				Name:     "John Doe",
				Email:    "john@example.com",
				Age:      15,
				Username: "john_doe",
				Password: "SecureP@ss123",
			},
			wantErr: true,
		},
		{
			name: "weak password",
			data: User{
				ID:       1,
				Name:     "John Doe",
				Email:    "john@example.com",
				Age:      25,
				Username: "john_doe",
				Password: "weak",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	tests := []struct {
		name string
		fn   func() bool
		want bool
	}{
		{"valid email", func() bool { return validation.IsEmail("test@example.com") }, true},
		{"invalid email", func() bool { return validation.IsEmail("not-an-email") }, false},
		{"valid URL", func() bool { return validation.IsURL("https://example.com") }, true},
		{"invalid URL", func() bool { return validation.IsURL("not-a-url") }, false},
		{"valid UUID", func() bool { return validation.IsUUID("550e8400-e29b-41d4-a716-446655440000") }, true},
		{"invalid UUID", func() bool { return validation.IsUUID("not-a-uuid") }, false},
		{"valid alpha", func() bool { return validation.IsAlpha("abcDEF") }, true},
		{"invalid alpha", func() bool { return validation.IsAlpha("abc123") }, false},
		{"valid alphanum", func() bool { return validation.IsAlphanumeric("abc123") }, true},
		{"invalid alphanum", func() bool { return validation.IsAlphanumeric("abc-123") }, false},
		{"valid numeric", func() bool { return validation.IsNumeric("12345") }, true},
		{"invalid numeric", func() bool { return validation.IsNumeric("12.34") }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(); got != tt.want {
				t.Errorf("Helper function = %v, want %v", got, tt.want)
			}
		})
	}
}
