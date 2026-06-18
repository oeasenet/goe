package http

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/stretchr/testify/assert"
)

// reqIDBody registers a route that writes the captured request id to the body.
func reqIDApp(withMiddleware bool) *fiber.App {
	app := fiber.New()
	if withMiddleware {
		app.Use(requestid.New())
	}
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("[" + requestIDFrom(c, fiber.HeaderXRequestID) + "]")
	})
	return app
}

func doGet(t *testing.T, app *fiber.App, header string) string {
	t.Helper()
	req := httptest.NewRequest("GET", "/", nil)
	if header != "" {
		req.Header.Set(fiber.HeaderXRequestID, header)
	}
	resp, err := app.Test(req)
	assert.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

func TestRequestIDFrom_Middleware_ReusesUpstream(t *testing.T) {
	// With the middleware, an incoming X-Request-ID is reused, not regenerated.
	got := doGet(t, reqIDApp(true), "upstream-abc")
	assert.Equal(t, "[upstream-abc]", got)
}

func TestRequestIDFrom_Middleware_GeneratesWhenAbsent(t *testing.T) {
	got := doGet(t, reqIDApp(true), "")
	assert.NotEqual(t, "[]", got, "middleware should generate an id when none is supplied")
}

func TestRequestIDFrom_NoMiddleware_FallsBackToHeader(t *testing.T) {
	// Without the middleware, the raw header is still captured (proxy -> backend).
	got := doGet(t, reqIDApp(false), "proxy-123")
	assert.Equal(t, "[proxy-123]", got)
}

func TestRequestIDFrom_NoMiddleware_NoHeader_Empty(t *testing.T) {
	got := doGet(t, reqIDApp(false), "")
	assert.Equal(t, "[]", got)
}

func TestRequestIDFrom_EmptyHeaderArgDefaultsToXRequestID(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("[" + requestIDFrom(c, "") + "]") // empty header -> defaults
	})
	got := doGet(t, app, "viaheader")
	assert.Equal(t, "[viaheader]", got)
}

func TestRequestIDEnabled(t *testing.T) {
	t.Run("default on when unset", func(t *testing.T) {
		c := &MockConfig{}
		c.On("Has", "HTTP_REQUEST_ID").Return(false)
		assert.True(t, requestIDEnabled(c))
	})
	t.Run("explicit true", func(t *testing.T) {
		c := &MockConfig{}
		c.On("Has", "HTTP_REQUEST_ID").Return(true)
		c.On("GetBool", "HTTP_REQUEST_ID").Return(true)
		assert.True(t, requestIDEnabled(c))
	})
	t.Run("explicit false disables", func(t *testing.T) {
		c := &MockConfig{}
		c.On("Has", "HTTP_REQUEST_ID").Return(true)
		c.On("GetBool", "HTTP_REQUEST_ID").Return(false)
		assert.False(t, requestIDEnabled(c))
	})
}
