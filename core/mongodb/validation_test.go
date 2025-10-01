package mongodb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDatabaseModule_ValidateConfig_ValidDefaultConfiguration(t *testing.T) {
	mockConfig := new(MockConfig)
	mockLogger := new(MockLogger)

	// Setup valid default configuration
	mockConfig.On("GetString", "MONGO_CONNECTION").Return("")
	mockConfig.On("GetString", "MONGO_CONNECTIONS").Return("")

	// Mock Has method to return true for required keys
	mockConfig.On("Has", "MONGO_URI").Return(true)
	mockConfig.On("Has", "MONGO_DB_NAME").Return(true)
	mockConfig.On("Has", mock.AnythingOfType("string")).Return(false)

	dbm := NewDBModule(mockConfig, mockLogger)
	err := dbm.ValidateConfig()

	assert.NoError(t, err)
	mockConfig.AssertExpectations(t)
}

func TestDatabaseModule_ValidateConfig_MissingRequiredFields(t *testing.T) {
	mockConfig := new(MockConfig)
	mockLogger := new(MockLogger)

	// Setup missing required fields
	mockConfig.On("GetString", "MONGO_CONNECTION").Return("")
	mockConfig.On("GetString", "MONGO_CONNECTIONS").Return("")

	// Mock Has method to return false for all keys (simulating missing configuration)
	mockConfig.On("Has", mock.AnythingOfType("string")).Return(false)

	dbm := NewDBModule(mockConfig, mockLogger)
	err := dbm.ValidateConfig()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MONGO_URI")
	assert.Contains(t, err.Error(), "MONGO_DB_NAME")
	mockConfig.AssertExpectations(t)
}

func TestDatabaseModule_ValidateConfig_WithNamedConnection(t *testing.T) {
	mockConfig := new(MockConfig)
	mockLogger := new(MockLogger)

	// Setup named connection configuration
	mockConfig.On("GetString", "MONGO_CONNECTION").Return("primary")
	mockConfig.On("GetString", "MONGO_CONNECTIONS").Return("")

	// Mock Has method to return true for required named connection keys
	mockConfig.On("Has", "MONGO_primary_URI").Return(true)
	mockConfig.On("Has", "MONGO_primary_DB_NAME").Return(true)
	mockConfig.On("Has", mock.AnythingOfType("string")).Return(false)

	dbm := NewDBModule(mockConfig, mockLogger)
	err := dbm.ValidateConfig()

	assert.NoError(t, err)
	mockConfig.AssertExpectations(t)
}

func TestDatabaseModule_ValidateConfig_WithMultipleConnections(t *testing.T) {
	mockConfig := new(MockConfig)
	mockLogger := new(MockLogger)

	// Setup multiple connections
	mockConfig.On("GetString", "MONGO_CONNECTION").Return("primary")
	mockConfig.On("GetString", "MONGO_CONNECTIONS").Return("secondary, readonly")

	// Mock Has method to return true for all required connection keys
	mockConfig.On("Has", "MONGO_primary_URI").Return(true)
	mockConfig.On("Has", "MONGO_primary_DB_NAME").Return(true)
	mockConfig.On("Has", "MONGO_SECONDARY_URI").Return(true)
	mockConfig.On("Has", "MONGO_SECONDARY_DB_NAME").Return(true)
	mockConfig.On("Has", "MONGO_READONLY_URI").Return(true)
	mockConfig.On("Has", "MONGO_READONLY_DB_NAME").Return(true)
	mockConfig.On("Has", mock.AnythingOfType("string")).Return(false)

	dbm := NewDBModule(mockConfig, mockLogger)
	err := dbm.ValidateConfig()

	assert.NoError(t, err)
	mockConfig.AssertExpectations(t)
}

func TestDatabaseModule_ValidateConfig_WithOptionalPoolSettings(t *testing.T) {
	mockConfig := new(MockConfig)
	mockLogger := new(MockLogger)

	// Setup configuration with optional pool settings
	mockConfig.On("GetString", "MONGO_CONNECTION").Return("")
	mockConfig.On("GetString", "MONGO_CONNECTIONS").Return("")

	// Mock Has method - required keys exist, optional pool settings exist
	mockConfig.On("Has", "MONGO_URI").Return(true)
	mockConfig.On("Has", "MONGO_DB_NAME").Return(true)
	mockConfig.On("Has", "MONGO_MIN_POOL_SIZE").Return(true)
	mockConfig.On("Has", "MONGO_MAX_POOL_SIZE").Return(true)
	mockConfig.On("Has", "MONGO_MAX_CONN_IDLE_TIME").Return(true)
	mockConfig.On("Has", mock.AnythingOfType("string")).Return(false)

	// Mock Get calls for validation
	mockConfig.On("Get", "MONGO_MIN_POOL_SIZE").Return(5)
	mockConfig.On("Get", "MONGO_MAX_POOL_SIZE").Return(10)
	mockConfig.On("Get", "MONGO_MAX_CONN_IDLE_TIME").Return(30 * time.Second)
	mockConfig.On("GetDuration", "MONGO_MAX_CONN_IDLE_TIME").Return(30 * time.Second)

	dbm := NewDBModule(mockConfig, mockLogger)
	err := dbm.ValidateConfig()

	assert.NoError(t, err)
	mockConfig.AssertExpectations(t)
}

func TestDatabaseModule_ValidateConfig_InvalidPoolSettings(t *testing.T) {
	mockConfig := new(MockConfig)
	mockLogger := new(MockLogger)

	// Setup configuration with invalid pool settings
	mockConfig.On("GetString", "MONGO_CONNECTION").Return("")
	mockConfig.On("GetString", "MONGO_CONNECTIONS").Return("")

	// Mock Has method - required keys exist, invalid pool settings exist
	mockConfig.On("Has", "MONGO_URI").Return(true)
	mockConfig.On("Has", "MONGO_DB_NAME").Return(true)
	mockConfig.On("Has", "MONGO_MIN_POOL_SIZE").Return(true)
	mockConfig.On("Has", "MONGO_MAX_CONN_IDLE_TIME").Return(true)
	mockConfig.On("Has", mock.AnythingOfType("string")).Return(false)

	// Mock Get calls for validation with invalid values
	mockConfig.On("Get", "MONGO_MIN_POOL_SIZE").Return(-1)
	mockConfig.On("Get", "MONGO_MAX_CONN_IDLE_TIME").Return(-30 * time.Second)
	mockConfig.On("GetDuration", "MONGO_MAX_CONN_IDLE_TIME").Return(-30 * time.Second) // Invalid negative duration

	dbm := NewDBModule(mockConfig, mockLogger)
	err := dbm.ValidateConfig()

	assert.Error(t, err)
	mockConfig.AssertExpectations(t)
}

func TestParseConnectionNames(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single connection",
			input:    "secondary",
			expected: []string{"SECONDARY"},
		},
		{
			name:     "multiple connections",
			input:    "secondary,readonly,analytics",
			expected: []string{"SECONDARY", "READONLY", "ANALYTICS"},
		},
		{
			name:     "connections with spaces",
			input:    " secondary , readonly , analytics ",
			expected: []string{"SECONDARY", "READONLY", "ANALYTICS"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "only spaces and commas",
			input:    " , , ",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseConnectionNames(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
