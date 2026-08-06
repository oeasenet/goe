# GOE Framework Roadmap

> Vision document for GOE framework evolution - capturing ideas for future development.

## Current State (v2)

### What GOE Does Well

**Architecture & Design**
- Clean contract-driven design with interfaces - excellent for testability
- Smart choice of Uber Fx for DI - battle-tested, type-safe
- Good module lifecycle management (OnStart/OnStop)
- Sensible defaults with configuration flexibility

**Technology Stack**
- GoFiber v3 - fastest Go HTTP framework
- Uber Zap - production-grade structured logging
- GORM - mature ORM with good ecosystem
- Redis - background job processing and distributed locking

**Production Readiness**
- Health checks with liveness/readiness probes (K8s-ready)
- Prometheus metrics out of the box
- OpenTelemetry tracing
- Graceful shutdown handling

### Current Modules

| Module | Status | Description |
|--------|--------|-------------|
| `http` | ✅ Stable | GoFiber v3 HTTP server with middleware |
| `config` | ✅ Stable | Environment-based configuration |
| `log` | ✅ Stable | Uber Zap structured logging |
| `db` | ✅ Stable | GORM SQL database (MySQL, PostgreSQL, SQLite, SQL Server) |
| `mongodb` | ✅ Stable | MongoDB native driver |
| `cache` | ✅ Stable | Pluggable cache drivers (memory, Redis, Badger, bbolt, custom) |
| `job` | ✅ Stable | Background job processing with scheduling and retries |
| `lock` | ✅ Stable | Distributed locking (Redlock) |
| `health` | ✅ Stable | Health checks (liveness/readiness) |
| `metrics` | ✅ Stable | Prometheus metrics |
| `otel` | ✅ Stable | OpenTelemetry tracing |
| `shutdown` | ✅ Stable | Graceful shutdown management |
| `validation` | ✅ Stable | Request validation with struct tags |

---

## Go Toolchain Follow-ups

- **`encoding/json/v2`**: becomes the default `encoding/json` backend in Go
  1.27 — evaluate the cache serialization path and webresult once GOE's
  baseline moves to 1.27 (it is GOEXPERIMENT-only on 1.26).
- **Goroutine-leak profile**: experimental in 1.26
  (`GOEXPERIMENT=goroutineleakprofile`), aimed at on-by-default in 1.27 —
  consider surfacing `/debug/pprof/goroutineleak` alongside the existing
  observability endpoints when it stabilizes.

## Gap Analysis

### Critical (Needed by ~95% of Apps)

#### 1. Authentication & Authorization
**Current State:** Not provided - users must build from scratch

**What's Missing:**
- JWT middleware with access/refresh token handling
- Session management (cookie-based, Redis-backed)
- Role-Based Access Control (RBAC)
- Permission system
- OAuth2/OIDC provider integration (Google, GitHub, etc.)
- API key authentication
- Multi-tenancy support

**Impact:** Very high - almost every app needs authentication

#### 2. Background Jobs / Task Queue ✅ IMPLEMENTED
**Current State:** Fully implemented with Redis backend

**Features:**
- ✅ Delayed job execution ("run this in 5 minutes")
- ✅ Scheduled/cron jobs with human-friendly syntax
- ✅ Persistent job queue (Redis-backed)
- ✅ Retry with exponential backoff
- ✅ Dead letter queue for failed jobs
- ✅ Multiple named queues
- ✅ Concurrency control
- ✅ Job uniqueness/deduplication
- ✅ Health checks and statistics

**Impact:** Completed - enable with `WithJob: true`

#### 3. Database Migrations
**Current State:** Only GORM AutoMigrate (not suitable for production)

**What's Missing:**
- Migration file management (up/down)
- Version tracking in database
- Rollback support
- Migration status command
- SQL and Go-based migrations
- Seed data management

**Impact:** High - essential for production deployments

### Important (Developer Experience)

#### 4. CLI / Command Runner
**Current State:** No CLI support

**What's Missing:**
- Built-in command runner
- Standard commands (migrate, seed, serve)
- Custom command registration
- Code generators (`goe make:handler`, `goe make:model`)
- Interactive prompts

**Impact:** Medium-high - significantly improves DX

#### 5. Testing Utilities
**Current State:** No testing helpers provided

**What's Missing:**
- HTTP test request builders
- Mock implementations of all contracts
- Test database helpers (transactions, cleanup)
- Fixture/factory system for test data
- Integration test setup helpers

**Impact:** Medium-high - testing is crucial for quality

#### 6. API Documentation
**Current State:** No auto-documentation

**What's Missing:**
- OpenAPI/Swagger spec generation
- Auto-generated docs from handler annotations
- Interactive API explorer
- Export to Postman/Insomnia

**Impact:** Medium - important for API consumers

#### 7. Security Middleware Bundle
**Current State:** Fiber has these, but not GOE-wrapped with sensible defaults

**What's Missing:**
- CORS configuration helper with presets
- CSRF protection middleware
- Security headers (X-Frame-Options, CSP, HSTS, etc.)
- Rate limiting (per-user, per-IP, distributed)
- Request sanitization
- SQL injection protection helpers

**Impact:** Medium - security is often overlooked

### Nice to Have

#### 8. File Storage Abstraction
**What's Missing:**
- Unified interface for Local/S3/GCS/Azure Blob
- File upload helpers with validation
- Image processing integration
- Temporary URL generation

#### 9. Mail / Notifications
**What's Missing:**
- Email sending abstraction (SMTP, SendGrid, SES, etc.)
- Email templates
- Notification channels (email, SMS, push, Slack, webhook)
- Queueable notifications
- Notification preferences per user

#### 10. WebSocket Support
**What's Missing:**
- WebSocket connection manager
- Rooms/channels abstraction
- Broadcast helpers
- Presence tracking

#### 11. GraphQL Support
**What's Missing:**
- GraphQL handler integration
- Schema-first or code-first approach
- Subscription support via WebSocket

#### 12. Service Discovery
**What's Missing:**
- Service registration (Consul, etcd)
- Client-side load balancing
- Health-aware routing

---

## Proposed Roadmap

### Phase 1: Core Completeness (High Priority)

```
goe/auth          Authentication & authorization module
goe/jobs          Background job processing & scheduling
goe/migrate       Database migration management
goe/cli           Command-line interface & runner
```

### Phase 2: Developer Experience

```
goe/testing       Test utilities, mocks, and helpers
goe/openapi       OpenAPI/Swagger documentation
goe/security      CORS, CSRF, headers, rate limiting bundle
goe/pagination    Cursor & offset pagination helpers
```

### Phase 3: Extended Features

```
goe/storage       File storage abstraction
goe/mail          Email sending
goe/notify        Multi-channel notifications
goe/ws            WebSocket manager
```

### Phase 4: Advanced

```
goe/graphql       GraphQL support
goe/discovery     Service discovery
goe/saga          Distributed transaction patterns
```

---

## Module Design Specifications

### Auth Module (`goe/auth`)

**Goals:**
- Zero-config JWT authentication with sensible defaults
- Pluggable session backends
- Flexible RBAC system
- Easy OAuth2 integration

**Proposed API:**

```go
// Enable auth module
goe.New(goe.Options{
    WithAuth: true,
})

// Configuration (.env)
// AUTH_JWT_SECRET=your-secret-key
// AUTH_JWT_EXPIRY=15m
// AUTH_REFRESH_EXPIRY=7d
// AUTH_SESSION_DRIVER=redis

// Middleware usage
app.Use(auth.Middleware())                           // Validates JWT
app.Get("/admin", auth.RequireRole("admin"), handler)
app.Get("/users", auth.RequirePermission("users:read"), handler)
app.Get("/public", auth.Optional(), handler)         // Parses but doesn't require

// In handlers
func handler(c fiber.Ctx) error {
    user := auth.User(c)           // Get authenticated user
    if auth.Can(c, "posts:delete") {
        // has permission
    }
    return nil
}

// Token generation
tokens, err := auth.GenerateTokens(userID, roles, permissions)
// tokens.AccessToken, tokens.RefreshToken

// Token refresh
newTokens, err := auth.RefreshTokens(refreshToken)

// Logout (invalidate refresh token)
auth.Logout(refreshToken)
```

**Contract:**

```go
type AuthManager interface {
    // Token operations
    GenerateTokens(userID string, claims map[string]any) (*TokenPair, error)
    RefreshTokens(refreshToken string) (*TokenPair, error)
    ValidateToken(token string) (*Claims, error)
    Logout(refreshToken string) error

    // Authorization
    HasRole(ctx fiber.Ctx, role string) bool
    HasPermission(ctx fiber.Ctx, permission string) bool
    HasAnyRole(ctx fiber.Ctx, roles ...string) bool
    HasAllPermissions(ctx fiber.Ctx, permissions ...string) bool
}

type TokenPair struct {
    AccessToken  string
    RefreshToken string
    ExpiresAt    time.Time
}

type Claims struct {
    UserID      string
    Roles       []string
    Permissions []string
    Custom      map[string]any
}
```

---

### Jobs Module (`goe/jobs`)

**Goals:**
- Simple job definition and dispatch
- Reliable execution with retries
- Scheduled/cron jobs
- Multiple queue support with priorities

**Proposed API:**

```go
// Enable jobs module
goe.New(goe.Options{
    WithJobs: true,
})

// Configuration (.env)
// JOBS_DRIVER=redis
// JOBS_REDIS_URL=redis://localhost:6379
// JOBS_DEFAULT_QUEUE=default
// JOBS_CONCURRENCY=10

// Define a job
type SendWelcomeEmail struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
}

func (j SendWelcomeEmail) Handle(ctx context.Context) error {
    // Send the email
    return mailer.Send(j.Email, "Welcome!", "...")
}

func (j SendWelcomeEmail) Queue() string {
    return "emails" // Optional: specify queue
}

func (j SendWelcomeEmail) Retries() int {
    return 3 // Optional: max retries
}

func (j SendWelcomeEmail) Timeout() time.Duration {
    return 30 * time.Second // Optional: job timeout
}

// Dispatch jobs
jobs.Dispatch(SendWelcomeEmail{UserID: "123", Email: "user@example.com"})

// Delayed dispatch
jobs.Delay(5*time.Minute, SendWelcomeEmail{...})

// Scheduled jobs (cron)
jobs.Schedule("0 9 * * *", DailyReport{})      // Every day at 9am
jobs.Schedule("*/5 * * * *", CleanupTemp{})    // Every 5 minutes

// Batch dispatch
jobs.Batch(
    SendWelcomeEmail{UserID: "1"},
    SendWelcomeEmail{UserID: "2"},
    SendWelcomeEmail{UserID: "3"},
)

// Chain jobs (run sequentially)
jobs.Chain(
    ProcessOrder{OrderID: "123"},
    SendConfirmation{OrderID: "123"},
    NotifyWarehouse{OrderID: "123"},
)
```

**Contract:**

```go
type JobManager interface {
    Dispatch(job Job) error
    Delay(delay time.Duration, job Job) error
    Schedule(cron string, job Job) error
    Batch(jobs ...Job) error
    Chain(jobs ...Job) error

    // Management
    Cancel(jobID string) error
    Retry(jobID string) error
    Purge(queue string) error
    Stats() (*QueueStats, error)
}

type Job interface {
    Handle(ctx context.Context) error
}

// Optional interfaces jobs can implement
type QueueableJob interface {
    Queue() string
}

type RetryableJob interface {
    Retries() int
    RetryDelay() time.Duration
}

type TimeoutJob interface {
    Timeout() time.Duration
}
```

---

### Migrate Module (`goe/migrate`)

**Goals:**
- Version-controlled database changes
- Safe rollbacks
- Both SQL and Go migrations
- CLI integration

**Proposed API:**

```go
// Enable migrations
goe.New(goe.Options{
    WithMigrate: true,
})

// Migration file: migrations/20240201_create_users_table.go
package migrations

import "go.oease.dev/goe/v2/migrate"

func init() {
    migrate.Register(&CreateUsersTable{})
}

type CreateUsersTable struct{}

func (m *CreateUsersTable) Version() string {
    return "20240201_create_users_table"
}

func (m *CreateUsersTable) Up(db *gorm.DB) error {
    return db.AutoMigrate(&User{})
}

func (m *CreateUsersTable) Down(db *gorm.DB) error {
    return db.Migrator().DropTable("users")
}

// SQL migrations also supported
// migrations/20240202_add_email_index.sql
// -- +migrate Up
// CREATE INDEX idx_users_email ON users(email);
//
// -- +migrate Down
// DROP INDEX idx_users_email;

// CLI commands
// goe migrate              # Run pending migrations
// goe migrate:rollback     # Rollback last batch
// goe migrate:reset        # Rollback all
// goe migrate:status       # Show migration status
// goe migrate:make name    # Create new migration file

// Programmatic usage
migrate.Run()                    // Run pending
migrate.Rollback()               // Rollback last batch
migrate.RollbackSteps(3)         // Rollback N migrations
migrate.Reset()                  // Rollback all
migrate.Status() []MigrationInfo // Get status
```

---

### CLI Module (`goe/cli`)

**Goals:**
- Built-in essential commands
- Easy custom command registration
- Code generators
- Interactive mode support

**Proposed API:**

```go
// Register custom commands
goe.New(goe.Options{
    Commands: []cli.Command{
        {
            Name:        "seed",
            Description: "Seed the database with test data",
            Handler:     seedDatabase,
        },
        {
            Name:        "cleanup",
            Description: "Clean up old records",
            Flags: []cli.Flag{
                {Name: "days", Default: "30", Usage: "Records older than N days"},
            },
            Handler:     cleanupOldData,
        },
    },
})

func seedDatabase(ctx *cli.Context) error {
    db := ctx.DB()
    // seed logic
    return nil
}

func cleanupOldData(ctx *cli.Context) error {
    days := ctx.Int("days")
    // cleanup logic
    return nil
}

// Built-in commands:
// goe serve              # Start the server
// goe migrate            # Run migrations
// goe migrate:rollback   # Rollback migrations
// goe seed               # Run seeders
// goe make:handler name  # Generate handler
// goe make:model name    # Generate model
// goe make:migration name # Generate migration
// goe routes             # List all routes
// goe env                # Show environment info
```

---

### Testing Module (`goe/testing`)

**Goals:**
- Make testing GOE apps easy
- Provide mocks for all contracts
- HTTP testing helpers
- Database test utilities

**Proposed API:**

```go
import (
    "testing"
    goetest "go.oease.dev/goe/v2/testing"
)

func TestCreateUser(t *testing.T) {
    // Create test app with isolated database
    app := goetest.NewTestApp(t, goetest.Config{
        WithHTTP: true,
        WithDB:   true,
    })
    defer app.Cleanup()

    // Make HTTP request
    resp := app.Post("/users").
        WithJSON(map[string]any{
            "name":  "John",
            "email": "john@example.com",
        }).
        WithHeader("Authorization", "Bearer "+app.AuthToken("admin")).
        Execute()

    // Assertions
    resp.AssertStatus(201)
    resp.AssertJSON("name", "John")
    resp.AssertJSONPath("$.data.id").Exists()

    // Database assertions
    app.AssertDatabaseHas("users", map[string]any{
        "email": "john@example.com",
    })
}

func TestWithMocks(t *testing.T) {
    // Use mock implementations
    mockCache := goetest.NewMockCache()
    mockCache.OnGet("user:123").Return(&User{ID: "123"}, nil)

    app := goetest.NewTestApp(t, goetest.Config{
        Mocks: map[string]any{
            "cache": mockCache,
        },
    })

    // Test with mocked cache
    resp := app.Get("/users/123").Execute()
    resp.AssertStatus(200)

    // Verify mock was called
    mockCache.AssertCalled(t, "Get", "user:123")
}

func TestDatabaseTransaction(t *testing.T) {
    // Each test runs in a transaction that's rolled back
    app := goetest.NewTestApp(t, goetest.Config{
        WithDB:            true,
        TransactionPerTest: true, // Auto-rollback after each test
    })

    // Create test data
    user := app.Factory().User().Create()

    // Test
    resp := app.Get("/users/" + user.ID).Execute()
    resp.AssertStatus(200)

    // Transaction automatically rolled back - database is clean
}
```

---

### Security Module (`goe/security`)

**Goals:**
- One-line security hardening
- Sensible defaults for web apps
- Configurable rate limiting

**Proposed API:**

```go
// Enable security module
goe.New(goe.Options{
    WithSecurity: true,
})

// Configuration (.env)
// SECURITY_CORS_ORIGINS=https://example.com,https://app.example.com
// SECURITY_CORS_CREDENTIALS=true
// SECURITY_RATE_LIMIT=100
// SECURITY_RATE_WINDOW=1m
// SECURITY_CSRF_ENABLED=true

// Or configure programmatically
app.Use(security.New(security.Config{
    // CORS
    CORS: security.CORSConfig{
        AllowOrigins:     []string{"https://example.com"},
        AllowCredentials: true,
    },

    // Rate limiting
    RateLimit: security.RateLimitConfig{
        Max:      100,
        Window:   time.Minute,
        KeyFunc:  func(c fiber.Ctx) string { return c.IP() },
        Storage:  redisStorage, // Distributed rate limiting
    },

    // Security headers
    Headers: security.HeadersConfig{
        XFrameOptions:        "DENY",
        XContentTypeOptions:  "nosniff",
        XSSProtection:        "1; mode=block",
        ContentSecurityPolicy: "default-src 'self'",
        StrictTransportSecurity: "max-age=31536000; includeSubDomains",
    },

    // CSRF
    CSRF: security.CSRFConfig{
        Enabled:    true,
        CookieName: "_csrf",
        HeaderName: "X-CSRF-Token",
    },
}))

// Or use presets
app.Use(security.Strict())  // Maximum security
app.Use(security.API())     // For APIs (no CSRF, strict CORS)
app.Use(security.Web())     // For web apps (CSRF, relaxed headers)
```

---

## Design Principles

### 1. Batteries-Included but Swappable
Every module should:
- Work out of the box with zero config
- Be fully replaceable via contracts
- Have sensible, secure defaults

### 2. Contract-First
- Define interfaces before implementations
- Allow multiple implementations per contract
- Enable easy mocking for tests

### 3. Configuration Hierarchy
```
Defaults → .env → .local.env → .{env}.env → System ENV → Code
```

### 4. Fail Fast, Fail Loud
- Validate configuration at startup
- Clear error messages
- No silent failures

### 5. Observable by Default
- All modules emit metrics
- Structured logging throughout
- Trace propagation

---

## Contributing

If you'd like to contribute to any of these modules:

1. Open an issue to discuss the design
2. Follow the existing code patterns
3. Include tests and documentation
4. Submit a PR against the `v2` branch

---

## References

- [Uber Fx Documentation](https://uber-go.github.io/fx/)
- [GoFiber v3 Documentation](https://docs.gofiber.io/)
- [GORM Documentation](https://gorm.io/docs/)
- [Redis Streams](https://redis.io/docs/data-types/streams/)

---

*Last updated: February 2026*
