# GOE Framework Testing Guide 🧪

## Overview

This comprehensive guide covers testing practices, patterns, and procedures specifically for the GOE framework. It's designed for both AI agents and human developers to understand how testing works in GOE, what the best practices are, and how to effectively contribute to the framework's test suite.

## Table of Contents

1. [Quick Start](#quick-start)
2. [GOE Testing Architecture](#goe-testing-architecture)
3. [Running Tests](#running-tests)
4. [Test Categories](#test-categories)
5. [GOE-Specific Testing Patterns](#goe-specific-testing-patterns)
6. [Best Practices for GOE](#best-practices-for-goe)
7. [Coverage Analysis](#coverage-analysis)
8. [Troubleshooting](#troubleshooting)
9. [Contributing Tests to GOE](#contributing-tests-to-goe)
10. [AI Agent Guidelines](#ai-agent-guidelines)

## Quick Start

### Prerequisites
- Go 1.21+ installed
- GOE framework dependencies (`go mod download`)
- Understanding of GOE's dependency injection (Fx-based)

### Run All Tests
```bash
make test
```

### Run Specific Test Categories
```bash
make test_contract     # Interface/contract tests (100% coverage)
make test_core        # Core implementation tests (~73% coverage)
make test_integration # Integration tests (~50% coverage)
```

### Run with Coverage
```bash
make test_coverage
```

## GOE Testing Architecture

The GOE framework uses a three-tier testing architecture that aligns with its modular design:

```
tests/
├── contract/          # Interface compliance tests (GOE contracts)
├── core/             # Implementation unit tests (GOE core modules)
├── integration/      # Cross-component integration tests
├── middlewares/      # HTTP middleware tests
└── utils/           # Utility function tests
```

### GOE Test Organization Principles

1. **Contract-First Testing**: GOE's interface-based design requires testing contracts before implementations
2. **Module Isolation**: Each GOE module (cache, db, http, log, etc.) is tested independently
3. **Dependency Injection Testing**: Tests use GOE's Fx-based DI for realistic scenarios
4. **Integration Validation**: Tests verify GOE component interactions work correctly

## Running Tests

### Makefile Commands (GOE-Specific)

| Command | Purpose | Coverage | GOE Components |
|---------|---------|----------|----------------|
| `make test` | Run all tests | Combined | All GOE modules |
| `make test_contract` | Test GOE interfaces | 100% | contract.* interfaces |
| `make test_core` | Test GOE implementations | ~73% | core.* packages |
| `make test_integration` | Test GOE integrations | ~50% | Cross-module scenarios |
| `make test_coverage` | Generate coverage report | All packages | Full GOE framework |
| `make clean` | Remove coverage files | N/A | Cleanup |

### GOE Test Flags

All GOE tests run with these flags for reliability:
- `-v`: Verbose output (important for debugging GOE's complex DI)
- `-race`: Race condition detection (critical for GOE's concurrent operations)
- `-coverpkg`: Coverage calculation across GOE packages

### Manual Test Execution for GOE

```bash
# Test specific GOE modules
go test -v ./tests/contract/cache_test.go    # Cache contract
go test -v ./tests/core/cache/...           # Cache implementation
go test -v ./tests/integration/...          # GOE integrations

# Test GOE with race detection (recommended)
go test -race ./tests/...

# Test specific GOE functionality
go test -v -run TestCacheInterface ./tests/contract/...
go test -v -run TestAppIntegration ./tests/integration/...
```

## Test Categories

### 1. Contract Tests (`tests/contract/`)

**Purpose**: Verify GOE interface compliance and contract behavior

**Coverage**: 100% (GOE interfaces only)

**GOE-Specific Files**:
- `app_test.go` - GOE Application interface tests
- `cache_test.go` - GOE Cache interface tests (contract.Cache, contract.CacheManager)
- `config_test.go` - GOE Configuration interface tests
- `db_test.go` - GOE Database interface tests
- `http_test.go` - GOE HTTP interface tests
- `log_test.go` - GOE Logging interface tests (contract.Logger)
- `module_test.go` - GOE Module interface tests (contract.Module)
- `observability_test.go` - GOE Observability interface tests

**GOE Contract Testing Pattern**:
```go
func TestGOECacheInterface(t *testing.T) {
    // Verify GOE interface compliance
    var _ contract.Cache = (*MockCache)(nil)
    
    // Create mock with GOE-specific expectations
    cache := new(MockCache)
    cache.On("Get", "goe:key").Return("goe:value", nil)
    cache.On("Set", "goe:key", "goe:value", time.Minute).Return(nil)
    
    // Test GOE interface methods
    value, err := cache.Get("goe:key")
    assert.NoError(t, err)
    assert.Equal(t, "goe:value", value)
    
    // Verify GOE expectations
    cache.AssertExpectations(t)
}
```

### 2. Core Tests (`tests/core/`)

**Purpose**: Test GOE concrete implementations and business logic

**Coverage**: ~73% (GOE implementation code)

**GOE Core Structure**:
```
tests/core/
├── app/           # GOE Application core tests
├── cache/         # GOE Cache implementation tests
├── config/        # GOE Configuration tests
├── db/           # GOE Database tests
├── log/          # GOE Logging tests
└── otel/         # GOE Observability tests
```

**GOE Implementation Testing Pattern**:
```go
func TestGOECacheImplementation(t *testing.T) {
    // Setup GOE mocks
    store := new(MockCacheStore)
    store.On("Get", "goe:prefix:key").Return([]byte(`"value"`), nil)
    store.On("Set", "goe:prefix:key", mock.Anything, time.Minute).Return(nil)
    
    // Create real GOE implementation
    cache := cache.New(store, "goe:prefix")
    
    // Test actual GOE behavior
    err := cache.Set("key", "value", time.Minute)
    assert.NoError(t, err)
    
    value, err := cache.Get("key")
    assert.NoError(t, err)
    assert.Equal(t, "value", value)
    
    store.AssertExpectations(t)
}
```

### 3. Integration Tests (`tests/integration/`)

**Purpose**: Test GOE cross-component interactions and real scenarios

**Coverage**: ~50% (GOE integration scenarios)

**GOE Integration Files**:
- `metrics_integration_test.go` - GOE Metrics collection integration
- `module_integration_test.go` - GOE Module lifecycle integration
- `observability_integration_test.go` - GOE Observability integration
- `tracing_integration_test.go` - GOE Distributed tracing integration

**GOE Integration Testing Pattern**:
```go
func TestGOEModuleIntegration(t *testing.T) {
    // Create real GOE application
    app := app.New("Test GOE App", "1.0.0", "test")
    
    // Add real GOE modules
    module := NewTestModule("goe-test")
    err := app.AddModule(module)
    assert.NoError(t, err)
    
    // Test real GOE lifecycle
    ctx := context.Background()
    err = app.Start(ctx)
    assert.NoError(t, err)
    assert.True(t, app.IsRunning())
    
    // Verify GOE integration behavior
    assert.True(t, module.startCalled)
    
    err = app.Stop(ctx)
    assert.NoError(t, err)
    assert.True(t, module.stopCalled)
}
```

### 4. Middleware Tests (`tests/middlewares/`)

**Purpose**: Test GOE HTTP middleware components

**GOE Middleware Files**:
- `cdn_cache_control_test.go` - GOE CDN cache control middleware
- `spa_test.go` - GOE SPA serving middleware

### 5. Utility Tests (`tests/utils/`)

**Purpose**: Test GOE utility functions

**GOE Utility Files**:
- `strings_test.go` - GOE string utility tests
- `human_readable_sizes_test.go` - GOE size conversion utility tests

## GOE-Specific Testing Patterns

### 1. GOE Mock Creation Pattern

```go
// GOE-style mock with contract compliance
type MockGOEService struct {
    mock.Mock
}

func (m *MockGOEService) Method(param string) (string, error) {
    args := m.Called(param)
    return args.String(0), args.Error(1)
}

// Ensure GOE contract compliance
var _ contract.Service = (*MockGOEService)(nil)
```

### 2. GOE Interface Compliance Testing

```go
func TestGOEInterfaceCompliance(t *testing.T) {
    // Test that implementation satisfies GOE contract
    var _ contract.Cache = (*cache.Cache)(nil)
    var _ contract.Logger = (*log.Logger)(nil)
    var _ contract.Module = (*MyGOEModule)(nil)
}
```

### 3. GOE Dependency Injection Testing

```go
func TestGOEDependencyInjection(t *testing.T) {
    // Create GOE app with test configuration
    app := app.New("Test", "1.0.0", "test")
    
    // Add GOE providers
    err := app.AddProvider(func() contract.Logger {
        return &MockLogger{}
    })
    assert.NoError(t, err)
    
    // Test GOE DI resolution
    var logger contract.Logger
    err = app.AddInvoker(func(l contract.Logger) {
        logger = l
    })
    assert.NoError(t, err)
    
    // Start GOE app to trigger DI
    ctx := context.Background()
    err = app.Start(ctx)
    assert.NoError(t, err)
    defer app.Stop(ctx)
    
    assert.NotNil(t, logger)
}
```

### 4. GOE HTTP Testing Pattern

```go
func TestGOEHTTPMiddleware(t *testing.T) {
    // Create Fiber app (GOE's HTTP framework)
    app := fiber.New()
    
    // Apply GOE middleware
    app.Use(middlewares.NewCDNCacheControlMiddleware().Handler())
    
    // Add test route
    app.Get("/test", func(c fiber.Ctx) error {
        return c.SendString("GOE response")
    })
    
    // Test GOE HTTP behavior
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    resp, err := app.Test(req)
    
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    resp.Body.Close()
}
```

### 5. GOE Module Testing Pattern

```go
func TestGOEModule(t *testing.T) {
    // Create GOE module
    module := &MyGOEModule{name: "test-module"}
    
    // Test GOE module interface compliance
    var _ contract.Module = module
    
    // Test GOE module lifecycle
    ctx := context.Background()
    
    err := module.OnStart(ctx)
    assert.NoError(t, err)
    
    err = module.OnStop(ctx)
    assert.NoError(t, err)
    
    assert.Equal(t, "test-module", module.Name())
}
```

## Best Practices for GOE

### 1. GOE Test Organization

- **Follow GOE structure**: Mirror the GOE package structure in tests
- **Use GOE naming**: Prefix test files with the GOE component name
- **GOE package isolation**: Use `_test` package suffix for integration tests
- **GOE module grouping**: Group tests by GOE module (cache, db, http, etc.)

### 2. GOE Mock Management

- **Use GOE contracts**: Always mock GOE interfaces, not implementations
- **GOE expectation verification**: Always call `AssertExpectations(t)` for GOE mocks
- **GOE dependency mocking**: Mock external dependencies, test GOE logic
- **GOE mock reset**: Reset mocks between GOE test cases

### 3. GOE Test Data

- **Use GOE-realistic data**: Use data that reflects real GOE usage
- **GOE configuration**: Use test-specific GOE configuration
- **GOE isolation**: Ensure GOE tests don't interfere with each other
- **GOE factories**: Create factories for complex GOE objects

### 4. GOE Assertions

- **GOE-specific assertions**: Test GOE-specific behavior and contracts
- **GOE error handling**: Test GOE error conditions and edge cases
- **GOE state verification**: Verify GOE component state changes
- **GOE integration checks**: Verify GOE component interactions

### 5. GOE Performance

- **Avoid real GOE I/O**: Use mocks for GOE database, HTTP calls
- **GOE parallel tests**: Use `t.Parallel()` for independent GOE tests
- **GOE timeouts**: Use timeout contexts for GOE integration tests
- **GOE benchmarks**: Add benchmarks for performance-critical GOE code

## Coverage Analysis

### Current GOE Coverage Status

| GOE Test Category | Coverage | GOE Components |
|-------------------|----------|----------------|
| Contract | 100% | All GOE interfaces |
| Core | 73% | GOE implementations |
| Integration | 50% | GOE cross-component scenarios |
| Middlewares | Added | GOE HTTP middlewares |
| Utils | Added | GOE utility functions |

### GOE Coverage Gaps

**Untested GOE Packages**:
- `webresult/` - GOE web response utilities
- Additional middleware components
- Some utility functions

**GOE Improvement Opportunities**:
1. Increase GOE core implementation coverage to 85%+
2. Add more GOE integration scenarios
3. Add GOE performance benchmarks
4. Add GOE end-to-end tests

### Generating GOE Coverage Reports

```bash
# Generate GOE coverage file
make test_coverage

# View GOE coverage in browser
go tool cover -html=coverage.txt

# View GOE coverage summary
go tool cover -func=coverage.txt | grep -E "(contract|core|integration)"
```

## Troubleshooting

### Common GOE Testing Issues

1. **GOE Dependency Injection Failures**
   ```
   Error: failed to build dependency graph
   ```
   - **Solution**: Ensure all GOE dependencies are properly provided in test setup

2. **GOE Mock Expectation Failures**
   ```
   mock: Unexpected Method Call
   ```
   - **Solution**: Verify GOE mock setup matches actual GOE interface calls

3. **GOE Race Condition Warnings**
   ```
   WARNING: DATA RACE in GOE component
   ```
   - **Solution**: Fix shared state access in GOE components or use proper synchronization

4. **GOE Context Timeout Issues**
   ```
   context deadline exceeded in GOE operation
   ```
   - **Solution**: Increase timeout for GOE operations or optimize test performance

### GOE Debug Techniques

1. **GOE Verbose Output**: Use `-v` flag for detailed GOE test output
2. **GOE Specific Tests**: Run individual GOE tests with `-run` flag
3. **GOE Race Detection**: Always use `-race` flag for GOE tests
4. **GOE Logging**: Add debug logs in GOE tests (remove before commit)

## Contributing Tests to GOE

### Adding New GOE Tests

1. **Choose appropriate GOE category**:
   - Contract: Testing GOE interfaces
   - Core: Testing GOE implementations
   - Integration: Testing GOE component interactions

2. **Follow GOE naming conventions**:
   - File: `goe_feature_test.go`
   - Function: `TestGOEFeatureBehavior`
   - Mock: `MockGOEInterface`

3. **Include GOE test cases**:
   - Happy path for GOE functionality
   - GOE error conditions
   - GOE edge cases
   - GOE boundary conditions

### GOE Test Review Checklist

- [ ] Tests follow GOE structure and patterns
- [ ] GOE mocks are properly configured
- [ ] GOE expectations are verified
- [ ] GOE error cases are tested
- [ ] Tests are independent of other GOE tests
- [ ] GOE coverage is maintained/improved
- [ ] No race conditions in GOE components
- [ ] Descriptive GOE test names
- [ ] GOE contracts are properly tested

### Example: Adding GOE Cache Driver Test

```go
// tests/core/cache/redis_driver_test.go
package cache_test

import (
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "go.oease.dev/goe/v2/core/cache"
    "go.oease.dev/goe/v2/contract"
)

func TestGOERedisDriver(t *testing.T) {
    // Setup GOE Redis driver
    driver := cache.NewRedisDriver("localhost:6379")
    
    // Ensure GOE contract compliance
    var _ contract.CacheStore = driver
    
    // Test GOE connection
    err := driver.Connect()
    assert.NoError(t, err)
    defer driver.Close()
    
    // Test GOE operations
    err = driver.Set("goe:test:key", []byte("goe:value"), time.Minute)
    assert.NoError(t, err)
    
    value, err := driver.Get("goe:test:key")
    assert.NoError(t, err)
    assert.Equal(t, []byte("goe:value"), value)
    
    // Cleanup GOE test data
    err = driver.Delete("goe:test:key")
    assert.NoError(t, err)
}
```

## AI Agent Guidelines

### For AI Agents Working with GOE Tests

1. **Understand GOE Architecture**: 
   - GOE uses dependency injection (Fx)
   - GOE follows contract-first design
   - GOE has modular architecture (cache, db, http, log, etc.)

2. **GOE Testing Priorities**:
   - Always test GOE contracts first
   - Mock GOE dependencies, not implementations
   - Use GOE-specific patterns and conventions
   - Follow GOE's three-tier testing approach

3. **GOE Test Creation Guidelines**:
   - Start with contract tests for new GOE interfaces
   - Add implementation tests for GOE core functionality
   - Include integration tests for GOE component interactions
   - Use GOE-realistic test data and scenarios

4. **GOE Code Quality**:
   - Ensure 100% coverage for GOE contracts
   - Aim for 80%+ coverage for GOE implementations
   - Include error handling for all GOE operations
   - Add benchmarks for performance-critical GOE code

5. **GOE Test Maintenance**:
   - Keep GOE tests independent and isolated
   - Update tests when GOE interfaces change
   - Maintain GOE test documentation
   - Follow GOE naming and organization conventions

### AI Agent Quick Reference

**GOE Test Commands**:
```bash
make test              # All GOE tests
make test_contract     # GOE interface tests
make test_core        # GOE implementation tests
make test_integration # GOE integration tests
```

**GOE Test Patterns**:
- Contract: `var _ contract.Interface = (*Implementation)(nil)`
- Mock: `mock.On("Method", "param").Return("result", nil)`
- Assert: `assert.NoError(t, err)` and `mock.AssertExpectations(t)`
- Integration: Use real GOE app with `app.New()` and lifecycle methods

**GOE Test Structure**:
```
tests/
├── contract/     # GOE interface compliance
├── core/        # GOE implementation testing
├── integration/ # GOE component interactions
├── middlewares/ # GOE HTTP middleware
└── utils/       # GOE utility functions
```

## Recommended GOE Improvements

### 1. Missing GOE Test Coverage

**Add tests for**:
- `webresult/` - GOE web response utilities
- Additional GOE middleware components
- More GOE utility functions
- GOE error handling edge cases

### 2. GOE Test Infrastructure

**Create GOE test helpers**:
```go
// tests/helpers/goe_app.go
func NewTestGOEApp() *app.Application {
    return app.New("Test GOE", "1.0.0", "test")
}

// tests/helpers/goe_mocks.go
func NewMockGOELogger() contract.Logger {
    return &MockLogger{}
}
```

### 3. GOE Environment-Specific Tests

**Add GOE test configurations**:
```env
# .env.test
GOE_ENV=test
LOG_LEVEL=error
DB_DRIVER=sqlite
DB_DATABASE=:memory:
CACHE_DRIVER=memory
```

### 4. GOE Benchmark Tests

**Add GOE performance tests**:
```go
func BenchmarkGOECacheGet(b *testing.B) {
    cache := setupGOECache()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        cache.Get("goe:benchmark:key")
    }
}
```

## Conclusion

The GOE framework has a comprehensive testing foundation that supports its modular, contract-first architecture:

✅ **Complete GOE Documentation**: Detailed testing guide for GOE developers and AI agents
✅ **GOE Interface Testing**: 100% coverage of GOE contracts
✅ **GOE Implementation Testing**: Good coverage of GOE core functionality
✅ **GOE Integration Testing**: Real-world GOE scenario coverage
✅ **GOE Quality Patterns**: Established GOE-specific best practices
✅ **GOE Test Infrastructure**: Organized, maintainable GOE test structure

**GOE Areas for improvement**:
- Add missing GOE package tests (webresult, additional middlewares)
- Increase GOE core implementation coverage
- Add more GOE integration scenarios
- Create GOE test helper utilities
- Add GOE benchmark tests

This guide provides the foundation for maintaining and extending the GOE test suite while ensuring high code quality and reliability as the GOE framework evolves.

**GOE Testing Summary**:
- **Total GOE Test Files**: 20+ test files
- **GOE Coverage**: Contract (100%), Core (73%), Integration (50%)
- **GOE Test Lines**: 1,500+ lines of comprehensive testing
- **GOE Components Tested**: All major GOE modules (cache, db, http, log, etc.)

This comprehensive GOE testing enhancement ensures the framework maintains high quality standards and provides clear guidance for future GOE development and testing efforts.