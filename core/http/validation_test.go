package http

import (
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type signupRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,strong_password"`
	Username string `json:"username" validate:"required,username"`
	Phone    string `json:"phone"    validate:"omitempty,phone"`
}

func asValidationError(t *testing.T, err error) *ValidationError {
	t.Helper()
	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	return ve
}

func TestBundledValidator_UsesJSONTagNames(t *testing.T) {
	v := newStructValidator()

	err := v.Validate(&signupRequest{Password: "Str0ng!pass", Username: "tony"})
	ve := asValidationError(t, err)

	require.NotEmpty(t, ve.Fields)
	// The failing field is reported under its json name, not the Go name.
	assert.Equal(t, "email", ve.Fields[0].Field)
	assert.Equal(t, "required", ve.Fields[0].Tag)
}

func TestBundledValidator_JSONDashFallsBackToFieldName(t *testing.T) {
	type dst struct {
		Secret string `json:"-" validate:"required"`
	}
	v := newStructValidator()

	ve := asValidationError(t, v.Validate(&dst{}))
	require.NotEmpty(t, ve.Fields)
	assert.Equal(t, "Secret", ve.Fields[0].Field)
}

func TestBundledValidator_CustomRules(t *testing.T) {
	v := newStructValidator()

	t.Run("valid values pass", func(t *testing.T) {
		err := v.Validate(&signupRequest{
			Email:    "tony@example.com",
			Password: "Str0ng!pass",
			Username: "tony_an",
			Phone:    "+1 555-0100-99",
		})
		assert.NoError(t, err)
	})

	t.Run("invalid values fail with their tag", func(t *testing.T) {
		err := v.Validate(&signupRequest{
			Email:    "tony@example.com",
			Password: "weak",
			Username: "x",
			Phone:    "abc",
		})
		ve := asValidationError(t, err)

		tags := map[string]string{}
		for _, fe := range ve.Fields {
			tags[fe.Field] = fe.Tag
		}
		assert.Equal(t, "strong_password", tags["password"])
		assert.Equal(t, "username", tags["username"])
		assert.Equal(t, "phone", tags["phone"])
	})
}

func TestValidationError_MessageAndUnwrap(t *testing.T) {
	v := newStructValidator()
	err := v.Validate(&signupRequest{Password: "Str0ng!pass", Username: "tony"})

	// Human-readable flat form for logs and text responses.
	assert.Contains(t, err.Error(), "validation failed")
	assert.Contains(t, err.Error(), "email")

	// The raw go-playground errors stay reachable for advanced callers.
	var raw validator.ValidationErrors
	assert.ErrorAs(t, err, &raw)
}

func TestValidationError_DoesNotEchoSubmittedValues(t *testing.T) {
	// A failed strong_password must never reflect the submitted secret back
	// in the response payload.
	v := newStructValidator()
	err := v.Validate(&signupRequest{
		Email:    "tony@example.com",
		Password: "hunter2secret", // fails strong_password: no upper, no special
		Username: "tony",
	})
	ve := asValidationError(t, err)
	require.Equal(t, "password", ve.Fields[0].Field)

	payload, marshalErr := json.Marshal(ve.Fields)
	require.NoError(t, marshalErr)
	assert.NotContains(t, string(payload), "hunter2secret")
}

func TestBundledValidator_NonStructErrorsPassThrough(t *testing.T) {
	// Programmer errors (validating a non-struct) are not client errors and
	// must not be dressed up as a ValidationError.
	v := newStructValidator()
	err := v.Validate(42)
	require.Error(t, err)
	var ve *ValidationError
	assert.False(t, errors.As(err, &ve))
}

// newErrorHandlerApp builds a fiber app wired exactly like GOE's kernel:
// bundled validator, sonic decoders, and GOE's default error handler.
func newErrorHandlerApp(t *testing.T) *fiber.App {
	t.Helper()
	if tpl == nil {
		parsed, err := template.ParseFS(templateFS, "error_page.gohtml")
		require.NoError(t, err)
		tpl = parsed
	}
	s := defaultSettings(defaultErrorHandler())
	app := fiber.New(s.fiber)

	app.Post("/signup", func(c fiber.Ctx) error {
		var req signupRequest
		if err := c.Bind().JSON(&req); err != nil {
			return err
		}
		return c.JSON(fiber.Map{"ok": true})
	})

	return app
}

func postSignup(t *testing.T, app *fiber.App, body, accept string) (int, string) {
	t.Helper()
	req := httptest.NewRequest("POST", "/signup", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	if accept != "" {
		req.Header.Set(fiber.HeaderAccept, accept)
	}
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, string(b)
}

func TestErrorHandler_ValidationFailureIs400WithMessage(t *testing.T) {
	app := newErrorHandlerApp(t)

	status, body := postSignup(t, app,
		`{"email":"not-an-email","password":"Str0ng!pass","username":"tony"}`,
		fiber.MIMEApplicationJSON)

	assert.Equal(t, fiber.StatusBadRequest, status)

	// The error itself is the message, ready to show to a user.
	var payload struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &payload))
	assert.Equal(t, "email must be a valid email address", payload.Message)
}

func TestErrorHandler_OneErrorAtATime(t *testing.T) {
	// Multiple rules fail, but the response presents only the first, in
	// declaration order — the client fixes it, resubmits, and sees the next.
	app := newErrorHandlerApp(t)

	status, body := postSignup(t, app,
		`{"email":"not-an-email","password":"weakpass","username":"x"}`,
		fiber.MIMEApplicationJSON)

	assert.Equal(t, fiber.StatusBadRequest, status)

	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &payload))
	assert.Equal(t, "email must be a valid email address", payload["message"])
	assert.NotContains(t, payload, "errors")
	assert.NotContains(t, body, "password")
}

func TestErrorHandler_MalformedBodyIs400(t *testing.T) {
	app := newErrorHandlerApp(t)

	status, body := postSignup(t, app, `{"email": not-json`, fiber.MIMEApplicationJSON)

	assert.Equal(t, fiber.StatusBadRequest, status)
	assert.Contains(t, body, "body")
}

func TestErrorHandler_RawValidatorErrorsAre400(t *testing.T) {
	// A replacement validator (WithStructValidator) may return raw
	// go-playground errors; they must still classify as a client error.
	app := newErrorHandlerApp(t)
	app.Get("/raw", func(c fiber.Ctx) error {
		return validator.New().Struct(struct {
			Name string `validate:"required"`
		}{})
	})

	req := httptest.NewRequest("GET", "/raw", nil)
	req.Header.Set(fiber.HeaderAccept, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(b), "required")
}

func TestErrorHandler_OtherErrorsStay500(t *testing.T) {
	app := newErrorHandlerApp(t)
	app.Get("/boom", func(c fiber.Ctx) error {
		return errors.New("boom")
	})

	req := httptest.NewRequest("GET", "/boom", nil)
	req.Header.Set(fiber.HeaderAccept, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestErrorHandler_FiberErrorsKeepTheirStatus(t *testing.T) {
	app := newErrorHandlerApp(t)
	app.Get("/teapot", func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusTeapot, "short and stout")
	})

	req := httptest.NewRequest("GET", "/teapot", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusTeapot, resp.StatusCode)
}

func TestErrorHandler_ValidationFailureHTMLNegotiation(t *testing.T) {
	app := newErrorHandlerApp(t)

	status, body := postSignup(t, app,
		`{"email":"not-an-email","password":"Str0ng!pass","username":"tony"}`,
		fiber.MIMETextHTML)

	assert.Equal(t, fiber.StatusBadRequest, status)
	assert.Contains(t, body, "400")
	// The flat message names the failing field even in the html page.
	assert.Contains(t, body, "email")
}

func TestErrorHandler_QueryValidationIs400(t *testing.T) {
	app := newErrorHandlerApp(t)
	app.Get("/search", func(c fiber.Ctx) error {
		var q struct {
			Limit int `query:"limit" validate:"required,gte=1,lte=100"`
		}
		if err := c.Bind().Query(&q); err != nil {
			return err
		}
		return c.JSON(fiber.Map{"limit": q.Limit})
	})

	req := httptest.NewRequest("GET", "/search?limit=500", nil)
	req.Header.Set(fiber.HeaderAccept, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
