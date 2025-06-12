# 5. Configuration Management ⚙️

Effective configuration management is essential for any application, allowing it to behave differently across various environments (development, testing, production) without code changes. Goe provides a flexible and straightforward configuration system.

## Configuration Sources & Loading Order

Goe loads configuration from multiple sources, with later sources overriding earlier ones. This provides a clear precedence:

1.  **`.env` file**: If a file named `.env` exists in the root of your project, it's loaded first. This is suitable for project-wide defaults that are not secret.
2.  **`.local.env` file**: If a file named `.local.env` exists in the project root, it's loaded next. This file is intended for local overrides (e.g., specific developer settings) and **should be added to your `.gitignore` file** to avoid committing it.
3.  **Environment-specific `.env` file (e.g., `.dev.env`, `.prod.env`)**:
    *   Goe checks the `GOE_ENV` environment variable.
    *   If `GOE_ENV` is set (e.g., to `dev`, `prod`, `staging`), Goe attempts to load a corresponding file like `.dev.env` or `.prod.env`.
    *   If `GOE_ENV` is not set, it defaults to `dev`, so Goe will attempt to load `.dev.env`.
    *   This allows you to have environment-specific configuration files.
4.  **System Environment Variables**: Finally, Goe loads configuration values from the system's environment variables. This is the highest precedence and is the recommended way to provide sensitive credentials or settings in production environments.

**Example Precedence:**

*   If `DB_HOST` is in `.env` as `localhost` and in `.prod.env` (with `GOE_ENV=prod`) as `prod.db.example.com`, and also set as a system environment variable to `override.db.example.com`, the value used will be `override.db.example.com`.
*   If `API_KEY` is only in `.local.env`, that value will be used (unless overridden by a system environment variable).

### `.env` File Format

`.env` files should contain key-value pairs, one per line:

```env
# This is a comment
APP_NAME="My Goe Application"
APP_PORT=8080
DEBUG_MODE=true
DATABASE_URL="postgres://user:pass@host:port/dbname"

# Values with spaces should be quoted
FEATURE_FLAG_MESSAGE="Hello World from Config"

# Empty lines are ignored
```

*   Keys are typically uppercase with underscores.
*   Values are parsed as strings. Goe's accessor methods will handle type conversion.
*   Quotes around values are optional but good practice if values contain spaces or special characters; the quotes will be trimmed during parsing.
*   Lines starting with `#` are treated as comments.

## Accessing Configuration Values

Goe provides two primary ways to access configuration values:

### 1. Global Accessor (`goe.Config()`)

You can easily access the global configuration instance anywhere in your code:

```go
package main

import (
	"fmt"
	"go.oease.dev/goe/v2"
	"time"
)

func main() {
	// Initialize Goe (Config module is implicitly initialized)
	_ = goe.New(goe.Options{}) // Basic initialization

	// Accessing config values
	appName := goe.Config().GetString("APP_NAME")
	appPort := goe.Config().GetInt("APP_PORT")
	debugMode := goe.Config().GetBool("DEBUG_MODE")
	timeout := goe.Config().GetDuration("API_TIMEOUT") // e.g., API_TIMEOUT=30s in .env
	adminUsers := goe.Config().GetStringSlice("ADMIN_USERS") // e.g., ADMIN_USERS=admin,editor,viewer

	fmt.Printf("App Name: %s
", appName)
	fmt.Printf("App Port: %d
", appPort)
	fmt.Printf("Debug Mode: %t
", debugMode)
	fmt.Printf("API Timeout: %s
", timeout.String())
	fmt.Printf("Admin Users: %v
", adminUsers)

    // Check if a key exists
    if goe.Config().Has("NON_EXISTENT_KEY") {
        fmt.Println("NON_EXISTENT_KEY exists")
    } else {
        fmt.Println("NON_EXISTENT_KEY does not exist")
    }

    // Get all configuration values
    // allConfig := goe.Config().All()
    // fmt.Printf("All config: %v
", allConfig) // Be careful logging all config in production
}
```

### 2. Dependency Injection (`contract.Config`)

For better testability and explicit dependencies, especially in your services and handlers, inject the `contract.Config` interface:

```go
package myapp

import (
	"fmt"
	"go.oease.dev/goe/v2/contract"
	"time"
)

type MyService struct {
	config contract.Config
	logger contract.Logger
}

// NewMyService is a constructor Fx can use
func NewMyService(cfg contract.Config, log contract.Logger) *MyService {
	return &MyService{config: cfg, logger: log}
}

func (s *MyService) PerformAction() {
	apiEndpoint := s.config.GetString("EXTERNAL_API_ENDPOINT")
	retryAttempts := s.config.GetInt("RETRY_ATTEMPTS")

	s.logger.Info(fmt.Sprintf("Contacting API: %s with %d retries", apiEndpoint, retryAttempts))
	// ... perform action
}
```
To make `MyService` available via Fx, you would provide its constructor:
```go
// In your main.go or a module file
// fx.Provide(NewMyService)
```

### Available Accessor Methods

The `contract.Config` interface provides the following methods:

*   `Get(key string) any`: Retrieves a value as `any` (interface{}). Returns `nil` if not found.
*   `GetString(key string) string`: Retrieves a value as a string. Returns `""` if not found or not a string.
*   `GetInt(key string) int`: Retrieves a value as an int. Returns `0` if not found or conversion fails.
*   `GetInt64(key string) int64`: Retrieves a value as an int64. Returns `0` if not found or conversion fails.
*   `GetFloat64(key string) float64`: Retrieves a value as a float64. Returns `0.0` if not found or conversion fails.
*   `GetBool(key string) bool`: Retrieves a value as a boolean.
    *   Truthy values: `"true"`, `"1"`, `"yes"`, `"on"` (case-insensitive).
    *   Falsy values: `"false"`, `"0"`, `"no"`, `"off"` (case-insensitive), or if the key is not found.
*   `GetDuration(key string) time.Duration`: Retrieves a value as `time.Duration` (e.g., "30s", "5m", "1h"). Returns `0` if not found or conversion fails.
*   `GetStringSlice(key string) []string`: Retrieves a value as a slice of strings. Expects a comma-separated string (e.g., `USER_ROLES=admin,user,guest`). Returns an empty slice if not found.
*   `GetStringMap(key string) map[string]any`: Retrieves a map where keys are sub-keys of the provided key (e.g., for `DB.HOST=localhost`, `GetStringMap("DB")` would include `{"HOST": "localhost"}`).
*   `Has(key string) bool`: Checks if a configuration key exists.
*   `All() map[string]any`: Returns all loaded configuration values as a map. **Use with caution in production, as it might expose sensitive data in logs if printed.**
*   `Set(key string, value any)`: Sets a configuration value at runtime. This primarily affects the in-memory representation and might not be suitable for all use cases.
*   `Reload() error`: Manually triggers a reload of configuration from all sources.

## Hot Reloading (Experimental)

Goe's configuration module has a basic built-in mechanism to periodically check for changes in `.env` files and reload them. This is done via a simple ticker that runs `Reload()` every 30 seconds (as seen in `core/config/config.go`'s `watchEnvFiles`).

**Important Considerations for Hot Reloading:**

*   **Simplicity**: The current implementation is a basic periodic poll. For more robust file watching in production, a library like `fsnotify` would be more appropriate (this might be enhanced in future Goe versions).
*   **Scope**: This hot reloading primarily applies to values loaded from the `.env` files that Goe directly monitors. System environment variables are typically immutable for a running process.
*   **Application Awareness**: Your application components need to be written to re-fetch configuration if they expect to see changes, or you need a mechanism (like an event bus) to notify them of configuration changes. Simply reloading the config in Goe's config module doesn't automatically update values already read and stored by your application parts.
*   **Custom Sources**: If you add custom `ConfigSource` implementations, their `Watch` method would be responsible for handling changes from that source.

Due to these factors, relying heavily on automatic hot reloading for critical runtime changes should be approached with caution. It's often safer to restart an application instance to pick up significant configuration changes, especially in production.

## Best Practices for Configuration

*   **Environment Variables for Production**: Use system environment variables for all settings in production, especially secrets. This is a standard and secure practice.
*   **`.env` for Development**: Use `.env` and `.local.env` for ease of development.
*   **`.env.example`**: Always include an `.env.example` file in your repository. This file should list all necessary environment variables with placeholder or default values, serving as documentation for developers. **Do not commit actual secrets to this file.**
*   **`.gitignore`**: Add `.env`, `.local.env`, and any environment-specific `.env` files (like `.prod.env` if it contains secrets) to your `.gitignore`.
*   **Type Safety**: Define structs that map to your configuration sections and populate them at startup. This provides better type safety than using string-based lookups everywhere. Goe's current accessors are by key, but you can build this layer on top.
*   **Centralized Access**: While Goe provides global access, consider wrapping config access for specific domains or modules within your own services or utility functions if it simplifies logic or improves testability.
*   **Defaults in Code**: For non-sensitive parameters, you can provide defaults directly in your code when accessing configuration values if a key might be missing.
    ```go
    // Example of providing a default
    port := goe.Config().GetInt("APP_PORT")
    if port == 0 { // GetInt returns 0 if not found or conversion error
        port = 8080 // Default port
    }
    ```

## Common Configuration Keys

Here are some common configuration keys used by Goe's core modules (refer to specific module documentation for exhaustive lists and defaults):

*   **Application:**
    *   `GOE_ENV`: (e.g., `dev`, `prod`, `test`) Determines which `.env.{GOE_ENV}` file is loaded. Defaults to `dev`.
    *   `APP_NAME`: Application name (used in Fiber, potentially logging).
    *   `APP_VERSION`: Application version.

*   **Logging (`LOG_*`):**
    *   `LOG_LEVEL`: (e.g., `debug`, `info`, `warn`, `error`). Default: `info`.
    *   `LOG_FORMAT`: (e.g., `text`, `json`). Default: `text` (pretty for dev, json for prod if `GOE_ENV=prod`).
    *   `LOG_OUTPUT`: Comma-separated list (e.g., `console`, `file`). Default: `console`.
    *   `LOG_CALLER`: (`true`/`false`) Whether to include caller information. Default: `true`.
    *   `LOG_STACKTRACE`: (`true`/`false`) Whether to include stack traces on errors. Default: `false`.

*   **HTTP Server (`HTTP_*`, `FIBER_*`):**
    *   `HTTP_HOST`: Host to bind to. Default: `0.0.0.0`.
    *   `HTTP_PORT`: Port to listen on. Default: `8080`.
    *   `HTTP_READ_TIMEOUT`, `HTTP_WRITE_TIMEOUT`, `HTTP_IDLE_TIMEOUT`: Server timeouts. Defaults: `10s`, `10s`, `30s`.
    *   `FIBER_SERVER_HEADER`: Value for the `Server` HTTP header. Default: "Goe".
    *   `FIBER_BODY_LIMIT`: Max request body size in bytes. Default: `4 * 1024 * 1024` (4MB).
    *   Many other `FIBER_*` keys exist for fine-tuning GoFiber behavior (see HTTP module docs or Fiber docs).

*   **Database (`DB_*`):**
    *   `DB_DRIVER`: (e.g., `mysql`, `postgres`, `sqlite`, `sqlserver`).
    *   `DB_HOST`, `DB_PORT`, `DB_DATABASE`, `DB_USERNAME`, `DB_PASSWORD`.
    *   `DB_DSN`: Full Data Source Name (can override individual parts).
    *   `DB_MAX_IDLE_CONNS`, `DB_MAX_OPEN_CONNS`, `DB_CONN_MAX_LIFETIME`, `DB_CONN_MAX_IDLE_TIME`: Connection pool settings.
    *   `DB_LOG_MODE`: (`true`/`false`) Enable/disable GORM's logger.
    *   `DB_AUTO_MIGRATE`: (`true`/`false`) Hint for auto-migration (actual migration requires code).

*   **Cache (`CACHE_*`):**
    *   `CACHE_DRIVER`: Default cache driver (e.g., `memory`, `redis`). Default: `memory`.
    *   `CACHE_PREFIX`: Global prefix for all cache keys.
    *   `CACHE_REDIS_HOST`, `CACHE_REDIS_PORT`, `CACHE_REDIS_PASSWORD`, `CACHE_REDIS_DB`: Example for Redis driver.

Always refer to the specific module's documentation for the most accurate and complete list of configuration options.

Next, let's look into [Logging](06-logging.md).
```
