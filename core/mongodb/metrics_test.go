package mongodb

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.opentelemetry.io/otel/codes"
)

func TestNewMetricsWrapper(t *testing.T) {
	mockMongoDB := new(TestMockMongoDB)
	mockMetrics := new(TestMockMetricsManager)
	mockTracing := new(TestMockTracingManager)

	wrapper := NewMetricsWrapper(mockMongoDB, mockMetrics, mockTracing)

	assert.NotNil(t, wrapper)
	assert.IsType(t, &MetricsWrapper{}, wrapper)
}

func TestNewMetricsWrapper_NilMetrics(t *testing.T) {
	mockMongoDB := new(TestMockMongoDB)
	mockTracing := new(TestMockTracingManager)

	wrapper := NewMetricsWrapper(mockMongoDB, nil, mockTracing)

	assert.Equal(t, mockMongoDB, wrapper)
}

func TestNewMetricsWrapper_NilTracing(t *testing.T) {
	mockMongoDB := new(TestMockMongoDB)
	mockMetrics := new(TestMockMetricsManager)

	wrapper := NewMetricsWrapper(mockMongoDB, mockMetrics, nil)

	assert.Equal(t, mockMongoDB, wrapper)
}

func TestMetricsWrapper_Instance_Success(t *testing.T) {
	mockMongoDB := new(TestMockMongoDB)
	mockMetrics := new(TestMockMetricsManager)
	mockTracing := new(TestMockTracingManager)
	mockSpan := new(TestMockSpan)
	mockCounter := new(TestMockCounter)
	mockHistogram := new(TestMockHistogram)

	// Mock database response
	mockDB := &mongo.Database{}
	mockMongoDB.On("Instance").Return(mockDB)

	// Mock tracing
	mockTracing.On("StartSpan", mock.Anything, "mongodb.get_instance", mock.Anything).Return(context.Background(), mockSpan)
	mockSpan.On("SetStatus", codes.Ok, "Instance retrieved successfully")
	mockSpan.On("End", mock.Anything)

	// Mock metrics
	mockMetrics.On("Counter", "mongodb_operations_total", mock.Anything).Return(mockCounter)
	mockMetrics.On("Histogram", "mongodb_operation_duration_seconds", mock.Anything).Return(mockHistogram)
	mockCounter.On("Add", mock.Anything, int64(1), mock.Anything)
	mockHistogram.On("Record", mock.Anything, mock.AnythingOfType("float64"), mock.Anything)

	wrapper := NewMetricsWrapper(mockMongoDB, mockMetrics, mockTracing)
	result := wrapper.Instance()

	assert.Equal(t, mockDB, result)
	mockMongoDB.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
	mockTracing.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
	mockCounter.AssertExpectations(t)
	mockHistogram.AssertExpectations(t)
}

func TestMetricsWrapper_Instance_Failure(t *testing.T) {
	mockMongoDB := new(TestMockMongoDB)
	mockMetrics := new(TestMockMetricsManager)
	mockTracing := new(TestMockTracingManager)
	mockSpan := new(TestMockSpan)
	mockCounter := new(TestMockCounter)
	mockHistogram := new(TestMockHistogram)

	// Mock database response (nil for failure)
	mockMongoDB.On("Instance").Return(nil)

	// Mock tracing
	mockTracing.On("StartSpan", mock.Anything, "mongodb.get_instance", mock.Anything).Return(context.Background(), mockSpan)
	mockSpan.On("SetStatus", codes.Error, "Failed to retrieve instance")
	mockSpan.On("End", mock.Anything)

	// Mock metrics - should record error
	mockMetrics.On("Counter", "mongodb_operations_total", mock.Anything).Return(mockCounter)
	mockMetrics.On("Histogram", "mongodb_operation_duration_seconds", mock.Anything).Return(mockHistogram)
	mockMetrics.On("Counter", "mongodb_errors_total", mock.Anything).Return(mockCounter)
	mockCounter.On("Add", mock.Anything, int64(1), mock.Anything)
	mockHistogram.On("Record", mock.Anything, mock.AnythingOfType("float64"), mock.Anything)

	wrapper := NewMetricsWrapper(mockMongoDB, mockMetrics, mockTracing)
	result := wrapper.Instance()

	assert.Nil(t, result)
	mockMongoDB.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
	mockTracing.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
	mockCounter.AssertExpectations(t)
	mockHistogram.AssertExpectations(t)
}

func TestMetricsWrapper_Connection_Success(t *testing.T) {
	mockMongoDB := new(TestMockMongoDB)
	mockMetrics := new(TestMockMetricsManager)
	mockTracing := new(TestMockTracingManager)
	mockSpan := new(TestMockSpan)
	mockCounter := new(TestMockCounter)
	mockHistogram := new(TestMockHistogram)

	connectionName := "test_connection"
	mockDB := &mongo.Database{}

	// Mock database response
	mockMongoDB.On("Connection", connectionName).Return(mockDB, nil)

	// Mock tracing
	mockTracing.On("StartSpan", mock.Anything, "mongodb.get_connection", mock.Anything).Return(context.Background(), mockSpan)
	mockSpan.On("SetStatus", codes.Ok, "Connection retrieved successfully")
	mockSpan.On("End", mock.Anything)

	// Mock metrics
	mockMetrics.On("Counter", "mongodb_operations_total", mock.Anything).Return(mockCounter)
	mockMetrics.On("Histogram", "mongodb_operation_duration_seconds", mock.Anything).Return(mockHistogram)
	mockCounter.On("Add", mock.Anything, int64(1), mock.Anything)
	mockHistogram.On("Record", mock.Anything, mock.AnythingOfType("float64"), mock.Anything)

	wrapper := NewMetricsWrapper(mockMongoDB, mockMetrics, mockTracing)
	result, err := wrapper.Connection(connectionName)

	assert.NoError(t, err)
	assert.Equal(t, mockDB, result)
	mockMongoDB.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
	mockTracing.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
	mockCounter.AssertExpectations(t)
	mockHistogram.AssertExpectations(t)
}

func TestMetricsWrapper_Connection_Failure(t *testing.T) {
	mockMongoDB := new(TestMockMongoDB)
	mockMetrics := new(TestMockMetricsManager)
	mockTracing := new(TestMockTracingManager)
	mockSpan := new(TestMockSpan)
	mockCounter := new(TestMockCounter)
	mockHistogram := new(TestMockHistogram)

	connectionName := "test_connection"
	testError := errors.New("connection failed")

	// Mock database response
	mockMongoDB.On("Connection", connectionName).Return(nil, testError)

	// Mock tracing
	mockTracing.On("StartSpan", mock.Anything, "mongodb.get_connection", mock.Anything).Return(context.Background(), mockSpan)
	mockSpan.On("RecordError", testError, mock.Anything)
	mockSpan.On("SetStatus", codes.Error, "Failed to retrieve connection")
	mockSpan.On("End", mock.Anything)

	// Mock metrics - should record error
	mockMetrics.On("Counter", "mongodb_operations_total", mock.Anything).Return(mockCounter)
	mockMetrics.On("Histogram", "mongodb_operation_duration_seconds", mock.Anything).Return(mockHistogram)
	mockMetrics.On("Counter", "mongodb_errors_total", mock.Anything).Return(mockCounter)
	mockCounter.On("Add", mock.Anything, int64(1), mock.Anything)
	mockHistogram.On("Record", mock.Anything, mock.AnythingOfType("float64"), mock.Anything)

	wrapper := NewMetricsWrapper(mockMongoDB, mockMetrics, mockTracing)
	result, err := wrapper.Connection(connectionName)

	assert.Error(t, err)
	assert.Equal(t, testError, err)
	assert.Nil(t, result)
	mockMongoDB.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
	mockTracing.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
	mockCounter.AssertExpectations(t)
	mockHistogram.AssertExpectations(t)
}

// Removed TestMetricsWrapper_SetMonitor as SetMonitor is no longer part of the contract

// Removed TestMetricsWrapper_IsNoDocumentsError as IsNoDocumentsError is now in utils package

// Removed TestMetricsWrapper_Ctx as Ctx is no longer part of the contract

func TestNewMetricsCommandMonitor(t *testing.T) {
	mockMetrics := new(TestMockMetricsManager)
	mockTracing := new(TestMockTracingManager)
	mockLogger := new(TestMockLogger)

	monitor := NewMetricsCommandMonitor(mockMetrics, mockTracing, mockLogger)

	assert.NotNil(t, monitor)
	assert.NotNil(t, monitor.Started)
	assert.NotNil(t, monitor.Succeeded)
	assert.NotNil(t, monitor.Failed)
}

func TestProvideMongoDBWithMetrics(t *testing.T) {
	mockMongoDB := new(TestMockMongoDB)
	mockMetrics := new(TestMockMetricsManager)
	mockTracing := new(TestMockTracingManager)

	result := ProvideMongoDBWithMetrics(mockMongoDB, mockMetrics, mockTracing)

	assert.IsType(t, &MetricsWrapper{}, result)
}
