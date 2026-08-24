package log

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// kvConfig is a map-backed contract.Config for tests that need specific keys.
type kvConfig map[string]any

func (c kvConfig) Get(key string) any { return c[key] }
func (c kvConfig) GetString(key string) string {
	if v, ok := c[key].(string); ok {
		return v
	}
	return ""
}
func (c kvConfig) GetInt(key string) int {
	if v, ok := c[key].(int); ok {
		return v
	}
	return 0
}
func (c kvConfig) GetInt64(key string) int64 {
	if v, ok := c[key].(int64); ok {
		return v
	}
	return 0
}
func (c kvConfig) GetFloat64(key string) float64 {
	if v, ok := c[key].(float64); ok {
		return v
	}
	return 0
}
func (c kvConfig) GetBool(key string) bool {
	if v, ok := c[key].(bool); ok {
		return v
	}
	return false
}
func (c kvConfig) GetDuration(key string) time.Duration {
	if v, ok := c[key].(time.Duration); ok {
		return v
	}
	return 0
}
func (c kvConfig) GetStringSlice(key string) []string {
	if v, ok := c[key].([]string); ok {
		return v
	}
	return nil
}
func (c kvConfig) GetStringMap(key string) map[string]any {
	if v, ok := c[key].(map[string]any); ok {
		return v
	}
	return nil
}
func (c kvConfig) Set(key string, value any) { c[key] = value }
func (c kvConfig) Has(key string) bool {
	_, ok := c[key]
	return ok
}
func (c kvConfig) All() map[string]any { return c }
func (c kvConfig) Reload() error       { return nil }

// captureStdout runs fn with os.Stdout redirected to a pipe and returns what fn
// wrote. Loggers must be constructed inside fn: buildLogger captures the
// *os.File value at construction time.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() { os.Stdout = old }()
	fn()
	require.NoError(t, w.Close())
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	return buf.String()
}

// lastJSONLine parses the last non-empty line of out as a JSON object.
func lastJSONLine(t *testing.T, out string) map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	require.NotEmpty(t, lines, "expected at least one log line, got none")
	last := lines[len(lines)-1]
	var entry map[string]any
	require.NoError(t, json.Unmarshal([]byte(last), &entry), "not a JSON log line: %q", last)
	return entry
}

func TestJSONFormat_LowercaseLevel(t *testing.T) {
	out := captureStdout(t, func() {
		logger := New(&defaultLoggerConfig{level: "info", format: "json", output: []string{"console"}})
		logger.Info("hello")
	})

	entry := lastJSONLine(t, out)
	assert.Equal(t, "info", entry["level"], "JSON logs must carry lowercase levels")
}

func TestNewModule_JSON_BaseIdentityFields(t *testing.T) {
	cfg := kvConfig{
		"LOG_FORMAT":  "json",
		"APP_NAME":    "orders-api",
		"APP_VERSION": "1.2.3",
		"OEASE_ENV":   "prod",
	}

	out := captureStdout(t, func() {
		NewModule(cfg).Provide().Info("hello")
	})

	entry := lastJSONLine(t, out)
	assert.Equal(t, "orders-api", entry["service"])
	assert.Equal(t, "prod", entry["env"])
	assert.Equal(t, "1.2.3", entry["version"])
}

func TestNewModule_JSON_EnvFallsBackToGOEEnv(t *testing.T) {
	cfg := kvConfig{
		"LOG_FORMAT": "json",
		"GOE_ENV":    "staging",
	}

	out := captureStdout(t, func() {
		NewModule(cfg).Provide().Info("hello")
	})

	entry := lastJSONLine(t, out)
	assert.Equal(t, "staging", entry["env"], "env should fall back to GOE_ENV when OEASE_ENV is unset")
}

func TestNewModule_JSON_EmptyIdentityFieldsOmitted(t *testing.T) {
	// APP_NAME is set, version and env are not: only service may appear. Unset
	// identity must be omitted, never emitted as an empty string.
	cfg := kvConfig{
		"LOG_FORMAT": "json",
		"APP_NAME":   "orders-api",
	}

	out := captureStdout(t, func() {
		NewModule(cfg).Provide().Info("hello")
	})

	entry := lastJSONLine(t, out)
	assert.Equal(t, "orders-api", entry["service"])
	assert.NotContains(t, entry, "env")
	assert.NotContains(t, entry, "version")
}

func TestNewModule_TextFormat_NoBaseFields(t *testing.T) {
	// Base identity fields are for the machine-readable contract; the dev
	// console stays clean.
	cfg := kvConfig{
		"APP_NAME":    "orders-api",
		"APP_VERSION": "1.2.3",
		"OEASE_ENV":   "prod",
	}

	out := captureStdout(t, func() {
		NewModule(cfg).Provide().Info("hello")
	})

	assert.NotContains(t, out, `"service"`)
	assert.NotContains(t, out, `"version"`)
}

func TestNewModule_JSON_FxLoggerCarriesBaseFields(t *testing.T) {
	cfg := kvConfig{
		"LOG_FORMAT": "json",
		"APP_NAME":   "orders-api",
	}

	out := captureStdout(t, func() {
		// Fx defaults to warn via defaultModuleLevels, so log at warn.
		NewModule(cfg).ProvideZap().Warn("fx event")
	})

	entry := lastJSONLine(t, out)
	assert.Equal(t, "orders-api", entry["service"], "the fx logger must inherit base identity fields")
	assert.Equal(t, "fx", entry["module"])
}
