# Cache Module

The cache module provides a common interface built on Fiber's storage drivers. It supports in-memory, Redis and many other backends.

Enable it when creating the application:

```go
_ = goe.New(goe.Options{WithCache: true})
```

## Using the Cache

You can retrieve the default store either globally or by injection.

```go
func globalExample() {
    goe.Cache().Set("foo", "bar", time.Minute)
}

func injected(c contract.Cache) {
    c.Set("foo", "bar", time.Minute)
}
```

Generic helpers in `core/cache` allow type-safe operations:

```go
value, _ := cache.GetT[string](goe.Cache(), "foo")
```

Multiple stores are supported via `contract.CacheManager`.

```go
func useStores(m contract.CacheManager) {
    primary := m.Store()
    sessions := m.Store("sessions")
    _ = primary
    _ = sessions
}
```

Configuration options like driver, prefix and TTL are documented in [Configuration](configuration.md).
