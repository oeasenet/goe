package cachecontrol

import "github.com/gofiber/fiber/v3"

// New creates a new CDN cache control middleware handler.
//
// This middleware sets Cache-Control headers based on URL patterns,
// optimized for CDN usage (e.g., Fastly, CloudFront, Cloudflare).
//
// Usage example:
//
//	app.Use(cache_control.New())
//
// Custom configuration:
//
//	app.Use(cache_control.New(cache_control.Config{
//		Rules: []cache_control.CacheRule{
//			{
//				Pattern:      regexp.MustCompile(`(?i)^.*/static/.*$`),
//				CacheControl: "public, max-age=86400",
//			},
//		},
//		DefaultCacheControl: "public, max-age=60",
//	}))
func New(config ...Config) fiber.Handler {
	// Set default config
	cfg := configDefault(config...)

	// Return new handler
	return func(c fiber.Ctx) error {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		// Get the full URL
		fullURL := c.OriginalURL()

		// Apply caching rules based on regex matching
		for _, rule := range cfg.Rules {
			if rule.Pattern.MatchString(fullURL) {
				c.Set(cfg.CacheControlHeader, rule.CacheControl)
				if rule.ContentType != "" {
					c.Set(fiber.HeaderContentType, rule.ContentType)
				}
				return c.Next()
			}
		}

		// Default cache setting for unmatched URLs
		c.Set(cfg.CacheControlHeader, cfg.DefaultCacheControl)

		// Proceed to the next middleware or route
		return c.Next()
	}
}

// Helper function to set default values
func configDefault(config ...Config) Config {
	// Return default config if nothing provided
	if len(config) < 1 {
		return ConfigDefault
	}

	// Override default config
	cfg := config[0]

	// Set default values
	if cfg.Next == nil {
		cfg.Next = ConfigDefault.Next
	}

	// If no rules provided, use default rules
	if len(cfg.Rules) == 0 {
		cfg.Rules = ConfigDefault.Rules
	}

	// Set default cache control if isn't provided
	if cfg.DefaultCacheControl == "" {
		cfg.DefaultCacheControl = ConfigDefault.DefaultCacheControl
	}

	// Set the default header name if not provided
	if cfg.CacheControlHeader == "" {
		cfg.CacheControlHeader = ConfigDefault.CacheControlHeader
	}

	return cfg
}
