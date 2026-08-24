package goe

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"go.oease.dev/goe/v2/core/log"
)

// TestOptions_LogCodeFirst_VersionWired proves Options.Log survives the option
// merge in New and reaches log.NewModule: a WithVersion option must show up as
// the version base field on JSON log lines.
func TestOptions_LogCodeFirst_VersionWired(t *testing.T) {
	// The logger captures os.Stdout at construction, so the pipe must be in
	// place before New runs.
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	_ = New(Options{
		Log:             []log.Option{log.WithVersion("wired-9.9.9")},
		ConfigOverrides: map[string]any{"LOG_FORMAT": "json"},
	})
	Log().Info("wiring probe")

	os.Stdout = old
	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	// Clean up the singleton like the other root tests.
	instance.app = nil
	instance.config = nil
	instance.logger = nil
	instance.http = nil

	var probe map[string]any
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		var entry map[string]any
		if json.Unmarshal([]byte(line), &entry) != nil {
			continue
		}
		if entry["msg"] == "wiring probe" {
			probe = entry
			break
		}
	}
	if probe == nil {
		t.Fatalf("wiring probe line not found in output:\n%s", buf.String())
	}
	if got := probe["version"]; got != "wired-9.9.9" {
		t.Errorf("version = %v, want wired-9.9.9 (Options.Log not wired through)", got)
	}
}
