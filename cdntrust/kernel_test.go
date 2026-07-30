package cdntrust

import (
	"context"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
	goehttp "go.oease.dev/goe/v2/core/http"
	"go.uber.org/zap"
)

// applyToFiberConfig runs the options through the real HTTP kernel and returns
// the resulting fiber.Config.
//
// Options carry an unexported settings type, so they can only be applied by
// goehttp.New. Going through it is the point: it proves the pair Options returns
// is actually accepted, rather than tripping the kernel's own rule that
// TrustProxyConfig without TrustProxy(true) is a misconfiguration.
func applyToFiberConfig(t *testing.T, opts []goehttp.Option) fiber.Config {
	t.Helper()

	kernel := goehttp.New(emptyConfig{}, discardLogger{}, opts...)

	// Option failures are reported through the module's config validation, so a
	// rejected combination shows up here rather than as a bad Config.
	require.NoError(t, goehttp.NewModule(emptyConfig{}, discardLogger{}, opts...).ValidateConfig(),
		"the kernel rejected the options cdntrust produced")

	return kernel.App().Config()
}

// emptyConfig is a contract.Config with nothing set, so the kernel falls back to
// GOE defaults and only the options under test affect the result.
type emptyConfig struct{}

func (emptyConfig) Get(string) any                     { return nil }
func (emptyConfig) GetString(string) string            { return "" }
func (emptyConfig) GetInt(string) int                  { return 0 }
func (emptyConfig) GetInt64(string) int64              { return 0 }
func (emptyConfig) GetFloat64(string) float64          { return 0 }
func (emptyConfig) GetBool(string) bool                { return false }
func (emptyConfig) GetDuration(string) time.Duration   { return 0 }
func (emptyConfig) GetStringSlice(string) []string     { return nil }
func (emptyConfig) GetStringMap(string) map[string]any { return nil }
func (emptyConfig) Set(string, any)                    {}
func (emptyConfig) Has(string) bool                    { return false }
func (emptyConfig) All() map[string]any                { return nil }
func (emptyConfig) Reload() error                      { return nil }

// discardLogger satisfies contract.Logger without emitting anything.
type discardLogger struct{}

func (discardLogger) Debug(string, ...any)  {}
func (discardLogger) Info(string, ...any)   {}
func (discardLogger) Warn(string, ...any)   {}
func (discardLogger) Error(string, ...any)  {}
func (discardLogger) Fatal(string, ...any)  {}
func (discardLogger) Debugf(string, ...any) {}
func (discardLogger) Infof(string, ...any)  {}
func (discardLogger) Warnf(string, ...any)  {}
func (discardLogger) Errorf(string, ...any) {}
func (discardLogger) Fatalf(string, ...any) {}
func (discardLogger) Panicf(string, ...any) {}
func (discardLogger) Debugw(string, ...any) {}
func (discardLogger) Infow(string, ...any)  {}
func (discardLogger) Warnw(string, ...any)  {}
func (discardLogger) Errorw(string, ...any) {}
func (discardLogger) Fatalw(string, ...any) {}

func (d discardLogger) With(...any) contract.Logger                 { return d }
func (d discardLogger) WithContext(context.Context) contract.Logger { return d }
func (d discardLogger) WithError(error) contract.Logger             { return d }
func (discardLogger) GetLogger() *zap.SugaredLogger                 { return zap.NewNop().Sugar() }
