package http

import (
	"io"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
)

// routeLimit is one WithRouteLimits entry.
type routeLimit struct {
	prefix string
	limits RouteLimits
}

// guard is the middleware GOE puts in front of the app's routes: it rejects
// paths with dot segments and, when strict body limits are on, bodies over
// their limit.
type guard struct {
	allowDotSegments bool
	strict           bool
	bodyLimit        int
	routes           []routeLimit // longest prefix first
}

func newGuard(s *settings) *guard {
	routes := slices.Clone(s.routeLimits)
	slices.SortStableFunc(routes, func(a, b routeLimit) int { return len(b.prefix) - len(a.prefix) })
	return &guard{
		allowDotSegments: s.allowDotSegments,
		strict:           s.strictBodyLimits(),
		bodyLimit:        s.fiber.BodyLimit,
		routes:           routes,
	}
}

func (g *guard) handle(c fiber.Ctx) error {
	if !g.allowDotSegments && hasDotSegment(c.Request().URI().PathOriginal()) {
		return fiber.ErrBadRequest
	}
	if !g.strict {
		return c.Next()
	}

	limits := g.limitsFor(c.Path())
	if limits.ReadTimeout > 0 {
		ctx := c.RequestCtx()
		_ = ctx.SetReadDeadline(ctx.Time().Add(limits.ReadTimeout))
	}
	limit := limits.BodyLimit
	if limit <= 0 {
		limit = g.bodyLimit
	}

	// fasthttp reads at most a few KB of a streamed body before handlers run,
	// so a declared length over the limit is refused before the body arrives.
	contentLength := c.Request().Header.ContentLength()
	if contentLength > limit {
		c.Response().SetConnectionClose()
		return fiber.ErrRequestEntityTooLarge
	}

	// A chunked body has no declared length: cap the stream instead.
	var capped *cappedBody
	if contentLength == -1 && c.Request().IsBodyStream() {
		capped = &cappedBody{r: c.Request().BodyStream(), remaining: int64(limit)}
		c.Request().SetBodyStream(capped, -1)
	}

	err := c.Next()
	// The handler failed after the body went over its limit (reading the
	// stream, or binding the error text c.Body() returns in that case).
	if capped != nil && capped.exceeded && err != nil {
		c.Response().SetConnectionClose()
		return fiber.ErrRequestEntityTooLarge
	}
	return err
}

// limitsFor returns the limits of the longest prefix path falls under.
func (g *guard) limitsFor(path string) RouteLimits {
	for _, route := range g.routes {
		if route.prefix == "/" || path == route.prefix || strings.HasPrefix(path, route.prefix+"/") {
			return route.limits
		}
	}
	return RouteLimits{}
}

// cappedBody fails once more than remaining bytes have been read.
type cappedBody struct {
	r         io.Reader
	remaining int64
	exceeded  bool
}

func (b *cappedBody) Read(p []byte) (int, error) {
	if b.exceeded {
		return 0, fasthttp.ErrBodyTooLarge
	}
	if int64(len(p)) > b.remaining+1 {
		p = p[:b.remaining+1]
	}
	n, err := b.r.Read(p)
	b.remaining -= int64(n)
	if b.remaining < 0 {
		b.exceeded = true
		return n + int(b.remaining), fasthttp.ErrBodyTooLarge
	}
	return n, err
}

// hasDotSegment reports whether a raw request path has a "." or ".." segment,
// including when those dots, or the slashes around them, are percent-encoded.
// Invalid escapes stay literal: they cannot spell a dot.
func hasDotSegment(path []byte) bool {
	if !slices.Contains(path, '.') && !slices.Contains(path, '%') {
		return false
	}
	length, onlyDots := 0, true
	endSegment := func() bool {
		dotSegment := onlyDots && (length == 1 || length == 2)
		length, onlyDots = 0, true
		return dotSegment
	}
	for i := 0; i < len(path); i++ {
		ch := path[i]
		if ch == '%' && i+2 < len(path) && isHex(path[i+1]) && isHex(path[i+2]) {
			ch = unhex(path[i+1])<<4 | unhex(path[i+2])
			i += 2
		}
		if ch == '/' || ch == '\\' {
			if endSegment() {
				return true
			}
			continue
		}
		length++
		onlyDots = onlyDots && ch == '.'
	}
	return endSegment()
}

func isHex(c byte) bool {
	return '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F'
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	default:
		return c - 'A' + 10
	}
}
