package config_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.oease.dev/goe/v2/core/config"
)

func TestConfig(t *testing.T) {
	// Create a new config
	cfg := config.New()

	// Test module name
	t.Run("Module Name", func(t *testing.T) {
		name := cfg.Name()
		if name != "config" {
			t.Errorf("Expected module name to be 'config', got '%s'", name)
		}
	})

	// Test basic get and set
	t.Run("Get and Set", func(t *testing.T) {
		// Set a value
		cfg.Set("test_key", "test_value")

		// Get the value
		value := cfg.Get("test_key")
		if value != "test_value" {
			t.Errorf("Expected value to be 'test_value', got '%s'", value)
		}

		// Get a non-existent value
		value = cfg.Get("non_existent_key")
		if value != "" {
			t.Errorf("Expected value to be empty, got '%s'", value)
		}

		// Get with default value
		value = cfg.GetDefault("non_existent_key", "default_value")
		if value != "default_value" {
			t.Errorf("Expected value to be 'default_value', got '%s'", value)
		}

		// Check if key exists
		if !cfg.Has("test_key") {
			t.Errorf("Expected key 'test_key' to exist")
		}

		if cfg.Has("non_existent_key") {
			t.Errorf("Expected key 'non_existent_key' to not exist")
		}
	})

	// Test environment variables
	t.Run("Environment Variables", func(t *testing.T) {
		// Set an environment variable
		os.Setenv("TEST_ENV_VAR", "test_env_value")
		defer os.Unsetenv("TEST_ENV_VAR")

		// Create a new config to load environment variables
		newCfg := config.New()
		err := newCfg.Initialize(context.Background())
		if err != nil {
			t.Errorf("Failed to initialize config: %v", err)
		}

		// Get the environment variable
		value := newCfg.Get("TEST_ENV_VAR")
		if value != "test_env_value" {
			t.Errorf("Expected value to be 'test_env_value', got '%s'", value)
		}
	})

	// Test .env file loading
	t.Run("Env File Loading", func(t *testing.T) {
		// Create a temporary directory
		tempDir, err := os.MkdirTemp("", "config_test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create a .env file
		envContent := `
TEST_KEY=test_value
GOE_ENV=test
`
		err = os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write .env file: %v", err)
		}

		// Create a .test.env file
		testEnvContent := `
TEST_KEY_OVERRIDE=override_value
TEST_KEY_ENV_SPECIFIC=env_specific_value
`
		err = os.WriteFile(filepath.Join(tempDir, ".test.env"), []byte(testEnvContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write .test.env file: %v", err)
		}

		// Create a new config with the temp dir as base path
		newCfg := config.New()
		newCfg.SetBasePath(tempDir)

		// Initialize the config
		err = newCfg.Initialize(context.Background())
		if err != nil {
			t.Errorf("Failed to initialize config: %v", err)
		}

		// Check if the values were loaded
		if newCfg.Get("TEST_KEY") != "test_value" {
			t.Errorf("Expected TEST_KEY to be 'test_value', got '%s'", newCfg.Get("TEST_KEY"))
		}

		if newCfg.Get("GOE_ENV") != "test" {
			t.Errorf("Expected GOE_ENV to be 'test', got '%s'", newCfg.Get("GOE_ENV"))
		}

		if newCfg.Get("TEST_KEY_OVERRIDE") != "override_value" {
			t.Errorf("Expected TEST_KEY_OVERRIDE to be 'override_value', got '%s'", newCfg.Get("TEST_KEY_OVERRIDE"))
		}

		if newCfg.Get("TEST_KEY_ENV_SPECIFIC") != "env_specific_value" {
			t.Errorf("Expected TEST_KEY_ENV_SPECIFIC to be 'env_specific_value', got '%s'", newCfg.Get("TEST_KEY_ENV_SPECIFIC"))
		}
	})

	// Test module lifecycle
	t.Run("Module Lifecycle", func(t *testing.T) {
		// Initialize
		err := cfg.Initialize(context.Background())
		if err != nil {
			t.Errorf("Failed to initialize config: %v", err)
		}

		// Start
		err = cfg.Start(context.Background())
		if err != nil {
			t.Errorf("Failed to start config: %v", err)
		}

		// Stop
		err = cfg.Stop(context.Background())
		if err != nil {
			t.Errorf("Failed to stop config: %v", err)
		}
	})

	// Test reload
	t.Run("Reload", func(t *testing.T) {
		// Set a value
		cfg.Set("reload_test", "before_reload")

		// Reload
		err := cfg.Reload(context.Background())
		if err != nil {
			t.Errorf("Failed to reload config: %v", err)
		}

		// The value should be gone after reload
		if cfg.Has("reload_test") {
			t.Errorf("Expected key 'reload_test' to not exist after reload")
		}
	})

	// Test GetInt and GetIntDefault
	t.Run("GetInt and GetIntDefault", func(t *testing.T) {
		// Set integer values
		cfg.Set("int_key", "123")
		cfg.Set("invalid_int", "not_an_int")

		// Test GetInt with valid value
		intValue, err := cfg.GetInt("int_key")
		if err != nil {
			t.Errorf("Failed to get int value: %v", err)
		}
		if intValue != 123 {
			t.Errorf("Expected int value to be 123, got %d", intValue)
		}

		// Test GetInt with invalid value
		_, err = cfg.GetInt("invalid_int")
		if err == nil {
			t.Errorf("Expected error when getting invalid int value")
		}

		// Test GetInt with non-existent key
		_, err = cfg.GetInt("non_existent_key")
		if err == nil {
			t.Errorf("Expected error when getting non-existent key")
		}

		// Test GetIntDefault with valid value
		intValue = cfg.GetIntDefault("int_key", 456)
		if intValue != 123 {
			t.Errorf("Expected int value to be 123, got %d", intValue)
		}

		// Test GetIntDefault with invalid value
		intValue = cfg.GetIntDefault("invalid_int", 456)
		if intValue != 456 {
			t.Errorf("Expected int value to be 456, got %d", intValue)
		}

		// Test GetIntDefault with non-existent key
		intValue = cfg.GetIntDefault("non_existent_key", 456)
		if intValue != 456 {
			t.Errorf("Expected int value to be 456, got %d", intValue)
		}
	})

	// Test GetBool and GetBoolDefault
	t.Run("GetBool and GetBoolDefault", func(t *testing.T) {
		// Set boolean values
		cfg.Set("bool_true_1", "true")
		cfg.Set("bool_true_2", "yes")
		cfg.Set("bool_true_3", "1")
		cfg.Set("bool_true_4", "on")
		cfg.Set("bool_false_1", "false")
		cfg.Set("bool_false_2", "no")
		cfg.Set("bool_false_3", "0")
		cfg.Set("bool_false_4", "off")
		cfg.Set("invalid_bool", "not_a_bool")

		// Test GetBool with valid values
		testCases := []struct {
			key      string
			expected bool
		}{
			{"bool_true_1", true},
			{"bool_true_2", true},
			{"bool_true_3", true},
			{"bool_true_4", true},
			{"bool_false_1", false},
			{"bool_false_2", false},
			{"bool_false_3", false},
			{"bool_false_4", false},
		}

		for _, tc := range testCases {
			boolValue, err := cfg.GetBool(tc.key)
			if err != nil {
				t.Errorf("Failed to get bool value for %s: %v", tc.key, err)
			}
			if boolValue != tc.expected {
				t.Errorf("Expected bool value for %s to be %v, got %v", tc.key, tc.expected, boolValue)
			}
		}

		// Test GetBool with invalid value
		_, err := cfg.GetBool("invalid_bool")
		if err == nil {
			t.Errorf("Expected error when getting invalid bool value")
		}

		// Test GetBool with non-existent key
		_, err = cfg.GetBool("non_existent_key")
		if err == nil {
			t.Errorf("Expected error when getting non-existent key")
		}

		// Test GetBoolDefault with valid value
		boolValue := cfg.GetBoolDefault("bool_true_1", false)
		if !boolValue {
			t.Errorf("Expected bool value to be true, got false")
		}

		// Test GetBoolDefault with invalid value
		boolValue = cfg.GetBoolDefault("invalid_bool", true)
		if !boolValue {
			t.Errorf("Expected bool value to be true, got false")
		}

		// Test GetBoolDefault with non-existent key
		boolValue = cfg.GetBoolDefault("non_existent_key", true)
		if !boolValue {
			t.Errorf("Expected bool value to be true, got false")
		}
	})

	// Test GetFloat and GetFloatDefault
	t.Run("GetFloat and GetFloatDefault", func(t *testing.T) {
		// Set float values
		cfg.Set("float_key", "123.456")
		cfg.Set("invalid_float", "not_a_float")

		// Test GetFloat with valid value
		floatValue, err := cfg.GetFloat("float_key")
		if err != nil {
			t.Errorf("Failed to get float value: %v", err)
		}
		if floatValue != 123.456 {
			t.Errorf("Expected float value to be 123.456, got %f", floatValue)
		}

		// Test GetFloat with invalid value
		_, err = cfg.GetFloat("invalid_float")
		if err == nil {
			t.Errorf("Expected error when getting invalid float value")
		}

		// Test GetFloat with non-existent key
		_, err = cfg.GetFloat("non_existent_key")
		if err == nil {
			t.Errorf("Expected error when getting non-existent key")
		}

		// Test GetFloatDefault with valid value
		floatValue = cfg.GetFloatDefault("float_key", 456.789)
		if floatValue != 123.456 {
			t.Errorf("Expected float value to be 123.456, got %f", floatValue)
		}

		// Test GetFloatDefault with invalid value
		floatValue = cfg.GetFloatDefault("invalid_float", 456.789)
		if floatValue != 456.789 {
			t.Errorf("Expected float value to be 456.789, got %f", floatValue)
		}

		// Test GetFloatDefault with non-existent key
		floatValue = cfg.GetFloatDefault("non_existent_key", 456.789)
		if floatValue != 456.789 {
			t.Errorf("Expected float value to be 456.789, got %f", floatValue)
		}
	})

	// Test GetDuration and GetDurationDefault
	t.Run("GetDuration and GetDurationDefault", func(t *testing.T) {
		// Set duration values
		cfg.Set("duration_key", "1h30m")
		cfg.Set("invalid_duration", "not_a_duration")

		// Test GetDuration with valid value
		durationValue, err := cfg.GetDuration("duration_key")
		if err != nil {
			t.Errorf("Failed to get duration value: %v", err)
		}
		expected := 90 * time.Minute
		if durationValue != expected {
			t.Errorf("Expected duration value to be %v, got %v", expected, durationValue)
		}

		// Test GetDuration with invalid value
		_, err = cfg.GetDuration("invalid_duration")
		if err == nil {
			t.Errorf("Expected error when getting invalid duration value")
		}

		// Test GetDuration with non-existent key
		_, err = cfg.GetDuration("non_existent_key")
		if err == nil {
			t.Errorf("Expected error when getting non-existent key")
		}

		// Test GetDurationDefault with valid value
		durationValue = cfg.GetDurationDefault("duration_key", 2*time.Hour)
		if durationValue != expected {
			t.Errorf("Expected duration value to be %v, got %v", expected, durationValue)
		}

		// Test GetDurationDefault with invalid value
		durationValue = cfg.GetDurationDefault("invalid_duration", 2*time.Hour)
		if durationValue != 2*time.Hour {
			t.Errorf("Expected duration value to be %v, got %v", 2*time.Hour, durationValue)
		}

		// Test GetDurationDefault with non-existent key
		durationValue = cfg.GetDurationDefault("non_existent_key", 2*time.Hour)
		if durationValue != 2*time.Hour {
			t.Errorf("Expected duration value to be %v, got %v", 2*time.Hour, durationValue)
		}
	})

	// Test GetStringSlice
	t.Run("GetStringSlice", func(t *testing.T) {
		// Set string slice values
		cfg.Set("slice_key", "a,b,c")
		cfg.Set("slice_key_with_spaces", " a , b , c ")
		cfg.Set("slice_key_custom_sep", "a|b|c")
		cfg.Set("empty_slice", "")

		// Test GetStringSlice with default separator
		sliceValue := cfg.GetStringSlice("slice_key", "")
		if len(sliceValue) != 3 || sliceValue[0] != "a" || sliceValue[1] != "b" || sliceValue[2] != "c" {
			t.Errorf("Expected slice value to be [a b c], got %v", sliceValue)
		}

		// Test GetStringSlice with spaces
		sliceValue = cfg.GetStringSlice("slice_key_with_spaces", "")
		if len(sliceValue) != 3 || sliceValue[0] != "a" || sliceValue[1] != "b" || sliceValue[2] != "c" {
			t.Errorf("Expected slice value to be [a b c], got %v", sliceValue)
		}

		// Test GetStringSlice with custom separator
		sliceValue = cfg.GetStringSlice("slice_key_custom_sep", "|")
		if len(sliceValue) != 3 || sliceValue[0] != "a" || sliceValue[1] != "b" || sliceValue[2] != "c" {
			t.Errorf("Expected slice value to be [a b c], got %v", sliceValue)
		}

		// Test GetStringSlice with empty value
		sliceValue = cfg.GetStringSlice("empty_slice", "")
		if len(sliceValue) != 0 {
			t.Errorf("Expected slice value to be empty, got %v", sliceValue)
		}

		// Test GetStringSlice with non-existent key
		sliceValue = cfg.GetStringSlice("non_existent_key", "")
		if sliceValue != nil {
			t.Errorf("Expected slice value to be nil, got %v", sliceValue)
		}
	})
}

// Benchmark config operations
func BenchmarkConfig(b *testing.B) {
	cfg := config.New()
	cfg.Set("bench_key", "bench_value")
	cfg.Set("bench_int", "123")
	cfg.Set("bench_bool", "true")
	cfg.Set("bench_float", "123.456")
	cfg.Set("bench_duration", "1h30m")
	cfg.Set("bench_slice", "a,b,c")

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cfg.Get("bench_key")
		}
	})

	b.Run("GetDefault", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cfg.GetDefault("non_existent_key", "default")
		}
	})

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cfg.Set("bench_key", "bench_value")
		}
	})

	b.Run("Has", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cfg.Has("bench_key")
		}
	})

	b.Run("GetInt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = cfg.GetInt("bench_int")
		}
	})

	b.Run("GetIntDefault", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cfg.GetIntDefault("bench_int", 456)
		}
	})

	b.Run("GetBool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = cfg.GetBool("bench_bool")
		}
	})

	b.Run("GetBoolDefault", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cfg.GetBoolDefault("bench_bool", false)
		}
	})

	b.Run("GetFloat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = cfg.GetFloat("bench_float")
		}
	})

	b.Run("GetFloatDefault", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cfg.GetFloatDefault("bench_float", 456.789)
		}
	})

	b.Run("GetDuration", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = cfg.GetDuration("bench_duration")
		}
	})

	b.Run("GetDurationDefault", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cfg.GetDurationDefault("bench_duration", 2*time.Hour)
		}
	})

	b.Run("GetStringSlice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cfg.GetStringSlice("bench_slice", "")
		}
	})
}
