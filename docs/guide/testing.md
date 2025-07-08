# Testing

GOE's architecture makes testing straightforward through dependency injection, clear interfaces, and modular design. This guide covers testing strategies for different layers of your application.

## Testing Philosophy

GOE promotes:
- **Unit Testing**: Test individual components in isolation
- **Integration Testing**: Test component interactions
- **Contract Testing**: Test interface implementations
- **HTTP Testing**: Test HTTP handlers and middleware

## Test Structure

Follow GOE's recommended test structure:

```
your-project/
├── internal/
│   ├── service/
│   │   ├── user_service.go
│   │   └── user_service_test.go
│   ├── repository/
│   │   ├── user_repository.go
│   │   └── user_repository_test.go
│   └── handler/
│       ├── user_handler.go
│       └── user_handler_test.go
├── tests/
│   ├── integration/
│   │   └── user_api_test.go
│   └── e2e/
│       └── user_flow_test.go
└── testutil/
    ├── mocks/
    └── fixtures/
```

## Unit Testing

### Testing Services

```go
package service

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "your-project/internal/domain"
)

// Mock repository
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) FindByID(id int) (*domain.User, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Create(user *domain.User) error {
    args := m.Called(user)
    return args.Error(0)
}

// Mock logger
type MockLogger struct {
    mock.Mock
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
    args := []interface{}{msg}
    args = append(args, keysAndValues...)
    m.Called(args...)
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
    args := []interface{}{msg}
    args = append(args, keysAndValues...)
    m.Called(args...)
}

// Test service
func TestUserService_GetUser(t *testing.T) {
    // Setup
    mockRepo := new(MockUserRepository)
    mockLogger := new(MockLogger)
    
    expectedUser := &domain.User{ID: 1, Name: "John Doe"}
    mockRepo.On("FindByID", 1).Return(expectedUser, nil)
    mockLogger.On("Info", "Getting user", "id", 1)
    
    service := NewUserService(mockRepo, mockLogger)
    
    // Execute
    user, err := service.GetUser(1)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expectedUser, user)
    
    // Verify mocks
    mockRepo.AssertExpectations(t)
    mockLogger.AssertExpectations(t)
}

func TestUserService_GetUser_NotFound(t *testing.T) {
    // Setup
    mockRepo := new(MockUserRepository)
    mockLogger := new(MockLogger)
    
    mockRepo.On("FindByID", 999).Return(nil, domain.ErrUserNotFound)
    mockLogger.On("Info", "Getting user", "id", 999)
    
    service := NewUserService(mockRepo, mockLogger)
    
    // Execute
    user, err := service.GetUser(999)
    
    // Assert
    assert.Error(t, err)
    assert.Nil(t, user)
    assert.Equal(t, domain.ErrUserNotFound, err)
    
    mockRepo.AssertExpectations(t)
    mockLogger.AssertExpectations(t)
}
```

### Testing Repositories

```go
package repository

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "your-project/internal/domain"
)

func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    
    // Auto-migrate
    err = db.AutoMigrate(&domain.User{})
    require.NoError(t, err)
    
    return db
}

func TestUserRepository_Create(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    mockDB := &MockDB{instance: db}
    
    repo := NewUserRepository(mockDB)
    
    user := &domain.User{
        Name:  "John Doe",
        Email: "john@example.com",
    }
    
    // Execute
    err := repo.Create(user)
    
    // Assert
    assert.NoError(t, err)
    assert.NotZero(t, user.ID)
    
    // Verify in database
    var found domain.User
    err = db.First(&found, user.ID).Error
    assert.NoError(t, err)
    assert.Equal(t, user.Name, found.Name)
}

func TestUserRepository_FindByID(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    mockDB := &MockDB{instance: db}
    
    // Create test user
    testUser := &domain.User{Name: "Jane Doe", Email: "jane@example.com"}
    db.Create(testUser)
    
    repo := NewUserRepository(mockDB)
    
    // Execute
    user, err := repo.FindByID(int(testUser.ID))
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, testUser.Name, user.Name)
    assert.Equal(t, testUser.Email, user.Email)
}
```

### Testing HTTP Handlers

```go
package handler

import (
    "bytes"
    "encoding/json"
    "net/http/httptest"
    "testing"
    "github.com/gofiber/fiber/v3"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "your-project/internal/domain"
)

type MockUserService struct {
    mock.Mock
}

func (m *MockUserService) GetUser(id int) (*domain.User, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) CreateUser(name, email string) (*domain.User, error) {
    args := m.Called(name, email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.User), args.Error(1)
}

func TestUserHandler_GetUser(t *testing.T) {
    // Setup
    app := fiber.New()
    mockService := new(MockUserService)
    mockLogger := new(MockLogger)
    
    handler := NewUserHandler(mockService, mockLogger)
    
    expectedUser := &domain.User{ID: 1, Name: "John Doe"}
    mockService.On("GetUser", 1).Return(expectedUser, nil)
    
    app.Get("/users/:id", handler.GetUser)
    
    // Execute
    req := httptest.NewRequest("GET", "/users/1", nil)
    resp, err := app.Test(req)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
    
    var user domain.User
    err = json.NewDecoder(resp.Body).Decode(&user)
    assert.NoError(t, err)
    assert.Equal(t, expectedUser.Name, user.Name)
    
    mockService.AssertExpectations(t)
}

func TestUserHandler_CreateUser(t *testing.T) {
    // Setup
    app := fiber.New()
    mockService := new(MockUserService)
    mockLogger := new(MockLogger)
    
    handler := NewUserHandler(mockService, mockLogger)
    
    requestData := map[string]string{
        "name":  "John Doe",
        "email": "john@example.com",
    }
    
    expectedUser := &domain.User{ID: 1, Name: "John Doe", Email: "john@example.com"}
    mockService.On("CreateUser", "John Doe", "john@example.com").Return(expectedUser, nil)
    
    app.Post("/users", handler.CreateUser)
    
    // Execute
    body, _ := json.Marshal(requestData)
    req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := app.Test(req)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, 201, resp.StatusCode)
    
    var user domain.User
    err = json.NewDecoder(resp.Body).Decode(&user)
    assert.NoError(t, err)
    assert.Equal(t, expectedUser.Name, user.Name)
    
    mockService.AssertExpectations(t)
}
```

## Integration Testing

### Testing with Real Database

```go
package integration

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "your-project/internal/domain"
    "your-project/internal/repository"
    "your-project/internal/service"
)

func setupIntegrationTest(t *testing.T) (*gorm.DB, func()) {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    
    // Auto-migrate
    err = db.AutoMigrate(&domain.User{})
    require.NoError(t, err)
    
    cleanup := func() {
        sqlDB, _ := db.DB()
        sqlDB.Close()
    }
    
    return db, cleanup
}

func TestUserService_Integration(t *testing.T) {
    // Setup
    db, cleanup := setupIntegrationTest(t)
    defer cleanup()
    
    mockDB := &MockDB{instance: db}
    mockLogger := new(MockLogger)
    
    repo := repository.NewUserRepository(mockDB)
    service := service.NewUserService(repo, mockLogger)
    
    // Test create user
    user, err := service.CreateUser("John Doe", "john@example.com")
    assert.NoError(t, err)
    assert.NotZero(t, user.ID)
    
    // Test get user
    found, err := service.GetUser(int(user.ID))
    assert.NoError(t, err)
    assert.Equal(t, user.Name, found.Name)
    assert.Equal(t, user.Email, found.Email)
}
```

### Testing with GOE Application

```go
package integration

import (
    "context"
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "go.uber.org/fx"
    "go.oease.dev/goe/v2"
    "your-project/internal/handler"
    "your-project/internal/service"
    "your-project/internal/repository"
)

func TestGOEApplication(t *testing.T) {
    // Setup test app
    app := goe.New(goe.Options{
        WithHTTP: true,
        WithDB:   true,
        Providers: []any{
            repository.NewUserRepository,
            service.NewUserService,
            handler.NewUserHandler,
        },
        Invokers: []any{
            func(handler *handler.UserHandler) {
                // Register test routes
                app := goe.HTTP().App()
                app.Get("/users/:id", handler.GetUser)
                app.Post("/users", handler.CreateUser)
            },
        },
    })
    
    // Start app
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    err := app.Start(ctx)
    assert.NoError(t, err)
    
    // Test HTTP endpoints
    // ... make HTTP requests to test endpoints
    
    // Stop app
    err = app.Stop(ctx)
    assert.NoError(t, err)
}
```

## Contract Testing

### Testing Interface Implementations

```go
package contract

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "your-project/internal/domain"
    "your-project/internal/repository"
)

// Test that repository implements the contract
func TestUserRepository_ImplementsContract(t *testing.T) {
    db := setupTestDB(t)
    mockDB := &MockDB{instance: db}
    
    var repo domain.UserRepository = repository.NewUserRepository(mockDB)
    
    // Test interface methods
    user := &domain.User{Name: "John Doe", Email: "john@example.com"}
    
    err := repo.Create(user)
    assert.NoError(t, err)
    
    found, err := repo.FindByID(int(user.ID))
    assert.NoError(t, err)
    assert.Equal(t, user.Name, found.Name)
}
```

## Test Fixtures

### Creating Test Data

```go
package fixtures

import (
    "time"
    "your-project/internal/domain"
)

func NewUser(overrides ...func(*domain.User)) *domain.User {
    user := &domain.User{
        Name:      "John Doe",
        Email:     "john@example.com",
        Active:    true,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    for _, override := range overrides {
        override(user)
    }
    
    return user
}

func NewUserWithID(id uint) *domain.User {
    return NewUser(func(u *domain.User) {
        u.ID = id
    })
}

func NewInactiveUser() *domain.User {
    return NewUser(func(u *domain.User) {
        u.Active = false
    })
}

// Usage in tests
func TestUserService_WithFixtures(t *testing.T) {
    user := fixtures.NewUser()
    inactiveUser := fixtures.NewInactiveUser()
    
    // Use fixtures in tests
    assert.Equal(t, "John Doe", user.Name)
    assert.False(t, inactiveUser.Active)
}
```

## Test Utilities

### Mock Builders

```go
package testutil

import (
    "github.com/stretchr/testify/mock"
    "your-project/internal/domain"
)

type MockUserServiceBuilder struct {
    mock *MockUserService
}

func NewMockUserService() *MockUserServiceBuilder {
    return &MockUserServiceBuilder{
        mock: new(MockUserService),
    }
}

func (b *MockUserServiceBuilder) WithGetUser(id int, user *domain.User, err error) *MockUserServiceBuilder {
    b.mock.On("GetUser", id).Return(user, err)
    return b
}

func (b *MockUserServiceBuilder) WithCreateUser(name, email string, user *domain.User, err error) *MockUserServiceBuilder {
    b.mock.On("CreateUser", name, email).Return(user, err)
    return b
}

func (b *MockUserServiceBuilder) Build() *MockUserService {
    return b.mock
}

// Usage in tests
func TestUserHandler_WithBuilder(t *testing.T) {
    user := &domain.User{ID: 1, Name: "John Doe"}
    
    mockService := NewMockUserService().
        WithGetUser(1, user, nil).
        Build()
    
    handler := NewUserHandler(mockService, new(MockLogger))
    
    // Test handler
    // ...
}
```

## Test Configuration

### Environment Variables

```go
package testutil

import (
    "os"
    "testing"
)

func SetupTestEnv(t *testing.T) func() {
    // Set test environment variables
    os.Setenv("APP_ENV", "test")
    os.Setenv("LOG_LEVEL", "error")
    os.Setenv("DB_DRIVER", "sqlite")
    os.Setenv("DB_PATH", ":memory:")
    
    return func() {
        // Cleanup
        os.Unsetenv("APP_ENV")
        os.Unsetenv("LOG_LEVEL")
        os.Unsetenv("DB_DRIVER")
        os.Unsetenv("DB_PATH")
    }
}

func TestWithEnv(t *testing.T) {
    cleanup := SetupTestEnv(t)
    defer cleanup()
    
    // Test with environment variables set
    // ...
}
```

## Benchmarking

### Performance Testing

```go
package service

import (
    "testing"
    "your-project/internal/domain"
)

func BenchmarkUserService_GetUser(b *testing.B) {
    // Setup
    mockRepo := new(MockUserRepository)
    mockLogger := new(MockLogger)
    
    user := &domain.User{ID: 1, Name: "John Doe"}
    mockRepo.On("FindByID", 1).Return(user, nil)
    mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything)
    
    service := NewUserService(mockRepo, mockLogger)
    
    // Benchmark
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.GetUser(1)
    }
}
```

## Test Coverage

### Running Tests with Coverage

```bash
# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Check coverage percentage
go tool cover -func=coverage.out
```

### Coverage Configuration

```makefile
# Makefile
test:
	go test -v -race ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

test-integration:
	go test -v -race -tags=integration ./tests/integration/...

.PHONY: test test-coverage test-integration
```

## Best Practices

### 1. Use Table-Driven Tests

```go
func TestUserService_GetUser(t *testing.T) {
    tests := []struct {
        name        string
        userID      int
        mockUser    *domain.User
        mockError   error
        expectError bool
    }{
        {
            name:        "successful get",
            userID:      1,
            mockUser:    &domain.User{ID: 1, Name: "John Doe"},
            mockError:   nil,
            expectError: false,
        },
        {
            name:        "user not found",
            userID:      999,
            mockUser:    nil,
            mockError:   domain.ErrUserNotFound,
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(MockUserRepository)
            mockLogger := new(MockLogger)
            
            mockRepo.On("FindByID", tt.userID).Return(tt.mockUser, tt.mockError)
            mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything)
            
            service := NewUserService(mockRepo, mockLogger)
            
            user, err := service.GetUser(tt.userID)
            
            if tt.expectError {
                assert.Error(t, err)
                assert.Nil(t, user)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.mockUser, user)
            }
            
            mockRepo.AssertExpectations(t)
        })
    }
}
```

### 2. Test Edge Cases

```go
func TestUserService_EdgeCases(t *testing.T) {
    // Test with nil repository
    service := NewUserService(nil, new(MockLogger))
    user, err := service.GetUser(1)
    assert.Error(t, err)
    assert.Nil(t, user)
    
    // Test with empty ID
    mockRepo := new(MockUserRepository)
    mockLogger := new(MockLogger)
    service = NewUserService(mockRepo, mockLogger)
    
    user, err = service.GetUser(0)
    assert.Error(t, err)
    assert.Nil(t, user)
}
```

### 3. Test Error Scenarios

```go
func TestUserService_DatabaseError(t *testing.T) {
    mockRepo := new(MockUserRepository)
    mockLogger := new(MockLogger)
    
    dbError := errors.New("database connection failed")
    mockRepo.On("FindByID", 1).Return(nil, dbError)
    mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything)
    mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
    
    service := NewUserService(mockRepo, mockLogger)
    
    user, err := service.GetUser(1)
    
    assert.Error(t, err)
    assert.Nil(t, user)
    assert.Equal(t, dbError, err)
    
    mockRepo.AssertExpectations(t)
    mockLogger.AssertExpectations(t)
}
```

## Continuous Integration

### GitHub Actions

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

## Next Steps

- [**Best Practices**](./best-practices.md) - Learn development best practices
- [**Deployment**](./deployment.md) - Deploy your tested application
- [**Observability**](./observability.md) - Monitor your application in production