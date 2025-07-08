# Authentication Example

This example demonstrates how to implement JWT-based authentication in a GOE application.

## JWT Service

### JWT Configuration

```go
package jwt

import (
    "time"
    "go.oease.dev/goe/v2/contract"
)

type JWTConfig struct {
    Secret     string
    Expiration time.Duration
    Issuer     string
}

func NewJWTConfig(config contract.Config) *JWTConfig {
    return &JWTConfig{
        Secret:     config.GetString("JWT_SECRET"),
        Expiration: config.GetDurationWithDefault("JWT_EXPIRATION", 24*time.Hour),
        Issuer:     config.GetStringWithDefault("JWT_ISSUER", "goe-app"),
    }
}
```

### JWT Service Implementation

```go
package jwt

import (
    "errors"
    "time"
    "github.com/golang-jwt/jwt/v5"
    "go.oease.dev/goe/v2/contract"
)

type JWTService interface {
    GenerateToken(userID uint, email string) (string, error)
    ValidateToken(tokenString string) (*Claims, error)
}

type Claims struct {
    UserID uint   `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

type jwtService struct {
    config *JWTConfig
    logger contract.Logger
}

func NewJWTService(config *JWTConfig, logger contract.Logger) JWTService {
    return &jwtService{
        config: config,
        logger: logger,
    }
}

func (s *jwtService) GenerateToken(userID uint, email string) (string, error) {
    claims := &Claims{
        UserID: userID,
        Email:  email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.Expiration)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    s.config.Issuer,
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(s.config.Secret))
    
    if err != nil {
        s.logger.Error("Failed to generate JWT token", "error", err)
        return "", err
    }
    
    s.logger.Info("JWT token generated", "user_id", userID, "email", email)
    return tokenString, nil
}

func (s *jwtService) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("invalid signing method")
        }
        return []byte(s.config.Secret), nil
    })
    
    if err != nil {
        s.logger.Error("Failed to parse JWT token", "error", err)
        return nil, err
    }
    
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, errors.New("invalid token")
}
```

## Authentication Service

### User Model

```go
package model

import (
    "time"
    "gorm.io/gorm"
    "golang.org/x/crypto/bcrypt"
)

type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Email     string         `gorm:"size:100;uniqueIndex;not null" json:"email"`
    Password  string         `gorm:"size:255;not null" json:"-"`
    Name      string         `gorm:"size:100;not null" json:"name"`
    Role      string         `gorm:"size:50;default:user" json:"role"`
    Active    bool           `gorm:"default:true" json:"active"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *User) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
    return err == nil
}

func (u *User) SetPassword(password string) error {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.Password = string(hashedPassword)
    return nil
}
```

### Authentication Service

```go
package service

import (
    "errors"
    "go.oease.dev/goe/v2/contract"
    "your-app/jwt"
    "your-app/model"
    "your-app/repository"
)

var (
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrUserNotFound      = errors.New("user not found")
    ErrUserInactive      = errors.New("user is inactive")
)

type AuthService struct {
    userRepo   *repository.UserRepository
    jwtService jwt.JWTService
    logger     contract.Logger
}

func NewAuthService(userRepo *repository.UserRepository, jwtService jwt.JWTService, logger contract.Logger) *AuthService {
    return &AuthService{
        userRepo:   userRepo,
        jwtService: jwtService,
        logger:     logger,
    }
}

func (s *AuthService) Register(name, email, password string) (*model.User, string, error) {
    s.logger.Info("User registration attempt", "email", email)
    
    // Check if user already exists
    if existingUser, err := s.userRepo.GetByEmail(email); err == nil && existingUser != nil {
        return nil, "", errors.New("user already exists")
    }
    
    // Create new user
    user := &model.User{
        Name:   name,
        Email:  email,
        Role:   "user",
        Active: true,
    }
    
    if err := user.SetPassword(password); err != nil {
        s.logger.Error("Failed to hash password", "error", err)
        return nil, "", err
    }
    
    if err := s.userRepo.Create(user); err != nil {
        s.logger.Error("Failed to create user", "error", err)
        return nil, "", err
    }
    
    // Generate JWT token
    token, err := s.jwtService.GenerateToken(user.ID, user.Email)
    if err != nil {
        s.logger.Error("Failed to generate token", "error", err)
        return nil, "", err
    }
    
    s.logger.Info("User registered successfully", "user_id", user.ID, "email", email)
    return user, token, nil
}

func (s *AuthService) Login(email, password string) (*model.User, string, error) {
    s.logger.Info("User login attempt", "email", email)
    
    // Find user by email
    user, err := s.userRepo.GetByEmail(email)
    if err != nil {
        s.logger.Error("User not found", "email", email, "error", err)
        return nil, "", ErrUserNotFound
    }
    
    // Check if user is active
    if !user.Active {
        s.logger.Warn("Login attempt for inactive user", "email", email)
        return nil, "", ErrUserInactive
    }
    
    // Check password
    if !user.CheckPassword(password) {
        s.logger.Warn("Invalid password", "email", email)
        return nil, "", ErrInvalidCredentials
    }
    
    // Generate JWT token
    token, err := s.jwtService.GenerateToken(user.ID, user.Email)
    if err != nil {
        s.logger.Error("Failed to generate token", "error", err)
        return nil, "", err
    }
    
    s.logger.Info("User logged in successfully", "user_id", user.ID, "email", email)
    return user, token, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*model.User, error) {
    claims, err := s.jwtService.ValidateToken(tokenString)
    if err != nil {
        return nil, err
    }
    
    user, err := s.userRepo.GetByID(claims.UserID)
    if err != nil {
        return nil, err
    }
    
    if !user.Active {
        return nil, ErrUserInactive
    }
    
    return user, nil
}
```

## Authentication Middleware

```go
package middleware

import (
    "strings"
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2/contract"
    "your-app/service"
)

type AuthMiddleware struct {
    authService *service.AuthService
    logger      contract.Logger
}

func NewAuthMiddleware(authService *service.AuthService, logger contract.Logger) *AuthMiddleware {
    return &AuthMiddleware{
        authService: authService,
        logger:      logger,
    }
}

func (m *AuthMiddleware) RequireAuth(c fiber.Ctx) error {
    // Get token from header
    authHeader := c.Get("Authorization")
    if authHeader == "" {
        return c.Status(401).JSON(fiber.Map{
            "error": "Authorization header required",
        })
    }
    
    // Extract token from "Bearer <token>"
    tokenParts := strings.Split(authHeader, " ")
    if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
        return c.Status(401).JSON(fiber.Map{
            "error": "Invalid authorization header format",
        })
    }
    
    token := tokenParts[1]
    
    // Validate token
    user, err := m.authService.ValidateToken(token)
    if err != nil {
        m.logger.Warn("Invalid token", "error", err)
        return c.Status(401).JSON(fiber.Map{
            "error": "Invalid token",
        })
    }
    
    // Store user in context
    c.Locals("user", user)
    c.Locals("user_id", user.ID)
    
    return c.Next()
}

func (m *AuthMiddleware) RequireRole(role string) fiber.Handler {
    return func(c fiber.Ctx) error {
        user, ok := c.Locals("user").(*model.User)
        if !ok {
            return c.Status(401).JSON(fiber.Map{
                "error": "Authentication required",
            })
        }
        
        if user.Role != role {
            return c.Status(403).JSON(fiber.Map{
                "error": "Insufficient permissions",
            })
        }
        
        return c.Next()
    }
}

func (m *AuthMiddleware) OptionalAuth(c fiber.Ctx) error {
    authHeader := c.Get("Authorization")
    if authHeader == "" {
        return c.Next()
    }
    
    tokenParts := strings.Split(authHeader, " ")
    if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
        return c.Next()
    }
    
    token := tokenParts[1]
    
    user, err := m.authService.ValidateToken(token)
    if err != nil {
        return c.Next()
    }
    
    c.Locals("user", user)
    c.Locals("user_id", user.ID)
    
    return c.Next()
}
```

## Authentication Handlers

```go
package handler

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2/contract"
    "your-app/model"
    "your-app/service"
)

type AuthHandler struct {
    authService *service.AuthService
    logger      contract.Logger
}

func NewAuthHandler(authService *service.AuthService, logger contract.Logger) *AuthHandler {
    return &AuthHandler{
        authService: authService,
        logger:      logger,
    }
}

type RegisterRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
    var req RegisterRequest
    
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
    
    user, token, err := h.authService.Register(req.Name, req.Email, req.Password)
    if err != nil {
        h.logger.Error("Registration failed", "error", err)
        return c.Status(400).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    
    return c.Status(201).JSON(fiber.Map{
        "user":  user,
        "token": token,
    })
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
    var req LoginRequest
    
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
    
    user, token, err := h.authService.Login(req.Email, req.Password)
    if err != nil {
        h.logger.Error("Login failed", "error", err)
        return c.Status(401).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    
    return c.JSON(fiber.Map{
        "user":  user,
        "token": token,
    })
}

func (h *AuthHandler) Profile(c fiber.Ctx) error {
    user, ok := c.Locals("user").(*model.User)
    if !ok {
        return c.Status(401).JSON(fiber.Map{
            "error": "Authentication required",
        })
    }
    
    return c.JSON(user)
}

func (h *AuthHandler) RefreshToken(c fiber.Ctx) error {
    user, ok := c.Locals("user").(*model.User)
    if !ok {
        return c.Status(401).JSON(fiber.Map{
            "error": "Authentication required",
        })
    }
    
    token, err := h.authService.jwtService.GenerateToken(user.ID, user.Email)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to refresh token",
        })
    }
    
    return c.JSON(fiber.Map{
        "token": token,
    })
}
```

## Main Application

```go
package main

import (
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
    "your-app/handler"
    "your-app/jwt"
    "your-app/middleware"
    "your-app/model"
    "your-app/repository"
    "your-app/service"
)

func main() {
    goe.New(goe.Options{
        WithHTTP: true,
        WithDB:   true,
        Providers: []any{
            jwt.NewJWTConfig,
            jwt.NewJWTService,
            repository.NewUserRepository,
            service.NewAuthService,
            handler.NewAuthHandler,
            middleware.NewAuthMiddleware,
        },
        Invokers: []any{
            AutoMigrate,
            RegisterRoutes,
        },
    })
    
    goe.Run()
}

func AutoMigrate(db contract.DB) {
    db.Instance().AutoMigrate(&model.User{})
}

func RegisterRoutes(
    httpKernel contract.HTTPKernel,
    authHandler *handler.AuthHandler,
    authMiddleware *middleware.AuthMiddleware,
) {
    app := httpKernel.App()
    
    // Public routes
    auth := app.Group("/auth")
    auth.Post("/register", authHandler.Register)
    auth.Post("/login", authHandler.Login)
    
    // Protected routes
    api := app.Group("/api")
    api.Use(authMiddleware.RequireAuth)
    api.Get("/profile", authHandler.Profile)
    api.Post("/refresh", authHandler.RefreshToken)
    
    // Admin routes
    admin := api.Group("/admin")
    admin.Use(authMiddleware.RequireRole("admin"))
    admin.Get("/users", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{"message": "Admin only endpoint"})
    })
}
```

## Configuration

```bash
# .env
JWT_SECRET=your-super-secret-jwt-key-here
JWT_EXPIRATION=24h
JWT_ISSUER=your-app-name

HTTP_PORT=8080
DB_DRIVER=sqlite
DB_PATH=./database.db
LOG_LEVEL=info
```

## Testing the Authentication

### Register a new user

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Access protected route

```bash
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE"
```

## Testing Authentication

```go
package service

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "your-app/model"
)

type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) GetByEmail(email string) (*model.User, error) {
    args := m.Called(email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Create(user *model.User) error {
    args := m.Called(user)
    return args.Error(0)
}

func TestAuthService_Login(t *testing.T) {
    // Setup
    mockRepo := new(MockUserRepository)
    mockJWT := new(MockJWTService)
    mockLogger := new(MockLogger)
    
    service := NewAuthService(mockRepo, mockJWT, mockLogger)
    
    // Create test user
    user := &model.User{
        ID:     1,
        Email:  "test@example.com",
        Active: true,
    }
    user.SetPassword("password123")
    
    mockRepo.On("GetByEmail", "test@example.com").Return(user, nil)
    mockJWT.On("GenerateToken", uint(1), "test@example.com").Return("token123", nil)
    mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything)
    
    // Execute
    resultUser, token, err := service.Login("test@example.com", "password123")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, user, resultUser)
    assert.Equal(t, "token123", token)
    
    mockRepo.AssertExpectations(t)
    mockJWT.AssertExpectations(t)
}
```

This example demonstrates:
- JWT token generation and validation
- User registration and login
- Password hashing and verification
- Authentication middleware
- Role-based access control
- Protected routes
- Proper error handling
- Testing authentication logic