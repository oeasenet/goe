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

### Basic Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `HTTP_HOST` | string | `0.0.0.0` | HTTP server bind address |
| `HTTP_PORT` | int | `8080` | HTTP server port |
| `HTTP_READ_TIMEOUT` | duration | `10s` | Maximum duration for reading the entire request |
| `HTTP_WRITE_TIMEOUT` | duration | `10s` | Maximum duration for writing the response |
| `HTTP_IDLE_TIMEOUT` | duration | `30s` | Maximum idle time for keep-alive connections |

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
| `LOG_LEVEL` | string | `info` | Log level: `debug`, `info`, `warn`, `error`, `panic`, `fatal` |
| `LOG_FORMAT` | string | `text` | Log format: `text` (console) or `json` |
| `LOG_OUTPUT` | []string | `console` | Comma-separated outputs: `console`, `stdout`, `stderr`, or file path |
| `LOG_CALLER` | bool | `false` | Include caller information in logs |
| `LOG_STACKTRACE` | bool | `false` | Include stack traces for error logs |

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
| `DB_LOG_MODE` | bool | `false` | Enable GORM SQL logging |
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
| `CACHE_STORE` | string | `memory` | Cache store type: `memory`, `redis`, `memcache`, `badger`, `sqlite3`, `postgres`, `mysql`, `mongodb`, `dynamodb`, `s3` |
| `CACHE_DRIVER` | string | - | Alias for `CACHE_STORE` |
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
| `CACHE_REDIS_DB` | int | - | Alias for `CACHE_REDIS_DATABASE` |
| `CACHE_REDIS_CLIENT_NAME` | string | - | Client name for Redis connection |
| `CACHE_REDIS_POOL_SIZE` | int | - | Connection pool size |
| `CACHE_REDIS_RESET` | bool | `false` | Reset (flush) Redis database on startup |

#### Redis Cluster/Sentinel

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CACHE_REDIS_ADDRS` | []string | - | Comma-separated cluster node addresses |
| `CACHE_REDIS_MASTER_NAME` | string | - | Sentinel master name |
| `CACHE_REDIS_IS_CLUSTER_MODE` | bool | `false` | Enable Redis Cluster mode |

### Database-Backed Stores

#### PostgreSQL/MySQL

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CACHE_DB_HOST` | string | **required** | Database host |
| `CACHE_DB_PORT` | int | driver-specific | Database port |
| `CACHE_DB_DATABASE` | string | **required** | Database name |
| `CACHE_DB_USERNAME` | string | **required** | Database username |
| `CACHE_DB_PASSWORD` | string | - | Database password |

#### Memcache

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CACHE_MEMCACHE_SERVERS` | string | **required** | Comma-separated server addresses |

#### MongoDB

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CACHE_MONGODB_URI` | string | **required** | MongoDB connection URI |
| `CACHE_MONGODB_DATABASE` | string | **required** | Database name |

#### DynamoDB

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CACHE_DYNAMODB_TABLE` | string | **required** | DynamoDB table name |
| `CACHE_DYNAMODB_REGION` | string | **required** | AWS region |

#### S3

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CACHE_S3_BUCKET` | string | **required** | S3 bucket name |
| `CACHE_S3_REGION` | string | **required** | AWS region |

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
| `LOCK_REDIS_URL` | string | `redis://localhost:6379/0` | Primary Redis connection URL |
| `LOCK_REDIS_URLS` | []string | - | Multiple URLs for Redlock algorithm |
| `LOCK_REDIS_ADDR` | string | - | Redis address (fallback) |
| `LOCK_REDIS_HOST` | string | - | Redis host (fallback) |
| `LOCK_REDIS_HOSTS` | string | - | Redis hosts (fallback) |
| `LOCK_REDIS_DB` | int | - | Redis database number |
| `LOCK_REDIS_POOL_SIZE` | int | - | Alias for `LOCK_POOL_SIZE` |

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
