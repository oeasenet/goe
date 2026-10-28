package spa

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

// Config defines the config for SPA (Single Page Application) middleware.
type Config struct {
	// Next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	Next func(c fiber.Ctx) bool

	// Root is the root directory path of the SPA files.
	//
	// Required. Default: ""
	// Example: "./web/ui/dist" or "./build"
	Root string

	// Index is the name of the index file to serve for non-file routes.
	//
	// Optional. Default: "index.html"
	Index string

	// NotFoundFile is the file to serve when a route is not found.
	// If empty, Index will be used (standard SPA behavior).
	//
	// Optional. Default: "" (uses Index)
	NotFoundFile string

	// Exclude lists path prefixes the middleware leaves to the next handler —
	// never a file, never the index — so that, for example, "/api" keeps a
	// missing API route from being answered with the app's HTML. Prefixes match
	// whole segments: "/api" covers "/api" and "/api/x", not "/apiary".
	//
	// Optional. Default: nil
	Exclude []string

	// NotFoundForMissingFiles answers 404 for a path with a file extension that
	// does not exist, instead of the index: a stale hashed asset after a deploy
	// then fails visibly rather than arriving as HTML.
	//
	// Optional. Default: false
	NotFoundForMissingFiles bool

	// MaxAge defines the max-age header value for static files in seconds.
	// Set to 0 to disable caching headers.
	//
	// Optional. Default: 0
	MaxAge int

	// Browse enables directory browsing when a directory is requested.
	//
	// Optional. Default: false
	Browse bool

	// ByteRange enables byte range requests for static files.
	//
	// Optional. Default: false
	ByteRange bool

	// Compress enables gzip compression for static files.
	//
	// Optional. Default: false
	Compress bool

	// CacheDuration is the expiration time for inactive file handlers.
	// Use a negative value to disable the file handler cache.
	//
	// Optional. Default: 10 * time.Second
	CacheDuration time.Duration
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	Next:          nil,
	Root:          "",
	Index:         "index.html",
	NotFoundFile:  "",
	MaxAge:        0,
	Browse:        false,
	ByteRange:     false,
	Compress:      false,
	CacheDuration: 10 * time.Second,
}
