package metrics

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"go.oease.dev/goe/v2/contract"
)

// Option configures the metrics module from code. Options are applied after
// the METRICS_* environment variables, so a value set in code wins.
type Option func(*settings) error

type settings struct {
	host     string
	port     int
	registry *prometheus.Registry
	pprof    bool
}

func (s *settings) applyEnv(cfg contract.Config) {
	if port := cfg.GetInt("METRICS_PORT"); port != 0 {
		s.port = port
	}
	if host := cfg.GetString("METRICS_HOST"); host != "" {
		s.host = host
	}
	if cfg.Has("METRICS_PPROF") {
		s.pprof = cfg.GetBool("METRICS_PPROF")
	}
}

// WithPort serves the metrics endpoint on its own listener at port instead of
// on the application's HTTP server, which then does not serve it at all.
// Replaces METRICS_PORT. The listener starts with the module and fails startup
// when the port is taken.
func WithPort(port int) Option {
	return func(s *settings) error {
		if port < 1 || port > 65535 {
			return fmt.Errorf("WithPort: port %d is out of range 1-65535", port)
		}
		s.port = port
		return nil
	}
}

// WithHost sets the interface the separate metrics listener binds to. Replaces
// METRICS_HOST. Defaults to all interfaces.
func WithHost(host string) Option {
	return func(s *settings) error { s.host = host; return nil }
}

// WithRegistry makes the module use registry — the application's own —
// instead of creating one: GOE's metrics are registered into it, and the
// endpoint serves everything it holds.
func WithRegistry(registry *prometheus.Registry) Option {
	return func(s *settings) error {
		if registry == nil {
			return errors.New("WithRegistry: registry must not be nil")
		}
		s.registry = registry
		return nil
	}
}

// WithPprof serves Go's profiles under /debug/pprof/ — goroutineleak included
// — on the separate metrics listener. Replaces METRICS_PPROF. Profiles are
// never served on the application's port: without WithPort they are not
// served at all.
func WithPprof(enabled bool) Option {
	return func(s *settings) error { s.pprof = enabled; return nil }
}
