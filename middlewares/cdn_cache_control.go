package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"regexp"
)

// CacheRule represents a single caching rule with a regex pattern and corresponding cache control value
type CacheRule struct {
	// Pattern is the regex pattern to match against the URL
	Pattern *regexp.Regexp
	// CacheControl is the Cache-Control header value to set when the pattern matches
	CacheControl string
	// ContentType is an optional Content-Type header to set when the pattern matches
	ContentType string
}

// CDNCacheControlConfig holds the configuration for the CDN cache control middleware
type CDNCacheControlConfig struct {
	// Rules is a slice of CacheRule that will be applied in order
	Rules []CacheRule
	// DefaultCacheControl is the Cache-Control header value to set when no pattern matches
	DefaultCacheControl string
}

// CDNCacheControlMiddleware is a middleware for setting Cache-Control headers based on URL patterns
type CDNCacheControlMiddleware struct {
	config CDNCacheControlConfig
}

// DefaultCDNCacheControlConfig returns the default configuration for the CDN cache control middleware
func DefaultCDNCacheControlConfig() CDNCacheControlConfig {
	return CDNCacheControlConfig{
		Rules: []CacheRule{
			{
				Pattern:      regexp.MustCompile(`(?i)^.*/index\.html$`),
				CacheControl: "no-cache, max-age=0, s-maxage=0, must-revalidate",
			},
			{
				Pattern:      regexp.MustCompile(`(?i)^.*/(manifest\.webmanifest)$`),
				CacheControl: "no-cache, max-age=0, s-maxage=0, must-revalidate",
				ContentType:  "application/manifest+json",
			},
			{
				Pattern:      regexp.MustCompile(`(?i)^.*/(registerSW\.js|sw\.js)$`),
				CacheControl: "no-cache, max-age=0, s-maxage=0, must-revalidate",
			},
			{
				Pattern:      regexp.MustCompile(`(?i)^.*/api/.*$`),
				CacheControl: "private, no-store",
			},
			{
				Pattern:      regexp.MustCompile(`(?i)^.*/\.well-known/.*$`),
				CacheControl: "private, no-store",
			},
			{
				Pattern:      regexp.MustCompile(`(?i)^.*/.*\.(css|js|png|jpg|jpeg|gif|svg|webp|ico|woff|woff2|ttf|otf|eot|mp4|webm|m4v|mp3|ogg|wav|flac)$`),
				CacheControl: "public, max-age=3600",
			},
		},
		DefaultCacheControl: "no-cache, max-age=0, s-maxage=0, must-revalidate",
	}
}

// NewCDNCacheControlMiddleware creates a new CDN cache control middleware with default configuration.
// Optimized for Fastly CDN.
//
// Usage example:
//
//	app.Use(middlewares.NewCDNCacheControlMiddleware().Handler())
func NewCDNCacheControlMiddleware() *CDNCacheControlMiddleware {
	return &CDNCacheControlMiddleware{
		config: DefaultCDNCacheControlConfig(),
	}
}

// NewCDNCacheControlMiddlewareWithConfig creates a new CDN cache control middleware with custom configuration.
//
// Usage example:
//
//	config := middlewares.DefaultCDNCacheControlConfig()
//	config.DefaultCacheControl = "public, max-age=60"
//	app.Use(middlewares.NewCDNCacheControlMiddlewareWithConfig(config).Handler())
func NewCDNCacheControlMiddlewareWithConfig(config CDNCacheControlConfig) *CDNCacheControlMiddleware {
	return &CDNCacheControlMiddleware{
		config: config,
	}
}

// WithRule adds a new rule to the middleware configuration.
// Returns the middleware instance for method chaining.
//
// Usage example:
//
//	middleware := middlewares.NewCDNCacheControlMiddleware().
//		WithRule(regexp.MustCompile(`(?i)^.*/custom/.*$`), "public, max-age=120", "").
//		WithRule(regexp.MustCompile(`(?i)^.*/special\.json$`), "private, no-store", "application/json")
//	app.Use(middleware.Handler())
func (m *CDNCacheControlMiddleware) WithRule(pattern *regexp.Regexp, cacheControl string, contentType string) *CDNCacheControlMiddleware {
	m.config.Rules = append(m.config.Rules, CacheRule{
		Pattern:      pattern,
		CacheControl: cacheControl,
		ContentType:  contentType,
	})
	return m
}

// WithDefaultCacheControl sets the default cache control value.
// Returns the middleware instance for method chaining.
//
// Usage example:
//
//	middleware := middlewares.NewCDNCacheControlMiddleware().
//		WithDefaultCacheControl("public, max-age=60")
//	app.Use(middleware.Handler())
func (m *CDNCacheControlMiddleware) WithDefaultCacheControl(cacheControl string) *CDNCacheControlMiddleware {
	m.config.DefaultCacheControl = cacheControl
	return m
}

// Handler returns a Fiber middleware handler function.
func (m *CDNCacheControlMiddleware) Handler() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		// Get the full URL
		fullURL := ctx.OriginalURL()

		// Apply caching rules based on regex matching
		for _, rule := range m.config.Rules {
			if rule.Pattern.MatchString(fullURL) {
				ctx.Set("Cache-Control", rule.CacheControl)
				if rule.ContentType != "" {
					ctx.Set("Content-Type", rule.ContentType)
				}
				return ctx.Next()
			}
		}

		// Default cache setting for unmatched URLs
		ctx.Set("Cache-Control", m.config.DefaultCacheControl)

		// Proceed to the next middleware or route
		return ctx.Next()
	}
}

// Deprecated: Use NewCDNCacheControlMiddleware().Handler() instead.
// This function is kept for backward compatibility.
func NewCDNCacheControlMiddlewareHandler() fiber.Handler {
	return NewCDNCacheControlMiddleware().Handler()
}
