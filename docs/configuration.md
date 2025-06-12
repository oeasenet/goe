# Configuration

Configuration in Goe is environment driven. Values are loaded from `.env` files and overridden by environment variables. The active environment is determined by `GOE_ENV` which defaults to `dev`.

## Files

1. `.env` – base settings committed to version control
2. `.local.env` – developer specific overrides (usually gitignored)
3. `.<GOE_ENV>.env` – environment specific (e.g. `.production.env`)

Later files override earlier ones. Finally, system environment variables override all files.

## Accessing Configuration

Use the global helper:

```go
port := goe.Config().GetInt("HTTP_PORT")
```

Or inject the `contract.Config` interface:

```go
func NewService(cfg contract.Config) *Service {
    timeout := cfg.GetDuration("REQUEST_TIMEOUT")
    return &Service{timeout: timeout}
}
```

In HTTP handlers you can retrieve configuration from the request context:

```go
func handler(c fiber.Ctx) error {
    cfg := http.GetConfig(c)
    return c.SendString(cfg.GetString("APP_NAME"))
}
```

## Environment Variables

The table below lists common variables. See the examples and module guides for additional options.

| Variable | Description | Default |
|----------|-------------|---------|
| `GOE_ENV` | Application environment | `dev` |
| `APP_NAME` | Application name | `Goe Application` |
| `APP_VERSION` | Application version | `1.0.0` |
| `HTTP_PORT` | HTTP listen port | `8080` |
| `LOG_LEVEL` | Log level (`debug`, `info`, ...) | `info` |
| `CACHE_DRIVER` | Cache driver (e.g. `memory`, `redis`) | `memory` |
