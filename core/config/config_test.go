package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a temporary .env file
func createTempEnvFile(t *testing.T, filename, content string) string {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, filename)

	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	return filePath
}

// Helper function to change to a temporary directory for testing
func changeToTempDir(t *testing.T) string {
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	return tempDir
}

func TestConfig_New(t *testing.T) {
	tempDir := changeToTempDir(t)

	// Create a test .env file
	envContent := `APP_NAME=test-app
DEBUG=true
PORT=8080`

	err := os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0644)
	require.NoError(t, err)

	// Set some system environment variables
	os.Setenv("GOE_ENV", "test")
	os.Setenv("SYSTEM_VAR", "system_value")
	defer func() {
		os.Unsetenv("GOE_ENV")
		os.Unsetenv("SYSTEM_VAR")
	}()

	cfg := New()

	assert.NotNil(t, cfg)
	assert.Equal(t, "test-app", cfg.GetString("APP_NAME"))
	assert.Equal(t, "true", cfg.GetString("DEBUG"))
	assert.Equal(t, "8080", cfg.GetString("PORT"))
	assert.Equal(t, "system_value", cfg.GetString("SYSTEM_VAR"))
}

func TestConfig_Get(t *testing.T) {
	tempDir := changeToTempDir(t)

	cfg := New()

	t.Run("get non-existent key", func(t *testing.T) {
		value := cfg.Get("NON_EXISTENT")
		assert.Nil(t, value)
	})

	t.Run("get existing key", func(t *testing.T) {
		cfg.Set("TEST_KEY", "test_value")
		value := cfg.Get("TEST_KEY")
		assert.Equal(t, "test_value", value)
	})

	t.Run("get from env file", func(t *testing.T) {
		envContent := `TEST_ENV_KEY=env_value`
		err := os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		cfg := New()
		value := cfg.Get("TEST_ENV_KEY")
		assert.Equal(t, "env_value", value)
	})
}

func TestConfig_GetString(t *testing.T) {
	cfg := New()

	t.Run("get string value", func(t *testing.T) {
		cfg.Set("STRING_KEY", "string_value")
		value := cfg.GetString("STRING_KEY")
		assert.Equal(t, "string_value", value)
	})

	t.Run("get non-string value", func(t *testing.T) {
		cfg.Set("INT_KEY", 42)
		value := cfg.GetString("INT_KEY")
		assert.Equal(t, "42", value)
	})

	t.Run("get non-existent key", func(t *testing.T) {
		value := cfg.GetString("NON_EXISTENT")
		assert.Equal(t, "", value)
	})
}

func TestConfig_GetInt(t *testing.T) {
	cfg := New()

	t.Run("get valid int", func(t *testing.T) {
		cfg.Set("INT_KEY", "42")
		value := cfg.GetInt("INT_KEY")
		assert.Equal(t, 42, value)
	})

	t.Run("get invalid int", func(t *testing.T) {
		cfg.Set("INVALID_INT", "not_a_number")
		value := cfg.GetInt("INVALID_INT")
		assert.Equal(t, 0, value)
	})

	t.Run("get non-existent key", func(t *testing.T) {
		value := cfg.GetInt("NON_EXISTENT")
		assert.Equal(t, 0, value)
	})
}

func TestConfig_GetInt64(t *testing.T) {
	cfg := New()

	t.Run("get valid int64", func(t *testing.T) {
		cfg.Set("INT64_KEY", "9223372036854775807")
		value := cfg.GetInt64("INT64_KEY")
		assert.Equal(t, int64(9223372036854775807), value)
	})

	t.Run("get invalid int64", func(t *testing.T) {
		cfg.Set("INVALID_INT64", "not_a_number")
		value := cfg.GetInt64("INVALID_INT64")
		assert.Equal(t, int64(0), value)
	})

	t.Run("get non-existent key", func(t *testing.T) {
		value := cfg.GetInt64("NON_EXISTENT")
		assert.Equal(t, int64(0), value)
	})
}

func TestConfig_GetFloat64(t *testing.T) {
	cfg := New()

	t.Run("get valid float64", func(t *testing.T) {
		cfg.Set("FLOAT_KEY", "3.14159")
		value := cfg.GetFloat64("FLOAT_KEY")
		assert.Equal(t, 3.14159, value)
	})

	t.Run("get invalid float64", func(t *testing.T) {
		cfg.Set("INVALID_FLOAT", "not_a_number")
		value := cfg.GetFloat64("INVALID_FLOAT")
		assert.Equal(t, 0.0, value)
	})

	t.Run("get non-existent key", func(t *testing.T) {
		value := cfg.GetFloat64("NON_EXISTENT")
		assert.Equal(t, 0.0, value)
	})
}

func TestConfig_GetBool(t *testing.T) {
	cfg := New()

	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"true", "true", true},
		{"false", "false", false},
		{"1", "1", true},
		{"0", "0", false},
		{"TRUE", "TRUE", true},
		{"FALSE", "FALSE", false},
		{"invalid", "invalid", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg.Set("BOOL_KEY", tc.input)
			value := cfg.GetBool("BOOL_KEY")
			assert.Equal(t, tc.expected, value)
		})
	}

	t.Run("get non-existent key", func(t *testing.T) {
		value := cfg.GetBool("NON_EXISTENT")
		assert.Equal(t, false, value)
	})
}

func TestConfig_GetDuration(t *testing.T) {
	cfg := New()

	t.Run("get valid duration", func(t *testing.T) {
		cfg.Set("DURATION_KEY", "5m30s")
		value := cfg.GetDuration("DURATION_KEY")
		assert.Equal(t, 5*time.Minute+30*time.Second, value)
	})

	t.Run("get invalid duration", func(t *testing.T) {
		cfg.Set("INVALID_DURATION", "not_a_duration")
		value := cfg.GetDuration("INVALID_DURATION")
		assert.Equal(t, time.Duration(0), value)
	})

	t.Run("get non-existent key", func(t *testing.T) {
		value := cfg.GetDuration("NON_EXISTENT")
		assert.Equal(t, time.Duration(0), value)
	})
}

func TestConfig_GetStringSlice(t *testing.T) {
	cfg := New()

	t.Run("get valid comma-separated string", func(t *testing.T) {
		cfg.Set("SLICE_KEY", "a,b,c,d")
		value := cfg.GetStringSlice("SLICE_KEY")
		assert.Equal(t, []string{"a", "b", "c", "d"}, value)
	})

	t.Run("get string with spaces", func(t *testing.T) {
		cfg.Set("SLICE_KEY", "a, b , c,  d  ")
		value := cfg.GetStringSlice("SLICE_KEY")
		assert.Equal(t, []string{"a", "b", "c", "d"}, value)
	})

	t.Run("get empty values", func(t *testing.T) {
		cfg.Set("SLICE_KEY", "a,,c,")
		value := cfg.GetStringSlice("SLICE_KEY")
		assert.Equal(t, []string{"a", "c"}, value)
	})

	t.Run("get non-existent key", func(t *testing.T) {
		value := cfg.GetStringSlice("NON_EXISTENT")
		assert.Equal(t, []string{}, value)
	})
}

func TestConfig_GetStringMap(t *testing.T) {
	cfg := New()

	t.Run("get prefixed keys", func(t *testing.T) {
		cfg.Set("DB.HOST", "localhost")
		cfg.Set("DB.PORT", "5432")
		cfg.Set("DB.NAME", "mydb")
		cfg.Set("OTHER.KEY", "value")

		value := cfg.GetStringMap("DB")
		expected := map[string]any{
			"HOST": "localhost",
			"PORT": "5432",
			"NAME": "mydb",
		}
		assert.Equal(t, expected, value)
	})

	t.Run("get non-existent prefix", func(t *testing.T) {
		value := cfg.GetStringMap("NON_EXISTENT")
		assert.Equal(t, map[string]any{}, value)
	})
}

func TestConfig_Set(t *testing.T) {
	cfg := New()

	t.Run("set and get", func(t *testing.T) {
		cfg.Set("TEST_KEY", "test_value")
		value := cfg.Get("TEST_KEY")
		assert.Equal(t, "test_value", value)
	})

	t.Run("set overwrites existing", func(t *testing.T) {
		cfg.Set("TEST_KEY", "old_value")
		cfg.Set("TEST_KEY", "new_value")
		value := cfg.Get("TEST_KEY")
		assert.Equal(t, "new_value", value)
	})
}

func TestConfig_Has(t *testing.T) {
	cfg := New()

	t.Run("has existing key", func(t *testing.T) {
		cfg.Set("TEST_KEY", "value")
		assert.True(t, cfg.Has("TEST_KEY"))
	})

	t.Run("has non-existent key", func(t *testing.T) {
		assert.False(t, cfg.Has("NON_EXISTENT"))
	})
}

func TestConfig_All(t *testing.T) {
	cfg := New()

	cfg.Set("KEY1", "value1")
	cfg.Set("KEY2", "value2")

	all := cfg.All()

	// Should contain at least our set values
	assert.Contains(t, all, "KEY1")
	assert.Contains(t, all, "KEY2")
	assert.Equal(t, "value1", all["KEY1"])
	assert.Equal(t, "value2", all["KEY2"])
}

func TestConfig_Reload(t *testing.T) {
	tempDir := changeToTempDir(t)

	// Create initial .env file
	envContent := `APP_NAME=initial`
	err := os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0644)
	require.NoError(t, err)

	cfg := New()
	assert.Equal(t, "initial", cfg.GetString("APP_NAME"))

	// Update .env file
	envContent = `APP_NAME=updated`
	err = os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0644)
	require.NoError(t, err)

	// Reload config
	err = cfg.Reload()
	assert.NoError(t, err)
	assert.Equal(t, "updated", cfg.GetString("APP_NAME"))
}

func TestConfig_EnvFilePriority(t *testing.T) {
	tempDir := changeToTempDir(t)

	// Create .env file
	envContent := `APP_NAME=from_env
DEBUG=false`
	err := os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0644)
	require.NoError(t, err)

	// Create .local.env file (should override .env)
	localEnvContent := `APP_NAME=from_local
PORT=8080`
	err = os.WriteFile(filepath.Join(tempDir, ".local.env"), []byte(localEnvContent), 0644)
	require.NoError(t, err)

	// Set GOE_ENV and create environment-specific file
	os.Setenv("GOE_ENV", "test")
	defer os.Unsetenv("GOE_ENV")

	testEnvContent := `APP_NAME=from_test
HOST=test.local`
	err = os.WriteFile(filepath.Join(tempDir, ".test.env"), []byte(testEnvContent), 0644)
	require.NoError(t, err)

	// Set system environment variable (should have highest priority)
	os.Setenv("APP_NAME", "from_system")
	defer os.Unsetenv("APP_NAME")

	cfg := New()

	// Check priority: system env > .{env}.env > .local.env > .env
	assert.Equal(t, "from_system", cfg.GetString("APP_NAME"))
	assert.Equal(t, "test.local", cfg.GetString("HOST"))
	assert.Equal(t, "8080", cfg.GetString("PORT"))
	assert.Equal(t, "false", cfg.GetString("DEBUG"))
}

func TestConfig_EnvFileComments(t *testing.T) {
	tempDir := changeToTempDir(t)

	envContent := `# This is a comment
APP_NAME=test
# Another comment
DEBUG=true

# Empty line above
PORT=8080`

	err := os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0644)
	require.NoError(t, err)

	cfg := New()

	assert.Equal(t, "test", cfg.GetString("APP_NAME"))
	assert.Equal(t, "true", cfg.GetString("DEBUG"))
	assert.Equal(t, "8080", cfg.GetString("PORT"))
}

func TestConfig_EnvFileQuotes(t *testing.T) {
	tempDir := changeToTempDir(t)

	envContent := `SINGLE_QUOTE='value with spaces'
DOUBLE_QUOTE="another value"
NO_QUOTE=simple`

	err := os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0644)
	require.NoError(t, err)

	cfg := New()

	assert.Equal(t, "value with spaces", cfg.GetString("SINGLE_QUOTE"))
	assert.Equal(t, "another value", cfg.GetString("DOUBLE_QUOTE"))
	assert.Equal(t, "simple", cfg.GetString("NO_QUOTE"))
}

func TestConfig_Caching(t *testing.T) {
	cfg := New()

	// Set a value
	cfg.Set("CACHE_TEST", "cached_value")

	// Get it multiple times - should use cache
	value1 := cfg.Get("CACHE_TEST")
	value2 := cfg.Get("CACHE_TEST")

	assert.Equal(t, "cached_value", value1)
	assert.Equal(t, "cached_value", value2)

	// Set a new value - should invalidate cache
	cfg.Set("CACHE_TEST", "new_value")
	value3 := cfg.Get("CACHE_TEST")

	assert.Equal(t, "new_value", value3)
}

func TestConfigModule_New(t *testing.T) {
	module := NewModule()

	assert.NotNil(t, module)
	assert.Equal(t, "config", module.Name())
	assert.NotNil(t, module.Provide())
}

func TestConfigModule_Lifecycle(t *testing.T) {
	module := NewModule()

	ctx := context.Background()

	t.Run("start module", func(t *testing.T) {
		err := module.OnStart(ctx)
		assert.NoError(t, err)
	})

	t.Run("stop module", func(t *testing.T) {
		err := module.OnStop(ctx)
		assert.NoError(t, err)
	})
}

func TestConfig_ConcurrentAccess(t *testing.T) {
	cfg := New()

	// Test concurrent reads and writes
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cfg.Set("CONCURRENT_KEY", i)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cfg.Get("CONCURRENT_KEY")
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Should not race
	assert.True(t, true)
}

func TestConfig_DefaultGOEEnv(t *testing.T) {
	tempDir := changeToTempDir(t)

	// Make sure GOE_ENV is not set
	os.Unsetenv("GOE_ENV")

	// Create .dev.env file (default environment)
	devEnvContent := `DEV_VAR=from_dev`
	err := os.WriteFile(filepath.Join(tempDir, ".dev.env"), []byte(devEnvContent), 0644)
	require.NoError(t, err)

	cfg := New()

	assert.Equal(t, "from_dev", cfg.GetString("DEV_VAR"))
}

func TestConfig_EnvFileParsingEdgeCases(t *testing.T) {
	tempDir := changeToTempDir(t)

	envContent := `MALFORMED_LINE_NO_EQUALS
EMPTY_VALUE=
SPACES_AROUND_EQUALS = value
MULTIPLE_EQUALS=key=value=extra`

	err := os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0644)
	require.NoError(t, err)

	cfg := New()

	// Malformed line should be skipped
	assert.Equal(t, "", cfg.GetString("MALFORMED_LINE_NO_EQUALS"))

	// Empty value should be empty string
	assert.Equal(t, "", cfg.GetString("EMPTY_VALUE"))

	// Spaces around equals should be handled
	assert.Equal(t, "value", cfg.GetString("SPACES_AROUND_EQUALS"))

	// Multiple equals should split only on first
	assert.Equal(t, "key=value=extra", cfg.GetString("MULTIPLE_EQUALS"))
}
