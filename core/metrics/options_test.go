package metrics

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
)

// warnRecorder keeps the warnings the module logs.
type warnRecorder struct {
	*testLogger
	mu    sync.Mutex
	warns []string
}

func (r *warnRecorder) Warn(msg string, args ...any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.warns = append(r.warns, msg)
}

func (r *warnRecorder) With(...any) contract.Logger { return r }

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = ln.Close() }()
	return ln.Addr().(*net.TCPAddr).Port
}

func fetch(t *testing.T, url string) (int, string) {
	t.Helper()
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Get(url)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, string(body)
}

func startModule(t *testing.T, m *Module) {
	t.Helper()
	require.NoError(t, m.OnStart(context.Background()))
	t.Cleanup(func() { _ = m.OnStop(context.Background()) })
}

func scrape(t *testing.T, m *Module) string {
	t.Helper()
	rec := httptest.NewRecorder()
	m.Manager().Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	return rec.Body.String()
}

func TestModule_ServesOnASeparatePort(t *testing.T) {
	port := freePort(t)
	m := NewModule(newTestConfig(), &testLogger{t: t}, WithHost("127.0.0.1"), WithPort(port))
	app := fiber.New()
	m.RegisterRoutes(app)
	startModule(t, m)

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/metrics", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "metrics leave the app's port when they have their own")

	status, body := fetch(t, "http://127.0.0.1:"+strconv.Itoa(port)+"/metrics")
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, body, "goe_build_info")
}

func TestModule_PortFromEnv(t *testing.T) {
	port := freePort(t)
	cfg := newTestConfig()
	cfg.values["METRICS_PORT"] = port
	cfg.values["METRICS_HOST"] = "127.0.0.1"
	startModule(t, NewModule(cfg, &testLogger{t: t}))

	status, _ := fetch(t, "http://127.0.0.1:"+strconv.Itoa(port)+"/metrics")
	assert.Equal(t, http.StatusOK, status)
}

func TestModule_PortInUseFailsStart(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = ln.Close() }()
	port := ln.Addr().(*net.TCPAddr).Port

	m := NewModule(newTestConfig(), &testLogger{t: t}, WithHost("127.0.0.1"), WithPort(port))
	err = m.OnStart(context.Background())
	assert.ErrorContains(t, err, "metrics")
}

func TestModule_InvalidOptionsFailStart(t *testing.T) {
	for name, opt := range map[string]Option{
		"port out of range": WithPort(70000),
		"nil registry":      WithRegistry(nil),
	} {
		m := NewModule(newTestConfig(), &testLogger{t: t}, opt)
		assert.Error(t, m.OnStart(context.Background()), name)
	}
}

func TestModule_UsesTheInjectedRegistry(t *testing.T) {
	reg := prometheus.NewRegistry()
	jobs := prometheus.NewCounter(prometheus.CounterOpts{Name: "app_jobs_total", Help: "Jobs run."})
	reg.MustRegister(jobs)
	jobs.Inc()

	m := NewModule(newTestConfig(), &testLogger{t: t}, WithRegistry(reg))
	m.Manager().Counter("orders_total", "Orders.").Inc()

	body := scrape(t, m)
	assert.Contains(t, body, "app_jobs_total 1", "the app's own collectors are served")
	assert.Contains(t, body, "goe_orders_total 1", "GOE's metrics go into the app's registry")

	families, err := reg.Gather()
	require.NoError(t, err)
	var names []string
	for _, f := range families {
		names = append(names, f.GetName())
	}
	assert.Contains(t, names, "goe_build_info")
}

// The application's registry may hold collectors GOE registers too, set up
// differently: its own build_info, a Go collector with more runtime metrics.
// GOE keeps them, leaves its own out with a warning, and still serves.
func TestModule_InjectedRegistryWithConflictingCollectors(t *testing.T) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector(collectors.WithGoCollectorRuntimeMetrics(collectors.MetricsAll)))
	appInfo := prometheus.NewGauge(prometheus.GaugeOpts{Name: "goe_build_info", Help: "The app's own."})
	reg.MustRegister(appInfo)
	appInfo.Set(1)

	rec := &warnRecorder{testLogger: &testLogger{t: t}}
	var m *Module
	require.NotPanics(t, func() { m = NewModule(newTestConfig(), rec, WithRegistry(reg)) })
	m.Manager().Counter("orders_total", "Orders.").Inc()

	body := scrape(t, m)
	assert.Contains(t, body, "goe_orders_total 1")
	assert.Contains(t, body, "goe_build_info 1", "the application's collector stays")
	assert.Contains(t, rec.warns, "Metrics collector not registered; the registry holds a conflicting one")
}

func TestModule_BuildInfo(t *testing.T) {
	cfg := newTestConfig()
	cfg.values["APP_VERSION"] = "1.4.2"
	body := scrape(t, NewModule(cfg, &testLogger{t: t}))
	assert.Contains(t, body, `go_version="`+runtime.Version()+`"`)
	assert.Contains(t, body, `version="1.4.2"`)
	assert.Regexp(t, `goe_build_info\{[^}]*goe_version="[^"]+"`, body)
}

func TestModule_PprofOnlyOnTheSeparatePort(t *testing.T) {
	port := freePort(t)
	m := NewModule(newTestConfig(), &testLogger{t: t}, WithHost("127.0.0.1"), WithPort(port), WithPprof(true))
	startModule(t, m)
	base := "http://127.0.0.1:" + strconv.Itoa(port)

	status, body := fetch(t, base+"/debug/pprof/")
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, body, "goroutineleak", "Go 1.27's goroutine leak profile is listed")
	status, body = fetch(t, base+"/debug/pprof/goroutine?debug=1")
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, body, "goroutine profile")

	noPort := &warnRecorder{testLogger: &testLogger{t: t}}
	startModule(t, NewModule(newTestConfig(), noPort, WithPprof(true)))
	assert.Contains(t, noPort.warns, "Metrics pprof needs a separate port (METRICS_PORT); not served")
}

func TestStreamMetrics_TracksActiveStreamsAndDurations(t *testing.T) {
	m := NewModule(newTestConfig(), &testLogger{t: t})
	streams := NewStreamMetrics(m.Manager())

	streams.StreamStarted("/events")
	assert.Contains(t, scrape(t, m), `goe_http_streams_active{route="/events"} 1`)

	streams.StreamEnded("/events", "completed", 2*time.Second)
	body := scrape(t, m)
	assert.Contains(t, body, `goe_http_streams_active{route="/events"} 0`)
	assert.True(t, strings.Contains(body,
		`goe_http_stream_duration_seconds_count{outcome="completed",route="/events"} 1`),
		fmt.Sprintf("duration histogram missing:\n%s", body))
}
