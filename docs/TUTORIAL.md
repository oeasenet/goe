# Goe Framework Tutorial

This tutorial will guide you through building a complete REST API using the Goe framework. We'll build a task management API with authentication, database integration, and proper architecture.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Project Structure](#project-structure)
3. [Basic Application](#basic-application)
4. [Configuration Management](#configuration-management)
5. [Logging](#logging)
6. [HTTP Routes and Handlers](#http-routes-and-handlers)
7. [Dependency Injection](#dependency-injection)
8. [Database Integration](#database-integration)
9. [Authentication](#authentication)
10. [Testing](#testing)
11. [Deployment](#deployment)

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Basic knowledge of Go
- PostgreSQL (for database examples)

### Installation

Create a new project:

```bash
mkdir task-api
cd task-api
go mod init github.com/yourusername/task-api
```

Install Goe:

```bash
go get go.oease.dev/goe/v2
```

## Project Structure

Organize your project following domain-driven design:

```
task-api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── auth/
│   │   ├── service.go
│   │   ├── controller.go
│   │   ├── middleware.go
│   │   └── jwt.go
│   ├── task/
│   │   ├── service.go
│   │   ├── controller.go
│   │   ├── repository.go
│   │   └── models.go
│   └── database/
│       └── postgres.go
├── pkg/
│   ├── validator/
│   │   └── validator.go
│   └── response/
│       └── response.go
├── migrations/
│   └── 001_create_tables.sql
├── .env
├── .env.production
├── Dockerfile
└── go.mod
```

## Basic Application

Let's start with a minimal Goe application:

```go
// cmd/api/main.go
package main

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
)

func main() {
    // Create application
    _ = goe.New(goe.Options{
        Name:        "Task API",
        Version:     "1.0.0",
        Environment: "dev",
        WithHTTP:    true,
        Invokers: []any{
            setupRoutes,
        },
    })
    
    // Start application
    goe.Log().Info("Starting Task API...")
    goe.Run()
}

func setupRoutes(http contract.HTTPKernel, logger contract.Logger) {
    app := http.App()
    
    // Health check
    app.Get("/health", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status": "healthy",
            "app":    goe.App().Name(),
            "version": goe.App().Version(),
        })
    })
    
    logger.Info("Routes configured")
}
```

Run the application:

```bash
go run cmd/api/main.go
```

Test it:

```bash
curl http://localhost:8080/health
```

## Configuration Management

### Environment Files

Create `.env` for development:

```env
# Application
APP_NAME=Task API
APP_ENV=development
DEBUG=true

# HTTP Server
HTTP_HOST=0.0.0.0
HTTP_PORT=8080
HTTP_READ_TIMEOUT=10s
HTTP_WRITE_TIMEOUT=10s

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=taskapi
DB_USER=postgres
DB_PASSWORD=secret
DB_SSL_MODE=disable

# JWT
JWT_SECRET=your-secret-key-change-this
JWT_EXPIRY=24h

# Logging
LOG_LEVEL=debug
LOG_FORMAT=console
```

Create `.env.production` for production:

```env
APP_ENV=production
DEBUG=false
LOG_FORMAT=json
DB_SSL_MODE=require
```

### Using Configuration

```go
// internal/database/postgres.go
package database

import (
    "database/sql"
    "fmt"
    
    _ "github.com/lib/pq"
    "go.oease.dev/goe/v2/contract"
    "go.uber.org/fx"
)

type DatabaseParams struct {
    fx.In
    Config contract.Config
    Logger contract.Logger
}

func NewDatabase(params DatabaseParams) (*sql.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        params.Config.GetString("DB_HOST"),
        params.Config.GetInt("DB_PORT"),
        params.Config.GetString("DB_USER"),
        params.Config.GetString("DB_PASSWORD"),
        params.Config.GetString("DB_NAME"),
        params.Config.GetString("DB_SSL_MODE"),
    )
    
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    
    if err := db.Ping(); err != nil {
        return nil, err
    }
    
    params.Logger.Info("Database connected")
    return db, nil
}
```

## Logging

### Structured Logging

```go
// internal/task/service.go
package task

import (
    "go.oease.dev/goe/v2/contract"
    "go.oease.dev/goe/v2/core/log"
)

type Service struct {
    logger contract.Logger
    repo   *Repository
}

func (s *Service) CreateTask(task *Task) error {
    s.logger.Info("Creating task",
        log.NewField("title", task.Title),
        log.NewField("user_id", task.UserID),
    )
    
    if err := s.repo.Create(task); err != nil {
        s.logger.Error("Failed to create task",
            log.NewField("error", err),
            log.NewField("user_id", task.UserID),
        )
        return err
    }
    
    s.logger.Debug("Task created successfully",
        log.NewField("task_id", task.ID),
    )
    
    return nil
}
```

### Request Logging

```go
// Automatic request logging is enabled
// Each request logs:
// - Method
// - Path
// - Status
// - Duration
// - Request ID

// Get request-scoped logger in handlers
func handler(c fiber.Ctx) error {
    logger := http.GetLogger(c) // Includes request_id
    logger.Info("Processing task creation")
    return nil
}
```

## HTTP Routes and Handlers

### Models

```go
// internal/task/models.go
package task

import (
    "time"
)

type Task struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Status      string    `json:"status"`
    UserID      string    `json:"user_id"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type CreateTaskRequest struct {
    Title       string `json:"title" validate:"required,min=3,max=100"`
    Description string `json:"description" validate:"max=500"`
}

type UpdateTaskRequest struct {
    Title       *string `json:"title" validate:"omitempty,min=3,max=100"`
    Description *string `json:"description" validate:"omitempty,max=500"`
    Status      *string `json:"status" validate:"omitempty,oneof=pending completed"`
}
```

### Controller

```go
// internal/task/controller.go
package task

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2/contract"
    "go.oease.dev/goe/v2/core/http"
    "go.uber.org/fx"
)

type Controller struct {
    service *Service
    logger  contract.Logger
}

type ControllerParams struct {
    fx.In
    Service *Service
    Logger  contract.Logger
}

func NewController(params ControllerParams) *Controller {
    return &Controller{
        service: params.Service,
        logger:  params.Logger,
    }
}

func (c *Controller) RegisterRoutes(app *fiber.App, authMiddleware fiber.Handler) {
    tasks := app.Group("/api/v1/tasks", authMiddleware)
    
    tasks.Get("/", c.ListTasks)
    tasks.Post("/", c.CreateTask)
    tasks.Get("/:id", c.GetTask)
    tasks.Put("/:id", c.UpdateTask)
    tasks.Delete("/:id", c.DeleteTask)
}

func (c *Controller) ListTasks(ctx fiber.Ctx) error {
    logger := http.GetLogger(ctx)
    userID := ctx.Locals("userID").(string)
    
    logger.Info("Listing tasks for user")
    
    tasks, err := c.service.ListByUser(userID)
    if err != nil {
        logger.Error("Failed to list tasks", log.NewField("error", err))
        return ctx.Status(500).JSON(fiber.Map{
            "error": "Failed to retrieve tasks",
        })
    }
    
    return ctx.JSON(fiber.Map{
        "tasks": tasks,
        "count": len(tasks),
    })
}

func (c *Controller) CreateTask(ctx fiber.Ctx) error {
    logger := http.GetLogger(ctx)
    
    var req CreateTaskRequest
    if err := ctx.Bind().Body(&req); err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }
    
    // Validate request
    if err := validate.Struct(req); err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "Validation failed",
            "details": formatValidationErrors(err),
        })
    }
    
    userID := ctx.Locals("userID").(string)
    
    task := &Task{
        Title:       req.Title,
        Description: req.Description,
        UserID:      userID,
        Status:      "pending",
    }
    
    if err := c.service.CreateTask(task); err != nil {
        logger.Error("Failed to create task", log.NewField("error", err))
        return ctx.Status(500).JSON(fiber.Map{
            "error": "Failed to create task",
        })
    }
    
    return ctx.Status(201).JSON(task)
}

func (c *Controller) GetTask(ctx fiber.Ctx) error {
    taskID := ctx.Params("id")
    userID := ctx.Locals("userID").(string)
    
    task, err := c.service.GetTask(taskID, userID)
    if err != nil {
        if err == ErrTaskNotFound {
            return ctx.Status(404).JSON(fiber.Map{
                "error": "Task not found",
            })
        }
        return ctx.Status(500).JSON(fiber.Map{
            "error": "Failed to retrieve task",
        })
    }
    
    return ctx.JSON(task)
}

func (c *Controller) UpdateTask(ctx fiber.Ctx) error {
    taskID := ctx.Params("id")
    userID := ctx.Locals("userID").(string)
    
    var req UpdateTaskRequest
    if err := ctx.Bind().Body(&req); err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }
    
    if err := validate.Struct(req); err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "Validation failed",
            "details": formatValidationErrors(err),
        })
    }
    
    task, err := c.service.UpdateTask(taskID, userID, &req)
    if err != nil {
        if err == ErrTaskNotFound {
            return ctx.Status(404).JSON(fiber.Map{
                "error": "Task not found",
            })
        }
        return ctx.Status(500).JSON(fiber.Map{
            "error": "Failed to update task",
        })
    }
    
    return ctx.JSON(task)
}

func (c *Controller) DeleteTask(ctx fiber.Ctx) error {
    taskID := ctx.Params("id")
    userID := ctx.Locals("userID").(string)
    
    if err := c.service.DeleteTask(taskID, userID); err != nil {
        if err == ErrTaskNotFound {
            return ctx.Status(404).JSON(fiber.Map{
                "error": "Task not found",
            })
        }
        return ctx.Status(500).JSON(fiber.Map{
            "error": "Failed to delete task",
        })
    }
    
    return ctx.Status(204).Send(nil)
}
```

## Dependency Injection

### Service Layer

```go
// internal/task/service.go
package task

import (
    "errors"
    
    "go.oease.dev/goe/v2/contract"
    "go.uber.org/fx"
)

var (
    ErrTaskNotFound = errors.New("task not found")
)

type Service struct {
    logger contract.Logger
    repo   *Repository
}

type ServiceParams struct {
    fx.In
    Logger     contract.Logger
    Repository *Repository
}

func NewService(params ServiceParams) *Service {
    return &Service{
        logger: params.Logger,
        repo:   params.Repository,
    }
}

func (s *Service) CreateTask(task *Task) error {
    task.ID = generateID()
    task.CreatedAt = time.Now()
    task.UpdatedAt = time.Now()
    
    return s.repo.Create(task)
}

func (s *Service) GetTask(id, userID string) (*Task, error) {
    task, err := s.repo.GetByID(id)
    if err != nil {
        return nil, err
    }
    
    if task.UserID != userID {
        return nil, ErrTaskNotFound
    }
    
    return task, nil
}

func (s *Service) ListByUser(userID string) ([]*Task, error) {
    return s.repo.ListByUserID(userID)
}

func (s *Service) UpdateTask(id, userID string, req *UpdateTaskRequest) (*Task, error) {
    task, err := s.GetTask(id, userID)
    if err != nil {
        return nil, err
    }
    
    if req.Title != nil {
        task.Title = *req.Title
    }
    if req.Description != nil {
        task.Description = *req.Description
    }
    if req.Status != nil {
        task.Status = *req.Status
    }
    
    task.UpdatedAt = time.Now()
    
    if err := s.repo.Update(task); err != nil {
        return nil, err
    }
    
    return task, nil
}

func (s *Service) DeleteTask(id, userID string) error {
    task, err := s.GetTask(id, userID)
    if err != nil {
        return err
    }
    
    return s.repo.Delete(task.ID)
}
```

### Repository Layer

```go
// internal/task/repository.go
package task

import (
    "database/sql"
    
    "go.uber.org/fx"
)

type Repository struct {
    db *sql.DB
}

type RepositoryParams struct {
    fx.In
    DB *sql.DB
}

func NewRepository(params RepositoryParams) *Repository {
    return &Repository{
        db: params.DB,
    }
}

func (r *Repository) Create(task *Task) error {
    query := `
        INSERT INTO tasks (id, title, description, status, user_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
    
    _, err := r.db.Exec(
        query,
        task.ID,
        task.Title,
        task.Description,
        task.Status,
        task.UserID,
        task.CreatedAt,
        task.UpdatedAt,
    )
    
    return err
}

func (r *Repository) GetByID(id string) (*Task, error) {
    query := `
        SELECT id, title, description, status, user_id, created_at, updated_at
        FROM tasks
        WHERE id = $1
    `
    
    task := &Task{}
    err := r.db.QueryRow(query, id).Scan(
        &task.ID,
        &task.Title,
        &task.Description,
        &task.Status,
        &task.UserID,
        &task.CreatedAt,
        &task.UpdatedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, ErrTaskNotFound
    }
    
    return task, err
}

func (r *Repository) ListByUserID(userID string) ([]*Task, error) {
    query := `
        SELECT id, title, description, status, user_id, created_at, updated_at
        FROM tasks
        WHERE user_id = $1
        ORDER BY created_at DESC
    `
    
    rows, err := r.db.Query(query, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var tasks []*Task
    for rows.Next() {
        task := &Task{}
        err := rows.Scan(
            &task.ID,
            &task.Title,
            &task.Description,
            &task.Status,
            &task.UserID,
            &task.CreatedAt,
            &task.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        tasks = append(tasks, task)
    }
    
    return tasks, rows.Err()
}

func (r *Repository) Update(task *Task) error {
    query := `
        UPDATE tasks
        SET title = $2, description = $3, status = $4, updated_at = $5
        WHERE id = $1
    `
    
    _, err := r.db.Exec(
        query,
        task.ID,
        task.Title,
        task.Description,
        task.Status,
        task.UpdatedAt,
    )
    
    return err
}

func (r *Repository) Delete(id string) error {
    query := `DELETE FROM tasks WHERE id = $1`
    _, err := r.db.Exec(query, id)
    return err
}
```

## Database Integration

### Migration

```sql
-- migrations/001_create_tables.sql
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tasks_user_id ON tasks(user_id);
CREATE INDEX idx_tasks_status ON tasks(status);
```

### Database Module

```go
// internal/database/module.go
package database

import (
    "context"
    "database/sql"
    
    "go.oease.dev/goe/v2/contract"
    "go.uber.org/fx"
)

type Module struct {
    db     *sql.DB
    logger contract.Logger
}

type ModuleParams struct {
    fx.In
    DB     *sql.DB
    Logger contract.Logger
}

func NewModule(params ModuleParams) contract.Module {
    return &Module{
        db:     params.DB,
        logger: params.Logger,
    }
}

func (m *Module) Name() string {
    return "database"
}

func (m *Module) OnStart(ctx context.Context) error {
    m.logger.Info("Running database migrations")
    
    // Run migrations here
    // You can use golang-migrate or similar tools
    
    return nil
}

func (m *Module) OnStop(ctx context.Context) error {
    m.logger.Info("Closing database connection")
    return m.db.Close()
}
```

## Authentication

### JWT Service

```go
// internal/auth/jwt.go
package auth

import (
    "errors"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
    "go.oease.dev/goe/v2/contract"
    "go.uber.org/fx"
)

var (
    ErrInvalidToken = errors.New("invalid token")
    ErrExpiredToken = errors.New("expired token")
)

type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

type JWTService struct {
    secret string
    expiry time.Duration
    logger contract.Logger
}

type JWTServiceParams struct {
    fx.In
    Config contract.Config
    Logger contract.Logger
}

func NewJWTService(params JWTServiceParams) *JWTService {
    return &JWTService{
        secret: params.Config.GetString("JWT_SECRET"),
        expiry: params.Config.GetDuration("JWT_EXPIRY"),
        logger: params.Logger,
    }
}

func (s *JWTService) GenerateToken(userID, email string) (string, error) {
    claims := &Claims{
        UserID: userID,
        Email:  email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.secret))
}

func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, ErrInvalidToken
        }
        return []byte(s.secret), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, ErrInvalidToken
    }
    
    if time.Now().After(claims.ExpiresAt.Time) {
        return nil, ErrExpiredToken
    }
    
    return claims, nil
}
```

### Auth Middleware

```go
// internal/auth/middleware.go
package auth

import (
    "strings"
    
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2/core/http"
)

func NewAuthMiddleware(jwtService *JWTService) fiber.Handler {
    return func(c fiber.Ctx) error {
        logger := http.GetLogger(c)
        
        // Get token from Authorization header
        auth := c.Get("Authorization")
        if auth == "" {
            return c.Status(401).JSON(fiber.Map{
                "error": "Missing authorization header",
            })
        }
        
        // Extract token
        parts := strings.Split(auth, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            return c.Status(401).JSON(fiber.Map{
                "error": "Invalid authorization format",
            })
        }
        
        token := parts[1]
        
        // Validate token
        claims, err := jwtService.ValidateToken(token)
        if err != nil {
            logger.Debug("Invalid token", log.NewField("error", err))
            return c.Status(401).JSON(fiber.Map{
                "error": "Invalid or expired token",
            })
        }
        
        // Store user info in context
        c.Locals("userID", claims.UserID)
        c.Locals("email", claims.Email)
        
        return c.Next()
    }
}
```

### Auth Controller

```go
// internal/auth/controller.go
package auth

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2/contract"
    "go.oease.dev/goe/v2/core/http"
    "go.uber.org/fx"
)

type Controller struct {
    service *Service
    logger  contract.Logger
}

type ControllerParams struct {
    fx.In
    Service *Service
    Logger  contract.Logger
}

func NewController(params ControllerParams) *Controller {
    return &Controller{
        service: params.Service,
        logger:  params.Logger,
    }
}

func (c *Controller) RegisterRoutes(app *fiber.App) {
    auth := app.Group("/api/v1/auth")
    
    auth.Post("/register", c.Register)
    auth.Post("/login", c.Login)
}

type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Name     string `json:"name" validate:"required,min=2"`
}

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

func (c *Controller) Register(ctx fiber.Ctx) error {
    logger := http.GetLogger(ctx)
    
    var req RegisterRequest
    if err := ctx.Bind().Body(&req); err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }
    
    if err := validate.Struct(req); err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "Validation failed",
            "details": formatValidationErrors(err),
        })
    }
    
    user, token, err := c.service.Register(req.Email, req.Password, req.Name)
    if err != nil {
        if err == ErrEmailExists {
            return ctx.Status(409).JSON(fiber.Map{
                "error": "Email already registered",
            })
        }
        
        logger.Error("Registration failed", log.NewField("error", err))
        return ctx.Status(500).JSON(fiber.Map{
            "error": "Registration failed",
        })
    }
    
    return ctx.Status(201).JSON(fiber.Map{
        "user": fiber.Map{
            "id":    user.ID,
            "email": user.Email,
            "name":  user.Name,
        },
        "token": token,
    })
}

func (c *Controller) Login(ctx fiber.Ctx) error {
    logger := http.GetLogger(ctx)
    
    var req LoginRequest
    if err := ctx.Bind().Body(&req); err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }
    
    if err := validate.Struct(req); err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "Validation failed",
            "details": formatValidationErrors(err),
        })
    }
    
    user, token, err := c.service.Login(req.Email, req.Password)
    if err != nil {
        if err == ErrInvalidCredentials {
            return ctx.Status(401).JSON(fiber.Map{
                "error": "Invalid email or password",
            })
        }
        
        logger.Error("Login failed", log.NewField("error", err))
        return ctx.Status(500).JSON(fiber.Map{
            "error": "Login failed",
        })
    }
    
    return ctx.JSON(fiber.Map{
        "user": fiber.Map{
            "id":    user.ID,
            "email": user.Email,
            "name":  user.Name,
        },
        "token": token,
    })
}
```

## Testing

### Unit Tests

```go
// internal/task/service_test.go
package task

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "go.oease.dev/goe/v2/core/log"
)

// Mock repository
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Create(task *Task) error {
    args := m.Called(task)
    return args.Error(0)
}

func (m *MockRepository) GetByID(id string) (*Task, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*Task), args.Error(1)
}

func TestService_CreateTask(t *testing.T) {
    // Setup
    mockRepo := new(MockRepository)
    logger := log.NewTestLogger()
    
    service := &Service{
        logger: logger,
        repo:   mockRepo,
    }
    
    task := &Task{
        Title:       "Test Task",
        Description: "Test Description",
        UserID:      "user123",
    }
    
    // Expectations
    mockRepo.On("Create", mock.AnythingOfType("*task.Task")).Return(nil)
    
    // Test
    err := service.CreateTask(task)
    
    // Assertions
    assert.NoError(t, err)
    assert.NotEmpty(t, task.ID)
    assert.NotZero(t, task.CreatedAt)
    assert.NotZero(t, task.UpdatedAt)
    assert.Equal(t, "pending", task.Status)
    
    mockRepo.AssertExpectations(t)
}

func TestService_GetTask_NotFound(t *testing.T) {
    // Setup
    mockRepo := new(MockRepository)
    logger := log.NewTestLogger()
    
    service := &Service{
        logger: logger,
        repo:   mockRepo,
    }
    
    // Expectations
    mockRepo.On("GetByID", "task123").Return(nil, ErrTaskNotFound)
    
    // Test
    task, err := service.GetTask("task123", "user123")
    
    // Assertions
    assert.Error(t, err)
    assert.Equal(t, ErrTaskNotFound, err)
    assert.Nil(t, task)
    
    mockRepo.AssertExpectations(t)
}

func TestService_GetTask_WrongUser(t *testing.T) {
    // Setup
    mockRepo := new(MockRepository)
    logger := log.NewTestLogger()
    
    service := &Service{
        logger: logger,
        repo:   mockRepo,
    }
    
    existingTask := &Task{
        ID:     "task123",
        UserID: "user456", // Different user
    }
    
    // Expectations
    mockRepo.On("GetByID", "task123").Return(existingTask, nil)
    
    // Test
    task, err := service.GetTask("task123", "user123")
    
    // Assertions
    assert.Error(t, err)
    assert.Equal(t, ErrTaskNotFound, err)
    assert.Nil(t, task)
    
    mockRepo.AssertExpectations(t)
}
```

### Integration Tests

```go
// internal/task/controller_test.go
package task

import (
    "bytes"
    "encoding/json"
    "net/http/httptest"
    "testing"
    
    "github.com/gofiber/fiber/v3"
    "github.com/stretchr/testify/assert"
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
)

func TestController_CreateTask_Integration(t *testing.T) {
    // Setup Goe application for testing
    _ = goe.New(goe.Options{
        Name:        "Test App",
        Environment: "test",
        WithHTTP:    true,
        Providers: []any{
            NewMockDatabase,
            NewRepository,
            NewService,
            NewController,
        },
        Invokers: []any{
            func(http contract.HTTPKernel, controller *Controller) {
                app := http.App()
                
                // Mock auth middleware
                mockAuth := func(c fiber.Ctx) error {
                    c.Locals("userID", "user123")
                    return c.Next()
                }
                
                controller.RegisterRoutes(app, mockAuth)
            },
        },
    })
    
    app := goe.HTTP().App()
    
    // Prepare request
    reqBody := CreateTaskRequest{
        Title:       "Integration Test Task",
        Description: "Test Description",
    }
    
    body, _ := json.Marshal(reqBody)
    req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    // Test
    resp, err := app.Test(req)
    
    // Assertions
    assert.NoError(t, err)
    assert.Equal(t, 201, resp.StatusCode)
    
    var task Task
    json.NewDecoder(resp.Body).Decode(&task)
    
    assert.NotEmpty(t, task.ID)
    assert.Equal(t, reqBody.Title, task.Title)
    assert.Equal(t, reqBody.Description, task.Description)
    assert.Equal(t, "user123", task.UserID)
    assert.Equal(t, "pending", task.Status)
}
```

## Deployment

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o task-api cmd/api/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/task-api .
COPY --from=builder /app/.env.production .env

EXPOSE 8080

CMD ["./task-api"]
```

### Docker Compose

```yaml
version: '3.8'

services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - GOE_ENV=production
      - DB_HOST=postgres
      - DB_PASSWORD=${DB_PASSWORD}
    depends_on:
      - postgres
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=taskapi
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    restart: unless-stopped

volumes:
  postgres_data:
```

### Production Configuration

```go
// cmd/api/main.go - Complete example
package main

import (
    "database/sql"
    
    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/fiber/v3/middleware/cors"
    "github.com/gofiber/fiber/v3/middleware/helmet"
    "github.com/gofiber/fiber/v3/middleware/limiter"
    
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
    "go.uber.org/fx"
    
    "github.com/yourusername/task-api/internal/auth"
    "github.com/yourusername/task-api/internal/database"
    "github.com/yourusername/task-api/internal/task"
)

func main() {
    _ = goe.New(goe.Options{
        Name:        "Task API",
        Version:     "1.0.0",
        Environment: goe.GetEnvironment(),
        WithHTTP:    true,
        Modules: []contract.Module{
            database.NewModule,
        },
        Providers: []any{
            // Database
            database.NewDatabase,
            
            // Auth
            auth.NewJWTService,
            auth.NewRepository,
            auth.NewService,
            auth.NewController,
            
            // Task
            task.NewRepository,
            task.NewService,
            task.NewController,
        },
        Invokers: []any{
            setupMiddleware,
            setupRoutes,
        },
    })
    
    goe.Log().Info("Starting Task API",
        log.NewField("version", goe.App().Version()),
        log.NewField("environment", goe.GetEnvironment()),
    )
    
    goe.Run()
}

func setupMiddleware(http contract.HTTPKernel, config contract.Config) {
    app := http.App()
    
    // Security headers
    app.Use(helmet.New())
    
    // CORS
    app.Use(cors.New(cors.Config{
        AllowOrigins: config.GetString("CORS_ORIGINS"),
        AllowHeaders: "Origin, Content-Type, Accept, Authorization",
        AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
    }))
    
    // Rate limiting
    app.Use(limiter.New(limiter.Config{
        Max:        config.GetInt("RATE_LIMIT_MAX"),
        Expiration: config.GetDuration("RATE_LIMIT_WINDOW"),
    }))
}

func setupRoutes(params struct {
    fx.In
    HTTP           contract.HTTPKernel
    Logger         contract.Logger
    JWTService     *auth.JWTService
    AuthController *auth.Controller
    TaskController *task.Controller
}) {
    app := params.HTTP.App()
    
    // Health check
    app.Get("/health", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status":  "healthy",
            "version": goe.App().Version(),
        })
    })
    
    // API routes
    api := app.Group("/api/v1")
    
    // Auth routes (public)
    params.AuthController.RegisterRoutes(api)
    
    // Protected routes
    authMiddleware := auth.NewAuthMiddleware(params.JWTService)
    params.TaskController.RegisterRoutes(api, authMiddleware)
    
    // 404 handler
    app.Use(func(c fiber.Ctx) error {
        return c.Status(404).JSON(fiber.Map{
            "error": "Not found",
        })
    })
    
    params.Logger.Info("All routes configured")
}
```

### Production Checklist

1. **Environment Variables**
   - Use strong JWT secret
   - Enable SSL for database
   - Set appropriate timeouts
   - Configure CORS properly

2. **Security**
   - Enable HTTPS
   - Use security headers (helmet)
   - Implement rate limiting
   - Validate all inputs
   - Sanitize outputs

3. **Monitoring**
   - Use JSON logging
   - Set up health checks
   - Monitor error rates
   - Track response times

4. **Performance**
   - Use connection pooling
   - Implement caching
   - Optimize database queries
   - Use pagination

5. **Deployment**
   - Use CI/CD pipeline
   - Run database migrations
   - Graceful shutdown
   - Rolling updates

## Conclusion

You've now built a complete REST API with:

- ✅ Clean architecture
- ✅ Dependency injection
- ✅ Database integration
- ✅ Authentication
- ✅ Input validation
- ✅ Error handling
- ✅ Testing
- ✅ Production-ready setup

The Goe framework provides a solid foundation for building scalable Go applications with excellent developer experience and maintainability.