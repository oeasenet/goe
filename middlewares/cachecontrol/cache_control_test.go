package cachecontrol

import (
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

func Test_CacheControl_Default(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, "no-cache, max-age=0, s-maxage=0, must-revalidate", resp.Header.Get(fiber.HeaderCacheControl))
}

func Test_CacheControl_StaticAssets(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(New())

	tests := []struct {
		path       string
		expectedCC string
		expectedCT string
	}{
		{"/static/style.css", "public, max-age=3600", ""},
		{"/assets/app.js", "public, max-age=3600", ""},
		{"/images/logo.png", "public, max-age=3600", ""},
		{"/fonts/roboto.woff2", "public, max-age=3600", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, tt.path, nil))
			require.NoError(t, err)
			require.Equal(t, tt.expectedCC, resp.Header.Get(fiber.HeaderCacheControl))
			if tt.expectedCT != "" {
				require.Equal(t, tt.expectedCT, resp.Header.Get(fiber.HeaderContentType))
			}
		})
	}
}

func Test_CacheControl_SpecialFiles(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(New())

	// Add route handlers to prevent Fiber from setting default Content-Type
	app.Get("/index.html", func(c fiber.Ctx) error {
		return c.SendString("<html></html>")
	})
	app.Get("/manifest.webmanifest", func(c fiber.Ctx) error {
		return c.SendString("{}")
	})
	app.Get("/sw.js", func(c fiber.Ctx) error {
		return c.SendString("// service worker")
	})
	app.Get("/registerSW.js", func(c fiber.Ctx) error {
		return c.SendString("// register service worker")
	})

	tests := []struct {
		path       string
		expectedCC string
		expectedCT string
	}{
		{"/index.html", "no-cache, max-age=0, s-maxage=0, must-revalidate", ""},
		{"/manifest.webmanifest", "no-cache, max-age=0, s-maxage=0, must-revalidate", "application/manifest+json"},
		{"/sw.js", "no-cache, max-age=0, s-maxage=0, must-revalidate", ""},
		{"/registerSW.js", "no-cache, max-age=0, s-maxage=0, must-revalidate", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, tt.path, nil))
			require.NoError(t, err)
			require.Equal(t, tt.expectedCC, resp.Header.Get(fiber.HeaderCacheControl))
			if tt.expectedCT != "" {
				require.Equal(t, tt.expectedCT, resp.Header.Get(fiber.HeaderContentType))
			}
		})
	}
}

func Test_CacheControl_API_Routes(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(New())

	tests := []struct {
		path       string
		expectedCC string
	}{
		{"/api/users", "private, no-store"},
		{"/api/v1/products", "private, no-store"},
		{"/api/health", "private, no-store"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, tt.path, nil))
			require.NoError(t, err)
			require.Equal(t, tt.expectedCC, resp.Header.Get(fiber.HeaderCacheControl))
		})
	}
}

func Test_CacheControl_CustomRules(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(New(Config{
		Rules: []CacheRule{
			{
				Pattern:      regexp.MustCompile(`(?i)^.*/documents/.*\.pdf$`),
				CacheControl: "public, max-age=604800",
				ContentType:  "application/pdf",
			},
			{
				Pattern:      regexp.MustCompile(`(?i)^.*/temp/.*$`),
				CacheControl: "no-store",
			},
		},
		DefaultCacheControl: "public, max-age=300",
	}))

	// Add route handlers to prevent Fiber from setting default Content-Type
	app.Get("/documents/report.pdf", func(c fiber.Ctx) error {
		return c.SendString("PDF content")
	})
	app.Get("/temp/file.txt", func(c fiber.Ctx) error {
		return c.SendString("temp content")
	})
	app.Get("/other/file.txt", func(c fiber.Ctx) error {
		return c.SendString("other content")
	})

	tests := []struct {
		path       string
		expectedCC string
		expectedCT string
	}{
		{"/documents/report.pdf", "public, max-age=604800", "application/pdf"},
		{"/temp/file.txt", "no-store", ""},
		{"/other/file.txt", "public, max-age=300", ""}, // default
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, tt.path, nil))
			require.NoError(t, err)
			require.Equal(t, tt.expectedCC, resp.Header.Get(fiber.HeaderCacheControl))
			if tt.expectedCT != "" {
				require.Equal(t, tt.expectedCT, resp.Header.Get(fiber.HeaderContentType))
			}
		})
	}
}

func Test_CacheControl_Next(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(New(Config{
		Next: func(c fiber.Ctx) bool {
			return c.Path() == "/skip"
		},
	}))

	// Route that should be skipped
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/skip", nil))
	require.NoError(t, err)
	require.Empty(t, resp.Header.Get(fiber.HeaderCacheControl))

	// Route that should not be skipped
	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/process", nil))
	require.NoError(t, err)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderCacheControl))
}

func Test_CacheControl_CustomHeader(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(New(Config{
		CacheControlHeader:  "X-Custom-Cache",
		DefaultCacheControl: "public, max-age=60",
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, "public, max-age=60", resp.Header.Get("X-Custom-Cache"))
	require.Empty(t, resp.Header.Get(fiber.HeaderCacheControl))
}

func Test_CacheControl_CaseInsensitive(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(New())

	tests := []struct {
		path       string
		expectedCC string
	}{
		{"/INDEX.HTML", "no-cache, max-age=0, s-maxage=0, must-revalidate"},
		{"/STYLE.CSS", "public, max-age=3600"},
		{"/API/USERS", "private, no-store"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, tt.path, nil))
			require.NoError(t, err)
			require.Equal(t, tt.expectedCC, resp.Header.Get(fiber.HeaderCacheControl))
		})
	}
}

func Test_CacheControl_RuleOrder(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(New(Config{
		Rules: []CacheRule{
			{
				Pattern:      regexp.MustCompile(`(?i)^.*/special/.*\.js$`),
				CacheControl: "no-cache",
			},
			{
				Pattern:      regexp.MustCompile(`(?i)^.*/.*\.js$`),
				CacheControl: "public, max-age=3600",
			},
		},
	}))

	// The first matching rule should win
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/special/app.js", nil))
	require.NoError(t, err)
	require.Equal(t, "no-cache", resp.Header.Get(fiber.HeaderCacheControl))

	// Regular JS file should match second rule
	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/regular/app.js", nil))
	require.NoError(t, err)
	require.Equal(t, "public, max-age=3600", resp.Header.Get(fiber.HeaderCacheControl))
}

// Benchmark tests
func Benchmark_CacheControl(b *testing.B) {
	app := fiber.New()
	app.Use(New())

	app.Get("/static/style.css", func(c fiber.Ctx) error {
		return c.SendString("body { margin: 0; }")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/static/style.css", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := app.Test(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_CacheControl_CustomRules(b *testing.B) {
	app := fiber.New()
	app.Use(New(Config{
		Rules: []CacheRule{
			{Pattern: regexp.MustCompile(`(?i)^.*/api/.*$`), CacheControl: "no-cache"},
			{Pattern: regexp.MustCompile(`(?i)^.*/static/.*$`), CacheControl: "public, max-age=86400"},
			{Pattern: regexp.MustCompile(`(?i)^.*/documents/.*\.pdf$`), CacheControl: "public, max-age=604800"},
			{Pattern: regexp.MustCompile(`(?i)^.*/temp/.*$`), CacheControl: "no-store"},
			{Pattern: regexp.MustCompile(`(?i)^.*/.*\.(jpg|jpeg|png|gif)$`), CacheControl: "public, max-age=2592000"},
		},
		DefaultCacheControl: "public, max-age=3600",
	}))

	app.Get("/static/image.jpg", func(c fiber.Ctx) error {
		return c.SendString("image data")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/static/image.jpg", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := app.Test(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
