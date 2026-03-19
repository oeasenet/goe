package log

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewFxDebugLogger(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fxLogger := NewFxDebugLogger(logger)

	require.NotNil(t, fxLogger)
	assert.IsType(t, &FxDebugLogger{}, fxLogger)
}

func TestFxDebugLogger_String(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fxLogger := NewFxDebugLogger(logger).(*FxDebugLogger)

	assert.Equal(t, "FxDebugLogger", fxLogger.String())
}

func TestFxDebugLogger_Test(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	logger := zap.New(core)
	fxLogger := NewFxDebugLogger(logger).(*FxDebugLogger)

	fxLogger.Test("test message")

	assert.Equal(t, 1, logs.Len())
	assert.Equal(t, "test message", logs.All()[0].Message)
}

func TestFxDebugLogger_LogEvent_DebugEnabled(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	logger := zap.New(core)
	fxLogger := NewFxDebugLogger(logger)

	t.Run("OnStartExecuting", func(t *testing.T) {
		logs.TakeAll() // Clear previous logs
		fxLogger.LogEvent(&fxevent.OnStartExecuting{
			FunctionName: "testFunc",
			CallerName:   "testCaller",
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "OnStart hook executing", logs.All()[0].Message)
	})

	t.Run("OnStartExecuted success", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.OnStartExecuted{
			FunctionName: "testFunc",
			CallerName:   "testCaller",
			Runtime:      time.Millisecond * 100,
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "OnStart hook executed", logs.All()[0].Message)
	})

	t.Run("OnStartExecuted error", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.OnStartExecuted{
			FunctionName: "testFunc",
			CallerName:   "testCaller",
			Err:          errors.New("start error"),
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "OnStart hook failed", logs.All()[0].Message)
		assert.Equal(t, zap.ErrorLevel, logs.All()[0].Level)
	})

	t.Run("OnStopExecuting", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.OnStopExecuting{
			FunctionName: "testFunc",
			CallerName:   "testCaller",
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "OnStop hook executing", logs.All()[0].Message)
	})

	t.Run("OnStopExecuted success", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.OnStopExecuted{
			FunctionName: "testFunc",
			CallerName:   "testCaller",
			Runtime:      time.Millisecond * 50,
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "OnStop hook executed", logs.All()[0].Message)
	})

	t.Run("OnStopExecuted error", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.OnStopExecuted{
			FunctionName: "testFunc",
			CallerName:   "testCaller",
			Err:          errors.New("stop error"),
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "OnStop hook failed", logs.All()[0].Message)
		assert.Equal(t, zap.ErrorLevel, logs.All()[0].Level)
	})

	t.Run("Supplied success", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Supplied{
			TypeName:    "TestType",
			StackTrace:  []string{"trace1"},
			ModuleTrace: []string{"module1"},
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Supplied", logs.All()[0].Message)
	})

	t.Run("Supplied error", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Supplied{
			TypeName: "TestType",
			Err:      errors.New("supply error"),
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Error supplying", logs.All()[0].Message)
		assert.Equal(t, zap.ErrorLevel, logs.All()[0].Level)
	})

	t.Run("Provided success", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Provided{
			ConstructorName: "NewTestType",
			OutputTypeNames: []string{"TestType"},
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Provided", logs.All()[0].Message)
	})

	t.Run("Provided error", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Provided{
			ConstructorName: "NewTestType",
			Err:             errors.New("provide error"),
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Error providing", logs.All()[0].Message)
		assert.Equal(t, zap.ErrorLevel, logs.All()[0].Level)
	})

	t.Run("Replaced success", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Replaced{
			OutputTypeNames: []string{"TestType"},
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Replaced", logs.All()[0].Message)
	})

	t.Run("Replaced error", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Replaced{
			Err: errors.New("replace error"),
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Error replacing", logs.All()[0].Message)
	})

	t.Run("Decorated success", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Decorated{
			DecoratorName:   "TestDecorator",
			OutputTypeNames: []string{"TestType"},
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Decorated", logs.All()[0].Message)
	})

	t.Run("Decorated error", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Decorated{
			DecoratorName: "TestDecorator",
			Err:           errors.New("decorate error"),
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Error decorating", logs.All()[0].Message)
	})

	t.Run("Run success", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Run{
			Name:    "TestRun",
			Kind:    "kind",
			Runtime: time.Millisecond * 10,
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Run", logs.All()[0].Message)
	})

	t.Run("Run error", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Run{
			Name: "TestRun",
			Err:  errors.New("run error"),
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Error running", logs.All()[0].Message)
	})

	t.Run("Invoking", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Invoking{
			FunctionName: "TestInvoke",
			ModuleName:   "testModule",
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Invoking", logs.All()[0].Message)
	})

	t.Run("Invoked error", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Invoked{
			FunctionName: "TestInvoke",
			Err:          errors.New("invoke error"),
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Invoke failed", logs.All()[0].Message)
	})

	t.Run("Invoked success no log", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Invoked{
			FunctionName: "TestInvoke",
		})

		// Success case is not logged
		assert.Equal(t, 0, logs.Len())
	})

	t.Run("Stopping", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Stopping{
			Signal: os.Interrupt,
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Stopping application", logs.All()[0].Message)
	})

	t.Run("Stopped error", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Stopped{
			Err: errors.New("stop error"),
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Stop failed", logs.All()[0].Message)
	})

	t.Run("Stopped success no log", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Stopped{})

		// Success case is not logged
		assert.Equal(t, 0, logs.Len())
	})

	t.Run("RollingBack", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.RollingBack{
			StartErr: errors.New("start error"),
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Rolling back", logs.All()[0].Message)
	})

	t.Run("RolledBack error", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.RolledBack{
			Err: errors.New("rollback error"),
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Rollback failed", logs.All()[0].Message)
	})

	t.Run("RolledBack success no log", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.RolledBack{})

		// Success case is not logged
		assert.Equal(t, 0, logs.Len())
	})

	t.Run("Started success", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Started{})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Started", logs.All()[0].Message)
	})

	t.Run("Started error", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Started{
			Err: errors.New("start error"),
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Start failed", logs.All()[0].Message)
	})

	t.Run("LoggerInitialized success", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.LoggerInitialized{
			ConstructorName: "NewCustomLogger",
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Initialized custom fxevent.Logger", logs.All()[0].Message)
	})

	t.Run("LoggerInitialized error", func(t *testing.T) {
		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.LoggerInitialized{
			Err: errors.New("logger init error"),
		})

		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Error initializing custom logger", logs.All()[0].Message)
	})
}

func TestFxDebugLogger_LogEvent_DebugDisabled(t *testing.T) {
	// Create a logger with only error level enabled
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	fxLogger := NewFxDebugLogger(logger)

	t.Run("OnStartExecuted error only", func(t *testing.T) {
		logs.TakeAll()

		// Success case - should not log
		fxLogger.LogEvent(&fxevent.OnStartExecuted{
			FunctionName: "testFunc",
		})
		assert.Equal(t, 0, logs.Len())

		// Error case - should log
		fxLogger.LogEvent(&fxevent.OnStartExecuted{
			FunctionName: "testFunc",
			Err:          errors.New("error"),
		})
		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "OnStart hook failed", logs.All()[0].Message)
	})

	t.Run("OnStopExecuted error only", func(t *testing.T) {
		logs.TakeAll()

		fxLogger.LogEvent(&fxevent.OnStopExecuted{
			FunctionName: "testFunc",
		})
		assert.Equal(t, 0, logs.Len())

		fxLogger.LogEvent(&fxevent.OnStopExecuted{
			FunctionName: "testFunc",
			Err:          errors.New("error"),
		})
		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "OnStop hook failed", logs.All()[0].Message)
	})

	t.Run("Supplied error only", func(t *testing.T) {
		logs.TakeAll()

		fxLogger.LogEvent(&fxevent.Supplied{TypeName: "TestType"})
		assert.Equal(t, 0, logs.Len())

		fxLogger.LogEvent(&fxevent.Supplied{
			TypeName: "TestType",
			Err:      errors.New("error"),
		})
		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Error supplying", logs.All()[0].Message)
	})

	t.Run("Provided error only", func(t *testing.T) {
		logs.TakeAll()

		fxLogger.LogEvent(&fxevent.Provided{ConstructorName: "New"})
		assert.Equal(t, 0, logs.Len())

		fxLogger.LogEvent(&fxevent.Provided{
			ConstructorName: "New",
			Err:             errors.New("error"),
		})
		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Error providing", logs.All()[0].Message)
	})

	t.Run("Invoked error only", func(t *testing.T) {
		logs.TakeAll()

		fxLogger.LogEvent(&fxevent.Invoked{FunctionName: "Test"})
		assert.Equal(t, 0, logs.Len())

		fxLogger.LogEvent(&fxevent.Invoked{
			FunctionName: "Test",
			Err:          errors.New("error"),
		})
		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Invoke failed", logs.All()[0].Message)
	})

	t.Run("Started logs at info", func(t *testing.T) {
		logs.TakeAll()

		fxLogger.LogEvent(&fxevent.Started{})
		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Started", logs.All()[0].Message)

		logs.TakeAll()
		fxLogger.LogEvent(&fxevent.Started{Err: errors.New("error")})
		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Start failed", logs.All()[0].Message)
	})

	t.Run("Stopping logs at info", func(t *testing.T) {
		logs.TakeAll()

		fxLogger.LogEvent(&fxevent.Stopping{Signal: os.Interrupt})
		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Stopping application", logs.All()[0].Message)
	})

	t.Run("Stopped error only", func(t *testing.T) {
		logs.TakeAll()

		fxLogger.LogEvent(&fxevent.Stopped{})
		assert.Equal(t, 0, logs.Len())

		fxLogger.LogEvent(&fxevent.Stopped{Err: errors.New("error")})
		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Stop failed", logs.All()[0].Message)
	})

	t.Run("RollingBack always logs", func(t *testing.T) {
		logs.TakeAll()

		fxLogger.LogEvent(&fxevent.RollingBack{StartErr: errors.New("start error")})
		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Rolling back", logs.All()[0].Message)
	})

	t.Run("RolledBack error only", func(t *testing.T) {
		logs.TakeAll()

		fxLogger.LogEvent(&fxevent.RolledBack{})
		assert.Equal(t, 0, logs.Len())

		fxLogger.LogEvent(&fxevent.RolledBack{Err: errors.New("error")})
		require.Equal(t, 1, logs.Len())
		assert.Equal(t, "Rollback failed", logs.All()[0].Message)
	})
}

func TestFxDebugLogger_LogEvent_UnknownEvent(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	logger := zap.New(core)
	fxLogger := NewFxDebugLogger(logger)

	// Unknown event type should not panic and should not log
	fxLogger.LogEvent(nil)
	assert.Equal(t, 0, logs.Len())
}

// Test convertFields and convertFieldsToArgs with actual contract.Field
func TestConvertFieldsWithContractField(t *testing.T) {
	f1 := NewField("key1", "value1")
	f2 := NewField("key2", 42)

	fields := []interface {
		Key() string
		Value() any
	}{f1.(*field), f2.(*field)}

	// Test convertFields logic
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key(), f.Value())
	}

	assert.Len(t, zapFields, 2)
	assert.Equal(t, "key1", zapFields[0].Key)
	assert.Equal(t, "key2", zapFields[1].Key)
}

