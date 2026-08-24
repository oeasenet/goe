package http

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
)

// This file exposes every field of fiber.Config as an Option, except the four
// listed in fiberConfigSkipList (see options_coverage_test.go), which stay
// reachable through WithFiberConfig. Option names mirror Fiber's field names
// exactly so that Fiber's own documentation is the reference: field X is
// WithX, taking Fiber's own type.

// WithServerHeader sets fiber.Config.ServerHeader, the value sent in the Server
// response header. Replaces FIBER_SERVER_HEADER.
func WithServerHeader(header string) Option {
	return func(s *settings) error { s.fiber.ServerHeader = header; return nil }
}

// WithStrictRouting sets fiber.Config.StrictRouting, which makes "/foo" and
// "/foo/" distinct routes. Replaces FIBER_STRICT_ROUTING.
func WithStrictRouting(enabled bool) Option {
	return func(s *settings) error { s.fiber.StrictRouting = enabled; return nil }
}

// WithCaseSensitive sets fiber.Config.CaseSensitive, which makes "/Foo" and
// "/foo" distinct routes. Replaces FIBER_CASE_SENSITIVE.
func WithCaseSensitive(enabled bool) Option {
	return func(s *settings) error { s.fiber.CaseSensitive = enabled; return nil }
}

// WithSkipUnmatchedRoutes sets fiber.Config.SkipUnmatchedRoutes: requests whose
// path and method match no registered route are answered with 404 (or 405)
// before the middleware chain runs, so no work is spent on bots and scanners.
// Warning: middleware never runs for skipped requests — catch-all Use handlers,
// the access log and metrics will not see them. Replaces
// FIBER_SKIP_UNMATCHED_ROUTES.
func WithSkipUnmatchedRoutes(enabled bool) Option {
	return func(s *settings) error { s.fiber.SkipUnmatchedRoutes = enabled; return nil }
}

// WithDisableHeadAutoRegister sets fiber.Config.DisableHeadAutoRegister, which
// stops Fiber from registering a HEAD route for every GET route.
func WithDisableHeadAutoRegister(disabled bool) Option {
	return func(s *settings) error { s.fiber.DisableHeadAutoRegister = disabled; return nil }
}

// WithImmutable sets fiber.Config.Immutable, which makes values returned by
// context methods valid after the handler returns, at the cost of allocation.
// Replaces FIBER_IMMUTABLE.
func WithImmutable(enabled bool) Option {
	return func(s *settings) error { s.fiber.Immutable = enabled; return nil }
}

// WithUnescapePath sets fiber.Config.UnescapePath, which URL-decodes path
// segments before matching. Replaces FIBER_UNESCAPE_PATH.
func WithUnescapePath(enabled bool) Option {
	return func(s *settings) error { s.fiber.UnescapePath = enabled; return nil }
}

// WithBodyLimit sets fiber.Config.BodyLimit, the maximum request body size in
// bytes. A negative value disables the limit. Replaces FIBER_BODY_LIMIT.
func WithBodyLimit(bytes int) Option {
	return func(s *settings) error { s.fiber.BodyLimit = bytes; return nil }
}

// WithMaxRanges sets fiber.Config.MaxRanges, the maximum number of ranges
// accepted in a Range header.
func WithMaxRanges(maxRanges int) Option {
	return func(s *settings) error {
		if maxRanges < 0 {
			return fmt.Errorf("WithMaxRanges: value must not be negative, got %d", maxRanges)
		}
		s.fiber.MaxRanges = maxRanges
		return nil
	}
}

// WithConcurrency sets fiber.Config.Concurrency, the maximum number of
// concurrent connections. Replaces FIBER_CONCURRENCY.
func WithConcurrency(n int) Option {
	return func(s *settings) error {
		if n <= 0 {
			return fmt.Errorf("WithConcurrency: value must be positive, got %d", n)
		}
		s.fiber.Concurrency = n
		return nil
	}
}

// WithViews sets fiber.Config.Views, the template engine.
//
// Use this to supply any Fiber-compatible engine. For the common case of
// html/template rendering, WithHTMLViews constructs the engine for you.
func WithViews(engine fiber.Views) Option {
	return func(s *settings) error { s.fiber.Views = engine; return nil }
}

// WithViewsLayout sets fiber.Config.ViewsLayout, the global layout template.
// Replaces VIEWS_LAYOUT.
func WithViewsLayout(layout string) Option {
	return func(s *settings) error { s.fiber.ViewsLayout = layout; return nil }
}

// WithPassLocalsToViews sets fiber.Config.PassLocalsToViews.
//
// GOE enables this by default. Disabling it stops handler locals from reaching
// templates.
func WithPassLocalsToViews(enabled bool) Option {
	return func(s *settings) error { s.fiber.PassLocalsToViews = enabled; return nil }
}

// WithPassLocalsToContext sets fiber.Config.PassLocalsToContext.
//
// GOE enables this by default and depends on it: http.WithReqCtx and
// http.GetServices resolve request-scoped values through it. Disabling it
// turns those into no-ops.
func WithPassLocalsToContext(enabled bool) Option {
	return func(s *settings) error { s.fiber.PassLocalsToContext = enabled; return nil }
}

// WithReadTimeout sets fiber.Config.ReadTimeout, the time allowed to read the
// full request. Replaces HTTP_READ_TIMEOUT.
func WithReadTimeout(d time.Duration) Option {
	return func(s *settings) error {
		if d < 0 {
			return fmt.Errorf("WithReadTimeout: value must not be negative, got %s", d)
		}
		s.fiber.ReadTimeout = d
		return nil
	}
}

// WithWriteTimeout sets fiber.Config.WriteTimeout, the time allowed to write
// the full response. Replaces HTTP_WRITE_TIMEOUT.
func WithWriteTimeout(d time.Duration) Option {
	return func(s *settings) error {
		if d < 0 {
			return fmt.Errorf("WithWriteTimeout: value must not be negative, got %s", d)
		}
		s.fiber.WriteTimeout = d
		return nil
	}
}

// WithIdleTimeout sets fiber.Config.IdleTimeout, how long to wait for the next
// request on a keep-alive connection. Replaces HTTP_IDLE_TIMEOUT.
func WithIdleTimeout(d time.Duration) Option {
	return func(s *settings) error {
		if d < 0 {
			return fmt.Errorf("WithIdleTimeout: value must not be negative, got %s", d)
		}
		s.fiber.IdleTimeout = d
		return nil
	}
}

// WithReadBufferSize sets fiber.Config.ReadBufferSize, the per-connection read
// buffer. Increase it for requests with very large headers.
func WithReadBufferSize(bytes int) Option {
	return func(s *settings) error {
		if bytes <= 0 {
			return fmt.Errorf("WithReadBufferSize: value must be positive, got %d", bytes)
		}
		s.fiber.ReadBufferSize = bytes
		return nil
	}
}

// WithWriteBufferSize sets fiber.Config.WriteBufferSize, the per-connection
// write buffer.
func WithWriteBufferSize(bytes int) Option {
	return func(s *settings) error {
		if bytes <= 0 {
			return fmt.Errorf("WithWriteBufferSize: value must be positive, got %d", bytes)
		}
		s.fiber.WriteBufferSize = bytes
		return nil
	}
}

// WithCompressedFileSuffixes sets fiber.Config.CompressedFileSuffixes, mapping
// content encodings to the file suffix used for cached compressed responses.
func WithCompressedFileSuffixes(suffixes map[string]string) Option {
	return func(s *settings) error { s.fiber.CompressedFileSuffixes = suffixes; return nil }
}

// WithProxyHeader sets fiber.Config.ProxyHeader, the header Ctx.IP reads the
// client address from. Defaults to X-Forwarded-For, which stays inert until
// trusted proxies are enabled. Replaces FIBER_PROXY_HEADER.
func WithProxyHeader(header string) Option {
	return func(s *settings) error { s.fiber.ProxyHeader = header; return nil }
}

// WithGETOnly sets fiber.Config.GETOnly, which rejects every method other than
// GET.
func WithGETOnly(enabled bool) Option {
	return func(s *settings) error { s.fiber.GETOnly = enabled; return nil }
}

// WithErrorHandler sets fiber.Config.ErrorHandler, replacing GOE's handler.
//
// GOE's default renders an HTML error page, or JSON/plain text based on the
// Accept header and a "format" query parameter. Replacing it is fully
// supported; this is the way to do it.
func WithErrorHandler(handler fiber.ErrorHandler) Option {
	return func(s *settings) error {
		if handler == nil {
			return errors.New("WithErrorHandler: handler must not be nil")
		}
		s.fiber.ErrorHandler = handler
		return nil
	}
}

// WithDisableKeepalive sets fiber.Config.DisableKeepalive, closing every
// connection after its response.
func WithDisableKeepalive(disabled bool) Option {
	return func(s *settings) error { s.fiber.DisableKeepalive = disabled; return nil }
}

// WithDisableDefaultDate sets fiber.Config.DisableDefaultDate, omitting the
// Date response header.
func WithDisableDefaultDate(disabled bool) Option {
	return func(s *settings) error { s.fiber.DisableDefaultDate = disabled; return nil }
}

// WithDisableDefaultContentType sets fiber.Config.DisableDefaultContentType,
// omitting the default Content-Type response header.
func WithDisableDefaultContentType(disabled bool) Option {
	return func(s *settings) error { s.fiber.DisableDefaultContentType = disabled; return nil }
}

// WithDisableHeaderNormalizing sets fiber.Config.DisableHeaderNormalizing,
// leaving header names exactly as received rather than canonicalising them.
func WithDisableHeaderNormalizing(disabled bool) Option {
	return func(s *settings) error { s.fiber.DisableHeaderNormalizing = disabled; return nil }
}

// WithAppName sets fiber.Config.AppName, shown in the startup banner. GOE
// defaults it to APP_NAME.
func WithAppName(name string) Option {
	return func(s *settings) error { s.fiber.AppName = name; return nil }
}

// WithSharedStorage sets fiber.Config.SharedStorage, the storage backend Fiber
// middleware shares state through.
func WithSharedStorage(storage fiber.Storage) Option {
	return func(s *settings) error { s.fiber.SharedStorage = storage; return nil }
}

// WithSharedStatePrefix sets fiber.Config.SharedStatePrefix, the key prefix
// used in the shared storage backend.
func WithSharedStatePrefix(prefix string) Option {
	return func(s *settings) error { s.fiber.SharedStatePrefix = prefix; return nil }
}

// WithStreamRequestBody sets fiber.Config.StreamRequestBody, which hands the
// body to the handler as a stream instead of buffering it. GOE enables this by
// default. Replaces FIBER_STREAM_REQUEST_BODY.
func WithStreamRequestBody(enabled bool) Option {
	return func(s *settings) error { s.fiber.StreamRequestBody = enabled; return nil }
}

// WithDisablePreParseMultipartForm sets
// fiber.Config.DisablePreParseMultipartForm, deferring multipart parsing to the
// handler.
func WithDisablePreParseMultipartForm(disabled bool) Option {
	return func(s *settings) error { s.fiber.DisablePreParseMultipartForm = disabled; return nil }
}

// WithReduceMemoryUsage sets fiber.Config.ReduceMemoryUsage, trading CPU for a
// smaller memory footprint. Replaces FIBER_REDUCE_MEMORY.
func WithReduceMemoryUsage(enabled bool) Option {
	return func(s *settings) error { s.fiber.ReduceMemoryUsage = enabled; return nil }
}

// WithJSONEncoder sets fiber.Config.JSONEncoder. GOE defaults to sonic.Marshal.
func WithJSONEncoder(encoder utils.JSONMarshal) Option {
	return func(s *settings) error {
		if encoder == nil {
			return errors.New("WithJSONEncoder: encoder must not be nil")
		}
		s.fiber.JSONEncoder = encoder
		return nil
	}
}

// WithJSONDecoder sets fiber.Config.JSONDecoder. GOE defaults to
// sonic.Unmarshal.
func WithJSONDecoder(decoder utils.JSONUnmarshal) Option {
	return func(s *settings) error {
		if decoder == nil {
			return errors.New("WithJSONDecoder: decoder must not be nil")
		}
		s.fiber.JSONDecoder = decoder
		return nil
	}
}

// WithMsgPackEncoder sets fiber.Config.MsgPackEncoder.
func WithMsgPackEncoder(encoder utils.MsgPackMarshal) Option {
	return func(s *settings) error {
		if encoder == nil {
			return errors.New("WithMsgPackEncoder: encoder must not be nil")
		}
		s.fiber.MsgPackEncoder = encoder
		return nil
	}
}

// WithMsgPackDecoder sets fiber.Config.MsgPackDecoder.
func WithMsgPackDecoder(decoder utils.MsgPackUnmarshal) Option {
	return func(s *settings) error {
		if decoder == nil {
			return errors.New("WithMsgPackDecoder: decoder must not be nil")
		}
		s.fiber.MsgPackDecoder = decoder
		return nil
	}
}

// WithCBOREncoder sets fiber.Config.CBOREncoder.
func WithCBOREncoder(encoder utils.CBORMarshal) Option {
	return func(s *settings) error {
		if encoder == nil {
			return errors.New("WithCBOREncoder: encoder must not be nil")
		}
		s.fiber.CBOREncoder = encoder
		return nil
	}
}

// WithCBORDecoder sets fiber.Config.CBORDecoder.
func WithCBORDecoder(decoder utils.CBORUnmarshal) Option {
	return func(s *settings) error {
		if decoder == nil {
			return errors.New("WithCBORDecoder: decoder must not be nil")
		}
		s.fiber.CBORDecoder = decoder
		return nil
	}
}

// WithXMLEncoder sets fiber.Config.XMLEncoder. GOE defaults to xml.Marshal.
func WithXMLEncoder(encoder utils.XMLMarshal) Option {
	return func(s *settings) error {
		if encoder == nil {
			return errors.New("WithXMLEncoder: encoder must not be nil")
		}
		s.fiber.XMLEncoder = encoder
		return nil
	}
}

// WithXMLDecoder sets fiber.Config.XMLDecoder.
func WithXMLDecoder(decoder utils.XMLUnmarshal) Option {
	return func(s *settings) error {
		if decoder == nil {
			return errors.New("WithXMLDecoder: decoder must not be nil")
		}
		s.fiber.XMLDecoder = decoder
		return nil
	}
}

// WithTrustProxy sets fiber.Config.TrustProxy, which must be enabled for
// WithTrustProxyConfig to take effect. Replaces FIBER_TRUST_PROXY.
func WithTrustProxy(enabled bool) Option {
	return func(s *settings) error { s.fiber.TrustProxy = enabled; return nil }
}

// WithTrustProxyConfig sets fiber.Config.TrustProxyConfig, describing which
// upstream addresses may be trusted for X-Forwarded-* headers. Replaces the
// FIBER_TRUST_PROXIES, FIBER_TRUST_LINK_LOCAL, FIBER_TRUST_LOOPBACK and
// FIBER_TRUST_PRIVATE variables.
//
// Fiber ignores this unless TrustProxy is enabled, so GOE rejects the
// combination at startup rather than letting it fail silently.
func WithTrustProxyConfig(cfg fiber.TrustProxyConfig) Option {
	return func(s *settings) error { s.fiber.TrustProxyConfig = cfg; return nil }
}

// WithEnableIPValidation sets fiber.Config.EnableIPValidation, which validates
// addresses parsed out of proxy headers and walks X-Forwarded-For past trusted
// hops rather than returning it raw. On by default in GOE. Replaces
// FIBER_ENABLE_IP_VALIDATION.
func WithEnableIPValidation(enabled bool) Option {
	return func(s *settings) error { s.fiber.EnableIPValidation = enabled; return nil }
}

// WithColorScheme sets fiber.Config.ColorScheme used by the startup banner.
func WithColorScheme(colors fiber.Colors) Option {
	return func(s *settings) error { s.fiber.ColorScheme = colors; return nil }
}

// WithStructValidator sets fiber.Config.StructValidator, replacing GOE's
// validator.
//
// GOE installs its own validator so that Ctx.Bind validates structs using the
// tags understood by the goe validation package. Replacing it is supported;
// this is the way to do it.
func WithStructValidator(v fiber.StructValidator) Option {
	return func(s *settings) error {
		if v == nil {
			return errors.New("WithStructValidator: validator must not be nil")
		}
		s.fiber.StructValidator = v
		return nil
	}
}

// WithRequestMethods sets fiber.Config.RequestMethods, the set of HTTP methods
// the router accepts.
//
// Fiber's default (fiber.DefaultMethods) includes QUERY as of v3.4. Supplying a
// narrower list silently makes the omitted methods unroutable, so prefer
// appending to fiber.DefaultMethods rather than writing a list from scratch.
func WithRequestMethods(methods []string) Option {
	return func(s *settings) error {
		if len(methods) == 0 {
			return errors.New("WithRequestMethods: at least one method is required")
		}
		s.fiber.RequestMethods = methods
		return nil
	}
}

// WithEnableSplittingOnParsers sets fiber.Config.EnableSplittingOnParsers,
// which splits comma-separated query and form values into slices.
func WithEnableSplittingOnParsers(enabled bool) Option {
	return func(s *settings) error { s.fiber.EnableSplittingOnParsers = enabled; return nil }
}
