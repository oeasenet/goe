package contract

import "strings"

import "fmt"

// ConfigValidator defines the interface for configuration validation
type ConfigValidator interface {
	// ValidateConfig validates the configuration for a module
	// Returns nil if valid, or an error describing what's missing/invalid
	ValidateConfig() error
}

// ConfigRequirement represents a required configuration field
type ConfigRequirement struct {
	Key         string
	Description string
	Required    bool
	ValidateFn  func(value any) error
}

// ConfigValidationError represents a validation error with details
type ConfigValidationError struct {
	Module      string
	MissingKeys []string
	InvalidKeys map[string]string // key -> error message
}

func (e *ConfigValidationError) Error() string {
	var msg strings.Builder
	fmt.Fprintf(&msg, "Configuration validation failed for module '%s'", e.Module)

	if len(e.MissingKeys) > 0 {
		fmt.Fprintf(&msg, "\n  Missing required configuration keys: %v", e.MissingKeys)
	}

	if len(e.InvalidKeys) > 0 {
		msg.WriteString("\n  Invalid configuration values:")
		for key, err := range e.InvalidKeys {
			fmt.Fprintf(&msg, "\n    - %s: %s", key, err)
		}
	}

	return msg.String()
}

// ModuleWithValidator represents a module that supports configuration validation
type ModuleWithValidator interface {
	Module
	ConfigValidator
}
