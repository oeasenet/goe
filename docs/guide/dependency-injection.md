# Dependency Injection

GOE is built entirely on Uber's Fx dependency injection framework. Understanding how to use Fx effectively is crucial for building maintainable and testable GOE applications.

## Overview

Dependency injection in GOE provides:
- **Automatic Dependency Resolution**: Fx automatically wires dependencies
- **Lifecycle Management**: Proper startup and shutdown ordering
- **Testability**: Easy mocking and testing of components
- **Modularity**: Clear separation of concerns

## Basic Concepts

### Providers

Providers are constructor functions that tell Fx how to create dependencies:

```go
// Provider function
func NewUserService(db contract.DB, logger contract.Logger) *UserService {
    return &UserService{db: db, logger: logger}
}

// Register provider
goe.New(goe.Options{
    Providers: []any{NewUserService},
})
```

### Invokers

Invokers are functions that are called at startup with their dependencies injected:

```go
// Invoker function
func SetupRoutes(httpKernel contract.HTTPKernel, userService *UserService) {
    app := httpKernel.App()
    app.Get("/users", userService.GetUsers)
}

// Register invoker
goe.New(goe.Options{
    Invokers: []any{SetupRoutes},
})
```

## Dependency Patterns

### Simple Dependency

```go
type UserService struct {
    logger contract.Logger
}

func NewUserService(logger contract.Logger) *UserService {
    return &UserService{logger: logger}
}

func (s *UserService) GetUsers() []User {
    s.logger.Info("Getting users")
    return []User{}
}
```

### Multiple Dependencies

```go
type UserService struct {
    db     contract.DB
    logger contract.Logger
    cache  contract.Cache
}

func NewUserService(db contract.DB, logger contract.Logger, cache contract.Cache) *UserService {
    return &UserService{
        db:     db,
        logger: logger,
        cache:  cache,
    }
}
```

### Interface Dependencies

```go
type EmailService interface {
    SendEmail(to, subject, body string) error
}

type UserService struct {
    emailService EmailService
    logger       contract.Logger
}

func NewUserService(emailService EmailService, logger contract.Logger) *UserService {
    return &UserService{
        emailService: emailService,
        logger:       logger,
    }
}
```

## Lifecycle Hooks

### OnStart Hook

```go
func NewDatabaseService(config contract.Config, logger contract.Logger) *DatabaseService {
    return &DatabaseService{config: config, logger: logger}
}

func (s *DatabaseService) OnStart(ctx context.Context) error {
    s.logger.Info("Connecting to database")
    
    // Connect to database
    conn, err := sql.Open("postgres", s.config.GetString("DATABASE_URL"))
    if err != nil {
        return err
    }
    
    s.connection = conn
    return nil
}

// Register with lifecycle
goe.New(goe.Options{
    Providers: []any{NewDatabaseService},
    Invokers: []any{
        func(lc fx.Lifecycle, dbService *DatabaseService) {
            lc.Append(fx.Hook{
                OnStart: dbService.OnStart,
                OnStop:  dbService.OnStop,
            })
        },
    },
})
```

### OnStop Hook

```go
func (s *DatabaseService) OnStop(ctx context.Context) error {
    s.logger.Info("Closing database connection")
    
    if s.connection != nil {
        return s.connection.Close()
    }
    
    return nil
}
```

## Advanced Patterns

### Optional Dependencies

```go
type UserService struct {
    db     contract.DB
    logger contract.Logger
    cache  contract.Cache // Optional
}

func NewUserService(db contract.DB, logger contract.Logger, cache contract.Cache) *UserService {
    return &UserService{
        db:     db,
        logger: logger,
        cache:  cache,
    }
}

// Make cache optional
type UserServiceParams struct {
    fx.In
    DB     contract.DB
    Logger contract.Logger
    Cache  contract.Cache `optional:"true"`
}

func NewUserService(params UserServiceParams) *UserService {
    return &UserService{
        db:     params.DB,
        logger: params.Logger,
        cache:  params.Cache,
    }
}
```

### Named Dependencies

```go
type DatabaseConfig struct {
    Primary   string `name:"primary"`
    Secondary string `name:"secondary"`
}

func NewPrimaryDatabase(config contract.Config) *sql.DB {
    // Connect to primary database
    conn, _ := sql.Open("postgres", config.GetString("PRIMARY_DB_URL"))
    return conn
}

func NewSecondaryDatabase(config contract.Config) *sql.DB {
    // Connect to secondary database
    conn, _ := sql.Open("postgres", config.GetString("SECONDARY_DB_URL"))
    return conn
}

type UserService struct {
    primaryDB   *sql.DB `name:"primary"`
    secondaryDB *sql.DB `name:"secondary"`
}

func NewUserService(primaryDB *sql.DB, secondaryDB *sql.DB) *UserService {
    return &UserService{
        primaryDB:   primaryDB,
        secondaryDB: secondaryDB,
    }
}

// Register named providers
goe.New(goe.Options{
    Providers: []any{
        fx.Annotate(NewPrimaryDatabase, fx.ResultTags(`name:"primary"`)),
        fx.Annotate(NewSecondaryDatabase, fx.ResultTags(`name:"secondary"`)),
        fx.Annotate(NewUserService, fx.ParamTags(`name:"primary"`, `name:"secondary"`)),
    },
})
```

### Group Dependencies

```go
type Handler interface {
    Pattern() string
    Handle(c fiber.Ctx) error
}

type UserHandler struct{}

func (h *UserHandler) Pattern() string { return "/users" }
func (h *UserHandler) Handle(c fiber.Ctx) error { return c.SendString("Users") }

type PostHandler struct{}

func (h *PostHandler) Pattern() string { return "/posts" }
func (h *PostHandler) Handle(c fiber.Ctx) error { return c.SendString("Posts") }

func NewUserHandler() Handler {
    return &UserHandler{}
}

func NewPostHandler() Handler {
    return &PostHandler{}
}

type RouterParams struct {
    fx.In
    Handlers []Handler `group:"handlers"`
}

func SetupRoutes(httpKernel contract.HTTPKernel, params RouterParams) {
    app := httpKernel.App()
    
    for _, handler := range params.Handlers {
        app.Get(handler.Pattern(), handler.Handle)
    }
}

// Register with groups
goe.New(goe.Options{
    Providers: []any{
        fx.Annotate(NewUserHandler, fx.As(new(Handler)), fx.ResultTags(`group:"handlers"`)),
        fx.Annotate(NewPostHandler, fx.As(new(Handler)), fx.ResultTags(`group:"handlers"`)),
    },
    Invokers: []any{SetupRoutes},
})
```

## Testing with Dependency Injection

### Unit Testing

```go
func TestUserService(t *testing.T) {
    // Create mocks
    mockDB := &MockDB{}
    mockLogger := &MockLogger{}
    
    // Create service with mocked dependencies
    service := NewUserService(mockDB, mockLogger)
    
    // Test the service
    users := service.GetUsers()
    assert.NotNil(t, users)
    
    // Verify mock calls
    mockLogger.AssertCalled(t, "Info", "Getting users")
}
```

### Integration Testing

```go
func TestUserService_Integration(t *testing.T) {
    app := fx.New(
        fx.Provide(
            NewUserService,
            NewMockDB,
            NewMockLogger,
        ),
        fx.Invoke(func(service *UserService) {
            // Test service with real dependencies
            users := service.GetUsers()
            assert.NotNil(t, users)
        }),
    )
    
    ctx := context.Background()
    
    err := app.Start(ctx)
    assert.NoError(t, err)
    
    err = app.Stop(ctx)
    assert.NoError(t, err)
}
```

### Test Doubles

```go
type MockDB struct {
    users []User
}

func (m *MockDB) Instance() *gorm.DB {
    // Return mock GORM instance
    return nil
}

func NewMockDB() contract.DB {
    return &MockDB{
        users: []User{
            {ID: 1, Name: "John Doe"},
            {ID: 2, Name: "Jane Smith"},
        },
    }
}

type MockLogger struct {
    logs []string
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
    m.logs = append(m.logs, msg)
}

func NewMockLogger() contract.Logger {
    return &MockLogger{}
}
```

## Best Practices

### 1. Use Interfaces

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
    db     *sql.DB
    logger *zap.Logger
}
```

### 2. Keep Constructors Simple

```go
// Good - simple constructor
func NewUserService(db contract.DB, logger contract.Logger) *UserService {
    return &UserService{db: db, logger: logger}
}

// Avoid - complex initialization in constructor
func NewUserService(config contract.Config, logger contract.Logger) *UserService {
    // Complex initialization logic
    db, err := sql.Open("postgres", config.GetString("DATABASE_URL"))
    if err != nil {
        logger.Fatal("Failed to connect to database", "error", err)
    }
    
    return &UserService{db: db, logger: logger}
}
```

### 3. Use Lifecycle Hooks for Setup

```go
func NewDatabaseService(config contract.Config, logger contract.Logger) *DatabaseService {
    return &DatabaseService{config: config, logger: logger}
}

func (s *DatabaseService) OnStart(ctx context.Context) error {
    // Initialize database connection
    conn, err := sql.Open("postgres", s.config.GetString("DATABASE_URL"))
    if err != nil {
        return err
    }
    
    s.connection = conn
    return nil
}
```

### 4. Use Groups for Similar Dependencies

```go
type Middleware func(fiber.Handler) fiber.Handler

func NewAuthMiddleware() Middleware {
    return func(next fiber.Handler) fiber.Handler {
        return func(c fiber.Ctx) error {
            // Auth logic
            return next(c)
        }
    }
}

func NewLoggingMiddleware() Middleware {
    return func(next fiber.Handler) fiber.Handler {
        return func(c fiber.Ctx) error {
            // Logging logic
            return next(c)
        }
    }
}

type MiddlewareParams struct {
    fx.In
    Middlewares []Middleware `group:"middlewares"`
}

func SetupMiddleware(httpKernel contract.HTTPKernel, params MiddlewareParams) {
    app := httpKernel.App()
    
    for _, middleware := range params.Middlewares {
        app.Use(middleware)
    }
}
```

## Error Handling

### Constructor Errors

```go
func NewUserService(db contract.DB, logger contract.Logger) (*UserService, error) {
    // Validate dependencies
    if db == nil {
        return nil, errors.New("database is required")
    }
    
    if logger == nil {
        return nil, errors.New("logger is required")
    }
    
    return &UserService{db: db, logger: logger}, nil
}
```

### Lifecycle Errors

```go
func (s *DatabaseService) OnStart(ctx context.Context) error {
    conn, err := sql.Open("postgres", s.config.GetString("DATABASE_URL"))
    if err != nil {
        return fmt.Errorf("failed to connect to database: %w", err)
    }
    
    if err := conn.PingContext(ctx); err != nil {
        return fmt.Errorf("failed to ping database: %w", err)
    }
    
    s.connection = conn
    return nil
}
```

## Common Patterns

### Service Layer

```go
type UserService struct {
    repository UserRepository
    logger     contract.Logger
}

func NewUserService(repository UserRepository, logger contract.Logger) *UserService {
    return &UserService{repository: repository, logger: logger}
}

func (s *UserService) GetUser(id int) (*User, error) {
    s.logger.Info("Getting user", "id", id)
    
    user, err := s.repository.FindByID(id)
    if err != nil {
        s.logger.Error("Failed to get user", "id", id, "error", err)
        return nil, err
    }
    
    return user, nil
}
```

### Repository Layer

```go
type UserRepository interface {
    FindByID(id int) (*User, error)
    Create(user *User) error
    Update(user *User) error
    Delete(id int) error
}

type userRepository struct {
    db contract.DB
}

func NewUserRepository(db contract.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) FindByID(id int) (*User, error) {
    var user User
    err := r.db.Instance().First(&user, id).Error
    return &user, err
}
```

### Handler Layer

```go
type UserHandler struct {
    userService *UserService
    logger      contract.Logger
}

func NewUserHandler(userService *UserService, logger contract.Logger) *UserHandler {
    return &UserHandler{userService: userService, logger: logger}
}

func (h *UserHandler) GetUser(c fiber.Ctx) error {
    id, err := c.ParamsInt("id")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
    }
    
    user, err := h.userService.GetUser(id)
    if err != nil {
        h.logger.Error("Failed to get user", "id", id, "error", err)
        return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
    }
    
    return c.JSON(user)
}
```

## Complete Example

```go
package main

import (
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
)

// Domain
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

// Repository
type UserRepository interface {
    FindAll() ([]User, error)
    FindByID(id int) (*User, error)
}

type userRepository struct {
    db contract.DB
}

func NewUserRepository(db contract.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) FindAll() ([]User, error) {
    var users []User
    err := r.db.Instance().Find(&users).Error
    return users, err
}

func (r *userRepository) FindByID(id int) (*User, error) {
    var user User
    err := r.db.Instance().First(&user, id).Error
    return &user, err
}

// Service
type UserService struct {
    repository UserRepository
    logger     contract.Logger
}

func NewUserService(repository UserRepository, logger contract.Logger) *UserService {
    return &UserService{repository: repository, logger: logger}
}

func (s *UserService) GetAllUsers() ([]User, error) {
    s.logger.Info("Getting all users")
    return s.repository.FindAll()
}

func (s *UserService) GetUser(id int) (*User, error) {
    s.logger.Info("Getting user", "id", id)
    return s.repository.FindByID(id)
}

// Handler
type UserHandler struct {
    userService *UserService
}

func NewUserHandler(userService *UserService) *UserHandler {
    return &UserHandler{userService: userService}
}

func (h *UserHandler) GetUsers(c fiber.Ctx) error {
    users, err := h.userService.GetAllUsers()
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
    }
    return c.JSON(users)
}

func (h *UserHandler) GetUser(c fiber.Ctx) error {
    id, err := c.ParamsInt("id")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
    }
    
    user, err := h.userService.GetUser(id)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
    }
    
    return c.JSON(user)
}

// Routes
func RegisterRoutes(httpKernel contract.HTTPKernel, userHandler *UserHandler) {
    app := httpKernel.App()
    
    api := app.Group("/api")
    api.Get("/users", userHandler.GetUsers)
    api.Get("/users/:id", userHandler.GetUser)
}

// Main
func main() {
    goe.New(goe.Options{
        WithHTTP: true,
        WithDB:   true,
        Providers: []any{
            NewUserRepository,
            NewUserService,
            NewUserHandler,
        },
        Invokers: []any{
            RegisterRoutes,
        },
    })
    
    goe.Run()
}
```

## Next Steps

- [**Testing**](./testing.md) - Test your dependency-injected components
- [**Modules**](./modules.md) - Create custom modules with dependency injection
- [**Best Practices**](./best-practices.md) - Learn development best practices