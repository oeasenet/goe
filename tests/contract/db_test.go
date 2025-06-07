package contract_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.oease.dev/goe/v2/contract"
	"gorm.io/gorm"
)

// MockDB is a mock implementation of the DB interface
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Instance() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Connection(name string) (*gorm.DB, error) {
	args := m.Called(name)
	return args.Get(0).(*gorm.DB), args.Error(1)
}

func (m *MockDB) AutoMigrate(dst ...interface{}) error {
	args := m.Called(dst)
	return args.Error(0)
}

func (m *MockDB) AutoMigrateOnConnection(connectionName string, dst ...interface{}) error {
	args := m.Called(connectionName, dst)
	return args.Error(0)
}

// TestModel is a simple model for testing auto migration
type TestModel struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:255"`
}

func TestDBInterface(t *testing.T) {
	// This test verifies that MockDB implements the DB interface
	var _ contract.DB = (*MockDB)(nil)

	// Create a mock DB
	db := new(MockDB)

	// Create a mock GORM DB
	mockGormDB := &gorm.DB{}

	// Set up expectations
	db.On("Instance").Return(mockGormDB)
	db.On("Connection", "test_connection").Return(mockGormDB, nil)
	db.On("AutoMigrate", mock.Anything).Return(nil)
	db.On("AutoMigrateOnConnection", "test_connection", mock.Anything).Return(nil)

	// Test the methods
	assert.Equal(t, mockGormDB, db.Instance())

	conn, err := db.Connection("test_connection")
	assert.Nil(t, err)
	assert.Equal(t, mockGormDB, conn)

	// Test auto migration with a model
	model := &TestModel{}
	assert.Nil(t, db.AutoMigrate(model))

	// Test auto migration on a specific connection
	assert.Nil(t, db.AutoMigrateOnConnection("test_connection", model))

	// Verify expectations
	db.AssertExpectations(t)
}
