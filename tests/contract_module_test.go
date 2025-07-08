package tests

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

	ctx := context.Background()
	assert.NoError(t, module.OnStart(ctx))
	assert.NoError(t, module.OnStop(ctx))

	// Verify expectations
	module.AssertExpectations(t)
}

func TestModuleLifecycle(t *testing.T) {
	// Test module lifecycle
	module := new(MockModule)

	// Set up expectations for lifecycle
	module.On("Name").Return("lifecycle_module")
	module.On("OnStart", mock.Anything).Return(nil)
	module.On("OnStop", mock.Anything).Return(nil)

	ctx := context.Background()

	// Test name
	assert.Equal(t, "lifecycle_module", module.Name())

	// Test start
	err := module.OnStart(ctx)
	assert.NoError(t, err)

	// Test stop
	err = module.OnStop(ctx)
	assert.NoError(t, err)

	// Verify expectations
	module.AssertExpectations(t)
}

func TestModuleWithError(t *testing.T) {
	// Test module with error
	module := new(MockModule)

	// Set up expectations for error scenarios
	module.On("Name").Return("error_module")
	module.On("OnStart", mock.Anything).Return(assert.AnError)
	module.On("OnStop", mock.Anything).Return(assert.AnError)

	ctx := context.Background()

	// Test name
	assert.Equal(t, "error_module", module.Name())

	// Test start error
	err := module.OnStart(ctx)
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)

	// Test stop error
	err = module.OnStop(ctx)
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)

	// Verify expectations
	module.AssertExpectations(t)
}
