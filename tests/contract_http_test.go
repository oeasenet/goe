package tests

import (
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/contract"
)

// MockHTTPKernel is a mock implementation of the HTTPKernel interface
type MockHTTPKernel struct {
	mock.Mock
}

func (m *MockHTTPKernel) App() *fiber.App {
	args := m.Called()
	return args.Get(0).(*fiber.App)
}

func (m *MockHTTPKernel) Listen(addr string) error {
	args := m.Called(addr)
	return args.Error(0)
}

func (m *MockHTTPKernel) Shutdown() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockHTTPKernel) Validator() any {
	args := m.Called()
	return args.Get(0)
}

func TestHTTPKernelInterface(t *testing.T) {
	// This test verifies that MockHTTPKernel implements the HTTPKernel interface
	var _ contract.HTTPKernel = (*MockHTTPKernel)(nil)

	// Create a mock HTTP kernel
	kernel := new(MockHTTPKernel)

	// Set up expectations
	app := fiber.New()
	kernel.On("App").Return(app)
	kernel.On("Listen", ":8080").Return(nil)
	kernel.On("Shutdown").Return(nil)
	kernel.On("Validator").Return(struct{}{})

	// Test the methods
	assert.Equal(t, app, kernel.App())
	assert.NoError(t, kernel.Listen(":8080"))
	assert.NoError(t, kernel.Shutdown())
	assert.NotNil(t, kernel.Validator())

	// Verify expectations
	kernel.AssertExpectations(t)
}

func TestHTTPConfig(t *testing.T) {
	// Test HTTPConfig struct
	config := contract.HTTPConfig{
		Port:           8080,
		Host:           "localhost",
		ReadTimeout:    "30s",
		WriteTimeout:   "30s",
		IdleTimeout:    "120s",
		BodyLimit:      1 << 20,
		TrustedProxies: []string{"127.0.0.1"},
	}

	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, "30s", config.ReadTimeout)
	assert.Equal(t, "30s", config.WriteTimeout)
	assert.Equal(t, "120s", config.IdleTimeout)
	assert.Equal(t, 1<<20, config.BodyLimit)
	assert.Equal(t, []string{"127.0.0.1"}, config.TrustedProxies)
}

func TestRouteInfo(t *testing.T) {
	// Test RouteInfo struct
	handler := func(c fiber.Ctx) error {
		return c.SendString("Hello")
	}

	route := contract.RouteInfo{
		Method:  "POST",
		Path:    "/api/users",
		Name:    "users.create",
		Handler: handler,
	}

	assert.Equal(t, "POST", route.Method)
	assert.Equal(t, "/api/users", route.Path)
	assert.Equal(t, "users.create", route.Name)
	assert.NotNil(t, route.Handler)
}
