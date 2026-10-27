package health

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
)

// countingChecker reports up and counts how often it ran.
type countingChecker struct {
	name string
	runs atomic.Int32
}

func (c *countingChecker) Name() string { return c.name }
func (c *countingChecker) Check(context.Context) contract.HealthCheckResult {
	c.runs.Add(1)
	return contract.HealthCheckResult{Status: contract.HealthStatusUp}
}

func get(t *testing.T, app *fiber.App, path string) (int, healthResponse) {
	t.Helper()
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, path, nil))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var report healthResponse
	require.NoError(t, json.Unmarshal(body, &report), string(body))
	return resp.StatusCode, report
}

func newHealthApp(t *testing.T, opts ...Option) (*Module, *fiber.App) {
	t.Helper()
	module := NewModule(newTestConfig(), &testLogger{t: t}, opts...)
	app := fiber.New()
	module.RegisterRoutes(app)
	return module, app
}

func TestReadiness_ServiceUnavailableOnceStopping(t *testing.T) {
	stopping := make(chan struct{})
	module, app := newHealthApp(t, WithStopping(stopping))
	db := &countingChecker{name: "db"}
	module.Manager().RegisterChecker(db)

	status, _ := get(t, app, "/health/ready")
	require.Equal(t, fiber.StatusOK, status)
	require.Equal(t, int32(1), db.runs.Load())

	close(stopping)
	status, report := get(t, app, "/health/ready")
	assert.Equal(t, fiber.StatusServiceUnavailable, status)
	assert.Equal(t, "down", report.Status)
	require.Contains(t, report.Checks, "shutdown")
	assert.Equal(t, "down", report.Checks["shutdown"].Status)
	assert.Equal(t, "shutting down", report.Checks["shutdown"].Message)
	assert.Equal(t, int32(1), db.runs.Load(), "checkers are not run while stopping")
}

func TestLiveness_StaysUpWhileStopping(t *testing.T) {
	stopping := make(chan struct{})
	_, app := newHealthApp(t, WithStopping(stopping))
	close(stopping)

	status, report := get(t, app, "/health/live")
	assert.Equal(t, fiber.StatusOK, status)
	assert.Equal(t, "up", report.Status)
}

func TestLiveness_DownLivenessCheckerReturns503(t *testing.T) {
	module, app := newHealthApp(t)
	module.Manager().RegisterLivenessChecker(NewCustomChecker("control-loop",
		func(context.Context) contract.HealthCheckResult {
			return contract.HealthCheckResult{Status: contract.HealthStatusDown, Message: "stalled"}
		}))

	status, report := get(t, app, "/health/live")
	assert.Equal(t, fiber.StatusServiceUnavailable, status)
	assert.Equal(t, "down", report.Status)
	require.Contains(t, report.Checks, "control-loop")
	assert.Equal(t, "stalled", report.Checks["control-loop"].Message)
}

func TestLiveness_NoLivenessCheckersIsUp(t *testing.T) {
	module, app := newHealthApp(t)
	module.Manager().RegisterChecker(&mockChecker{name: "db", status: contract.HealthStatusDown})

	status, report := get(t, app, "/health/live")
	assert.Equal(t, fiber.StatusOK, status, "readiness checkers do not affect liveness")
	assert.Equal(t, "up", report.Status)
}
