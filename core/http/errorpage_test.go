package http

import (
	"fmt"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

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
	// Create a noop logger for GetLogger
	zapLogger, _ := zap.NewProduction()
	logger.On("GetLogger").Return(zapLogger.Sugar())

	kernel := New(config, logger)
	app := kernel.App()

	// Add test routes that trigger errors
	app.Get("/error", func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadRequest, "Bad Request")
	})

	app.Get("/panic", func(c fiber.Ctx) error {
		panic("Test panic")
	})

	// A handler returning a nil pointer of a concrete error type yields a
	// non-nil error interface whose Error() dereferences a nil receiver.
	app.Get("/typed-nil", func(c fiber.Ctx) error {
		var e *derefError
		return e
	})

	app.Get("/typed-nil-fiber", func(c fiber.Ctx) error {
		var e *fiber.Error
		return e
	})

	app.Get("/wrapped-typed-nil-fiber", func(c fiber.Ctx) error {
		var e *fiber.Error
		return fmt.Errorf("wrapped: %w", e)
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

	// Mirrors the typed-nil hardening Fiber added to its own DefaultErrorHandler
	// in v3.4 (gofiber/fiber#4407, #4372). Without the guard these routes panic
	// inside the error handler instead of rendering a 500.
	for _, tc := range []string{"/typed-nil", "/typed-nil-fiber", "/wrapped-typed-nil-fiber"} {
		t.Run("typed-nil error"+tc, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc+"?format=text", nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, "Internal Server Error", string(body))
		})
	}
}

// derefError is an error whose Error() dereferences its receiver, so calling it
// on a typed-nil value panics.
type derefError struct{ msg string }

func (e *derefError) Error() string { return e.msg }
