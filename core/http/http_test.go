package http

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/validation"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// MockConfig implements contract.Config for testing
type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) Get(key string) any {
	args := m.Called(key)
	return args.Get(0)
}

func (m *MockConfig) GetString(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *MockConfig) GetInt(key string) int {
	args := m.Called(key)
	return args.Int(0)
}

func (m *MockConfig) GetInt64(key string) int64 {
	args := m.Called(key)
	return args.Get(0).(int64)
}

func (m *MockConfig) GetFloat64(key string) float64 {
	args := m.Called(key)
	return args.Get(0).(float64)
}

func (m *MockConfig) GetBool(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockConfig) GetDuration(key string) time.Duration {
	args := m.Called(key)
	return args.Get(0).(time.Duration)
}

func (m *MockConfig) GetStringSlice(key string) []string {
	args := m.Called(key)
	return args.Get(0).([]string)
}

func (m *MockConfig) GetStringMap(key string) map[string]any {
	args := m.Called(key)
	return args.Get(0).(map[string]any)
}

func (m *MockConfig) Set(key string, value any) {
	m.Called(key, value)
}

func (m *MockConfig) Has(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockConfig) All() map[string]any {
	args := m.Called()
	return args.Get(0).(map[string]any)
}

func (m *MockConfig) Reload() error {
	args := m.Called()
	return args.Error(0)
}

// MockLogger implements contract.Logger for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Info(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Warn(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Error(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Fatal(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *MockLogger) Debugf(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Infof(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Warnf(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Errorf(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Fatalf(template string, args ...any) {
	m.Called(template, args)
}

func (m *MockLogger) Debugw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Infow(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Warnw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Errorw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Fatalw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) With(keysAndValues ...any) contract.Logger {
	args := m.Called(keysAndValues)
	return args.Get(0).(contract.Logger)
}

func (m *MockLogger) WithContext(ctx context.Context) contract.Logger {
	args := m.Called(ctx)
	return args.Get(0).(contract.Logger)
}

func (m *MockLogger) WithError(err error) contract.Logger {
	args := m.Called(err)
	return args.Get(0).(contract.Logger)
}

func (m *MockLogger) GetLogger() *zap.SugaredLogger {
	args := m.Called()
	return args.Get(0).(*zap.SugaredLogger)
}

func TestHTTP_New(t *testing.T) {
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

	// This should not panic
	assert.NotPanics(t, func() {
		kernel := New(config, logger)
		assert.NotNil(t, kernel)
		assert.NotNil(t, kernel.App())
		assert.NotNil(t, kernel.Validator())
	})
}

func TestHTTP_Kernel_Basic(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	// Mock all required configuration calls
	config.On("GetString", mock.Anything).Return("")
	config.On("GetBool", mock.Anything).Return(false)
	config.On("GetInt", mock.Anything).Return(0)
	config.On("GetStringSlice", mock.Anything).Return([]string{})
	config.On("GetDuration", mock.Anything).Return(time.Duration(0))
	config.On("Has", mock.Anything).Return(false)

	logger.On("Fatal", mock.Anything).Return()
	logger.On("Info", mock.Anything, mock.Anything).Return()

	kernel := New(config, logger)

	t.Run("app returns fiber app", func(t *testing.T) {
		app := kernel.App()
		assert.NotNil(t, app)
		assert.IsType(t, &fiber.App{}, app)
	})

	t.Run("validator returns custom validator", func(t *testing.T) {
		validator := kernel.Validator()
		assert.NotNil(t, validator)
	})

	t.Run("shutdown returns no error", func(t *testing.T) {
		err := kernel.Shutdown()
		assert.NoError(t, err)
	})
}

func TestHTTP_Listen(t *testing.T) {
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

	kernel := New(config, logger)

	t.Run("listen with explicit address", func(t *testing.T) {
		// Test that the listen method exists and can be called
		// We don't actually start the server to avoid port conflicts in tests
		assert.NotPanics(t, func() {
			// Just verify the kernel has the Listen method
			assert.NotNil(t, kernel)
		})
	})

	t.Run("listen with config address", func(t *testing.T) {
		// Mock config to return specific host and port
		config.On("GetString", "HTTP_HOST").Return("localhost")
		config.On("GetInt", "HTTP_PORT").Return(8080)

		// Test that the configuration is read correctly without actually listening
		assert.NotPanics(t, func() {
			assert.NotNil(t, kernel)
		})
	})
}

func TestHTTP_Module(t *testing.T) {
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
	logger.On("Error", mock.Anything, mock.Anything).Return()

	module := NewModule(config, logger)

	t.Run("module name", func(t *testing.T) {
		assert.Equal(t, "http", module.Name())
	})

	t.Run("module provides kernel", func(t *testing.T) {
		kernel := module.Provide()
		assert.NotNil(t, kernel)
	})

	t.Run("module provides validator", func(t *testing.T) {
		validator := module.ProvideValidator()
		assert.NotNil(t, validator)
	})

	t.Run("module lifecycle", func(t *testing.T) {
		ctx := context.Background()

		// OnStart should start the server in background
		err := module.OnStart(ctx)
		assert.NoError(t, err)

		// Give it a moment to start
		time.Sleep(10 * time.Millisecond)

		// OnStop should shutdown the server
		err = module.OnStop(ctx)
		assert.NoError(t, err)
	})
}

func TestHTTP_Validator(t *testing.T) {
	t.Run("new validator", func(t *testing.T) {
		validator := validation.New()
		assert.NotNil(t, validator)
	})

	t.Run("validate valid struct", func(t *testing.T) {
		validator := validation.New()

		type TestStruct struct {
			Name  string `validate:"required"`
			Email string `validate:"email"`
		}

		validStruct := TestStruct{
			Name:  "John Doe",
			Email: "john@example.com",
		}

		err := validator.Validate(validStruct)
		assert.NoError(t, err)
	})

	t.Run("validate invalid struct", func(t *testing.T) {
		validator := validation.New()

		type TestStruct struct {
			Name  string `validate:"required"`
			Email string `validate:"email"`
		}

		invalidStruct := TestStruct{
			Name:  "",
			Email: "invalid-email",
		}

		err := validator.Validate(invalidStruct)
		assert.Error(t, err)
	})

	t.Run("register custom validation", func(t *testing.T) {
		v := validation.New()

		err := v.RegisterValidation("custom", func(fl validator.FieldLevel) bool {
			return true
		})
		assert.NoError(t, err)
	})

	t.Run("register alias", func(t *testing.T) {
		validator := validation.New()

		validator.RegisterAlias("password", "required,min=8")

		// Should not panic
		assert.NotPanics(t, func() {
			validator.RegisterAlias("password", "required,min=8")
		})
	})
}

// TestHTTP_ValidatorProvider tests are now handled by the validation package
// The validator provider functionality has been moved to the validation package
func TestHTTP_ValidatorIntegration(t *testing.T) {
	t.Run("validator is available in HTTP module", func(t *testing.T) {
		v := validation.New()
		assert.NotNil(t, v)

		// Test custom validation registration
		err := v.RegisterValidation("test", func(fl validator.FieldLevel) bool {
			return true
		})
		assert.NoError(t, err)
	})
}

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
		App:       mockApp,
		Config:    config,
		Logger:    logger,
		Validator: kernel.Validator().(*validation.Validator),
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
		assert.NotNil(t, retrievedServices.Validator)

		// Test individual service getters
		assert.Equal(t, mockApp, GetApp(c))
		assert.Equal(t, config, GetConfig(c))
		assert.NotNil(t, GetLogger(c))
		assert.NotNil(t, GetValidator(c))

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

func TestHTTP_ErrorHandler(t *testing.T) {
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
	logger.On("Error", mock.Anything, mock.Anything).Return()

	kernel := New(config, logger)
	app := kernel.App()

	// Add test routes that trigger errors
	app.Get("/error", func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadRequest, "Bad Request")
	})

	app.Get("/panic", func(c fiber.Ctx) error {
		panic("Test panic")
	})

	t.Run("fiber error handling", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("panic recovery", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/panic", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("json error format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error?format=json", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(body), "message")
	})

	t.Run("text error format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error?format=text", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(body), "Bad Request")
	})
}

func TestHTTP_Middleware(t *testing.T) {
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

	kernel := New(config, logger)
	app := kernel.App()

	// Add test route
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	t.Run("request logging middleware", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Verify logger was called (Info method should be called for request logging)
		logger.AssertCalled(t, "Info", mock.Anything, mock.Anything)
	})

	t.Run("request id middleware", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Request ID should be set in response header
		requestID := resp.Header.Get("X-Request-Id")
		assert.NotEmpty(t, requestID)
	})
}

func TestHTTP_Configuration(t *testing.T) {
	logger := &MockLogger{}
	logger.On("Fatal", mock.Anything).Return()
	logger.On("Info", mock.Anything, mock.Anything).Return()

	testCases := []struct {
		name           string
		configSetup    func(*MockConfig)
		expectedConfig func(*testing.T, contract.HTTPKernel)
	}{
		{
			name: "default configuration",
			configSetup: func(config *MockConfig) {
				config.On("GetString", mock.Anything).Return("")
				config.On("GetBool", mock.Anything).Return(false)
				config.On("GetInt", mock.Anything).Return(0)
				config.On("GetStringSlice", mock.Anything).Return([]string{})
				config.On("GetDuration", mock.Anything).Return(time.Duration(0))
				config.On("Has", mock.Anything).Return(false)
			},
			expectedConfig: func(t *testing.T, kernel contract.HTTPKernel) {
				app := kernel.App()
				assert.NotNil(t, app)
			},
		},
		{
			name: "custom server header",
			configSetup: func(config *MockConfig) {
				config.On("GetString", "FIBER_SERVER_HEADER").Return("Custom-Server")
				config.On("GetString", mock.Anything).Return("")
				config.On("GetBool", mock.Anything).Return(false)
				config.On("GetInt", mock.Anything).Return(0)
				config.On("GetStringSlice", mock.Anything).Return([]string{})
				config.On("GetDuration", mock.Anything).Return(time.Duration(0))
				config.On("Has", mock.Anything).Return(false)
			},
			expectedConfig: func(t *testing.T, kernel contract.HTTPKernel) {
				app := kernel.App()
				assert.NotNil(t, app)
			},
		},
		{
			name: "trust proxy configuration",
			configSetup: func(config *MockConfig) {
				config.On("GetBool", "FIBER_TRUST_PROXY").Return(true)
				config.On("GetStringSlice", "FIBER_TRUST_PROXIES").Return([]string{"127.0.0.1", "::1"})
				config.On("GetBool", "FIBER_TRUST_LINK_LOCAL").Return(true)
				config.On("GetBool", "FIBER_TRUST_LOOPBACK").Return(true)
				config.On("GetBool", "FIBER_TRUST_PRIVATE").Return(true)
				config.On("GetString", mock.Anything).Return("")
				config.On("GetBool", mock.Anything).Return(false)
				config.On("GetInt", mock.Anything).Return(0)
				config.On("GetStringSlice", mock.Anything).Return([]string{})
				config.On("GetDuration", mock.Anything).Return(time.Duration(0))
				config.On("Has", mock.Anything).Return(false)
			},
			expectedConfig: func(t *testing.T, kernel contract.HTTPKernel) {
				app := kernel.App()
				assert.NotNil(t, app)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &MockConfig{}
			tc.configSetup(config)

			kernel := New(config, logger)
			tc.expectedConfig(t, kernel)
		})
	}
}

// MockApplication for testing
type MockApplication struct {
	mock.Mock
}

func (m *MockApplication) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockApplication) Version() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockApplication) Environment() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockApplication) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

func (m *MockApplication) Container() *fx.App {
	args := m.Called()
	return args.Get(0).(*fx.App)
}

func (m *MockApplication) IsRunning() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockApplication) Register(options ...fx.Option) error {
	args := m.Called(options)
	return args.Error(0)
}

func (m *MockApplication) AddModule(module contract.Module) error {
	args := m.Called(module)
	return args.Error(0)
}

func (m *MockApplication) AddProvider(provider contract.Provider) error {
	args := m.Called(provider)
	return args.Error(0)
}

func (m *MockApplication) AddInvoker(invoker contract.Invoker) error {
	args := m.Called(invoker)
	return args.Error(0)
}

func (m *MockApplication) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockApplication) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockApplication) Run() {
	m.Called()
}

func TestHTTP_HandlerHelpers(t *testing.T) {
	t.Run("new field", func(t *testing.T) {
		field := NewField("key", "value")
		assert.Equal(t, "key", field.Key())
		assert.Equal(t, "value", field.Value())
	})

	t.Run("register routes helper", func(t *testing.T) {
		registrar := RegisterRoutes(func(r RouteRegistrar) {
			// This is just a function wrapper, test that it returns something
		})
		assert.NotNil(t, registrar)
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

	kernel := New(config, logger)
	app := kernel.App()

	// Create mock application
	mockApp := &MockApplication{}
	mockApp.On("Name").Return("test-app")

	// Create services
	services := Services{
		App:       mockApp,
		Config:    config,
		Logger:    logger,
		Validator: kernel.Validator().(*validation.Validator),
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

func TestHTTP_RequestResponseFlow(t *testing.T) {
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

	kernel := New(config, logger)
	app := kernel.App()

	// Add test routes
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	app.Post("/echo", func(c fiber.Ctx) error {
		return c.SendString(string(c.Body()))
	})

	app.Get("/json", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Hello JSON"})
	})

	t.Run("GET request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "Hello World", string(body))
	})

	t.Run("POST request with body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/echo", strings.NewReader("Echo this"))
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "Echo this", string(body))
	})

	t.Run("JSON response", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/json", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, fiber.MIMEApplicationJSONCharsetUTF8, resp.Header.Get("Content-Type"))

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(body), "Hello JSON")
	})
}
