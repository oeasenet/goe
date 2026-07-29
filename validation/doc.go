// Package validation offers HTTP request validation helpers for the HTTP module.
//
// Deprecated: request validation belongs to Fiber's Bind, which parses and
// validates in one step by calling the fiber.StructValidator that the GOE HTTP
// kernel installs by default:
//
//	if err := c.Bind().JSON(&dto); err != nil {
//	    return err // dto is parsed and validated
//	}
//
// Bind covers JSON, Query, URI, Form, Header, Cookie, XML, CBOR and MsgPack, so
// the middleware and parse-and-validate helpers here duplicate it. See
// https://docs.gofiber.io/guide/validation, and use core/http's
// WithValidatorSetup to register custom rules or WithStructValidator to replace
// the validator entirely.
//
// The validator type itself remains useful for validating structs outside a
// request — that is what Validator.Validate is for. Note that a Validator built
// here is a separate instance from the one the HTTP kernel installs, so rules
// registered on one do not apply to the other; register request rules through
// WithValidatorSetup.
package validation
