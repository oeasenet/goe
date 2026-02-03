package goe

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
)

func TestCircularDependencies(t *testing.T) {
	t.Run("no circular dependencies in core modules", func(t *testing.T) {
		resetGlobalInstance()

		// Test with all modules enabled to check for circular dependencies
		app := New(Options{
			WithHTTP:    true,
			WithCache:   true,
			WithDB:      false, // Skip DB to avoid connection requirements
			WithMongoDB: false, // Skip MongoDB to avoid connection requirements
		})

		require.NotNil(t, app, "Application should be created without circular dependency issues")

		// Verify all services are accessible
		assert.NotNil(t, Config())
		assert.NotNil(t, Log())
		assert.NotNil(t, HTTP())
		assert.NotNil(t, Cache())
	})

	t.Run("complex provider dependency chain", func(t *testing.T) {
		resetGlobalInstance()

		// Test a complex dependency chain to ensure no circular dependencies
		type ServiceA struct{ Value string }
		type ServiceB struct{ Value string }
		type ServiceC struct{ Value string }

		var serviceA ServiceA
		var serviceB ServiceB
		var serviceC ServiceC

		app := New(Options{
			WithHTTP:  true,
			WithCache: true,
			Providers: []any{
				// ServiceA depends on config and logger (framework dependencies)
				func(config contract.Config, logger contract.Logger) ServiceA {
					logger.Debug("ServiceA created")
					return ServiceA{Value: "serviceA"}
				},
				// ServiceB depends on ServiceA and cache
				func(serviceA ServiceA, cache contract.Cache) ServiceB {
					return ServiceB{Value: "serviceB-depends-on-" + serviceA.Value}
				},
				// ServiceC depends on ServiceA and ServiceB
				func(serviceA ServiceA, serviceB ServiceB) ServiceC {
					return ServiceC{Value: "serviceC-depends-on-" + serviceA.Value + "-and-" + serviceB.Value}
				},
			},
			Invokers: []any{
				func(sA ServiceA, sB ServiceB, sC ServiceC) {
					serviceA = sA
					serviceB = sB
					serviceC = sC
				},
			},
		})

		require.NotNil(t, app)
		assert.Equal(t, "serviceA", serviceA.Value)
		assert.Equal(t, "serviceB-depends-on-serviceA", serviceB.Value)
		assert.Equal(t, "serviceC-depends-on-serviceA-and-serviceB-depends-on-serviceA", serviceC.Value)
	})

	t.Run("module dependency resolution order", func(t *testing.T) {
		resetGlobalInstance()

		var initOrder []string

		// Test module initialization order to ensure proper dependency resolution
		app := New(Options{
			WithHTTP:  true,
			WithCache: true,
			Invokers: []any{
				// This invoker depends on services from multiple modules
				func(
					httpKernel contract.HTTPKernel, // From HTTP module
					cache contract.Cache, // From Cache module
					config contract.Config, // From Config module
					logger contract.Logger, // From Log module
				) {
					initOrder = append(initOrder, "invoker")
					assert.NotNil(t, httpKernel)
					assert.NotNil(t, cache)
					assert.NotNil(t, config)
					assert.NotNil(t, logger)
				},
			},
		})

		require.NotNil(t, app)
		assert.Contains(t, initOrder, "invoker")
	})

	t.Run("provider failure does not cause circular dependency panic", func(t *testing.T) {
		resetGlobalInstance()

		// Test that provider failure is handled gracefully, not as circular dependency
		type ValidService struct{ Value string }
		assert.NotPanics(t, func() {
			app := New(Options{
				Providers: []any{
					// Valid provider
					func(config contract.Config) ValidService {
						return ValidService{Value: "valid-service"}
					},
					// This would cause a provider creation issue, not circular dependency
					func() *invalidService {
						return nil // This could cause issues but shouldn't be circular dependency
					},
				},
			})
			require.NotNil(t, app)
		})
	})
}

type invalidService struct{}

func TestDependencyResolutionOrder(t *testing.T) {
	t.Run("framework dependencies before user dependencies", func(t *testing.T) {
		resetGlobalInstance()

		type UserService struct{ Value string }
		var resolutionOrder []string

		app := New(Options{
			WithHTTP:  true,
			WithCache: true,
			Providers: []any{
				// User provider that depends on framework services
				func(config contract.Config, logger contract.Logger, cache contract.Cache) UserService {
					resolutionOrder = append(resolutionOrder, "user-provider")
					assert.NotNil(t, config)
					assert.NotNil(t, logger)
					assert.NotNil(t, cache)
					return UserService{Value: "user-service"}
				},
			},
			Invokers: []any{
				// User invoker that depends on both framework and user services
				func(userService UserService, httpKernel contract.HTTPKernel) {
					resolutionOrder = append(resolutionOrder, "user-invoker")
					assert.Equal(t, "user-service", userService.Value)
					assert.NotNil(t, httpKernel)
				},
			},
		})

		require.NotNil(t, app)
		assert.Contains(t, resolutionOrder, "user-provider")
		assert.Contains(t, resolutionOrder, "user-invoker")
	})
}
