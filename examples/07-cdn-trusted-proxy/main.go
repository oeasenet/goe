// Package main demonstrates trusting a CDN's edge IPs so that Ctx.IP() reports
// the real client address instead of the CDN's.
//
// This example shows:
// - Fetching published edge ranges with the optional cdntrust helper
// - Failing startup when the fetch fails, rather than booting untrusted
// - Composing the resulting options with the rest of your HTTP config
// - Relying on the default X-Forwarded-For client-IP header, and when to override
//
// Run (requires internet access):
//
//	go run main.go
//
// Test:
//
//	curl -s http://localhost:8080/whoami
//	curl -s -H 'X-Forwarded-For: 203.0.113.9' http://localhost:8080/whoami
//
// The second call still reports your loopback address: 127.0.0.1 is not a
// trusted proxy, so the forged header is correctly ignored. That is the whole
// point — only hops in the trusted list may rewrite the client address.
package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/cdntrust"
	"go.oease.dev/goe/v2/contract"
	goehttp "go.oease.dev/goe/v2/core/http"
)

func main() {
	// Bound the fetch. Without a deadline cdntrust applies its own
	// (cdntrust.DefaultTimeout) so a hung endpoint cannot hang startup forever.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// One-shot fetch, at boot. There is no cache, retry or refresh: redeploy
	// periodically to pick up provider changes.
	cdnOpts, err := cdntrust.Options(ctx, cdntrust.Cloudflare, cdntrust.Fastly)
	if err != nil {
		// Deliberate fail-fast. Starting with an unknown trust set is worse than
		// not starting: the app would either ignore X-Forwarded-For entirely, or
		// trust a partial list of hops.
		log.Fatalf("cdn trust: %v", err)
	}

	_ = goe.New(goe.Options{
		HTTP: append(cdnOpts,
			goehttp.WithPort(8080),
			// No WithProxyHeader needed: GOE defaults it to X-Forwarded-For,
			// which Cloudflare, Fastly and Bunny all set. Override only for a
			// provider-specific header — e.g. WithProxyHeader("CF-Connecting-IP")
			// when Cloudflare is the sole edge.
		),
		Invokers: []any{registerRoutes},
	})

	goe.Run()
}

func registerRoutes(kernel contract.HTTPKernel, logger contract.Logger) {
	app := kernel.App()

	app.Get("/whoami", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			// With the CDN trusted, this is the end user's address rather than
			// the edge node that forwarded the request.
			"client_ip":       c.IP(),
			"forwarded_chain": c.IPs(),
			"note":            "a forged header from an untrusted hop is ignored",
		})
	})

	trusted := kernel.App().Config().TrustProxyConfig.Proxies
	logger.Info("CDN edge ranges trusted", "count", len(trusted))
}
