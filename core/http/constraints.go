package http

import (
	"regexp"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// Custom route constraints for GoFiber v3
// These constraints can be used in route definitions like:
//   app.Get("/users/:id<uuid>", handler)
//   app.Get("/posts/:page<uint>", handler)

// UUIDConstraint validates that a route parameter is a valid UUID
type UUIDConstraint struct {
	fiber.CustomConstraint
}

// Name returns the constraint name for use in route definitions
func (*UUIDConstraint) Name() string {
	return "uuid"
}

// Execute validates the parameter value
func (*UUIDConstraint) Execute(param string, args ...string) bool {
	_, err := uuid.Parse(param)
	return err == nil
}

// UIntConstraint validates that a route parameter is a valid unsigned integer
type UIntConstraint struct {
	fiber.CustomConstraint
}

// Name returns the constraint name for use in route definitions
func (*UIntConstraint) Name() string {
	return "uint"
}

// Execute validates the parameter value
func (*UIntConstraint) Execute(param string, args ...string) bool {
	val, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		return false
	}
	return val > 0
}

// SlugConstraint validates that a route parameter is a valid URL slug
// (lowercase alphanumeric with hyphens)
type SlugConstraint struct {
	fiber.CustomConstraint
}

// Name returns the constraint name for use in route definitions
func (*SlugConstraint) Name() string {
	return "slug"
}

// Execute validates the parameter value
func (*SlugConstraint) Execute(param string, args ...string) bool {
	if param == "" {
		return false
	}
	// Slug pattern: lowercase letters, numbers, and hyphens
	// Cannot start or end with hyphen, no consecutive hyphens
	matched, _ := regexp.MatchString(`^[a-z0-9]+(?:-[a-z0-9]+)*$`, param)
	return matched
}

// EmailConstraint validates that a route parameter looks like an email
type EmailConstraint struct {
	fiber.CustomConstraint
}

// Name returns the constraint name for use in route definitions
func (*EmailConstraint) Name() string {
	return "email"
}

// Execute validates the parameter value
func (*EmailConstraint) Execute(param string, args ...string) bool {
	if param == "" {
		return false
	}
	// Simple email validation pattern
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, param)
	return matched
}

// RegisterDefaultConstraints registers all default GOE constraints with the Fiber app
// Call this in your application setup to enable constraint-based route validation:
//
//	app := kernel.App()
//	http.RegisterDefaultConstraints(app)
//	app.Get("/users/:id<uuid>", handler)
//	app.Get("/posts/:page<uint>", handler)
//	app.Get("/articles/:slug<slug>", handler)
func RegisterDefaultConstraints(app *fiber.App) {
	app.RegisterCustomConstraint(&UUIDConstraint{})
	app.RegisterCustomConstraint(&UIntConstraint{})
	app.RegisterCustomConstraint(&SlugConstraint{})
	app.RegisterCustomConstraint(&EmailConstraint{})
}
