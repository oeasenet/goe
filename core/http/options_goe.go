package http

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

// This file holds the options GOE defines itself, because Fiber has no
// equivalent field: the listener address, the request-id middleware, the
// html/template engine GOE constructs on your behalf, and the request guard. Everything else follows
// Fiber's field names (see options_fiber.go and options_listen.go).

// WithHost sets the interface the server binds to. Replaces HTTP_HOST.
// Defaults to 0.0.0.0.
func WithHost(host string) Option {
	return func(s *settings) error {
		if host == "" {
			return errors.New("WithHost: host must not be empty")
		}
		s.host = host
		return nil
	}
}

// WithPort sets the port the server binds to. Replaces HTTP_PORT and the
// deprecated goe.Options.HTTPPort. Defaults to 8080.
//
// Port 0 asks the operating system for an arbitrary free port, which is useful
// in tests.
func WithPort(port int) Option {
	return func(s *settings) error {
		if port < 0 || port > 65535 {
			return fmt.Errorf("WithPort: port %d is out of range 0-65535", port)
		}
		s.port = port
		return nil
	}
}

// WithRequestID enables or disables GOE's request-id middleware. Replaces
// HTTP_REQUEST_ID. Enabled by default.
//
// The middleware reuses an inbound request id when the configured header is
// present and generates one otherwise. Disable it to own the middleware stack
// yourself; http.WithReqCtx still falls back to reading the raw header.
func WithRequestID(enabled bool) Option {
	return func(s *settings) error { s.requestIDEnabled = enabled; return nil }
}

// WithRequestIDHeader sets the header the request-id middleware reads and
// writes. Replaces HTTP_REQUEST_ID_HEADER. Defaults to X-Request-ID.
func WithRequestIDHeader(header string) Option {
	return func(s *settings) error {
		if header == "" {
			return errors.New("WithRequestIDHeader: header must not be empty")
		}
		s.requestIDHeader = header
		return nil
	}
}

// WithHTMLViews configures GOE to build a html/template view engine rooted at
// root, loading files with the given extension. Replaces VIEWS_ENGINE,
// VIEWS_ROOT and VIEWS_EXT. Use WithViewsLayout to set a global layout.
//
// This is a convenience over WithViews for the common case. If you supply your
// own engine through WithViews, that engine is used and this option is ignored.
func WithHTMLViews(root, ext string) Option {
	return func(s *settings) error {
		if root == "" {
			return errors.New("WithHTMLViews: root must not be empty")
		}
		if ext == "" {
			return errors.New("WithHTMLViews: ext must not be empty")
		}
		s.htmlViews = true
		s.htmlViewsRoot = root
		s.htmlViewsExt = ext
		return nil
	}
}

// WithFiberConfig exposes the assembled fiber.Config for direct modification,
// covering any field GOE does not wrap.
//
// Hooks run last, after every other option and after GOE has translated its own
// settings into Fiber's, so whatever you write here wins. That includes fields
// GOE's own features rely on: clearing StructValidator or disabling
// PassLocalsToContext will break struct validation or http.WithReqCtx
// respectively. GOE logs a warning naming the field and the feature it broke,
// but does not override your choice.
//
//	goehttp.WithFiberConfig(func(c *fiber.Config) {
//	    c.EnableSplittingOnParsers = true
//	})
//
// Multiple hooks are applied in the order given.
func WithFiberConfig(fn func(*fiber.Config)) Option {
	return func(s *settings) error {
		if fn == nil {
			return errors.New("WithFiberConfig: function must not be nil")
		}
		s.fiberHooks = append(s.fiberHooks, fn)
		return nil
	}
}

// WithListenConfig exposes the assembled fiber.ListenConfig for direct
// modification, covering GracefulContext and ShutdownTimeout, which GOE does
// not wrap because its own shutdown manager owns them.
//
// Setting ShutdownTimeout here competes with goe.Options.ShutdownTimeout and
// DrainTimeout; prefer those.
//
// Hooks run last and are applied in the order given.
func WithListenConfig(fn func(*fiber.ListenConfig)) Option {
	return func(s *settings) error {
		if fn == nil {
			return errors.New("WithListenConfig: function must not be nil")
		}
		s.listenHooks = append(s.listenHooks, fn)
		return nil
	}
}

// WithAllowDotSegments lets requests through whose raw path has "." or ".."
// segments, also percent-encoded ones. Replaces HTTP_ALLOW_DOT_SEGMENTS.
//
// GOE answers such requests 400 by default. Browsers never send them, and since
// Fiber v3.6 c.Path() is normalized while c.OriginalURL() stays raw, so code
// that routes or authorizes on one and forwards the other can be steered to a
// different route.
func WithAllowDotSegments(allowed bool) Option {
	return func(s *settings) error { s.allowDotSegments = allowed; return nil }
}

// WithStrictBodyLimit makes BodyLimit hold while request bodies are streamed —
// GOE streams them by default (StreamRequestBody), and then Fiber reads past
// BodyLimit. Replaces HTTP_STRICT_BODY_LIMIT. Off by default.
//
// When on, a declared Content-Length above the limit is answered 413 before any
// handler runs and the connection is closed; a chunked body fails once it
// passes the limit (413 when the handler returns the error); and multipart
// forms are parsed when a handler asks for them rather than up front, since
// parsing up front reads the whole form to temporary files first.
func WithStrictBodyLimit(enabled bool) Option {
	return func(s *settings) error { s.strictBodyLimit = enabled; return nil }
}

// RouteLimits overrides limits for the requests under a path prefix. Zero fields
// keep the server's values.
type RouteLimits struct {
	// BodyLimit is the largest body accepted, in bytes.
	BodyLimit int
	// ReadTimeout bounds reading the request, body included, from its start.
	// Set it for uploads that take longer than the server's ReadTimeout.
	ReadTimeout time.Duration
}

// WithRouteLimits sets limits for the requests whose normalized path is prefix
// or below it (segment by segment: "/files" covers "/files/x", not
// "/filesx"). The longest matching prefix wins. Route limits imply
// WithStrictBodyLimit(true) for every route.
func WithRouteLimits(prefix string, limits RouteLimits) Option {
	return func(s *settings) error {
		if !strings.HasPrefix(prefix, "/") {
			return fmt.Errorf("WithRouteLimits: prefix %q must start with /", prefix)
		}
		if limits.BodyLimit < 0 || limits.ReadTimeout < 0 {
			return fmt.Errorf("WithRouteLimits: limits for %q must not be negative", prefix)
		}
		if len(prefix) > 1 {
			prefix = strings.TrimSuffix(prefix, "/")
		}
		s.routeLimits = append(s.routeLimits, routeLimit{prefix: prefix, limits: limits})
		return nil
	}
}

// WithStoppingSignal gives the kernel a channel that closes when the
// application starts stopping: SSE and SendStream end their streams on it.
// goe.New passes the shutdown manager's Stopping channel; without one, streams
// end when the server shuts down.
func WithStoppingSignal(stopping <-chan struct{}) Option {
	return func(s *settings) error { s.stopping = stopping; return nil }
}

// WithStreamObserver reports the start and end of every stream served by SSE or
// SendStream to observer. GOE's metrics module supplies one.
func WithStreamObserver(observer StreamObserver) Option {
	return func(s *settings) error { s.streamObserver = observer; return nil }
}
