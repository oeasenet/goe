package cache

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The badger and bbolt drivers are embedded stores, so unlike redis they are
// exercised end to end: a real database in a temp directory, through the full
// manager -> factory -> contract.Cache path.

func TestBadgerDriver_RoundTrip(t *testing.T) {
	dir := t.TempDir()

	m := NewModule(newMapConfig(), &nopLogger{},
		WithStore("badger"),
		WithBadgerDatabase(filepath.Join(dir, "cache.badger")),
	)
	require.NoError(t, m.ValidateConfig())
	assert.Equal(t, "badger", m.manager.Driver())

	store := m.Provide().Store()
	require.NotNil(t, store)

	require.NoError(t, store.Set("greeting", "hello badger", time.Minute))
	var got string
	require.NoError(t, store.Get("greeting", &got))
	assert.Equal(t, "hello badger", got)

	require.NoError(t, store.Forget("greeting"))
	var missing string
	require.NoError(t, store.Get("greeting", &missing))
	assert.Empty(t, missing)

	require.NoError(t, m.OnStop(t.Context()))
}

func TestBboltDriver_RoundTrip(t *testing.T) {
	dir := t.TempDir()

	m := NewModule(newMapConfig(), &nopLogger{},
		WithStore("bbolt"),
		WithBboltDatabase(filepath.Join(dir, "cache.db")),
		WithBboltBucket("test_bucket"),
	)
	require.NoError(t, m.ValidateConfig())
	assert.Equal(t, "bbolt", m.manager.Driver())

	store := m.Provide().Store()
	require.NotNil(t, store)

	require.NoError(t, store.Set("greeting", "hello bbolt", time.Minute))
	var got string
	require.NoError(t, store.Get("greeting", &got))
	assert.Equal(t, "hello bbolt", got)

	require.NoError(t, m.OnStop(t.Context()))
}

func TestBboltDriver_PersistsAcrossReopen(t *testing.T) {
	// The point of an embedded store over memory: data survives a restart.
	path := filepath.Join(t.TempDir(), "cache.db")

	m1 := NewModule(newMapConfig(), &nopLogger{},
		WithStore("bbolt"), WithBboltDatabase(path))
	store1 := m1.Provide().Store()
	require.NoError(t, store1.Forever("persistent", "still here"))
	require.NoError(t, m1.OnStop(t.Context()))

	m2 := NewModule(newMapConfig(), &nopLogger{},
		WithStore("bbolt"), WithBboltDatabase(path))
	store2 := m2.Provide().Store()
	var got string
	require.NoError(t, store2.Get("persistent", &got))
	assert.Equal(t, "still here", got)
	require.NoError(t, m2.OnStop(t.Context()))
}

func TestBuildBadgerConfig_MapsEnvironment(t *testing.T) {
	cfg := newMapConfig()
	cfg.Set("CACHE_BADGER_DATABASE", "/tmp/custom.badger")
	cfg.Set("CACHE_BADGER_RESET", true)
	cfg.Set("CACHE_BADGER_GC_INTERVAL", 30*time.Second)

	bc := buildBadgerConfig(cfg)
	assert.Equal(t, "/tmp/custom.badger", bc.Database)
	assert.True(t, bc.Reset)
	assert.Equal(t, 30*time.Second, bc.GCInterval)
}

func TestBuildBadgerConfig_Defaults(t *testing.T) {
	bc := buildBadgerConfig(newMapConfig())
	assert.Empty(t, bc.Database, "empty means the storage package's default")
	assert.False(t, bc.Reset)
	assert.Zero(t, bc.GCInterval, "zero means the storage package's default")
}

func TestBuildBboltConfig_MapsEnvironment(t *testing.T) {
	cfg := newMapConfig()
	cfg.Set("CACHE_BBOLT_DATABASE", "/tmp/custom.db")
	cfg.Set("CACHE_BBOLT_BUCKET", "mybucket")
	cfg.Set("CACHE_BBOLT_TIMEOUT", 5*time.Second)
	cfg.Set("CACHE_BBOLT_RESET", true)

	bc := buildBboltConfig(cfg)
	assert.Equal(t, "/tmp/custom.db", bc.Database)
	assert.Equal(t, "mybucket", bc.Bucket)
	assert.Equal(t, 5*time.Second, bc.Timeout)
	assert.True(t, bc.Reset)
}

func TestBboltFactory_OpenFailureIsError(t *testing.T) {
	// The storage packages panic when the database cannot be opened; the
	// factory converts that into a labeled error instead of a raw panic.
	cfg := newMapConfig()
	// A directory cannot be opened as a bbolt database file.
	cfg.Set("CACHE_BBOLT_DATABASE", t.TempDir())
	cfg.Set("CACHE_BBOLT_TIMEOUT", 200*time.Millisecond)

	_, err := BboltStoreFactory(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bbolt")
}

func TestEmbeddedDrivers_Options(t *testing.T) {
	m := NewModule(newMapConfig(), &nopLogger{},
		WithBadgerDatabase("/data/cache.badger"),
		WithBadgerReset(true),
		WithBadgerGCInterval(time.Minute),
		WithBboltDatabase("/data/cache.db"),
		WithBboltBucket("app"),
		WithBboltTimeout(10*time.Second),
		WithBboltReset(true),
	)

	assert.Equal(t, "/data/cache.badger", m.config.GetString("CACHE_BADGER_DATABASE"))
	assert.True(t, m.config.GetBool("CACHE_BADGER_RESET"))
	assert.Equal(t, time.Minute, m.config.GetDuration("CACHE_BADGER_GC_INTERVAL"))
	assert.Equal(t, "/data/cache.db", m.config.GetString("CACHE_BBOLT_DATABASE"))
	assert.Equal(t, "app", m.config.GetString("CACHE_BBOLT_BUCKET"))
	assert.Equal(t, 10*time.Second, m.config.GetDuration("CACHE_BBOLT_TIMEOUT"))
	assert.True(t, m.config.GetBool("CACHE_BBOLT_RESET"))
}

func TestEmbeddedDrivers_OptionValidation(t *testing.T) {
	cases := []struct {
		name string
		opt  Option
	}{
		{"WithBadgerDatabase empty", WithBadgerDatabase("")},
		{"WithBadgerGCInterval zero", WithBadgerGCInterval(0)},
		{"WithBboltDatabase empty", WithBboltDatabase("")},
		{"WithBboltBucket empty", WithBboltBucket("")},
		{"WithBboltTimeout zero", WithBboltTimeout(0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, errs := resolveOverrides([]Option{tc.opt})
			assert.NotEmpty(t, errs)
		})
	}
}
