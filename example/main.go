package main

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/http"
	"go.oease.dev/goe/v2/core/log"
	"go.uber.org/fx"
)

// UserService is an example service that will be injected
type UserService struct {
	logger contract.Logger
	config contract.Config
}

// NewUserService creates a new user service
func NewUserService(logger contract.Logger, config contract.Config) *UserService {
	return &UserService{
		logger: logger,
		config: config,
	}
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(id string) map[string]any {
	s.logger.Info("Getting user", log.NewField("user_id", id))

	// Simulate fetching user from database
	return map[string]any{
		"id":    id,
		"name":  "John Doe",
		"email": "john@example.com",
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(data map[string]any) map[string]any {
	s.logger.Info("Creating user", log.NewField("data", data))

	// Simulate creating user
	data["id"] = "123"
	data["created_at"] = time.Now()

	return data
}

// UserController handles user-related HTTP requests
type UserController struct {
	userService *UserService
	logger      contract.Logger
}

// NewUserController creates a new user controller
func NewUserController(userService *UserService, logger contract.Logger) *UserController {
	return &UserController{
		userService: userService,
		logger:      logger,
	}
}

// RegisterRoutes registers user routes
func (ctrl *UserController) RegisterRoutes(app *fiber.App) {
	// Method 1: Using context helpers to get services
	app.Get("/api/v1/users/:id", func(c fiber.Ctx) error {
		logger := http.GetLogger(c)
		config := http.GetConfig(c)

		logger.Info("Handler using context helpers",
			log.NewField("api_version", config.GetString("API_VERSION")),
		)

		user := ctrl.userService.GetUser(c.Params("id"))
		return c.JSON(user)
	})

	// Method 2: Using the injected services directly
	app.Post("/api/v1/users", func(c fiber.Ctx) error {
		var data map[string]any
		if err := c.Bind().JSON(&data); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		user := ctrl.userService.CreateUser(data)
		return c.Status(fiber.StatusCreated).JSON(user)
	})
}

// HealthController handles health check endpoints
type HealthController struct {
	fx.In

	Logger contract.Logger
	Config contract.Config
	App    contract.Application
}

// RegisterHealthRoutes registers health check routes
// This demonstrates using Fx parameter objects for cleaner DI
func RegisterHealthRoutes(params HealthController) func(*fiber.App) {
	return func(app *fiber.App) {
		app.Get("/health", func(c fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"status":      "healthy",
				"app_name":    params.Config.GetString("APP_NAME"),
				"version":     params.App.Version(),
				"environment": params.App.Environment(),
				"uptime":      time.Since(c.Locals("start_time").(time.Time)).String(),
			})
		})

		app.Get("/ready", func(c fiber.Ctx) error {
			// Check if all services are ready
			ready := params.App.IsRunning()

			if ready {
				return c.JSON(fiber.Map{"ready": true})
			}

			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"ready": false,
				"error": "Service not ready",
			})
		})
	}
}

// CustomMiddleware demonstrates creating middleware with DI
func CustomMiddleware(logger contract.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Store start time for uptime calculation
		c.Locals("start_time", time.Now())

		// Log custom middleware execution
		logger.Debug("Custom middleware executed",
			log.NewField("path", c.Path()),
			log.NewField("method", c.Method()),
		)

		return c.Next()
	}
}

// RouteRegistrar handles all route registrations
type RouteRegistrar struct {
	fx.In

	HTTP           contract.HTTPKernel
	UserController *UserController
	Logger         contract.Logger
}

// RegisterAllRoutes registers all application routes
func RegisterAllRoutes(r RouteRegistrar, health HealthController) {
	app := r.HTTP.App()

	// Add custom middleware
	app.Use(CustomMiddleware(r.Logger))

	// Register health routes
	RegisterHealthRoutes(health)(app)

	// Register user routes
	r.UserController.RegisterRoutes(app)

	// Add a simple home route
	app.Get("/", func(c fiber.Ctx) error {
		// Get services from context
		logger := http.GetLogger(c)
		config := http.GetConfig(c)

		logger.Info("Home page accessed")

		return c.JSON(fiber.Map{
			"message": "Welcome to " + config.GetString("APP_NAME"),
			"version": goe.App().Version(),
			"docs":    "/api/docs",
		})
	})

	// API documentation route
	app.Get("/api/docs", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"endpoints": []fiber.Map{
				{
					"method":      "GET",
					"path":        "/",
					"description": "Home page",
				},
				{
					"method":      "GET",
					"path":        "/health",
					"description": "Health check",
				},
				{
					"method":      "GET",
					"path":        "/ready",
					"description": "Readiness check",
				},
				{
					"method":      "GET",
					"path":        "/api/v1/users/:id",
					"description": "Get user by ID",
				},
				{
					"method":      "POST",
					"path":        "/api/v1/users",
					"description": "Create new user",
				},
			},
		})
	})

	// 404 handler
	app.Use(func(c fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found",
			"path":  c.Path(),
		})
	})
}

func main() {
	// Create application with HTTP enabled
	_ = goe.New(goe.Options{
		Name:        "Goe HTTP Example",
		Version:     "1.0.0",
		Environment: "dev",
		WithHTTP:    true, // Enable HTTP module
		Providers: []any{
			// Provide services
			NewUserService,
			NewUserController,
		},
		Invokers: []any{
			// Register routes after all dependencies are ready
			RegisterAllRoutes,

			// Log application startup
			func(logger contract.Logger, config contract.Config) {
				logger.Info("Application configured",
					log.NewField("http_port", config.GetInt("HTTP_PORT")),
					log.NewField("debug", config.GetBool("DEBUG")),
				)
			},
		},
	})

	// Access HTTP module globally
	goe.Log().Info("Starting HTTP server...")

	// You can also access the Fiber app directly for advanced configuration
	app := goe.HTTP().App()

	// Add global configuration
	app.Hooks().OnListen(func(data fiber.ListenData) error {
		if fiber.IsChild() {
			return nil
		}
		scheme := "http"
		if data.TLS {
			scheme = "https"
		}
		goe.Log().Info("Server started",
			log.NewField("url", scheme+"://"+data.Host+":"+data.Port),
		)
		return nil
	})

	// Run the application
	goe.Run()
}
