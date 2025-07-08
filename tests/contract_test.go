package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/fx"
)

// MockApplication is a mock implementation of the Application interface
type MockApplication struct {
	mock.Mock
}

func (m *MockApplication) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockApplication) Version() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockApplication) Environment() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockApplication) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

func (m *MockApplication) Container() *fx.App {
	args := m.Called()
	return args.Get(0).(*fx.App)
}

func (m *MockApplication) IsRunning() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockApplication) Register(options ...fx.Option) error {
	args := m.Called(options)
	return args.Error(0)
}

func (m *MockApplication) AddModule(module contract.Module) error {
	args := m.Called(module)
	return args.Error(0)
}

func (m *MockApplication) AddProvider(provider contract.Provider) error {
	args := m.Called(provider)
	return args.Error(0)
}

func (m *MockApplication) AddInvoker(invoker contract.Invoker) error {
	args := m.Called(invoker)
	return args.Error(0)
}

func (m *MockApplication) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockApplication) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockApplication) Run() {
	m.Called()
}

func TestApplicationInterface(t *testing.T) {
	// This test verifies that MockApplication implements the Application interface
	var _ contract.Application = (*MockApplication)(nil)

	// Create a mock application
	app := new(MockApplication)

	// Set up expectations
	app.On("Name").Return("Test App")
	app.On("Version").Return("1.0.0")
	app.On("Environment").Return("test")
	app.On("Context").Return(context.Background())
	app.On("Container").Return(fx.New())
	app.On("IsRunning").Return(true)
	app.On("Register", mock.Anything).Return(nil)

	// Test the methods
	assert.Equal(t, "Test App", app.Name())
	assert.Equal(t, "1.0.0", app.Version())
	assert.Equal(t, "test", app.Environment())
	assert.NotNil(t, app.Context())
	assert.NotNil(t, app.Container())
	assert.True(t, app.IsRunning())
	assert.Nil(t, app.Register(fx.Options()))

	// Verify expectations
	app.AssertExpectations(t)
}

// MockCache is a mock implementation of the Cache interface
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(key string) (any, error) {
	args := m.Called(key)
	return args.Get(0), args.Error(1)
}

func (m *MockCache) Set(key string, value any, ttl time.Duration) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Forever(key string, value any) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockCache) Forget(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockCache) Flush() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCache) Has(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockCache) Remember(key string, ttl time.Duration, callback func() (any, error)) (any, error) {
	args := m.Called(key, ttl, callback)
	return args.Get(0), args.Error(1)
}

func (m *MockCache) RememberForever(key string, callback func() (any, error)) (any, error) {
	args := m.Called(key, callback)
	return args.Get(0), args.Error(1)
}

func (m *MockCache) Pull(key string) (any, error) {
	args := m.Called(key)
	return args.Get(0), args.Error(1)
}

func (m *MockCache) Add(key string, value any, ttl time.Duration) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Increment(key string, value ...int64) (int64, error) {
	args := m.Called(key, value)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCache) Decrement(key string, value ...int64) (int64, error) {
	args := m.Called(key, value)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCache) Store() contract.CacheStore {
	args := m.Called()
	return args.Get(0).(contract.CacheStore)
}

func (m *MockCache) GetPrefix() string {
	args := m.Called()
	return args.String(0)
}

func TestCacheInterface(t *testing.T) {
	// This test verifies that MockCache implements the Cache interface
	var _ contract.Cache = (*MockCache)(nil)

	// Create a mock cache
	cache := new(MockCache)

	// Set up expectations
	cache.On("Get", "test_key").Return("test_value", nil)
	cache.On("Set", "test_key", "test_value", time.Minute).Return(nil)
	cache.On("Forget", "test_key").Return(nil)
	cache.On("Flush").Return(nil)
	cache.On("Has", "test_key").Return(true)

	// Test the methods
	value, err := cache.Get("test_key")
	assert.NoError(t, err)
	assert.Equal(t, "test_value", value)

	err = cache.Set("test_key", "test_value", time.Minute)
	assert.NoError(t, err)

	err = cache.Forget("test_key")
	assert.NoError(t, err)

	err = cache.Flush()
	assert.NoError(t, err)

	assert.True(t, cache.Has("test_key"))

	// Verify expectations
	cache.AssertExpectations(t)
}

// MockCacheStore is a mock implementation of the CacheStore interface
type MockCacheStore struct {
	mock.Mock
}

func (m *MockCacheStore) Get(key string) ([]byte, error) {
	args := m.Called(key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCacheStore) Set(key string, val []byte, exp time.Duration) error {
	args := m.Called(key, val, exp)
	return args.Error(0)
}

func (m *MockCacheStore) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockCacheStore) Reset() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCacheStore) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestCacheStoreInterface(t *testing.T) {
	// This test verifies that MockCacheStore implements the CacheStore interface
	var _ contract.CacheStore = (*MockCacheStore)(nil)

	// Create a mock cache store
	store := new(MockCacheStore)

	// Set up expectations
	store.On("Get", "test_key").Return([]byte("test_value"), nil)
	store.On("Set", "test_key", []byte("test_value"), time.Minute).Return(nil)
	store.On("Delete", "test_key").Return(nil)
	store.On("Reset").Return(nil)
	store.On("Close").Return(nil)

	// Test the methods
	value, err := store.Get("test_key")
	assert.NoError(t, err)
	assert.Equal(t, []byte("test_value"), value)

	err = store.Set("test_key", []byte("test_value"), time.Minute)
	assert.NoError(t, err)

	err = store.Delete("test_key")
	assert.NoError(t, err)

	err = store.Reset()
	assert.NoError(t, err)

	err = store.Close()
	assert.NoError(t, err)

	// Verify expectations
	store.AssertExpectations(t)
}

// MockCacheManager is a mock implementation of the CacheManager interface
type MockCacheManager struct {
	mock.Mock
}

func (m *MockCacheManager) Store(name ...string) contract.Cache {
	args := m.Called(name)
	return args.Get(0).(contract.Cache)
}

func (m *MockCacheManager) Driver() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCacheManager) Extend(driver string, factory contract.CacheStoreFactory) {
	m.Called(driver, factory)
}

func TestCacheManagerInterface(t *testing.T) {
	// This test verifies that MockCacheManager implements the CacheManager interface
	var _ contract.CacheManager = (*MockCacheManager)(nil)

	// Create a mock cache manager
	manager := new(MockCacheManager)
	store := new(MockCacheStore)
	cache := new(MockCache)

	// Set up expectations
	manager.On("Store", []string{"memory"}).Return(cache)
	manager.On("Driver").Return("memory")
	manager.On("Extend", "memory", mock.Anything).Return()

	// Test the methods
	retrievedCache := manager.Store("memory")
	assert.Equal(t, cache, retrievedCache)

	defaultDriver := manager.Driver()
	assert.Equal(t, "memory", defaultDriver)

	factory := func(config contract.Config) (contract.CacheStore, error) {
		return store, nil
	}
	manager.Extend("memory", factory)

	// Verify expectations
	manager.AssertExpectations(t)
}

// MockConfig is a mock implementation of the Config interface
type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) Get(key string) interface{} {
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

func (m *MockConfig) GetBool(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockConfig) GetFloat64(key string) float64 {
	args := m.Called(key)
	return args.Get(0).(float64)
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

func TestConfigInterface(t *testing.T) {
	// This test verifies that MockConfig implements the Config interface
	var _ contract.Config = (*MockConfig)(nil)

	// Create a mock config
	config := new(MockConfig)

	// Set up expectations
	config.On("Get", "test_key").Return("test_value")
	config.On("GetString", "string_key").Return("string_value")
	config.On("GetInt", "int_key").Return(42)
	config.On("GetInt64", "int64_key").Return(int64(42))
	config.On("GetBool", "bool_key").Return(true)
	config.On("GetFloat64", "float_key").Return(3.14)
	config.On("GetDuration", "duration_key").Return(time.Second)
	config.On("GetStringSlice", "slice_key").Return([]string{"a", "b", "c"})
	config.On("GetStringMap", "map_key").Return(map[string]any{"key": "value"})
	config.On("Has", "test_key").Return(true)
	config.On("All").Return(map[string]any{"key1": "value1", "key2": "value2"})
	config.On("Reload").Return(nil)

	// Test the methods
	assert.Equal(t, "test_value", config.Get("test_key"))
	assert.Equal(t, "string_value", config.GetString("string_key"))
	assert.Equal(t, 42, config.GetInt("int_key"))
	assert.Equal(t, int64(42), config.GetInt64("int64_key"))
	assert.True(t, config.GetBool("bool_key"))
	assert.Equal(t, 3.14, config.GetFloat64("float_key"))
	assert.Equal(t, time.Second, config.GetDuration("duration_key"))
	assert.Equal(t, []string{"a", "b", "c"}, config.GetStringSlice("slice_key"))
	assert.Equal(t, map[string]any{"key": "value"}, config.GetStringMap("map_key"))
	assert.True(t, config.Has("test_key"))
	assert.Equal(t, map[string]any{"key1": "value1", "key2": "value2"}, config.All())
	assert.NoError(t, config.Reload())

	// Verify expectations
	config.AssertExpectations(t)
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

func TestConfigSourceInterface(t *testing.T) {
	// This test verifies that MockConfigSource implements the ConfigSource interface
	var _ contract.ConfigSource = (*MockConfigSource)(nil)

	// Create a mock config source
	source := new(MockConfigSource)

	// Set up expectations
	source.On("Name").Return("test_source")
	source.On("Load").Return(map[string]any{"key": "value"}, nil)
	source.On("Watch", mock.Anything).Return(nil)

	// Test the methods
	assert.Equal(t, "test_source", source.Name())

	data, err := source.Load()
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"key": "value"}, data)

	callback := func(data map[string]any) {}
	err = source.Watch(callback)
	assert.NoError(t, err)

	// Verify expectations
	source.AssertExpectations(t)
}
