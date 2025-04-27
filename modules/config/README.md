# Config Module

The Config module provides a simple interface for accessing configuration values from environment variables or configuration files. It implements the [`contracts.Config`](https://github.com/oeasenet/goe/blob/main/contracts/config.go) interface and offers a flexible configuration management solution for your application.

## Features

- Environment variable support
- Configuration file support (.env files)
- Environment-specific configuration files
- Local override configuration files
- Default values for missing configuration
- Type conversion helpers (string, int, bool, slices)
- Automatic loading of configuration files
- Simple and intuitive API

## Usage

### Initialization

The config module is automatically initialized by the GOE framework:

```go
// Initialize with a specific config directory
configModule := config.New("./configs")
```

In a GOE application, you can access the config module using:

```go
// Get the config instance
config := goe.UseCfg()
```

### Configuration Files

The config module supports multiple configuration files:

1. **Base Configuration**: `.env` file in the specified directory
2. **Environment-Specific Configuration**: `.{env}.env` file (e.g., `.dev.env`, `.prod.env`)
3. **Local Override Configuration**: `.local.env` file for local development

The loading order is:
1. Base configuration (`.env`)
2. Environment-specific configuration (if `APP_ENV` is set)
3. Local override configuration (if `APP_ENV` is not set)

Example `.env` file:

```
# App Configuration
APP_NAME=MyApp
APP_VERSION=1.0.0
APP_ENV=dev

# Server Configuration
HTTP_PORT=3000
HTTP_HOST=localhost

# Database Configuration
MONGODB_URI=mongodb://localhost:27017
MONGODB_DB=myapp
```

Example `.dev.env` file (overrides values from `.env` when `APP_ENV=dev`):

```
# Development-specific configuration
DEBUG=true
LOG_LEVEL=debug
```

Example `.local.env` file (overrides values from `.env` for local development):

```
# Local development overrides
HTTP_PORT=8080
MONGODB_URI=mongodb://localhost:27018
```

### Accessing Configuration Values

```go
// Get string values
appName := config.GetString("APP_NAME")
appVersion := config.GetString("APP_VERSION")

// Get string values with defaults
appEnv := config.GetOrDefaultString("APP_ENV", "dev")
logLevel := config.GetOrDefaultString("LOG_LEVEL", "info")

// Get integer values
port := config.GetInt("HTTP_PORT")
workerCount := config.GetOrDefaultInt("WORKER_COUNT", 4)

// Get boolean values
debug := config.GetBool("DEBUG")
enableCache := config.GetOrDefaultBool("ENABLE_CACHE", true)

// Get slice values
allowedOrigins := config.GetStringSlice("ALLOWED_ORIGINS")
ports := config.GetIntSlice("PORTS")
features := config.GetBoolSlice("FEATURES")
```

### Configuration Patterns

#### Feature Flags

```go
// Define feature flags in your configuration
// FEATURES=user_profiles,advanced_search,notifications

// Check if a feature is enabled
features := config.GetStringSlice("FEATURES")
hasAdvancedSearch := false
for _, feature := range features {
    if feature == "advanced_search" {
        hasAdvancedSearch = true
        break
    }
}

if hasAdvancedSearch {
    // Enable advanced search functionality
}
```

#### Environment-Based Configuration

```go
// Get the current environment
env := config.GetOrDefaultString("APP_ENV", "dev")

// Apply environment-specific settings
switch env {
case "dev":
    // Development settings
    enableDebugLogging()
    useMockServices()
case "staging":
    // Staging settings
    enableMetrics()
    useRealServicesWithSandboxData()
case "prod":
    // Production settings
    enableCaching()
    useRealServices()
}
```

## API Reference

The Config module implements the [`contracts.Config`](https://github.com/oeasenet/goe/blob/main/contracts/config.go) interface:

```go
type Config interface {
    // Get retrieves a configuration value as a string
    Get(string) string
    
    // GetString retrieves a configuration value as a string
    GetString(key string) string
    
    // GetInt retrieves a configuration value as an integer
    GetInt(key string) int
    
    // GetBool retrieves a configuration value as a boolean
    GetBool(key string) bool
    
    // GetStringSlice retrieves a comma-separated configuration value as a string slice
    GetStringSlice(key string) []string
    
    // GetIntSlice retrieves a comma-separated configuration value as an integer slice
    GetIntSlice(key string) []int
    
    // GetBoolSlice retrieves a comma-separated configuration value as a boolean slice
    GetBoolSlice(key string) []bool
    
    // GetOrDefaultString retrieves a configuration value as a string with a default value
    GetOrDefaultString(key string, defaultValue string) string
    
    // GetOrDefaultInt retrieves a configuration value as an integer with a default value
    GetOrDefaultInt(key string, defaultValue int) int
    
    // GetOrDefaultBool retrieves a configuration value as a boolean with a default value
    GetOrDefaultBool(key string, defaultValue bool) bool
}
```

## Implementation Details

The config module is implemented in the [`Config`](https://github.com/oeasenet/goe/blob/main/modules/config/config.go) struct, which provides:

- Loading configuration from `.env` files using [godotenv](https://github.com/joho/godotenv)
- Support for environment-specific configuration files
- Support for local override configuration files
- Reading from environment variables
- Type conversion for various data types
- Default values for missing configuration

### Configuration Loading Process

1. The module first loads the base configuration from the `.env` file in the specified directory
2. If `APP_ENV` is set, it loads the environment-specific configuration from `.{env}.env` file
3. If `APP_ENV` is not set, it loads the local override configuration from `.local.env` file
4. All environment variables are loaded into an in-memory map for fast access
5. When a configuration value is requested, it is retrieved from the map and converted to the requested type

### Type Conversion

The module provides type conversion for various data types:

- **String**: Direct retrieval from the environment variables
- **Integer**: Conversion using `strconv.Atoi`
- **Boolean**: Conversion using `strconv.ParseBool`
- **Slices**: Splitting comma-separated values and converting each element

### Error Handling

The module handles errors gracefully:

- If a configuration file is not found, it continues without error
- If a type conversion fails, it returns a default value (0 for integers, false for booleans)
- If a key is not found, it returns an empty string, 0, or false depending on the requested type