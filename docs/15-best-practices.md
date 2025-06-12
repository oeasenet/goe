# 15. Goe Framework Best Practices 🌟

To make the most of the Goe framework and build high-quality, maintainable, and scalable applications, consider the following best practices. These are guidelines, not strict rules, and should be adapted to your project's specific needs.

## 1. Project Structure

*   **Follow Standard Go Layout**: Adhere to a structure similar to the one outlined in the [Project Structure](03-project-structure.md) guide (e.g., `cmd/`, `internal/`, `pkg/`). This improves consistency and navigability.
*   **Domain-Driven Grouping**: Within `internal/`, group code by domain or feature (e.g., `internal/user/`, `internal/product/`). Each domain package can contain its own models, services, repositories, and handlers.
*   **Clear Separation**: Maintain a clear separation between:
    *   **`cmd/`**: Application entry points (minimal code).
    *   **`internal/app` or `internal/<domain>/handler`**: Request/response handling (HTTP, gRPC, etc.).
    *   **`internal/<domain>/service`**: Business logic orchestration.
    *   **`internal/domain`**: Core domain entities and business rules, independent of delivery mechanisms.
    *   **`internal/platform` or `internal/<domain>/repository`**: Data persistence and external service interactions.

## 2. Configuration Management

*   **Environment Variables for Production**: Store all configuration, especially secrets, in environment variables for production deployments.
*   **`.env` Files for Development**: Use `.env` and `.local.env` (git-ignored) for local development convenience.
*   **`.env.example`**: Always commit an `.env.example` file listing all required environment variables with placeholder or default values.
*   **Type-Safe Access**: While Goe provides direct accessors like `GetString`, `GetInt`, consider creating typed configuration structs for your application and populating them at startup for better compile-time safety.
*   **Centralized Defaults**: Define sensible default values for configurations either in code (when a config key is missing) or in a base `.env` file.

## 3. Logging

*   **Structured Logging**: Always use structured logging with key-value fields (`contract.NewField(...)`). Avoid formatting data directly into log messages.
*   **Log at Appropriate Levels**: Use `Debug`, `Info`, `Warn`, `Error`, `Fatal` meaningfully. Don't overuse `Error` for non-critical issues.
*   **Contextual Logging**: Use `logger.With(...)` to add persistent contextual fields (e.g., `request_id`, `user_id`, `module_name`) to your loggers. The HTTP module already provides a request-scoped logger.
*   **Log Errors, Don't Ignore Them**: When an error is handled or bubbles up to a boundary, log it with sufficient context.
*   **Avoid Sensitive Information**: Be extremely careful not to log passwords, API keys, or other sensitive PII unless absolutely necessary and properly secured/masked.
*   **Consistent Field Names**: Establish conventions for common log field names (e.g., `user_id`, `trace_id`, `error_code`).

## 4. Error Handling

*   **Explicit Error Returns**: Follow Go's standard practice of returning `error` values.
*   **Wrap Errors for Context**: Use `fmt.Errorf("operation failed: %w", err)` to add context while preserving the original error. This is crucial for debugging.
*   **Sentinel vs. Custom Types**:
    *   Use sentinel errors (e.g., `var ErrNotFound = errors.New("not found")`) for common, fixed error conditions. Check with `errors.Is()`.
    *   Use custom error types (structs implementing `error`) when you need to convey more structured information with an error. Check with `errors.As()`.
*   **HTTP Error Codes**: In HTTP handlers, translate internal errors into appropriate HTTP status codes using `fiber.NewError(statusCode, message)`. Let unhandled internal errors become 500s (which Goe logs).
*   **User-Friendly Messages**: For client-facing errors (especially 4xx), provide clear, user-friendly messages. Avoid exposing internal error details directly.

## 5. Dependency Injection (Fx)

*   **Prefer Constructor Injection**: Inject dependencies via constructors. This makes components easier to test and their dependencies explicit.
*   **Depend on Interfaces (Contracts)**: Design your services and components to depend on interfaces (like Goe's `contract.Logger`, `contract.DB`, or your own repository/service interfaces) rather than concrete implementations.
*   **`fx.In` for Readability**: If a constructor takes many dependencies (e.g., >3), use an `fx.In` struct to group them.
*   **`fx.Module` for Organization**: Group related Fx providers and invokers into `fx.Module`s for better organization in larger applications.
*   **Clear Provider Functions**: Name your provider functions descriptively (e.g., `NewUserService`, `NewPostgresTaskRepository`).
*   **Avoid Global Accessors in Core Logic**: While Goe provides global accessors like `goe.Log()`, try to avoid using them within your core service and business logic layers. Pass dependencies explicitly for better testability. Globals are fine in `main.go` or very high-level setup.

## 6. Module Design (`contract.Module`)

*   **Single Responsibility**: Design Goe modules to have a clear, single responsibility or manage a cohesive set of functionalities.
*   **Resource Management**: Use `OnStart` to acquire/initialize resources and `OnStop` to release/clean them up gracefully.
*   **Context Awareness**: Pay attention to the `context.Context` passed to `OnStart` and `OnStop`. Use it for timeouts and cancellation signals, especially for background goroutines started by your module.
*   **Minimal Dependencies in `New...Module`**: The constructor for your `contract.Module` implementation should ideally take only essential dependencies that are needed for its configuration or for it to provide its own services to Fx. Runtime dependencies for its operations are often resolved within `OnStart` or methods called after startup.

## 7. HTTP Handling (GoFiber)

*   **Thin Handlers/Controllers**: Keep your HTTP handlers (or controller methods) thin. Their primary responsibility should be:
    1.  Parsing and validating input (request body, query params, path params).
    2.  Calling appropriate service methods to perform business logic.
    3.  Formatting the service's response into an HTTP response.
*   **Input Validation**: Always validate incoming data from users or external systems using Goe's validator (`goehttp.GetValidator(c)`) or other validation mechanisms.
*   **Service Layer for Business Logic**: Encapsulate business logic within service layers, keeping it separate from HTTP concerns.
*   **DTOs (Data Transfer Objects)**: Use specific structs for request payloads and response bodies rather than directly exposing your internal domain models, especially if they differ.

## 8. Database Interaction (GORM)

*   **Repository Pattern**: Abstract GORM (or any database interaction) logic behind repository interfaces. Your services should depend on these interfaces, not directly on GORM.
*   **Production Migrations**: Use dedicated migration tools (Goose, Atlas, GORM's migrator tool) for managing database schema changes in production. Avoid relying solely on GORM's `AutoMigrate` for production environments.
*   **Connection Pooling**: Configure connection pool settings (`DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, etc.) appropriately for your expected load.
*   **Error Handling**: Check for `gorm.ErrRecordNotFound` specifically and handle it as a "not found" case rather than a generic error where appropriate.

## 9. Caching

*   **Cache Wisely**: Cache data that is expensive to compute or fetch and doesn't change too frequently.
*   **Appropriate TTLs**: Set sensible Time-To-Live (TTL) values.
*   **Cache Invalidation Strategy**: Plan for how to invalidate cache entries when underlying data changes.
*   **Consistent Key Naming**: Use a clear, consistent naming convention for cache keys. Use prefixes to avoid collisions. Goe's cache module helps with global/store-specific prefixes.

## 10. Testing

*   **Write Tests!**: Aim for good test coverage, especially for business logic (unit tests) and key API endpoints (integration tests).
*   **Unit Tests for Isolation**: Mock dependencies to test components in isolation.
*   **Integration Tests for Interactions**: Test how different parts of your application (e.g., handler -> service -> repository) work together. Use a test database for these.
*   **`fxtest` for Module/Fx Testing**: Utilize `go.uber.org/fx/fxtest` for testing components within an Fx lifecycle or for testing Goe modules.
*   **Test Environment Configuration**: Use a separate configuration (e.g., `.env.test`, `GOE_ENV=test`) for your tests, often with in-memory databases or mock external services.

## 11. Security Considerations (Brief)

*   **Input Validation**: Always validate all incoming data from users or external systems to prevent injection attacks, data corruption, etc.
*   **HTTPS**: Run your application behind a reverse proxy (like Nginx or Caddy) that handles HTTPS termination in production.
*   **Secrets Management**: Never hardcode secrets. Use environment variables (managed by your deployment system/secrets manager) for sensitive data.
*   **Dependency Updates**: Keep your Go modules (including Goe and its dependencies) up to date to patch security vulnerabilities. Use `go list -u -m all` and `go get -u`.
*   **Rate Limiting/Authentication**: Implement rate limiting and proper authentication/authorization middleware for your API endpoints. Fiber offers middleware for these.

## 12. Performance

*   **Profiling**: If you encounter performance bottlenecks, use Go's profiling tools (`pprof`) to identify hot spots.
*   **Efficient Queries**: Write efficient database queries and ensure proper indexing.
*   **Caching**: Utilize caching for frequently accessed, expensive operations.
*   **Concurrency**: Leverage Go's goroutines for concurrent processing where appropriate, but be mindful of race conditions (use channels, mutexes correctly).
*   **Minimize Allocations**: In performance-critical code paths, be mindful of memory allocations. Tools like `benchmem` can help. GoFiber itself is designed for low allocations.

By keeping these best practices in mind, you can build robust, maintainable, and efficient applications using the Goe framework.

Next, we will discuss [Deployment Strategies](16-deployment.md).
```
