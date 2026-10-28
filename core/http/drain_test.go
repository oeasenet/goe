package http

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newServingModule builds an HTTP module bound to an ephemeral loopback port and
// returns a func that reports the bound address once OnStart returned.
func newServingModule(t *testing.T, opts ...Option) (*Module, *MockLogger, func() string) {
	t.Helper()
	var mu sync.Mutex
	var addr string
	opts = append([]Option{
		WithHost("127.0.0.1"),
		WithPort(0),
		WithDisableStartupMessage(true),
		WithListenerAddrFunc(func(a net.Addr) {
			mu.Lock()
			defer mu.Unlock()
			addr = a.String()
		}),
	}, opts...)
	logger := newOptionsLogger()
	mod := NewModule(newFakeConfig(nil), logger, opts...)
	return mod, logger, func() string {
		mu.Lock()
		defer mu.Unlock()
		return addr
	}
}

// warnFields returns the key/value fields of every Warn call with msg.
func warnFields(logger *MockLogger, msg string) []map[string]any {
	var out []map[string]any
	for _, call := range logger.Calls {
		if call.Method != "Warn" || len(call.Arguments) < 2 || call.Arguments[0] != msg {
			continue
		}
		args, _ := call.Arguments[1].([]any)
		fields := map[string]any{}
		for i := 0; i+1 < len(args); i += 2 {
			if key, ok := args[i].(string); ok {
				fields[key] = args[i+1]
			}
		}
		out = append(out, fields)
	}
	return out
}

func waitFor(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}

func TestModule_DrainOverrunClosesConnectionsAndReturnsNil(t *testing.T) {
	mod, logger, addr := newServingModule(t)
	entered := make(chan struct{})
	mod.Provide().App().Get("/slow", func(c fiber.Ctx) error {
		close(entered)
		time.Sleep(3 * time.Second)
		return c.SendString("done")
	})
	require.NoError(t, mod.OnStart(context.Background()))

	type result struct {
		err   error
		after time.Duration
	}
	got := make(chan result, 1)
	start := time.Now()
	go func() {
		resp, err := http.Get("http://" + addr() + "/slow")
		if err == nil {
			_, err = io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if err == nil {
				err = errors.New("the request completed")
			}
		}
		got <- result{err: err, after: time.Since(start)}
	}()
	waitFor(t, entered, "the slow handler")

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	stopStart := time.Now()
	require.NoError(t, mod.OnStop(ctx), "an incomplete drain is reported, not returned")
	assert.Less(t, time.Since(stopStart), time.Second)

	select {
	case r := <-got:
		assert.Error(t, r.err, "the connection still open at the drain deadline must be closed")
		assert.Less(t, r.after, 2*time.Second, "the client must not wait for the handler to finish")
	case <-time.After(5 * time.Second):
		t.Fatal("the client never saw its connection end")
	}

	fields := warnFields(logger, "HTTP drain incomplete")
	require.Len(t, fields, 1)
	assert.Equal(t, 1, fields[0]["open_connections"])
}

func TestModule_ResponsesDuringDrainCarryConnectionClose(t *testing.T) {
	mod, _, addr := newServingModule(t)
	entered := make(chan struct{})
	mod.Provide().App().Get("/work", func(c fiber.Ctx) error {
		close(entered)
		time.Sleep(300 * time.Millisecond)
		return c.SendString("ok")
	})
	require.NoError(t, mod.OnStart(context.Background()))

	responses := make(chan *http.Response, 1)
	go func() {
		resp, err := http.Get("http://" + addr() + "/work")
		if err != nil {
			responses <- nil
			return
		}
		responses <- resp
	}()
	waitFor(t, entered, "the handler")

	stopped := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		stopped <- mod.OnStop(ctx)
	}()

	resp := <-responses
	require.NotNil(t, resp)
	defer func() { _ = resp.Body.Close() }()
	assert.True(t, resp.Close, "a response written during the drain must carry Connection: close")
	require.NoError(t, <-stopped)
}

func TestModule_DrainWithoutOpenConnectionsIsQuiet(t *testing.T) {
	mod, logger, _ := newServingModule(t)
	require.NoError(t, mod.OnStart(context.Background()))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, mod.OnStop(ctx))
	assert.Empty(t, warnFields(logger, "HTTP drain incomplete"))
}

// OnStart reports the server ready from Fiber's listen hooks, just before
// fasthttp's Serve registers the listener. A stop that lands in between must
// still close it: fasthttp's shutdown finds no listener and returns at once,
// and Serve would then serve on after the app stopped.
func TestModule_StopBeforeServeRegisteredClosesTheListener(t *testing.T) {
	release := make(chan struct{})
	mod, _, addr := newServingModule(t, WithBeforeServeFunc(func(*fiber.App) error {
		<-release // hold Listen between the ready signal and Serve
		return nil
	}))
	shutDown := make(chan struct{})
	mod.Provide().App().Hooks().OnPostShutdown(func(error) error {
		close(shutDown)
		return nil
	})
	require.NoError(t, mod.OnStart(context.Background()))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stopped := make(chan error, 1)
	go func() { stopped <- mod.OnStop(ctx) }()
	<-shutDown     // the shutdown ran before Serve registered the listener
	close(release) // now Serve registers it
	require.NoError(t, <-stopped)

	_, err := net.DialTimeout("tcp", addr(), time.Second)
	assert.Error(t, err, "the listener must be closed once OnStop returned")
}
