package goe

import (
	"os"
	"testing"
)

func TestConfigOverride_UnitTests(t *testing.T) {
	// Clean up any existing env vars
	os.Unsetenv("HTTP_PORT")
	os.Unsetenv("TEST_VAR")

	t.Run("HTTPPort_override_config_only", func(t *testing.T) {
		// Test just the config override without starting HTTP server
		_ = New(Options{
			HTTPPort: 5555,
		})

		// Test that the config override is working
		config := Config()
		if port := config.GetInt("HTTP_PORT"); port != 5555 {
			t.Errorf("Expected HTTP_PORT to be 5555, got %d", port)
		}

		// Clean up
		instance.app = nil
		instance.config = nil
		instance.logger = nil
		instance.http = nil
	})

	t.Run("ConfigOverrides_general_config_only", func(t *testing.T) {
		// Test just the config override without starting HTTP server
		_ = New(Options{
			ConfigOverrides: map[string]any{
				"HTTP_PORT": 4444,
				"TEST_VAR":  "test_value_override",
				"DB_HOST":   "127.0.0.1",
				"DEBUG":     true,
			},
		})

		// Test that the config override is working
		config := Config()
		if port := config.GetInt("HTTP_PORT"); port != 4444 {
			t.Errorf("Expected HTTP_PORT to be 4444, got %d", port)
		}

		if testVar := config.GetString("TEST_VAR"); testVar != "test_value_override" {
			t.Errorf("Expected TEST_VAR to be 'test_value_override', got %s", testVar)
		}

		if dbHost := config.GetString("DB_HOST"); dbHost != "127.0.0.1" {
			t.Errorf("Expected DB_HOST to be '127.0.0.1', got %s", dbHost)
		}

		if debug := config.GetBool("DEBUG"); debug != true {
			t.Errorf("Expected DEBUG to be true, got %v", debug)
		}

		// Clean up
		instance.app = nil
		instance.config = nil
		instance.logger = nil
		instance.http = nil
	})

	t.Run("HTTPPort_precedence_over_ConfigOverrides", func(t *testing.T) {
		// Test that HTTPPort takes precedence over ConfigOverrides
		_ = New(Options{
			HTTPPort: 3333,
			ConfigOverrides: map[string]any{
				"HTTP_PORT": 2222, // This should be overridden by HTTPPort
				"OTHER_VAR": "other_value",
			},
		})

		// Test that the config override is working
		config := Config()
		if port := config.GetInt("HTTP_PORT"); port != 3333 {
			t.Errorf("Expected HTTP_PORT to be 3333 (HTTPPort takes precedence), got %d", port)
		}

		if otherVar := config.GetString("OTHER_VAR"); otherVar != "other_value" {
			t.Errorf("Expected OTHER_VAR to be 'other_value', got %s", otherVar)
		}

		// Clean up
		instance.app = nil
		instance.config = nil
		instance.logger = nil
		instance.http = nil
	})

	t.Run("fallback_to_env_vars", func(t *testing.T) {
		// Set an env var to test fallback
		os.Setenv("HTTP_PORT", "8888")
		defer os.Unsetenv("HTTP_PORT")

		// Test with no overrides - should use env var
		_ = New(Options{})

		config := Config()
		if port := config.GetInt("HTTP_PORT"); port != 8888 {
			t.Errorf("Expected HTTP_PORT to be 8888 (from env var), got %d", port)
		}

		// Clean up
		instance.app = nil
		instance.config = nil
		instance.logger = nil
		instance.http = nil
	})

	t.Run("override_takes_precedence_over_env_vars", func(t *testing.T) {
		// Set an env var
		os.Setenv("HTTP_PORT", "7777")
		defer os.Unsetenv("HTTP_PORT")

		// Test with override - should use override not env var
		_ = New(Options{
			HTTPPort: 6666,
		})

		config := Config()
		if port := config.GetInt("HTTP_PORT"); port != 6666 {
			t.Errorf("Expected HTTP_PORT to be 6666 (override takes precedence), got %d", port)
		}

		// Clean up
		instance.app = nil
		instance.config = nil
		instance.logger = nil
		instance.http = nil
	})
}
