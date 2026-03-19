package spa

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// New creates a new SPA serving middleware handler.
//
// This middleware serves static files and falls back to serving an index file
// for client-side routing in Single Page Applications (SPAs).
//
// Usage example:
//
//	app.Use(spa.New(spa.Config{
//		Root: "./web/ui/dist",
//	}))
//
// Advanced configuration:
//
//	app.Use(spa.New(spa.Config{
//		Root:     "./build",
//		Index:    "app.html",
//		MaxAge:   3600,
//		Compress: true,
//	}))
func New(config ...Config) fiber.Handler {
	// Set default config
	cfg := configDefault(config...)

	// Validate configuration
	if cfg.Root == "" {
		panic("spa: Root path is required")
	}

	// Ensure root path is absolute
	absRoot, err := filepath.Abs(cfg.Root)
	if err != nil {
		panic("spa: failed to get absolute path: " + err.Error())
	}
	cfg.Root = absRoot

	// Create a static file handler with the configuration
	// Note: This is a simplified approach. In production, you might want to use
	// the actual static middleware with proper configuration
	serveStatic := func(c fiber.Ctx, filePath string) error {
		// Set cache control headers if MaxAge is specified
		if cfg.MaxAge > 0 {
			c.Set(fiber.HeaderCacheControl, "public, max-age="+strconv.Itoa(cfg.MaxAge))
		}

		// Serve the file
		return c.SendFile(filePath)
	}

	// Determine the fallback file
	fallbackFile := cfg.Index
	if cfg.NotFoundFile != "" {
		fallbackFile = cfg.NotFoundFile
	}

	return func(c fiber.Ctx) error {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		// Get the requested path
		path := c.Path()

		// Check if this could be a file request (has extension or ends with /)
		// This is a heuristic to avoid file system checks for obvious routes
		if mightBeFile(path) {
			// Try to serve as static file first
			// We'll check if the file exists before serving
			fullPath := filepath.Join(cfg.Root, path)
			// Prevent path traversal outside the root directory
			if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(cfg.Root)) {
				return c.Next()
			}
			if fileExists(fullPath) {
				return serveStatic(c, fullPath)
			}

			}

		// For SPA routing, serve the index file for all non-file routes
		// This allows client-side routing to work
		indexPath := filepath.Join(cfg.Root, fallbackFile)

		// Check if index file exists
		if !fileExists(indexPath) {
			// If index doesn't exist, continue to next handler
			// This prevents infinite loops and allows proper 404 handling
			return c.Next()
		}

		// Serve the index file
		return c.SendFile(indexPath)
	}
}

// Helper function to set default values
func configDefault(config ...Config) Config {
	// Return default config if nothing provided
	if len(config) < 1 {
		cfg := ConfigDefault
		return cfg
	}

	// Override default config
	cfg := config[0]

	// Set default values
	if cfg.Index == "" {
		cfg.Index = ConfigDefault.Index
	}

	if cfg.CacheDuration == 0 {
		cfg.CacheDuration = ConfigDefault.CacheDuration
	}

	return cfg
}

// mightBeFile checks if a path might be a file request
func mightBeFile(path string) bool {
	// Check if path has a file extension
	if strings.Contains(filepath.Base(path), ".") {
		return true
	}

	// Check if path ends with / (directory request)
	if strings.HasSuffix(path, "/") {
		return true
	}

	// Special cases for common static directories
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) > 0 {
		firstSegment := segments[0]
		staticDirs := []string{"assets", "static", "public", "dist", "img", "images", "js", "css", "fonts"}
		if slices.Contains(staticDirs, firstSegment) {
			return true
		}
	}

	return false
}

// fileExists checks if a file exists and is not a directory
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

