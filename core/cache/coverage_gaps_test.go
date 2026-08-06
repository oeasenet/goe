package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/config"
)

// erroringStore fails every operation, for exercising error propagation.
type erroringStore struct{ err error }

func (s *erroringStore) Get(string) ([]byte, error)              { return nil, s.err }
func (s *erroringStore) Set(string, []byte, time.Duration) error { return s.err }
func (s *erroringStore) Delete(string) error                     { return s.err }
func (s *erroringStore) Reset() error                            { return s.err }
func (s *erroringStore) Close() error                            { return s.err }

func TestIncrement_ValueFormats(t *testing.T) {
	t.Run("float64 stored value truncates", func(t *testing.T) {
		store := newTTLRecordingStore()
		c := New(store, "", 0)
		require.NoError(t, c.Set("f", 2.9, time.Minute))

		n, err := c.Increment("f")
		require.NoError(t, err)
		assert.Equal(t, int64(3), n)
	})

	t.Run("non-numeric stored value errors", func(t *testing.T) {
		store := newTTLRecordingStore()
		c := New(store, "", 0)
		require.NoError(t, c.Set("s", "not a number", time.Minute))

		_, err := c.Increment("s")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a number")
	})

	t.Run("store get error propagates", func(t *testing.T) {
		c := New(&erroringStore{err: errors.New("backend down")}, "", 0)
		_, err := c.Increment("n")
		assert.ErrorContains(t, err, "backend down")
	})
}

func TestTypedHelpers_StoreErrors(t *testing.T) {
	c := New(&erroringStore{err: errors.New("backend down")}, "", 0)

	_, _, err := Get[string](c, "k")
	assert.ErrorContains(t, err, "backend down")

	_, err = GetOr(c, "k", "fb")
	assert.ErrorContains(t, err, "backend down")

	_, _, err = Pull[string](c, "k")
	assert.ErrorContains(t, err, "backend down")

	_, err = Remember(c, "k", time.Minute, func() (string, error) { return "v", nil })
	assert.ErrorContains(t, err, "backend down")
}

// TestTypedHelpers_ThroughModule proves the typed API composes with a
// module-built cache: prefix and default TTL from configuration apply.
func TestTypedHelpers_ThroughModule(t *testing.T) {
	store := newTTLRecordingStore()
	m := newTestModule(t, func(cfg *config.Module) {
		cfg.Provide().Set("CACHE_PREFIX", "typed")
		cfg.Provide().Set("CACHE_TTL", "45m")
	},
		WithDriver("recording"),
		WithCustomDriver("recording", func(contract.Config) (contract.CacheStore, error) { return store, nil }),
	)
	require.NoError(t, m.ValidateConfig())
	c := m.Provide()

	got, err := Remember(c, "n", 0, func() (int, error) { return 9, nil })
	require.NoError(t, err)
	assert.Equal(t, 9, got)
	assert.Equal(t, 45*time.Minute, store.lastExp, "typed Remember must inherit the configured default TTL")

	_, ok := store.data["typed:n"]
	assert.True(t, ok, "typed helpers must inherit the configured prefix")
}

func TestModule_NameAndOnStart(t *testing.T) {
	m := newTestModule(t, nil)
	assert.Equal(t, "cache", m.Name())
	assert.NoError(t, m.OnStart(t.Context()))
}

func TestValidateConfig_DriverSpecificKeys(t *testing.T) {
	cases := []struct {
		name        string
		setup       func(*config.Module)
		expectError bool
	}{
		{
			name: "badger invalid GC interval rejected",
			setup: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_DRIVER", "badger")
				cfg.Provide().Set("CACHE_BADGER_GC_INTERVAL", "-5s")
			},
			expectError: true,
		},
		{
			name: "badger valid GC interval accepted",
			setup: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_DRIVER", "badger")
				cfg.Provide().Set("CACHE_BADGER_GC_INTERVAL", "30s")
			},
		},
		{
			name: "bbolt invalid timeout rejected",
			setup: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_DRIVER", "bbolt")
				cfg.Provide().Set("CACHE_BBOLT_TIMEOUT", "0s")
			},
			expectError: true,
		},
		{
			name: "bbolt valid timeout accepted",
			setup: func(cfg *config.Module) {
				cfg.Provide().Set("CACHE_DRIVER", "bbolt")
				cfg.Provide().Set("CACHE_BBOLT_TIMEOUT", "10s")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestModule(t, tc.setup)
			err := m.ValidateConfig()
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
