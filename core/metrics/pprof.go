package metrics

import (
	"fmt"
	"html"
	"net/http"
	"runtime/pprof"
	"runtime/trace"
	"strconv"
	"strings"
	"time"
)

// registerPprof serves Go's profiles on mux under /debug/pprof/. It is written
// against runtime/pprof rather than net/http/pprof, whose import registers the
// same handlers on http.DefaultServeMux — exposing them on any server of the
// application that uses the default mux.
func registerPprof(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprofProfile)
	mux.HandleFunc("/debug/pprof/profile", pprofCPU)
	mux.HandleFunc("/debug/pprof/trace", pprofTrace)
}

// pprofProfile lists the profiles at /debug/pprof/ and serves one at
// /debug/pprof/<name>; ?debug=1 or 2 returns text.
func pprofProfile(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/debug/pprof/")
	if name == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, "<html><body><h1>/debug/pprof/</h1><ul>")
		for _, p := range pprof.Profiles() {
			n := html.EscapeString(p.Name())
			_, _ = fmt.Fprintf(w, `<li><a href="%s?debug=1">%s</a> (%d)</li>`, n, n, p.Count())
		}
		_, _ = fmt.Fprint(w, `<li><a href="profile">profile</a> (CPU, ?seconds=)</li>`+
			`<li><a href="trace">trace</a> (?seconds=)</li></ul></body></html>`)
		return
	}
	profile := pprof.Lookup(name)
	if profile == nil {
		http.Error(w, "unknown profile", http.StatusNotFound)
		return
	}
	level, _ := strconv.Atoi(r.URL.Query().Get("debug"))
	if level > 0 {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	}
	_ = profile.WriteTo(w, level)
}

// pprofCPU serves a CPU profile of ?seconds= (default 30).
func pprofCPU(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="profile"`)
	if err := pprof.StartCPUProfile(w); err != nil {
		http.Error(w, "could not start the CPU profile: "+err.Error(), http.StatusInternalServerError)
		return
	}
	sleep(r, seconds(r, 30))
	pprof.StopCPUProfile()
}

// pprofTrace serves an execution trace of ?seconds= (default 1).
func pprofTrace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="trace"`)
	if err := trace.Start(w); err != nil {
		http.Error(w, "could not start the trace: "+err.Error(), http.StatusInternalServerError)
		return
	}
	sleep(r, seconds(r, 1))
	trace.Stop()
}

func seconds(r *http.Request, fallback int) time.Duration {
	if n, err := strconv.Atoi(r.URL.Query().Get("seconds")); err == nil && n > 0 {
		return time.Duration(n) * time.Second
	}
	return time.Duration(fallback) * time.Second
}

func sleep(r *http.Request, d time.Duration) {
	select {
	case <-time.After(d):
	case <-r.Context().Done():
	}
}
