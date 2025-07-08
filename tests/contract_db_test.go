package tests

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

func (m *MockDB) RegisterModelsForMigration(dst ...interface{}) {
	m.Called(dst)
}

func (m *MockDB) RegisterModelsForMigrationOnConnection(connectionName string, dst ...interface{}) {
	m.Called(connectionName, dst)
}

func TestDBInterface(t *testing.T) {
	// This test verifies that MockDB implements the DB interface
	var _ contract.DB = (*MockDB)(nil)

	// Create a mock DB
	db := new(MockDB)

	// Set up expectations
	gormDB := &gorm.DB{}
	db.On("Instance").Return(gormDB)
	db.On("Connection", "default").Return(gormDB, nil)
	db.On("AutoMigrate", mock.Anything).Return(nil)
	db.On("AutoMigrateOnConnection", "default", mock.Anything).Return(nil)
	db.On("RegisterModelsForMigration", mock.Anything).Return()
	db.On("RegisterModelsForMigrationOnConnection", "default", mock.Anything).Return()

	// Test the methods
	assert.Equal(t, gormDB, db.Instance())

	conn, err := db.Connection("default")
	assert.NoError(t, err)
	assert.Equal(t, gormDB, conn)

	err = db.AutoMigrate(&struct{}{})
	assert.NoError(t, err)

	err = db.AutoMigrateOnConnection("default", &struct{}{})
	assert.NoError(t, err)

	db.RegisterModelsForMigration(&struct{}{})
	db.RegisterModelsForMigrationOnConnection("default", &struct{}{})

	// Verify expectations
	db.AssertExpectations(t)
}
