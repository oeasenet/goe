package http

import (
	"io/fs"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

// StaticConfig holds configuration for serving static files
// This is a wrapper around Fiber v3's static middleware config
type StaticConfig struct {
	// Root directory to serve static files from
	Root string

	// Index file names to look for when serving directories
	// Default: ["index.html"]
	IndexNames []string

	// Enable directory browsing
	// Default: false
	Browse bool

	// MaxAge in seconds for the Cache-Control header
	// Default: 0 (no caching)
	MaxAge int

	// Compress enables gzip/brotli compression
	// Default: false
	Compress bool

	// Download enables Content-Disposition: attachment header
	// Default: false
	Download bool

	// Next defines a function to skip this middleware when returned true
	Next func(c fiber.Ctx) bool
}

// Static creates a static file serving handler using Fiber v3's static middleware
//
// Usage:
//
//	// Serve files from ./public at root
//	app.Get("/*", http.Static("./public"))
//
//	// Serve files from ./assets at /static prefix
//	app.Get("/static/*", http.Static("./assets"))
//
//	// With custom configuration
//	app.Get("/*", http.Static("./public", http.StaticConfig{
//		IndexNames: []string{"index.html", "index.htm"},
//		MaxAge:     3600,
//		Compress:   true,
//	}))
func Static(root string, config ...StaticConfig) fiber.Handler {
	cfg := static.Config{
		FS: nil, // Use filesystem
	}

	// Apply custom config if provided
	if len(config) > 0 {
		c := config[0]

		if len(c.IndexNames) > 0 {
			cfg.IndexNames = c.IndexNames
		}
		cfg.Browse = c.Browse
		cfg.MaxAge = c.MaxAge
		cfg.Compress = c.Compress
		cfg.Download = c.Download
		cfg.Next = c.Next
	}

	return static.New(root, cfg)
}

// StaticFS creates a static file serving handler from an embedded filesystem
//
// Usage with embed.FS:
//
//	//go:embed static/*
//	var staticFS embed.FS
//
//	app.Get("/*", http.StaticFS(staticFS, http.StaticConfig{
//		IndexNames: []string{"index.html"},
//	}))
func StaticFS(filesystem fs.FS, config ...StaticConfig) fiber.Handler {
	cfg := static.Config{
		FS: filesystem,
	}

	// Apply custom config if provided
	if len(config) > 0 {
		c := config[0]

		if len(c.IndexNames) > 0 {
			cfg.IndexNames = c.IndexNames
		}
		cfg.Browse = c.Browse
		cfg.MaxAge = c.MaxAge
		cfg.Compress = c.Compress
		cfg.Download = c.Download
		cfg.Next = c.Next
	}

	return static.New("", cfg)
}
