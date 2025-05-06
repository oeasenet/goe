# Validation Module

The Validation module provides a simple interface for validating data in the GOE framework. It integrates with Fiber to provide request validation capabilities.

## Features

- Request data validation
- Custom validation rules
- Error message customization
- Integration with Fiber's context

## Usage

### Basic Validation

The validation module can be used to validate request data in Fiber handlers:

```go
// Import the validation module
import "go.oease.dev/goe/modules/validation"

// Create a Fiber handler with validation
func CreateUserHandler(ctx fiber.Ctx) error {
    // Define the validation rules
    validator := validation.NewFiberValidator(ctx)
    validator.AddRule("name", "required|min:3|max:50")
    validator.AddRule("email", "required|email")
    validator.AddRule("age", "required|numeric|min:18")
    
    // Validate the request
    if !validator.Validate() {
        // Return validation errors
        return webresult.ValidationFailed(ctx, validator.Errors())
    }
    
    // Process the validated data
    name := validator.GetString("name")
    email := validator.GetString("email")
    age := validator.GetInt("age")
    
    // ... create user logic ...
    
    return webresult.SendSucceed(ctx, "User created successfully")
}
```

### Custom Validation Messages

You can customize the validation error messages:

```go
validator := validation.NewFiberValidator(ctx)
validator.AddRule("name", "required|min:3|max:50")
validator.AddRule("email", "required|email")

// Add custom error messages
validator.AddMessage("name.required", "Please provide your name")
validator.AddMessage("name.min", "Name must be at least 3 characters")
validator.AddMessage("email.email", "Please provide a valid email address")

// Validate
if !validator.Validate() {
    return webresult.ValidationFailed(ctx, validator.Errors())
}
```

### Accessing Validated Data

After validation, you can access the validated data:

```go
// Get string value
name := validator.GetString("name")

// Get int value
age := validator.GetInt("age")

// Get float value
score := validator.GetFloat("score")

// Get bool value
active := validator.GetBool("active")

// Get time value
createdAt := validator.GetTime("created_at")

// Get array value
tags := validator.GetStringArray("tags")
```

### Available Validation Rules

The validation module supports many validation rules, including:

- `required`: Field must be present and not empty
- `email`: Field must be a valid email address
- `url`: Field must be a valid URL
- `numeric`: Field must be a number
- `alpha`: Field must contain only letters
- `alphanumeric`: Field must contain only letters and numbers
- `min`: Minimum value (for numbers) or length (for strings)
- `max`: Maximum value (for numbers) or length (for strings)
- `in`: Field must be one of the specified values
- `not_in`: Field must not be one of the specified values
- `regex`: Field must match the regular expression
- `date`: Field must be a valid date
- `before`: Date must be before the specified date
- `after`: Date must be after the specified date

## Implementation Details

The validation module is built on top of the [Gookit Validate](https://github.com/gookit/validate) library and provides a simplified interface for use with Fiber.

The module integrates with Fiber's context to automatically parse and validate request data from various sources (query parameters, form data, JSON body, etc.).