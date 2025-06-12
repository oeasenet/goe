# 13. Testing Your Goe Application 🧪

Testing is a crucial part of software development, ensuring your application behaves as expected and preventing regressions. Goe's design, with its emphasis on dependency injection and interfaces (contracts), facilitates various types of testing.

## Testing Philosophy

*   **Go's Built-in `testing` Package**: Leverage Go's standard library for writing tests.
*   **Unit Tests**: Test individual components (functions, structs, services, repositories) in isolation. Dependencies should be mocked or stubbed.
*   **Integration Tests**: Test the interaction between several components (e.g., a service and its repository, or an HTTP handler with its underlying service). May involve real external dependencies like a test database or a mock HTTP server.
*   **End-to-End (E2E) Tests**: Test the entire application flow from an external perspective (e.g., making HTTP requests to a running instance of your application). These are generally slower and more complex but provide high confidence.
*   **Testable Code**: Write your code with testability in mind from the start. This often means:
    *   Preferring dependency injection over global state.
    *   Depending on interfaces rather than concrete types.
    *   Keeping functions and methods focused on a single responsibility.

## Unit Testing

Unit tests focus on the smallest parts of your application.

### Testing a Service

Consider a simple service:

```go
// internal/services/calculator/calculator_service.go
package calculator

import "go.oease.dev/goe/v2/contract"

type CalculatorService struct {
	logger contract.Logger
}

func NewCalculatorService(logger contract.Logger) *CalculatorService {
	return &CalculatorService{logger: logger}
}

func (s *CalculatorService) Add(a, b int) int {
	result := a + b
	s.logger.Info("Addition performed",
		contract.NewField("a", a),
		contract.NewField("b", b),
		contract.NewField("result", result),
	)
	return result
}
```

To test `CalculatorService.Add`:

```go
// internal/services/calculator/calculator_service_test.go
package calculator

import (
	"context" // Added for WithContext
	"testing"
	"go.oease.dev/goe/v2/contract" // Changed from core/log
	"go.uber.org/zap"             // Added for GetLogger return type if not nil
	// "github.com/stretchr/testify/assert" // Popular assertion library (optional)
)

// MockLogger for testing (basic example)
type MockLogger struct {
	// Add fields to capture log messages if needed for assertions
}

func (m *MockLogger) Debug(msg string, fields ...contract.Field) {}
func (m *MockLogger) Info(msg string, fields ...contract.Field)  {}
func (m *MockLogger) Warn(msg string, fields ...contract.Field)  {}
func (m *MockLogger) Error(msg string, fields ...contract.Field) {}
func (m *MockLogger) Fatal(msg string, fields ...contract.Field) {} // In tests, maybe panic or set a flag
func (m *MockLogger) With(fields ...contract.Field) contract.Logger   { return m }
func (m *MockLogger) WithContext(ctx context.Context) contract.Logger { return m }
func (m *MockLogger) WithError(err error) contract.Logger        { return m }
func (m *MockLogger) GetLogger() *zap.SugaredLogger { return nil } // Return a dummy if not used

func TestCalculatorService_Add(t *testing.T) {
	mockLogger := &MockLogger{} // Use your preferred mocking tool or a simple stub
	calcService := NewCalculatorService(mockLogger)

	testCases := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"add positive numbers", 2, 3, 5},
		{"add with zero", 5, 0, 5},
		{"add negative numbers", -2, -3, -5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calcService.Add(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("Add(%d, %d) = %d; want %d", tc.a, tc.b, result, tc.expected)
			}
			// Using testify/assert:
			// assert.Equal(t, tc.expected, result)
		})
	}
}
```

### Mocking Dependencies

When testing a component that has dependencies (like `contract.Logger`, `contract.DB`, or other services), you'll need to provide mock implementations of these dependencies.

*   **Manual Mocks**: Create simple structs that implement the required interface, like `MockLogger` above.
*   **Mocking Libraries**: Tools like [GoMock](https://github.com/golang/mock) or [Testify's mock package](https://github.com/stretchr/testify#mock-package) can generate mock implementations from interfaces, providing more features like call expectations and argument matching.

## Integration Testing

Integration tests verify that different parts of your system work together correctly.

### Testing HTTP Handlers

To test HTTP handlers, you can use the `net/http/httptest` package. Since Goe uses Fiber, which is not directly compatible with `net/http.Handler`, you'll typically test against the Fiber app instance.

```go
// internal/handlers/userhandler/user_handler_test.go
package userhandler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert" // Using testify for assertions
	"go.oease.dev/goe/v2/contract"
	corehttp "go.oease.dev/goe/v2/core/http"
	"go.oease.dev/goe/v2/core/log"     // For a test logger
	// "yourproject/internal/services/userservice" // Assume you have a user service
)

// MockUserService for testing the handler
type MockUserService struct {
	GetUserByIDFunc func(id string) (any, error)
}
// func (m *MockUserService) GetUserByID(id string) (*userservice.UserDTO, error) { /* ... */ }
// Implement other methods of your actual user service interface

func TestUserHandler_GetUser(t *testing.T) {
	// Setup: Create a Fiber app instance for testing
	app := fiber.New()

	// Mock dependencies
	mockLogger := &log.MockLogger{} // Assuming you have a mock logger
	// mockUserService := &MockUserService{
	// 	GetUserByIDFunc: func(id string) (any, error) {
	// 		if id == "1" {
	// 			return map[string]string{"id": "1", "name": "Test User"}, nil
	// 		}
	// 		return nil, userservice.ErrUserNotFound // Assuming your service has this
	// 	},
	// }
	// userCtrl := NewUserController(mockLogger, mockUserService) // Your handler/controller

	// Setup a route on the test app
	// For simplicity, directly defining a handler here.
	// In a real scenario, you'd register your actual handler/controller methods.
	app.Get("/users/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		if id == "1" {
			return c.JSON(fiber.Map{"id": "1", "name": "Test User"})
		}
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	})


	// Test Case 1: User Found
	reqFound := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	respFound, _ := app.Test(reqFound) // Fiber's Test method

	assert.Equal(t, http.StatusOK, respFound.StatusCode)
	var bodyFound map[string]string
	json.NewDecoder(respFound.Body).Decode(&bodyFound)
	assert.Equal(t, "1", bodyFound["id"])
	assert.Equal(t, "Test User", bodyFound["name"])
	respFound.Body.Close()

	// Test Case 2: User Not Found
	reqNotFound := httptest.NewRequest(http.MethodGet, "/users/2", nil)
	respNotFound, _ := app.Test(reqNotFound)

	assert.Equal(t, http.StatusNotFound, respNotFound.StatusCode)
	respNotFound.Body.Close()
}
```

**Key points for testing Fiber handlers:**

*   Create a `fiber.New()` instance for your tests.
*   Register the specific handler(s) you want to test on this app.
*   Use `app.Test(req)` which takes an `*http.Request` and returns an `*http.Response`. This allows you to use `httptest.NewRequest`.
*   Make assertions about the response status code, headers, and body.

### Testing with a Real Database (Test Database)

For integration tests involving database interactions (e.g., testing a repository or a service that uses a repository):

*   **Use a Test Database**: Configure Goe to connect to a separate test database (e.g., a local PostgreSQL instance, an in-memory SQLite database).
*   **Environment Configuration**: Set environment variables specifically for the test environment (e.g., in a `.env.test` file or via test script setup).
    ```env
    # .env.test
    GOE_ENV=test
    DB_DRIVER=sqlite
    DB_DATABASE=file:test_db.sqlite?cache=shared&mode=memory
    # Or use a real DB connection string for a dedicated test DB
    DB_LOG_MODE=false
    LOG_LEVEL=error # Keep test logs quiet unless errors occur
    ```
*   **Migrations**: Ensure your test database schema is up-to-date. You can run migrations at the beginning of your test suite or for specific test cases.
*   **Data Seeding & Teardown**: Seed necessary data before tests and clean up data afterwards to ensure tests are isolated and repeatable. Transactions can be useful here.

```go
// internal/repositories/userrepo/user_repository_integration_test.go
package userrepo_test // Use _test package for integration tests

import (
	"testing"
	// "go.oease.dev/goe/v2"
	// "go.oease.dev/goe/v2/contract"
	// "yourproject/internal/models"
	// Test setup helpers
	// "github.com/stretchr/testify/suite"
)

// This is a simplified example. A full setup would involve:
// 1. Initializing Goe with test-specific config (e.g., pointing to a test DB).
// 2. Running migrations.
// 3. Using the DB instance from Goe for the repository.

func TestUserRepository_Integration_CreateAndGetUser(t *testing.T) {
	// TODO: Full Goe app setup for integration tests is more involved.
	// It requires initializing Fx with test configurations.
	// For now, this is a conceptual placeholder.

	// --- Setup (Conceptual - A helper function would typically do this) ---
	// 1. Load .env.test or set env vars
	// os.Setenv("GOE_ENV", "test")
	// os.Setenv("DB_DRIVER", "sqlite")
	// os.Setenv("DB_DATABASE", "file::memory:?cache=shared")

	// 2. Initialize Goe
	// testApp := goe.New(goe.Options{WithDB: true, WithLog: true})
	// defer testApp.Container().Stop(context.Background()) // Ensure cleanup

	// 3. Get DB instance (this would require Fx to be running or specific DI for test)
	// var db contract.DB
	// testApp.Container().Invoke(func(d contract.DB) { db = d })
	// if db == nil || db.Instance() == nil {
	// 	t.Fatal("Failed to get DB instance for test")
	// }
	// gormDB := db.Instance()

	// 4. Run migrations for test models
	// err := gormDB.AutoMigrate(&models.User{})
	// if err != nil {
	// 	t.Fatalf("Failed to migrate test database: %v", err)
	// }
	// defer gormDB.Migrator().DropTable(&models.User{}) // Teardown

	// --- Test Logic (using the prepared gormDB) ---
	// userRepo := userrepo.NewGormUserRepository(gormDB) // Assuming such a constructor

	// createdUser, err := userRepo.Create("Test User", "test@example.com")
	// assert.NoError(t, err)
	// assert.NotZero(t, createdUser.ID)

	// fetchedUser, err := userRepo.GetByID(createdUser.ID)
	// assert.NoError(t, err)
	// assert.Equal(t, "Test User", fetchedUser.Name)

	t.Log("Integration test placeholder - full Fx app setup for tests is complex.")
}
```
**Note**: Setting up a full Goe/Fx application instance for integration tests can be complex. You might need test-specific Fx options or helper packages to manage this. Libraries like `testify/suite` can help structure test suites with setup/teardown logic.

## Setting Up a Test Environment

*   **Configuration Files**: Use a dedicated configuration file for tests (e.g., `.env.test`). Set `GOE_ENV=test` when running tests.
*   **Quiet Logging**: Configure `LOG_LEVEL=error` or `warn` for tests to keep output clean, unless you are specifically testing logging behavior.
*   **In-Memory Databases**: For faster tests and isolation, use in-memory SQLite if your application is compatible.
*   **Parallel Tests**: Be cautious when running tests in parallel (`go test -parallel N`) if they share state (e.g., a single database instance, global variables). Design tests to be independent or use appropriate synchronization. Fx itself generally constructs a new DI container for each test if you're bootstrapping Fx per test (which can be slow).

## Testing Goe Modules (`contract.Module`)

If you create custom Goe modules implementing `contract.Module`:

*   **Unit Test `OnStart` / `OnStop` Logic**: If these methods contain significant logic, unit test them by providing mock dependencies to your module's constructor.
*   **Integration Test Lifecycle**: To test a module's behavior within the Fx lifecycle, you would need to construct a minimal Fx app in your test, provide your module, and then start/stop the Fx app to trigger the hooks.

```go
// internal/modules/mymodule/my_module_test.go
package mymodule

import (
	"context"
	"testing"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest" // Fx's testing utilities
	"go.oease.dev/goe/v2/core/log"
	"github.com/stretchr/testify/assert"
)

func TestMyModule_Lifecycle(t *testing.T) {
	var moduleInstance *MyModule // To access the instance if needed for assertions
	mockLogger := &log.MockLogger{}

	app := fxtest.New(t,
		fx.Provide(func() contract.Logger { return mockLogger }), // Provide mock logger
		// Provide other dependencies your module needs for construction
		fx.Provide(NewMyModule), // Provide your module's constructor
		fx.Populate(&moduleInstance), // Populate the moduleInstance variable
	)
	// fxtest.Lifecycle automatically starts and stops the app.
	// If you need manual control:
	// lc := fxtest.NewLifecycle(t)
	// app := fx.New(
	//     // ... providers ...
	//     fx.Supply(lc), // Supply the lifecycle
	// )
	// lc.RequireStart()
	// // ... assertions ...
	// lc.RequireStop()


	app.RequireStart() // Start the app, OnStart hooks run
	// Assertions about what OnStart should have done
	// e.g., if MyModule sets a flag on start:
	// assert.True(t, moduleInstance.startedSuccessfully)

	app.RequireStop() // Stop the app, OnStop hooks run
	// Assertions about what OnStop should have done
	// assert.True(t, moduleInstance.stoppedCleanly)
}
```
`fxtest` provides utilities that simplify testing Fx applications by managing the application lifecycle within a test.

## End-to-End (E2E) Tests

E2E tests involve running your actual compiled application and interacting with it as a user would (e.g., making real HTTP calls to its exposed endpoints).

*   **Setup**: Requires building your application binary, running it (often in a controlled environment like Docker), and ensuring any external dependencies (database, other services) are available.
*   **Tools**: Use HTTP client libraries (Go's `net/http` or higher-level ones) or dedicated E2E testing frameworks (e.g., [Playwright](https://playwright.dev/docs/intro) if testing UIs, though less common for pure Go backend testing).
*   **Assertions**: Verify responses, side effects (e.g., data created in the database), and overall application behavior.
*   **Complexity**: E2E tests are powerful but are the most complex and slowest to run. Use them judiciously for critical user flows.

Writing testable code and having a comprehensive test suite are essential for maintaining a healthy Goe application. Start with unit tests for core logic and gradually add integration and E2E tests for broader coverage.

Next, we'll look at some [Practical Examples and Use Cases](14-examples.md).
```
