package http

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/utils/v2"
	"github.com/valyala/fasthttp"
	"go.oease.dev/goe/v2/contract"
)

// Long-lived responses — Server-Sent Events, proxied LLM streams — break on
// two server defaults. fasthttp applies WriteTimeout once to the whole
// response, so a stream longer than it is cut mid-way; and a handler's
// context is never cancelled, so it learns neither that the client left nor
// that the server is stopping. SSE and SendStream fix both: the write deadline
// moves forward before every write (each write gets WriteTimeout, the stream
// as long as it keeps writing), and the stream context is cancelled when the
// client is gone or the application starts stopping.

var (
	// ErrServerStopping is the cause of a stream's context when the
	// application started shutting down.
	ErrServerStopping = errors.New("goe: server is stopping")
	// ErrClientGone is the cause of a stream's context, and is wrapped by the
	// write error, when a write to the client failed: it disconnected or
	// stopped reading.
	ErrClientGone = errors.New("goe: stream client is gone")
	// ErrStreamClosed is returned by writes made after the stream ended.
	ErrStreamClosed = errors.New("goe: stream is closed")
)

// Stream outcomes, as logged when a stream ends and reported to a
// StreamObserver.
const (
	StreamCompleted      = "completed"
	StreamClientGone     = "client_gone"
	StreamServerStopping = "server_stopping"
	StreamFailed         = "error"
)

// StreamObserver is told when a stream served by SSE or SendStream begins and
// ends. GOE's metrics module supplies one when metrics are enabled.
type StreamObserver interface {
	StreamStarted(route string)
	StreamEnded(route, outcome string, d time.Duration)
}

// Keys under which the kernel shares stream dependencies through app.State().
const (
	stateStopping       = "goe.http.stopping"
	stateStreamLogger   = "goe.http.stream_logger"
	stateStreamObserver = "goe.http.stream_observer"
)

// StreamOption configures SSE and SendStream.
type StreamOption func(*streamSettings)

type streamSettings struct {
	heartbeat       time.Duration
	reconnectDelay  time.Duration
	writeTimeout    time.Duration
	writeTimeoutSet bool
}

// WithHeartbeat sets how often SSE writes a comment frame while the handler is
// quiet. Heartbeats keep proxies from closing idle streams and detect clients
// that left. Defaults to 15 seconds; zero or less disables them.
func WithHeartbeat(interval time.Duration) StreamOption {
	return func(s *streamSettings) { s.heartbeat = interval }
}

// WithReconnectDelay sets the retry hint SSE sends when the application starts
// stopping, telling clients how soon to reconnect — with Last-Event-ID, to
// another instance. Defaults to 1 second.
func WithReconnectDelay(d time.Duration) StreamOption {
	return func(s *streamSettings) { s.reconnectDelay = d }
}

// WithStreamWriteTimeout sets the time each write of the stream may take.
// Defaults to the server's WriteTimeout; zero disables write deadlines.
func WithStreamWriteTimeout(d time.Duration) StreamOption {
	return func(s *streamSettings) { s.writeTimeout = d; s.writeTimeoutSet = true }
}

func newStreamSettings(opts []StreamOption) streamSettings {
	s := streamSettings{heartbeat: 15 * time.Second, reconnectDelay: time.Second}
	for _, opt := range opts {
		if opt != nil {
			opt(&s)
		}
	}
	return s
}

// Event is one Server-Sent Event.
type Event struct {
	// ID sets the id field, which clients send back as Last-Event-ID when they
	// reconnect.
	ID string
	// Name sets the event field.
	Name string
	// Data is written as one data field per line. Strings, byte slices and
	// json.RawMessage are written as they are; other values are encoded with
	// the app's JSON encoder. Nil writes no data field.
	Data any
	// Retry sets the retry field (the reconnect delay) when positive.
	Retry time.Duration
}

// EventStream is a Server-Sent Events stream. Its methods are safe for
// concurrent use.
type EventStream struct {
	stream      *stream
	lastEventID string
	encode      utils.JSONMarshal
}

// Send writes ev and flushes it to the client.
func (s *EventStream) Send(ev Event) error {
	frame, err := encodeEvent(ev, s.encode)
	if err != nil {
		return err
	}
	return s.stream.send(frame, true)
}

// Comment writes a comment frame, which clients ignore.
func (s *EventStream) Comment(text string) error {
	return s.stream.send(encodeComment(text), false)
}

// Context is cancelled when the client is gone (cause ErrClientGone), the
// application starts stopping (cause ErrServerStopping), or the stream ends.
// Pass it to the work that produces the events, such as an upstream LLM call.
func (s *EventStream) Context() context.Context { return s.stream.ctx }

// LastEventID returns the Last-Event-ID a reconnecting client sent.
func (s *EventStream) LastEventID() string { return s.lastEventID }

// SSE returns a handler that serves a Server-Sent Events stream written by fn.
//
// fn runs after the handler returned, while the response is being written; it
// returns when the stream should end. Use s.Context() to stop producing events
// when the client leaves or the application stops. When the application starts
// stopping, SSE sends a retry hint (WithReconnectDelay) once fn returned, then
// ends the response; clients resume with Last-Event-ID elsewhere. A panic in fn
// is recovered and logged. One line is logged when the stream ends.
func SSE(fn func(c fiber.Ctx, s *EventStream) error, opts ...StreamOption) fiber.Handler {
	settings := newStreamSettings(opts)
	return func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextEventStream)
		c.Set(fiber.HeaderCacheControl, "no-cache")
		c.Set("X-Accel-Buffering", "no")
		lastEventID := c.Get(fiber.HeaderLastEventID)
		encode := c.App().Config().JSONEncoder

		return runStream(c, settings, true, func(st *stream) error {
			events := &EventStream{stream: st, lastEventID: lastEventID, encode: encode}
			stop := st.heartbeat(settings.heartbeat)
			defer stop()
			return fn(c, events)
		}, func(st *stream) {
			_ = st.write(retryFrame(settings.reconnectDelay), true)
		})
	}
}

// StreamWriter writes a raw streamed response. It is safe for concurrent use.
type StreamWriter struct {
	stream *stream
}

// Write buffers p; Flush sends what is buffered. The write deadline moves
// forward before every write that reaches the connection.
func (w *StreamWriter) Write(p []byte) (int, error) { return w.stream.writeN(p) }

// Flush sends the buffered bytes to the client.
func (w *StreamWriter) Flush() error { return w.stream.write(nil, true) }

// Context is cancelled when the client is gone, the application starts
// stopping, or the stream ends (see EventStream.Context).
func (w *StreamWriter) Context() context.Context { return w.stream.ctx }

// SendStream streams a response body written by fn — a proxied upstream stream,
// for example — with SSE's deadline and shutdown handling: each write gets the
// write timeout instead of the whole response, and w.Context() is cancelled when
// the client is gone or the application stops. Set the content type before
// calling it. fn runs after the handler returned; a panic in fn is recovered and
// logged, and one line is logged when the stream ends.
func SendStream(c fiber.Ctx, fn func(w *StreamWriter) error, opts ...StreamOption) error {
	settings := newStreamSettings(opts)
	return runStream(c, settings, false, func(st *stream) error {
		return fn(&StreamWriter{stream: st})
	}, nil)
}

// stream is what SSE and SendStream share: the writer, its deadline and its
// context.
type stream struct {
	rc      *fasthttp.RequestCtx
	timeout time.Duration
	ctx     context.Context
	cancel  context.CancelCauseFunc

	mu     sync.Mutex
	w      *bufio.Writer
	closed bool
	gone   bool
	bytes  int64
	events int64
}

// runStream starts a streamed response whose body runs body, and on shutdown
// runs onStopping before the response ends.
func runStream(c fiber.Ctx, settings streamSettings, countEvents bool, body func(st *stream) error, onStopping func(st *stream)) error {
	rc := c.RequestCtx()
	app := c.App()
	timeout := settings.writeTimeout
	if !settings.writeTimeoutSet {
		timeout = app.Config().WriteTimeout
	}
	// A deadline the handler sets keeps fasthttp from applying WriteTimeout to
	// the whole response; it is moved before every write.
	if timeout > 0 {
		_ = rc.SetWriteDeadline(time.Now().Add(timeout))
	}

	stopping, _ := fiber.GetState[<-chan struct{}](app.State(), stateStopping)
	logger, _ := fiber.GetState[contract.Logger](app.State(), stateStreamLogger)
	observer, _ := fiber.GetState[StreamObserver](app.State(), stateStreamObserver)
	route := c.Route().Path
	requestID := requestid.FromContext(c)

	// The writer runs after the handler returned; keep Fiber from reusing c.
	c.Abandon()

	return c.SendStreamWriter(func(w *bufio.Writer) {
		ctx, cancel := context.WithCancelCause(context.Background())
		st := &stream{rc: rc, timeout: timeout, ctx: ctx, cancel: cancel, w: w}
		start := time.Now()
		if observer != nil {
			observer.StreamStarted(route)
		}
		go st.watch(stopping, rc.Done())

		err, panicked := st.run(body, logger, requestID, route)

		outcome := StreamCompleted
		switch {
		case panicked:
			outcome = StreamFailed
		case st.clientGone():
			outcome = StreamClientGone
		case errors.Is(context.Cause(ctx), ErrServerStopping):
			outcome = StreamServerStopping
			if onStopping != nil {
				onStopping(st)
			}
		case err != nil && !errors.Is(err, context.Canceled):
			outcome = StreamFailed
		}
		st.close()

		d := time.Since(start)
		if observer != nil {
			observer.StreamEnded(route, outcome, d)
		}
		if logger == nil {
			return
		}
		fields := []any{
			"request_id", requestID,
			"route", route,
			"duration_ms", d.Milliseconds(),
			"bytes", st.bytes,
			"outcome", outcome,
		}
		if countEvents {
			fields = append(fields, "events", st.events)
		}
		if outcome == StreamFailed && err != nil {
			logger.Warn("Stream ended", append(fields, "error", err.Error())...)
			return
		}
		logger.Info("Stream ended", fields...)
	})
}

// run calls body, recovering a panic.
func (s *stream) run(body func(st *stream) error, logger contract.Logger, requestID, route string) (err error, panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
			if logger != nil {
				logger.Error("Stream handler panicked",
					"request_id", requestID,
					"route", route,
					"panic", r,
					"stack", string(debug.Stack()),
				)
			}
		}
	}()
	return body(s), false
}

// watch cancels the stream when the application starts stopping or the server
// shuts down, and returns when the stream ends.
func (s *stream) watch(stopping, shutdown <-chan struct{}) {
	select {
	case <-stopping:
		s.cancel(ErrServerStopping)
	case <-shutdown:
		s.cancel(ErrServerStopping)
	case <-s.ctx.Done():
	}
}

// heartbeat writes a comment frame every interval until the returned stop is
// called or the stream ends.
func (s *stream) heartbeat(interval time.Duration) (stop func()) {
	if interval <= 0 {
		return func() {}
	}
	quit, done := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if s.write(heartbeatFrame, true) != nil {
					return
				}
			case <-quit:
				return
			case <-s.ctx.Done():
				return
			}
		}
	}()
	return func() {
		close(quit)
		<-done
	}
}

var heartbeatFrame = []byte(":\n\n")

// send writes one whole frame and flushes it, counting it as an event when
// event is set.
func (s *stream) send(frame []byte, event bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.writeLocked(frame, true); err != nil {
		return err
	}
	if event {
		s.events++
	}
	return nil
}

func (s *stream) write(p []byte, flush bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.writeLocked(p, flush)
	return err
}

func (s *stream) writeN(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writeLocked(p, false)
}

// writeLocked writes p and optionally flushes, moving the write deadline
// before each step. A failure marks the client gone and cancels the context.
func (s *stream) writeLocked(p []byte, flush bool) (int, error) {
	switch {
	case s.closed:
		return 0, ErrStreamClosed
	case s.gone:
		return 0, ErrClientGone
	}
	var n int
	var err error
	if len(p) > 0 {
		s.extend()
		n, err = s.w.Write(p)
		s.bytes += int64(n)
	}
	if err == nil && flush {
		s.extend()
		err = s.w.Flush()
	}
	if err != nil {
		s.gone = true
		s.cancel(ErrClientGone)
		return n, fmt.Errorf("%w: %w", ErrClientGone, err)
	}
	return n, nil
}

func (s *stream) extend() {
	if s.timeout > 0 {
		_ = s.rc.SetWriteDeadline(time.Now().Add(s.timeout))
	}
}

func (s *stream) clientGone() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.gone
}

// close ends the stream: later writes fail and the context is cancelled.
func (s *stream) close() {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
	s.cancel(context.Canceled)
}

func retryFrame(d time.Duration) []byte {
	return []byte("retry: " + strconv.FormatInt(retryMilliseconds(d), 10) + "\n\n")
}

func retryMilliseconds(d time.Duration) int64 {
	if ms := d.Milliseconds(); ms > 0 {
		return ms
	}
	return 1
}

var errInvalidEventField = errors.New("goe: SSE id and event fields must not contain CR or LF")

func encodeEvent(ev Event, encode utils.JSONMarshal) ([]byte, error) {
	var frame bytes.Buffer
	for _, field := range [][2]string{{"id", ev.ID}, {"event", ev.Name}} {
		if field[1] == "" {
			continue
		}
		if strings.ContainsAny(field[1], "\r\n") {
			return nil, fmt.Errorf("%w (%s)", errInvalidEventField, field[0])
		}
		if value := strings.Trim(field[1], " "); value != "" {
			frame.WriteString(field[0] + ": " + value + "\n")
		}
	}
	if ev.Retry > 0 {
		frame.WriteString("retry: " + strconv.FormatInt(retryMilliseconds(ev.Retry), 10) + "\n")
	}

	var data string
	switch value := ev.Data.(type) {
	case nil:
	case string:
		data = value
	case []byte:
		data = string(value)
	case json.RawMessage:
		data = string(value)
	default:
		if encode == nil {
			encode = json.Marshal
		}
		encoded, err := encode(value)
		if err != nil {
			return nil, fmt.Errorf("goe: SSE data: %w", err)
		}
		data = string(encoded)
	}
	if ev.Data != nil {
		for line := range strings.SplitSeq(normalizeNewlines(data), "\n") {
			frame.WriteString("data: " + line + "\n")
		}
	}
	frame.WriteByte('\n')
	return frame.Bytes(), nil
}

func encodeComment(text string) []byte {
	text = normalizeNewlines(text)
	if text == "" {
		return heartbeatFrame
	}
	var frame bytes.Buffer
	for line := range strings.SplitSeq(text, "\n") {
		frame.WriteString(": " + line + "\n")
	}
	frame.WriteByte('\n')
	return frame.Bytes()
}

func normalizeNewlines(s string) string {
	if !strings.ContainsRune(s, '\r') {
		return s
	}
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
}
