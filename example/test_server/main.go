package main

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/http"
	"go.oease.dev/goe/v2/core/log"
)

func main() {
	// Create application with HTTP enabled
	_ = goe.New(goe.Options{
		Name:        "Test Server",
		Version:     "1.0.0",
		Environment: "dev",
		WithHTTP:    true,
		Invokers: []any{
			func(httpKernel contract.HTTPKernel, logger contract.Logger, config contract.Config) {
				app := httpKernel.App()

				// Simple health check
				app.Get("/health", func(c fiber.Ctx) error {
					return c.JSON(fiber.Map{
						"status": "healthy",
						"app":    config.GetString("APP_NAME"),
					})
				})

				// Test context helpers
				app.Get("/context-test", func(c fiber.Ctx) error {
					// Get logger from context
					ctxLogger := http.GetLogger(c)
					ctxLogger.Info("Context test endpoint called")

					// Get config from context
					ctxConfig := http.GetConfig(c)

					return c.JSON(fiber.Map{
						"app_name": ctxConfig.GetString("APP_NAME"),
						"env":      ctxConfig.GetString("GOE_ENV"),
						"debug":    ctxConfig.GetBool("DEBUG"),
					})
				})

				// Test direct injection
				app.Get("/direct-test", func(c fiber.Ctx) error {
					logger.Info("Direct test endpoint called",
						log.NewField("method", c.Method()),
						log.NewField("path", c.Path()),
					)

					return c.JSON(fiber.Map{
						"message": "Direct injection works",
						"config": fiber.Map{
							"app_name": config.GetString("APP_NAME"),
							"port":     config.GetInt("HTTP_PORT"),
						},
					})
				})

				logger.Info("Test routes registered")
			},
		},
	})

	goe.Log().Info("Test server starting on http://localhost:8080")
	goe.Run()
}
