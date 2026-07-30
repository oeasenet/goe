// Package cdntrust fetches the published edge IP ranges of CDN providers so they
// can be configured as trusted proxies.
//
// This package is entirely optional. Nothing in GOE imports it, so neither the
// network calls nor this code end up in your binary unless you use it.
//
// # Why you need this
//
// When your app sits behind a CDN, every request arrives from the CDN's edge, so
// the client address is in X-Forwarded-For rather than the socket. Fiber only
// honours that header for addresses you have declared trustworthy — otherwise
// anyone could forge it and spoof their IP. This package supplies that list from
// each provider's own published source.
//
// # Fetching happens once, at startup, and can stop the app from starting
//
// Ranges are fetched when you call Fetch, Options or TrustProxyConfig — normally
// during construction, before goe.Run. There is no caching, no retry and no
// background refresh: one request per provider, then done.
//
// That means a network failure surfaces as an error you must handle, and the
// usual handling is to abort startup:
//
//	cdnOpts, err := cdntrust.Options(ctx, cdntrust.Cloudflare)
//	if err != nil {
//	    log.Fatalf("cdn trust: %v", err) // do not start with an unknown trust set
//	}
//
// That trade-off is deliberate. Booting with an *unknown* set of trusted proxies
// is worse than not booting: either the header is ignored and every client looks
// like the CDN, or a partial list silently trusts the wrong hops. If you would
// rather tolerate a failed fetch, pin the ranges in configuration instead and use
// goehttp.WithTrustProxyConfig directly.
//
// Because the list is a point-in-time snapshot, restart or redeploy periodically
// to pick up provider changes.
package cdntrust

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"slices"
	"time"

	"github.com/gofiber/fiber/v3"
	goehttp "go.oease.dev/goe/v2/core/http"
	"golang.org/x/sync/errgroup"
)

// Provider identifies a CDN whose edge ranges can be fetched.
type Provider string

const (
	// Cloudflare fetches from https://api.cloudflare.com/client/v4/ips
	Cloudflare Provider = "cloudflare"

	// Fastly fetches from https://api.fastly.com/public-ip-list
	Fastly Provider = "fastly"

	// BunnyCDN fetches from https://bunnycdn.com/api/system/edgeserverlist.
	//
	// Bunny publishes individual edge addresses rather than CIDR blocks, so this
	// returns several hundred entries. Fiber accepts bare addresses alongside
	// CIDRs, so they are usable as-is.
	BunnyCDN Provider = "bunnycdn"
)

// DefaultTimeout bounds the whole fetch when the supplied context has no
// deadline of its own. Without it a hung endpoint would hang startup
// indefinitely, which is a worse failure than a clear timeout.
const DefaultTimeout = 15 * time.Second

// ErrNoProviders is returned when Fetch is called without any provider.
var ErrNoProviders = errors.New("cdntrust: no providers specified")

// endpoints maps each provider to its source URLs. It is a variable so tests can
// point at local servers instead of the live APIs.
var endpoints = map[Provider][]string{
	Cloudflare: {"https://api.cloudflare.com/client/v4/ips"},
	Fastly:     {"https://api.fastly.com/public-ip-list"},
	BunnyCDN: {
		"https://bunnycdn.com/api/system/edgeserverlist",
		"https://bunnycdn.com/api/system/edgeserverlist/ipv6",
	},
}

// Fetcher fetches edge ranges. The zero value is ready to use.
type Fetcher struct {
	// HTTPClient performs the requests. When nil, a client bounded by
	// DefaultTimeout is used.
	HTTPClient *http.Client
}

// Fetch returns the union of the given providers' edge ranges, sorted and
// deduplicated so the resulting configuration is reproducible.
//
// Every entry is checked to parse as an IP address or CIDR block; anything else
// is an error rather than being silently passed to Fiber. Providers are fetched
// concurrently, and the first failure fails the whole call — a partial trust list
// is not a safe fallback.
func Fetch(ctx context.Context, providers ...Provider) ([]string, error) {
	return (&Fetcher{}).Fetch(ctx, providers...)
}

// Fetch is the method form of the package-level Fetch.
func (f *Fetcher) Fetch(ctx context.Context, providers ...Provider) ([]string, error) {
	if len(providers) == 0 {
		return nil, ErrNoProviders
	}
	for _, p := range providers {
		if _, ok := endpoints[p]; !ok {
			return nil, fmt.Errorf("cdntrust: unknown provider %q", p)
		}
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultTimeout)
		defer cancel()
	}

	client := f.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: DefaultTimeout}
	}

	results := make([][]string, len(providers))
	group, groupCtx := errgroup.WithContext(ctx)
	for i, provider := range providers {
		i, provider := i, provider
		group.Go(func() error {
			ranges, err := fetchProvider(groupCtx, client, provider)
			if err != nil {
				return fmt.Errorf("cdntrust: %s: %w", provider, err)
			}
			results[i] = ranges
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}

	var all []string
	for _, ranges := range results {
		all = append(all, ranges...)
	}
	if len(all) == 0 {
		return nil, errors.New("cdntrust: providers returned no ranges")
	}

	slices.Sort(all)
	return slices.Compact(all), nil
}

// TrustProxyConfig fetches the providers' ranges and returns them as a
// fiber.TrustProxyConfig, ready to pass to goehttp.WithTrustProxyConfig.
//
// Only the CDN ranges are trusted. LinkLocal, Loopback, Private and UnixSocket
// are left off deliberately: trusting a private range in addition to the CDN
// would let anything inside your network forge a client address. Set those fields
// yourself if your topology needs them.
func TrustProxyConfig(ctx context.Context, providers ...Provider) (fiber.TrustProxyConfig, error) {
	ranges, err := Fetch(ctx, providers...)
	if err != nil {
		return fiber.TrustProxyConfig{}, err
	}
	return fiber.TrustProxyConfig{Proxies: ranges}, nil
}

// Options returns the goehttp options that trust the given providers' edges.
//
// It enables trust-proxy mode and installs the fetched ranges. Fiber ignores
// TrustProxyConfig unless TrustProxy is on, so both are returned together:
//
//	cdnOpts, err := cdntrust.Options(ctx, cdntrust.Cloudflare, cdntrust.Fastly)
//	if err != nil {
//	    log.Fatalf("cdn trust: %v", err)
//	}
//
//	goe.New(goe.Options{
//	    HTTP: append(cdnOpts, goehttp.WithPort(8080)),
//	})
//
// GOE defaults ProxyHeader to X-Forwarded-For, which these CDNs set, so client
// addresses resolve without further configuration. Use goehttp.WithProxyHeader
// only for a provider-specific header such as Cloudflare's CF-Connecting-IP.
func Options(ctx context.Context, providers ...Provider) ([]goehttp.Option, error) {
	cfg, err := TrustProxyConfig(ctx, providers...)
	if err != nil {
		return nil, err
	}
	return []goehttp.Option{
		goehttp.WithTrustProxy(true),
		goehttp.WithTrustProxyConfig(cfg),
	}, nil
}

// fetchProvider retrieves and parses every source for one provider.
func fetchProvider(ctx context.Context, client *http.Client, provider Provider) ([]string, error) {
	var out []string
	for _, url := range endpoints[provider] {
		body, err := get(ctx, client, url)
		if err != nil {
			return nil, err
		}

		var ranges []string
		switch provider {
		case Cloudflare:
			ranges, err = parseCloudflare(body)
		case Fastly:
			ranges, err = parseFastly(body)
		case BunnyCDN:
			ranges, err = parseStringArray(body)
		default:
			return nil, fmt.Errorf("unknown provider %q", provider)
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", url, err)
		}

		validated, err := validate(ranges)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", url, err)
		}
		out = append(out, validated...)
	}
	return out, nil
}

func get(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}

	// Cap the read: these lists are a few KB, and an unbounded read from a
	// remote source is an easy way to exhaust memory at startup. Reading one byte
	// past the cap lets an oversized body be reported rather than truncated.
	const maxBody = 4 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxBody {
		return nil, fmt.Errorf("response exceeds %d bytes", maxBody)
	}
	return body, nil
}

// cloudflareResponse mirrors https://api.cloudflare.com/client/v4/ips, which
// wraps its payload in the standard Cloudflare envelope.
type cloudflareResponse struct {
	Result struct {
		IPv4CIDRs []string `json:"ipv4_cidrs"`
		IPv6CIDRs []string `json:"ipv6_cidrs"`
	} `json:"result"`
	Success bool     `json:"success"`
	Errors  []any    `json:"errors"`
	Message []string `json:"messages"`
}

func parseCloudflare(body []byte) ([]string, error) {
	var resp cloudflareResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	// The envelope reports failure with HTTP 200, so status alone is not enough.
	if !resp.Success {
		return nil, fmt.Errorf("api reported failure: %v", resp.Errors)
	}
	return append(slices.Clone(resp.Result.IPv4CIDRs), resp.Result.IPv6CIDRs...), nil
}

// fastlyResponse mirrors https://api.fastly.com/public-ip-list.
type fastlyResponse struct {
	Addresses     []string `json:"addresses"`
	IPv6Addresses []string `json:"ipv6_addresses"`
}

func parseFastly(body []byte) ([]string, error) {
	var resp fastlyResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return append(slices.Clone(resp.Addresses), resp.IPv6Addresses...), nil
}

// parseStringArray handles endpoints that return a bare JSON array of addresses,
// which is how BunnyCDN publishes its edge list.
func parseStringArray(body []byte) ([]string, error) {
	var out []string
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return out, nil
}

// validate rejects anything that is not a usable address or prefix. A malformed
// entry must not reach Fiber's trust list, where it would silently never match
// and leave a gap in the set of trusted hops.
func validate(entries []string) ([]string, error) {
	if len(entries) == 0 {
		return nil, errors.New("no addresses returned")
	}
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		if _, err := netip.ParsePrefix(entry); err == nil {
			out = append(out, entry)
			continue
		}
		if _, err := netip.ParseAddr(entry); err == nil {
			out = append(out, entry)
			continue
		}
		return nil, fmt.Errorf("%q is neither an IP address nor a CIDR block", entry)
	}
	return out, nil
}
