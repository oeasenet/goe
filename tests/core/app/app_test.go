package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/core/app"
	"go.uber.org/fx"
)

func TestNew(t *testing.T) {
	// Test creating a new application
	application := app.New("Test App", "1.0.0", "test")

	// Verify the application properties
	assert.Equal(t, "Test App", application.Name())
	assert.Equal(t, "1.0.0", application.Version())
	assert.Equal(t, "test", application.Environment())
	assert.NotNil(t, application.Context())
	assert.Nil(t, application.Container()) // Container is nil until Register is called
	assert.False(t, application.IsRunning())
}

func TestRegister(t *testing.T) {
	// Create a new application
	application := app.New("Test App", "1.0.0", "test")

	// Register a simple option
	err := application.Register(fx.Supply(42))

	// Verify registration was successful
	assert.Nil(t, err)
	assert.NotNil(t, application.Container())
}

func TestRegisterMultiple(t *testing.T) {
	// Create a new application
	application := app.New("Test App", "1.0.0", "test")

	// Register initial options
	err := application.Register(fx.Supply(42))
	assert.Nil(t, err)

	// Get the initial container
	initialContainer := application.Container()
	assert.NotNil(t, initialContainer)

	// Register additional options (should rebuild the container)
	err = application.Register(fx.Supply("test"))
	assert.Nil(t, err)

	// Get the new container
	newContainer := application.Container()
	assert.NotNil(t, newContainer)

	// Verify the container was rebuilt (should be a different instance)
	assert.NotEqual(t, initialContainer, newContainer)
}

func TestRegisterError(t *testing.T) {
	// Create a new application
	application := app.New("Test App", "1.0.0", "test")

	// Create an option that will cause an error
	// fx.Error returns an option that fails with the given error
	err := application.Register(fx.Error(assert.AnError))

	// Verify registration failed with the expected error
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestApplicationProperties(t *testing.T) {
	// Test different application properties
	tests := []struct {
		name        string
		version     string
		environment string
	}{
		{"App1", "1.0.0", "dev"},
		{"App2", "2.0.0", "prod"},
		{"App3", "3.0.0", "test"},
		{"", "", ""}, // Test empty values
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			application := app.New(tt.name, tt.version, tt.environment)

			assert.Equal(t, tt.name, application.Name())
			assert.Equal(t, tt.version, application.Version())
			assert.Equal(t, tt.environment, application.Environment())
		})
	}
}

// MockModule is a mock implementation of the Module interface for testing
type MockModule struct {
	mock.Mock
	startCalled bool
	stopCalled  bool
}

func (m *MockModule) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockModule) OnStart(ctx context.Context) error {
	m.startCalled = true
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockModule) OnStop(ctx context.Context) error {
	m.stopCalled = true
	args := m.Called(ctx)
	return args.Error(0)
}

func TestAddModule(t *testing.T) {
	// Create a new application
	application := app.New("Test App", "1.0.0", "test")

	// Create a mock module
	module := new(MockModule)
	module.On("Name").Return("test_module")

	// Add the module
	err := application.AddModule(module)

	// Verify module was added successfully
	assert.Nil(t, err)
	assert.NotNil(t, application.Container())

	// Verify expectations
	module.AssertExpectations(t)
}

func TestAddMultipleModules(t *testing.T) {
	// Create a new application
	application := app.New("Test App", "1.0.0", "test")

	// Create mock modules
	module1 := new(MockModule)
	module1.On("Name").Return("module1")

	module2 := new(MockModule)
	module2.On("Name").Return("module2")

	// Add the modules
	err := application.AddModule(module1)
	assert.Nil(t, err)

	err = application.AddModule(module2)
	assert.Nil(t, err)

	// Verify container exists
	assert.NotNil(t, application.Container())

	// Verify expectations
	module1.AssertExpectations(t)
	module2.AssertExpectations(t)
}

func TestModuleLifecycle(t *testing.T) {
	// Create a new application
	application := app.New("Test App", "1.0.0", "test")

	// Create a mock module
	module := new(MockModule)
	module.On("Name").Return("test_module")
	module.On("OnStart", mock.Anything).Return(nil)
	module.On("OnStop", mock.Anything).Return(nil)

	// Add the module
	err := application.AddModule(module)
	assert.Nil(t, err)

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = application.Start(ctx)
	assert.Nil(t, err)
	assert.True(t, application.IsRunning())

	// Verify module OnStart was called
	assert.True(t, module.startCalled)

	// Stop the application
	err = application.Stop(ctx)
	assert.Nil(t, err)
	assert.False(t, application.IsRunning())

	// Verify module OnStop was called
	assert.True(t, module.stopCalled)

	// Verify expectations
	module.AssertExpectations(t)
}

func TestModuleStartError(t *testing.T) {
	// Create a new application
	application := app.New("Test App", "1.0.0", "test")

	// Create a mock module that fails on start
	module := new(MockModule)
	module.On("Name").Return("failing_module")
	module.On("OnStart", mock.Anything).Return(assert.AnError)

	// Add the module
	err := application.AddModule(module)
	assert.Nil(t, err)

	// Start the application (should fail)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = application.Start(ctx)
	assert.Error(t, err)

	// Verify expectations
	module.AssertExpectations(t)
}

func TestApplicationStartStop(t *testing.T) {
	// Create a new application
	application := app.New("Test App", "1.0.0", "test")

	// Register a simple provider to make the container valid
	err := application.Register(fx.Supply(42))
	assert.Nil(t, err)

	// Test starting the application
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = application.Start(ctx)
	assert.Nil(t, err)
	assert.True(t, application.IsRunning())

	// Test stopping the application
	err = application.Stop(ctx)
	assert.Nil(t, err)
	assert.False(t, application.IsRunning())
}

func TestApplicationStartStopWithoutContainer(t *testing.T) {
	// Create a new application without registering anything
	application := app.New("Test App", "1.0.0", "test")

	// Test starting the application (should not fail but do nothing)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := application.Start(ctx)
	assert.Nil(t, err)
	assert.True(t, application.IsRunning())

	// Test stopping the application
	err = application.Stop(ctx)
	assert.Nil(t, err)
	assert.False(t, application.IsRunning())
}

func TestAddProvider(t *testing.T) {
	// Create a new application
	application := app.New("Test App", "1.0.0", "test")

	// Create a simple provider function
	provider := func() string {
		return "test_value"
	}

	// Add the provider
	err := application.AddProvider(provider)

	// Verify provider was added successfully
	assert.Nil(t, err)
	assert.NotNil(t, application.Container())
}

func TestAddInvoker(t *testing.T) {
	// Create a new application
	application := app.New("Test App", "1.0.0", "test")

	// Create a simple invoker function
	invoker := func() {
		// This function will be invoked by Fx
	}

	// Add the invoker
	err := application.AddInvoker(invoker)

	// Verify invoker was added successfully
	assert.Nil(t, err)
	assert.NotNil(t, application.Container())
}

func TestAddMultipleProvidersAndInvokers(t *testing.T) {
	// Test adding providers and invokers separately to avoid rebuild conflicts
	t.Run("AddProviders", func(t *testing.T) {
		application := app.New("Test App", "1.0.0", "test")

		provider := func() string {
			return "value1"
		}

		err := application.AddProvider(provider)
		assert.Nil(t, err)
		assert.NotNil(t, application.Container())
	})

	t.Run("AddInvokers", func(t *testing.T) {
		application := app.New("Test App", "1.0.0", "test")

		invoker := func() {
			// Test invoker
		}

		err := application.AddInvoker(invoker)
		assert.Nil(t, err)
		assert.NotNil(t, application.Container())
	})
}
