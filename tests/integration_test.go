package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.oease.dev/goe/v2/core/app"
)

// TestModule is a simple test module for integration testing
type TestModule struct {
	name        string
	startCalled bool
	stopCalled  bool
}

func NewTestModule(name string) *TestModule {
	return &TestModule{
		name: name,
	}
}

func (m *TestModule) Name() string {
	return m.name
}

func (m *TestModule) OnStart(ctx context.Context) error {
	m.startCalled = true
	return nil
}

func (m *TestModule) OnStop(ctx context.Context) error {
	m.stopCalled = true
	return nil
}

func TestIntegrationModuleLifecycle(t *testing.T) {
	// Create a new application
	application := app.New("Test App", "1.0.0", "test")

	// Create test modules
	module1 := NewTestModule("module1")
	module2 := NewTestModule("module2")

	// Add modules using the new AddModule method
	err := application.AddModule(module1)
	assert.NoError(t, err)

	err = application.AddModule(module2)
	assert.NoError(t, err)

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = application.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, application.IsRunning())

	// Verify modules were started
	assert.True(t, module1.startCalled)
	assert.True(t, module2.startCalled)

	// Stop the application
	err = application.Stop(ctx)
	assert.NoError(t, err)
	assert.False(t, application.IsRunning())

	// Verify modules were stopped
	assert.True(t, module1.stopCalled)
	assert.True(t, module2.stopCalled)
}

// TestProviderInvokerIntegration tests that providers and invokers work correctly with DI
func TestIntegrationProviderInvoker(t *testing.T) {
	// Create a new application
	application := app.New("Test App", "1.0.0", "test")

	// Track if invoker was called
	var invokerCalled bool
	var receivedValue string

	// Add a provider that provides a string
	provider := func() string {
		return "test_dependency_value"
	}
	err := application.AddProvider(provider)
	assert.NoError(t, err)

	// Add an invoker that depends on the provided string
	invoker := func(value string) {
		invokerCalled = true
		receivedValue = value
	}
	err = application.AddInvoker(invoker)
	assert.NoError(t, err)

	// Start the application to trigger DI
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = application.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, application.IsRunning())

	// Verify that the invoker was called with the correct dependency
	assert.True(t, invokerCalled)
	assert.Equal(t, "test_dependency_value", receivedValue)

	// Stop the application
	err = application.Stop(ctx)
	assert.NoError(t, err)
	assert.False(t, application.IsRunning())
}
