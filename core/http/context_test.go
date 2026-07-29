package http

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestHTTP_Context(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	// Mock required configuration
	config.On("GetString", mock.Anything).Return("")
	config.On("GetBool", mock.Anything).Return(false)
	config.On("GetInt", mock.Anything).Return(0)
	config.On("GetStringSlice", mock.Anything).Return([]string{})
	config.On("GetDuration", mock.Anything).Return(time.Duration(0))
	config.On("Has", mock.Anything).Return(false)

	logger.On("Fatal", mock.Anything).Return()
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("With", mock.Anything).Return(logger)
	// Create a noop logger for GetLogger
	zapLogger, _ := zap.NewProduction()
	logger.On("GetLogger").Return(zapLogger.Sugar())

	kernel := New(config, logger)
	app := kernel.App()

	// Create mock application
	mockApp := &MockApplication{}
	mockApp.On("Name").Return("test-app")
	mockApp.On("Version").Return("1.0.0")
	mockApp.On("Environment").Return("test")
	mockApp.On("IsRunning").Return(true)
	mockApp.On("Context").Return(context.Background())

	// Create services
	services := Services{
		App:    mockApp,
		Config: config,
		Logger: logger,
	}

	// Add service injection middleware
	app.Use(InjectServices(services))

	// Add test route
	app.Get("/test", func(c fiber.Ctx) error {
		// Test service retrieval
		retrievedServices := GetServices(c)
		assert.NotNil(t, retrievedServices.App)
		assert.NotNil(t, retrievedServices.Config)
		assert.NotNil(t, retrievedServices.Logger)

		// Test individual service getters
		assert.Equal(t, mockApp, GetApp(c))
		assert.Equal(t, config, GetConfig(c))
		assert.NotNil(t, GetLogger(c))
		// Validation is no longer an injected service: Fiber's Bind calls the
		// StructValidator installed on the app directly.

		return c.SendString("OK")
	})

	t.Run("service injection and retrieval", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "OK", string(body))
	})
}

func TestHTTP_HandlerHelpers(t *testing.T) {
	// RouteRegistrar is an fx.In struct, so its value comes from the container.
	// The contract that matters is that it carries the dependencies route
	// registration needs; an invoker taking it must be usable as-is.
	t.Run("route registrar carries registration dependencies", func(t *testing.T) {
		var invoker any = func(r RouteRegistrar) {
			_ = r.HTTP
			_ = r.Config
			_ = r.Logger
		}
		assert.NotNil(t, invoker)
	})
}

func TestHTTP_GroupRouter(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	// Mock required configuration
	config.On("GetString", mock.Anything).Return("")
	config.On("GetBool", mock.Anything).Return(false)
	config.On("GetInt", mock.Anything).Return(0)
	config.On("GetStringSlice", mock.Anything).Return([]string{})
	config.On("GetDuration", mock.Anything).Return(time.Duration(0))
	config.On("Has", mock.Anything).Return(false)

	logger.On("Fatal", mock.Anything).Return()
	logger.On("Info", mock.Anything, mock.Anything).Return()
	// Create a noop logger for GetLogger
	zapLogger, _ := zap.NewProduction()
	logger.On("GetLogger").Return(zapLogger.Sugar())

	kernel := New(config, logger)
	app := kernel.App()

	// Create mock application
	mockApp := &MockApplication{}
	mockApp.On("Name").Return("test-app")

	// Create services
	services := Services{
		App:    mockApp,
		Config: config,
		Logger: logger,
	}

	// Create group
	group := NewGroup(app.Group("/api"), services)

	t.Run("group GET method", func(t *testing.T) {
		group.GET("/test", func(c fiber.Ctx, deps Services) error {
			assert.NotNil(t, deps.App)
			assert.NotNil(t, deps.Config)
			assert.NotNil(t, deps.Logger)
			return c.SendString("GET OK")
		})

		req := httptest.NewRequest("GET", "/api/test", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "GET OK", string(body))
	})

	t.Run("group POST method", func(t *testing.T) {
		group.POST("/test", func(c fiber.Ctx, deps Services) error {
			return c.SendString("POST OK")
		})

		req := httptest.NewRequest("POST", "/api/test", bytes.NewBufferString("test"))
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "POST OK", string(body))
	})

	t.Run("group PUT method", func(t *testing.T) {
		group.PUT("/test", func(c fiber.Ctx, deps Services) error {
			return c.SendString("PUT OK")
		})

		req := httptest.NewRequest("PUT", "/api/test", bytes.NewBufferString("test"))
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("group DELETE method", func(t *testing.T) {
		group.DELETE("/test", func(c fiber.Ctx, deps Services) error {
			return c.SendString("DELETE OK")
		})

		req := httptest.NewRequest("DELETE", "/api/test", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("group PATCH method", func(t *testing.T) {
		group.PATCH("/test", func(c fiber.Ctx, deps Services) error {
			return c.SendString("PATCH OK")
		})

		req := httptest.NewRequest("PATCH", "/api/test", bytes.NewBufferString("test"))
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	// QUERY (RFC 10008) carries the query expression in the request body and is
	// enabled by default from Fiber v3.4 onwards.
	t.Run("group QUERY method", func(t *testing.T) {
		group.QUERY("/test", func(c fiber.Ctx, deps Services) error {
			assert.NotNil(t, deps.App)
			return c.Send(c.Body())
		})

		req := httptest.NewRequest(fiber.MethodQuery, "/api/test", bytes.NewBufferString("select * where id = 1"))
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "select * where id = 1", string(body))
	})
}

func TestHTTP_AsHandler(t *testing.T) {
	t.Run("as handler conversion", func(t *testing.T) {
		handler := AsHandler(func(c fiber.Ctx, deps string) error {
			return c.SendString("Handler with deps: " + deps)
		}, "test-deps")

		assert.NotNil(t, handler)

		// Create a test fiber app to test the handler
		app := fiber.New()
		app.Get("/test", handler)

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "Handler with deps: test-deps", string(body))
	})
}

func TestHTTP_ServiceProvider(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	// Mock required configuration
	config.On("GetString", mock.Anything).Return("")
	config.On("GetBool", mock.Anything).Return(false)
	config.On("GetInt", mock.Anything).Return(0)
	config.On("GetStringSlice", mock.Anything).Return([]string{})
	config.On("GetDuration", mock.Anything).Return(time.Duration(0))
	config.On("Has", mock.Anything).Return(false)

	logger.On("Fatal", mock.Anything).Return()
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()

	// Create mock application
	mockApp := &MockApplication{}
	mockApp.On("Name").Return("test-app")

	provider := ServiceProvider{
		App:    mockApp,
		Config: config,
		Logger: logger,
	}

	t.Run("create service middleware", func(t *testing.T) {
		middleware := CreateServiceMiddleware(provider)
		assert.NotNil(t, middleware)

		// Test the middleware works
		app := fiber.New()
		app.Use(middleware)
		app.Get("/test", func(c fiber.Ctx) error {
			services := GetServices(c)
			assert.NotNil(t, services.App)
			assert.NotNil(t, services.Config)
			assert.NotNil(t, services.Logger)
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}
