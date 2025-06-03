package http_test

import (
	"context"
	"io"
	stdhttp "net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v3"
	corehttp "go.oease.dev/goe/v2/core/http"
)

func TestHttp(t *testing.T) {
	// Create a new http module
	h := corehttp.New()

	// Test module name
	t.Run("Module Name", func(t *testing.T) {
		name := h.Name()
		if name != "http" {
			t.Errorf("Expected module name to be 'http', got '%s'", name)
		}
	})

	// Test module lifecycle
	t.Run("Module Lifecycle", func(t *testing.T) {
		// Initialize
		err := h.Initialize(context.Background())
		if err != nil {
			t.Errorf("Failed to initialize http: %v", err)
		}

		// Start
		err = h.Start(context.Background())
		if err != nil {
			t.Errorf("Failed to start http: %v", err)
		}

		// Stop
		err = h.Stop(context.Background())
		if err != nil {
			t.Errorf("Failed to stop http: %v", err)
		}
	})

	// Test Fiber app
	t.Run("Fiber App", func(t *testing.T) {
		app := h.Fiber()
		if app == nil {
			t.Errorf("Expected Fiber app to not be nil")
		}
	})

	// Test route registration
	t.Run("Route Registration", func(t *testing.T) {
		// Create a new http module for this test
		h := corehttp.New()

		// Register a GET route
		h.Get("/test", func(c fiber.Ctx) error {
			return c.SendString("GET test")
		})

		// Register a POST route
		h.Post("/test", func(c fiber.Ctx) error {
			return c.SendString("POST test")
		})

		// Register a PUT route
		h.Put("/test", func(c fiber.Ctx) error {
			return c.SendString("PUT test")
		})

		// Register a DELETE route
		h.Delete("/test", func(c fiber.Ctx) error {
			return c.SendString("DELETE test")
		})

		// Register a PATCH route
		h.Patch("/test", func(c fiber.Ctx) error {
			return c.SendString("PATCH test")
		})

		// Register an OPTIONS route
		h.Options("/test", func(c fiber.Ctx) error {
			return c.SendString("OPTIONS test")
		})

		// Register a HEAD route
		h.Head("/test", func(c fiber.Ctx) error {
			return c.SendString("HEAD test")
		})

		// Register an ALL route
		h.All("/all", func(c fiber.Ctx) error {
			return c.SendString("ALL test")
		})

		// Register a CONNECT route
		h.Connect("/test", func(c fiber.Ctx) error {
			return c.SendString("CONNECT test")
		})

		// Register a TRACE route
		h.Trace("/test", func(c fiber.Ctx) error {
			return c.SendString("TRACE test")
		})

		// Register a route with multiple methods
		h.Add([]string{"GET", "POST"}, "/multi", func(c fiber.Ctx) error {
			return c.SendString("MULTI test")
		})

		// Test the routes with httptest
		app := h.Fiber()

		// Test GET route
		req := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Errorf("Failed to test GET route: %v", err)
		}
		if resp.StatusCode != stdhttp.StatusOK {
			t.Errorf("Expected status code %d, got %d", stdhttp.StatusOK, resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "GET test" {
			t.Errorf("Expected body 'GET test', got '%s'", string(body))
		}

		// Test POST route
		req = httptest.NewRequest(stdhttp.MethodPost, "/test", nil)
		resp, err = app.Test(req)
		if err != nil {
			t.Errorf("Failed to test POST route: %v", err)
		}
		if resp.StatusCode != stdhttp.StatusOK {
			t.Errorf("Expected status code %d, got %d", stdhttp.StatusOK, resp.StatusCode)
		}
		body, _ = io.ReadAll(resp.Body)
		if string(body) != "POST test" {
			t.Errorf("Expected body 'POST test', got '%s'", string(body))
		}

		// Test multi-method route
		req = httptest.NewRequest(stdhttp.MethodGet, "/multi", nil)
		resp, err = app.Test(req)
		if err != nil {
			t.Errorf("Failed to test multi-method route: %v", err)
		}
		if resp.StatusCode != stdhttp.StatusOK {
			t.Errorf("Expected status code %d, got %d", stdhttp.StatusOK, resp.StatusCode)
		}
		body, _ = io.ReadAll(resp.Body)
		if string(body) != "MULTI test" {
			t.Errorf("Expected body 'MULTI test', got '%s'", string(body))
		}
	})

	// Test route groups
	t.Run("Route Groups", func(t *testing.T) {
		// Create a new http module for this test
		h := corehttp.New()

		// Create a route group
		group := h.Group("/api")

		// Register a GET route in the group
		group.Get("/test", func(c fiber.Ctx) error {
			return c.SendString("API GET test")
		})

		// Register a POST route in the group
		group.Post("/test", func(c fiber.Ctx) error {
			return c.SendString("API POST test")
		})

		// Create a nested group
		nestedGroup := group.Group("/v1")

		// Register a GET route in the nested group
		nestedGroup.Get("/test", func(c fiber.Ctx) error {
			return c.SendString("API V1 GET test")
		})

		// Test the routes with httptest
		app := h.Fiber()

		// Test group GET route
		req := httptest.NewRequest(stdhttp.MethodGet, "/api/test", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Errorf("Failed to test group GET route: %v", err)
		}
		if resp.StatusCode != stdhttp.StatusOK {
			t.Errorf("Expected status code %d, got %d", stdhttp.StatusOK, resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "API GET test" {
			t.Errorf("Expected body 'API GET test', got '%s'", string(body))
		}

		// Test nested group GET route
		req = httptest.NewRequest(stdhttp.MethodGet, "/api/v1/test", nil)
		resp, err = app.Test(req)
		if err != nil {
			t.Errorf("Failed to test nested group GET route: %v", err)
		}
		if resp.StatusCode != stdhttp.StatusOK {
			t.Errorf("Expected status code %d, got %d", stdhttp.StatusOK, resp.StatusCode)
		}
		body, _ = io.ReadAll(resp.Body)
		if string(body) != "API V1 GET test" {
			t.Errorf("Expected body 'API V1 GET test', got '%s'", string(body))
		}
	})

	// Test middleware
	t.Run("Middleware", func(t *testing.T) {
		// Create a new http module for this test
		h := corehttp.New()

		// Register a middleware
		h.Use(func(c fiber.Ctx) error {
			c.Locals("test", "middleware")
			return c.Next()
		})

		// Register a route that uses the middleware
		h.Get("/middleware", func(c fiber.Ctx) error {
			return c.SendString(c.Locals("test").(string))
		})

		// Test the route with httptest
		app := h.Fiber()

		// Test middleware
		req := httptest.NewRequest(stdhttp.MethodGet, "/middleware", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Errorf("Failed to test middleware: %v", err)
		}
		if resp.StatusCode != stdhttp.StatusOK {
			t.Errorf("Expected status code %d, got %d", stdhttp.StatusOK, resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "middleware" {
			t.Errorf("Expected body 'middleware', got '%s'", string(body))
		}
	})

	// Test static file serving
	t.Run("Static Files", func(t *testing.T) {
		// Create a temporary directory
		tempDir, err := os.MkdirTemp("", "http_test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create a test file
		testFile := "test.txt"
		testContent := "This is a test file"
		err = os.WriteFile(tempDir+"/"+testFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		// Create a new http module for this test
		h := corehttp.New()

		// Serve static files
		h.Static("/static", tempDir)

		// Test the static file with httptest
		app := h.Fiber()

		// Test static file
		req := httptest.NewRequest(stdhttp.MethodGet, "/static/"+testFile, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Errorf("Failed to test static file: %v", err)
		}
		if resp.StatusCode != stdhttp.StatusOK {
			t.Errorf("Expected status code %d, got %d", stdhttp.StatusOK, resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if string(body) != testContent {
			t.Errorf("Expected body '%s', got '%s'", testContent, string(body))
		}
	})
}

// BenchmarkHttp benchmarks the http operations
func BenchmarkHttp(b *testing.B) {
	// Create a new http module
	h := corehttp.New()

	// Register a simple route
	h.Get("/bench", func(c fiber.Ctx) error {
		return c.SendString("Benchmark")
	})

	// Get the Fiber app
	app := h.Fiber()

	// Benchmark route registration
	b.Run("RouteRegistration", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			path := "/bench" + string(rune(i%26+97)) // a-z
			h.Get(path, func(c fiber.Ctx) error {
				return c.SendString("Benchmark")
			})
		}
	})

	// Benchmark route handling
	b.Run("RouteHandling", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest(stdhttp.MethodGet, "/bench", nil)
			_, _ = app.Test(req)
		}
	})

	// Benchmark middleware
	b.Run("Middleware", func(b *testing.B) {
		// Create a new http module for this benchmark
		h := corehttp.New()

		// Register a middleware
		h.Use(func(c fiber.Ctx) error {
			c.Locals("test", "middleware")
			return c.Next()
		})

		// Register a route that uses the middleware
		h.Get("/middleware", func(c fiber.Ctx) error {
			return c.SendString(c.Locals("test").(string))
		})

		// Get the Fiber app
		app := h.Fiber()

		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest(stdhttp.MethodGet, "/middleware", nil)
			_, _ = app.Test(req)
		}
	})

	// Benchmark group creation
	b.Run("GroupCreation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			path := "/group" + string(rune(i%26+97)) // a-z
			_ = h.Group(path)
		}
	})
}
