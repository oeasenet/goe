# Migration Guide

Breaking changes and how to move past them. Releases without breaking changes
are not listed here.

---

## v2.2.0

### ⚠️ Breaking: removed inert HTTP types and helpers

Four exported symbols were removed. Every one of them was **dead on arrival** —
nothing in GOE ever read them, so code that set them had no effect. They are
listed individually below with a replacement.

If you never referenced these names, this release is a drop-in upgrade.

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
versioning. This ships as a **v2 minor** deliberately: all four removals are
inert declarations — code referencing them compiled but had no runtime effect —
so the practical blast radius is a compile error with a one-line fix, not changed
behaviour. If you hit one of them, the replacement is listed above.

---

## Earlier releases

See the [GitHub releases](https://github.com/oeasenet/goe/releases) page.
