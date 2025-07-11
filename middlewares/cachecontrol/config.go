package cachecontrol

import (
	"github.com/gofiber/fiber/v3"
	"regexp"
)

// Config defines the config for CDN Cache Control middleware.
type Config struct {
	// Next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	Next func(c fiber.Ctx) bool

	// Rules is a slice of CacheRule that will be applied in order
	//
	// Optional. Default: See ConfigDefault
	Rules []CacheRule

	// DefaultCacheControl is the Cache-Control header value to set when no pattern matches
	//
	// Optional. Default: "no-cache, max-age=0, s-maxage=0, must-revalidate"
	DefaultCacheControl string

	// CacheControlHeader allows you to customize the header name for cache control
	//
	// Optional. Default: "Cache-Control"
	CacheControlHeader string
}

// CacheRule represents a single caching rule with a regex pattern and corresponding cache control value
type CacheRule struct {
	// Pattern is the regex pattern to match against the URL
	Pattern *regexp.Regexp

	// CacheControl is the Cache-Control header value to set when the pattern matches
	CacheControl string

	// ContentType is an optional Content-Type header to set when the pattern matches
	ContentType string
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	Next: nil,
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
	CacheControlHeader:  fiber.HeaderCacheControl,
}
