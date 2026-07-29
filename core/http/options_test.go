package http

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// fakeConfig is a map-backed contract.Config. The precedence tests need
// per-key values, which reads far better here than layered testify matchers.
type fakeConfig struct {
	mu     sync.RWMutex
	values map[string]any
}

func newFakeConfig(values map[string]any) *fakeConfig {
	if values == nil {
		values = map[string]any{}
	}
	return &fakeConfig{values: values}
}

func (c *fakeConfig) get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.values[key]
	return v, ok
}

func (c *fakeConfig) Get(key string) any {
	v, _ := c.get(key)
	return v
}

func (c *fakeConfig) GetString(key string) string {
	if v, ok := c.get(key); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (c *fakeConfig) GetInt(key string) int {
	if v, ok := c.get(key); ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return 0
}

func (c *fakeConfig) GetInt64(key string) int64 { return int64(c.GetInt(key)) }

func (c *fakeConfig) GetFloat64(key string) float64 {
	if v, ok := c.get(key); ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

func (c *fakeConfig) GetBool(key string) bool {
	if v, ok := c.get(key); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func (c *fakeConfig) GetDuration(key string) time.Duration {
	if v, ok := c.get(key); ok {
		if d, ok := v.(time.Duration); ok {
			return d
		}
	}
	return 0
}

func (c *fakeConfig) GetStringSlice(key string) []string {
	if v, ok := c.get(key); ok {
		if s, ok := v.([]string); ok {
			return s
		}
	}
	return nil
}

func (c *fakeConfig) GetStringMap(key string) map[string]any {
	if v, ok := c.get(key); ok {
		if m, ok := v.(map[string]any); ok {
			return m
		}
	}
	return nil
}

func (c *fakeConfig) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
}

func (c *fakeConfig) Has(key string) bool {
	_, ok := c.get(key)
	return ok
}

func (c *fakeConfig) All() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]any, len(c.values))
	for k, v := range c.values {
		out[k] = v
	}
	return out
}

func (c *fakeConfig) Reload() error { return nil }

// newOptionsLogger returns a MockLogger that accepts every log call and records
// them, so tests can assert on emitted warnings.
func newOptionsLogger() *MockLogger {
	logger := &MockLogger{}
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()
	logger.On("Fatal", mock.Anything, mock.Anything).Return()
	zapLogger, _ := zap.NewDevelopment()
	logger.On("GetLogger").Return(zapLogger.Sugar())
	return logger
}

// warnings returns every message passed to logger.Warn.
func warnings(logger *MockLogger) []string {
	var out []string
	for _, call := range logger.Calls {
		if call.Method == "Warn" && len(call.Arguments) > 0 {
			if msg, ok := call.Arguments[0].(string); ok {
				out = append(out, msg)
			}
		}
	}
	return out
}

// newTestKernel builds a kernel from the given env values and options.
func newTestKernel(t *testing.T, env map[string]any, opts ...Option) (*kernel, *MockLogger) {
	t.Helper()
	logger := newOptionsLogger()
	k := New(newFakeConfig(env), logger, opts...)
	return k.(*kernel), logger
}

func TestOptions_Precedence(t *testing.T) {
	t.Run("code wins over env", func(t *testing.T) {
		k, _ := newTestKernel(t,
			map[string]any{"FIBER_BODY_LIMIT": 8 << 20},
			WithBodyLimit(16<<20),
		)
		assert.Equal(t, 16<<20, k.App().Config().BodyLimit)
	})

	t.Run("env fills gaps code leaves alone", func(t *testing.T) {
		k, _ := newTestKernel(t,
			map[string]any{"FIBER_BODY_LIMIT": 8 << 20, "FIBER_CONCURRENCY": 512},
			WithBodyLimit(16<<20),
		)
		cfg := k.App().Config()
		assert.Equal(t, 16<<20, cfg.BodyLimit, "code-set field")
		assert.Equal(t, 512, cfg.Concurrency, "env-set field")
	})

	t.Run("defaults apply when neither sets a value", func(t *testing.T) {
		k, _ := newTestKernel(t, nil)
		cfg := k.App().Config()
		assert.Equal(t, 4<<20, cfg.BodyLimit)
		assert.Equal(t, "Goe", cfg.ServerHeader)
		assert.Equal(t, 10*time.Second, cfg.ReadTimeout)
		assert.Equal(t, 30*time.Second, cfg.IdleTimeout)
		assert.True(t, cfg.StreamRequestBody)
		assert.True(t, cfg.PassLocalsToContext)
	})

	t.Run("code wins for booleans env set to true", func(t *testing.T) {
		// Booleans are the case where "unset" and "false" are easy to confuse,
		// so an option must be able to turn an env-enabled flag back off.
		k, _ := newTestKernel(t,
			map[string]any{"FIBER_STRICT_ROUTING": true},
			WithStrictRouting(false),
		)
		assert.False(t, k.App().Config().StrictRouting)
	})

	t.Run("host and port", func(t *testing.T) {
		k, _ := newTestKernel(t,
			map[string]any{"HTTP_HOST": "127.0.0.1", "HTTP_PORT": 9999},
			WithPort(3000),
		)
		assert.Equal(t, "127.0.0.1", k.host, "env fills host")
		assert.Equal(t, 3000, k.port, "code wins for port")
		assert.Equal(t, "127.0.0.1:3000", k.addr())
	})
}

func TestOptions_Ordering(t *testing.T) {
	t.Run("last option wins", func(t *testing.T) {
		k, _ := newTestKernel(t, nil,
			WithBodyLimit(1<<20),
			WithBodyLimit(2<<20),
		)
		assert.Equal(t, 2<<20, k.App().Config().BodyLimit)
	})

	t.Run("escape hatch runs after options regardless of position", func(t *testing.T) {
		// The hook is listed first but must still win, otherwise "developer
		// always wins" would depend on argument order.
		k, _ := newTestKernel(t, nil,
			WithFiberConfig(func(c *fiber.Config) { c.BodyLimit = 99 }),
			WithBodyLimit(2<<20),
		)
		assert.Equal(t, 99, k.App().Config().BodyLimit)
	})

	t.Run("escape hatch overrides materialised views", func(t *testing.T) {
		k, _ := newTestKernel(t, nil,
			WithHTMLViews("./views", ".gohtml"),
			WithFiberConfig(func(c *fiber.Config) { c.Views = nil }),
		)
		assert.Nil(t, k.App().Config().Views,
			"hooks run after materialisation, so clearing Views must stick")
	})

	t.Run("multiple hooks apply in order", func(t *testing.T) {
		k, _ := newTestKernel(t, nil,
			WithFiberConfig(func(c *fiber.Config) { c.ServerHeader = "first" }),
			WithFiberConfig(func(c *fiber.Config) { c.ServerHeader = "second" }),
		)
		assert.Equal(t, "second", k.App().Config().ServerHeader)
	})
}

func TestOptions_AllOrNothing(t *testing.T) {
	k, _ := newTestKernel(t,
		map[string]any{"FIBER_BODY_LIMIT": 8 << 20},
		WithBodyLimit(16<<20),    // valid
		WithConcurrency(-1),      // invalid
		WithServerHeader("nope"), // valid
	)

	assert.NotEmpty(t, k.optErrs, "the invalid option must be recorded")

	cfg := k.App().Config()
	assert.Equal(t, 8<<20, cfg.BodyLimit,
		"a failed option discards every option, leaving the env value")
	assert.Equal(t, "Goe", cfg.ServerHeader,
		"valid options are discarded alongside the invalid one")

	module := &Module{kernel: k}
	err := module.ValidateConfig()
	require.Error(t, err, "option errors must fail startup")
	assert.Contains(t, err.Error(), "WithConcurrency")
}

func TestOptions_AllErrorsReportedTogether(t *testing.T) {
	k, _ := newTestKernel(t, nil,
		WithConcurrency(-1),
		WithPort(70000),
	)
	err := errors.Join(k.optErrs...)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "WithConcurrency")
	assert.Contains(t, err.Error(), "WithPort")
}

func TestOptions_CrossFieldValidation(t *testing.T) {
	t.Run("trust proxy config without trust proxy", func(t *testing.T) {
		// Fiber silently ignores TrustProxyConfig unless TrustProxy is on,
		// which is exactly the kind of thing that is only noticed in production.
		k, _ := newTestKernel(t, nil,
			WithTrustProxyConfig(fiber.TrustProxyConfig{Proxies: []string{"10.0.0.0/8"}}),
		)
		require.NotEmpty(t, k.optErrs)
		assert.Contains(t, errors.Join(k.optErrs...).Error(), "WithTrustProxy(true)")
	})

	t.Run("trust proxy config with trust proxy is accepted", func(t *testing.T) {
		k, _ := newTestKernel(t, nil,
			WithTrustProxy(true),
			WithTrustProxyConfig(fiber.TrustProxyConfig{Proxies: []string{"10.0.0.0/8"}}),
		)
		assert.Empty(t, k.optErrs)
		assert.True(t, k.App().Config().TrustProxy)
	})

	t.Run("certificate without key", func(t *testing.T) {
		k, _ := newTestKernel(t, nil, WithCertFile("/tmp/cert.pem"))
		require.NotEmpty(t, k.optErrs)
		assert.Contains(t, errors.Join(k.optErrs...).Error(), "WithCertKeyFile")
	})

	t.Run("key without certificate", func(t *testing.T) {
		k, _ := newTestKernel(t, nil, WithCertKeyFile("/tmp/key.pem"))
		require.NotEmpty(t, k.optErrs)
		assert.Contains(t, errors.Join(k.optErrs...).Error(), "WithCertFile")
	})
}

func TestOptions_ClobberWarning(t *testing.T) {
	t.Run("warns but does not override", func(t *testing.T) {
		k, logger := newTestKernel(t, nil,
			WithFiberConfig(func(c *fiber.Config) {
				c.PassLocalsToContext = false
				c.StructValidator = nil
			}),
		)

		cfg := k.App().Config()
		assert.False(t, cfg.PassLocalsToContext, "the developer's choice must stand")
		assert.Nil(t, cfg.StructValidator, "the developer's choice must stand")

		msgs := strings.Join(warnings(logger), "\n")
		assert.Contains(t, msgs, "PassLocalsToContext")
		assert.Contains(t, msgs, "WithReqCtx")
		assert.Contains(t, msgs, "StructValidator")
	})

	t.Run("silent when nothing load-bearing changed", func(t *testing.T) {
		_, logger := newTestKernel(t, nil,
			WithFiberConfig(func(c *fiber.Config) { c.ServerHeader = "custom" }),
		)
		assert.Empty(t, warnings(logger))
	})

	t.Run("silent when replacing rather than clearing", func(t *testing.T) {
		_, logger := newTestKernel(t, nil,
			WithErrorHandler(func(fiber.Ctx, error) error { return nil }),
		)
		assert.Empty(t, warnings(logger),
			"replacing the error handler is supported, not a mistake")
	})
}

func TestOptions_GOELevel(t *testing.T) {
	t.Run("request id disabled", func(t *testing.T) {
		k, _ := newTestKernel(t, nil, WithRequestID(false))
		k.App().Get("/", func(c fiber.Ctx) error { return c.SendString("ok") })

		resp, err := k.App().Test(httptest.NewRequest("GET", "/", nil))
		require.NoError(t, err)
		assert.Empty(t, resp.Header.Get(fiber.HeaderXRequestID))
	})

	t.Run("custom request id header", func(t *testing.T) {
		k, _ := newTestKernel(t, nil, WithRequestIDHeader("X-Trace-Id"))
		k.App().Get("/", func(c fiber.Ctx) error { return c.SendString("ok") })

		resp, err := k.App().Test(httptest.NewRequest("GET", "/", nil))
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Header.Get("X-Trace-Id"))
	})

	t.Run("html views engine is constructed", func(t *testing.T) {
		k, _ := newTestKernel(t, nil, WithHTMLViews("./views", ".gohtml"))
		assert.NotNil(t, k.App().Config().Views)
	})

	t.Run("explicit views engine beats html views", func(t *testing.T) {
		engine := &stubViews{}
		k, _ := newTestKernel(t, nil,
			WithHTMLViews("./views", ".gohtml"),
			WithViews(engine),
		)
		assert.Same(t, engine, k.App().Config().Views)
	})

	t.Run("unix listener network uses host as socket path", func(t *testing.T) {
		k, _ := newTestKernel(t, nil,
			WithListenerNetwork(fiber.NetworkUnix),
			WithHost("/tmp/goe-test.sock"),
		)
		assert.Equal(t, "/tmp/goe-test.sock", k.addr())
	})
}

// stubViews is a minimal fiber.Views implementation for identity assertions.
type stubViews struct{}

func (*stubViews) Load() error { return nil }

func (*stubViews) Render(_ io.Writer, _ string, _ any, _ ...string) error { return nil }

func TestOptions_ListenerNetworkDefaultsToDualStack(t *testing.T) {
	// Fiber defaults ListenerNetwork to tcp4. GOE bound dual-stack before
	// options existed, so the default must stay tcp or IPv6 clients silently
	// stop being served.
	k, _ := newTestKernel(t, nil)
	assert.Equal(t, fiber.NetworkTCP, k.listenCfg.ListenerNetwork)

	k, _ = newTestKernel(t, nil, WithListenerNetwork(fiber.NetworkTCP4))
	assert.Equal(t, fiber.NetworkTCP4, k.listenCfg.ListenerNetwork, "override must still work")
}

func TestOptions_ListenPathAcceptsIPv6(t *testing.T) {
	var boundAddr net.Addr
	k, _ := newTestKernel(t, nil,
		WithHost("::1"),
		WithPort(0),
		WithDisableStartupMessage(true),
		WithListenerAddrFunc(func(a net.Addr) { boundAddr = a }),
	)
	require.Empty(t, k.optErrs)
	assert.Equal(t, "[::1]:0", k.addr(), "IPv6 hosts must be bracketed")

	k.App().Get("/ping", func(c fiber.Ctx) error { return c.SendString("pong") })

	module := &Module{kernel: k}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	require.NoError(t, module.OnStart(ctx))
	t.Cleanup(func() { _ = module.OnStop(context.Background()) })
	require.NotNil(t, boundAddr)

	resp, err := http.Get("http://" + boundAddr.String() + "/ping")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestOptions_TLSMinVersionRejectsValuesFiberPanicsOn(t *testing.T) {
	// Fiber panics for anything other than TLS 1.2/1.3, so these must be
	// rejected as startup errors rather than reaching Fiber.
	for _, version := range []uint16{tls.VersionTLS10, tls.VersionTLS11, 0x9999} {
		k, _ := newTestKernel(t, nil, WithTLSMinVersion(version))
		require.NotEmptyf(t, k.optErrs, "version %#04x must be rejected", version)
		assert.Contains(t, errors.Join(k.optErrs...).Error(), "WithTLSMinVersion")
	}

	for _, version := range []uint16{tls.VersionTLS12, tls.VersionTLS13} {
		k, _ := newTestKernel(t, nil, WithTLSMinVersion(version))
		assert.Emptyf(t, k.optErrs, "version %#04x must be accepted", version)
	}
}

func TestOptions_ListenPathStartsAndStops(t *testing.T) {
	var boundAddr net.Addr
	k, _ := newTestKernel(t, nil,
		WithHost("127.0.0.1"),
		WithPort(0), // let the OS choose, so the test cannot race another listener
		WithDisableStartupMessage(true),
		WithListenerAddrFunc(func(a net.Addr) { boundAddr = a }),
	)
	k.App().Get("/ping", func(c fiber.Ctx) error { return c.SendString("pong") })

	module := &Module{kernel: k}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	require.NoError(t, module.OnStart(ctx))
	t.Cleanup(func() { _ = module.OnStop(context.Background()) })

	require.NotNil(t, boundAddr, "the developer's ListenerAddrFunc must still be called")

	// OnStart returning nil must mean the port is already accepting: connect
	// immediately, with no retry loop to paper over a race.
	resp, err := http.Get("http://" + boundAddr.String() + "/ping")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestOptions_ListenPathReportsBindFailure(t *testing.T) {
	// Occupy a port, then ask the server to bind the same one.
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = occupied.Close() }()

	port := occupied.Addr().(*net.TCPAddr).Port

	k, _ := newTestKernel(t, nil,
		WithHost("127.0.0.1"),
		WithPort(port),
		WithDisableStartupMessage(true),
	)

	module := &Module{kernel: k}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = module.OnStart(ctx)
	require.Error(t, err, "a bind failure must fail startup rather than be logged and swallowed")
	assert.Contains(t, err.Error(), fmt.Sprintf("%d", port))
}

func TestOptions_ListenPathServesTLS(t *testing.T) {
	certFile, keyFile, pool := generateSelfSignedCert(t)

	var boundAddr net.Addr
	k, _ := newTestKernel(t, nil,
		WithHost("127.0.0.1"),
		WithPort(0),
		WithDisableStartupMessage(true),
		WithCertFile(certFile),
		WithCertKeyFile(keyFile),
		WithTLSMinVersion(tls.VersionTLS12),
		WithListenerAddrFunc(func(a net.Addr) { boundAddr = a }),
	)
	require.Empty(t, k.optErrs)
	k.App().Get("/secure", func(c fiber.Ctx) error { return c.SendString("tls ok") })

	module := &Module{kernel: k}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	require.NoError(t, module.OnStart(ctx))
	t.Cleanup(func() { _ = module.OnStop(context.Background()) })
	require.NotNil(t, boundAddr)

	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{
			RootCAs:    pool,
			ServerName: "localhost",
			MinVersion: tls.VersionTLS12,
		}},
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("https://127.0.0.1:%d/secure", boundAddr.(*net.TCPAddr).Port)
	resp, err := client.Get(url)
	require.NoError(t, err, "GOE had no TLS path before ListenConfig was wired up")
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, resp.TLS.HandshakeComplete)
}

// generateSelfSignedCert writes a throwaway certificate/key pair for the TLS
// test and returns their paths plus a pool trusting the certificate.
func generateSelfSignedCert(t *testing.T) (certPath, keyPath string, pool *x509.CertPool) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "localhost"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	require.NoError(t, err)

	dir := t.TempDir()
	certPath = filepath.Join(dir, "cert.pem")
	keyPath = filepath.Join(dir, "key.pem")

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	require.NoError(t, os.WriteFile(certPath, certPEM, 0o600))

	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	require.NoError(t, os.WriteFile(keyPath, keyPEM, 0o600))

	pool = x509.NewCertPool()
	require.True(t, pool.AppendCertsFromPEM(certPEM))

	return certPath, keyPath, pool
}
