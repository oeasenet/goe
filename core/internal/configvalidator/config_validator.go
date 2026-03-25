package configvalidator

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"go.oease.dev/goe/v2/contract"
)

// ConfigValidator provides utilities for validating configuration
type ConfigValidator struct {
	config       contract.Config
	module       string
	requirements []contract.ConfigRequirement
}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator(config contract.Config, module string) *ConfigValidator {
	return &ConfigValidator{
		config:       config,
		module:       module,
		requirements: make([]contract.ConfigRequirement, 0),
	}
}

// AddRequirement adds a configuration requirement
func (v *ConfigValidator) AddRequirement(req contract.ConfigRequirement) *ConfigValidator {
	v.requirements = append(v.requirements, req)
	return v
}

// Require adds a required configuration key
func (v *ConfigValidator) Require(key, description string) *ConfigValidator {
	return v.AddRequirement(contract.ConfigRequirement{
		Key:         key,
		Description: description,
		Required:    true,
	})
}

// RequireWithValidator adds a required configuration key with custom validation
func (v *ConfigValidator) RequireWithValidator(key, description string, validateFn func(value any) error) *ConfigValidator {
	return v.AddRequirement(contract.ConfigRequirement{
		Key:         key,
		Description: description,
		Required:    true,
		ValidateFn:  validateFn,
	})
}

// Optional adds an optional configuration key with validation
func (v *ConfigValidator) Optional(key, description string, validateFn func(value any) error) *ConfigValidator {
	return v.AddRequirement(contract.ConfigRequirement{
		Key:         key,
		Description: description,
		Required:    false,
		ValidateFn:  validateFn,
	})
}

// Validate performs the validation
func (v *ConfigValidator) Validate() error {
	validationErr := &contract.ConfigValidationError{
		Module:      v.module,
		MissingKeys: make([]string, 0),
		InvalidKeys: make(map[string]string),
	}

	hasErrors := false

	for _, req := range v.requirements {
		if req.Required && !v.config.Has(req.Key) {
			validationErr.MissingKeys = append(validationErr.MissingKeys, req.Key)
			hasErrors = true
			continue
		}

		// If key exists and has validator, run it
		if v.config.Has(req.Key) && req.ValidateFn != nil {
			value := v.config.Get(req.Key)
			if err := req.ValidateFn(value); err != nil {
				validationErr.InvalidKeys[req.Key] = err.Error()
				hasErrors = true
			}
		}
	}

	if hasErrors {
		return validationErr
	}

	return nil
}

// Common validators

// ValidateNotEmpty validates that a string value is not empty
func ValidateNotEmpty(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string value")
	}
	if strings.TrimSpace(str) == "" {
		return fmt.Errorf("value cannot be empty")
	}
	return nil
}

// ValidatePort validates that a value is a valid port number
func ValidatePort(value any) error {
	var port int
	switch v := value.(type) {
	case int:
		port = v
	case int64:
		port = int(v)
	case string:
		if v == "" {
			return fmt.Errorf("port cannot be empty")
		}
		_, err := fmt.Sscanf(v, "%d", &port)
		if err != nil {
			return fmt.Errorf("invalid port format: %v", err)
		}
	default:
		return fmt.Errorf("expected numeric port value")
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}

// ValidateHostPort validates host:port format
func ValidateHostPort(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string value")
	}

	host, port, err := net.SplitHostPort(str)
	if err != nil {
		return fmt.Errorf("invalid host:port format: %v", err)
	}

	if host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	// Validate port
	if err := ValidatePort(port); err != nil {
		return err
	}

	return nil
}

// ValidateURL validates that a value is a valid URL
func ValidateURL(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string value")
	}

	u, err := url.Parse(str)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	if u.Scheme == "" {
		return fmt.Errorf("URL scheme is required")
	}

	if u.Host == "" {
		return fmt.Errorf("URL host is required")
	}

	return nil
}

// ValidateDatabaseDriver validates database driver names
func ValidateDatabaseDriver(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string value")
	}

	validDrivers := []string{"mysql", "postgres", "postgresql", "pgsql", "sqlite", "sqlite3", "sqlserver", "mssql"}
	driver := strings.ToLower(str)

	if slices.Contains(validDrivers, driver) {
		return nil
	}

	return fmt.Errorf("unsupported database driver: %s (valid options: %s)", str, strings.Join(validDrivers, ", "))
}

// ValidateEmail validates email format
func ValidateEmail(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string value")
	}

	// Simple email regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(str) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidateOneOf validates that value is one of allowed options
func ValidateOneOf(options ...string) func(value any) error {
	return func(value any) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string value")
		}

		if slices.Contains(options, str) {
			return nil
		}

		return fmt.Errorf("value must be one of: %s", strings.Join(options, ", "))
	}
}

// ValidateNonNegativeInt validates that value is a non-negative integer (>= 0)
func ValidateNonNegativeInt(value any) error {
	var num int
	switch v := value.(type) {
	case int:
		num = v
	case int64:
		num = int(v)
	case string:
		if v == "" {
			return fmt.Errorf("value cannot be empty")
		}
		_, err := fmt.Sscanf(v, "%d", &num)
		if err != nil {
			return fmt.Errorf("invalid numeric format: %v", err)
		}
	default:
		return fmt.Errorf("expected numeric value")
	}

	if num < 0 {
		return fmt.Errorf("value must be non-negative")
	}

	return nil
}

// ValidatePositiveInt validates that value is a positive integer
func ValidatePositiveInt(value any) error {
	var num int
	switch v := value.(type) {
	case int:
		num = v
	case int64:
		num = int(v)
	case string:
		if v == "" {
			return fmt.Errorf("value cannot be empty")
		}
		_, err := fmt.Sscanf(v, "%d", &num)
		if err != nil {
			return fmt.Errorf("invalid numeric format: %v", err)
		}
	default:
		return fmt.Errorf("expected numeric value")
	}

	if num <= 0 {
		return fmt.Errorf("value must be positive")
	}

	return nil
}
