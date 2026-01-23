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
		{
			name:     "mixed case connections",
			input:    "Primary,SECONDARY,Tertiary",
			expected: []string{"PRIMARY", "SECONDARY", "TERTIARY"},
		},
		{
			name:     "single trailing comma",
			input:    "primary,",
			expected: []string{"PRIMARY"},
		},
		{
			name:     "single leading comma",
			input:    ",primary",
			expected: []string{"PRIMARY"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseConnectionNames(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDatabaseModule_ValidateConfig_AdditionalConnectionWithPoolSettings(t *testing.T) {
	mockConfig := new(MockConfig)
	mockLogger := new(MockLogger)

	// Setup configuration with additional connection having pool settings
	mockConfig.On("GetString", "MONGO_CONNECTION").Return("primary")
	mockConfig.On("GetString", "MONGO_CONNECTIONS").Return("secondary")

	// Mock Has method for primary and secondary connections
	mockConfig.On("Has", "MONGO_primary_URI").Return(true)
	mockConfig.On("Has", "MONGO_primary_DB_NAME").Return(true)
	mockConfig.On("Has", "MONGO_SECONDARY_URI").Return(true)
	mockConfig.On("Has", "MONGO_SECONDARY_DB_NAME").Return(true)
	mockConfig.On("Has", "MONGO_SECONDARY_MIN_POOL_SIZE").Return(true)
	mockConfig.On("Has", "MONGO_SECONDARY_MAX_POOL_SIZE").Return(true)
	mockConfig.On("Has", "MONGO_SECONDARY_MAX_CONN_IDLE_TIME").Return(true)
	mockConfig.On("Has", mock.AnythingOfType("string")).Return(false)

	// Mock Get calls for additional connection pool settings
	mockConfig.On("Get", "MONGO_SECONDARY_MIN_POOL_SIZE").Return(5)
	mockConfig.On("Get", "MONGO_SECONDARY_MAX_POOL_SIZE").Return(10)
	mockConfig.On("Get", "MONGO_SECONDARY_MAX_CONN_IDLE_TIME").Return(30 * time.Second)
	mockConfig.On("GetDuration", "MONGO_SECONDARY_MAX_CONN_IDLE_TIME").Return(30 * time.Second)

	dbm := NewDBModule(mockConfig, mockLogger)
	err := dbm.ValidateConfig()

	assert.NoError(t, err)
	mockConfig.AssertExpectations(t)
}

func TestDatabaseModule_ValidateConfig_AdditionalConnectionInvalidPoolSettings(t *testing.T) {
	mockConfig := new(MockConfig)
	mockLogger := new(MockLogger)

	// Setup configuration with additional connection having invalid pool settings
	mockConfig.On("GetString", "MONGO_CONNECTION").Return("primary")
	mockConfig.On("GetString", "MONGO_CONNECTIONS").Return("secondary")

	// Mock Has method for primary and secondary connections
	mockConfig.On("Has", "MONGO_primary_URI").Return(true)
	mockConfig.On("Has", "MONGO_primary_DB_NAME").Return(true)
	mockConfig.On("Has", "MONGO_SECONDARY_URI").Return(true)
	mockConfig.On("Has", "MONGO_SECONDARY_DB_NAME").Return(true)
	mockConfig.On("Has", "MONGO_SECONDARY_MIN_POOL_SIZE").Return(true)
	mockConfig.On("Has", "MONGO_SECONDARY_MAX_CONN_IDLE_TIME").Return(true)
	mockConfig.On("Has", mock.AnythingOfType("string")).Return(false)

	// Mock Get calls with invalid values
	mockConfig.On("Get", "MONGO_SECONDARY_MIN_POOL_SIZE").Return(-1) // Invalid negative
	mockConfig.On("Get", "MONGO_SECONDARY_MAX_CONN_IDLE_TIME").Return(-30 * time.Second)
	mockConfig.On("GetDuration", "MONGO_SECONDARY_MAX_CONN_IDLE_TIME").Return(-30 * time.Second) // Invalid negative

	dbm := NewDBModule(mockConfig, mockLogger)
	err := dbm.ValidateConfig()

	assert.Error(t, err)
	mockConfig.AssertExpectations(t)
}

func TestDatabaseModule_ValidateConfig_SkipDuplicateDefaultConnection(t *testing.T) {
	mockConfig := new(MockConfig)
	mockLogger := new(MockLogger)

	// Setup where MONGO_CONNECTIONS includes the default connection name (uppercase)
	// Note: parseConnectionNames uppercases all names, but defaultConnectionName keeps original case
	// So to test the skip logic, the default must already be uppercase
	mockConfig.On("GetString", "MONGO_CONNECTION").Return("PRIMARY")
	mockConfig.On("GetString", "MONGO_CONNECTIONS").Return("PRIMARY,secondary") // PRIMARY matches default

	// Mock Has method - PRIMARY is validated once, secondary also needs validation
	mockConfig.On("Has", "MONGO_PRIMARY_URI").Return(true)
	mockConfig.On("Has", "MONGO_PRIMARY_DB_NAME").Return(true)
	mockConfig.On("Has", "MONGO_SECONDARY_URI").Return(true)
	mockConfig.On("Has", "MONGO_SECONDARY_DB_NAME").Return(true)
	mockConfig.On("Has", mock.AnythingOfType("string")).Return(false)

	dbm := NewDBModule(mockConfig, mockLogger)
	err := dbm.ValidateConfig()

	assert.NoError(t, err)
	mockConfig.AssertExpectations(t)
}

func TestDatabaseModule_ValidateConfig_EmptyConnectionInList(t *testing.T) {
	mockConfig := new(MockConfig)
	mockLogger := new(MockLogger)

	// Setup where MONGO_CONNECTIONS has empty entries
	mockConfig.On("GetString", "MONGO_CONNECTION").Return("")
	mockConfig.On("GetString", "MONGO_CONNECTIONS").Return(",secondary,,tertiary,") // has empty entries

	// Mock Has method
	mockConfig.On("Has", "MONGO_URI").Return(true)
	mockConfig.On("Has", "MONGO_DB_NAME").Return(true)
	mockConfig.On("Has", "MONGO_SECONDARY_URI").Return(true)
	mockConfig.On("Has", "MONGO_SECONDARY_DB_NAME").Return(true)
	mockConfig.On("Has", "MONGO_TERTIARY_URI").Return(true)
	mockConfig.On("Has", "MONGO_TERTIARY_DB_NAME").Return(true)
	mockConfig.On("Has", mock.AnythingOfType("string")).Return(false)

	dbm := NewDBModule(mockConfig, mockLogger)
	err := dbm.ValidateConfig()

	assert.NoError(t, err)
	mockConfig.AssertExpectations(t)
}
