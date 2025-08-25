package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Error ValidationError represents a validation error with field details
type Error struct {
	Errors []FieldError `json:"errors"`
}

// FieldError represents a single field validation error
type FieldError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value,omitempty"`
	Tag     string      `json:"tag"`
	Message string      `json:"message"`
}

// Error returns the error message
func (e Error) Error() string {
	var messages []string
	for _, err := range e.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return fmt.Sprintf("validation failed: %s", strings.Join(messages, ", "))
}

// NewValidationError creates a ValidationError from validator errors
func NewValidationError(err error) error {
	var ve validator.ValidationErrors
	var errs validator.ValidationErrors
	if errors.As(err, &errs) {
		ve = errs
	}

	var fieldErrors []FieldError
	for _, e := range ve {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   e.Field(),
			Value:   e.Value(),
			Tag:     e.Tag(),
			Message: getErrorMessage(e),
		})
	}

	return Error{
		Errors: fieldErrors,
	}
}

// ParseError represents a parsing error
type ParseError struct {
	Message string `json:"message"`
}

// Error returns the error message
func (e ParseError) Error() string {
	return fmt.Sprintf("parse error: %s", e.Message)
}

// NewParseError creates a ParseError
func NewParseError(err error) error {
	return ParseError{
		Message: err.Error(),
	}
}

// getErrorMessage returns a human-readable error message for a field error
func getErrorMessage(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()
	param := e.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "numeric":
		return fmt.Sprintf("%s must be a valid number", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uri":
		return fmt.Sprintf("%s must be a valid URI", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "uuid3":
		return fmt.Sprintf("%s must be a valid UUID v3", field)
	case "uuid4":
		return fmt.Sprintf("%s must be a valid UUID v4", field)
	case "uuid5":
		return fmt.Sprintf("%s must be a valid UUID v5", field)
	case "ascii":
		return fmt.Sprintf("%s must contain only ASCII characters", field)
	case "contains":
		return fmt.Sprintf("%s must contain '%s'", field, param)
	case "containsany":
		return fmt.Sprintf("%s must contain at least one of '%s'", field, param)
	case "excludes":
		return fmt.Sprintf("%s must not contain '%s'", field, param)
	case "excludesall":
		return fmt.Sprintf("%s must not contain any of '%s'", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "phone":
		return fmt.Sprintf("%s must be a valid phone number", field)
	case "username":
		return fmt.Sprintf("%s must be a valid username (3-30 characters, alphanumeric and underscore only)", field)
	case "strong_password":
		return fmt.Sprintf("%s must be at least 8 characters and contain uppercase, lowercase, number, and special character", field)
	case "datetime":
		return fmt.Sprintf("%s must be a valid datetime in format %s", field, param)
	case "startswith":
		return fmt.Sprintf("%s must start with '%s'", field, param)
	case "endswith":
		return fmt.Sprintf("%s must end with '%s'", field, param)
	case "ip":
		return fmt.Sprintf("%s must be a valid IP address", field)
	case "ipv4":
		return fmt.Sprintf("%s must be a valid IPv4 address", field)
	case "ipv6":
		return fmt.Sprintf("%s must be a valid IPv6 address", field)
	case "cidr":
		return fmt.Sprintf("%s must be a valid CIDR notation", field)
	case "cidrv4":
		return fmt.Sprintf("%s must be a valid IPv4 CIDR notation", field)
	case "cidrv6":
		return fmt.Sprintf("%s must be a valid IPv6 CIDR notation", field)
	case "mac":
		return fmt.Sprintf("%s must be a valid MAC address", field)
	case "hostname":
		return fmt.Sprintf("%s must be a valid hostname", field)
	case "fqdn":
		return fmt.Sprintf("%s must be a valid FQDN", field)
	case "json":
		return fmt.Sprintf("%s must be valid JSON", field)
	case "base64":
		return fmt.Sprintf("%s must be valid base64", field)
	case "base64url":
		return fmt.Sprintf("%s must be valid base64 URL encoding", field)
	case "isbn":
		return fmt.Sprintf("%s must be a valid ISBN", field)
	case "isbn10":
		return fmt.Sprintf("%s must be a valid ISBN-10", field)
	case "isbn13":
		return fmt.Sprintf("%s must be a valid ISBN-13", field)
	case "btc_addr":
		return fmt.Sprintf("%s must be a valid Bitcoin address", field)
	case "eth_addr":
		return fmt.Sprintf("%s must be a valid Ethereum address", field)
	case "hexcolor":
		return fmt.Sprintf("%s must be a valid hex color", field)
	case "rgb":
		return fmt.Sprintf("%s must be a valid RGB color", field)
	case "rgba":
		return fmt.Sprintf("%s must be a valid RGBA color", field)
	case "hsl":
		return fmt.Sprintf("%s must be a valid HSL color", field)
	case "hsla":
		return fmt.Sprintf("%s must be a valid HSLA color", field)
	case "latitude":
		return fmt.Sprintf("%s must be a valid latitude", field)
	case "longitude":
		return fmt.Sprintf("%s must be a valid longitude", field)
	case "ssn":
		return fmt.Sprintf("%s must be a valid SSN", field)
	case "semver":
		return fmt.Sprintf("%s must be a valid semantic version", field)
	default:
		return fmt.Sprintf("%s failed validation on tag '%s'", field, tag)
	}
}
