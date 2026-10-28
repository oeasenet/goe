package http

import (
	"net"
	"sync"

	"github.com/valyala/fasthttp"
)

// connTracker records the connections the server holds open, so a drain that
// runs out of time can close the ones still open instead of leaving them to
// requests that would outlive the modules they use.
type connTracker struct {
	mu    sync.Mutex
	conns map[net.Conn]struct{}
}

func newConnTracker() *connTracker {
	return &connTracker{conns: make(map[net.Conn]struct{})}
}

// track installs the tracker as the server's ConnState callback, keeping a
// callback already set. Fiber wraps whatever is set here when it starts serving.
func (t *connTracker) track(server *fasthttp.Server) {
	previous := server.ConnState
	server.ConnState = func(conn net.Conn, state fasthttp.ConnState) {
		t.observe(conn, state)
		if previous != nil {
			previous(conn, state)
		}
	}
}

func (t *connTracker) observe(conn net.Conn, state fasthttp.ConnState) {
	t.mu.Lock()
	defer t.mu.Unlock()
	switch state {
	case fasthttp.StateNew:
		t.conns[conn] = struct{}{}
	case fasthttp.StateClosed, fasthttp.StateHijacked:
		// A hijacked connection belongs to its handler (WebSocket upgrades),
		// and fasthttp reports nothing after it.
		delete(t.conns, conn)
	}
}

// closeAll closes every connection still open and returns how many there were.
func (t *connTracker) closeAll() int {
	t.mu.Lock()
	conns := make([]net.Conn, 0, len(t.conns))
	for conn := range t.conns {
		conns = append(conns, conn)
	}
	t.mu.Unlock()

	for _, conn := range conns {
		_ = conn.Close()
	}
	return len(conns)
}
