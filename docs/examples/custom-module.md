# Custom Module Example

This example shows how to create a custom GOE module for email functionality.

## Email Module

### Email Service Interface

```go
package email

import "context"

type EmailService interface {
    SendEmail(to, subject, body string) error
    SendTemplateEmail(to, template string, data interface{}) error
}

type EmailConfig struct {
    SMTPHost     string
    SMTPPort     int
    SMTPUsername string
    SMTPPassword string
    FromEmail    string
    FromName     string
}
```

### Email Service Implementation

```go
package email

import (
    "fmt"
    "net/smtp"
    "go.oease.dev/goe/v2/contract"
)

type emailService struct {
    config *EmailConfig
    logger contract.Logger
}

func NewEmailService(config *EmailConfig, logger contract.Logger) EmailService {
    return &emailService{
        config: config,
        logger: logger,
    }
}

func (s *emailService) SendEmail(to, subject, body string) error {
    s.logger.Info("Sending email", "to", to, "subject", subject)
    
    // Set up authentication
    auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
    
    // Compose message
    message := fmt.Sprintf("From: %s <%s>\r\n", s.config.FromName, s.config.FromEmail)
    message += fmt.Sprintf("To: %s\r\n", to)
    message += fmt.Sprintf("Subject: %s\r\n", subject)
    message += "\r\n"
    message += body
    
    // Send email
    addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
    err := smtp.SendMail(addr, auth, s.config.FromEmail, []string{to}, []byte(message))
    
    if err != nil {
        s.logger.Error("Failed to send email", "to", to, "error", err)
        return err
    }
    
    s.logger.Info("Email sent successfully", "to", to)
    return nil
}

func (s *emailService) SendTemplateEmail(to, template string, data interface{}) error {
    // Template rendering logic would go here
    body := fmt.Sprintf("Template: %s, Data: %v", template, data)
    return s.SendEmail(to, "Template Email", body)
}
```

### Email Module

```go
package email

import (
    "context"
    "go.oease.dev/goe/v2/contract"
)

type Module struct {
    service EmailService
    config  *EmailConfig
    logger  contract.Logger
}

func NewModule(config contract.Config, logger contract.Logger) *Module {
    emailConfig := &EmailConfig{
        SMTPHost:     config.GetString("SMTP_HOST"),
        SMTPPort:     config.GetInt("SMTP_PORT"),
        SMTPUsername: config.GetString("SMTP_USERNAME"),
        SMTPPassword: config.GetString("SMTP_PASSWORD"),
        FromEmail:    config.GetString("FROM_EMAIL"),
        FromName:     config.GetString("FROM_NAME"),
    }
    
    service := NewEmailService(emailConfig, logger)
    
    return &Module{
        service: service,
        config:  emailConfig,
        logger:  logger,
    }
}

func (m *Module) Name() string {
    return "email-module"
}

func (m *Module) OnStart(ctx context.Context) error {
    m.logger.Info("Email module starting")
    
    // Validate configuration
    if m.config.SMTPHost == "" {
        return fmt.Errorf("SMTP_HOST is required")
    }
    
    if m.config.FromEmail == "" {
        return fmt.Errorf("FROM_EMAIL is required")
    }
    
    m.logger.Info("Email module started successfully")
    return nil
}

func (m *Module) OnStop(ctx context.Context) error {
    m.logger.Info("Email module stopping")
    
    // Cleanup logic here if needed
    
    m.logger.Info("Email module stopped")
    return nil
}

func (m *Module) EmailService() EmailService {
    return m.service
}
```

## Using the Custom Module

### Main Application

```go
package main

import (
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
    "your-app/email"
    "your-app/handler"
    "your-app/service"
)

func main() {
    goe.New(goe.Options{
        WithHTTP: true,
        Providers: []any{
            email.NewModule,
            service.NewUserService,
            handler.NewUserHandler,
        },
        Invokers: []any{
            RegisterRoutes,
        },
    })
    
    goe.Run()
}

func RegisterRoutes(httpKernel contract.HTTPKernel, userHandler *handler.UserHandler) {
    app := httpKernel.App()
    
    app.Post("/users", userHandler.CreateUser)
    app.Post("/users/:id/send-email", userHandler.SendEmailToUser)
}
```

### User Service with Email

```go
package service

import (
    "go.oease.dev/goe/v2/contract"
    "your-app/email"
    "your-app/model"
)

type UserService struct {
    repository   UserRepository
    emailService email.EmailService
    logger       contract.Logger
}

func NewUserService(repository UserRepository, emailModule *email.Module, logger contract.Logger) *UserService {
    return &UserService{
        repository:   repository,
        emailService: emailModule.EmailService(),
        logger:       logger,
    }
}

func (s *UserService) CreateUser(name, email string) (*model.User, error) {
    s.logger.Info("Creating user", "name", name, "email", email)
    
    user := &model.User{
        Name:  name,
        Email: email,
    }
    
    if err := s.repository.Create(user); err != nil {
        s.logger.Error("Failed to create user", "error", err)
        return nil, err
    }
    
    // Send welcome email
    go func() {
        err := s.emailService.SendEmail(
            user.Email,
            "Welcome to Our App",
            fmt.Sprintf("Hello %s, welcome to our application!", user.Name),
        )
        if err != nil {
            s.logger.Error("Failed to send welcome email", "user_id", user.ID, "error", err)
        }
    }()
    
    s.logger.Info("User created successfully", "user_id", user.ID)
    return user, nil
}

func (s *UserService) SendEmailToUser(userID uint, subject, body string) error {
    user, err := s.repository.GetByID(userID)
    if err != nil {
        return err
    }
    
    return s.emailService.SendEmail(user.Email, subject, body)
}
```

### Handler with Email

```go
package handler

import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2/contract"
    "your-app/service"
)

type UserHandler struct {
    userService *service.UserService
    logger      contract.Logger
}

func NewUserHandler(userService *service.UserService, logger contract.Logger) *UserHandler {
    return &UserHandler{
        userService: userService,
        logger:      logger,
    }
}

type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

type SendEmailRequest struct {
    Subject string `json:"subject" validate:"required"`
    Body    string `json:"body" validate:"required"`
}

func (h *UserHandler) CreateUser(c fiber.Ctx) error {
    var req CreateUserRequest
    
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
    }
    
    if err := validate.Struct(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Validation failed"})
    }
    
    user, err := h.userService.CreateUser(req.Name, req.Email)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
    }
    
    return c.Status(201).JSON(user)
}

func (h *UserHandler) SendEmailToUser(c fiber.Ctx) error {
    userID, err := c.ParamsInt("id")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
    }
    
    var req SendEmailRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
    }
    
    if err := validate.Struct(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Validation failed"})
    }
    
    if err := h.userService.SendEmailToUser(uint(userID), req.Subject, req.Body); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to send email"})
    }
    
    return c.JSON(fiber.Map{"message": "Email sent successfully"})
}
```

## Background Task Module

### Task Module

```go
package task

import (
    "context"
    "time"
    "go.oease.dev/goe/v2/contract"
)

type TaskService interface {
    StartPeriodicTask(name string, interval time.Duration, task func() error)
    StopTask(name string)
}

type taskService struct {
    tasks  map[string]chan struct{}
    logger contract.Logger
}

func NewTaskService(logger contract.Logger) TaskService {
    return &taskService{
        tasks:  make(map[string]chan struct{}),
        logger: logger,
    }
}

func (s *taskService) StartPeriodicTask(name string, interval time.Duration, task func() error) {
    s.logger.Info("Starting periodic task", "name", name, "interval", interval)
    
    done := make(chan struct{})
    s.tasks[name] = done
    
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                s.logger.Debug("Running periodic task", "name", name)
                if err := task(); err != nil {
                    s.logger.Error("Periodic task failed", "name", name, "error", err)
                }
            case <-done:
                s.logger.Info("Periodic task stopped", "name", name)
                return
            }
        }
    }()
}

func (s *taskService) StopTask(name string) {
    if done, exists := s.tasks[name]; exists {
        close(done)
        delete(s.tasks, name)
    }
}

type Module struct {
    service TaskService
    logger  contract.Logger
}

func NewModule(logger contract.Logger) *Module {
    service := NewTaskService(logger)
    
    return &Module{
        service: service,
        logger:  logger,
    }
}

func (m *Module) Name() string {
    return "task-module"
}

func (m *Module) OnStart(ctx context.Context) error {
    m.logger.Info("Task module started")
    
    // Start cleanup task
    m.service.StartPeriodicTask("cleanup", 1*time.Hour, func() error {
        m.logger.Info("Running cleanup task")
        // Cleanup logic here
        return nil
    })
    
    return nil
}

func (m *Module) OnStop(ctx context.Context) error {
    m.logger.Info("Task module stopping")
    
    // Stop all tasks
    m.service.StopTask("cleanup")
    
    m.logger.Info("Task module stopped")
    return nil
}

func (m *Module) TaskService() TaskService {
    return m.service
}
```

## Configuration

Add configuration for the custom modules:

```bash
# .env
# Email configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
FROM_EMAIL=your-email@gmail.com
FROM_NAME=Your App Name

# Other configuration
HTTP_PORT=8080
LOG_LEVEL=info
```

## Testing the Custom Module

```bash
# Create a user (will trigger welcome email)
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com"}'

# Send email to user
curl -X POST http://localhost:8080/users/1/send-email \
  -H "Content-Type: application/json" \
  -d '{"subject":"Test Email","body":"This is a test email"}'
```

## Testing Custom Modules

```go
package email

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "go.oease.dev/goe/v2/contract"
)

type MockConfig struct {
    mock.Mock
}

func (m *MockConfig) GetString(key string) string {
    args := m.Called(key)
    return args.String(0)
}

func (m *MockConfig) GetInt(key string) int {
    args := m.Called(key)
    return args.Int(0)
}

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

func TestEmailModule_OnStart(t *testing.T) {
    mockConfig := new(MockConfig)
    mockLogger := new(MockLogger)
    
    mockConfig.On("GetString", "SMTP_HOST").Return("smtp.gmail.com")
    mockConfig.On("GetInt", "SMTP_PORT").Return(587)
    mockConfig.On("GetString", "SMTP_USERNAME").Return("test@gmail.com")
    mockConfig.On("GetString", "SMTP_PASSWORD").Return("password")
    mockConfig.On("GetString", "FROM_EMAIL").Return("test@gmail.com")
    mockConfig.On("GetString", "FROM_NAME").Return("Test App")
    
    mockLogger.On("Info", mock.Anything, mock.Anything).Return()
    
    module := NewModule(mockConfig, mockLogger)
    
    err := module.OnStart(context.Background())
    
    assert.NoError(t, err)
    mockConfig.AssertExpectations(t)
    mockLogger.AssertExpectations(t)
}
```

This example demonstrates:
- Creating a custom module with lifecycle management
- Implementing a service interface
- Using configuration for module setup
- Integration with other services
- Background task management
- Proper error handling and logging
- Testing custom modules