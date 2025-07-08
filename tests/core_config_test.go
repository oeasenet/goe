package tests

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.oease.dev/goe/v2/core/config"
)

func TestConfigNew(t *testing.T) {
	// Test creating a new config
	cfg := config.New()

	// Verify the config is not nil
	assert.NotNil(t, cfg)
}

func TestConfigGetMethods(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("TEST_STRING", "test_value")
	os.Setenv("TEST_INT", "42")
	os.Setenv("TEST_INT64", "42")
	os.Setenv("TEST_FLOAT", "3.14")
	os.Setenv("TEST_BOOL", "true")
	os.Setenv("TEST_DURATION", "1s")
	os.Setenv("TEST_SLICE", "value1,value2,value3")
	os.Setenv("TEST_MAP.KEY1", "value1")
	os.Setenv("TEST_MAP.KEY2", "value2")
	defer func() {
		os.Unsetenv("TEST_STRING")
		os.Unsetenv("TEST_INT")
		os.Unsetenv("TEST_INT64")
		os.Unsetenv("TEST_FLOAT")
		os.Unsetenv("TEST_BOOL")
		os.Unsetenv("TEST_DURATION")
		os.Unsetenv("TEST_SLICE")
		os.Unsetenv("TEST_MAP.KEY1")
		os.Unsetenv("TEST_MAP.KEY2")
	}()

	// Create a new config
	cfg := config.New()

	// Test Get
	assert.Equal(t, "test_value", cfg.Get("TEST_STRING"))

	// Test GetString
	assert.Equal(t, "test_value", cfg.GetString("TEST_STRING"))
	assert.Equal(t, "", cfg.GetString("NON_EXISTENT_KEY"))

	// Test GetInt
	assert.Equal(t, 42, cfg.GetInt("TEST_INT"))
	assert.Equal(t, 0, cfg.GetInt("NON_EXISTENT_KEY"))

	// Test GetInt64
	assert.Equal(t, int64(42), cfg.GetInt64("TEST_INT64"))
	assert.Equal(t, int64(0), cfg.GetInt64("NON_EXISTENT_KEY"))

	// Test GetFloat64
	assert.Equal(t, 3.14, cfg.GetFloat64("TEST_FLOAT"))
	assert.Equal(t, 0.0, cfg.GetFloat64("NON_EXISTENT_KEY"))

	// Test GetBool
	assert.True(t, cfg.GetBool("TEST_BOOL"))
	assert.False(t, cfg.GetBool("NON_EXISTENT_KEY"))

	// Test GetDuration
	assert.Equal(t, time.Second, cfg.GetDuration("TEST_DURATION"))
	assert.Equal(t, time.Duration(0), cfg.GetDuration("NON_EXISTENT_KEY"))

	// Test GetStringSlice
	assert.Equal(t, []string{"value1", "value2", "value3"}, cfg.GetStringSlice("TEST_SLICE"))
	assert.Equal(t, []string{}, cfg.GetStringSlice("NON_EXISTENT_KEY"))

	// Test GetStringMap
	assert.Equal(t, map[string]any{"KEY1": "value1", "KEY2": "value2"}, cfg.GetStringMap("TEST_MAP"))
}

func TestConfigSetAndHas(t *testing.T) {
	// Create a new config
	cfg := config.New()

	// Test Set and Has
	assert.False(t, cfg.Has("TEST_KEY"))
	cfg.Set("TEST_KEY", "test_value")
	assert.True(t, cfg.Has("TEST_KEY"))
	assert.Equal(t, "test_value", cfg.GetString("TEST_KEY"))
}

func TestConfigAll(t *testing.T) {
	// Create a new config
	cfg := config.New()

	// Set some values
	cfg.Set("TEST_KEY1", "value1")
	cfg.Set("TEST_KEY2", "value2")

	// Test All
	all := cfg.All()
	assert.NotNil(t, all)
	assert.Contains(t, all, "TEST_KEY1")
	assert.Contains(t, all, "TEST_KEY2")
	assert.Equal(t, "value1", all["TEST_KEY1"])
	assert.Equal(t, "value2", all["TEST_KEY2"])
}

func TestConfigReload(t *testing.T) {
	// Create a new config
	cfg := config.New()

	// Set a value
	cfg.Set("TEST_KEY", "old_value")
	assert.Equal(t, "old_value", cfg.GetString("TEST_KEY"))

	// Set environment variable
	os.Setenv("TEST_KEY", "new_value")
	defer os.Unsetenv("TEST_KEY")

	// Reload config
	err := cfg.Reload()
	assert.Nil(t, err)

	// Verify the value was updated
	assert.Equal(t, "new_value", cfg.GetString("TEST_KEY"))
}

func TestConfigModule(t *testing.T) {
	// Create a new config module
	module := config.NewModule()

	// Verify the module name
	assert.Equal(t, "config", module.Name())

	// Verify the module provides a config
	cfg := module.Provide()
	assert.NotNil(t, cfg)

	// Test OnStart and OnStop (these should not return errors)
	assert.Nil(t, module.OnStart(nil))
	assert.Nil(t, module.OnStop(nil))
}

func TestConfigEnvFileLoading(t *testing.T) {
	// Create temporary .env file
	envContent := `
TEST_ENV_KEY=test_env_value
# This is a comment
TEST_ENV_INT=42
TEST_ENV_BOOL=true
`
	err := os.WriteFile(".env", []byte(envContent), 0644)
	assert.Nil(t, err)
	defer os.Remove(".env")

	// Create a new config
	cfg := config.New()

	// Verify values from .env file were loaded
	assert.Equal(t, "test_env_value", cfg.GetString("TEST_ENV_KEY"))
	assert.Equal(t, 42, cfg.GetInt("TEST_ENV_INT"))
	assert.True(t, cfg.GetBool("TEST_ENV_BOOL"))
}
