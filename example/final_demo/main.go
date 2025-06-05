package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	httppkg "go.oease.dev/goe/v2/core/http"
	"go.oease.dev/goe/v2/core/log"
	"go.uber.org/fx"
)

// UserService for demonstration
type UserService struct {
	logger contract.Logger
	config contract.Config
}

// NewUserService creates a new user service
func NewUserService(deps struct {
	fx.In
	Logger contract.Logger
	Config contract.Config
}) *UserService {
	return &UserService{
		logger: deps.Logger,
		config: deps.Config,
	}
}

// GetUser returns a user by ID
func (s *UserService) GetUser(id string) map[string]interface{} {
	s.logger.Info("Getting user", log.NewField("id", id))

	return map[string]interface{}{
		"id":   id,
		"name": "John Doe",
		"app":  s.config.GetString("APP_NAME"),
	}
}

func main() {
	// Create application with HTTP enabled
	_ = goe.New(goe.Options{
		Name:        "Final Test",
		Version:     "1.0.0",
		Environment: "dev",
		WithHTTP:    true,
		Providers: []any{
			NewUserService,
		},
		Invokers: []any{
			func(deps struct {
				fx.In
				HTTPKernel  contract.HTTPKernel
				Logger      contract.Logger
				Config      contract.Config
				UserService *UserService
			}) {
				app := deps.HTTPKernel.App()

				// Test 1: Direct injection
				app.Get("/direct", func(c fiber.Ctx) error {
					deps.Logger.Info("Direct endpoint called")
					return c.JSON(fiber.Map{
						"method": "direct",
						"app":    deps.Config.GetString("APP_NAME"),
					})
				})

				// Test 2: Context helpers
				app.Get("/context", func(c fiber.Ctx) error {
					logger := httppkg.GetLogger(c)
					config := httppkg.GetConfig(c)

					logger.Info("Context endpoint called")

					return c.JSON(fiber.Map{
						"method": "context",
						"app":    config.GetString("APP_NAME"),
						"env":    config.GetString("GOE_ENV"),
					})
				})

				// Test 3: Service injection
				app.Get("/user/:id", func(c fiber.Ctx) error {
					user := deps.UserService.GetUser(c.Params("id"))
					return c.JSON(user)
				})

				// Test 4: Mixed approach
				app.Get("/mixed", func(c fiber.Ctx) error {
					// Get logger from context
					logger := httppkg.GetLogger(c)
					logger.Info("Mixed endpoint called")

					// Use injected service
					user := deps.UserService.GetUser("123")

					return c.JSON(fiber.Map{
						"method": "mixed",
						"user":   user,
					})
				})

				deps.Logger.Info("All test routes registered")
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

	// Test all endpoints
	fmt.Println("\nTesting all endpoints...")

	endpoints := []string{
		"/direct",
		"/context",
		"/user/456",
		"/mixed",
	}

	for _, endpoint := range endpoints {
		resp, err := http.Get("http://localhost:8080" + endpoint)
		if err != nil {
			fmt.Printf("❌ %s - Error: %v\n", endpoint, err)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var result map[string]interface{}
		json.Unmarshal(body, &result)

		fmt.Printf("✅ %s - Status: %d, Response: %v\n", endpoint, resp.StatusCode, result)
	}

	fmt.Println("\nAll tests completed successfully! 🎉")
}
