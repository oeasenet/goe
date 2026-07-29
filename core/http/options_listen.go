package http

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/acme/autocert"
)

// This file exposes every field of fiber.ListenConfig as an Option, except the
// two listed in listenConfigSkipList (see options_coverage_test.go): GOE's
// shutdown manager owns GracefulContext and ShutdownTimeout, so a second knob
// for them would silently fight goe.Options.ShutdownTimeout/DrainTimeout. Both
// remain reachable through WithListenConfig.
//
// Fiber splits server configuration across Config and ListenConfig; GOE does
// not, because that split is an implementation detail a developer should not
// have to know. The two structs share no field names, so a single flat
// With<Field> namespace covers both without ambiguity.

// WithCertFile sets fiber.ListenConfig.CertFile, the TLS certificate to serve.
// Must be paired with WithCertKeyFile.
func WithCertFile(path string) Option {
	return func(s *settings) error {
		if path == "" {
			return errors.New("WithCertFile: path must not be empty")
		}
		s.listen.CertFile = path
		return nil
	}
}

// WithCertKeyFile sets fiber.ListenConfig.CertKeyFile, the private key matching
// WithCertFile.
func WithCertKeyFile(path string) Option {
	return func(s *settings) error {
		if path == "" {
			return errors.New("WithCertKeyFile: path must not be empty")
		}
		s.listen.CertKeyFile = path
		return nil
	}
}

// WithCertClientFile sets fiber.ListenConfig.CertClientFile, the CA bundle used
// to verify client certificates for mutual TLS.
func WithCertClientFile(path string) Option {
	return func(s *settings) error {
		if path == "" {
			return errors.New("WithCertClientFile: path must not be empty")
		}
		s.listen.CertClientFile = path
		return nil
	}
}

// WithTLSConfig sets fiber.ListenConfig.TLSConfig. Supplying one takes
// precedence over WithCertFile/WithCertKeyFile and WithAutoCertManager.
func WithTLSConfig(cfg *tls.Config) Option {
	return func(s *settings) error {
		if cfg == nil {
			return errors.New("WithTLSConfig: config must not be nil")
		}
		s.listen.TLSConfig = cfg
		return nil
	}
}

// WithTLSConfigFunc sets fiber.ListenConfig.TLSConfigFunc, called to adjust the
// TLS configuration Fiber assembled from the certificate options.
func WithTLSConfigFunc(fn func(*tls.Config)) Option {
	return func(s *settings) error {
		if fn == nil {
			return errors.New("WithTLSConfigFunc: function must not be nil")
		}
		s.listen.TLSConfigFunc = fn
		return nil
	}
}

// WithTLSMinVersion sets fiber.ListenConfig.TLSMinVersion.
//
// Only tls.VersionTLS12 and tls.VersionTLS13 are accepted: Fiber panics on any
// other value, so this is rejected here as a startup error instead.
func WithTLSMinVersion(version uint16) Option {
	return func(s *settings) error {
		switch version {
		case tls.VersionTLS12, tls.VersionTLS13:
			s.listen.TLSMinVersion = version
			return nil
		default:
			return fmt.Errorf(
				"WithTLSMinVersion: unsupported version %#04x; Fiber accepts only tls.VersionTLS12 or tls.VersionTLS13",
				version)
		}
	}
}

// WithAutoCertManager sets fiber.ListenConfig.AutoCertManager for automatic
// ACME certificates. Cannot be combined with WithCertFile/WithCertKeyFile.
func WithAutoCertManager(manager *autocert.Manager) Option {
	return func(s *settings) error {
		if manager == nil {
			return errors.New("WithAutoCertManager: manager must not be nil")
		}
		s.listen.AutoCertManager = manager
		return nil
	}
}

// WithListenerNetwork sets fiber.ListenConfig.ListenerNetwork, for example
// fiber.NetworkTCP4 or fiber.NetworkUnix.
func WithListenerNetwork(network string) Option {
	return func(s *settings) error {
		switch network {
		case fiber.NetworkTCP, fiber.NetworkTCP4, fiber.NetworkTCP6, fiber.NetworkUnix:
			s.listen.ListenerNetwork = network
			return nil
		default:
			return fmt.Errorf("WithListenerNetwork: unsupported network %q; use a fiber.Network constant", network)
		}
	}
}

// WithUnixSocketFileMode sets fiber.ListenConfig.UnixSocketFileMode, the
// permissions applied to the socket file when listening on a unix network.
func WithUnixSocketFileMode(mode os.FileMode) Option {
	return func(s *settings) error { s.listen.UnixSocketFileMode = mode; return nil }
}

// WithListenerAddrFunc sets fiber.ListenConfig.ListenerAddrFunc, called with
// the bound address once the listener is created.
//
// GOE does not use this hook internally, so it is yours to use: readiness is
// detected through Fiber's OnListen hook, which fires in prefork masters too.
func WithListenerAddrFunc(fn func(net.Addr)) Option {
	return func(s *settings) error {
		if fn == nil {
			return errors.New("WithListenerAddrFunc: function must not be nil")
		}
		s.listen.ListenerAddrFunc = fn
		return nil
	}
}

// WithBeforeServeFunc sets fiber.ListenConfig.BeforeServeFunc, called after the
// listener is bound but before requests are served. Returning an error aborts
// startup.
func WithBeforeServeFunc(fn func(*fiber.App) error) Option {
	return func(s *settings) error {
		if fn == nil {
			return errors.New("WithBeforeServeFunc: function must not be nil")
		}
		s.listen.BeforeServeFunc = fn
		return nil
	}
}

// WithDisableStartupMessage sets fiber.ListenConfig.DisableStartupMessage,
// suppressing Fiber's ASCII banner. Useful when logs are machine-parsed.
func WithDisableStartupMessage(disabled bool) Option {
	return func(s *settings) error { s.listen.DisableStartupMessage = disabled; return nil }
}

// WithEnablePrintRoutes sets fiber.ListenConfig.EnablePrintRoutes, printing the
// full route table at startup.
func WithEnablePrintRoutes(enabled bool) Option {
	return func(s *settings) error { s.listen.EnablePrintRoutes = enabled; return nil }
}

// WithEnablePrefork sets fiber.ListenConfig.EnablePrefork, forking one child
// process per CPU core, each binding the port with SO_REUSEPORT.
//
// Prefork spawns child processes that re-execute the program, so anything
// created before goe.Run — database pools, background jobs — is created once
// per child. Verify your application tolerates that before enabling it.
func WithEnablePrefork(enabled bool) Option {
	return func(s *settings) error { s.listen.EnablePrefork = enabled; return nil }
}

// WithPreforkLogger sets fiber.ListenConfig.PreforkLogger, used for prefork
// master and child lifecycle messages.
func WithPreforkLogger(logger fiber.PreforkLogger) Option {
	return func(s *settings) error {
		if logger == nil {
			return errors.New("WithPreforkLogger: logger must not be nil")
		}
		s.listen.PreforkLogger = logger
		return nil
	}
}

// WithPreforkRecoverThreshold sets
// fiber.ListenConfig.PreforkRecoverThreshold, how many times a crashed child
// may be respawned before the master gives up.
func WithPreforkRecoverThreshold(n int) Option {
	return func(s *settings) error {
		if n < 0 {
			return fmt.Errorf("WithPreforkRecoverThreshold: value must not be negative, got %d", n)
		}
		s.listen.PreforkRecoverThreshold = n
		return nil
	}
}

// WithPreforkRecoverInterval sets
// fiber.ListenConfig.PreforkRecoverInterval, how long the master waits before
// respawning a crashed child. Added in Fiber v3.4.
func WithPreforkRecoverInterval(d time.Duration) Option {
	return func(s *settings) error {
		if d < 0 {
			return fmt.Errorf("WithPreforkRecoverInterval: value must not be negative, got %s", d)
		}
		s.listen.PreforkRecoverInterval = d
		return nil
	}
}

// WithPreforkShutdownGracePeriod sets
// fiber.ListenConfig.PreforkShutdownGracePeriod, how long the master waits
// after SIGTERM before sending SIGKILL to children. Added in Fiber v3.4.
func WithPreforkShutdownGracePeriod(d time.Duration) Option {
	return func(s *settings) error {
		if d < 0 {
			return fmt.Errorf("WithPreforkShutdownGracePeriod: value must not be negative, got %s", d)
		}
		s.listen.PreforkShutdownGracePeriod = d
		return nil
	}
}
