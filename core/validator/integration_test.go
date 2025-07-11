package validator

import (
	"os"
	"testing"

	"go.oease.dev/goe/v2/contract"
)

func TestConfigValidationFramework(t *testing.T) {
	// Test basic validation framework without import cycles

	// Mock config
	config := &mockConfig{
		values: map[string]interface{}{
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

func createTempConfigFile(t *testing.T, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test-config-*.env")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	return tmpFile.Name()
}
