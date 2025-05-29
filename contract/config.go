package contract

import (
	"context"
	"time"
)

// Config represents the configuration module interface
type Config interface {
	Module

	// Get retrieves a configuration value as a string
	Get(key string) string

	// GetDefault retrieves a configuration value as a string with a default value
	GetDefault(key string, defaultValue string) string

	// GetInt retrieves a configuration value as an integer
	GetInt(key string) (int, error)

	// GetIntDefault retrieves a configuration value as an integer with a default value
	GetIntDefault(key string, defaultValue int) int

	// GetBool retrieves a configuration value as a boolean
	GetBool(key string) (bool, error)

	// GetBoolDefault retrieves a configuration value as a boolean with a default value
	GetBoolDefault(key string, defaultValue bool) bool

	// GetFloat retrieves a configuration value as a float64
	GetFloat(key string) (float64, error)

	// GetFloatDefault retrieves a configuration value as a float64 with a default value
	GetFloatDefault(key string, defaultValue float64) float64

	// GetDuration retrieves a configuration value as a duration
	GetDuration(key string) (time.Duration, error)

	// GetDurationDefault retrieves a configuration value as a duration with a default value
	GetDurationDefault(key string, defaultValue time.Duration) time.Duration

	// GetStringSlice retrieves a configuration value as a string slice
	GetStringSlice(key string, separator string) []string

	// Has checks if a configuration key exists
	Has(key string) bool

	// Set sets a configuration value
	Set(key string, value interface{})

	// Load loads configuration from a specific source
	Load(ctx context.Context) error

	// Reload reloads configuration from all sources
	Reload(ctx context.Context) error
}

// ConfigProvider is a function that provides a Config instance
type ConfigProvider func() Config
