package contract_test

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

	// Create a fiber app for testing
	app := fiber.New()

	// Set up expectations
	kernel.On("App").Return(app)
	kernel.On("Listen", ":8080").Return(nil)
	kernel.On("Shutdown").Return(nil)
	kernel.On("Validator").Return(nil)

	// Test the methods
	assert.Equal(t, app, kernel.App())
	assert.Nil(t, kernel.Listen(":8080"))
	assert.Nil(t, kernel.Shutdown())
	assert.Nil(t, kernel.Validator())

	// Verify expectations
	kernel.AssertExpectations(t)
}

func TestHTTPConfig(t *testing.T) {
	// Test the HTTPConfig struct
	config := contract.HTTPConfig{
		Host:          "localhost",
		Port:          8080,
		Prefork:       true,
		ServerHeader:  "Goe",
		StrictRouting: true,
		CaseSensitive: true,
		BodyLimit:     1024 * 1024,
		ReadTimeout:   "10s",
		WriteTimeout:  "10s",
		IdleTimeout:   "30s",
		TrustedProxies: []string{
			"127.0.0.1",
			"::1",
		},
	}

	// Verify the fields
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 8080, config.Port)
	assert.True(t, config.Prefork)
	assert.Equal(t, "Goe", config.ServerHeader)
	assert.True(t, config.StrictRouting)
	assert.True(t, config.CaseSensitive)
	assert.Equal(t, 1024*1024, config.BodyLimit)
	assert.Equal(t, "10s", config.ReadTimeout)
	assert.Equal(t, "10s", config.WriteTimeout)
	assert.Equal(t, "30s", config.IdleTimeout)
	assert.Equal(t, []string{"127.0.0.1", "::1"}, config.TrustedProxies)
}

func TestRouteInfo(t *testing.T) {
	// Test the RouteInfo struct
	handler := func(c fiber.Ctx) error { return nil }
	info := contract.RouteInfo{
		Method:  "GET",
		Path:    "/test",
		Name:    "test-route",
		Handler: handler,
	}

	// Verify the fields
	assert.Equal(t, "GET", info.Method)
	assert.Equal(t, "/test", info.Path)
	assert.Equal(t, "test-route", info.Name)
	assert.NotNil(t, info.Handler)
}
