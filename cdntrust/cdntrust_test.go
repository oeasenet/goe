package cdntrust

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Payload fragments captured from the live endpoints, so the parsers are tested
// against each provider's real envelope shape rather than an invented one.
const (
	cloudflareBody = `{"result":{"ipv4_cidrs":["173.245.48.0/20","103.21.244.0/22"],` +
		`"ipv6_cidrs":["2400:cb00::/32","2606:4700::/32"],"etag":"38f79d05"},` +
		`"success":true,"errors":[],"messages":[]}`

	fastlyBody = `{"addresses":["23.235.32.0/20","43.249.72.0/22"],` +
		`"ipv6_addresses":["2a04:4e40::/32"]}`

	// Bunny publishes bare addresses, not CIDR blocks.
	bunnyV4Body = `["89.187.188.227","89.187.162.249"]`
	bunnyV6Body = `["2400:52e0:1500::714:1"]`
)

// serveEndpoints points every provider at a local server and restores the real
// URLs afterwards.
func serveEndpoints(t *testing.T, bodies map[string]string) {
	t.Helper()

	mux := http.NewServeMux()
	for path, body := range bodies {
		body := body
		mux.HandleFunc(path, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(body))
		})
	}
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	original := endpoints
	t.Cleanup(func() { endpoints = original })

	endpoints = map[Provider][]string{
		Cloudflare: {server.URL + "/cloudflare"},
		Fastly:     {server.URL + "/fastly"},
		BunnyCDN:   {server.URL + "/bunny4", server.URL + "/bunny6"},
	}
}

func allBodies() map[string]string {
	return map[string]string{
		"/cloudflare": cloudflareBody,
		"/fastly":     fastlyBody,
		"/bunny4":     bunnyV4Body,
		"/bunny6":     bunnyV6Body,
	}
}

func TestFetch_PerProvider(t *testing.T) {
	t.Run("cloudflare returns both address families", func(t *testing.T) {
		serveEndpoints(t, allBodies())
		got, err := Fetch(context.Background(), Cloudflare)
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{
			"173.245.48.0/20", "103.21.244.0/22", "2400:cb00::/32", "2606:4700::/32",
		}, got)
	})

	t.Run("fastly returns both address families", func(t *testing.T) {
		serveEndpoints(t, allBodies())
		got, err := Fetch(context.Background(), Fastly)
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{
			"23.235.32.0/20", "43.249.72.0/22", "2a04:4e40::/32",
		}, got)
	})

	t.Run("bunnycdn merges its two endpoints", func(t *testing.T) {
		serveEndpoints(t, allBodies())
		got, err := Fetch(context.Background(), BunnyCDN)
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{
			"89.187.188.227", "89.187.162.249", "2400:52e0:1500::714:1",
		}, got)
	})
}

func TestFetch_CombinesSortsAndDeduplicates(t *testing.T) {
	serveEndpoints(t, allBodies())

	got, err := Fetch(context.Background(), Cloudflare, Fastly, BunnyCDN)
	require.NoError(t, err)
	assert.Len(t, got, 10, "every entry from all three providers")
	assert.IsIncreasing(t, got, "output must be sorted so config is reproducible")

	t.Run("overlapping providers are deduplicated", func(t *testing.T) {
		// Serve the same payload as two providers to force a collision.
		serveEndpoints(t, map[string]string{
			"/cloudflare": cloudflareBody,
			"/fastly":     `{"addresses":["173.245.48.0/20"],"ipv6_addresses":[]}`,
			"/bunny4":     bunnyV4Body,
			"/bunny6":     bunnyV6Body,
		})
		got, err := Fetch(context.Background(), Cloudflare, Fastly)
		require.NoError(t, err)
		assert.Equal(t, len(got), len(uniqueOf(got)), "no duplicates")
		assert.Contains(t, got, "173.245.48.0/20")
	})
}

func TestFetch_RejectsBadInput(t *testing.T) {
	t.Run("no providers", func(t *testing.T) {
		_, err := Fetch(context.Background())
		assert.ErrorIs(t, err, ErrNoProviders)
	})

	t.Run("unknown provider", func(t *testing.T) {
		_, err := Fetch(context.Background(), Provider("akamai"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "akamai")
	})

	t.Run("non-200 response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		t.Cleanup(server.Close)
		original := endpoints
		t.Cleanup(func() { endpoints = original })
		endpoints = map[Provider][]string{Cloudflare: {server.URL}}

		_, err := Fetch(context.Background(), Cloudflare)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "503")
	})

	t.Run("malformed json", func(t *testing.T) {
		serveEndpoints(t, map[string]string{
			"/cloudflare": `{not json`,
			"/fastly":     fastlyBody, "/bunny4": bunnyV4Body, "/bunny6": bunnyV6Body,
		})
		_, err := Fetch(context.Background(), Cloudflare)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "decoding response")
	})

	t.Run("cloudflare envelope reporting failure with HTTP 200", func(t *testing.T) {
		// Cloudflare signals errors in the body, so status alone is not enough.
		serveEndpoints(t, map[string]string{
			"/cloudflare": `{"result":{},"success":false,"errors":[{"code":1000}]}`,
			"/fastly":     fastlyBody, "/bunny4": bunnyV4Body, "/bunny6": bunnyV6Body,
		})
		_, err := Fetch(context.Background(), Cloudflare)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "api reported failure")
	})

	t.Run("garbage address is rejected, not forwarded to Fiber", func(t *testing.T) {
		serveEndpoints(t, map[string]string{
			"/cloudflare": `{"result":{"ipv4_cidrs":["not-an-ip"],"ipv6_cidrs":[]},"success":true}`,
			"/fastly":     fastlyBody, "/bunny4": bunnyV4Body, "/bunny6": bunnyV6Body,
		})
		_, err := Fetch(context.Background(), Cloudflare)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "neither an IP address nor a CIDR block")
	})

	t.Run("empty list is an error", func(t *testing.T) {
		serveEndpoints(t, map[string]string{
			"/cloudflare": `{"result":{"ipv4_cidrs":[],"ipv6_cidrs":[]},"success":true}`,
			"/fastly":     fastlyBody, "/bunny4": bunnyV4Body, "/bunny6": bunnyV6Body,
		})
		_, err := Fetch(context.Background(), Cloudflare)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no addresses returned")
	})
}

// TestFetch_OneFailureFailsAll checks that a partial trust list is never
// returned: trusting some of the hops is not a safe halfway house.
func TestFetch_OneFailureFailsAll(t *testing.T) {
	serveEndpoints(t, map[string]string{
		"/cloudflare": cloudflareBody,
		"/fastly":     `{"broken`,
		"/bunny4":     bunnyV4Body,
		"/bunny6":     bunnyV6Body,
	})

	got, err := Fetch(context.Background(), Cloudflare, Fastly, BunnyCDN)
	require.Error(t, err)
	assert.Nil(t, got, "no partial list on failure")
	assert.Contains(t, err.Error(), "fastly", "the error should name the provider")
}

func TestFetch_RespectsContext(t *testing.T) {
	t.Run("cancelled context aborts", func(t *testing.T) {
		serveEndpoints(t, allBodies())
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := Fetch(ctx, Cloudflare)
		require.Error(t, err)
	})

	t.Run("slow endpoint hits the deadline rather than hanging", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-time.After(10 * time.Second):
			case <-r.Context().Done():
			}
		}))
		t.Cleanup(server.Close)
		original := endpoints
		t.Cleanup(func() { endpoints = original })
		endpoints = map[Provider][]string{Cloudflare: {server.URL}}

		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		start := time.Now()
		_, err := Fetch(ctx, Cloudflare)
		require.Error(t, err)
		assert.Less(t, time.Since(start), 3*time.Second,
			"the supplied deadline must bound the fetch")
	})
}

func TestTrustProxyConfig(t *testing.T) {
	serveEndpoints(t, allBodies())

	cfg, err := TrustProxyConfig(context.Background(), Cloudflare)
	require.NoError(t, err)
	assert.Len(t, cfg.Proxies, 4)

	// Only the CDN edges are trusted; trusting private ranges as well would let
	// anything inside the network forge a client address.
	assert.False(t, cfg.Private, "Private must not be enabled implicitly")
	assert.False(t, cfg.Loopback, "Loopback must not be enabled implicitly")
	assert.False(t, cfg.LinkLocal, "LinkLocal must not be enabled implicitly")
	assert.False(t, cfg.UnixSocket, "UnixSocket must not be enabled implicitly")
}

// TestOptions_AreAcceptedByTheKernel is the integration point that matters:
// Fiber ignores TrustProxyConfig unless TrustProxy is on, and GOE rejects that
// combination at startup, so Options must return both.
func TestOptions_AreAcceptedByTheKernel(t *testing.T) {
	serveEndpoints(t, allBodies())

	opts, err := Options(context.Background(), Cloudflare, Fastly)
	require.NoError(t, err)
	require.Len(t, opts, 2, "trust-proxy toggle plus the range list")

	cfg := applyToFiberConfig(t, opts)
	assert.True(t, cfg.TrustProxy, "Options must enable trust-proxy mode")
	assert.NotEmpty(t, cfg.TrustProxyConfig.Proxies)
	assert.Contains(t, cfg.TrustProxyConfig.Proxies, "173.245.48.0/20")
}

func TestOptions_PropagatesFetchFailure(t *testing.T) {
	serveEndpoints(t, map[string]string{
		"/cloudflare": `{"broken`,
		"/fastly":     fastlyBody, "/bunny4": bunnyV4Body, "/bunny6": bunnyV6Body,
	})

	opts, err := Options(context.Background(), Cloudflare)
	require.Error(t, err, "a failed fetch must be reported, not silently skipped")
	assert.Nil(t, opts)
}

func uniqueOf(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	var out []string
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
