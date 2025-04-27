# Config Module

The Config module provides a simple interface for accessing configuration values from environment variables or configuration files. It implements the `contracts.Config` interface.

## Features

- Environment variable support
- Configuration file support (JSON, YAML, etc.)
- Default values for missing configuration
- Type conversion helpers

## Usage

### Initialization

The config module is automatically initialized by the GOE framework:

```go
// Initialize with a specific config directory
configModule := config.New("./configs")
```

### Accessing Configuration Values

```go
// Get the config instance
config := goe.UseCfg()

// Get a string value with a default
appName := config.GetOrDefaultString("APP_NAME", "MyApp")

// Get an integer value with a default
port := config.GetOrDefaultInt("HTTP_PORT", 3000)

// Get a boolean value with a default
debug := config.GetOrDefaultBool("DEBUG", false)

// Get a string slice
allowedOrigins := config.GetStringSlice("ALLOWED_ORIGINS")

// Get a raw value (returns interface{})
value := config.Get("SOME_KEY")
```

### Configuration Files

The config module can load configuration from files in the specified directory. The files should be named according to the environment (e.g., `dev.json`, `prod.json`).

Example JSON configuration file:

```json
{
  "APP_NAME": "MyApp",
  "APP_VERSION": "1.0.0",
  "HTTP_PORT": 3000,
  "DEBUG": true,
  "ALLOWED_ORIGINS": ["http://localhost:3000", "https://example.com"]
}
```

## Implementation Details

The config module uses a simple key-value store to manage configuration values. It first checks for environment variables, then falls back to configuration files if the value is not found in the environment.

The module supports various data types and provides helper methods for type conversion.