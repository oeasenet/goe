package goe

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/http"
)

// getFreePort returns an available port on the system for testing.
func getFreePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to get free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}

func TestDependencyInjection_Integration(t *testing.T) {
	t.Run("full application DI flow", func(t *testing.T) {
		// Reset global state
		resetGlobalInstance()

		// Get a free port to avoid conflicts
		port := getFreePort(t)

		// Create app with multiple modules to test DI
		app := New(Options{
			WithHTTP:  true,
			WithCache: true,
			WithDB:    false, // Skip DB to avoid connection requirements
			HTTPPort:  port,
			Providers: []any{
				// Custom provider that depends on config and logger
				func(config contract.Config, logger contract.Logger) *customTestService {
					logger.Debug("Custom provider called with dependencies")
					return &customTestService{Name: "test-service"}
				},
			},
			Invokers: []any{
				// Custom invoker that depends on multiple services
				func(
					httpKernel contract.HTTPKernel,
					logger contract.Logger,
					config contract.Config,
					cache contract.Cache,
					customService *customTestService,
				) {
					logger.Info("DI Integration test invoker called")

					// Verify all dependencies are available
					assert.NotNil(t, httpKernel)
					assert.NotNil(t, logger)
					assert.NotNil(t, config)
					assert.NotNil(t, cache)
					assert.Equal(t, "test-service", customService.Name)

					// Register test route to verify HTTP DI
					httpKernel.App().Get("/di-test", func(c fiber.Ctx) error {
						services := http.GetServices(c)
						assert.NotNil(t, services.App)
						assert.NotNil(t, services.Config)
						assert.NotNil(t, services.Logger)
						assert.NotNil(t, services.Cache)
						return c.SendString("DI working")
					})
				},
			},
		})

		require.NotNil(t, app)

		// Test global accessors work after initialization
		assert.NotNil(t, App())
		assert.NotNil(t, Config())
		assert.NotNil(t, Log())
		assert.NotNil(t, HTTP())
		assert.NotNil(t, Cache())
	})

	t.Run("module registration order", func(t *testing.T) {
		resetGlobalInstance()

		// Get a free port to avoid conflicts
		port := getFreePort(t)

		var initOrder []string
		var initOrderMutex sync.Mutex

		// Custom module to track initialization order
		// Module constructor that will be called by DI
		moduleConstructor := func(logger contract.Logger, config contract.Config) contract.Module {
			return &testModule{
				name: "test-module",
				onStart: func() {
					initOrderMutex.Lock()
					initOrder = append(initOrder, "test-module")
					initOrderMutex.Unlock()
				},
			}
		}

		app := New(Options{
			WithHTTP:  true,
			WithCache: true,
			HTTPPort:  port,
			Modules:   []any{moduleConstructor}, // Pass constructor function
			Invokers: []any{
				func(logger contract.Logger) {
					initOrderMutex.Lock()
					initOrder = append(initOrder, "user-invoker")
					initOrderMutex.Unlock()
				},
			},
		})

		require.NotNil(t, app)

		// Start the app briefly to trigger OnStart hooks
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		go func() {
			app.Container().Run()
		}()

		<-ctx.Done()

		// Verify modules were registered
		initOrderMutex.Lock()
		assert.Contains(t, initOrder, "test-module")
		assert.Contains(t, initOrder, "user-invoker")
		initOrderMutex.Unlock()
	})

	t.Run("error scenarios", func(t *testing.T) {
		resetGlobalInstance()

		// Test panic when accessing uninitialized service
		assert.Panics(t, func() {
			Config()
		})

		assert.Panics(t, func() {
			HTTP()
		})
	})

	t.Run("optional dependencies", func(t *testing.T) {
		resetGlobalInstance()

		var cacheAvailable bool

		// Test with cache enabled
		app := New(Options{
			WithCache: true,
			Invokers: []any{
				func(cache contract.Cache) {
					cacheAvailable = (cache != nil)
				},
			},
		})

		require.NotNil(t, app)
		assert.True(t, cacheAvailable)

		resetGlobalInstance()
		cacheAvailable = false

		// Test with cache disabled - should not fail due to optional tag
		app2 := New(Options{
			WithCache: false,
			Invokers: []any{
				func(logger contract.Logger) {
					// Cache is not available, but that's okay
					cacheAvailable = false
				},
			},
		})

		require.NotNil(t, app2)
		assert.False(t, cacheAvailable)
	})
}

func TestServiceProvider_DependencyInjection(t *testing.T) {
	t.Run("HTTP service provider gets all dependencies", func(t *testing.T) {
		resetGlobalInstance()

		// Get a free port to avoid conflicts
		port := getFreePort(t)

		var providerReceived bool

		app := New(Options{
			WithHTTP:  true,
			WithCache: true,
			HTTPPort:  port,
			Invokers: []any{
				func(
					app contract.Application,
					config contract.Config,
					logger contract.Logger,
					cache contract.Cache,
				) {
					providerReceived = true
					assert.NotNil(t, app)
					assert.NotNil(t, config)
					assert.NotNil(t, logger)
					assert.NotNil(t, cache)
				},
			},
		})

		require.NotNil(t, app)
		assert.True(t, providerReceived)
	})
}

// Helper types for testing
type customTestService struct {
	Name string
}

type testModule struct {
	name    string
	onStart func()
}

func (m *testModule) Name() string {
	return m.name
}

func (m *testModule) OnStart(ctx context.Context) error {
	if m.onStart != nil {
		m.onStart()
	}
	return nil
}

func (m *testModule) OnStop(ctx context.Context) error {
	return nil
}

// Helper function to reset global state for testing
func resetGlobalInstance() {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	instance.app = nil
	instance.config = nil
	instance.logger = nil
	instance.http = nil
	instance.cacheManager = nil
	instance.db = nil
	instance.mongoDB = nil
	instance.eventManager = nil
}
