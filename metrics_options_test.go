package goe_test

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	goehttp "go.oease.dev/goe/v2/core/http"
	"go.oease.dev/goe/v2/core/metrics"
)

func freeLoopbackPort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = ln.Close() }()
	return ln.Addr().(*net.TCPAddr).Port
}

func scrapeMetrics(t *testing.T, port int) string {
	t.Helper()
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Get("http://127.0.0.1:" + strconv.Itoa(port) + "/metrics")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(body)
}

func TestMetrics_OptionsImplyTheModuleAndServeOnTheirPort(t *testing.T) {
	port := freeLoopbackPort(t)
	app := goe.New(goe.Options{
		Metrics: []metrics.Option{metrics.WithHost("127.0.0.1"), metrics.WithPort(port)},
	})
	require.NotNil(t, goe.Metrics(), "metrics options imply the module")

	ctx := context.Background()
	require.NoError(t, app.Start(ctx))
	defer func() { _ = app.Stop(ctx) }()
	assert.Contains(t, scrapeMetrics(t, port), "goe_build_info")
}

func TestMetrics_StreamsAreRecorded(t *testing.T) {
	port := freeLoopbackPort(t)
	srv := &serverAddr{}
	app := goe.New(goe.Options{
		HTTP:    shutdownTestHTTP(srv),
		Metrics: []metrics.Option{metrics.WithHost("127.0.0.1"), metrics.WithPort(port)},
		Invokers: []any{
			func(k contract.HTTPKernel) {
				k.App().Get("/events", goehttp.SSE(func(c fiber.Ctx, s *goehttp.EventStream) error {
					return s.Send(goehttp.Event{Data: "x"})
				}))
			},
		},
	})
	ctx := context.Background()
	require.NoError(t, app.Start(ctx))
	defer func() { _ = app.Stop(ctx) }()

	resp, err := (&http.Client{Timeout: 5 * time.Second}).Get("http://" + srv.get() + "/events")
	require.NoError(t, err)
	_, _ = io.Copy(io.Discard, bufio.NewReader(resp.Body))
	_ = resp.Body.Close()

	require.Eventually(t, func() bool {
		return strings.Contains(scrapeMetrics(t, port),
			`goe_http_stream_duration_seconds_count{outcome="completed",route="/events"} 1`)
	}, 2*time.Second, 20*time.Millisecond, "the stream is recorded")
}
