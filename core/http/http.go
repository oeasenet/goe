package http

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/validator"
	"go.uber.org/fx"
)

// kernel implements the HTTPKernel interface
type kernel struct {
	app       *fiber.App
	config    contract.Config
	logger    contract.Logger
	validator *CustomValidator
}

// New creates a new HTTP kernel
func New(config contract.Config, logger contract.Logger) contract.HTTPKernel {
	// Create validator
	validator := NewValidator()

	// Init error page template
	var err error
	tpl, err = template.ParseFS(templateFS, "error_page.gohtml")
	if err != nil {
		logger.Fatal(err.Error())
		panic(err)
		return nil
	}

	// Create fiber config with all supported options
	fiberConfig := fiber.Config{
		ServerHeader:       config.GetString("FIBER_SERVER_HEADER"),
		StrictRouting:      config.GetBool("FIBER_STRICT_ROUTING"),
		CaseSensitive:      config.GetBool("FIBER_CASE_SENSITIVE"),
		Immutable:          config.GetBool("FIBER_IMMUTABLE"),
		UnescapePath:       config.GetBool("FIBER_UNESCAPE_PATH"),
		BodyLimit:          config.GetInt("FIBER_BODY_LIMIT"),
		StreamRequestBody:  config.GetBool("FIBER_STREAM_REQUEST_BODY"),
		Concurrency:        config.GetInt("FIBER_CONCURRENCY"),
		ProxyHeader:        config.GetString("FIBER_PROXY_HEADER"),
		AppName:            config.GetString("APP_NAME"),
		ReduceMemoryUsage:  config.GetBool("FIBER_REDUCE_MEMORY"),
		JSONEncoder:        sonic.Marshal,
		JSONDecoder:        sonic.Unmarshal,
		XMLEncoder:         xml.Marshal,
		EnableIPValidation: config.GetBool("FIBER_ENABLE_IP_VALIDATION"),
		ColorScheme:        fiber.DefaultColors,
		StructValidator:    validator,
		ErrorHandler:       defaultErrorHandler(logger),
	}

	// Handle TrustProxy configuration
	if config.GetBool("FIBER_TRUST_PROXY") {
		fiberConfig.TrustProxy = true

		// Parse trusted proxies
		trustedProxies := config.GetStringSlice("FIBER_TRUST_PROXIES")
		if len(trustedProxies) > 0 {
			fiberConfig.TrustProxyConfig = fiber.TrustProxyConfig{
				Proxies:   trustedProxies,
				LinkLocal: config.GetBool("FIBER_TRUST_LINK_LOCAL"),
				Loopback:  config.GetBool("FIBER_TRUST_LOOPBACK"),
				Private:   config.GetBool("FIBER_TRUST_PRIVATE"),
			}
			// Set defaults for trust proxy config if not specified
			if !config.Has("FIBER_TRUST_LINK_LOCAL") {
				fiberConfig.TrustProxyConfig.LinkLocal = true
			}
			if !config.Has("FIBER_TRUST_LOOPBACK") {
				fiberConfig.TrustProxyConfig.Loopback = true
			}
			if !config.Has("FIBER_TRUST_PRIVATE") {
				fiberConfig.TrustProxyConfig.Private = true
			}
		}
	}

	// Set defaults
	if fiberConfig.ServerHeader == "" {
		fiberConfig.ServerHeader = "Goe"
	}
	if fiberConfig.BodyLimit == 0 {
		fiberConfig.BodyLimit = 4 * 1024 * 1024 // 4MB
	}
	if !config.Has("FIBER_STREAM_REQUEST_BODY") {
		fiberConfig.StreamRequestBody = true
	}
	if fiberConfig.Concurrency == 0 {
		fiberConfig.Concurrency = 256 * 1024 // Fiber's default
	}

	// Set timeouts from HTTP_ prefixed config for backward compatibility
	if fiberConfig.ReadTimeout == 0 {
		fiberConfig.ReadTimeout = config.GetDuration("HTTP_READ_TIMEOUT")
		if fiberConfig.ReadTimeout == 0 {
			fiberConfig.ReadTimeout = 10 * time.Second
		}
	}
	if fiberConfig.WriteTimeout == 0 {
		fiberConfig.WriteTimeout = config.GetDuration("HTTP_WRITE_TIMEOUT")
		if fiberConfig.WriteTimeout == 0 {
			fiberConfig.WriteTimeout = 10 * time.Second
		}
	}
	if fiberConfig.IdleTimeout == 0 {
		fiberConfig.IdleTimeout = config.GetDuration("HTTP_IDLE_TIMEOUT")
		if fiberConfig.IdleTimeout == 0 {
			fiberConfig.IdleTimeout = 30 * time.Second
		}
	}

	// Create fiber app
	app := fiber.New(fiberConfig)

	// Add default middleware
	app.Use(recover.New())
	app.Use(requestid.New())

	// Add request logging middleware
	app.Use(func(c fiber.Ctx) error {
		start := time.Now()

		// Retrieve request ID set by the requestid middleware
		requestID := requestid.FromContext(c)
		if requestID == "" {
			requestID = c.Get(fiber.HeaderXRequestID)
		}
		c.Locals(string(RequestIDKey), requestID)

		// Continue to next middleware
		err := c.Next()

		// Log request
		logger.Info("HTTP Request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration", time.Since(start).String(),
			"request_id", requestID,
		)

		return err
	})

	return &kernel{
		app:       app,
		config:    config,
		logger:    logger,
		validator: validator,
	}
}

// App returns the underlying Fiber app
func (k *kernel) App() *fiber.App {
	return k.app
}

// Validator returns the struct validator
func (k *kernel) Validator() any {
	return k.validator
}

// Listen starts the HTTP server
func (k *kernel) Listen(addr string) error {
	if addr == "" {
		host := k.config.GetString("HTTP_HOST")
		if host == "" {
			host = "0.0.0.0"
		}

		port := k.config.GetInt("HTTP_PORT")
		if port == 0 {
			port = 8080
		}

		addr = fmt.Sprintf("%s:%d", host, port)
	}

	k.logger.Info("HTTP server starting",
		"address", addr,
	)

	return k.app.Listen(addr)
}

// Shutdown gracefully shuts down the server
func (k *kernel) Shutdown() error {
	k.logger.Info("HTTP server shutting down")
	return k.app.Shutdown()
}

// defaultErrorHandler creates a default error handler
func defaultErrorHandler(logger contract.Logger) fiber.ErrorHandler {
	return func(ctx fiber.Ctx, err error) error {
		// Status code defaults to 500
		respCode := fiber.StatusInternalServerError
		// Set error message
		message := err.Error()
		// Check if it's a fiber.Error type
		var e *fiber.Error
		if errors.As(err, &e) {
			respCode = e.Code
			message = e.Message
		}
		ctx.Status(respCode)

		// Log error for 5xx errors
		if respCode >= 500 {
			logger.Error("HTTP Error",
				"error", err.Error(),
				"path", ctx.Path(),
				"method", ctx.Method(),
				"status", respCode,
			)
		}

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

// Module represents the HTTP module for Fx
type Module struct {
	kernel    contract.HTTPKernel
	validator *CustomValidator
}

// NewModule creates a new HTTP module
func NewModule(config contract.Config, logger contract.Logger) *Module {
	kernel := New(config, logger)
	return &Module{
		kernel:    kernel,
		validator: kernel.Validator().(*CustomValidator),
	}
}

// Name returns the module name
func (m *Module) Name() string {
	return "http"
}

// OnStart is called when the module starts
func (m *Module) OnStart(ctx context.Context) error {
	// Get listen address
	host := m.kernel.(*kernel).config.GetString("HTTP_HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	port := m.kernel.(*kernel).config.GetInt("HTTP_PORT")
	if port == 0 {
		port = 8080
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	// Start server in background
	go func() {
		if err := m.kernel.Listen(addr); err != nil {
			m.kernel.(*kernel).logger.Error("HTTP server error",
				"error", err.Error(),
			)
		}
	}()

	return nil
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
func (m *Module) ProvideValidator() *CustomValidator {
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
	v := validator.NewConfigValidator(k.config, "http")

	// HTTP port is optional but should be valid if set
	if k.config.Has("HTTP_PORT") {
		v.Optional("HTTP_PORT", "HTTP server port", validator.ValidatePort)
	}

	// Validate Fiber-specific configurations if set
	if k.config.Has("FIBER_BODY_LIMIT") {
		v.Optional("FIBER_BODY_LIMIT", "Request body size limit", validator.ValidatePositiveInt)
	}

	if k.config.Has("FIBER_CONCURRENCY") {
		v.Optional("FIBER_CONCURRENCY", "Maximum concurrent connections", validator.ValidatePositiveInt)
	}

	// Validate trust proxy configuration
	if k.config.GetBool("FIBER_TRUST_PROXY") && k.config.Has("FIBER_TRUST_PROXIES") {
		// Validate each proxy in the list
		proxies := k.config.GetStringSlice("FIBER_TRUST_PROXIES")
		for _, proxy := range proxies {
			// Simple validation - could be IP or CIDR
			if proxy == "" {
				return &contract.ConfigValidationError{
					Module:      "http",
					InvalidKeys: map[string]string{"FIBER_TRUST_PROXIES": "proxy address cannot be empty"},
				}
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
