package http

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net"
	"reflect"
	"slices"
	"strconv"
	"sync"

	fiberzap "github.com/gofiber/contrib/v3/zap"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/utils/v2"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/internal/configvalidator"
	"go.oease.dev/goe/v2/validation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// kernel implements the HTTPKernel interface
type kernel struct {
	app       *fiber.App
	config    contract.Config
	logger    contract.Logger
	validator *validation.Validator

	// listenCfg, host and port are resolved once at construction from
	// defaults, environment and Options, and drive both Listen and Module.OnStart.
	listenCfg fiber.ListenConfig
	host      string
	port      int

	// optErrs holds failures from Option application. They are reported by
	// ValidateConfig so that startup aborts; the kernel itself was built from
	// the environment-only configuration and is never served.
	optErrs []error
}

// New creates a new HTTP kernel.
//
// Configuration resolves in layers: GOE defaults, then FIBER_*/HTTP_*/VIEWS_*
// environment variables, then opts. Anything set through an Option wins over
// the environment; the environment still supplies everything code leaves
// alone, so calling New without options behaves exactly as before options
// existed.
func New(config contract.Config, logger contract.Logger, opts ...Option) contract.HTTPKernel {
	// Tag loggers once: "http" for kernel/server logs, "access" for the
	// per-request access log so it can be controlled independently
	// (e.g. LOG_MODULE_LEVELS=access:warn shows only 4xx/5xx requests).
	klog := logger.With("module", "http")
	accessLogger := logger.With("module", "access")

	// Create validator
	validator := validation.New()

	// Init error page template
	var err error
	tpl, err = template.ParseFS(templateFS, "error_page.gohtml")
	if err != nil {
		klog.Fatal(err.Error())
		panic(err)
	}

	// Layer 1 and 2: GOE defaults, then the environment.
	base := defaultSettings(validator, defaultErrorHandler(klog))
	base.applyEnv(config)

	if engine := config.GetString("VIEWS_ENGINE"); engine != "" && engine != "html" {
		klog.Warn("Unsupported VIEWS_ENGINE specified; views not initialized", "engine", engine)
	}

	// Layer 3 onwards: options, cross-field validation, materialisation and
	// the raw escape hatches. Any option error discards every option and
	// leaves the environment-only configuration in place; ValidateConfig then
	// aborts startup before the server can serve a request.
	resolved, optErrs := resolve(klog, base, opts)

	// Create fiber app
	app := fiber.New(resolved.fiber)

	// Register default route constraints (uuid, uint, slug, email)
	RegisterDefaultConstraints(app)

	// Add default middleware
	app.Use(recover.New())

	// Request ID middleware: on by default (set HTTP_REQUEST_ID=false or use
	// WithRequestID(false) to disable and own your own middleware stack). It
	// reuses an upstream request id from the configured header and generates
	// one when absent.
	reqIDHeader := resolved.requestIDHeader
	if resolved.requestIDEnabled {
		app.Use(requestid.New(requestid.Config{Header: reqIDHeader}))
	}

	//Add request logging middleware using fiber's official middleware
	// fiberzap maps by status: index 0 = >=500, 1 = >=400, 2 = other.
	// So success requests log at Info, 4xx at Warn, 5xx at Error — and the
	// "access" module tag lets operators silence success with access:warn.
	fiberZap := fiberzap.New(fiberzap.Config{
		SkipURIs: []string{
			"/.well-known/liveness",
			"/.well-known/readiness",
			"/.well-known/health",
		},
		Logger: accessLogger.GetLogger().Desugar(),
		Fields: []string{"method", "status", "latency", "ip", "url"},
		FieldsFunc: func(c fiber.Ctx) []zap.Field {
			fields := []zap.Field{zap.String("request_id", requestIDFrom(c, reqIDHeader))}
			if sc := trace.SpanContextFromContext(c.Context()); sc.IsValid() {
				fields = append(fields, zap.String("trace_id", sc.TraceID().String()))
			}
			return fields
		},
		Messages: []string{"http server error", "http client error", "http request"},
		Levels:   []zapcore.Level{zapcore.ErrorLevel, zapcore.WarnLevel, zapcore.InfoLevel},
	})

	app.Use(fiberZap)

	return &kernel{
		app:       app,
		config:    config,
		logger:    klog,
		validator: validator,
		listenCfg: resolved.listen,
		host:      resolved.host,
		port:      resolved.port,
		optErrs:   optErrs,
	}
}

// addr returns the address the server binds to. For a unix listener network
// the host carries the socket path and no port is appended.
func (k *kernel) addr() string {
	if k.listenCfg.ListenerNetwork == fiber.NetworkUnix {
		return k.host
	}
	return net.JoinHostPort(k.host, strconv.Itoa(k.port))
}

// App returns the underlying Fiber app
func (k *kernel) App() *fiber.App {
	return k.app
}

// Validator returns the struct validator
func (k *kernel) Validator() any {
	return k.validator
}

// HTTPValidator returns the validator as an HTTPValidator interface
func (k *kernel) HTTPValidator() contract.HTTPValidator {
	return k.validator
}

// Listen starts the HTTP server, blocking until it stops.
//
// An empty addr uses the resolved host and port. The resolved fiber.ListenConfig
// is applied either way, so TLS, prefork and unix-socket options configured in
// code take effect here too.
func (k *kernel) Listen(addr string) error {
	if addr == "" {
		addr = k.addr()
	}

	k.logger.Info("HTTP server starting",
		"address", addr,
	)

	return k.app.Listen(addr, k.listenCfg)
}

// Shutdown gracefully shuts down the server
func (k *kernel) Shutdown() error {
	k.logger.Info("HTTP server shutting down")
	return k.app.Shutdown()
}

// isNilError reports whether err is nil or holds a typed-nil value. A handler
// that returns a nil pointer of a concrete error type produces a non-nil error
// interface whose Error() panics when it dereferences its receiver.
func isNilError(err error) bool {
	if err == nil {
		return true
	}
	switch v := reflect.ValueOf(err); v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

// defaultErrorHandler creates a default error handler
func defaultErrorHandler(logger contract.Logger) fiber.ErrorHandler {
	return func(ctx fiber.Ctx, err error) error {
		// Fiber v3.4 hardened its DefaultErrorHandler against typed-nil errors
		// (gofiber/fiber#4407, #4372). GOE replaces that handler, so it has to
		// carry the same guard: without it a typed-nil error panics here, and
		// this handler runs outside the recover middleware on Fiber's own path.
		if isNilError(err) {
			err = nil
		}

		// Status code defaults to 500
		respCode := fiber.StatusInternalServerError
		// Set error message
		message := utils.StatusMessage(respCode)
		// Check if it's a fiber.Error type. errors.As can match a wrapped
		// typed-nil *fiber.Error, so e itself must be checked as well.
		var e *fiber.Error
		switch matched := errors.As(err, &e); {
		case matched && e != nil:
			respCode = e.Code
			message = e.Message
		case err != nil && !matched:
			message = err.Error()
		}
		ctx.Status(respCode)

		// Log error for 5xx errors, removing this because fiber already has the similar thing
		//if respCode >= 500 {
		//	logger.Error("HTTP Error",
		//		"error", err.Error(),
		//		"path", ctx.Path(),
		//		"method", ctx.Method(),
		//		"status", respCode,
		//	)
		//}

		// If the format is forced to json or text through query parameter, then return the response in that format
		if ctx.Query("format") == "json" {
			return ctx.JSON(fiber.Map{
				"message": message,
			})
		}

		// If the format is forced to text through query parameter, then return the response in that format
		if ctx.Query("format") == "text" {
			ctx.Response().Header.SetContentType(fiber.MIMETextPlain)
			return ctx.SendString(message)
		}

		if ctx.Accepts(fiber.MIMETextHTML) == fiber.MIMETextHTML {
			// default response, html error page
			ctx.Response().Header.SetContentType(fiber.MIMETextHTMLCharsetUTF8)
			return ErrorPage(ctx, fmt.Sprintf("ERROR %d", respCode), fmt.Sprintf("%d", respCode), message, "/")
		}

		// If the format is not forced, then check the accept header
		if ctx.Accepts(fiber.MIMEApplicationJSON) == fiber.MIMEApplicationJSON {
			return ctx.JSON(fiber.Map{
				"message": message,
			})
		}

		if ctx.Accepts(fiber.MIMETextPlain) == fiber.MIMETextPlain {
			ctx.Response().Header.SetContentType(fiber.MIMETextPlain)
			return ctx.SendString(message)
		}

		// default response, html error page
		ctx.Response().Header.SetContentType(fiber.MIMETextHTMLCharsetUTF8)
		return ErrorPage(ctx, fmt.Sprintf("ERROR %d", respCode), fmt.Sprintf("%d", respCode), message, "/")
	}
}

// getValidatorFromKernel safely extracts the validator from the kernel
func getValidatorFromKernel(kernel contract.HTTPKernel) *validation.Validator {
	if v, ok := kernel.Validator().(*validation.Validator); ok {
		return v
	}
	return nil
}

// Module represents the HTTP module for Fx
type Module struct {
	kernel    contract.HTTPKernel
	validator *validation.Validator
}

// NewModule creates a new HTTP module. See New for how opts resolve against
// GOE defaults and the environment.
func NewModule(config contract.Config, logger contract.Logger, opts ...Option) *Module {
	kernel := New(config, logger, opts...)
	return &Module{
		kernel:    kernel,
		validator: getValidatorFromKernel(kernel),
	}
}

// Name returns the module name
func (m *Module) Name() string {
	return "http"
}

// OnStart starts the HTTP server and returns once it is accepting connections.
//
// The server runs through app.Listen rather than a pre-bound net.Listener so
// that fiber.ListenConfig applies — that is what makes TLS, prefork and unix
// sockets work. Binding therefore happens on the serving goroutine, so
// readiness is awaited explicitly to keep the guarantee callers already rely
// on: when OnStart returns nil the port is bound, and a bind failure such as
// EADDRINUSE is returned rather than logged and swallowed.
//
// Two signals are needed because Fiber reports readiness differently per mode:
// OnListen hooks fire in the normal path and in the prefork master, while
// prefork children only get ListenerAddrFunc. Whichever arrives first wins.
func (m *Module) OnStart(ctx context.Context) error {
	k := m.kernel.(*kernel)
	addr := k.addr()

	ready := make(chan struct{})
	var once sync.Once
	signalReady := func() { once.Do(func() { close(ready) }) }

	// Copy so the kernel's own config keeps the developer's callback intact.
	listenCfg := k.listenCfg
	userAddrFunc := listenCfg.ListenerAddrFunc
	listenCfg.ListenerAddrFunc = func(a net.Addr) {
		if userAddrFunc != nil {
			userAddrFunc(a)
		}
		signalReady()
	}
	k.app.Hooks().OnListen(func(fiber.ListenData) error {
		signalReady()
		return nil
	})

	// Buffered so the goroutine never blocks once OnStart has returned.
	errc := make(chan error, 1)
	go func() {
		errc <- k.app.Listen(addr, listenCfg)
	}()

	select {
	case <-ready:
		k.logger.Info("HTTP server started", "address", addr)
		return nil
	case err := <-errc:
		if err == nil {
			return fmt.Errorf("HTTP server on %s stopped before it started serving", addr)
		}
		return fmt.Errorf("failed to start HTTP server on %s: %w", addr, err)
	case <-ctx.Done():
		return fmt.Errorf("timed out starting HTTP server on %s: %w", addr, ctx.Err())
	}
}

// OnStop is called when the module stops
func (m *Module) OnStop(ctx context.Context) error {
	return m.kernel.Shutdown()
}

// Provide returns the HTTP kernel instance for Fx
func (m *Module) Provide() contract.HTTPKernel {
	return m.kernel
}

// ProvideValidator returns the validator instance
func (m *Module) ProvideValidator() *validation.Validator {
	return m.validator
}

// SetupServiceMiddleware sets up the service injection middleware
func (m *Module) SetupServiceMiddleware(app contract.Application, config contract.Config, logger contract.Logger) {
	services := Services{
		App:       app,
		Config:    config,
		Logger:    logger,
		Validator: m.validator,
	}
	m.kernel.App().Use(InjectServices(services))
}

// ValidateConfig validates the HTTP module configuration
func (m *Module) ValidateConfig() error {
	// Get the kernel's config
	k := m.kernel.(*kernel)

	// Option failures are reported first and abort startup. The kernel fell
	// back to the environment-only configuration when this happened, so
	// returning here guarantees a half-configured server is never served.
	if len(k.optErrs) > 0 {
		return errors.Join(k.optErrs...)
	}

	v := configvalidator.NewConfigValidator(k.config, "http")

	// HTTP port is optional but should be valid if set
	if k.config.Has("HTTP_PORT") {
		v.Optional("HTTP_PORT", "HTTP server port", configvalidator.ValidatePort)
	}

	// Validate Fiber-specific configurations if set
	if k.config.Has("FIBER_BODY_LIMIT") {
		v.Optional("FIBER_BODY_LIMIT", "Request body size limit", configvalidator.ValidatePositiveInt)
	}

	if k.config.Has("FIBER_CONCURRENCY") {
		v.Optional("FIBER_CONCURRENCY", "Maximum concurrent connections", configvalidator.ValidatePositiveInt)
	}

	// Validate trust proxy configuration
	if k.config.GetBool("FIBER_TRUST_PROXY") && k.config.Has("FIBER_TRUST_PROXIES") {
		// Validate each proxy in the list
		proxies := k.config.GetStringSlice("FIBER_TRUST_PROXIES")
		if slices.Contains(proxies, "") {
			return &contract.ConfigValidationError{
				Module:      "http",
				InvalidKeys: map[string]string{"FIBER_TRUST_PROXIES": "proxy address cannot be empty"},
			}
		}
	}

	// Validate view engine configuration if present
	if k.config.Has("VIEWS_ENGINE") {
		engine := k.config.GetString("VIEWS_ENGINE")
		if engine != "html" {
			return &contract.ConfigValidationError{
				Module:      "http",
				InvalidKeys: map[string]string{"VIEWS_ENGINE": "unsupported engine; only 'html' is supported"},
			}
		}
		// Optional: ensure non-empty strings if provided
		if k.config.Has("VIEWS_ROOT") && k.config.GetString("VIEWS_ROOT") == "" {
			return &contract.ConfigValidationError{
				Module:      "http",
				InvalidKeys: map[string]string{"VIEWS_ROOT": "cannot be empty when VIEWS_ENGINE is set"},
			}
		}
		if k.config.Has("VIEWS_EXT") && k.config.GetString("VIEWS_EXT") == "" {
			return &contract.ConfigValidationError{
				Module:      "http",
				InvalidKeys: map[string]string{"VIEWS_EXT": "cannot be empty when VIEWS_ENGINE is set"},
			}
		}
	}

	return v.Validate()
}

// field implementation for HTTP module
type field struct {
	key   string
	value any
}

func (f *field) Key() string {
	return f.key
}

func (f *field) Value() any {
	return f.value
}

// NewField creates a new field for structured logging
func NewField(key string, value any) contract.Field {
	return &field{key: key, value: value}
}

// HandlerParams is used for dependency injection in HTTP handlers
type HandlerParams struct {
	fx.In

	Config contract.Config
	Logger contract.Logger
	// Add other dependencies that handlers might need
}

// NewHandler creates a fiber handler with dependency injection
// This helper function makes it easy to create handlers that have access to DI services
func NewHandler(fn func(c fiber.Ctx, params HandlerParams) error) func(params HandlerParams) fiber.Handler {
	return func(params HandlerParams) fiber.Handler {
		return func(c fiber.Ctx) error {
			return fn(c, params)
		}
	}
}

// RouteRegistrar is a helper for registering routes with DI
type RouteRegistrar struct {
	fx.In

	HTTP   contract.HTTPKernel
	Config contract.Config
	Logger contract.Logger
}

// RegisterRoutes is a helper function that can be used to register routes
// Example usage:
//
//	fx.Invoke(http.RegisterRoutes(func(r http.RouteRegistrar) {
//	    r.HTTP.App().Get("/", myHandler)
//	}))
func RegisterRoutes(fn func(RouteRegistrar)) any {
	return fn
}
