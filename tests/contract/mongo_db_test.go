package contract_test

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/contract"
)

// MockMongoDB is a mock implementation of the MongoDB interface
type MockMongoDB struct {
	mock.Mock
}

func (m *MockMongoDB) Instance() *mongo.Database {
	args := m.Called()
	return args.Get(0).(*mongo.Database)
}

func (m *MockMongoDB) Connection(name string) (*mongo.Database, error) {
	args := m.Called()
	return args.Get(0).(*mongo.Database), args.Error(1)
}

func (m *MockMongoDB) IsNoDocumentsError(err error) bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMongoDB) Ctx() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

// TestModel is a simple model for testing auto migration
type TestMongoModel struct {
	Name string
	Age  int
}

func TestMongoDBInterface(t *testing.T) {
	// This test verifies that MockMongoDB implements the MongoDB interface
	var _ contract.MongoDB = (*MockMongoDB)(nil)

	// Create a mock DB
	db := new(MockMongoDB)

	// Create a mock GORM DB
	mockMongoDB := &mongo.Database{}

	// Set up expectations
	db.On("Instance").Return(mockMongoDB)
	db.On("Connection", mock.Anything).Return(mockMongoDB, nil)
	db.On("IsNoDocumentsError", mock.Anything).Return(false)
	db.On("Ctx").Return(context.Background())

	// Test the methods
	assert.Equal(t, mockMongoDB, db.Instance())

	conn, err := db.Connection("test_connection")
	assert.Nil(t, err)
	assert.Equal(t, mockMongoDB, conn)

	assert.False(t, db.IsNoDocumentsError(errors.New("")))

	assert.Equal(t, context.Background(), db.Ctx())

	// Verify expectations
	db.AssertExpectations(t)
}
