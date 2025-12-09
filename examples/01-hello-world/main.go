// Package main demonstrates the minimal GOE application with HTTP server.
//
// This example shows:
// - Basic GOE initialization
// - HTTP route registration with Fiber
// - Using the Logger via dependency injection
// - Accessing global config and logger
//
// Run:
//
//	go run main.go
//
// Test:
//
//	curl http://localhost:3000/
//	curl http://localhost:3000/hello/World
//	curl http://localhost:3000/health
package main

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/webresult"
)

func main() {
	// Create a new GOE application with HTTP support enabled
	_ = goe.New(goe.Options{
		WithHTTP: true, // Enable the HTTP server (Fiber)

		// Invokers are functions called after all dependencies are ready
		// They're perfect for registering routes, starting background tasks, etc.
		Invokers: []any{
			registerRoutes,
		},
	})

	// Run starts the application and blocks until shutdown
	// Handles graceful shutdown on SIGINT/SIGTERM
	goe.Run()
}

// registerRoutes sets up all HTTP routes for the application.
// Dependencies are automatically injected by the Fx container.
func registerRoutes(kernel contract.HTTPKernel, logger contract.Logger, config contract.Config) {
	// Get the underlying Fiber app instance
	app := kernel.App()

	// Log startup information using the injected logger
	logger.Infow("Registering routes",
		"port", config.GetInt("HTTP_PORT"),
		"env", config.GetString("GOE_ENV"),
	)

	// Basic route - returns a simple greeting
	app.Get("/", func(c fiber.Ctx) error {
		logger.Info("Root endpoint accessed")
		return c.SendString("Hello from GOE Framework!")
	})

	// Route with path parameter
	app.Get("/hello/:name", func(c fiber.Ctx) error {
		name := c.Params("name")
		logger.Infow("Hello endpoint accessed", "name", name)
		return c.SendString("Hello, " + name + "!")
	})

	// JSON response using webresult helper
	app.Get("/health", func(c fiber.Ctx) error {
		return webresult.SendSucceed(c, fiber.Map{
			"status":  "healthy",
			"version": config.GetString("APP_VERSION"),
		})
	})

	// POST endpoint with JSON body parsing
	app.Post("/echo", func(c fiber.Ctx) error {
		// Parse JSON body into a map
		var body map[string]any
		if err := c.Bind().JSON(&body); err != nil {
			return webresult.SendFailed(c, "Invalid JSON body")
		}
		return webresult.SendSucceed(c, body)
	})

	logger.Info("All routes registered successfully")
}
