package http

import (
	"bufio"
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
