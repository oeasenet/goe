package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.oease.dev/goe/v2/contract"
)

// Config implements the contract.Config interface
type Config struct {
	mu       sync.RWMutex
	values   map[string]string
	env      string
	basePath string
}

// New creates a new Config instance
func New() *Config {
	return &Config{
		values: make(map[string]string),
		env:    "dev", // Default environment
	}
}

// Name returns the name of the module
func (c *Config) Name() string {
	return "config"
}

// Initialize initializes the config module
func (c *Config) Initialize(ctx context.Context) error {
	// Set base path to current working directory if not set
	if c.basePath == "" {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		c.basePath = dir
	}

	// Load default .env file first
	defaultEnvPath := filepath.Join(c.basePath, ".env")
	if _, err := os.Stat(defaultEnvPath); err == nil {
		if err := c.loadEnvFile(defaultEnvPath); err != nil {
			return err
		}
	}

	// Check if GOE_ENV is set in the default .env file or environment
	if envVal := c.Get("GOE_ENV"); envVal != "" {
		c.env = envVal
	} else if envVal := os.Getenv("GOE_ENV"); envVal != "" {
		c.env = envVal
		c.Set("GOE_ENV", envVal)
	}

	// Load environment-specific .env file
	envFile := filepath.Join(c.basePath, "."+c.env+".env")
	if _, err := os.Stat(envFile); err == nil {
		if err := c.loadEnvFile(envFile); err != nil {
			return err
		}
	}

	// Load environment variables
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			c.Set(pair[0], pair[1])
		}
	}

	return nil
}

// Start starts the config module
func (c *Config) Start(ctx context.Context) error {
	return nil
}

// Stop stops the config module
func (c *Config) Stop(ctx context.Context) error {
	return nil
}

// Get retrieves a configuration value as a string
func (c *Config) Get(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.values[key]
}

// GetDefault retrieves a configuration value as a string with a default value
func (c *Config) GetDefault(key string, defaultValue string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if val, ok := c.values[key]; ok {
		return val
	}
	return defaultValue
}

// GetInt retrieves a configuration value as an integer
func (c *Config) GetInt(key string) (int, error) {
	// Implementation will be added
	return 0, nil
}

// GetIntDefault retrieves a configuration value as an integer with a default value
func (c *Config) GetIntDefault(key string, defaultValue int) int {
	// Implementation will be added
	return defaultValue
}

// GetBool retrieves a configuration value as a boolean
func (c *Config) GetBool(key string) (bool, error) {
	// Implementation will be added
	return false, nil
}

// GetBoolDefault retrieves a configuration value as a boolean with a default value
func (c *Config) GetBoolDefault(key string, defaultValue bool) bool {
	// Implementation will be added
	return defaultValue
}

// GetFloat retrieves a configuration value as a float64
func (c *Config) GetFloat(key string) (float64, error) {
	// Implementation will be added
	return 0, nil
}

// GetFloatDefault retrieves a configuration value as a float64 with a default value
func (c *Config) GetFloatDefault(key string, defaultValue float64) float64 {
	// Implementation will be added
	return defaultValue
}

// GetDuration retrieves a configuration value as a duration
func (c *Config) GetDuration(key string) (time.Duration, error) {
	// Implementation will be added
	return 0, nil
}

// GetDurationDefault retrieves a configuration value as a duration with a default value
func (c *Config) GetDurationDefault(key string, defaultValue time.Duration) time.Duration {
	// Implementation will be added
	return defaultValue
}

// GetStringSlice retrieves a configuration value as a string slice
func (c *Config) GetStringSlice(key string, separator string) []string {
	// Implementation will be added
	return nil
}

// Has checks if a configuration key exists
func (c *Config) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.values[key]
	return ok
}

// Set sets a configuration value
func (c *Config) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = fmt.Sprintf("%v", value)
}

// Load loads configuration from a specific source
func (c *Config) Load(ctx context.Context) error {
	return c.Initialize(ctx)
}

// Reload reloads configuration from all sources
func (c *Config) Reload(ctx context.Context) error {
	c.mu.Lock()
	c.values = make(map[string]string)
	c.mu.Unlock()
	return c.Initialize(ctx)
}

// loadEnvFile loads environment variables from a file
func (c *Config) loadEnvFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) > 1 && (value[0] == '"' && value[len(value)-1] == '"' || value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}

		c.Set(key, value)
	}

	return nil
}

// Provider provides a Config instance
func Provider() contract.Config {
	return New()
}
