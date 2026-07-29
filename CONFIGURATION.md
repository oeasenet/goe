# GOE Framework Configuration Reference

This document provides a comprehensive reference for all environment variables supported by the GOE framework.

## Configuration Loading

GOE loads configuration from multiple sources in the following order (later sources override earlier ones):

1. `.env` - Base environment file
2. `.local.env` - Local overrides (gitignored)
3. `.{GOE_ENV}.env` - Environment-specific file (e.g., `.prod.env`, `.dev.env`)
4. System environment variables - Highest priority

### File Format

Environment files use standard `KEY=VALUE` format:

```bash
# Comments start with #
APP_NAME=My Application
HTTP_PORT=8080
DB_PASSWORD="password with spaces"
```

---

## Core Configuration

### Application Environment

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `GOE_ENV` | string | `dev` | Application environment. Common values: `dev`, `prod`, `production`, `staging`, `test` |
| `APP_NAME` | string | `Goe Application` | Application name used in logging and server headers |
| `APP_VERSION` | string | `1.0.0` | Application version |
| `DISABLE_CONFIG_VALIDATION` | bool | `false` | Skip configuration validation during startup (development only) |

---

## HTTP Server (Fiber)

### Configuring in Go code

Every variable in this section has a code equivalent. Pass options through
`goe.Options.HTTP` and they take precedence over the environment:

```go
import (
    "go.oease.dev/goe/v2"
    goehttp "go.oease.dev/goe/v2/core/http"
)

goe.New(goe.Options{
    // Supplying HTTP options enables the module; WithHTTP: true is not needed.
    HTTP: []goehttp.Option{
        goehttp.WithPort(8080),
        goehttp.WithBodyLimit(16 << 20),
        goehttp.WithTrustProxy(true),
        goehttp.WithTrustProxyConfig(fiber.TrustProxyConfig{
            Proxies: []string{"10.0.0.0/8"},
        }),
    },
})
```

**Precedence — code wins, the environment fills the gaps:**

| Layer | Example | Wins over |
|-------|---------|-----------|
| 1. GOE defaults | `BodyLimit` 4MB | — |
| 2. Environment | `FIBER_BODY_LIMIT=8388608` | GOE defaults |
| 3. Options | `goehttp.WithBodyLimit(16 << 20)` | environment and defaults |
| 4. Escape hatches | `goehttp.WithFiberConfig(func(c *fiber.Config) { ... })` | everything, including GOE's own fields |

A field you do not set in code keeps its environment value, so adding options to
an existing application changes nothing else. Options apply in the order given,
so the last write wins.

**Naming rule.** Every field of `fiber.Config` and `fiber.ListenConfig` is
exposed as `With<FieldName>` taking Fiber's own type — read [Fiber's
documentation](https://docs.gofiber.io), prepend `With`, and that is the option.
GOE only invents names where Fiber has no matching field: `WithHost`,
`WithPort`, `WithRequestID`, `WithRequestIDHeader` and `WithHTMLViews`.

**Beyond the environment.** Options also reach `fiber.ListenConfig`, which has no
environment equivalent at all — most importantly TLS:

```go
HTTP: []goehttp.Option{
    goehttp.WithCertFile("/etc/certs/server.crt"),
    goehttp.WithCertKeyFile("/etc/certs/server.key"),
    goehttp.WithTLSMinVersion(tls.VersionTLS13),
},
```

Also available there: `WithAutoCertManager` (ACME), `WithListenerNetwork` for
unix sockets, `WithEnablePrefork`, `WithDisableStartupMessage` and
`WithEnablePrintRoutes`.

**Escape hatch.** Four `fiber.Config` fields and two `fiber.ListenConfig` fields
are intentionally not wrapped — Fiber's `Services` lifecycle (GOE uses fx
modules), `RegexHandler`, and `GracefulContext`/`ShutdownTimeout` (GOE's
shutdown manager owns those; use `goe.Options.ShutdownTimeout` and
`DrainTimeout`). Reach them, and anything a future Fiber release adds, directly:

```go
goehttp.WithFiberConfig(func(c *fiber.Config) {
    c.RegexHandler = myEngine
}),
```

Hooks run last and always win — including over fields GOE's own features depend
on. Disabling `PassLocalsToContext` breaks `http.WithReqCtx`, and clearing
`StructValidator` disables `Bind` validation. GOE logs a warning naming the field
and the affected feature, but does not override the choice. Use
`WithErrorHandler` and `WithStructValidator` to *replace* those rather than clear
them.

**Errors fail fast.** An invalid option (`WithPort(70000)`), or an invalid
combination (`WithTrustProxyConfig` without `WithTrustProxy(true)`,
`WithCertFile` without `WithCertKeyFile`), aborts startup with every problem
listed at once. When that happens no option is applied at all, so a
half-configured server is never served.

### Basic Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `HTTP_HOST` | string | `0.0.0.0` | HTTP server bind address |
| `HTTP_PORT` | int | `8080` | HTTP server port |
| `HTTP_READ_TIMEOUT` | duration | `10s` | Maximum duration for reading the entire request |
| `HTTP_WRITE_TIMEOUT` | duration | `10s` | Maximum duration for writing the response |
| `HTTP_IDLE_TIMEOUT` | duration | `30s` | Maximum idle time for keep-alive connections |
| `HTTP_REQUEST_ID` | bool | `true` | Register the request-id middleware (reuses an upstream id, generates one when absent, echoes the response header). Set `false` to disable and use your own. |
| `HTTP_REQUEST_ID_HEADER` | string | `X-Request-ID` | Header used to read/set the request id. |

### Fiber-Specific Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `FIBER_SERVER_HEADER` | string | `Goe` | Value of the `Server` response header |
| `FIBER_STRICT_ROUTING` | bool | `false` | Enable strict routing (`/foo` != `/foo/`) |
| `FIBER_CASE_SENSITIVE` | bool | `false` | Enable case-sensitive routing |
| `FIBER_IMMUTABLE` | bool | `false` | Enable immutable context values |
| `FIBER_UNESCAPE_PATH` | bool | `false` | Unescape path before processing |
| `FIBER_BODY_LIMIT` | int | `4194304` | Maximum allowed request body size in bytes (default: 4MB) |
| `FIBER_STREAM_REQUEST_BODY` | bool | `true` | Stream request body to reduce memory usage |
| `FIBER_CONCURRENCY` | int | `262144` | Maximum number of concurrent connections |
| `FIBER_PROXY_HEADER` | string | - | Header used to obtain client IP (e.g., `X-Forwarded-For`) |
| `FIBER_REDUCE_MEMORY` | bool | `false` | Reduce memory usage at the cost of performance |
| `FIBER_ENABLE_IP_VALIDATION` | bool | `false` | Enable IP address validation |

### Trust Proxy Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `FIBER_TRUST_PROXY` | bool | `false` | Enable trusted proxy mode |
| `FIBER_TRUST_PROXIES` | []string | - | Comma-separated list of trusted proxy IPs/CIDRs |
| `FIBER_TRUST_LINK_LOCAL` | bool | `true` | Trust link-local addresses when `FIBER_TRUST_PROXY` is enabled |
| `FIBER_TRUST_LOOPBACK` | bool | `true` | Trust loopback addresses when `FIBER_TRUST_PROXY` is enabled |
| `FIBER_TRUST_PRIVATE` | bool | `true` | Trust private network addresses when `FIBER_TRUST_PROXY` is enabled |

### View Engine Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `VIEWS_ENGINE` | string | - | Template engine to use (currently only `html` is supported) |
| `VIEWS_ROOT` | string | `./views` | Directory containing view templates |
| `VIEWS_EXT` | string | `.gohtml` | File extension for view templates |
| `VIEWS_LAYOUT` | string | - | Default layout template name |

---

## Logging

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `LOG_LEVEL` | string | `info` | Global log level (baseline for every module): `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | string | `text` | Log format: `text` (console) or `json` |
| `LOG_OUTPUT` | []string | `console` | Comma-separated outputs: `console`, `stdout`, `stderr`, or file path |
| `LOG_CALLER` | bool | `false` | Include caller information in logs |
| `LOG_STACKTRACE` | bool | `false` | Include stack traces for error logs |
| `LOG_MODULE_LEVELS` | string | `""` | Per-module level overrides, comma-separated `module:level` pairs (e.g. `gorm:warn,job:debug`). Overrides `LOG_LEVEL` for those modules, up or down. |

**Per-module logging.** `LOG_LEVEL` sets the global baseline; `LOG_MODULE_LEVELS`
adjusts individual modules. Valid levels: `debug`, `info`, `warn`, `error`. Module
names: `log`, `app`, `db`, `gorm`, `mongo`, `mongo-migrate`, `cache`, `lock`,
`job`, `http`, `access`, `health`, `metrics`, `otel`. The `access` module is the
per-request HTTP log (2xx/3xx at Info, 4xx at Warn, 5xx at Error), so `access:warn`
shows only failing requests. Invalid entries are ignored with a startup warning.

```bash
LOG_MODULE_LEVELS=job:debug              # debug only the job module
LOG_LEVEL=debug
LOG_MODULE_LEVELS=gorm:warn,access:warn  # debug everything except GORM and access
```

**What lands where (default `info`):** module **started / stopped / failed**, the HTTP
listen address, "database connected", migrations applied, and graceful shutdown stay at
**Info**. Per-request, per-query, and per-job detail is at **Debug** (opt in per module).
Recoverable issues (slow SQL, job retries) are **Warn**; failures are **Error**.

**HTTP access log** (`access` module): one line per request with fields `method`, `status`,
`latency`, `ip`, `url`, `request_id` (plus `trace_id` when OpenTelemetry is active). The
`/.well-known/{liveness,readiness,health}` URIs are skipped. Status maps to level: 2xx/3xx → Info,
4xx → Warn, 5xx → Error.

**Request-scoped logging.** In a handler, `log := http.WithReqCtx(c)` (alias `http.GetLogger(c)`)
returns the app logger enriched with `request_id` (reused from an upstream `X-Request-ID` or the
configured header) and `trace_id`/`span_id` when a span is active — so every line you log is
correlated to the request and its trace.

To see SQL or MongoDB commands, enable `DB_LOG_MODE` / `MONGO_DEBUG` — they surface under the
`gorm` / `mongo` module tags (e.g. `LOG_MODULE_LEVELS=gorm:debug`).

---

## Database (GORM)

GOE supports multiple database connections. The default connection uses `DB_*` prefix, while named connections use `DB_{NAME}_*` prefix (e.g., `DB_REPLICA_HOST`).

### Connection Management

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `DB_CONNECTION` | string | `default` | Name of the default database connection |
| `DB_CONNECTIONS` | string | - | Comma-separated list of additional connection names |

### Connection Settings

For each connection, use the prefix `DB_` (default) or `DB_{NAME}_` (named connection).

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `DB_DRIVER` | string | **required** | Database driver: `mysql`, `postgres`, `pgsql`, `postgresql`, `sqlite`, `sqlite3`, `sqlserver`, `mssql` |
| `DB_HOST` | string | **required*** | Database host (not required for SQLite) |
| `DB_PORT` | string | driver-specific | Database port (MySQL: 3306, PostgreSQL: 5432, SQLServer: 1433) |
| `DB_DATABASE` | string | **required*** | Database name or SQLite file path |
| `DB_USERNAME` | string | - | Database username |
| `DB_PASSWORD` | string | - | Database password |
| `DB_DSN` | string | - | Direct DSN string (overrides individual settings) |

### Driver-Specific Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `DB_CHARSET` | string | `utf8mb4` | Character set (MySQL) |
| `DB_TIMEZONE` | string | `Local`/`UTC` | Timezone (MySQL uses `Local`, PostgreSQL uses `UTC`) |
| `DB_SSLMODE` | string | `disable` | SSL mode: `disable`, `require`, `verify-ca`, `verify-full` (PostgreSQL) |
| `DB_ENCRYPT` | string | `disable` | Encryption mode (SQL Server) |
| `DB_TRUST_SERVER_CERTIFICATE` | string | - | Trust server certificate (SQL Server) |

### Connection Pool Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `DB_MAX_IDLE_CONNS` | int | `10` | Maximum idle connections in pool |
| `DB_MAX_OPEN_CONNS` | int | `100` | Maximum open connections |
| `DB_CONN_MAX_LIFETIME` | duration | - | Maximum connection lifetime |
| `DB_CONN_MAX_IDLE_TIME` | duration | - | Maximum idle time for connections |

### GORM Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `DB_LOG_MODE` | bool | `false` | Enable GORM SQL logging (queries at Debug, slow queries at Warn, under the `gorm` log module). OR'd with per-connection `DB_{NAME}_LOG_MODE`. |
| `DB_IGNORE_RECORD_NOT_FOUND_ERROR` | bool | `false` | Suppress record not found errors |
| `DB_DISABLE_FOREIGN_KEY_CONSTRAINT_WHEN_MIGRATING` | bool | `false` | Disable FK constraints during migration |

### Auto-Migration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `DB_AUTO_MIGRATE` | bool | `false` | Enable auto-migration for default connection |
| `DB_AUTO_MIGRATE_ANY` | bool | `false` | Enable auto-migration for all connections |
| `DB_{NAME}_AUTO_MIGRATE` | bool | `false` | Enable auto-migration for named connection |

---

## MongoDB

GOE supports multiple MongoDB connections. The default connection uses `MONGO_*` prefix, while named connections use `MONGO_{NAME}_*` prefix.

### Connection Management

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `MONGO_CONNECTION` | string | `default` | Name of the default MongoDB connection |
| `MONGO_CONNECTIONS` | string | - | Comma-separated list of additional connection names |

### Connection Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `MONGO_URI` | string | **required** | MongoDB connection URI (e.g., `mongodb://localhost:27017`) |
| `MONGO_DB_NAME` | string | **required** | Database name |
| `MONGO_MIN_POOL_SIZE` | int | - | Minimum connection pool size |
| `MONGO_MAX_POOL_SIZE` | int | - | Maximum connection pool size |
| `MONGO_MAX_CONN_IDLE_TIME` | duration | - | Maximum idle time for connections |
| `MONGO_DEBUG` | bool | `false` | Log MongoDB commands via a command monitor (under the `mongo` log module); also `MONGO_{NAME}_DEBUG` per named connection |

---

## MongoDB Migrations

MongoDB migrations provide schema versioning, distributed locking, and automatic document-level versioning. Enable with `WithMigrate: true` (requires `WithMongoDB: true`).

### Migration Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `MONGODB_MIGRATE_COLLECTION` | string | `_goe_migrations` | Collection name for tracking migration state |
| `MONGODB_MIGRATE_TIMEOUT` | duration | `5m` | Maximum duration for a single migration execution |
| `MONGODB_MIGRATE_LOCK_TIMEOUT` | duration | `30m` | Maximum duration for distributed migration lock |
| `MONGODB_MIGRATE_LOCK_HEARTBEAT` | duration | `30s` | Interval for lock heartbeat updates |
| `MONGODB_MIGRATE_USE_TRANSACTIONS` | bool | `true` | Use transactions if MongoDB replica set is available |
| `MONGODB_MIGRATE_VERIFY_CHECKSUMS` | bool | `true` | Verify migration checksums on startup to detect modifications |
| `MONGODB_MIGRATE_AUTO` | bool | `false` | Automatically run pending migrations on application start |
| `MONGODB_MIGRATE_VERSION_SCHEME` | string | `sequential` | Version scheme: `sequential` (1, 2, 3...) or `timestamp` |
| `MONGODB_MIGRATE_SCHEMA_VERSION_FIELD` | string | `_goe_sv` | Field name for document-level schema versioning |
| `MONGODB_MIGRATE_DRY_RUN` | bool | `false` | Enable dry-run mode by default (preview without applying) |

---

## Job System (Background Processing)

The job system provides Redis-backed background job processing with scheduling, retries, and dead letter queues. Enable with `WithJob: true`.

### Redis Connection

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `JOB_REDIS_URL` | string | - | Redis connection URL (e.g., `redis://localhost:6379/0`) |
| `JOB_REDIS_HOSTS` | []string | `localhost:6379` | Comma-separated Redis hosts (fallback if URL not provided) |
| `JOB_REDIS_USERNAME` | string | - | Redis username |
| `JOB_REDIS_PASSWORD` | string | - | Redis password |
| `JOB_REDIS_DB` | int | `0` | Redis database number |
| `JOB_REDIS_POOL_SIZE` | int | `10` | Redis connection pool size |
| `JOB_REDIS_ADDR` | string | - | **Deprecated**: Use `JOB_REDIS_URL` or `JOB_REDIS_HOSTS` |

### Worker Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `JOB_CONCURRENCY` | int | `5` | Number of concurrent workers per queue |
| `JOB_MAX_CONCURRENCY` | int | `100` | Maximum total concurrent workers |
| `JOB_POLL_INTERVAL` | duration | `1s` | How often to poll for new jobs |
| `JOB_SHUTDOWN_TIMEOUT` | duration | `30s` | Timeout for graceful shutdown |
| `JOB_HEARTBEAT_INTERVAL` | duration | `15s` | Worker heartbeat interval |
| `JOB_KEY_PREFIX` | string | `goe:job:` | Prefix for all job-related Redis keys |

### Job Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `JOB_DEFAULT_QUEUE` | string | `default` | Default queue name |
| `JOB_DEFAULT_MAX_ATTEMPTS` | int | `3` | Default maximum retry attempts |
| `JOB_DEFAULT_TIMEOUT` | duration | `30m` | Default job execution timeout |
| `JOB_RETRY_BACKOFF` | duration | `1s` | Initial retry backoff duration |
| `JOB_MAX_RETRY_BACKOFF` | duration | `5m` | Maximum retry backoff duration |
| `JOB_RETRY_BACKOFF_FACTOR` | float | `2.0` | Retry backoff multiplier |
| `JOB_DEFAULT_UNIQUE_TTL` | duration | `1h` | Default TTL for job uniqueness |

### Scheduler Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `JOB_SCHEDULER_ENABLED` | bool | `true` | Enable the job scheduler |
| `JOB_SCHEDULER_INTERVAL` | duration | `1s` | How often to check for scheduled jobs |

### Dead Letter Queue

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `JOB_DLQ_ENABLED` | bool | `true` | Enable dead letter queue for failed jobs |
| `JOB_DLQ_TTL` | duration | `168h` | How long to keep jobs in DLQ (default: 7 days) |

### Metrics

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `JOB_METRICS_ENABLED` | bool | `true` | Enable job metrics collection |

---

## Cache

### General Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CACHE_STORE` | string | `memory` | Cache store. **Only `memory` and `redis` are supported**; any other value is rejected at startup. |
| `CACHE_DRIVER` | string | `memory` | Backing driver (`memory` or `redis`); resolved from `CACHE_{STORE}_DRIVER`, then `CACHE_DRIVER`, then `memory` |
| `CACHE_PREFIX` | string | `{APP_NAME}` | Prefix for all cache keys |
| `CACHE_TTL` | duration | `2h` | Default cache TTL |

### Memory Store

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CACHE_MEMORY_GC_INTERVAL` | duration | `10s` | Garbage collection interval |

### Redis Store

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CACHE_REDIS_URL` | string | - | Redis connection URL (overrides individual settings) |
| `CACHE_REDIS_HOST` | string | `127.0.0.1` | Redis host |
| `CACHE_REDIS_PORT` | int | `6379` | Redis port |
| `CACHE_REDIS_USERNAME` | string | - | Redis username (Redis 6+) |
| `CACHE_REDIS_PASSWORD` | string | - | Redis password |
| `CACHE_REDIS_DATABASE` | int | `0` | Redis database number |
| `CACHE_REDIS_CLIENT_NAME` | string | - | Client name for Redis connection |
| `CACHE_REDIS_POOL_SIZE` | int | - | Connection pool size |
| `CACHE_REDIS_RESET` | bool | `false` | Reset (flush) Redis database on startup |

#### Redis Cluster/Sentinel

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CACHE_REDIS_ADDRS` | []string | - | Comma-separated cluster node addresses |
| `CACHE_REDIS_MASTER_NAME` | string | - | Sentinel master name |
| `CACHE_REDIS_IS_CLUSTER_MODE` | bool | `false` | Enable Redis Cluster mode |

### Other store types

> Only `memory` and `redis` have registered drivers. Any other `CACHE_STORE` value
> (`postgres`, `mysql`, `memcache`, `mongodb`, `dynamodb`, `s3`, `badger`, `sqlite3`) is
> **rejected at startup** by config validation.

---

## Lock System (Distributed Mutex)

The lock system provides distributed locking using Redis with support for single instance, Sentinel, Cluster, and Redlock algorithms.

### Redis Connection

The connection URL scheme determines the mode:
- `redis://` - Single Redis instance
- `rediss://` - Single Redis with TLS
- `redis-sentinel://` - Redis Sentinel
- `redis-cluster://` - Redis Cluster

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `LOCK_REDIS_URL` | string | `redis://localhost:6379/0` | Primary Redis connection URL (database is taken from the URL path) |
| `LOCK_REDIS_URLS` | []string | - | Multiple URLs for the Redlock algorithm |

> The lock connection is configured **only** via `LOCK_REDIS_URL` (or `LOCK_REDIS_URLS` for
> Redlock), with the pool size from `LOCK_POOL_SIZE`. `LOCK_REDIS_ADDR`, `LOCK_REDIS_HOST`,
> `LOCK_REDIS_HOSTS`, `LOCK_REDIS_DB`, and `LOCK_REDIS_POOL_SIZE` are **not used** (and no longer
> validated) — setting only one of them is rejected, so use the URL form.

### Lock Defaults

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `LOCK_DEFAULT_EXPIRY` | duration | `8s` | Default lock expiration time |
| `LOCK_DEFAULT_TRIES` | int | `32` | Default number of acquisition attempts |
| `LOCK_DEFAULT_RETRY_DELAY` | duration | `500ms` | Default delay between retries |
| `LOCK_DEFAULT_DRIFT_FACTOR` | float | `0.01` | Clock drift factor for Redlock |
| `LOCK_KEY_PREFIX` | string | `lock:` | Prefix for all lock keys |
| `LOCK_POOL_SIZE` | int | `10` | Redis connection pool size |

### TLS Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `LOCK_TLS_INSECURE_SKIP_VERIFY` | bool | `false` | Skip TLS certificate verification (not recommended) |

---

## Observability — OpenTelemetry

Distributed tracing (and optional metric export) via OpenTelemetry. Enable with `WithOTel: true`.

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `OTEL_ENABLED` | bool | `true` | Master switch for OpenTelemetry (when `WithOTel` is set) |
| `OTEL_SERVICE_NAME` | string | `APP_NAME` → `goe-app` | Service name reported on traces |
| `OTEL_SERVICE_VERSION` | string | `APP_VERSION` → `1.0.0` | Service version reported on traces |
| `OTEL_TRACES_ENABLED` | bool | `true` | Enable trace generation |
| `OTEL_METRICS_ENABLED` | bool | `false` | Enable OTel metric export (Prometheus is used by default) |
| `OTEL_TRACES_SAMPLER` | string | `parentbased_traceidratio` | Trace sampler strategy |
| `OTEL_TRACES_SAMPLER_ARG` | float | `0.1` | Sampler ratio (10%) for ratio-based samplers |
| `OTEL_EXPORTER_TYPE` | string | `otlp` | Exporter: `otlp`, `stdout`, or `none` (falls back to `OTEL_EXPORTER`) |
| `OTEL_EXPORTER` | string | `otlp` | Legacy fallback for `OTEL_EXPORTER_TYPE` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | string | `localhost:4317` | OTLP collector endpoint |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | string | `grpc` | OTLP protocol: `grpc` or `http` |
| `OTEL_EXPORTER_OTLP_INSECURE` | bool | `true` | Disable TLS for the OTLP exporter (intended for local dev) |
| `OTEL_EXPORTER_OTLP_HEADERS` | string | - | Comma-separated `key=value` headers for OTLP requests |
| `OTEL_PROPAGATORS` | string | `tracecontext,baggage` | Comma-separated context propagators |
| `OTEL_RESOURCE_ATTRIBUTES` | string | - | Comma-separated `key=value` resource attributes |

---

## Metrics (Prometheus)

Prometheus metrics endpoint and collectors. Enable with `WithMetrics: true`.

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `METRICS_ENABLED` | bool | `true` | Enable metrics collection (when `WithMetrics` is set) |
| `METRICS_PATH` | string | `/metrics` | HTTP path for the Prometheus endpoint |
| `METRICS_NAMESPACE` | string | `goe` | Prometheus namespace (metric name prefix) |
| `METRICS_SUBSYSTEM` | string | - | Optional Prometheus subsystem |
| `METRICS_GO_ENABLED` | bool | `true` | Collect Go runtime metrics |
| `METRICS_PROCESS_ENABLED` | bool | `true` | Collect process metrics |

---

## Health Checks

Liveness/readiness endpoints. Enable with `WithHealth: true`.

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `HEALTH_ENABLED` | bool | `true` | Enable health checks (when `WithHealth` is set) |
| `HEALTH_PATH` | string | `/health` | Base path for the health endpoint |
| `HEALTH_LIVENESS_PATH` | string | `/health/live` | Liveness probe path |
| `HEALTH_READINESS_PATH` | string | `/health/ready` | Readiness probe path |
| `HEALTH_TIMEOUT` | duration | `5s` | Per-check timeout |

---

## Graceful Shutdown

Timeouts applied when the application shuts down (SIGINT/SIGTERM).

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `SHUTDOWN_TIMEOUT` | duration | `30s` | Total graceful-shutdown timeout |
| `SHUTDOWN_DRAIN_TIMEOUT` | duration | `5s` | HTTP connection drain timeout |

---

## Duration Format

Duration values accept Go duration format:
- `s` - seconds (e.g., `30s`)
- `m` - minutes (e.g., `5m`)
- `h` - hours (e.g., `24h`)
- `ms` - milliseconds (e.g., `500ms`)

Examples: `30s`, `5m`, `24h`, `500ms`, `1h30m`

## Boolean Format

Boolean values accept: `true`, `false`, `1`, `0`, `yes`, `no`

---

## Named Connection Pattern

For modules supporting multiple connections (DB, MongoDB), named connections use the pattern:

```bash
# Default connection
DB_DRIVER=postgres
DB_HOST=localhost

# Named connection "replica"
DB_REPLICA_DRIVER=postgres
DB_REPLICA_HOST=replica.example.com

# Register the named connection
DB_CONNECTIONS=replica
```

---

## Best Practices

1. **Use `.local.env` for local overrides** - This file should be gitignored
2. **Use environment-specific files** - Create `.prod.env`, `.staging.env` for different environments
3. **Use `GOE_ENV`** - Set this to control which environment file is loaded
4. **Secrets in system env** - Store sensitive values in system environment variables, not files
5. **Validate in development** - Leave `DISABLE_CONFIG_VALIDATION=false` during development
6. **Use connection URLs** - When available, use URL format (`*_URL`) for cleaner configuration
