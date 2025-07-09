# Testing

GOE Framework provides comprehensive testing capabilities that leverage Go's built-in testing framework along with dependency injection to make testing easier and more maintainable.

## Testing Philosophy

GOE promotes testable code through:

- **Dependency Injection**: Easy mocking of dependencies
- **Interface-based Design**: Clean separation of concerns
- **Modular Architecture**: Isolated testing of components
- **Contract-based Testing**: Consistent testing patterns

## Unit Testing

### Testing Services

GOE makes testing services straightforward with dependency injection:

```go
func TestUserService(t *testing.T) {
    // Create mocks
    mockRepo := &mocks.UserRepository{}
    mockLogger := &mocks.Logger{}
    
    // Setup expectations
    mockRepo.On("FindByID", "123").Return(&User{ID: "123", Name: "John"}, nil)
    
    // Create service with mocks
    service := NewUserService(mockRepo, mockLogger)
    
    // Test
    user, err := service.GetUser("123")
    
    // Assertions
    assert.NoError(t, err)
    assert.Equal(t, "John", user.Name)
    mockRepo.AssertExpectations(t)
}
```

### Testing HTTP Handlers

Test your HTTP handlers using Fiber's testing utilities:

```go
func TestUserHandler(t *testing.T) {
    // Create test app
    app := fiber.New()
    
    // Create handler with mocked dependencies
    userService := &mocks.UserService{}
    handler := NewUserHandler(userService)
    
    // Setup routes
    app.Get("/users/:id", handler.GetUser)
    
    // Setup mock expectations
    userService.On("GetUser", "123").Return(&User{ID: "123", Name: "John"}, nil)
    
    // Create test request
    req := httptest.NewRequest("GET", "/users/123", nil)
    
    // Perform request
    resp, err := app.Test(req)
    
    // Assertions
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
    
    // Parse response
    var user User
    body, _ := io.ReadAll(resp.Body)
    json.Unmarshal(body, &user)
    assert.Equal(t, "John", user.Name)
}
```

## Integration Testing

### Testing with Real Database

For integration tests, you can use a test database:

```go
func TestUserRepository_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Create repository
    repo := NewUserRepository(db)
    
    // Test create
    user := &User{Name: "John", Email: "john@example.com"}
    err := repo.Create(user)
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
    
    // Test find
    found, err := repo.FindByID(user.ID)
    assert.NoError(t, err)
    assert.Equal(t, user.Name, found.Name)
}
```

### Testing Modules

Test entire modules in isolation:

```go
func TestUserModule(t *testing.T) {
    // Create test container
    container := dig.New()
    
    // Register test dependencies
    container.Provide(func() Database { return &TestDB{} })
    container.Provide(func() Logger { return &TestLogger{} })
    
    // Register module
    module := NewUserModule()
    err := module.Register(container)
    assert.NoError(t, err)
    
    // Test service resolution
    var service UserService
    err = container.Invoke(func(s UserService) {
        service = s
    })
    assert.NoError(t, err)
    assert.NotNil(t, service)
}
```

## Test Utilities

### Test Configuration

Create test-specific configuration:

```go
func TestConfig() *Config {
    return &Config{
        Database: DatabaseConfig{
            DSN: "sqlite::memory:",
        },
        Logger: LoggerConfig{
            Level: "debug",
        },
        Server: ServerConfig{
            Port: 0, // Random port
        },
    }
}
```

### Test Helpers

Common test helpers:

```go
// Setup test database
func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    require.NoError(t, err)
    
    // Run migrations
    err = runMigrations(db)
    require.NoError(t, err)
    
    return db
}

// Create test HTTP request
func createTestRequest(method, url string, body interface{}) *http.Request {
    var buf bytes.Buffer
    if body != nil {
        json.NewEncoder(&buf).Encode(body)
    }
    
    req := httptest.NewRequest(method, url, &buf)
    req.Header.Set("Content-Type", "application/json")
    return req
}
```

## Mocking

### Using testify/mock

GOE works well with testify/mock for creating mocks:

```go
//go:generate mockery --name=UserRepository
type UserRepository interface {
    FindByID(id string) (*User, error)
    Create(user *User) error
    Update(user *User) error
    Delete(id string) error
}

// In tests
mockRepo := &mocks.UserRepository{}
mockRepo.On("FindByID", "123").Return(&User{ID: "123"}, nil)
```

### Interface Segregation

Design interfaces for easy testing:

```go
// Good - focused interface
type UserFinder interface {
    FindByID(id string) (*User, error)
}

// Better than large interface
type UserRepository interface {
    FindByID(id string) (*User, error)
    Create(user *User) error
    Update(user *User) error
    Delete(id string) error
    FindByEmail(email string) (*User, error)
    List(offset, limit int) ([]*User, error)
}
```

## Test Organization

### File Structure

Organize tests alongside source code:

```
user/
├── handler.go
├── handler_test.go
├── repository.go
├── repository_test.go
├── service.go
└── service_test.go
```

### Test Suites

Group related tests:

```go
func TestUserService(t *testing.T) {
    t.Run("GetUser", func(t *testing.T) {
        t.Run("Success", func(t *testing.T) {
            // Test success case
        })
        
        t.Run("NotFound", func(t *testing.T) {
            // Test not found case
        })
    })
    
    t.Run("CreateUser", func(t *testing.T) {
        // Test user creation
    })
}
```

## Coverage

### Running Tests with Coverage

```bash
# Run all tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Check coverage percentage
go tool cover -func=coverage.out
```

### Coverage Goals

Aim for:
- **Unit tests**: 80%+ coverage
- **Integration tests**: Critical paths covered
- **End-to-end tests**: Happy path scenarios

## Best Practices

1. **Test Structure**: Use Arrange-Act-Assert pattern
2. **Naming**: Use descriptive test names
3. **Independence**: Tests should not depend on each other
4. **Fast Tests**: Unit tests should run quickly
5. **Mocking**: Mock external dependencies
6. **Table Tests**: Use table-driven tests for multiple scenarios
7. **Error Cases**: Test both success and error scenarios
8. **Setup/Teardown**: Use proper setup and cleanup

## Example Test Suite

```go
func TestUserService_GetUser(t *testing.T) {
    tests := []struct {
        name           string
        userID         string
        repoResponse   *User
        repoError      error
        expectedUser   *User
        expectedError  error
    }{
        {
            name:         "success",
            userID:       "123",
            repoResponse: &User{ID: "123", Name: "John"},
            repoError:    nil,
            expectedUser: &User{ID: "123", Name: "John"},
            expectedError: nil,
        },
        {
            name:          "user not found",
            userID:        "456",
            repoResponse:  nil,
            repoError:     ErrUserNotFound,
            expectedUser:  nil,
            expectedError: ErrUserNotFound,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            mockRepo := &mocks.UserRepository{}
            mockRepo.On("FindByID", tt.userID).Return(tt.repoResponse, tt.repoError)
            
            service := NewUserService(mockRepo)
            
            // Act
            user, err := service.GetUser(tt.userID)
            
            // Assert
            assert.Equal(t, tt.expectedError, err)
            assert.Equal(t, tt.expectedUser, user)
            mockRepo.AssertExpectations(t)
        })
    }
}
```

This comprehensive testing approach ensures your GOE applications are reliable, maintainable, and well-tested.