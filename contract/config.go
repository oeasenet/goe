package contract

import "time"

// Config defines the configuration interface
type Config interface {
	// Get retrieves a configuration value by key
	Get(key string) any

	// GetString retrieves a string configuration value
	GetString(key string) string

	// GetInt retrieves an integer configuration value
	GetInt(key string) int

	// GetInt64 retrieves an int64 configuration value
	GetInt64(key string) int64

	// GetFloat64 retrieves a float64 configuration value
	GetFloat64(key string) float64

	// GetBool retrieves a boolean configuration value
	GetBool(key string) bool

	// GetDuration retrieves a time.Duration configuration value
	GetDuration(key string) time.Duration

	// GetStringSlice retrieves a string slice configuration value
	GetStringSlice(key string) []string

	// GetStringMap retrieves a string map configuration value
	GetStringMap(key string) map[string]any

	// Set sets a configuration value
	Set(key string, value any)

	// Has checks if a configuration key exists
	Has(key string) bool

	// All returns all configuration values
	All() map[string]any

	// Reload reloads the configuration from sources
	Reload() error
}

// ConfigSource represents a source of configuration values
type ConfigSource interface {
	// Name returns the name of the configuration source
	Name() string

	// Load loads configuration from the source
	Load() (map[string]any, error)

	// Watch watches for configuration changes
	Watch(callback func(map[string]any)) error
}
