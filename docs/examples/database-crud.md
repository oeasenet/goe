# Database CRUD Example

This example demonstrates how to implement Create, Read, Update, Delete operations with GOE's database module.

## Setup

```go
package main

import (
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
)

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
            AutoMigrate,
            RegisterRoutes,
        },
    })
    
    goe.Run()
}
```

## User Model

```go
package model

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Name      string         `gorm:"size:100;not null" json:"name"`
    Email     string         `gorm:"size:100;uniqueIndex;not null" json:"email"`
    Age       int            `gorm:"not null" json:"age"`
    Active    bool           `gorm:"default:true" json:"active"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

## Repository Layer

```go
package repository

import (
    "go.oease.dev/goe/v2/contract"
    "your-app/model"
)

type UserRepository struct {
    db contract.DB
}

func NewUserRepository(db contract.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
    return r.db.Instance().Create(user).Error
}

func (r *UserRepository) GetByID(id uint) (*model.User, error) {
    var user model.User
    err := r.db.Instance().First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) GetAll(limit, offset int) ([]model.User, error) {
    var users []model.User
    err := r.db.Instance().
        Limit(limit).
        Offset(offset).
        Order("created_at DESC").
        Find(&users).Error
    return users, err
}

func (r *UserRepository) Update(user *model.User) error {
    return r.db.Instance().Save(user).Error
}

func (r *UserRepository) Delete(id uint) error {
    return r.db.Instance().Delete(&model.User{}, id).Error
}

func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
    var user model.User
    err := r.db.Instance().Where("email = ?", email).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}
```

## Service Layer

```go
package service

import (
    "errors"
    "go.oease.dev/goe/v2/contract"
    "your-app/model"
    "your-app/repository"
)

type UserService struct {
    repo   *repository.UserRepository
    logger contract.Logger
}

func NewUserService(repo *repository.UserRepository, logger contract.Logger) *UserService {
    return &UserService{
        repo:   repo,
        logger: logger,
    }
}

func (s *UserService) CreateUser(name, email string, age int) (*model.User, error) {
    s.logger.Info("Creating user", "name", name, "email", email)
    
    // Check if user already exists
    existingUser, err := s.repo.GetByEmail(email)
    if err == nil && existingUser != nil {
        return nil, errors.New("user with this email already exists")
    }
    
    user := &model.User{
        Name:   name,
        Email:  email,
        Age:    age,
        Active: true,
    }
    
    if err := s.repo.Create(user); err != nil {
        s.logger.Error("Failed to create user", "error", err)
        return nil, err
    }
    
    s.logger.Info("User created successfully", "user_id", user.ID)
    return user, nil
}

func (s *UserService) GetUser(id uint) (*model.User, error) {
    s.logger.Info("Getting user", "user_id", id)
    
    user, err := s.repo.GetByID(id)
    if err != nil {
        s.logger.Error("Failed to get user", "user_id", id, "error", err)
        return nil, err
    }
    
    return user, nil
}

func (s *UserService) GetUsers(page, limit int) ([]model.User, error) {
    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 10
    }
    
    offset := (page - 1) * limit
    
    s.logger.Info("Getting users", "page", page, "limit", limit)
    
    users, err := s.repo.GetAll(limit, offset)
    if err != nil {
        s.logger.Error("Failed to get users", "error", err)
        return nil, err
    }
    
    return users, nil
}

func (s *UserService) UpdateUser(id uint, name, email string, age int) (*model.User, error) {
    s.logger.Info("Updating user", "user_id", id)
    
    user, err := s.repo.GetByID(id)
    if err != nil {
        s.logger.Error("User not found", "user_id", id, "error", err)
        return nil, err
    }
    
    user.Name = name
    user.Email = email
    user.Age = age
    
    if err := s.repo.Update(user); err != nil {
        s.logger.Error("Failed to update user", "user_id", id, "error", err)
        return nil, err
    }
    
    s.logger.Info("User updated successfully", "user_id", id)
    return user, nil
}

func (s *UserService) DeleteUser(id uint) error {
    s.logger.Info("Deleting user", "user_id", id)
    
    // Check if user exists
    _, err := s.repo.GetByID(id)
    if err != nil {
        s.logger.Error("User not found", "user_id", id, "error", err)
        return err
    }
    
    if err := s.repo.Delete(id); err != nil {
        s.logger.Error("Failed to delete user", "user_id", id, "error", err)
        return err
    }
    
    s.logger.Info("User deleted successfully", "user_id", id)
    return nil
}
```

## Handler Layer

```go
package handler

import (
    "strconv"
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2/contract"
    "your-app/service"
)

type UserHandler struct {
    service *service.UserService
    logger  contract.Logger
}

func NewUserHandler(service *service.UserService, logger contract.Logger) *UserHandler {
    return &UserHandler{
        service: service,
        logger:  logger,
    }
}

type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2,max=100"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"required,min=1,max=120"`
}

type UpdateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2,max=100"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"required,min=1,max=120"`
}

func (h *UserHandler) CreateUser(c fiber.Ctx) error {
    var req CreateUserRequest
    
    if err := c.BodyParser(&req); err != nil {
        h.logger.Error("Invalid request body", "error", err)
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }
    
    // Validate request
    if err := validate.Struct(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Validation failed",
            "details": err.Error(),
        })
    }
    
    user, err := h.service.CreateUser(req.Name, req.Email, req.Age)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    
    return c.Status(201).JSON(user)
}

func (h *UserHandler) GetUser(c fiber.Ctx) error {
    id, err := strconv.ParseUint(c.Params("id"), 10, 32)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid user ID",
        })
    }
    
    user, err := h.service.GetUser(uint(id))
    if err != nil {
        return c.Status(404).JSON(fiber.Map{
            "error": "User not found",
        })
    }
    
    return c.JSON(user)
}

func (h *UserHandler) GetUsers(c fiber.Ctx) error {
    page, _ := strconv.Atoi(c.Query("page", "1"))
    limit, _ := strconv.Atoi(c.Query("limit", "10"))
    
    users, err := h.service.GetUsers(page, limit)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to get users",
        })
    }
    
    return c.JSON(fiber.Map{
        "users": users,
        "page":  page,
        "limit": limit,
    })
}

func (h *UserHandler) UpdateUser(c fiber.Ctx) error {
    id, err := strconv.ParseUint(c.Params("id"), 10, 32)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid user ID",
        })
    }
    
    var req UpdateUserRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }
    
    if err := validate.Struct(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Validation failed",
            "details": err.Error(),
        })
    }
    
    user, err := h.service.UpdateUser(uint(id), req.Name, req.Email, req.Age)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    
    return c.JSON(user)
}

func (h *UserHandler) DeleteUser(c fiber.Ctx) error {
    id, err := strconv.ParseUint(c.Params("id"), 10, 32)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid user ID",
        })
    }
    
    if err := h.service.DeleteUser(uint(id)); err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    
    return c.Status(204).Send(nil)
}
```

## Routes Registration

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2/contract"
    "your-app/handler"
    "your-app/model"
)

func AutoMigrate(db contract.DB) {
    db.Instance().AutoMigrate(&model.User{})
}

func RegisterRoutes(httpKernel contract.HTTPKernel, userHandler *handler.UserHandler) {
    app := httpKernel.App()
    
    // API routes
    api := app.Group("/api")
    
    // User routes
    users := api.Group("/users")
    users.Post("/", userHandler.CreateUser)
    users.Get("/", userHandler.GetUsers)
    users.Get("/:id", userHandler.GetUser)
    users.Put("/:id", userHandler.UpdateUser)
    users.Delete("/:id", userHandler.DeleteUser)
    
    // Health check
    app.Get("/health", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status": "healthy",
            "service": "user-api",
        })
    })
}
```

## Testing the API

```bash
# Create a user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","age":30}'

# Get all users
curl http://localhost:8080/api/users

# Get user by ID
curl http://localhost:8080/api/users/1

# Update user
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"John Smith","email":"john.smith@example.com","age":31}'

# Delete user
curl -X DELETE http://localhost:8080/api/users/1
```

## Complete Example

Run the complete example:

```bash
# Create project
mkdir goe-crud-example
cd goe-crud-example
go mod init goe-crud-example

# Install dependencies
go get go.oease.dev/goe/v2
go get github.com/go-playground/validator/v10

# Create .env file
echo "DB_DRIVER=sqlite
DB_PATH=./database.db
HTTP_PORT=8080
LOG_LEVEL=info" > .env

# Run the application
go run main.go
```

This example demonstrates:
- Complete CRUD operations
- Repository pattern for data access
- Service layer for business logic
- HTTP handlers for API endpoints
- Proper error handling and logging
- Input validation
- Database migration