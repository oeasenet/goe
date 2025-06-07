package app_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
