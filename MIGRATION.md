# Migration Guide

Breaking changes, and behaviour changes that need no code edit but will alter what
your application does. Releases with neither are not listed here.

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
