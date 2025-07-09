package middlewares

import (
	"io"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultCDNCacheControlConfig(t *testing.T) {
	config := DefaultCDNCacheControlConfig()

	// Check that we have default rules
	assert.NotEmpty(t, config.Rules)
	assert.Equal(t, "no-cache, max-age=0, s-maxage=0, must-revalidate", config.DefaultCacheControl)

	// Check that default rules cover expected patterns by testing actual URLs
	testUrls := []struct {
		url                  string
		shouldMatch          bool
		expectedCacheControl string
	}{
		{"/index.html", true, "no-cache, max-age=0, s-maxage=0, must-revalidate"},
		{"/manifest.webmanifest", true, "no-cache, max-age=0, s-maxage=0, must-revalidate"},
		{"/registerSW.js", true, "no-cache, max-age=0, s-maxage=0, must-revalidate"},
		{"/sw.js", true, "no-cache, max-age=0, s-maxage=0, must-revalidate"},
		{"/api/users", true, "private, no-store"},
		{"/.well-known/security.txt", true, "private, no-store"},
		{"/assets/styles.css", true, "public, max-age=3600"},
		{"/assets/app.js", true, "public, max-age=3600"},
		{"/images/logo.png", true, "public, max-age=3600"},
		{"/images/photo.jpg", true, "public, max-age=3600"},
		{"/images/photo.jpeg", true, "public, max-age=3600"},
		{"/images/icon.gif", true, "public, max-age=3600"},
		{"/icons/star.svg", true, "public, max-age=3600"},
		{"/images/bg.webp", true, "public, max-age=3600"},
		{"/favicon.ico", true, "public, max-age=3600"},
		{"/fonts/font.woff", true, "public, max-age=3600"},
		{"/fonts/font.woff2", true, "public, max-age=3600"},
		{"/fonts/font.ttf", true, "public, max-age=3600"},
		{"/fonts/font.otf", true, "public, max-age=3600"},
		{"/fonts/font.eot", true, "public, max-age=3600"},
		{"/videos/demo.mp4", true, "public, max-age=3600"},
		{"/videos/demo.webm", true, "public, max-age=3600"},
		{"/videos/demo.m4v", true, "public, max-age=3600"},
		{"/audio/sound.mp3", true, "public, max-age=3600"},
		{"/audio/sound.ogg", true, "public, max-age=3600"},
		{"/audio/sound.wav", true, "public, max-age=3600"},
		{"/audio/sound.flac", true, "public, max-age=3600"},
		{"/random/path", false, "no-cache, max-age=0, s-maxage=0, must-revalidate"},
	}

	for _, testUrl := range testUrls {
		matched := false
		var matchedCacheControl string

		// Test each rule to see if it matches
		for _, rule := range config.Rules {
			if rule.Pattern.MatchString(testUrl.url) {
				matched = true
				matchedCacheControl = rule.CacheControl
				break
			}
		}

		if !matched {
			matchedCacheControl = config.DefaultCacheControl
		}

		if testUrl.shouldMatch {
			assert.Equal(t, testUrl.expectedCacheControl, matchedCacheControl,
				"URL %s should match with cache control %s", testUrl.url, testUrl.expectedCacheControl)
		}
	}
}

func TestNewCDNCacheControlMiddleware(t *testing.T) {
	middleware := NewCDNCacheControlMiddleware()

	assert.NotNil(t, middleware)
	assert.NotEmpty(t, middleware.config.Rules)
	assert.Equal(t, "no-cache, max-age=0, s-maxage=0, must-revalidate", middleware.config.DefaultCacheControl)
}

func TestNewCDNCacheControlMiddlewareWithConfig(t *testing.T) {
	config := CDNCacheControlConfig{
		Rules: []CacheRule{
			{
				Pattern:      regexp.MustCompile(`(?i)^.*/test\.html$`),
				CacheControl: "public, max-age=300",
				ContentType:  "text/html",
			},
		},
		DefaultCacheControl: "private, no-store",
	}

	middleware := NewCDNCacheControlMiddlewareWithConfig(config)

	assert.NotNil(t, middleware)
	assert.Len(t, middleware.config.Rules, 1)
	assert.Equal(t, "private, no-store", middleware.config.DefaultCacheControl)
}

func TestCDNCacheControlMiddleware_WithRule(t *testing.T) {
	middleware := NewCDNCacheControlMiddleware()
	originalRulesCount := len(middleware.config.Rules)

	pattern := regexp.MustCompile(`(?i)^.*/custom/.*$`)
	result := middleware.WithRule(pattern, "public, max-age=120", "application/json")

	// Should return the same instance for method chaining
	assert.Equal(t, middleware, result)

	// Should have added one more rule
	assert.Len(t, middleware.config.Rules, originalRulesCount+1)

	// Check the new rule
	newRule := middleware.config.Rules[len(middleware.config.Rules)-1]
	assert.Equal(t, pattern, newRule.Pattern)
	assert.Equal(t, "public, max-age=120", newRule.CacheControl)
	assert.Equal(t, "application/json", newRule.ContentType)
}

func TestCDNCacheControlMiddleware_WithDefaultCacheControl(t *testing.T) {
	middleware := NewCDNCacheControlMiddleware()

	result := middleware.WithDefaultCacheControl("public, max-age=60")

	// Should return the same instance for method chaining
	assert.Equal(t, middleware, result)
	assert.Equal(t, "public, max-age=60", middleware.config.DefaultCacheControl)
}

func TestCDNCacheControlMiddleware_Handler(t *testing.T) {
	tests := []struct {
		name                 string
		url                  string
		expectedCacheControl string
		expectedContentType  string
	}{
		{
			name:                 "index.html",
			url:                  "/index.html",
			expectedCacheControl: "no-cache, max-age=0, s-maxage=0, must-revalidate",
			expectedContentType:  "",
		},
		{
			name:                 "manifest.webmanifest",
			url:                  "/manifest.webmanifest",
			expectedCacheControl: "no-cache, max-age=0, s-maxage=0, must-revalidate",
			expectedContentType:  "application/manifest+json",
		},
		{
			name:                 "service worker",
			url:                  "/registerSW.js",
			expectedCacheControl: "no-cache, max-age=0, s-maxage=0, must-revalidate",
			expectedContentType:  "",
		},
		{
			name:                 "sw.js",
			url:                  "/sw.js",
			expectedCacheControl: "no-cache, max-age=0, s-maxage=0, must-revalidate",
			expectedContentType:  "",
		},
		{
			name:                 "api endpoint",
			url:                  "/api/users",
			expectedCacheControl: "private, no-store",
			expectedContentType:  "",
		},
		{
			name:                 "well-known",
			url:                  "/.well-known/security.txt",
			expectedCacheControl: "private, no-store",
			expectedContentType:  "",
		},
		{
			name:                 "css file",
			url:                  "/assets/styles.css",
			expectedCacheControl: "public, max-age=3600",
			expectedContentType:  "",
		},
		{
			name:                 "js file",
			url:                  "/assets/app.js",
			expectedCacheControl: "public, max-age=3600",
			expectedContentType:  "",
		},
		{
			name:                 "png image",
			url:                  "/images/logo.png",
			expectedCacheControl: "public, max-age=3600",
			expectedContentType:  "",
		},
		{
			name:                 "jpg image",
			url:                  "/images/photo.jpg",
			expectedCacheControl: "public, max-age=3600",
			expectedContentType:  "",
		},
		{
			name:                 "svg image",
			url:                  "/icons/star.svg",
			expectedCacheControl: "public, max-age=3600",
			expectedContentType:  "",
		},
		{
			name:                 "woff font",
			url:                  "/fonts/font.woff",
			expectedCacheControl: "public, max-age=3600",
			expectedContentType:  "",
		},
		{
			name:                 "woff2 font",
			url:                  "/fonts/font.woff2",
			expectedCacheControl: "public, max-age=3600",
			expectedContentType:  "",
		},
		{
			name:                 "mp4 video",
			url:                  "/videos/demo.mp4",
			expectedCacheControl: "public, max-age=3600",
			expectedContentType:  "",
		},
		{
			name:                 "mp3 audio",
			url:                  "/audio/sound.mp3",
			expectedCacheControl: "public, max-age=3600",
			expectedContentType:  "",
		},
		{
			name:                 "unmatched url",
			url:                  "/some/random/path",
			expectedCacheControl: "no-cache, max-age=0, s-maxage=0, must-revalidate",
			expectedContentType:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			middleware := NewCDNCacheControlMiddleware()

			app.Use(middleware.Handler())
			app.Get("/*", func(c fiber.Ctx) error {
				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", tt.url, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCacheControl, resp.Header.Get("Cache-Control"))

			if tt.expectedContentType != "" {
				assert.Equal(t, tt.expectedContentType, resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestCDNCacheControlMiddleware_Handler_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name                 string
		url                  string
		expectedCacheControl string
	}{
		{
			name:                 "INDEX.HTML uppercase",
			url:                  "/INDEX.HTML",
			expectedCacheControl: "no-cache, max-age=0, s-maxage=0, must-revalidate",
		},
		{
			name:                 "App.JS mixed case",
			url:                  "/assets/App.JS",
			expectedCacheControl: "public, max-age=3600",
		},
		{
			name:                 "LOGO.PNG uppercase",
			url:                  "/images/LOGO.PNG",
			expectedCacheControl: "public, max-age=3600",
		},
		{
			name:                 "API endpoint uppercase",
			url:                  "/API/users",
			expectedCacheControl: "private, no-store",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			middleware := NewCDNCacheControlMiddleware()

			app.Use(middleware.Handler())
			app.Get("/*", func(c fiber.Ctx) error {
				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", tt.url, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCacheControl, resp.Header.Get("Cache-Control"))
		})
	}
}

func TestCDNCacheControlMiddleware_Handler_CustomConfig(t *testing.T) {
	config := CDNCacheControlConfig{
		Rules: []CacheRule{
			{
				Pattern:      regexp.MustCompile(`(?i)^.*/test\.html$`),
				CacheControl: "public, max-age=300",
				ContentType:  "text/html",
			},
			{
				Pattern:      regexp.MustCompile(`(?i)^.*/custom/.*$`),
				CacheControl: "private, max-age=60",
			},
		},
		DefaultCacheControl: "public, max-age=120",
	}

	middleware := NewCDNCacheControlMiddlewareWithConfig(config)

	tests := []struct {
		name                 string
		url                  string
		expectedCacheControl string
		expectedContentType  string
	}{
		{
			name:                 "matching test.html",
			url:                  "/test.html",
			expectedCacheControl: "public, max-age=300",
			expectedContentType:  "text/html",
		},
		{
			name:                 "matching custom path",
			url:                  "/custom/file.txt",
			expectedCacheControl: "private, max-age=60",
			expectedContentType:  "",
		},
		{
			name:                 "default rule",
			url:                  "/other/file.txt",
			expectedCacheControl: "public, max-age=120",
			expectedContentType:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Use(middleware.Handler())
			app.Get("/*", func(c fiber.Ctx) error {
				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", tt.url, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCacheControl, resp.Header.Get("Cache-Control"))

			if tt.expectedContentType != "" {
				assert.Equal(t, tt.expectedContentType, resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestCDNCacheControlMiddleware_Handler_RuleOrder(t *testing.T) {
	// Test that rules are applied in order (first match wins)
	config := CDNCacheControlConfig{
		Rules: []CacheRule{
			{
				Pattern:      regexp.MustCompile(`(?i)^.*/.*\.js$`),
				CacheControl: "public, max-age=300",
			},
			{
				Pattern:      regexp.MustCompile(`(?i)^.*/special\.js$`),
				CacheControl: "private, no-store",
			},
		},
		DefaultCacheControl: "public, max-age=120",
	}

	middleware := NewCDNCacheControlMiddlewareWithConfig(config)

	app := fiber.New()
	app.Use(middleware.Handler())
	app.Get("/*", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// The first rule should match and apply, even though the second rule would also match
	req := httptest.NewRequest("GET", "/special.js", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, "public, max-age=300", resp.Header.Get("Cache-Control"))
}

func TestCDNCacheControlMiddleware_MethodChaining(t *testing.T) {
	middleware := NewCDNCacheControlMiddleware().
		WithRule(regexp.MustCompile(`(?i)^.*/custom/.*$`), "public, max-age=120", "").
		WithRule(regexp.MustCompile(`(?i)^.*/special\.json$`), "private, no-store", "application/json").
		WithDefaultCacheControl("public, max-age=60")

	assert.NotNil(t, middleware)
	assert.Equal(t, "public, max-age=60", middleware.config.DefaultCacheControl)

	// Should have original rules plus 2 new ones
	expectedRulesCount := len(DefaultCDNCacheControlConfig().Rules) + 2
	assert.Len(t, middleware.config.Rules, expectedRulesCount)
}

func TestNewCDNCacheControlMiddlewareHandler_Deprecated(t *testing.T) {
	app := fiber.New()
	handler := NewCDNCacheControlMiddlewareHandler()

	app.Use(handler)
	app.Get("/*", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test.css", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, "public, max-age=3600", resp.Header.Get("Cache-Control"))
}

func TestCDNCacheControlMiddleware_Handler_Integration(t *testing.T) {
	app := fiber.New()
	middleware := NewCDNCacheControlMiddleware()

	app.Use(middleware.Handler())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Home")
	})
	app.Get("/api/users", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"users": []string{"user1", "user2"}})
	})
	app.Get("/assets/app.js", func(c fiber.Ctx) error {
		return c.SendString("/* JavaScript code */")
	})

	tests := []struct {
		name                 string
		url                  string
		expectedCacheControl string
		expectedStatus       int
	}{
		{
			name:                 "home page",
			url:                  "/",
			expectedCacheControl: "no-cache, max-age=0, s-maxage=0, must-revalidate",
			expectedStatus:       200,
		},
		{
			name:                 "api endpoint",
			url:                  "/api/users",
			expectedCacheControl: "private, no-store",
			expectedStatus:       200,
		},
		{
			name:                 "javascript asset",
			url:                  "/assets/app.js",
			expectedCacheControl: "public, max-age=3600",
			expectedStatus:       200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, tt.expectedCacheControl, resp.Header.Get("Cache-Control"))

			// Verify the response body
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.NotEmpty(t, body)
		})
	}
}
