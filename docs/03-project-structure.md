# 3. Recommended Project Structure 📂

A well-organized project structure is crucial for maintainability, scalability, and collaboration. While Goe doesn't enforce a strict layout, this section outlines a recommended structure that aligns with common Go practices and works well with Goe applications.

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
│   │       └── model/      # Data transfer objects, request/response models (distinct from domain)
│   ├── domain/             # Core domain models and business logic (agnostic of delivery mechanism)
│   │   ├── user/           # Example: user domain
│   │   │   ├── user.go     # User entity, core logic
│   │   │   └── store.go    # Interface for user persistence
│   │   └── product/
│   │       ├── product.go
│   │       └── store.go
│   ├── config/             # Configuration loading and struct definitions
│   ├── module/             # Custom Goe modules specific to your application
│   ├── platform/           # Platform-level concerns
│   │   ├── database/       # Database interaction (GORM setup, migrations if not external)
│   │   ├── cache/          # Cache interaction setup
│   │   └── log/            # Logging setup if more customization is needed
│   └── repository/         # Data persistence implementations (e.g., GORM implementations of domain.Store)
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
├── migrations/             # Database migration files (if using a migration tool like Goose, Atlas)
│
├── build/                  # Build artifacts, scripts, Dockerfiles
│   ├── package/            # Packaging scripts, configurations
│   ├── docker/             # Dockerfiles for different services
│   │   ├── myapi.Dockerfile
│   │   └── myworker.Dockerfile
│   └── ci/                 # CI/CD pipeline configurations
│
├── docs/                   # Project documentation (like this one!)
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

*   **`cmd/`**:
    *   Contains the `main` packages for your executables.
    *   Typically, each subdirectory in `cmd/` corresponds to a single binary (e.g., a web server, a CLI tool, a worker process).
    *   Code here should be minimal, primarily responsible for initializing and running the application using packages from `internal/` or `pkg/`.
    *   **How Goe fits**: Your `main.go` in these directories will call `goe.New()` and `goe.Run()`.

*   **`internal/`**:
    *   This is where most of your application's private code resides. Go enforces that code in `internal/` can only be imported by code within the same parent directory (i.e., your project).
    *   **`app/`**: Can be used to structure application-specific logic, such as defining services, handlers, and request/response models specific to one of your `cmd/` applications. This helps if you have multiple applications in `cmd/` that share some `domain/` logic but have different front-facing APIs.
    *   **`domain/`**: Contains your core business logic and entities. This code should be independent of how it's exposed (e.g., HTTP, gRPC) or how it's stored. It defines *what* your application does.
    *   **`config/`**: Manages application configuration loading and provides typed access. Goe's config module handles much of this, but you might store your specific config structs here.
    *   **`module/`**: If you create custom Goe modules (see the [Modules](10-modules.md) chapter), they can live here.
    *   **`platform/`**: For setting up and configuring platform-level concerns like database connections, cache clients, or custom logging initialization that might be shared across your internal applications. Goe's modules provide a lot of this, but you might have app-specific wiring here.
    *   **`repository/`**: Implements data access logic, such as database interactions using GORM. These repositories often implement interfaces defined in your `domain/` layer (e.g., `UserRepository` implementing `user.Store`).
    *   **How Goe fits**: Most of your application-specific Fx providers, invokers, services, handlers, and repositories will live within `internal/`. Goe's core modules (`contract.Config`, `contract.Logger`, etc.) will be injected into your components here.

*   **`pkg/`**:
    *   Use this directory for code that's safe to be imported and used by external applications. If you don't plan to share any code, you might not need this directory.
    *   Be mindful of what you put here, as it becomes part of your public API if the repository is public.

*   **`configs/`**:
    *   Store example configuration files (`.env.example`) and potentially default configurations.
    *   Actual `.env` files (containing secrets) should be in your project root (and listed in `.gitignore`) or managed via environment variables in deployment.
    *   **How Goe fits**: Goe's config module will load `.env` files from the root or based on `GOE_ENV`. This directory helps manage templates for those files.

*   **`web/`**:
    *   For projects serving HTML directly or hosting Single Page Applications.
    *   **How Goe fits**: GoFiber (Goe's HTTP engine) can be configured to serve static files from `web/static` or render templates from `web/templates`.

*   **`migrations/`**:
    *   Essential for managing database schema changes over time in a controlled manner.
    *   Use tools like [Goose](https://github.com/pressly/goose), [Atlas](https://atlasgo.io/), or GORM's own migration features.
    *   **How Goe fits**: While Goe's DB module can do auto-migration, for production, it's better to use dedicated migration files stored here.

*   **`build/`**:
    *   Contains scripts, Dockerfiles, and configurations related to building and packaging your application.

*   **`docs/`**:
    *   Your project's detailed documentation.

*   **`scripts/`**:
    *   Utility scripts for development, testing, deployment, etc. (e.g., shell scripts, Makefiles).

*   **`test/`**:
    *   Can house end-to-end tests, integration tests that span multiple internal packages, and common test data or helpers. Unit tests usually live alongside the code they test (e.g., `user_test.go` next to `user.go`).

## Benefits of This Structure

*   **Clear Separation of Concerns**: Different aspects of your application (domain logic, HTTP handling, data access) are neatly organized.
*   **Improved Navigability**: Easier for developers (including your future self) to find code.
*   **Scalability**: As your project grows, this structure can accommodate new features and components more gracefully.
*   **Testability**: Well-defined boundaries between packages make unit and integration testing easier.
*   **Maintainability**: Easier to understand, debug, and refactor code.

## How Goe Adapts

Goe is designed to be flexible and doesn't impose this structure. However, its features complement this layout well:

*   **Dependency Injection (Fx)**: Encourages defining components (services, handlers, repositories) in their respective packages within `internal/` and wiring them together in your `cmd/.../main.go` or through dedicated Fx modules.
*   **Modules**: Custom Goe modules can be placed in `internal/module/` to encapsulate specific functionalities with their own lifecycle hooks.
*   **Configuration**: Goe's config system loads `.env` files from the project root, aligning with placing `configs/.env.example` for reference.
*   **Contracts**: Goe's use of interfaces (`contract.Logger`, `contract.DB`, etc.) makes it easy to inject these framework services into your application components located anywhere in `internal/`.

Start with a simpler version of this structure and add directories as your project's needs evolve. The key is consistency and ensuring your team understands the layout.

Next, let's dive into [Goe's Architecture](04-architecture.md).
