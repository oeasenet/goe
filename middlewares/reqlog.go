package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/valyala/bytebufferpool"
	"go.oease.dev/goe"
	"strconv"
	"strings"
)

func NewRequestLoggingMiddleware(skipStaticRec ...bool) fiber.Handler {
	skipStatic := false
	if len(skipStaticRec) > 0 {
		skipStatic = skipStaticRec[0]
	}
	defaultCfg := logger.ConfigDefault
	defaultCfg.Next = func(c fiber.Ctx) bool {
		if skipStatic {
			return staticResourceSkipper(c)
		}
		return false
	}
	defaultCfg.LoggerFunc = func(c fiber.Ctx, data *logger.Data, cfg logger.Config) error {
		// Get new buffer
		buf := bytebufferpool.Get()

		// Format error if exist
		formatErr := ""
		if data.ChainErr != nil {
			formatErr = " | " + data.ChainErr.Error()
		}
		buf.WriteString(c.Method())
		buf.WriteString(" | ")
		buf.WriteString(strconv.Itoa(c.Response().StatusCode()))
		buf.WriteString(" | ")
		buf.WriteString(data.Stop.Sub(data.Start).String())
		buf.WriteString(" | ")
		buf.WriteString(c.IP())
		buf.WriteString(" | ")
		buf.WriteString(c.Path())
		buf.WriteString(" | ")
		buf.WriteString(formatErr)

		if cfg.Done != nil {
			cfg.Done(c, buf.Bytes())
		}

		// Write buffer to output
		goe.UseLog().Info(string(buf.Bytes()))

		// Put buffer back to pool
		bytebufferpool.Put(buf)

		// End chain
		return nil
	}
	return logger.New(defaultCfg)
}

// Skipper function to skip requests for frontend static resources
func staticResourceSkipper(c fiber.Ctx) bool {
	staticExtensions := []string{".html", ".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico"}
	path := c.Path()

	for _, ext := range staticExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}
