package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"go.oease.dev/goe/v2/middlewares"
)

func TestNewSPAServingMiddleware(t *testing.T) {
	// Test creating a new SPA middleware
	middleware := middlewares.NewSPAServingMiddleware("./dist", "index.html")
	assert.NotNil(t, middleware)
}

func TestSPAMiddlewareHandleIndexFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test index.html file
	indexContent := `<!DOCTYPE html>
<html>
<head>
    <title>Test SPA</title>
</head>
<body>
    <div id="app">Test SPA Application</div>
</body>
</html>`

	indexPath := filepath.Join(tempDir, "index.html")
	err := os.WriteFile(indexPath, []byte(indexContent), 0644)
	assert.NoError(t, err)

	// Create SPA middleware
	spaMiddleware := middlewares.NewSPAServingMiddleware(tempDir, "index.html")

	// Create a new Fiber app
	app := fiber.New()

	// Add the index file handler for catch-all routes
	app.Get("/*", spaMiddleware.HandleIndexFile())

	tests := []struct {
		name string
		path string
	}{
		{"root path", "/"},
		{"nested route", "/users"},
		{"deep nested route", "/users/123/profile"},
		{"route with query params", "/search?q=test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))

			resp.Body.Close()
		})
	}
}

func TestSPAMiddlewareHandleStaticFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create test static files
	cssContent := "body { margin: 0; padding: 0; }"
	cssPath := filepath.Join(tempDir, "style.css")
	err := os.WriteFile(cssPath, []byte(cssContent), 0644)
	assert.NoError(t, err)

	jsContent := "console.log('Hello, World!');"
	jsPath := filepath.Join(tempDir, "app.js")
	err = os.WriteFile(jsPath, []byte(jsContent), 0644)
	assert.NoError(t, err)

	// Create SPA middleware
	spaMiddleware := middlewares.NewSPAServingMiddleware(tempDir, "index.html")

	// Create a new Fiber app
	app := fiber.New()

	// Add the static file handler
	app.Use("/", spaMiddleware.HandleStaticFiles())

	tests := []struct {
		name                string
		path                string
		expectedContent     string
		expectedContentType string
	}{
		{
			name:                "CSS file",
			path:                "/style.css",
			expectedContent:     cssContent,
			expectedContentType: "text/css; charset=utf-8",
		},
		{
			name:                "JS file",
			path:                "/app.js",
			expectedContent:     jsContent,
			expectedContentType: "application/javascript; charset=utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, tt.expectedContentType, resp.Header.Get("Content-Type"))

			resp.Body.Close()
		})
	}
}

func TestSPAMiddlewareIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create test files
	indexContent := `<!DOCTYPE html>
<html>
<head>
    <title>Test SPA</title>
    <link rel="stylesheet" href="/style.css">
</head>
<body>
    <div id="app">Test SPA Application</div>
    <script src="/app.js"></script>
</body>
</html>`

	cssContent := "body { margin: 0; padding: 0; background: #f0f0f0; }"
	jsContent := "document.addEventListener('DOMContentLoaded', function() { console.log('SPA loaded'); });"

	// Write files
	indexPath := filepath.Join(tempDir, "index.html")
	err := os.WriteFile(indexPath, []byte(indexContent), 0644)
	assert.NoError(t, err)

	cssPath := filepath.Join(tempDir, "style.css")
	err = os.WriteFile(cssPath, []byte(cssContent), 0644)
	assert.NoError(t, err)

	jsPath := filepath.Join(tempDir, "app.js")
	err = os.WriteFile(jsPath, []byte(jsContent), 0644)
	assert.NoError(t, err)

	// Create SPA middleware
	spaMiddleware := middlewares.NewSPAServingMiddleware(tempDir, "index.html")

	// Create a new Fiber app with proper SPA setup
	app := fiber.New()

	// First handle static files
	app.Use("/", spaMiddleware.HandleStaticFiles())

	// Then handle SPA routes (catch-all)
	app.Get("/*", spaMiddleware.HandleIndexFile())

	tests := []struct {
		name                string
		path                string
		expectedStatus      int
		expectedContentType string
		shouldContain       string
	}{
		{
			name:                "static CSS file",
			path:                "/style.css",
			expectedStatus:      http.StatusOK,
			expectedContentType: "text/css; charset=utf-8",
			shouldContain:       "background: #f0f0f0",
		},
		{
			name:                "static JS file",
			path:                "/app.js",
			expectedStatus:      http.StatusOK,
			expectedContentType: "application/javascript; charset=utf-8",
			shouldContain:       "SPA loaded",
		},
		{
			name:                "SPA route - root",
			path:                "/",
			expectedStatus:      http.StatusOK,
			expectedContentType: "text/html; charset=utf-8",
			shouldContain:       "Test SPA Application",
		},
		{
			name:                "SPA route - nested",
			path:                "/users/123",
			expectedStatus:      http.StatusOK,
			expectedContentType: "text/html; charset=utf-8",
			shouldContain:       "Test SPA Application",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, tt.expectedContentType, resp.Header.Get("Content-Type"))

			resp.Body.Close()
		})
	}
}

func TestSPAMiddlewareNonExistentIndexFile(t *testing.T) {
	// Create a temporary directory without index file
	tempDir := t.TempDir()

	// Create SPA middleware pointing to non-existent index file
	spaMiddleware := middlewares.NewSPAServingMiddleware(tempDir, "nonexistent.html")

	// Create a new Fiber app
	app := fiber.New()

	// Add the index file handler
	app.Get("/*", spaMiddleware.HandleIndexFile())

	// Test request should return 404 or error
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	// Should return 404 or 500 when index file doesn't exist
	assert.True(t, resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusInternalServerError)

	resp.Body.Close()
}

func TestSPAMiddlewareNonExistentStaticFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create SPA middleware
	spaMiddleware := middlewares.NewSPAServingMiddleware(tempDir, "index.html")

	// Create a new Fiber app
	app := fiber.New()

	// Add the static file handler
	app.Use("/", spaMiddleware.HandleStaticFiles())

	// Test request for non-existent static file
	req := httptest.NewRequest(http.MethodGet, "/nonexistent.css", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	// Should return 404 when static file doesn't exist
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	resp.Body.Close()
}
