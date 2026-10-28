package http

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// newAccessLogKernel builds a kernel whose access log writes into an in-memory
// observer, so tests can assert the exact fields fiberzap emits per request.
func newAccessLogKernel(t *testing.T) (contract.HTTPKernel, *observer.ObservedLogs) {
	t.Helper()

	config := &MockConfig{}
	config.On("GetString", mock.Anything).Return("")
	config.On("GetBool", mock.Anything).Return(false)
	config.On("GetInt", mock.Anything).Return(0)
	config.On("GetStringSlice", mock.Anything).Return([]string{})
	config.On("GetDuration", mock.Anything).Return(time.Duration(0))
	config.On("Has", mock.Anything).Return(false)

	core, logs := observer.New(zap.DebugLevel)
	logger := &MockLogger{}
	logger.On("Fatal", mock.Anything).Return()
	logger.On("GetLogger").Return(zap.New(core).Sugar())

	return New(config, logger), logs
}

func TestAccessLog_ContractFields(t *testing.T) {
	kernel, logs := newAccessLogKernel(t)
	app := kernel.App()
	app.Get("/users/:id", func(c fiber.Ctx) error {
		return c.SendString("hello!")
	})

	req := httptest.NewRequest("GET", "/users/42?page=1", nil)
	req.Host = "api.example.com"
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	entries := logs.FilterMessage("http request").All()
	require.Len(t, entries, 1)
	fields := entries[0].ContextMap()

	// New contract fields.
	assert.Equal(t, "/users/:id", fields["route"], "route must be the template, not the concrete URL")
	assert.Equal(t, "api.example.com", fields["host"])
	assert.EqualValues(t, len("hello!"), fields["bytesSent"])

	durationMS, ok := fields["duration_ms"].(int64)
	require.True(t, ok, "duration_ms must be an integer, got %T", fields["duration_ms"])
	assert.GreaterOrEqual(t, durationMS, int64(0))
	assert.Less(t, durationMS, int64(60_000), "duration_ms must be measured from the request start, not the zero time")

	// Existing fields must survive.
	assert.Equal(t, "GET", fields["method"])
	assert.EqualValues(t, fiber.StatusOK, fields["status"])
	assert.Contains(t, fields, "latency")
	assert.Equal(t, "/users/42?page=1", fields["url"])
	assert.Contains(t, fields, "ip")
	assert.Contains(t, fields, "request_id")
}

// readStreamedResponse opens a raw HTTP/1.1 connection to addr, requests path and
// records when each "data:" line arrived, returning the arrivals and the raw wire bytes.
func readStreamedResponse(t *testing.T, addr, path string) ([]time.Duration, string) {
	t.Helper()
	conn, err := (&net.Dialer{}).DialContext(t.Context(), "tcp", addr)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()
	start := time.Now()
	_, _ = fmt.Fprintf(conn, "GET %s HTTP/1.1\r\nHost: %s\r\n\r\n", path, addr)
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))
	r := bufio.NewReader(conn)
	var arrivals []time.Duration
	var raw strings.Builder
	for {
		line, err := r.ReadString('\n')
		raw.WriteString(line)
		if strings.HasPrefix(line, "data:") {
			arrivals = append(arrivals, time.Since(start))
		}
		if err != nil || strings.HasPrefix(line, "0\r") { // EOF or the final chunk
			break
		}
	}
	return arrivals, raw.String()
}

// TestAccessLog_StreamedResponseIsNotDrained pins that logging a request whose body is a
// stream (SendStreamWriter — SSE) leaves the stream alone: reading the body for bytesSent
// drains it into a buffer and the client receives every frame at once, as Content-Length.
func TestAccessLog_StreamedResponseIsNotDrained(t *testing.T) {
	kernel, logs := newAccessLogKernel(t)
	app := kernel.App()
	app.Get("/events", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextEventStream)
		return c.SendStreamWriter(func(w *bufio.Writer) {
			for i := range 3 {
				_, _ = fmt.Fprintf(w, "id: %d\nevent: tick\ndata: t%d\n\n", i+1, i+1)
				if err := w.Flush(); err != nil {
					return
				}
				time.Sleep(200 * time.Millisecond)
			}
		})
	})

	ln, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go func() { _ = app.Listener(ln, fiber.ListenConfig{DisableStartupMessage: true}) }()
	t.Cleanup(func() { _ = app.Shutdown() })

	arrivals, raw := readStreamedResponse(t, ln.Addr().String(), "/events")
	require.Len(t, arrivals, 3, raw)
	assert.Less(t, arrivals[0], 250*time.Millisecond,
		"first frame arrived at %v: the access log drained the stream before it was written:\n%s", arrivals[0], raw)
	assert.GreaterOrEqual(t, arrivals[2], 350*time.Millisecond, "third frame arrived before it could have been produced")
	assert.Contains(t, raw, "Transfer-Encoding: chunked", "a streamed body must stay chunked")
	assert.NotContains(t, raw, "Content-Length:", "a Content-Length means the stream was buffered")

	entries := logs.FilterMessage("http request").All()
	require.Len(t, entries, 1)
	fields := entries[0].ContextMap()
	assert.Equal(t, true, fields["streamed"], "a streamed response must be marked streamed")
	assert.NotContains(t, fields, "bytesSent", "bytesSent is unknown at log time for a stream")
	assert.EqualValues(t, fiber.StatusOK, fields["status"])
}

// newObservedKernel builds a kernel from env and opts whose access log writes
// into an observer, and whose kernel logger records calls.
func newObservedKernel(t *testing.T, env map[string]any, opts ...Option) (contract.HTTPKernel, *observer.ObservedLogs, *MockLogger) {
	t.Helper()
	core, logs := observer.New(zap.DebugLevel)
	logger := &MockLogger{}
	for _, method := range []string{"Debug", "Info", "Warn", "Error", "Fatal"} {
		logger.On(method, mock.Anything, mock.Anything).Return()
	}
	logger.On("GetLogger").Return(zap.New(core).Sugar())
	return New(newFakeConfig(env), logger, opts...), logs, logger
}

func accessLine(t *testing.T, logs *observer.ObservedLogs) map[string]any {
	t.Helper()
	var lines []observer.LoggedEntry
	for _, entry := range logs.All() {
		if strings.HasPrefix(entry.Message, "http ") {
			lines = append(lines, entry)
		}
	}
	require.Len(t, lines, 1)
	return lines[0].ContextMap()
}

func TestAccessLog_RedactsSensitiveQueryValues(t *testing.T) {
	kernel, logs, _ := newObservedKernel(t, nil)
	kernel.App().Get("/cb", func(c fiber.Ctx) error { return c.SendString("ok") })

	_, err := kernel.App().Test(httptest.NewRequest("GET", "/cb?code=abc&state=x&Token=t", nil))
	require.NoError(t, err)
	assert.Equal(t, "/cb?code=REDACTED&state=x&Token=REDACTED", accessLine(t, logs)["url"])
}

func TestAccessLog_RedactListReplacesTheDefault(t *testing.T) {
	kernel, logs, _ := newObservedKernel(t, nil, WithAccessLogRedactQuery("state"))
	kernel.App().Get("/cb", func(c fiber.Ctx) error { return c.SendString("ok") })
	_, err := kernel.App().Test(httptest.NewRequest("GET", "/cb?code=abc&state=x", nil))
	require.NoError(t, err)
	assert.Equal(t, "/cb?code=abc&state=REDACTED", accessLine(t, logs)["url"])

	kernel, logs, _ = newObservedKernel(t, map[string]any{"HTTP_ACCESS_LOG_REDACT_QUERY": "state"})
	kernel.App().Get("/cb", func(c fiber.Ctx) error { return c.SendString("ok") })
	_, err = kernel.App().Test(httptest.NewRequest("GET", "/cb?code=abc&state=x", nil))
	require.NoError(t, err)
	assert.Equal(t, "/cb?code=abc&state=REDACTED", accessLine(t, logs)["url"], "from the environment")
}

func TestAccessLog_CustomURL(t *testing.T) {
	kernel, logs, _ := newObservedKernel(t, nil,
		WithAccessLogURL(func(c fiber.Ctx) string { return c.Path() }))
	kernel.App().Get("/cb", func(c fiber.Ctx) error { return c.SendString("ok") })
	_, err := kernel.App().Test(httptest.NewRequest("GET", "/cb?email=a@b.c", nil))
	require.NoError(t, err)
	assert.Equal(t, "/cb", accessLine(t, logs)["url"])
}

func TestAccessLog_SkipPredicate(t *testing.T) {
	kernel, logs, _ := newObservedKernel(t, nil,
		WithAccessLogSkip(func(c fiber.Ctx) bool { return c.Path() == "/quiet" }))
	kernel.App().Get("/quiet", func(c fiber.Ctx) error { return c.SendString("ok") })
	kernel.App().Get("/loud", func(c fiber.Ctx) error { return c.SendString("ok") })

	_, err := kernel.App().Test(httptest.NewRequest("GET", "/quiet", nil))
	require.NoError(t, err)
	_, err = kernel.App().Test(httptest.NewRequest("GET", "/loud", nil))
	require.NoError(t, err)
	assert.Equal(t, "/loud", accessLine(t, logs)["url"])
}

func TestAccessLog_ErrorOnlyOn5xx(t *testing.T) {
	kernel, logs, _ := newObservedKernel(t, nil)
	kernel.App().Get("/boom", func(c fiber.Ctx) error { return errors.New("database down") })

	_, err := kernel.App().Test(httptest.NewRequest("GET", "/boom", nil))
	require.NoError(t, err)
	assert.Equal(t, "database down", accessLine(t, logs)["error"])

	kernel, logs, _ = newObservedKernel(t, nil)
	_, err = kernel.App().Test(httptest.NewRequest("GET", "/missing", nil))
	require.NoError(t, err)
	assert.NotContains(t, accessLine(t, logs), "error", "a 404 is not an error worth a field")
}

func TestRecover_LogsPanicWithStackAndRequestID(t *testing.T) {
	kernel, logs, logger := newObservedKernel(t, nil)
	kernel.App().Get("/panic", func(c fiber.Ctx) error { panic("kaboom") })

	resp, err := kernel.App().Test(httptest.NewRequest("GET", "/panic", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	var panics []map[string]any
	for _, call := range logger.Calls {
		if call.Method == "Error" && call.Arguments[0] == "Request panicked" {
			args, _ := call.Arguments[1].([]any)
			fields := map[string]any{}
			for i := 0; i+1 < len(args); i += 2 {
				fields[args[i].(string)] = args[i+1]
			}
			panics = append(panics, fields)
		}
	}
	require.Len(t, panics, 1)
	assert.Equal(t, "kaboom", fmt.Sprint(panics[0]["panic"]))
	assert.Contains(t, panics[0]["stack"], "accesslog_test.go")
	assert.NotEmpty(t, panics[0]["request_id"])
	assert.Equal(t, "/panic", panics[0]["route"])

	line := accessLine(t, logs)
	assert.Equal(t, int64(500), line["status"], "a panicked request is in the access log")
	assert.Contains(t, line["error"], "kaboom")
}

func TestRequestID_IgnoresInboundWhenNotTrusted(t *testing.T) {
	req := func(kernel contract.HTTPKernel) string {
		kernel.App().Get("/x", func(c fiber.Ctx) error { return c.SendString("ok") })
		r := httptest.NewRequest("GET", "/x", nil)
		r.Header.Set("X-Request-ID", "client-chosen")
		resp, err := kernel.App().Test(r)
		require.NoError(t, err)
		return resp.Header.Get("X-Request-ID")
	}

	trusted, _, _ := newObservedKernel(t, nil)
	assert.Equal(t, "client-chosen", req(trusted), "by default an inbound id is reused")

	untrusted, _, _ := newObservedKernel(t, nil, WithRequestIDTrustInbound(false))
	id := req(untrusted)
	assert.NotEmpty(t, id)
	assert.NotEqual(t, "client-chosen", id)

	fromEnv, _, _ := newObservedKernel(t, map[string]any{"HTTP_REQUEST_ID_TRUST_INBOUND": false})
	assert.NotEqual(t, "client-chosen", req(fromEnv))
}
