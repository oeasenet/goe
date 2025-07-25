package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func (m *MockConfig) GetStringMapString(key string) map[string]string {
	args := m.Called(key)
	return args.Get(0).(map[string]string)
}

func (m *MockConfig) GetTime(key string) time.Time {
	args := m.Called(key)
	return args.Get(0).(time.Time)
}

func (m *MockConfig) Set(key string, value any) {
	m.Called(key, value)
}

func (m *MockConfig) IsSet(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockConfig) All() map[string]any {
	args := m.Called()
	return args.Get(0).(map[string]any)
}

func (m *MockConfig) Has(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

// Tests
func TestNew(t *testing.T) {
	store := &MockCacheStore{}
	prefix := "test"
	cache := New(store, prefix)

	assert.NotNil(t, cache)
	assert.Equal(t, store, cache.Store())
	assert.Equal(t, prefix, cache.GetPrefix())
}

func TestCache_Get(t *testing.T) {
	t.Run("successful get string", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		jsonValue := []byte(`"hello"`)
		store.On("Get", "test:key").Return(jsonValue, nil)

		var value string
		err := cache.Get("key", &value)
		assert.NoError(t, err)
		assert.Equal(t, "hello", value)

		store.AssertExpectations(t)
	})

	t.Run("successful get struct", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		type testStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		jsonValue := []byte(`{"name":"John","age":30}`)
		store.On("Get", "test:key").Return(jsonValue, nil)

		var value testStruct
		err := cache.Get("key", &value)
		assert.NoError(t, err)
		assert.Equal(t, "John", value.Name)
		assert.Equal(t, 30, value.Age)

		store.AssertExpectations(t)
	})

	t.Run("key not found", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(nil), nil)

		var value string
		err := cache.Get("key", &value)
		assert.NoError(t, err)     // Cache miss is not an error
		assert.Equal(t, "", value) // Should remain zero value

		store.AssertExpectations(t)
	})

	t.Run("store error", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(nil), errors.New("store error"))

		var value string
		err := cache.Get("key", &value)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "store error")

		store.AssertExpectations(t)
	})

	t.Run("unmarshal error", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte("invalid json"), nil)

		var value string
		err := cache.Get("key", &value)
		assert.Error(t, err)

		store.AssertExpectations(t)
	})

	t.Run("non-pointer value", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")

		var value string
		err := cache.Get("key", value) // Not a pointer
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "value must be a non-nil pointer")

		store.AssertExpectations(t)
	})

	t.Run("nil pointer value", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")

		var value *string
		err := cache.Get("key", value) // Nil pointer
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "value must be a non-nil pointer")

		store.AssertExpectations(t)
	})
}

func TestCache_GetWithDefault(t *testing.T) {
	t.Run("key exists", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		jsonValue := []byte(`"cached"`)
		store.On("Get", "test:key").Return(jsonValue, nil)

		var value string
		err := cache.GetWithDefault("key", &value, "default")
		assert.NoError(t, err)
		assert.Equal(t, "cached", value)

		store.AssertExpectations(t)
	})

	t.Run("key not found - use default", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(nil), nil)

		var value string
		err := cache.GetWithDefault("key", &value, "default")
		assert.NoError(t, err)
		assert.Equal(t, "default", value)

		store.AssertExpectations(t)
	})

	t.Run("store error", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(nil), errors.New("store error"))

		var value string
		err := cache.GetWithDefault("key", &value, "default")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "store error")

		store.AssertExpectations(t)
	})

	t.Run("non-pointer value", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")

		var value string
		err := cache.GetWithDefault("key", value, "default") // Not a pointer
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "value must be a non-nil pointer")

		store.AssertExpectations(t)
	})
}

func TestCache_Set(t *testing.T) {
	t.Run("successful set", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Set", "test:key", []byte(`"value"`), time.Minute).Return(nil)

		err := cache.Set("key", "value", time.Minute)
		assert.NoError(t, err)

		store.AssertExpectations(t)
	})

	t.Run("marshal error", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")

		// Create an unmarshalable value
		unmarshalable := make(chan int)
		err := cache.Set("key", unmarshalable, time.Minute)
		assert.Error(t, err)

		store.AssertExpectations(t)
	})

	t.Run("store error", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Set", "test:key", []byte(`"value"`), time.Minute).Return(errors.New("store error"))

		err := cache.Set("key", "value", time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "store error")

		store.AssertExpectations(t)
	})
}

func TestCache_Forever(t *testing.T) {
	store := &MockCacheStore{}
	cache := New(store, "test")
	store.On("Set", "test:key", []byte(`"value"`), time.Duration(0)).Return(nil)

	err := cache.Forever("key", "value")
	assert.NoError(t, err)

	store.AssertExpectations(t)
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
		store.On("Delete", "test:key").Return(errors.New("store error"))

		err := cache.Forget("key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "store error")

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
		store.On("Reset").Return(errors.New("store error"))

		err := cache.Flush()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "store error")

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

	t.Run("key not exists", func(t *testing.T) {
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
		store.On("Get", "test:key").Return([]byte(nil), errors.New("store error"))

		exists := cache.Has("key")
		assert.False(t, exists)

		store.AssertExpectations(t)
	})
}

func TestCache_Remember(t *testing.T) {
	t.Run("value in cache", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		jsonValue := []byte(`"cached"`)
		store.On("Get", "test:key").Return(jsonValue, nil)

		called := false
		callback := func() (any, error) {
			called = true
			return "computed", nil
		}

		var value string
		err := cache.Remember("key", &value, time.Minute, callback)
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

		var value string
		err := cache.Remember("key", &value, time.Minute, callback)
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

		var value string
		err := cache.Remember("key", &value, time.Minute, callback)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "callback error")

		store.AssertExpectations(t)
	})
}

func TestCache_RememberForever(t *testing.T) {
	store := &MockCacheStore{}
	cache := New(store, "test")
	store.On("Get", "test:key").Return([]byte(nil), nil)
	store.On("Set", "test:key", []byte(`"computed"`), time.Duration(0)).Return(nil)

	callback := func() (any, error) {
		return "computed", nil
	}

	var value string
	err := cache.RememberForever("key", &value, callback)
	assert.NoError(t, err)
	assert.Equal(t, "computed", value)

	store.AssertExpectations(t)
}

func TestCache_Pull(t *testing.T) {
	t.Run("successful pull", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		jsonValue := []byte(`"value"`)
		store.On("Get", "test:key").Return(jsonValue, nil)
		store.On("Delete", "test:key").Return(nil)

		var value string
		err := cache.Pull("key", &value)
		assert.NoError(t, err)
		assert.Equal(t, "value", value)

		store.AssertExpectations(t)
	})

	t.Run("key not found", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(nil), nil)

		var value string
		err := cache.Pull("key", &value)
		assert.NoError(t, err)     // Cache miss is not an error
		assert.Equal(t, "", value) // Should remain zero value

		store.AssertExpectations(t)
	})
}

func TestCache_Add(t *testing.T) {
	t.Run("key doesn't exist", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:key").Return([]byte(nil), nil)
		store.On("Set", "test:key", []byte(`"value"`), time.Minute).Return(nil)

		err := cache.Add("key", "value", time.Minute)
		assert.NoError(t, err)

		store.AssertExpectations(t)
	})

	t.Run("key already exists", func(t *testing.T) {
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
	t.Run("increment existing value", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:counter").Return([]byte(`10`), nil)
		store.On("Set", "test:counter", []byte(`11`), time.Duration(0)).Return(nil)

		newValue, err := cache.Increment("counter")
		assert.NoError(t, err)
		assert.Equal(t, int64(11), newValue)

		store.AssertExpectations(t)
	})

	t.Run("increment non-existing value", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:counter").Return([]byte(nil), nil)
		store.On("Set", "test:counter", []byte(`1`), time.Duration(0)).Return(nil)

		newValue, err := cache.Increment("counter")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), newValue)

		store.AssertExpectations(t)
	})

	t.Run("increment with custom value", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:counter").Return([]byte(`10`), nil)
		store.On("Set", "test:counter", []byte(`15`), time.Duration(0)).Return(nil)

		newValue, err := cache.Increment("counter", 5)
		assert.NoError(t, err)
		assert.Equal(t, int64(15), newValue)

		store.AssertExpectations(t)
	})
}

func TestCache_Decrement(t *testing.T) {
	t.Run("decrement existing value", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:counter").Return([]byte(`10`), nil)
		store.On("Set", "test:counter", []byte(`9`), time.Duration(0)).Return(nil)

		newValue, err := cache.Decrement("counter")
		assert.NoError(t, err)
		assert.Equal(t, int64(9), newValue)

		store.AssertExpectations(t)
	})

	t.Run("decrement with custom value", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test")
		store.On("Get", "test:counter").Return([]byte(`10`), nil)
		store.On("Set", "test:counter", []byte(`5`), time.Duration(0)).Return(nil)

		newValue, err := cache.Decrement("counter", 5)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), newValue)

		store.AssertExpectations(t)
	})
}

func TestCache_prefixKey(t *testing.T) {
	t.Run("with prefix", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "app")
		store.On("Get", "app:key").Return([]byte(nil), nil)

		var value string
		cache.Get("key", &value)

		store.AssertExpectations(t)
	})

	t.Run("without prefix", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "")
		store.On("Get", "key").Return([]byte(nil), nil)

		var value string
		cache.Get("key", &value)

		store.AssertExpectations(t)
	})
}
