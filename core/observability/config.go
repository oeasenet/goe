package observability

import (
	"fmt"
	"strings"
	"time"

	"go.oease.dev/goe/v2/contract"
)

// config implements the ObservabilityConfig interface
type config struct {
	baseConfig contract.Config
}

// NewConfig creates a new observability configuration
func NewConfig(baseConfig contract.Config) contract.ObservabilityConfig {
	return &config{
		baseConfig: baseConfig,
	}
}

// Enabled returns whether observability is enabled
func (c *config) Enabled() bool {
	return c.baseConfig.GetBool("OTEL_ENABLED")
}

// ServiceName returns the service name for tracing
func (c *config) ServiceName() string {
	serviceName := c.baseConfig.GetString("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = c.baseConfig.GetString("APP_NAME")
		if serviceName == "" {
			serviceName = "goe-app"
		}
	}
	return serviceName
}

// ServiceVersion returns the service version
func (c *config) ServiceVersion() string {
	version := c.baseConfig.GetString("OTEL_SERVICE_VERSION")
	if version == "" {
		version = c.baseConfig.GetString("APP_VERSION")
		if version == "" {
			version = "1.0.0"
		}
	}
	return version
}

// Environment returns the environment name
func (c *config) Environment() string {
	env := c.baseConfig.GetString("OTEL_ENVIRONMENT")
	if env == "" {
		env = c.baseConfig.GetString("GOE_ENV")
		if env == "" {
			env = "dev"
		}
	}
	return env
}

// Metrics returns metrics configuration
func (c *config) Metrics() contract.MetricsConfig {
	return &metricsConfig{
		baseConfig: c.baseConfig,
	}
}

// Tracing returns tracing configuration
func (c *config) Tracing() contract.TracingConfig {
	return &tracingConfig{
		baseConfig: c.baseConfig,
	}
}

// metricsConfig implements the MetricsConfig interface
type metricsConfig struct {
	baseConfig contract.Config
}

// Enabled returns whether metrics are enabled
func (m *metricsConfig) Enabled() bool {
	// Check specific metrics flag first
	if m.baseConfig.Has("OTEL_METRICS_ENABLED") {
		return m.baseConfig.GetBool("OTEL_METRICS_ENABLED")
	}
	// Fall back to general observability flag
	return m.baseConfig.GetBool("OTEL_ENABLED")
}

// Endpoint returns the metrics endpoint (for Prometheus)
func (m *metricsConfig) Endpoint() string {
	return m.baseConfig.GetString("OTEL_METRICS_ENDPOINT")
}

// Port returns the metrics server port
func (m *metricsConfig) Port() int {
	port := m.baseConfig.GetInt("OTEL_METRICS_PORT")
	if port == 0 {
		port = 9090 // Default Prometheus port
	}
	return port
}

// Path returns the metrics endpoint path
func (m *metricsConfig) Path() string {
	path := m.baseConfig.GetString("OTEL_METRICS_PATH")
	if path == "" {
		path = "/metrics"
	}
	return path
}

// Interval returns the metrics collection interval
func (m *metricsConfig) Interval() time.Duration {
	interval := m.baseConfig.GetDuration("OTEL_METRICS_INTERVAL")
	if interval == 0 {
		interval = 30 * time.Second
	}
	return interval
}

// Exporters returns the list of enabled exporters
func (m *metricsConfig) Exporters() []string {
	exporters := m.baseConfig.GetStringSlice("OTEL_METRICS_EXPORTERS")
	if len(exporters) == 0 {
		exportersStr := m.baseConfig.GetString("OTEL_METRICS_EXPORTERS")
		if exportersStr != "" {
			exporters = strings.Split(exportersStr, ",")
			for i, exporter := range exporters {
				exporters[i] = strings.TrimSpace(exporter)
			}
		} else {
			exporters = []string{"prometheus"} // Default exporter
		}
	}
	return exporters
}

// tracingConfig implements the TracingConfig interface
type tracingConfig struct {
	baseConfig contract.Config
}

// Enabled returns whether tracing is enabled
func (t *tracingConfig) Enabled() bool {
	// Check specific tracing flag first
	if t.baseConfig.Has("OTEL_TRACING_ENABLED") {
		return t.baseConfig.GetBool("OTEL_TRACING_ENABLED")
	}
	// Fall back to general observability flag
	return t.baseConfig.GetBool("OTEL_ENABLED")
}

// Endpoint returns the tracing endpoint (OTLP)
func (t *tracingConfig) Endpoint() string {
	endpoint := t.baseConfig.GetString("OTEL_TRACING_ENDPOINT")
	if endpoint == "" {
		endpoint = t.baseConfig.GetString("OTEL_EXPORTER_OTLP_ENDPOINT")
		if endpoint == "" {
			endpoint = "localhost:4318" // Default OTLP HTTP endpoint (without http://)
		}
	}
	return endpoint
}

// SamplingRatio returns the sampling ratio (0.0 to 1.0)
func (t *tracingConfig) SamplingRatio() float64 {
	ratio := t.baseConfig.GetFloat64("OTEL_TRACING_SAMPLING_RATIO")
	if ratio == 0 {
		ratio = 1.0 // Default to 100% sampling
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	return ratio
}

// Exporters returns the list of enabled exporters
func (t *tracingConfig) Exporters() []string {
	exporters := t.baseConfig.GetStringSlice("OTEL_TRACING_EXPORTERS")
	if len(exporters) == 0 {
		exportersStr := t.baseConfig.GetString("OTEL_TRACING_EXPORTERS")
		if exportersStr != "" {
			exporters = strings.Split(exportersStr, ",")
			for i, exporter := range exporters {
				exporters[i] = strings.TrimSpace(exporter)
			}
		} else {
			exporters = []string{"otlp"} // Default exporter
		}
	}
	return exporters
}

// Headers returns additional headers for tracing export
func (t *tracingConfig) Headers() map[string]string {
	headers := make(map[string]string)

	// Get all configuration with prefix OTEL_TRACING_HEADERS_
	prefix := "OTEL_TRACING_HEADERS_"
	allConfig := t.baseConfig.All()

	for key, value := range allConfig {
		if len(key) > len(prefix) && strings.HasPrefix(key, prefix) {
			headerKey := strings.ToLower(key[len(prefix):])
			headerKey = strings.ReplaceAll(headerKey, "_", "-")
			headers[headerKey] = fmt.Sprintf("%v", value)
		}
	}

	// Also check for generic OTLP headers
	if len(headers) == 0 {
		headersStr := t.baseConfig.GetString("OTEL_EXPORTER_OTLP_HEADERS")
		if headersStr != "" {
			pairs := strings.Split(headersStr, ",")
			for _, pair := range pairs {
				parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
				if len(parts) == 2 {
					headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	return headers
}
