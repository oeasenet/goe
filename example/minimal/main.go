package main

import (
	"fmt"
	"time"

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
			func(http contract.HTTPKernel) {
				// Register a simple route directly
				app := http.App()

				app.Get("/", func(c fiber.Ctx) error {
					return c.JSON(fiber.Map{
						"message": "Hello from Goe!",
						"time":    time.Now(),
					})
				})

				app.Get("/test", func(c fiber.Ctx) error {
					logger := goe.Log()
					logger.Info("Test endpoint called", log.NewField("path", c.Path()))

					return c.JSON(fiber.Map{
						"status": "ok",
						"app":    goe.App().Name(),
					})
				})
			},
		},
	})

	fmt.Println("Starting minimal test server...")
	goe.Log().Info("Server starting on http://localhost:8080")

	// Run the application
	goe.Run()
}
