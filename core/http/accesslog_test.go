package http

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// newAccessLogKernel builds a kernel whose access log writes into an in-memory
// observer, so tests can assert the exact fields fiberzap emits per request.
func newAccessLogKernel(t *testing.T) (contract.HTTPKernel, *observer.ObservedLogs) {
	t.Helper()

	config := &MockConfig{}
	config.On("GetString", mock.Anything).Return("")
	config.On("GetBool", mock.Anything).Return(false)
	config.On("GetInt", mock.Anything).Return(0)
	config.On("GetStringSlice", mock.Anything).Return([]string{})
	config.On("GetDuration", mock.Anything).Return(time.Duration(0))
	config.On("Has", mock.Anything).Return(false)

	core, logs := observer.New(zap.DebugLevel)
	logger := &MockLogger{}
	logger.On("Fatal", mock.Anything).Return()
	logger.On("GetLogger").Return(zap.New(core).Sugar())

	return New(config, logger), logs
}

func TestAccessLog_ContractFields(t *testing.T) {
	kernel, logs := newAccessLogKernel(t)
	app := kernel.App()
	app.Get("/users/:id", func(c fiber.Ctx) error {
		return c.SendString("hello!")
	})

	req := httptest.NewRequest("GET", "/users/42?page=1", nil)
	req.Host = "api.example.com"
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	entries := logs.FilterMessage("http request").All()
	require.Len(t, entries, 1)
	fields := entries[0].ContextMap()

	// New contract fields.
	assert.Equal(t, "/users/:id", fields["route"], "route must be the template, not the concrete URL")
	assert.Equal(t, "api.example.com", fields["host"])
	assert.EqualValues(t, len("hello!"), fields["bytesSent"])

	durationMS, ok := fields["duration_ms"].(int64)
	require.True(t, ok, "duration_ms must be an integer, got %T", fields["duration_ms"])
	assert.GreaterOrEqual(t, durationMS, int64(0))
	assert.Less(t, durationMS, int64(60_000), "duration_ms must be measured from the request start, not the zero time")

	// Existing fields must survive.
	assert.Equal(t, "GET", fields["method"])
	assert.EqualValues(t, fiber.StatusOK, fields["status"])
	assert.Contains(t, fields, "latency")
	assert.Equal(t, "/users/42?page=1", fields["url"])
	assert.Contains(t, fields, "ip")
	assert.Contains(t, fields, "request_id")
}
