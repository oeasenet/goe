package main

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/http"
	"go.oease.dev/goe/v2/core/log"
)

// User represents a user registration request
type User struct {
	Username string `json:"username" validate:"required,min=3,max=20,alphanum"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,containsany=!@#$%^&*"`
	Age      int    `json:"age" validate:"required,min=18,max=120"`
	Phone    string `json:"phone" validate:"required,e164|customphone"`
	Website  string `json:"website" validate:"omitempty,url"`
	Bio      string `json:"bio" validate:"max=500"`
}

// Address represents an address with custom validation
type Address struct {
	Street     string `json:"street" validate:"required"`
	City       string `json:"city" validate:"required"`
	State      string `json:"state" validate:"required,len=2,alpha"`
	PostalCode string `json:"postal_code" validate:"required,uspostal"`
	Country    string `json:"country" validate:"required,iso3166_1_alpha2"`
}

func main() {
	// Create the application
	_ = goe.New(goe.Options{
		WithHTTP: true,
		Invokers: []any{
			setupRoutes,
			registerCustomValidations,
		},
	})

	goe.Log().Info("Validation demo server starting",
		log.NewField("host", goe.Config().GetString("HTTP_HOST")),
		log.NewField("port", goe.Config().GetInt("HTTP_PORT")),
	)

	// Run the application
	goe.Run()
}

func registerCustomValidations(httpKernel contract.HTTPKernel) {
	// Get the validator
	customValidator := httpKernel.Validator().(*http.CustomValidator)

	// Register custom phone validation (alternative format)
	customValidator.RegisterValidation("customphone", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		// Simple validation: 10-15 digits, can start with +
		if strings.HasPrefix(phone, "+") {
			phone = phone[1:]
		}
		if len(phone) < 10 || len(phone) > 15 {
			return false
		}
		for _, ch := range phone {
			if ch < '0' || ch > '9' {
				return false
			}
		}
		return true
	})

	// Register US postal code validation
	customValidator.RegisterValidation("uspostal", func(fl validator.FieldLevel) bool {
		postal := fl.Field().String()
		// Simple validation: 5 digits or 5+4 format
		if len(postal) == 5 {
			for _, ch := range postal {
				if ch < '0' || ch > '9' {
					return false
				}
			}
			return true
		}
		if len(postal) == 10 && postal[5] == '-' {
			for i, ch := range postal {
				if i == 5 {
					continue
				}
				if ch < '0' || ch > '9' {
					return false
				}
			}
			return true
		}
		return false
	})

	// Register struct-level validation
	customValidator.RegisterStructValidation(func(sl validator.StructLevel) {
		user := sl.Current().Interface().(User)

		// Example: Premium users (with website) must have a bio
		if user.Website != "" && user.Bio == "" {
			sl.ReportError(user.Bio, "bio", "Bio", "required_with_website", "")
		}

		// Example: Users under 21 cannot have certain usernames
		if user.Age < 21 && strings.Contains(strings.ToLower(user.Username), "admin") {
			sl.ReportError(user.Username, "username", "Username", "no_admin_for_minors", "")
		}
	}, User{})

	goe.Log().Info("Custom validations registered")
}

func setupRoutes(httpKernel contract.HTTPKernel) {
	app := httpKernel.App()

	// Welcome route
	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to Goe Validation Demo",
			"endpoints": fiber.Map{
				"POST /users":     "Create a new user with validation",
				"POST /addresses": "Create an address with validation",
				"GET /schema":     "Get validation schema information",
			},
		})
	})

	// User registration with validation
	app.Post("/users", func(c fiber.Ctx) error {
		logger := http.GetLogger(c)
		validator := http.GetValidator(c)

		var user User
		// Parse body
		if err := c.Bind().JSON(&user); err != nil {
			logger.Warn("Invalid JSON in request", log.NewField("error", err.Error()))
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid JSON format",
			})
		}

		// Validate
		if err := validator.Validate(user); err != nil {
			logger.Info("Validation failed", log.NewField("error", err.Error()))

			// Parse validation errors for better response
			validationErrors := make(map[string]string)
			if errs, ok := err.(validator.ValidationErrors); ok {
				for _, e := range errs {
					field := strings.ToLower(e.Field())
					switch e.Tag() {
					case "required":
						validationErrors[field] = field + " is required"
					case "min":
						validationErrors[field] = field + " must be at least " + e.Param() + " characters"
					case "max":
						validationErrors[field] = field + " must be at most " + e.Param() + " characters"
					case "email":
						validationErrors[field] = field + " must be a valid email address"
					case "alphanum":
						validationErrors[field] = field + " must contain only letters and numbers"
					case "containsany":
						validationErrors[field] = field + " must contain at least one special character"
					case "customphone", "e164":
						validationErrors[field] = field + " must be a valid phone number"
					case "url":
						validationErrors[field] = field + " must be a valid URL"
					case "required_with_website":
						validationErrors[field] = "bio is required when website is provided"
					case "no_admin_for_minors":
						validationErrors[field] = "users under 21 cannot have 'admin' in username"
					default:
						validationErrors[field] = field + " validation failed: " + e.Tag()
					}
				}
			}

			return c.Status(400).JSON(fiber.Map{
				"error":   "Validation failed",
				"details": validationErrors,
			})
		}

		logger.Info("User registration successful",
			log.NewField("username", user.Username),
			log.NewField("email", user.Email),
		)

		return c.Status(201).JSON(fiber.Map{
			"message": "User created successfully",
			"user": fiber.Map{
				"username": user.Username,
				"email":    user.Email,
				"age":      user.Age,
				"phone":    user.Phone,
				"website":  user.Website,
			},
		})
	})

	// Address validation example
	app.Post("/addresses", func(c fiber.Ctx) error {
		validator := http.GetValidator(c)

		var address Address
		if err := c.Bind().JSON(&address); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid JSON format",
			})
		}

		if err := validator.Validate(address); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error":   "Validation failed",
				"details": err.Error(),
			})
		}

		return c.Status(201).JSON(fiber.Map{
			"message": "Address created successfully",
			"address": address,
		})
	})

	// Get validation schema information
	app.Get("/schema", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_schema": fiber.Map{
				"username": "required, 3-20 chars, alphanumeric only",
				"email":    "required, valid email format",
				"password": "required, min 8 chars, must contain special character",
				"age":      "required, 18-120",
				"phone":    "required, E.164 format or 10-15 digits",
				"website":  "optional, valid URL",
				"bio":      "max 500 chars, required if website is provided",
			},
			"address_schema": fiber.Map{
				"street":      "required",
				"city":        "required",
				"state":       "required, 2 letter state code",
				"postal_code": "required, US postal code (12345 or 12345-6789)",
				"country":     "required, ISO 3166-1 alpha-2 country code",
			},
			"custom_validations": []string{
				"customphone: Alternative phone validation",
				"uspostal: US postal code validation",
				"struct: Bio required with website",
				"struct: No 'admin' in username for users under 21",
			},
		})
	})

	// Example valid requests
	app.Get("/examples", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"valid_user": fiber.Map{
				"username": "johndoe123",
				"email":    "john@example.com",
				"password": "SecureP@ss123",
				"age":      25,
				"phone":    "+14155552671",
				"website":  "https://example.com",
				"bio":      "Software developer from San Francisco",
			},
			"valid_address": fiber.Map{
				"street":      "123 Main St",
				"city":        "San Francisco",
				"state":       "CA",
				"postal_code": "94105",
				"country":     "US",
			},
		})
	})
}
