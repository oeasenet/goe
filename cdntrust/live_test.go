//go:build integration

package cdntrust

import (
	"context"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLive_Providers hits the real endpoints.
//
// The offline tests pin the parsers to payloads captured at the time of writing;
// this one catches the thing those cannot — a provider changing its URL, envelope
// or field names. Requires internet access, which is why it is behind the
// integration tag.
func TestLive_Providers(t *testing.T) {
	for _, provider := range []Provider{Cloudflare, Fastly, BunnyCDN} {
		t.Run(string(provider), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			ranges, err := Fetch(ctx, provider)
			require.NoErrorf(t, err, "%s: endpoint or schema may have changed", provider)
			require.NotEmpty(t, ranges)

			var v4, v6 int
			for _, entry := range ranges {
				// Every entry must be usable by Fiber as an address or prefix.
				if prefix, perr := netip.ParsePrefix(entry); perr == nil {
					if prefix.Addr().Is4() {
						v4++
					} else {
						v6++
					}
					continue
				}
				addr, aerr := netip.ParseAddr(entry)
				require.NoErrorf(t, aerr, "%q is neither an address nor a prefix", entry)
				if addr.Is4() {
					v4++
				} else {
					v6++
				}
			}

			t.Logf("%s: %d entries (%d IPv4, %d IPv6)", provider, len(ranges), v4, v6)
			assert.Positive(t, v4, "%s should publish IPv4 space", provider)
			assert.Positive(t, v6, "%s should publish IPv6 space", provider)

			// A trust list that accidentally contained a private or loopback range
			// would let internal traffic forge client addresses.
			for _, entry := range ranges {
				addr, ok := parseAddrOrPrefixAddr(entry)
				if !ok {
					continue
				}
				assert.Falsef(t, addr.IsLoopback(), "%s published loopback %q", provider, entry)
				assert.Falsef(t, addr.IsPrivate(), "%s published private range %q", provider, entry)
				assert.Falsef(t, addr.IsUnspecified(), "%s published unspecified %q", provider, entry)
			}
		})
	}
}

// TestLive_AllProvidersCombined checks the aggregate path end to end.
func TestLive_AllProvidersCombined(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	opts, err := Options(ctx, Cloudflare, Fastly, BunnyCDN)
	require.NoError(t, err)
	require.Len(t, opts, 2)

	cfg := applyToFiberConfig(t, opts)
	require.True(t, cfg.TrustProxy)

	proxies := cfg.TrustProxyConfig.Proxies
	assert.Greater(t, len(proxies), 50, "three CDNs should yield a substantial list")
	assert.IsIncreasing(t, proxies, "list should be sorted")
	t.Logf("combined trust list: %d entries", len(proxies))

	// Spot-check that a well-known Cloudflare block survived the round trip.
	var found bool
	for _, p := range proxies {
		if strings.HasPrefix(p, "173.245.48.") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected a known Cloudflare range in the combined list")
}

func parseAddrOrPrefixAddr(entry string) (netip.Addr, bool) {
	if prefix, err := netip.ParsePrefix(entry); err == nil {
		return prefix.Addr(), true
	}
	if addr, err := netip.ParseAddr(entry); err == nil {
		return addr, true
	}
	return netip.Addr{}, false
}
