# 12. Error Handling Strategies  Fehler! (Error!) 🚧

Robust error handling is critical for building reliable applications. Go has a unique approach to error handling with its explicit error returns. Goe embraces this and provides mechanisms, particularly in its HTTP layer, to manage errors gracefully.

## Go's Error Handling Philosophy

Recall Go's core error handling principles:

*   **Explicit Error Returns**: Functions that can fail return an `error` as their last value.
    ```go
    func DoSomething() (ResultType, error) {
        // ...
        if someConditionFails {
            return nil, errors.New("something went wrong")
        }
        // ...
        return result, nil
    }
    ```
*   **Checking Errors**: Callers are responsible for checking returned errors.
    ```go
    result, err := DoSomething()
    if err != nil {
        // Handle error: log it, return it, etc.
        return fmt.Errorf("DoSomething failed: %w", err) // %w wraps the original error
    }
    // Use result
    ```
*   **`error` is an Interface**: Any type that implements the `Error() string` method satisfies the `error` interface. This allows for custom error types.
*   **Error Wrapping (`fmt.Errorf` with `%w`)**: Introduced in Go 1.13, error wrapping allows you to add context to an error while preserving the original error. Use `errors.Is` and `errors.As` to inspect wrapped errors.

Goe applications should adhere to these standard Go practices.

## Error Handling in Goe

### 1. HTTP Error Handling

Goe's HTTP module (built on GoFiber) has a default error handler that catches errors returned from your HTTP handlers or middleware.

**Default Behavior (`core/http/http.go#defaultErrorHandler`):**

*   **Error Source**: If a handler returns an error, Fiber passes it to the error handler.
*   **`fiber.Error`**: If the returned error is a `*fiber.Error`, the error handler uses its specified status code and message. This is the recommended way to return specific HTTP client errors (4xx).
    ```go
    // In a handler
    if !resourceFound {
        return fiber.NewError(fiber.StatusNotFound, "The requested resource was not found.")
    }
    if badRequestInput {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid input: name field is required.")
    }
    ```
*   **Standard Errors**: If a standard Go error (not `*fiber.Error`) is returned, it's treated as a `500 Internal Server Error`.
*   **Logging**: For `5xx` errors (server-side issues), the default error handler logs the error message, request path, method, and status code using Goe's logger. This helps in diagnosing server problems. `4xx` client errors are generally not logged as server errors by default, as they often indicate client mistakes.
*   **Response Format**: The error handler tries to send a response in a format acceptable to the client:
    *   If `?format=json` query param is present: JSON response `{"message": "error message"}`.
    *   If `?format=text` query param is present: Plain text response.
    *   Otherwise, it checks `Accept` header:
        *   `application/json`: JSON response.
        *   `text/html`: HTML error page (a simple default is provided by Goe).
        *   `text/plain`: Plain text response.
    *   Fallback is the HTML error page.

**Customizing HTTP Error Handling**:
While Goe's default is quite robust, you can provide a custom error handler to Fiber if needed:
```go
// When creating Fiber app (advanced, usually not needed with Goe's setup)
// fiberApp := fiber.New(fiber.Config{
//     ErrorHandler: func(c fiber.Ctx, err error) error {
//         // ... your custom logic ...
//         code := fiber.StatusInternalServerError
//         // ... determine code and message ...
//         return c.Status(code).JSON(fiber.Map{"error": message})
//     },
// })
// goeKernel := goe.HTTP().(*corehttp.kernel) // Assuming you can access and modify
// goeKernel.SetApp(fiberApp) // Hypothetical setter, direct modification is tricky
```
Modifying the error handler after Goe initializes the HTTP kernel requires careful manipulation or would need more explicit support from Goe's API. For most cases, returning `fiber.Error` or standard errors from handlers is sufficient.

### 2. Service Layer and Other Components

In your services, repositories, and other non-HTTP components, use standard Go error handling:

*   Return errors explicitly.
*   Wrap errors to add context.
*   Define custom error types if they help convey specific failure modes that callers need to handle differently.

```go
// internal/services/userservice/user_service.go
package userservice

import (
	"errors"
	"fmt"
	"go.oease.dev/goe/v2/contract"
	// "yourproject/internal/models"
	// "yourproject/internal/customerrors"
)

var ErrUserNotFound = errors.New("user not found") // Sentinel error

type UserService struct {
	logger contract.Logger
	db     contract.DB
}

func NewUserService(logger contract.Logger, db contract.DB) *UserService {
	return &UserService{logger: logger, db: db}
}

func (s *UserService) GetUserByID(id uint) (/*models.User*/ any, error) {
	// var user models.User
	// err := s.db.Instance().First(&user, id).Error
	var err error // Simulate DB error
	if id == 0 { // Simulate not found
		err = gorm.ErrRecordNotFound // Assuming gorm is used by db.Instance()
	} else if id == 99 { // Simulate other DB error
        err = errors.New("simulated database connection error")
    }


	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("User lookup failed: not found", contract.NewField("user_id", id))
			// return models.User{}, customerrors.NewNotFoundError(fmt.Sprintf("user with ID %d not found", id))
			// Or, return a specific sentinel error if preferred for this layer
			return nil, fmt.Errorf("could not find user with ID %d: %w", id, ErrUserNotFound)
		}
		s.logger.Error("Database error fetching user", contract.NewField("user_id", id), contract.NewField("error", err))
		// return models.User{}, customerrors.NewInfrastructureError(fmt.Errorf("failed to query user %d: %w", id, err))
		return nil, fmt.Errorf("failed to query user %d: %w", id, err) // Wrap the original error
	}
	// return user, nil
	return map[string]any{"id": id, "name": "Mock User"}, nil // Simulate found user
}
```

## Best Practices for Error Handling in Goe Applications

1.  **Always Check Errors**: Don't ignore errors. This is fundamental Go.

2.  **Wrap Errors for Context**: When an error passes through layers, wrap it with `fmt.Errorf("context: %w", err)` to add context about where and why the error occurred. This creates an error chain that can be inspected.

3.  **Define Custom Error Types or Sentinel Errors**:
    *   **Sentinel Errors**: Pre-defined error values (e.g., `var ErrUserNotFound = errors.New("user not found")`). Use `errors.Is()` to check for them. Useful for common, known error conditions.
    *   **Custom Error Types**: Structs that implement the `error` interface. Can carry more data (e.g., error codes, parameters). Use `errors.As()` to check and extract data.
    ```go
    // Example Custom Error
    type ValidationError struct {
        Field   string
        Message string
    }

    func (e *ValidationError) Error() string {
        return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
    }
    // ...
    // return &ValidationError{Field: "email", Message: "is not a valid email format"}
    ```

4.  **Log Errors Appropriately**:
    *   **Where to Log**: Log errors at the point where they are handled or become significant for operational insight. Often, this is at the boundary of your application (e.g., HTTP handler, main goroutine of a worker) or when a service decides it cannot recover. Avoid redundant logging of the same error at multiple levels of the call stack if it's just being passed up.
    *   **What to Log**: Log the error itself and any relevant contextual information (e.g., user ID, request parameters, operation being performed) as structured fields. Goe's logger handles `error` types well.
    ```go
    userID := "user123"
    err := processUserData(userID)
    if err != nil {
        // Log here, at the handling boundary
        logger.Error("Failed to process user data",
            contract.NewField("user_id", userID),
            contract.NewField("error", err), // Original error is preserved
        )
        // Return an appropriate error to the caller / HTTP response
    }
    ```

5.  **Specific HTTP Status Codes**: In HTTP handlers, translate internal application errors into appropriate HTTP status codes. Use `fiber.NewError` for client errors (4xx) and let standard errors become 5xx server errors (which Goe logs).

6.  **Graceful Degradation**: For non-critical operations, consider if a failure should halt the entire request or if the application can degrade gracefully (e.g., show a page without personalized content if a recommendation service fails).

7.  **Panic for Unrecoverable States Only**: `panic` should generally be reserved for truly exceptional, unrecoverable situations (e.g., programmer errors like nil pointer dereferences in critical paths where the state is unknown, or critical resource unavailability at startup). Goe's HTTP `recover` middleware will catch panics in handlers and convert them to 500 errors.

8.  **Consistent Error Responses (APIs)**: For APIs, define a consistent JSON error response format.
    ```json
    // Example error response
    {
      "error": {
        "code": "VALIDATION_ERROR", // Or "NOT_FOUND", "UNAUTHENTICATED"
        "message": "The email field is required.",
        "details": [ // Optional: more specific errors
          { "field": "email", "issue": "cannot be empty" }
        ]
      }
    }
    ```
    You can achieve this by having a helper function that constructs these responses in your HTTP error handling logic or by customizing Fiber's error handler.

By consistently applying these error handling principles, your Goe applications will be more robust, easier to debug, and provide better feedback to users and developers.

Up next: Strategies for [Testing Your Goe Application](13-testing.md).
```
