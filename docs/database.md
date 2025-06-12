# Database Module

The optional database module provides a GORM based connection manager. Enable it with `WithDB: true` when constructing the application.

```go
_ = goe.New(goe.Options{WithDB: true})
```

A database handle (`*gorm.DB`) can be obtained by injection or via the global helper `goe.DB()`.

```go
func injected(db *gorm.DB) {
    // use db
}

func global() {
    db := goe.DB()
    // use db
}
```

Database configuration comes from environment variables such as `DB_DSN` or `DB_HOST`, `DB_PORT` and friends. See [Configuration](configuration.md) for details.
