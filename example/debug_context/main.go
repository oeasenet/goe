package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	httppkg "go.oease.dev/goe/v2/core/http"
)

func main() {
	// Create application with HTTP enabled
	_ = goe.New(goe.Options{
		Name:        "Debug Context",
		Version:     "1.0.0",
		Environment: "dev",
		WithHTTP:    true,
		Invokers: []any{
			func(httpKernel contract.HTTPKernel, logger contract.Logger, config contract.Config) {
				app := httpKernel.App()

				// Debug endpoint to check context
				app.Get("/debug", func(c fiber.Ctx) error {
					// Check if services are in context
					services := c.Locals(string(httppkg.ServicesKey))

					if services == nil {
						return c.JSON(fiber.Map{
							"error":  "Services not found in context",
							"locals": fmt.Sprintf("%v", c.Locals("")),
						})
					}

					// Try to get services
					svc, ok := services.(httppkg.Services)
					if !ok {
						return c.JSON(fiber.Map{
							"error": "Services wrong type",
							"type":  fmt.Sprintf("%T", services),
						})
					}

					return c.JSON(fiber.Map{
						"success":   true,
						"hasConfig": svc.Config != nil,
						"hasLogger": svc.Logger != nil,
						"hasApp":    svc.App != nil,
					})
				})

				logger.Info("Debug routes registered")
			},
		},
	})

	// Start server in background
	go func() {
		goe.Log().Info("Debug server starting on http://localhost:8080")
		goe.Run()
	}()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Test the endpoint
	fmt.Println("\nTesting debug endpoint...")

	resp, err := http.Get("http://localhost:8080/debug")
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response: %s\n", body)
}
