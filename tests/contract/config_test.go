package contract_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/contract"
)

// MockConfig is a mock implementation of the Config interface
type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) Get(key string) any {
	args := m.Called(key)
	return args.Get(0)
}

func (m *MockConfig) GetString(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *MockConfig) GetInt(key string) int {
	args := m.Called(key)
	return args.Int(0)
}

func (m *MockConfig) GetInt64(key string) int64 {
	args := m.Called(key)
	return args.Get(0).(int64)
}

func (m *MockConfig) GetFloat64(key string) float64 {
	args := m.Called(key)
	return args.Get(0).(float64)
}

func (m *MockConfig) GetBool(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockConfig) GetDuration(key string) time.Duration {
	args := m.Called(key)
	return args.Get(0).(time.Duration)
}

func (m *MockConfig) GetStringSlice(key string) []string {
	args := m.Called(key)
	return args.Get(0).([]string)
}

func (m *MockConfig) GetStringMap(key string) map[string]any {
	args := m.Called(key)
	return args.Get(0).(map[string]any)
}

func (m *MockConfig) Set(key string, value any) {
	m.Called(key, value)
}

func (m *MockConfig) Has(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockConfig) All() map[string]any {
	args := m.Called()
	return args.Get(0).(map[string]any)
}

func (m *MockConfig) Reload() error {
	args := m.Called()
	return args.Error(0)
}

// MockConfigSource is a mock implementation of the ConfigSource interface
type MockConfigSource struct {
	mock.Mock
}

func (m *MockConfigSource) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConfigSource) Load() (map[string]any, error) {
	args := m.Called()
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *MockConfigSource) Watch(callback func(map[string]any)) error {
	args := m.Called(callback)
	return args.Error(0)
}

func TestConfigInterface(t *testing.T) {
	// This test verifies that MockConfig implements the Config interface
	var _ contract.Config = (*MockConfig)(nil)

	// Create a mock config
	config := new(MockConfig)

	// Set up expectations
	config.On("Get", "test_key").Return("test_value")
	config.On("GetString", "test_key").Return("test_value")
	config.On("GetInt", "test_key").Return(42)
	config.On("GetInt64", "test_key").Return(int64(42))
	config.On("GetFloat64", "test_key").Return(42.0)
	config.On("GetBool", "test_key").Return(true)
	config.On("GetDuration", "test_key").Return(time.Second)
	config.On("GetStringSlice", "test_key").Return([]string{"test1", "test2"})
	config.On("GetStringMap", "test_key").Return(map[string]any{"key": "value"})
	config.On("Set", "test_key", "test_value").Return()
	config.On("Has", "test_key").Return(true)
	config.On("All").Return(map[string]any{"test_key": "test_value"})
	config.On("Reload").Return(nil)

	// Test the methods
	assert.Equal(t, "test_value", config.Get("test_key"))
	assert.Equal(t, "test_value", config.GetString("test_key"))
	assert.Equal(t, 42, config.GetInt("test_key"))
	assert.Equal(t, int64(42), config.GetInt64("test_key"))
	assert.Equal(t, 42.0, config.GetFloat64("test_key"))
	assert.True(t, config.GetBool("test_key"))
	assert.Equal(t, time.Second, config.GetDuration("test_key"))
	assert.Equal(t, []string{"test1", "test2"}, config.GetStringSlice("test_key"))
	assert.Equal(t, map[string]any{"key": "value"}, config.GetStringMap("test_key"))
	config.Set("test_key", "test_value")
	assert.True(t, config.Has("test_key"))
	assert.Equal(t, map[string]any{"test_key": "test_value"}, config.All())
	assert.Nil(t, config.Reload())

	// Verify expectations
	config.AssertExpectations(t)
}

func TestConfigSourceInterface(t *testing.T) {
	// This test verifies that MockConfigSource implements the ConfigSource interface
	var _ contract.ConfigSource = (*MockConfigSource)(nil)

	// Create a mock config source
	source := new(MockConfigSource)

	// Set up expectations
	source.On("Name").Return("test_source")
	source.On("Load").Return(map[string]any{"test_key": "test_value"}, nil)
	source.On("Watch", mock.AnythingOfType("func(map[string]interface {})")).Return(nil)

	// Test the methods
	assert.Equal(t, "test_source", source.Name())
	data, err := source.Load()
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{"test_key": "test_value"}, data)
	assert.Nil(t, source.Watch(func(data map[string]any) {}))

	// Verify expectations
	source.AssertExpectations(t)
}
