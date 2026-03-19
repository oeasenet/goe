package otel

import (
	"strings"

	"go.oease.dev/goe/v2/contract"
)

// Config holds configuration for the OpenTelemetry module
type Config struct {
	config contract.Config
}

// NewConfig creates a new OpenTelemetry configuration from the application config
func NewConfig(cfg contract.Config) *Config {
	return &Config{config: cfg}
}

// Enabled returns whether OpenTelemetry is enabled
func (c *Config) Enabled() bool {
	if c.config.Has("OTEL_ENABLED") {
		return c.config.GetBool("OTEL_ENABLED")
	}
	return true // Enabled by default when WithOTel is set
}

// ServiceName returns the service name
func (c *Config) ServiceName() string {
	// Check OTEL_SERVICE_NAME first, then fall back to APP_NAME
	name := c.config.GetString("OTEL_SERVICE_NAME")
	if name == "" {
		name = c.config.GetString("APP_NAME")
	}
	if name == "" {
		return "goe-app"
	}
	return name
}

// ServiceVersion returns the service version
func (c *Config) ServiceVersion() string {
	version := c.config.GetString("OTEL_SERVICE_VERSION")
	if version == "" {
		version = c.config.GetString("APP_VERSION")
	}
	if version == "" {
		return "1.0.0"
	}
	return version
}

// TracesEnabled returns whether traces are enabled
func (c *Config) TracesEnabled() bool {
	if c.config.Has("OTEL_TRACES_ENABLED") {
		return c.config.GetBool("OTEL_TRACES_ENABLED")
	}
	return true // Enabled by default
}

// MetricsEnabled returns whether OTel metrics export is enabled
func (c *Config) MetricsEnabled() bool {
	if c.config.Has("OTEL_METRICS_ENABLED") {
		return c.config.GetBool("OTEL_METRICS_ENABLED")
	}
	return false // Disabled by default (use Prometheus metrics instead)
}

// TraceSampler returns the trace sampling strategy
func (c *Config) TraceSampler() string {
	sampler := c.config.GetString("OTEL_TRACES_SAMPLER")
	if sampler == "" {
		return "parentbased_traceidratio"
	}
	return sampler
}

// TraceSamplerArg returns the argument for ratio-based samplers
func (c *Config) TraceSamplerArg() float64 {
	arg := c.config.GetFloat64("OTEL_TRACES_SAMPLER_ARG")
	if arg <= 0 {
		return 0.1 // Default 10% sampling
	}
	return arg
}

// ExporterType returns the exporter type (otlp, stdout, none)
func (c *Config) ExporterType() string {
	exporter := c.config.GetString("OTEL_EXPORTER_TYPE")
	if exporter == "" {
		exporter = c.config.GetString("OTEL_EXPORTER")
		if exporter == "" {
			return "otlp"
		}
	}
	return exporter
}

// OTLPEndpoint returns the OTLP collector endpoint
func (c *Config) OTLPEndpoint() string {
	endpoint := c.config.GetString("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		return "localhost:4317"
	}
	return endpoint
}

// OTLPProtocol returns the OTLP protocol (grpc, http)
func (c *Config) OTLPProtocol() string {
	protocol := c.config.GetString("OTEL_EXPORTER_OTLP_PROTOCOL")
	if protocol == "" {
		return "grpc"
	}
	return protocol
}

// OTLPInsecure returns whether TLS should be disabled
func (c *Config) OTLPInsecure() bool {
	if c.config.Has("OTEL_EXPORTER_OTLP_INSECURE") {
		return c.config.GetBool("OTEL_EXPORTER_OTLP_INSECURE")
	}
	return true // Insecure by default for local development
}

// OTLPHeaders returns additional headers for OTLP requests
func (c *Config) OTLPHeaders() map[string]string {
	headers := c.config.GetString("OTEL_EXPORTER_OTLP_HEADERS")
	if headers == "" {
		return nil
	}

	result := make(map[string]string)
	pairs := strings.SplitSeq(headers, ",")
	for pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return result
}

// Propagators returns the list of propagators to use
func (c *Config) Propagators() []string {
	propagators := c.config.GetString("OTEL_PROPAGATORS")
	if propagators == "" {
		return []string{"tracecontext", "baggage"}
	}
	return strings.Split(propagators, ",")
}

// ResourceAttributes returns additional resource attributes
func (c *Config) ResourceAttributes() map[string]string {
	attrs := c.config.GetString("OTEL_RESOURCE_ATTRIBUTES")
	if attrs == "" {
		return nil
	}

	result := make(map[string]string)
	pairs := strings.SplitSeq(attrs, ",")
	for pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return result
}

// ToContractConfig converts to contract.OTelConfig
func (c *Config) ToContractConfig() contract.OTelConfig {
	return contract.OTelConfig{
		Enabled:            c.Enabled(),
		ServiceName:        c.ServiceName(),
		ServiceVersion:     c.ServiceVersion(),
		TracesEnabled:      c.TracesEnabled(),
		MetricsEnabled:     c.MetricsEnabled(),
		TraceSampler:       c.TraceSampler(),
		TraceSamplerArg:    c.TraceSamplerArg(),
		ExporterType:       c.ExporterType(),
		OTLPEndpoint:       c.OTLPEndpoint(),
		OTLPProtocol:       c.OTLPProtocol(),
		OTLPInsecure:       c.OTLPInsecure(),
		OTLPHeaders:        c.OTLPHeaders(),
		Propagators:        c.Propagators(),
		ResourceAttributes: c.ResourceAttributes(),
	}
}
