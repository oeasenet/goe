package http

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationError is what the bundled validator returns when `validate` tags
// fail. GOE's default error handler renders it as an HTTP 400 whose message is
// the first failed rule's message — one error at a time, in field declaration
// order. Handlers that want a different rendering (all fields at once, a
// custom envelope) can catch it and use Fields:
//
//	if err := c.Bind().JSON(&req); err != nil {
//	    var ve *goehttp.ValidationError
//	    if errors.As(err, &ve) {
//	        return c.Status(400).JSON(myShape(ve.Fields))
//	    }
//	    return err
//	}
//
// Unwrap exposes the underlying validator.ValidationErrors, so errors.As
// reaches the raw go-playground form too.
type ValidationError struct {
	Fields []FieldError

	// raw is the original go-playground error, kept so advanced callers can
	// reach parameters this simplified form does not carry.
	raw validator.ValidationErrors
}

// FieldError describes one failed rule on one field. Field carries the json
// name when the struct declares one.
//
// The submitted value is deliberately not included: request DTOs routinely
// carry passwords and tokens, and a validation response must never echo them
// back. Read it from the request in the handler if you truly need it.
type FieldError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Param   string `json:"param,omitempty"`
	Message string `json:"message"`
}

// Error implements error with a flat, log-friendly summary.
func (e *ValidationError) Error() string {
	parts := make([]string, 0, len(e.Fields))
	for _, fe := range e.Fields {
		parts = append(parts, fmt.Sprintf("%s: %s", fe.Field, fe.Message))
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

// FirstMessage returns the first failed rule's message — what the default
// error handler presents, one error at a time, in field declaration order.
func (e *ValidationError) FirstMessage() string {
	if len(e.Fields) == 0 {
		return "Validation failed"
	}
	return e.Fields[0].Message
}

// Unwrap exposes the raw validator.ValidationErrors.
func (e *ValidationError) Unwrap() error {
	return e.raw
}

// newValidationError converts go-playground field errors into the typed form.
func newValidationError(errs validator.ValidationErrors) *ValidationError {
	fields := make([]FieldError, 0, len(errs))
	for _, fe := range errs {
		fields = append(fields, FieldError{
			Field:   fe.Field(),
			Tag:     fe.Tag(),
			Param:   fe.Param(),
			Message: fieldMessage(fe),
		})
	}
	return &ValidationError{Fields: fields, raw: errs}
}

// fieldMessage returns a human-readable message for a failed rule.
func fieldMessage(e validator.FieldError) string {
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
	case "uuid", "uuid3", "uuid4", "uuid5":
		return fmt.Sprintf("%s must be a valid UUID", field)
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
	case "cidr", "cidrv4", "cidrv6":
		return fmt.Sprintf("%s must be a valid CIDR notation", field)
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
	case "isbn", "isbn10", "isbn13":
		return fmt.Sprintf("%s must be a valid ISBN", field)
	case "hexcolor":
		return fmt.Sprintf("%s must be a valid hex color", field)
	case "latitude":
		return fmt.Sprintf("%s must be a valid latitude", field)
	case "longitude":
		return fmt.Sprintf("%s must be a valid longitude", field)
	case "semver":
		return fmt.Sprintf("%s must be a valid semantic version", field)
	default:
		if param != "" {
			return fmt.Sprintf("%s failed validation on tag '%s=%s'", field, tag, param)
		}
		return fmt.Sprintf("%s failed validation on tag '%s'", field, tag)
	}
}
