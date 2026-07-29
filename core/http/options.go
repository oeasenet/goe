package http

import (
	"encoding/xml"
	"errors"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	htmltpl "github.com/gofiber/template/html/v3"
	"go.oease.dev/goe/v2/contract"
)

// Option configures the HTTP kernel from Go code.
//
// Options are applied after environment variables, so a value set in code
// always wins over the matching FIBER_*/HTTP_*/VIEWS_* variable. Environment
// variables still supply everything code does not set, which means an
// application that passes no options behaves exactly as it did before.
//
// The naming rule is mechanical: every field of fiber.Config and
// fiber.ListenConfig is exposed as With<FieldName> taking Fiber's own type.
// Read Fiber's documentation, prepend "With", and that is the option. GOE only
// invents names where Fiber has no equivalent field: WithHost, WithPort,
// WithRequestID, WithRequestIDHeader and WithHTMLViews.
//
// Options are applied in the order given, so the last write wins:
//
//	goehttp.WithBodyLimit(1 << 20),
//	goehttp.WithBodyLimit(2 << 20), // 2MB
//
// If any option returns an error, none of them are applied and the error is
// reported by Module.ValidateConfig, which fails application startup. A
// partially applied configuration is never used.
type Option func(*settings) error

// settings is the mutable state options write to. It is deliberately
// unexported: options are the only way in, which is what keeps "unset" and
// "set to the zero value" distinguishable. A caller who could build this struct
// literally would reintroduce that ambiguity.
type settings struct {
	fiber  fiber.Config
	listen fiber.ListenConfig

	// GOE-level settings that have no fiber.Config equivalent.
	host string
	port int

	// htmlViews requests that GOE construct a html/template view engine. It is
	// only materialised when no explicit fiber.Views engine has been supplied.
	htmlViews     bool
	htmlViewsRoot string
	htmlViewsExt  string

	requestIDEnabled bool
	requestIDHeader  string

	// Escape hatches, always applied last so that code always wins.
	fiberHooks  []func(*fiber.Config)
	listenHooks []func(*fiber.ListenConfig)
}

// defaultSettings returns GOE's opinionated baseline, the bottom layer of the
// resolution pipeline (defaults -> env -> options -> escape hatches).
func defaultSettings(validator fiber.StructValidator, errorHandler fiber.ErrorHandler) settings {
	return settings{
		fiber: fiber.Config{
			ServerHeader:        "Goe",
			BodyLimit:           4 * 1024 * 1024,
			Concurrency:         256 * 1024,
			StreamRequestBody:   true,
			ReadTimeout:         10 * time.Second,
			WriteTimeout:        10 * time.Second,
			IdleTimeout:         30 * time.Second,
			JSONEncoder:         sonic.Marshal,
			JSONDecoder:         sonic.Unmarshal,
			XMLEncoder:          xml.Marshal,
			ColorScheme:         fiber.DefaultColors,
			StructValidator:     validator,
			ErrorHandler:        errorHandler,
			PassLocalsToContext: true,
			PassLocalsToViews:   true,
		},
		host:             "0.0.0.0",
		port:             8080,
		requestIDEnabled: true,
		requestIDHeader:  fiber.HeaderXRequestID,
	}
}

// applyEnv overlays the FIBER_*, HTTP_* and VIEWS_* variables onto s. Only keys
// that are actually present override a default, so an absent variable leaves
// GOE's baseline intact.
func (s *settings) applyEnv(config contract.Config) {
	setString := func(key string, dst *string) {
		if v := config.GetString(key); v != "" {
			*dst = v
		}
	}
	setInt := func(key string, dst *int) {
		if v := config.GetInt(key); v != 0 {
			*dst = v
		}
	}
	setBool := func(key string, dst *bool) {
		if config.Has(key) {
			*dst = config.GetBool(key)
		}
	}
	setDuration := func(key string, dst *time.Duration) {
		if v := config.GetDuration(key); v != 0 {
			*dst = v
		}
	}

	// Server identity and limits.
	setString("FIBER_SERVER_HEADER", &s.fiber.ServerHeader)
	setInt("FIBER_BODY_LIMIT", &s.fiber.BodyLimit)
	setInt("FIBER_CONCURRENCY", &s.fiber.Concurrency)
	setString("FIBER_PROXY_HEADER", &s.fiber.ProxyHeader)
	if appName := config.GetString("APP_NAME"); appName != "" {
		s.fiber.AppName = appName
	}

	// Routing and request handling behaviour.
	setBool("FIBER_STRICT_ROUTING", &s.fiber.StrictRouting)
	setBool("FIBER_CASE_SENSITIVE", &s.fiber.CaseSensitive)
	setBool("FIBER_IMMUTABLE", &s.fiber.Immutable)
	setBool("FIBER_UNESCAPE_PATH", &s.fiber.UnescapePath)
	setBool("FIBER_STREAM_REQUEST_BODY", &s.fiber.StreamRequestBody)
	setBool("FIBER_REDUCE_MEMORY", &s.fiber.ReduceMemoryUsage)
	setBool("FIBER_ENABLE_IP_VALIDATION", &s.fiber.EnableIPValidation)

	// Timeouts.
	setDuration("HTTP_READ_TIMEOUT", &s.fiber.ReadTimeout)
	setDuration("HTTP_WRITE_TIMEOUT", &s.fiber.WriteTimeout)
	setDuration("HTTP_IDLE_TIMEOUT", &s.fiber.IdleTimeout)

	// Listener address.
	setString("HTTP_HOST", &s.host)
	setInt("HTTP_PORT", &s.port)

	// Request ID middleware.
	if config.Has("HTTP_REQUEST_ID") {
		s.requestIDEnabled = config.GetBool("HTTP_REQUEST_ID")
	}
	setString("HTTP_REQUEST_ID_HEADER", &s.requestIDHeader)

	// Trusted proxies.
	if config.GetBool("FIBER_TRUST_PROXY") {
		s.fiber.TrustProxy = true

		if proxies := config.GetStringSlice("FIBER_TRUST_PROXIES"); len(proxies) > 0 {
			// Historically these three default to true once a proxy list is
			// supplied, unless explicitly overridden.
			trustCfg := fiber.TrustProxyConfig{
				Proxies:   proxies,
				LinkLocal: true,
				Loopback:  true,
				Private:   true,
			}
			if config.Has("FIBER_TRUST_LINK_LOCAL") {
				trustCfg.LinkLocal = config.GetBool("FIBER_TRUST_LINK_LOCAL")
			}
			if config.Has("FIBER_TRUST_LOOPBACK") {
				trustCfg.Loopback = config.GetBool("FIBER_TRUST_LOOPBACK")
			}
			if config.Has("FIBER_TRUST_PRIVATE") {
				trustCfg.Private = config.GetBool("FIBER_TRUST_PRIVATE")
			}
			s.fiber.TrustProxyConfig = trustCfg
		}
	}

	// View engine.
	if engine := config.GetString("VIEWS_ENGINE"); engine == "html" {
		s.htmlViews = true
		s.htmlViewsRoot = "./views"
		s.htmlViewsExt = ".gohtml"
		setString("VIEWS_ROOT", &s.htmlViewsRoot)
		setString("VIEWS_EXT", &s.htmlViewsExt)
		setString("VIEWS_LAYOUT", &s.fiber.ViewsLayout)
	}
}

// applyOptions runs opts against s, collecting every error rather than stopping
// at the first, so a developer sees all of their mistakes in one startup
// failure instead of one per run.
func applyOptions(s *settings, opts []Option) []error {
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(s); err != nil {
			errs = append(errs, fmt.Errorf("http option %d: %w", i+1, err))
		}
	}
	return errs
}

// validate performs the cross-field checks that cannot be expressed by a single
// option. These catch configurations Fiber would otherwise accept and then
// silently ignore, which is precisely the class of bug that is hardest to find
// in production.
func (s *settings) validate() []error {
	var errs []error

	if s.port < 0 || s.port > 65535 {
		errs = append(errs, fmt.Errorf("http port %d is out of range 0-65535", s.port))
	}

	// Fiber ignores TrustProxyConfig entirely unless TrustProxy is true.
	if !s.fiber.TrustProxy && trustProxyConfigured(s.fiber.TrustProxyConfig) {
		errs = append(errs, errors.New(
			"WithTrustProxyConfig has no effect unless trusted proxies are enabled; add WithTrustProxy(true)"))
	}

	// A half-configured certificate pair fails at listen time with a less
	// obvious message, so reject it up front.
	switch {
	case s.listen.CertFile != "" && s.listen.CertKeyFile == "":
		errs = append(errs, errors.New("WithCertFile requires WithCertKeyFile"))
	case s.listen.CertKeyFile != "" && s.listen.CertFile == "":
		errs = append(errs, errors.New("WithCertKeyFile requires WithCertFile"))
	}

	// Fiber returns ErrAutoCertWithCertFile for this combination.
	if s.listen.AutoCertManager != nil && (s.listen.CertFile != "" || s.listen.CertKeyFile != "") {
		errs = append(errs, errors.New(
			"WithAutoCertManager cannot be combined with WithCertFile/WithCertKeyFile"))
	}

	return errs
}

// trustProxyConfigured reports whether any trusted-proxy setting was supplied.
func trustProxyConfigured(c fiber.TrustProxyConfig) bool {
	return len(c.Proxies) > 0 || c.LinkLocal || c.Loopback || c.Private || c.UnixSocket
}

// materialise translates GOE-level settings into their fiber.Config equivalents.
// It runs before the escape hatches so that a raw hook can still override
// anything decided here.
func (s *settings) materialise(logger contract.Logger) {
	// Build the html/template engine only when the developer has not supplied
	// an engine of their own through WithViews.
	if s.htmlViews && s.fiber.Views == nil {
		root := s.htmlViewsRoot
		if root == "" {
			root = "./views"
		}
		ext := s.htmlViewsExt
		if ext == "" {
			ext = ".gohtml"
		}
		s.fiber.Views = htmltpl.New(root, ext)
		logger.Debug("HTTP views engine initialized", "engine", "html", "root", root, "ext", ext)
	}
}

// applyHooks runs the raw escape hatches. They are applied last, after
// materialisation, which is what makes "code always wins" true: were they
// applied at their position in the option list, materialisation could silently
// overwrite a field the developer had just set.
func (s *settings) applyHooks() {
	for _, hook := range s.fiberHooks {
		if hook != nil {
			hook(&s.fiber)
		}
	}
	for _, hook := range s.listenHooks {
		if hook != nil {
			hook(&s.listen)
		}
	}
}

// loadBearing captures the fiber.Config fields GOE's own features depend on, so
// that clearing one through an escape hatch can be reported.
type loadBearing struct {
	passLocalsToContext bool
	passLocalsToViews   bool
	structValidator     fiber.StructValidator
	errorHandler        fiber.ErrorHandler
}

func snapshotLoadBearing(cfg fiber.Config) loadBearing {
	return loadBearing{
		passLocalsToContext: cfg.PassLocalsToContext,
		passLocalsToViews:   cfg.PassLocalsToViews,
		structValidator:     cfg.StructValidator,
		errorHandler:        cfg.ErrorHandler,
	}
}

// warnOnClobber reports fields an escape hatch disabled that GOE relies on. It
// never restores them: WithFiberConfig is documented as always winning. The
// warning exists so that the resulting breakage reads as a deliberate choice
// rather than a GOE bug.
func warnOnClobber(before loadBearing, cfg fiber.Config, logger contract.Logger) {
	if before.passLocalsToContext && !cfg.PassLocalsToContext {
		logger.Warn("PassLocalsToContext was disabled via WithFiberConfig; " +
			"http.WithReqCtx and http.GetServices will no longer resolve request-scoped values")
	}
	if before.passLocalsToViews && !cfg.PassLocalsToViews {
		logger.Warn("PassLocalsToViews was disabled via WithFiberConfig; " +
			"handler locals will no longer be visible to templates")
	}
	if before.structValidator != nil && cfg.StructValidator == nil {
		logger.Warn("StructValidator was cleared via WithFiberConfig; " +
			"struct validation on Bind will no longer run. Use WithStructValidator to replace it instead")
	}
	if before.errorHandler != nil && cfg.ErrorHandler == nil {
		logger.Warn("ErrorHandler was cleared via WithFiberConfig; " +
			"GOE error pages are disabled and Fiber's default handler will be used. " +
			"Use WithErrorHandler to replace it instead")
	}
}

// resolve runs the full pipeline: defaults -> env -> options -> validation ->
// materialisation -> escape hatches -> clobber scan.
//
// When an option or cross-field check fails, every option is discarded and the
// environment-only configuration is used instead. The returned errors are
// surfaced by Module.ValidateConfig, which aborts startup, so the fallback
// configuration is never actually served.
func resolve(logger contract.Logger, base settings, opts []Option) (settings, []error) {
	s := base
	errs := applyOptions(&s, opts)
	if len(errs) == 0 {
		errs = s.validate()
	}
	if len(errs) > 0 {
		s = base
		s.fiberHooks = nil
		s.listenHooks = nil
	}

	s.materialise(logger)

	before := snapshotLoadBearing(s.fiber)
	s.applyHooks()
	warnOnClobber(before, s.fiber, logger)

	return s, errs
}
