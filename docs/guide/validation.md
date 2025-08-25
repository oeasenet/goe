# HTTP Validation

GOE provides a powerful and flexible validation system for HTTP requests built on top of the [go-playground/validator](https://github.com/go-playground/validator) library. The validation package offers struct validation, custom validators, middleware support, and seamless integration with Fiber.

## Overview

The validation system in GOE is designed to:
- **Validate request data**: Body, query parameters, URL parameters, headers, and forms
- **Provide clear error messages**: Human-readable validation errors with field details
- **Support custom validators**: Extend with your own validation logic
- **Integrate seamlessly**: Works perfectly with Fiber and GOE's dependency injection

## Installation & Setup

The validation package is included with GOE and is automatically available when you enable the HTTP module:

```go
package main

import (
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/validation"
)

func main() {
    goe.New(goe.Options{
        WithHTTP: true,
    })
    
    // Validation is automatically available
    goe.Run()
}
```

## Basic Usage

### Struct Validation

Define your structs with validation tags:

```go
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"required,gte=18,lte=120"`
    Password string `json:"password" validate:"required,min=8"`
}
```

### Manual Validation in Handlers

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/validation"
)

func main() {
    goe.New(goe.Options{WithHTTP: true})
    
    app := goe.HTTP().App()
    validator := validation.New()
    
    app.Post("/users", func(c fiber.Ctx) error {
        var req CreateUserRequest
        
        // Parse and validate request body
        if err := validator.ValidateRequest(c, &req); err != nil {
            if ve, ok := err.(validation.ValidationError); ok {
                return c.Status(400).JSON(fiber.Map{
                    "error": "Validation failed",
                    "details": ve.Errors,
                })
            }
            return c.Status(400).JSON(fiber.Map{
                "error": err.Error(),
            })
        }
        
        // Process valid request
        return c.JSON(fiber.Map{
            "message": "User created",
            "user": req,
        })
    })
    
    goe.Run()
}
```

### Using Validation from Context

When using GOE's HTTP module, the validator is automatically available in the context:

```go
func CreateUserHandler(c fiber.Ctx) error {
    validator := http.GetValidator(c)
    
    var req CreateUserRequest
    if err := validator.ValidateRequest(c, &req); err != nil {
        return err // Error handler will format the response
    }
    
    // Process valid request
    return c.JSON(req)
}
```

## Validation Methods

The validation package provides several methods for different types of data:

### ValidateRequest
Parses and validates JSON request body:

```go
var req LoginRequest
err := validator.ValidateRequest(c, &req)
```

### ValidateQuery
Parses and validates query parameters:

```go
type PaginationQuery struct {
    Page  int `query:"page" validate:"min=1"`
    Limit int `query:"limit" validate:"min=1,max=100"`
}

var query PaginationQuery
err := validator.ValidateQuery(c, &query)
```

### ValidateParams
Parses and validates URL parameters:

```go
type UserParams struct {
    ID string `params:"id" validate:"required,uuid"`
}

var params UserParams
err := validator.ValidateParams(c, &params)
```

### ValidateHeaders
Parses and validates request headers:

```go
type AuthHeaders struct {
    Authorization string `header:"Authorization" validate:"required"`
    APIKey        string `header:"X-API-Key" validate:"required,len=32"`
}

var headers AuthHeaders
err := validator.ValidateHeaders(c, &headers)
```

### ValidateForm
Parses and validates form data:

```go
type UploadForm struct {
    Title       string `form:"title" validate:"required,max=100"`
    Description string `form:"description" validate:"max=500"`
}

var form UploadForm
err := validator.ValidateForm(c, &form)
```

## Validation Middleware

Use middleware for automatic validation:

### Global Validation Middleware

```go
app := goe.HTTP().App()

// Add validation middleware globally
app.Use(validation.NewMiddleware())

// Now validator is available in all handlers
app.Post("/api/users", func(c fiber.Ctx) error {
    validator := validation.GetValidator(c)
    // Use validator...
})
```

### Route-Specific Validation

```go
// Validate body for specific route
app.Post("/users", 
    validation.ValidateBody(&CreateUserRequest{}),
    func(c fiber.Ctx) error {
        // Get validated data from context
        req := validation.GetValidatedBody(c).(*CreateUserRequest)
        
        // Process validated request
        return c.JSON(req)
    },
)

// Validate query parameters
app.Get("/users",
    validation.ValidateQuery(&PaginationQuery{}),
    func(c fiber.Ctx) error {
        query := validation.GetValidatedQuery(c).(*PaginationQuery)
        
        // Use validated query params
        return c.JSON(fiber.Map{
            "page": query.Page,
            "limit": query.Limit,
        })
    },
)
```

## Common Validation Tags

GOE supports all standard go-playground/validator tags:

### Required & Optional
- `required` - Field must be present and not zero value
- `omitempty` - Field can be omitted or empty

### String Validation
- `min=n` - Minimum length
- `max=n` - Maximum length
- `len=n` - Exact length
- `email` - Valid email address
- `url` - Valid URL
- `uri` - Valid URI
- `alpha` - Alphabetic characters only
- `alphanum` - Alphanumeric characters only
- `numeric` - Numeric string
- `hexadecimal` - Hexadecimal string
- `hexcolor` - Hex color string
- `rgb` - RGB color
- `rgba` - RGBA color
- `uuid` - Valid UUID
- `uuid3` - Valid UUID v3
- `uuid4` - Valid UUID v4
- `uuid5` - Valid UUID v5
- `ascii` - ASCII characters only
- `base64` - Valid base64
- `json` - Valid JSON

### Numeric Validation
- `gt=n` - Greater than
- `gte=n` - Greater than or equal
- `lt=n` - Less than
- `lte=n` - Less than or equal
- `eq=n` - Equal to
- `ne=n` - Not equal to

### Network Validation
- `ip` - Valid IP address
- `ipv4` - Valid IPv4
- `ipv6` - Valid IPv6
- `cidr` - Valid CIDR notation
- `cidrv4` - Valid IPv4 CIDR
- `cidrv6` - Valid IPv6 CIDR
- `mac` - Valid MAC address
- `hostname` - Valid hostname
- `fqdn` - Valid FQDN

### Date & Time
- `datetime=2006-01-02` - Valid datetime in specified format

### Comparison
- `eqfield=Field` - Equal to another field
- `nefield=Field` - Not equal to another field
- `gtfield=Field` - Greater than another field
- `gtefield=Field` - Greater than or equal to another field
- `ltfield=Field` - Less than another field
- `ltefield=Field` - Less than or equal to another field

### Other
- `contains=substring` - Contains substring
- `containsany=chars` - Contains any of the characters
- `excludes=substring` - Excludes substring
- `excludesall=chars` - Excludes all characters
- `startswith=prefix` - Starts with prefix
- `endswith=suffix` - Ends with suffix
- `oneof=val1 val2` - One of the specified values

## Custom Validators

GOE includes several built-in custom validators:

### phone
Validates phone numbers:
```go
type Contact struct {
    Phone string `json:"phone" validate:"phone"`
}
```

### username
Validates usernames (3-30 chars, alphanumeric and underscore):
```go
type User struct {
    Username string `json:"username" validate:"username"`
}
```

### strong_password
Validates password strength (min 8 chars, uppercase, lowercase, number, special char):
```go
type Security struct {
    Password string `json:"password" validate:"strong_password"`
}
```

### Adding Custom Validators

Register your own validators:

```go
validator := validation.New()

// Register custom postal code validator
validator.RegisterValidation("postalcode", func(fl validator.FieldLevel) bool {
    code := fl.Field().String()
    // US postal code validation
    matched, _ := regexp.MatchString(`^\d{5}(-\d{4})?$`, code)
    return matched
})

// Use in struct
type Address struct {
    PostalCode string `json:"postal_code" validate:"required,postalcode"`
}
```

## Error Handling

Validation errors provide detailed information about what failed:

```go
app.Post("/users", func(c fiber.Ctx) error {
    var req CreateUserRequest
    
    if err := validator.ValidateRequest(c, &req); err != nil {
        if ve, ok := err.(validation.ValidationError); ok {
            // Structured validation errors
            return c.Status(400).JSON(fiber.Map{
                "error": "Validation failed",
                "details": ve.Errors, // Array of field errors
            })
        }
        
        if pe, ok := err.(validation.ParseError); ok {
            // Parse errors (malformed JSON, etc.)
            return c.Status(400).JSON(fiber.Map{
                "error": pe.Message,
            })
        }
        
        // Other errors
        return c.Status(500).JSON(fiber.Map{
            "error": "Internal server error",
        })
    }
    
    return c.JSON(req)
})
```

Example error response:
```json
{
    "error": "Validation failed",
    "details": [
        {
            "field": "email",
            "value": "invalid",
            "tag": "email",
            "message": "email must be a valid email address"
        },
        {
            "field": "age",
            "value": 15,
            "tag": "gte",
            "message": "age must be greater than or equal to 18"
        }
    ]
}
```

## Helper Functions

The validation package includes helper functions for common validations:

```go
// Email validation
if validation.IsEmail("user@example.com") {
    // Valid email
}

// URL validation
if validation.IsURL("https://example.com") {
    // Valid URL
}

// UUID validation
if validation.IsUUID("550e8400-e29b-41d4-a716-446655440000") {
    // Valid UUID
}

// Alpha validation
if validation.IsAlpha("abcDEF") {
    // Contains only letters
}

// Alphanumeric validation
if validation.IsAlphanumeric("abc123") {
    // Contains only letters and numbers
}

// Numeric validation
if validation.IsNumeric("12345") {
    // Contains only numbers
}
```

## Best Practices

### 1. Use Struct Tags
Define validation rules using struct tags for clarity and reusability:

```go
type User struct {
    ID       int    `json:"id" validate:"required,min=1"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"omitempty,gte=0,lte=120"`
    Status   string `json:"status" validate:"required,oneof=active inactive pending"`
}
```

### 2. Create Request/Response Types
Separate your domain models from HTTP request/response types:

```go
// Request type with validation
type CreateProductRequest struct {
    Name        string  `json:"name" validate:"required,min=3,max=100"`
    Description string  `json:"description" validate:"max=500"`
    Price       float64 `json:"price" validate:"required,gt=0"`
    Stock       int     `json:"stock" validate:"required,gte=0"`
}

// Domain model
type Product struct {
    ID          int
    Name        string
    Description string
    Price       float64
    Stock       int
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 3. Use Middleware for Common Routes
Apply validation middleware to route groups:

```go
api := app.Group("/api", validation.NewMiddleware())

v1 := api.Group("/v1")
v1.Post("/users", createUser)
v1.Put("/users/:id", updateUser)
```

### 4. Custom Error Messages
Provide user-friendly error messages:

```go
func handleValidationError(c fiber.Ctx, err error) error {
    if ve, ok := err.(validation.ValidationError); ok {
        messages := make([]string, 0, len(ve.Errors))
        for _, e := range ve.Errors {
            messages = append(messages, e.Message)
        }
        
        return c.Status(400).JSON(fiber.Map{
            "success": false,
            "message": "Please correct the following errors",
            "errors":  messages,
        })
    }
    
    return c.Status(500).JSON(fiber.Map{
        "success": false,
        "message": "An unexpected error occurred",
    })
}
```

### 5. Validate Early
Validate input as early as possible in your handler:

```go
func CreateOrderHandler(c fiber.Ctx) error {
    // Validate first
    var req CreateOrderRequest
    if err := validator.ValidateRequest(c, &req); err != nil {
        return handleValidationError(c, err)
    }
    
    // Then process
    order, err := orderService.Create(req)
    if err != nil {
        return handleServiceError(c, err)
    }
    
    return c.Status(201).JSON(order)
}
```

## Examples

### Complete User Registration Example

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/validation"
)

type RegisterRequest struct {
    Username        string `json:"username" validate:"required,username"`
    Email           string `json:"email" validate:"required,email"`
    Password        string `json:"password" validate:"required,strong_password"`
    ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
    Age             int    `json:"age" validate:"required,gte=18"`
    AcceptTerms     bool   `json:"accept_terms" validate:"required,eq=true"`
}

func main() {
    goe.New(goe.Options{WithHTTP: true})
    
    app := goe.HTTP().App()
    validator := validation.New()
    
    app.Post("/register", func(c fiber.Ctx) error {
        var req RegisterRequest
        
        // Validate request
        if err := validator.ValidateRequest(c, &req); err != nil {
            if ve, ok := err.(validation.ValidationError); ok {
                return c.Status(400).JSON(fiber.Map{
                    "success": false,
                    "errors":  ve.Errors,
                })
            }
            return c.Status(400).JSON(fiber.Map{
                "success": false,
                "message": err.Error(),
            })
        }
        
        // Registration logic here...
        
        return c.JSON(fiber.Map{
            "success": true,
            "message": "Registration successful",
            "user": fiber.Map{
                "username": req.Username,
                "email":    req.Email,
            },
        })
    })
    
    goe.Run()
}
```

### API with Pagination and Filtering

```go
type ListUsersQuery struct {
    Page     int    `query:"page" validate:"min=1"`
    Limit    int    `query:"limit" validate:"min=1,max=100"`
    SortBy   string `query:"sort_by" validate:"omitempty,oneof=name email created_at"`
    Order    string `query:"order" validate:"omitempty,oneof=asc desc"`
    Status   string `query:"status" validate:"omitempty,oneof=active inactive all"`
    Search   string `query:"search" validate:"omitempty,max=100"`
}

app.Get("/users", func(c fiber.Ctx) error {
    var query ListUsersQuery
    
    // Set defaults
    query.Page = 1
    query.Limit = 20
    query.Order = "asc"
    query.Status = "all"
    
    // Validate query parameters
    if err := validator.ValidateQuery(c, &query); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid query parameters",
            "details": err.Error(),
        })
    }
    
    // Fetch users with validated parameters
    users, total, err := userService.List(query)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to fetch users",
        })
    }
    
    return c.JSON(fiber.Map{
        "users": users,
        "pagination": fiber.Map{
            "page":  query.Page,
            "limit": query.Limit,
            "total": total,
        },
    })
})
```

### File Upload with Validation

```go
type FileUploadForm struct {
    Title       string `form:"title" validate:"required,max=100"`
    Description string `form:"description" validate:"max=500"`
    Category    string `form:"category" validate:"required,oneof=document image video"`
}

app.Post("/upload", func(c fiber.Ctx) error {
    var form FileUploadForm
    
    // Validate form data
    if err := validator.ValidateForm(c, &form); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid form data",
            "details": err.Error(),
        })
    }
    
    // Get uploaded file
    file, err := c.FormFile("file")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "File is required",
        })
    }
    
    // Validate file size (10MB max)
    if file.Size > 10*1024*1024 {
        return c.Status(400).JSON(fiber.Map{
            "error": "File size must not exceed 10MB",
        })
    }
    
    // Process upload...
    
    return c.JSON(fiber.Map{
        "success": true,
        "message": "File uploaded successfully",
        "file": fiber.Map{
            "title":    form.Title,
            "category": form.Category,
            "size":     file.Size,
        },
    })
})
```

## Contracts and Interfaces

GOE provides well-defined contracts for validation components to ensure loose coupling and testability:

### HTTPValidator Contract

The `contract.HTTPValidator` interface defines the core validation operations:

```go
type HTTPValidator interface {
    // Validate validates a struct according to its tags
    Validate(i interface{}) error
    
    // ValidateRequest validates and parses JSON request body
    ValidateRequest(c fiber.Ctx, dst interface{}) error
    
    // ValidateQuery validates and parses query parameters
    ValidateQuery(c fiber.Ctx, dst interface{}) error
    
    // ValidateParams validates and parses URL parameters
    ValidateParams(c fiber.Ctx, dst interface{}) error
    
    // ValidateHeaders validates and parses request headers
    ValidateHeaders(c fiber.Ctx, dst interface{}) error
    
    // ValidateForm validates and parses form data
    ValidateForm(c fiber.Ctx, dst interface{}) error
}
```

### ValidationProvider Contract

For dependency injection scenarios:

```go
type ValidationProvider interface {
    // GetValidator returns an HTTP validator instance
    GetValidator() HTTPValidator
}
```

### ValidationMiddleware Contract

For middleware providers:

```go
type ValidationMiddleware interface {
    // ValidateBody creates middleware that validates request body
    ValidateBody(dst interface{}) fiber.Handler
    
    // ValidateQuery creates middleware that validates query parameters
    ValidateQuery(dst interface{}) fiber.Handler
    
    // ValidateParams creates middleware that validates URL parameters
    ValidateParams(dst interface{}) fiber.Handler
}
```

### Using Contracts

```go
// HTTP Kernel provides validator through contract
httpKernel := goe.HTTP()
validator := httpKernel.HTTPValidator()

// Or use the validation provider
provider := validation.NewProvider()
httpValidator := provider.GetValidator()

// Middleware provider
middlewareProvider := validation.NewMiddlewareProvider()
bodyHandler := middlewareProvider.ValidateBody(&CreateUserRequest{})
```

## Integration with Dependency Injection

When using GOE's dependency injection system:

```go
package main

import (
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
    "go.oease.dev/goe/v2/validation"
    "github.com/gofiber/fiber/v3"
)

type UserController struct {
    validator contract.HTTPValidator
    logger    contract.Logger
    db        contract.DB
}

func NewUserController(
    httpKernel contract.HTTPKernel,
    logger contract.Logger,
    db contract.DB,
) *UserController {
    return &UserController{
        validator: httpKernel.HTTPValidator(),
        logger:    logger,
        db:        db,
    }
}

// Alternative using validation provider
func NewUserControllerWithProvider(
    validationProvider contract.ValidationProvider,
    logger contract.Logger,
    db contract.DB,
) *UserController {
    return &UserController{
        validator: validationProvider.GetValidator(),
        logger:    logger,
        db:        db,
    }
}

func (uc *UserController) Create(c fiber.Ctx) error {
    var req CreateUserRequest
    
    if err := uc.validator.ValidateRequest(c, &req); err != nil {
        uc.logger.Warn("Validation failed", "error", err)
        return c.Status(400).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    
    // Create user in database
    user, err := uc.createUser(req)
    if err != nil {
        uc.logger.Error("Failed to create user", "error", err)
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to create user",
        })
    }
    
    return c.Status(201).JSON(user)
}

func main() {
    goe.New(goe.Options{
        WithHTTP: true,
        WithDB:   true,
        Providers: []any{
            NewUserController,
            // Or use validation provider approach
            validation.NewProvider,
            // NewUserControllerWithProvider,
        },
        Invokers: []any{
            func(
                httpKernel contract.HTTPKernel,
                controller *UserController,
            ) {
                app := httpKernel.App()
                app.Post("/users", controller.Create)
            },
        },
    })
    
    goe.Run()
}
```

## Testing Validation

Write tests for your validation logic:

```go
package main

import (
    "testing"
    "go.oease.dev/goe/v2/validation"
)

func TestUserValidation(t *testing.T) {
    v := validation.New()
    
    tests := []struct {
        name    string
        user    User
        wantErr bool
    }{
        {
            name: "valid user",
            user: User{
                Name:     "John Doe",
                Email:    "john@example.com",
                Age:      25,
                Username: "john_doe",
                Password: "SecureP@ss123",
            },
            wantErr: false,
        },
        {
            name: "invalid email",
            user: User{
                Name:     "John Doe",
                Email:    "invalid-email",
                Age:      25,
                Username: "john_doe",
                Password: "SecureP@ss123",
            },
            wantErr: true,
        },
        {
            name: "weak password",
            user: User{
                Name:     "John Doe",
                Email:    "john@example.com",
                Age:      25,
                Username: "john_doe",
                Password: "weak",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := v.Validate(tt.user)
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Performance Considerations

1. **Reuse Validators**: Create validator instances once and reuse them
2. **Cache Validation Rules**: Struct validation rules are cached after first use
3. **Use Middleware Wisely**: Apply validation middleware only where needed
4. **Validate Once**: Don't re-validate the same data multiple times

## Summary

GOE's validation system provides:
- Comprehensive request validation for all HTTP data types
- Clear, actionable error messages
- Extensibility through custom validators
- Seamless integration with Fiber and dependency injection
- High performance through efficient caching

Use validation to ensure data integrity, improve user experience with clear error messages, and maintain clean, secure application code.