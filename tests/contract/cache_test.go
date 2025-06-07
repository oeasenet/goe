package contract_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/contract"
)

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

func TestCacheInterface(t *testing.T) {
	// This test verifies that MockCache implements the Cache interface
	var _ contract.Cache = (*MockCache)(nil)

	// Create a mock cache
	cache := new(MockCache)

	// Set up expectations
	cache.On("Get", "test_key").Return("test_value", nil)
	cache.On("Set", "test_key", "test_value", time.Minute).Return(nil)
	cache.On("Forever", "test_key", "test_value").Return(nil)
	cache.On("Forget", "test_key").Return(nil)
	cache.On("Flush").Return(nil)
	cache.On("Has", "test_key").Return(true)
	cache.On("Remember", "test_key", time.Minute, mock.AnythingOfType("func() (interface {}, error)")).Return("test_value", nil)
	cache.On("RememberForever", "test_key", mock.AnythingOfType("func() (interface {}, error)")).Return("test_value", nil)
	cache.On("Pull", "test_key").Return("test_value", nil)
	cache.On("Add", "test_key", "test_value", time.Minute).Return(nil)
	cache.On("Increment", "test_key", []int64{1}).Return(int64(2), nil)
	cache.On("Decrement", "test_key", []int64{1}).Return(int64(0), nil)
	cache.On("Store").Return(new(MockCacheStore))
	cache.On("GetPrefix").Return("test_prefix")

	// Test the methods
	value, err := cache.Get("test_key")
	assert.Nil(t, err)
	assert.Equal(t, "test_value", value)

	assert.Nil(t, cache.Set("test_key", "test_value", time.Minute))
	assert.Nil(t, cache.Forever("test_key", "test_value"))
	assert.Nil(t, cache.Forget("test_key"))
	assert.Nil(t, cache.Flush())
	assert.True(t, cache.Has("test_key"))

	value, err = cache.Remember("test_key", time.Minute, func() (any, error) {
		return "test_value", nil
	})
	assert.Nil(t, err)
	assert.Equal(t, "test_value", value)

	value, err = cache.RememberForever("test_key", func() (any, error) {
		return "test_value", nil
	})
	assert.Nil(t, err)
	assert.Equal(t, "test_value", value)

	value, err = cache.Pull("test_key")
	assert.Nil(t, err)
	assert.Equal(t, "test_value", value)

	assert.Nil(t, cache.Add("test_key", "test_value", time.Minute))

	newValue, err := cache.Increment("test_key", 1)
	assert.Nil(t, err)
	assert.Equal(t, int64(2), newValue)

	newValue, err = cache.Decrement("test_key", 1)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), newValue)

	assert.NotNil(t, cache.Store())
	assert.Equal(t, "test_prefix", cache.GetPrefix())

	// Verify expectations
	cache.AssertExpectations(t)
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
	assert.Nil(t, err)
	assert.Equal(t, []byte("test_value"), value)

	assert.Nil(t, store.Set("test_key", []byte("test_value"), time.Minute))
	assert.Nil(t, store.Delete("test_key"))
	assert.Nil(t, store.Reset())
	assert.Nil(t, store.Close())

	// Verify expectations
	store.AssertExpectations(t)
}

func TestCacheManagerInterface(t *testing.T) {
	// This test verifies that MockCacheManager implements the CacheManager interface
	var _ contract.CacheManager = (*MockCacheManager)(nil)

	// Create a mock cache manager
	manager := new(MockCacheManager)

	// Set up expectations
	manager.On("Store", []string{"test_store"}).Return(new(MockCache))
	manager.On("Driver").Return("memory")
	manager.On("Extend", "redis", mock.AnythingOfType("contract.CacheStoreFactory")).Return()

	// Test the methods
	assert.NotNil(t, manager.Store("test_store"))
	assert.Equal(t, "memory", manager.Driver())

	// Test Extend with a factory function
	factory := func(config contract.Config) (contract.CacheStore, error) {
		return new(MockCacheStore), nil
	}
	manager.Extend("redis", factory)

	// Verify expectations
	manager.AssertExpectations(t)
}
