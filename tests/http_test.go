package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	corehttp "go.oease.dev/goe/v2/core/http"
	"go.oease.dev/goe/v2/core/log"
	"go.uber.org/fx"
)

func TestHTTPModule(t *testing.T) {
	// Create app with HTTP enabled
	app := goe.New(goe.Options{
		Name:        "Test HTTP App",
		Version:     "1.0.0",
		Environment: "test",
		WithHTTP:    true,
	})

	if app == nil {
		t.Fatal("Failed to create application with HTTP")
	}

	// Test HTTP accessor
	httpKernel := goe.HTTP()
	if httpKernel == nil {
		t.Fatal("HTTP kernel is nil")
	}

	// Test Fiber app access
	fiberApp := httpKernel.App()
	if fiberApp == nil {
		t.Fatal("Fiber app is nil")
	}
}

func TestHTTPRouting(t *testing.T) {
	_ = goe.New(goe.Options{
		Name:        "Test Routing App",
		Version:     "1.0.0",
		Environment: "test",
		WithHTTP:    true,
		Invokers: []any{
			func(http contract.HTTPKernel) {
				app := http.App()

				// Test basic route
				app.Get("/test", func(c fiber.Ctx) error {
					return c.JSON(fiber.Map{"message": "test"})
				})

				// Test route with params
				app.Get("/users/:id", func(c fiber.Ctx) error {
					return c.JSON(fiber.Map{"id": c.Params("id")})
				})

				// Test POST route
				app.Post("/users", func(c fiber.Ctx) error {
					var body map[string]any
					if err := c.Bind().JSON(&body); err != nil {
						return err
					}
					return c.Status(fiber.StatusCreated).JSON(body)
				})
			},
		},
	})

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test GET request
	t.Run("GET /test", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/test")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result["message"] != "test" {
			t.Errorf("Expected message 'test', got %v", result["message"])
		}
	})

	// Test route with params
	t.Run("GET /users/:id", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/users/123")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result["id"] != "123" {
			t.Errorf("Expected id '123', got %v", result["id"])
		}
	})

	// Test POST request
	t.Run("POST /users", func(t *testing.T) {
		body := map[string]any{"name": "John", "email": "john@example.com"}
		jsonBody, _ := json.Marshal(body)

		resp, err := http.Post("http://localhost:8080/users", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 201 {
			t.Errorf("Expected status 201, got %d", resp.StatusCode)
		}

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result["name"] != "John" {
			t.Errorf("Expected name 'John', got %v", result["name"])
		}
	})
}

func TestHTTPContextHelpers(t *testing.T) {
	_ = goe.New(goe.Options{
		Name:        "Test Context App",
		Version:     "1.0.0",
		Environment: "test",
		WithHTTP:    true,
		Invokers: []any{
			func(http contract.HTTPKernel, logger contract.Logger, config contract.Config) {
				app := http.App()

				// Test context helpers
				app.Get("/context-test", func(c fiber.Ctx) error {
					// Get logger from context
					ctxLogger := corehttp.GetLogger(c)
					if ctxLogger == nil {
						return c.Status(500).JSON(fiber.Map{"error": "logger is nil"})
					}

					// Get config from context
					ctxConfig := corehttp.GetConfig(c)
					if ctxConfig == nil {
						return c.Status(500).JSON(fiber.Map{"error": "config is nil"})
					}

					// Log something
					ctxLogger.Info("Context test successful")

					return c.JSON(fiber.Map{
						"app_name": ctxConfig.GetString("APP_NAME"),
						"success":  true,
					})
				})
			},
		},
	})

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://localhost:8080/context-test")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d, body: %s", resp.StatusCode, body)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["success"] != true {
		t.Errorf("Expected success true, got %v", result["success"])
	}
}

func TestHTTPMiddleware(t *testing.T) {
	requestCount := 0

	_ = goe.New(goe.Options{
		Name:        "Test Middleware App",
		Version:     "1.0.0",
		Environment: "test",
		WithHTTP:    true,
		Invokers: []any{
			func(http contract.HTTPKernel, logger contract.Logger) {
				app := http.App()

				// Custom middleware
				app.Use(func(c fiber.Ctx) error {
					requestCount++
					c.Locals("request_count", requestCount)
					logger.Info("Middleware executed", log.NewField("count", requestCount))
					return c.Next()
				})

				// Route that uses middleware data
				app.Get("/middleware-test", func(c fiber.Ctx) error {
					count := c.Locals("request_count")
					return c.JSON(fiber.Map{"request_count": count})
				})
			},
		},
	})

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Make multiple requests
	for i := 1; i <= 3; i++ {
		resp, err := http.Get("http://localhost:8080/middleware-test")
		if err != nil {
			t.Fatalf("Failed to make request %d: %v", i, err)
		}
		defer resp.Body.Close()

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response %d: %v", i, err)
		}

		// The count should match the request number
		if int(result["request_count"].(float64)) != i {
			t.Errorf("Request %d: Expected count %d, got %v", i, i, result["request_count"])
		}
	}
}

func TestHTTPErrorHandling(t *testing.T) {
	_ = goe.New(goe.Options{
		Name:        "Test Error App",
		Version:     "1.0.0",
		Environment: "test",
		WithHTTP:    true,
		Invokers: []any{
			func(http contract.HTTPKernel) {
				app := http.App()

				// Route that returns an error
				app.Get("/error", func(c fiber.Ctx) error {
					return fiber.NewError(fiber.StatusBadRequest, "Test error")
				})

				// Route that panics (should be recovered)
				app.Get("/panic", func(c fiber.Ctx) error {
					panic("test panic")
				})
			},
		},
	})

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test error response
	t.Run("Error handling", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/error")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 400 {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result["error"] != "Test error" {
			t.Errorf("Expected error 'Test error', got %v", result["error"])
		}
	})

	// Test panic recovery
	t.Run("Panic recovery", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/panic")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Should return 500 due to panic
		if resp.StatusCode != 500 {
			t.Errorf("Expected status 500, got %d", resp.StatusCode)
		}
	})
}

func TestHTTPDependencyInjection(t *testing.T) {
	// Test service
	type TestService struct {
		logger contract.Logger
		value  string
	}

	_ = goe.New(goe.Options{
		Name:        "Test DI App",
		Version:     "1.0.0",
		Environment: "test",
		WithHTTP:    true,
		Providers: []any{
			func(logger contract.Logger) *TestService {
				return &TestService{
					logger: logger,
					value:  "injected-value",
				}
			},
		},
		Invokers: []any{
			func(http contract.HTTPKernel, service *TestService) {
				app := http.App()

				// Route using injected service
				app.Get("/di-test", func(c fiber.Ctx) error {
					service.logger.Info("DI test endpoint called")
					return c.JSON(fiber.Map{
						"value":      service.value,
						"has_logger": service.logger != nil,
					})
				})
			},
			// Test Fx parameter objects
			func(params struct {
				fx.In
				HTTP    contract.HTTPKernel
				Logger  contract.Logger
				Config  contract.Config
				Service *TestService
			}) {
				app := params.HTTP.App()

				app.Get("/di-params-test", func(c fiber.Ctx) error {
					return c.JSON(fiber.Map{
						"has_logger":    params.Logger != nil,
						"has_config":    params.Config != nil,
						"has_service":   params.Service != nil,
						"service_value": params.Service.value,
					})
				})
			},
		},
	})

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test service injection
	t.Run("Service injection", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/di-test")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result["value"] != "injected-value" {
			t.Errorf("Expected value 'injected-value', got %v", result["value"])
		}

		if result["has_logger"] != true {
			t.Error("Logger was not injected")
		}
	})

	// Test Fx parameter objects
	t.Run("Fx parameter objects", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/di-params-test")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result["has_logger"] != true {
			t.Error("Logger was not provided")
		}
		if result["has_config"] != true {
			t.Error("Config was not provided")
		}
		if result["has_service"] != true {
			t.Error("Service was not provided")
		}
		if result["service_value"] != "injected-value" {
			t.Errorf("Expected service_value 'injected-value', got %v", result["service_value"])
		}
	})
}
