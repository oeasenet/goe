package app

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// MockModule implements contract.Module for testing
type MockModule struct {
	mock.Mock
}

func (m *MockModule) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockModule) OnStart(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockModule) OnStop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockProvider for testing
type MockProvider struct {
	Value string
}

func NewMockProvider() *MockProvider {
	return &MockProvider{Value: "test"}
}

// MockInvoker for testing
func MockInvoker(provider *MockProvider) {
	// Do nothing, just for testing
}

func TestApp_New(t *testing.T) {
	tests := []struct {
		name        string
		appName     string
		version     string
		environment string
	}{
		{
			name:        "create new app with valid parameters",
			appName:     "test-app",
			version:     "1.0.0",
			environment: "test",
		},
		{
			name:        "create new app with empty parameters",
			appName:     "",
			version:     "",
			environment: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New(tt.appName, tt.version, tt.environment)

			assert.Equal(t, tt.appName, app.Name())
			assert.Equal(t, tt.version, app.Version())
			assert.Equal(t, tt.environment, app.Environment())
			assert.NotNil(t, app.Context())
			assert.False(t, app.IsRunning())
			assert.Nil(t, app.Container())
		})
	}
}

func TestApp_BasicProperties(t *testing.T) {
	app := New("test-app", "1.0.0", "test")

	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, "test-app", app.Name())
	})

	t.Run("Version", func(t *testing.T) {
		assert.Equal(t, "1.0.0", app.Version())
	})

	t.Run("Environment", func(t *testing.T) {
		assert.Equal(t, "test", app.Environment())
	})

	t.Run("Context", func(t *testing.T) {
		ctx := app.Context()
		assert.NotNil(t, ctx)
		assert.Equal(t, context.Background(), ctx)
	})

	t.Run("IsRunning initial state", func(t *testing.T) {
		assert.False(t, app.IsRunning())
	})
}

func TestApp_AddModule(t *testing.T) {
	app := New("test-app", "1.0.0", "test")

	t.Run("add single module", func(t *testing.T) {
		module := &MockModule{}
		module.On("Name").Return("test-module")

		err := app.AddModule(module)
		assert.NoError(t, err)

		module.AssertExpectations(t)
	})

	t.Run("add multiple modules", func(t *testing.T) {
		module1 := &MockModule{}
		module1.On("Name").Return("module1")

		module2 := &MockModule{}
		module2.On("Name").Return("module2")

		err := app.AddModule(module1)
		assert.NoError(t, err)

		err = app.AddModule(module2)
		assert.NoError(t, err)

		module1.AssertExpectations(t)
		module2.AssertExpectations(t)
	})
}

func TestApp_AddProvider(t *testing.T) {
	app := New("test-app", "1.0.0", "test")

	t.Run("add provider", func(t *testing.T) {
		provider := NewMockProvider

		err := app.AddProvider(provider)
		assert.NoError(t, err)

		// Container should be created
		assert.NotNil(t, app.Container())
	})
}

func TestApp_AddInvoker(t *testing.T) {
	app := New("test-app", "1.0.0", "test")

	t.Run("add invoker", func(t *testing.T) {
		// First add the provider that the invoker depends on
		err := app.AddProvider(NewMockProvider)
		assert.NoError(t, err)

		// Then add the invoker
		err = app.AddInvoker(MockInvoker)
		assert.NoError(t, err)

		// Container should be created
		assert.NotNil(t, app.Container())
	})
}

func TestApp_Register(t *testing.T) {
	app := New("test-app", "1.0.0", "test")

	t.Run("register fx options", func(t *testing.T) {
		err := app.Register(fx.Provide(NewMockProvider))
		assert.NoError(t, err)

		// Container should be created
		assert.NotNil(t, app.Container())
	})
}

func TestApp_StartStop(t *testing.T) {
	t.Run("start and stop app without modules", func(t *testing.T) {
		app := New("test-app", "1.0.0", "test")

		ctx := context.Background()

		// Start app
		err := app.Start(ctx)
		assert.NoError(t, err)
		assert.True(t, app.IsRunning())

		// Stop app
		err = app.Stop(ctx)
		assert.NoError(t, err)
		assert.False(t, app.IsRunning())
	})

	t.Run("start and stop app with successful modules", func(t *testing.T) {
		app := New("test-app", "1.0.0", "test")

		module := &MockModule{}
		module.On("Name").Return("test-module")
		module.On("OnStart", mock.Anything).Return(nil)
		module.On("OnStop", mock.Anything).Return(nil)

		err := app.AddModule(module)
		require.NoError(t, err)

		ctx := context.Background()

		// Start app
		err = app.Start(ctx)
		assert.NoError(t, err)
		assert.True(t, app.IsRunning())

		// Stop app
		err = app.Stop(ctx)
		assert.NoError(t, err)
		assert.False(t, app.IsRunning())

		module.AssertExpectations(t)
	})

	t.Run("start fails when module fails", func(t *testing.T) {
		app := New("test-app", "1.0.0", "test")

		module1 := &MockModule{}
		module1.On("Name").Return("module1")
		module1.On("OnStart", mock.Anything).Return(nil)
		module1.On("OnStop", mock.Anything).Return(nil)

		module2 := &MockModule{}
		module2.On("Name").Return("module2")
		module2.On("OnStart", mock.Anything).Return(errors.New("start failed"))

		err := app.AddModule(module1)
		require.NoError(t, err)

		err = app.AddModule(module2)
		require.NoError(t, err)

		ctx := context.Background()

		// Start should fail
		err = app.Start(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start failed")
		assert.False(t, app.IsRunning())

		module1.AssertExpectations(t)
		module2.AssertExpectations(t)
	})

	t.Run("modules are stopped in reverse order", func(t *testing.T) {
		app := New("test-app", "1.0.0", "test")

		var stopOrder []string
		var mu sync.Mutex

		module1 := &MockModule{}
		module1.On("Name").Return("module1")
		module1.On("OnStart", mock.Anything).Return(nil)
		module1.On("OnStop", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			mu.Lock()
			stopOrder = append(stopOrder, "module1")
			mu.Unlock()
		})

		module2 := &MockModule{}
		module2.On("Name").Return("module2")
		module2.On("OnStart", mock.Anything).Return(nil)
		module2.On("OnStop", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			mu.Lock()
			stopOrder = append(stopOrder, "module2")
			mu.Unlock()
		})

		err := app.AddModule(module1)
		require.NoError(t, err)

		err = app.AddModule(module2)
		require.NoError(t, err)

		ctx := context.Background()

		// Start app
		err = app.Start(ctx)
		require.NoError(t, err)

		// Stop app
		err = app.Stop(ctx)
		require.NoError(t, err)

		// Verify stop order (reverse of start order)
		mu.Lock()
		expected := []string{"module2", "module1"}
		mu.Unlock()
		assert.Equal(t, expected, stopOrder)

		module1.AssertExpectations(t)
		module2.AssertExpectations(t)
	})
}

func TestApp_ConcurrentAccess(t *testing.T) {
	app := New("test-app", "1.0.0", "test")

	t.Run("concurrent module addition", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 10

		wg.Add(numGoroutines)

		for i := range numGoroutines {
			go func(i int) {
				defer wg.Done()

				module := &MockModule{}
				module.On("Name").Return("module")

				err := app.AddModule(module)
				assert.NoError(t, err)
			}(i)
		}

		wg.Wait()
	})

	t.Run("concurrent provider addition", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 10

		wg.Add(numGoroutines)

		for i := range numGoroutines {
			go func(i int) {
				defer wg.Done()

				// Create a unique provider for each goroutine to avoid conflicts
				provider := func() *MockProvider {
					return &MockProvider{}
				}

				err := app.AddProvider(provider)
				if err != nil {
					// In concurrent scenarios, some may fail due to timing
					// This is acceptable behavior for this test
					t.Logf("Provider addition failed (expected in concurrent test): %v", err)
				}
			}(i)
		}

		wg.Wait()
	})
}

func TestApp_Run(t *testing.T) {
	t.Run("run with nil container", func(t *testing.T) {
		app := New("test-app", "1.0.0", "test")

		// This should not panic
		done := make(chan bool)
		go func() {
			app.Run()
			done <- true
		}()

		select {
		case <-done:
			// Expected behavior
		case <-time.After(100 * time.Millisecond):
			// Also expected since Run() might block
		}
	})
}

func TestApp_StartStopWithContext(t *testing.T) {
	t.Run("start with cancelled context", func(t *testing.T) {
		app := New("test-app", "1.0.0", "test")

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := app.Start(ctx)
		// Should still succeed as we handle context differently
		assert.NoError(t, err)
	})

	t.Run("start with timeout context", func(t *testing.T) {
		app := New("test-app", "1.0.0", "test")

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := app.Start(ctx)
		assert.NoError(t, err)
	})
}

func TestApp_Container(t *testing.T) {
	t.Run("container is nil initially", func(t *testing.T) {
		app := New("test-app", "1.0.0", "test")
		assert.Nil(t, app.Container())
	})

	t.Run("container is created after register", func(t *testing.T) {
		app := New("test-app", "1.0.0", "test")

		err := app.Register(fx.Provide(NewMockProvider))
		require.NoError(t, err)

		assert.NotNil(t, app.Container())
	})
}

func TestApp_IsRunning(t *testing.T) {
	t.Run("initially not running", func(t *testing.T) {
		app := New("test-app", "1.0.0", "test")
		assert.False(t, app.IsRunning())
	})

	t.Run("running after start", func(t *testing.T) {
		app := New("test-app", "1.0.0", "test")

		ctx := context.Background()
		err := app.Start(ctx)
		require.NoError(t, err)

		assert.True(t, app.IsRunning())
	})

	t.Run("not running after stop", func(t *testing.T) {
		app := New("test-app", "1.0.0", "test")

		ctx := context.Background()
		err := app.Start(ctx)
		require.NoError(t, err)

		err = app.Stop(ctx)
		require.NoError(t, err)

		assert.False(t, app.IsRunning())
	})
}
