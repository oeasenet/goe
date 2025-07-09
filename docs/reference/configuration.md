# Configuration Options

GOE Framework provides comprehensive configuration options for all core modules. Configuration can be loaded from environment variables, configuration files, or set programmatically.

## Configuration Structure

```go
type Config struct {
    Server       ServerConfig       `yaml:"server"`
    Database     DatabaseConfig     `yaml:"database"`
    Cache        CacheConfig        `yaml:"cache"`
    Logger       LoggerConfig       `yaml:"logger"`
    Observability ObservabilityConfig `yaml:"observability"`
}
```

## Server Configuration

```go
type ServerConfig struct {
    Port                int    `yaml:"port" env:"SERVER_PORT" default:"8080"`
    Host                string `yaml:"host" env:"SERVER_HOST" default:"localhost"`
    ReadTimeout         int    `yaml:"read_timeout" env:"SERVER_READ_TIMEOUT" default:"30"`
    WriteTimeout        int    `yaml:"write_timeout" env:"SERVER_WRITE_TIMEOUT" default:"30"`
    IdleTimeout         int    `yaml:"idle_timeout" env:"SERVER_IDLE_TIMEOUT" default:"120"`
    MaxRequestBodySize  int    `yaml:"max_request_body_size" env:"SERVER_MAX_REQUEST_BODY_SIZE" default:"4194304"`
    Prefork             bool   `yaml:"prefork" env:"SERVER_PREFORK" default:"false"`
    DisableKeepalive    bool   `yaml:"disable_keepalive" env:"SERVER_DISABLE_KEEPALIVE" default:"false"`
}
```

## Database Configuration

```go
type DatabaseConfig struct {
    Driver          string `yaml:"driver" env:"DB_DRIVER" default:"postgres"`
    DSN             string `yaml:"dsn" env:"DB_DSN" required:"true"`
    MaxOpenConns    int    `yaml:"max_open_conns" env:"DB_MAX_OPEN_CONNS" default:"25"`
    MaxIdleConns    int    `yaml:"max_idle_conns" env:"DB_MAX_IDLE_CONNS" default:"25"`
    ConnMaxLifetime int    `yaml:"conn_max_lifetime" env:"DB_CONN_MAX_LIFETIME" default:"300"`
    ConnMaxIdleTime int    `yaml:"conn_max_idle_time" env:"DB_CONN_MAX_IDLE_TIME" default:"300"`
    MigrationsPath  string `yaml:"migrations_path" env:"DB_MIGRATIONS_PATH" default:"migrations"`
}
```

## Cache Configuration

```go
type CacheConfig struct {
    Driver       string `yaml:"driver" env:"CACHE_DRIVER" default:"memory"`
    DSN          string `yaml:"dsn" env:"CACHE_DSN"`
    DefaultTTL   int    `yaml:"default_ttl" env:"CACHE_DEFAULT_TTL" default:"3600"`
    MaxKeys      int    `yaml:"max_keys" env:"CACHE_MAX_KEYS" default:"1000"`
    CleanupInterval int `yaml:"cleanup_interval" env:"CACHE_CLEANUP_INTERVAL" default:"600"`
}
```

## Logger Configuration

```go
type LoggerConfig struct {
    Level       string `yaml:"level" env:"LOG_LEVEL" default:"info"`
    Format      string `yaml:"format" env:"LOG_FORMAT" default:"json"`
    Output      string `yaml:"output" env:"LOG_OUTPUT" default:"stdout"`
    Filename    string `yaml:"filename" env:"LOG_FILENAME"`
    MaxSize     int    `yaml:"max_size" env:"LOG_MAX_SIZE" default:"100"`
    MaxAge      int    `yaml:"max_age" env:"LOG_MAX_AGE" default:"28"`
    MaxBackups  int    `yaml:"max_backups" env:"LOG_MAX_BACKUPS" default:"3"`
    Compress    bool   `yaml:"compress" env:"LOG_COMPRESS" default:"false"`
}
```

## Observability Configuration

```go
type ObservabilityConfig struct {
    Metrics   MetricsConfig   `yaml:"metrics"`
    Tracing   TracingConfig   `yaml:"tracing"`
    Health    HealthConfig    `yaml:"health"`
}

type MetricsConfig struct {
    Enabled   bool   `yaml:"enabled" env:"METRICS_ENABLED" default:"true"`
    Path      string `yaml:"path" env:"METRICS_PATH" default:"/metrics"`
    Namespace string `yaml:"namespace" env:"METRICS_NAMESPACE" default:"goe"`
}

type TracingConfig struct {
    Enabled     bool    `yaml:"enabled" env:"TRACING_ENABLED" default:"false"`
    ServiceName string  `yaml:"service_name" env:"TRACING_SERVICE_NAME" default:"goe-app"`
    Endpoint    string  `yaml:"endpoint" env:"TRACING_ENDPOINT"`
    SampleRate  float64 `yaml:"sample_rate" env:"TRACING_SAMPLE_RATE" default:"1.0"`
}

type HealthConfig struct {
    Enabled bool   `yaml:"enabled" env:"HEALTH_ENABLED" default:"true"`
    Path    string `yaml:"path" env:"HEALTH_PATH" default:"/health"`
}
```

## Configuration Loading

### From Environment Variables

```bash
export SERVER_PORT=8080
export DB_DSN="postgres://user:pass@localhost/db"
export LOG_LEVEL="debug"
```

### From Configuration File

```yaml
# config.yaml
server:
  port: 8080
  host: "0.0.0.0"

database:
  driver: "postgres"
  dsn: "postgres://user:pass@localhost/db"
  max_open_conns: 25

logger:
  level: "info"
  format: "json"

observability:
  metrics:
    enabled: true
    path: "/metrics"
  tracing:
    enabled: true
    service_name: "my-app"
    endpoint: "http://jaeger:14268/api/traces"
```

### Programmatic Configuration

```go
config := &Config{
    Server: ServerConfig{
        Port: 8080,
        Host: "0.0.0.0",
    },
    Database: DatabaseConfig{
        Driver: "postgres",
        DSN:    "postgres://user:pass@localhost/db",
    },
    Logger: LoggerConfig{
        Level:  "info",
        Format: "json",
    },
}
```

## Configuration Validation

GOE automatically validates configuration using struct tags:

```go
type ServerConfig struct {
    Port int `yaml:"port" validate:"required,min=1,max=65535"`
    Host string `yaml:"host" validate:"required,hostname"`
}
```

## Environment-Specific Configuration

Use different configuration files for different environments:

```
config/
├── config.yaml          # Default configuration
├── config.dev.yaml      # Development overrides
├── config.prod.yaml     # Production overrides
└── config.test.yaml     # Test overrides
```

Load environment-specific config:

```go
config, err := LoadConfig("config", "dev")
if err != nil {
    log.Fatal(err)
}
```

## Configuration Best Practices

1. **Use environment variables for secrets**: Never store sensitive data in configuration files
2. **Provide sensible defaults**: All configuration should have reasonable defaults
3. **Validate configuration**: Use validation tags to ensure configuration is correct
4. **Document configuration**: Provide clear documentation for all configuration options
5. **Use type-safe configuration**: Prefer structured configuration over string-based keys