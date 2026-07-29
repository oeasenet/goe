package http

import (
	"errors"

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
func (v *structValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

// newStructValidator returns GOE's default validator.
func newStructValidator() *structValidator {
	return &structValidator{validate: validator.New()}
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
