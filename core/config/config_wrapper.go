package config

import (
	"sync"
	"time"

	"go.oease.dev/goe/v2/contract"
)

// ConfigWrapper wraps a base configuration and allows overriding specific values
type ConfigWrapper struct {
	base      contract.Config
	overrides map[string]any
	mu        sync.RWMutex
}

// NewConfigWrapper creates a new configuration wrapper with overrides
func NewConfigWrapper(base contract.Config, overrides map[string]any) contract.Config {
	if overrides == nil {
		overrides = make(map[string]any)
	}

	return &ConfigWrapper{
		base:      base,
		overrides: overrides,
	}
}

// Get retrieves a configuration value by key, checking overrides first
func (w *ConfigWrapper) Get(key string) any {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Check overrides first
	if value, exists := w.overrides[key]; exists {
		return value
	}

	// Fall back to base config
	return w.base.Get(key)
}

// GetString retrieves a string configuration value
func (w *ConfigWrapper) GetString(key string) string {
	val := w.Get(key)
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	case int:
		return string(rune(v))
	case int64:
		return string(rune(v))
	case float64:
		return string(rune(int(v)))
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return w.base.GetString(key)
	}
}

// GetInt retrieves an integer configuration value
func (w *ConfigWrapper) GetInt(key string) int {
	val := w.Get(key)
	if val == nil {
		return w.base.GetInt(key)
	}

	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		return w.base.GetInt(key) // Let base handle string parsing
	default:
		return w.base.GetInt(key)
	}
}

// GetInt64 retrieves an int64 configuration value
func (w *ConfigWrapper) GetInt64(key string) int64 {
	val := w.Get(key)
	if val == nil {
		return w.base.GetInt64(key)
	}

	switch v := val.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		return w.base.GetInt64(key) // Let base handle string parsing
	default:
		return w.base.GetInt64(key)
	}
}

// GetFloat64 retrieves a float64 configuration value
func (w *ConfigWrapper) GetFloat64(key string) float64 {
	val := w.Get(key)
	if val == nil {
		return w.base.GetFloat64(key)
	}

	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		return w.base.GetFloat64(key) // Let base handle string parsing
	default:
		return w.base.GetFloat64(key)
	}
}

// GetBool retrieves a boolean configuration value
func (w *ConfigWrapper) GetBool(key string) bool {
	val := w.Get(key)
	if val == nil {
		return w.base.GetBool(key)
	}

	switch v := val.(type) {
	case bool:
		return v
	case string:
		return w.base.GetBool(key) // Let base handle string parsing
	default:
		return w.base.GetBool(key)
	}
}

// GetDuration retrieves a time.Duration configuration value
func (w *ConfigWrapper) GetDuration(key string) time.Duration {
	val := w.Get(key)
	if val == nil {
		return w.base.GetDuration(key)
	}

	switch v := val.(type) {
	case time.Duration:
		return v
	case string:
		return w.base.GetDuration(key) // Let base handle string parsing
	default:
		return w.base.GetDuration(key)
	}
}

// GetStringSlice retrieves a string slice configuration value
func (w *ConfigWrapper) GetStringSlice(key string) []string {
	val := w.Get(key)
	if val == nil {
		return w.base.GetStringSlice(key)
	}

	switch v := val.(type) {
	case []string:
		return v
	case string:
		return w.base.GetStringSlice(key) // Let base handle string parsing
	default:
		return w.base.GetStringSlice(key)
	}
}

// GetStringMap retrieves a string map configuration value
func (w *ConfigWrapper) GetStringMap(key string) map[string]any {
	val := w.Get(key)
	if val == nil {
		return w.base.GetStringMap(key)
	}

	switch v := val.(type) {
	case map[string]any:
		return v
	default:
		return w.base.GetStringMap(key)
	}
}

// Set sets a configuration value in the overrides
func (w *ConfigWrapper) Set(key string, value any) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.overrides[key] = value
}

// Has checks if a configuration key exists in overrides or base config
func (w *ConfigWrapper) Has(key string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Check overrides first
	if _, exists := w.overrides[key]; exists {
		return true
	}

	// Fall back to base config
	return w.base.Has(key)
}

// All returns all configuration values merged from base and overrides
func (w *ConfigWrapper) All() map[string]any {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Start with base config
	result := w.base.All()

	// Apply overrides
	for key, value := range w.overrides {
		result[key] = value
	}

	return result
}

// Reload reloads the base configuration
func (w *ConfigWrapper) Reload() error {
	return w.base.Reload()
}

// SetOverride sets a specific override value
func (w *ConfigWrapper) SetOverride(key string, value any) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.overrides[key] = value
}

// GetOverrides returns a copy of the current overrides
func (w *ConfigWrapper) GetOverrides() map[string]any {
	w.mu.RLock()
	defer w.mu.RUnlock()

	result := make(map[string]any)
	for key, value := range w.overrides {
		result[key] = value
	}
	return result
}
