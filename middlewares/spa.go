package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

type SPAMiddleware struct {
	// rootPath is the root path of the SPA, e.g. './web/ui/dist'
	rootPath string
	// indexFileFileName is the name of the index file, e.g. 'index.html'
	indexFileFileName string
}

// NewSPAServingMiddleware create a new instance of the SPAMiddleware struct.
// It takes the rootPath and indexFile as parameters and initializes the struct with these values.
// IMPORTANT: It's really important to handle both static files and index file in a SPA application.
// Usage example:
// middleware := NewSPAServingMiddleware("./web/ui/dist", "index.html")
// app.Use("/", middleware.HandleStaticFiles())
// app.Get("/*", middleware.HandleIndexFile())
func NewSPAServingMiddleware(rootPath string, indexFile string) *SPAMiddleware {
	return &SPAMiddleware{
		rootPath:          rootPath,
		indexFileFileName: indexFile,
	}
}

// HandleStaticFiles returns a middleware handler for serving static files.
// It creates a new static file server using the root path specified in the SPAMiddleware struct.
// Usage example:
// app.Use("/", spm.HandleStaticFiles())
func (s *SPAMiddleware) HandleStaticFiles() fiber.Handler {
	return static.New(s.rootPath)
}

// HandleIndexFile returns a middleware handler for serving the index file.
// It creates a new handler that sends the index file specified in the SPAMiddleware struct.
// Usage example:
// app.Get("/*", spm.HandleIndexFile())
func (s *SPAMiddleware) HandleIndexFile() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		return ctx.SendFile(s.rootPath + "/" + s.indexFileFileName)
	}
}
