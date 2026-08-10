package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/contract"
)

// MockAtomicCacheStore implements contract.CacheStore plus the optional
// contract.AtomicCacheStore capability, standing in for stores (like the
// builtin Redis driver) whose backend executes compound operations natively.
type MockAtomicCacheStore struct {
	MockCacheStore
}

func (m *MockAtomicCacheStore) Increment(key string, delta int64) (int64, error) {
	args := m.Called(key, delta)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAtomicCacheStore) SetIfNotExists(key string, val []byte, exp time.Duration) (bool, error) {
	args := m.Called(key, val, exp)
	return args.Bool(0), args.Error(1)
}

func (m *MockAtomicCacheStore) GetDelete(key string) ([]byte, error) {
	args := m.Called(key)
	return args.Get(0).([]byte), args.Error(1)
}

var (
	_ contract.CacheStore       = (*MockAtomicCacheStore)(nil)
	_ contract.AtomicCacheStore = (*MockAtomicCacheStore)(nil)
)

func TestCache_Increment_NativeStore(t *testing.T) {
	t.Run("routes through the store's native increment", func(t *testing.T) {
		store := &MockAtomicCacheStore{}
		cache := New(store, "test", 0)
		store.On("Increment", "test:counter", int64(1)).Return(int64(6), nil)

		newValue, err := cache.Increment("counter")
		assert.NoError(t, err)
		assert.Equal(t, int64(6), newValue)

		store.AssertExpectations(t)
		store.AssertNotCalled(t, "Get", mock.Anything)
		store.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("passes a custom delta", func(t *testing.T) {
		store := &MockAtomicCacheStore{}
		cache := New(store, "test", 0)
		store.On("Increment", "test:counter", int64(5)).Return(int64(15), nil)

		newValue, err := cache.Increment("counter", 5)
		assert.NoError(t, err)
		assert.Equal(t, int64(15), newValue)

		store.AssertExpectations(t)
	})

	t.Run("decrement negates the delta", func(t *testing.T) {
		store := &MockAtomicCacheStore{}
		cache := New(store, "test", 0)
		store.On("Increment", "test:counter", int64(-3)).Return(int64(7), nil)

		newValue, err := cache.Decrement("counter", 3)
		assert.NoError(t, err)
		assert.Equal(t, int64(7), newValue)

		store.AssertExpectations(t)
	})

	t.Run("propagates store errors", func(t *testing.T) {
		store := &MockAtomicCacheStore{}
		cache := New(store, "test", 0)
		store.On("Increment", "test:counter", int64(1)).Return(int64(0), errors.New("boom"))

		newValue, err := cache.Increment("counter")
		assert.EqualError(t, err, "boom")
		assert.Equal(t, int64(0), newValue)

		store.AssertExpectations(t)
	})
}

func TestCache_Add_NativeStore(t *testing.T) {
	t.Run("stores through the native check-and-set", func(t *testing.T) {
		store := &MockAtomicCacheStore{}
		cache := New(store, "test", 0)
		store.On("SetIfNotExists", "test:key", []byte(`"value"`), time.Minute).Return(true, nil)

		err := cache.Add("key", "value", time.Minute)
		assert.NoError(t, err)

		store.AssertExpectations(t)
		store.AssertNotCalled(t, "Get", mock.Anything)
		store.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("reports an existing key with the emulated path's error", func(t *testing.T) {
		store := &MockAtomicCacheStore{}
		cache := New(store, "test", 0)
		store.On("SetIfNotExists", "test:key", []byte(`"value"`), time.Minute).Return(false, nil)

		err := cache.Add("key", "value", time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key already exists")

		store.AssertExpectations(t)
	})

	t.Run("resolves ttl 0 to the configured default", func(t *testing.T) {
		store := &MockAtomicCacheStore{}
		cache := New(store, "test", 45*time.Minute)
		store.On("SetIfNotExists", "test:key", []byte(`"value"`), 45*time.Minute).Return(true, nil)

		err := cache.Add("key", "value", 0)
		assert.NoError(t, err)

		store.AssertExpectations(t)
	})

	t.Run("propagates store errors", func(t *testing.T) {
		store := &MockAtomicCacheStore{}
		cache := New(store, "test", 0)
		store.On("SetIfNotExists", "test:key", []byte(`"value"`), time.Minute).Return(false, errors.New("boom"))

		err := cache.Add("key", "value", time.Minute)
		assert.EqualError(t, err, "boom")

		store.AssertExpectations(t)
	})
}

func TestCache_Pull_NativeStore(t *testing.T) {
	t.Run("consumes through the native get-delete", func(t *testing.T) {
		store := &MockAtomicCacheStore{}
		cache := New(store, "test", 0)
		store.On("GetDelete", "test:key").Return([]byte(`"value"`), nil)

		var value string
		err := cache.Pull("key", &value)
		assert.NoError(t, err)
		assert.Equal(t, "value", value)

		store.AssertExpectations(t)
		store.AssertNotCalled(t, "Get", mock.Anything)
		store.AssertNotCalled(t, "Delete", mock.Anything)
	})

	t.Run("miss keeps nil error and untouched destination", func(t *testing.T) {
		store := &MockAtomicCacheStore{}
		cache := New(store, "test", 0)
		store.On("GetDelete", "test:key").Return([]byte(nil), nil)

		value := "untouched"
		err := cache.Pull("key", &value)
		assert.NoError(t, err)
		assert.Equal(t, "untouched", value)

		store.AssertExpectations(t)
	})

	t.Run("rejects non-pointer destinations before touching the store", func(t *testing.T) {
		store := &MockAtomicCacheStore{}
		cache := New(store, "test", 0)

		var value string
		err := cache.Pull("key", value)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a non-nil pointer")

		store.AssertNotCalled(t, "GetDelete", mock.Anything)
	})

	t.Run("propagates store errors", func(t *testing.T) {
		store := &MockAtomicCacheStore{}
		cache := New(store, "test", 0)
		store.On("GetDelete", "test:key").Return([]byte(nil), errors.New("boom"))

		var value string
		err := cache.Pull("key", &value)
		assert.EqualError(t, err, "boom")

		store.AssertExpectations(t)
	})
}

// The typed helpers wrap the untyped methods, so they must pick up the native
// path transparently — found=false on a consumed-nothing miss, found=true with
// the decoded value on a hit.
func TestTypedPull_NativeStore(t *testing.T) {
	t.Run("hit", func(t *testing.T) {
		store := &MockAtomicCacheStore{}
		c := New(store, "test", 0)
		store.On("GetDelete", "test:otp").Return([]byte(`{"purpose":"signup","email":"a@b.c"}`), nil)

		type stash struct {
			Purpose string `json:"purpose"`
			Email   string `json:"email"`
		}
		v, found, err := Pull[stash](c, "otp")
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, stash{Purpose: "signup", Email: "a@b.c"}, v)

		store.AssertExpectations(t)
	})

	t.Run("miss", func(t *testing.T) {
		store := &MockAtomicCacheStore{}
		c := New(store, "test", 0)
		store.On("GetDelete", "test:otp").Return([]byte(nil), nil)

		v, found, err := Pull[string](c, "otp")
		assert.NoError(t, err)
		assert.False(t, found)
		assert.Empty(t, v)

		store.AssertExpectations(t)
	})
}
