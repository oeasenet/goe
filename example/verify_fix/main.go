package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	httppkg "go.oease.dev/goe/v2/core/http"
	"go.oease.dev/goe/v2/core/log"
)

func main() {
	// Create application with HTTP enabled
	_ = goe.New(goe.Options{
		WithHTTP: true,
		Invokers: []any{
			func(httpKernel contract.HTTPKernel, logger contract.Logger, config contract.Config) {
				app := httpKernel.App()

				// Test 1: Simple endpoint
				app.Get("/test1", func(c fiber.Ctx) error {
					return c.JSON(fiber.Map{"test": 1, "status": "ok"})
				})

				// Test 2: Context helpers
				app.Get("/test2", func(c fiber.Ctx) error {
					// This should work now with the fix
					ctxLogger := httppkg.GetLogger(c)
					ctxConfig := httppkg.GetConfig(c)

					ctxLogger.Info("Test 2 called")

					return c.JSON(fiber.Map{
						"test":     2,
						"app_name": ctxConfig.GetString("APP_NAME"),
						"env":      ctxConfig.GetString("GOE_ENV"),
					})
				})

				// Test 3: Direct injection
				app.Get("/test3", func(c fiber.Ctx) error {
					logger.Info("Test 3 called", log.NewField("path", c.Path()))
					return c.JSON(fiber.Map{
						"test":     3,
						"app_name": config.GetString("APP_NAME"),
					})
				})

				logger.Info("Fix verification routes registered")
			},
		},
	})

	// Start server in background
	go func() {
		goe.Log().Info("Server starting on http://localhost:8080")
		goe.Run()
	}()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Test the endpoints
	fmt.Println("\nTesting endpoints...")

	// Test 1
	resp1, err := http.Get("http://localhost:8080/test1")
	if err != nil {
		fmt.Printf("Test 1 failed: %v\n", err)
	} else {
		fmt.Printf("Test 1: %d\n", resp1.StatusCode)
		resp1.Body.Close()
	}

	// Test 2 - This should work now
	resp2, err := http.Get("http://localhost:8080/test2")
	if err != nil {
		fmt.Printf("Test 2 failed: %v\n", err)
	} else {
		fmt.Printf("Test 2: %d\n", resp2.StatusCode)
		resp2.Body.Close()
	}

	// Test 3
	resp3, err := http.Get("http://localhost:8080/test3")
	if err != nil {
		fmt.Printf("Test 3 failed: %v\n", err)
	} else {
		fmt.Printf("Test 3: %d\n", resp3.StatusCode)
		resp3.Body.Close()
	}

	fmt.Println("\nAll tests completed!")
}
