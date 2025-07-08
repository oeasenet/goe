# Project Structure

A well-organized project structure is crucial for maintainability, scalability, and collaboration. While GOE doesn't enforce a strict layout, this section outlines a recommended structure that aligns with common Go practices and works well with GOE applications.

## Standard Go Project Layout

This structure is influenced by the [Standard Go Project Layout](https://github.com/golang-standards/project-layout) and other community best practices. You might not need all of these directories for every project, especially smaller ones. Adapt it to your needs.

```
your-project-name/
├── cmd/                    # Application entry points (main packages)
│   ├── myapi/              # Example: your main web API application
│   │   └── main.go
│   └── myworker/           # Example: a background worker application
│       └── main.go
│
├── internal/               # Private application and library code
│   ├── app/                # Core application logic, orchestration
│   │   └── myapi/          # Application-specific logic for 'myapi'
│   │       ├── service/    # Business logic services
│   │       ├── handler/    # HTTP handlers, gRPC handlers, etc.
│   │       └── model/      # Data transfer objects, request/response models
│   ├── domain/             # Core domain models and business logic
│   │   ├── user/           # Example: user domain
│   │   │   ├── user.go     # User entity, core logic
│   │   │   └── store.go    # Interface for user persistence
│   │   └── product/
│   │       ├── product.go
│   │       └── store.go
│   ├── config/             # Configuration loading and struct definitions
│   ├── module/             # Custom GOE modules specific to your application
│   ├── platform/           # Platform-level concerns
│   │   ├── database/       # Database interaction (GORM setup, migrations)
│   │   ├── cache/          # Cache interaction setup
│   │   └── log/            # Logging setup if more customization is needed
│   └── repository/         # Data persistence implementations
│
├── pkg/                    # Public library code, shareable with external applications
│   ├── mypubliclib/        # Example: a library you want to share
│   │   └── lib.go
│
├── configs/                # Configuration files (e.g., .env.example, default values)
│   ├── .env.development
│   ├── .env.production
│   └── .env.example
│
├── web/                    # Web assets (if serving HTML, SPAs)
│   ├── static/             # CSS, JavaScript, images
│   ├── templates/          # HTML templates
│   └── spa/                # Single Page Application build output
│
├── migrations/             # Database migration files
│
├── build/                  # Build artifacts, scripts, Dockerfiles
│   ├── package/            # Packaging scripts, configurations
│   ├── docker/             # Dockerfiles for different services
│   │   ├── myapi.Dockerfile
│   │   └── myworker.Dockerfile
│   └── ci/                 # CI/CD pipeline configurations
│
├── docs/                   # Project documentation
│
├── scripts/                # Helper scripts (build, test, lint, deploy, etc.)
│   ├── build.sh
│   └── test.sh
│
├── test/                   # Additional test files, E2E tests, test data
│   └── e2e/
│
├── .git/                   # Git repository data
├── .gitignore              # Files and directories to ignore by Git
├── go.mod                  # Go module definition file
├── go.sum                  # Go module checksums
└── README.md               # Project root README
```

## Rationale Behind the Structure

### `cmd/`
- Contains the `main` packages for your executables
- Each subdirectory corresponds to a single binary (web server, CLI tool, worker process)
- Code here should be minimal, primarily responsible for initializing and running the application
- **How GOE fits**: Your `main.go` in these directories will call `goe.New()` and `goe.Run()`

### `internal/`
- Most of your application's private code resides here
- Go enforces that code in `internal/` can only be imported by code within the same parent directory

**Sub-directories:**
- **`app/`**: Application-specific logic, services, handlers, and request/response models
- **`domain/`**: Core business logic and entities, independent of how it's exposed or stored
- **`config/`**: Application configuration loading and provides typed access
- **`module/`**: Custom GOE modules (see [Modules](./modules.md) chapter)
- **`platform/`**: Platform-level concerns like database connections, cache clients
- **`repository/`**: Data access logic, database interactions using GORM

**How GOE fits**: Most of your Fx providers, invokers, services, handlers, and repositories will live within `internal/`. GOE's core modules (`contract.Config`, `contract.Logger`, etc.) will be injected into your components here.

### `pkg/`
- Code that's safe to be imported and used by external applications
- If you don't plan to share any code, you might not need this directory
- Be mindful of what you put here, as it becomes part of your public API

### `configs/`
- Store example configuration files (`.env.example`) and default configurations
- Actual `.env` files should be in your project root (and listed in `.gitignore`)
- **How GOE fits**: GOE's config module will load `.env` files from the root or based on `GOE_ENV`

### `web/`
- For projects serving HTML directly or hosting Single Page Applications
- **How GOE fits**: GoFiber (GOE's HTTP engine) can serve static files from `web/static` or render templates from `web/templates`

### `migrations/`
- Essential for managing database schema changes over time
- Use tools like [Goose](https://github.com/pressly/goose), [Atlas](https://atlasgo.io/), or GORM's migration features
- **How GOE fits**: While GOE's DB module can do auto-migration, for production, use dedicated migration files

### Other directories
- **`build/`**: Build scripts, Dockerfiles, and packaging configurations
- **`docs/`**: Project documentation
- **`scripts/`**: Development, testing, deployment scripts
- **`test/`**: End-to-end tests, integration tests, and common test data

## Benefits of This Structure

- **Clear Separation of Concerns**: Different aspects of your application are neatly organized
- **Improved Navigability**: Easier for developers to find code
- **Scalability**: Structure can accommodate new features and components gracefully
- **Testability**: Well-defined boundaries make unit and integration testing easier
- **Maintainability**: Easier to understand, debug, and refactor code

## How GOE Adapts

GOE is designed to be flexible and doesn't impose this structure. However, its features complement this layout well:

- **Dependency Injection (Fx)**: Encourages defining components in their respective packages within `internal/` and wiring them together in your `cmd/.../main.go`
- **Modules**: Custom GOE modules can be placed in `internal/module/` to encapsulate specific functionalities with their own lifecycle hooks
- **Configuration**: GOE's config system loads `.env` files from the project root, aligning with placing `configs/.env.example` for reference
- **Contracts**: GOE's use of interfaces (`contract.Logger`, `contract.DB`, etc.) makes it easy to inject these framework services into your application components

## Example: User Service Implementation

Here's how a typical user service might be structured:

```go
// internal/domain/user/user.go
package user

import "time"

type User struct {
    ID        uint      `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

type Store interface {
    Create(user *User) error
    FindByID(id uint) (*User, error)
    FindByEmail(email string) (*User, error)
    Update(user *User) error
    Delete(id uint) error
}
```

```go
// internal/app/myapi/service/user.go
package service

import (
    "go.oease.dev/goe/v2/contract"
    "yourproject/internal/domain/user"
)

type UserService struct {
    store  user.Store
    logger contract.Logger
}

func NewUserService(store user.Store, logger contract.Logger) *UserService {
    return &UserService{
        store:  store,
        logger: logger,
    }
}

func (s *UserService) CreateUser(name, email string) (*user.User, error) {
    u := &user.User{
        Name:  name,
        Email: email,
    }
    
    if err := s.store.Create(u); err != nil {
        s.logger.Error("Failed to create user", "error", err)
        return nil, err
    }
    
    s.logger.Info("User created", "id", u.ID, "email", u.Email)
    return u, nil
}
```

```go
// internal/repository/user.go
package repository

import (
    "go.oease.dev/goe/v2/contract"
    "yourproject/internal/domain/user"
)

type UserRepository struct {
    db contract.DB
}

func NewUserRepository(db contract.DB) user.Store {
    return &UserRepository{db: db}
}

func (r *UserRepository) Create(u *user.User) error {
    return r.db.Instance().Create(u).Error
}

func (r *UserRepository) FindByID(id uint) (*user.User, error) {
    var u user.User
    if err := r.db.Instance().First(&u, id).Error; err != nil {
        return nil, err
    }
    return &u, nil
}

// ... other methods
```

```go
// internal/app/myapi/handler/user.go
package handler

import (
    "github.com/gofiber/fiber/v3"
    "yourproject/internal/app/myapi/service"
)

type UserHandler struct {
    userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
    return &UserHandler{userService: userService}
}

func (h *UserHandler) CreateUser(c fiber.Ctx) error {
    var req struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    user, err := h.userService.CreateUser(req.Name, req.Email)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
    }
    
    return c.JSON(user)
}
```

```go
// cmd/myapi/main.go
package main

import (
    "go.oease.dev/goe/v2"
    "yourproject/internal/app/myapi/handler"
    "yourproject/internal/app/myapi/service"
    "yourproject/internal/repository"
)

func main() {
    goe.New(goe.Options{
        WithHTTP: true,
        WithDB:   true,
        Providers: []any{
            repository.NewUserRepository,
            service.NewUserService,
            handler.NewUserHandler,
        },
        Invokers: []any{registerRoutes},
    })
    
    goe.Run()
}

func registerRoutes(userHandler *handler.UserHandler) {
    app := goe.HTTP().App()
    
    api := app.Group("/api")
    api.Post("/users", userHandler.CreateUser)
}
```

## Getting Started

Start with a simpler version of this structure and add directories as your project's needs evolve. The key is consistency and ensuring your team understands the layout.

For a small project, you might start with:

```
my-simple-goe-app/
├── cmd/
│   └── app/
│       └── main.go
├── internal/
│   ├── handler/
│   ├── service/
│   └── repository/
├── .env.example
├── go.mod
└── README.md
```

## Next Steps

- [**Architecture**](./architecture.md) - Deep dive into GOE's architecture
- [**Configuration**](./configuration.md) - Learn about configuration management
- [**Modules**](./modules.md) - Create custom modules for your application