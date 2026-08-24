package log

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithVersion_WinsOverAppVersion(t *testing.T) {
	cfg := kvConfig{
		"LOG_FORMAT":  "json",
		"APP_VERSION": "1.0.0",
	}

	out := captureStdout(t, func() {
		NewModule(cfg, WithVersion("9.9.9+ldflags")).Provide().Info("hello")
	})

	entry := lastJSONLine(t, out)
	assert.Equal(t, "9.9.9+ldflags", entry["version"], "a code-first version must win over APP_VERSION")
}

func TestWithVersion_EmptyFallsBackToAppVersion(t *testing.T) {
	cfg := kvConfig{
		"LOG_FORMAT":  "json",
		"APP_VERSION": "1.0.0",
	}

	out := captureStdout(t, func() {
		NewModule(cfg, WithVersion("")).Provide().Info("hello")
	})

	entry := lastJSONLine(t, out)
	assert.Equal(t, "1.0.0", entry["version"], "an empty WithVersion must not mask APP_VERSION")
}

func TestWithBaseFields_AddsPairs(t *testing.T) {
	cfg := kvConfig{"LOG_FORMAT": "json"}

	out := captureStdout(t, func() {
		NewModule(cfg, WithBaseFields("region", "us-east-1", "team", "payments")).Provide().Info("hello")
	})

	entry := lastJSONLine(t, out)
	assert.Equal(t, "us-east-1", entry["region"])
	assert.Equal(t, "payments", entry["team"])
}

func TestWithBaseFields_TextFormatSkipped(t *testing.T) {
	// Base fields follow the same rule as the identity fields: JSON only.
	cfg := kvConfig{}

	out := captureStdout(t, func() {
		NewModule(cfg, WithBaseFields("region", "us-east-1")).Provide().Info("hello")
	})

	assert.NotContains(t, out, "us-east-1")
}

func TestWithBaseFields_OddCount_FailsValidateConfig(t *testing.T) {
	module := NewModule(kvConfig{"LOG_FORMAT": "json"}, WithBaseFields("dangling-key"))

	err := module.ValidateConfig()
	require.Error(t, err, "an odd key/value count must abort startup")
	assert.Contains(t, err.Error(), "WithBaseFields")
}

func TestWithBaseFields_NonStringKey_FailsValidateConfig(t *testing.T) {
	module := NewModule(kvConfig{"LOG_FORMAT": "json"}, WithBaseFields(42, "value"))

	err := module.ValidateConfig()
	require.Error(t, err, "a non-string key must abort startup")
	assert.Contains(t, err.Error(), "WithBaseFields")
}

func TestOptions_ValidateConfigStillValidatesEnv(t *testing.T) {
	// Option errors must not mask the existing env validation, and vice versa:
	// a bad LOG_LEVEL still fails with no options involved.
	module := NewModule(kvConfig{"LOG_LEVEL": "loud"})

	err := module.ValidateConfig()
	require.Error(t, err)
}

func TestOptionErrors_AbortOnStart(t *testing.T) {
	// Built-in modules are not registered with the startup validator, so
	// OnStart is where a bad option must abort the application (the job
	// module follows the same pattern).
	module := NewModule(kvConfig{}, WithBaseFields("dangling-key"))

	err := module.OnStart(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "WithBaseFields")
}

func TestValidateConfig_AcceptsTextFormat(t *testing.T) {
	// "text" is the module's own default format; the validator must not
	// reject it.
	module := NewModule(kvConfig{"LOG_FORMAT": "text"})

	assert.NoError(t, module.ValidateConfig())
}
