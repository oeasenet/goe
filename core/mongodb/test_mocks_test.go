package mongodb

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTestMockConfig(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		mock := &TestMockConfig{}
		mock.On("Get", "test_key").Return("test_value")

		result := mock.Get("test_key")
		assert.Equal(t, "test_value", result)
		mock.AssertExpectations(t)
	})

	t.Run("GetString", func(t *testing.T) {
		mock := &TestMockConfig{}
		mock.On("GetString", "string_key").Return("string_value")

		result := mock.GetString("string_key")
		assert.Equal(t, "string_value", result)
		mock.AssertExpectations(t)
	})

	t.Run("GetInt", func(t *testing.T) {
		mock := &TestMockConfig{}
		mock.On("GetInt", "int_key").Return(42)

		result := mock.GetInt("int_key")
		assert.Equal(t, 42, result)
		mock.AssertExpectations(t)
	})

	t.Run("GetInt64", func(t *testing.T) {
		mock := &TestMockConfig{}
		mock.On("GetInt64", "int64_key").Return(int64(9223372036854775807))

		result := mock.GetInt64("int64_key")
		assert.Equal(t, int64(9223372036854775807), result)
		mock.AssertExpectations(t)
	})

	t.Run("GetFloat64", func(t *testing.T) {
		mock := &TestMockConfig{}
		mock.On("GetFloat64", "float_key").Return(3.14159)

		result := mock.GetFloat64("float_key")
		assert.Equal(t, 3.14159, result)
		mock.AssertExpectations(t)
	})

	t.Run("GetBool", func(t *testing.T) {
		mock := &TestMockConfig{}
		mock.On("GetBool", "bool_key").Return(true)

		result := mock.GetBool("bool_key")
		assert.True(t, result)
		mock.AssertExpectations(t)
	})

	t.Run("GetDuration", func(t *testing.T) {
		mock := &TestMockConfig{}
		mock.On("GetDuration", "duration_key").Return(5 * time.Second)

		result := mock.GetDuration("duration_key")
		assert.Equal(t, 5*time.Second, result)
		mock.AssertExpectations(t)
	})

	t.Run("GetStringSlice", func(t *testing.T) {
		mock := &TestMockConfig{}
		mock.On("GetStringSlice", "slice_key").Return([]string{"a", "b", "c"})

		result := mock.GetStringSlice("slice_key")
		assert.Equal(t, []string{"a", "b", "c"}, result)
		mock.AssertExpectations(t)
	})

	t.Run("Has", func(t *testing.T) {
		mock := &TestMockConfig{}
		mock.On("Has", "existing_key").Return(true)
		mock.On("Has", "missing_key").Return(false)

		assert.True(t, mock.Has("existing_key"))
		assert.False(t, mock.Has("missing_key"))
		mock.AssertExpectations(t)
	})
}

func TestTestMockLogger(t *testing.T) {
	t.Run("Info", func(t *testing.T) {
		mock := &TestMockLogger{}
		mock.On("Info", "test message", []interface{}{"key", "value"}).Return()

		mock.Info("test message", "key", "value")
		mock.AssertExpectations(t)
	})

	t.Run("Debug", func(t *testing.T) {
		mock := &TestMockLogger{}
		mock.On("Debug", "debug message", []interface{}(nil)).Return()

		mock.Debug("debug message")
		mock.AssertExpectations(t)
	})

	t.Run("Warn", func(t *testing.T) {
		mock := &TestMockLogger{}
		mock.On("Warn", "warn message", []interface{}(nil)).Return()

		mock.Warn("warn message")
		mock.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mock := &TestMockLogger{}
		mock.On("Error", "error message", []interface{}(nil)).Return()

		mock.Error("error message")
		mock.AssertExpectations(t)
	})

	t.Run("Fatal", func(t *testing.T) {
		mock := &TestMockLogger{}
		mock.On("Fatal", "fatal message", []interface{}(nil)).Return()

		mock.Fatal("fatal message")
		mock.AssertExpectations(t)
	})

	t.Run("Panic", func(t *testing.T) {
		mock := &TestMockLogger{}
		mock.On("Panic", "panic message", []interface{}(nil)).Return()

		mock.Panic("panic message")
		mock.AssertExpectations(t)
	})

	t.Run("Panicf", func(t *testing.T) {
		mock := &TestMockLogger{}
		mock.On("Panicf", "panic %s", []interface{}{"formatted"}).Return()

		mock.Panicf("panic %s", "formatted")
		mock.AssertExpectations(t)
	})

	t.Run("Debugf", func(t *testing.T) {
		mock := &TestMockLogger{}
		mock.On("Debugf", "debug %s", []interface{}{"formatted"}).Return()

		mock.Debugf("debug %s", "formatted")
		mock.AssertExpectations(t)
	})

	t.Run("Infof", func(t *testing.T) {
		mock := &TestMockLogger{}
		mock.On("Infof", "info %s", []interface{}{"formatted"}).Return()

		mock.Infof("info %s", "formatted")
		mock.AssertExpectations(t)
	})

	t.Run("Warnf", func(t *testing.T) {
		mock := &TestMockLogger{}
		mock.On("Warnf", "warn %s", []interface{}{"formatted"}).Return()

		mock.Warnf("warn %s", "formatted")
		mock.AssertExpectations(t)
	})

	t.Run("Errorf", func(t *testing.T) {
		mock := &TestMockLogger{}
		mock.On("Errorf", "error %s", []interface{}{"formatted"}).Return()

		mock.Errorf("error %s", "formatted")
		mock.AssertExpectations(t)
	})

	t.Run("Fatalf", func(t *testing.T) {
		mock := &TestMockLogger{}
		mock.On("Fatalf", "fatal %s", []interface{}{"formatted"}).Return()

		mock.Fatalf("fatal %s", "formatted")
		mock.AssertExpectations(t)
	})

	t.Run("Debugw", func(t *testing.T) {
		mock := &TestMockLogger{}
		mock.On("Debugw", "debug structured", []interface{}{"key", "value"}).Return()

		mock.Debugw("debug structured", "key", "value")
		mock.AssertExpectations(t)
	})

	t.Run("Infow", func(t *testing.T) {
		mock := &TestMockLogger{}
		mock.On("Infow", "info structured", []interface{}{"key", "value"}).Return()

		mock.Infow("info structured", "key", "value")
		mock.AssertExpectations(t)
	})

	t.Run("Warnw", func(t *testing.T) {
		mock := &TestMockLogger{}
		mock.On("Warnw", "warn structured", []interface{}{"key", "value"}).Return()

		mock.Warnw("warn structured", "key", "value")
		mock.AssertExpectations(t)
	})

	t.Run("Errorw", func(t *testing.T) {
		mock := &TestMockLogger{}
		mock.On("Errorw", "error structured", []interface{}{"key", "value"}).Return()

		mock.Errorw("error structured", "key", "value")
		mock.AssertExpectations(t)
	})

	t.Run("Fatalw", func(t *testing.T) {
		mock := &TestMockLogger{}
		mock.On("Fatalw", "fatal structured", []interface{}{"key", "value"}).Return()

		mock.Fatalw("fatal structured", "key", "value")
		mock.AssertExpectations(t)
	})

	t.Run("With", func(t *testing.T) {
		mock := &TestMockLogger{}
		mock.On("With", []interface{}{"key", "value"}).Return(mock)

		result := mock.With("key", "value")
		assert.Equal(t, mock, result)
		mock.AssertExpectations(t)
	})

	t.Run("WithContext", func(t *testing.T) {
		mock := &TestMockLogger{}
		ctx := context.Background()
		mock.On("WithContext", ctx).Return(mock)

		result := mock.WithContext(ctx)
		assert.Equal(t, mock, result)
		mock.AssertExpectations(t)
	})

	t.Run("WithError", func(t *testing.T) {
		mock := &TestMockLogger{}
		err := errors.New("test error")
		mock.On("WithError", err).Return(mock)

		result := mock.WithError(err)
		assert.Equal(t, mock, result)
		mock.AssertExpectations(t)
	})

	t.Run("GetLogger", func(t *testing.T) {
		mock := &TestMockLogger{}
		zapLogger := zap.NewNop().Sugar()
		mock.On("GetLogger").Return(zapLogger)

		result := mock.GetLogger()
		assert.Equal(t, zapLogger, result)
		mock.AssertExpectations(t)
	})
}

func TestTestMockMongoDB(t *testing.T) {
	t.Run("Instance returns nil", func(t *testing.T) {
		mock := &TestMockMongoDB{}
		mock.On("Instance").Return(nil)

		result := mock.Instance()
		assert.Nil(t, result)
		mock.AssertExpectations(t)
	})

	t.Run("Connection returns nil with error", func(t *testing.T) {
		mock := &TestMockMongoDB{}
		expectedErr := errors.New("connection not found")
		mock.On("Connection", "nonexistent").Return(nil, expectedErr)

		result, err := mock.Connection("nonexistent")
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectations(t)
	})
}
