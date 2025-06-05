package main

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/log"
)

func main() {
	// Create application with HTTP enabled
	_ = goe.New(goe.Options{
		WithHTTP: true,
		Invokers: []any{
			func(http contract.HTTPKernel, logger contract.Logger, config contract.Config) {
				app := http.App()

				// Simple route without context helpers
				app.Get("/simple", func(c fiber.Ctx) error {
					logger.Info("Simple endpoint called")
					return c.JSON(fiber.Map{
						"status":  "ok",
						"message": "Simple test working",
					})
				})

				// Route using injected services directly
				app.Get("/info", func(c fiber.Ctx) error {
					logger.Info("Info endpoint called",
						log.NewField("app_name", config.GetString("APP_NAME")),
					)

					return c.JSON(fiber.Map{
						"app": fiber.Map{
							"name":  config.GetString("APP_NAME"),
							"env":   config.GetString("GOE_ENV"),
							"debug": config.GetBool("DEBUG"),
						},
					})
				})
			},
		},
	})

	goe.Log().Info("Simple test server starting on http://localhost:8080")
	goe.Run()
}
