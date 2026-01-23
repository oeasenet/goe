package metrics

import (
	"go.oease.dev/goe/v2/contract"
)

// Config holds configuration for the metrics module
type Config struct {
	config contract.Config
}

// NewConfig creates a new metrics configuration from the application config
func NewConfig(cfg contract.Config) *Config {
	return &Config{config: cfg}
}

// Enabled returns whether metrics collection is enabled
func (c *Config) Enabled() bool {
	if c.config.Has("METRICS_ENABLED") {
		return c.config.GetBool("METRICS_ENABLED")
	}
	return true // Enabled by default when WithMetrics is set
}

// Path returns the path for the metrics endpoint
func (c *Config) Path() string {
	path := c.config.GetString("METRICS_PATH")
	if path == "" {
		return "/metrics"
	}
	return path
}

// Namespace returns the prefix for all metrics
func (c *Config) Namespace() string {
	ns := c.config.GetString("METRICS_NAMESPACE")
	if ns == "" {
		return "goe"
	}
	return ns
}

// Subsystem returns the optional subsystem name
func (c *Config) Subsystem() string {
	return c.config.GetString("METRICS_SUBSYSTEM")
}

// EnableGoMetrics returns whether Go runtime metrics are enabled
func (c *Config) EnableGoMetrics() bool {
	if c.config.Has("METRICS_GO_ENABLED") {
		return c.config.GetBool("METRICS_GO_ENABLED")
	}
	return true // Enabled by default
}

// EnableProcessMetrics returns whether process metrics are enabled
func (c *Config) EnableProcessMetrics() bool {
	if c.config.Has("METRICS_PROCESS_ENABLED") {
		return c.config.GetBool("METRICS_PROCESS_ENABLED")
	}
	return true // Enabled by default
}

// HTTPRequestDurationBuckets returns histogram buckets for HTTP request duration
func (c *Config) HTTPRequestDurationBuckets() []float64 {
	// Could be configured via METRICS_HTTP_BUCKETS but for now use defaults
	return contract.DefaultHTTPRequestDurationBuckets()
}

// ToContractConfig converts to contract.MetricsConfig
func (c *Config) ToContractConfig() contract.MetricsConfig {
	return contract.MetricsConfig{
		Enabled:                    c.Enabled(),
		Path:                       c.Path(),
		Namespace:                  c.Namespace(),
		Subsystem:                  c.Subsystem(),
		EnableGoMetrics:            c.EnableGoMetrics(),
		EnableProcessMetrics:       c.EnableProcessMetrics(),
		HTTPRequestDurationBuckets: c.HTTPRequestDurationBuckets(),
	}
}
