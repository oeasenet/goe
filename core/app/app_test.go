package app_test

import (
	"context"
	"testing"
	"time"

	"go.oease.dev/goe/v2/core/app"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// MockModule is a simple implementation of contract.Module for testing
type MockModule struct {
	name            string
	initializeCalls int
	startCalls      int
	stopCalls       int
	initializeError error
	startError      error
	stopError       error
}

func NewMockModule(name string) *MockModule {
	return &MockModule{name: name}
}

func (m *MockModule) Name() string {
	return m.name
}

func (m *MockModule) Initialize(ctx context.Context) error {
	m.initializeCalls++
	return m.initializeError
}

func (m *MockModule) Start(ctx context.Context) error {
	m.startCalls++
	return m.startError
}

func (m *MockModule) Stop(ctx context.Context) error {
	m.stopCalls++
	return m.stopError
}

// TestAppLifecycle tests the lifecycle of the app module
func TestAppLifecycle(t *testing.T) {
	// Create a new app
	a := app.New()

	// Create mock modules
	module1 := NewMockModule("module1")
	module2 := NewMockModule("module2")

	// Register modules
	a.RegisterModules(module1, module2)

	// Test initialization
	err := a.Initialize(context.Background())
	if err != nil {
		t.Errorf("Failed to initialize app: %v", err)
	}

	// Verify that modules were initialized
	if module1.initializeCalls != 1 {
		t.Errorf("Expected module1 to be initialized once, got %d", module1.initializeCalls)
	}
	if module2.initializeCalls != 1 {
		t.Errorf("Expected module2 to be initialized once, got %d", module2.initializeCalls)
	}

	// Test start
	err = a.Start(context.Background())
	if err != nil {
		t.Errorf("Failed to start app: %v", err)
	}

	// Verify that modules were started
	if module1.startCalls != 1 {
		t.Errorf("Expected module1 to be started once, got %d", module1.startCalls)
	}
	if module2.startCalls != 1 {
		t.Errorf("Expected module2 to be started once, got %d", module2.startCalls)
	}

	// Test stop
	err = a.Stop(context.Background())
	if err != nil {
		t.Errorf("Failed to stop app: %v", err)
	}

	// Verify that modules were stopped
	if module1.stopCalls != 1 {
		t.Errorf("Expected module1 to be stopped once, got %d", module1.stopCalls)
	}
	if module2.stopCalls != 1 {
		t.Errorf("Expected module2 to be stopped once, got %d", module2.stopCalls)
	}
}

// TestAppWithFx tests the app module with Fx
func TestAppWithFx(t *testing.T) {
	// Create a test Fx app
	testApp := fxtest.New(t,
		// Provide the app module
		fx.Provide(app.New),

		// Invoke a function that uses the app module
		fx.Invoke(func(a *app.App) {
			// Register a mock module
			mockModule := NewMockModule("fx-module")
			a.RegisterModule(mockModule)

			// Verify that the module was registered
			if a.Container() == nil {
				t.Error("Expected container to be initialized")
			}
		}),

		// Set a short timeout for the test
		fx.StartTimeout(time.Second),
		fx.StopTimeout(time.Second),
	)

	// Start and stop the test app
	testApp.RequireStart()
	testApp.RequireStop()
}

// TestAppProviderRegistration tests registering providers with the app
func TestAppProviderRegistration(t *testing.T) {
	// Create a new app
	a := app.New()

	// Define a test type and provider
	type TestService struct {
		Name string
	}

	// Register a provider
	a.RegisterProvider(func() *TestService {
		return &TestService{Name: "test"}
	})

	// Create a channel to signal when the service is provided
	serviceProvided := make(chan struct{})

	// Register an invoker that uses the service
	a.Invoke(func(service *TestService) {
		if service.Name != "test" {
			t.Errorf("Expected service name to be 'test', got '%s'", service.Name)
		}
		close(serviceProvided)
	})

	// Start the app in a goroutine
	go func() {
		if err := a.Run(); err != nil {
			t.Errorf("Failed to run app: %v", err)
		}
	}()

	// Wait for the service to be provided or timeout
	select {
	case <-serviceProvided:
		// Service was provided successfully
	case <-time.After(time.Second):
		t.Error("Timed out waiting for service to be provided")
	}

	// Stop the app
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := a.Stop(ctx); err != nil {
		t.Errorf("Failed to stop app: %v", err)
	}
}
