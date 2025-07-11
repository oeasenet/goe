# Best Practices

This guide outlines best practices for developing applications with the GOE framework, covering architecture, performance, security, and maintainability.

## Architecture Best Practices

### 1. Dependency Injection

**Use interfaces for dependencies:**

```go
// Good - depend on interfaces
type UserService struct {
    repository UserRepository
    logger     contract.Logger
}

func NewUserService(repository UserRepository, logger contract.Logger) *UserService {
    return &UserService{repository: repository, logger: logger}
}

// Avoid - depending on concrete types
type UserService struct {
    db     *gorm.DB
    logger *zap.Logger
}
```

**Keep constructors simple:**

```go
// Good - simple constructor
func NewUserService(repository UserRepository, logger contract.Logger) *UserService {
    return &UserService{repository: repository, logger: logger}
}

// Avoid - complex initialization
func NewUserService(config contract.Config, logger contract.Logger) *UserService {
    dsn := config.GetString("DATABASE_URL")
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        logger.Fatal("Database connection failed", "error", err)
    }
    
    return &UserService{db: db, logger: logger}
}
```

### 2. Layer Separation

**Follow the layered architecture:**

```go
// Domain layer - business logic
type User struct {
    ID    int
    Name  string
    Email string
}

type UserRepository interface {
    Create(user *User) error
    FindByID(id int) (*User, error)
}

// Service layer - application logic
type UserService struct {
    repository UserRepository
    logger     contract.Logger
}

func (s *UserService) CreateUser(name, email string) (*User, error) {
    // Validation
    if name == "" {
        return nil, errors.New("name is required")
    }
    
    // Business logic
    user := &User{Name: name, Email: email}
    
    if err := s.repository.Create(user); err != nil {
        s.logger.Error("Failed to create user", "error", err)
        return nil, err
    }
    
    return user, nil
}

// Handler layer - presentation logic
type UserHandler struct {
    service *UserService
}

func (h *UserHandler) CreateUser(c fiber.Ctx) error {
    var req CreateUserRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    user, err := h.service.CreateUser(req.Name, req.Email)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
    }
    
    return c.Status(201).JSON(user)
}
```

### 3. Error Handling

**Create custom error types:**

```go
type UserError struct {
    Code    string
    Message string
    Cause   error
}

func (e *UserError) Error() string {
    return e.Message
}

func (e *UserError) Unwrap() error {
    return e.Cause
}

var (
    ErrUserNotFound     = &UserError{Code: "USER_NOT_FOUND", Message: "User not found"}
    ErrUserAlreadyExists = &UserError{Code: "USER_EXISTS", Message: "User already exists"}
    ErrInvalidEmail     = &UserError{Code: "INVALID_EMAIL", Message: "Invalid email format"}
)
```

**Handle errors at appropriate levels:**

```go
// Repository level - return domain errors
func (r *userRepository) FindByID(id int) (*User, error) {
    var user User
    err := r.db.Instance().First(&user, id).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrUserNotFound
        }
        return nil, err
    }
    
    return &user, nil
}

// Service level - add context and logging
func (s *UserService) GetUser(id int) (*User, error) {
    s.logger.Info("Getting user", "id", id)
    
    user, err := s.repository.FindByID(id)
    if err != nil {
        s.logger.Error("Failed to get user", "id", id, "error", err)
        return nil, err
    }
    
    return user, nil
}

// Handler level - convert to HTTP responses
func (h *UserHandler) GetUser(c fiber.Ctx) error {
    id, err := c.ParamsInt("id")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
    }
    
    user, err := h.service.GetUser(id)
    if err != nil {
        var userErr *UserError
        if errors.As(err, &userErr) {
            switch userErr.Code {
            case "USER_NOT_FOUND":
                return c.Status(404).JSON(fiber.Map{"error": userErr.Message})
            }
        }
        return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
    }
    
    return c.JSON(user)
}
```

## Performance Best Practices

### 1. Database Optimization

**Use indexes:**

```go
type User struct {
    ID    uint   `gorm:"primaryKey"`
    Email string `gorm:"size:100;uniqueIndex;not null"`
    Name  string `gorm:"size:100;index"`
}
```

**Use pagination:**

```go
func (r *userRepository) FindAllWithPagination(page, limit int) ([]User, int64, error) {
    var users []User
    var total int64
    
    offset := (page - 1) * limit
    
    if err := r.db.Instance().Model(&User{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    err := r.db.Instance().
        Offset(offset).
        Limit(limit).
        Order("created_at DESC").
        Find(&users).Error
    
    return users, total, err
}
```

**Use select to limit fields:**

```go
func (r *userRepository) FindAllBasicInfo() ([]User, error) {
    var users []User
    err := r.db.Instance().
        Select("id", "name", "email").
        Find(&users).Error
    return users, err
}
```

### 2. Caching Strategies

**Cache expensive operations:**

```go
func (s *UserService) GetUserProfile(id int) (*UserProfile, error) {
    cacheKey := fmt.Sprintf("user_profile:%d", id)
    
    // Try cache first
    if cached, err := s.cache.Get(cacheKey); err == nil {
        if profile, ok := cached.(*UserProfile); ok {
            return profile, nil
        }
    }
    
    // Fetch from database
    profile, err := s.buildUserProfile(id)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    s.cache.Set(cacheKey, profile, 15*time.Minute)
    
    return profile, nil
}
```

**Cache invalidation:**

```go
func (s *UserService) UpdateUser(id int, updates map[string]interface{}) error {
    if err := s.repository.Update(id, updates); err != nil {
        return err
    }
    
    // Invalidate cache
    s.cache.Delete(fmt.Sprintf("user_profile:%d", id))
    s.cache.Delete(fmt.Sprintf("user:%d", id))
    
    return nil
}
```

### 3. HTTP Performance

**Use connection pooling:**

```bash
# .env
HTTP_READ_TIMEOUT=30s
HTTP_WRITE_TIMEOUT=30s
HTTP_IDLE_TIMEOUT=60s
```

**Implement rate limiting:**

```go
import "github.com/gofiber/fiber/v3/middleware/limiter"

func SetupRateLimiting(app *fiber.App) {
    app.Use(limiter.New(limiter.Config{
        Max:        100,
        Expiration: 1 * time.Minute,
    }))
}
```

## Security Best Practices

### 1. Input Validation

**Validate all input:**

```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2,max=100"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=0,max=150"`
}

func (h *UserHandler) CreateUser(c fiber.Ctx) error {
    var req CreateUserRequest
    
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
    }
    
    if err := validate.Struct(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Validation failed",
            "details": err.Error(),
        })
    }
    
    // Process request
    return nil
}
```

**Sanitize input:**

```go
import "html"

func sanitizeInput(input string) string {
    // Remove HTML tags
    input = html.EscapeString(input)
    
    // Trim whitespace
    input = strings.TrimSpace(input)
    
    return input
}
```

### 2. Authentication & Authorization

**Use middleware for authentication:**

```go
func AuthMiddleware(c fiber.Ctx) error {
    token := c.Get("Authorization")
    if token == "" {
        return c.Status(401).JSON(fiber.Map{"error": "Token required"})
    }
    
    claims, err := validateToken(token)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
    }
    
    c.Locals("user_id", claims.UserID)
    return c.Next()
}
```

**Implement role-based access:**

```go
func RequireRole(role string) fiber.Handler {
    return func(c fiber.Ctx) error {
        userID := c.Locals("user_id").(int)
        
        user, err := getUserByID(userID)
        if err != nil {
            return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
        }
        
        if user.Role != role {
            return c.Status(403).JSON(fiber.Map{"error": "Insufficient permissions"})
        }
        
        return c.Next()
    }
}
```

### 3. Data Protection

**Hash passwords:**

```go
import "golang.org/x/crypto/bcrypt"

func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func checkPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

**Don't log sensitive data:**

```go
// Good - don't log sensitive information
logger.Info("User login attempt", "email", user.Email)

// Bad - logging sensitive data
logger.Info("User login", "email", user.Email, "password", password)
```

## Configuration Best Practices

### 1. Environment-Based Configuration

**Use environment variables:**

```go
type Config struct {
    Port        int    `env:"HTTP_PORT" default:"8080"`
    DatabaseURL string `env:"DATABASE_URL" required:"true"`
    JWTSecret   string `env:"JWT_SECRET" required:"true"`
    Debug       bool   `env:"DEBUG" default:"false"`
}
```

**Validate configuration:**

```go
func ValidateConfig(config contract.Config) error {
    required := []string{
        "DATABASE_URL",
        "JWT_SECRET",
    }
    
    for _, key := range required {
        if !config.Has(key) {
            return fmt.Errorf("required configuration missing: %s", key)
        }
    }
    
    return nil
}
```

### 2. Configuration Management

**Use configuration structs:**

```go
type DatabaseConfig struct {
    Host     string
    Port     int
    Database string
    Username string
    Password string
}

func NewDatabaseConfig(config contract.Config) *DatabaseConfig {
    return &DatabaseConfig{
        Host:     config.GetString("DB_HOST"),
        Port:     config.GetInt("DB_PORT"),
        Database: config.GetString("DB_NAME"),
        Username: config.GetString("DB_USER"),
        Password: config.GetString("DB_PASSWORD"),
    }
}
```

## Testing Best Practices

### 1. Test Structure

**Use AAA pattern (Arrange, Act, Assert):**

```go
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := new(MockUserRepository)
    mockLogger := new(MockLogger)
    service := NewUserService(mockRepo, mockLogger)
    
    expectedUser := &User{ID: 1, Name: "John Doe", Email: "john@example.com"}
    mockRepo.On("Create", mock.AnythingOfType("*User")).Return(nil)
    
    // Act
    user, err := service.CreateUser("John Doe", "john@example.com")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expectedUser.Name, user.Name)
    assert.Equal(t, expectedUser.Email, user.Email)
    mockRepo.AssertExpectations(t)
}
```

**Use table-driven tests:**

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name     string
        email    string
        expected bool
    }{
        {"valid email", "user@example.com", true},
        {"invalid email", "invalid-email", false},
        {"empty email", "", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := validateEmail(tt.email)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 2. Test Coverage

**Aim for high test coverage:**

```bash
# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Set coverage threshold
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//' | awk '{if($1 < 80) exit 1}'
```

## Monitoring Best Practices

### 1. Structured Logging

**Use consistent log levels:**

```go
// Debug - detailed information for debugging
logger.Debug("Processing user request", "user_id", userID, "action", "create")

// Info - general information
logger.Info("User created successfully", "user_id", userID)

// Warn - potentially harmful situations
logger.Warn("High memory usage detected", "usage", memUsage)

// Error - error events
logger.Error("Failed to process request", "error", err, "user_id", userID)
```

**Include context in logs:**

```go
logger.Info("Processing order",
    "order_id", orderID,
    "user_id", userID,
    "amount", amount,
    "payment_method", paymentMethod,
    "request_id", requestID,
)
```

### 2. Metrics and Monitoring

**Track key metrics:**

```go
type Metrics struct {
    RequestCount    prometheus.Counter
    RequestDuration prometheus.Histogram
    ErrorCount      prometheus.Counter
}

func (m *Metrics) RecordRequest(duration time.Duration, success bool) {
    m.RequestCount.Inc()
    m.RequestDuration.Observe(duration.Seconds())
    
    if !success {
        m.ErrorCount.Inc()
    }
}
```

## Deployment Best Practices

### 1. Docker Configuration

**Use multi-stage builds:**

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/.env .

EXPOSE 8080
CMD ["./main"]
```

### 2. Health Checks

**Implement health endpoints:**

```go
func (h *HealthHandler) CheckHealth(c fiber.Ctx) error {
    checks := map[string]interface{}{
        "database": h.checkDatabase(),
        "cache":    h.checkCache(),
        "status":   "healthy",
    }
    
    return c.JSON(checks)
}

func (h *HealthHandler) checkDatabase() interface{} {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := h.db.Instance().WithContext(ctx).Exec("SELECT 1").Error; err != nil {
        return map[string]interface{}{
            "status": "unhealthy",
            "error":  err.Error(),
        }
    }
    
    return map[string]interface{}{
        "status": "healthy",
    }
}
```

### 3. Configuration Management

**Use environment-specific configs:**

```bash
# .env.production
APP_ENV=production
LOG_LEVEL=warn
LOG_FORMAT=json

# Database
DB_HOST=prod-db.example.com
DB_SSL_MODE=require

# Performance
HTTP_READ_TIMEOUT=30s
HTTP_WRITE_TIMEOUT=30s
```

## Code Quality Best Practices

### 1. Code Organization

**Use consistent naming:**

```go
// Good naming
type UserService struct{}
func (s *UserService) CreateUser() {}
func (s *UserService) GetUser() {}

// Consistent error naming
var (
    ErrUserNotFound     = errors.New("user not found")
    ErrUserAlreadyExists = errors.New("user already exists")
)
```

**Group related functionality:**

```go
// internal/user/
//   ├── handler.go
//   ├── service.go
//   ├── repository.go
//   └── models.go
```

### 2. Documentation

**Document public APIs:**

```go
// UserService provides user management functionality.
type UserService struct {
    repository UserRepository
    logger     contract.Logger
}

// CreateUser creates a new user with the given name and email.
// Returns ErrUserAlreadyExists if a user with the same email already exists.
func (s *UserService) CreateUser(name, email string) (*User, error) {
    // Implementation
}
```

**Use godoc conventions:**

```go
// Package user provides user management functionality.
//
// This package includes services for creating, updating, and retrieving users.
// It follows the repository pattern for data access and includes proper error handling.
package user
```

## Summary

Following these best practices will help you build:
- **Maintainable**: Code that's easy to understand and modify
- **Testable**: Components that can be easily tested in isolation
- **Performant**: Applications that scale well under load
- **Secure**: Systems that protect user data and prevent attacks
- **Reliable**: Applications that handle errors gracefully

Remember to:
1. Use dependency injection for better testability
2. Follow the layered architecture pattern
3. Implement proper error handling
4. Validate all inputs
5. Use caching strategically
6. Monitor your application
7. Write comprehensive tests
8. Keep security in mind at all times

## Next Steps

- [**Deployment**](./deployment.md) - Learn how to deploy your application
- [**Examples**](../examples/basic-app.md) - See practical examples of these practices