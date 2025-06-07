package contract_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/contract"
)

// MockModule is a mock implementation of the Module interface
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

func TestModuleInterface(t *testing.T) {
	// This test verifies that MockModule implements the Module interface
	var _ contract.Module = (*MockModule)(nil)

	// Create a mock module
	module := new(MockModule)

	// Set up expectations
	module.On("Name").Return("test_module")
	module.On("OnStart", mock.Anything).Return(nil)
	module.On("OnStop", mock.Anything).Return(nil)

	// Test the methods
	assert.Equal(t, "test_module", module.Name())
	assert.Nil(t, module.OnStart(context.Background()))
	assert.Nil(t, module.OnStop(context.Background()))

	// Verify expectations
	module.AssertExpectations(t)
}

// TestModuleLifecycle tests the lifecycle of a module
func TestModuleLifecycle(t *testing.T) {
	// Create a mock module
	module := new(MockModule)

	// Set up expectations with a specific order
	module.On("Name").Return("test_module")
	module.On("OnStart", mock.Anything).Return(nil)
	module.On("OnStop", mock.Anything).Return(nil)

	// Simulate module lifecycle
	ctx := context.Background()

	// 1. Get module name
	name := module.Name()
	assert.Equal(t, "test_module", name)

	// 2. Start the module
	err := module.OnStart(ctx)
	assert.Nil(t, err)

	// 3. Stop the module
	err = module.OnStop(ctx)
	assert.Nil(t, err)

	// Verify expectations and their order
	module.AssertExpectations(t)
}

// TestModuleWithError tests error handling in module lifecycle
func TestModuleWithError(t *testing.T) {
	// Create a mock module
	module := new(MockModule)

	// Set up expectations with errors
	module.On("Name").Return("test_module")
	module.On("OnStart", mock.Anything).Return(assert.AnError)

	// Simulate module lifecycle with error
	ctx := context.Background()

	// 1. Get module name
	name := module.Name()
	assert.Equal(t, "test_module", name)

	// 2. Start the module (should return error)
	err := module.OnStart(ctx)
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)

	// Verify expectations
	module.AssertExpectations(t)
}
