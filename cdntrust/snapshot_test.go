package cdntrust

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fastRetries shortens the backoff for the duration of a test.
func fastRetries(t *testing.T) {
	t.Helper()
	original := retryDelay
	retryDelay = time.Millisecond
	t.Cleanup(func() { retryDelay = original })
}

// flakyEndpoints serves Cloudflare and Fastly; fastly fails with status while
// failing reports true.
func flakyEndpoints(t *testing.T, failing *atomic.Bool, status int) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/cloudflare", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(cloudflareBody))
	})
	mux.HandleFunc("/fastly", func(w http.ResponseWriter, _ *http.Request) {
		if failing.Load() {
			w.WriteHeader(status)
			return
		}
		_, _ = w.Write([]byte(fastlyBody))
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	original := endpoints
	t.Cleanup(func() { endpoints = original })
	endpoints = map[Provider][]string{
		Cloudflare: {server.URL + "/cloudflare"},
		Fastly:     {server.URL + "/fastly"},
	}
}

func TestFetch_RetriesTransientFailures(t *testing.T) {
	fastRetries(t)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(cloudflareBody))
	}))
	t.Cleanup(server.Close)
	original := endpoints
	t.Cleanup(func() { endpoints = original })
	endpoints = map[Provider][]string{Cloudflare: {server.URL}}

	ranges, err := Fetch(context.Background(), Cloudflare)
	require.NoError(t, err)
	assert.Contains(t, ranges, "173.245.48.0/20")
	assert.Equal(t, int32(3), requests.Load())
}

func TestFetch_DoesNotRetryAMalformedBody(t *testing.T) {
	fastRetries(t)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte(`{"broken`))
	}))
	t.Cleanup(server.Close)
	original := endpoints
	t.Cleanup(func() { endpoints = original })
	endpoints = map[Provider][]string{Cloudflare: {server.URL}}

	_, err := Fetch(context.Background(), Cloudflare)
	require.Error(t, err)
	assert.Equal(t, int32(1), requests.Load())
}

func TestLoad_WritesASnapshotAfterSuccess(t *testing.T) {
	fastRetries(t)
	var failing atomic.Bool
	flakyEndpoints(t, &failing, http.StatusServiceUnavailable)
	file := filepath.Join(t.TempDir(), "cdntrust.json")

	cfg, err := Load(context.Background(), Config{Providers: []Provider{Cloudflare, Fastly}, SnapshotFile: file})
	require.NoError(t, err)
	assert.Contains(t, cfg.Proxies, "23.235.32.0/20")

	data, err := os.ReadFile(file)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"fastly"`)
	assert.Contains(t, string(data), "23.235.32.0/20")
}

func TestLoad_UsesTheSnapshotForAProviderThatStaysDown(t *testing.T) {
	fastRetries(t)
	var failing atomic.Bool
	flakyEndpoints(t, &failing, http.StatusBadGateway)
	file := filepath.Join(t.TempDir(), "cdntrust.json")
	providers := []Provider{Cloudflare, Fastly}

	_, err := Load(context.Background(), Config{Providers: providers, SnapshotFile: file})
	require.NoError(t, err)

	failing.Store(true)
	var logs bytes.Buffer
	cfg, err := Load(context.Background(), Config{
		Providers:    providers,
		SnapshotFile: file,
		Logger:       slog.New(slog.NewTextHandler(&logs, nil)),
	})
	require.NoError(t, err, "a provider that is down falls back to its last-known-good ranges")
	assert.Contains(t, cfg.Proxies, "23.235.32.0/20", "fastly from the snapshot")
	assert.Contains(t, cfg.Proxies, "173.245.48.0/20", "cloudflare fetched fresh")
	assert.Contains(t, logs.String(), "level=WARN")
	assert.Contains(t, logs.String(), "provider=fastly")
}

func TestLoad_FailsWithoutSnapshotCoverage(t *testing.T) {
	fastRetries(t)
	var failing atomic.Bool
	flakyEndpoints(t, &failing, http.StatusServiceUnavailable)
	file := filepath.Join(t.TempDir(), "cdntrust.json")

	// The snapshot only knows Cloudflare.
	_, err := Load(context.Background(), Config{Providers: []Provider{Cloudflare}, SnapshotFile: file})
	require.NoError(t, err)

	failing.Store(true)
	_, err = Load(context.Background(), Config{Providers: []Provider{Cloudflare, Fastly}, SnapshotFile: file})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fastly")
}

func TestLoad_BackoffRespectsContext(t *testing.T) {
	original := retryDelay
	retryDelay = 5 * time.Second
	t.Cleanup(func() { retryDelay = original })
	var failing atomic.Bool
	failing.Store(true)
	flakyEndpoints(t, &failing, http.StatusServiceUnavailable)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := Load(ctx, Config{Providers: []Provider{Fastly}})
	require.Error(t, err)
	assert.Less(t, time.Since(start), 2*time.Second, "waiting to retry must stop at the deadline")
}
