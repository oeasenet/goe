package http

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/fx"
)

// kernel implements the HTTPKernel interface
type kernel struct {
	app    *fiber.App
	config contract.Config
	logger contract.Logger
}

// New creates a new HTTP kernel
func New(config contract.Config, logger contract.Logger) contract.HTTPKernel {
	// Create fiber config
	fiberConfig := fiber.Config{
		ServerHeader:  config.GetString("HTTP_SERVER_HEADER"),
		StrictRouting: config.GetBool("HTTP_STRICT_ROUTING"),
		CaseSensitive: config.GetBool("HTTP_CASE_SENSITIVE"),
		BodyLimit:     config.GetInt("HTTP_BODY_LIMIT"),
		ReadTimeout:   config.GetDuration("HTTP_READ_TIMEOUT"),
		WriteTimeout:  config.GetDuration("HTTP_WRITE_TIMEOUT"),
		IdleTimeout:   config.GetDuration("HTTP_IDLE_TIMEOUT"),
		AppName:       config.GetString("APP_NAME"),
		ErrorHandler:  defaultErrorHandler(logger),
	}

	// Set defaults
	if fiberConfig.ServerHeader == "" {
		fiberConfig.ServerHeader = "Goe"
	}
	if fiberConfig.BodyLimit == 0 {
		fiberConfig.BodyLimit = 4 * 1024 * 1024 // 4MB
	}
	if fiberConfig.ReadTimeout == 0 {
		fiberConfig.ReadTimeout = 10 * time.Second
	}
	if fiberConfig.WriteTimeout == 0 {
		fiberConfig.WriteTimeout = 10 * time.Second
	}

	// Create fiber app
	app := fiber.New(fiberConfig)

	// Add default middleware
	app.Use(recover.New())
	app.Use(requestid.New())

	// Add request logging middleware
	app.Use(func(c fiber.Ctx) error {
		start := time.Now()

		// Store request ID in locals for use in handlers
		requestID := c.Get("X-Request-ID")
		c.Locals("requestID", requestID)

		// Continue to next middleware
		err := c.Next()

		// Log request
		logger.Info("HTTP Request",
			contract.Field(newField("method", c.Method())),
			contract.Field(newField("path", c.Path())),
			contract.Field(newField("status", c.Response().StatusCode())),
			contract.Field(newField("duration", time.Since(start).String())),
			contract.Field(newField("request_id", requestID)),
		)

		return err
	})

	return &kernel{
		app:    app,
		config: config,
		logger: logger,
	}
}

// App returns the underlying Fiber app
func (k *kernel) App() *fiber.App {
	return k.app
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
		contract.Field(newField("address", addr)),
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
				contract.Field(newField("error", err.Error())),
				contract.Field(newField("path", ctx.Path())),
				contract.Field(newField("method", ctx.Method())),
				contract.Field(newField("status", respCode)),
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
			ctx.Response().Header.SetContentType(fiber.MIMETextHTML)
			return ctx.SendString(ErrorPage(fmt.Sprintf("ERROR %d", respCode), fmt.Sprintf("%d", respCode), message, "/"))
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
		ctx.Response().Header.SetContentType(fiber.MIMETextHTML)
		return ctx.SendString(ErrorPage(fmt.Sprintf("ERROR %d", respCode), fmt.Sprintf("%d", respCode), message, "/"))
	}
}

// Module represents the HTTP module for Fx
type Module struct {
	kernel contract.HTTPKernel
}

// NewModule creates a new HTTP module
func NewModule(config contract.Config, logger contract.Logger) *Module {
	return &Module{
		kernel: New(config, logger),
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
				contract.Field(newField("error", err.Error())),
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

func newField(key string, value any) contract.Field {
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
