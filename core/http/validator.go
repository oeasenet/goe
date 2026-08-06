package http

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// structValidator adapts go-playground/validator to fiber.StructValidator.
//
// This is the pattern documented at https://docs.gofiber.io/guide/validation.
// Fiber defines the one-method interface and calls it from every Bind method,
// but ships no validator of its own: with StructValidator unset,
// c.Bind().JSON(&dto) parses the body and skips validation silently. GOE
// installs this by default so `validate` struct tags work out of the box.
//
// Validation belongs to the HTTP kernel rather than a separate injected service
// because Fiber owns the call site — Bind invokes it, not application code.
type structValidator struct {
	validate *validator.Validate
}

// Validate implements fiber.StructValidator.
//
// Fiber only calls this for struct destinations; binding into a map or other
// non-struct skips it.
//
// Failed rules come back as *ValidationError, which the default error handler
// renders as an HTTP 400 with structured field errors. Anything else — such as
// go-playground's InvalidValidationError for non-struct input — is a
// programmer error and passes through untouched, so it keeps reading as the
// server fault it is.
func (v *structValidator) Validate(out any) error {
	err := v.validate.Struct(out)

	if fieldErrs, ok := errors.AsType[validator.ValidationErrors](err); ok {
		return newValidationError(fieldErrs)
	}
	return err
}

// newStructValidator returns GOE's default validator: field names come from
// json tags, and the bundled custom rules are pre-registered.
func newStructValidator() *structValidator {
	v := validator.New()

	// Report fields under their json names, so a client that sent {"email"}
	// is told about "email", not "Email". A json:"-" field falls back to the
	// Go field name.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	registerBundledRules(v)

	return &structValidator{validate: v}
}

// registerBundledRules adds GOE's custom validation tags. Override any of them
// with WithValidatorSetup: re-registering a tag replaces it.
//
// RegisterValidation only errors on an empty tag name, which is impossible
// here, so the returned errors are discarded.
func registerBundledRules(v *validator.Validate) {
	// phone: permissive international phone shape — digits with optional
	// +, -, and spaces, 10 to 15 characters.
	_ = v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		if len(phone) < 10 || len(phone) > 15 {
			return false
		}
		for _, r := range phone {
			if r != '+' && r != '-' && r != ' ' && (r < '0' || r > '9') {
				return false
			}
		}
		return true
	})

	// username: 3-30 characters, alphanumeric and underscore.
	_ = v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		username := fl.Field().String()
		if len(username) < 3 || len(username) > 30 {
			return false
		}
		for _, r := range username {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
				return false
			}
		}
		return true
	})

	// strong_password: at least 8 characters with upper, lower, digit and
	// special character.
	_ = v.RegisterValidation("strong_password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		if len(password) < 8 {
			return false
		}

		var hasUpper, hasLower, hasNumber, hasSpecial bool
		for _, r := range password {
			switch {
			case r >= 'A' && r <= 'Z':
				hasUpper = true
			case r >= 'a' && r <= 'z':
				hasLower = true
			case r >= '0' && r <= '9':
				hasNumber = true
			case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", r):
				hasSpecial = true
			}
		}

		return hasUpper && hasLower && hasNumber && hasSpecial
	})
}

// ValidatorSetup customises the bundled validator, for registering custom rules
// or tag name functions. Returning an error aborts startup.
type ValidatorSetup func(*validator.Validate) error

// WithValidatorSetup registers custom rules on the bundled validator.
//
// Use this when the defaults are fine but you need an extra rule or two —
// registering one validation should not require reimplementing the adapter:
//
//	goehttp.WithValidatorSetup(func(v *validator.Validate) error {
//	    return v.RegisterValidation("slug", isSlug)
//	}),
//
// Setups run in the order given, during kernel construction, so a rule is
// available before the first request. To replace validation wholesale — a
// different library, or entirely custom behaviour — use WithStructValidator
// instead; combining the two is rejected, since a setup for the bundled
// validator cannot apply to a replacement that does not use it.
func WithValidatorSetup(setup ValidatorSetup) Option {
	return func(s *settings) error {
		if setup == nil {
			return errValidatorSetupNil
		}
		s.validatorSetups = append(s.validatorSetups, setup)
		return nil
	}
}

// errValidatorSetupNil is returned when WithValidatorSetup is given no function.
var errValidatorSetupNil = errors.New("WithValidatorSetup: setup must not be nil")
