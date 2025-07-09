package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/contract"
)

// MockCacheStore implements contract.CacheStore for testing
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

// MockConfig implements contract.Config for testing
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

func TestCache_New(t *testing.T) {
	store := &MockCacheStore{}
	prefix := "test"

	cache := New(store, prefix)

	assert.NotNil(t, cache)
	assert.Equal(t, store, cache.Store())
	assert.Equal(t, prefix, cache.GetPrefix())
}

func TestCache_Get(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		jsonValue := []byte(`"hello"`)
		store.On("Get", "test:key").Return(jsonValue, nil)

		value, err := cache.Get("key")
		assert.NoError(t, err)
		assert.Equal(t, "hello", value)

		store.AssertExpectations(t)
	})

	t.Run("key not found", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(nil), nil)

		value, err := cache.Get("key")
		assert.NoError(t, err)
		assert.Nil(t, value)

		store.AssertExpectations(t)
	})

	t.Run("store error", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(nil), errors.New("store error"))

		value, err := cache.Get("key")
		assert.Error(t, err)
		assert.Nil(t, value)
		assert.Contains(t, err.Error(), "store error")

		store.AssertExpectations(t)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		invalidJSON := []byte(`invalid json`)
		store.On("Get", "test:key").Return(invalidJSON, nil)

		value, err := cache.Get("key")
		assert.Error(t, err)
		assert.Nil(t, value)

		store.AssertExpectations(t)
	})
}

func TestCache_Set(t *testing.T) {
	t.Run("successful set", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		expectedData := []byte(`"hello"`)
		ttl := time.Minute

		store.On("Set", "test:key", expectedData, ttl).Return(nil)

		err := cache.Set("key", "hello", ttl)
		assert.NoError(t, err)

		store.AssertExpectations(t)
	})

	t.Run("store error", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Set", "test:key", mock.Anything, mock.Anything).Return(errors.New("store error"))

		err := cache.Set("key", "hello", time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "store error")

		store.AssertExpectations(t)
	})

	t.Run("complex object", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		obj := map[string]any{"name": "test", "value": 42}
		store.On("Set", "test:key", mock.Anything, mock.Anything).Return(nil)

		err := cache.Set("key", obj, time.Minute)
		assert.NoError(t, err)

		store.AssertExpectations(t)
	})
}

func TestCache_Forever(t *testing.T) {
	t.Run("forever sets with zero TTL", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		expectedData := []byte(`"hello"`)

		store.On("Set", "test:key", expectedData, time.Duration(0)).Return(nil)

		err := cache.Forever("key", "hello")
		assert.NoError(t, err)

		store.AssertExpectations(t)
	})
}

func TestCache_Forget(t *testing.T) {
	t.Run("successful forget", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Delete", "test:key").Return(nil)

		err := cache.Forget("key")
		assert.NoError(t, err)

		store.AssertExpectations(t)
	})

	t.Run("store error", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Delete", "test:key").Return(errors.New("delete error"))

		err := cache.Forget("key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete error")

		store.AssertExpectations(t)
	})
}

func TestCache_Flush(t *testing.T) {
	t.Run("successful flush", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Reset").Return(nil)

		err := cache.Flush()
		assert.NoError(t, err)

		store.AssertExpectations(t)
	})

	t.Run("store error", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Reset").Return(errors.New("reset error"))

		err := cache.Flush()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "reset error")

		store.AssertExpectations(t)
	})
}

func TestCache_Has(t *testing.T) {
	t.Run("key exists", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(`"value"`), nil)

		exists := cache.Has("key")
		assert.True(t, exists)

		store.AssertExpectations(t)
	})

	t.Run("key does not exist", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(nil), nil)

		exists := cache.Has("key")
		assert.False(t, exists)

		store.AssertExpectations(t)
	})

	t.Run("store error", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(nil), errors.New("get error"))

		exists := cache.Has("key")
		assert.False(t, exists)

		store.AssertExpectations(t)
	})
}

func TestCache_Remember(t *testing.T) {
	t.Run("value exists in cache", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		jsonValue := []byte(`"cached"`)
		store.On("Get", "test:key").Return(jsonValue, nil)

		called := false
		callback := func() (any, error) {
			called = true
			return "computed", nil
		}

		value, err := cache.Remember("key", time.Minute, callback)
		assert.NoError(t, err)
		assert.Equal(t, "cached", value)
		assert.False(t, called)

		store.AssertExpectations(t)
	})

	t.Run("value not in cache, compute and store", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(nil), nil)
		store.On("Set", "test:key", []byte(`"computed"`), time.Minute).Return(nil)

		called := false
		callback := func() (any, error) {
			called = true
			return "computed", nil
		}

		value, err := cache.Remember("key", time.Minute, callback)
		assert.NoError(t, err)
		assert.Equal(t, "computed", value)
		assert.True(t, called)

		store.AssertExpectations(t)
	})

	t.Run("callback error", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(nil), nil)

		callback := func() (any, error) {
			return nil, errors.New("callback error")
		}

		value, err := cache.Remember("key", time.Minute, callback)
		assert.Error(t, err)
		assert.Nil(t, value)
		assert.Contains(t, err.Error(), "callback error")

		store.AssertExpectations(t)
	})
}

func TestCache_RememberForever(t *testing.T) {
	t.Run("remember forever", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(nil), nil)
		store.On("Set", "test:key", []byte(`"computed"`), time.Duration(0)).Return(nil)

		callback := func() (any, error) {
			return "computed", nil
		}

		value, err := cache.RememberForever("key", callback)
		assert.NoError(t, err)
		assert.Equal(t, "computed", value)

		store.AssertExpectations(t)
	})
}

func TestCache_Pull(t *testing.T) {
	t.Run("pull existing value", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		jsonValue := []byte(`"value"`)
		store.On("Get", "test:key").Return(jsonValue, nil)
		store.On("Delete", "test:key").Return(nil)

		value, err := cache.Pull("key")
		assert.NoError(t, err)
		assert.Equal(t, "value", value)

		store.AssertExpectations(t)
	})

	t.Run("pull non-existing value", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(nil), nil)

		value, err := cache.Pull("key")
		assert.NoError(t, err)
		assert.Nil(t, value)

		store.AssertExpectations(t)
	})

	t.Run("get error", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(nil), errors.New("get error"))

		value, err := cache.Pull("key")
		assert.Error(t, err)
		assert.Nil(t, value)

		store.AssertExpectations(t)
	})
}

func TestCache_Add(t *testing.T) {
	t.Run("add new key", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(nil), nil)
		store.On("Set", "test:key", []byte(`"value"`), time.Minute).Return(nil)

		err := cache.Add("key", "value", time.Minute)
		assert.NoError(t, err)

		store.AssertExpectations(t)
	})

	t.Run("add existing key", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(`"existing"`), nil)

		err := cache.Add("key", "value", time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key already exists")

		store.AssertExpectations(t)
	})
}

func TestCache_Increment(t *testing.T) {
	t.Run("increment new key", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(nil), nil)
		store.On("Set", "test:key", []byte(`1`), time.Duration(0)).Return(nil)

		value, err := cache.Increment("key")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), value)

		store.AssertExpectations(t)
	})

	t.Run("increment existing int64", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(`5`), nil)
		store.On("Set", "test:key", []byte(`8`), time.Duration(0)).Return(nil)

		value, err := cache.Increment("key", 3)
		assert.NoError(t, err)
		assert.Equal(t, int64(8), value)

		store.AssertExpectations(t)
	})

	t.Run("increment existing float64", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(`5.5`), nil)
		store.On("Set", "test:key", []byte(`6`), time.Duration(0)).Return(nil)

		value, err := cache.Increment("key")
		assert.NoError(t, err)
		assert.Equal(t, int64(6), value)

		store.AssertExpectations(t)
	})

	t.Run("increment non-numeric value", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(`"string"`), nil)

		value, err := cache.Increment("key")
		assert.Error(t, err)
		assert.Equal(t, int64(0), value)
		assert.Contains(t, err.Error(), "value is not a number")

		store.AssertExpectations(t)
	})
}

func TestCache_Decrement(t *testing.T) {
	t.Run("decrement existing value", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(`5`), nil)
		store.On("Set", "test:key", []byte(`3`), time.Duration(0)).Return(nil)

		value, err := cache.Decrement("key", 2)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), value)

		store.AssertExpectations(t)
	})

	t.Run("decrement with default", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(`5`), nil)
		store.On("Set", "test:key", []byte(`4`), time.Duration(0)).Return(nil)

		value, err := cache.Decrement("key")
		assert.NoError(t, err)
		assert.Equal(t, int64(4), value)

		store.AssertExpectations(t)
	})
}

func TestCache_PrefixKey(t *testing.T) {
	t.Run("with prefix", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")

		store.On("Get", "test:key").Return([]byte(nil), nil)

		cache.Get("key")

		store.AssertExpectations(t)
	})

	t.Run("without prefix", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "")

		store.On("Get", "key").Return([]byte(nil), nil)

		cache.Get("key")

		store.AssertExpectations(t)
	})
}

func TestCacheManager_New(t *testing.T) {
	config := &MockConfig{}
	manager := NewManager(config)

	assert.NotNil(t, manager)
}

func TestCacheManager_Store(t *testing.T) {
	t.Run("get default store", func(t *testing.T) {
		config := &MockConfig{}
		manager := NewManager(config)

		config.On("GetString", mock.Anything).Return("")
		config.On("All").Return(map[string]any{})

		// Mock factory
		factory := func(config contract.Config) (contract.CacheStore, error) {
			return &MockCacheStore{}, nil
		}
		manager.Extend("memory", factory)

		store := manager.Store()
		assert.NotNil(t, store)
	})

	t.Run("get named store", func(t *testing.T) {
		config := &MockConfig{}
		manager := NewManager(config)

		config.On("GetString", mock.Anything).Return("")
		config.On("All").Return(map[string]any{})

		// Mock factory
		factory := func(config contract.Config) (contract.CacheStore, error) {
			return &MockCacheStore{}, nil
		}
		manager.Extend("memory", factory)

		store := manager.Store("custom")
		assert.NotNil(t, store)
	})
}

func TestCacheManager_Driver(t *testing.T) {
	t.Run("get default driver", func(t *testing.T) {
		config := &MockConfig{}
		manager := NewManager(config)

		config.On("GetString", mock.Anything).Return("")

		driver := manager.Driver()
		assert.Equal(t, "memory", driver) // Fallback to memory
	})

	t.Run("fallback to memory", func(t *testing.T) {
		config := &MockConfig{}
		manager := NewManager(config)

		config.On("GetString", mock.Anything).Return("")

		driver := manager.Driver()
		assert.Equal(t, "memory", driver)
	})
}

func TestCacheManager_Extend(t *testing.T) {
	config := &MockConfig{}
	manager := NewManager(config)

	t.Run("extend with custom driver", func(t *testing.T) {
		factory := func(config contract.Config) (contract.CacheStore, error) {
			return &MockCacheStore{}, nil
		}

		manager.Extend("custom", factory)

		// Should not panic
		assert.NotPanics(t, func() {
			manager.Extend("custom", factory)
		})
	})
}

func TestCacheManager_StoreConfig(t *testing.T) {
	t.Run("driver fallback logic", func(t *testing.T) {
		config := &MockConfig{}
		manager := NewManager(config)

		config.On("GetString", mock.Anything).Return("")
		config.On("All").Return(map[string]any{})

		factory := func(config contract.Config) (contract.CacheStore, error) {
			return &MockCacheStore{}, nil
		}
		manager.Extend("memory", factory)

		store := manager.Store("test")
		assert.NotNil(t, store)
	})
}
