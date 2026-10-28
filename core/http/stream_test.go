package http

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// serveStreams starts a module whose routes are registered by setup before it
// starts, and stops it when the test ends.
func serveStreams(t *testing.T, setup func(app *fiber.App), opts ...Option) (*Module, *MockLogger, string) {
	t.Helper()
	mod, logger, addr := newServingModule(t, opts...)
	setup(mod.Provide().App())
	require.NoError(t, mod.OnStart(context.Background()))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = mod.OnStop(ctx)
	})
	return mod, logger, "http://" + addr()
}

type sseClient struct {
	resp *http.Response
	r    *bufio.Reader
}

func openSSE(t *testing.T, url string, header ...string) *sseClient {
	t.Helper()
	req, err := http.NewRequest(fiber.MethodGet, url, nil)
	require.NoError(t, err)
	for i := 0; i+1 < len(header); i += 2 {
		req.Header.Set(header[i], header[i+1])
	}
	// A timeout turns a stream that never ends into a failure, not a hang.
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	return &sseClient{resp: resp, r: bufio.NewReader(resp.Body)}
}

// next returns the lines of the next frame, up to the blank line ending it.
func (s *sseClient) next() ([]string, error) {
	var lines []string
	for {
		line, err := s.r.ReadString('\n')
		if err != nil {
			return lines, err
		}
		line = strings.TrimSuffix(line, "\n")
		if line == "" {
			return lines, nil
		}
		lines = append(lines, line)
	}
}

// dataFrames reads frames until the stream ends and returns their data lines.
func (s *sseClient) dataFrames() (data []string, err error) {
	for {
		frame, err := s.next()
		for _, line := range frame {
			if value, ok := strings.CutPrefix(line, "data: "); ok {
				data = append(data, value)
			}
		}
		if err != nil {
			return data, err
		}
	}
}

// logCalls returns the key/value fields of every call to method with msg.
func logCalls(logger *MockLogger, method, msg string) []map[string]any {
	var out []map[string]any
	for _, call := range logger.Calls {
		if call.Method != method || len(call.Arguments) < 2 || call.Arguments[0] != msg {
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

func TestSSE_OutlivesTheWriteTimeout(t *testing.T) {
	_, _, base := serveStreams(t, func(app *fiber.App) {
		app.Get("/events", SSE(func(c fiber.Ctx, s *EventStream) error {
			for i := range 10 {
				if err := s.Send(Event{Data: fmt.Sprint(i)}); err != nil {
					return err
				}
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		}, WithHeartbeat(0)))
		app.Get("/plain", func(c fiber.Ctx) error {
			return c.SendStreamWriter(func(w *bufio.Writer) {
				for i := range 10 {
					_, _ = fmt.Fprintf(w, "data: %d\n\n", i)
					if w.Flush() != nil {
						return
					}
					time.Sleep(100 * time.Millisecond)
				}
			})
		})
	}, WithWriteTimeout(300*time.Millisecond))

	plain, _ := openSSE(t, base+"/plain").dataFrames()
	assert.Less(t, len(plain), 10, "control: a plain stream is cut at the write timeout")

	events, err := openSSE(t, base+"/events").dataFrames()
	assert.ErrorIs(t, err, io.EOF)
	assert.Len(t, events, 10, "every write gets the write timeout, not the whole stream")
}

func TestSSE_SendsHeartbeats(t *testing.T) {
	_, _, base := serveStreams(t, func(app *fiber.App) {
		app.Get("/events", SSE(func(c fiber.Ctx, s *EventStream) error {
			time.Sleep(300 * time.Millisecond)
			return nil
		}, WithHeartbeat(50*time.Millisecond)))
	})

	client := openSSE(t, base+"/events")
	assert.Equal(t, "text/event-stream", client.resp.Header.Get("Content-Type"))
	assert.Equal(t, "no-cache", client.resp.Header.Get("Cache-Control"))
	assert.Equal(t, "no", client.resp.Header.Get("X-Accel-Buffering"))

	heartbeats := 0
	for {
		frame, err := client.next()
		if len(frame) == 1 && frame[0] == ":" {
			heartbeats++
		}
		if err != nil {
			break
		}
	}
	assert.GreaterOrEqual(t, heartbeats, 3)
}

func TestSSE_ClientDisconnectCancelsTheContext(t *testing.T) {
	causes := make(chan error, 1)
	_, _, base := serveStreams(t, func(app *fiber.App) {
		app.Get("/events", SSE(func(c fiber.Ctx, s *EventStream) error {
			if err := s.Send(Event{Data: "first"}); err != nil {
				return err
			}
			<-s.Context().Done()
			causes <- context.Cause(s.Context())
			return nil
		}, WithHeartbeat(50*time.Millisecond)))
	})

	client := openSSE(t, base+"/events")
	frame, err := client.next()
	require.NoError(t, err)
	require.Equal(t, []string{"data: first"}, frame)
	_ = client.resp.Body.Close()

	select {
	case cause := <-causes:
		assert.ErrorIs(t, cause, ErrClientGone)
	case <-time.After(2 * time.Second):
		t.Fatal("a disconnected client did not cancel the stream context")
	}
}

func TestSSE_StoppingEndsTheStreamWithAReconnectHint(t *testing.T) {
	stopping := make(chan struct{})
	causes := make(chan error, 1)
	_, _, base := serveStreams(t, func(app *fiber.App) {
		app.Get("/events", SSE(func(c fiber.Ctx, s *EventStream) error {
			if err := s.Send(Event{Data: "hello"}); err != nil {
				return err
			}
			<-s.Context().Done()
			causes <- context.Cause(s.Context())
			return nil
		}, WithHeartbeat(0), WithReconnectDelay(250*time.Millisecond)))
	}, WithStoppingSignal(stopping))

	client := openSSE(t, base+"/events")
	frame, err := client.next()
	require.NoError(t, err)
	require.Equal(t, []string{"data: hello"}, frame)

	close(stopping)
	frame, err = client.next()
	require.NoError(t, err)
	assert.Equal(t, []string{"retry: 250"}, frame, "the client is told to reconnect")
	_, err = client.next()
	assert.ErrorIs(t, err, io.EOF, "then the stream ends")

	select {
	case cause := <-causes:
		assert.ErrorIs(t, cause, ErrServerStopping)
	case <-time.After(time.Second):
		t.Fatal("stopping did not cancel the stream context")
	}
}

func TestSSE_ExposesLastEventID(t *testing.T) {
	_, _, base := serveStreams(t, func(app *fiber.App) {
		app.Get("/events", SSE(func(c fiber.Ctx, s *EventStream) error {
			return s.Send(Event{Data: s.LastEventID()})
		}))
	})
	data, _ := openSSE(t, base+"/events", "Last-Event-ID", "41").dataFrames()
	assert.Equal(t, []string{"41"}, data)
}

func TestSSE_WritesEventFields(t *testing.T) {
	_, _, base := serveStreams(t, func(app *fiber.App) {
		app.Get("/events", SSE(func(c fiber.Ctx, s *EventStream) error {
			if err := s.Send(Event{ID: "a\nb"}); err == nil {
				return errors.New("an id with a newline was accepted")
			}
			return s.Send(Event{ID: "7", Name: "delta", Data: "a\r\nb", Retry: 2 * time.Second})
		}, WithHeartbeat(0)))
	})
	frame, err := openSSE(t, base+"/events").next()
	require.NoError(t, err)
	assert.Equal(t, []string{"id: 7", "event: delta", "retry: 2000", "data: a", "data: b"}, frame)
}

func TestSSE_JSONDataUsesTheAppEncoder(t *testing.T) {
	_, _, base := serveStreams(t, func(app *fiber.App) {
		app.Get("/events", SSE(func(c fiber.Ctx, s *EventStream) error {
			return s.Send(Event{Data: map[string]int{"n": 1}})
		}))
	}, WithJSONEncoder(func(any) ([]byte, error) { return []byte(`{"encoder":"app"}`), nil }))
	data, _ := openSSE(t, base+"/events").dataFrames()
	assert.Equal(t, []string{`{"encoder":"app"}`}, data)
}

func TestSSE_PanicIsRecoveredAndLogged(t *testing.T) {
	_, logger, base := serveStreams(t, func(app *fiber.App) {
		app.Get("/events", SSE(func(c fiber.Ctx, s *EventStream) error {
			_ = s.Send(Event{Data: "before"})
			panic("kaboom")
		}))
	})
	data, err := openSSE(t, base+"/events").dataFrames()
	assert.Equal(t, []string{"before"}, data)
	assert.ErrorIs(t, err, io.EOF)

	require.Eventually(t, func() bool { return len(logCalls(logger, "Error", "Stream handler panicked")) == 1 },
		time.Second, 10*time.Millisecond)
	fields := logCalls(logger, "Error", "Stream handler panicked")[0]
	assert.Equal(t, "kaboom", fmt.Sprint(fields["panic"]))
	assert.Contains(t, fields["stack"], "stream_test.go")
	assert.NotEmpty(t, fields["request_id"])
}

func TestSSE_LogsWhenTheStreamEnds(t *testing.T) {
	_, logger, base := serveStreams(t, func(app *fiber.App) {
		app.Get("/events/:topic", SSE(func(c fiber.Ctx, s *EventStream) error {
			_ = s.Send(Event{Data: "one"})
			return s.Send(Event{Data: "two"})
		}))
	})
	_, _ = openSSE(t, base+"/events/news").dataFrames()

	require.Eventually(t, func() bool { return len(logCalls(logger, "Info", "Stream ended")) == 1 },
		time.Second, 10*time.Millisecond)
	fields := logCalls(logger, "Info", "Stream ended")[0]
	assert.Equal(t, "completed", fields["outcome"])
	assert.Equal(t, "/events/:topic", fields["route"])
	assert.Equal(t, int64(2), fields["events"])
	assert.Positive(t, fields["bytes"])
	assert.Contains(t, fields, "duration_ms")
	assert.NotEmpty(t, fields["request_id"])
}

func TestSendStream_ProxiesChunksPastTheWriteTimeout(t *testing.T) {
	_, _, base := serveStreams(t, func(app *fiber.App) {
		app.Get("/proxy", func(c fiber.Ctx) error {
			return SendStream(c, func(w *StreamWriter) error {
				for i := range 10 {
					if _, err := fmt.Fprintf(w, "chunk-%d\n", i); err != nil {
						return err
					}
					if err := w.Flush(); err != nil {
						return err
					}
					time.Sleep(100 * time.Millisecond)
				}
				return nil
			})
		})
	}, WithWriteTimeout(300*time.Millisecond))

	resp, err := http.Get(base + "/proxy")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, 10, strings.Count(string(body), "chunk-"))
}

type recordingObserver struct {
	mu      sync.Mutex
	started []string
	ended   []string
}

func (o *recordingObserver) StreamStarted(route string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.started = append(o.started, route)
}

func (o *recordingObserver) StreamEnded(route, outcome string, d time.Duration) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if d > 0 {
		o.ended = append(o.ended, route+" "+outcome)
	}
}

func TestStreams_ReportToTheObserver(t *testing.T) {
	observer := &recordingObserver{}
	_, _, base := serveStreams(t, func(app *fiber.App) {
		app.Get("/events", SSE(func(c fiber.Ctx, s *EventStream) error {
			return s.Send(Event{Data: "x"})
		}))
	}, WithStreamObserver(observer))
	_, _ = openSSE(t, base+"/events").dataFrames()

	require.Eventually(t, func() bool {
		observer.mu.Lock()
		defer observer.mu.Unlock()
		return len(observer.ended) == 1
	}, time.Second, 10*time.Millisecond)
	assert.Equal(t, []string{"/events"}, observer.started)
	assert.Equal(t, []string{"/events completed"}, observer.ended)
}

func TestSSE_HandlerIgnoringContextEndsAtDrain(t *testing.T) {
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	mod, _, base := serveStreams(t, func(app *fiber.App) {
		app.Get("/events", SSE(func(c fiber.Ctx, s *EventStream) error {
			_ = s.Send(Event{Data: "first"})
			<-release // ignores its context
			return nil
		}, WithHeartbeat(0)))
	})

	client := openSSE(t, base+"/events")
	_, err := client.next()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	start := time.Now()
	require.NoError(t, mod.OnStop(ctx))
	assert.Less(t, time.Since(start), time.Second)

	_, err = client.next()
	assert.Error(t, err, "the drain end closes the stream the handler would not end")
}

func TestSSE_ConcurrentSendsAndHeartbeatDoNotInterleave(t *testing.T) {
	_, _, base := serveStreams(t, func(app *fiber.App) {
		app.Get("/events", SSE(func(c fiber.Ctx, s *EventStream) error {
			var wg sync.WaitGroup
			for g := range 10 {
				wg.Go(func() {
					for i := range 50 {
						_ = s.Send(Event{Data: fmt.Sprintf("%d-%d", g, i)})
					}
				})
			}
			wg.Wait()
			return nil
		}, WithHeartbeat(time.Millisecond)))
	})

	client := openSSE(t, base+"/events")
	wellFormed := regexp.MustCompile(`^data: \d+-\d+$`)
	data := 0
	for {
		frame, err := client.next()
		switch {
		case len(frame) == 0:
		case len(frame) == 1 && frame[0] == ":":
		case len(frame) == 1 && wellFormed.MatchString(frame[0]):
			data++
		default:
			t.Fatalf("malformed frame %q", frame)
		}
		if err != nil {
			break
		}
	}
	assert.Equal(t, 500, data)
}
