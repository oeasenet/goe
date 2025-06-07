package contract_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/fx"
)

// MockApplication is a mock implementation of the Application interface
type MockApplication struct {
	mock.Mock
}

func (m *MockApplication) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockApplication) Version() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockApplication) Environment() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockApplication) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

func (m *MockApplication) Container() *fx.App {
	args := m.Called()
	return args.Get(0).(*fx.App)
}

func (m *MockApplication) IsRunning() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockApplication) Register(options ...fx.Option) error {
	args := m.Called(options)
	return args.Error(0)
}

func TestApplicationInterface(t *testing.T) {
	// This test verifies that MockApplication implements the Application interface
	var _ contract.Application = (*MockApplication)(nil)

	// Create a mock application
	app := new(MockApplication)

	// Set up expectations
	app.On("Name").Return("Test App")
	app.On("Version").Return("1.0.0")
	app.On("Environment").Return("test")
	app.On("Context").Return(context.Background())
	app.On("Container").Return(fx.New())
	app.On("IsRunning").Return(true)
	app.On("Register", mock.Anything).Return(nil)

	// Test the methods
	assert.Equal(t, "Test App", app.Name())
	assert.Equal(t, "1.0.0", app.Version())
	assert.Equal(t, "test", app.Environment())
	assert.NotNil(t, app.Context())
	assert.NotNil(t, app.Container())
	assert.True(t, app.IsRunning())
	assert.Nil(t, app.Register(fx.Options()))

	// Verify expectations
	app.AssertExpectations(t)
}
