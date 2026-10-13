package goe_test

import (
	"context"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	goehttp "go.oease.dev/goe/v2/core/http"
	"go.oease.dev/goe/v2/core/shutdown"
)

// serverAddr records where the HTTP server bound, so a test can probe whether
// it is accepting connections at a given point of the lifecycle.
type serverAddr struct {
	mu   sync.Mutex
	addr string
}

func (s *serverAddr) set(a net.Addr) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.addr = a.String()
}

func (s *serverAddr) get() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addr
}

// accepting reports whether the server accepts connections right now: false
// before it has bound and after its listener closed.
func (s *serverAddr) accepting() bool {
	addr := s.get()
	if addr == "" {
		return false
	}
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func shutdownTestHTTP(srv *serverAddr) []goehttp.Option {
	return []goehttp.Option{
		goehttp.WithHost("127.0.0.1"),
		goehttp.WithPort(0),
		goehttp.WithDisableStartupMessage(true),
		goehttp.WithListenerAddrFunc(srv.set),
	}
}

func TestShutdown_RegisteredHooksRunOnStop(t *testing.T) {
	var ran atomic.Bool
	app := goe.New(goe.Options{
		Invokers: []any{
			func(m *shutdown.Manager) {
				m.RegisterHook("flush", shutdown.PriorityTelemetry, func(context.Context) error {
					ran.Store(true)
					return nil
				})
			},
		},
	})

	ctx := context.Background()
	require.NoError(t, app.Start(ctx))
	require.NoError(t, app.Stop(ctx))

	assert.True(t, ran.Load(), "a hook registered with the shutdown manager must run when the app stops")
}

func TestShutdown_HooksRunAroundTheHTTPDrainByPriority(t *testing.T) {
	type observation struct{ ran, serving bool }

	srv := &serverAddr{}
	var mu sync.Mutex
	seen := map[string]observation{}
	observe := func(name string) contract.ShutdownHook {
		return func(context.Context) error {
			serving := srv.accepting()
			mu.Lock()
			defer mu.Unlock()
			seen[name] = observation{ran: true, serving: serving}
			return nil
		}
	}

	app := goe.New(goe.Options{
		HTTP: shutdownTestHTTP(srv),
		Invokers: []any{
			func(m *shutdown.Manager) {
				m.RegisterHook("deregister", shutdown.PriorityHTTP+10, observe("deregister"))
				m.RegisterHook("finish-jobs", shutdown.PriorityWorkers, observe("finish-jobs"))
			},
		},
	})

	ctx := context.Background()
	require.NoError(t, app.Start(ctx))
	require.True(t, srv.accepting(), "the server must accept connections once the app started")
	require.NoError(t, app.Stop(ctx))

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, observation{ran: true, serving: true}, seen["deregister"],
		"a hook above PriorityHTTP runs while the server still accepts connections")
	assert.Equal(t, observation{ran: true, serving: false}, seen["finish-jobs"],
		"a hook below PriorityHTTP runs once the server stopped accepting connections")
}

// probeModule records whether the HTTP server accepted connections when the
// module started and when it stopped.
type probeModule struct {
	srv                                *serverAddr
	started, stopped                   atomic.Bool
	servingAtStart, servingWhenStopped atomic.Bool
}

func (m *probeModule) Name() string { return "probe" }

func (m *probeModule) OnStart(context.Context) error {
	m.servingAtStart.Store(m.srv.accepting())
	m.started.Store(true)
	return nil
}

func (m *probeModule) OnStop(context.Context) error {
	m.servingWhenStopped.Store(m.srv.accepting())
	m.stopped.Store(true)
	return nil
}

func TestShutdown_CustomModulesStartBeforeAndStopAfterTheServer(t *testing.T) {
	srv := &serverAddr{}
	probe := &probeModule{srv: srv}

	app := goe.New(goe.Options{
		HTTP: shutdownTestHTTP(srv),
		Modules: []any{
			func(contract.Logger, contract.Config) contract.Module { return probe },
		},
	})

	ctx := context.Background()
	require.NoError(t, app.Start(ctx))
	require.NoError(t, app.Stop(ctx))

	require.True(t, probe.started.Load())
	require.True(t, probe.stopped.Load())
	assert.False(t, probe.servingAtStart.Load(),
		"a custom module must be started before the server accepts requests that may use it")
	assert.False(t, probe.servingWhenStopped.Load(),
		"a custom module must stop only after the server drained the requests that may use it")
}

func TestShutdown_HTTPDrainIsBoundedByDrainTimeout(t *testing.T) {
	srv := &serverAddr{}
	entered := make(chan struct{})
	release := make(chan struct{})

	app := goe.New(goe.Options{
		HTTP:         shutdownTestHTTP(srv),
		DrainTimeout: 200 * time.Millisecond,
		Invokers: []any{
			func(k contract.HTTPKernel) {
				k.App().Get("/slow", func(c fiber.Ctx) error {
					close(entered)
					select {
					case <-release:
					case <-time.After(5 * time.Second):
					}
					return c.SendString("done")
				})
			},
		},
	})

	ctx := context.Background()
	require.NoError(t, app.Start(ctx))

	clientDone := make(chan struct{})
	go func() {
		defer close(clientDone)
		resp, err := http.Get("http://" + srv.get() + "/slow")
		if err == nil {
			_ = resp.Body.Close()
		}
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("the slow request never reached its handler")
	}

	start := time.Now()
	err := app.Stop(ctx)
	elapsed := time.Since(start)
	close(release)
	<-clientDone

	assert.Less(t, elapsed, 2*time.Second,
		"stopping must not wait for an in-flight request beyond the drain timeout")
	assert.ErrorIs(t, err, context.DeadlineExceeded,
		"a drain cut short by its timeout is reported")
}

func TestShutdown_StopTimeoutFollowsShutdownTimeout(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		app := goe.New(goe.Options{})
		assert.Equal(t, 30*time.Second, app.Container().StopTimeout())
	})

	t.Run("option", func(t *testing.T) {
		app := goe.New(goe.Options{ShutdownTimeout: 7 * time.Second})
		assert.Equal(t, 7*time.Second, app.Container().StopTimeout())
	})

	t.Run("negative falls back to the default", func(t *testing.T) {
		// A negative bound would expire the stop at once, skipping every hook.
		app := goe.New(goe.Options{ShutdownTimeout: -time.Second})
		assert.Equal(t, 30*time.Second, app.Container().StopTimeout())
	})
}

func TestShutdown_NegativeDrainTimeoutFallsBackToTheDefault(t *testing.T) {
	var drain time.Duration
	_ = goe.New(goe.Options{
		DrainTimeout: -time.Second,
		Invokers: []any{
			func(m *shutdown.Manager) { drain = m.DrainTimeout() },
		},
	})
	assert.Equal(t, 5*time.Second, drain)
}
