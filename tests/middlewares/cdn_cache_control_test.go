package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"go.oease.dev/goe/v2/middlewares"
)

func TestCDNCacheControlDefaultRules(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expectedHeader string
	}{
		{
			name:           "index.html should not be cached",
			path:           "/index.html",
			expectedHeader: "no-cache, max-age=0, s-maxage=0, must-revalidate",
		},
		{
			name:           "nested index.html should not be cached",
			path:           "/app/index.html",
			expectedHeader: "no-cache, max-age=0, s-maxage=0, must-revalidate",
		},
		{
			name:           "manifest.webmanifest should not be cached",
			path:           "/manifest.webmanifest",
			expectedHeader: "no-cache, max-age=0, s-maxage=0, must-revalidate",
		},
		{
			name:           "service worker should not be cached",
			path:           "/sw.js",
			expectedHeader: "no-cache, max-age=0, s-maxage=0, must-revalidate",
		},
		{
			name:           "register service worker should not be cached",
			path:           "/registerSW.js",
			expectedHeader: "no-cache, max-age=0, s-maxage=0, must-revalidate",
		},
		{
			name:           "API endpoints should be private",
			path:           "/api/users",
			expectedHeader: "private, no-store",
		},
		{
			name:           "well-known endpoints should be private",
			path:           "/.well-known/security.txt",
			expectedHeader: "private, no-store",
		},
		{
			name:           "CSS files should be cached",
			path:           "/assets/style.css",
			expectedHeader: "public, max-age=3600",
		},
		{
			name:           "JS files should be cached",
			path:           "/assets/app.js",
			expectedHeader: "public, max-age=3600",
		},
		{
			name:           "image files should be cached",
			path:           "/images/logo.png",
			expectedHeader: "public, max-age=3600",
		},
		{
			name:           "font files should be cached",
			path:           "/fonts/roboto.woff2",
			expectedHeader: "public, max-age=3600",
		},
		{
			name:           "unknown paths should use default",
			path:           "/unknown/path",
			expectedHeader: "no-cache, max-age=0, s-maxage=0, must-revalidate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Fiber app
			app := fiber.New()

			// Apply the CDN cache control middleware with default config
			app.Use(middlewares.NewCDNCacheControlMiddleware().Handler())

			// Add a test route
			app.Get("/*", func(c fiber.Ctx) error {
				return c.SendString("test response")
			})

			// Create a test request
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			resp, err := app.Test(req)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, tt.expectedHeader, resp.Header.Get("Cache-Control"))

			resp.Body.Close()
		})
	}
}

func TestCDNCacheControlCustomConfig(t *testing.T) {
	// Create a custom configuration
	config := middlewares.DefaultCDNCacheControlConfig()
	config.DefaultCacheControl = "public, max-age=60"

	// Create a new Fiber app
	app := fiber.New()

	// Apply the CDN cache control middleware with custom config
	app.Use(middlewares.NewCDNCacheControlMiddlewareWithConfig(config).Handler())

	// Add a test route
	app.Get("/custom", func(c fiber.Ctx) error {
		return c.SendString("custom response")
	})

	// Create a test request for a path that doesn't match any rule
	req := httptest.NewRequest(http.MethodGet, "/custom", nil)
	resp, err := app.Test(req)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "public, max-age=60", resp.Header.Get("Cache-Control"))

	resp.Body.Close()
}

func TestCDNCacheControlWithRule(t *testing.T) {
	// Create a new Fiber app
	app := fiber.New()

	// Apply the CDN cache control middleware with custom rule
	middleware := middlewares.NewCDNCacheControlMiddleware().
		WithRule(regexp.MustCompile(`(?i)^.*/custom/.*$`), "public, max-age=120", "").
		WithDefaultCacheControl("public, max-age=30")

	app.Use(middleware.Handler())

	// Add test routes
	app.Get("/*", func(c fiber.Ctx) error {
		return c.SendString("test response")
	})

	tests := []struct {
		name           string
		path           string
		expectedHeader string
	}{
		{
			name:           "custom rule should match",
			path:           "/custom/file.txt",
			expectedHeader: "public, max-age=120",
		},
		{
			name:           "non-matching path should use default",
			path:           "/other/file.txt",
			expectedHeader: "public, max-age=30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, tt.expectedHeader, resp.Header.Get("Cache-Control"))

			resp.Body.Close()
		})
	}
}

func TestCDNCacheControlWithContentType(t *testing.T) {
	// Create a new Fiber app
	app := fiber.New()

	// Apply the CDN cache control middleware with default config
	app.Use(middlewares.NewCDNCacheControlMiddleware().Handler())

	// Add a test route
	app.Get("/*", func(c fiber.Ctx) error {
		return c.SendString("test response")
	})

	// Test manifest.webmanifest which should set both Cache-Control and Content-Type
	req := httptest.NewRequest(http.MethodGet, "/manifest.webmanifest", nil)
	resp, err := app.Test(req)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "no-cache, max-age=0, s-maxage=0, must-revalidate", resp.Header.Get("Cache-Control"))
	assert.Equal(t, "application/manifest+json", resp.Header.Get("Content-Type"))

	resp.Body.Close()
}

func TestCDNCacheControlDeprecatedHandler(t *testing.T) {
	// Create a new Fiber app
	app := fiber.New()

	// Apply the deprecated middleware handler
	app.Use(middlewares.NewCDNCacheControlMiddlewareHandler())

	// Add a test route
	app.Get("/test.css", func(c fiber.Ctx) error {
		return c.SendString("body { margin: 0; }")
	})

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/test.css", nil)
	resp, err := app.Test(req)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "public, max-age=3600", resp.Header.Get("Cache-Control"))

	resp.Body.Close()
}
