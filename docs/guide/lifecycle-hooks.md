# Lifecycle Hooks

GOE provides a simple, developer-friendly API for hooking into your application's lifecycle. Under the hood, these hooks use [Uber's Fx lifecycle system](https://uber-go.github.io/fx/lifecycle.html), but with a much simpler interface.

## Overview

Lifecycle hooks allow you to:
- Run initialization code after all dependencies are wired
- Perform cleanup before the application shuts down
- Access all GOE services and custom providers
- Control execution order through Fx's dependency graph

## Basic Usage

### Using Options

The primary way to register hooks is when creating the application:

```go
app := goe.New(goe.Options{
    WithHTTP: true,
    WithDB: true,
    OnStart: []func(context.Context) error{
        initializeCache,
        seedDatabase,
        startBackgroundWorkers,
    },
    OnStop: []func(context.Context) error{
        stopBackgroundWorkers,
        saveApplicationState,
    },
})
```

### Using WithHooks Helper

For a more fluent API, use the `WithHooks` helper:

```go
app := goe.New(
    goe.WithHooks(
        goe.WithOnStart(func(ctx context.Context) error {
            logger := goe.Log()
            logger.Info("Starting application")
            return nil
        }),
        goe.WithOnStop(func(ctx context.Context) error {
            logger := goe.Log()
            logger.Info("Stopping application")
            return nil
        }),
    ),
    goe.Options{
        WithHTTP: true,
        WithDB: true,
    },
)
```

## Hook Execution Order

Understanding when hooks execute is crucial:

1. **Dependency Injection Phase**: All providers are instantiated and wired
2. **Module OnStart**: Core modules (Config, Log, DB, HTTP, etc.) start in dependency order
3. **User OnStart Hooks**: Your OnStart hooks execute after all modules
4. **Application Running**: HTTP server starts listening, application serves requests
5. **Shutdown Signal**: Application receives shutdown signal (Ctrl+C, SIGTERM, etc.)
6. **User OnStop Hooks**: Your OnStop hooks execute first (in reverse order)
7. **Module OnStop**: Core modules stop in reverse dependency order

## Accessing Services in Hooks

All GOE services are available in hooks via global accessors:

```go
app := goe.New(goe.Options{
    WithHTTP: true,
    WithDB: true,
    WithCache: true,
    OnStart: []func(context.Context) error{
        func(ctx context.Context) error {
            // Access configuration
            config := goe.Config()
            apiKey := config.GetString("API_KEY")
            
            // Access logger
            logger := goe.Log()
            logger.Info("Initializing with API key", "key_length", len(apiKey))
            
            // Access database
            db := goe.DB()
            if err := db.Ping(); err != nil {
                return fmt.Errorf("database not ready: %w", err)
            }
            
            // Access HTTP server
            http := goe.HTTP()
            http.App().Get("/health", healthHandler)
            
            // Access cache
            cache := goe.Cache()
            return cache.Set(ctx, "app:started", time.Now().Unix(), 0)
        },
    },
})
```

## Using Dependency Injection in Hooks

For more complex scenarios, you can use Fx's dependency injection directly:

```go
type Services struct {
    fx.In
    DB     contract.DB
    Cache  contract.CacheManager
    Logger contract.Logger
    Config contract.Config
}

app := goe.New(goe.Options{
    WithDB: true,
    WithCache: true,
    Invokers: []any{
        func(lc fx.Lifecycle, services Services) {
            lc.Append(fx.Hook{
                OnStart: func(ctx context.Context) error {
                    services.Logger.Info("Starting with injected services")
                    return initializeWithServices(services)
                },
                OnStop: func(ctx context.Context) error {
                    return cleanupServices(services)
                },
            })
        },
    },
})
```

## Common Use Cases

### Database Migrations

```go
app := goe.New(goe.Options{
    WithDB: true,
    OnStart: []func(context.Context) error{
        func(ctx context.Context) error {
            db := goe.DB()
            
            // Auto-migrate models
            err := db.AutoMigrate(
                &User{},
                &Product{},
                &Order{},
            )
            if err != nil {
                return fmt.Errorf("migration failed: %w", err)
            }
            
            goe.Log().Info("Database migrations completed")
            return nil
        },
    },
})
```

### Background Workers

```go
var workerCtx context.Context
var workerCancel context.CancelFunc

app := goe.New(goe.Options{
    OnStart: []func(context.Context) error{
        func(ctx context.Context) error {
            workerCtx, workerCancel = context.WithCancel(context.Background())
            
            // Start background worker
            go func() {
                ticker := time.NewTicker(5 * time.Minute)
                defer ticker.Stop()
                
                for {
                    select {
                    case <-workerCtx.Done():
                        return
                    case <-ticker.C:
                        processBackgroundTasks()
                    }
                }
            }()
            
            goe.Log().Info("Background worker started")
            return nil
        },
    },
    OnStop: []func(context.Context) error{
        func(ctx context.Context) error {
            if workerCancel != nil {
                workerCancel()
            }
            
            // Wait for worker to finish with timeout
            select {
            case <-time.After(5 * time.Second):
                goe.Log().Warn("Worker shutdown timeout")
            case <-workerCtx.Done():
                goe.Log().Info("Worker stopped gracefully")
            }
            
            return nil
        },
    },
})
```

### External Service Initialization

```go
var redisClient *redis.Client

app := goe.New(goe.Options{
    OnStart: []func(context.Context) error{
        func(ctx context.Context) error {
            config := goe.Config()
            
            redisClient = redis.NewClient(&redis.Options{
                Addr:     config.GetString("REDIS_ADDR"),
                Password: config.GetString("REDIS_PASSWORD"),
                DB:       config.GetInt("REDIS_DB"),
            })
            
            // Test connection
            _, err := redisClient.Ping(ctx).Result()
            if err != nil {
                return fmt.Errorf("redis connection failed: %w", err)
            }
            
            goe.Log().Info("Redis connected")
            return nil
        },
    },
    OnStop: []func(context.Context) error{
        func(ctx context.Context) error {
            if redisClient != nil {
                return redisClient.Close()
            }
            return nil
        },
    },
})
```

### Cache Warming

```go
app := goe.New(goe.Options{
    WithDB: true,
    WithCache: true,
    OnStart: []func(context.Context) error{
        func(ctx context.Context) error {
            cache := goe.Cache()
            db := goe.DB()
            
            // Load frequently accessed data into cache
            var configs []AppConfig
            if err := db.Find(&configs).Error; err != nil {
                return err
            }
            
            for _, cfg := range configs {
                key := fmt.Sprintf("config:%s", cfg.Key)
                if err := cache.Set(ctx, key, cfg.Value, 1*time.Hour); err != nil {
                    goe.Log().Warn("Failed to cache config", "key", cfg.Key, "error", err)
                }
            }
            
            goe.Log().Info("Cache warmed up", "items", len(configs))
            return nil
        },
    },
})
```

## Error Handling

### OnStart Errors

If an OnStart hook returns an error:
- The application startup is aborted
- Any successfully started components are stopped
- The error is logged and the application exits

```go
app := goe.New(goe.Options{
    OnStart: []func(context.Context) error{
        func(ctx context.Context) error {
            if !isLicenseValid() {
                return fmt.Errorf("invalid license key")  // This will abort startup
            }
            return nil
        },
    },
})
```

### OnStop Errors

If an OnStop hook returns an error:
- The error is logged
- Shutdown continues with remaining hooks
- The application still exits

```go
app := goe.New(goe.Options{
    OnStop: []func(context.Context) error{
        func(ctx context.Context) error {
            if err := saveState(); err != nil {
                // Log error but don't prevent shutdown
                goe.Log().Error("Failed to save state", "error", err)
                return err  // Error is logged but shutdown continues
            }
            return nil
        },
    },
})
```

## Best Practices

1. **Keep hooks focused**: Each hook should have a single responsibility
2. **Handle context cancellation**: Respect the context timeout
3. **Log appropriately**: Use the logger to track hook execution
4. **Fail fast on critical errors**: Return errors for critical initialization failures
5. **Graceful degradation**: For non-critical features, log warnings instead of failing

```go
app := goe.New(goe.Options{
    OnStart: []func(context.Context) error{
        func(ctx context.Context) error {
            logger := goe.Log()
            
            // Critical initialization - fail if this doesn't work
            if err := initializeCriticalService(); err != nil {
                return fmt.Errorf("critical service failed: %w", err)
            }
            
            // Non-critical feature - log but don't fail
            if err := initializeOptionalFeature(); err != nil {
                logger.Warn("Optional feature unavailable", "error", err)
                // Continue without failing
            }
            
            return nil
        },
    },
})
```

## Testing Hooks

You can test hooks by creating a test application:

```go
func TestMyHook(t *testing.T) {
    var hookExecuted bool
    
    app := goe.New(goe.Options{
        WithDB: false,  // Minimal setup for testing
        OnStart: []func(context.Context) error{
            func(ctx context.Context) error {
                hookExecuted = true
                return nil
            },
        },
    })
    
    ctx := context.Background()
    err := app.Start(ctx)
    require.NoError(t, err)
    
    assert.True(t, hookExecuted)
    
    err = app.Stop(ctx)
    require.NoError(t, err)
}
```

## Advanced: Custom Module with Hooks

For complex initialization, consider creating a custom module:

```go
type BackgroundWorkerModule struct {
    workers []Worker
    logger  contract.Logger
}

func NewBackgroundWorkerModule(logger contract.Logger) contract.Module {
    return &BackgroundWorkerModule{
        logger: logger,
    }
}

func (m *BackgroundWorkerModule) Name() string {
    return "background-workers"
}

func (m *BackgroundWorkerModule) OnStart(ctx context.Context) error {
    for _, worker := range m.workers {
        if err := worker.Start(ctx); err != nil {
            return fmt.Errorf("failed to start worker %s: %w", worker.Name(), err)
        }
    }
    m.logger.Info("All workers started", "count", len(m.workers))
    return nil
}

func (m *BackgroundWorkerModule) OnStop(ctx context.Context) error {
    for _, worker := range m.workers {
        if err := worker.Stop(ctx); err != nil {
            m.logger.Error("Failed to stop worker", "name", worker.Name(), "error", err)
        }
    }
    return nil
}
```

## Summary

GOE's lifecycle hooks provide a clean, simple API built on top of Fx's powerful lifecycle management. Whether you need to initialize services, start background workers, or perform cleanup, hooks give you the control you need at exactly the right moment in your application's lifecycle.