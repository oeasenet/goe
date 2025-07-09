package middlewares

import (
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSPAServingMiddleware(t *testing.T) {
	rootPath := "./test-spa"
	indexFile := "index.html"

	middleware := NewSPAServingMiddleware(rootPath, indexFile)

	assert.NotNil(t, middleware)
	assert.Equal(t, rootPath, middleware.rootPath)
	assert.Equal(t, indexFile, middleware.indexFileFileName)
}

func TestSPAMiddleware_HandleStaticFiles(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "spa-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFiles := map[string]string{
		"app.js":        "console.log('Hello World');",
		"styles.css":    "body { margin: 0; }",
		"image.png":     "fake-png-data",
		"manifest.json": `{"name": "Test App"}`,
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)
	}

	middleware := NewSPAServingMiddleware(tmpDir, "index.html")
	app := fiber.New()

	app.Use("/", middleware.HandleStaticFiles())

	// Test each file
	for filename, expectedContent := range testFiles {
		t.Run("serve "+filename, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/"+filename, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, 200, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, expectedContent, string(body))
		})
	}
}

func TestSPAMiddleware_HandleStaticFiles_NotFound(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "spa-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	middleware := NewSPAServingMiddleware(tmpDir, "index.html")
	app := fiber.New()

	app.Use("/", middleware.HandleStaticFiles())

	req := httptest.NewRequest("GET", "/nonexistent.js", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, 404, resp.StatusCode)
}

func TestSPAMiddleware_HandleIndexFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "spa-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create index.html
	indexContent := `<!DOCTYPE html>
<html>
<head>
    <title>SPA Test</title>
</head>
<body>
    <div id="app">Loading...</div>
</body>
</html>`

	indexPath := filepath.Join(tmpDir, "index.html")
	err = os.WriteFile(indexPath, []byte(indexContent), 0644)
	require.NoError(t, err)

	middleware := NewSPAServingMiddleware(tmpDir, "index.html")
	app := fiber.New()

	app.Get("/*", middleware.HandleIndexFile())

	tests := []string{
		"/",
		"/about",
		"/users/123",
		"/products/category/electronics",
		"/deeply/nested/path/structure",
	}

	for _, path := range tests {
		t.Run("serve index for "+path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, 200, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, indexContent, string(body))
		})
	}
}

func TestSPAMiddleware_HandleIndexFile_NotFound(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "spa-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Don't create index.html to test 404 behavior
	middleware := NewSPAServingMiddleware(tmpDir, "index.html")
	app := fiber.New()

	app.Get("/*", middleware.HandleIndexFile())

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, 404, resp.StatusCode)
}

func TestSPAMiddleware_HandleIndexFile_CustomIndexFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "spa-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create custom index file
	customIndexContent := `<!DOCTYPE html>
<html>
<head>
    <title>Custom SPA</title>
</head>
<body>
    <div id="root">Custom App</div>
</body>
</html>`

	customIndexPath := filepath.Join(tmpDir, "app.html")
	err = os.WriteFile(customIndexPath, []byte(customIndexContent), 0644)
	require.NoError(t, err)

	middleware := NewSPAServingMiddleware(tmpDir, "app.html")
	app := fiber.New()

	app.Get("/*", middleware.HandleIndexFile())

	req := httptest.NewRequest("GET", "/custom-route", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, customIndexContent, string(body))
}

func TestSPAMiddleware_Integration_StaticFilesAndIndex(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "spa-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test files
	indexContent := `<!DOCTYPE html>
<html>
<head>
    <title>SPA Integration Test</title>
    <link rel="stylesheet" href="/styles.css">
</head>
<body>
    <div id="app">Loading...</div>
    <script src="/app.js"></script>
</body>
</html>`

	cssContent := `body { 
    margin: 0; 
    padding: 20px; 
    font-family: Arial, sans-serif; 
}`

	jsContent := `console.log('SPA App loaded');
document.getElementById('app').innerHTML = 'SPA App Ready!';`

	files := map[string]string{
		"index.html": indexContent,
		"styles.css": cssContent,
		"app.js":     jsContent,
	}

	for filename, content := range files {
		filePath := filepath.Join(tmpDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)
	}

	middleware := NewSPAServingMiddleware(tmpDir, "index.html")
	app := fiber.New()

	// Set up middleware for serving static files
	app.Use("/", middleware.HandleStaticFiles())

	// Set up catch-all route for SPA routing
	app.Get("/*", middleware.HandleIndexFile())

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "static CSS file",
			path:           "/styles.css",
			expectedStatus: 200,
			expectedBody:   cssContent,
		},
		{
			name:           "static JS file",
			path:           "/app.js",
			expectedStatus: 200,
			expectedBody:   jsContent,
		},
		{
			name:           "root path serves index",
			path:           "/",
			expectedStatus: 200,
			expectedBody:   indexContent,
		},
		{
			name:           "SPA route serves index",
			path:           "/about",
			expectedStatus: 200,
			expectedBody:   indexContent,
		},
		{
			name:           "nested SPA route serves index",
			path:           "/users/123/profile",
			expectedStatus: 200,
			expectedBody:   indexContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, string(body))
		})
	}
}

func TestSPAMiddleware_Integration_StaticFilesPriority(t *testing.T) {
	// Test that static files take priority over index file serving
	tmpDir, err := os.MkdirTemp("", "spa-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	indexContent := `<!DOCTYPE html><html><body>Index</body></html>`
	realFileContent := `/* This is a real CSS file */`

	// Create both index.html and styles.css
	err = os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte(indexContent), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "styles.css"), []byte(realFileContent), 0644)
	require.NoError(t, err)

	middleware := NewSPAServingMiddleware(tmpDir, "index.html")
	app := fiber.New()

	// Static files middleware first
	app.Use("/", middleware.HandleStaticFiles())

	// Catch-all for SPA routing
	app.Get("/*", middleware.HandleIndexFile())

	// Request for existing static file should serve the static file
	req := httptest.NewRequest("GET", "/styles.css", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, realFileContent, string(body))
	assert.NotEqual(t, indexContent, string(body))
}

func TestSPAMiddleware_MultipleInstances(t *testing.T) {
	// Test that multiple instances work independently
	tmpDir1, err := os.MkdirTemp("", "spa-test-1-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir1)

	tmpDir2, err := os.MkdirTemp("", "spa-test-2-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir2)

	// Create different index files
	index1Content := `<!DOCTYPE html><html><body>App 1</body></html>`
	index2Content := `<!DOCTYPE html><html><body>App 2</body></html>`

	err = os.WriteFile(filepath.Join(tmpDir1, "index.html"), []byte(index1Content), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir2, "app.html"), []byte(index2Content), 0644)
	require.NoError(t, err)

	middleware1 := NewSPAServingMiddleware(tmpDir1, "index.html")
	middleware2 := NewSPAServingMiddleware(tmpDir2, "app.html")

	// Test first middleware
	app1 := fiber.New()
	app1.Get("/*", middleware1.HandleIndexFile())

	req1 := httptest.NewRequest("GET", "/test", nil)
	resp1, err := app1.Test(req1)
	require.NoError(t, err)

	body1, err := io.ReadAll(resp1.Body)
	require.NoError(t, err)
	assert.Equal(t, index1Content, string(body1))

	// Test second middleware
	app2 := fiber.New()
	app2.Get("/*", middleware2.HandleIndexFile())

	req2 := httptest.NewRequest("GET", "/test", nil)
	resp2, err := app2.Test(req2)
	require.NoError(t, err)

	body2, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)
	assert.Equal(t, index2Content, string(body2))

	// Verify they are different
	assert.NotEqual(t, string(body1), string(body2))
}

func TestSPAMiddleware_HandleStaticFiles_SubDirectories(t *testing.T) {
	// Test serving files from subdirectories
	tmpDir, err := os.MkdirTemp("", "spa-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create subdirectories and files
	assetsDir := filepath.Join(tmpDir, "assets")
	err = os.MkdirAll(assetsDir, 0755)
	require.NoError(t, err)

	imagesDir := filepath.Join(assetsDir, "images")
	err = os.MkdirAll(imagesDir, 0755)
	require.NoError(t, err)

	// Create files in subdirectories
	testFiles := map[string]string{
		"assets/app.js":          "console.log('App JS');",
		"assets/styles.css":      "body { color: red; }",
		"assets/images/logo.png": "fake-png-data",
	}

	for relativePath, content := range testFiles {
		fullPath := filepath.Join(tmpDir, relativePath)
		err := os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}

	middleware := NewSPAServingMiddleware(tmpDir, "index.html")
	app := fiber.New()

	app.Use("/", middleware.HandleStaticFiles())

	// Test serving files from subdirectories
	for relativePath, expectedContent := range testFiles {
		t.Run("serve "+relativePath, func(t *testing.T) {
			// Convert to URL path (use forward slashes)
			urlPath := "/" + strings.ReplaceAll(relativePath, "\\", "/")

			req := httptest.NewRequest("GET", urlPath, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, 200, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, expectedContent, string(body))
		})
	}
}
