package tests

import (
	"os"
	"testing"

	"go.oease.dev/goe/v2/core/config"
)

func TestEnvironmentLoadingPriority(t *testing.T) {
	// Create test env files
	envContent := `TEST_VAR=from_env
OVERRIDE_VAR=env_value
ENV_ONLY=env_only_value`

	localEnvContent := `TEST_VAR=from_local_env
OVERRIDE_VAR=local_env_value
LOCAL_ONLY=local_only_value`

	// Write test files
	if err := os.WriteFile(".env", []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(".env")

	if err := os.WriteFile(".local.env", []byte(localEnvContent), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(".local.env")

	// Set system environment variable to test highest priority
	os.Setenv("OVERRIDE_VAR", "system_env_value")
	os.Setenv("SYSTEM_ONLY", "system_only_value")
	defer os.Unsetenv("OVERRIDE_VAR")
	defer os.Unsetenv("SYSTEM_ONLY")

	// Create config instance
	cfg := config.New()

	// Test priority: system env > .local.env > .env
	tests := []struct {
		key      string
		expected string
		desc     string
	}{
		{"TEST_VAR", "from_local_env", ".local.env should override .env"},
		{"OVERRIDE_VAR", "system_env_value", "System env should have highest priority"},
		{"ENV_ONLY", "env_only_value", "Value only in .env should be available"},
		{"LOCAL_ONLY", "local_only_value", "Value only in .local.env should be available"},
		{"SYSTEM_ONLY", "system_only_value", "Value only in system env should be available"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := cfg.GetString(tt.key)
			if got != tt.expected {
				t.Errorf("GetString(%q) = %q, want %q", tt.key, got, tt.expected)
			}
		})
	}

	// Test reload maintains the same priority
	if err := cfg.Reload(); err != nil {
		t.Fatal(err)
	}

	// Verify priority is maintained after reload
	for _, tt := range tests {
		t.Run("After reload: "+tt.desc, func(t *testing.T) {
			got := cfg.GetString(tt.key)
			if got != tt.expected {
				t.Errorf("After reload: GetString(%q) = %q, want %q", tt.key, got, tt.expected)
			}
		})
	}
}

func TestGOEENVEnvironmentFile(t *testing.T) {
	// Create test env files
	devEnvContent := `DEV_VAR=dev_value
ENV_SPECIFIC=dev_specific`

	prodEnvContent := `PROD_VAR=prod_value
ENV_SPECIFIC=prod_specific`

	// Write test files
	if err := os.WriteFile(".dev.env", []byte(devEnvContent), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(".dev.env")

	if err := os.WriteFile(".prod.env", []byte(prodEnvContent), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(".prod.env")

	// Test with GOE_ENV=prod
	os.Setenv("GOE_ENV", "prod")
	defer os.Unsetenv("GOE_ENV")

	cfg := config.New()

	// Should load .prod.env instead of .dev.env
	if got := cfg.GetString("PROD_VAR"); got != "prod_value" {
		t.Errorf("GetString(\"PROD_VAR\") = %q, want %q", got, "prod_value")
	}

	if got := cfg.GetString("ENV_SPECIFIC"); got != "prod_specific" {
		t.Errorf("GetString(\"ENV_SPECIFIC\") = %q, want %q", got, "prod_specific")
	}

	// DEV_VAR should not be loaded
	if got := cfg.GetString("DEV_VAR"); got != "" {
		t.Errorf("GetString(\"DEV_VAR\") = %q, want empty string", got)
	}
}
