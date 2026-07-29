package configvalidator

import (
	"testing"

	"go.oease.dev/goe/v2/contract"
)

func TestConfigValidationFramework(t *testing.T) {
	// Test basic validation framework without import cycles

	// Mock config
	config := &mockConfig{
		values: map[string]any{
			"VALID_KEY":    "value",
			"EMPTY_KEY":    "",
			"INVALID_PORT": "99999",
			"VALID_PORT":   "8080",
		},
	}

	// Test basic requirements
	requirements := []contract.ConfigRequirement{
		{
			Key:         "VALID_KEY",
			Description: "A valid key",
			Required:    true,
		},
		{
			Key:         "MISSING_KEY",
			Description: "A missing key",
			Required:    true,
		},
		{
			Key:         "VALID_PORT",
			Description: "A valid port",
			Required:    false,
			ValidateFn:  ValidatePort,
		},
	}

	err := ValidateConfig(config, "test", requirements)
	if err == nil {
		t.Error("Expected validation to fail due to missing key")
	}

	// Test with all valid requirements
	validRequirements := []contract.ConfigRequirement{
		{
			Key:         "VALID_KEY",
			Description: "A valid key",
			Required:    true,
		},
		{
			Key:         "VALID_PORT",
			Description: "A valid port",
			Required:    false,
			ValidateFn:  ValidatePort,
		},
	}

	err = ValidateConfig(config, "test", validRequirements)
	if err != nil {
		t.Errorf("Expected validation to pass but got: %v", err)
	}
}
