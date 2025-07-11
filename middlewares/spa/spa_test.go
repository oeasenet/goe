package spa

import (
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// Test helper to create temporary SPA directory structure
func createTestSPAStructure(t interface {
	Helper()
	Errorf(format string, args ...interface{})
	FailNow()
}) string {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "spa-test-*")
	require.NoError(t, err)

	// Create directory structure
	dirs := []string{
		"assets/css",
		"assets/js",
		"assets/images",
		"api",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755)
		require.NoError(t, err)
	}

	// Create test files
	files := map[string]string{
		"index.html":          `<!DOCTYPE html><html><body>SPA Index</body></html>`,
		"404.html":            `<!DOCTYPE html><html><body>Not Found</body></html>`,
		"assets/css/main.css": `body { margin: 0; }`,
		"assets/js/app.js":    `console.log("app");`,
		"robots.txt":          `User-agent: *`,
	}

	for file, content := range files {
		filePath := filepath.Join(tmpDir, file)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)
	}

	return tmpDir
}

func Test_New_PanicOnEmptyRoot(t *testing.T) {
	t.Parallel()

	require.Panics(t, func() {
		New(Config{})
	})

	require.Panics(t, func() {
		New()
	})
}

func Test_SPA_Default(t *testing.T) {
	tmpDir := createTestSPAStructure(t)
	defer os.RemoveAll(tmpDir)

	app := fiber.New()
	app.Use(New(Config{
		Root: tmpDir,
	}))

	// Test serving index.html for root
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "SPA Index")
}

func Test_SPA_StaticFiles(t *testing.T) {
	tmpDir := createTestSPAStructure(t)
	defer os.RemoveAll(tmpDir)

	app := fiber.New()
	app.Use(New(Config{
		Root: tmpDir,
	}))

	tests := []struct {
		path         string
		expectedBody string
		statusCode   int
	}{
		{"/assets/css/main.css", "body { margin: 0; }", fiber.StatusOK},
		{"/assets/js/app.js", `console.log("app");`, fiber.StatusOK},
		{"/robots.txt", "User-agent: *", fiber.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, tt.path, nil))
			require.NoError(t, err)
			require.Equal(t, tt.statusCode, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, tt.expectedBody, string(body))
		})
	}
}

func Test_SPA_ClientRouting(t *testing.T) {
	tmpDir := createTestSPAStructure(t)
	defer os.RemoveAll(tmpDir)

	app := fiber.New()
	app.Use(New(Config{
		Root: tmpDir,
	}))

	// Test client-side routes (should all return index.html)
	routes := []string{
		"/about",
		"/users/123",
		"/dashboard",
		"/settings/profile",
		"/products/category/items",
	}

	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, route, nil))
			require.NoError(t, err)
			require.Equal(t, fiber.StatusOK, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Contains(t, string(body), "SPA Index")
		})
	}
}

func Test_SPA_NotFoundFile(t *testing.T) {
	tmpDir := createTestSPAStructure(t)
	defer os.RemoveAll(tmpDir)

	app := fiber.New()
	app.Use(New(Config{
		Root:         tmpDir,
		NotFoundFile: "404.html",
	}))

	// Non-existent route should return 404.html content
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/non-existent", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "Not Found")
}

func Test_SPA_CustomIndex(t *testing.T) {
	tmpDir := createTestSPAStructure(t)
	defer os.RemoveAll(tmpDir)

	// Create custom index file
	customIndex := `<!DOCTYPE html><html><body>Custom App</body></html>`
	err := os.WriteFile(filepath.Join(tmpDir, "app.html"), []byte(customIndex), 0644)
	require.NoError(t, err)

	app := fiber.New()
	app.Use(New(Config{
		Root:  tmpDir,
		Index: "app.html",
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "Custom App")
}

func Test_SPA_MaxAge(t *testing.T) {
	tmpDir := createTestSPAStructure(t)
	defer os.RemoveAll(tmpDir)

	app := fiber.New()
	app.Use(New(Config{
		Root:   tmpDir,
		MaxAge: 3600,
	}))

	// Test static file with cache headers
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/assets/css/main.css", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.Equal(t, "public, max-age=3600", resp.Header.Get(fiber.HeaderCacheControl))
}

func Test_SPA_Next(t *testing.T) {
	tmpDir := createTestSPAStructure(t)
	defer os.RemoveAll(tmpDir)

	app := fiber.New()

	// Skip SPA middleware for /api routes
	app.Use(New(Config{
		Root: tmpDir,
		Next: func(c fiber.Ctx) bool {
			return len(c.Path()) >= 4 && c.Path()[:4] == "/api"
		},
	}))

	// API handler
	app.Get("/api/users", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"users": []string{"alice", "bob"}})
	})

	// Test that /api routes skip SPA middleware
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/users", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "alice")
	require.NotContains(t, string(body), "SPA Index")

	// Test that non-API routes still get SPA treatment
	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/about", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "SPA Index")
}

func Test_SPA_MissingIndexFile(t *testing.T) {
	tmpDir := createTestSPAStructure(t)
	defer os.RemoveAll(tmpDir)

	// Remove index.html
	os.Remove(filepath.Join(tmpDir, "index.html"))

	app := fiber.New()
	app.Use(New(Config{
		Root: tmpDir,
	}))

	// Add a 404 handler
	app.Use(func(c fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).SendString("404 Not Found")
	})

	// Should pass through to 404 handler
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func Test_SPA_Browse(t *testing.T) {
	tmpDir := createTestSPAStructure(t)
	defer os.RemoveAll(tmpDir)

	app := fiber.New()
	app.Use(New(Config{
		Root:   tmpDir,
		Browse: true,
	}))

	// Directory browsing is simplified in this implementation
	// It falls back to index.html
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/assets/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func Test_MightBeFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/index.html", true},
		{"/assets/css/main.css", true},
		{"/robots.txt", true},
		{"/favicon.ico", true},
		{"/about", false},
		{"/users/123", false},
		{"/api/users", false},
		{"/assets/images/", true},
		{"/static/file.js", true},
		{"/dist/bundle.js", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := mightBeFile(tt.path)
			require.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests
func Benchmark_SPA_Index(b *testing.B) {
	tmpDir := createTestSPAStructure(b)
	defer os.RemoveAll(tmpDir)

	app := fiber.New()
	app.Use(New(Config{
		Root: tmpDir,
	}))

	req := httptest.NewRequest(fiber.MethodGet, "/about", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := app.Test(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_SPA_Static(b *testing.B) {
	tmpDir := createTestSPAStructure(b)
	defer os.RemoveAll(tmpDir)

	app := fiber.New()
	app.Use(New(Config{
		Root: tmpDir,
	}))

	req := httptest.NewRequest(fiber.MethodGet, "/assets/css/main.css", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := app.Test(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_SPA_WithNext(b *testing.B) {
	tmpDir := createTestSPAStructure(b)
	defer os.RemoveAll(tmpDir)

	app := fiber.New()
	app.Use(New(Config{
		Root: tmpDir,
		Next: func(c fiber.Ctx) bool {
			return c.Path()[:4] == "/api"
		},
	}))

	req := httptest.NewRequest(fiber.MethodGet, "/api/users", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := app.Test(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
