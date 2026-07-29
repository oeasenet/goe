package contract

// HTTPValidator validates a struct according to its `validate` tags.
//
// This is deliberately the same shape as fiber.StructValidator. Fiber owns the
// call site — Ctx.Bind invokes the validator for every JSON, Query, URI, Form,
// Header, Cookie, XML, CBOR and MsgPack binding — so one Validate method is the
// whole contract. See https://docs.gofiber.io/guide/validation.
//
// The parse-and-validate methods that used to live here (ValidateRequest,
// ValidateQuery, ValidateParams, ValidateHeaders, ValidateForm) duplicated
// Ctx.Bind, as did the ValidationProvider and ValidationMiddleware interfaces.
// Use Bind directly — it parses and validates in one step:
//
//	if err := c.Bind().JSON(&dto); err != nil {
//	    return err // dto is parsed and validated
//	}
type HTTPValidator interface {
	// Validate validates a struct according to its tags
	Validate(i any) error
}
