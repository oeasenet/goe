package goe

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
)

// accessorProbe is built by a provider and consumed by an invoker so that fx
// executes both while New() is still on the stack. Each side reads the global
// accessors instead of taking fx-injected parameters — the usage pattern
// reported to freeze consumer apps on v2.2.0.
type accessorProbe struct {
	logger contract.Logger
	config contract.Config
}

// TestAccessorsUsableDuringNew guards against goe.New() holding instance.mu
// across app.Register, which runs the fx graph synchronously: any goe.*
// accessor called from a module constructor, provider, or invoker then blocks
// on RLock of the write-locked mutex and the process freezes with no output.
func TestAccessorsUsableDuringNew(t *testing.T) {
	resetGlobalInstance()

	var (
		fromModuleCtor contract.Logger
		fromInvoker    contract.Logger
		probe          *accessorProbe
	)

	done := make(chan contract.Application, 1)
	go func() {
		done <- New(Options{
			Modules: []any{
				func(logger contract.Logger, config contract.Config) contract.Module {
					fromModuleCtor = Log()
					return &testModule{name: "accessor-probe-module"}
				},
			},
			Providers: []any{
				func() *accessorProbe {
					return &accessorProbe{logger: Log(), config: Config()}
				},
			},
			Invokers: []any{
				func(p *accessorProbe) {
					probe = p
					fromInvoker = Log()
					_ = App()
				},
			},
		})
	}()

	select {
	case app := <-done:
		require.NotNil(t, app)
	case <-time.After(15 * time.Second):
		t.Fatal("goe.New() deadlocked: goe.* accessor called from user code during New() blocked on instance.mu")
	}

	assert.NotNil(t, fromModuleCtor, "Log() inside a module constructor should return the logger")
	assert.NotNil(t, fromInvoker, "Log() inside an invoker should return the logger")
	require.NotNil(t, probe, "provider should have run during New()")
	assert.NotNil(t, probe.logger, "Log() inside a provider should return the logger")
	assert.NotNil(t, probe.config, "Config() inside a provider should return the config")
}

// TestDisabledModuleAccessorPanicsDuringNew pins the failure mode for the
// other half of the footgun: an accessor for a module that is not enabled,
// called from an invoker, must fail with the descriptive panic — not freeze.
func TestDisabledModuleAccessorPanicsDuringNew(t *testing.T) {
	resetGlobalInstance()

	var recovered any
	done := make(chan struct{})
	go func() {
		defer close(done)
		New(Options{
			Invokers: []any{
				func() {
					defer func() { recovered = recover() }()
					_ = DB() // WithDB is not set
				},
			},
		})
	}()

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("goe.New() deadlocked instead of panicking for a disabled module accessor")
	}

	require.NotNil(t, recovered, "DB() without WithDB should panic, not hang or return nil")
	assert.Contains(t, recovered.(string), "DB module not initialized")
}
