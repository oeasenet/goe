# Migration Guide

Breaking changes, and behaviour changes that need no code edit but will alter what
your application does. Releases with neither are not listed here.

---

## v2.8.0

No public API is removed. This release makes graceful shutdown do what its
configuration says, gives long-lived responses (SSE, proxied LLM streams) helpers
that survive the write timeout and end cleanly at shutdown, and puts a guard in
front of your routes. It also raises the minimum Go version to 1.27 and moves
every dependency to its latest release — Fiber v3.6.1 and fasthttp v1.77.0 among
them — closing several security holes. The behaviour changes below need no code
edit, but check the upgrade checklist first.

### Upgrade checklist

1. **Build with Go 1.27 or later.** GOE's `go.mod` now says `go 1.27.0`: update
   builder images (`golang:1.27`) and CI. With `GOTOOLCHAIN=local` an older Go
   fails with "go.mod requires go >= 1.27.0".
2. **Set `SHUTDOWN_TIMEOUT` per service, below its stop grace period** (Compose
   `stop_grace_period` defaults to 10s). The stop now lasts up to
   `SHUTDOWN_TIMEOUT` (30s by default; Fx's own 15s applied before), so an unset
   value is longer than most grace periods.
3. **Move pre-drain work to a shutdown hook above `shutdown.PriorityHTTP`.** Fx
   hooks appended by your Providers and Invokers now stop *after* the HTTP drain
   (they stopped while the server still served). Work that must happen while
   requests are still accepted — deregistration, ending runs that hold open
   streams — belongs in `RegisterHook(name, shutdown.PriorityHTTP+10, fn)`.
4. **Replace workarounds:** self-SIGTERM → `goe.Shutdown(code)`; a custom run
   loop to escape Fx's 15s stop limit → `goe.Options.ShutdownTimeout`; untracked
   goroutines that should finish → `goe.Go(name, fn)`; a global multi-minute
   `HTTP_WRITE_TIMEOUT` for streams → `goehttp.SSE` / `goehttp.SendStream`;
   hand-rolled metrics servers → `metrics.WithPort`; string-matching
   "key already exists" → `errors.Is(err, cache.ErrKeyExists)`.
5. **Proxies:** if a handler routes or authorizes on `c.Path()` and forwards
   `c.OriginalURL()`, forward `c.Path()` plus the query instead. Fiber v3.6
   normalizes the former and not the latter; GOE now refuses dot-segment paths,
   which closes the known way to exploit the mismatch.
6. **Uploads:** `FIBER_BODY_LIMIT` does nothing while bodies stream (GOE's
   default). Turn on `HTTP_STRICT_BODY_LIMIT` or use `goehttp.WithRouteLimits`.

### Behaviour changes

| Change | Action needed |
|---|---|
| GOE requires Go 1.27 | Update builders and CI (checklist 1) |
| Shutdown hooks registered with `shutdown.Manager.RegisterHook` now run when the app stops — the manager was never invoked, so they never ran — highest priority first, equal priorities in registration order | Review your hooks: they execute for the first time |
| Stop order: `goe.Options.OnStop` hooks, then the shutdown manager's phases (hooks above `PriorityHTTP`; the HTTP drain; hooks below it, with the wait for `goe.Go` tasks at `PriorityWorkers`), then Fx teardown (Provider/Invoker hooks, custom modules, built-in modules). The server starts after Provider/Invoker hooks and custom modules, and drains before they stop | Checklist 3 |
| `SHUTDOWN_TIMEOUT` (30s) bounds the whole stop and is Fx's stop timeout; the manager's phases get three quarters of it, the rest is kept for teardown. `SHUTDOWN_DRAIN_TIMEOUT` (5s) bounds the drain within that | Checklist 2. A drain timeout above three quarters of the shutdown timeout is capped, with a startup warning |
| A drain that runs out of time closes the connections still open and logs one warning, `HTTP drain incomplete`; a hook that runs out of budget logs `Shutdown hook ran out of time`. Both exit 0. Before, the drain was unbounded and Fx's 15s limit ended the stop with exit code 1 | None |
| Responses written during the drain carry `Connection: close`; readiness answers 503 once shutdown starts | None |
| A request whose raw path has a `.` or `..` segment (percent-encoded too) is answered 400 | None for browsers. `HTTP_ALLOW_DOT_SEGMENTS=true` restores routing them |
| The access log redacts the values of sensitive query parameters (`code`, `token`, `password`, … — see [Access log](CONFIGURATION.md#access-log)), adds `error` to lines with status ≥ 500, and logs requests whose handler panicked as 500 (they were missing); the panic is logged at Error with its stack | Dashboards keyed on the full query see `REDACTED` values. `HTTP_ACCESS_LOG_REDACT_QUERY=` (empty) turns redaction off |
| MongoDB connects and pings while `goe.New` builds the app, so `DB()` works in Providers and Invokers; an unreachable database ends the process in `goe.New` ("Failed to connect to MongoDB") instead of at start | None. `goe.New` may take up to `MONGO_PING_TIMEOUT` when the database is slow |
| `LOG_OUTPUT=stdout` and `stderr` work (they silently dropped every line); an unknown value logs to the console with a warning | None |
| A config value that does not parse — `SHUTDOWN_TIMEOUT=30` without a unit — still falls back to the default, now with a warning | Fix the values the warnings name |
| `.env` values in double quotes may span lines and understand `\n`, `\r`, `\t`, `\"`, `\\`; single-quoted values may span lines; a `# comment` after the closing quote and a space is no longer part of the value. A value whose quotes do not enclose it — never closed, or with more text after the closing quote, such as `"p@ss"word"` — is read as before, with the warning `Env file value has unbalanced quotes` naming the key | A double-quoted value that contains a backslash sequence by accident (a Windows path such as `"C:\new"`) changes — single-quote it. Fix the values the warning names |
| `contract.ShutdownManager` gains `Stopping()` and `Go(name, fn)`; `contract.HealthManager` gains `RegisterLivenessChecker` | Only third-party implementations of these interfaces must add the methods |
| fasthttp v1.77: the default file server, used by Fiber's static middleware, refuses symlinks that leave its root and absolute symlinks | Copy files instead of linking them, or serve that directory yourself |
| fasthttp v1.76: bytes after a multipart body are drained; a request with both `Content-Length` and `Transfer-Encoding` closes the connection; IPv6 clients count against `MaxConnsPerIP` | None |
| Fiber's default `ReadBufferSize` and `WriteBufferSize` double to 8 KB. The read buffer caps request-line plus header size, so requests with 4–8 KB of headers are now served instead of answered with 431 | None for most apps. Pin the previous limit with `http.WithReadBufferSize(4096)` and `http.WithWriteBufferSize(4096)` |
| Request paths are normalized before routing (RFC 3986): escapes of unreserved characters are decoded and `/./` and `/../` resolved, so `/api/./users` matches `/api/users`. `FIBER_UNESCAPE_PATH` additionally decodes reserved escapes such as `%2F` | None, unless a route or middleware depended on a raw path (and see checklist 5) |
| Requests the HTTP parser rejects — oversized headers (431), malformed (400), read timeout (408), unknown transfer coding (501), and with `FIBER_STREAM_REQUEST_BODY=false` a body over the limit (413) — reach the access log and metrics with their real status. They were logged at info level as 200 | None. Dashboards counting these requests see them under their real status |
| `http.WithBodyLimit` rejects zero and negative values at startup. Fiber replaced them with its 4 MB default, so they never disabled the limit as documented | Pass a positive byte count |
| `http.WithGETOnly` admits HEAD and answers other methods with 405 and `Allow: GET, HEAD` | None |
| Fiber v3.6.1: with `PassLocalsToContext` (GOE's default) the session middleware no longer hands a pooled middleware to a retained context; client `Debug()` now logs raw requests and responses | None |
| The Postgres DSN built from `DB_HOST`, `DB_USERNAME`, `DB_PASSWORD`, … now quotes every value. pgx v5.11 parses DSNs like libpq, so a password with a backslash lost it, and a space or trailing backslash broke the connection | None. A hand-written `DB_DSN` in keyword/value form must quote values containing backslashes or spaces; in URL form, `+` is now literal and a repeated parameter's last value wins |
| pgx v5.11 returns `timestamptz` values read in text format — simple protocol, as behind PgBouncer — in `time.Local` rather than the server's offset | None for the instant itself; JSON that renders the offset changes |
| mongo-driver v2.8.2 → v2.9.2 requires MongoDB server 4.4 or newer, and a client `timeoutMS` now also bounds `CommitTransaction`/`AbortTransaction` | Upgrade MongoDB servers older than 4.4 first: the startup ping fails against them |
| go-redis v9.23 gives each client a separate pool for pipelines, up to 10 extra connections opened on demand (the job module pipelines dead-letter moves) | Size Redis `maxclients` for it, or append `pipeline_pool_size=-1` to `JOB_REDIS_URL` |
| A Redis URL whose `skip_verify` is not a boolean fails startup instead of being read as false; `CACHE_REDIS_URL` errors no longer surface as a raw panic from the storage package | Use `true`/`false` |
| fasthttp v1.75 drains a request body the handler left unread, up to 256 KiB, and closes the connection above that; it rejects HTTP/1.0 requests carrying `Transfer-Encoding` and an out-of-range `Content-Length` with 400 | None |

### New

- **Lifecycle:** `goe.Stopping()`, `goe.Go(name, fn)`, `goe.Shutdown(code)`,
  `goe.Options.StartTimeout` / `STARTUP_TIMEOUT`, liveness checkers
  (`RegisterLivenessChecker`). See [Graceful Shutdown](CONFIGURATION.md#graceful-shutdown).
- **Streams:** `goehttp.SSE`, `goehttp.SendStream`. See
  [Streaming responses](CONFIGURATION.md#streaming-responses-sse).
- **Request guard:** `HTTP_STRICT_BODY_LIMIT`, `goehttp.WithRouteLimits`,
  `HTTP_ALLOW_DOT_SEGMENTS`. See [Request guard](CONFIGURATION.md#request-guard).
- **Metrics:** `goe.Options.Metrics` with `metrics.WithPort` / `METRICS_PORT`,
  `metrics.WithRegistry`, `metrics.WithPprof` / `METRICS_PPROF`, build info and
  stream metrics.
- **cdntrust:** retries, and `cdntrust.Load` with a last-known-good snapshot.
- **Smaller:** `WithAccessLogRedactQuery`, `WithAccessLogURL`,
  `WithAccessLogSkip`, `WithRequestIDTrustInbound`, `goehttp.NewValidator`,
  `cache.ErrKeyExists`, `goe.IsProduction`, `spa.Config.Exclude` and
  `NotFoundForMissingFiles`.

### Security fixes in this release

- **Process crash from one request (GHSA-fpmh-frq8-7rj8).** A zstd-encoded
  request body of about 27 KB, sent to any handler that reads the body or calls
  `Bind`, killed the process from a decoder goroutine that `recover` cannot
  reach. Fixed by klauspost/compress v1.20.2+ (v1.20.3 here).
- **Request smuggling with streamed request bodies.** GOE enables
  `StreamRequestBody` by default; before fasthttp v1.75, body bytes a handler
  left unread were parsed as the next request on the same connection.
- **Dot-segment paths through proxies.** Fiber v3.6 normalizes `c.Path()` but
  not `c.OriginalURL()`; a proxy checking one and forwarding the other could be
  steered onto routes it never meant to expose. GOE answers such paths 400.
- **Unbounded uploads.** With streamed request bodies, multipart forms were
  parsed to temporary files without any size limit before a handler could
  refuse them. `HTTP_STRICT_BODY_LIMIT` bounds them (opt-in).
- **Process crash in the JSON decoder.** sonic v1.15.4 fixes an amd64 JIT
  decoder crash when a request fills a fixed-size array field whose elements
  hold pointers, such as `[4]string`; v1.15.5 restores a register after native
  calls.
- **SPA middleware read files from sibling directories.** With `Root: "./dist"`,
  `GET /../dist-private/secret.txt` was served from `./dist-private`, or any
  sibling whose name starts with the root's. Fiber v3.6's path normalization
  now blocks the request, and `spa` checks containment itself, which also
  covers Windows backslash paths.
- **fasthttp v1.76/v1.77 hardening:** outgoing framing validated, static files
  confined to their root, a rejected TLS connection no longer blocks the accept
  loop, kept hijacked connections no longer read a recycled request context.
- **mongo-driver v2.9.2**: CVE-2026-107324 and CVE-2026-107325 (`bson.Raw` and
  `RawArray` validation panics) and CVE-2026-88031 (GridFS delete matching file
  IDs too loosely). It keeps v2.8.2's CVE-2026-81521 fix, which v2.9.0 and
  v2.9.1 lost — do not pin those two.

---

## v2.7.1

The access log no longer reads the response body to size it. fiberzap's
`bytesSent` calls `Response.Body()`, which on a streamed response (SSE via
`SendStreamWriter`, a proxy with `StreamResponseBody`) copies the whole stream
into a buffer before the first byte reaches the client — every frame arrived at
once, as `Content-Length`. No code edits are required. Every dependency also
moves to its latest release; the upgrades that change what a running
application does are listed below.

| Change | Action needed |
|---|---|
| Streamed responses are delivered live and their access-log line carries `streamed: true` instead of `bytesSent`; buffered responses keep `bytesSent` | None. A log query that requires `bytesSent` on every line should treat `streamed: true` as "size unknown" |
| Cache drivers upgraded: fiber storage redis 3.5.1 → 3.6.0, badger 2.1.10 → 2.2.0, bbolt 2.1.9 → 2.2.0 | None for normal use. After `Close` (app shutdown) every store operation returns the driver's `ErrClosed` instead of reaching a closed client; badger TTLs round **up** to the next whole second instead of truncating down, so a key can no longer expire up to a second early; `CACHE_REDIS_RESET` on a cluster flushes every master instead of one shard |
| mongo-driver v2.8.0 → v2.8.2 | None. Fixes CVE-2026-81521 (`Client.BulkWrite` with a `.` in a caller-controlled database name could address another namespace) and stops a failed first write attempt (`NoWritesPerformed`) from returning as success — callers now see the real server error |

---

## v2.7.0

JSON logging now natively emits the log-contract shape that a collector
previously had to normalize. Nothing changes for text/console output, and no
code edits are required — but anything that parses the JSON log stream sees
different lines.

| Change | Action needed |
|---|---|
| JSON `level` values are lowercase: `"INFO"` → `"info"` (text/console output keeps `INFO`) | Update log queries, alerts, or parsers that match capitalized level values |
| With `LOG_FORMAT=json`, every line gains `service` (from `APP_NAME`), `env` (from `OEASE_ENV`, falling back to `GOE_ENV`) and `version` (from `APP_VERSION`) when set; unset values are omitted | None for tolerant consumers. Strict-schema pipelines should allow the new keys |
| The access log adds `route` (matched template), `host`, `bytesSent` and integer `duration_ms`; `latency` and all previous fields remain | None. Prefer `route` over `url` for grouping and `duration_ms` over the human-readable `latency` string |
| `Logger.WithContext(ctx)` was a no-op and now adds `request_id`/`trace_id`/`span_id` when the context carries them | None — lines logged through it gain correlation fields |
| Fiber upgraded v3.4.0 → v3.5.0 | None for GOE-managed configuration. The new `SkipUnmatchedRoutes` fast path is exposed as `http.WithSkipUnmatchedRoutes` / `FIBER_SKIP_UNMATCHED_ROUTES` and stays off by default |

---

## v2.6.0

The redis driver now executes the cache's compound operations — `Add`, `Pull`,
`Increment`/`Decrement` — as single server-side commands (`SET NX`, `GETDEL`,
`INCRBY`). Choosing redis means the cache is shared across processes, and the
previous lock-emulated operations were only atomic within one process: two app
replicas could lose increments or both consume the same `Pull`ed key. The
embedded drivers (memory, badger, bbolt) are unchanged — their process is the
store's only writer, so the in-process lock was already correct.

| Change | Action needed |
|---|---|
| Redis `Increment`/`Decrement` use `INCRBY`: atomic across processes, and an existing key's TTL is now **preserved** (previously reset to "never expires") | None for typical code — counters seeded with `Add(key, 0, window)` now expire with the window, which is the standard rate-limiter pattern. Keys *created* by `Increment` still never expire. If you relied on `Increment` immortalizing a TTL'd key, re-store the value without a TTL instead |
| Redis `Increment` on non-integer values returns Redis's error | Previously stored floats were truncated to int64 and strings rejected with `value is not a number`. Counters written by `Increment` were always integers, so this only affects keys that mixed `Set(float)` with `Increment` |
| Redis `Pull` uses `GETDEL`: atomic consume across processes | Requires Redis 6.2+ (2021). The key is now consumed even when unmarshaling into the destination fails; previously a failed unmarshal left it in place |
| Redis `Add` uses `SET NX`: atomic across processes | None — same semantics and same "key already exists" error |
| New optional `contract.AtomicCacheStore` capability interface | None. Custom drivers may implement it on their store to get the native path; stores without it keep the emulated per-process behaviour |

### Behaviour: compound cache operations are distributed-atomic on redis

The lock inside the cache is process-local, so on a shared Redis the emulated
`Increment` (get, add, set) under-counted when several replicas raced, `Add`
could double-admit, and `Pull` could hand the same one-time value to two
processes. Auth-shaped code — attempt lockouts, one-time codes — silently
depended on running a single replica. The native commands close exactly that
gap; the guarantee is per-key atomicity on a healthy shard, not durability
(Redis replication is asynchronous, so a failover can drop the newest writes).

---

## v2.5.0

The cache module drops its multi-store layer the same way MongoDB dropped
multi-connection in v2.4.0: GOE now manages exactly one cache, selected by a
single driver setting. The advertised-but-inert default TTL starts working.

| Change | Action needed |
|---|---|
| ⚠️ Multi-store support removed: `CACHE_STORE` and `CACHE_{name}_DRIVER` are rejected at startup; `cache.WithStore`/`WithStoreDriver`/`WithStorePrefix`/`WithStoreTTL` are gone | Set `CACHE_DRIVER=redis` (or `cache.WithDriver("redis")`). Startup validation names the replacement in its error message |
| ⚠️ `contract.CacheManager`, `contract.CacheConfig`, `contract.CacheStoreConfig` removed; DI provides `contract.Cache`; `goe.Cache()` returns `contract.Cache` | Inject `contract.Cache` instead of `contract.CacheManager`; drop `.Store()` calls: `goe.Cache().Store().Set(...)` → `goe.Cache().Set(...)` |
| ⚠️ `CacheManager.Extend` removed | Register custom backends with `cache.WithCustomDriver(name, factory)` — available before validation, unlike runtime `Extend` |
| `CACHE_TTL`/`cache.WithTTL` now actually apply | Callers passing `ttl == 0` to `Set`/`Add`/`Remember` get the configured default instead of "never expires". Unset `CACHE_TTL` keeps the old behaviour exactly. `Forever`/`RememberForever` and counters still never expire |
| Cache now **fails fast** at startup | The cache is built inside `goe.New` like the job and lock connections; stores were previously created lazily on first use, so an unreachable Redis let the app start and panic later. Verify deployments actually reach the cache backend |

### ⚠️ Breaking: one cache, not many stores

The store layer never earned its surface. Connection settings were always
per-driver (`CACHE_REDIS_*`), so two redis-backed stores could not point at
different hosts or databases — the per-store `Connection()` plumbing had no
callers, per-store TTLs were never read, `contract.CacheConfig` had no
implementation, and nothing in the framework or examples used a named store.
What remained was an alias of driver + prefix, at the cost of the
store/driver split and a name-matching resolution rule that made
`WithStore("redis")` quietly select a driver.

Selection is now one setting: `CACHE_DRIVER` / `cache.WithDriver`, defaulting
to `memory`. Removed keys fail startup validation with a migration message
instead of being silently ignored, so an app that relied on them cannot come
up with a differently-wired cache. Apps that only ever used the default store
with `CACHE_DRIVER` set — or no cache configuration at all — are unaffected.

Need two caches with different prefixes? Construct the second one directly
from the exported building blocks:

```go
sessions := cache.New(store, "sessions", 24*time.Hour)
```

### Behaviour: the default TTL is real now

`CACHE_TTL` (and `cache.WithTTL`) were validated and documented but never
consumed — every write path passed the caller's ttl straight to the store, so
`Set(key, v, 0)` always meant "no expiration". With a default configured,
`ttl == 0` in `Set`, `Add` and `Remember` now resolves to it. Without one,
nothing changes. `Forever` and `RememberForever` keep storing without
expiration, and `Increment`/`Decrement` counters never expire either way.

---

## v2.4.0

The MongoDB and migration modules join the code-first configuration pattern
(`goe.Options.MongoDB`/`Migrate` — additive, credentials stay
environment-only), MongoDB startup becomes fail-fast, and multi-connection
support is removed: GOE now manages exactly one MongoDB connection.

| Change | Action needed |
|---|---|
| ⚠️ Multi-connection support removed: `contract.MongoDB` loses `Connection(name)` and `ColFrom(...)`; `MONGO_CONNECTION`, `MONGO_CONNECTIONS` and `MONGO_{NAME}_*` are no longer read | None if you only used the default connection (`MONGO_URI`/`MONGO_DB_NAME` — unchanged). Apps with a second data source construct their own `mongo.Client` via the driver |
| MongoDB startup now **fails fast** when the connection is unreachable | Verify your deployments actually reach MongoDB; previously the app started anyway and panicked on first use. Tune `MONGO_PING_TIMEOUT` (default 5s) for slow clusters |
| New `MONGO_USERNAME`/`MONGO_PASSWORD` merge into credential-free URIs | None — optional; URI-embedded credentials still win |
| ⚠️ MongoDB test mocks (`TestMockConfig`, `TestMockLogger`, …) left the public package | They were test scaffolding; define your own mocks or use the `contract` interfaces directly |

### ⚠️ Breaking: one MongoDB connection, not many

Named connections (`MONGO_CONNECTION`, `MONGO_CONNECTIONS`,
`MONGO_{NAME}_URI` and friends) existed since v2's early days but were never
exercised in production, cost a per-request map lookup, and carried their own
validation and configuration surface. The module now manages a single
connection configured by the unprefixed `MONGO_*` variables — which is exactly
what every known deployment already does, so for those apps this release
changes nothing.

Removed with it: `contract.MongoDB.Connection`/`ColFrom` (their callers were
the only consumers of multi-connection), and the connection-scoped code
options. A second data source is the driver's job:

```go
client, err := mongo.Connect(options.Client().ApplyURI(analyticsURI))
```

### Behaviour: MongoDB startup verifies the connection

`mongo.Connect` performs no I/O, so earlier releases logged
"connected successfully" without ever reaching the server, swallowed failures,
and left a nil database behind — the crash surfaced later, inside a handler,
far from the cause. Startup now pings the connection (bounded by
`MONGO_PING_TIMEOUT`, default 5s) and aborts with a clear error when MongoDB
is unreachable, matching how the job and lock modules have always behaved.

### Data-access API: unchanged

`contract.MongoDB`'s `DB`, `Col` and `Client`, `goe.Mongo()`, the migration
registration API (`migrate.Register`, `migrate.New`, the helpers) and
`contract.Migrator` are untouched. Existing single-connection
`MONGO_*`/`MONGODB_MIGRATE_*` environment configuration keeps working
unchanged.

---

## v2.3.0

The cache, job and lock modules join HTTP as fully code-configurable
(`goe.Options.Cache/Job/Lock` — additive, credentials stay environment-only),
and request validation now lives entirely in the HTTP kernel, reached only
through `Ctx.Bind`. The standalone `validation` package is gone, validation
failures finally render as client errors, and three configuration behaviours
changed.

| Change | Action needed |
|---|---|
| ⚠️ `go.oease.dev/goe/v2/validation` package removed | Use `c.Bind()` (it validates); register custom rules with `goehttp.WithValidatorSetup` |
| ⚠️ `contract.HTTPValidator` removed; `HTTPKernel` loses `Validator()`/`HTTPValidator()` | Drop the calls — `Bind` invokes the validator itself |
| Bind/validation failures now render **400**, previously 500 | None — `return err` after `Bind` is now the right thing to do |
| Validation messages use json field names, one error at a time | None — clients see `email`, no longer `Email` |
| `phone`, `username`, `strong_password` tags now work through `Bind` | None — previously they panicked (unregistered) on the Bind path |
| `CACHE_STORE` naming a registered driver now selects that driver | Check apps setting `CACHE_STORE=redis` without `CACHE_DRIVER`: they silently ran on the memory driver and now actually use Redis |
| `JOB_REDIS_URL`/`CACHE_REDIS_URL` now honour the separate credential variables | None — they were silently ignored whenever a URL was set; URL-embedded credentials still win |
| New `LOCK_REDIS_USERNAME`/`LOCK_REDIS_PASSWORD` | None — optional, applies to any lock URL that embeds no userinfo |

### Behaviour: `CACHE_STORE` naming a registered driver selects it

`CACHE_STORE=redis` (or `cache.WithStore("redis")`) previously resolved the
driver as `CACHE_{STORE}_DRIVER` → `CACHE_DRIVER` → memory — so without a
separate `CACHE_DRIVER=redis` it silently ran on the in-memory driver, while
startup validation implied the store name was enough. A store named after a
registered driver (`memory`, `redis`, or a custom driver added through
`Extend`) now uses that driver directly. An explicitly configured driver still
wins, unknown store names are still rejected at startup, and named stores with
a configured registered driver now pass validation instead of being rejected.

### Behaviour: separate Redis credential variables always apply

`JOB_REDIS_USERNAME`/`JOB_REDIS_PASSWORD` and
`CACHE_REDIS_USERNAME`/`CACHE_REDIS_PASSWORD` were silently ignored whenever
the corresponding `*_REDIS_URL` was set. They now fill in whenever the URL
embeds no `user:pass` of its own (URL-embedded credentials win; for the cache
they are injected into the URL, leaving scheme, database and query parameters
untouched). The lock system gains `LOCK_REDIS_USERNAME`/`LOCK_REDIS_PASSWORD`
with the same semantics across every mode — single, sentinel, cluster and
redlock. This is what lets code-side options pick an endpoint while the
environment supplies the secret.

### ⚠️ Breaking: the `validation` package is removed

Everything it still offered was either a duplicate of `Ctx.Bind` (the
`ValidateRequest`/`ValidateQuery` wrappers, the middleware) or has moved into
the HTTP kernel (typed errors, human-readable messages, json tag names, the
custom rules). There is one way to validate a request now:

```go
if err := c.Bind().JSON(&req); err != nil {
    return err // parsed AND validated; renders as a structured 400
}
```

Custom rules register on the bundled validator at startup:

```go
goe.New(goe.Options{
    HTTP: []goehttp.Option{
        goehttp.WithValidatorSetup(func(v *validator.Validate) error {
            return v.RegisterValidation("slug", isSlug)
        }),
    },
})
```

Replacing the validator wholesale is still `goehttp.WithStructValidator`.

### Behaviour: validation and bind failures are 400s, one message at a time

Previously a raw `return err` after a failed `Bind` fell through GOE's error
handler as a 500, because neither `validator.ValidationErrors` nor Fiber's
parse failures are `*fiber.Error`. The handler now classifies them:

- Failed `validate` tags → **400**; the message *is* the first failed rule's
  message, in field declaration order — the client fixes it, resubmits, and
  sees the next. The same message feeds the error page for browsers and
  `format=text`.
- Malformed body / unconvertible parameter (Fiber's `*BindError`) → **400**.
- Raw `validator.ValidationErrors` from a replacement validator → **400**.

```json
{
  "message": "email must be a valid email address"
}
```

Submitted values are deliberately not echoed back — request DTOs carry
passwords. Handlers that want a different rendering (every field at once, a
custom envelope) can catch the typed error and use its fields:

```go
var ve *goehttp.ValidationError
if errors.As(err, &ve) { /* ve.Fields: field, tag, param, message */ }
```

---

## v2.2.2

No code edits required. Two defaults changed and one startup warning was added,
all around trusted proxies.

### Behaviour: `ProxyHeader` now defaults to `X-Forwarded-For`

`fiber.Config.ProxyHeader` was previously empty, so enabling trusted proxies
(`FIBER_TRUST_PROXY`/`WithTrustProxy` plus a proxy list) still left `Ctx.IP()`
returning the socket peer — behind a CDN, the edge's address rather than the
client's. Fiber only consults the header for peers in the trusted list, so the
new default is inert until trusted proxies are enabled.

- Not behind a proxy, or trusted proxies never enabled: no change.
- Trusted proxies enabled: `Ctx.IP()` and the access log `ip` field now report
  the client address instead of the proxy's. If you relied on logging the
  proxy's own address, restore it with `goehttp.WithProxyHeader("")`.

### Behaviour: `EnableIPValidation` now defaults to true

With validation off, Fiber returns the proxy header raw, so a client that
prepends a forged `X-Forwarded-For` entry pollutes the resolved IP. With
validation on, Fiber walks the chain right-to-left past trusted hops and
returns the first address a trusted proxy vouched for. Set
`FIBER_ENABLE_IP_VALIDATION=false` if you need the raw value.

### New: startup warning when trust is configured without a proxy header

Clearing `ProxyHeader` while trusted proxies are configured logs a warning at
startup, because `Ctx.IP()` cannot report client addresses in that state. The
configuration stays legal — the trust list alone still governs
`X-Forwarded-Proto`/`X-Forwarded-Host` handling — so it warns rather than
fails.

---

## v2.2.0

Everything below ships in one release. Two groups of breaking change, and three
behaviour changes worth knowing about even though they need no code edits.

**At a glance**

| Change | Action needed |
|---|---|
| Request validation moved to `Ctx.Bind` | Replace `ValidateRequest`-style calls; drop redundant validation after `Bind` |
| Four inert types/helpers removed | Compile error with a one-line fix, or none if you never referenced them |
| `goe.Options.HTTPPort` deprecated | None — prefer `goehttp.WithPort` |
| Scheduled jobs no longer fire at startup | None, unless you relied on the old behaviour |
| Fx dependency-injection logs default to `warn` | None — set `LOG_MODULE_LEVELS=fx:debug` to restore |
| HTTP server starts via `app.Listen` | None |

---

### ⚠️ Breaking: request validation moved to Fiber's Bind

Validation is no longer a separate injected service. Fiber owns the call site —
`Ctx.Bind` validates every binding — so GOE now installs a
`fiber.StructValidator` on the app and gets out of the way. This is the pattern
from [Fiber's validation guide](https://docs.gofiber.io/guide/validation).

**If you already use `c.Bind()`, nothing changes.** It validated before and it
validates now; you may find you can delete a redundant validation call.

#### `contract.HTTPValidator` reduced to one method

```go
// Before
type HTTPValidator interface {
    Validate(i any) error
    ValidateRequest(c fiber.Ctx, dst any) error
    ValidateQuery(c fiber.Ctx, dst any) error
    ValidateParams(c fiber.Ctx, dst any) error
    ValidateHeaders(c fiber.Ctx, dst any) error
    ValidateForm(c fiber.Ctx, dst any) error
}

// After — the same shape as fiber.StructValidator
type HTTPValidator interface {
    Validate(i any) error
}
```

The removed methods duplicated `Bind`:

```go
// Before
if err := v.ValidateRequest(c, &dto); err != nil { return err }

// After — parses and validates in one step
if err := c.Bind().JSON(&dto); err != nil { return err }
```

`Bind` covers `JSON`, `Query`, `URI`, `Form`, `Header`, `Cookie`, `XML`, `CBOR`
and `MsgPack`, plus `SkipValidation()` and `WithAutoHandling()`.

#### `contract.ValidationProvider` and `contract.ValidationMiddleware` — removed

Both existed only so the validation package could be injected. Nothing consumed
them. `validation.NewProvider` and `validation.NewMiddlewareProvider` are gone
with them.

#### Validation removed from dependency injection

`http.Services.Validator`, `http.GetValidator(c)`, `http.ServiceProvider.Validator`
and `Module.ProvideValidator()` are removed. Fiber calls the validator itself, so
handlers never needed a reference:

```go
// Before
v := http.GetValidator(c)
if err := v.Validate(&dto); err != nil { return err }

// After
if err := c.Bind().JSON(&dto); err != nil { return err }
```

`kernel.Validator()` and `kernel.HTTPValidator()` still work but are deprecated.

#### Customising validation

```go
// Add a rule to the bundled validator
goe.New(goe.Options{
    HTTP: []goehttp.Option{
        goehttp.WithValidatorSetup(func(v *validator.Validate) error {
            return v.RegisterValidation("slug", isSlug)
        }),
    },
})

// Or replace it entirely
goehttp.WithStructValidator(myValidator)
```

Combining the two is rejected at startup: a setup configures the bundled
validator, so it cannot apply to a replacement.

#### The `validation` package is deprecated

It still compiles and its tests still pass, so existing code keeps working. The
middleware (`NewMiddleware`, `ValidateBody`, `GetValidatedBody`, …), the
parse-and-validate helpers and the `IsEmail`/`IsUUID`-style sugar are all marked
`Deprecated:` with their replacements. `Validator` itself remains useful for
validating structs outside a request.

One caveat: a `validation.New()` validator is a **separate instance** from the one
the HTTP kernel installs, so rules registered on it do not affect `Bind`. Use
`WithValidatorSetup` for request rules.

---

### ⚠️ Breaking: removed inert HTTP types and helpers

Four exported symbols were removed. Every one of them was **dead on arrival** —
nothing in GOE ever read them, so code that set them had no effect. They are
listed individually below with a replacement.

If you never referenced these four names, this group costs you nothing. (The
validation change above is separate and may still apply.)

#### `contract.HTTPConfig` — removed

Nothing in GOE ever read this struct. Filling it in configured nothing, and its
`Prefork` field was actively misleading: prefork is a `fiber.ListenConfig`
setting, not one GOE ever applied.

HTTP configuration now lives in `core/http` as options.

```go
// Before — compiled, but did nothing
cfg := contract.HTTPConfig{
    Host:      "0.0.0.0",
    Port:      8080,
    BodyLimit: 16 << 20,
    Prefork:   true,
}

// After
import goehttp "go.oease.dev/goe/v2/core/http"

goe.New(goe.Options{
    HTTP: []goehttp.Option{
        goehttp.WithHost("0.0.0.0"),
        goehttp.WithPort(8080),
        goehttp.WithBodyLimit(16 << 20),
        goehttp.WithEnablePrefork(true), // now actually applied
    },
})
```

See [CONFIGURATION.md](CONFIGURATION.md#configuring-in-go-code) for the full
option list.

#### `contract.RouteInfo` — removed

Never populated or returned by any GOE API. For route introspection, use Fiber
directly:

```go
// After
for _, routes := range goe.HTTP().App().Stack() {
    for _, r := range routes {
        fmt.Println(r.Method, r.Path)
    }
}
```

#### `http.RegisterRoutes` — removed

It was an identity function (`func RegisterRoutes(fn) any { return fn }`), so it
only added a layer of indirection. Pass the function straight to `Invokers`:

```go
// Before
Invokers: []any{
    http.RegisterRoutes(func(r http.RouteRegistrar) {
        r.HTTP.App().Get("/", myHandler)
    }),
}

// After
Invokers: []any{
    func(r http.RouteRegistrar) {
        r.HTTP.App().Get("/", myHandler)
    },
}
```

`http.RouteRegistrar` itself is unchanged and still works as an `fx.In`
parameter.

#### `http.NewField` — removed

A duplicate of `log.NewField` that happened to live in the HTTP package. The
logging one is canonical:

```go
// Before
f := http.NewField("key", "value")

// After
import "go.oease.dev/goe/v2/core/log"

f := log.NewField("key", "value")
```

### Deprecated: `goe.Options.HTTPPort`

Still works, no action required. Prefer `goehttp.WithPort`, which takes
precedence over it:

```go
// Before
goe.New(goe.Options{WithHTTP: true, HTTPPort: 8080})

// After
goe.New(goe.Options{HTTP: []goehttp.Option{goehttp.WithPort(8080)}})
```

### Behaviour change: scheduled jobs no longer fire at startup

**This was a bug, and the fix changes when your jobs run.**

Last-run times were held in a process-local map, so a schedule the process had not
seen yet had a zero last-run — and `Schedule.Next(zeroTime)` returns a moment in
year 1, which is always in the past. Every schedule therefore dispatched
immediately on the first scheduler tick, regardless of its period: a
`DailyAt(3, 0)` report went out on every deploy, and each replica fired from its
own private map, defeating the distributed tick lock.

Last-run times now live in Redis under `<KeyPrefix>schedule:lastrun:<name>`, with a
30-day refreshing TTL, cleared when a schedule is unregistered. The first time any
instance sees a schedule it anchors to the current time and does *not* dispatch;
the first run happens at the schedule's next genuine occurrence. Missed
occurrences are not backfilled, matching cron.

No code change is required. If you were relying — knowingly or not — on jobs
running at boot, dispatch them explicitly instead:

```go
// Run once at startup, in addition to the schedule
if _, err := goe.Job().Dispatch(ctx, &contract.JobDefinition{Name: "my-job"}); err != nil {
    return err
}
```

### Behaviour change: invalid schedules are now rejected at registration

`RegisterSchedule` previously accepted a mistyped cron expression and the job
simply never ran, because `Cron()` cannot return an error and turns a parse failure
into a schedule whose next run is in the year 9999. It now returns
`ErrInvalidSchedule`, naming the schedule and the parse error — including when the
expression is wrapped in `Between` or `SkipWeekends`.

A nil `Schedule` or `Handler`, or an empty name, is also rejected. A nil `Schedule`
previously panicked inside the scheduler goroutine, which has no recover, taking
the process down.

If a schedule of yours was silently never running, this will now surface as a
startup error. That is the point.

### Behaviour change: Fx dependency-injection logs default to `warn`

Fx emits a line for every constructor supplied, provided, decorated and run, plus
each lifecycle hook. That was tied to the global log level, so raising
`LOG_LEVEL=debug` for your own code buried it under dependency-graph narration —
19 lines on a small app.

`fx` is now a log module in its own right, defaulting to `warn`. Provide and invoke
*failures* are still logged. To get the detail back when diagnosing wiring:

```bash
LOG_MODULE_LEVELS=fx:debug
```

GOE's own boot output was trimmed the same way: ten `WithX flag` lines collapsed
into one debug line, and `Registering X module` moved to debug.

### Internal change: HTTP server startup

No action required. Recorded here because it touches the startup path.

The HTTP module now starts through `app.Listen(addr, listenConfig)` instead of
binding its own `net.Listener` and calling `app.Listener(ln)`. This is what
makes TLS, prefork and unix sockets configurable at all — Fiber ignores
`ListenConfig` entirely on the `Listener` path, and logs
`"Prefork isn't supported for custom listeners"` there.

Existing behaviour is preserved deliberately:

- `OnStart` still returns only once the port is accepting connections.
- A bind failure such as `EADDRINUSE` is still returned from `OnStart` rather
  than logged and swallowed. Its message now carries the address it tried.
- The listener network is pinned to `tcp` (dual-stack). Fiber's own default is
  `tcp4`, which would have silently dropped IPv6; override with
  `goehttp.WithListenerNetwork` if you want to narrow it.
- Fiber's startup banner is unchanged — it was already printed on the
  `Listener` path. Suppress it with
  `goehttp.WithDisableStartupMessage(true)`.

### Note on versioning

Removing exported symbols is normally a major-version change under semantic
versioning. This ships as a **v2 minor** deliberately, but the two groups of
removal differ in how much they can actually affect you, so they are worth
separating honestly:

- **`contract.HTTPConfig`, `contract.RouteInfo`, `http.RegisterRoutes`,
  `http.NewField`** were inert or duplicated. Code referencing them compiled but
  had no runtime effect (or delegated straight to an identical function), so the
  blast radius is a compile error with a one-line fix — not changed behaviour.

- **The validation API** is a genuine functional removal. Nothing inside GOE used
  `ValidateRequest`, `ValidateBody`, `ValidationProvider` or the injected
  `Services.Validator`, but they worked, and your code may call them. Migration is
  mechanical — replace them with `c.Bind()` — and `Bind` already validated for
  anyone using it, so behaviour is unchanged in the common case.

The `validation` package itself still compiles, so nothing breaks merely by
upgrading; you get deprecation notices pointing at the replacements.

If you would rather not take a functional removal in a minor, pin to `v2.1.x`.

---

## Earlier releases

See the [GitHub releases](https://github.com/oeasenet/goe/releases) page.
