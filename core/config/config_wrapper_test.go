package config

import (
	"maps"
	"testing"
	"time"
)

// mockConfig is a simple mock implementation of the Config interface for testing
type mockConfig struct {
	data map[string]any
}

func newMockConfig() *mockConfig {
	return &mockConfig{
		data: make(map[string]any),
	}
}

func (m *mockConfig) Get(key string) any {
	return m.data[key]
}

func (m *mockConfig) GetString(key string) string {
	val, ok := m.data[key]
	if !ok {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

func (m *mockConfig) GetInt(key string) int {
	val, ok := m.data[key]
	if !ok {
		return 0
	}
	if i, ok := val.(int); ok {
		return i
	}
	return 0
}

func (m *mockConfig) GetInt64(key string) int64 {
	val, ok := m.data[key]
	if !ok {
		return 0
	}
	if i, ok := val.(int64); ok {
		return i
	}
	return 0
}

func (m *mockConfig) GetFloat64(key string) float64 {
	val, ok := m.data[key]
	if !ok {
		return 0
	}
	if f, ok := val.(float64); ok {
		return f
	}
	return 0
}

func (m *mockConfig) GetBool(key string) bool {
	val, ok := m.data[key]
	if !ok {
		return false
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}

func (m *mockConfig) GetDuration(key string) time.Duration {
	val, ok := m.data[key]
	if !ok {
		return 0
	}
	if d, ok := val.(time.Duration); ok {
		return d
	}
	return 0
}

func (m *mockConfig) GetStringSlice(key string) []string {
	val, ok := m.data[key]
	if !ok {
		return []string{}
	}
	if s, ok := val.([]string); ok {
		return s
	}
	return []string{}
}

func (m *mockConfig) GetStringMap(key string) map[string]any {
	val, ok := m.data[key]
	if !ok {
		return make(map[string]any)
	}
	if m, ok := val.(map[string]any); ok {
		return m
	}
	return make(map[string]any)
}

func (m *mockConfig) Set(key string, value any) {
	m.data[key] = value
}

func (m *mockConfig) Has(key string) bool {
	_, ok := m.data[key]
	return ok
}

func (m *mockConfig) All() map[string]any {
	result := make(map[string]any)
	maps.Copy(result, m.data)
	return result
}

func (m *mockConfig) Reload() error {
	return nil
}

func TestConfigWrapper_Get(t *testing.T) {
	base := newMockConfig()
	base.Set("base_key", "base_value")

	overrides := map[string]any{
		"override_key": "override_value",
		"base_key":     "overridden_value",
	}

	wrapper := NewConfigWrapper(base, overrides)

	// Test override takes precedence
	if val := wrapper.Get("base_key"); val != "overridden_value" {
		t.Errorf("Expected override value, got %v", val)
	}

	// Test override key
	if val := wrapper.Get("override_key"); val != "override_value" {
		t.Errorf("Expected override_value, got %v", val)
	}

	// Test fallback to base
	base.Set("base_only", "base_only_value")
	if val := wrapper.Get("base_only"); val != "base_only_value" {
		t.Errorf("Expected base_only_value, got %v", val)
	}
}

func TestConfigWrapper_GetInt(t *testing.T) {
	base := newMockConfig()
	base.Set("base_port", 8080)

	overrides := map[string]any{
		"override_port": 3000,
		"base_port":     9000,
	}

	wrapper := NewConfigWrapper(base, overrides)

	// Test override takes precedence
	if val := wrapper.GetInt("base_port"); val != 9000 {
		t.Errorf("Expected 9000, got %d", val)
	}

	// Test override key
	if val := wrapper.GetInt("override_port"); val != 3000 {
		t.Errorf("Expected 3000, got %d", val)
	}

	// Test fallback to base
	base.Set("base_only_port", 8888)
	if val := wrapper.GetInt("base_only_port"); val != 8888 {
		t.Errorf("Expected 8888, got %d", val)
	}
}

func TestConfigWrapper_Has(t *testing.T) {
	base := newMockConfig()
	base.Set("base_key", "base_value")

	overrides := map[string]any{
		"override_key": "override_value",
	}

	wrapper := NewConfigWrapper(base, overrides)

	// Test override key exists
	if !wrapper.Has("override_key") {
		t.Error("Expected override_key to exist")
	}

	// Test base key exists
	if !wrapper.Has("base_key") {
		t.Error("Expected base_key to exist")
	}

	// Test non-existent key
	if wrapper.Has("non_existent") {
		t.Error("Expected non_existent to not exist")
	}
}

func TestConfigWrapper_All(t *testing.T) {
	base := newMockConfig()
	base.Set("base_key", "base_value")
	base.Set("common_key", "base_common")

	overrides := map[string]any{
		"override_key": "override_value",
		"common_key":   "override_common",
	}

	wrapper := NewConfigWrapper(base, overrides)
	all := wrapper.All()

	// Test override takes precedence
	if val := all["common_key"]; val != "override_common" {
		t.Errorf("Expected override_common, got %v", val)
	}

	// Test override key exists
	if val := all["override_key"]; val != "override_value" {
		t.Errorf("Expected override_value, got %v", val)
	}

	// Test base key exists
	if val := all["base_key"]; val != "base_value" {
		t.Errorf("Expected base_value, got %v", val)
	}

	// Test expected number of keys
	if len(all) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(all))
	}
}

func TestConfigWrapper_SetOverride(t *testing.T) {
	base := newMockConfig()
	base.Set("base_key", "base_value")

	wrapper := NewConfigWrapper(base, nil).(*ConfigWrapper)

	// Test setting override
	wrapper.SetOverride("new_key", "new_value")

	if val := wrapper.Get("new_key"); val != "new_value" {
		t.Errorf("Expected new_value, got %v", val)
	}

	// Test override takes precedence over base
	wrapper.SetOverride("base_key", "overridden_value")

	if val := wrapper.Get("base_key"); val != "overridden_value" {
		t.Errorf("Expected overridden_value, got %v", val)
	}
}
