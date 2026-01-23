package health

import (
	"time"

	"go.oease.dev/goe/v2/contract"
)

// Config holds configuration for the health module
type Config struct {
	config contract.Config
}

// NewConfig creates a new health configuration from the application config
func NewConfig(cfg contract.Config) *Config {
	return &Config{config: cfg}
}

// Enabled returns whether health checks are enabled
func (c *Config) Enabled() bool {
	if c.config.Has("HEALTH_ENABLED") {
		return c.config.GetBool("HEALTH_ENABLED")
	}
	return true // Enabled by default when WithHealth is set
}

// Path returns the base path for health endpoints
func (c *Config) Path() string {
	path := c.config.GetString("HEALTH_PATH")
	if path == "" {
		return "/health"
	}
	return path
}

// LivenessPath returns the path for liveness probe
func (c *Config) LivenessPath() string {
	path := c.config.GetString("HEALTH_LIVENESS_PATH")
	if path == "" {
		return "/health/live"
	}
	return path
}

// ReadinessPath returns the path for readiness probe
func (c *Config) ReadinessPath() string {
	path := c.config.GetString("HEALTH_READINESS_PATH")
	if path == "" {
		return "/health/ready"
	}
	return path
}

// Timeout returns the maximum time for a health check
func (c *Config) Timeout() time.Duration {
	timeout := c.config.GetDuration("HEALTH_TIMEOUT")
	if timeout == 0 {
		return 5 * time.Second
	}
	return timeout
}

// ToContractConfig converts to contract.HealthConfig
func (c *Config) ToContractConfig() contract.HealthConfig {
	return contract.HealthConfig{
		Enabled:       c.Enabled(),
		Path:          c.Path(),
		LivenessPath:  c.LivenessPath(),
		ReadinessPath: c.ReadinessPath(),
		Timeout:       c.Timeout(),
	}
}
