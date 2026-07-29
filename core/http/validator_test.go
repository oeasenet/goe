package http

import (
	"bytes"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type person struct {
	Name string `json:"name" validate:"required"`
	Age  int    `json:"age" validate:"gte=18,lte=60"`
}

// postJSON registers a Bind-based handler and posts body to it.
func postJSON(t *testing.T, k *kernel, body string) *httptest.ResponseRecorder {
	t.Helper()
	k.App().Post("/people", func(c fiber.Ctx) error {
		p := new(person)
		if err := c.Bind().JSON(p); err != nil {
			return err
		}
		return c.SendString("ok:" + p.Name)
	})

	req := httptest.NewRequest("POST", "/people", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := k.App().Test(req)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	rec.Code = resp.StatusCode
	b, _ := io.ReadAll(resp.Body)
	rec.Body = bytes.NewBuffer(b)
	return rec
}

// TestValidator_BindValidatesByDefault is the point of bundling a validator:
// Fiber ships none, and with StructValidator unset Bind parses the body and
// skips validation silently. See https://docs.gofiber.io/guide/validation.
func TestValidator_BindValidatesByDefault(t *testing.T) {
	t.Run("valid payload passes", func(t *testing.T) {
		k, _ := newTestKernel(t, nil)
		rec := postJSON(t, k, `{"name":"Ada","age":36}`)
		assert.Equal(t, fiber.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "ok:Ada")
	})

	t.Run("missing required field is rejected", func(t *testing.T) {
		k, _ := newTestKernel(t, nil)
		rec := postJSON(t, k, `{"age":36}`)
		assert.GreaterOrEqual(t, rec.Code, 400,
			"a payload violating `validate:\"required\"` must not reach the handler")
	})

	t.Run("out-of-range value is rejected", func(t *testing.T) {
		k, _ := newTestKernel(t, nil)
		rec := postJSON(t, k, `{"name":"Ada","age":9}`)
		assert.GreaterOrEqual(t, rec.Code, 400,
			"age=9 violates gte=18 and must be rejected")
	})

	t.Run("the app really has a validator installed", func(t *testing.T) {
		k, _ := newTestKernel(t, nil)
		require.NotNil(t, k.App().Config().StructValidator)
		assert.Error(t, k.App().Config().StructValidator.Validate(&person{}))
		assert.NoError(t, k.App().Config().StructValidator.Validate(&person{Name: "Ada", Age: 36}))
	})
}

func TestValidator_WithValidatorSetup(t *testing.T) {
	t.Run("custom rule is registered and enforced", func(t *testing.T) {
		k, _ := newTestKernel(t, nil, WithValidatorSetup(func(v *validator.Validate) error {
			return v.RegisterValidation("lowercase_only", func(fl validator.FieldLevel) bool {
				return fl.Field().String() == strings.ToLower(fl.Field().String())
			})
		}))
		require.Empty(t, k.optErrs)

		type doc struct {
			Slug string `validate:"lowercase_only"`
		}
		sv := k.App().Config().StructValidator
		assert.NoError(t, sv.Validate(&doc{Slug: "all-lower"}))
		assert.Error(t, sv.Validate(&doc{Slug: "HasUpper"}))
	})

	t.Run("setups apply in order", func(t *testing.T) {
		var order []string
		k, _ := newTestKernel(t, nil,
			WithValidatorSetup(func(*validator.Validate) error { order = append(order, "first"); return nil }),
			WithValidatorSetup(func(*validator.Validate) error { order = append(order, "second"); return nil }),
		)
		require.Empty(t, k.optErrs)
		assert.Equal(t, []string{"first", "second"}, order)
	})

	t.Run("nil setup is rejected", func(t *testing.T) {
		k, _ := newTestKernel(t, nil, WithValidatorSetup(nil))
		require.NotEmpty(t, k.optErrs)
	})

	t.Run("combining with WithStructValidator is rejected", func(t *testing.T) {
		// A setup configures the bundled validator; if that validator has been
		// replaced the setup would silently do nothing, so say so instead.
		k, _ := newTestKernel(t, nil,
			WithStructValidator(&noopValidator{}),
			WithValidatorSetup(func(*validator.Validate) error { return nil }),
		)
		require.NotEmpty(t, k.optErrs)
		assert.Contains(t, errorsJoined(k.optErrs), "WithValidatorSetup")
	})
}

func TestValidator_WithStructValidatorReplaces(t *testing.T) {
	replacement := &noopValidator{}
	k, _ := newTestKernel(t, nil, WithStructValidator(replacement))
	require.Empty(t, k.optErrs)
	assert.Same(t, replacement, k.App().Config().StructValidator,
		"WithStructValidator should replace the bundled validator outright")
}

// noopValidator accepts everything.
type noopValidator struct{}

func (*noopValidator) Validate(any) error { return nil }

func errorsJoined(errs []error) string {
	var b strings.Builder
	for _, err := range errs {
		b.WriteString(err.Error())
		b.WriteString("\n")
	}
	return b.String()
}
