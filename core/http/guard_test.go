package http

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startGuarded serves a module with a catch-all route that echoes the routed
// path and reads the whole body. It returns the address and a count of the
// requests that reached the route.
func startGuarded(t *testing.T, opts ...Option) (string, *atomic.Int32) {
	t.Helper()
	mod, _, addr := newServingModule(t, opts...)
	handled := &atomic.Int32{}
	mod.Provide().App().All("/*", func(c fiber.Ctx) error {
		handled.Add(1)
		n, err := io.Copy(io.Discard, bodyReader(c))
		if err != nil {
			return err
		}
		return c.SendString(fmt.Sprintf("routed %s %d", c.Path(), n))
	})
	require.NoError(t, mod.OnStart(context.Background()))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = mod.OnStop(ctx)
	})
	return addr(), handled
}

// bodyReader reads a streamed body from its stream and a buffered one from memory.
func bodyReader(c fiber.Ctx) io.Reader {
	if c.Request().IsBodyStream() {
		return c.Request().BodyStream()
	}
	return bytes.NewReader(c.Body())
}

// raw sends a raw HTTP/1.1 request and returns the status code and body. The
// request is written in the background: a server may answer, and close the
// connection, before a large body is through.
func raw(t *testing.T, addr, request string) (int, string) {
	t.Helper()
	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	go func() { _, _ = conn.Write([]byte(request)) }()
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func get(path string) string {
	return "GET " + path + " HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n"
}

func TestGuard_RejectsDotSegments(t *testing.T) {
	addr, _ := startGuarded(t)
	for _, path := range []string{
		"/a/../b", "/a/%2e%2e/b", "/a/.%2E/b", "/./a", "/a/..", "/a/.", "/a/%2E",
		"/a/..%2Fb", "/a/%2e%2e%2fb", "/a/%2e%2e%5cb",
	} {
		status, _ := raw(t, addr, get(path))
		assert.Equal(t, fiber.StatusBadRequest, status, path)
	}

	// With UnescapePath, Fiber decodes %2F before routing, so the routed path
	// no longer shows the segment; the guard reads the raw one.
	t.Run("WithUnescapePath", func(t *testing.T) {
		addr, _ := startGuarded(t, WithUnescapePath(true))
		for _, path := range []string{"/a/..%2Fb", "/a/%2e%2e%2fb"} {
			status, _ := raw(t, addr, get(path))
			assert.Equal(t, fiber.StatusBadRequest, status, path)
		}
	})
}

func TestGuard_RoutesPathsWithoutDotSegments(t *testing.T) {
	addr, _ := startGuarded(t)
	for _, path := range []string{
		"/a//b", "/a/.well-known/x", "/a/b.c", "/..a", "/a/...", "/a/b..", "/.a", "/%2e%2e%2e",
	} {
		status, body := raw(t, addr, get(path))
		assert.Equal(t, fiber.StatusOK, status, path)
		assert.Contains(t, body, "routed", path)
	}
}

func TestGuard_DotSegmentsAllowedWhenOptedOut(t *testing.T) {
	addr, _ := startGuarded(t, WithAllowDotSegments(true))
	status, _ := raw(t, addr, get("/a/../b"))
	assert.Equal(t, fiber.StatusOK, status)
}

func TestGuard_DotSegmentsAllowedFromEnv(t *testing.T) {
	k, _ := newTestKernel(t, map[string]any{"HTTP_ALLOW_DOT_SEGMENTS": true})
	k.App().Get("/*", func(c fiber.Ctx) error { return c.SendString("routed") })
	req, err := http.NewRequest(fiber.MethodGet, "http://example.com/a/../b", nil)
	require.NoError(t, err)
	resp, err := k.App().Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func post(path string, body []byte, headers ...string) string {
	return "POST " + path + " HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n" +
		strings.Join(headers, "") +
		"Content-Length: " + strconv.Itoa(len(body)) + "\r\n\r\n" + string(body)
}

func TestGuard_NotStrictLeavesStreamingUnchanged(t *testing.T) {
	addr, _ := startGuarded(t, WithBodyLimit(1024))
	status, body := raw(t, addr, post("/upload", bytes.Repeat([]byte("x"), 64*1024)))
	assert.Equal(t, fiber.StatusOK, status, "without strict limits a streamed body passes BodyLimit")
	assert.Contains(t, body, "65536")
}

func TestGuard_StrictRejectsDeclaredLengthOverTheLimit(t *testing.T) {
	addr, handled := startGuarded(t, WithBodyLimit(1024), WithStrictBodyLimit(true))

	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	// Declare 10 MB and send only the start of it: the answer must not wait
	// for the rest.
	start := append([]byte("POST /exact HTTP/1.1\r\nHost: example.com\r\nContent-Length: 10485760\r\n\r\n"),
		bytes.Repeat([]byte("x"), 8192)...)
	_, err = conn.Write(start)
	require.NoError(t, err)
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	require.NoError(t, err)
	_ = resp.Body.Close()

	assert.Equal(t, fiber.StatusRequestEntityTooLarge, resp.StatusCode)
	assert.True(t, resp.Close, "the connection is closed instead of reading the body")
	assert.Zero(t, handled.Load(), "the handler must not run")
}

func TestGuard_StrictFromEnv(t *testing.T) {
	k, _ := newTestKernel(t, map[string]any{"HTTP_STRICT_BODY_LIMIT": true, "FIBER_BODY_LIMIT": 1024})
	k.App().Post("/x", func(c fiber.Ctx) error { return c.SendString("ok") })
	req, err := http.NewRequest(fiber.MethodPost, "http://example.com/x", bytes.NewReader(make([]byte, 4096)))
	require.NoError(t, err)
	resp, err := k.App().Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusRequestEntityTooLarge, resp.StatusCode)
}

func TestGuard_RouteLimitAdmitsALargerBody(t *testing.T) {
	addr, _ := startGuarded(t, WithBodyLimit(1024),
		WithRouteLimits("/api/files", RouteLimits{BodyLimit: 1 << 20}))

	status, body := raw(t, addr, post("/api/files/upload", bytes.Repeat([]byte("x"), 64*1024)))
	assert.Equal(t, fiber.StatusOK, status)
	assert.Contains(t, body, "65536")

	status, _ = raw(t, addr, post("/api/filesx", bytes.Repeat([]byte("x"), 64*1024)))
	assert.Equal(t, fiber.StatusRequestEntityTooLarge, status, "a prefix matches whole segments only")

	status, _ = raw(t, addr, post("/other", bytes.Repeat([]byte("x"), 64*1024)))
	assert.Equal(t, fiber.StatusRequestEntityTooLarge, status, "routes limits turn strict limits on")
}

func TestGuard_LongestRoutePrefixWins(t *testing.T) {
	addr, _ := startGuarded(t, WithBodyLimit(1024),
		WithRouteLimits("/api", RouteLimits{BodyLimit: 1 << 20}),
		WithRouteLimits("/api/small", RouteLimits{BodyLimit: 2048}))

	status, _ := raw(t, addr, post("/api/small/x", bytes.Repeat([]byte("x"), 4096)))
	assert.Equal(t, fiber.StatusRequestEntityTooLarge, status)
	status, _ = raw(t, addr, post("/api/big", bytes.Repeat([]byte("x"), 4096)))
	assert.Equal(t, fiber.StatusOK, status)
}

func TestGuard_ChunkedBodyOverTheLimitIs413(t *testing.T) {
	addr, _ := startGuarded(t, WithBodyLimit(1024), WithStrictBodyLimit(true))

	chunk := strings.Repeat("y", 4096)
	request := "POST /chunked HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\nTransfer-Encoding: chunked\r\n\r\n" +
		fmt.Sprintf("%x\r\n%s\r\n0\r\n\r\n", len(chunk), chunk)
	status, _ := raw(t, addr, request)
	assert.Equal(t, fiber.StatusRequestEntityTooLarge, status)

	small := strings.Repeat("y", 100)
	request = "POST /chunked HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\nTransfer-Encoding: chunked\r\n\r\n" +
		fmt.Sprintf("%x\r\n%s\r\n0\r\n\r\n", len(small), small)
	status, body := raw(t, addr, request)
	assert.Equal(t, fiber.StatusOK, status)
	assert.Contains(t, body, "100")
}

// multipartBody builds a form with one file part of the given size.
func multipartBody(t *testing.T, size int) ([]byte, string) {
	t.Helper()
	var form bytes.Buffer
	writer := multipart.NewWriter(&form)
	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": {`form-data; name="file"; filename="big.bin"`},
		"Content-Type":        {"application/octet-stream"},
	})
	require.NoError(t, err)
	_, err = part.Write(bytes.Repeat([]byte("z"), size))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	return form.Bytes(), writer.FormDataContentType()
}

func TestGuard_StrictParsesMultipartOnDemand(t *testing.T) {
	mod, _, addr := newServingModule(t, WithBodyLimit(1<<20), WithStrictBodyLimit(true))
	var unparsedOnArrival atomic.Bool
	mod.Provide().App().Post("/upload", func(c fiber.Ctx) error {
		// Parsed up front, the form would already sit in memory or on disk.
		unparsedOnArrival.Store(c.Request().IsBodyStream())
		file, err := c.FormFile("file")
		if err != nil {
			return err
		}
		return c.SendString(strconv.FormatInt(file.Size, 10))
	})
	require.NoError(t, mod.OnStart(context.Background()))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = mod.OnStop(ctx)
	})

	body, contentType := multipartBody(t, 64<<10)
	status, got := raw(t, addr(), post("/upload", body, "Content-Type: "+contentType+"\r\n"))
	require.Equal(t, fiber.StatusOK, status, got)
	assert.Equal(t, "65536", got, "the handler parses the form when it asks for it")
	assert.True(t, unparsedOnArrival.Load(), "strict limits must not parse multipart forms before handlers run")
}

func TestGuard_MultipartOverTheLimitIs413(t *testing.T) {
	addr, handled := startGuarded(t, WithBodyLimit(1<<20), WithStrictBodyLimit(true))
	body, contentType := multipartBody(t, 20<<20)
	status, _ := raw(t, addr, post("/upload", body, "Content-Type: "+contentType+"\r\n"))
	assert.Equal(t, fiber.StatusRequestEntityTooLarge, status)
	assert.Zero(t, handled.Load())
}

func TestGuard_RouteReadTimeoutLetsASlowUploadFinish(t *testing.T) {
	slowUpload := func(addr string) (int, error) {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return 0, err
		}
		defer func() { _ = conn.Close() }()
		_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
		// Past fasthttp's 8 KB read-ahead, so the rest is read by the handler.
		head := bytes.Repeat([]byte("h"), 8192)
		tail := []byte("tail-of-the-upload")
		header := "POST /uploads/slow HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\nContent-Length: " +
			strconv.Itoa(len(head)+len(tail)) + "\r\n\r\n"
		if _, err := conn.Write(append([]byte(header), head...)); err != nil {
			return 0, err
		}
		for _, b := range tail {
			time.Sleep(80 * time.Millisecond) // about 1.4 s in all
			if _, err := conn.Write([]byte{b}); err != nil {
				return 0, err
			}
		}
		resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
		if err != nil {
			return 0, err
		}
		_ = resp.Body.Close()
		return resp.StatusCode, nil
	}

	slow, _ := startGuarded(t, WithReadTimeout(500*time.Millisecond), WithStrictBodyLimit(true))
	status, err := slowUpload(slow)
	assert.True(t, err != nil || status != fiber.StatusOK, "the server read timeout cuts the slow upload off")

	patient, _ := startGuarded(t, WithReadTimeout(500*time.Millisecond),
		WithRouteLimits("/uploads", RouteLimits{ReadTimeout: 3 * time.Second}))
	status, err = slowUpload(patient)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, status, "a route read timeout lets the upload finish")
}
