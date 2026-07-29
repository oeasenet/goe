package http

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHTTP_Module(t *testing.T) {
	config := &MockConfig{}
	logger := &MockLogger{}

	// Mock required configuration
	config.On("GetString", mock.Anything).Return("")
	config.On("GetBool", mock.Anything).Return(false)
	config.On("GetInt", mock.Anything).Return(0)
	config.On("GetStringSlice", mock.Anything).Return([]string{})
	config.On("GetDuration", mock.Anything).Return(time.Duration(0))
	config.On("Has", mock.Anything).Return(false)

	logger.On("Fatal", mock.Anything).Return()
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()
	// Create a noop logger for GetLogger
	zapLogger, _ := zap.NewProduction()
	logger.On("GetLogger").Return(zapLogger.Sugar())

	module := NewModule(config, logger)

	t.Run("module name", func(t *testing.T) {
		assert.Equal(t, "http", module.Name())
	})

	t.Run("module provides kernel", func(t *testing.T) {
		kernel := module.Provide()
		assert.NotNil(t, kernel)
	})

	t.Run("module provides validator", func(t *testing.T) {
		validator := module.ProvideValidator()
		assert.NotNil(t, validator)
	})

	t.Run("module lifecycle", func(t *testing.T) {
		// Use an OS-assigned free port to avoid conflicts in CI
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		assert.NoError(t, err)
		freePort := ln.Addr().(*net.TCPAddr).Port
		require.NoError(t, ln.Close())

		lcConfig := &MockConfig{}
		lcLogger := &MockLogger{}

		lcConfig.On("GetInt", "HTTP_PORT").Return(freePort)
		lcConfig.On("GetString", "HTTP_HOST").Return("127.0.0.1")
		lcConfig.On("GetString", mock.Anything).Return("")
		lcConfig.On("GetBool", mock.Anything).Return(false)
		lcConfig.On("GetInt", mock.Anything).Return(0)
		lcConfig.On("GetStringSlice", mock.Anything).Return([]string{})
		lcConfig.On("GetDuration", mock.Anything).Return(time.Duration(0))
		lcConfig.On("Has", mock.Anything).Return(false)

		lcLogger.On("Fatal", mock.Anything).Return()
		lcLogger.On("Info", mock.Anything, mock.Anything).Return()
		lcLogger.On("Error", mock.Anything, mock.Anything).Return()
		zapLogger, _ := zap.NewProduction()
		lcLogger.On("GetLogger").Return(zapLogger.Sugar())

		lcModule := NewModule(lcConfig, lcLogger)

		ctx := context.Background()

		// OnStart should start the server in background
		err = lcModule.OnStart(ctx)
		assert.NoError(t, err)

		// Give it a moment to start
		time.Sleep(10 * time.Millisecond)

		// OnStop should shutdown the server
		err = lcModule.OnStop(ctx)
		assert.NoError(t, err)
	})
}
