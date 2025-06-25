# 7. Building Web Applications with the HTTP Server 🌐

Goe's HTTP module provides a robust and high-performance foundation for building web applications and APIs. It's built on top of [GoFiber](https://gofiber.io/), a web framework inspired by Express.js and built on Fasthttp, the fastest HTTP engine for Go.

## Enabling the HTTP Module

To use the HTTP server, enable it when initializing your Goe application:

```go
package main

import (
	"go.oease.dev/goe/v2"
	// ... other imports
)

func main() {
	app := goe.New(goe.Options{
		WithHTTP: true, // Enable the HTTP module
		// ... other options, like Invokers to register routes
	})
	goe.Run()
}
```

## Configuration

The HTTP server can be configured using various environment variables, typically prefixed with `HTTP_` or `FIBER_`. Key options include:

*   `HTTP_HOST`: The host address the server binds to. Default: `0.0.0.0`.
*   `HTTP_PORT`: The port the server listens on. Default: `8080`.
*   `HTTP_READ_TIMEOUT`: Max duration for reading the entire request. Default: `10s`.
*   `HTTP_WRITE_TIMEOUT`: Max duration for writing the response. Default: `10s`.
*   `HTTP_IDLE_TIMEOUT`: Max duration an idle connection is kept alive. Default: `30s`.
*   `FIBER_SERVER_HEADER`: Custom value for the `Server` HTTP header. Default: `"Goe"`.
*   `FIBER_APP_NAME`: Application name, can be used by Fiber internally.
*   `FIBER_BODY_LIMIT`: Maximum request body size in bytes. Default: `4194304` (4MB).
*   `FIBER_STRICT_ROUTING`: Enable/disable strict routing.
*   `FIBER_CASE_SENSITIVE`: Enable/disable case-sensitive routing.
*   `FIBER_TRUST_PROXY`: (`true`/`false`) Enables trusting of `X-Forwarded-*` headers.
    *   `FIBER_TRUST_PROXIES`: Comma-separated list of trusted proxy IPs/ranges.
*   `FIBER_JSON_ENCODER` / `FIBER_JSON_DECODER`: Goe defaults to using `github.com/bytedance/sonic` for faster JSON operations.

Refer to the [Configuration](05-configuration.md#common-configuration-keys) guide and the [GoFiber documentation](https://docs.gofiber.io/api/fiber#config) for a complete list of Fiber configuration options.

## Routing

Routing involves defining URL patterns and associating them with handler functions. You access the underlying Fiber app instance via `contract.HTTPKernel`.

```go
package main

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	goehttp "go.oease.dev/goe/v2/core/http" // For GetLogger, etc.
)

func main() {
	_ = goe.New(goe.Options{
		WithHTTP: true,
		Invokers: []any{RegisterRoutes}, // Register routes via Fx invoker
	})
	goe.Run()
}

// RegisterRoutes is an Fx invoker
func RegisterRoutes(kernel contract.HTTPKernel, logger contract.Logger) {
	app := kernel.App() // Get the *fiber.App instance

	// Simple GET route
	app.Get("/hello", func(c fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// Route with path parameters
	app.Get("/users/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		// user, err := fetchUser(id) ...
		return c.JSON(fiber.Map{"user_id": id, "name": "Demo User"})
	})

	// POST route
	app.Post("/submit", func(c fiber.Ctx) error {
		// ... handle form submission ...
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Resource created"})
	})

	// Route Grouping
	api := app.Group("/api/v1", apiV1Middleware) // Group with optional middleware
	api.Get("/status", func(c fiber.Ctx) error {
		reqLogger := goehttp.GetLogger(c) // Access request-scoped logger
		reqLogger.Info("API v1 status requested")
		return c.JSON(fiber.Map{"status": "API v1 is operational"})
	})
	api.Post("/data", func(c fiber.Ctx) error {
		// ... handle data submission for API v1 ...
		return c.SendStatus(fiber.StatusOK)
	})

	logger.Info("HTTP routes registered successfully.")
}

func apiV1Middleware(c fiber.Ctx) error {
	c.Set("X-API-Version", "v1")
	return c.Next()
}
```

GoFiber supports various routing features:
*   Path parameters (e.g., `/:name`, `/:id<int>`, `/:file<path>`)
*   Optional parameters (e.g., `/:name?`)
*   Wildcards (e.g., `/api/*`)
*   Route naming, route metadata.

Consult the [GoFiber Routing Documentation](https://docs.gofiber.io/routing) for more details.

## Handlers

Handlers are functions that process incoming requests and generate responses. They receive a `fiber.Ctx` (aliased as `contract.Context`) object.

### Accessing Request Data:

*   **Path Parameters**: `c.Params("key")`, `c.ParamsInt("key")`, etc.
*   **Query Parameters**: `c.Query("key")`, `c.QueryInt("key")`, etc.
*   **Request Body**:
    *   `c.Body()`: Returns the raw body as `[]byte`.
    *   `c.BodyParser(out interface{}) error`: Parses the body into a struct (supports JSON, XML, form data based on `Content-Type`).
    *   `c.FormValue("key")`: For form data.
*   **Headers**: `c.Get("Header-Name")`.
*   **Cookies**: `c.Cookies("cookieName")`.
*   **IP Address**: `c.IP()`.

### Sending Responses:

*   **Strings**: `c.SendString("Hello")`.
*   **JSON**: `c.JSON(data interface{})`. Goe configures Fiber to use `sonic` for faster JSON.
*   **Status Codes**: `c.Status(fiber.StatusOK).JSON(...)` or `c.SendStatus(fiber.StatusNotFound)`.
*   **Files**: `c.SendFile("path/to/file.txt")`.
*   **HTML Templates**: Fiber supports various template engines. You'd configure this on the `fiber.App` instance. (See Fiber docs for templating).
*   **Redirects**: `c.Redirect("/new-location", fiber.StatusTemporaryRedirect)`.

```go
type CreateUserPayload struct {
	Name  string `json:"name" xml:"name" form:"name" validate:"required,min=3"`
	Email string `json:"email" xml:"email" form:"email" validate:"required,email"`
}

func CreateUserHandler(c fiber.Ctx) error {
	payload := new(CreateUserPayload)

	// Parse request body (JSON, XML, form) into payload struct
	if err := c.BodyParser(payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	// Validate the payload (see Request Validation section)
	// validator := goehttp.GetValidator(c) // Assuming validator is needed
	// if err := validator.Validate(payload); err != nil { ... }


	// Use services obtained from context or DI
	logger := goehttp.GetLogger(c)
	logger.Info("Creating user", contract.NewField("name", payload.Name))

	// ... save user ...

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"name":    payload.Name,
	})
}
```

## Middleware

Middleware functions have access to the request and response objects and can execute code, make changes, or terminate the request chain.

### Built-in Goe/Fiber Middleware:

Goe automatically registers several useful middleware:

1.  **`recover`**: Recovers from panics anywhere in the request chain and converts them into a `500 Internal Server Error` response.
2.  **`requestid`**: Assigns a unique ID to each incoming request (available via `c.Locals(string(goehttp.RequestIDKey))` or `requestid.FromContext(c)`). This ID is also automatically included in logs made by the request-scoped logger.
3.  **Service Injection Middleware**: Injects core Goe services (`Config`, `Logger`, `App`, `Validator`, `Cache`) into `c.Locals(string(goehttp.ServicesKey))`. These can then be retrieved using helper functions like `goehttp.GetLogger(c)`, `goehttp.GetConfig(c)`.
4.  **Request Logging Middleware**: Logs information about each request (method, path, status, duration, request ID) using Goe's logger.

### Custom Middleware:

You can easily write your own middleware. A middleware is just a `fiber.Handler`.

```go
// Example: Authentication Middleware
func AuthMiddleware(c fiber.Ctx) error {
	apiKey := c.Get("X-API-Key")
	if apiKey != "my-secret-key" { // In reality, check against config or DB
		// You can get services from context if needed
		// logger := goehttp.GetLogger(c)
		// logger.Warn("Invalid API key attempt", contract.NewField("key_provided", apiKey))
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid API Key")
	}
	// User is authenticated, set some user context if needed
	c.Locals("authenticated_user", "service_user")
	return c.Next() // Call c.Next() to pass control to the next middleware or handler
}

// ... in RegisterRoutes ...
// app.Use(AuthMiddleware) // Apply to all routes
// apiGroup := app.Group("/protected", AuthMiddleware) // Apply to a group
// apiGroup.Get("/resource", myProtectedHandler)
```

## Request Validation

Goe integrates the [go-playground/validator](https://github.com/go-playground/validator) library for request struct validation. The validator instance is available via `contract.HTTPKernel.Validator()` or, within a handler, `goehttp.GetValidator(c)`.

```go
package main

import (
	"fmt"
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	goehttp "go.oease.dev/goe/v2/core/http"
)

// Product defines a struct with validation tags
type Product struct {
	Name  string  `json:"name" validate:"required,min=3,max=100"`
	SKU   string  `json:"sku" validate:"required,alphanum,len=8"`
	Price float64 `json:"price" validate:"required,gt=0"`
	Email string  `json:"email" validate:"required,email"` // Example custom tag if registered
}

func main() {
	_ = goe.New(goe.Options{
		WithHTTP: true,
		Invokers: []any{
			func(k contract.HTTPKernel) {
				// Example: Register a custom validation
				// validator := k.Validator().(*goehttp.CustomValidator)
				// validator.RegisterValidation("custom_tag", myCustomValidationFunc)

				app := k.App()
				app.Post("/products", CreateProductHandler)
			},
		},
	})
	goe.Run()
}

func CreateProductHandler(c fiber.Ctx) error {
	payload := new(Product)

	if err := c.BodyParser(payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	// Get the validator
	validator := goehttp.GetValidator(c) // From context
	// OR, if injected into a controller struct that has the validator:
	// validator := myController.validator

	// Validate the struct
	if err := validator.Validate(payload); err != nil {
		// err will contain validation errors
		// You might want to format these errors into a more user-friendly response
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	// ... process valid payload ...
	logger := goehttp.GetLogger(c)
	logger.Info("Product validated and being created", contract.NewField("product_name", payload.Name))

	return c.Status(fiber.StatusCreated).JSON(payload)
}
```

You can register custom validation functions and aliases using the `*goehttp.CustomValidator` instance. See `core/http/validator.go` and the `go-playground/validator` documentation for more.

## Service Access in Handlers

As shown in examples:

1.  **Via Context Helpers**: `goehttp.GetLogger(c)`, `goehttp.GetConfig(c)`, etc., retrieve services injected by the `InjectServices` middleware. This is convenient for direct Fiber handlers.
2.  **Via Dependency Injection**: If your handlers are methods on structs (controllers/services) that are themselves managed by Fx, Fx will inject dependencies into these structs.

    ```go
    // Controller struct
    type UserController struct {
        logger    contract.Logger
        userService *myinternal.UserService // Assume UserService is an Fx component
    }

    // Fx provider for UserController
    func NewUserController(logger contract.Logger, us *myinternal.UserService) *UserController {
        return &UserController{logger: logger, userService: us}
    }

    // Handler method
    func (uc *UserController) GetUser(c fiber.Ctx) error {
        userID := c.Params("id")
        uc.logger.Info("Fetching user", contract.NewField("user_id", userID))
        // user, err := uc.userService.FindByID(userID) ...
        return c.JSON(fiber.Map{"id": userID, "name": "From DI Controller"})
    }

    // In main.go or module registration:
    // fx.Provide(myinternal.NewUserService)
    // fx.Provide(NewUserController)
    // fx.Invoke(func(kernel contract.HTTPKernel, uc *UserController) {
    //    kernel.App().Get("/users/:id", uc.GetUser) // Register method handler
    // })
    ```

## Error Handling

Goe's HTTP module sets up a default error handler:

*   It catches errors returned by handlers or middleware.
*   If the error is a `fiber.Error`, its status code and message are used.
*   Otherwise, it defaults to a `500 Internal Server Error`.
*   For `5xx` errors, it logs the error along with request details.
*   It attempts to send a response in a format acceptable to the client (HTML, JSON, or plain text). An HTML error page is used as a fallback.
    *   You can force JSON or text error responses via `?format=json` or `?format=text` query parameters.

You can customize the error handler via `fiber.Config.ErrorHandler` when setting up the Fiber app if needed, but Goe's default handler (see `core/http/http.go#defaultErrorHandler`) is quite comprehensive.

To return specific HTTP errors from your handlers:

```go
// In a handler
if userNotFound {
    return fiber.NewError(fiber.StatusNotFound, "User not found with the given ID.")
}
if invalidInput {
    // You can also return a standard error, which will become a 500,
    // or a fiber.Error for specific client error codes.
    return fiber.NewError(fiber.StatusBadRequest, "Invalid input provided: <reason>")
}
```

## Graceful Shutdown

When the Goe application receives a shutdown signal (e.g., SIGINT/Ctrl+C), `goe.Run()` initiates a graceful shutdown process:

*   The HTTP server (`kernel.Shutdown()`) calls Fiber's `app.Shutdown()`, which stops accepting new connections.
*   It waits for active connections to complete their requests (up to a timeout).
*   Other modules also perform their `OnStop` cleanup.

This ensures that ongoing requests are not abruptly terminated.

## Accessing the HTTP Kernel and Validator Globally

While dependency injection is preferred for most application components, you can access the HTTP kernel or its validator globally if needed:

*   `goe.HTTP()`: Returns the `contract.HTTPKernel` instance.
    *   `goe.HTTP().App()`: Returns the `*fiber.App`.
    *   `goe.HTTP().Validator()`: Returns the validator (`any`, needs type assertion to `*goehttp.CustomValidator`).

```go
// Example of global access (less common for routing, more for one-off needs)
func someUtilityFunction() {
    if goe.IsModuleEnabled(goe.ModuleHTTP) { // Check if HTTP module is even running
        validatorInstance, ok := goe.HTTP().Validator().(*goehttp.CustomValidator)
        if ok {
            // ... use validatorInstance ...
        }
    }
}
```

This covers the core aspects of using Goe's HTTP module. For more advanced Fiber features, always refer to the official [GoFiber Documentation](https://docs.gofiber.io/).

Next, we'll explore how Goe handles [Database Interactions](08-database.md).
