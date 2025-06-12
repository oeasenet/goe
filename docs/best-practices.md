# Best Practices

This page collects recommendations for structuring and maintaining your Goe projects.

## Project Layout

- Keep the main application entry point under `cmd/` and domain packages under `internal/`.
- Separate HTTP handlers, services and repositories within each domain package.
- Add a `pkg/` directory for reusable utilities or middleware that might be shared outside `internal/`.

```
myapp/
├── cmd/api/main.go
├── internal/
│   └── user/
│       ├── handler.go
│       ├── service.go
│       └── repository.go
├── pkg/
│   └── middleware/
└── go.mod
```

## Dependency Injection

Prefer constructor functions that take a single parameters struct annotated with `fx.In`. This keeps dependencies explicit and avoids long argument lists.

```go
type ServiceParams struct {
    fx.In
    Logger contract.Logger
    Config contract.Config
}

func NewService(p ServiceParams) *Service {
    return &Service{log: p.Logger, cfg: p.Config}
}
```

## Graceful Shutdown

Modules implementing `contract.Module` should release resources in `OnStop`. Background goroutines should respect the context passed to these hooks.

## Logging

Use structured logging through `contract.Logger` and avoid formatting log messages manually. Attach contextual fields whenever possible.

## Error Handling

Wrap underlying errors with additional context and log them at the appropriate level. Use sentinel errors for expected failure cases.
