// Package main demonstrates configuring the HTTP server from Go code instead
// of environment variables.
//
// This example shows:
// - Configuring Fiber through goe.Options.HTTP rather than FIBER_*/HTTP_* vars
// - How code, environment and GOE defaults layer together
// - Reaching settings that have no environment equivalent, such as TLS
// - The escape hatch for anything GOE does not wrap
//
// Run:
//
//	go run main.go
//
// Test:
//
//	curl -i http://localhost:8080/
//	curl -i http://localhost:8080/config
package main

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
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

		Invokers: []any{registerRoutes},
	})

	goe.Run()
}

func registerRoutes(kernel contract.HTTPKernel, logger contract.Logger) {
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

	logger.Info("routes registered")
}
