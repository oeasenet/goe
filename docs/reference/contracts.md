# Contracts

GOE Framework uses interface-based contracts to ensure loose coupling and easy testing. This page documents the core contracts and their usage.

## Core Contracts

### Logger Contract
```go
type Logger interface {
    Info(msg string, fields ...zap.Field)
    Debug(msg string, fields ...zap.Field)
    Warn(msg string, fields ...zap.Field)
    Error(msg string, fields ...zap.Field)
    Fatal(msg string, fields ...zap.Field)
}
```

### Database Contract
```go
type Database interface {
    Query(query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(query string, args ...interface{}) *sql.Row
    Exec(query string, args ...interface{}) (sql.Result, error)
    Begin() (*sql.Tx, error)
    Close() error
}
```

### Cache Contract
```go
type Cache interface {
    Get(key string) ([]byte, error)
    Set(key string, value []byte, expiration time.Duration) error
    Delete(key string) error
    Clear() error
}
```

### Config Contract
```go
type Config interface {
    Get(key string) interface{}
    GetString(key string) string
    GetInt(key string) int
    GetBool(key string) bool
    Set(key string, value interface{})
}
```

## Usage Examples

### Using Contracts in Services
```go
type UserService struct {
    db     Database
    cache  Cache
    logger Logger
}

func NewUserService(db Database, cache Cache, logger Logger) *UserService {
    return &UserService{
        db:     db,
        cache:  cache,
        logger: logger,
    }
}
```

### Testing with Contracts
```go
func TestUserService(t *testing.T) {
    mockDB := &mocks.Database{}
    mockCache := &mocks.Cache{}
    mockLogger := &mocks.Logger{}
    
    service := NewUserService(mockDB, mockCache, mockLogger)
    
    // Test implementation
}
```

## Contract Implementation

Contracts are implemented by concrete types:

```go
type ZapLogger struct {
    logger *zap.Logger
}

func (l *ZapLogger) Info(msg string, fields ...zap.Field) {
    l.logger.Info(msg, fields...)
}

// Implement other methods...
```