package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAdd_ReturnsErrKeyExists(t *testing.T) {
	t.Run("emulated path", func(t *testing.T) {
		store := &MockCacheStore{}
		cache := New(store, "test", 0)
		store.On("Get", "test:key").Return([]byte(`"there"`), nil)

		err := cache.Add("key", "value", time.Minute)
		assert.True(t, errors.Is(err, ErrKeyExists))
		assert.Equal(t, "key already exists", err.Error(), "the text apps matched on is kept")
	})

	t.Run("native store", func(t *testing.T) {
		store := &MockAtomicCacheStore{}
		cache := New(store, "test", 0)
		store.On("SetIfNotExists", "test:key", mock.Anything, time.Minute).Return(false, nil)

		err := cache.Add("key", "value", time.Minute)
		assert.True(t, errors.Is(err, ErrKeyExists))
		assert.Equal(t, "key already exists", err.Error())
	})
}
