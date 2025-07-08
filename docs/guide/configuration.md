# Configuration

GOE provides a flexible configuration system that loads settings from environment variables and `.env` files. This makes it easy to manage different configurations for development, staging, and production environments.

## Configuration Sources

GOE loads configuration from multiple sources in the following priority order:

1. **System environment variables** (highest priority)
2. **`.env.{GOE_ENV}` file** (e.g., `.env.production`)
3. **`.env.local` file**
4. **`.env` file** (lowest priority)

## Environment Files

### Basic .env File

Create a `.env` file in your project root:

```bash
# Application
APP_NAME=My GOE App
APP_VERSION=1.0.0
APP_ENV=development

# HTTP Server
HTTP_HOST=0.0.0.0
HTTP_PORT=8080

# Database
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp
DB_USER=postgres
DB_PASSWORD=secret
DB_SSL_MODE=disable

# Caching
CACHE_DRIVER=memory
CACHE_PREFIX=myapp
CACHE_TTL=30m

# Logging
LOG_LEVEL=info
LOG_FORMAT=console
```

### Environment-Specific Files

Create different files for different environments:

```bash
# .env.development
APP_ENV=development
LOG_LEVEL=debug
LOG_FORMAT=console
DB_HOST=localhost

# .env.production
APP_ENV=production
LOG_LEVEL=warn
LOG_FORMAT=json
DB_HOST=prod-db.example.com

# .env.test
APP_ENV=test
LOG_LEVEL=error
DB_NAME=myapp_test
```

Set the environment:

```bash
export GOE_ENV=production
```

## Accessing Configuration

### Using Global Accessor

```go
import "go.oease.dev/goe/v2"

func someFunction() {
    config := goe.Config()
    
    // Get string values
    appName := config.GetString("APP_NAME")
    dbHost := config.GetString("DB_HOST")
    
    // Get integer values
    httpPort := config.GetInt("HTTP_PORT")
    dbPort := config.GetInt("DB_PORT")
    
    // Get boolean values
    debugMode := config.GetBool("DEBUG_MODE")
    
    // Get values with defaults
    timeout := config.GetIntWithDefault("TIMEOUT", 30)
    env := config.GetStringWithDefault("APP_ENV", "development")
}
```

### Using Dependency Injection

```go
import "go.oease.dev/goe/v2/contract"

type DatabaseService struct {
    config contract.Config
}

func NewDatabaseService(config contract.Config) *DatabaseService {
    return &DatabaseService{config: config}
}

func (s *DatabaseService) Connect() error {
    host := s.config.GetString("DB_HOST")
    port := s.config.GetInt("DB_PORT")
    database := s.config.GetString("DB_NAME")
    
    // Use configuration to connect to database
    connectionString := fmt.Sprintf("host=%s port=%d dbname=%s", host, port, database)
    
    return nil
}
```

## Configuration Methods

### Basic Getters

```go
config := goe.Config()

// String values
appName := config.GetString("APP_NAME")

// Integer values
port := config.GetInt("HTTP_PORT")

// Boolean values
debug := config.GetBool("DEBUG_MODE")

// Float values
timeout := config.GetFloat64("TIMEOUT")

// Duration values (parses strings like "30s", "5m", "1h")
cacheTTL := config.GetDuration("CACHE_TTL")
```

### Getters with Defaults

```go
config := goe.Config()

// With default values
httpPort := config.GetIntWithDefault("HTTP_PORT", 8080)
logLevel := config.GetStringWithDefault("LOG_LEVEL", "info")
enableMetrics := config.GetBoolWithDefault("ENABLE_METRICS", false)
requestTimeout := config.GetDurationWithDefault("REQUEST_TIMEOUT", 30*time.Second)
```

### Checking if Keys Exist

```go
config := goe.Config()

if config.Has("DATABASE_URL") {
    databaseURL := config.GetString("DATABASE_URL")
    // Use database URL
}

// Get all keys
keys := config.Keys()
for _, key := range keys {
    fmt.Printf("Key: %s, Value: %s\n", key, config.GetString(key))
}
```

## Configuration Patterns

### Database Configuration

```go
type DatabaseConfig struct {
    Driver   string
    Host     string
    Port     int
    Database string
    Username string
    Password string
    SSLMode  string
}

func NewDatabaseConfig(config contract.Config) *DatabaseConfig {
    return &DatabaseConfig{
        Driver:   config.GetString("DB_DRIVER"),
        Host:     config.GetString("DB_HOST"),
        Port:     config.GetInt("DB_PORT"),
        Database: config.GetString("DB_NAME"),
        Username: config.GetString("DB_USER"),
        Password: config.GetString("DB_PASSWORD"),
        SSLMode:  config.GetStringWithDefault("DB_SSL_MODE", "disable"),
    }
}

func (c *DatabaseConfig) ConnectionString() string {
    return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
        c.Host, c.Port, c.Database, c.Username, c.Password, c.SSLMode)
}
```

### HTTP Server Configuration

```go
type HTTPConfig struct {
    Host         string
    Port         int
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    IdleTimeout  time.Duration
}

func NewHTTPConfig(config contract.Config) *HTTPConfig {
    return &HTTPConfig{
        Host:         config.GetStringWithDefault("HTTP_HOST", "0.0.0.0"),
        Port:         config.GetIntWithDefault("HTTP_PORT", 8080),
        ReadTimeout:  config.GetDurationWithDefault("HTTP_READ_TIMEOUT", 30*time.Second),
        WriteTimeout: config.GetDurationWithDefault("HTTP_WRITE_TIMEOUT", 30*time.Second),
        IdleTimeout:  config.GetDurationWithDefault("HTTP_IDLE_TIMEOUT", 60*time.Second),
    }
}
```

## Environment-Specific Configuration

### Development Setup

```bash
# .env.development
APP_ENV=development
LOG_LEVEL=debug
LOG_FORMAT=console

# Enable debug features
DEBUG_MODE=true
ENABLE_PPROF=true

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp_dev

# Cache
CACHE_DRIVER=memory
```

### Production Setup

```bash
# .env.production
APP_ENV=production
LOG_LEVEL=warn
LOG_FORMAT=json

# Production database
DB_HOST=prod-db.example.com
DB_PORT=5432
DB_NAME=myapp_prod
DB_SSL_MODE=require

# Redis cache
CACHE_DRIVER=redis
CACHE_REDIS_HOST=redis.example.com
CACHE_REDIS_PORT=6379
CACHE_REDIS_PASSWORD=secret
```

### Running Different Environments

```bash
# Development (default)
go run main.go

# Production
GOE_ENV=production go run main.go

# Staging
GOE_ENV=staging go run main.go

# Testing
GOE_ENV=test go test ./...
```

## Configuration in HTTP Handlers

```go
import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2/contract"
)

func NewAppHandler(config contract.Config) *AppHandler {
    return &AppHandler{config: config}
}

type AppHandler struct {
    config contract.Config
}

func (h *AppHandler) GetInfo(c fiber.Ctx) error {
    return c.JSON(fiber.Map{
        "app_name":    h.config.GetString("APP_NAME"),
        "version":     h.config.GetString("APP_VERSION"),
        "environment": h.config.GetString("APP_ENV"),
    })
}

func (h *AppHandler) GetHealth(c fiber.Ctx) error {
    return c.JSON(fiber.Map{
        "status": "healthy",
        "env":    h.config.GetString("APP_ENV"),
    })
}
```

## Configuration Validation

```go
import (
    "errors"
    "go.oease.dev/goe/v2/contract"
)

func ValidateConfig(config contract.Config) error {
    // Check required configuration
    requiredKeys := []string{
        "APP_NAME",
        "DB_HOST",
        "DB_PORT",
        "DB_NAME",
        "DB_USER",
    }
    
    for _, key := range requiredKeys {
        if !config.Has(key) {
            return fmt.Errorf("required configuration key missing: %s", key)
        }
    }
    
    // Validate specific values
    if config.GetInt("HTTP_PORT") <= 0 {
        return errors.New("HTTP_PORT must be a positive integer")
    }
    
    if config.GetString("LOG_LEVEL") == "" {
        return errors.New("LOG_LEVEL cannot be empty")
    }
    
    return nil
}

// Use in your application
func main() {
    goe.New(goe.Options{
        WithHTTP: true,
        WithDB:   true,
        Invokers: []any{
            func(config contract.Config) {
                if err := ValidateConfig(config); err != nil {
                    goe.Log().Fatal("Configuration validation failed", "error", err)
                }
            },
        },
    })
    
    goe.Run()
}
```

## Best Practices

### 1. Use Environment Variables for Secrets

Never commit sensitive data to version control:

```bash
# Good - use environment variables
export DB_PASSWORD=secret123
export API_KEY=key123

# Bad - don't put secrets in .env files in version control
```

### 2. Provide Sensible Defaults

```go
// Always provide defaults for non-critical configuration
httpPort := config.GetIntWithDefault("HTTP_PORT", 8080)
logLevel := config.GetStringWithDefault("LOG_LEVEL", "info")
```

### 3. Group Related Configuration

```bash
# Group related settings with prefixes
HTTP_HOST=0.0.0.0
HTTP_PORT=8080
HTTP_READ_TIMEOUT=30s
HTTP_WRITE_TIMEOUT=30s

DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp
DB_USER=postgres
```

### 4. Use .env.example

Create a `.env.example` file with all required configuration keys:

```bash
# .env.example
APP_NAME=My GOE App
APP_VERSION=1.0.0
APP_ENV=development

HTTP_HOST=0.0.0.0
HTTP_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp
DB_USER=postgres
DB_PASSWORD=change_me
```

### 5. Validate Configuration Early

Validate configuration during application startup:

```go
func main() {
    goe.New(goe.Options{
        WithHTTP: true,
        Invokers: []any{
            func(config contract.Config, logger contract.Logger) {
                if !config.Has("DB_HOST") {
                    logger.Fatal("DB_HOST is required")
                }
                
                if config.GetInt("HTTP_PORT") <= 0 {
                    logger.Fatal("HTTP_PORT must be positive")
                }
            },
        },
    })
    
    goe.Run()
}
```

## Testing Configuration

```go
import (
    "testing"
    "go.oease.dev/goe/v2/contract"
)

type MockConfig struct {
    data map[string]string
}

func (c *MockConfig) GetString(key string) string {
    return c.data[key]
}

func (c *MockConfig) GetInt(key string) int {
    // Implementation for testing
    return 0
}

func TestDatabaseService(t *testing.T) {
    config := &MockConfig{
        data: map[string]string{
            "DB_HOST": "localhost",
            "DB_PORT": "5432",
            "DB_NAME": "testdb",
        },
    }
    
    service := NewDatabaseService(config)
    // Test service with mock config
}
```

## Next Steps

- [**Logging**](./logging.md) - Learn about structured logging
- [**Database**](./database.md) - Set up database connections
- [**HTTP Server**](./http-server.md) - Configure HTTP server settings