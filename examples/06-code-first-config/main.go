// Package main demonstrates configuring GOE modules from Go code instead of
// environment variables.
//
// This example shows:
// - Configuring Fiber through goe.Options.HTTP rather than FIBER_*/HTTP_* vars
// - Configuring the cache through goe.Options.Cache rather than CACHE_* vars
// - How code, environment and GOE defaults layer together
// - Reaching settings that have no environment equivalent, such as TLS
// - The escape hatch for anything GOE does not wrap
// - Where credentials live: always the environment, never code
//
// Run:
//
//	go run main.go
//
// Test:
//
//	curl -i http://localhost:8080/
//	curl -i http://localhost:8080/config
//	curl -i http://localhost:8080/cache
package main

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	goecache "go.oease.dev/goe/v2/core/cache"
	goehttp "go.oease.dev/goe/v2/core/http"
)

func main() {
	_ = goe.New(goe.Options{
		// Passing HTTP options enables the HTTP module, so WithHTTP: true is
		// not needed. Anything set here wins over the matching environment
		// variable; anything left out still reads from the environment.
		HTTP: []goehttp.Option{
			// GOE-level settings. These have no fiber.Config field, so GOE
			// names them itself.
			goehttp.WithHost("0.0.0.0"),
			goehttp.WithPort(8080),
			goehttp.WithRequestIDHeader("X-Request-ID"),

			// Fiber settings. The option name is always Fiber's field name
			// with "With" in front, so Fiber's documentation is the reference.
			goehttp.WithServerHeader("goe-example"),
			goehttp.WithBodyLimit(16 << 20), // fiber.Config.BodyLimit
			goehttp.WithStrictRouting(true),
			goehttp.WithCaseSensitive(true),

			// Listener settings, which have no environment equivalent at all.
			goehttp.WithDisableStartupMessage(false),
			goehttp.WithEnablePrintRoutes(true),

			// Serving HTTPS is a matter of pointing at a key pair:
			//
			//	goehttp.WithCertFile("/etc/certs/server.crt"),
			//	goehttp.WithCertKeyFile("/etc/certs/server.key"),
			//	goehttp.WithTLSMinVersion(tls.VersionTLS13),

			// The escape hatch covers everything GOE does not wrap, including
			// fields a future Fiber release adds. It runs last, so it always
			// wins — including over fields GOE itself relies on. GOE logs a
			// warning if you disable one of those, but never overrides you.
			goehttp.WithFiberConfig(func(c *fiber.Config) {
				c.EnableSplittingOnParsers = true
			}),
		},

		// The same pattern configures the cache, job, and lock modules, and
		// passing options enables the module just like WithCache: true would.
		// The option name is always the environment key with the module
		// prefix dropped: CACHE_TTL is WithTTL, CACHE_REDIS_HOST is
		// WithRedisHost.
		Cache: []goecache.Option{
			// The memory driver keeps this example runnable without Redis. A
			// store named after a registered driver uses that driver, so
			// goecache.WithStore("redis") is all a Redis-backed cache needs.
			goecache.WithStore("memory"),
			goecache.WithPrefix("example06"),
			goecache.WithTTL(5 * time.Minute),

			// A production setup points at Redis the same way:
			//
			//	goecache.WithStore("redis"),
			//	goecache.WithRedisHost("redis.internal"),
			//	goecache.WithRedisPort(6380),
			//
			// Credentials are the deliberate exception: there is no
			// WithRedisPassword. CACHE_REDIS_USERNAME/CACHE_REDIS_PASSWORD
			// come from the environment and apply to the endpoint chosen
			// here — secrets never live in source.
		},

		// Job and lock follow the same rules (they need a running Redis, so
		// they stay disabled in this example):
		//
		//	Job: []goejob.Option{
		//	    goejob.WithRedisHosts("redis.internal:6379"), // JOB_REDIS_PASSWORD still applies
		//	    goejob.WithConcurrency(10),
		//	    goejob.WithDefaultQueue("critical"),
		//	},
		//	Lock: []goelock.Option{
		//	    // Topology in code; lock.WithRedisURL rejects user:pass in the
		//	    // URL — LOCK_REDIS_USERNAME/LOCK_REDIS_PASSWORD supply it.
		//	    goelock.WithRedisURL("redis-sentinel://mymaster@s1:26379,s2:26379/0"),
		//	    goelock.WithDefaultExpiry(10 * time.Second),
		//	},

		Invokers: []any{registerRoutes},
	})

	goe.Run()
}

func registerRoutes(kernel contract.HTTPKernel, cache contract.Cache, logger contract.Logger) {
	app := kernel.App()

	app.Get("/", func(c fiber.Ctx) error {
		// WithReqCtx returns a logger already carrying request_id and, when a
		// span is active, trace_id. It relies on PassLocalsToContext, which GOE
		// enables by default.
		goehttp.WithReqCtx(c).Info("root endpoint accessed")
		return c.SendString("Configured from Go code")
	})

	// Echo back the settings that came from code, so the layering is visible.
	app.Get("/config", func(c fiber.Ctx) error {
		cfg := c.App().Config()
		return c.JSON(fiber.Map{
			"server_header":  cfg.ServerHeader,  // from WithServerHeader
			"body_limit":     cfg.BodyLimit,     // from WithBodyLimit
			"strict_routing": cfg.StrictRouting, // from WithStrictRouting
			"case_sensitive": cfg.CaseSensitive, // from WithCaseSensitive
			"read_timeout":   cfg.ReadTimeout.String(),
			"note":           "read_timeout was not set in code, so it comes from HTTP_READ_TIMEOUT or the GOE default",
		})
	})

	// The cache configured through goe.Options.Cache, exercised end to end.
	app.Get("/cache", func(c fiber.Ctx) error {
		visits, err := cache.Increment("visits")
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{
			"visits": visits,
			"note":   "counter lives in the memory store configured in code (WithStore, WithPrefix, WithTTL)",
		})
	})

	logger.Info("routes registered")
}
