package shutdown

import (
	"time"

	"go.oease.dev/goe/v2/contract"
)

// Config holds configuration for the shutdown module
type Config struct {
	config contract.Config
}

// NewConfig creates a new shutdown configuration from the application config
func NewConfig(cfg contract.Config) *Config {
	return &Config{config: cfg}
}

// Timeout returns the total shutdown timeout
func (c *Config) Timeout() time.Duration {
	timeout := c.config.GetDuration("SHUTDOWN_TIMEOUT")
	if timeout == 0 {
		return 30 * time.Second
	}
	return timeout
}

// DrainTimeout returns the HTTP drain timeout
func (c *Config) DrainTimeout() time.Duration {
	timeout := c.config.GetDuration("SHUTDOWN_DRAIN_TIMEOUT")
	if timeout == 0 {
		return 5 * time.Second
	}
	return timeout
}
