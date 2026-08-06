package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypedGet(t *testing.T) {
	c := New(newTTLRecordingStore(), "t", 0)

	t.Run("miss returns zero, false, nil", func(t *testing.T) {
		v, found, err := Get[string](c, "absent")
		require.NoError(t, err)
		assert.False(t, found)
		assert.Empty(t, v)
	})

	t.Run("hit returns value, true", func(t *testing.T) {
		require.NoError(t, c.Set("k", 42, time.Minute))
		v, found, err := Get[int](c, "k")
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, 42, v)
	})

	t.Run("stored zero value is found", func(t *testing.T) {
		require.NoError(t, c.Set("zero", 0, time.Minute))
		v, found, err := Get[int](c, "zero")
		require.NoError(t, err)
		assert.True(t, found)
		assert.Zero(t, v)
	})

	t.Run("type mismatch errors", func(t *testing.T) {
		require.NoError(t, c.Set("s", "text", time.Minute))
		_, _, err := Get[int](c, "s")
		assert.Error(t, err)
	})

	t.Run("struct round trip", func(t *testing.T) {
		type user struct{ Name string }
		require.NoError(t, c.Set("u", user{Name: "tony"}, time.Minute))
		v, found, err := Get[user](c, "u")
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "tony", v.Name)
	})
}

func TestTypedGetOr(t *testing.T) {
	c := New(newTTLRecordingStore(), "t", 0)

	v, err := GetOr(c, "absent", "fallback")
	require.NoError(t, err)
	assert.Equal(t, "fallback", v)

	require.NoError(t, c.Set("k", "stored", time.Minute))
	v, err = GetOr(c, "k", "fallback")
	require.NoError(t, err)
	assert.Equal(t, "stored", v)
}

func TestTypedRemember(t *testing.T) {
	c := New(newTTLRecordingStore(), "t", 0)

	calls := 0
	compute := func() (int, error) { calls++; return 7, nil }

	v, err := Remember(c, "n", time.Minute, compute)
	require.NoError(t, err)
	assert.Equal(t, 7, v)

	v, err = Remember(c, "n", time.Minute, compute)
	require.NoError(t, err)
	assert.Equal(t, 7, v)
	assert.Equal(t, 1, calls, "second call must hit the cache")

	_, err = Remember(c, "err", time.Minute, func() (int, error) { return 0, errors.New("boom") })
	assert.ErrorContains(t, err, "boom")
}

func TestTypedPull(t *testing.T) {
	c := New(newTTLRecordingStore(), "t", 0)
	require.NoError(t, c.Set("k", "v", time.Minute))

	v, found, err := Pull[string](c, "k")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "v", v)

	_, found, err = Pull[string](c, "k")
	require.NoError(t, err)
	assert.False(t, found, "Pull must remove the key")
}
