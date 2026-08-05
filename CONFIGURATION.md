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

### Request validation

GOE installs a `fiber.StructValidator` on the app, so `validate` struct tags are
enforced by `Ctx.Bind` with no setup:

```go
type Person struct {
    Name string `json:"name" validate:"required"`
    Age  int    `json:"age"  validate:"gte=18,lte=60"`
}

app.Post("/people", func(c fiber.Ctx) error {
    p := new(Person)
    if err := c.Bind().JSON(p); err != nil {
        return err // parsed AND validated
    }
    return c.JSON(p)
})
```

This matters because Fiber ships no validator of its own — it defines the
one-method `StructValidator` interface and calls it from every binding, but with
the field unset `Bind` parses the body and skips validation *silently*. GOE fills
it in. See [Fiber's validation guide](https://docs.gofiber.io/guide/validation).
`Bind` is the only validation path — there is no separate validator service to
inject or call.

`Bind` covers `JSON`, `Query`, `URI`, `Form`, `Header`, `Cookie`, `XML`, `CBOR`
and `MsgPack`. It only validates struct destinations; binding into a map skips the
validator. Use `c.Bind().SkipValidation()` to parse without validating.

**What the bundled validator gives you:**

- **json field names.** Failures are reported under the json tag (`email`, not
  `Email`), because that is the name the client sent.
- **Bundled rules** beyond go-playground's built-ins: `phone` (10-15 chars,
  digits with `+`, `-`, spaces), `username` (3-30 chars, alphanumeric and
  underscore) and `strong_password` (8+ chars with upper, lower, digit and
  special). Re-register a tag via `WithValidatorSetup` to replace its rule.
- **Client errors render as client errors.** `return err` after a failed
  `Bind` produces a **400** whose message *is* the failed rule's message —
  one error at a time, in field declaration order, ready to show to a user.
  The same message feeds the error page for browsers and `format=text`.
  Malformed bodies and unconvertible parameters (Fiber's `*BindError`) are
  400s too. Submitted values are never echoed back, so failed password rules
  do not leak secrets.

```json
{
  "message": "age must be greater than or equal to 18"
}
```

To render failures differently — every field at once, a custom envelope —
catch the typed error; `Fields` carries field, tag, param and message for
each failed rule:

```go
if err := c.Bind().JSON(&req); err != nil {
    var ve *goehttp.ValidationError
    if errors.As(err, &ve) {
        return c.Status(fiber.StatusBadRequest).JSON(myErrorShape(ve.Fields))
    }
    return err
}
```

**Adding a rule** — when the defaults are fine but you need one more:

```go
goe.New(goe.Options{
    HTTP: []goehttp.Option{
        goehttp.WithValidatorSetup(func(v *validator.Validate) error {
            return v.RegisterValidation("slug", isSlug)
        }),
    },
})
```

Setups run in order during kernel construction, so rules exist before the first
request. Returning an error aborts startup.

**Replacing it entirely** — a different library, or custom behaviour:

```go
goehttp.WithStructValidator(myValidator) // implements Validate(any) error
```

Combining the two is rejected at startup, since a setup for the bundled validator
cannot apply to a replacement.

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
| `FIBER_PROXY_HEADER` | string | `X-Forwarded-For` | Header used to obtain client IP; only consulted when trusted proxies are enabled |
| `FIBER_REDUCE_MEMORY` | bool | `false` | Reduce memory usage at the cost of performance |
| `FIBER_ENABLE_IP_VALIDATION` | bool | `true` | Validate proxy-header IPs and walk past trusted hops instead of returning the raw header |

### Trust Proxy Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `FIBER_TRUST_PROXY` | bool | `false` | Enable trusted proxy mode |
| `FIBER_TRUST_PROXIES` | []string | - | Comma-separated list of trusted proxy IPs/CIDRs |
| `FIBER_TRUST_LINK_LOCAL` | bool | `true` | Trust link-local addresses when `FIBER_TRUST_PROXY` is enabled |
| `FIBER_TRUST_LOOPBACK` | bool | `true` | Trust loopback addresses when `FIBER_TRUST_PROXY` is enabled |
| `FIBER_TRUST_PRIVATE` | bool | `true` | Trust private network addresses when `FIBER_TRUST_PROXY` is enabled |

### Trusting a CDN's edge IPs

Behind a CDN, every request arrives from an edge node, so the client address is in
a forwarded header rather than the socket. Fiber only honours that header for hops
you have declared trustworthy — otherwise anyone could forge it. The optional
`cdntrust` package fetches those ranges from each provider's published source:

```go
import "go.oease.dev/goe/v2/cdntrust"

ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
defer cancel()

cdnOpts, err := cdntrust.Options(ctx, cdntrust.Cloudflare, cdntrust.Fastly)
if err != nil {
    log.Fatalf("cdn trust: %v", err) // do not start with an unknown trust set
}

goe.New(goe.Options{
    HTTP: append(cdnOpts,
        goehttp.WithPort(8080),
        goehttp.WithProxyHeader("CF-Connecting-IP"), // Cloudflare's client header
    ),
})
```

Supported providers: `cdntrust.Cloudflare`, `cdntrust.Fastly`, `cdntrust.BunnyCDN`
(Cloudflare and Fastly publish CIDR blocks; Bunny publishes several hundred
individual edge addresses, which Fiber accepts equally).

**It fetches once, at boot, and can stop the app from starting.** There is no
cache, no retry and no background refresh — one request per provider. A network
failure is returned as an error, and the expected response is to abort startup:
booting with an unknown trust set means either ignoring the forwarded header
entirely or trusting the wrong hops. Because the list is a point-in-time snapshot,
redeploy periodically to pick up provider changes. If you would rather tolerate a
failed fetch, pin the ranges in config and use `WithTrustProxyConfig` directly.

The package is genuinely optional: nothing in GOE imports it, so neither the code
nor the network calls exist in your binary unless you use it. It adds no
dependency beyond what GOE already requires.

Only the CDN ranges are trusted. `LinkLocal`, `Loopback`, `Private` and
`UnixSocket` are deliberately left off — trusting a private range alongside the
CDN would let anything inside your network forge a client address. Set those
fields yourself if your topology needs them.

See [examples/07-cdn-trusted-proxy](examples/07-cdn-trusted-proxy/).

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

**Module defaults.** The `fx` module (Uber Fx dependency injection) defaults to
`warn` rather than following `LOG_LEVEL`. Fx logs one line per constructor
supplied, provided, decorated and run, plus every lifecycle hook — dozens of
lines of graph detail that previously buried your own output whenever
`LOG_LEVEL=debug` was set. Provide and invoke *failures* are still logged. Set
`LOG_MODULE_LEVELS=fx:debug` to see the wiring when diagnosing injection itself.

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

GOE manages a single MongoDB connection. Enable with `WithMongoDB: true` or by passing MongoDB options. Applications needing a second data source construct their own `mongo.Client` via the driver.

### Configuring in Go code

Every `MONGO_*` variable below — except the URI and credentials — has a code
equivalent. Pass options through `goe.Options.MongoDB` (which also enables the
module) and they take precedence over the environment:

```go
import goemongo "go.oease.dev/goe/v2/core/mongodb"

goe.New(goe.Options{
    MongoDB: []goemongo.Option{
        goemongo.WithDatabase("myapp"),
        goemongo.WithMaxPoolSize(50),
    },
})
```

The naming rule is mechanical: the environment key with the `MONGO_` prefix
dropped (`MONGO_MAX_POOL_SIZE` → `WithMaxPoolSize`). `WithCommandMonitor`
installs a custom `*event.CommandMonitor` — a live object with no environment
equivalent — and wins over the `MONGO_DEBUG` monitor. An invalid option
discards every option and fails startup validation with all problems listed
at once.

**Credentials stay in the environment.** `MONGO_URI` (which can embed
`user:pass`), `MONGO_USERNAME` and `MONGO_PASSWORD` have no option — secrets
never belong in source. They compose with code: options tune databases and
pools while the environment supplies the URI and its credentials.

### Connection Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `MONGO_URI` | string | **required** | MongoDB connection URI (e.g., `mongodb://localhost:27017`). Environment-only |
| `MONGO_DB_NAME` | string | **required** | Database name |
| `MONGO_USERNAME` | string | - | Username. Environment-only; applies to a URI that embeds no userinfo |
| `MONGO_PASSWORD` | string | - | Password. Environment-only; applies to a URI that embeds no userinfo |
| `MONGO_MIN_POOL_SIZE` | int | - | Minimum connection pool size |
| `MONGO_MAX_POOL_SIZE` | int | - | Maximum connection pool size |
| `MONGO_MAX_CONN_IDLE_TIME` | duration | - | Maximum idle time for connections |
| `MONGO_PING_TIMEOUT` | duration | `5s` | Startup reachability ping timeout |
| `MONGO_DEBUG` | bool | `false` | Log MongoDB commands via a command monitor (under the `mongo` log module) |

> Credentials embedded in `MONGO_URI` win over `MONGO_USERNAME`/`MONGO_PASSWORD`; the separate
> variables fill in whenever the URI carries none.
>
> **Behaviour change:** startup now verifies the connection with a ping (bounded by
> `MONGO_PING_TIMEOUT`) and **fails fast** when MongoDB is unreachable. Earlier releases logged
> the failure and continued with a nil database, deferring the crash to the first
> `Col()`/`DB()` call inside a handler.

---

## MongoDB Migrations

MongoDB migrations provide schema versioning, distributed locking, and automatic document-level versioning. Enable with `WithMigrate: true` (requires `WithMongoDB: true`), or by passing migrate options — which imply both modules.

### Configuring in Go code

Every `MONGODB_MIGRATE_*` variable below has a code equivalent. Pass options
through `goe.Options.Migrate` (which also enables MongoDB and migrations) and
they take precedence over the environment:

```go
import "go.oease.dev/goe/v2/core/mongodb/migrate"

goe.New(goe.Options{
    Migrate: []migrate.Option{
        migrate.WithAutoMigrate(true),
        migrate.WithCollection("_migrations"),
        migrate.WithVersionScheme("timestamp"),
    },
})
```

The naming rule is mechanical: every field of `migrate.Config` is exposed as
`With<FieldName>` (`MONGODB_MIGRATE_LOCK_TIMEOUT` → `WithLockTimeout`,
`MONGODB_MIGRATE_DRY_RUN` → `WithDryRunByDefault`). Options validate strictly
— `WithVersionScheme` accepts only `sequential` or `timestamp` — and an
invalid option aborts startup inside `goe.New` with every problem listed at
once. These are module options; the lower-level `MigratorOption` family
(`WithLogger`, `WithConfig`, `WithDryRun`, `WithHostname`) still configures a
hand-built `Migrator` as before.

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

The job system provides Redis-backed background job processing with scheduling, retries, and dead letter queues. Enable with `WithJob: true` or by passing job options.

### Configuring in Go code

Every `JOB_*` variable below — except the connection URL and credentials — has a
code equivalent. Pass options through `goe.Options.Job` (which also enables the
module) and they take precedence over the environment; anything left unset keeps
its environment value:

```go
import goejob "go.oease.dev/goe/v2/core/job"

goe.New(goe.Options{
    Job: []goejob.Option{
        goejob.WithRedisHosts("redis.internal:6379"),
        goejob.WithConcurrency(10),
        goejob.WithDefaultQueue("critical"),
        goejob.WithDLQTTL(48 * time.Hour),
    },
})
```

The naming rule is mechanical: every field of `job.Config` is exposed as
`With<FieldName>` (`JOB_MAX_CONCURRENCY` → `WithMaxConcurrency`). Options apply
in the order given; an invalid option aborts startup inside `goe.New` with every
problem listed at once, before any Redis connection is attempted.

**Credentials stay in the environment.** `JOB_REDIS_URL` (which can embed
`user:pass`), `JOB_REDIS_USERNAME` and `JOB_REDIS_PASSWORD` have no option on
purpose — secrets never belong in source. They compose with code: an endpoint
chosen with `WithRedisHosts` still authenticates with the environment's
username and password.

### Redis Connection

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `JOB_REDIS_URL` | string | - | Redis connection URL (e.g., `redis://localhost:6379/0`). Environment-only; takes priority over hosts |
| `JOB_REDIS_HOSTS` | []string | `localhost:6379` | Comma-separated Redis hosts (fallback if URL not provided) |
| `JOB_REDIS_USERNAME` | string | - | Redis username. Environment-only; applies to hosts **and** to a URL that embeds no userinfo |
| `JOB_REDIS_PASSWORD` | string | - | Redis password. Environment-only; applies to hosts **and** to a URL that embeds no userinfo |
| `JOB_REDIS_DB` | int | `0` | Redis database number |
| `JOB_REDIS_POOL_SIZE` | int | `10` | Redis connection pool size |
| `JOB_REDIS_ADDR` | string | - | **Deprecated**: Use `JOB_REDIS_URL` or `JOB_REDIS_HOSTS` |

> Credentials embedded in `JOB_REDIS_URL` win over `JOB_REDIS_USERNAME`/`JOB_REDIS_PASSWORD`.
> A credential-free URL plus the separate credential variables now works — earlier releases
> silently ignored the separate variables whenever a URL was set.

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

### Configuring in Go code

Every `CACHE_*` variable below — except the connection URL and credentials —
has a code equivalent. Pass options through `goe.Options.Cache` (which also
enables the module) and they take precedence over the environment:

```go
import goecache "go.oease.dev/goe/v2/core/cache"

goe.New(goe.Options{
    Cache: []goecache.Option{
        goecache.WithStore("redis"),
        goecache.WithTTL(30 * time.Minute),
        goecache.WithRedisHost("redis.internal"),
    },
})
```

The naming rule is mechanical: the environment key with the `CACHE_` prefix
dropped (`CACHE_REDIS_POOL_SIZE` → `WithRedisPoolSize`). Store-scoped keys take
the store name as their first argument: `WithStoreDriver("sessions", "redis")`
is `CACHE_sessions_DRIVER`, and `WithStorePrefix`/`WithStoreTTL` follow suit.
Under the hood options become a configuration overlay, so custom drivers
registered through `Extend` read code-configured values exactly as they read
environment variables. An invalid option discards every option and fails
startup validation with all problems listed at once.

**Credentials stay in the environment.** `CACHE_REDIS_URL` (which can embed
`user:pass`), `CACHE_REDIS_USERNAME` and `CACHE_REDIS_PASSWORD` have no option
on purpose — secrets never belong in source. They compose with code: an
endpoint chosen with `WithRedisHost` still authenticates with the environment's
username and password.

### General Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CACHE_STORE` | string | `memory` | Cache store. A store named after a registered driver (`memory`, `redis`, or a custom driver added via `Extend`) uses that driver directly; any other name must have a driver configured, or startup is rejected. |
| `CACHE_DRIVER` | string | `memory` | Backing driver; resolved from `CACHE_{STORE}_DRIVER`, then `CACHE_DRIVER`, then the store name itself if it is a registered driver, then `memory` |
| `CACHE_PREFIX` | string | `{APP_NAME}` | Prefix for all cache keys |
| `CACHE_TTL` | duration | `2h` | Default cache TTL |

> **Behavior fix:** `CACHE_STORE=redis` alone now uses the redis driver. Earlier
> releases required `CACHE_DRIVER=redis` alongside it and silently fell back to
> the memory driver otherwise. An explicitly configured driver still wins.

### Memory Store

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CACHE_MEMORY_GC_INTERVAL` | duration | `10s` | Garbage collection interval |

### Redis Store

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CACHE_REDIS_URL` | string | - | Redis connection URL (overrides individual settings). Environment-only |
| `CACHE_REDIS_HOST` | string | `127.0.0.1` | Redis host |
| `CACHE_REDIS_PORT` | int | `6379` | Redis port |
| `CACHE_REDIS_USERNAME` | string | - | Redis username (Redis 6+). Environment-only; applies to host/port **and** to a URL that embeds no userinfo |
| `CACHE_REDIS_PASSWORD` | string | - | Redis password. Environment-only; applies to host/port **and** to a URL that embeds no userinfo |
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

> Credentials embedded in `CACHE_REDIS_URL` win over `CACHE_REDIS_USERNAME`/`CACHE_REDIS_PASSWORD`.
> A credential-free URL plus the separate credential variables now works — the credentials are
> injected into the URL (scheme, database and query parameters untouched); earlier releases
> silently ignored the separate variables whenever a URL was set.

### Other store types

> Only `memory` and `redis` ship as built-in drivers. A `CACHE_STORE` that neither names a
> registered driver nor has one configured via `CACHE_{STORE}_DRIVER`/`CACHE_DRIVER` is
> **rejected at startup** by config validation. Custom drivers registered through
> `goe.Cache().Extend(name, factory)` count as registered — a store may name one directly.

---

## Lock System (Distributed Mutex)

The lock system provides distributed locking using Redis with support for single instance, Sentinel, Cluster, and Redlock algorithms. Enable with `WithLock: true` or by passing lock options.

### Configuring in Go code

Every `LOCK_*` variable below — except the credentials — has a code
equivalent. Pass options through `goe.Options.Lock` (which also enables the
module) and they take precedence over the environment:

```go
import goelock "go.oease.dev/goe/v2/core/lock"

goe.New(goe.Options{
    Lock: []goelock.Option{
        goelock.WithRedisURL("redis-sentinel://mymaster@s1:26379,s2:26379/0"),
        goelock.WithDefaultExpiry(10 * time.Second),
        goelock.WithKeyPrefix("myapp:lock:"),
    },
})
```

The naming rule is mechanical: every field of `lock.Config` is exposed as
`With<FieldName>` (`LOCK_DEFAULT_EXPIRY` → `WithDefaultExpiry`). An invalid
option aborts startup inside `goe.New` with every problem listed at once,
before any Redis connection is attempted.

**Credentials stay in the environment.** `WithRedisURL` and `WithRedisURLs`
**reject** URLs that embed `user:pass` — the URL in code declares the topology,
and `LOCK_REDIS_USERNAME`/`LOCK_REDIS_PASSWORD` from the environment supply the
secret, whatever the mode (single, sentinel, cluster, redlock).

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
| `LOCK_REDIS_USERNAME` | string | - | Redis username. Environment-only; applies to any URL that embeds no userinfo |
| `LOCK_REDIS_PASSWORD` | string | - | Redis password. Environment-only; applies to any URL that embeds no userinfo |

> The lock connection is configured **only** via `LOCK_REDIS_URL` (or `LOCK_REDIS_URLS` for
> Redlock), with the pool size from `LOCK_POOL_SIZE`. `LOCK_REDIS_ADDR`, `LOCK_REDIS_HOST`,
> `LOCK_REDIS_HOSTS`, `LOCK_REDIS_DB`, and `LOCK_REDIS_POOL_SIZE` are **not used** (and no longer
> validated) — setting only one of them is rejected, so use the URL form.
>
> Credentials embedded in the URL win over `LOCK_REDIS_USERNAME`/`LOCK_REDIS_PASSWORD`; the
> separate variables fill in whenever the URL carries none, in every connection mode.

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

For the SQL DB module, which supports multiple connections, named connections use the pattern:

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
