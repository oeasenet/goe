# Environment Variables Reference

This document lists all configuration environment variables recognized by **Goe**.
The defaults shown are the values used when a variable is unset.

| Variable | Description | Default |
|----------|-------------|---------|
| `GOE_ENV` | Application environment | `dev` |
| `APP_NAME` | Application name | `"Goe Application"` |
| `APP_VERSION` | Application version | `"1.0.0"` |
| `DEBUG` | Enable debug mode | `false` |
| `HTTP_HOST` | HTTP server host | `"0.0.0.0"` |
| `HTTP_PORT` | HTTP server port | `8080` |
| `HTTP_READ_TIMEOUT` | HTTP read timeout | `10s` |
| `HTTP_WRITE_TIMEOUT` | HTTP write timeout | `10s` |
| `HTTP_IDLE_TIMEOUT` | HTTP idle timeout | `30s` |
| `LOG_LEVEL` | Logging level | `"info"` |
| `LOG_FORMAT` | Logging format | `"text"` |
| `LOG_OUTPUT` | Logger outputs | `"console"` |
| `LOG_CALLER` | Include caller info | `false` |
| `LOG_STACKTRACE` | Include stack traces | `false` |
| `FIBER_SERVER_HEADER` | HTTP server header | `"Goe"` |
| `FIBER_STRICT_ROUTING` | Enable strict routing | `false` |
| `FIBER_CASE_SENSITIVE` | Case sensitive routing | `false` |
| `FIBER_IMMUTABLE` | Immutable mode | `false` |
| `FIBER_UNESCAPE_PATH` | Unescape path values | `false` |
| `FIBER_BODY_LIMIT` | Max body size in bytes | `4194304` |
| `FIBER_STREAM_REQUEST_BODY` | Stream request body | `true` |
| `FIBER_CONCURRENCY` | Maximum concurrent connections | `262144` |
| `FIBER_REDUCE_MEMORY` | Reduce memory usage | `false` |
| `FIBER_ENABLE_IP_VALIDATION` | Enable IP validation | `false` |
| `FIBER_TRUST_PROXY` | Trust proxy headers | `false` |
| `FIBER_PROXY_HEADER` | Custom proxy header | `"X-Forwarded-For"` |
| `FIBER_TRUST_PROXIES` | Trusted proxy IPs or CIDRs | *(none)* |
| `FIBER_TRUST_LINK_LOCAL` | Trust link-local addresses | `true` |
| `FIBER_TRUST_LOOPBACK` | Trust loopback addresses | `true` |
| `FIBER_TRUST_PRIVATE` | Trust private addresses | `true` |
| `CACHE_DRIVER` | Default cache driver | `memory` |
| `CACHE_PREFIX` | Cache key prefix | value of `APP_NAME` |
| `CACHE_TTL` | Default TTL for cache entries | `2h` |
| `CACHE_MEMORY_GC_INTERVAL` | Memory driver GC interval | `10s` |
| `CACHE_REDIS_URL` | Redis connection URL | *(none)* |
| `CACHE_REDIS_HOSTS` | Redis hosts | `localhost:6379` |
| `CACHE_REDIS_USERNAME` | Redis username | *(none)* |
| `CACHE_REDIS_PASSWORD` | Redis password | *(none)* |
| `CACHE_REDIS_DATABASE` | Redis database index | `0` |
| `CACHE_REDIS_CLIENT_NAME` | Redis client name | *(none)* |
| `CACHE_REDIS_CACHE_SIZE` | Client-side cache size | `134217728` |
| `CACHE_REDIS_BLOCKING_POOL_SIZE` | Blocking command pool size | `1000` |
| `CACHE_REDIS_PIPELINE_MULTIPLEX` | Pipelining connections | `2` |
| `CACHE_REDIS_DISABLE_RETRY` | Disable retry on network errors | `false` |
| `CACHE_REDIS_DISABLE_CACHE` | Disable client-side caching | `false` |
| `CACHE_REDIS_ALWAYS_PIPELINING` | Always pipeline commands | `true` |
| `CACHE_REDIS_CACHE_TTL` | Client-side cache TTL | `1m` |
| `CACHE_SQLITE_DATABASE` | SQLite database path | `./cache.db` |
| `CACHE_SQLITE_TABLE` | SQLite table name | `cache` |
| `CACHE_STORE` | Default cache store name | `default` |
| `CACHE_<name>_DRIVER` | Driver for named cache store | inherits `CACHE_DRIVER` |
| `CACHE_<name>_PREFIX` | Prefix for named store | inherits `CACHE_PREFIX` |
| `CACHE_<name>_TTL` | TTL for named store | inherits `CACHE_TTL` |

The table above consolidates settings documented throughout the repository. The original environment variable table can also be found in [docs/API.md](API.md).
